package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AstraProtocol/astra/v3/cmd/config"
	feeburntype "github.com/AstraProtocol/astra/v3/x/feeburn/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"github.com/stretchr/testify/mock"
	"math"
	"math/big"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v12/encoding"
	evmConfig "github.com/evmos/evmos/v12/server/config"
	ethermint "github.com/evmos/evmos/v12/types"
	"github.com/evmos/evmos/v12/x/evm/statedb"
	evm "github.com/evmos/evmos/v12/x/evm/types"
	feemarkettypes "github.com/evmos/evmos/v12/x/feemarket/types"

	"github.com/AstraProtocol/astra/v3/app"
	"github.com/AstraProtocol/astra/v3/contracts"
	"github.com/evmos/evmos/v12/x/erc20/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx              sdk.Context
	app              *app.Astra
	queryClientEvm   evm.QueryClient
	queryClient      types.QueryClient
	address          common.Address
	consAddress      sdk.ConsAddress
	validator        stakingtypes.Validator
	clientCtx        client.Context
	ethSigner        ethtypes.Signer
	signer           keyring.Signer
	mintFeeCollector bool
}

var (
	contract  common.Address
	contract2 common.Address
)

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

func (suite *KeeperTestSuite) setupRegisterIBCVoucher() (banktypes.Metadata, *types.TokenPair) {
	suite.SetupTest()

	validMetadata := banktypes.Metadata{
		Description: "ATOM IBC voucher (channel 14)",
		Base:        ibcBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcBase,
				Exponent: 0,
			},
		},
		Name:    "ATOM channel-14",
		Symbol:  "ibcATOM-14",
		Display: ibcBase,
	}

	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
	suite.Require().NoError(err)

	// pair := types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
	pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, validMetadata)
	suite.Require().NoError(err)
	suite.Commit()
	return validMetadata, pair
}

// DoSetupTest Test helpers
func (suite *KeeperTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = utiltx.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	// setup feemarketGenesis paramsFeeBurn
	feemarketGenesis := feemarkettypes.DefaultGenesisState()
	feemarketGenesis.Params.EnableHeight = 1
	feemarketGenesis.Params.NoBaseFee = false

	suite.app = app.Setup(false, feemarkettypes.DefaultGenesisState())

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         app.TestnetChainID + "-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelperEvm, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelperEvm)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.Erc20Keeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	paramsFeeBurn := feeburntype.DefaultParams()
	paramsFeeBurn.EnableFeeBurn = true
	suite.app.FeeBurnKeeper.SetParams(suite.ctx, paramsFeeBurn)

	stakingParams := suite.app.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = config.BaseDenom
	suite.app.StakingKeeper.SetParams(suite.ctx, stakingParams)

	mintParams := suite.app.MintKeeper.GetParams(suite.ctx)
	mintParams.MintDenom = config.BaseDenom
	suite.app.MintKeeper.SetParams(suite.ctx, mintParams)

	evmParams := suite.app.EvmKeeper.GetParams(suite.ctx)
	evmParams.EvmDenom = config.BaseDenom
	suite.app.EvmKeeper.SetParams(suite.ctx, evmParams)

	// Set Validator
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	validators := suite.app.StakingKeeper.GetValidators(suite.ctx, 1)
	suite.validator = validators[0]

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 1))}})
	require.NoError(t, err)
	suite.Commit()
	feePoolBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 1))}}
	err = suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, feePoolBalance)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, feePoolBalance)
	suite.Require().NoError(err)
	// Deploy contracts
	contract, err = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	require.NoError(t, err)
	contract2, err = suite.DeployContract(erc20Name2, erc20Symbol2, erc20Decimals)
	require.NoError(t, err)
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *KeeperTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash().Bytes())))
}

func (suite *KeeperTestSuite) MintFeeCollector(coins sdk.Coins) {
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, types.ModuleName, authtypes.FeeCollectorName, coins)
	suite.Require().NoError(err)
}

// DeployContract deploys the ERC20MinterBurnerDecimalsContract.
func (suite *KeeperTestSuite) DeployContract(name, symbol string, decimals uint8) (common.Address, error) {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("", name, symbol, decimals)
	if err != nil {
		return common.Address{}, err
	}

	data := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	if err != nil {
		return common.Address{}, err
	}

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: evmConfig.DefaultGasCap,
	})
	if err != nil {
		return common.Address{}, err
	}

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTx(&evm.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  res.Gas,
		Input:     data,
		GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		GasPrice:  nil,
		ChainID:   chainID,
		Amount:    nil,
		GasTipCap: big.NewInt(1),
		Accesses:  &ethtypes.AccessList{},
	})

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	if err != nil {
		return common.Address{}, err
	}

	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	if err != nil {
		return common.Address{}, err
	}

	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce), nil
}

func (suite *KeeperTestSuite) DeployContractMaliciousDelayed(name string, symbol string) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20MaliciousDelayedContract.ABI.Pack("", big.NewInt(1000000000000000000))
	suite.Require().NoError(err)

	data := append(contracts.ERC20MaliciousDelayedContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(evmConfig.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTx(&evm.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  res.Gas,
		Input:     data,
		GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		GasPrice:  nil,
		ChainID:   chainID,
		Amount:    nil,
		GasTipCap: big.NewInt(1),
		Accesses:  &ethtypes.AccessList{},
	})

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) DeployContractDirectBalanceManipulation(name string, symbol string) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contracts.ERC20DirectBalanceManipulationContract.ABI.Pack("", big.NewInt(1000000000000000000))
	suite.Require().NoError(err)

	data := append(contracts.ERC20DirectBalanceManipulationContract.Bin, ctorArgs...)
	args, err := json.Marshal(&evm.TransactionArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(evmConfig.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	erc20DeployTx := evm.NewTx(&evm.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  res.Gas,
		Input:     data,
		GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
		GasPrice:  nil,
		ChainID:   chainID,
		Amount:    nil,
		GasTipCap: big.NewInt(1),
		Accesses:  &ethtypes.AccessList{},
	})

	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) Commit() {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()
	header.Height += 1
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) MintERC20Token(contractAddr, from, to common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("mint", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) BurnERC20Token(contractAddr, from common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", types.ModuleAddress, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) GrantERC20Token(contractAddr, from, to common.Address, role_string string) *evm.MsgEthereumTx {
	// 0xCc508cD0818C85b8b8a1aB4cEEef8d981c8956A6 MINTER_ROLE
	role := crypto.Keccak256([]byte(role_string))
	// needs to be an array not a slice
	var v [32]byte
	copy(v[:], role)

	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("grantRole", v, to)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

func (suite *KeeperTestSuite) sendTx(contractAddr, from common.Address, transferData []byte) *evm.MsgEthereumTx {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	args, err := json.Marshal(&evm.TransactionArgs{To: &contractAddr, From: &from, Data: (*hexutil.Bytes)(&transferData)})
	suite.Require().NoError(err)
	res, err := suite.queryClientEvm.EstimateGas(ctx, &evm.EthCallRequest{
		Args:   args,
		GasCap: uint64(evmConfig.DefaultGasCap),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address)

	// Mint the max gas to the FeeCollector to ensure balance in case of refund
	suite.MintFeeCollector(sdk.NewCoins(sdk.NewCoin(config.BaseDenom, sdk.NewInt(suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx).Int64()*int64(res.Gas)))))

	ercTransferTx := evm.NewTx(
		&evm.EvmTxArgs{
			Nonce:     nonce,
			GasLimit:  res.Gas,
			Input:     transferData,
			GasFeeCap: suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx),
			GasPrice:  nil,
			ChainID:   chainID,
			Amount:    nil,
			GasTipCap: big.NewInt(1),
			To:        nil,
			Accesses:  &ethtypes.AccessList{},
		},
	)

	ercTransferTx.From = suite.address.Hex()
	err = ercTransferTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, ercTransferTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return ercTransferTx
}

func (suite *KeeperTestSuite) BalanceOf(contract, account common.Address) interface{} {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, false, "balanceOf", account)
	if err != nil {
		return nil
	}

	unpacked, err := erc20.Unpack("balanceOf", res.Ret)
	if len(unpacked) == 0 {
		return nil
	}

	return unpacked[0]
}

func (suite *KeeperTestSuite) NameOf(contract common.Address) string {
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	res, err := suite.app.Erc20Keeper.CallEVM(suite.ctx, erc20, types.ModuleAddress, contract, false, "name")
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	unpacked, err := erc20.Unpack("name", res.Ret)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(unpacked)

	return fmt.Sprintf("%v", unpacked[0])
}

func (suite *KeeperTestSuite) TransferERC20Token(contractAddr, from, to common.Address, amount *big.Int) *evm.MsgEthereumTx {
	transferData, err := contracts.ERC20MinterBurnerDecimalsContract.ABI.Pack("transfer", to, amount)
	suite.Require().NoError(err)
	return suite.sendTx(contractAddr, from, transferData)
}

const (
	contractMinterBurner = iota + 1
	contractDirectBalanceManipulation
	contractMaliciousDelayed
)

const (
	erc20Name          = "Coin Token"
	erc20Symbol        = "CTKN"
	erc20Name2         = "Coin Token2"
	erc20Symbol2       = "CTKN2"
	erc20Decimals      = uint8(18)
	cosmosTokenBase    = "acoin"
	cosmosTokenDisplay = "coin"
	cosmosDecimals     = uint8(6)
	defaultExponent    = uint32(18)
	zeroExponent       = uint32(0)
	ibcBase            = "ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2"
)

func (suite *KeeperTestSuite) setupRegisterERC20TwoPair(contractType int) (common.Address, common.Address) {
	suite.SetupTest()

	var contractAddr common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contractAddr = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol)
	case contractMaliciousDelayed:
		contractAddr = suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	default:
		contractAddr, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}
	suite.Commit()

	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	var contractAddr2 common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contractAddr2 = suite.DeployContractDirectBalanceManipulation(erc20Name2, erc20Symbol2)
	case contractMaliciousDelayed:
		contractAddr2 = suite.DeployContractMaliciousDelayed(erc20Name2, erc20Symbol2)
	default:
		contractAddr2, _ = suite.DeployContract(erc20Name2, erc20Symbol2, erc20Decimals)
	}
	suite.Commit()

	_, err = suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr2)
	suite.Require().NoError(err)
	return contractAddr, contractAddr2
}

func (suite *KeeperTestSuite) setupRegisterERC20TwoPairSameErc20NameSymbol(contractType int) (common.Address, common.Address) {
	suite.SetupTest()

	var contractAddr common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contractAddr = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol)
	case contractMaliciousDelayed:
		contractAddr = suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	default:
		contractAddr, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}
	suite.Commit()

	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
	suite.Require().NoError(err)

	var contractAddr2 common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contractAddr2 = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol)
	case contractMaliciousDelayed:
		contractAddr2 = suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	default:
		contractAddr2, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}
	suite.Commit()

	_, err = suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr2)
	suite.Require().NoError(err)
	return contractAddr, contractAddr2
}

func (suite *KeeperTestSuite) setupRegisterERC20Pair(contractType int) common.Address {
	suite.SetupTest()

	var contractAddr common.Address
	// Deploy contract
	switch contractType {
	case contractDirectBalanceManipulation:
		contractAddr = suite.DeployContractDirectBalanceManipulation(erc20Name, erc20Symbol)
	case contractMaliciousDelayed:
		contractAddr = suite.DeployContractMaliciousDelayed(erc20Name, erc20Symbol)
	default:
		contractAddr, _ = suite.DeployContract(erc20Name, erc20Symbol, erc20Decimals)
	}
	suite.Commit()

	_, err := suite.app.Erc20Keeper.RegisterERC20(suite.ctx, contractAddr)
	suite.Require().NoError(err)
	return contractAddr
}

func (suite *KeeperTestSuite) setupRegisterCoin() (banktypes.Metadata, *types.TokenPair) {
	suite.SetupTest()
	validMetadata := banktypes.Metadata{
		Description: "description of the token",
		Base:        cosmosTokenBase,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    cosmosTokenBase,
				Exponent: 0,
			},
			{
				Denom:    cosmosTokenBase[1:],
				Exponent: uint32(18),
			},
		},
		Name:    cosmosTokenBase,
		Symbol:  erc20Symbol,
		Display: cosmosTokenBase,
	}

	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, sdk.Coins{sdk.NewInt64Coin(validMetadata.Base, 1)})
	suite.Require().NoError(err)

	// pair := types.NewTokenPair(contractAddr, cosmosTokenBase, true, types.OWNER_MODULE)
	pair, err := suite.app.Erc20Keeper.RegisterCoin(suite.ctx, validMetadata)
	suite.Require().NoError(err)
	suite.Commit()
	return validMetadata, pair
}

// CommitAfter Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	_ = suite.app.Commit()
	header := suite.ctx.BlockHeader()
	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	evm.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClientEvm = evm.NewQueryClient(queryHelper)
}

var _ types.EVMKeeper = &MockEVMKeeper{}

type MockEVMKeeper struct {
	mock.Mock
}

func (m *MockEVMKeeper) GetParams(ctx sdk.Context) evm.Params {
	args := m.Called(mock.Anything)
	return args.Get(0).(evm.Params)
}

func (m *MockEVMKeeper) GetAccountWithoutBalance(ctx sdk.Context, addr common.Address) *statedb.Account {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*statedb.Account)
}

func (m *MockEVMKeeper) EstimateGas(c context.Context, req *evm.EthCallRequest) (*evm.EstimateGasResponse, error) {
	args := m.Called(mock.Anything, mock.Anything)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evm.EstimateGasResponse), args.Error(1)
}

func (m *MockEVMKeeper) ApplyMessage(ctx sdk.Context, msg core.Message, tracer vm.EVMLogger, commit bool) (*evm.MsgEthereumTxResponse, error) {
	args := m.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*evm.MsgEthereumTxResponse), args.Error(1)
}

var _ types.BankKeeper = &MockBankKeeper{}

type MockBankKeeper struct {
	mock.Mock
}

func (b *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	args := b.Called(mock.Anything, mock.Anything, mock.Anything)
	return args.Error(0)
}

func (b *MockBankKeeper) IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	args := b.Called(mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool) {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Get(0).(banktypes.Metadata), args.Bool(1)
}

func (b *MockBankKeeper) SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata) {

}

func (b *MockBankKeeper) HasSupply(ctx sdk.Context, denom string) bool {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Bool(0)
}

func (b *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	args := b.Called(mock.Anything, mock.Anything)
	return args.Get(0).(sdk.Coin)
}

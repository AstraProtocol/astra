package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/app"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"math"
	"math/big"
)

var (
	DEFAULT_FEE int64 = 1000000
	privs       []*ethsecp256k1.PrivKey
)

func (suite *KeeperTestSuite) TestBurnFeeCosmosTxDelegate() {
	suite.SetupTest()
	priv0, _ := ethsecp256k1.GenerateKey()
	addr := getAddr(priv0)
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	s.Commit()
	suite.Require().NoError(err)
	totalSupplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupply", totalSupplyBefore)
	delegateAmount := sdk.NewCoin(config.BaseDenom, sdk.NewInt(1))
	delegate(priv0, delegateAmount)
	s.Commit()
	mintedCoin := getMintedCoin()
	totalSupplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupplyAfter", totalSupplyAfter)
	expectAmount := totalSupplyAfter.Amount.Sub(totalSupplyBefore.Amount)
	expectAmount = mintedCoin.Amount.Sub(expectAmount)
	fmt.Println("expectAmount", expectAmount)
	s.Require().Equal(getExpectTotalFeeBurn(1), expectAmount)
}

func (suite *KeeperTestSuite) TestBurnFeeCosmosTxSend() {
	suite.SetupTest()
	accountCount := 50
	for i := 0; i < accountCount; i++ {
		priv, _ := ethsecp256k1.GenerateKey()
		privs = append(privs, priv)
		addr := getAddr(priv)
		accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
		err := suite.FundAccount(suite.ctx, addr, accBalance)
		s.Require().NoError(err)
	}
	s.Commit()
	totalSupplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupply", totalSupplyBefore)
	sendAmount := sdk.NewCoin(config.BaseDenom, sdk.NewInt(10))
	for i := 0; i < accountCount; i++ {
		send(privs[i], sendAmount)
	}
	s.Commit()

	totalSupplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupplyAfter", totalSupplyAfter)
	expectAmount := totalSupplyAfter.Amount.Sub(totalSupplyBefore.Amount)
	mintedCoin := getMintedCoin()
	expectAmount = mintedCoin.Amount.Sub(expectAmount)
	fmt.Println("expectAmount", expectAmount)
	s.Require().Equal(getExpectTotalFeeBurn(accountCount), expectAmount)
	s.Commit()
	totalSupplyAfter1 := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupplyAfter1", totalSupplyAfter1)

	s.Require().Equal(getMintedCoin(), totalSupplyAfter1.Sub(totalSupplyAfter))
}

func (suite *KeeperTestSuite) TestEvmTx() {
	suite.SetupTest()
	priv, _ := ethsecp256k1.GenerateKey()
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	addr := getAddr(priv)
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	s.Require().NoError(err)
	sendEthToSelf(priv)
}

func getExpectTotalFeeBurn(numberTx int) sdk.Int {
	feeBurnParams := s.app.FeeBurnKeeper.GetParams(s.ctx)
	return feeBurnParams.FeeBurn.MulInt(sdk.NewInt(DEFAULT_FEE).Mul(sdk.NewInt(int64(numberTx)))).RoundInt()
}

func getMintedCoin() sdk.Coin {
	mintParams := s.app.MintKeeper.GetParams(s.ctx)
	return s.app.MintKeeper.GetMinter(s.ctx).BlockProvision(mintParams)
}

func getAddr(priv *ethsecp256k1.PrivKey) sdk.AccAddress {
	return priv.PubKey().Address().Bytes()
}

func delegate(priv *ethsecp256k1.PrivKey, delegateAmount sdk.Coin) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(s.validator.OperatorAddress)
	s.Require().NoError(err)

	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, delegateAmount)
	deliverTx(priv, delegateMsg)
}

func send(priv *ethsecp256k1.PrivKey, sendAmount sdk.Coin) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	privNew, _ := ethsecp256k1.GenerateKey()
	addrRecv := getAddr(privNew)

	amount := sdk.Coins{sendAmount}
	sendMsg := banktypes.NewMsgSend(accountAddress, addrRecv, amount)
	deliverTx(priv, sendMsg, sendMsg)
}

func prepareCosmosTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	gasPrice := sdk.NewInt(1)
	fees := &sdk.Coins{{Denom: config.BaseDenom, Amount: gasPrice.MulRaw(DEFAULT_FEE)}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

func sendEthToSelf(priv *ethsecp256k1.PrivKey) {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	msgEthereumTx := evmtypes.NewTx(chainID, nonce, &from, nil, 100000, nil,
		s.app.FeeMarketKeeper.GetBaseFee(s.ctx), big.NewInt(1), nil, &ethtypes.AccessList{})
	msgEthereumTx.From = from.String()
	performEthTx(priv, msgEthereumTx)
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	// Assemble transaction from fields
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	tx, err := msgEthereumTx.BuildTx(txBuilder, config.BaseDenom)
	s.Require().NoError(err)

	// Encode transaction by default Tx encoder and broadcasted over the network
	txEncoder := encodingConfig.TxConfig.TxEncoder()
	bz, err := txEncoder(tx)
	s.Require().NoError(err)

	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	fmt.Println(res)
	s.Require().Equal(true, res.IsOK())
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

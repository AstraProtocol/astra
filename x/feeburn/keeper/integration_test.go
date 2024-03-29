package keeper_test

import (
	"fmt"
	utiltx "github.com/evmos/evmos/v12/testutil/tx"
	"math"
	"math/big"

	"github.com/AstraProtocol/astra/v3/app"
	"github.com/AstraProtocol/astra/v3/cmd/config"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	"github.com/evmos/evmos/v12/encoding"
	evmtypes "github.com/evmos/evmos/v12/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	DEFAULT_FEE int64 = 1000000000000
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

func (suite *KeeperTestSuite) TestBurnFeeCosmosTxDelegateFail() {
	suite.SetupTest()
	priv0, _ := ethsecp256k1.GenerateKey()
	addr := getAddr(priv0)
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	feeParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	feeParams.MinGasPrice = sdk.NewDec(10_000_000_000_000) // > DEFAULT_FEE
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, feeParams)
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

	// transaction send fail, expect fee burn = 0
	s.Require().Equal(true, expectAmount.Equal(sdk.NewInt(0)))
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
	totalEvmFeeBurn := sdk.NewInt(21875000000000)
	priv, _ := ethsecp256k1.GenerateKey()
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	addr := getAddr(priv)
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	s.Require().NoError(err)
	s.Commit()
	totalSupplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupply", totalSupplyBefore)
	res := sendEth(priv)
	s.Require().Equal(true, res.IsOK())
	s.Commit()
	totalSupplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	expectAmount := totalSupplyAfter.Amount.Sub(totalSupplyBefore.Amount)
	s.Require().Equal(getMintedCoin().Amount.Sub(totalEvmFeeBurn), expectAmount)
}

func (suite *KeeperTestSuite) TestEvmTxFail() {
	suite.SetupTest()
	priv, _ := ethsecp256k1.GenerateKey()
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(16) * 1))}}
	addr := getAddr(priv)
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	s.Require().NoError(err)
	feeParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	feeParams.MinGasPrice = sdk.NewDec(50_000_000_000_000) // > DEFAULT_FEE
	feeParams.BaseFee = sdk.NewInt(100_000_000_000)        // > DEFAULT_FEE
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, feeParams)
	s.Commit()
	totalFeeBurnBefore := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
	res := sendEth(priv)
	s.Require().Equal(false, res.IsOK())
	s.Commit()
	totalFeeBurnAfter := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
	fmt.Println("totalFeeBurnAfter", totalFeeBurnAfter)
	s.Require().Equal(true, totalFeeBurnAfter.Equal(totalFeeBurnBefore))

}

func (suite *KeeperTestSuite) TestEvmTxWhenDisableBurnFee() {
	suite.SetupTest()
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	params.EnableFeeBurn = false
	suite.app.FeeBurnKeeper.SetParams(suite.ctx, params)

	priv, _ := ethsecp256k1.GenerateKey()
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	addr := getAddr(priv)
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	s.Require().NoError(err)
	s.Commit()
	totalSupplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupply", totalSupplyBefore)
	res := sendEth(priv)
	s.Require().Equal(true, res.IsOK())
	s.Commit()
	totalSupplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	expectAmount := totalSupplyAfter.Amount.Sub(totalSupplyBefore.Amount)
	s.Require().Equal(getMintedCoin().Amount, expectAmount)
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

func sendEth(priv *ethsecp256k1.PrivKey) abci.ResponseDeliverTx {

	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := s.app.EvmKeeper.GetNonce(s.ctx, from)

	privNew, _ := ethsecp256k1.GenerateKey()
	addrRecv := common.BytesToAddress(privNew.PubKey().Address().Bytes())

	msgEthereumTx := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		Nonce:     nonce,
		GasLimit:  100000,
		Input:     nil,
		GasFeeCap: s.app.FeeMarketKeeper.GetBaseFee(s.ctx),
		GasPrice:  big.NewInt(10000),
		ChainID:   s.app.EvmKeeper.ChainID(),
		Amount:    big.NewInt(100),
		GasTipCap: big.NewInt(1),
		To:        &addrRecv,
		Accesses:  &ethtypes.AccessList{},
	})
	msgEthereumTx.From = from.String()
	return performEthTx(priv, msgEthereumTx)
}

func performEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	// Sign transaction
	err := msgEthereumTx.Sign(s.ethSigner, utiltx.NewSigner(priv))
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
	return res
}

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

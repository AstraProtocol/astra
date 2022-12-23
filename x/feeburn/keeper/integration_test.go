package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/app"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/encoding"
	abci "github.com/tendermint/tendermint/abci/types"
	"math"
)

var (
	DEFAULT_FEE int64 = 1000000
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
	mintParams := suite.app.MintKeeper.GetParams(suite.ctx)
	mintedCoin := suite.app.MintKeeper.GetMinter(suite.ctx).BlockProvision(mintParams)
	totalSupplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	fmt.Println("totalSupplyAfter", totalSupplyAfter)
	feeBurnParams := suite.app.FeeBurnKeeper.GetParams(s.ctx)
	expectAmount := totalSupplyAfter.Amount.Sub(totalSupplyBefore.Amount)
	expectAmount = mintedCoin.Amount.Sub(expectAmount)
	fmt.Println("expectAmount", expectAmount)
	s.Require().Equal(feeBurnParams.FeeBurn.MulInt(sdk.NewInt(DEFAULT_FEE)).RoundInt(), expectAmount)
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

func deliverTx(priv *ethsecp256k1.PrivKey, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

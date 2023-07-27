package ante_test

import (
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/feeburn/ante"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/evmos/v12/crypto/ethsecp256k1"
	"math"
)

func (suite *AnteTestSuite) TestFeeBurnDecorator() {
	suite.SetupTest(false) // reset

	fbd := ante.NewFeeBurnDecorator(suite.app.BankKeeper, suite.app.FeeBurnKeeper)
	antehandler := sdk.ChainAnteDecorators(fbd)

	priv, _ := ethsecp256k1.GenerateKey()
	addr := getAddr(priv)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	privNew, _ := ethsecp256k1.GenerateKey()
	addrRecv := getAddr(privNew)

	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	suite.Require().NoError(err)

	sendAmount := sdk.NewCoin(config.BaseDenom, sdk.NewInt(10))
	amount := sdk.Coins{sendAmount}
	sendMsg := banktypes.NewMsgSend(accountAddress, addrRecv, amount)
	txBuilder := prepareCosmosTx(priv, sendMsg)
	_, err = antehandler(suite.ctx, txBuilder.GetTx(), false)

	suite.Require().NoError(err, "Did not error on invalid tx")
}

func (suite *AnteTestSuite) TestFeeBurnDecoratorWhenTxNull() {
	suite.SetupTest(false) // reset

	fbd := ante.NewFeeBurnDecorator(suite.app.BankKeeper, suite.app.FeeBurnKeeper)
	antehandler := sdk.ChainAnteDecorators(fbd)

	priv, _ := ethsecp256k1.GenerateKey()
	addr := getAddr(priv)
	accBalance := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(math.Pow10(18) * 2))}}
	err := suite.FundAccount(suite.ctx, addr, accBalance)
	suite.Require().NoError(err)
	_, err = antehandler(suite.ctx, nil, false)
	suite.Require().Error(err, "Tx must be a FeeTx")
}

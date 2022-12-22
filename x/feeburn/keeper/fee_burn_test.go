package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (suite *KeeperTestSuite) TestBurnErrorWhenFeeCollectorIsZeroAmount() {
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(1000)}}
	err := suite.app.FeeBurnKeeper.FeeBurnPayout(suite.ctx, suite.app.BankKeeper, totalFees, params)
	suite.Require().Error(err, types.ErrFeeBurnSend)
}

func (suite *KeeperTestSuite) TestBurnSuccess() {
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(1000)}}
	supplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	// send coin to feecollector module
	err := suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, totalFees)
	suite.Require().NoError(err)
	err = suite.app.FeeBurnKeeper.FeeBurnPayout(suite.ctx, suite.app.BankKeeper, totalFees, params)
	suite.Require().NoError(err)

	supplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
	suite.Require().Equal(supplyBefore.Amount.Sub(totalFees.AmountOf(config.BaseDenom).Quo(sdk.NewInt(2))), supplyAfter.Amount)

}

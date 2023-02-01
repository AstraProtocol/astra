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
	err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFees, params)
	suite.Require().Error(err, types.ErrFeeBurnSend)
}

func (suite *KeeperTestSuite) TestBurnWhenDisableBurnFee() {
	paramsUpdate := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	paramsUpdate.EnableFeeBurn = false
	suite.app.FeeBurnKeeper.SetParams(suite.ctx, paramsUpdate)
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	for i := 0; i < 10000; i++ {
		totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(i))}}
		err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFees, params)
		suite.Require().NoError(err, types.ErrFeeBurnSend)
	}
}

func (suite *KeeperTestSuite) TestBurnWhenTotalFeeIsZero() {
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(0)}}
	err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFees, params)
	suite.Require().NoError(err, types.ErrFeeBurnSend)
}

func (suite *KeeperTestSuite) TestBurnSuccess() {
	for j := 0; j < 10; j++ {
		newParams := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
		newParams.FeeBurn = sdk.NewDec(int64(j)).Quo(sdk.NewDec(int64(10)))
		suite.app.FeeBurnKeeper.SetParams(suite.ctx, newParams)
		for i := 0; i < 10000; i++ {
			params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
			totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(i))}}
			supplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
			totalFeeBurnBefore := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			// send coin to feecollector module
			if i > 0 {
				err := suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, totalFees)
				suite.Require().NoError(err)
			}
			err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFees, params)
			suite.Require().NoError(err)
			totalFeeBurnAfter := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			supplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
			totalFeeBurn := params.FeeBurn.MulInt(sdk.NewInt(int64(i))).RoundInt()
			suite.Require().Equal(supplyBefore.Amount.Sub(totalFeeBurn), supplyAfter.Amount)
			suite.Require().Equal(true, totalFeeBurn.ToDec().Equal(totalFeeBurnAfter.Sub(totalFeeBurnBefore)))
		}
	}
}

func (suite *KeeperTestSuite) TestBurnWhenFeeNegative() {
	for j := 0; j < 10; j++ {
		newParams := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
		newParams.FeeBurn = sdk.NewDec(int64(j)).Quo(sdk.NewDec(int64(10)))
		suite.app.FeeBurnKeeper.SetParams(suite.ctx, newParams)
		for i := 0; i < 10000; i++ {
			params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
			totalFeesNegative := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(-i))}}
			totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: sdk.NewInt(int64(i))}}
			supplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
			totalFeeBurnBefore := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			// send coin to feecollector module
			if i > 0 {
				err := suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, totalFees)
				suite.Require().NoError(err)
			}
			err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFeesNegative, params)
			suite.Require().NoError(err)
			totalFeeBurnAfter := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			supplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)

			// when fee is negative then fee burn is zero
			totalFeeBurn := params.FeeBurn.MulInt(sdk.NewInt(int64(0))).RoundInt()
			suite.Require().Equal(supplyBefore.Amount.Sub(totalFeeBurn), supplyAfter.Amount)
			suite.Require().Equal(true, totalFeeBurn.ToDec().Equal(totalFeeBurnAfter.Sub(totalFeeBurnBefore)))
		}
	}
}

func (suite *KeeperTestSuite) TestBurnWhenManyFeeDenom() {
	for j := 0; j < 10; j++ {
		newParams := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
		newParams.FeeBurn = sdk.NewDec(int64(j)).Quo(sdk.NewDec(int64(10)))
		suite.app.FeeBurnKeeper.SetParams(suite.ctx, newParams)
		for i := 0; i < 10000; i++ {
			params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
			denom := config.BaseDenom
			if i%2 == 0 {
				denom = "test"
			}
			totalFees := sdk.Coins{{Denom: denom, Amount: sdk.NewInt(int64(i))}}
			supplyBefore := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
			totalFeeBurnBefore := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			// send coin to feecollector module
			if i%2 == 1 {
				err := suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, totalFees)
				suite.Require().NoError(err)
			}
			err := suite.app.FeeBurnKeeper.BurnFee(suite.ctx, suite.app.BankKeeper, totalFees, params)
			suite.Require().NoError(err)
			totalFeeBurnAfter := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
			supplyAfter := suite.app.BankKeeper.GetSupply(suite.ctx, config.BaseDenom)
			totalFeeBurn := params.FeeBurn.MulInt(sdk.NewInt(int64(i))).RoundInt()
			if i%2 == 0 {
				totalFeeBurn = params.FeeBurn.MulInt(sdk.NewInt(int64(0))).RoundInt()
			}
			suite.Require().Equal(supplyBefore.Amount.Sub(totalFeeBurn), supplyAfter.Amount)
			suite.Require().Equal(true, totalFeeBurn.ToDec().Equal(totalFeeBurnAfter.Sub(totalFeeBurnBefore)))
		}
	}
}

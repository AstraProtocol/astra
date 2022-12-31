package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestTotalFeeBurn() {
	totalFeeBurn := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
	suite.Require().Equal(sdk.NewDec(0), totalFeeBurn)
	totalFeeBurn = totalFeeBurn.Add(sdk.NewDec(2))

	suite.app.FeeBurnKeeper.SetTotalFeeBurn(suite.ctx, totalFeeBurn)
	newTotalFeeBurn := suite.app.FeeBurnKeeper.GetTotalFeeBurn(suite.ctx)
	suite.Require().Equal(sdk.NewDec(2), newTotalFeeBurn)

}

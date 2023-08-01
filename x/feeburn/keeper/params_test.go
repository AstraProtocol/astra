package keeper_test

import (
	"github.com/AstraProtocol/astra/v3/x/feeburn/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	params.EnableFeeBurn = true
	suite.Require().Equal(types.DefaultParams(), params)

	enableFeeBurn := suite.app.FeeBurnKeeper.EnableFeeBurn(suite.ctx)
	suite.Require().Equal(true, enableFeeBurn)

	feeBurn := suite.app.FeeBurnKeeper.FeeBurn(suite.ctx)
	suite.Require().Equal(params.FeeBurn, feeBurn)

	params.EnableFeeBurn = false
	suite.app.FeeBurnKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)

	enableFeeBurn = suite.app.FeeBurnKeeper.EnableFeeBurn(suite.ctx)
	suite.Require().Equal(false, enableFeeBurn)
}

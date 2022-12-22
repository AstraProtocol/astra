package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	params.EnableFeeBurn = true
	suite.Require().Equal(types.DefaultParams(), params)
	params.EnableFeeBurn = false
	suite.app.FeeBurnKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeeBurnKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}

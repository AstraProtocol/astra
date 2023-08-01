package keeper_test

import (
	"github.com/AstraProtocol/astra/v3/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestQueryParamsError() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	res, err := suite.app.FeeBurnKeeper.Params(ctx, nil)
	suite.Require().Error(err)
	suite.Require().Nil(res)
}

func (suite *KeeperTestSuite) TestQueryParamsSuccess() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()
	expParams.EnableFeeBurn = true

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

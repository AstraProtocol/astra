package keeper_test

import (
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryTotalFeeBurn() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	totalFeeBurn := sdk.NewDec(100000000000)
	suite.app.FeeBurnKeeper.SetTotalFeeBurn(suite.ctx, totalFeeBurn)

	res, err := suite.queryClient.TotalFeeBurn(ctx, &types.QueryTotalFeeBurnRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.TotalFeeBurn.Amount, totalFeeBurn)
}

package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v3/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	for i := 0; i < numTests; i++ {
		params := randomizedValidParams()
		suite.app.MintKeeper.SetParams(suite.ctx, params)

		res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(res.Params, params)
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryInflation() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	for i := 0; i < numTests; i++ {
		minter := randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, minter)

		res, err := suite.queryClient.Inflation(ctx, &types.QueryInflationRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(res.Inflation, minter.Inflation)
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryAnnualProvisions() {
	for i := 0; i < numTests; i++ {
		suite.SetupTest()
		ctx := sdk.WrapSDKContext(suite.ctx)

		minter := randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, minter)

		res, err := suite.queryClient.AnnualProvisions(ctx, &types.QueryAnnualProvisionsRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(res.AnnualProvisions, minter.AnnualProvisions)
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryTotalMintedProvision() {
	var (
		req    *types.QueryTotalMintedProvisionRequest
		expRes *types.QueryTotalMintedProvisionResponse
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"genesis value",
			func() {
				req = &types.QueryTotalMintedProvisionRequest{}
				expRes = &types.QueryTotalMintedProvisionResponse{
					TotalMintedProvision: sdk.NewDecCoinFromDec(denomMint, sdk.NewDec(0)),
				}
			},
		},
		{
			"newly set total minted provision",
			func() {
				newTotalMintedProvision := sdk.NewDec(1_000_000)
				suite.app.MintKeeper.SetTotalMintProvision(suite.ctx, newTotalMintedProvision)
				suite.CommitAndBeginBlock()

				req = &types.QueryTotalMintedProvisionRequest{}
				expRes = &types.QueryTotalMintedProvisionResponse{
					TotalMintedProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, newTotalMintedProvision)}
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.TotalMintedProvision(ctx, req)
			suite.Require().NoError(err)
			suite.Require().Equal(expRes, res)
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryBlockProvision() {
	for i := 0; i < numTests; i++ {
		suite.SetupTest()
		ctx := sdk.WrapSDKContext(suite.ctx)

		minter := randomMinter()
		suite.app.MintKeeper.SetMinter(suite.ctx, minter)

		params := randomizedValidParams()
		suite.app.MintKeeper.SetParams(suite.ctx, params)

		res, err := suite.queryClient.BlockProvision(ctx, &types.QueryBlockProvisionRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(res.Provision, minter.BlockProvision(suite.app.MintKeeper.GetParams(suite.ctx)))
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryCirculatingSupply() {
	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)
	oldSupply := suite.app.MintKeeper.StakingTokenSupply(suite.ctx)
	baseAmount, _ := sdk.NewIntFromString("1200000000000000000000000000")
	for i := 0; i < numTests; i++ {
		expectedMint := randRate(sdk.ZeroDec(), sdk.OneDec()).MulInt(baseAmount).TruncateInt()
		err := suite.app.MintKeeper.MintCoins(
			suite.ctx,
			sdk.NewCoin(types.DefaultInflationDenom, expectedMint),
		)
		suite.Require().NoError(err)

		res, err := suite.queryClient.CirculatingSupply(ctx, &types.QueryCirculatingSupplyRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(res.CirculatingSupply.Amount.TruncateInt(), oldSupply.Add(expectedMint))

		oldSupply = res.CirculatingSupply.Amount.TruncateInt()
	}
}

func (suite *KeeperTestSuite) TestGRPCBondedRatio() {
	initialSupply, _ := sdk.NewIntFromString("1200000000000000000000000000")
	for i := 0; i < numTests; i++ {
		suite.SetupTest()

		bondedRate := randRate(sdk.ZeroDec(), sdk.OneDec())
		suite.mintAndBondWithRate(types.DefaultInflationDenom, initialSupply, bondedRate)

		ctx := sdk.WrapSDKContext(suite.ctx)
		res, err := suite.queryClient.GetBondedRatio(ctx, &types.QueryBondedRatioRequest{})
		suite.Require().NoError(err)
		suite.Require().True(res.BondedRatio.Sub(bondedRate).Abs().LTE(sdk.NewDecWithPrec(1, 10)))
	}
}

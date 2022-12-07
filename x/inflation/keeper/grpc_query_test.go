package keeper_test

import (
	"fmt"
	epochstypes "github.com/evmos/evmos/v6/x/epochs/types"
	"github.com/tendermint/tendermint/libs/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/AstraProtocol/astra/v2/x/inflation/types"
)

func (suite *KeeperTestSuite) TestPeriod() {
	var (
		req    *types.QueryInflationPeriodRequest
		expRes *types.QueryInflationPeriodResponse
	)

	defaultPeriodInfo := &types.QueryInflationPeriodResponse{
		Period:          0,
		EpochsPerPeriod: 365,
		EpochIdentifier: epochstypes.DayEpochID,
	}

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default period information",
			func() {
				req = &types.QueryInflationPeriodRequest{}
				expRes = defaultPeriodInfo
			},
			true,
		},
		{
			"newly set period",
			func() {
				period := uint64(9)
				suite.app.InflationKeeper.SetPeriod(suite.ctx, period)
				suite.Commit()

				req = &types.QueryInflationPeriodRequest{}
				expRes = &types.QueryInflationPeriodResponse{
					Period:          period,
					EpochsPerPeriod: defaultPeriodInfo.EpochsPerPeriod,
					EpochIdentifier: defaultPeriodInfo.EpochIdentifier,
				}
			},
			true,
		},
		{
			"newly set epochIdentifier",
			func() {
				epochIdentifier := epochstypes.HourEpochID
				suite.app.InflationKeeper.SetEpochIdentifier(suite.ctx, epochIdentifier)
				suite.Commit()

				req = &types.QueryInflationPeriodRequest{}
				expRes = &types.QueryInflationPeriodResponse{
					Period:          defaultPeriodInfo.Period,
					EpochsPerPeriod: defaultPeriodInfo.EpochsPerPeriod,
					EpochIdentifier: epochIdentifier,
				}
			},
			true,
		},
		{
			"newly set epochsPerPeriod",
			func() {
				epochsPerPeriod := int64(101)
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, epochsPerPeriod)
				suite.Commit()

				req = &types.QueryInflationPeriodRequest{}
				expRes = &types.QueryInflationPeriodResponse{
					Period:          defaultPeriodInfo.Period,
					EpochsPerPeriod: uint64(epochsPerPeriod),
					EpochIdentifier: defaultPeriodInfo.EpochIdentifier,
				}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.InflationPeriod(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes.Period, res.Period)
				suite.Require().Equal(expRes.EpochIdentifier, res.EpochIdentifier)
				suite.Require().Equal(expRes.EpochsPerPeriod, res.EpochsPerPeriod)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestEpochMintProvision() {
	var (
		req    *types.QueryEpochMintProvisionRequest
		expRes *types.QueryEpochMintProvisionResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default epochMintProvision",
			func() {
				params := types.DefaultParams()
				defaultEpochMintProvision := types.CalculateEpochMintProvision(
					params,
					uint64(0),
					365,
				)
				req = &types.QueryEpochMintProvisionRequest{}
				expRes = &types.QueryEpochMintProvisionResponse{
					EpochMintProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, defaultEpochMintProvision),
				}
			},
			true,
		},
		{
			"set epochMintProvision",
			func() {
				epochMintProvision := sdk.NewDec(1_000_000)
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, epochMintProvision)
				suite.Commit()

				req = &types.QueryEpochMintProvisionRequest{}
				expRes = &types.QueryEpochMintProvisionResponse{EpochMintProvision: sdk.NewDecCoinFromDec(types.DefaultInflationDenom, epochMintProvision)}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.EpochMintProvision(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSkippedEpochs() {
	var (
		req    *types.QuerySkippedEpochsRequest
		expRes *types.QuerySkippedEpochsResponse
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"default skipped epochs",
			func() {
				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{}
			},
			true,
		},
		{
			"set skipped epochs",
			func() {
				skippedEpochs := uint64(9)
				suite.app.InflationKeeper.SetSkippedEpochs(suite.ctx, skippedEpochs)
				suite.Commit()

				req = &types.QuerySkippedEpochsRequest{}
				expRes = &types.QuerySkippedEpochsResponse{SkippedEpochs: skippedEpochs}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			ctx := sdk.WrapSDKContext(suite.ctx)
			tc.malleate()

			res, err := suite.queryClient.SkippedEpochs(ctx, req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCirculatingSupply() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	// Mint coins to increase supply
	mintDenom := suite.app.InflationKeeper.GetParams(suite.ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(400_000_000), sdk.DefaultPowerReduction))
	err := suite.app.InflationKeeper.MintCoins(suite.ctx, mintCoin)
	suite.Require().NoError(err)

	expCirculatingSupply := sdk.NewDecCoin(mintDenom, mintCoin.Amount)

	res, err := suite.queryClient.CirculatingSupply(ctx, &types.QueryCirculatingSupplyRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expCirculatingSupply, res.CirculatingSupply)

	// mint a number of new coins and check circulating supply
	numCoins := rand.Int() % 100
	for i := 0; i < numCoins; i++ {
		// create a new coin
		newCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(1+rand.Uint64()%400_000_000), sdk.DefaultPowerReduction))
		expCirculatingSupply = expCirculatingSupply.Add(sdk.NewDecCoin(mintDenom, newCoin.Amount))

		// mint this new coin
		err := suite.app.InflationKeeper.MintCoins(suite.ctx, newCoin)
		suite.Require().NoError(err)

		res, err := suite.queryClient.CirculatingSupply(ctx, &types.QueryCirculatingSupplyRequest{})
		suite.Require().NoError(err)
		suite.Require().Equal(expCirculatingSupply, res.CirculatingSupply)
	}

}

func (suite *KeeperTestSuite) TestQueryInflationRate() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	// Mint coins to increase supply
	mintDenom := suite.app.InflationKeeper.GetParams(suite.ctx).MintDenom
	mintCoin := sdk.NewCoin(mintDenom, sdk.TokensFromConsensusPower(int64(1_200_000_000), sdk.DefaultPowerReduction))
	err := suite.app.InflationKeeper.MintCoins(suite.ctx, mintCoin)
	suite.Require().NoError(err)

	expInflationRate := sdk.MustNewDecFromStr("17.333333333333333300")
	res, err := suite.queryClient.InflationRate(ctx, &types.QueryInflationRateRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expInflationRate, res.InflationRate)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

func (suite *KeeperTestSuite) TestGetSetTotalMintedProvision() {
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
				suite.app.InflationKeeper.SetTotalMintProvision(suite.ctx, newTotalMintedProvision)
				suite.Commit()

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

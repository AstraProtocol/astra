package keeper_test

import (
	"github.com/AstraProtocol/astra/v3/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/rand"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.MintKeeper.GetParams(suite.ctx)
	expParams := types.DefaultParams()

	suite.Require().Equal(expParams, params)

	for i := 0; i < numTests; i++ {
		params = randomizedValidParams()
		suite.app.MintKeeper.SetParams(suite.ctx, params)
		newParams := suite.app.MintKeeper.GetParams(suite.ctx)
		suite.Require().Equal(newParams, params)
	}
}

func randomizedValidParams() types.Params {
	return types.Params{
		MintDenom:             randomValidDenom(),
		InflationParameters:   randomValidInflationParameters(),
		InflationDistribution: randomValidInflationDistribution(),
		FoundationAddress:     randomAstraAddress(),
	}
}

func randomValidDenom() string {
	a := []string{"astra", "aastra"}
	return a[rand.Intn(2)]
}

func randomValidInflationDistribution() types.InflationDistribution {
	total := int64(100)
	staking := rand.Int63n(total + 1)
	total -= staking
	foundation := rand.Int63n(total + 1)
	total -= foundation

	return types.InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(staking, 2),
		Foundation:     sdk.NewDecWithPrec(foundation, 2),
		CommunityPool:  sdk.NewDecWithPrec(total, 2),
	}
}

func randomValidInflationParameters() types.InflationParameters {
	for {
		inflationMin := rand.Int63n(101)
		inflationMax := inflationMin + rand.Int63n(101-inflationMin)
		if inflationMax > 100 {
			continue
		}
		goalBonded := 1 + rand.Int63n(100)      // make sure 1% <= goalBonded <= 100%
		inflationRateChange := rand.Int63n(101) // 0% <= inflationRateChange <= 100%

		return types.InflationParameters{
			InflationRateChange: sdk.NewDecWithPrec(inflationRateChange, 2),
			InflationMin:        sdk.NewDecWithPrec(inflationMin, 2),
			InflationMax:        sdk.NewDecWithPrec(inflationMax, 2),
			GoalBonded:          sdk.NewDecWithPrec(goalBonded, 2),
			BlocksPerYear:       uint64(1 + rand.Int63n(60*60*8766)),
		}
	}
}

func randomAstraAddress() string {
	address := sdk.AccAddress(rand.Bytes(32))

	return address.String()
}

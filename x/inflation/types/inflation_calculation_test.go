package types

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	ethermint "github.com/evmos/ethermint/types"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InflationTestSuite struct {
	suite.Suite
}

func TestInflationSuite(t *testing.T) {
	suite.Run(t, new(InflationTestSuite))
}

func (suite *InflationTestSuite) TestCalculateEpochMintProvision() {
	epochsPerPeriod := int64(365)
	defaultParams := DefaultParams()
	defaultParams.InflationParameters.R = sdk.NewDecWithPrec(26, 2)
	baseParams := defaultParams
	baseParams.MintDenom = config.DisplayDenom
	baseParams.InflationParameters.MaxStakingRewards = baseParams.InflationParameters.MaxStakingRewards.Quo(ethermint.PowerReduction.ToDec())

	testCases := []struct {
		name              string
		params            Params
		period            uint64
		expEpochProvision sdk.Dec
		expPass           bool
	}{
		{
			"pass - default param - initial period",
			defaultParams,
			uint64(0),
			sdk.MustNewDecFromStr("569863013698630136986301.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 1",
			defaultParams,
			uint64(1),
			sdk.MustNewDecFromStr("421698630136986301369863.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 2",
			defaultParams,
			uint64(2),
			sdk.MustNewDecFromStr("312056986301369863013698.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 3",
			defaultParams,
			uint64(3),
			sdk.MustNewDecFromStr("230922169863013698630136.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 4",
			defaultParams,
			uint64(4),
			sdk.MustNewDecFromStr("170882405698630136986301.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 5",
			defaultParams,
			uint64(5),
			sdk.MustNewDecFromStr("126452980216986301369863.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 6",
			defaultParams,
			uint64(6),
			sdk.MustNewDecFromStr("93575205360569863013698.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 20",
			defaultParams,
			uint64(20),
			sdk.MustNewDecFromStr("1381671709073041469584.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 21",
			defaultParams,
			uint64(21),
			sdk.MustNewDecFromStr("1022437064714050687492.000000000000000000"),
			true,
		},
		{
			"pass - `astra` denom - period 0",
			baseParams,
			uint64(0),
			sdk.MustNewDecFromStr("569863013698630136986301.000000000000000000"),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			epochMintProvisions := CalculateEpochMintProvision(
				tc.params,
				tc.period,
				epochsPerPeriod,
			)

			suite.Require().Equal(tc.expEpochProvision, epochMintProvisions)
		})
	}
}

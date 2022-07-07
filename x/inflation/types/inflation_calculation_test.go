package types

import (
	fmt "fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
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
	baseParams := defaultParams
	baseParams.MintDenom = config.BaseDenom

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
			sdk.MustNewDecFromStr("608821917808219178082192.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 1",
			defaultParams,
			uint64(1),
			sdk.MustNewDecFromStr("547939726027397260273973.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 2",
			defaultParams,
			uint64(2),
			sdk.MustNewDecFromStr("493145753424657534246575.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 3",
			defaultParams,
			uint64(3),
			sdk.MustNewDecFromStr("443831178082191780821918.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 20",
			defaultParams,
			uint64(20),
			sdk.MustNewDecFromStr("74018532008537829106301.000000000000000000"),
			true,
		},
		{
			"pass - default param - period 21",
			defaultParams,
			uint64(21),
			sdk.MustNewDecFromStr("66616678807684045586849.000000000000000000"),
			true,
		},
		{
			"pass - `aastra` denom - period 0",
			baseParams,
			uint64(0),
			sdk.MustNewDecFromStr("608821.000000000000000000"),
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

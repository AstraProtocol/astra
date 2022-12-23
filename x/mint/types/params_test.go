package types

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/rand"
	"testing"
)

const (
	numRandomTests = 1000
)

type TypesTestSuite struct {
	suite.Suite
}

func TestTypesTestSuite(t *testing.T) {
	config.SetBech32Prefixes(sdk.GetConfig())
	suite.Run(t, new(TypesTestSuite))
}

func (suite *TypesTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

type validateParamTestCase struct {
	name          string
	params        Params
	errorExpected bool
}

func (suite *TypesTestSuite) TestValidateParams() {
	defaultParams := DefaultParams()
	tests := []validateParamTestCase{
		{"default params", defaultParams, false},
		{
			"invalid - denom with backslash",
			NewParams(
				"/aastra",
				defaultParams.InflationParameters,
				defaultParams.InflationDistribution,
				defaultParams.FoundationAddress,
			),
			true,
		},
		{
			"invalid - empty denom",
			Params{
				MintDenom:             "",
				InflationParameters:   defaultParams.InflationParameters,
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{
			"invalid - not allowed denom",
			Params{
				MintDenom:             "aaastra",
				InflationParameters:   defaultParams.InflationParameters,
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - negative InflationRateChange",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: sdk.NewDecWithPrec(-10, 2),
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - InflationRateChange > 1",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: sdk.NewDecWithPrec(101, 2),
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - negative InflationMax",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        sdk.NewDecWithPrec(-10, 2),
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - InflationMax > 1",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        sdk.NewDecWithPrec(101, 2),
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - negative InflationMin",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        sdk.NewDecWithPrec(-10, 2),
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - InflationMin > 1",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        sdk.NewDecWithPrec(101, 2),
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - InflationMin > InflationMax",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        sdk.NewDecWithPrec(8, 2),
					InflationMin:        sdk.NewDecWithPrec(10, 2),
					GoalBonded:          defaultParams.InflationParameters.GoalBonded,
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - 0% GoalBonded",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          sdk.ZeroDec(),
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - negative GoalBonded",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          sdk.NewDecWithPrec(-10, 2),
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - GoalBonded > 1",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          sdk.NewDecWithPrec(101, 2),
					BlocksPerYear:       defaultParams.InflationParameters.BlocksPerYear,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation parameters - zero BlocksPerYears",
			Params{
				MintDenom: defaultParams.MintDenom,
				InflationParameters: InflationParameters{
					InflationRateChange: defaultParams.InflationParameters.InflationRateChange,
					InflationMax:        defaultParams.InflationParameters.InflationMax,
					InflationMin:        defaultParams.InflationParameters.InflationMin,
					GoalBonded:          sdk.NewDecWithPrec(101, 2),
					BlocksPerYear:       0,
				},
				InflationDistribution: defaultParams.InflationDistribution,
				FoundationAddress:     defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - staking rewards < 0",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(-10, 2),
					Foundation:     sdk.NewDecWithPrec(0, 2),
					CommunityPool:  sdk.NewDecWithPrec(0, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - staking rewards > 1",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(101, 2),
					Foundation:     sdk.NewDecWithPrec(0, 2),
					CommunityPool:  sdk.NewDecWithPrec(0, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - foundation < 0",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(0, 2),
					Foundation:     sdk.NewDecWithPrec(-10, 2),
					CommunityPool:  sdk.NewDecWithPrec(0, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - foundation > 1",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(0, 2),
					Foundation:     sdk.NewDecWithPrec(101, 2),
					CommunityPool:  sdk.NewDecWithPrec(0, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - CommunityPool < 0",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(0, 2),
					Foundation:     sdk.NewDecWithPrec(0, 2),
					CommunityPool:  sdk.NewDecWithPrec(-10, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - CommunityPool > 1",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(0, 2),
					Foundation:     sdk.NewDecWithPrec(0, 2),
					CommunityPool:  sdk.NewDecWithPrec(101, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"invalid inflation distribution - total != 1",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(80, 2),
					Foundation:     sdk.NewDecWithPrec(10, 2),
					CommunityPool:  sdk.NewDecWithPrec(20, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			true,
		},
		{"valid inflation distribution",
			Params{
				MintDenom:           defaultParams.MintDenom,
				InflationParameters: defaultParams.InflationParameters,
				InflationDistribution: InflationDistribution{
					StakingRewards: sdk.NewDecWithPrec(20, 2),
					Foundation:     sdk.NewDecWithPrec(20, 2),
					CommunityPool:  sdk.NewDecWithPrec(60, 2),
				},
				FoundationAddress: defaultParams.FoundationAddress,
			},
			false,
		},
	}

	for i := 0; i < numRandomTests; i++ {
		tests = append(tests, validateParamTestCase{
			name:          fmt.Sprintf("RandomValidTestCase-%v", rand.Int63()),
			params:        randomizedValidParamsTestCase(),
			errorExpected: false,
		})
	}

	for _, tc := range tests {
		err := tc.params.Validate()
		if tc.errorExpected {
			suite.Require().Error(err, fmt.Sprintf("Test %v: expected error, got nil", tc.name))
		} else {
			suite.Require().NoError(err, fmt.Sprintf("Test %v: expected no error, got %v", tc.name, err))
		}
	}
}

func randomizedValidParamsTestCase() Params {
	return Params{
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

func randomValidInflationDistribution() InflationDistribution {
	total := int64(100)
	staking := rand.Int63n(total + 1)
	total -= staking
	foundation := rand.Int63n(total + 1)
	total -= foundation

	return InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(staking, 2),
		Foundation:     sdk.NewDecWithPrec(foundation, 2),
		CommunityPool:  sdk.NewDecWithPrec(total, 2),
	}
}

func randomValidInflationParameters() InflationParameters {
	for {
		inflationMin := rand.Int63n(101)
		inflationMax := inflationMin + rand.Int63n(101-inflationMin)
		if inflationMax > 100 {
			continue
		}
		goalBonded := 1 + rand.Int63n(100)
		inflationRateChange := rand.Int63n(101)

		return InflationParameters{
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

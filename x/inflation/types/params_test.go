package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	validInflationParameters := InflationParameters{
		MaxStakingRewards: sdk.NewDec(2222200000),
		R:                 sdk.NewDecWithPrec(5, 1),
	}

	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{
			"default",
			DefaultParams(),
			false,
		},
		{
			"valid",
			NewParams(
				"aastra",
				validInflationParameters,
				true,
			),
			false,
		},
		{
			"valid param literal",
			Params{
				MintDenom:           "aastra",
				InflationParameters: validInflationParameters,
				EnableInflation:     true,
			},
			false,
		},
		{
			"invalid - denom with backslash",
			NewParams(
				"/aastra",
				validInflationParameters,
				true,
			),
			true,
		},
		{
			"invalid - empty denom",
			Params{
				MintDenom:           "",
				InflationParameters: validInflationParameters,
				EnableInflation:     true,
			},
			true,
		},
		{
			"invalid - not allowed denom",
			Params{
				MintDenom:           "aaastra",
				InflationParameters: validInflationParameters,
				EnableInflation:     true,
			},
			true,
		},
		{
			"invalid - inflation parameters - R greater than 1",
			Params{
				MintDenom: "aastra",
				InflationParameters: InflationParameters{
					MaxStakingRewards: validInflationParameters.MaxStakingRewards,
					R:                 sdk.NewDecWithPrec(5, 0),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation parameters - negative R",
			Params{
				MintDenom: "aastra",
				InflationParameters: InflationParameters{
					MaxStakingRewards: validInflationParameters.MaxStakingRewards,
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation parameters - negative maximum staking rewards < -1",
			Params{
				MintDenom: "aastra",
				InflationParameters: InflationParameters{
					MaxStakingRewards: sdk.NewDec(-100101010101),
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				EnableInflation: true,
			},
			true,
		},
		{
			"invalid - inflation parameters - maximum staking rewards == 0",
			Params{
				MintDenom: "aastra",
				InflationParameters: InflationParameters{
					MaxStakingRewards: sdk.NewDec(0),
					R:                 sdk.NewDecWithPrec(-5, 1),
				},
				EnableInflation: true,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *TypesTestSuite) TestNextInflation() {
	minter := DefaultInitialMinter()
	params := DefaultParams()
	blocksPerYr := sdk.NewDec(int64(params.InflationParameters.BlocksPerYear))

	// Governing Mechanism:
	//    inflationRateChangePerYear = (1 - BondedRatio/ GoalBonded) * MaxInflationRateChange

	tests := []struct {
		name                                 string
		bondedRatio, setInflation, expChange sdk.Dec
	}{
		// with 0% bonded astra supply the inflation should increase by InflationRateChange
		{
			"zero bonded - expect increased", sdk.ZeroDec(), sdk.NewDecWithPrec(10, 2), params.InflationParameters.InflationRateChange.Quo(blocksPerYr),
		},

		// 100% bonded, starting at 10% inflation and being reduced
		// (1 - (1/0.5))*(0.60/10519200)
		{
			"100% bonded - expect decreased",
			sdk.OneDec(), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.OneDec().Quo(params.InflationParameters.GoalBonded)).Mul(params.InflationParameters.InflationRateChange).Quo(blocksPerYr),
		},

		// 40% bonded, starting at 10% inflation and being increased
		{
			"under-bonded - expect increased",
			sdk.NewDecWithPrec(4, 1), sdk.NewDecWithPrec(10, 2),
			sdk.OneDec().Sub(sdk.NewDecWithPrec(4, 1).Quo(params.InflationParameters.GoalBonded)).Mul(params.InflationParameters.InflationRateChange).Quo(blocksPerYr),
		},

		// test 3% minimum stop (testing with 100% bonded)
		{"minimum inflation rate + over-bonded - expect unchanged", sdk.OneDec(), sdk.NewDecWithPrec(3, 2), sdk.ZeroDec()},
		{"near minimum inflation rate + over-bonded - expect reaching minimum", sdk.OneDec(), sdk.NewDecWithPrec(300000001, 10), sdk.NewDecWithPrec(-1, 10)},

		// test 15% maximum stop (testing with 0% bonded)
		{"maximum inflation rate + under-bonded - expect unchanged", sdk.ZeroDec(), sdk.NewDecWithPrec(15, 2), sdk.ZeroDec()},
		{"near maximum inflation rate + under-bonded - expect reaching minimum", sdk.ZeroDec(), sdk.NewDecWithPrec(1499999999, 10), sdk.NewDecWithPrec(1, 10)},

		// perfect balance shouldn't change inflation
		{"perfect ratio - expect unchanged", sdk.NewDecWithPrec(50, 2), sdk.NewDecWithPrec(10, 2), sdk.ZeroDec()},
	}
	for _, tc := range tests {
		minter.Inflation = tc.setInflation

		inflation := minter.NextInflationRate(params, tc.bondedRatio)
		diffInflation := inflation.Sub(tc.setInflation)

		suite.Require().True(diffInflation.Equal(tc.expChange),
			"Test Name: %v\nDiff:  %v\nExpected: %v\n", tc.name, diffInflation, tc.expChange)
	}
}

func (suite *TypesTestSuite) TestBlockProvision() {
	minter := InitialMinter(sdk.NewDecWithPrec(1, 1))
	params := DefaultParams()

	blocksPerYear := int64(60 * 60 * 8766 / 3)

	tests := []struct {
		annualProvisions int64
		expProvisions    int64
	}{
		{blocksPerYear, 1},
		{blocksPerYear + 1, 1},
		{blocksPerYear * 2, 2},
		{blocksPerYear / 2, 0},
	}
	for i, tc := range tests {
		minter.AnnualProvisions = sdk.NewDec(tc.annualProvisions)
		provisions := minter.BlockProvision(params)

		expProvisions := sdk.NewCoin(params.MintDenom,
			sdk.NewInt(tc.expProvisions))

		suite.Require().True(expProvisions.IsEqual(provisions),
			"test: %v\n\tExp: %v\n\tGot: %v\n",
			i, tc.expProvisions, provisions)
	}
}

func (suite *TypesTestSuite) TestValidateMinter() {
	tests := []struct {
		name      string
		inflation sdk.Dec
		expErr    bool
	}{
		{
			"valid minter", sdk.NewDecWithPrec(10, 2), false,
		},
		{
			"invalid minter - negative inflation", sdk.NewDecWithPrec(-10, 2), true,
		},
	}

	for _, tc := range tests {
		minter := DefaultInitialMinter()
		minter.Inflation = tc.inflation
		err := ValidateMinter(minter)
		if tc.expErr {
			suite.Require().Error(err, fmt.Sprintf("Test %v: expected error, got nil", tc.name))
		} else {
			suite.Require().NoError(err, fmt.Sprintf("Test %v: expected no error, got %v", tc.name, err))
		}
	}
}

package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/tendermint/tendermint/libs/rand"
)

type tcMintAndAllocateInflation struct {
	name                string
	mintCoin            sdk.Coin
	expStakingRewardAmt sdk.Coin
	expFoundationAmt    sdk.Coin
	expCommunityAmt     sdk.Coin
}

func newRandomizedMintAndAllocateInflationTestCase(param types.Params) tcMintAndAllocateInflation {
	mintCoinAmount := int64(rand.Int())
	mintCoin := sdk.Coin{Denom: denomMint, Amount: sdk.NewInt(mintCoinAmount)}

	expStakingRewardAmt := sdk.NewCoin(denomMint,
		sdk.NewDecFromInt(sdk.NewInt(mintCoinAmount)).Mul(param.InflationDistribution.StakingRewards).TruncateInt())

	expFoundationAmt := sdk.NewCoin(denomMint,
		sdk.NewDecFromInt(sdk.NewInt(mintCoinAmount)).Mul(param.InflationDistribution.Foundation).TruncateInt())

	return tcMintAndAllocateInflation{
		name:                fmt.Sprintf("randomized-%v", mintCoinAmount),
		mintCoin:            mintCoin,
		expStakingRewardAmt: expStakingRewardAmt,
		expFoundationAmt:    expFoundationAmt,
		expCommunityAmt:     mintCoin.Sub(expFoundationAmt.Add(expStakingRewardAmt)),
	}
}

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []tcMintAndAllocateInflation{
		{
			"pass",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.NewCoin(denomMint, sdk.NewInt(880_000)),
			sdk.NewCoin(denomMint, sdk.NewInt(100_000)),
			sdk.NewCoin(denomMint, sdk.NewInt(20_000)),
		},
		{
			"pass - no coins minted",
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
		},
	}

	for i := 0; i < numRandTests; i++ {
		testCases = append(testCases, newRandomizedMintAndAllocateInflationTestCase(suite.app.InflationKeeper.GetParams(suite.ctx)))
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			param := suite.app.InflationKeeper.GetParams(suite.ctx)
			foundationAddress := sdk.MustAccAddressFromBech32(param.FoundationAddress)
			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			communityPool := suite.app.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)

			oldFoundationBalance := suite.app.BankKeeper.GetBalance(suite.ctx, foundationAddress, denomMint)
			oldFeeCollectorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, denomMint)
			oldCommunityPoolBalance := suite.app.BankKeeper.GetBalance(suite.ctx, communityPool, denomMint)

			err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)
			suite.Require().NoError(err, tc.name)

			// get current module balance
			moduleBalance := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			// get new balances of distribution components
			newFoundationBalance := suite.app.BankKeeper.GetBalance(suite.ctx, foundationAddress, denomMint)
			newFeeCollectorBalance := suite.app.BankKeeper.GetBalance(suite.ctx, feeCollector, denomMint)
			newCommunityPoolBalance := suite.app.BankKeeper.GetBalance(suite.ctx, communityPool, denomMint)

			suite.Require().True(moduleBalance.IsZero())
			suite.Require().Equal(tc.expFoundationAmt, newFoundationBalance.Sub(oldFoundationBalance))
			suite.Require().Equal(tc.expStakingRewardAmt, newFeeCollectorBalance.Sub(oldFeeCollectorBalance))
			suite.Require().Equal(tc.expCommunityAmt, newCommunityPoolBalance.Sub(oldCommunityPoolBalance))
		})
	}
}

func (suite *KeeperTestSuite) TestGetCirculatingSupplyAndInflationRate() {
	testCases := []struct {
		name             string
		bankSupply       int64
		malleate         func()
		expInflationRate sdk.Dec
	}{
		{
			"no mint provision",
			1_200_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, sdk.ZeroDec())
			},
			sdk.ZeroDec(),
		},
		{
			"no epochs per period",
			1_200_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 0)
			},
			sdk.ZeroDec(),
		},
		{
			"usual supply",
			1_200_000_000,
			func() {},
			sdk.MustNewDecFromStr("17.333333333333333300"),
		},
		{
			"high supply",
			1_600_000_000,
			func() {},
			sdk.MustNewDecFromStr("13.000000000000000000"),
		},
		{
			"low supply",
			800_000_000,
			func() {},
			sdk.MustNewDecFromStr("26.000000000000000000"),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			suite.ctx = suite.ctx.WithChainID("astra_11112-1")

			// Mint initial supply
			coin := sdk.NewCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(tc.bankSupply, sdk.DefaultPowerReduction))
			decCoin := sdk.NewDecCoinFromCoin(coin)

			err := suite.app.InflationKeeper.MintCoins(suite.ctx, coin)
			suite.Require().NoError(err)

			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx)
			suite.Require().Equal(decCoin.Amount, circulatingSupply)

			tc.malleate()

			inflationRate := s.app.InflationKeeper.GetInflationRate(suite.ctx)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
		})
	}
}

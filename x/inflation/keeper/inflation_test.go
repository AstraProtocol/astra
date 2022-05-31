package keeper_test

import (
	"fmt"
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (suite *KeeperTestSuite) TestMintAndAllocateInflation() {
	testCases := []struct {
		name                string
		mintCoin            sdk.Coin
		malleate            func()
		expStakingRewardAmt sdk.Coin
		expCommunityPoolAmt sdk.DecCoins
		expPass             bool
	}{
		{
			"pass",
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			func() {},
			sdk.NewCoin(denomMint, sdk.NewInt(1_000_000)),
			sdk.NewDecCoins(sdk.NewDecCoin(denomMint, sdk.NewInt(0))),
			true,
		},
		{
			"pass - no coins minted ",
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			func() {},
			sdk.NewCoin(denomMint, sdk.ZeroInt()),
			sdk.DecCoins(nil),
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			err := suite.app.InflationKeeper.MintAndAllocateInflation(suite.ctx, tc.mintCoin)

			// Get balances
			balanceModule := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				suite.app.AccountKeeper.GetModuleAddress(types.ModuleName),
				denomMint,
			)

			feeCollector := suite.app.AccountKeeper.GetModuleAddress(authtypes.FeeCollectorName)
			balanceStakingRewards := suite.app.BankKeeper.GetBalance(
				suite.ctx,
				feeCollector,
				denomMint,
			)

			if tc.expPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().True(balanceModule.IsZero())
				suite.Require().Equal(tc.expStakingRewardAmt, balanceStakingRewards)
			} else {
				suite.Require().Error(err)
			}
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
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochMintProvision(suite.ctx, sdk.ZeroDec())
			},
			sdk.ZeroDec(),
		},
		{
			"no epochs per period",
			400_000_000,
			func() {
				suite.app.InflationKeeper.SetEpochsPerPeriod(suite.ctx, 0)
			},
			sdk.ZeroDec(),
		},
		{
			"high supply",
			800_000_000,
			func() {},
			sdk.MustNewDecFromStr("51.562500000000000000"),
		},
		{
			"low supply",
			400_000_000,
			func() {},
			sdk.MustNewDecFromStr("154.687500000000000000"),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			// Team allocation is only set on mainnet
			suite.ctx = suite.ctx.WithChainID("astra_11112-1")
			tc.malleate()

			// Mint coins to increase supply
			coin := sdk.NewCoin(types.DefaultInflationDenom, sdk.TokensFromConsensusPower(tc.bankSupply, sdk.DefaultPowerReduction))
			decCoin := sdk.NewDecCoinFromCoin(coin)
			suite.app.InflationKeeper.MintCoins(suite.ctx, coin)

			circulatingSupply := s.app.InflationKeeper.GetCirculatingSupply(suite.ctx)

			suite.Require().Equal(decCoin.Amount, circulatingSupply)

			inflationRate := s.app.InflationKeeper.GetInflationRate(suite.ctx)
			suite.Require().Equal(tc.expInflationRate, inflationRate)
		})
	}
}

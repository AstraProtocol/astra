<!--
order: 3
-->

# Begin-Block

Minting parameters are recalculated and inflation
paid at the beginning of each block.

## NextInflationRate

The target annual inflation rate is recalculated each block.
The inflation is also subject to a rate change (positive or negative)
depending on the distance from the desired ratio (50%). The maximum rate change
possible is defined to be 60% per year, however the annual inflation is capped
as between 3% and 15%.

```go
// NextInflationRate returns the new inflation rate for the next hour.
func (m Minter) NextInflationRate(params Params, bondedRatio sdk.Dec) sdk.Dec {
	// The target annual inflation rate is recalculated for each provision cycle. The
	// inflation is also subject to a rate change (positive or negative) depending on
	// the distance from the desired ratio (50%). The maximum rate change possible is
	// defined to be 60% per year, however the annual inflation is capped as between
	// 3% and 15%.

	// (1 - bondedRatio/GoalBonded) * InflationRateChange
	inflationRateChangePerYear := sdk.OneDec().
		Sub(bondedRatio.Quo(params.InflationParameters.GoalBonded)).
		Mul(params.InflationParameters.InflationRateChange)
	inflationRateChange := inflationRateChangePerYear.Quo(sdk.NewDec(int64(params.InflationParameters.BlocksPerYear)))

	// adjust the new annual inflation for this next cycle
	inflation := m.Inflation.Add(inflationRateChange) // note inflationRateChange may be negative
	if inflation.GT(params.InflationParameters.InflationMax) {
		inflation = params.InflationParameters.InflationMax
	}
	if inflation.LT(params.InflationParameters.InflationMin) {
		inflation = params.InflationParameters.InflationMin
	}

	return inflation
}
```

## NextAnnualProvisions

Calculate the annual provisions based on current total supply and inflation
rate. This parameter is calculated once per block.

```go
// NextAnnualProvisions returns the annual provisions based on current total
// supply and inflation rate.
func (m Minter) NextAnnualProvisions(_ Params, totalSupply sdk.Int) sdk.Dec {
	return m.Inflation.MulInt(totalSupply)
}
```

## BlockProvision

Calculate the provisions generated for each block based on current annual provisions. The provisions are then minted by the `mint` module's `ModuleMinterAccount` and then distributed to the respective modules.

```go
// BlockProvision returns the provisions for a block based on the annual
// provisions rate.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	provisionAmt := m.AnnualProvisions.QuoInt(sdk.NewInt(int64(params.InflationParameters.BlocksPerYear)))
	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}
```

## AllocateInflation
Allocate coins from the inflation to external modules according to allocation proportions:
   - staking rewards -> sdk `auth` module fee collector
   - foundation -> a multi-sig address of the Astra Foundation
   - community pool -> sdk `dist` module community pool

```go
// AllocateInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
// 	 - foundation -> a multi-sig address of the Astra Foundation
//	 - community pool -> sdk `dist` module community pool
func (k Keeper) AllocateInflation(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	distribution := params.InflationDistribution

	// allocate staking rewards to the fee collector account
	err := k.AllocateCollectedFees(ctx, sdk.NewCoins(k.GetProportions(ctx, mintedCoin, distribution.StakingRewards)))
	if err != nil {
		return err
	}

	// allocate a portion of mintedProvision to the Foundation
	err = k.AllocateFoundation(ctx, sdk.NewCoins(k.GetProportions(ctx, mintedCoin, distribution.Foundation)))
	if err != nil {
		return err
	}

	// allocate the remaining balance of this module for the community pool
	err = k.AllocateAllBalanceToCommunity(ctx)
	if err != nil {
		return err
	}

	// update the total minted provision
	totalMintedProvision := k.GetTotalMintProvision(ctx)
	mintedAmt := mintedCoin.Amount.ToDec()
	if mintedCoin.Denom == config.DisplayDenom {
		mintedAmt = mintedAmt.Mul(ethermint.PowerReduction.ToDec())
	}
	newTotalMintedProvision := totalMintedProvision.Add(mintedAmt)
	k.SetTotalMintProvision(ctx, newTotalMintedProvision)

	return nil
}
```
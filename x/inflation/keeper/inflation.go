package keeper

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/ethermint/types"
)

// GetProportions returns the amount of coins given its distribution and the total amount of coins.
func (k Keeper) GetProportions(
	_ sdk.Context,
	coin sdk.Coin,
	distribution sdk.Dec,
) sdk.Coin {
	return sdk.NewCoin(
		coin.Denom,
		sdk.NewDecFromInt(coin.Amount).Mul(distribution).TruncateInt(),
	)
}

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, coin sdk.Coin) error {
	// Mint coins for distribution
	if err := k.MintCoins(ctx, coin); err != nil {
		return err
	}

	// Allocate minted coins according to the staking module
	return k.AllocateInflation(ctx, coin)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoin sdk.Coin) error {
	newCoins := sdk.NewCoins(newCoin)

	// skip as no coins need to be minted
	if newCoins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AllocateInflation allocates coins from the inflation to external
// modules according to allocation proportions:
//   - staking rewards -> sdk `auth` module fee collector
// 	 - foundation -> a multi-sig address of the Astra Foundation
//	 - community pool -> sdk `dist` module community pool
func (k Keeper) AllocateInflation(ctx sdk.Context, mintedCoin sdk.Coin) error {
	params := k.GetParams(ctx)
	distribution := params.InflationDistribution

	// allocate staking rewards to the fee collector account
	stakingRewardsAmt := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, distribution.StakingRewards))
	err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		stakingRewardsAmt,
	)
	if err != nil {
		return err
	}

	// allocate a portion of mintedProvision to the Foundation
	foundationAmt := sdk.NewCoins(k.GetProportions(ctx, mintedCoin, distribution.Foundation))
	foundationAddr, err := sdk.AccAddressFromBech32(params.FoundationAddress)
	if err != nil {
		return fmt.Errorf("invalid foudation address %v: %v", params.FoundationAddress, err)
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		foundationAddr,
		foundationAmt,
	)
	if err != nil {
		return err
	}

	// allocate the remaining balance of this module for the community pool
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	communityAmt := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	err = k.distrKeeper.FundCommunityPool(
		ctx,
		communityAmt,
		moduleAddr,
	)
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

// GetCirculatingSupply returns the bank supply.
func (k Keeper) GetCirculatingSupply(ctx sdk.Context) sdk.Dec {
	circulatingSupply := k.bankKeeper.GetSupply(ctx, config.BaseDenom).Amount.ToDec()

	return circulatingSupply
}

// GetInflationRate returns the inflation rate for the current period.
func (k Keeper) GetInflationRate(ctx sdk.Context) sdk.Dec {
	epochMintProvision, _ := k.GetEpochMintProvision(ctx)
	if epochMintProvision.IsZero() {
		return sdk.ZeroDec()
	}

	epp := k.GetEpochsPerPeriod(ctx)
	if epp == 0 {
		return sdk.ZeroDec()
	}

	epochsPerPeriod := sdk.NewDec(epp)

	circulatingSupply := k.GetCirculatingSupply(ctx)
	if circulatingSupply.IsZero() {
		return sdk.ZeroDec()
	}

	// epochMintProvision * 365 / circulatingSupply * 100
	return epochMintProvision.Mul(epochsPerPeriod).Quo(circulatingSupply).Mul(sdk.NewDec(100))
}

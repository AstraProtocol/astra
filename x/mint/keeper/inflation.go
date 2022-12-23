package keeper

import (
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/mint/types"
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

	if newCoins.Empty() {
		// skip as no coins need to be minted
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

// AllocateCollectedFees implements an alias call to the underlying supply keeper's
// AllocateCollectedFees to be used in BeginBlocker.
func (k Keeper) AllocateCollectedFees(ctx sdk.Context, fees sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

// AllocateFoundation allocates tokens to the Foundation Address.
func (k Keeper) AllocateFoundation(ctx sdk.Context, amount sdk.Coins) error {
	params := k.GetParams(ctx)
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx,
		types.ModuleName,
		sdk.MustAccAddressFromBech32(params.FoundationAddress),
		amount,
	)
}

// AllocateAllBalanceToCommunity allocates all remaining balance for the Community Pool.
func (k Keeper) AllocateAllBalanceToCommunity(ctx sdk.Context) error {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	communityAmt := k.bankKeeper.GetAllBalances(ctx, moduleAddr)
	if !communityAmt.IsZero() {
		return k.distrKeeper.FundCommunityPool(ctx, communityAmt, moduleAddr)
	}

	return nil
}

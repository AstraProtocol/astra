package keeper

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTotalMintProvision returns the total amount of minted provision via block rewards.
// The returned amount is measured in config.BaseDenom (i.e, aastra).
func (k Keeper) GetTotalMintProvision(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixTotalMintedProvision)
	if len(bz) == 0 {
		return sdk.ZeroDec()
	}

	var totalMintedProvision sdk.Dec
	err := totalMintedProvision.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal totalMintedProvision value: %w", err))
	}

	return totalMintedProvision
}

// SetTotalMintProvision sets the current TotalMintedProvision.
// totalMintedProvision must be converted to config.BaseDenom (i.e, aastra).
func (k Keeper) SetTotalMintProvision(ctx sdk.Context, totalMintedProvision sdk.Dec) {
	bz, err := totalMintedProvision.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value: %w", err))
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixTotalMintedProvision, bz)
}

package keeper

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTotalFeeBurn returns the total fee already burn.
// The returned amount is measured in config.BaseDenom (i.e, aastra).
func (k Keeper) GetTotalFeeBurn(ctx sdk.Context) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixTotalFeeBurn)
	if len(bz) == 0 {
		return sdk.ZeroDec()
	}

	var totalFeeBurn sdk.Dec
	err := totalFeeBurn.Unmarshal(bz)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal totalFeeBurn value: %w", err))
	}

	return totalFeeBurn
}

// SetTotalFeeBurn sets the current totalFeeBurn.
// totalFeeBurn must be converted to config.BaseDenom (i.e, aastra).
func (k Keeper) SetTotalFeeBurn(ctx sdk.Context, totalFeeBurn sdk.Dec) {
	bz, err := totalFeeBurn.Marshal()
	if err != nil {
		panic(fmt.Errorf("unable to marshal amount value: %w", err))
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixTotalFeeBurn, bz)
}

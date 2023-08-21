package keeper

import (
	"github.com/AstraProtocol/astra/v3/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSetIfExists(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// EnableFeeBurn returns the EnableFeeBurn param
func (k Keeper) EnableFeeBurn(ctx sdk.Context) (res bool) {
	k.paramstore.Get(ctx, types.KeyEnableFeeBurn, &res)
	return
}

// FeeBurn returns the FeeBurn param
func (k Keeper) FeeBurn(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyFeeBurn, &res)
	return
}

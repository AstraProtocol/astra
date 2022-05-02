package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
)

func RunForkLogic(ctx sdk.Context, mintkeeper mintkeeper.Keeper) {
	params := mintkeeper.GetParams(ctx)
	params.BlocksPerYear = 15778800
	mintkeeper.SetParams(ctx, params)
}

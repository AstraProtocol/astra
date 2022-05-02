package app

import (
	tv1 "github.com/AstraProtocol/astra/v1/app/upgrades/testnet/v1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

// BeginBlockForks executes any necessary fork logic based upon the current block height.
func BeginBlockForks(ctx sdk.Context, app *Astra) {
	switch ctx.BlockHeight() {
	case tv1.UpgradeHeight:
		// NOTE: only run for testnet
		if !strings.HasPrefix(ctx.ChainID(), TestnetChainID) {
			return
		}
		tv1.RunForkLogic(ctx, app.MintKeeper)
	default:
		// do nothing
		return
	}
	return
}

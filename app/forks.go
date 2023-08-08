package app

import (
	v1 "github.com/AstraProtocol/astra/v3/app/upgrades/v1"
	v3 "github.com/AstraProtocol/astra/v3/app/upgrades/v3"
	"github.com/AstraProtocol/astra/v3/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// BeginBlockForks executes any necessary fork logic based upon the current block height.
func BeginBlockForks(ctx sdk.Context, app *Astra) {
	upgradePlan := upgradetypes.Plan{
		Height: ctx.BlockHeight(),
	}

	switch ctx.BlockHeight() {
	case v1.UpgradeHeight:
		// NOTE: only run for testnet
		if types.IsMainnet(ctx.ChainID()) {
			return
		}

		upgradePlan.Name = v1.UpgradeName
		upgradePlan.Info = v1.UpgradeInfo
	case v3.TestnetUpgradeHeight:
		if types.IsMainnet(ctx.ChainID()) {
			return
		}
		upgradePlan.Name = v3.UpgradeName
		upgradePlan.Info = v3.UpgradeInfo
	case v3.MainnetUpgradeHeight:
		if !types.IsMainnet(ctx.ChainID()) {
		}

		upgradePlan.Name = v3.UpgradeName
		upgradePlan.Info = v3.UpgradeInfo
	default:
		// do nothing
		return
	}

	err := app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
	if err != nil {
		panic(err)
	}
}

package v1

import (
	"time"

	"github.com/AstraProtocol/astra/v3/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	gk govkeeper.Keeper,
	pk paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		if types.IsTestnet(ctx.ChainID()) {
			logger.Info("updating gov params...")
			UpdateGovParams(ctx, gk, pk)
		}
		// Leave modules are as-is to avoid running InitGenesis.
		logger.Info("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

func UpdateGovParams(ctx sdk.Context, gk govkeeper.Keeper, pk paramskeeper.Keeper) {
	tmp := time.Hour
	votingParams := gk.GetVotingParams(ctx)
	votingParams.VotingPeriod = &tmp
	gk.SetVotingParams(ctx, votingParams)
}

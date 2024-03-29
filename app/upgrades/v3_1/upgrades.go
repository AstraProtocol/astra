package v31

import (
	"github.com/AstraProtocol/astra/v3/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v8.1
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	stakingKeeper stakingkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// mainnet only
		if types.IsMainnet(ctx.ChainID()) {
			setMinCommissionRate(ctx, stakingKeeper)
		}
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		// Leave modules are as-is to avoid running InitGenesis.
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// setMinCommissionRate sets the minimum commission rate for validators
// to 5%.
func setMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) {
	stakingParams := stakingtypes.Params{
		UnbondingTime:     sk.UnbondingTime(ctx),
		MaxValidators:     sk.MaxValidators(ctx),
		MaxEntries:        sk.MaxEntries(ctx),
		HistoricalEntries: sk.HistoricalEntries(ctx),
		BondDenom:         sk.BondDenom(ctx),
		MinCommissionRate: sdk.NewDecWithPrec(5, 2), // 5%
	}

	sk.SetParams(ctx, stakingParams)
}

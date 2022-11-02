package v1_1

import (
	"github.com/AstraProtocol/astra/v2/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v1_1
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	sk stakingkeeper.Keeper,
	pk paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)
		// Refs:
		// - https://docs.cosmos.network/master/building-modules/upgrade.html#registering-migrations
		// - https://docs.cosmos.network/master/migrations/chain-upgrade-guide-044.html#chain-upgrade

		if types.IsTestnet(ctx.ChainID()) {
			logger.Info("updating Tendermint consensus params...")
			UpdateConsensusParams(ctx, sk, pk)
		}
		// Leave modules are as-is to avoid running InitGenesis.
		logger.Info("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// UpdateConsensusParams updates the Tendermint Consensus Evidence params (MaxAgeDuration and
// MaxAgeNumBlocks) to match the unbonding period and use the expected avg block time based on the
// node configuration.
func UpdateConsensusParams(ctx sdk.Context, sk stakingkeeper.Keeper, pk paramskeeper.Keeper) {
	subspace, found := pk.GetSubspace(baseapp.Paramspace)
	if !found {
		return
	}

	var evidenceParams tmproto.EvidenceParams
	subspace.GetIfExists(ctx, baseapp.ParamStoreKeyEvidenceParams, &evidenceParams)

	// safety check: no-op if the evidence params is empty (shouldn't happen)
	if evidenceParams.Equal(tmproto.EvidenceParams{}) {
		return
	}

	stakingParams := sk.GetParams(ctx)
	evidenceParams.MaxAgeDuration = stakingParams.UnbondingTime

	maxAgeNumBlocks := sdk.NewInt(int64(evidenceParams.MaxAgeDuration)).QuoRaw(int64(AvgBlockTime))
	evidenceParams.MaxAgeNumBlocks = maxAgeNumBlocks.Int64()
	subspace.Set(ctx, baseapp.ParamStoreKeyEvidenceParams, evidenceParams)

	// update maxGas
	var blockParams abci.BlockParams
	subspace.GetIfExists(ctx, baseapp.ParamStoreKeyBlockParams, &blockParams)
	blockParams.MaxGas = NewMaxGas
	subspace.Set(ctx, baseapp.ParamStoreKeyBlockParams, blockParams)
}

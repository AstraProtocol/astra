package inflation

import (
	"github.com/AstraProtocol/astra/v1/x/inflation/keeper"
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	_ types.StakingKeeper,
	data types.GenesisState,
) {
	// Ensure inflation module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the inflation module account has not been set")
	}

	// Set genesis state
	params := data.Params
	k.SetParams(ctx, params)

	period := data.Period
	k.SetPeriod(ctx, period)

	epochIdentifier := data.EpochIdentifier
	k.SetEpochIdentifier(ctx, epochIdentifier)

	epochsPerPeriod := data.EpochsPerPeriod
	k.SetEpochsPerPeriod(ctx, epochsPerPeriod)

	skippedEpochs := data.SkippedEpochs
	k.SetSkippedEpochs(ctx, skippedEpochs)

	// Calculate epoch mint provision
	epochMintProvision := types.CalculateEpochMintProvision(
		params,
		period,
		epochsPerPeriod,
	)
	k.SetEpochMintProvision(ctx, epochMintProvision)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params:          k.GetParams(ctx),
		Period:          k.GetPeriod(ctx),
		EpochIdentifier: k.GetEpochIdentifier(ctx),
		EpochsPerPeriod: k.GetEpochsPerPeriod(ctx),
		SkippedEpochs:   k.GetSkippedEpochs(ctx),
	}
}

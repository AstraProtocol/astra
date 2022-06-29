package v3

import (
	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	feemarketv010types "github.com/evmos/ethermint/x/feemarket/migrations/v010/types"
	feemarketv011 "github.com/evmos/ethermint/x/feemarket/migrations/v011"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// MigrateGenesis migrates exported state from v4 to v5 genesis state.
// It performs a no-op if the migration errors.
func MigrateGenesis(appState genutiltypes.AppMap, clientCtx client.Context) genutiltypes.AppMap {
	// Migrate x/feemarket.
	if appState[feemarkettypes.ModuleName] == nil {
		return appState
	}

	// unmarshal relative source genesis application state
	var oldFeeMarketState feemarketv010types.GenesisState
	if err := clientCtx.Codec.UnmarshalJSON(appState[feemarkettypes.ModuleName], &oldFeeMarketState); err != nil {
		return appState
	}

	// delete deprecated x/feemarket genesis state
	delete(appState, feemarkettypes.ModuleName)

	// Migrate relative source genesis application state and marshal it into
	// the respective key.
	newFeeMarketState := feemarketv011.MigrateJSON(oldFeeMarketState)

	feeMarketBz, err := clientCtx.Codec.MarshalJSON(&newFeeMarketState)
	if err != nil {
		return appState
	}

	appState[feemarkettypes.ModuleName] = feeMarketBz

	return appState
}

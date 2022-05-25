package inflation_test

import (
	"testing"

	keepertest "github.com/AstraProtocol/astra/v1/testutil/keeper"
	"github.com/AstraProtocol/astra/v1/testutil/nullify"
	"github.com/AstraProtocol/astra/v1/x/inflation"
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.InflationKeeper(t)
	inflation.InitGenesis(ctx, *k, genesisState)
	got := inflation.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}

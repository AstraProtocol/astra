package keeper_test

import (
	"testing"

	testkeeper "github.com/AstraProtocol/astra/v1/testutil/keeper"
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.InflationKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}

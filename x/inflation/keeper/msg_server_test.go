package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/AstraProtocol/astra/v1/testutil/keeper"
	"github.com/AstraProtocol/astra/v1/x/inflation/keeper"
	"github.com/AstraProtocol/astra/v1/x/inflation/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.InflationKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

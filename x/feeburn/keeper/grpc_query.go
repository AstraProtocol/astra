package keeper

import (
	"context"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) TotalFeeBurn(c context.Context, request *types.QueryTotalFeeBurnRequest) (*types.QueryTotalFeeBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	totalFeeBurn := k.GetTotalFeeBurn(ctx)
	return &types.QueryTotalFeeBurnResponse{TotalFeeBurn: sdk.NewDecCoinFromDec(config.BaseDenom, totalFeeBurn)}, nil
}

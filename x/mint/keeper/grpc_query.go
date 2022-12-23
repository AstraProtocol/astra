package keeper

import (
	"context"
	"github.com/AstraProtocol/astra/v2/cmd/config"

	"github.com/AstraProtocol/astra/v2/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (k Keeper) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (k Keeper) AnnualProvisions(c context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}

// TotalMintedProvision returns the total amount of provisions minted via block rewards.
func (k Keeper) TotalMintedProvision(
	c context.Context,
	_ *types.QueryTotalMintedProvisionRequest,
) (*types.QueryTotalMintedProvisionResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	totalMintedProvision := k.GetTotalMintProvision(ctx)
	return &types.QueryTotalMintedProvisionResponse{TotalMintedProvision: sdk.NewDecCoinFromDec(config.BaseDenom, totalMintedProvision)}, nil
}

// BlockProvision returns current block provisions.
func (k Keeper) BlockProvision(
	c context.Context,
	_ *types.QueryBlockProvisionRequest,
) (*types.QueryBlockProvisionResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	blockProvision := minter.BlockProvision(params)

	return &types.QueryBlockProvisionResponse{Provision: blockProvision}, nil
}

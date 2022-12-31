package keeper

import (
	feeburntype "github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// BurnFee takes the total fees and burn 50% (or param set)
func (k Keeper) BurnFee(ctx sdk.Context, bankKeeper feeburntype.BankKeeper,
	totalFees sdk.Coins,
	params feeburntype.Params) error {
	if !params.EnableFeeBurn {
		return nil
	}
	if totalFees.IsZero() {
		return nil
	}

	var feeBurn sdk.Coins
	totalBurnAmount := sdk.NewDec(0)
	for _, c := range totalFees {
		burnAmount := params.FeeBurn.MulInt(c.Amount).RoundInt()
		if !burnAmount.IsZero() {
			feeBurn = feeBurn.Add(sdk.NewCoin(c.Denom, burnAmount))
			totalBurnAmount = totalBurnAmount.Add(params.FeeBurn.MulInt(c.Amount))
		}
	}

	err := bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, feeburntype.ModuleName, feeBurn)
	if err != nil {
		return sdkerrors.Wrapf(err, feeburntype.ErrFeeBurnSend.Error())
	}
	err = bankKeeper.BurnCoins(ctx, feeburntype.ModuleName, feeBurn)
	if err != nil {
		return sdkerrors.Wrapf(err, feeburntype.ErrFeeBurn.Error())
	}
	totalFeeBurn := k.GetTotalFeeBurn(ctx)
	totalFeeBurn = totalFeeBurn.Add(totalBurnAmount)
	k.SetTotalFeeBurn(ctx, totalFeeBurn)
	return nil
}

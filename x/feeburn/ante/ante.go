package ante

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FeeBurnPayoutDecorator Run his after we already deduct the fee from the account with
// the ante.NewDeductFeeDecorator() decorator. We pull funds from the FeeCollector ModuleAccount
type FeeBurnPayoutDecorator struct {
	bankKeeper    BankKeeper
	feeBurnKeeper FeeBurnKeeper
}

func NewFeeBurnPayoutDecorator(bk BankKeeper, fb FeeBurnKeeper) FeeBurnPayoutDecorator {
	return FeeBurnPayoutDecorator{
		bankKeeper:    bk,
		feeBurnKeeper: fb,
	}
}

func (fsd FeeBurnPayoutDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}
	err = FeeBurnPayout(ctx, fsd.bankKeeper, feeTx.GetFee(), fsd.feeBurnKeeper)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return next(ctx, tx, simulate)
}

// FeeBurnPayout takes the total fees and burn 50% (or param set)
func FeeBurnPayout(ctx sdk.Context, bankKeeper BankKeeper, totalFees sdk.Coins, burnKeeper FeeBurnKeeper) error {
	params := burnKeeper.GetParams(ctx)
	if !params.EnableFeeBurn {
		return nil
	}
	if totalFees.Empty() {
		return nil
	}

	var feeBurn sdk.Coins
	for _, c := range totalFees {
		burnAmount := params.FeeBurn.MulInt(c.Amount).RoundInt()
		if !burnAmount.IsZero() {
			feeBurn = feeBurn.Add(sdk.NewCoin(c.Denom, burnAmount))
		}
	}
	fmt.Println("total fee", totalFees)
	fmt.Println("total fee burn", feeBurn)
	err := bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, types.ModuleName, feeBurn)
	if err != nil {
		fmt.Println("send coin failed", err)
		return err
	}
	err = bankKeeper.BurnCoins(ctx, types.ModuleName, feeBurn)
	if err != nil {
		fmt.Println("burn coin failed", err)
		return err
	}

	return nil
}

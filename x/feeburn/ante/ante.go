package ante

import (
	"fmt"
	feeburnkeeper "github.com/AstraProtocol/astra/v2/x/feeburn/keeper"
	feeburntype "github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FeeBurnPayoutDecorator Run his after we already deduct the fee from the account with
// the ante.NewDeductFeeDecorator() decorator. We pull funds from the FeeCollector ModuleAccount
type FeeBurnPayoutDecorator struct {
	bankKeeper    feeburntype.BankKeeper
	feeBurnKeeper feeburntype.FeeBurnKeeper
}

func NewFeeBurnPayoutDecorator(bk feeburntype.BankKeeper, fb feeburntype.FeeBurnKeeper) FeeBurnPayoutDecorator {
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
	fmt.Println(tx.GetMsgs())
	params := fsd.feeBurnKeeper.GetParams(ctx)
	err = feeburnkeeper.FeeBurnPayout(ctx, fsd.bankKeeper, feeTx.GetFee(), params)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return next(ctx, tx, simulate)
}

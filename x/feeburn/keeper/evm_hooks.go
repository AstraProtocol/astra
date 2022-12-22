package keeper

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/AstraProtocol/astra/v2/x/feeburn/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer (or, if set,
// the withdraw address) receives a share from the transaction fees paid by the
// transaction sender.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	// check if the fees are globally enabled
	params := k.GetParams(ctx)
	if !params.EnableFeeBurn {
		return nil
	}

	totalFeeAmount := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	totalFees := sdk.Coins{{Denom: config.BaseDenom, Amount: totalFeeAmount}}
	fmt.Println("evm total fee", totalFees)
	err := FeeBurnPayout(ctx, k.bankKeeper, totalFees, params)
	if err != nil {
		return err
	}

	return nil
}

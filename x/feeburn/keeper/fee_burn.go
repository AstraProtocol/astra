package keeper

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	feeburntype "github.com/AstraProtocol/astra/v2/x/feeburn/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// FeeBurnPayout takes the total fees and burn 50% (or param set)
func FeeBurnPayout(ctx sdk.Context, bankKeeper feeburntype.BankKeeper,
	totalFees sdk.Coins,
	params feeburntype.Params) error {
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
	totalSupply := bankKeeper.GetSupply(ctx, config.BaseDenom)
	fmt.Println("total supply before", totalSupply)
	fmt.Println("total fee", totalFees)
	fmt.Println("total fee burn", feeBurn)
	err := bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, feeburntype.ModuleName, feeBurn)
	if err != nil {
		fmt.Println("send coin failed", err)
		return err
	}
	err = bankKeeper.BurnCoins(ctx, feeburntype.ModuleName, feeBurn)
	if err != nil {
		fmt.Println("burn coin failed", err)
		return err
	}
	totalSupply = bankKeeper.GetSupply(ctx, config.BaseDenom)
	fmt.Println("total supply after", totalSupply)

	return nil
}

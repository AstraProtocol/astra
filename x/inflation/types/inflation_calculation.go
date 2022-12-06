package types

import (
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/evmos/ethermint/types"
)

// CalculateEpochMintProvision returns mint provision per epoch.
// The returned value's demon is always `aastra`.
func CalculateEpochMintProvision(
	params Params,
	period uint64,
	epochsPerPeriod int64,
) sdk.Dec {
	x := period                       // period
	r := params.InflationParameters.R // reduction factor

	// periodProvision = exponentialDecay * MaxStakingRewards
	// where exponentialDecay := r * (1 - r) ^ x
	//
	// to work-around float-point precision loss, we recursively multiply periodProvision with `decay`
	// instead of calculating the whole exponentialDecay := r * (1 - r) ^ x.
	periodProvision := r.Mul(params.InflationParameters.MaxStakingRewards)
	decay := sdk.OneDec().Sub(r)
	for i := uint64(0); i < x; i++ {
		periodProvision = periodProvision.Mul(decay)
	}

	// epochProvision = periodProvision / epochsPerPeriod
	epochProvision := periodProvision.Quo(sdk.NewDec(epochsPerPeriod))

	// If the `denom` is already in `aastra`, multiply epochMintProvision with power reduction (10^18 for astra)
	// as the issued tokens need to be given in `aastra`.
	if params.MintDenom == config.DisplayDenom {
		epochProvision = epochProvision.Mul(ethermint.PowerReduction.ToDec())
	}

	// The returned value is in `aastra`, therefore, it has to be rounded.
	return epochProvision.TruncateInt().ToDec()
}

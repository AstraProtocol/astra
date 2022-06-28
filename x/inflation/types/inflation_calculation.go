package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/evmos/ethermint/types"
)

// CalculateEpochMintProvision returns mint provision per epoch
func CalculateEpochMintProvision(
	params Params,
	period uint64,
	epochsPerPeriod int64,
) sdk.Dec {
	x := period                       // period
	r := params.InflationParameters.R // reduction factor

	// exponentialDecay := r * (1 - r) ^ x
	decay := sdk.OneDec().Sub(r)
	exponentialDecay := r.Mul(decay.Power(x))

	// periodProvision = exponentialDecay * maxStakingRewards
	periodProvision := exponentialDecay.Mul(params.InflationParameters.MaxStakingRewards)

	// epochProvision = periodProvision / epochsPerPeriod
	epochProvision := periodProvision.Quo(sdk.NewDec(epochsPerPeriod))

	// Multiply epochMintProvision with power reduction (10^18 for astra) as the
	// calculation is based on `astra` and the issued tokens need to be given in
	// `aastra`
	epochProvision = epochProvision.Mul(ethermint.PowerReduction.ToDec())
	return epochProvision
}

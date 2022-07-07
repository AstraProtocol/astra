package types

import (
	"github.com/AstraProtocol/astra/v2/cmd/config"
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

	// If the `denom` is already in `aastra`, multiply epochMintProvision with power reduction (10^18 for astra)
	// as the issued tokens need to be given in `aastra`.
	if params.MintDenom == config.DisplayDenom {
		epochProvision = epochProvision.Mul(ethermint.PowerReduction.ToDec())
	}

	// The returned value is in `aastra`, therefore, it has to be rounded.
	return epochProvision.TruncateInt().ToDec()
}

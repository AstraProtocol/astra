package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	"math/rand"

	"github.com/AstraProtocol/astra/v2/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const (
	Inflation             = "inflation"
	InflationParameters   = "inflation_parameters"
	InflationDistribution = "inflation_distribution"
	FoundationAddress     = "foundation_address"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRateChange randomized InflationRateChange
func GenInflationRateChange(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(_ *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(15, 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(_ *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(3, 2)
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(_ *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(50, 2)
}

// GenInflationDistribution randomizes InflationDistribution.
func GenInflationDistribution(r *rand.Rand) types.InflationDistribution {
	total := 100

	staking := r.Intn(total - 1)
	total -= staking

	foundation := r.Intn(total - 1)
	community := 100 - staking - foundation

	return types.InflationDistribution{
		StakingRewards: sdk.NewDecWithPrec(int64(staking), 2),
		Foundation:     sdk.NewDecWithPrec(int64(foundation), 2),
		CommunityPool:  sdk.NewDecWithPrec(int64(community), 2),
	}
}

// GenFoundationAddress randomized FoundationAddress.
func GenFoundationAddress(_ *rand.Rand) string {
	return "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen"
}

func GenInflationParameters(r *rand.Rand) types.InflationParameters {
	return types.InflationParameters{
		InflationRateChange: GenInflationRateChange(r),
		InflationMax:        GenInflationMax(r),
		InflationMin:        GenInflationMin(r),
		GoalBonded:          GenGoalBonded(r),
		BlocksPerYear:       uint64(60 * 60 * 8766 / 3),
	}
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Inflation, &inflation, simState.Rand,
		func(r *rand.Rand) { inflation = GenInflation(r) },
	)

	// params
	var inflationParameters types.InflationParameters
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationParameters, &inflationParameters, simState.Rand,
		func(r *rand.Rand) { inflationParameters = GenInflationParameters(r) },
	)

	mintDenom := config.BaseDenom

	var inflationDistribution types.InflationDistribution
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationDistribution, &inflationDistribution, simState.Rand,
		func(r *rand.Rand) { inflationDistribution = GenInflationDistribution(r) },
	)

	var foundationAddress string
	simState.AppParams.GetOrGenerate(
		simState.Cdc, FoundationAddress, &foundationAddress, simState.Rand,
		func(r *rand.Rand) { foundationAddress = GenFoundationAddress(r) },
	)

	params := types.NewParams(mintDenom,
		inflationParameters,
		inflationDistribution,
		foundationAddress,
	)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}

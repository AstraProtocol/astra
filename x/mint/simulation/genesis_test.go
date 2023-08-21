package simulation_test

import (
	"cosmossdk.io/math"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AstraProtocol/astra/v3/x/mint/simulation"
	"github.com/AstraProtocol/astra/v3/x/mint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// TestRandomizedGenState tests the normal scenario of applying RandomizedGenState.
// Abonormal scenarios are not tested here.
func TestRandomizedGenState(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 3),
		InitialStake: math.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	simulation.RandomizedGenState(&simState)

	var mintGenesis types.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[types.ModuleName], &mintGenesis)

	dec1, _ := sdk.NewDecFromStr("0.500000000000000000")
	dec2, _ := sdk.NewDecFromStr("0.150000000000000000")
	dec3, _ := sdk.NewDecFromStr("0.030000000000000000")

	require.Equal(t, uint64(10519200), mintGenesis.Params.InflationParameters.BlocksPerYear)
	require.Equal(t, dec1, mintGenesis.Params.InflationParameters.GoalBonded)
	require.Equal(t, dec2, mintGenesis.Params.InflationParameters.InflationMax)
	require.Equal(t, dec3, mintGenesis.Params.InflationParameters.InflationMin)
	require.Equal(t, "aastra", mintGenesis.Params.MintDenom)
	require.Equal(t, "0aastra", mintGenesis.Minter.BlockProvision(mintGenesis.Params).String())
	require.Equal(t, "0.100000000000000000", mintGenesis.Minter.NextAnnualProvisions(mintGenesis.Params, sdk.OneInt()).String())
	require.Equal(t, "0.099999983839075215", mintGenesis.Minter.NextInflationRate(mintGenesis.Params, sdk.OneDec()).String())
	require.Equal(t, "0.100000000000000000", mintGenesis.Minter.Inflation.String())
	require.Equal(t, "0.000000000000000000", mintGenesis.Minter.AnnualProvisions.String())
}

// TestRandomizedGenState tests abnormal scenarios of applying RandomizedGenState.
func TestRandomizedGenState1(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	s := rand.NewSource(1)
	r := rand.New(s)
	// all these tests will panic
	tests := []struct {
		simState module.SimulationState
		panicMsg string
	}{
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{}, "invalid memory address or nil pointer dereference"},
		{ // panic => reason: incomplete initialization of the simState
			module.SimulationState{
				AppParams: make(simtypes.AppParams),
				Cdc:       cdc,
				Rand:      r,
			}, "assignment to entry in nil map"},
	}

	for _, tt := range tests {
		require.Panicsf(t, func() { simulation.RandomizedGenState(&tt.simState) }, tt.panicMsg)
	}
}

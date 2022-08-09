package types

import (
	"fmt"
	"github.com/AstraProtocol/astra/v1/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ethermint "github.com/evmos/ethermint/types"
)

var DefaultInflationDenom = config.BaseDenom

// Parameter store keys
var (
	ParamStoreKeyMintDenom           = []byte("ParamStoreKeyMintDenom")
	ParamStoreKeyInflationParameters = []byte("ParamStoreKeyInflationParameters")
	ParamStoreKeyEnableInflation     = []byte("ParamStoreKeyEnableInflation")
)

// ParamKeyTable ParamTable for inflation module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	inflationParameters InflationParameters,
	enableInflation bool,
) Params {
	return Params{
		MintDenom:           mintDenom,
		InflationParameters: inflationParameters,
		EnableInflation:     enableInflation,
	}
}

// DefaultParams default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom: DefaultInflationDenom,
		InflationParameters: InflationParameters{
			MaxStakingRewards: sdk.NewDec(2222200000).Mul(ethermint.PowerReduction.ToDec()),
			R:                 sdk.NewDecWithPrec(5, 2), // decayFactor = 5%
		},
		EnableInflation: true,
	}
}

// ParamSetPairs Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyInflationParameters, &p.InflationParameters, validateInflationParameters),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableInflation, &p.EnableInflation, validateBool),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v != config.BaseDenom && v != config.DisplayDenom {
		return fmt.Errorf("mint denom must be one of [%v, %v]", config.BaseDenom, config.DisplayDenom)
	}

	return nil
}

func validateInflationParameters(i interface{}) error {
	v, ok := i.(InflationParameters)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.MaxStakingRewards.LTE(sdk.NewDec(0)) {
		return fmt.Errorf("MaxStakingRewards cannot be less than or equal to 0")
	}

	// validate reduction factor
	if v.R.GTE(sdk.NewDec(1)) {
		return fmt.Errorf("DecayFactor cannot be greater than or equal to 1")
	}

	if v.R.LTE(sdk.NewDec(0)) {
		return fmt.Errorf("DecayFactor cannot be negative or equal to 0")
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationParameters(p.InflationParameters); err != nil {
		return err
	}

	return validateBool(p.EnableInflation)
}

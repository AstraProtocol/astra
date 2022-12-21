package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyEnableFeeBurn = []byte("EnableFeeBurn")
	// TODO: Determine the default value
	DefaultEnableFeeBurn bool = true
)

var (
	KeyFeeBurn = []byte("FeeBurn")
	// TODO: Determine the default value
	DefaultFeeBurn = sdk.NewDecWithPrec(50, 2) // 50%
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enableFeeBurn bool,
	feeBurn sdk.Dec,
) Params {
	return Params{
		EnableFeeBurn: enableFeeBurn,
		FeeBurn:       feeBurn,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultEnableFeeBurn,
		DefaultFeeBurn,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEnableFeeBurn, &p.EnableFeeBurn, validateEnableFeeBurn),
		paramtypes.NewParamSetPair(KeyFeeBurn, &p.FeeBurn, validateFeeBurn),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEnableFeeBurn(p.EnableFeeBurn); err != nil {
		return err
	}

	if err := validateFeeBurn(p.FeeBurn); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateEnableFeeBurn validates the EnableFeeBurn param
func validateEnableFeeBurn(v interface{}) error {
	enableFeeBurn, ok := v.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = enableFeeBurn

	return nil
}

// validateFeeBurn validates the FeeBurn param
func validateFeeBurn(v interface{}) error {
	feeBurn, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = feeBurn

	return nil
}

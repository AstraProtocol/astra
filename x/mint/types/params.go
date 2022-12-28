package types

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var DefaultInflationDenom = config.BaseDenom

// Parameter store keys
var (
	KeyMintDenom             = []byte("MintDenom")
	KeyInflationParameters   = []byte("InflationParameters")
	KeyInflationDistribution = []byte("InflationDistribution")
	KeyFoundationAddress     = []byte("FoundationAddress")
)

// ParamKeyTable returns the KeyTable for this module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	inflationParameters InflationParameters,
	inflationDistribution InflationDistribution,
	foundationAddress string,
) Params {
	return Params{
		MintDenom:             mintDenom,
		InflationParameters:   inflationParameters,
		InflationDistribution: inflationDistribution,
		FoundationAddress:     foundationAddress,
	}
}

// DefaultParams returns the default parameters for this module.
func DefaultParams() Params {
	return Params{
		MintDenom: DefaultInflationDenom,
		InflationParameters: InflationParameters{
			InflationRateChange: sdk.NewDecWithPrec(60, 2),
			InflationMax:        sdk.NewDecWithPrec(15, 2),
			InflationMin:        sdk.NewDecWithPrec(3, 2),
			GoalBonded:          sdk.NewDecWithPrec(50, 2),
			BlocksPerYear:       uint64(60 * 60 * 8766 / 3), // assuming 3-second block times
		},
		InflationDistribution: InflationDistribution{
			StakingRewards: sdk.NewDecWithPrec(88, 2), // 88%
			Foundation:     sdk.NewDecWithPrec(10, 2), // 10%
			CommunityPool:  sdk.NewDecWithPrec(2, 2),  // 2%
		},
		FoundationAddress: "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen",
	}
}

// Validate checks the correctness of the current Params.
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationParameters(p.InflationParameters); err != nil {
		return err
	}
	if err := validateInflationDistribution(p.InflationDistribution); err != nil {
		return err
	}
	if err := validateAstraAddress(p.FoundationAddress); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs implements params.ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyInflationParameters, &p.InflationParameters, validateInflationParameters),
		paramtypes.NewParamSetPair(KeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair(KeyFoundationAddress, &p.FoundationAddress, validateAstraAddress),
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

	if err := validateInflationRateChange(v.InflationRateChange); err != nil {
		return err
	}
	if err := validateInflationMax(v.InflationMax); err != nil {
		return err
	}
	if err := validateInflationMin(v.InflationMin); err != nil {
		return err
	}
	if err := validateGoalBonded(v.GoalBonded); err != nil {
		return err
	}
	if err := validateBlocksPerYear(v.BlocksPerYear); err != nil {
		return err
	}
	if v.InflationMax.LT(v.InflationMin) {
		return fmt.Errorf(
			"max inflation (%s) must be greater than or equal to min inflation (%s)",
			v.InflationMax, v.InflationMin,
		)
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("goal bonded must be positive: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateInflationDistribution(i interface{}) error {
	v, ok := i.(InflationDistribution)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.StakingRewards.LT(sdk.ZeroDec()) || v.StakingRewards.GT(sdk.OneDec()) {
		return fmt.Errorf("StakingRewards proportion cannot be less than 0 or greater than 1")
	}

	if v.CommunityPool.LT(sdk.ZeroDec()) || v.CommunityPool.GT(sdk.OneDec()) {
		return fmt.Errorf("CommunityPool proportion cannot be less than 0 or greater than 1")
	}

	if v.Foundation.LT(sdk.ZeroDec()) || v.Foundation.GT(sdk.OneDec()) {
		return fmt.Errorf("AstraFoundation proportion cannot be less than 0 or greater than 1")
	}

	totalProportion := v.Foundation.Add(v.CommunityPool)
	totalProportion = totalProportion.Add(v.StakingRewards)
	if !totalProportion.Equal(sdk.OneDec()) {
		return fmt.Errorf("expected total proportion to be equal to 1, got %v", totalProportion.String())
	}

	return nil
}

func validateAstraAddress(i interface{}) error {
	addrStr, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if len(addrStr) == 0 {
		return fmt.Errorf("empty address")
	}
	_, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return fmt.Errorf("invalid Cosmos address %v: %v", addrStr, err)
	}

	return nil
}

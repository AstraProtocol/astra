package types

import (
	"fmt"
	"github.com/AstraProtocol/astra/v2/cmd/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ethermint "github.com/evmos/ethermint/types"
)

var DefaultInflationDenom = config.BaseDenom

// Parameter store keys
var (
	ParamStoreKeyMintDenom             = []byte("ParamStoreKeyMintDenom")
	ParamStoreKeyInflationParameters   = []byte("ParamStoreKeyInflationParameters")
	ParamStoreKeyInflationDistribution = []byte("ParamStoreKeyInflationDistribution")
	ParamStoreKeyFoundationAddress     = []byte("ParamStoreKeyFoundationAddress")
	ParamStoreKeyEnableInflation       = []byte("ParamStoreKeyEnableInflation")
)

// ParamKeyTable ParamTable for inflation module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	inflationParameters InflationParameters,
	inflationDistribution InflationDistribution,
	foundationAddress string,
	enableInflation bool,
) Params {
	return Params{
		MintDenom:             mintDenom,
		InflationParameters:   inflationParameters,
		InflationDistribution: inflationDistribution,
		FoundationAddress:     foundationAddress,
		EnableInflation:       enableInflation,
	}
}

// DefaultParams default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom: DefaultInflationDenom,
		InflationParameters: InflationParameters{
			MaxStakingRewards: sdk.NewDec(800000000).Mul(ethermint.PowerReduction.ToDec()),
			R:                 sdk.NewDecWithPrec(26, 2), // decayFactor = 26%
		},
		InflationDistribution: InflationDistribution{
			StakingRewards: sdk.NewDecWithPrec(88, 2), // 88%
			Foundation:     sdk.NewDecWithPrec(10, 2), // 10%
			CommunityPool:  sdk.NewDecWithPrec(2, 2),  // 2%
		},
		FoundationAddress: "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen",
		EnableInflation:   true,
	}
}

// ParamSetPairs Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyInflationParameters, &p.InflationParameters, validateInflationParameters),
		paramtypes.NewParamSetPair(ParamStoreKeyInflationDistribution, &p.InflationDistribution, validateInflationDistribution),
		paramtypes.NewParamSetPair(ParamStoreKeyFoundationAddress, &p.FoundationAddress, validateAstraAddress),
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
	_, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return fmt.Errorf("invalid Cosmos address %v: %v", addrStr, err)
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
	if err := validateInflationDistribution(p.InflationDistribution); err != nil {
		return err
	}
	if err := validateAstraAddress(p.FoundationAddress); err != nil {
		return err
	}

	return validateBool(p.EnableInflation)
}

<!--
order: 5
-->

# Parameters

The `x/inflation` module contains the parameters described below. All parameters
can be modified via governance.

| Key                      | Type                    | Default Value                                                                                                      |
|:-------------------------|:------------------------|:-------------------------------------------------------------------------------------------------------------------|
| `MintDenom`              | string                  | `config.BaseDenom` // “aastra”                                                                                     |
| `InflationParameters`    | InflationParameters     | `MaxStakingRewards: sdk.NewDec(2222200000).Mul(ethermint.PowerReduction.ToDec())`                                  |
|                          |                         | `R:                 sdk.NewDecWithPrec(10, 2)` // decayFactor = 10%                                                |
| `EnableInflation`        | bool                    | `true`                                                                                                             |

## Mint Denom

The `MintDenom` parameter sets the denomination in which new coins are minted.

## Inflation Parameters

The `InflationParameters` parameter holds all values required for the
calculation of the `epochMintProvision`. More detail can be found [here](01_concepts.md).

Here is the detail of the inflation parameters:
```go
// InflationParameters defines the distribution along with the parameters in which inflation is
// allocated through minting on each epoch. It excludes the genesis-enabled vesting distribution for team,
// genesis partners or reward providers, as they are only minted once at genesis.
// The rest of the total supply (i.e, 30%) will be gradually allocated for staking rewards through
// epoch provisions (e.g, block rewards).
// The epoch provision on each period is calculated as follows:
// periodProvision  = exponentialDecay       *  totalInflation
// f(x)             = r * (1 - r)^x  *  R
type InflationParameters struct {
	// max_staking_rewards defines the maximum amount of the staking reward inflation
	// distributed through block rewards (i.e, R). The max_staking_rewards accounts for 20% of the total supply.
	MaxStakingRewards github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=max_staking_rewards,json=maxStakingRewards,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"max_staking_rewards"`
	// r holds the value of the decay factor.
	R github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,2,opt,name=r,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"r"`
}
```

## Enable Inflation

The `EnableInflation` parameter enables the daily inflation. If it is disabled,
no tokens are minted and the number of skipped epochs increases for each passed
epoch.

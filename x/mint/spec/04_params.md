<!--
order: 4
-->

# Parameters

The minting module contains the following parameters:

| Key                   | Type                  | Example                                        |
|-----------------------|-----------------------|------------------------------------------------|
| MintDenom             | string                | "aastra"                                       |
| InflationParameters   | InflationParameters   | below                                          |
| InflationDistribution | InflationDistribution | below                                          |
| FoundationAddress     | string                | "astra13wjs7d3z8hra6rp7vjmryuulwxjrd232sceuen" |

## Inflation Parameters

| Key                 | Type            | Example                |
|---------------------|-----------------|------------------------|
| InflationRateChange | string (dec)    | "0.600000000000000000" |
| InflationMax        | string (dec)    | "0.150000000000000000" |
| InflationMin        | string (dec)    | "0.030000000000000000" |
| GoalBonded          | string (dec)    | "0.500000000000000000" |
| BlocksPerYear       | string (uint64) | "10519200"             |

## Inflation Distribution

| Key              | Type            | Example                |
|------------------|-----------------|------------------------|
| StakingRewards   | string (dec)    | "0.880000000000000000" |
| Foundation       | string (dec)    | "0.100000000000000000" |
| CommunityPool    | string (dec)    | "0.020000000000000000" |

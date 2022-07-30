<!--
order: 6
-->

# Clients

A user can query the `x/inflation` module using the CLI, JSON-RPC, gRPC or
REST.

## CLI

Find below a list of `astrad` commands added with the `x/inflation`module. You
can obtain the full list by using the`astrad -h`command.

### Queries

The`query`commands allow users to query`inflation`state.

**`period`**

Allows users to query the current inflation period.

```go
astrad query inflation period [flags]
```

**`epoch-mint-provision`**

Allows users to query the current inflation epoch provisions value.

```go
astrad query inflation epoch-mint-provision [flags]
```

**`skipped-epochs`**

Allows users to query the current number of skipped epochs.

```go
astrad query inflation skipped-epochs [flags]
```

**`circulating-supply`**

Allows users to query the supply of tokens in circulation.

```go
astrad query inflation circulating-supply [flags]
```

**`inflation-rate`**

Allows users to query the inflation rate of the current period.

```go
astrad query inflation inflation-rate [flags]
```

**`params`**

Allows users to query the current inflation parameters.

```go
astrad query inflation params [flags]
```

## gRPC

### Queries

|    Verb | Method                                     | Description                                    |
|--------:|:-------------------------------------------|:-----------------------------------------------|
|  `gRPC` | `astra.inflation.Query/InflationPeriod`    | Gets current inflation period                  |
|  `gRPC` | `astra.inflation.Query/EpochMintProvision` | Gets current inflation epoch provisions value  |
|  `gRPC` | `astra.inflation.Query/Params`             | Gets current inflation parameters              |
|  `gRPC` | `astra.inflation.Query/SkippedEpochs`      | Gets current number of skipped epochs          |
|  `gRPC` | `astra.inflation.Query/CirculatingSupply`  | Gets current total supply                      |
|  `gRPC` | `astra.inflation.Query/InflationRate`      | Gets current inflation rate                    |
|   `GET` | `/astra/inflation/inflation_period`        | Gets current inflation period                  |
|   `GET` | `/astra/inflation/epoch_mint_provision`    | Gets current inflation epoch provisions value  |
|   `GET` | `/astra/inflation/skipped_epochs`          | Gets current number of skipped epochs          |
|   `GET` | `/astra/inflation/circulating_supply`      | Gets current total supply                      |
|   `GET` | `/astra/inflation/inflation_rate`          | Gets current inflation rate                    |
|   `GET` | `/astra/inflation/params`                  | Gets current inflation parameters              |

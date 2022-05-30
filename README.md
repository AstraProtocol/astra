<!--
parent:
  order: false
-->

<div align="center">
  <h1> Astra </h1>
</div>

<!-- TODO: add banner -->
<!-- ![banner](docs/ethermint.jpg) -->

Astra is a scalable, high-throughput Proof-of-Stake blockchain that is fully compatible and
interoperable with Ethereum. It's built using the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/) which runs on top of [Tendermint Core](https://github.com/tendermint/tendermint) consensus engine.

**Note**: Requires [Go 1.17.5+](https://golang.org/dl/)

## Installation

For prerequisites and detailed build instructions please read the Installation instructions. Once the dependencies are installed, run:

```bash
make install
```

## Integration test
### Run test
    cd integration_tests
    pytest -s -vv

### Run single test
    pytest -k test_gov


Or check out the latest [release](https://github.com/AstraProtocol/astra/releases).

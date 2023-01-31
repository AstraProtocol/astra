<a name="v2.2.0"></a>
## [v2.2.0](https://github.com/AstraProtocol/astra/compare/v2.1.3...v2.2.0) (2023-01-31)

### Chore

* update init.sh
* upadte init.sh
* update script gen
* update swagger
* remove code unused
* remove proto unused

### Docs

* add module feeburn

### Feat

* add get total_fee_burn in module fee
* add feeburn rest api
* add module feeburn
* add NewMinGasPriceDecorator

### Fix

* integer overflow ([#161](https://github.com/AstraProtocol/astra/issues/161))
* totalBurnAmount incorrect
* miss feeburn in ante handler

### Refactor
* remove default MinGasPrice
* remove upgrade handler
* add default DefaultMinGasPrice, DefaultMinGasMultiplier
* add handler errror ([#160](https://github.com/AstraProtocol/astra/issues/160))
* update proto
* reorder module
* update astra app

### Test

* update unittest for mint module
* update integration tests
* update test_fee_payer ([#159](https://github.com/AstraProtocol/astra/issues/159))
* update test_fee_payer
* update integration tests for mint
* update integration test cases
* upadte integration tests for mint
* update integration test
* add integration test evm tx
* update integration test feeburn
* add feeburn integration test
* fix unittest invalid denom


<a name="v2.1.3"></a>
## [v2.1.3](https://github.com/AstraProtocol/astra/compare/v2.1.2...v2.1.3) (2022-11-11)

### Chore

* update go release amd64 ([#136](https://github.com/AstraProtocol/astra/issues/136))


<a name="v2.1.2"></a>
## [v2.1.2](https://github.com/AstraProtocol/astra/compare/v2.1.1...v2.1.2) (2022-10-24)

### Deps

* upgrade evmos to v6.1.3-astra ([#129](https://github.com/AstraProtocol/astra/issues/129))
* upgrade evmos to cosmos-sdk v0.45.9 ([#128](https://github.com/AstraProtocol/astra/issues/128))


<a name="v2.1.1"></a>
## [v2.1.1](https://github.com/AstraProtocol/astra/compare/v2.1.0...v2.1.1) (2022-10-20)


<a name="v2.1.0"></a>
## [v2.1.0](https://github.com/AstraProtocol/astra/compare/v2.0.1...v2.1.0) (2022-09-19)


<a name="v2.0.1"></a>
## [v2.0.1](https://github.com/AstraProtocol/astra/compare/v2.0.0...v2.0.1) (2022-09-13)

### Fix

* double supply when add genesis vesting ([#108](https://github.com/AstraProtocol/astra/issues/108))


<a name="v2.0.0"></a>
## [v2.0.0](https://github.com/AstraProtocol/astra/compare/v1.2.2...v2.0.0) (2022-08-23)

### Chore

* update go version and remove package.json ([#100](https://github.com/AstraProtocol/astra/issues/100))

### Deps

* bump ethermint, go-ethereum version ([#98](https://github.com/AstraProtocol/astra/issues/98))

### Docs

* update inflation specs ([#96](https://github.com/AstraProtocol/astra/issues/96))


<a name="v1.2.2"></a>
## [v1.2.2](https://github.com/AstraProtocol/astra/compare/v1.2.1...v1.2.2) (2022-08-15)

### Deps

* Bump cosmos-sdk to v0.45.7, ibc-go to v3.2.0 ([#94](https://github.com/AstraProtocol/astra/issues/94))
* Bump Ethermint version to v0.18.0 ([#90](https://github.com/AstraProtocol/astra/issues/90))

### Docs

* update tokenomics numbers ([#88](https://github.com/AstraProtocol/astra/issues/88))

### Fix

* update docs ([#87](https://github.com/AstraProtocol/astra/issues/87))
* build tag rocksdb ([#84](https://github.com/AstraProtocol/astra/issues/84))

### Test

* update inflation unit-tests ([#92](https://github.com/AstraProtocol/astra/issues/92))


<a name="v1.2.1"></a>
## [v1.2.1](https://github.com/AstraProtocol/astra/compare/v1.2.0...v1.2.1) (2022-08-03)


<a name="v1.2.0"></a>
## [v1.2.0](https://github.com/AstraProtocol/astra/compare/v1.1.0...v1.2.0) (2022-08-01)

### Deps

* bump ethermint to 0.17.3-astra ([#77](https://github.com/AstraProtocol/astra/issues/77))

### Fix

* update version ethermint ([#79](https://github.com/AstraProtocol/astra/issues/79))


<a name="v1.1.0"></a>
## [v1.1.0](https://github.com/AstraProtocol/astra/compare/v1.0.0...v1.1.0) (2022-07-28)

### Chore

* add changelog ([#73](https://github.com/AstraProtocol/astra/issues/73))


<a name="v1.0.0"></a>
## [v1.0.0](https://github.com/AstraProtocol/astra/compare/v0.3.2...v1.0.0) (2022-07-13)

### Docs

* add codecov ([#60](https://github.com/AstraProtocol/astra/issues/60))

### Feat

* add module fees and update evmos to v6 ([#59](https://github.com/AstraProtocol/astra/issues/59))

### Fix

* update codeql ([#65](https://github.com/AstraProtocol/astra/issues/65))

### Refactor

* remove mint ([#72](https://github.com/AstraProtocol/astra/issues/72))
* update swagger ([#68](https://github.com/AstraProtocol/astra/issues/68))

### Test

* add test ([#70](https://github.com/AstraProtocol/astra/issues/70))


<a name="v0.3.2"></a>
## [v0.3.2](https://github.com/AstraProtocol/astra/compare/v0.3.1...v0.3.2) (2022-06-23)

### All

* bump version Astra to v2 ([#43](https://github.com/AstraProtocol/astra/issues/43))

### Bump

* ibc to v3.1.0 ([#45](https://github.com/AstraProtocol/astra/issues/45))

### Deps

* bump go-ethereum version from v1.10.16 -> v1.10.17 ([#48](https://github.com/AstraProtocol/astra/issues/48))
* bump go-ethereum version from v1.10.16 -> v1.10.18 ([#46](https://github.com/AstraProtocol/astra/issues/46))

### Refactor

* update workflows and bot ([#55](https://github.com/AstraProtocol/astra/issues/55))
* remove module epoch ([#53](https://github.com/AstraProtocol/astra/issues/53))


<a name="v0.3.1"></a>
## [v0.3.1](https://github.com/AstraProtocol/astra/compare/v0.3.0...v0.3.1) (2022-06-13)

### Feat

* add github workflows ([#42](https://github.com/AstraProtocol/astra/issues/42))


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/AstraProtocol/astra/compare/v0.2.0...v0.3.0) (2022-06-10)

### Deps

* bump ethermint to v0.16.1 ([#41](https://github.com/AstraProtocol/astra/issues/41))

### Fix

* buf protoc was moved to buf alpha protoc ([#25](https://github.com/AstraProtocol/astra/issues/25))


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/AstraProtocol/astra/compare/v0.2.1...v0.2.0) (2022-05-18)


<a name="v0.2.1"></a>
## [v0.2.1](https://github.com/AstraProtocol/astra/compare/v0.1.0...v0.2.1) (2022-05-18)

### Refactor

* off fork testnet ([#19](https://github.com/AstraProtocol/astra/issues/19))

### Test

* update erc20 testcase
* add TestConvertERC20NativeCoin
* add TestConvertCoinNativeCoin
* update TestRegisterCoin
* add TestRegisterCoin
* update test RegisterERC20
* add MọckEvmKeeper, MockBankKeeper
* add test for proposals ToggleConversion
* add crawback vesting accounts
* add crawback vesting accounts tokens
* refactor helper
* add test vesting

### Pull Requests

* Merge pull request [#12](https://github.com/AstraProtocol/astra/issues/12) from AstraProtocol/dependabot/go_modules/github.com/tharsis/evmos/v4-4.0.1
* Merge pull request [#11](https://github.com/AstraProtocol/astra/issues/11) from AstraProtocol/test/erc20
* Merge pull request [#10](https://github.com/AstraProtocol/astra/issues/10) from AstraProtocol/update/evmos
* Merge pull request [#8](https://github.com/AstraProtocol/astra/issues/8) from AstraProtocol/add_tests


<a name="v0.1.0"></a>
## v0.1.0 (2022-05-02)

### Docs

* update pull request template
* update pull request template
* Add github action for CI

### Feat

* add fork BlocksPerYear in testnet

### Test

* add test for ibc
* add test for erc20

### Pull Requests

* Merge pull request [#6](https://github.com/AstraProtocol/astra/issues/6) from AstraProtocol/upgrade_blockperyear
* Merge pull request [#2](https://github.com/AstraProtocol/astra/issues/2) from AstraProtocol/dependabot/go_modules/github.com/tendermint/tendermint-0.35.4
* Merge pull request [#5](https://github.com/AstraProtocol/astra/issues/5) from AstraProtocol/add_test
* Merge pull request [#4](https://github.com/AstraProtocol/astra/issues/4) from AstraProtocol/ci
* Merge pull request [#1](https://github.com/AstraProtocol/astra/issues/1) from AstraProtocol/ci


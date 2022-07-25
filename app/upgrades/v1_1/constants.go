package v1_1

import "time"

const (
	// UpgradeName is the shared upgrade plan name for testnet
	UpgradeName = "v1.0.1"
	// TestUpgradeHeight defines the Evmos mainnet block height on which the upgrade will take place
	TestUpgradeHeight = 253230 // (24 * 60 * 60) / 2.6 + 220000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = ``

	AvgBlockTime = 3 * time.Second

	NewMaxGas = 40_000_000
)

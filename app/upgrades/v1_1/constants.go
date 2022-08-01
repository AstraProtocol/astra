package v1_1

import "time"

const (
	// UpgradeName is the shared upgrade plan name for testnet
	UpgradeName = "v1.1.0"
	// TestUpgradeHeight defines the Astra testnet block height on which the upgrade will take place
	TestUpgradeHeight = 390000
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/astraProtocol/astra/releases/download/v1.1.0/astra_1.1.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/astraProtocol/astra/releases/download/v1.1.0/astra_1.1.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/astraProtocol/astra/releases/download/v1.1.0/astra_1.1.0_Linux_arm64.tar.gz","linux/x86_64":"https://github.com/astraProtocol/astra/releases/download/v1.1.0/astra_1.1.0_Linux_x86_64.tar.gz","windows/x86_64":"https://github.com/astraProtocol/astra/releases/download/v1.1.0/astra_1.1.0_Windows_x86_64.zip"}}'`

	AvgBlockTime = 3 * time.Second

	NewMaxGas = int64(40_000_000)
)

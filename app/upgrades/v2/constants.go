package v2

const (
	// UpgradeName is the shared upgrade plan name for testnet
	UpgradeName = "v2.1.1"
	// TestUpgradeHeight defines the Astra testnet block height on which the upgrade will take place
	TestUpgradeHeight = 3222222
	// UpgradeInfo defines the binaries that will be used for the upgrade
	UpgradeInfo = `'{"binaries":{"darwin/arm64":"https://github.com/astraProtocol/astra/releases/download/v2.1.1/astra_2.1.1_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/astraProtocol/astra/releases/download/v2.1.1/astra_2.1.1_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/astraProtocol/astra/releases/download/2.1.1/astra_2.1.1_Linux_arm64.tar.gz","linux/amd64":"https://github.com/astraProtocol/astra/releases/download/v2.1.1/astra_2.1.1_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/astraProtocol/astra/releases/download/v2.1.1/astra_2.1.1_Windows_x86_64.zip"}}'`
)

package main_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/AstraProtocol/astra/v2/app"
	astrad "github.com/AstraProtocol/astra/v2/cmd/astrad"
)

func TestInitCmd(t *testing.T) {
	rootCmd, _ := astrad.NewRootCmd()
	rootCmd.SetArgs([]string{
		"init",       // Test the init cmd
		"astra-test", // Moniker
		fmt.Sprintf("--%s=%s", cli.FlagOverwrite, "true"), // Overwrite genesis.json, in case it already exists
		fmt.Sprintf("--%s=%s", flags.FlagChainID, "astra_11110-1"),
	})

	err := svrcmd.Execute(rootCmd, app.DefaultNodeHome)
	require.NoError(t, err)
}

func TestAddKeyLedgerCmd(t *testing.T) {
	rootCmd, _ := astrad.NewRootCmd()
	rootCmd.SetArgs([]string{
		"keys",
		"add",
		"mykey",
		fmt.Sprintf("--%s", flags.FlagUseLedger),
	})

	err := svrcmd.Execute(rootCmd, app.DefaultNodeHome)
	require.Error(t, err)
}

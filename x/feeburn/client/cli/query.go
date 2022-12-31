package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/AstraProtocol/astra/v2/x/feeburn/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group feeburn queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(GetCmdQueryTotalFeeBurn())

	return cmd
}

// GetCmdQueryTotalFeeBurn implements a command to return the total fee burn.
func GetCmdQueryTotalFeeBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-fee-burn",
		Short: "Query the total ASAs burn through the `feeburn` module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryTotalFeeBurnRequest{}
			res, err := queryClient.TotalFeeBurn(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.TotalFeeBurn)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

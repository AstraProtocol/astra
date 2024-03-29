package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AstraProtocol/astra/v3/x/mint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the minting module.
func GetQueryCmd() *cobra.Command {
	mintingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the minting module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	mintingQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetCmdQueryInflation(),
		GetCmdQueryAnnualProvisions(),
		GetCmdQueryTotalMintedProvision(),
		GetCmdQueryBlockProvision(),
		GetCmdQueryCirculatingSupply(),
		GetCmdQueryBondedRatio(),
	)

	return mintingQueryCmd
}

// GetCmdQueryParams implements a command to return the current minting
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current minting parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryInflation implements a command to return the current minting
// inflation value.
func GetCmdQueryInflation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation",
		Short: "Query the current minting inflation value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryInflationRequest{}
			res, err := queryClient.Inflation(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.Inflation))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAnnualProvisions implements a command to return the current minting
// annual provisions value.
func GetCmdQueryAnnualProvisions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "annual-provisions",
		Short: "Query the current minting annual provisions value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAnnualProvisionsRequest{}
			res, err := queryClient.AnnualProvisions(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.AnnualProvisions))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalMintedProvision implements a command to return the total minted provision.
func GetCmdQueryTotalMintedProvision() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-minted-provision",
		Short: "Query the total minted ASAs through the `mint` module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryTotalMintedProvisionRequest{}
			res, err := queryClient.TotalMintedProvision(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.TotalMintedProvision)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBlockProvision implements a command to return the current block provision.
func GetCmdQueryBlockProvision() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-provision",
		Short: "Query the current block provision",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryBlockProvisionRequest{}
			res, err := queryClient.BlockProvision(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Provision)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryCirculatingSupply implements a command to return the current circulating supply.
func GetCmdQueryCirculatingSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circulating-supply",
		Short: "Query the current circulating supply",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryCirculatingSupplyRequest{}
			res, err := queryClient.CirculatingSupply(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.CirculatingSupply)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryBondedRatio implements a command to return the current bonded ratio.
func GetCmdQueryBondedRatio() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bonded-ratio",
		Short: "Query the current bonded ratio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryBondedRatioRequest{}
			res, err := queryClient.GetBondedRatio(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", res.BondedRatio))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

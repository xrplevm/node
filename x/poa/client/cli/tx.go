package cli

import (
	"fmt"
	"strings"

	"github.com/Peersyst/exrp/x/poa/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "PoA transaction subcommands",
	}

	cmd.AddCommand(NewSubmitAddValidatorProposalTxCmd())
	cmd.AddCommand(NewSubmitRemoveValidatorProposalTxCmd())

	return cmd
}

// NewSubmitAddValidatorProposalTxCmd returns a CLI command handler for creating
// a new validator proposal governance transaction.
func NewSubmitAddValidatorProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validator [address] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an AddValidator proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an AddValidator proposal

Example:
$ %s tx poa add-validator <address> --from=<key_or_address>
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			address := args[0]
			content := types.NewAddValidatorProposal("add validator", "proposal to add a new validator in the chain", address)
			fmt.Println(address)
			fmt.Printf("%+v\n", content)
			from := clientCtx.GetFromAddress()
			fmt.Printf("%+v", from)

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				panic(err)
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitRemoveValidatorProposalTxCmd returns a CLI command handler for creating
// a remove validator proposal governance transaction.
func NewSubmitRemoveValidatorProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-validator [address] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit an RemoveValidator proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an RemoveValidator proposal

Example:
$ %s tx poa remove-validator <address> --from=<key_or_address>
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			address := args[0]
			content := types.NewRemoveValidatorProposal("remove validator", "proposal to remove an existent validator", address)

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				panic(err)
			}

			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

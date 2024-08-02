package cli

import (
	"fmt"
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/xrplevm/node/v2/x/poa/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/version"
)

const (
	FlagAddress = "address"
)

func flagSetDescription() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(stakingcli.FlagMoniker, "", "The validator's name")
	fs.String(stakingcli.FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")

	return fs
}

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Proof of Authority module transaction subcommands",
	}

	cmd.AddCommand(NewSubmitAddValidatorProposalTxCmd())
	cmd.AddCommand(NewSubmitRemoveValidatorProposalTxCmd())

	return cmd
}

// NewSubmitAddValidatorProposalTxCmd returns a CLI command handler for creating
// a new validator proposal governance transaction.
func NewSubmitAddValidatorProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-validator [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a governance proposal for adding a new validator in the proof of authority",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a governance proposal for adding a new validator in the proof of authority

Example:
$ %s tx poa add-validator \
	--address <validator owner address> \
	--pubkey <validator tendermint public key> \
	--moniker <validator moniker> \
	--title <proposal title> \
	--summary <proposal summary> \
	--deposit 50000000000000000000axrp
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := cli.ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			fs := cmd.Flags()

			authority := sdk.AccAddress(address.Module("gov"))

			rawAddress, _ := cmd.Flags().GetString(FlagAddress)
			addr, err := sdk.AccAddressFromBech32(rawAddress)
			if err != nil {
				return err
			}

			pkStr, err := fs.GetString(stakingcli.FlagPubKey)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &pk); err != nil {
				return err
			}

			moniker, _ := fs.GetString(stakingcli.FlagMoniker)
			identity, _ := fs.GetString(stakingcli.FlagIdentity)

			msg, err := types.NewMsgAddValidator(authority.String(), addr.String(), pk,
				stakingtypes.Description{
					Moniker:  moniker,
					Identity: identity,
				},
			)
			if err != nil {
				return err
			}

			if err := proposal.SetMsgs([]sdk.Msg{
				msg,
			}); err != nil {
				return fmt.Errorf("failed to create add validator proposal message: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	cmd.Flags().String(FlagAddress, "", "Address of the validator to be added in the proof of authority")
	cmd.Flags().AddFlagSet(stakingcli.FlagSetPublicKey())
	cmd.Flags().AddFlagSet(flagSetDescription())

	flags.AddTxFlagsToCmd(cmd)

	cli.AddGovPropFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagAddress)
	_ = cmd.MarkFlagRequired(stakingcli.FlagMoniker)

	return cmd
}

// NewSubmitRemoveValidatorProposalTxCmd returns a CLI command handler for creating
// a remove validator proposal governance transaction.
func NewSubmitRemoveValidatorProposalTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-validator [address] [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a governance proposal for removing a current validator from the proof of authority",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a governance proposal for removing a current validator from the proof of authority

Example:
$ %s tx poa remove-validator \
	--address <validator owner address>
	--title <proposal title> \
	--summary <proposal summary> \
	--deposit 50000000000000000000axrp
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposal, err := cli.ReadGovPropFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			authority := sdk.AccAddress(address.Module("gov"))

			rawAddress, _ := cmd.Flags().GetString(FlagAddress)
			addr, err := sdk.AccAddressFromBech32(rawAddress)
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveValidator(authority.String(), addr.String())

			if err := proposal.SetMsgs([]sdk.Msg{
				msg,
			}); err != nil {
				return fmt.Errorf("failed to create remove validator proposal message: %w", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
		},
	}
	cmd.Flags().String(FlagAddress, "", "Address of the validator to be removed from the proof of authority")

	flags.AddTxFlagsToCmd(cmd)

	cli.AddGovPropFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired(FlagAddress)

	return cmd
}

package e2e

import (
	"fmt"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	"github.com/xrplevm/node/v2/x/poa/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/xrplevm/node/v2/testutil/network"
)

func transactionFlags(s *IntegrationTestSuite, val network.Validator) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewCoin(s.Cfg.TokenDenom, sdk.NewInt(1000000000))).String()),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "1000000"),
	}
}

func addValidatorMsg(ctx client.Context, newValidatorAddress string, pubKey cryptotypes.PubKey) string {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	msg, err := types.NewMsgAddValidator(authority.String(), newValidatorAddress, pubKey, stakingtypes.Description{Moniker: "Test node"})
	if err != nil {
		panic(err)
	}
	rawMsg, err := ctx.Codec.MarshalInterfaceJSON(msg)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf(`
{
	"messages": [
		%s
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"metadata": "ipfs://CID",
	"deposit": "%s"
}`, rawMsg, GetMinDeposit(ctx))
}

func removeValidatorMsg(ctx client.Context, newValidatorAddress string) string {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	return fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/packages.blockchain.poa.MsgRemoveValidator",
		    "authority": "%s",
		    "validator_address": "%s"
		}
	],
	"title": "My awesome title",
	"summary": "My awesome description",
	"metadata": "ipfs://CID",
	"deposit": "%s"
}`, authority, newValidatorAddress, GetMinDeposit(ctx))
}

func submitProposal(s *IntegrationTestSuite, val network.Validator, content string) string {
	propFile := testutil.WriteToNewTempFile(s.T(), content)
	defer propFile.Close()

	cmd := govcli.NewCmdSubmitProposal()
	clientCtx := val.ClientCtx

	args := []string{
		propFile.Name(),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, val)...))
	s.Require().NoError(err)

	fmt.Printf("Submitted new proposal transaction: %+v\n", out.String())

	return s.ConsumeProposalCount()
}

func voteProposal(s *IntegrationTestSuite, val network.Validator, id string) {
	cmd := govcli.NewCmdVote()
	clientCtx := val.ClientCtx

	args := []string{
		id,
		"yes",
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, val)...))
	s.Require().NoError(err)

	fmt.Printf("Submitted vote proposal transaction: %+v\n", out.String())
}

var (
	AddValidatorAction    = 0
	RemoveValidatorAction = 1
)

func ChangeValidator(
	s *IntegrationTestSuite,
	action int,
	address sdk.AccAddress,
	pubKey cryptotypes.PubKey,
	// pubKey string,
	validators []*network.Validator,
	waitStatus govtypesv1.ProposalStatus,
) {
	initiator := validators[0]
	var msg string
	if action == AddValidatorAction {
		msg = addValidatorMsg(initiator.ClientCtx, address.String(), pubKey)
	} else {
		msg = removeValidatorMsg(initiator.ClientCtx, address.String())
	}

	proposalId := submitProposal(s, *initiator, msg)

	s.Network.MustWaitForNextBlock()

	for _, validator := range validators {
		voteProposal(s, *validator, proposalId)
	}
	s.Network.MustWaitForNextBlock()

	timeLimit := time.Now().Add(time.Minute * 2)
	for time.Now().Before(timeLimit) &&
		waitStatus != govtypesv1.StatusNil &&
		GetProposal(initiator.ClientCtx, proposalId).Status != waitStatus {
		s.Network.MustWaitForNextBlock()
	}
}

func BondTokens(s *IntegrationTestSuite, validator *network.Validator, tokens sdk.Int) {
	clientCtx := validator.ClientCtx
	cmd := stakingcli.NewCreateValidatorCmd()

	json, _ := clientCtx.Codec.MarshalInterfaceJSON(validator.PubKey)

	args := []string{
		fmt.Sprintf("--%s=%s", stakingcli.FlagAmount, sdk.NewCoin(s.Cfg.BondDenom, tokens).String()),
		fmt.Sprintf("--%s=%s", stakingcli.FlagPubKey, json),
		fmt.Sprintf("--%s=%s", stakingcli.FlagMoniker, "moniker"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionRate, "0.1"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionMaxRate, "0.2"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionMaxChangeRate, "0.01"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagMinSelfDelegation, "1"),
	}

	ExecTransaction(s, validator, cmd, args)

	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}
}

func UnBondTokens(s *IntegrationTestSuite, validator *network.Validator, tokens sdk.Int, wait bool) string {
	cmd := stakingcli.NewUnbondCmd()

	args := []string{
		sdk.ValAddress(validator.Address).String(),
		sdk.NewCoin(s.Cfg.BondDenom, tokens).String(),
	}

	out := ExecTransaction(s, validator, cmd, args)

	if wait {
		time.Sleep(s.Cfg.UnBoundingTime)
	}

	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	return out
}

func Delegate(s *IntegrationTestSuite, delegator *network.Validator, validator *network.Validator, tokens sdk.Int) string {
	cmd := stakingcli.NewDelegateCmd()

	args := []string{
		sdk.ValAddress(validator.Address).String(),
		sdk.NewCoin(s.Cfg.BondDenom, tokens).String(),
	}

	out := ExecTransaction(s, validator, cmd, args)

	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.Network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	return out
}

func Redelegate(s *IntegrationTestSuite, src *network.Validator, dst *network.Validator, tokens sdk.Int) string {
	cmd := stakingcli.NewRedelegateCmd()
	args := []string{
		sdk.ValAddress(src.Address).String(),
		sdk.ValAddress(dst.Address).String(),
		sdk.NewCoin(s.Cfg.BondDenom, tokens).String(),
	}
	return ExecTransaction(s, src, cmd, args)
}

func ExecTransaction(s *IntegrationTestSuite, validator *network.Validator, cmd *cobra.Command, args []string) string {
	clientCtx := validator.ClientCtx
	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, *validator)...))
	s.Require().NoError(err)

	fmt.Printf("Submitted transaction: %+v\n", out.String())

	return out.String()
}

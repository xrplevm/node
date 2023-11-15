package poa

import (
	"fmt"
	"time"

	"github.com/Peersyst/exrp/testutil/network"
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
)

func transactionFlags(s *TestSuite, val network.Validator) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewCoin(s.cfg.TokenDenom, sdk.NewInt(1000000000))).String()),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "1000000"),
	}
}

func addValidatorMsg(ctx client.Context, newValidatorAddress string) string {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	return fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/exrp.poa.MsgAddValidator",
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

func removeValidatorMsg(ctx client.Context, newValidatorAddress string) string {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	return fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/exrp.poa.MsgRemoveValidator",
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

func submitProposal(s *TestSuite, val network.Validator, content string) string {
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

func voteProposal(s *TestSuite, val network.Validator, id string) {
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
	s *TestSuite,
	action int,
	address string,
	validators []*network.Validator,
	waitStatus govtypesv1.ProposalStatus,
) {
	initiator := validators[0]
	var msg string
	if action == AddValidatorAction {
		msg = addValidatorMsg(initiator.ClientCtx, address)
	} else {
		msg = removeValidatorMsg(initiator.ClientCtx, address)
	}

	proposalId := submitProposal(s, *initiator, msg)

	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	for _, validator := range validators {
		voteProposal(s, *validator, proposalId)
	}
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	timeLimit := time.Now().Add(time.Minute * 2)
	for time.Now().Before(timeLimit) &&
		waitStatus != govtypesv1.StatusNil &&
		GetProposal(initiator.ClientCtx, proposalId).Status != waitStatus {
		if err := s.network.WaitForNextBlock(); err != nil {
			panic(err)
		}
	}
}

func BondTokens(s *TestSuite, validator *network.Validator, tokens sdk.Int) {
	clientCtx := validator.ClientCtx
	cmd := stakingcli.NewCreateValidatorCmd()

	json, _ := clientCtx.Codec.MarshalInterfaceJSON(validator.PubKey)

	args := []string{
		fmt.Sprintf("--%s=%s", stakingcli.FlagAmount, sdk.NewCoin(s.cfg.BondDenom, tokens).String()),
		fmt.Sprintf("--%s=%s", stakingcli.FlagPubKey, json),
		fmt.Sprintf("--%s=%s", stakingcli.FlagMoniker, "moniker"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionRate, "0.1"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionMaxRate, "0.2"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagCommissionMaxChangeRate, "0.01"),
		fmt.Sprintf("--%s=%s", stakingcli.FlagMinSelfDelegation, "1"),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, *validator)...))
	s.Require().NoError(err)

	fmt.Printf("Submitted bond transaction: %+v\n", out.String())

	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
}

func UnBondTokens(s *TestSuite, validator *network.Validator, tokens sdk.Int, wait bool) {
	clientCtx := validator.ClientCtx
	cmd := stakingcli.NewUnbondCmd()

	args := []string{
		sdk.ValAddress(validator.Address).String(),
		sdk.NewCoin(s.cfg.BondDenom, tokens).String(),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, *validator)...))
	s.Require().NoError(err)

	fmt.Printf("Submitted unbound transaction: %+v\n", out.String())

	if wait {
		time.Sleep(s.cfg.UnBoundingTime)
	}

	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
}

func Delegate(s *TestSuite, delegator *network.Validator, validator *network.Validator, tokens sdk.Int) {
	clientCtx := delegator.ClientCtx
	cmd := stakingcli.NewDelegateCmd()

	args := []string{
		sdk.ValAddress(validator.Address).String(),
		sdk.NewCoin(s.cfg.BondDenom, tokens).String(),
	}

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, append(args, transactionFlags(s, *delegator)...))
	fmt.Printf("Submitted delegation transaction: %+v\n", out.String())
	s.Require().NoError(err)

	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
}

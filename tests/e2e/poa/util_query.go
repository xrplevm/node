package poa

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func GetMinDeposit(ctx client.Context) sdk.Coin {
	args := []string{
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, govcli.GetCmdQueryParams(), args)

	var params govtypesv1.Params
	if err := json.Unmarshal(res.Bytes(), &params); err != nil {
		fmt.Printf("Error unmarshaling json %v - %v", err, res.String())
	}

	return params.MinDeposit[0]
}

type GetBalanceResponse struct {
	Balances []sdk.Coin
}

func GetBalance(ctx client.Context, address string, denom string) sdk.Coin {
	args := []string{
		fmt.Sprint(address),
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), args)

	var parsedRes GetBalanceResponse
	if err := json.Unmarshal(res.Bytes(), &parsedRes); err != nil {
		panic(err)
	}

	for _, coin := range parsedRes.Balances {
		if coin.Denom == denom {
			return coin
		}
	}
	return sdk.NewCoin(denom, sdk.NewInt(0))
}

func GetValidator(ctx client.Context, address string) *stakingtypes.Validator {
	args := []string{
		fmt.Sprint(address),
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, stakingcli.GetCmdQueryValidator(), args)

	var parsed stakingtypes.Validator
	err := ctx.Codec.UnmarshalJSON(res.Bytes(), &parsed)
	if err != nil {
		return nil
	}

	return &parsed
}

func GetDelegation(ctx client.Context, validatorAddress string, delegatorAddress string) *stakingtypes.Delegation {
	args := []string{
		fmt.Sprint(delegatorAddress),
		fmt.Sprint(validatorAddress),
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, stakingcli.GetCmdQueryDelegation(), args)

	var parsed stakingtypes.Delegation
	err := ctx.Codec.UnmarshalJSON(res.Bytes(), &parsed)
	if err != nil {
		return nil
	}

	return &parsed
}

func GetProposal(ctx client.Context, id string) govtypesv1.Proposal {
	args := []string{
		fmt.Sprint(id),
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, govcli.GetCmdQueryProposal(), args)

	var parsed govtypesv1.Proposal
	ctx.Codec.MustUnmarshalJSON(res.Bytes(), &parsed)

	return parsed
}

type GetValidatorSetResponse struct {
	Validators []struct {
		Address     string
		VotingPower string
	}
}

func GetValidatorSet(ctx client.Context) *GetValidatorSetResponse {
	args := []string{
		fmt.Sprint(ctx.Height),
		"--output=json",
	}
	res, _ := clitestutil.ExecTestCLICmd(ctx, rpc.ValidatorCommand(), args)

	var parsed GetValidatorSetResponse
	if err := json.Unmarshal(res.Bytes(), &parsed); err != nil {
		return nil
	}

	return &parsed
}

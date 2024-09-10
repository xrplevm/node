package simulation

import (
	"math/rand"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/xrplevm/node/v3/x/poa/types"
)


const (
	OpWeightMsgAddValidator = "op_weight_msg_add_validator"
	DefaultWeightMsgAddValidator int = 80
	
	OpWeightMsgRemoveValidator = "op_weight_msg_remove_validator"
	DefaultWeightMsgRemoveValidator int = 20
)

func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgAddValidator,
			DefaultWeightMsgAddValidator,
			SimulateMsgAddValidator,
		),
		simulation.NewWeightedProposalMsg(
			OpWeightMsgRemoveValidator,
			DefaultWeightMsgRemoveValidator,
			SimulateMsgRemoveValidator,
		),
	}
}

// MsgAddValidator simulation functions

// randomDescription generates a random description for a validator
func randomDescription(r *rand.Rand) stakingtypes.Description {
	return stakingtypes.Description{
		Moniker: 		 simtypes.RandStringOfLength(r, 10),
		Identity: 		 simtypes.RandStringOfLength(r, 10),
		Website:         simtypes.RandStringOfLength(r, 10),
		SecurityContact: simtypes.RandStringOfLength(r, 10),
		Details:         simtypes.RandStringOfLength(r, 10),
	}
}

// randomMsgAddValidator generates a random MsgAddValidator message
func randomMsgAddValidator(r *rand.Rand, authAddr sdk.AccAddress) (*types.MsgAddValidator, error) {
	validatorAccs := simtypes.RandomAccounts(r, 1)

	validatorAcc := validatorAccs[0]
	pubkey, err := codectypes.NewAnyWithValue(validatorAcc.PubKey)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddValidator{
		Authority: authAddr.String(),
		Description: randomDescription(r),
		ValidatorAddress: validatorAcc.Address.String(),
		Pubkey: pubkey, 
	}, nil
}

// SimulateMsgAddValidator simulates the MsgAddValidator message
func SimulateMsgAddValidator(r *rand.Rand, _ sdk.Context, accs []simtypes.Account) sdk.Msg {
	var authAddr sdk.AccAddress = address.Module("gov")

	randMsg, err := randomMsgAddValidator(r, authAddr)
	if err != nil {
		panic(err)
	}
	return randMsg

}

// MsgRemoveValidator simulation functions

// randomMsgRemoveValidator generates a random MsgRemoveValidator message
func randomMsgRemoveValidator(r *rand.Rand, authAddr sdk.AccAddress, accs []simtypes.Account) *types.MsgRemoveValidator {
	rmValidator, _ := simtypes.RandomAcc(r, accs)

	return &types.MsgRemoveValidator{
		Authority: authAddr.String(),
		ValidatorAddress: rmValidator.Address.String(),
	}
}

// SimulateMsgRemoveValidator simulates the MsgRemoveValidator message
func SimulateMsgRemoveValidator(r *rand.Rand, _ sdk.Context, accs []simtypes.Account) sdk.Msg {
	var authAddr sdk.AccAddress = address.Module("gov")
	
	randMsg := randomMsgRemoveValidator(r, authAddr, accs)
	return randMsg
}

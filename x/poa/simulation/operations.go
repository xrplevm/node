package simulation

import (
	"github.com/Peersyst/exrp/x/poa/types"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateAddValidatorProposalContent() simulation.
	ContentSimulatorFn {
	return func(r *rand.Rand, _ sdk.Context, accs []simulation.Account) simulation.Content {
		simAccount, _ := simulation.RandomAcc(r, accs)

		msg := types.NewAddValidatorProposal(simulation.RandStringOfLength(r, 10), simulation.RandStringOfLength(r, 10), simAccount.Address.String())

		return msg
	}
}

func SimulateRemoveValidatorProposalContent(
	bk types.BankKeeper,
	sk types.StakingKeeper,
) simulation.
	ContentSimulatorFn {
	return func(r *rand.Rand, ctx sdk.Context, accs []simulation.Account) simulation.Content {
		var validators []sdk.Address
		bk.IterateAllBalances(ctx, func(addr sdk.AccAddress, coin sdk.Coin) bool {
			if coin.Denom == sk.GetParams(ctx).BondDenom && !coin.IsZero() {
				validators = append(validators, addr)
			}
			return false
		})

		if len(validators) == 0 {
			return nil
		}

		validatorAddress := RandomAccount(r, validators)

		msg := types.NewRemoveValidatorProposal(simulation.RandStringOfLength(r, 10), simulation.RandStringOfLength(r, 10), validatorAddress.String())

		return msg
	}
}

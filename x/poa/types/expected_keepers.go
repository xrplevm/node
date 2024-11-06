package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}

// StakingKeeper defines the expected interface needed to retrieve account balances.
type StakingKeeper interface {
	GetParams(ctx sdk.Context) stakingtypes.Params
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	GetValidators(ctx sdk.Context, maxRetrieve uint32) (validators []stakingtypes.Validator)
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetAllDelegations(ctx sdk.Context) (delegations []stakingtypes.Delegation)
	GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation
	GetUnbondingDelegationsFromValidator(ctx sdk.Context, validator sdk.ValAddress) []stakingtypes.UnbondingDelegation
	SlashUnbondingDelegation(ctx sdk.Context, ubd stakingtypes.UnbondingDelegation, infractionHeight int64, slashFactor sdk.Dec) (totalSlashAmount math.Int)
	RemoveDelegation(ctx sdk.Context, delegation stakingtypes.Delegation) error
	RemoveValidatorTokensAndShares(ctx sdk.Context, validator stakingtypes.Validator, sharesToRemove math.LegacyDec) (stakingtypes.Validator, math.Int)
	RemoveValidatorTokens(ctx sdk.Context, validator stakingtypes.Validator, tokensToRemove math.Int) stakingtypes.Validator
	BondDenom(ctx sdk.Context) string
}

type SlashingKeeper interface {
	GetParams(ctx sdk.Context) (params slashingtypes.Params)
}

type GovKeeper interface {
	SubmitProposal(ctx sdk.Context, messages []sdk.Msg, metadata string) (v1.Proposal, error)
}

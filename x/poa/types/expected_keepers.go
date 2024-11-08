package types

import (
	"context"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	IterateAllBalances(ctx context.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
}

// StakingKeeper defines the expected interface needed to retrieve account balances.
type StakingKeeper interface {
	GetParams(ctx context.Context) (stakingtypes.Params, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error)
	GetValidators(ctx context.Context, maxRetrieve uint32) (validators []stakingtypes.Validator, err error)
	GetAllValidators(ctx context.Context) (validators []stakingtypes.Validator, err error)
	GetAllDelegations(ctx context.Context) (delegations []stakingtypes.Delegation, err error)
	GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error)
	GetUnbondingDelegationsFromValidator(ctx context.Context, validator sdk.ValAddress) ([]stakingtypes.UnbondingDelegation, error)
	SlashUnbondingDelegation(ctx context.Context, ubd stakingtypes.UnbondingDelegation, infractionHeight int64, slashFactor math.LegacyDec) (totalSlashAmount math.Int, err error)
	RemoveDelegation(ctx context.Context, delegation stakingtypes.Delegation) error
	RemoveValidatorTokensAndShares(ctx context.Context, validator stakingtypes.Validator, sharesToRemove math.LegacyDec) (stakingtypes.Validator, math.Int, error)
	RemoveValidatorTokens(ctx context.Context, validator stakingtypes.Validator, tokensToRemove math.Int) (stakingtypes.Validator, error)
	BondDenom(ctx context.Context) (string, error)
}

type SlashingKeeper interface {
	GetParams(ctx context.Context) (params slashingtypes.Params, err error)
}

type GovKeeper interface {
	SubmitProposal(ctx context.Context, messages []sdk.Msg, metadata string) (v1.Proposal, error)
}

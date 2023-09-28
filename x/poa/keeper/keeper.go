package keeper

import (
	"fmt"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Peersyst/exrp/x/poa/types"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		paramstore     paramtypes.Subspace
		authority      string // the address capable of executing a poa change. Usually the gov module account
		bk             types.BankKeeper
		sk             stakingkeeper.Keeper
		slashingKeeper slashingkeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	bk types.BankKeeper,
	sk stakingkeeper.Keeper,
	slashingKeeper slashingkeeper.Keeper,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:            cdc,
		paramstore:     ps,
		authority:      authority,
		bk:             bk,
		sk:             sk,
		slashingKeeper: slashingKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) ExecuteAddValidator(ctx sdk.Context, validatorAddress string) error {
	// Check if the new validator already has staking power in the bank account
	accAddress, err := sdk.AccAddressFromBech32(validatorAddress)
	valAddress := sdk.ValAddress(accAddress)
	if err != nil {
		return err
	}
	denom := k.sk.GetParams(ctx).BondDenom
	balance := k.bk.GetBalance(ctx, accAddress, denom)
	if !balance.IsZero() {
		// Validator already has staking tokens in bank
		return types.ErrAddressHasBankTokens
	}

	// Check if the validator has bonded tokens in the staking module
	validator, found := k.sk.GetValidator(ctx, valAddress)
	if found && !validator.Tokens.IsZero() {
		// Validator already has staking tokens bonded
		return types.ErrAddressHasBondedTokens
	}

	delegations := k.sk.GetAllDelegatorDelegations(ctx, accAddress)
	// Check if the delegations are greater than 0
	// Validator has delegations to other validators, not eligible for new tokens
	for _, delegation := range delegations {
		if !delegation.Shares.IsZero() {
			delVal, found := k.sk.GetValidator(ctx, delegation.GetValidatorAddr())
			if !found {
				continue
			}
			if !delVal.Tokens.IsZero() {
				return types.ErrAddressHasDelegatedTokens
			}
		}
	}

	// Check if address has unbonding delegations with balance
	// If so, return error since the account already has staking power
	unbondingBalance := sdk.ZeroInt()
	ubds := k.sk.GetUnbondingDelegationsFromValidator(ctx, valAddress)
	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			unbondingBalance = unbondingBalance.Add(entry.Balance)
		}
	}
	if !unbondingBalance.IsZero() {
		return types.ErrAddressHasUnbondingTokens
	}

	// All checks passed, mint new validator tokens and send them to the address
	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.DefaultPowerReduction))
	err = k.bk.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddress, coins)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddValidator,
			sdk.NewAttribute(types.AttributeValidator, accAddress.String()),
			sdk.NewAttribute(types.AttributeHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeStakingTokens, fmt.Sprintf("%d", validator.Tokens)),
			sdk.NewAttribute(types.AttributeBankTokens, balance.String()),
		),
	)

	return nil
}

func (k Keeper) ExecuteRemoveValidator(ctx sdk.Context, validatorAddress string) error {
	accAddress, err := sdk.AccAddressFromBech32(validatorAddress)
	if err != nil {
		return err
	}
	denom := k.sk.GetParams(ctx).BondDenom
	valAddress := sdk.ValAddress(accAddress)

	// Check if address is a validator or has balance some balance in bank
	validator, found := k.sk.GetValidator(ctx, valAddress)
	balance := k.bk.GetBalance(ctx, accAddress, denom)
	if balance.IsZero() && !found {
		// Address has no balance in bank and is not a validator either
		// NOTE: Since delegations are not enabled in this version, we don't need to check for them
		return types.ErrAddressHasNoTokens
	}

	// If address is a validator, we need to check if there are unbonding delegations currently
	// and slash them. We also need to remove all the tokens from the validator and burn them
	// from the staking module account
	if found {
		ubds := k.sk.GetUnbondingDelegationsFromValidator(ctx, valAddress)
		for _, ubd := range ubds {
			k.sk.SlashUnbondingDelegation(ctx, ubd, 0, sdk.OneDec())
		}

		changedVal := k.sk.RemoveValidatorTokens(ctx, validator, validator.Tokens)
		switch changedVal.GetStatus() {
		case stakingtypes.Bonded:
			coins := sdk.NewCoins(sdk.NewCoin(k.sk.BondDenom(ctx), validator.Tokens))
			err = k.bk.BurnCoins(ctx, stakingtypes.BondedPoolName, coins)
			if err != nil {
				panic(err)
			}
		case stakingtypes.Unbonding, stakingtypes.Unbonded:
			coins := sdk.NewCoins(sdk.NewCoin(k.sk.BondDenom(ctx), validator.Tokens))
			err = k.bk.BurnCoins(ctx, stakingtypes.NotBondedPoolName, coins)
			if err != nil {
				panic(err)
			}
		default:
			return types.ErrInvalidValidatorStatus
		}

	}

	// If address also has tokens in the bank, we need to remove them and burn them
	if !balance.IsZero() {
		coins := sdk.NewCoins(balance)
		err = k.bk.SendCoinsFromAccountToModule(ctx, accAddress, types.ModuleName, coins)
		if err != nil {
			return err
		}

		err = k.bk.BurnCoins(ctx, types.ModuleName, coins)
		if err != nil {
			return err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveValidator,
			sdk.NewAttribute(types.AttributeValidator, validatorAddress),
			sdk.NewAttribute(types.AttributeHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeStakingTokens, fmt.Sprintf("%d", validator.Tokens)),
			sdk.NewAttribute(types.AttributeBankTokens, balance.String()),
		),
	)

	return nil
}

package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/xrplevm/node/v6/x/poa/types"
)

type (
	Keeper struct {
		cdc        codec.Codec
		paramstore paramtypes.Subspace
		authority  string                    // the address capable of executing a poa change. Usually the gov module account
		router     *baseapp.MsgServiceRouter // Msg server router
		bk         types.BankKeeper
		sk         types.StakingKeeper
		ck         types.SlashingKeeper
	}
)

func NewKeeper(
	cdc codec.Codec,
	ps paramtypes.Subspace,
	router *baseapp.MsgServiceRouter,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	ck types.SlashingKeeper,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		panic(err)
	}

	return &Keeper{
		cdc:        cdc,
		paramstore: ps,
		authority:  authority,
		router:     router,
		bk:         bk,
		sk:         sk,
		ck:         ck,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Router returns the gov keeper's router
func (k Keeper) Router() *baseapp.MsgServiceRouter {
	return k.router
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) ExecuteAddValidator(ctx sdk.Context, msg *types.MsgAddValidator) error {
	// Check if the new validator already has staking power in the bank account
	accAddress, err := sdk.AccAddressFromBech32(msg.ValidatorAddress)
	valAddress := sdk.ValAddress(accAddress)
	if err != nil {
		return err
	}
	params, err := k.sk.GetParams(ctx)
	if err != nil {
		return err
	}
	denom := params.BondDenom
	balance := k.bk.GetBalance(ctx, accAddress, denom)
	if !balance.IsZero() {
		// Validator already has staking tokens in bank
		return types.ErrAddressHasBankTokens
	}

	// Check if the validator has bonded tokens in the staking module
	validator, err := k.sk.GetValidator(ctx, valAddress)
	if err != nil && !errors.IsOf(err, stakingtypes.ErrNoValidatorFound) {
		return err
	} else if err == nil && !validator.Tokens.IsZero() {
		// Validator already has staking tokens bonded
		return types.ErrAddressHasBondedTokens
	}

	delegations, err := k.sk.GetAllDelegatorDelegations(ctx, accAddress)
	if err != nil {
		return err
	}
	// Check if the delegations are greater than 0
	// Validator has delegations to other validators, not eligible for new tokens
	for _, delegation := range delegations {
		if !delegation.Shares.IsZero() {
			delegationValidatorAddress, err := sdk.ValAddressFromBech32(delegation.GetValidatorAddr())
			if err != nil {
				return err
			}
			delVal, err := k.sk.GetValidator(ctx, delegationValidatorAddress)
			if err != nil && !errors.IsOf(err, stakingtypes.ErrNoValidatorFound) {
				return err
			} else if errors.IsOf(err, stakingtypes.ErrNoValidatorFound) {
				continue
			}
			if !delVal.Tokens.IsZero() {
				return types.ErrAddressHasDelegatedTokens
			}
		}
	}

	// Check if address has unbonding delegations with balance
	// If so, return error since the account already has staking power
	unbondingBalance := math.ZeroInt()
	ubds, err := k.sk.GetUnbondingDelegationsFromValidator(ctx, valAddress)
	if err != nil {
		return err
	}
	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			unbondingBalance = unbondingBalance.Add(entry.Balance)
		}
	}
	if !unbondingBalance.IsZero() {
		return types.ErrAddressHasUnbondingTokens
	}

	// All checks passed, mint new validator tokens and send them to the address
	coin := sdk.NewCoin(denom, sdk.DefaultPowerReduction)
	coins := sdk.NewCoins(coin)
	err = k.bk.MintCoins(ctx, types.ModuleName, coins)
	if err != nil {
		return err
	}
	err = k.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddress, coins)
	if err != nil {
		return err
	}

	pubKey, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pubKey)
	}
	createValidatorMsg, err := stakingtypes.NewMsgCreateValidator(
		valAddress.String(),
		pubKey,
		coin,
		msg.Description,
		stakingtypes.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
		math.OneInt(),
	)
	if err != nil {
		return err
	}
	handler := k.Router().Handler(createValidatorMsg)
	_, err = handler(ctx, createValidatorMsg)
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
	valAddress, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		return err
	}
	params, err := k.sk.GetParams(ctx)
	if err != nil {
		return err
	}
	denom := params.BondDenom

	accAddress := sdk.AccAddress(valAddress)

	// Check if address has some balance in bank and withdraw in case of having
	balance := k.bk.GetBalance(ctx, accAddress, denom)

	// If address also has a validator, we need to check additional conditions
	validator, err := k.sk.GetValidator(ctx, valAddress)
	if err != nil {
		ctx.Logger().Warn("Error getting validator", "error", err)
		if balance.IsZero() {
			// Address has no balance in bank and is not a validator either
			// NOTE: Since delegations are not enabled in this version, we don't need to check for them
			return types.ErrAddressHasNoTokens
		}
		// Validator does not exist, but we already took its balance from bank, we can safely return
		return nil
	}

	if err := k.sk.Hooks().BeforeValidatorModified(ctx, valAddress); err != nil {
		k.Logger(ctx).Error("failed to call before validator modified hook", "error", err)
	}
	// If address is a validator, we need to check if there are unbonding delegations currently
	// and slash them. We also need to remove all the tokens from the validator and burn them
	// from the staking module account
	ubds, err := k.sk.GetUnbondingDelegationsFromValidator(ctx, valAddress)
	if err != nil {
		return err
	}
	for _, ubd := range ubds {
		totalSlashAmount, err := k.sk.SlashUnbondingDelegation(ctx, ubd, 0, math.LegacyOneDec())
		if err != nil {
			return err
		}
		ctx.Logger().Debug("Unbonding delegation slashed", "delegator", ubd.DelegatorAddress, "amount", totalSlashAmount)
	}

	if validator.Tokens.IsPositive() {
		// call the before-slashed hook
		if err := k.sk.Hooks().BeforeValidatorSlashed(ctx, valAddress, math.LegacyOneDec()); err != nil {
			k.Logger(ctx).Error("failed to call before validator slashed hook", "error", err)
		}
	}

	changedVal, err := k.sk.RemoveValidatorTokens(ctx, validator, validator.Tokens)
	if err != nil {
		return err
	}

	switch changedVal.GetStatus() {
	case stakingtypes.Bonded:
		coins := sdk.NewCoins(sdk.NewCoin(denom, validator.Tokens))
		err = k.bk.BurnCoins(ctx, stakingtypes.BondedPoolName, coins)
		if err != nil {
			return err
		}
	case stakingtypes.Unbonding, stakingtypes.Unbonded:
	default:
		return types.ErrInvalidValidatorStatus
	}

	// Unbond self-delegation so the validator is removed after being unbonded
	_, err = k.sk.Unbond(ctx, sdk.AccAddress(valAddress), valAddress, changedVal.DelegatorShares)
	if err != nil {
		return err
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

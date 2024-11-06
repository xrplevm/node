package testutil

import (
	context "context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MsgServer is the server API for Msg service.
type StakingMsgServer interface {
	// CreateValidator defines a method for creating a new validator.
	CreateValidator(context.Context, *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error)
	// EditValidator defines a method for editing an existing validator.
	EditValidator(context.Context, *stakingtypes.MsgEditValidator) (*stakingtypes.MsgEditValidatorResponse, error)
	// Delegate defines a method for performing a delegation of coins
	// from a delegator to a validator.
	Delegate(context.Context, *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error)
	// BeginRedelegate defines a method for performing a redelegation
	// of coins from a delegator and source validator to a destination validator.
	BeginRedelegate(context.Context, *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error)
	// Undelegate defines a method for performing an undelegation from a
	// delegate and a validator.
	Undelegate(context.Context, *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error)
	// CancelUnbondingDelegation defines a method for performing canceling the unbonding delegation
	// and delegate back to previous validator.
	//
	// Since: cosmos-sdk 0.46
	CancelUnbondingDelegation(context.Context, *stakingtypes.MsgCancelUnbondingDelegation) (*stakingtypes.MsgCancelUnbondingDelegationResponse, error)
	// UpdateParams defines an operation for updating the x/staking module
	// parameters.
	// Since: cosmos-sdk 0.47
	UpdateParams(context.Context, *stakingtypes.MsgUpdateParams) (*stakingtypes.MsgUpdateParamsResponse, error)
}

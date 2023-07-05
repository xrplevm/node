package poa

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *TestSuite) Test_AddDelegationIsNotAllowedToOtherValidators() {
	fmt.Println("==== Test_AddDelegationIsNotAllowedToOtherValidators")
	validator := s.network.Validators[0]
	validatorAddress := validator.Address.String()
	delegator := s.network.Validators[s.cfg.NumBondedValidators]
	delegatorAddress := delegator.Address.String()

	// Validator should be bonded and have default bonded tokens power
	s.RequireValidator(validatorAddress, &bondedStatus, &DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, zero)
	// Delegator should not have any shares and should have default bonded tokens in bank
	s.RequireDelegation(validatorAddress, delegatorAddress, sdk.ZeroDec())
	s.RequireBondBalance(delegatorAddress, DefaultBondedTokens)

	Delegate(s, delegator, validator, DefaultBondedTokens)

	// Delegator should not have any shares and should have default bonded tokens in bank
	s.RequireDelegation(validatorAddress, delegatorAddress, sdk.ZeroDec())
	s.RequireBondBalance(delegatorAddress, DefaultBondedTokens)
	fmt.Println("==== [V] Test_AddDelegationIsNotAllowedToOtherValidators")
}

func (s *TestSuite) Test_AddDelegationIsAllowedToSelfValidator() {
	fmt.Println("==== Test_AddDelegationIsAllowedToSelfValidator")
	validator := s.network.Validators[s.cfg.NumBondedValidators]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator should not be bonded and have default bonded tokens power
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, DefaultBondedTokens)

	// EXEC:
	halfTokens := sdk.NewDec(DefaultBondedTokens.Int64()).Quo(sdk.NewDec(2)).RoundInt()

	BondTokens(s, validator, halfTokens)

	// Check validator is active and there are pending bonded tokens in bank
	s.RequireValidator(validatorAddress, &unbondedStatus, &halfTokens)
	s.RequireBondBalance(validatorAddress, halfTokens)

	Delegate(s, validator, validator, halfTokens)

	// POST:
	// Delegator should have all the tokens bonded and delegation should have happened
	s.RequireValidator(validatorAddress, &bondedStatus, &DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, sdk.ZeroInt())
	fmt.Println("==== [V] Test_AddDelegationIsAllowedToSelfValidator")
}

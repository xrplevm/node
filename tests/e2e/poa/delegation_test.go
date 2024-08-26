package poa_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/xrplevm/node/v2/tests/e2e"
)

func (s *TestSuite) Test_AddDelegationIsNotAllowedToOtherValidators() {
	s.T().Logf("==== Test_AddDelegationIsNotAllowedToOtherValidators")
	validator := s.Network.Validators[0]
	validatorAddress := validator.Address.String()
	delegator := s.Network.Validators[s.Cfg.NumBondedValidators]
	delegatorAddress := delegator.Address.String()

	// Validator should be bonded and have default bonded tokens power
	s.RequireValidator(validatorAddress, &e2e.BondedStatus, &e2e.DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	// Delegator should not have any shares and should have default bonded tokens in bank
	s.RequireDelegation(validatorAddress, delegatorAddress, sdk.ZeroDec())
	s.RequireBondBalance(delegatorAddress, e2e.DefaultBondedTokens)

	e2e.Delegate(&s.IntegrationTestSuite, validator, e2e.DefaultBondedTokens)

	// Delegator should not have any shares and should have default bonded tokens in bank
	s.RequireDelegation(validatorAddress, delegatorAddress, sdk.ZeroDec())
	s.RequireBondBalance(delegatorAddress, e2e.DefaultBondedTokens)
	s.T().Logf("==== [V] Test_AddDelegationIsNotAllowedToOtherValidators")
}

func (s *TestSuite) Test_AddDelegationIsAllowedToSelfValidator() {
	s.T().Logf("==== Test_AddDelegationIsAllowedToSelfValidator")
	validator := s.Network.Validators[s.Cfg.NumBondedValidators]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator should not be bonded and have default bonded tokens power
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, e2e.DefaultBondedTokens)

	// EXEC:
	halfTokens := sdk.NewDec(e2e.DefaultBondedTokens.Int64()).Quo(sdk.NewDec(2)).RoundInt()

	e2e.BondTokens(&s.IntegrationTestSuite, validator, halfTokens)

	// Check validator is active and there are pending bonded tokens in bank
	s.RequireValidator(validatorAddress, &e2e.UnbondedStatus, &halfTokens)
	s.RequireBondBalance(validatorAddress, halfTokens)

	e2e.Delegate(&s.IntegrationTestSuite, validator, halfTokens)

	// POST:
	// Delegator should have all the tokens bonded and delegation should have happened
	s.RequireValidator(validatorAddress, &e2e.BondedStatus, &e2e.DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, sdk.ZeroInt())
	s.T().Logf("==== [V] Test_AddDelegationIsAllowedToSelfValidator")
}

package poa_test

import (
	"fmt"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (s *IntegrationTestSuite) Test_AddNewValidator() {
	fmt.Println("==== Test_AddNewValidator")

	address := "evmos1ycvhcxthjju0466d4ga0j7du7wt8kmaep28zqv"
	validators := s.network.Validators

	// PRE:
	// Validator has no balance at the beginning
	s.RequireBondBalance(address, zero)

	// EXEC:
	// Make a PoA change to add validator
	ChangeValidator(s, AddValidatorAction, address, validators, govtypesv1.StatusPassed)

	// POST:
	// Validator should have bonded tokens in bank
	s.RequireBondBalance(address, DefaultBondedTokens)

	fmt.Println("==== [V] Test_AddNewValidator")
}

func (s *IntegrationTestSuite) Test_AddValidatorWithUnboundedTokens() {
	fmt.Println("==== Test_AddValidatorWithUnboundedTokens")

	address := "evmos1vnaenttkyalgvjus34xxt8h8k0fpuuurdlukaq"
	validators := s.network.Validators

	// PRE:
	// Validator has no initial balance
	s.RequireBondBalance(address, zero)

	// EXEC:
	// Add validator through PoA change twice
	ChangeValidator(s, AddValidatorAction, address, validators, govtypesv1.StatusPassed)
	s.RequireBondBalance(address, DefaultBondedTokens)
	ChangeValidator(s, AddValidatorAction, address, validators, govtypesv1.StatusFailed)

	// POST:
	// Only one of the proposals should have passed, validator should have the default power tokens
	s.RequireBondBalance(address, DefaultBondedTokens)

	fmt.Println("==== [V] Test_AddValidatorWithUnboundedTokens")
}

func (s *IntegrationTestSuite) Test_AddValidatorWithBondedTokens() {
	fmt.Println("==== Test_AddValidatorWithBondedTokens")

	address := s.network.Validators[0].Address.String()
	validators := s.network.Validators

	// PRE:
	// Validator has no balance in bank and bonded balance in staking
	s.RequireBondBalance(address, zero)
	s.RequireValidator(address, &bondedStatus, &DefaultBondedTokens)

	// EXEC:
	// Add validator through PoA Change
	ChangeValidator(s, AddValidatorAction, address, validators, govtypesv1.StatusFailed)

	// POST:
	// Validator should not have extra balance in bank
	s.RequireBondBalance(address, zero)
	s.RequireValidator(address, &bondedStatus, &DefaultBondedTokens)

	fmt.Println("==== [V] Test_AddValidatorWithBondedTokens")
}

func (s *IntegrationTestSuite) Test_AddUnbondingValidator() {
	fmt.Println("==== Test_AddUnbondingValidator")

	validator := s.network.Validators[1]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no balance in bank
	s.RequireValidator(validatorAddress, &bondedStatus, &DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().Contains(validator)

	// EXEC:
	// Add validator from a poa change but don't wait to be finished
	ChangeValidator(s, AddValidatorAction, validatorAddress, s.network.Validators, govtypesv1.StatusNil)
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	UnBondTokens(s, validator, DefaultBondedTokens, true)

	// POST:
	// Validator should not have any tokens in staking and bonded
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, DefaultBondedTokens)
	s.RequireValidatorSet().NotContains(validator)

	fmt.Println("==== [V] Test_AddUnbondingValidator")
}

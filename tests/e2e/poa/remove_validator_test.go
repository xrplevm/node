package poa_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"time"
)

func (s *IntegrationTestSuite) Test_RemoveUnexistentValidator() {
	fmt.Println("==== Test_RemoveUnexistentValidator")

	address := "evmos16qkupjv69m6r8zl2frckc9vkmlz9ll7law8uea"
	validators := s.network.Validators[0:s.cfg.NumBondedValidators]

	// PRE:
	// Validator address has no balance
	s.RequireBondBalance(address, zero)

	// EXEC:
	// Add balance through POA change
	ChangeValidator(s, AddValidatorAction, address, validators, govtypesv1.StatusPassed)
	s.RequireBondBalance(address, DefaultBondedTokens)
	ChangeValidator(s, RemoveValidatorAction, address, validators, govtypesv1.StatusPassed)

	// POST:
	// Validator has no balance
	s.RequireBondBalance(address, zero)

	fmt.Println("==== [V] Test_RemoveUnexistentValidator")
}

func (s *IntegrationTestSuite) Test_RemoveValidatorWithoutBondedAndBankTokens() {
	fmt.Println("==== Test_RemoveValidatorWithoutBondedAndBankTokens")

	address := "evmos1jp26kvnhf940p544awlfd75jkx23z6pjyyvvkz"
	validators := s.network.Validators[0:s.cfg.NumBondedValidators]

	// PRE:
	// Validator has no balance
	s.RequireBondBalance(address, zero)

	// EXEC:
	// Remove validator that has no balance, proposal has failed
	ChangeValidator(s, RemoveValidatorAction, address, validators, govtypesv1.StatusFailed)

	// POST:
	// Nothing has happened, validator still has no balance
	s.RequireBondBalance(address, zero)

	fmt.Println("==== [V] Test_RemoveValidatorWithoutBondedAndBankTokens")
}

func (s *IntegrationTestSuite) Test_RemoveFullyBondedValidator() {
	fmt.Println("==== Test_RemoveFullyBondedValidator")

	validator := s.network.Validators[0]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no tokens in bank
	s.RequireValidator(validatorAddress, &bondedStatus, &DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().Contains(validator)

	// EXEC:
	// Remove validator through PoA change
	ChangeValidator(s, RemoveValidatorAction, validatorAddress, s.network.Validators, govtypesv1.StatusPassed)
	time.Sleep(s.cfg.UnBoundingTime)
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}

	// POST:
	// Validator is unbonded and has no tokens in bank
	s.RequireValidator(validatorAddress, &unbondedStatus, &zero)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().NotContains(validator)

	fmt.Println("==== [V] Test_RemoveFullyBondedValidator")
}

func (s *IntegrationTestSuite) Test_RemoveUnbondedValidator() {
	fmt.Println("==== Test_RemoveUnbondedValidator")

	validator := s.network.Validators[s.cfg.NumBondedValidators+0]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator does not exist but has balance in bank
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, DefaultBondedTokens)

	// EXEC:
	// Bond some tokens that are not enough for being bonded to make validator status being unbonded
	// and then remove validator through PoA
	halfTokens := sdk.NewDec(DefaultBondedTokens.Int64()).Quo(sdk.NewDec(2)).RoundInt()
	BondTokens(s, validator, halfTokens)
	s.RequireValidator(validatorAddress, &unbondedStatus, &halfTokens)
	s.RequireBondBalance(validatorAddress, halfTokens)
	s.RequireValidatorSet().NotContains(validator)

	ChangeValidator(s, RemoveValidatorAction, validatorAddress, s.network.Validators, govtypesv1.StatusPassed)

	// POST:
	// Validator should not have any tokens in staking and bonded
	s.RequireValidator(validatorAddress, &unbondedStatus, &zero)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().NotContains(validator)

	fmt.Println("==== [V] Test_RemoveUnbondedValidator")
}

func (s *IntegrationTestSuite) Test_RemoveUnbondingValidator() {
	fmt.Println("==== Test_RemoveUnbondingValidator")

	validator := s.network.Validators[1]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no balance in bank
	s.RequireValidator(validatorAddress, &bondedStatus, &DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().Contains(validator)

	// EXEC:
	// Remove validator from a pool but don't wait to be finished
	ChangeValidator(s, RemoveValidatorAction, validatorAddress, s.network.Validators, govtypesv1.StatusNil)
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	if err := s.network.WaitForNextBlock(); err != nil {
		panic(err)
	}
	UnBondTokens(s, validator, DefaultBondedTokens, true)

	// POST:
	// Validator should not have any tokens in staking and bonded
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().NotContains(validator)

	fmt.Println("==== [V] Test_RemoveUnbondingValidator")
}

// TODO: Remove validator with Bonded state with some staking tokens and some bank tokens

// TODO: Remove validator with Unbounded state without tokens (bank & staking)
// TODO: Remove validator with Unbounded state with staking tokens ?
// TODO: Remove validator with Unbounded state with bank tokens
// TODO: Remove validator with Unbounded state with some staking tokens and some bank tokens

// TODO: Remove validator with UnBounding state without tokens (bank & staking)
// TODO: Remove validator with UnBounding state with staking tokens
// TODO: Remove validator with UnBounding state with bank tokens
// TODO: Remove validator with UnBounding state with some staking tokens and some bank tokens

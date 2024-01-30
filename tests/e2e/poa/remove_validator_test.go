package poa_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"sync"
	"time"
)

func (s *IntegrationTestSuite) Test_RemoveNonexistentValidator() {
	fmt.Println("==== Test_RemoveUnexistentValidator")

	address, _ := sdk.AccAddressFromBech32("evmos16qkupjv69m6r8zl2frckc9vkmlz9ll7law8uea")
	pubKey := GenPubKey()
	validators := s.network.Validators

	// PRE:
	// Validator address has no balance
	s.RequireBondBalance(address.String(), zero)
	s.RequireValidatorSet().NotContains(pubKey)

	// EXEC:
	// Remove nonexistent validator
	ChangeValidator(s, RemoveValidatorAction, address, pubKey, validators, govtypesv1.StatusFailed)

	// POST:
	// Validator has no balance
	s.RequireBondBalance(address.String(), zero)
	s.RequireValidatorSet().NotContains(pubKey)

	fmt.Println("==== [V] Test_RemoveUnexistentValidator")
}

func (s *IntegrationTestSuite) Test_RemoveValidatorWithBankTokens() {
	fmt.Println("==== Test_RemoveValidatorWithoutBondedAndBankTokens")

	validator := s.network.Validators[s.cfg.NumBondedValidators]
	validators := s.network.Validators

	// PRE:
	// Validator has balance and is not in the validator set
	s.RequireBondBalance(validator.Address.String(), DefaultBondedTokens)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	// EXEC:
	// Remove validator that has balance, proposal passes
	ChangeValidator(s, RemoveValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusPassed)

	// POST:
	// Validator still has no balance and is not in the validator set
	s.RequireBondBalance(validator.Address.String(), zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

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
	s.RequireValidatorSet().Contains(validator.PubKey)

	// EXEC:
	// Remove validator through PoA change
	ChangeValidator(s, RemoveValidatorAction, validator.Address, validator.PubKey, s.network.Validators, govtypesv1.StatusPassed)
	time.Sleep(s.cfg.UnBoundingTime)
	s.network.MustWaitForNextBlock()

	// POST:
	// Validator is unbonded and has no tokens in bank
	s.RequireValidator(validatorAddress, &unbondedStatus, &zero)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

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
	s.RequireValidatorSet().NotContains(validator.PubKey)
	// Bond some tokens that are not enough for being bonded to make validator status being unbonded
	// and then remove validator through PoA
	halfTokens := sdk.NewDec(DefaultBondedTokens.Int64()).Quo(sdk.NewDec(2)).RoundInt()
	BondTokens(s, validator, halfTokens)
	s.RequireValidator(validatorAddress, &unbondedStatus, &halfTokens)
	s.RequireBondBalance(validatorAddress, halfTokens)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	// EXEC:
	ChangeValidator(s, RemoveValidatorAction, validator.Address, validator.PubKey, s.network.Validators, govtypesv1.StatusPassed)

	// POST:
	// Validator should not have any tokens in staking and bonded
	s.RequireValidator(validatorAddress, &unbondedStatus, &zero)
	s.RequireBondBalance(validatorAddress, zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

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
	s.RequireValidatorSet().Contains(validator.PubKey)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// EXEC:
		// Remove validator from a pool but don't wait to be finished
		ChangeValidator(s, RemoveValidatorAction, validator.Address, validator.PubKey, s.network.Validators, govtypesv1.StatusPassed)
		s.network.MustWaitForNextBlock()

		// POST:
		// Validator should not have any tokens in staking and bonded
		s.RequireValidator(validatorAddress, &unbondedStatus, &zero)
		s.RequireBondBalance(validatorAddress, zero)
		s.RequireValidatorSet().NotContains(validator.PubKey)
	}()
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	if err := validator.TmNode.Stop(); err != nil {
		fmt.Printf("Error stopping node: %v\n", err)
	}

	wg.Wait()

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

package poa_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"sync"
)

func (s *IntegrationTestSuite) Test_AddNewValidator() {
	fmt.Println("==== Test_AddNewValidator")

	address, _ := sdk.AccAddressFromBech32("evmos1ycvhcxthjju0466d4ga0j7du7wt8kmaep28zqv")
	validators := s.network.Validators
	pubKey := GenPubKey()

	// PRE:
	// Validator should not be in the validator set
	s.RequireValidatorSet().NotContains(pubKey)

	// EXEC:
	// Make a PoA change to add validator
	ChangeValidator(s, AddValidatorAction, address, pubKey, validators, govtypesv1.StatusPassed)
	// Wait enough to be sure that the validator is in the validator set
	s.network.MustWaitForNextBlock()
	s.network.MustWaitForNextBlock()

	// POST:
	// Validator should be in the validator set
	s.RequireValidatorSet().Contains(pubKey)

	fmt.Println("==== [V] Test_AddNewValidator")
}

func (s *IntegrationTestSuite) Test_AddValidatorWithUnboundedTokens() {
	fmt.Println("==== Test_AddValidatorWithUnboundedTokens")

	validators := s.network.Validators
	validator := s.network.Validators[s.cfg.NumBondedValidators]

	// PRE:
	// Validator has no initial balance
	s.RequireBondBalance(validator.Address.String(), DefaultBondedTokens)

	// EXEC:
	// Add validator through PoA change twice
	ChangeValidator(s, AddValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusFailed)

	// POST:
	// Only one of the proposals should have passed, validator should have the default power tokens
	s.RequireBondBalance(validator.Address.String(), DefaultBondedTokens)

	fmt.Println("==== [V] Test_AddValidatorWithUnboundedTokens")
}

func (s *IntegrationTestSuite) Test_AddValidatorWithBondedTokens() {
	fmt.Println("==== Test_AddValidatorWithBondedTokens")

	validator := s.network.Validators[0]
	address := validator.Address.String()
	validators := s.network.Validators

	// PRE:
	// Validator has no balance in bank and bonded balance in staking
	s.RequireBondBalance(address, zero)
	s.RequireValidator(address, &bondedStatus, &DefaultBondedTokens)

	// EXEC:
	// Add validator through PoA Change
	ChangeValidator(s, AddValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusFailed)

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
	s.RequireValidatorSet().Contains(validator.PubKey)

	// EXEC:
	// Add validator from a poa change but don't wait to be finished
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ChangeValidator(s, AddValidatorAction, validator.Address, validator.PubKey, s.network.Validators, govtypesv1.StatusFailed)

		// POST:
		// Validator should have tokens bonded in its validator but not in bank
		s.RequireValidator(validatorAddress, &unbondingStatus, &DefaultBondedTokens)
		s.RequireBondBalance(validatorAddress, zero)
	}()
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	if err := validator.TmNode.Stop(); err != nil {
		fmt.Printf("Error stopping node: %v\n", err)
	}

	wg.Wait()

	fmt.Println("==== [V] Test_AddUnbondingValidator")
}

package unbonding_test

import (
	"github.com/Peersyst/exrp/tests/e2e"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"sync"
	"time"
)

func (s *TestSuite) Test_AddUnbondingValidator() {
	s.T().Logf("==== Test_AddUnbondingValidator")
	validator := s.Network.Validators[1]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no balance in bank
	s.RequireValidator(validatorAddress, &e2e.BondedStatus, &e2e.DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	s.RequireValidatorSet().Contains(validator.PubKey)

	// EXEC:
	// Add validator from a poa change but don't wait to be finished
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.AddValidatorAction, validator.Address, validator.PubKey, s.Network.Validators, govtypesv1.StatusFailed)

		// POST:
		// Validator should have tokens bonded in its validator but not in bank
		s.RequireValidator(validatorAddress, &e2e.UnbondingStatus, &e2e.DefaultBondedTokens)
		s.RequireBondBalance(validatorAddress, e2e.Zero)
	}()
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	err := validator.TmNode.Stop()
	s.Require().NoError(err)

	wg.Wait()

	s.T().Logf("==== [V] Test_AddUnbondingValidator")
}

func (s *TestSuite) Test_RemoveUnbondingValidator() {
	s.T().Logf("==== Test_RemoveUnbondingValidator")

	validator := s.Network.Validators[1]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no balance in bank
	s.RequireValidator(validatorAddress, &e2e.BondedStatus, &e2e.DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	s.RequireValidatorSet().Contains(validator.PubKey)

	var wg sync.WaitGroup
	wg.Add(1)

	// Height 10    -> Proposal is broadcasted
	// Height 10    -> Validator is disconnected
	// Height 11    -> Proposal is submitted
	// Height 13    -> Validator is slashed unsigned h11 & h12 -> status updated to unbonding
	// Height 15-16 -> Proposal is executed
	// Height 18    -> Validator becames unbonded

	go func() {
		defer wg.Done()
		s.Network.WaitForHeightWithTimeout(10, 2*time.Minute)
		// EXEC:
		// Remove validator from a pool but don't wait to be finished
		e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.RemoveValidatorAction, validator.Address, validator.PubKey, s.Network.Validators, govtypesv1.StatusPassed)

		// POST:
		// Validator should not have any tokens in staking and bonded
		s.RequireValidator(validatorAddress, &e2e.UnbondingStatus, &e2e.Zero)
		s.RequireBondBalance(validatorAddress, e2e.Zero)
		s.RequireValidatorSet().NotContains(validator.PubKey)
	}()
	// Execute unbond tokens so at the moment of the proposal execution the status is unbonding
	s.Network.WaitForHeightWithTimeout(10, 2*time.Minute)
	err := validator.TmNode.Stop()
	s.Require().NoError(err)

	wg.Wait()

	s.T().Logf("==== [V] Test_RemoveUnbondingValidator")
}

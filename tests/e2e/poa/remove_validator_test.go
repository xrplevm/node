package poa_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/xrplevm/node/v3/tests/e2e"
)

func (s *TestSuite) Test_RemoveNonexistentValidator() {
	s.T().Logf("==== Test_RemoveUnexistentValidator")

	address, _ := sdk.AccAddressFromBech32("evmos16qkupjv69m6r8zl2frckc9vkmlz9ll7law8uea")
	pubKey := e2e.GenPubKey()
	validators := s.Network.Validators

	// PRE:
	// Validator address has no balance
	s.RequireBondBalance(address.String(), e2e.Zero)
	s.RequireValidatorSet().NotContains(pubKey)

	// EXEC:
	// Remove nonexistent validator
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.RemoveValidatorAction, address, pubKey, validators, govtypesv1.StatusFailed)

	// POST:
	// Validator has no balance
	s.RequireBondBalance(address.String(), e2e.Zero)
	s.RequireValidatorSet().NotContains(pubKey)

	s.T().Logf("==== [V] Test_RemoveUnexistentValidator")
}

func (s *TestSuite) Test_RemoveValidatorWithBankTokens() {
	s.T().Logf("==== Test_RemoveValidatorWithoutBondedAndBankTokens")

	validator := s.Network.Validators[s.Cfg.NumBondedValidators]
	validators := s.Network.Validators

	// PRE:
	// Validator has balance and is not in the validator set
	s.RequireBondBalance(validator.Address.String(), e2e.DefaultBondedTokens)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	// EXEC:
	// Remove validator that has balance, proposal passes
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.RemoveValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusPassed)

	// POST:
	// Validator still has no balance and is not in the validator set
	s.RequireBondBalance(validator.Address.String(), e2e.Zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	s.T().Logf("==== [V] Test_RemoveValidatorWithoutBondedAndBankTokens")
}

func (s *TestSuite) Test_RemoveFullyBondedValidator() {
	s.T().Logf("==== Test_RemoveFullyBondedValidator")

	validator := s.Network.Validators[0]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator is bonded and has no tokens in bank
	s.RequireValidator(validatorAddress, &e2e.BondedStatus, &e2e.DefaultBondedTokens)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	s.RequireValidatorSet().Contains(validator.PubKey)

	// EXEC:
	// Remove validator through PoA change
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.RemoveValidatorAction, validator.Address, validator.PubKey, s.Network.Validators, govtypesv1.StatusPassed)
	time.Sleep(s.Cfg.UnBoundingTime)
	s.Network.MustWaitForNextBlock()

	// POST:
	// Validator is unbonded and has no tokens in bank
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	s.T().Logf("==== [V] Test_RemoveFullyBondedValidator")
}

func (s *TestSuite) Test_RemoveUnbondedValidator() {
	s.T().Logf("==== Test_RemoveUnbondedValidator")

	validator := s.Network.Validators[s.Cfg.NumBondedValidators+0]
	validatorAddress := validator.Address.String()

	// PRE:
	// Validator does not exist but has balance in bank
	s.RequireValidator(validatorAddress, nil, nil)
	s.RequireBondBalance(validatorAddress, e2e.DefaultBondedTokens)
	s.RequireValidatorSet().NotContains(validator.PubKey)
	// Bond some tokens that are not enough for being bonded to make validator status being unbonded
	// and then remove validator through PoA
	halfTokens := sdk.NewDec(e2e.DefaultBondedTokens.Int64()).Quo(sdk.NewDec(2)).RoundInt()
	e2e.BondTokens(&s.IntegrationTestSuite, validator, halfTokens)
	s.RequireValidator(validatorAddress, &e2e.UnbondedStatus, &halfTokens)
	s.RequireBondBalance(validatorAddress, halfTokens)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	// EXEC:
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.RemoveValidatorAction, validator.Address, validator.PubKey, s.Network.Validators, govtypesv1.StatusPassed)

	// POST:
	// Validator should not have any tokens in staking and bonded
	s.RequireValidator(validatorAddress, &e2e.UnbondedStatus, &e2e.Zero)
	s.RequireBondBalance(validatorAddress, e2e.Zero)
	s.RequireValidatorSet().NotContains(validator.PubKey)

	s.T().Logf("==== [V] Test_RemoveUnbondedValidator")
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

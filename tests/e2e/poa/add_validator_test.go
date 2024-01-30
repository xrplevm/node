package poa_test

import (
	"github.com/Peersyst/exrp/tests/e2e"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (s *TestSuite) Test_AddNewValidator() {
	s.T().Logf("==== Test_AddNewValidator")

	address, _ := sdk.AccAddressFromBech32("evmos1ycvhcxthjju0466d4ga0j7du7wt8kmaep28zqv")
	validators := s.Network.Validators
	pubKey := e2e.GenPubKey()

	// PRE:
	// Validator should not be in the validator set
	s.RequireValidatorSet().NotContains(pubKey)

	// EXEC:
	// Make a PoA change to add validator
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.AddValidatorAction, address, pubKey, validators, govtypesv1.StatusPassed)
	// Wait enough to be sure that the validator is in the validator set
	s.Network.MustWaitForNextBlock()

	// POST:
	// Validator should be in the validator set
	s.RequireValidatorSet().Contains(pubKey)

	s.T().Logf("==== [V] Test_AddNewValidator")
}

func (s *TestSuite) Test_AddValidatorWithUnboundedTokens() {
	s.T().Logf("==== Test_AddValidatorWithUnboundedTokens")

	validators := s.Network.Validators
	validator := s.Network.Validators[s.Cfg.NumBondedValidators]

	// PRE:
	// Validator has no initial balance
	s.RequireBondBalance(validator.Address.String(), e2e.DefaultBondedTokens)

	// EXEC:
	// Add validator through PoA change twice
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.AddValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusFailed)

	// POST:
	// Only one of the proposals should have passed, validator should have the default power tokens
	s.RequireBondBalance(validator.Address.String(), e2e.DefaultBondedTokens)

	s.T().Logf("==== [V] Test_AddValidatorWithUnboundedTokens")
}

func (s *TestSuite) Test_AddValidatorWithBondedTokens() {
	s.T().Logf("==== Test_AddValidatorWithBondedTokens")

	validator := s.Network.Validators[0]
	address := validator.Address.String()
	validators := s.Network.Validators

	// PRE:
	// Validator has no balance in bank and bonded balance in staking
	s.RequireBondBalance(address, e2e.Zero)
	s.RequireValidator(address, &e2e.BondedStatus, &e2e.DefaultBondedTokens)

	// EXEC:
	// Add validator through PoA Change
	e2e.ChangeValidator(&s.IntegrationTestSuite, e2e.AddValidatorAction, validator.Address, validator.PubKey, validators, govtypesv1.StatusFailed)

	// POST:
	// Validator should not have extra balance in bank
	s.RequireBondBalance(address, e2e.Zero)
	s.RequireValidator(address, &e2e.BondedStatus, &e2e.DefaultBondedTokens)

	s.T().Logf("==== [V] Test_AddValidatorWithBondedTokens")
}

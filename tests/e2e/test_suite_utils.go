package e2e

import (
	"cosmossdk.io/math"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	BondedStatus        = stakingtypes.Bonded
	UnbondedStatus      = stakingtypes.Unbonded
	UnbondingStatus     = stakingtypes.Unbonding
	Zero                = math.ZeroInt()
	DefaultBondedTokens = sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
)

func (s *IntegrationTestSuite) ConsumeProposalCount() string {
	s.ProposalCount++
	return strconv.Itoa(s.ProposalCount)
}

func (s *IntegrationTestSuite) GetCtx() client.Context {
	return s.Network.Validators[0].ClientCtx
}

//nolint:staticcheck
func (s *IntegrationTestSuite) RequireValidator(address string, status *stakingtypes.BondStatus, tokens *math.Int) {
	accAddr, _ := sdk.AccAddressFromBech32(address)
	validatorInfo := GetValidator(s.GetCtx(), sdk.ValAddress(accAddr).String())
	if validatorInfo == nil {
		s.Require().True(status == nil)
		s.Require().True(tokens == nil)
	} else {
		s.Require().Equal(*status, validatorInfo.Status)
		s.Require().Equal(*tokens, validatorInfo.Tokens)
	}
}

func (s *IntegrationTestSuite) RequireDelegation(valAddress string, delAddress string, shares math.LegacyDec) {
	accAddr, _ := sdk.AccAddressFromBech32(valAddress)
	valAddr := sdk.ValAddress(accAddr).String()
	delegation := GetDelegation(s.GetCtx(), valAddr, delAddress)
	if delegation == nil {
		s.Require().Equal(math.LegacyZeroDec(), shares)
	} else {
		s.Require().Equal(delegation.Shares, shares)
	}
}

//nolint:staticcheck
func (s *IntegrationTestSuite) RequireBondBalance(address string, balance math.Int) {
	originalBalance := GetBalance(s.GetCtx(), address, s.Cfg.BondDenom)
	expected := sdk.NewCoin(s.Cfg.BondDenom, balance)
	s.Require().True(originalBalance.Equal(expected))
}

func (s *IntegrationTestSuite) RequireValidatorSet() struct {
	Contains    func(validator cryptotypes.PubKey)
	NotContains func(validator cryptotypes.PubKey)
} {
	validatorSet, err := s.Handler.GetValidatorSet()
	if err != nil {
		s.T().Fatal(err)
	}
	validatorAddresses := make([]string, 0)
	for _, val := range validatorSet.Validators {
		validatorAddresses = append(validatorAddresses, val.Address)
	}
	return struct {
		Contains    func(pubKey cryptotypes.PubKey)
		NotContains func(pubKey cryptotypes.PubKey)
	}{
		Contains: func(pubKey cryptotypes.PubKey) {
			s.Require().Contains(validatorAddresses, sdk.ConsAddress(pubKey.Address()).String())
		},
		NotContains: func(pubKey cryptotypes.PubKey) {
			s.Require().NotContains(validatorAddresses, sdk.ConsAddress(pubKey.Address()).String())
		},
	}
}

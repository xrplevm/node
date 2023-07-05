package poa

import (
	"github.com/Peersyst/exrp/testutil/network"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"strconv"
)

var (
	bondedStatus        = stakingtypes.Bonded
	unbondedStatus      = stakingtypes.Unbonded
	zero                = sdk.ZeroInt()
	DefaultBondedTokens = sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)
)

type TestSuite struct {
	suite.Suite

	proposalCount int

	cfg     network.Config
	network *network.Network
}

func (s *TestSuite) ConsumeProposalCount() string {
	s.proposalCount = s.proposalCount + 1
	return strconv.Itoa(s.proposalCount)
}

func (s *TestSuite) GetCtx() client.Context {
	return s.network.Validators[0].ClientCtx
}

func (s *TestSuite) RequireValidator(address string, status *stakingtypes.BondStatus, tokens *sdk.Int) {
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

func (s *TestSuite) RequireDelegation(valAddress string, delAddress string, shares sdk.Dec) {
	accAddr, _ := sdk.AccAddressFromBech32(valAddress)
	valAddr := sdk.ValAddress(accAddr).String()
	delegation := GetDelegation(s.GetCtx(), valAddr, delAddress)
	if delegation == nil {
		s.Require().Equal(sdk.ZeroDec(), shares)
	} else {
		s.Require().Equal(delegation.Shares, shares)
	}
}

func (s *TestSuite) RequireBondBalance(address string, balance sdk.Int) {
	originalBalance := GetBalance(s.GetCtx(), address, s.cfg.BondDenom)
	expected := sdk.NewCoin(s.cfg.BondDenom, balance)
	s.Require().True(originalBalance.Equal(expected))
}

func (s *TestSuite) RequireValidatorSet() struct {
	Contains    func(validator *network.Validator)
	NotContains func(validator *network.Validator)
} {
	validatorSet := GetValidatorSet(s.GetCtx())
	validatorAddresses := make([]string, 0)
	for _, val := range validatorSet.Validators {
		validatorAddresses = append(validatorAddresses, val.Address)
	}
	return struct {
		Contains    func(validator *network.Validator)
		NotContains func(validator *network.Validator)
	}{
		Contains: func(validator *network.Validator) {
			s.Require().Contains(validatorAddresses, sdk.ConsAddress(validator.PubKey.Address()).String())
		},
		NotContains: func(validator *network.Validator) {
			s.Require().NotContains(validatorAddresses, sdk.ConsAddress(validator.PubKey.Address()).String())
		},
	}
}

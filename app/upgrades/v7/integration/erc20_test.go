//nolint:dupl
package integration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

func (s *UpgradeTestSuite) TestUpgrade_ERC20Params() {
	prevParams, err := s.network.GetERC20Client().Params(
		s.network.GetContext(),
		&erc20types.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetERC20Client().Params(
		s.network.GetContext(),
		&erc20types.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_ERC20_TokenPairs() {
	prevTokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postTokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified token pairs are the same
	s.Require().Equal(prevTokenPairs.TokenPairs, postTokenPairs.TokenPairs)
}

func (s *UpgradeTestSuite) TestUpgrade_ERC20_MintCoins() {
	tokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)
	s.Require().Equal(len(tokenPairs.TokenPairs), 1)

	tokenPair := tokenPairs.TokenPairs[0]

	sender, err := sdktypes.AccAddressFromBech32(tokenPair.OwnerAddress)
	s.Require().NoError(err)

	receiver, err := sdktypes.AccAddressFromBech32("ethm1dakgyqjulg29m5fmv992g2y66m9g2mjn6hahwg")
	s.Require().NoError(err)

	amount := sdktypes.NewInt64Coin(s.network.GetDenom(), 100)

	prevBalancesReceiver, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: receiver.String(),
			Denom:   s.network.GetDenom(),
		},
	)

	s.RunUpgrade(upgradeName)

	err = s.network.ERC20Keeper().MintCoins(
		s.network.GetContext(),
		sender,
		receiver,
		amount.Amount,
		s.network.GetDenom(),
	)
	s.Require().NoError(err)

	postBalancesReceiver, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: receiver.String(),
			Denom:   s.network.GetDenom(),
		},
	)
	s.Require().NoError(err)

	s.Require().Equal(prevBalancesReceiver.Balance.Amount.Add(amount.Amount).String(), postBalancesReceiver.Balance.Amount.String())
}

func (s *UpgradeTestSuite) TestUpgrade_ERC20_BurnCoins() {
	sender, err := sdktypes.AccAddressFromBech32("ethm1dakgyqjulg29m5fmv992g2y66m9g2mjn6hahwg")
	s.Require().NoError(err)

	amount := sdktypes.NewInt64Coin(s.network.GetDenom(), 100)

	prevBalancesSender, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: sender.String(),
			Denom:   s.network.GetDenom(),
		},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	err = s.network.ERC20Keeper().BurnCoins(
		s.network.GetContext(),
		sender,
		amount.Amount,
		s.network.GetDenom(),
	)
	s.Require().NoError(err)

	postBalancesSender, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: sender.String(),
			Denom:   s.network.GetDenom(),
		},
	)
	s.Require().NoError(err)

	s.Require().Equal(prevBalancesSender.Balance.Amount.Sub(amount.Amount).String(), postBalancesSender.Balance.Amount.String())
}

func (s *UpgradeTestSuite) TestUpgrade_ERC20_TransferOwnership() {
	tokenPairs, err := s.network.GetERC20Client().TokenPairs(
		s.network.GetContext(),
		&erc20types.QueryTokenPairsRequest{},
	)
	s.Require().NoError(err)
	s.Require().Equal(len(tokenPairs.TokenPairs), 1)

	tokenPair := tokenPairs.TokenPairs[0]

	sender, err := sdktypes.AccAddressFromBech32(tokenPair.OwnerAddress)
	s.Require().NoError(err)

	newOwner, err := sdktypes.AccAddressFromBech32("ethm1nqvn2hmte72e3z0xyqmh06hdwd9qu6hgdcavhh")
	s.Require().NoError(err)

	s.network.ERC20Keeper().TransferOwnership(
		s.network.GetContext(),
		sender,
		newOwner,
		tokenPair.Denom,
	)
	s.Require().NoError(err)

	postTokenPair, err := s.network.GetERC20Client().TokenPair(
		s.network.GetContext(),
		&erc20types.QueryTokenPairRequest{
			Token: tokenPair.Denom,
		},
	)
	s.Require().NoError(err)

	s.Require().Equal(newOwner.String(), postTokenPair.TokenPair.OwnerAddress)
}
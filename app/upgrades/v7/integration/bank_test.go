//nolint:dupl
package integration

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (s *UpgradeTestSuite) TestUpgrade_Bank_Params() {
	prevParams, err := s.network.GetBankClient().Params(
		s.network.GetContext(),
		&banktypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postParams, err := s.network.GetBankClient().Params(
		s.network.GetContext(),
		&banktypes.QueryParamsRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified params are the same
	s.Require().Equal(prevParams.Params, postParams.Params)
}

func (s *UpgradeTestSuite) TestUpgrade_Bank_TotalSupply() {
	res, err := s.network.GetBankClient().TotalSupply(
		s.network.GetContext(),
		&banktypes.QueryTotalSupplyRequest{},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	postRes, err := s.network.GetBankClient().TotalSupply(
		s.network.GetContext(),
		&banktypes.QueryTotalSupplyRequest{},
	)
	s.Require().NoError(err)

	// Check that not modified balances are the same
	s.Require().Equal(res.Supply, postRes.Supply)
}

func (s *UpgradeTestSuite) TestUpgrade_Bank_Send() {
	// Replace with the desired addresses
	sender, err := sdktypes.AccAddressFromBech32("ethm1dakgyqjulg29m5fmv992g2y66m9g2mjn6hahwg")
	s.Require().NoError(err)
	receiver, err := sdktypes.AccAddressFromBech32("ethm1nqvn2hmte72e3z0xyqmh06hdwd9qu6hgdcavhh")
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

	prevBalancesReceiver, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: receiver.String(),
			Denom:   s.network.GetDenom(),
		},
	)
	s.Require().NoError(err)

	s.RunUpgrade(upgradeName)

	err = s.network.BankKeeper().SendCoins(
		s.network.GetContext(),
		sender,
		receiver,
		sdktypes.NewCoins(amount),
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

	postBalancesReceiver, err := s.network.GetBankClient().Balance(
		s.network.GetContext(),
		&banktypes.QueryBalanceRequest{
			Address: receiver.String(),
			Denom:   s.network.GetDenom(),
		},
	)
	s.Require().NoError(err)

	s.Require().Equal(prevBalancesSender.Balance.Amount.Sub(amount.Amount).String(), postBalancesSender.Balance.Amount.String())
	s.Require().Equal(prevBalancesReceiver.Balance.Amount.Add(amount.Amount).String(), postBalancesReceiver.Balance.Amount.String())
}

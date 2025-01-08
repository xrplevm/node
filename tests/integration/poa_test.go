package integration

import (
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	poatypes "github.com/xrplevm/node/v5/x/poa/types"
)

func (s *TestSuite) TestRemoveValidator() {
	// Generate a random account
	randomAccs := simtypes.RandomAccounts(rand.New(rand.NewSource(time.Now().UnixNano())), 1)
	randomAcc := randomAccs[0]

	// Validators
	validators := s.Network().GetValidators()
	require.NotZero(s.T(), len(validators))

	valAddr, err := sdktypes.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(s.T(), err)
	valAccAddr := sdktypes.AccAddress(valAddr)

	tt := []struct {
		name string
		valAddress string
		expectedError error
		beforeRun func()
		afterRun func()
	}{
		{
			name: "remove unexisting validator - random address - no balance",
			valAddress: randomAcc.Address.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom: s.Network().GetDenom(),
					},
				)
				require.NoError(s.T(), err)
				require.Equal(s.T(), balance.Balance.Amount, sdkmath.NewInt(0))
			},
		},
		{
			name: "remove unexisting validator - random address - with balance",
			valAddress: randomAcc.Address.String(),
			expectedError: poatypes.ErrAddressHasNoTokens,
			beforeRun: func() {
				coins := sdktypes.NewCoin(
					s.Network().GetDenom(),
					sdkmath.NewInt(100000000000),
				)

				err = s.factory.FundAccount(
					s.keyring.GetKey(0),
					randomAcc.Address,
					sdktypes.NewCoins(
						coins,
					),
				)
				require.NoError(s.T(), err)
				require.NoError(s.T(), s.Network().NextBlock())

				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom: s.Network().GetDenom(),
					},
				)
				require.NoError(s.T(), err)
				require.Equal(s.T(), coins.Amount.BigInt(), balance.Balance.Amount.BigInt())
			},
			afterRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom: s.Network().GetDenom(),
					},
				)
				require.NoError(s.T(), err)
				require.Equal(s.T(), balance.Balance.Amount, sdkmath.NewInt(0))
			},
		},
		{
			name: "remove existing validator - with unbonding delegations - with tokens", 
			valAddress: valAccAddr.String(),
			afterRun: func() {
				balance, err := s.Network().GetBankClient().Balance(
					s.Network().GetContext(),
					&banktypes.QueryBalanceRequest{
						Address: randomAcc.Address.String(),
						Denom: s.Network().GetDenom(),
					},
				)
				require.NoError(s.T(), err)
				require.Equal(s.T(), balance.Balance.Amount, sdkmath.NewInt(0))	
			},
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			if tc.beforeRun != nil {
				tc.beforeRun()
			}

			err = s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.afterRun != nil {
				tc.afterRun()
			}

			if tc.expectedError != nil && err != nil {
				require.Error(s.T(), err)
				require.ErrorIs(s.T(), err, tc.expectedError)
			} else {
				require.NoError(s.T(), err)
			}
		})
	}
}


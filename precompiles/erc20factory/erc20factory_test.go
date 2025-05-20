package erc20factory_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/xrplevm/node/v6/precompiles/erc20factory"
)

func (s *PrecompileTestSuite) TestIsTransaction() {
	s.SetupTest()

	// Constants
	s.Require().Equal(s.precompile.Address().String(), erc20factory.Erc20FactoryAddress)

	// Queries
	s.Require().False(s.precompile.IsTransaction(erc20factory.CalculateAddressMethod))

	// Transactions
	s.Require().True(s.precompile.IsTransaction(erc20factory.CreateMethod))
}

func (s *PrecompileTestSuite) TestRequiredGas() {
	s.SetupTest()
	salt := common.HexToHash("0x4f5b6f778b28c4d67a9c12345678901234567890123456789012345678901234")

	testcases := []struct {
		name     string
		malleate func() []byte
		expGas   uint64
	}{
		{
			name: erc20factory.CreateMethod,
			malleate: func() []byte {
				bz, err := s.precompile.ABI.Pack(erc20factory.CreateMethod, uint8(0), salt, "Test token", "TT", uint8(3))
				s.Require().NoError(err, "expected no error packing ABI")
				return bz
			},
			expGas: erc20factory.GasCreate,
		},
		{
			name: erc20factory.CalculateAddressMethod,
			malleate: func() []byte {
				bz, err := s.precompile.ABI.Pack(erc20factory.CalculateAddressMethod, uint8(0), salt)
				s.Require().NoError(err, "expected no error packing ABI")
				return bz
			},
			expGas: erc20factory.GasCalculateAddress,
		},
		{
			name: "invalid method",
			malleate: func() []byte {
				return []byte("invalid method")
			},
			expGas: 0,
		},
		{
			name: "input bytes too short",
			malleate: func() []byte {
				return []byte{0x00, 0x00, 0x00}
			},
			expGas: 0,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			input := tc.malleate()

			s.Require().Equal(tc.expGas, s.precompile.RequiredGas(input))
		})
	}
}

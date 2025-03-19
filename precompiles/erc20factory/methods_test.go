package erc20factory_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
	"github.com/xrplevm/node/v6/precompiles/erc20factory"
)

func (s *PrecompileTestSuite) TestCreate() {
	method := s.precompile.Methods[erc20factory.CreateMethod]
	// fromAddr is the address of the keyring account used for testing.
	fromAddr := s.keyring.GetKey(0).Addr
	salt := [32]uint8(common.HexToHash("0x4f5b6f778b28c4d67a9c12345678901234567890123456789012345678901234").Bytes())
	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(ctx sdk.Context, output []byte)
		expErr      bool
		errContains string
	}{
		{
			"success - create token",
			func() []interface{} {

				return []interface{}{uint8(0), salt, "AAA", "aaa", uint8(3)}
			},
			func(ctx sdk.Context, output []byte) {
				res, err := method.Outputs.Unpack(output)
				s.Require().NoError(err, "expected no error unpacking output")
				s.Require().Len(res, 1, "expected one output")
				address, ok := res[0].(common.Address)
				s.Require().True(ok, "expected address type")

				tokenId := s.network.App.Erc20Keeper.GetTokenPairID(ctx, address.String())
				tokenPair, found := s.network.App.Erc20Keeper.GetTokenPair(ctx, tokenId)
				s.Require().True(found, "expected no error getting token pair")
				s.Require().Equal(uint32(0), tokenPair.TokenType, "expected TokenType to match")
				s.Require().Equal(fmt.Sprintf("erc20/%s", address.String()), tokenPair.Denom, "expected token Denom to match")
				s.Require().Equal("", tokenPair.OwnerAddress, "expected OwnerAddress to match")
				s.Require().Equal(true, tokenPair.Enabled, "expected Enabled to match")
				s.Require().Equal(address.String(), tokenPair.Erc20Address, "expected Erc20Address to match")
				s.Require().Equal(erc20types.OWNER_EXTERNAL, tokenPair.ContractOwner, "expected ContractOwner to match")
			},
			false,
			"",
		},
		{
			"fail - invalid tokenType",
			func() []interface{} {
				return []interface{}{"", salt, "AAA", "aaa", uint8(0)}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid tokenType",
		},
		{
			"fail - invalid salt",
			func() []interface{} {
				return []interface{}{uint8(0), "", "AAA", "aaa", uint8(0)}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid salt",
		},
		{
			"fail - invalid name (bad type)",
			func() []interface{} {
				return []interface{}{uint8(0), salt, 10, "aaa", uint8(0)}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid name",
		},
		{
			"fail - invalid name (too short)",
			func() []interface{} {
				return []interface{}{uint8(0), salt, "AA", "aaa", uint8(0)}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid name",
		},
		{
			"fail - invalid name (too long)",
			func() []interface{} {
				return []interface{}{
					uint8(0),
					salt,
					"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
					"aaa",
					uint8(0),
				}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid name",
		},
		{
			"fail - invalid symbol (bad type)",
			func() []interface{} {
				return []interface{}{
					uint8(0),
					salt,
					"AAA",
					0,
					uint8(0),
				}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid symbol",
		},
		{
			"fail - invalid symbol (too short)",
			func() []interface{} {
				return []interface{}{
					uint8(0),
					salt,
					"AAA",
					"AA",
					uint8(0),
				}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid symbol",
		},
		{
			"fail - invalid symbol (too long)",
			func() []interface{} {
				return []interface{}{
					uint8(0),
					salt,
					"AAA",
					"AAAAAAAAAAAAAAAAA",
					uint8(0),
				}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid symbol",
		},
		{
			"fail - invalid decimals",
			func() []interface{} {
				return []interface{}{
					uint8(0),
					salt,
					"AAA",
					"AAA",
					uint32(0),
				}
			},
			func(_ sdk.Context, _ []byte) {},
			true,
			"invalid decimals",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			ctx := s.network.GetContext()
			output, err := s.precompile.Create(ctx, &method, fromAddr, tc.malleate())
			if tc.expErr {
				s.Require().Error(err, "expected create to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected create to fail with specific error")
			} else {
				s.Require().NoError(err, "expected create succeeded")
				tc.postCheck(ctx, output)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestCalculateAddress() {
	method := s.precompile.Methods[erc20factory.CalculateAddressMethod]
	// fromAddr is the address of the keyring account used for testing.
	fromAddr := common.HexToAddress("0xDc411BaFB148ebDA2B63EBD5f3D8669DD4383Af5")
	salt := [32]uint8(common.HexToHash("0x4f5b6f778b28c4d67a9c12345678901234567890123456789012345678901234").Bytes())
	testcases := []struct {
		name        string
		malleate    func() []interface{}
		postCheck   func(output []byte)
		expErr      bool
		errContains string
	}{
		{
			"success - calculate address",
			func() []interface{} {
				return []interface{}{uint8(0), salt}
			},
			func(output []byte) {
				res, err := method.Outputs.Unpack(output)
				s.Require().NoError(err, "expected no error unpacking output")
				s.Require().Len(res, 1, "expected one output")
				address, ok := res[0].(common.Address)
				s.Require().True(ok, "expected address type")
				s.Require().Equal(address.String(), "0x188a919f3583f8e02183332E6c73E944E002C553", "expected address to match")
			},
			false,
			"",
		},
		{
			"fail - invalid tokenType",
			func() []interface{} {
				return []interface{}{"", salt}
			},
			func(_ []byte) {},
			true,
			"invalid tokenType",
		},
		{
			"fail - invalid salt",
			func() []interface{} {
				return []interface{}{uint8(0), ""}
			},
			func(_ []byte) {},
			true,
			"invalid salt",
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			output, err := s.precompile.CalculateAddress(&method, fromAddr, tc.malleate())
			if tc.expErr {
				s.Require().Error(err, "expected create to fail")
				s.Require().Contains(err.Error(), tc.errContains, "expected create to fail with specific error")
			} else {
				s.Require().NoError(err, "expected create succeeded")
				tc.postCheck(output)
			}
		})
	}
}

func (s *PrecompileTestSuite) TestAddressCalculateMatch() {
	calculateAddressMethod := s.precompile.Methods[erc20factory.CalculateAddressMethod]
	createMethod := s.precompile.Methods[erc20factory.CreateMethod]
	fromAddr := common.HexToAddress("0xDc411BaFB148ebDA2B63EBD5f3D8669DD4383Af5")
	salt := [32]uint8(common.HexToHash("0x4f5b6f778b28c4d67a9c12345678901234567890123456789012345678901234").Bytes())
	s.Run("calculated address match created", func() {
		s.SetupTest()
		calculateAddressOutput, err := s.precompile.CalculateAddress(&calculateAddressMethod, fromAddr, []interface{}{uint8(0), salt})
		s.Require().NoError(err, "expected no error calculating address")
		calculateAddressRes, err := calculateAddressMethod.Outputs.Unpack(calculateAddressOutput)
		s.Require().NoError(err, "expected no error unpacking output")
		calculatedAddress, ok := calculateAddressRes[0].(common.Address)
		s.Require().True(ok, "expected address type")

		createOutput, err := s.precompile.Create(s.network.GetContext(), &createMethod, fromAddr, []interface{}{uint8(0), salt, "AAA", "aaa", uint8(0)})
		s.Require().NoError(err, "expected no error creating token")
		createRes, err := createMethod.Outputs.Unpack(createOutput)
		s.Require().NoError(err, "expected no error unpacking output")
		createdAddress, ok := createRes[0].(common.Address)
		s.Require().True(ok, "expected address type")

		s.Require().Equal(calculatedAddress.String(), createdAddress.String(), "expected calculated address to match created address")
	})
}

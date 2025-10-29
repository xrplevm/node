// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package erc20factory

import (
	"cosmossdk.io/errors"
	"encoding/binary"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	erc20types "github.com/evmos/evmos/v20/x/erc20/types"
)

const (
	// CreateMethod defines the ABI method name to create a new ERC20 Token Pair
	CreateMethod           = "create"
	CalculateAddressMethod = "calculateAddress"
)

// Create CreateERC20Precompile creates a new ERC20 TokenPair
func (p Precompile) Create(
	ctx sdk.Context,
	method *abi.Method,
	caller common.Address,
	args []interface{},
) ([]byte, error) {
	if len(args) != 5 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	tokenType, ok := args[0].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid tokenType")
	}

	salt, ok := args[1].([32]uint8)
	if !ok {
		return nil, fmt.Errorf("invalid salt")
	}

	name, ok := args[2].(string)
	if !ok || len(name) < 3 || len(name) > 128 {
		return nil, fmt.Errorf("invalid name")
	}

	symbol, ok := args[3].(string)
	if !ok || len(symbol) < 3 || len(symbol) > 16 {
		return nil, fmt.Errorf("invalid symbol")
	}

	decimals, ok := args[4].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid decimals")
	}

	address := crypto.CreateAddress2(caller, salt, calculateCodeHash(tokenType))

	metadata, err := p.CreateCoinMetadata(ctx, address, name, symbol, decimals)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to create wrapped coin denom metadata for ERC20",
		)
	}

	if err := metadata.Validate(); err != nil {
		return nil, errors.Wrapf(
			err, "ERC20 token data is invalid for contract %s", address.String(),
		)
	}

	p.bankKeeper.SetDenomMetaData(ctx, *metadata)

	pair := erc20types.NewTokenPair(address, metadata.Name, erc20types.OWNER_EXTERNAL)
	pair.TokenType = uint32(tokenType)

	p.erc20Keeper.SetToken(ctx, pair)

	err = p.erc20Keeper.EnableDynamicPrecompiles(ctx, pair.GetERC20Contract())
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(address)
}

// CalculateAddress calculates the address of a new ERC20 Token Pair
func (p Precompile) CalculateAddress(
	method *abi.Method,
	caller common.Address,
	args []interface{},
) ([]byte, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	tokenType, ok := args[0].(uint8)
	if !ok {
		return nil, fmt.Errorf("invalid tokenType")
	}

	salt, ok := args[1].([32]uint8)
	if !ok {
		return nil, fmt.Errorf("invalid salt")
	}

	address := crypto.CreateAddress2(caller, salt, calculateCodeHash(tokenType))

	return method.Outputs.Pack(address)
}

func calculateCodeHash(tokenType uint8) []byte {
	tokenTypeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(tokenTypeBytes, uint32(tokenType))
	return tokenTypeBytes
}

func (p Precompile) CreateCoinMetadata(ctx sdk.Context, address common.Address, name string, symbol string, decimals uint8) (*banktypes.Metadata, error) {
	addressString := address.String()
	denom := erc20types.CreateDenom(addressString)

	_, found := p.bankKeeper.GetDenomMetaData(ctx, denom)
	if found {
		return nil, errors.Wrap(
			erc20types.ErrInternalTokenPair, "denom metadata already registered",
		)
	}

	if p.erc20Keeper.IsDenomRegistered(ctx, denom) {
		return nil, errors.Wrapf(
			erc20types.ErrInternalTokenPair, "coin denomination already registered: %s", name,
		)
	}

	// base denomination
	base := erc20types.CreateDenom(addressString)

	// create a bank denom metadata based on the ERC20 token ABI details
	// metadata name is should always be the contract since it's the key
	// to the bank store
	metadata := banktypes.Metadata{
		Description: erc20types.CreateDenomDescription(addressString),
		Base:        base,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    base,
				Exponent: 0,
			},
		},
		Name:    base,
		Symbol:  symbol,
		Display: base,
	}

	// only append metadata if decimals > 0, otherwise validation fails
	if decimals > 0 {
		nameSanitized := erc20types.SanitizeERC20Name(name)
		metadata.DenomUnits = append(
			metadata.DenomUnits,
			&banktypes.DenomUnit{
				Denom:    nameSanitized,
				Exponent: uint32(decimals), //#nosec G115
			},
		)
		metadata.Display = nameSanitized
	}

	return &metadata, nil
}

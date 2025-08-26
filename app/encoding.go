package app

import (
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/address"
	types2 "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	evmethsecp256k1 "github.com/cosmos/evm/crypto/ethsecp256k1"
	evmtypes "github.com/cosmos/evm/types"
	"github.com/xrplevm/node/v9/legacy/crypto/ethsecp256k1"
	legacytypes "github.com/xrplevm/node/v9/legacy/types"

	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	enccodec "github.com/cosmos/evm/encoding/codec"
	"github.com/cosmos/evm/ethereum/eip712"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// MakeEncodingConfig creates an EncodingConfig for testing
// func MakeEncodingConfig() params.EncodingConfig {
// 	return evmenc.MakeConfig(ModuleBasics)
// }

func MakeEncodingConfig(evmChainID uint64) sdktestutil.TestEncodingConfig {
	cdc := amino.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
			vmtypes.MsgEthereumTxCustomGetSigner.MsgType:      vmtypes.MsgEthereumTxCustomGetSigner.Fn,
			erc20types.MsgConvertERC20CustomGetSigner.MsgType: erc20types.MsgConvertERC20CustomGetSigner.Fn,
		},
	}

	interfaceRegistry, _ := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})

	interfaceRegistry.RegisterImplementations((*sdk.AccountI)(nil),
		&evmtypes.EthAccount{},    // cosmos-evm
		&legacytypes.EthAccount{}, // evmos (legacy)
	)
	interfaceRegistry.RegisterImplementations((*authtypes.AccountI)(nil),
		&evmtypes.EthAccount{},    // cosmos-evm
		&legacytypes.EthAccount{}, // evmos (legacy)
	)
	interfaceRegistry.RegisterImplementations((*types2.PubKey)(nil),
		&evmethsecp256k1.PubKey{}, // cosmos-evm
		&ethsecp256k1.PubKey{},    // evmos (legacy)
	)
	interfaceRegistry.RegisterImplementations((*types2.PrivKey)(nil),
		&evmethsecp256k1.PrivKey{}, // cosmos-evm
		&ethsecp256k1.PrivKey{},    // evmos (legacy)
	)

	codec := amino.NewProtoCodec(interfaceRegistry)
	enccodec.RegisterLegacyAminoCodec(cdc)
	enccodec.RegisterInterfaces(interfaceRegistry)

	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = cdc
	eip712.SetEncodingConfig(cdc, interfaceRegistry, evmChainID)

	return sdktestutil.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}
}

package app

import (
	"fmt"

	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec/address"
	legacytypes "github.com/xrplevm/node/v12/types/legacy/ethermint/types"

	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	enccodec "github.com/cosmos/evm/encoding/codec"
	"github.com/cosmos/evm/ethereum/eip712"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	vmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/cosmos/gogoproto/proto"
	ica "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts"
	"google.golang.org/protobuf/reflect/protoreflect"

	feemarketlegacytypes "github.com/xrplevm/node/v12/types/legacy/ethermint/feemarket"
	erc20legacytypes "github.com/xrplevm/node/v12/types/legacy/evmos/erc20"
	poalegacytypes "github.com/xrplevm/node/v12/x/poa/types/legacy"
)

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

	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create interface registry: %w", err))
	}

	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil),
		&poalegacytypes.MsgAddValidator{},
		&poalegacytypes.MsgRemoveValidator{},
		&feemarketlegacytypes.MsgUpdateParams{},
		&erc20legacytypes.MsgConvertERC20{},
		&erc20legacytypes.MsgConvertCoin{},
		&erc20legacytypes.MsgUpdateParams{},
		&erc20legacytypes.MsgTransferOwnership{},
		&erc20legacytypes.MsgMint{},
		&erc20legacytypes.MsgBurn{},
		&erc20legacytypes.MsgRegisterERC20{},
		&erc20legacytypes.MsgToggleConversion{},
	)

	interfaceRegistry.RegisterImplementations((*sdk.AccountI)(nil),
		&legacytypes.EthAccount{}, // evmos (legacy)
	)
	interfaceRegistry.RegisterImplementations((*authtypes.AccountI)(nil),
		&legacytypes.EthAccount{}, // evmos (legacy)
	)

	// x/authz was removed in v12. Keep its types so historical txs still decode.
	authz.RegisterInterfaces(interfaceRegistry)
	// ICA host was removed in v12. Keep the ICS-27 types so historical txs and
	// interchain accounts in the auth store still decode.
	ica.AppModuleBasic{}.RegisterInterfaces(interfaceRegistry)
	// x/feegrant was removed in v12. Keep its types so historical txs still decode.
	feegrant.RegisterInterfaces(interfaceRegistry)
	// x/crisis was removed in v12. Keep its types so historical txs still decode.
	crisistypes.RegisterInterfaces(interfaceRegistry)
	// x/params was removed in v12. Keep ParameterChangeProposal so historical
	// txs and stored gov proposals still decode, and EIP-712 sign bytes of a
	// tx that carries it do not change.
	paramsproposal.RegisterInterfaces(interfaceRegistry)
	paramsproposal.RegisterLegacyAminoCodec(cdc)

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
		TxConfig:          tx.NewTxConfig(codec, EnabledSignModes),
		Amino:             cdc,
	}
}

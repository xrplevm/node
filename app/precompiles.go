package app

import (
	"fmt"
	"github.com/xrplevm/node/v6/precompiles/erc20factory"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/common"
	bankprecompile "github.com/evmos/evmos/v20/precompiles/bank"
	"github.com/evmos/evmos/v20/precompiles/bech32"
	distprecompile "github.com/evmos/evmos/v20/precompiles/distribution"
	govprecompile "github.com/evmos/evmos/v20/precompiles/gov"
	ics20precompile "github.com/evmos/evmos/v20/precompiles/ics20"
	"github.com/evmos/evmos/v20/precompiles/p256"
	stakingprecompile "github.com/evmos/evmos/v20/precompiles/staking"
	erc20Keeper "github.com/evmos/evmos/v20/x/erc20/keeper"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
	transferkeeper "github.com/evmos/evmos/v20/x/ibc/transfer/keeper"
	stakingkeeper "github.com/evmos/evmos/v20/x/staking/keeper"
	"golang.org/x/exp/maps"
)

const bech32PrecompileBaseGas = 6_000

// AvailableStaticPrecompiles returns the list of all available static precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	govKeeper govkeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	bech32Precompile, err := bech32.NewPrecompile(bech32PrecompileBaseGas)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(
		distributionKeeper,
		stakingKeeper,
		authzKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(
		stakingKeeper,
		transferKeeper,
		channelKeeper,
		authzKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	govPrecompile, err := govprecompile.NewPrecompile(govKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate gov precompile: %w", err))
	}

	erc20factoryPrecompile, err := erc20factory.NewPrecompile(authzKeeper, erc20Keeper, bankKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate erc20factory precompile: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile
	precompiles[govPrecompile.Address()] = govPrecompile
	precompiles[erc20factoryPrecompile.Address()] = erc20factoryPrecompile
	return precompiles
}

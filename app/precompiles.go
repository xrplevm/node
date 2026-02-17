package app

import (
	"github.com/cosmos/cosmos-sdk/codec"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/evm/precompiles/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	precompiletypes "github.com/cosmos/evm/precompiles/types"

	erc20Keeper "github.com/cosmos/evm/x/erc20/keeper"

	transferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v10/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// AvailableStaticPrecompiles returns the list of all available static precompiled contracts.
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	codec codec.Codec,
	opts ...precompiletypes.Option,
) map[common.Address]vm.PrecompiledContract {
	newPrecompiles := types.NewStaticPrecompiles().
		WithBech32Precompile().
		WithP256Precompile().
		WithStakingPrecompile(stakingKeeper, bankKeeper, opts...).
		WithDistributionPrecompile(distributionKeeper, stakingKeeper, bankKeeper, opts...).
		WithICS20Precompile(bankKeeper, stakingKeeper, &transferKeeper, &channelKeeper, &erc20Keeper).
		WithBankPrecompile(bankKeeper, &erc20Keeper).
		WithGovPrecompile(govKeeper, bankKeeper, codec, opts...)

	return newPrecompiles
}

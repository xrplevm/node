package erc20factory

import (
	storetypes "cosmossdk.io/store/types"
	"embed"
	"fmt"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	cmn "github.com/evmos/evmos/v20/precompiles/common"
	erc20keeper "github.com/evmos/evmos/v20/x/erc20/keeper"
	"github.com/evmos/evmos/v20/x/evm/core/vm"
)

const (
	Erc20FactoryAddress = "0x0000000000000000000000000000000000000900"
	// GasCreate defines the gas required to create a new ERC20 Token Pair calculated from a ERC20 deploy transaction
	GasCreate = 3_000_000
	// GasCalculateAddress defines the gas required to calculate the address of a new ERC20 Token Pair
	GasCalculateAddress = 3_000
)

var _ vm.PrecompiledContract = &Precompile{}

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

// Precompile defines the precompiled contract for Bech32 encoding.
type Precompile struct {
	cmn.Precompile
	erc20Keeper erc20keeper.Keeper
	bankKeeper  bankkeeper.Keeper
}

// NewPrecompile creates a new bech32 Precompile instance as a
// PrecompiledContract interface.
func NewPrecompile(authzKeeper authzkeeper.Keeper, erc20Keeper erc20keeper.Keeper, bankKeeper bankkeeper.Keeper) (*Precompile, error) {
	newABI, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  newABI,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration, // should be configurable in the future.
		},
		erc20Keeper: erc20Keeper,
		bankKeeper:  bankKeeper,
	}

	// SetAddress defines the address of the distribution compile contract.
	p.SetAddress(common.HexToAddress(Erc20FactoryAddress))
	return p, nil
}

// Address defines the address of the bech32 precompiled contract.
func (Precompile) Address() common.Address {
	return common.HexToAddress(Erc20FactoryAddress)
}

// RequiredGas calculates the contract gas use.
func (p Precompile) RequiredGas(input []byte) uint64 {
	// NOTE: This check avoid panicking when trying to decode the method ID
	if len(input) < 4 {
		return 0
	}

	methodID := input[:4]
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	switch method.Name {
	// ERC-20 transactions
	case CreateMethod:
		return GasCreate
	case CalculateAddressMethod:
		return GasCalculateAddress
	default:
		return 0
	}
}

// Run executes the precompiled contract bech32 methods defined in the ABI.
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}
	// This handles any out of gas errors that may occur during the execution of a precompile query.
	// It avoids panics and returns the out of gas error so the EVM can continue gracefully.
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	switch method.Name {
	// Bank queries
	case CreateMethod:
		bz, err = p.Create(ctx, method, contract.Caller(), args)
	case CalculateAddressMethod:
		bz, err = p.CalculateAddress(method, contract.Caller(), args)
	default:
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	if err != nil {
		return nil, err
	}

	cost := ctx.GasMeter().GasConsumed() - initialGas

	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
}

// IsTransaction checks if the given method name corresponds to a transaction or query.
//
// Available ERC20 Factory transactions are:
//   - Create
func (Precompile) IsTransaction(methodName string) bool {
	switch methodName {
	case CreateMethod:
		return true
	default:
		return false
	}
}

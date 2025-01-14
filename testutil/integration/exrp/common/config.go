package exrpcommon

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/xrplevm/node/v5/app"
)

const (
	bip44CoinType = 60
	ChainID       = "exrp_1440002-1"
)

// Config defines the configuration for a chain.
// It allows for customization of the network to adjust to
// testing needs.
type Config struct {
	ChainID            string
	EIP155ChainID      *big.Int
	AmountOfValidators int
	PreFundedAccounts  []sdktypes.AccAddress
	Balances           []banktypes.Balance
	BondDenom          string
	Denom              string
	CustomGenesisState CustomGenesisState
	GenesisBytes       []byte
	OtherCoinDenom     []string
	OperatorsAddrs     []sdktypes.AccAddress
	CustomBaseAppOpts  []func(*baseapp.BaseApp)
	MinDepositAmt      sdkmath.Int
	Quorum             string
}

type CustomGenesisState map[string]interface{}

// DefaultConfig returns the default configuration for a chain.
func DefaultConfig() Config {
	return Config{
		ChainID:       ChainID,
		EIP155ChainID: big.NewInt(1440002),
		Balances:      nil,
		Denom:         app.BaseDenom,
	}
}

// ConfigOption defines a function that can modify the NetworkConfig.
// The purpose of this is to force to be declarative when the default configuration
// requires to be changed.
type ConfigOption func(*Config)

// WithChainID sets a custom chainID for the network. It panics if the chainID is invalid.
func WithChainID(chainID string) ConfigOption {
	chainIDNum, err := evmostypes.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}
	return func(cfg *Config) {
		cfg.ChainID = chainID
		cfg.EIP155ChainID = chainIDNum
	}
}

// WithAmountOfValidators sets the amount of validators for the network.
func WithAmountOfValidators(amount int) ConfigOption {
	return func(cfg *Config) {
		cfg.AmountOfValidators = amount
	}
}

// WithPreFundedAccounts sets the pre-funded accounts for the network.
func WithPreFundedAccounts(accounts ...sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.PreFundedAccounts = accounts
	}
}

// WithBalances sets the specific balances for the pre-funded accounts, that
// are being set up for the network.
func WithBalances(balances ...banktypes.Balance) ConfigOption {
	return func(cfg *Config) {
		cfg.Balances = append(cfg.Balances, balances...)
	}
}

// WithDenom sets the denom for the network.
func WithDenom(denom string) ConfigOption {
	return func(cfg *Config) {
		cfg.Denom = denom
	}
}

// WithBondDenom sets the bond denom for the network.
func WithBondDenom(denom string) ConfigOption {
	return func(cfg *Config) {
		cfg.BondDenom = denom
	}
}

// WithCustomGenesis sets the custom genesis of the network for specific modules.
func WithCustomGenesis(customGenesis CustomGenesisState) ConfigOption {
	return func(cfg *Config) {
		cfg.CustomGenesisState = customGenesis
	}
}

// WithOtherDenoms sets other possible coin denominations for the network.
func WithOtherDenoms(otherDenoms []string) ConfigOption {
	return func(cfg *Config) {
		cfg.OtherCoinDenom = otherDenoms
	}
}

// WithValidatorOperators overwrites the used operator address for the network instantiation.
func WithValidatorOperators(keys []sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.OperatorsAddrs = keys
	}
}

// WithCustomBaseAppOpts sets custom base app options for the network.
func WithCustomBaseAppOpts(opts ...func(*baseapp.BaseApp)) ConfigOption {
	return func(cfg *Config) {
		cfg.CustomBaseAppOpts = opts
	}
}

// WithMinDepositAmt sets the min deposit amount for the network.
func WithMinDepositAmt(minDepositAmt sdkmath.Int) ConfigOption {
	return func(cfg *Config) {
		cfg.MinDepositAmt = minDepositAmt
	}
}

func WithQuorum(quorum string) ConfigOption {
	return func(cfg *Config) {
		cfg.Quorum = quorum
	}
}

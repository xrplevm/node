// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpnetwork

import (
	"math/big"
	"os"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmostypes "github.com/evmos/evmos/v20/types"
	"github.com/xrplevm/node/v4/app"
)

const (
	ChainID = "exrp_1440002-1"
)

// Config defines the configuration for a chain.
// It allows for customization of the network to adjust to
// testing needs.
type Config struct {
	chainID            string
	eip155ChainID      *big.Int
	amountOfValidators int
	preFundedAccounts  []sdktypes.AccAddress
	balances           []banktypes.Balance
	denom              string
	customGenesisState CustomGenesisState
	genesisBytes       []byte
	otherCoinDenom     []string
	operatorsAddrs     []sdktypes.AccAddress
	customBaseAppOpts  []func(*baseapp.BaseApp)
}

type CustomGenesisState map[string]interface{}

// DefaultConfig returns the default configuration for a chain.
func DefaultConfig() Config {
	return Config{
		chainID:            ChainID,
		eip155ChainID:      big.NewInt(1440002),
		balances:           nil,
		// MODIFIED
		denom:              app.BaseDenom,
		customGenesisState: nil,
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
		cfg.chainID = chainID
		cfg.eip155ChainID = chainIDNum
	}
}

// WithAmountOfValidators sets the amount of validators for the network.
func WithAmountOfValidators(amount int) ConfigOption {
	return func(cfg *Config) {
		cfg.amountOfValidators = amount
	}
}

// WithPreFundedAccounts sets the pre-funded accounts for the network.
func WithPreFundedAccounts(accounts ...sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.preFundedAccounts = accounts
	}
}

// WithBalances sets the specific balances for the pre-funded accounts, that
// are being set up for the network.
func WithBalances(balances ...banktypes.Balance) ConfigOption {
	return func(cfg *Config) {
		cfg.balances = append(cfg.balances, balances...)
	}
}

// WithDenom sets the denom for the network.
func WithDenom(denom string) ConfigOption {
	return func(cfg *Config) {
		cfg.denom = denom
	}
}

// WithCustomGenesis sets the custom genesis of the network for specific modules.
func WithCustomGenesis(customGenesis CustomGenesisState) ConfigOption {
	return func(cfg *Config) {
		cfg.customGenesisState = customGenesis
	}
}

// WithOtherDenoms sets other possible coin denominations for the network.
func WithOtherDenoms(otherDenoms []string) ConfigOption {
	return func(cfg *Config) {
		cfg.otherCoinDenom = otherDenoms
	}
}

// WithValidatorOperators overwrites the used operator address for the network instantiation.
func WithValidatorOperators(keys []sdktypes.AccAddress) ConfigOption {
	return func(cfg *Config) {
		cfg.operatorsAddrs = keys
	}
}

// WithCustomBaseAppOpts sets custom base app options for the network.
func WithCustomBaseAppOpts(opts ...func(*baseapp.BaseApp)) ConfigOption {
	return func(cfg *Config) {
		cfg.customBaseAppOpts = opts
	}
}

func WithGenesisFile(genesisFile string) ConfigOption {
	return func(cfg *Config) {
		genesisBytes, err := os.ReadFile(genesisFile)
		if err != nil {
			panic(err)
		}
		cfg.genesisBytes = genesisBytes
	}
}

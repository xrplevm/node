// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package exrpintegration

import (
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	testtx "github.com/evmos/evmos/v20/testutil/tx"
	exrpcommon "github.com/xrplevm/node/v8/testutil/integration/exrp/common"
)

// DefaultIntegrationConfig returns the default configuration for a chain.
func DefaultIntegrationConfig() exrpcommon.Config {
	account, _ := testtx.NewAccAddressAndKey()
	config := exrpcommon.DefaultConfig()
	config.AmountOfValidators = 3
	config.PreFundedAccounts = []sdktypes.AccAddress{account}
	return config
}

// getGenAccountsAndBalances takes the network configuration and returns the used
// genesis accounts and balances.
//
// NOTE: If the balances are set, the pre-funded accounts are ignored.
func getGenAccountsAndBalances(cfg exrpcommon.Config, validators []stakingtypes.Validator) (genAccounts []authtypes.GenesisAccount, balances []banktypes.Balance) {
	if len(cfg.Balances) > 0 {
		balances = cfg.Balances
		accounts := getAccAddrsFromBalances(balances)
		genAccounts = createGenesisAccounts(accounts)
	} else {
		genAccounts = createGenesisAccounts(cfg.PreFundedAccounts)
		balances = createBalances(cfg.PreFundedAccounts, append(cfg.OtherCoinDenom, cfg.Denom))
	}

	// append validators to genesis accounts and balances
	valAccs := make([]sdktypes.AccAddress, len(validators))
	for i, v := range validators {
		valAddr, err := sdktypes.ValAddressFromBech32(v.OperatorAddress)
		if err != nil {
			panic(fmt.Sprintf("failed to derive validator address from %q: %s", v.OperatorAddress, err.Error()))
		}
		valAccs[i] = sdktypes.AccAddress(valAddr.Bytes())
	}
	genAccounts = append(genAccounts, createGenesisAccounts(valAccs)...)

	return
}

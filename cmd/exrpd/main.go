package main

import (
	"fmt"
	"os"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethermint "github.com/evmos/evmos/v20/types"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/xrplevm/node/v6/app"
	"github.com/xrplevm/node/v6/cmd/exrpd/cmd"
)

func main() {
	initSDKConfig()
	registerDenoms()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func initSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := app.AccountAddressPrefix + "pub"
	validatorAddressPrefix := app.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := app.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := app.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := app.AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.SetCoinType(app.Bip44CoinType)
	config.SetPurpose(sdk.Purpose) // Shared
	// config.SetFullFundraiserPath(ethermint.BIP44HDPath) // nolint: staticcheck
	config.Seal()
}

func registerDenoms() {
	if err := sdk.RegisterDenom(app.DisplayDenom, math.LegacyOneDec()); err != nil {
		panic(err)
	}

	if err := sdk.RegisterDenom(app.BaseDenom, math.LegacyNewDecWithPrec(1, ethermint.BaseDenomUnit)); err != nil {
		panic(err)
	}
}

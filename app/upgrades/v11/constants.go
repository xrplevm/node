package v11

import sdkmath "cosmossdk.io/math"

const (
	UpgradeName = "v11.0.0"

	// mainnet
	MainnetChainID    = "xrplevm_1440000-1"
	ElysChannelID     = "channel-1"
	WithdrawalAddress = "ethm1m2pp8zjwk3ystxyxvw5h3mrhhhnzcr2ltjntz9"

	// testnet
	TestnetChainID           = "xrplevm_1449000-1"
	TestnetElysChannelID     = "channel-17"
	TestnetWithdrawalAddress = "ethm1p95fctckyrxuxu6t47e2uuckjl9tfuxynuawsc"

	// devnet
	DevnetChainID           = "xrplevm_1449900-1"
	DevnetElysChannelID     = "channel-4"
	DevnetWithdrawalAddress = "ethm1p95fctckyrxuxu6t47e2uuckjl9tfuxynuawsc"
)

// twoXRP expresses 2 XRP in axrp base units (axrp is atto-XRP — 18 decimals).
var twoXRP = sdkmath.NewIntWithDecimal(2, 18)

// mainnetElysAmount is the exact XRP stranded in the mainnet Elys channel escrow.
var mainnetElysAmount, _ = sdkmath.NewIntFromString("6955539034646993768414")

// ElysRecovery holds, for a single network, the Elys transfer channel whose
// escrow holds the stranded XRP, the address that should receive it, and the
// amount of XRP (in axrp base units) to unescrow.
type ElysRecovery struct {
	ChannelID         string
	WithdrawalAddress string
	Amount            sdkmath.Int
}

// ElysRecoveryByNetwork maps each network's Cosmos chain ID to its Elys recovery
// parameters. The v11 handler selects the entry matching ctx.ChainID().
var ElysRecoveryByNetwork = map[string]ElysRecovery{
	MainnetChainID: {
		ChannelID:         ElysChannelID,
		WithdrawalAddress: WithdrawalAddress,
		Amount:            mainnetElysAmount,
	},
	TestnetChainID: {
		ChannelID:         TestnetElysChannelID,
		WithdrawalAddress: TestnetWithdrawalAddress,
		Amount:            twoXRP,
	},
	DevnetChainID: {
		ChannelID:         DevnetElysChannelID,
		WithdrawalAddress: DevnetWithdrawalAddress,
		Amount:            twoXRP,
	},
}

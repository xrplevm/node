package v11

const (
	UpgradeName = "v11.0.0"

	// mainnet
	MainnetChainID = "xrplevm_1440000-1"
	ElysChannelID  = "channel-1"
	// TODO:  Set mainnet withdrawal address.
	WithdrawalAddress = "ethm1p95fctckyrxuxu6t47e2uuckjl9tfuxynuawsc"

	// testnet
	TestnetChainID       = "xrplevm_1449000-1"
	TestnetElysChannelID = "channel-3"
	// TODO: set testnet withdrawal address
	TestnetWithdrawalAddress = "ethm1p95fctckyrxuxu6t47e2uuckjl9tfuxynuawsc"
)

// ElysRecovery holds, for a single network, the Elys transfer channel whose
// escrow holds the stranded XRP and the address that should receive it.
type ElysRecovery struct {
	ChannelID         string
	WithdrawalAddress string
}

// ElysRecoveryByNetwork maps each network's Cosmos chain ID to its Elys recovery
// parameters. The v11 handler selects the entry matching ctx.ChainID().
var ElysRecoveryByNetwork = map[string]ElysRecovery{
	MainnetChainID: {
		ChannelID:         ElysChannelID,
		WithdrawalAddress: WithdrawalAddress,
	},
	TestnetChainID: {
		ChannelID:         TestnetElysChannelID,
		WithdrawalAddress: TestnetWithdrawalAddress,
	},
}

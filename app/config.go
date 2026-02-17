package app

type EVMOptionsFn func(uint64) error

const (
	AccountAddressPrefix = "ethm"
	Bip44CoinType        = 60
	Name                 = "exrp"
	// BaseDenom defines to the default denomination used in EVM
	BaseDenom                  = "axrp"
	Denom                      = "xrp"
	DenomDescription           = "XRP is a digital asset that's native to the XRP Ledger and XRPL EVM Sidechain, an open-source permissionless and decentralized blockchain technology."
	DenomName                  = "XRP"
	DenomSymbol                = "XRP"
	NativeErc20ContractAddress = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

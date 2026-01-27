package cmd

const (
	LocalnetChainID string = "xrplevm_1449999-1"
	DevnetChainID   string = "xrplevm_1449900-1"
	TestnetChainID  string = "xrplevm_1449000-1"
	MainnetChainID  string = "xrplevm_1440000-1"
)

const (
	LocalnetEVMChainID uint64 = 1449999
	MainnetEVMChainID  uint64 = 1440000
	TestnetEVMChainID  uint64 = 1449000
	DevnetEVMChainID   uint64 = 1449900
)

var CosmosToEVMChainID = map[string]uint64{
	LocalnetChainID: LocalnetEVMChainID,
	DevnetChainID:   DevnetEVMChainID,
	TestnetChainID:  TestnetEVMChainID,
	MainnetChainID:  MainnetEVMChainID,
}

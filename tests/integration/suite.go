package integration

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/evmos/v20/x/evm/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v6/app"
	factory "github.com/xrplevm/node/v6/testutil/integration/common/factory"
	"github.com/xrplevm/node/v6/testutil/integration/common/grpc"
	"github.com/xrplevm/node/v6/testutil/integration/common/keyring"
	exrpcommon "github.com/xrplevm/node/v6/testutil/integration/exrp/common"
)

type TestSuite struct {
	suite.Suite

	network     *Network
	keyring     keyring.Keyring
	factory     factory.CoreTxFactory
	grpcHandler grpc.Handler
}

func (s *TestSuite) Network() *Network {
	return s.network
}

func (s *TestSuite) SetupSuite() {
	s.network.SetupSdkConfig()
	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")
}

func (s *TestSuite) SetupTest() {
	// Check that the network was created successfully
	kr := keyring.New(5)

	customGenesis := exrpcommon.CustomGenesisState{}

	evmGen := evmtypes.DefaultGenesisState()

	evmGen.Params.EvmDenom = app.BaseDenom

	customGenesis[evmtypes.ModuleName] = evmGen

	s.network = NewIntegrationNetwork(
		exrpcommon.WithPreFundedAccounts(kr.GetAllAccAddrs()...),
		exrpcommon.WithAmountOfValidators(5),
		exrpcommon.WithCustomGenesis(customGenesis),
		exrpcommon.WithBondDenom("apoa"),
		exrpcommon.WithMinDepositAmt(sdkmath.NewInt(1)),
		exrpcommon.WithValidatorOperators(kr.GetAllAccAddrs()),
	)
	s.Require().NotNil(s.network)

	grpcHandler := grpc.NewIntegrationHandler(s.network)

	s.factory = factory.New(s.network, grpcHandler)
	s.keyring = kr
	s.grpcHandler = grpcHandler
}

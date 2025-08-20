package integration

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/evm/testutil/integration/common/factory"
	"github.com/cosmos/evm/testutil/integration/os/grpc"
	"github.com/cosmos/evm/testutil/integration/os/keyring"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v9/app"
	exrpcommon "github.com/xrplevm/node/v9/testutil/integration/exrp/common"
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
		exrpcommon.WithMaxValidators(7),
		exrpcommon.WithMinDepositAmt(sdkmath.NewInt(1)),
		exrpcommon.WithValidatorOperators(kr.GetAllAccAddrs()),
	)
	s.Require().NotNil(s.network)

	// TODO: Update when migrating to v10
	grpcHandler := grpc.NewIntegrationHandler(s.network)

	// TODO: Update when migrating to v10
	s.factory = factory.New(s.network, grpcHandler)
	s.keyring = kr
	s.grpcHandler = grpcHandler
}

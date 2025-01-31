package integration

import (
	"os/exec"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v6/testutil/integration/common/grpc"
	"github.com/xrplevm/node/v6/testutil/integration/common/keyring"
	exrpcommon "github.com/xrplevm/node/v6/testutil/integration/exrp/common"
	exrpupgrade "github.com/xrplevm/node/v6/testutil/integration/exrp/upgrade"
)

const (
	upgradeName = "v7.0.0"
)

type UpgradeTestSuite struct {
	suite.Suite

	network     *UpgradeTestNetwork
	keyring     keyring.Keyring
	grpcHandler grpc.Handler
}

func (s *UpgradeTestSuite) Network() *UpgradeTestNetwork {
	return s.network
}

func (s *UpgradeTestSuite) SetupSuite() {
	// Setup the SDK config
	s.network.SetupSdkConfig()

	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")
}

func (s *UpgradeTestSuite) SetupTest() {
	// Check that the network was created successfully
	kr := keyring.New(5)

	s.Require().NoError(exec.Command("cp", "-r", ".exrpd", ".exrpd-v7").Run())

	// Create the network
	s.network = NewUpgradeTestNetwork(
		exrpupgrade.WithUpgradePlanName(upgradeName),
		exrpupgrade.WithDataDir(".exrpd-v7/data"),
		exrpupgrade.WithNodeDBName("application"),
		exrpcommon.WithBondDenom("apoa"),
		exrpcommon.WithDenom("token"),
	)

	rpcHandler := grpc.NewIntegrationHandler(s.network)

	s.grpcHandler = rpcHandler
	s.keyring = kr
}

func (s *UpgradeTestSuite) TearDownTest() {
	s.Require().NoError(exec.Command("rm", "-rf", ".exrpd-v7").Run())
}

func (s *UpgradeTestSuite) RunUpgrade(name string) {
	res, err := s.network.GetUpgradeClient().CurrentPlan(
		s.network.GetContext(),
		&upgradetypes.QueryCurrentPlanRequest{},
	)
	s.Require().NoError(err)
	s.Require().Equal(name, res.Plan.Name)

	s.Require().True(s.Network().UpgradeKeeper().HasHandler(name))

	err = s.network.UpgradeKeeper().ApplyUpgrade(
		s.Network().GetContext(),
		*res.Plan,
	)

	s.Require().NoError(err)
}

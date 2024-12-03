package testupgrade

import (
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	upgradenetwork "github.com/xrplevm/node/v4/testutil/integration/exrp/upgrade"
)

const defaultStateFile = "upgrade-state.json"

type UpgradeTestSuite struct {
	suite.Suite

	network *UpgradeTestNetwork
}

func (s *UpgradeTestSuite) Network() *UpgradeTestNetwork {
	return s.network
}

func (s *UpgradeTestSuite) SetupTest() {
	// Get the state file from the environment variable, or use the default one
	stateFile := os.Getenv("UPGRADE_STATE_FILE")
	if stateFile == "" {
		stateFile = defaultStateFile
	}
	s.Require().NotEmpty(stateFile)

	// Setup the SDK config
	s.network.SetupSdkConfig()

	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")

	// Create the network
	s.network = NewUpgradeTestNetwork(
		upgradenetwork.WithGenesisFile(stateFile),
	)

	// Check that the network was created successfully
	s.Require().NotNil(s.network)
}
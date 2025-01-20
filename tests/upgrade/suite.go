package testupgrade

import (
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v5/tests/upgrade/testutil"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"
)

const (
	DefaultNodeDBName = ".exrp-upgrade"
	DefaultNodeDBDir  = "exrp-upgrade"
)

type UpgradeTestSuite struct {
	suite.Suite

	network *UpgradeTestNetwork
}

func (s *UpgradeTestSuite) Network() *UpgradeTestNetwork {
	return s.network
}

func (s *UpgradeTestSuite) SetupTest() {
	// Setup the SDK config
	s.network.SetupSdkConfig()

	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")

	s.Require().NoError(testutil.CopyNodeDB(DefaultNodeDBDir, DefaultNodeDBDir+"-tmp"))

	db, err := dbm.NewGoLevelDB(DefaultNodeDBName, DefaultNodeDBDir+"-tmp", nil)
	s.Require().NoError(err)

	// Create the network
	s.network = NewUpgradeTestNetwork(
		exrpcommon.WithCustomBaseAppOpts(func(ba *baseapp.BaseApp) {
			ba.SetDB(db)
		}),
	)

	// Check that the network was created successfully
	s.Require().NotNil(s.network)
}

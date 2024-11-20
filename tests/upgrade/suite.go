package testupgrade

import (
	"encoding/json"
	"os"

	"github.com/stretchr/testify/suite"
	exrpnetwork "github.com/xrplevm/node/v4/testutil/integration/exrp/network"
)

type UpgradeTestSuite struct {
	suite.Suite

	network *UpgradeTestNetwork
}

func (s *UpgradeTestSuite) SetupTest() {
	// READ APP STATE FILE

	genesisBytes, err := os.ReadFile("../../exported-state.json")
	s.Require().NoError(err)

	var genesisState exrpnetwork.CustomGenesisState

	err = json.Unmarshal(genesisBytes, &genesisState)
	s.Require().NoError(err)

	appState := genesisState["app_state"].(map[string]interface{})

	s.network = NewUpgradeTestNetwork(
		// LOAD APP STATE FROM FILE
		exrpnetwork.WithCustomGenesis(appState),
	)

	s.Require().NotNil(s.network)
}
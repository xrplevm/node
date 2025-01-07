package integration

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	exrpcommon "github.com/xrplevm/node/v5/testutil/integration/exrp/common"
)

type TestSuite struct {
	suite.Suite

	network *Network
}

func (s *TestSuite) Network() *Network {
	return s.network
}

func (s *TestSuite) SetupTest() {
	// Setup the SDK config
	s.network.SetupSdkConfig()

	s.Require().Equal(sdk.GetConfig().GetBech32AccountAddrPrefix(), "ethm")

	// Check that the network was created successfully
	s.network = NewIntegrationNetwork(
		exrpcommon.WithAmountOfValidators(5),
	)
	s.Require().NotNil(s.network)
}

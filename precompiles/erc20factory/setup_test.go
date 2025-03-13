package erc20factory_test

import (
	"github.com/xrplevm/node/v6/precompiles/erc20factory"
	"testing"

	testkeyring "github.com/evmos/evmos/v20/testutil/integration/evmos/keyring"
	"github.com/evmos/evmos/v20/testutil/integration/evmos/network"
	"github.com/stretchr/testify/suite"
)

var s *PrecompileTestSuite

// PrecompileTestSuite is the implementation of the TestSuite interface for ERC20 precompile
// unit tests.
type PrecompileTestSuite struct {
	suite.Suite

	network *network.UnitTestNetwork
	keyring testkeyring.Keyring

	precompile *erc20factory.Precompile
}

func TestPrecompileTestSuite(t *testing.T) {
	s = new(PrecompileTestSuite)
	suite.Run(t, s)
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)
	integrationNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	s.keyring = keyring
	s.network = integrationNetwork

	precompile, err := erc20factory.NewPrecompile(integrationNetwork.App.AuthzKeeper, integrationNetwork.App.Erc20Keeper, integrationNetwork.App.BankKeeper)
	s.Require().NoError(err, "failed to create erc20factory precompile")

	s.precompile = precompile
}

package testupgrade

import (
	"fmt"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"github.com/xrplevm/node/v4/app"
	exrpcommon "github.com/xrplevm/node/v4/testutil/integration/exrp/common"
)

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	denom := s.network.GetDenom()
	s.Require().NotEmpty(denom)
	s.Require().Equal(denom, app.BaseDenom)

	balances, err := exrpcommon.GetBankClient(s.Network()).AllBalances(s.network.GetContext(), &banktypes.QueryAllBalancesRequest{
		Address: "ethm1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3w48d64",
	})

	fmt.Println("balances", balances)
	s.Require().NoError(err)
	fmt.Println(balances)

	err = s.network.NextBlock()
	s.Require().NoError(err)
}

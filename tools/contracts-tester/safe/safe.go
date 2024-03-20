package safe

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/Peersyst/exrp/cmd/exrpd/cmd"
	"github.com/Peersyst/exrp/tools/contracts-tester/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type SafeTestSuite struct {
	t         *testing.T
	safe      *Safe
	threshold int
	claimer   types.TestAccount
	owners    []types.TestAccount
}

func CreateSafeSuite(t *testing.T) types.ContractTestSuite {
	return &SafeTestSuite{
		t: t,
	}
}

func (suite *SafeTestSuite) SetupEnv(ctx context.Context) error {
	client, err := ethclient.Dial(types.GetNodeUrl())
	if err != nil {
		return fmt.Errorf("SAFE_TEST: Error creating client")
	}

	safe, err := NewSafe(common.HexToAddress(cmd.SafeProxyAddress), client)
	if err != nil {
		return fmt.Errorf("Error instantiating safe contract: '%+v'", err)
	}
	suite.safe = safe

	suite.claimer = types.GetClaimerAccount(ctx)
	suite.owners = types.GetWitnesses(ctx)
	suite.threshold = types.GetSafeThreshold()

	return nil
}

func (suite *SafeTestSuite) RunTests() {
	suite.runThresholdTest()
	suite.runModulesTests()
	suite.runOwnersTests()
}

func (suite *SafeTestSuite) runThresholdTest() {
	threshold, err := suite.safe.GetThreshold(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting threshold: '%+v'", err)
	} else if threshold.Int64() != int64(suite.threshold) {
		suite.t.Errorf("Invalid threshold - expected '%+v' got '%+v'", suite.threshold, threshold.Int64())
	}
}

func (suite *SafeTestSuite) runModulesTests() {
	modules, err := suite.safe.GetModulesPaginated(suite.claimer.CallOpts, common.HexToAddress("0x1"), big.NewInt(10))
	if err != nil {
		suite.t.Errorf("Error getting modules paginated: '%+v'", err)
	} else {
		if len(modules.Array) != 1 {
			suite.t.Errorf("Invalid safe modules size - expected '%+v' got '%+v'", 1, len(modules.Array))
		} else {
			expected := common.HexToAddress(cmd.BridgeProxyModuleAddress).Hex()
			if modules.Array[0].Hex() != expected {
				suite.t.Errorf("Invalid module address - expected '%+v' got '%+v'", expected, modules.Array[0].Hex())
			}
		}
	}

	enabled, err := suite.safe.IsModuleEnabled(suite.claimer.CallOpts, common.HexToAddress(cmd.BridgeProxyModuleAddress))
	if err != nil {
		suite.t.Errorf("Error getting module enabled: '%+v'", err)
	} else if !enabled {
		suite.t.Errorf("Invalid module enabled value - expected '%+v' got '%+v'", true, enabled)
	}
}

func (suite *SafeTestSuite) runOwnersTests() {
	owners, err := suite.safe.GetOwners(suite.claimer.CallOpts)
	if err != nil {
		suite.t.Errorf("Error getting owners: '%+v'", err)
	} else {
		if len(owners) != len(suite.owners) {
			suite.t.Errorf("Invalid owners length - expected '%+v' got '%+v'", len(suite.owners), len(owners))
		}
		for _, expOwner := range suite.owners {
			found := false
			for _, owner := range owners {
				if owner.Hex() == expOwner.EvmAddress.Hex() {
					found = true
				}
			}
			if !found {
				suite.t.Errorf("Owner not found - expected '%+v'", expOwner.EvmAddress.Hex())
			}
		}
	}
}

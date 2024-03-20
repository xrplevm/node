package contracts

import (
	"context"
	"testing"

	"github.com/Peersyst/exrp/tools/contracts-tester/bridge"
	"github.com/Peersyst/exrp/tools/contracts-tester/safe"
	"github.com/Peersyst/exrp/tools/contracts-tester/types"
)

func Test_TestContracts(t *testing.T) {
	if !types.GetRunTests() {
		t.SkipNow()
		return
	}

	ctx := context.Background()

	t.Logf("Setting up safe tests...")
	safeSuite := safe.CreateSafeSuite(t)
	safeSuite.SetupEnv(ctx)
	t.Logf("Running safe tests...")
	safeSuite.RunTests()

	t.Logf("Setting up bridge tests...")
	bridgeSuite := bridge.CreateBridgeSuite(t)
	bridgeSuite.SetupEnv(ctx)
	t.Logf("Running bridge tests...")
	bridgeSuite.RunTests()
}

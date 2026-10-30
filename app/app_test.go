package app_test

import (
	"sync"

	"github.com/xrplevm/node/v12/app"
	"github.com/xrplevm/node/v12/testutil/constants"
	exrpcommon "github.com/xrplevm/node/v12/testutil/integration/exrp/common"
)

var (
	testAppOnce sync.Once
	testApp     *app.App
)

// sharedTestApp returns one app for the whole package. The EVM keeper sets a
// global chain config, so without the test build tag (as in make test-app) a
// process can build only one app.
func sharedTestApp() *app.App {
	testAppOnce.Do(func() {
		testApp = exrpcommon.CreateExrpApp(constants.LocalnetChainID.ChainID)
	})
	return testApp
}

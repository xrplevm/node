// Command exrpd is the validator build of exrpd. It carries the exrpd release the chain
// runs now, the exrpd release its next on-chain upgrade needs, and Cosmovisor. "exrpd
// start" runs the node under that Cosmovisor, so it switches binaries at the upgrade
// height without downloading anything. Every other command runs with the exrpd that
// runs the node data.
//
// Only start, init on a new home and bundle setup install the binaries. Every other
// command runs only when the marker says that this bundle installed them, and otherwise
// refuses and names the next step.
//
// The binaries live in the stock Cosmovisor layout under <home>/cosmovisor:
//
//	upgrades/<plan>/bin/exrpd     exrpd of each plan, the genesis plan included
//	supervisor/<sha>/cosmovisor   the embedded Cosmovisor
//	current -> upgrades/<plan>
//	bundle.sha256                 the marker: the bundle that installed the binaries
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "exrpd:", err)
		os.Exit(1)
	}
}

// logf reports a change that exrpd made to the home on stderr, where the service
// manager's log keeps it.
func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "exrpd: "+format+"\n", args...)
}

// run resolves the home once, so Cosmovisor, the launcher and exrpd all use it, loads
// the bundle once, and dispatches on the command.
func run(args []string) error {
	inv, err := parseCommand(args, os.Getenv, os.UserHomeDir)
	if err != nil {
		return err
	}
	b, err := loadBundle(embeddedPayload())
	if err != nil {
		return err
	}
	n := node{Layout: newLayout(inv.home, b.m), Source: inv.source}
	switch inv.command {
	case startCommand:
		return runStart(b, n, inv.args)
	case initCommand:
		return runInit(b, n, inv.args)
	case bundleSetupCommand:
		_, err := b.stage(n)
		return err
	case bundleCommand:
		fmt.Println(bundleUsage)
		return nil
	case versionCommand:
		return runVersion(b, n)
	case versionBundleCommand:
		return runVersionBundle(b, n)
	case completeCommand:
		return runCompletion(b, n, inv.args)
	default:
		return runExrpd(b, n, inv.args)
	}
}

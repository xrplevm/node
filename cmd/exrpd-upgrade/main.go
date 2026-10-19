// Command exrpd-upgrade is the validator build of exrpd. It carries two exrpd releases
// (the one the chain runs now and the one its next on-chain upgrade needs) plus
// Cosmovisor, and runs exrpd under that Cosmovisor so the node switches binaries at the
// upgrade height without downloading anything.
//
// Every command works on the stock Cosmovisor layout under <home>/cosmovisor:
//
//	genesis/bin/exrpd             exrpd the chain runs before the upgrade
//	upgrades/<plan>/bin/exrpd     exrpd the upgrade plan needs
//	supervisor/<sha>/cosmovisor   the embedded Cosmovisor
//	current -> genesis | upgrades/<plan>
package main

import (
	"errors"
	"fmt"
	"os"
)

const usage = `Usage:
  exrpd-upgrade start [--home DIR] [exrpd start flags...]
      Stage the embedded binaries and run "exrpd start" under Cosmovisor.
  exrpd-upgrade stage [--home DIR] [--select genesis|PLAN]
      Stage the embedded binaries into DIR/cosmovisor and exit. --select points
      DIR/cosmovisor/current at the genesis binary or at the binary of upgrade PLAN,
      for data restored from a snapshot or state sync.
  exrpd-upgrade info
      Print the embedded manifest.

DIR defaults to $DAEMON_HOME, then $HOME/.exrpd.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "exrpd-upgrade:", err)
		os.Exit(1)
	}
}

// logf reports a change that exrpd-upgrade made to the home on stderr, where the
// service manager's log keeps it.
func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "exrpd-upgrade: "+format+"\n", args...)
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return errors.New("missing command")
	}
	command, args := args[0], args[1:]
	switch command {
	case "start":
		return runStart(args)
	case "stage":
		return runStage(args)
	case "info":
		return runInfo()
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", command)
	}
}

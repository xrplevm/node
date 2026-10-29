// Command exrpd is the validator build of exrpd. It carries the exrpd release the chain
// runs now, the exrpd releases its next on-chain upgrades need, and Cosmovisor. "exrpd
// start" runs the node under that Cosmovisor, so it switches binaries at the upgrade
// height without downloading anything. Every other command runs with the exrpd that
// runs the node data.
//
// The binaries live in the stock Cosmovisor layout under <home>/cosmovisor:
//
//	genesis/bin/exrpd             exrpd of a home without data/upgrade-info.json
//	upgrades/<plan>/bin/exrpd     exrpd of each plan, the genesis plan included
//	supervisor/<sha>/cosmovisor   the embedded Cosmovisor
//	current -> genesis | upgrades/<plan>
package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
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

// run resolves the home once, so Cosmovisor, the launcher and exrpd all use it, and
// dispatches on the command.
func run(args []string) error {
	home, source, command, args, err := parseCommand(args, os.Getenv, os.UserHomeDir)
	if err != nil {
		return err
	}
	switch command {
	case startCommand:
		return runStart(home, source, args)
	case versionCommand:
		return runVersion(home, source, args)
	case completeCommand:
		return runCompletion(home, args)
	case helpCommand:
		return runHelp(home, args)
	default:
		return runExrpd(home, args)
	}
}

// parseCommand returns the home and its source, the launcher command ("start",
// "version", "__complete", "help", or "" for a command that goes to exrpd) and its
// arguments. --home may come before the command, as with the standalone exrpd.
//
// A shell completion request (a TAB key) carries the words typed so far. A shell alias
// such as `alias exrpd='exrpd --home /h'` puts --home before them. The words from
// __complete on go to exrpd unchanged, so cobra completes what was typed, and a
// half-typed --home picks the home of the alias, else the default, instead of failing.
// "exrpd completion" writes a script and is an ordinary command.
func parseCommand(args []string, getenv func(string) string, userHomeDir func() (string, error)) (home, source, command string, rest []string, err error) {
	if at := completionAt(args); at >= 0 {
		prefix := args[:at]
		home, source, _, err = resolveHome(append(slices.Clip(prefix), args[at+1:]...), getenv, userHomeDir)
		if err != nil {
			home, source, _, err = resolveHome(prefix, getenv, userHomeDir)
		}
		if err != nil {
			home, source, _, err = resolveHome(nil, getenv, userHomeDir)
		}
		return home, source, completeCommand, args[at:], err
	}
	home, source, args, err = resolveHome(args, getenv, userHomeDir)
	if err != nil {
		return "", "", "", nil, err
	}
	// cobra skips empty words when it finds the command, so `exrpd "" start` is start. A
	// leading empty word cannot be the value of a flag.
	for len(args) > 0 && args[0] == "" {
		args = args[1:]
	}
	if len(args) > 0 && args[0] == startCommand {
		return home, source, startCommand, args[1:], nil
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "-") && slices.Contains(args, startCommand) {
		// The launcher cannot tell which flag takes a value, so "start" after a flag may
		// be the command. Without Cosmovisor the node would not upgrade.
		return "", "", "", nil, errors.New(`write the command first: exrpd start [flags]`)
	}
	if len(args) == 0 || args[0] == helpCommand || slices.Contains(args, "-h") || slices.Contains(args, "--help") {
		return home, source, helpCommand, args, nil
	}
	if args[0] == versionCommand {
		return home, source, versionCommand, args[1:], nil
	}
	return home, source, "", args, nil
}

// completionAt returns the index of the __complete word of a shell completion request,
// or -1. Only empty words and --home with its value may come before it.
func completionAt(args []string) int {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "" || strings.HasPrefix(args[i], "--home="):
		case args[i] == "--home":
			i++
		case args[i] == completeCommand || args[i] == "__completeNoDesc":
			return i
		default:
			return -1
		}
	}
	return -1
}

const (
	startCommand   = "start"
	versionCommand = "version"
	// completeCommand is cobra's hidden command for shell completion requests.
	completeCommand = "__complete"
	// helpCommand is a help request of any command but start, which always goes to
	// Cosmovisor: in "start --moniker --help", --help is the moniker.
	helpCommand = "help"
)

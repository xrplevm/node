package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// The launcher only owns --home and --select. Every other argument belongs to
// "exrpd start". A full flag parser would reject those, so flags are taken out of the
// argument list by name instead.

// takeFlag removes every "name VALUE" and "name=VALUE" from args. It returns the last
// value, the remaining arguments in order, and whether the flag was present.
func takeFlag(args []string, name string) (value string, rest []string, found bool, err error) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == name:
			if i+1 == len(args) || args[i+1] == "" {
				return "", nil, false, fmt.Errorf("%s requires a value", name)
			}
			i++
			value, found = args[i], true
		case strings.HasPrefix(arg, name+"="):
			value, found = strings.TrimPrefix(arg, name+"="), true
			if value == "" {
				return "", nil, false, fmt.Errorf("%s requires a value", name)
			}
		default:
			rest = append(rest, arg)
		}
	}
	return value, rest, found, nil
}

// resolveHome removes --home from args and returns the node home as an absolute path:
// the flag, else $DAEMON_HOME, else $HOME/.exrpd.
//
// It refuses "--". exrpd reads "--moniker --" as the moniker "--" and then reads a later
// --home. takeFlag cannot tell such a value from a terminator, so allowing "--" could
// give Cosmovisor and exrpd different homes. "exrpd start" ignores positional arguments,
// so "--" has no use there.
func resolveHome(args []string, getenv func(string) string, userHomeDir func() (string, error)) (home string, rest []string, err error) {
	if slices.Contains(args, "--") {
		return "", nil, errors.New(`"--" is not supported: write a flag value that is "--" as --flag=--`)
	}
	home, rest, _, err = takeFlag(args, "--home")
	if err != nil {
		return "", nil, err
	}
	if home == "" {
		home = getenv("DAEMON_HOME")
	}
	if home == "" {
		dir, err := userHomeDir()
		if err != nil {
			return "", nil, err
		}
		home = filepath.Join(dir, ".exrpd")
	}
	home, err = filepath.Abs(home)
	return home, rest, err
}

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// The launcher only owns --home and --select. Every other argument belongs to
// "exrpd start". A full flag parser would reject those, so flags are taken out of the
// argument list by name instead.

// takeFlag removes every "name VALUE" and "name=VALUE" before a "--" terminator from
// args. It returns the last value, the remaining arguments in order (the terminator and
// everything after it unchanged), and whether the flag was present.
func takeFlag(args []string, name string) (value string, rest []string, found bool, err error) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return value, append(rest, args[i:]...), found, nil
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
func resolveHome(args []string, getenv func(string) string, userHomeDir func() (string, error)) (home string, rest []string, err error) {
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

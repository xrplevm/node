package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// The launcher only owns --home. Every other argument belongs to the exrpd command. A
// full flag parser would reject those, so --home is taken out of the argument list by
// name instead.
//
// Taking a flag by name cannot see which tokens exrpd reads as flag values. exrpd reads
// "--moniker --home=/srv/name" as the moniker "--home=/srv/name", but the launcher would
// take /srv/name as the home and stage binaries there. The launcher does not know the
// flags of exrpd, so it cannot tell a boolean flag such as --trace, which does not take
// the next token, from a string flag such as --moniker, which does. A list of boolean
// flags could become incorrect when exrpd changes and then select the wrong home
// silently. Thus takeHome refuses --home when it directly follows a forwarded flag
// without "=". The user can write the forwarded flag as --flag=VALUE or --flag=true, or
// move the owned flag before it. This refuses some unambiguous forms, such as
// "--trace --home /x", but it never selects a home that exrpd does not use.

// takeHome removes every "--home VALUE" and "--home=VALUE" from args. It returns the
// last value, "" when there is none, and the remaining arguments in order. It refuses
// --home when the token before it is a flag without "=" that exrpd can read as taking
// --home as its value.
func takeHome(args []string) (value string, rest []string, err error) {
	const name = "--home"
	rest = make([]string, 0, len(args))
	valueAt := -1
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == name || strings.HasPrefix(arg, name+"=")) && i > 0 && i-1 != valueAt && mayTakeValue(args[i-1]) {
			prev := args[i-1]
			return "", nil, fmt.Errorf("%s follows %s, and exrpd can read %q as the value of %s: write %s=%s if that is the value, write %s=true if %s is a boolean flag, or put %s first",
				name, prev, arg, prev, prev, arg, prev, prev, name)
		}
		switch {
		case arg == name:
			if i+1 == len(args) || args[i+1] == "" {
				return "", nil, fmt.Errorf("%s requires a value", name)
			}
			i++
			valueAt = i
			value = args[i]
		case strings.HasPrefix(arg, name+"="):
			value = strings.TrimPrefix(arg, name+"=")
			if value == "" {
				return "", nil, fmt.Errorf("%s requires a value", name)
			}
		default:
			rest = append(rest, arg)
		}
	}
	return value, rest, nil
}

// mayTakeValue reports whether arg is a flag token that can take the next argument as
// its value. "-" alone is a positional argument. --bundle is the launcher's own boolean
// flag of "exrpd version".
func mayTakeValue(arg string) bool {
	return strings.HasPrefix(arg, "-") && arg != "-" && !strings.Contains(arg, "=") && arg != "--bundle"
}

// resolveHome removes --home from args and returns the node home as an absolute path,
// and its source: the flag, else $DAEMON_HOME, else $EXRPD_HOME, else $NODE_HOME, else
// $HOME/.exrpd ("default").
//
// It refuses "--" for every command. exrpd reads "--moniker --" as the moniker "--" and
// then reads a later --home. takeHome cannot tell such a value from a terminator, so
// allowing "--" could give Cosmovisor, the launcher and exrpd different homes. A value
// "--" can still be written as --flag=--.
func resolveHome(args []string, getenv func(string) string, userHomeDir func() (string, error)) (home, source string, rest []string, err error) {
	if slices.Contains(args, "--") {
		return "", "", nil, errors.New(`"--" is not supported: write a flag value that is "--" as --flag=--`)
	}
	home, rest, err = takeHome(args)
	if err != nil {
		return "", "", nil, err
	}
	source = "--home"
	// The standalone exrpd runs its node, and every command that opens the node's config
	// and data, on $EXRPD_HOME (viper prefix "exrpd"). It reads its default home from
	// $NODE_HOME (client/v2 helpers).
	for _, key := range []string{"DAEMON_HOME", "EXRPD_HOME", "NODE_HOME"} {
		if home == "" {
			home, source = getenv(key), "$"+key
		}
	}
	if home == "" {
		dir, err := userHomeDir()
		if err != nil {
			return "", "", nil, err
		}
		home, source = filepath.Join(dir, ".exrpd"), "default"
	}
	home, err = filepath.Abs(home)
	return home, source, rest, err
}

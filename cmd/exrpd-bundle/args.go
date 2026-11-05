package main

import (
	"cmp"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// invocation is a parsed command line.
type invocation struct {
	home, source, command string
	args                  []string
}

// parseCommand returns the home and its source, the launcher command ("start", "init",
// "bundle", "bundle setup", "version", "version --bundle", "__complete", or "" for a
// command that goes to exrpd) and its arguments. splitCommand reads the value of --home,
// and homeFrom turns it into the home once for every command.
func parseCommand(args []string, getenv func(string) string, userHomeDir func() (string, error)) (invocation, error) {
	value, inv, err := splitCommand(args)
	if err != nil {
		return invocation{}, err
	}
	if inv.home, inv.source, err = homeFrom(value, getenv, userHomeDir); err != nil {
		return invocation{}, err
	}
	return inv, nil
}

// splitCommand returns the value of --home, "" when there is none, and the command with
// its arguments. It finds the command word first. --home may come before it, as with the
// standalone exrpd. The launcher owns every argument of bundle and of version --bundle.
// Of any other command it owns only --home. A help request goes to exrpd like any other
// command, so "init --help" never installs.
//
// A shell completion request (a TAB key) carries the words typed so far. A shell alias
// puts its flags before them, such as `alias exrpd='exrpd --home /h'` or
// `alias exrpd='exrpd --log_level info'`: completeAt finds the request behind them.
// "exrpd completion" writes a script and is an ordinary command.
func splitCommand(args []string) (string, invocation, error) {
	if at := completeAt(args); at >= 0 {
		value, rest := splitCompletion(args, at)
		return value, invocation{command: completeCommand, args: rest}, nil
	}
	if at := commandAt(args); at < len(args) && (args[at] == bundleCommand || args[at] == versionCommand && slices.Contains(args[at+1:], "--bundle")) {
		return splitOwned(args, at)
	}
	value, args, err := forwardedHome(args)
	if err != nil {
		return "", invocation{}, err
	}
	// cobra skips empty words when it finds the command, so `exrpd "" start` is start. A
	// leading empty word cannot be the value of a flag.
	for len(args) > 0 && args[0] == "" {
		args = args[1:]
	}
	// start goes to Cosmovisor with all its arguments: in "start --moniker --help", --help
	// is the moniker.
	if len(args) > 0 && args[0] == startCommand {
		return value, invocation{command: startCommand, args: args[1:]}, nil
	}
	if len(args) > 0 && strings.HasPrefix(args[0], "-") && (slices.Contains(args, startCommand) || slices.Contains(args, initCommand)) {
		// The launcher cannot tell which flag takes a value, so "start" or "init" after a
		// flag may be the command. Without Cosmovisor the node would not upgrade, and
		// init would not install the bundle.
		return "", invocation{}, errors.New(`write the command first: exrpd init|start [flags]`)
	}
	if len(args) > 0 && args[0] == initCommand {
		for _, arg := range args[1:] {
			short, _, _ := strings.Cut(arg, "=")
			if len(short) > 2 && short[0] == '-' && short[1] != '-' && strings.ContainsRune(short, 'h') {
				return "", invocation{}, fmt.Errorf("%s may request help: write short flags separately for exrpd init", arg)
			}
		}
	}
	// version with any argument, a help flag included, goes to exrpd.
	if len(args) == 1 && args[0] == versionCommand ||
		len(args) > 0 && args[0] == initCommand && !slices.ContainsFunc(args, helpFlag) {
		return value, invocation{command: args[0], args: args[1:]}, nil
	}
	return value, invocation{args: args}, nil
}

// commandAt returns the index of the command word: the first word that is not empty and
// not --home with its value. It is len(args) when there is none.
func commandAt(args []string) int {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "" || strings.HasPrefix(args[i], homeFlag+"="):
		case args[i] == homeFlag:
			i++
		default:
			return i
		}
	}
	return len(args)
}

// isCompleteWord reports whether s is one of cobra's words for a shell completion
// request.
func isCompleteWord(s string) bool {
	return s == completeCommand || s == completeNoDescCommand
}

// completeAt returns the index of cobra's completion word when args are a shell
// completion request, and -1 otherwise. Every word before that word is empty, a flag, or
// the word after a flag that can take it as its value, as an alias such as
// `alias exrpd='exrpd --log_level info'` puts them. So "keys show __complete" is not a
// completion request. The word after --home is always its value, as commandAt reads it:
// "--home __complete start" is start.
func completeAt(args []string) int {
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == homeFlag:
			i++
		case isCompleteWord(arg):
			return i
		case arg == "", strings.HasPrefix(arg, "-"), i > 0 && mayTakeValue(args[i-1]):
		default:
			return -1
		}
	}
	return -1
}

// splitCompletion reads a completion request whose word is args[at]. exrpd gets the
// words without the empty words and the --home before the request, because runCompletion
// passes the home itself. The other words go to exrpd unchanged, so cobra completes what
// was typed. The home is the last --home of the typed words, else the --home of the
// alias, else the default: a half-typed or refused --home never fails a TAB key.
func splitCompletion(args []string, at int) (value string, rest []string) {
	var prefix []string
	for i := 0; i < at; i++ {
		switch arg := args[i]; {
		case arg == "":
		case arg == homeFlag:
			// completeAt skipped the value, so it comes before args[at].
			i++
			value = args[i]
		case strings.HasPrefix(arg, homeFlag+"="):
			value = strings.TrimPrefix(arg, homeFlag+"=")
		default:
			prefix = append(prefix, arg)
		}
	}
	if typed, _, err := forwardedHome(slices.Concat(prefix, args[at+1:])); err == nil && typed != "" {
		value = typed
	}
	return value, slices.Concat(prefix, args[at:])
}

// splitOwned reads a command line whose command, args[at], the launcher owns with every
// argument. version --bundle takes no other flag: passed to exrpd, --bundle would fail
// with "unknown flag". bundle takes setup, and help flags before or after it, which ask
// for the usage.
func splitOwned(args []string, at int) (string, invocation, error) {
	value, rest, err := ownedHome(args, at)
	if err != nil {
		return "", invocation{}, err
	}
	if args[at] == versionCommand {
		for _, arg := range rest {
			if arg != "--bundle" {
				return "", invocation{}, fmt.Errorf("version --bundle takes no other flag, got %s", arg)
			}
		}
		return value, invocation{command: versionBundleCommand}, nil
	}
	if words := slices.DeleteFunc(slices.Clone(rest), helpFlag); len(words) > 1 || len(words) == 1 && words[0] != setupCommand {
		return "", invocation{}, fmt.Errorf("unknown bundle command %q\n\n%s", strings.Join(rest, " "), bundleUsage)
	}
	if slices.Equal(rest, []string{setupCommand}) {
		return value, invocation{command: bundleSetupCommand}, nil
	}
	return value, invocation{command: bundleCommand}, nil
}

const (
	startCommand = "start"
	// initCommand creates a node home. On a new home it installs the bundle first.
	initCommand = "init"
	// bundleCommand holds the launcher's own subcommands. Alone or with a help flag, it
	// prints their usage.
	bundleCommand = "bundle"
	setupCommand  = "setup"
	// bundleSetupCommand is "bundle setup", which installs the bundle.
	bundleSetupCommand = "bundle setup"
	versionCommand     = "version"
	// versionBundleCommand is "version --bundle", which the launcher answers itself.
	versionBundleCommand = "version --bundle"
	// completeCommand and completeNoDescCommand are cobra's hidden commands for shell
	// completion requests.
	completeCommand       = "__complete"
	completeNoDescCommand = "__completeNoDesc"
)

// For a command that goes to exrpd, the launcher only owns --home. Every other argument
// belongs to the exrpd command. A full flag parser would reject those, so --home is
// taken out of the argument list by name instead.
//
// Taking a flag by name cannot see which tokens exrpd reads as flag values. exrpd reads
// "--moniker --home=/srv/name" as the moniker "--home=/srv/name", but the launcher would
// take /srv/name as the home and stage binaries there. The launcher does not know the
// flags of exrpd, so it cannot tell a boolean flag such as --trace, which does not take
// the next token, from a string flag such as --moniker, which does. A list of boolean
// flags could become incorrect when exrpd changes and then select the wrong home
// silently. Thus forwardedHome refuses --home when it directly follows a forwarded flag
// without "=". The user can write the forwarded flag as --flag=VALUE or --flag=true, or
// move the owned flag before it. This refuses some unambiguous forms, such as
// "--trace --home /x", but it never selects a home that exrpd does not use.
//
// The launcher owns every argument of its own commands, which parseCommand names. It
// knows each of their flags, and only --home takes a value, so --home may follow any of
// them.

// homeFlag is the only flag of exrpd that the launcher reads.
const homeFlag = "--home"

// takeHome removes every "--home VALUE" and "--home=VALUE" from args. It returns the
// last value, "" when there is none, and the remaining arguments in order. It refuses
// --home when takesValue reports that the token before it can take --home as its value.
func takeHome(args []string, takesValue func(string) bool) (value string, rest []string, err error) {
	rest = make([]string, 0, len(args))
	valueAt := -1
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if (arg == homeFlag || strings.HasPrefix(arg, homeFlag+"=")) && i > 0 && i-1 != valueAt && takesValue(args[i-1]) {
			prev := args[i-1]
			return "", nil, fmt.Errorf("%s follows %s, and exrpd can read %q as the value of %s: write %s=%s if that is the value, write %s=true if %s is a boolean flag, or put %s first",
				homeFlag, prev, arg, prev, prev, arg, prev, prev, homeFlag)
		}
		switch {
		case arg == homeFlag:
			if i+1 == len(args) || args[i+1] == "" {
				return "", nil, fmt.Errorf("%s requires a value", homeFlag)
			}
			i++
			valueAt = i
			value = args[i]
		case strings.HasPrefix(arg, homeFlag+"="):
			value = strings.TrimPrefix(arg, homeFlag+"=")
			if value == "" {
				return "", nil, fmt.Errorf("%s requires a value", homeFlag)
			}
		default:
			rest = append(rest, arg)
		}
	}
	return value, rest, nil
}

// mayTakeValue reports whether arg is a flag token that exrpd can read as taking the
// next argument as its value. "-" alone is a positional argument, and cobra's help flag
// is a boolean in every command.
func mayTakeValue(arg string) bool {
	return strings.HasPrefix(arg, "-") && arg != "-" && !strings.Contains(arg, "=") && !helpFlag(arg)
}

// helpFlag reports whether arg is cobra's help flag set to true: -h, --help, or either
// with a value that strconv.ParseBool reads as true, as pflag parses a boolean flag.
func helpFlag(arg string) bool {
	for _, name := range []string{"-h", "--help"} {
		if arg == name {
			return true
		}
		if v, ok := strings.CutPrefix(arg, name+"="); ok {
			b, err := strconv.ParseBool(v)
			return err == nil && b
		}
	}
	return false
}

// forwardedHome removes --home from the arguments of a command that goes to exrpd. It
// returns the value of the last --home, "" when there is none, and the other arguments.
//
// It refuses "--" for every such command. exrpd reads "--moniker --" as the moniker "--"
// and then reads a later --home. takeHome cannot tell such a value from a terminator, so
// allowing "--" could give Cosmovisor, the launcher and exrpd different homes. A value
// "--" can still be written as --flag=--.
func forwardedHome(args []string) (value string, rest []string, err error) {
	if slices.Contains(args, "--") {
		return "", nil, errors.New(`"--" is not supported: write a flag value that is "--" as --flag=--`)
	}
	return takeHome(args, mayTakeValue)
}

// ownedHome is forwardedHome for a command that the launcher owns, whose word is
// args[at]. No flag but --home takes a value there. A --home after the command word wins
// over one before it. rest is the arguments after the command word, without --home. A
// --home value that starts with "-" is refused: it is almost surely a mistyped flag.
func ownedHome(args []string, at int) (value string, rest []string, err error) {
	noValue := func(string) bool { return false }
	// Before the command word there are only --home and empty words.
	before, _, err := takeHome(args[:at], noValue)
	if err != nil {
		return "", nil, err
	}
	after, rest, err := takeHome(args[at+1:], noValue)
	if err != nil {
		return "", nil, err
	}
	value = cmp.Or(after, before)
	if strings.HasPrefix(value, "-") {
		return "", nil, fmt.Errorf("--home requires a value, got %q", value)
	}
	return value, rest, nil
}

// homeFrom returns the node home as an absolute path, and its source: value, the value
// of --home, else $DAEMON_HOME, else $EXRPD_HOME, else $NODE_HOME, else $HOME/.exrpd
// ("default").
func homeFrom(value string, getenv func(string) string, userHomeDir func() (string, error)) (home, source string, err error) {
	home, source = value, homeFlag
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
			return "", "", err
		}
		home, source = filepath.Join(dir, ".exrpd"), "default"
	}
	home, err = filepath.Abs(home)
	return home, source, err
}

package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// "start" must reach Cosmovisor wherever --home is, as with the standalone exrpd. A
// "start" that went to exrpd would run the node without the upgrade switch.
func TestParseCommand(t *testing.T) {
	noEnv := func(string) string { return "" }
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	userHomeDir := testUserHomeDir
	for _, tc := range []struct {
		args              []string
		command, wantHome string
		wantRest          []string
	}{
		{[]string{"start", "--home", "/h", "--x=1"}, "start", "/h", []string{"--x=1"}},
		{[]string{"--home", "/h", "start", "--x=1"}, "start", "/h", []string{"--x=1"}},
		{[]string{"keys", "list", "--home", "/h"}, "", "/h", []string{"keys", "list"}},
		{[]string{"keys", "add", "start"}, "", "/users/val/.exrpd", []string{"keys", "add", "start"}},
		// Shell completion (a TAB key): the words typed so far reach exrpd unchanged, the
		// home of the binary comes from them, and a half-typed --home falls back to the
		// default.
		{[]string{"__complete", "keys", "show", "", "--home", "/h"}, completeCommand, "/h", []string{"__complete", "keys", "show", "", "--home", "/h"}},
		{[]string{"__completeNoDesc", "--home", ""}, completeCommand, "/users/val/.exrpd", []string{"__completeNoDesc", "--home", ""}},
		// A shell alias puts --home or an empty word before the request.
		{[]string{"--home", "/h", "__complete", "ke"}, completeCommand, "/h", []string{"__complete", "ke"}},
		{[]string{"", "__complete", "ke"}, completeCommand, "/users/val/.exrpd", []string{"__complete", "ke"}},
		{[]string{"--home", "/h", "__complete", "keys", "--home", ""}, completeCommand, "/h", []string{"__complete", "keys", "--home", ""}},
		// An alias that adds other flags puts them before the request too. After a
		// command word, __complete is an ordinary argument.
		{[]string{"--log_level", "info", "__complete", "keys", ""}, completeCommand, "/users/val/.exrpd", []string{"--log_level", "info", "__complete", "keys", ""}},
		// Behind such an alias too, the typed --home reaches exrpd and a half-typed one falls
		// back to the default.
		{[]string{"--log_level", "info", "__complete", "start", "--home", "/h", ""}, completeCommand, "/h", []string{"--log_level", "info", "__complete", "start", "--home", "/h", ""}},
		{[]string{"--log_level", "info", "__complete", "start", "--home", ""}, completeCommand, "/users/val/.exrpd", []string{"--log_level", "info", "__complete", "start", "--home", ""}},
		// The word after --home is its value, also when it is __complete.
		{[]string{"--home", "/a", "--home", "__complete", "start"}, "start", filepath.Join(wd, "__complete"), nil},
		{[]string{"keys", "show", "__complete"}, "", "/users/val/.exrpd", []string{"keys", "show", "__complete"}},
		// "completion" writes a script and is an ordinary command: it never installs.
		{[]string{"--home", "/h", "completion", "--zsh"}, "", "/h", []string{"completion", "--zsh"}},
		// init installs on a new home. "init --help" is a help request, which goes to exrpd,
		// so it never installs.
		{[]string{"init", "val", "--chain-id", "c", "--home", "/h"}, initCommand, "/h", []string{"val", "--chain-id", "c"}},
		{[]string{"init", "--help"}, "", "/users/val/.exrpd", []string{"init", "--help"}},
		// cobra reads a boolean flag's value with strconv.ParseBool.
		{[]string{"init", "--help=true"}, "", "/users/val/.exrpd", []string{"init", "--help=true"}},
		{[]string{"init", "-h=1"}, "", "/users/val/.exrpd", []string{"init", "-h=1"}},
		{[]string{"init", "val", "--help=false"}, initCommand, "/users/val/.exrpd", []string{"val", "--help=false"}},
		{[]string{"init", "val", "--chain-id=-oh"}, initCommand, "/users/val/.exrpd", []string{"val", "--chain-id=-oh"}},
		// The launcher owns every argument of bundle: --home may come anywhere, and a help
		// flag before or after setup asks for the usage.
		{[]string{"bundle", "setup", "--home", "/h"}, bundleSetupCommand, "/h", nil},
		{[]string{"bundle", "--home", "/h", "setup"}, bundleSetupCommand, "/h", nil},
		{[]string{"--home", "/h", "bundle", "setup"}, bundleSetupCommand, "/h", nil},
		{[]string{"bundle", "--help"}, bundleCommand, "/users/val/.exrpd", nil},
		{[]string{"bundle", "-h", "setup"}, bundleCommand, "/users/val/.exrpd", nil},
		{[]string{"bundle", "setup", "--help"}, bundleCommand, "/users/val/.exrpd", nil},
		// The last --home wins, and an empty word before the command is skipped.
		{[]string{"--home", "-x", "bundle", "setup", "--home", "/h"}, bundleSetupCommand, "/h", nil},
		{[]string{"", "bundle", "setup"}, bundleSetupCommand, "/users/val/.exrpd", nil},
		// It owns every argument of version --bundle too: passed to exrpd, --bundle would
		// fail with "unknown flag". Other flags of version go to exrpd.
		{[]string{"--home=/h", "version", "--bundle"}, versionBundleCommand, "/h", nil},
		{[]string{"version", "--bundle", "--home", "/h"}, versionBundleCommand, "/h", nil},
		{[]string{"version"}, versionCommand, "/users/val/.exrpd", nil},
		{[]string{"version", "--long"}, "", "/users/val/.exrpd", []string{"version", "--long"}},
		// cobra skips empty words, so an empty word before start must not hide it.
		{[]string{"", "start", "--home", "/h"}, "start", "/h", nil},
		{[]string{"--home", "/h", "", "", "start"}, "start", "/h", nil},
		{[]string{"keys", "add", "val", "--note", ""}, "", "/users/val/.exrpd", []string{"keys", "add", "val", "--note", ""}},
		// A help request goes to exrpd like any other command, so it never installs. In
		// "start --moniker --help", --help is the moniker, so start still goes to Cosmovisor.
		{nil, "", "/users/val/.exrpd", nil},
		{[]string{"--home", "/h", "--help"}, "", "/h", []string{"--help"}},
		{[]string{"help", "keys"}, "", "/users/val/.exrpd", []string{"help", "keys"}},
		{[]string{"keys", "add", "-h"}, "", "/users/val/.exrpd", []string{"keys", "add", "-h"}},
		// --help never takes a value, so a --home after it is the home.
		{[]string{"keys", "--help", "--home", "/h"}, "", "/h", []string{"keys", "--help"}},
		{[]string{"keys", "-h", "--home=/h"}, "", "/h", []string{"keys", "-h"}},
		{[]string{"version", "--help"}, "", "/users/val/.exrpd", []string{"version", "--help"}},
		{[]string{"start", "--moniker", "--help"}, "start", "/users/val/.exrpd", []string{"--moniker", "--help"}},
	} {
		inv, err := parseCommand(tc.args, noEnv, userHomeDir)
		if err != nil || inv.command != tc.command || inv.home != tc.wantHome || strings.Join(inv.args, " ") != strings.Join(tc.wantRest, " ") {
			t.Errorf("parseCommand(%q) = %q, %q, %q, %v; want %q, %q, %q", tc.args, inv.home, inv.command, inv.args, err, tc.wantHome, tc.command, tc.wantRest)
		}
	}
	for _, args := range [][]string{
		{"--home", "/h", "--log_level", "info", "start"},
		// --help can be the value of --log_level, so this may be start.
		{"--log_level", "--help", "start"},
		// init after a flag may be init, which would not install the bundle.
		{"--log_level", "info", "init", "val"},
		{"--home", "/h", "init", "-oh"},
		{"init", "-ho"},
		{"init", "-oh=false"},
		// Not lenient like a TAB key: a refused --home must not fall back to the default home.
		{"completion", "--zsh", "--trace", "--home", "/h"},
		{"bundle", "setup", "--force"},
		{"bundle", "setup", "setup"},
		{"version", "--bundle", "--long"},
		{"version", "--long", "--bundle"},
		{"bundle", "setup", "--home", "--"},
		{"--home", "", "bundle", "setup"},
		{"version", "--bundle", "--home", "--help"},
	} {
		if _, err := parseCommand(args, noEnv, userHomeDir); err == nil {
			t.Errorf("parseCommand(%q) succeeded, want a refusal", args)
		}
	}
}

func TestTakeHome(t *testing.T) {
	value, rest, err := takeHome([]string{"--a=1", "--home", "x", "--b=2", "--home=y"}, mayTakeValue)
	if err != nil || value != "y" || !slices.Equal(rest, []string{"--a=1", "--b=2"}) {
		t.Errorf("takeHome() = %q, %q, %v; want last value y and other args in order", value, rest, err)
	}

	value, _, err = takeHome([]string{"--home", "-v", "--home=y"}, mayTakeValue)
	if err != nil || value != "y" {
		t.Errorf("takeHome() = %q, %v; want y, a value of --home is not a forwarded flag", value, err)
	}

	if value, _, _ := takeHome([]string{"--homes"}, mayTakeValue); value != "" {
		t.Error("takeHome() matched a flag that only shares the prefix")
	}
	for _, args := range [][]string{{"--home"}, {"--home="}, {"--home", ""}} {
		if _, _, err := takeHome(args, mayTakeValue); err == nil {
			t.Errorf("takeHome(%q) accepted a missing value", args)
		}
	}
}

func TestForwardedHome(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantValue string
		wantRest  []string
	}{
		{"flag", []string{"--home", "/a", "--x"}, "/a", []string{"--x"}},
		{"flag with equals", []string{"--x=1", "--home=/b"}, "/b", []string{"--x=1"}},
		{"after positional", []string{"-", "--home", "/c"}, "/c", []string{"-"}},
		{"none", []string{"--x"}, "", []string{"--x"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, rest, err := forwardedHome(tc.args)
			if err != nil || value != tc.wantValue || !slices.Equal(rest, tc.wantRest) {
				t.Errorf("forwardedHome() = %q, %q, %v; want %q, %q", value, rest, err, tc.wantValue, tc.wantRest)
			}
		})
	}
}

// The standalone exrpd opens the node's config and data on $EXRPD_HOME, else on
// $NODE_HOME. The bundle must pick the same home, after its own --home and $DAEMON_HOME.
// A different home would make a command like "comet unsafe-reset-all" wipe another data/.
// The source goes to the start log and to version --bundle, so an operator sees why.
func TestHomeFrom(t *testing.T) {
	userHomeDir := testUserHomeDir
	for _, tc := range []struct {
		value            string
		env              map[string]string
		want, wantSource string
	}{
		{"", nil, "/users/val/.exrpd", "default"},
		{"", map[string]string{"NODE_HOME": "/node"}, "/node", "$NODE_HOME"},
		{"", map[string]string{"NODE_HOME": "/node", "EXRPD_HOME": "/exrpd"}, "/exrpd", "$EXRPD_HOME"},
		{"", map[string]string{"NODE_HOME": "/node", "EXRPD_HOME": "/exrpd", "DAEMON_HOME": "/daemon"}, "/daemon", "$DAEMON_HOME"},
		{"/flag", map[string]string{"DAEMON_HOME": "/daemon"}, "/flag", "--home"},
	} {
		home, source, err := homeFrom(tc.value, func(k string) string { return tc.env[k] }, userHomeDir)
		if err != nil || home != tc.want || source != tc.wantSource {
			t.Errorf("homeFrom(%q) with %v = %q, %q, %v; want %q, %q", tc.value, tc.env, home, source, err, tc.want, tc.wantSource)
		}
	}
}

func TestHomeFromMakesRelativePathAbsolute(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	noEnv := func(string) string { return "" }
	home, _, err := homeFrom("rel", noEnv, os.UserHomeDir)
	if err != nil || home != filepath.Join(wd, "rel") {
		t.Errorf("homeFrom() = %q, %v; want %q", home, err, filepath.Join(wd, "rel"))
	}
}

// exrpd reads "--moniker --" as the moniker "--". A launcher that stopped at "--" would
// miss the --home that exrpd uses, and Cosmovisor would watch another home.
func TestForwardedHomeRefusesTerminator(t *testing.T) {
	for _, args := range [][]string{
		{"--moniker", "--", "--home", "/srv/exrpd"},
		{"--home", "/a", "--"},
	} {
		if value, _, err := forwardedHome(args); err == nil {
			t.Errorf("forwardedHome(%q) = %q, want an error for --", args, value)
		}
	}
}

// exrpd reads "--moniker --home=/srv/name" as the moniker "--home=/srv/name". The
// launcher must not select /srv/name as the home.
func TestForwardedHomeRefusesFlagAfterBareFlag(t *testing.T) {
	for _, args := range [][]string{
		{"--moniker", "--home=/srv/name"},
		{"--moniker", "--home", "/srv/name"},
		{"--trace", "--home", "/srv/name"},
		{"-m", "--home=/srv/name"},
	} {
		if value, _, err := forwardedHome(args); err == nil {
			t.Errorf("forwardedHome(%q) = %q, want an error for a flag that can take --home as its value", args, value)
		}
	}
}

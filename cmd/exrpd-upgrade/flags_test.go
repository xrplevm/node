package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestTakeHome(t *testing.T) {
	value, rest, err := takeHome([]string{"--a=1", "--home", "x", "--b=2", "--home=y"})
	if err != nil || value != "y" || !slices.Equal(rest, []string{"--a=1", "--b=2"}) {
		t.Errorf("takeHome() = %q, %q, %v; want last value y and other args in order", value, rest, err)
	}

	value, _, err = takeHome([]string{"--home", "-v", "--home=y"})
	if err != nil || value != "y" {
		t.Errorf("takeHome() = %q, %v; want y, a value of --home is not a forwarded flag", value, err)
	}

	if value, _, _ := takeHome([]string{"--homes"}); value != "" {
		t.Error("takeHome() matched a flag that only shares the prefix")
	}
	for _, args := range [][]string{{"--home"}, {"--home="}, {"--home", ""}} {
		if _, _, err := takeHome(args); err == nil {
			t.Errorf("takeHome(%q) accepted a missing value", args)
		}
	}
}

func TestResolveHome(t *testing.T) {
	userHomeDir := testUserHomeDir
	for _, tc := range []struct {
		name      string
		args      []string
		daemonEnv string
		wantHome  string
		wantRest  []string
	}{
		{"flag wins over env", []string{"--home", "/a", "--x"}, "/env", "/a", []string{"--x"}},
		{"flag with equals", []string{"--x=1", "--home=/b"}, "/env", "/b", []string{"--x=1"}},
		{"after positional", []string{"-", "--home", "/c"}, "/env", "/c", []string{"-"}},
		{"DAEMON_HOME", []string{"--x"}, "/env", "/env", []string{"--x"}},
		{"after the launcher's --bundle", []string{"version", "--bundle", "--home", "/d"}, "/env", "/d", []string{"version", "--bundle"}},
		{"default", nil, "", "/users/val/.exrpd", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(string) string { return tc.daemonEnv }
			home, _, rest, err := resolveHome(tc.args, getenv, userHomeDir)
			if err != nil || home != tc.wantHome || !slices.Equal(rest, tc.wantRest) {
				t.Errorf("resolveHome() = %q, %q, %v; want %q, %q", home, rest, err, tc.wantHome, tc.wantRest)
			}
		})
	}
}

// The standalone exrpd opens the node's config and data on $EXRPD_HOME, else on
// $NODE_HOME. The bundle must pick the same home, after its own --home and $DAEMON_HOME.
// A different home would make a command like "comet unsafe-reset-all" wipe another data/.
// The source goes to the start log and to version --bundle, so an operator sees why.
func TestResolveHomeEnv(t *testing.T) {
	userHomeDir := testUserHomeDir
	for _, tc := range []struct {
		args             []string
		env              map[string]string
		want, wantSource string
	}{
		{nil, nil, "/users/val/.exrpd", "default"},
		{nil, map[string]string{"NODE_HOME": "/node"}, "/node", "$NODE_HOME"},
		{nil, map[string]string{"NODE_HOME": "/node", "EXRPD_HOME": "/exrpd"}, "/exrpd", "$EXRPD_HOME"},
		{nil, map[string]string{"NODE_HOME": "/node", "EXRPD_HOME": "/exrpd", "DAEMON_HOME": "/daemon"}, "/daemon", "$DAEMON_HOME"},
		{[]string{"--home", "/flag"}, map[string]string{"DAEMON_HOME": "/daemon"}, "/flag", "--home"},
	} {
		home, source, _, err := resolveHome(tc.args, func(k string) string { return tc.env[k] }, userHomeDir)
		if err != nil || home != tc.want || source != tc.wantSource {
			t.Errorf("resolveHome(%q) with %v = %q, %q, %v; want %q, %q", tc.args, tc.env, home, source, err, tc.want, tc.wantSource)
		}
	}
}

func TestResolveHomeMakesRelativePathAbsolute(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	noEnv := func(string) string { return "" }
	home, _, _, err := resolveHome([]string{"--home", "rel"}, noEnv, os.UserHomeDir)
	if err != nil || home != filepath.Join(wd, "rel") {
		t.Errorf("resolveHome() = %q, %v; want %q", home, err, filepath.Join(wd, "rel"))
	}
}

// exrpd reads "--moniker --" as the moniker "--". A launcher that stopped at "--" would
// miss the --home that exrpd uses, and Cosmovisor would watch another home.
func TestResolveHomeRefusesTerminator(t *testing.T) {
	noEnv := func(string) string { return "" }
	for _, args := range [][]string{
		{"--moniker", "--", "--home", "/srv/exrpd"},
		{"--home", "/a", "--"},
	} {
		if home, _, _, err := resolveHome(args, noEnv, os.UserHomeDir); err == nil {
			t.Errorf("resolveHome(%q) = %q, want an error for --", args, home)
		}
	}
}

// exrpd reads "--moniker --home=/srv/name" as the moniker "--home=/srv/name". The
// launcher must not select /srv/name as the home.
func TestResolveHomeRefusesFlagAfterBareFlag(t *testing.T) {
	noEnv := func(string) string { return "" }
	for _, args := range [][]string{
		{"--moniker", "--home=/srv/name"},
		{"--moniker", "--home", "/srv/name"},
		{"--trace", "--home", "/srv/name"},
		{"-m", "--home=/srv/name"},
	} {
		if home, _, _, err := resolveHome(args, noEnv, os.UserHomeDir); err == nil {
			t.Errorf("resolveHome(%q) = %q, want an error for a flag that can take --home as its value", args, home)
		}
	}
}

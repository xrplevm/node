package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestTakeFlag(t *testing.T) {
	value, rest, found, err := takeFlag([]string{"--a", "--select", "x", "--b", "--select=y"}, "--select")
	if err != nil || !found || value != "y" || !slices.Equal(rest, []string{"--a", "--b"}) {
		t.Errorf("takeFlag() = %q, %q, %v, %v; want last value y and other args in order", value, rest, found, err)
	}

	if _, _, found, _ := takeFlag([]string{"--selected"}, "--select"); found {
		t.Error("takeFlag() matched a flag that only shares the prefix")
	}
	for _, args := range [][]string{{"--select"}, {"--select="}, {"--select", ""}} {
		if _, _, _, err := takeFlag(args, "--select"); err == nil {
			t.Errorf("takeFlag(%q) accepted a missing value", args)
		}
	}
}

func TestResolveHome(t *testing.T) {
	userHomeDir := func() (string, error) { return "/users/val", nil }
	for _, tc := range []struct {
		name      string
		args      []string
		daemonEnv string
		wantHome  string
		wantRest  []string
	}{
		{"flag wins over env", []string{"--home", "/a", "--x"}, "/env", "/a", []string{"--x"}},
		{"flag with equals", []string{"--x", "--home=/b"}, "/env", "/b", []string{"--x"}},
		{"DAEMON_HOME", []string{"--x"}, "/env", "/env", []string{"--x"}},
		{"default", nil, "", "/users/val/.exrpd", []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(string) string { return tc.daemonEnv }
			home, rest, err := resolveHome(tc.args, getenv, userHomeDir)
			if err != nil || home != tc.wantHome || !slices.Equal(rest, tc.wantRest) {
				t.Errorf("resolveHome() = %q, %q, %v; want %q, %q", home, rest, err, tc.wantHome, tc.wantRest)
			}
		})
	}
}

func TestResolveHomeMakesRelativePathAbsolute(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	noEnv := func(string) string { return "" }
	home, _, err := resolveHome([]string{"--home", "rel"}, noEnv, os.UserHomeDir)
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
		if home, _, err := resolveHome(args, noEnv, os.UserHomeDir); err == nil {
			t.Errorf("resolveHome(%q) = %q, want an error for --", args, home)
		}
	}
}

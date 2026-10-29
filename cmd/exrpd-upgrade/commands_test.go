package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every command runs the exrpd that runs the node data, picked by the same rule as
// start, and never moves `current`: the node may be running.
func TestExrpdFor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup homeSetup
		want  string // directory of the exrpd, relative to <home>/cosmovisor
	}{
		{"new home runs genesis", homeSetup{}, genesisName},
		{"genesis plan", homeSetup{current: genesisName, data: upgradeInfoJSON(testGenesisPlan)}, testGenesisDir},
		// Killed at the halt: `current` is stale, the data file names the upgrade.
		{"stale current", homeSetup{current: testGenesisDir, data: upgradeInfoJSON(testPlan)}, testUpgradeDir},
		{"resync keeps the upgrade", homeSetup{current: testUpgradeDir}, testUpgradeDir},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := tc.setup.build(t)

			got, err := exrpdFor(files, home, true)
			if err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(home, "cosmovisor", tc.want, "bin", "exrpd")
			if got != want {
				t.Errorf("exrpdFor() = %s, want %s", got, want)
			}
			if readFile(t, got) == "" {
				t.Errorf("%s was not installed", got)
			}
			if target := currentTarget(t, home); target != tc.setup.current {
				t.Errorf("current -> %q, want it unchanged at %q", target, tc.setup.current)
			}
		})
	}
}

func TestExrpdForRefusal(t *testing.T) {
	files, _ := testPayload(t)
	home := homeSetup{current: genesisName, data: upgradeInfoJSON("v12.1.0")}.build(t)
	if bin, err := exrpdFor(files, home, true); err == nil || !strings.Contains(err.Error(), "names plan v12.1.0") {
		t.Fatalf("exrpdFor() = %s, %v; want a refusal", bin, err)
	}
	if got := readFile(t, filepath.Join(home, "cosmovisor", genesisName, "bin", "exrpd")); got != "" {
		t.Errorf("a refused command installed genesis")
	}
}

// A tab key on a machine without a node home must not install the bundle: completion
// only uses an exrpd that is already there.
func TestExrpdForWithoutInstall(t *testing.T) {
	files, _ := testPayload(t)
	home := t.TempDir()
	if bin, err := exrpdFor(files, home, false); err != nil || bin != "" {
		t.Fatalf("exrpdFor() = %q, %v; want no binary and no error", bin, err)
	}
	if _, err := os.Stat(filepath.Join(home, "cosmovisor")); !os.IsNotExist(err) {
		t.Errorf("completion installed files (stat error %v)", err)
	}
}

// With stock Cosmovisor, bin/exrpd is often a symlink to /usr/local/bin/exrpd, where the
// bundle is installed too. Running it would run the launcher again without end, and
// version would start one process after another.
func TestRefuseSelf(t *testing.T) {
	files, _ := testPayload(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	home := homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	if err := os.MkdirAll(filepath.Join(home, "cosmovisor", testUpgradeDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(self, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd")); err != nil {
		t.Fatal(err)
	}
	if _, err := exrpdFor(files, home, true); err == nil || !strings.Contains(err.Error(), "is this exrpd bundle") {
		t.Errorf("exrpdFor() error = %v, want a refusal", err)
	}
	if _, err := versionOutput(files, home, "--home", false); err == nil || !strings.Contains(err.Error(), "is this exrpd bundle") {
		t.Errorf("versionOutput() error = %v, want a refusal", err)
	}
}

// --bundle belongs to the launcher anywhere after version. Passed to exrpd, it would fail
// with "unknown flag".
func TestVersionMode(t *testing.T) {
	for _, tc := range []struct {
		args          []string
		bundle, fails bool
	}{
		{nil, false, false},
		{[]string{"--long"}, false, false},
		{[]string{"--bundle"}, true, false},
		{[]string{"--bundle", "--long"}, false, true},
		{[]string{"--long", "--bundle"}, false, true},
	} {
		bundle, err := versionMode(tc.args)
		if bundle != tc.bundle || (err != nil) != tc.fails {
			t.Errorf("versionMode(%q) = %v, %v; want %v, error %v", tc.args, bundle, err, tc.bundle, tc.fails)
		}
	}
}

// version names the exrpd that runs the node data first, so a script that reads the
// first word sees the same version as with the standalone exrpd.
func TestVersionOutput(t *testing.T) {
	files, _ := testPayload(t)
	for _, tc := range []struct {
		name  string
		setup homeSetup
		want  string
	}{
		{"before the upgrade", homeSetup{}, "11.3.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
		{"after the upgrade", homeSetup{data: upgradeInfoJSON(testPlan)}, "12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := tc.setup.build(t)
			got, err := versionOutput(files, home, "--home", false)
			if err != nil || got != tc.want {
				t.Errorf("versionOutput() = %q, %v; want %q", got, err, tc.want)
			}
			if readFile(t, filepath.Join(home, "cosmovisor", genesisName, "bin", "exrpd")) != "" {
				t.Error("version installed files")
			}
		})
	}
}

// version --bundle tells an operator which home is used, why, and which exrpd commands
// run there. It never runs that exrpd, so it works also when that binary is broken.
func TestVersionBundle(t *testing.T) {
	files, m := testPayload(t)
	type output struct {
		Version    string
		Home       string
		HomeSource string   `json:"home_source"`
		WillRun    *willRun `json:"will_run"`
		Refused    *string
		Dirs       []string
		Manifest   Manifest
	}
	bundleJSON := func(t *testing.T, home string) (output, error) {
		t.Helper()
		raw, err := versionOutput(files, home, "$DAEMON_HOME", true)
		var got output
		if jsonErr := json.Unmarshal([]byte(raw), &got); jsonErr != nil {
			t.Fatalf("version --bundle is not JSON: %v\n%s", jsonErr, raw)
		}
		return got, err
	}

	home := t.TempDir()
	got, err := bundleJSON(t, home)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(home, "cosmovisor")
	wantDirs := []string{
		filepath.Join(root, genesisName),
		filepath.Join(root, testGenesisDir),
		filepath.Join(root, testUpgradeDir),
		filepath.Join(root, "supervisor", m.Cosmovisor.SHA256[:16]),
	}
	wantRun := willRun{genesisName, filepath.Join(root, genesisName, "bin", "exrpd"), "11.3.0", false, "no data/upgrade-info.json and no current: genesis"}
	if got.Version != "11.3.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3" || got.Home != home || got.HomeSource != "$DAEMON_HOME" ||
		got.WillRun == nil || *got.WillRun != wantRun || got.Refused != nil ||
		strings.Join(got.Dirs, " ") != strings.Join(wantDirs, " ") || got.Manifest.Upgrades[0].Plan != testPlan {
		t.Errorf("version --bundle = %+v, want will_run %+v and directories %q", got, wantRun, wantDirs)
	}

	// Another exrpd in the directory: a script that fails if it runs.
	home = homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	writeFile(t, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd"), "#!/bin/sh\nexit 1\n", 0o755)
	got, err = bundleJSON(t, home)
	if err != nil || got.WillRun == nil || got.WillRun.Version != "unknown" || !got.WillRun.Installed || got.WillRun.Dir != testUpgradeDir {
		t.Errorf("version --bundle with another exrpd = %+v, %v; want will_run version unknown, installed", got.WillRun, err)
	}

	home = homeSetup{current: genesisName, data: upgradeInfoJSON("v12.1.0")}.build(t)
	got, err = bundleJSON(t, home)
	if err == nil || got.Refused == nil || *got.Refused != err.Error() || got.WillRun != nil || got.Home != home {
		t.Errorf("version --bundle on a refused home = %+v, %v; want the refusal in the JSON and as the error", got, err)
	}
}

// After a patch bundle is installed and before the node restarts, the home still has the
// earlier binary, and commands run it. version must name that one, not the bundled patch.
func TestVersionOutputBeforeRestart(t *testing.T) {
	files, m := testPayload(t)
	m.Upgrades[0].Version = "12.0.1"
	setManifest(t, files, m)
	home := homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	// The binary answers only when it gets this home, as every child of the launcher must.
	writeFile(t, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd"),
		"#!/bin/sh\n[ \"$1 $2 $3\" = \"--home "+home+" version\" ] && echo 12.0.0\n", 0o755)

	got, err := versionOutput(files, home, "--home", false)
	if want := "12.0.0 - genesis 11.3.0 - upgrade 12.0.1 - cosmovisor 1.7.3"; err != nil || got != want {
		t.Errorf("versionOutput() = %q, %v; want %q", got, err, want)
	}
}

// A bundle with several upgrades, as a bundle for sync from genesis would carry, installs
// every binary in its plan directory and runs each plan's data with its own binary.
func TestBundleWithTwoUpgrades(t *testing.T) {
	files, m := testPayload(t)
	later := m.Upgrades[0]
	later.Plan, later.Version, later.File = "v12.1.0", "12.1.0", "exrpd-later.gz"
	files[later.File] = files[m.Upgrades[0].File]
	m.Upgrades = append(m.Upgrades, later)
	setManifest(t, files, m)

	for plan, wantVersion := range map[string]string{testPlan: "12.0.0", "v12.1.0": "12.1.0"} {
		home := homeSetup{current: genesisName, data: upgradeInfoJSON(plan)}.build(t)
		if _, err := prepare(files, home); err != nil {
			t.Fatal(err)
		}
		if got := currentTarget(t, home); got != planDir(plan) {
			t.Errorf("data for %s: current -> %q, want %q", plan, got, planDir(plan))
		}
		for _, dir := range []string{testGenesisDir, testUpgradeDir, "upgrades/v12.1.0"} {
			if readFile(t, filepath.Join(home, "cosmovisor", dir, "bin", "exrpd")) == "" {
				t.Errorf("%s/bin/exrpd not installed", dir)
			}
		}
		line, err := versionOutput(files, home, "--home", false)
		if want := wantVersion + " - genesis 11.3.0 - upgrade 12.0.0, 12.1.0 - cosmovisor 1.7.3"; err != nil || line != want {
			t.Errorf("data for %s: version = %q, %v; want %q", plan, line, err, want)
		}
	}
}

// "start" must reach Cosmovisor wherever --home is, as with the standalone exrpd. A
// "start" that went to exrpd would run the node without the upgrade switch.
func TestParseCommand(t *testing.T) {
	noEnv := func(string) string { return "" }
	userHomeDir := testUserHomeDir
	for _, tc := range []struct {
		args              []string
		command, wantHome string
		wantRest          []string
	}{
		{[]string{"start", "--home", "/h", "--x=1"}, "start", "/h", []string{"--x=1"}},
		{[]string{"--home", "/h", "start", "--x=1"}, "start", "/h", []string{"--x=1"}},
		{[]string{"--home=/h", "version", "--bundle"}, "version", "/h", []string{"--bundle"}},
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
		// "completion" writes a script and may install, so it is an ordinary command.
		{[]string{"--home", "/h", "completion", "--zsh"}, "", "/h", []string{"completion", "--zsh"}},
		// cobra skips empty words, so an empty word before start must not hide it.
		{[]string{"", "start", "--home", "/h"}, "start", "/h", nil},
		{[]string{"--home", "/h", "", "", "start"}, "start", "/h", nil},
		{[]string{"keys", "add", "val", "--note", ""}, "", "/users/val/.exrpd", []string{"keys", "add", "val", "--note", ""}},
		// Help never installs. In "start --moniker --help", --help is the moniker, so start
		// still goes to Cosmovisor.
		{nil, helpCommand, "/users/val/.exrpd", nil},
		{[]string{"--home", "/h", "--help"}, helpCommand, "/h", []string{"--help"}},
		{[]string{"help", "keys"}, helpCommand, "/users/val/.exrpd", []string{"help", "keys"}},
		{[]string{"keys", "add", "-h"}, helpCommand, "/users/val/.exrpd", []string{"keys", "add", "-h"}},
		{[]string{"version", "--help"}, helpCommand, "/users/val/.exrpd", []string{"version", "--help"}},
		{[]string{"start", "--moniker", "--help"}, "start", "/users/val/.exrpd", []string{"--moniker", "--help"}},
	} {
		home, _, command, rest, err := parseCommand(tc.args, noEnv, userHomeDir)
		if err != nil || command != tc.command || home != tc.wantHome || strings.Join(rest, " ") != strings.Join(tc.wantRest, " ") {
			t.Errorf("parseCommand(%q) = %q, %q, %q, %v; want %q, %q, %q", tc.args, home, command, rest, err, tc.wantHome, tc.command, tc.wantRest)
		}
	}
	for _, args := range [][]string{
		{"--home", "/h", "--log_level", "info", "start"},
		// --help can be the value of --log_level, so this may be start: never help.
		{"--log_level", "--help", "start"},
		// Not lenient like a TAB key: a refused --home must not install into the default home.
		{"completion", "--zsh", "--trace", "--home", "/h"},
	} {
		if _, _, _, _, err := parseCommand(args, noEnv, userHomeDir); err == nil {
			t.Errorf("parseCommand(%q) succeeded, want a refusal", args)
		}
	}
}

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
		{"new home runs the genesis plan", homeSetup{}, testGenesisDir},
		{"genesis plan", homeSetup{current: genesisName, data: upgradeInfoJSON(testGenesisPlan)}, testGenesisDir},
		// Killed at the halt: `current` is stale, the data file names the upgrade.
		{"stale current", homeSetup{current: testGenesisDir, data: upgradeInfoJSON(testPlan)}, testUpgradeDir},
		{"resync keeps the upgrade", homeSetup{current: testUpgradeDir}, testUpgradeDir},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := tc.setup.build(t)
			installAs(t, b, home, installMatch)

			got, err := b.exrpdFor(testNode(b, home))
			if err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(home, "cosmovisor", tc.want, "bin", "exrpd")
			if got != want {
				t.Errorf("exrpdFor() = %s, want %s", got, want)
			}
			if target := currentTarget(t, home); target != tc.setup.current {
				t.Errorf("current -> %q, want it unchanged at %q", target, tc.setup.current)
			}
		})
	}
}

func TestExrpdForRefusal(t *testing.T) {
	b := testBundle(t)
	home := homeSetup{current: genesisName, data: upgradeInfoJSON("v12.1.0")}.build(t)
	if bin, err := b.exrpdFor(testNode(b, home)); err == nil || !strings.Contains(err.Error(), "names plan v12.1.0") {
		t.Fatalf("exrpdFor() = %s, %v; want a refusal", bin, err)
	}
	if got := readFile(t, filepath.Join(home, "cosmovisor", testGenesisDir, "bin", "exrpd")); got != "" {
		t.Errorf("a refused command installed genesis")
	}
}

// Commands other than the installers run the exrpd of the home only when this bundle
// installed it. Otherwise they refuse with the next step for that home: client commands
// and help fail with it, and completion suggests nothing.
func TestExrpdForMarker(t *testing.T) {
	notInstalled := "The bundle is not installed in <home> (from --home). Restart the node service, or run `exrpd bundle setup --home <home>` while the node is stopped."
	for _, tc := range []struct {
		name    string
		setup   homeSetup
		genesis bool         // config/genesis.json exists
		install installState // marker after the install, "" for nothing installed
		deleted bool         // the genesis plan's exrpd deleted after the install
		want    string       // the refusal, "" to run the genesis plan's exrpd
	}{
		{"installed", homeSetup{}, false, installMatch, false, ""},
		{"nothing installed", homeSetup{}, false, "", false, "No exrpd is installed in <home> (from --home). New node: run `exrpd init <moniker> --chain-id <id> --home <home>`. Existing node: this is probably not its home, pass --home."},
		{"standalone exrpd home", homeSetup{data: upgradeInfoJSON(testPlan)}, false, "", false, notInstalled},
		{"node without marker", homeSetup{}, true, installMissing, false, notInstalled},
		{"installed exrpd deleted", homeSetup{}, true, installMatch, true, "The bundle is installed in <home> (from --home), but <home>/cosmovisor/" + testGenesisDir + "/bin/exrpd is missing. Restart the node service, or run `exrpd bundle setup --home <home>` while the node is stopped."},
		{"another bundle installed", homeSetup{}, true, installDiffers, false, "The bundle installed in <home> (from --home) differs from this exrpd. Restart the node service, or run `exrpd bundle setup --home <home>` while the node is stopped."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := tc.setup.build(t)
			if tc.genesis {
				writeFile(t, filepath.Join(home, "config", "genesis.json"), "{}", 0o600)
			}
			if tc.install != "" {
				installAs(t, b, home, tc.install)
			}
			genesisBin := filepath.Join(home, "cosmovisor", testGenesisDir, "bin", "exrpd")
			if tc.deleted {
				if err := os.Remove(genesisBin); err != nil {
					t.Fatal(err)
				}
			}
			marker := filepath.Join(home, "cosmovisor", "bundle.sha256")
			before := readFile(t, marker)

			bin, err := b.exrpdFor(testNode(b, home))
			if tc.want == "" {
				if err != nil || bin != genesisBin {
					t.Fatalf("exrpdFor() = %q, %v; want %s", bin, err, genesisBin)
				}
				return
			}
			var refusal *homeRefusal
			if want := strings.ReplaceAll(tc.want, "<home>", home); !errors.As(err, &refusal) || err.Error() != want {
				t.Errorf("exrpdFor() = %q, %v; want the refusal %q", bin, err, want)
			}
			if got := readFile(t, marker); got != before {
				t.Errorf("marker = %q, want it unchanged at %q", got, before)
			}
		})
	}
}

// With stock Cosmovisor, bin/exrpd is often a symlink to /usr/local/bin/exrpd, where the
// bundle is installed too. Running it would run the launcher again without end, and
// version would start one process after another.
func TestRefuseSelf(t *testing.T) {
	b := testBundle(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	home := homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	installAs(t, b, home, installMatch)
	bin := filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd")
	if err := os.Remove(bin); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(self, bin); err != nil {
		t.Fatal(err)
	}
	if _, err := b.exrpdFor(testNode(b, home)); err == nil || !strings.Contains(err.Error(), "is this exrpd bundle") {
		t.Errorf("exrpdFor() error = %v, want a refusal", err)
	}
	if _, err := b.versionLine(testNode(b, home)); err == nil || !strings.Contains(err.Error(), "is this exrpd bundle") {
		t.Errorf("versionLine() error = %v, want a refusal", err)
	}
}

// Stock Cosmovisor runs bin/exrpd of the directory `current` selects, with <root> as its
// working directory. When that is this bundle, start would install the bundle and replace
// `current` under the Cosmovisor that runs it. start refuses before it writes, whichever
// path to the slot names it.
func TestStartRefusesSelectedSlot(t *testing.T) {
	b := testBundle(t)
	for _, argv0 := range []func(n node) string{
		func(n node) string { return n.bin(genesisName) },
		func(node) string { return filepath.Join("current", "bin", "exrpd") },
	} {
		n := slotHome(t, b)
		if err := startAs(t, b, n, argv0(n), n.Root); err == nil || !strings.Contains(err.Error(), "external Cosmovisor") {
			t.Errorf("start as %s from the Cosmovisor root: error %v, want the refusal", argv0(n), err)
		}
		if got := currentTarget(t, n.Home); got != genesisName {
			t.Errorf("current = %q, want %q", got, genesisName)
		}
		if _, err := os.Stat(n.Marker()); !os.IsNotExist(err) {
			t.Errorf("start wrote the marker (stat error %v)", err)
		}
	}
}

// The installed exrpd starts from the Cosmovisor root, as after cd into an old stock
// home, and the slot starts from elsewhere: neither is a Cosmovisor running the slot.
func TestStartOutsideSelectedSlot(t *testing.T) {
	b := testBundle(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		argv0 func(n node) string
		cwd   func(n node) string
	}{
		{"installed path from the root", func(node) string { return self }, func(n node) string { return n.Root }},
		{"slot from the home", func(n node) string { return n.bin(genesisName) }, func(n node) string { return n.Home }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := slotHome(t, b)
			if err := startAs(t, b, n, tc.argv0(n), tc.cwd(n)); err == nil || strings.Contains(err.Error(), "external Cosmovisor") {
				t.Errorf("start error = %v, want the exec of the test Cosmovisor to fail", err)
			}
			if _, err := os.Stat(n.Marker()); err != nil {
				t.Errorf("start did not install the bundle: %v", err)
			}
		})
	}
}

// slotHome returns a stock Cosmovisor home whose `current` selects genesis, with
// genesis/bin/exrpd a symlink to this bundle.
func slotHome(t *testing.T, b bundle) node {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	n := testNode(b, homeSetup{current: genesisName}.build(t))
	if err := os.MkdirAll(filepath.Dir(n.bin(genesisName)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(self, n.bin(genesisName)); err != nil {
		t.Fatal(err)
	}
	return n
}

// startAs runs start for n as if invoked as argv0 from cwd.
func startAs(t *testing.T, b bundle, n node, argv0, cwd string) error {
	t.Helper()
	for _, v := range requiredEnv {
		t.Setenv(v.key, v.value)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	args := os.Args
	os.Args = append([]string{argv0}, args[1:]...)
	t.Cleanup(func() {
		os.Args = args
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
	return runStart(b, n, nil)
}

// version names the exrpd that runs the node data first, so a script that reads the
// first word sees the same version as with the standalone exrpd. Until this bundle is
// installed, or until the restart that installs it, the first word says so instead.
func TestVersionLine(t *testing.T) {
	b, single, multi := testBundle(t), testSingleBundle(t), testMultiBundle(t)
	for _, tc := range []struct {
		name    string
		b       bundle
		setup   homeSetup
		install installState // marker after the install, "" for nothing installed
		want    string
	}{
		{"before the upgrade", b, homeSetup{}, installMatch, "11.3.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
		{"after the upgrade", b, homeSetup{data: upgradeInfoJSON(testPlan)}, installMatch, "12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
		{"standalone exrpd home", b, homeSetup{data: upgradeInfoJSON(testPlan)}, "", "not-installed - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
		{"another bundle installed", b, homeSetup{data: upgradeInfoJSON(testPlan)}, installDiffers, "restart-needed - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"},
		// A single bundle has no upgrade part.
		{"single bundle on a new home", single, homeSetup{}, installMatch, "12.0.1 - genesis 12.0.1 - cosmovisor 1.7.3"},
		{"single bundle on its plan", single, homeSetup{data: upgradeInfoJSON(testPlan)}, installMatch, "12.0.1 - genesis 12.0.1 - cosmovisor 1.7.3"},
		{"single bundle not installed", single, homeSetup{data: upgradeInfoJSON(testPlan)}, "", "not-installed - genesis 12.0.1 - cosmovisor 1.7.3"},
		// A bundle with more than one upgrade lists them in plan order.
		{"multi-plan bundle on the middle plan", multi, homeSetup{data: upgradeInfoJSON(testMidPlan)}, installMatch, "11.4.0 - genesis 11.3.0 - upgrade 11.4.0 12.0.0 - cosmovisor 1.7.3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := tc.setup.build(t)
			if tc.install != "" {
				installAs(t, tc.b, home, tc.install)
			}
			got, err := tc.b.versionLine(testNode(tc.b, home))
			if err != nil || got != tc.want {
				t.Errorf("versionLine() = %q, %v; want %q", got, err, tc.want)
			}
			if _, err := os.Stat(filepath.Join(home, "cosmovisor")); tc.install == "" && !os.IsNotExist(err) {
				t.Errorf("version wrote to the home (stat error %v)", err)
			}
		})
	}
}

// version --bundle tells an operator which home is used, why, and which exrpd commands
// run there. It never runs that exrpd, so it works also when that binary is broken.
func TestVersionBundle(t *testing.T) {
	b := testBundle(t)
	type output struct {
		Version         string
		Home            string
		HomeSource      string   `json:"home_source"`
		WillRun         *willRun `json:"will_run"`
		Refused         *string
		Dirs            []string
		Manifest        Manifest
		InstalledBundle string `json:"installed_bundle"`
	}
	bundleJSON := func(t *testing.T, home string) (output, error) {
		t.Helper()
		raw, err := b.bundleReport(node{Layout: newLayout(home, b.m), Source: "$DAEMON_HOME"})
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
	if _, err := os.Stat(filepath.Join(home, "cosmovisor")); !os.IsNotExist(err) {
		t.Errorf("version --bundle wrote to the home (stat error %v)", err)
	}
	root := filepath.Join(home, "cosmovisor")
	wantDirs := []string{
		filepath.Join(root, testGenesisDir),
		filepath.Join(root, testUpgradeDir),
		filepath.Join(root, "supervisor", b.m.Cosmovisor.SHA256[:16]),
	}
	wantRun := willRun{testGenesisDir, filepath.Join(root, testGenesisDir, "bin", "exrpd"), "11.3.0", false, "no data/upgrade-info.json and no current: genesis plan " + testGenesisPlan}
	if got.Version != "11.3.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3" || got.Home != home || got.HomeSource != "$DAEMON_HOME" ||
		got.WillRun == nil || *got.WillRun != wantRun || got.Refused != nil || got.InstalledBundle != "none" ||
		strings.Join(got.Dirs, " ") != strings.Join(wantDirs, " ") || len(got.Manifest.Upgrades) != 1 || got.Manifest.Upgrades[0].Plan != testPlan ||
		got.Manifest.Launcher.Commit != testCommit {
		t.Errorf("version --bundle = %+v, want will_run %+v and directories %q", got, wantRun, wantDirs)
	}

	// Another exrpd in the directory: a script that fails if it runs.
	home = homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	writeFile(t, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd"), "#!/bin/sh\nexit 1\n", 0o755)
	got, err = bundleJSON(t, home)
	if err != nil || got.WillRun == nil || got.WillRun.Version != "unknown" || !got.WillRun.Exists || got.WillRun.Dir != testUpgradeDir {
		t.Errorf("version --bundle with another exrpd = %+v, %v; want will_run version unknown, exists", got.WillRun, err)
	}

	home = homeSetup{current: genesisName, data: upgradeInfoJSON("v12.1.0")}.build(t)
	writeFile(t, filepath.Join(home, "cosmovisor", "bundle.sha256"), b.digest+"\n", 0o644)
	got, err = bundleJSON(t, home)
	if err == nil || got.Refused == nil || *got.Refused != err.Error() || got.WillRun != nil || got.Home != home || got.InstalledBundle != "this" {
		t.Errorf("version --bundle on a refused home = %+v, %v; want the refusal in the JSON and as the error, and installed_bundle from the marker", got, err)
	}

	// installed_bundle is the bundle that the marker names, also when the exrpd that
	// commands run is missing: will_run.exists tells that.
	for _, tc := range []struct {
		state      installState
		removeBin  bool
		wantBundle string
		wantExists bool
	}{
		{installMatch, false, "this", true},
		{installDiffers, false, "other", true},
		{installMatch, true, "this", false},
	} {
		home := t.TempDir()
		installAs(t, b, home, tc.state)
		if tc.removeBin {
			if err := os.Remove(filepath.Join(home, "cosmovisor", testGenesisDir, "bin", "exrpd")); err != nil {
				t.Fatal(err)
			}
		}
		got, err := bundleJSON(t, home)
		if err != nil || got.InstalledBundle != tc.wantBundle || got.WillRun == nil || got.WillRun.Exists != tc.wantExists || got.Refused != nil {
			t.Errorf("version --bundle with the marker %s, binary removed %v = %+v, %v; want installed_bundle %q, exists %v", tc.state, tc.removeBin, got, err, tc.wantBundle, tc.wantExists)
		}
	}

	// A single bundle prints its manifest as is: no upgrades key, and the launcher commit.
	single := testSingleBundle(t)
	home = t.TempDir()
	raw, err := single.bundleReport(testNode(single, home))
	if err != nil {
		t.Fatal(err)
	}
	var singleOut struct {
		Version  string
		Dirs     []string
		Manifest map[string]json.RawMessage
	}
	if err := json.Unmarshal([]byte(raw), &singleOut); err != nil {
		t.Fatalf("version --bundle is not JSON: %v\n%s", err, raw)
	}
	root = filepath.Join(home, "cosmovisor")
	wantDirs = []string{filepath.Join(root, testUpgradeDir), filepath.Join(root, "supervisor", single.m.Cosmovisor.SHA256[:16])}
	_, hasUpgrade := singleOut.Manifest["upgrades"]
	var launcher Launcher
	if err := json.Unmarshal(singleOut.Manifest["launcher"], &launcher); err != nil {
		t.Fatal(err)
	}
	if singleOut.Version != "12.0.1 - genesis 12.0.1 - cosmovisor 1.7.3" || hasUpgrade ||
		launcher.Commit != testCommit || strings.Join(singleOut.Dirs, " ") != strings.Join(wantDirs, " ") {
		t.Errorf("single bundle version --bundle = %s, want no upgrade, launcher commit %s and directories %q", raw, testCommit, wantDirs)
	}
}

// The marker trusts the files. An exrpd replaced by hand under the marker of this
// bundle tells its version itself. When it fails, the error has its reason.
func TestVersionLineReplacedBinary(t *testing.T) {
	b := testBundle(t)
	home := homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}.build(t)
	installAs(t, b, home, installMatch)
	// The binary answers only when it gets this home, as every child of the launcher must.
	writeFile(t, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd"),
		"#!/bin/sh\n[ \"$1 $2 $3\" = \"--home "+home+" version\" ] && echo 12.0.0-custom\n", 0o755)

	got, err := b.versionLine(testNode(b, home))
	if want := "12.0.0-custom - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3"; err != nil || got != want {
		t.Errorf("versionLine() = %q, %v; want %q", got, err, want)
	}

	writeFile(t, filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd"),
		"#!/bin/sh\necho 'cannot open the home' >&2\nexit 1\n", 0o755)
	if _, err := b.versionLine(testNode(b, home)); err == nil || !strings.Contains(err.Error(), "cannot open the home") {
		t.Errorf("versionLine() error = %v, want the stderr of the failed exrpd", err)
	}
}

// A help request goes to exrpd like any other command, and so does any command after a
// leading flag, which hides version --bundle and completion from the launcher. On a
// home without the bundle they fail with the refusal and write nothing.
func TestForwardedRefusalWritesNothing(t *testing.T) {
	b := testBundle(t)
	noEnv := func(string) string { return "" }
	for _, args := range [][]string{
		{"--help"},
		{"init", "--help"},
		{"init", "--help", "--help=false"},
		{"init", "--help", "--chain-id", "--help=false"},
		{"--log_level", "info", "help", "keys"},
		{"--log_level", "info", "version", "--bundle"},
		{"--log_level", "info", "completion", "--zsh"},
		{"keys", "show", "__complete"},
	} {
		home := t.TempDir()
		inv, err := parseCommand(append([]string{"--home", home}, args...), noEnv, testUserHomeDir)
		if err != nil || inv.command != "" {
			t.Fatalf("parseCommand(%q) = %q, %v; want an exrpd command", args, inv.command, err)
		}
		var refusal *homeRefusal
		if err := runExrpd(b, node{Layout: newLayout(inv.home, b.m), Source: inv.source}, inv.args); !errors.As(err, &refusal) {
			t.Errorf("%q on a new home: error %v, want the refusal", args, err)
		}
		if _, err := os.Stat(filepath.Join(home, "cosmovisor")); !os.IsNotExist(err) {
			t.Errorf("%q wrote to the home (stat error %v)", args, err)
		}
	}
}

// init installs only on a new home: no marker and no config/genesis.json, which native
// init refuses without --overwrite too. config/client.toml alone comes from client
// commands of the standalone exrpd. A home that decide refuses is never new. A refused
// init writes nothing.
func TestInitExrpd(t *testing.T) {
	const existing = "(from --home) already has"
	for _, tc := range []struct {
		name    string
		config  string       // file under config/ before init, "" for none
		data    string       // file under data/ before init, "" for none
		install installState // marker after the install, "" for nothing installed
		refused string       // part of the refusal, "" when init runs
	}{
		{"new home", "", "", "", ""},
		{"client config only", "client.toml", "", "", ""},
		{"existing node without marker", "genesis.json", "", "", existing},
		{"another bundle installed", "", "", installDiffers, existing},
		{"installed, as for init --overwrite", "genesis.json", "", installMatch, ""},
		{"batch file", "", "upgrade-info.json.batch", "", "is a Cosmovisor batch upgrade file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := t.TempDir()
			if tc.config != "" {
				writeFile(t, filepath.Join(home, "config", tc.config), "{}", 0o600)
			}
			if tc.data != "" {
				writeFile(t, filepath.Join(home, "data", tc.data), "{}", 0o600)
			}
			if tc.install != "" {
				installAs(t, b, home, tc.install)
			}
			marker := filepath.Join(home, "cosmovisor", "bundle.sha256")
			before := readFile(t, marker)

			bin, err := b.initExrpd(testNode(b, home))
			if tc.refused != "" {
				if err == nil || !strings.Contains(err.Error(), tc.refused) {
					t.Fatalf("initExrpd() = %q, %v; want a refusal with %q", bin, err, tc.refused)
				}
				if got := readFile(t, marker); got != before {
					t.Errorf("marker = %q, want it unchanged at %q", got, before)
				}
				if _, err := os.Stat(filepath.Join(home, "cosmovisor")); tc.install == "" && !os.IsNotExist(err) {
					t.Errorf("a refused init wrote to the home (stat error %v)", err)
				}
				return
			}
			if want := filepath.Join(home, "cosmovisor", testGenesisDir, "bin", "exrpd"); err != nil || bin != want {
				t.Fatalf("initExrpd() = %q, %v; want %s", bin, err, want)
			}
			if got, err := b.exrpdFor(testNode(b, home)); err != nil || got != bin {
				t.Errorf("exrpdFor() after init = %q, %v; want %s: init must write the marker", got, err, bin)
			}
			if target := currentTarget(t, home); target != "" {
				t.Errorf("init created current -> %q, but only start moves it", target)
			}
		})
	}
}

// bundle setup installs and writes the marker, but never creates or moves `current`:
// the next start picks the exrpd for the data.
func TestSetupKeepsCurrent(t *testing.T) {
	for _, tc := range []struct {
		name    string
		setup   homeSetup
		install installState // marker before setup, "" for nothing installed
	}{
		{"new home", homeSetup{}, ""},
		{"plan reached on genesis", homeSetup{current: genesisName, data: upgradeInfoJSON(testPlan)}, ""},
		{"another bundle installed", homeSetup{current: testUpgradeDir, data: upgradeInfoJSON(testPlan)}, installDiffers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := tc.setup.build(t)
			if tc.install != "" {
				installAs(t, b, home, tc.install)
			}

			if _, err := b.stage(testNode(b, home)); err != nil {
				t.Fatal(err)
			}
			if got := currentTarget(t, home); got != tc.setup.current {
				t.Errorf("current -> %q, want it unchanged at %q", got, tc.setup.current)
			}
			if _, err := b.exrpdFor(testNode(b, home)); err != nil {
				t.Errorf("exrpdFor() after setup error = %v, want the exrpd of the data", err)
			}
		})
	}
}

// A failed exec names the binary, so an operator sees which one a noexec mount or a
// missing file stopped. A file without the execute bit fails like a noexec mount.
func TestExecBinNamesBinary(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "exrpd")
	writeFile(t, bin, "", 0o644)
	if err := execBin(bin, []string{"exrpd"}, os.Environ()); !errors.Is(err, syscall.EACCES) || !strings.Contains(err.Error(), bin) {
		t.Errorf("execBin() error = %v, want EACCES that names %s", err, bin)
	}
}

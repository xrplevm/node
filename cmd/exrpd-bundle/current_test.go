package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// homeSetup describes a node home before a command runs.
type homeSetup struct {
	current  string // `current` target, "" for none
	dangling bool   // current points at a directory that does not exist
	realDir  bool   // current is a directory, not a symlink. Leave current "" with it
	record   string // upgrade-info.json in the directory current points at
	data     string // data/upgrade-info.json, "" for none
	batch    string // data/upgrade-info.json.batch, "" for none
}

func (s homeSetup) build(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	switch {
	case s.realDir:
		if err := os.MkdirAll(filepath.Join(home, "cosmovisor", "current"), 0o755); err != nil {
			t.Fatal(err)
		}
	case s.current == "":
	case s.dangling:
		danglingCurrent(t, home, s.current)
	default:
		linkCurrent(t, home, s.current)
	}
	if s.record != "" {
		writeFile(t, filepath.Join(home, "cosmovisor", s.current, "upgrade-info.json"), s.record, 0o644)
	}
	if s.data != "" {
		writeDataUpgradeInfo(t, home, s.data)
	}
	if s.batch != "" {
		writeFile(t, filepath.Join(home, "data", "upgrade-info.json.batch"), s.batch, 0o644)
	}
	return home
}

// start must leave `current` on the bundle directory whose exrpd runs the data, with the
// plan recorded where Cosmovisor compares it with data/upgrade-info.json. A wrong target
// runs the wrong exrpd. A missing record makes Cosmovisor switch to a directory again.
func TestPrepareCurrent(t *testing.T) {
	v11 := upgradeInfoJSON(testGenesisPlan)
	v12 := upgradeInfoJSON(testPlan)
	name := func(plan string) string { return `{"name":"` + plan + `"}` }
	type prepareCase struct {
		name        string
		setup       homeSetup
		wantCurrent string
		wantRecord  string // upgrade-info.json under wantCurrent
	}
	twoBinary := []prepareCase{
		{"new home links the genesis plan", homeSetup{}, testGenesisDir, name(testGenesisPlan)},
		{"no data moves genesis to the genesis plan", homeSetup{current: genesisName}, testGenesisDir, name(testGenesisPlan)},

		// The genesis plan has its own directory: Cosmovisor finds upgrades/v11.3.0.
		{"genesis plan on a new home", homeSetup{data: v11}, testGenesisDir, v11},
		{"genesis plan on genesis", homeSetup{current: genesisName, data: v11}, testGenesisDir, v11},
		{"genesis plan on a stock Cosmovisor home", homeSetup{current: testGenesisDir, record: v11, data: v11}, testGenesisDir, v11},
		{"genesis plan moves current off the upgrade", homeSetup{current: testUpgradeDir, data: v11}, testGenesisDir, v11},

		{"plan reached on a new home", homeSetup{data: v12}, testUpgradeDir, v12},
		{"plan reached on genesis", homeSetup{current: genesisName, data: v12}, testUpgradeDir, v12},
		{"plan reached on the genesis plan", homeSetup{current: testGenesisDir, record: v11, data: v12}, testUpgradeDir, v12},
		{"plan reached on upgrade", homeSetup{current: testUpgradeDir, record: v12, data: v12}, testUpgradeDir, v12},
		{"plan reached in another case", homeSetup{current: genesisName, data: upgradeInfoJSON("V12.0.0")}, testUpgradeDir, upgradeInfoJSON("V12.0.0")},

		// A resync wipes data/, data/upgrade-info.json included. `current` is the only
		// evidence left of the plan the node runs.
		{"no data keeps the upgrade", homeSetup{current: testUpgradeDir, record: v12}, testUpgradeDir, v12},
		{"no data adds a missing record", homeSetup{current: testUpgradeDir}, testUpgradeDir, name(testPlan)},
		{"no data keeps the genesis plan", homeSetup{current: testGenesisDir, record: v11}, testGenesisDir, v11},
		{"deleted upgrade dir is created again", homeSetup{current: testUpgradeDir, dangling: true}, testUpgradeDir, name(testPlan)},
		{"deleted genesis dir moves to the genesis plan", homeSetup{current: genesisName, dangling: true}, testGenesisDir, name(testGenesisPlan)},

		// Another Cosmovisor setup: its record names the plan, or its upgrades/<plan> name.
		{"foreign dir recording the upgrade", homeSetup{current: "custom", record: v12}, testUpgradeDir, v12},
		{"foreign dir recording the genesis plan", homeSetup{current: "custom", record: v11}, testGenesisDir, v11},
		{"foreign dir with a name-only record", homeSetup{current: "custom", record: name(testPlan)}, testUpgradeDir, name(testPlan)},
		{"foreign plan dir without a record", homeSetup{current: "upgrades/V12.0.0"}, testUpgradeDir, name(testPlan)},
	}
	// A single bundle runs only its genesis plan, testPlan, from testUpgradeDir.
	single := []prepareCase{
		{"single bundle on a new home", homeSetup{}, testUpgradeDir, name(testPlan)},
		{"single bundle on genesis", homeSetup{current: genesisName}, testUpgradeDir, name(testPlan)},
		{"single bundle on its plan", homeSetup{data: v12}, testUpgradeDir, v12},
		// The upgrade directory of a two-binary bundle of the same plan.
		{"single bundle keeps the upgrade", homeSetup{current: testUpgradeDir, record: v12}, testUpgradeDir, v12},
		{"single bundle keeps the upgrade on its plan", homeSetup{current: testUpgradeDir, record: v12, data: v12}, testUpgradeDir, v12},
	}
	// A bundle with more than one upgrade runs each plan from its own directory.
	v114 := upgradeInfoJSON(testMidPlan)
	multi := []prepareCase{
		{"multi-plan bundle on a new home", homeSetup{}, testGenesisDir, name(testGenesisPlan)},
		{"multi-plan bundle on the middle plan", homeSetup{current: testGenesisDir, record: v11, data: v114}, testMidDir, v114},
		{"multi-plan bundle on the last plan", homeSetup{current: testMidDir, record: v114, data: v12}, testUpgradeDir, v12},
		{"multi-plan bundle keeps the middle plan", homeSetup{current: testMidDir, record: v114}, testMidDir, v114},
	}
	for _, set := range []struct {
		bundle func(*testing.T) bundle
		cases  []prepareCase
	}{{testBundle, twoBinary}, {testSingleBundle, single}, {testMultiBundle, multi}} {
		for _, tc := range set.cases {
			t.Run(tc.name, func(t *testing.T) {
				b := set.bundle(t)
				home := tc.setup.build(t)

				if err := b.prepare(testNode(b, home)); err != nil {
					t.Fatal(err)
				}
				if got := currentTarget(t, home); got != tc.wantCurrent {
					t.Fatalf("current -> %q, want %q", got, tc.wantCurrent)
				}
				if got := readFile(t, filepath.Join(home, "cosmovisor", tc.wantCurrent, "upgrade-info.json")); got != tc.wantRecord {
					t.Errorf("%s/upgrade-info.json = %q, want %q", tc.wantCurrent, got, tc.wantRecord)
				}
				if _, err := os.Stat(filepath.Join(home, "cosmovisor", "current", "bin", "exrpd")); err != nil {
					t.Errorf("current does not lead to an exrpd: %v", err)
				}
				if _, err := os.Stat(filepath.Join(home, "cosmovisor", genesisName, "bin", "exrpd")); !os.IsNotExist(err) {
					t.Errorf("genesis/bin/exrpd exists (stat error %v), want nothing installed in genesis/", err)
				}
				if got := readFile(t, filepath.Join(home, "data", "upgrade-info.json")); got != tc.setup.data {
					t.Errorf("data/upgrade-info.json = %q, want it unchanged", got)
				}
				if tmp, _ := filepath.Glob(filepath.Join(home, "cosmovisor/.exrpd-bundle-*")); len(tmp) != 0 {
					t.Errorf("temporary files left behind: %v", tmp)
				}
			})
		}
	}
}

// An absolute link to a directory of the bundle is that directory, and it is rewritten
// in the form Cosmovisor writes.
func TestPrepareCurrentAbsoluteLink(t *testing.T) {
	for _, tc := range []struct {
		name           string
		homeAlias      bool
		targetAlias    bool
		missingParents bool
	}{
		{"direct home", false, false, false},
		{"symlinked home", true, false, false},
		{"symlinked target", false, true, false},
		{"symlinked home without upgrades", true, false, true},
		{"symlinked target without upgrades", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := t.TempDir()
			runHome, targetHome := home, home
			if tc.homeAlias || tc.targetAlias {
				alias := filepath.Join(t.TempDir(), "alias")
				if err := os.Symlink(home, alias); err != nil {
					t.Fatal(err)
				}
				if tc.homeAlias {
					runHome = alias
				}
				if tc.targetAlias {
					targetHome = alias
				}
			}
			target := filepath.Join(targetHome, "cosmovisor", testUpgradeDir)
			danglingCurrent(t, home, target)
			if !tc.missingParents {
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			if err := b.prepare(testNode(b, runHome)); err != nil {
				t.Fatal(err)
			}
			if got := currentTarget(t, home); got != testUpgradeDir {
				t.Errorf("current -> %q, want %q", got, testUpgradeDir)
			}
			if got := readFile(t, filepath.Join(home, "cosmovisor", "current", "bin", "exrpd")); got != "v12 bytes" {
				t.Errorf("current binary = %q, want the restored upgrade binary", got)
			}
		})
	}
}

// A stock genesis/ is the genesis plan, also through an absolute link, a chain of links
// or a symlinked home. A genesis/ made a symlink to a plan directory by hand is that
// plan: moving it to the genesis plan would run an older exrpd.
func TestPrepareCurrentOnGenesis(t *testing.T) {
	// stockGenesis makes <home>/cosmovisor/genesis a real directory and returns its path.
	stockGenesis := func(t *testing.T, home string) string {
		t.Helper()
		genesis := filepath.Join(home, "cosmovisor", genesisName)
		if err := os.MkdirAll(genesis, 0o755); err != nil {
			t.Fatal(err)
		}
		return genesis
	}
	for _, tc := range []struct {
		name  string
		build func(t *testing.T, home string) string // returns the home path prepare runs on
		want  string
	}{
		{"absolute link to a stock genesis", func(t *testing.T, home string) string {
			danglingCurrent(t, home, stockGenesis(t, home))
			return home
		}, testGenesisDir},
		{"link to a link to a stock genesis", func(t *testing.T, home string) string {
			stockGenesis(t, home)
			if err := os.Symlink(genesisName, filepath.Join(home, "cosmovisor", "initial")); err != nil {
				t.Fatal(err)
			}
			danglingCurrent(t, home, "initial")
			return home
		}, testGenesisDir},
		{"absolute link to a stock genesis through a symlinked home", func(t *testing.T, home string) string {
			danglingCurrent(t, home, stockGenesis(t, home))
			alias := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(home, alias); err != nil {
				t.Fatal(err)
			}
			return alias
		}, testGenesisDir},
		{"genesis aliasing the upgrade", func(t *testing.T, home string) string {
			if err := os.MkdirAll(filepath.Join(home, "cosmovisor", testUpgradeDir), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(testUpgradeDir, filepath.Join(home, "cosmovisor", genesisName)); err != nil {
				t.Fatal(err)
			}
			danglingCurrent(t, home, genesisName)
			return home
		}, testUpgradeDir},
		{"genesis aliasing a deleted upgrade", func(t *testing.T, home string) string {
			danglingCurrent(t, home, genesisName)
			if err := os.Symlink(testUpgradeDir, filepath.Join(home, "cosmovisor", genesisName)); err != nil {
				t.Fatal(err)
			}
			return home
		}, testUpgradeDir},
		{"genesis aliasing a deleted unknown directory", func(t *testing.T, home string) string {
			danglingCurrent(t, home, genesisName)
			if err := os.Symlink(filepath.Join("upgrades", "v9.0.0"), filepath.Join(home, "cosmovisor", genesisName)); err != nil {
				t.Fatal(err)
			}
			return home
		}, testGenesisDir},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := t.TempDir()
			runHome := tc.build(t, home)

			if err := b.prepare(testNode(b, runHome)); err != nil {
				t.Fatal(err)
			}
			if got := currentTarget(t, home); got != tc.want {
				t.Errorf("current -> %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPrepareCurrentLogging(t *testing.T) {
	for _, tc := range []struct {
		name        string
		setup       homeSetup
		wantCurrent string
		reason      string // "" for no output
	}{
		{"reached plan", homeSetup{current: genesisName, data: upgradeInfoJSON(testPlan)}, testUpgradeDir, "data/upgrade-info.json names " + testPlan},
		{"missing current", homeSetup{data: upgradeInfoJSON(testPlan)}, testUpgradeDir, "data/upgrade-info.json names " + testPlan},
		{"foreign record", homeSetup{current: "custom", record: upgradeInfoJSON(testPlan)}, testUpgradeDir, "custom records plan " + testPlan},
		{"unchanged reached plan", homeSetup{current: testUpgradeDir, record: upgradeInfoJSON(testPlan), data: upgradeInfoJSON(testPlan)}, testUpgradeDir, ""},
		{"genesis moves to the genesis plan", homeSetup{current: genesisName}, testGenesisDir, "current points at genesis: genesis plan " + testGenesisPlan},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testBundle(t)
			home := tc.setup.build(t)
			var err error
			got := captureStderr(t, func() { err = b.prepare(testNode(b, home)) })
			if err != nil {
				t.Fatal(err)
			}
			// The home comes first. A home without cosmovisor/ also logs the first install.
			got, homeFirst := strings.CutPrefix(got, "exrpd: home "+home+" (from --home)\n")
			if !homeFirst {
				t.Errorf("prepare logged %q, want the home first", got)
			}
			got = strings.Replace(got, "exrpd: installed the bundle in "+filepath.Join(home, "cosmovisor")+"\n", "", 1)
			if target := currentTarget(t, home); target != tc.wantCurrent {
				t.Fatalf("current -> %q, want %q", target, tc.wantCurrent)
			}
			if tc.reason == "" {
				if got != "" {
					t.Errorf("unchanged current logged %q, want no output", got)
				}
				return
			}
			if strings.Count(got, "\n") != 1 {
				t.Errorf("prepare logged %q, want exactly one line", got)
			}
			old := tc.setup.current
			if old == "" {
				old = "missing"
			}
			for _, want := range []string{"exrpd:", old, tc.wantCurrent, tc.reason} {
				if !strings.Contains(got, want) {
					t.Errorf("prepare logged %q, want it to contain %q", got, want)
				}
			}
		})
	}
}

// A single bundle of the plan that a two-binary bundle upgraded to keeps `current` on
// that upgrade directory, its genesis directory, and logs the replaced exrpd. A resync
// leaves only `current`.
func TestPrepareSingleAfterTwoBinary(t *testing.T) {
	for _, resync := range []bool{false, true} {
		two, single := testBundle(t), testSingleBundle(t)
		home := homeSetup{data: upgradeInfoJSON(testPlan)}.build(t)
		captureStderr(t, func() {
			if err := two.prepare(testNode(two, home)); err != nil {
				t.Fatal(err)
			}
		})
		if resync {
			if err := os.Remove(filepath.Join(home, "data", "upgrade-info.json")); err != nil {
				t.Fatal(err)
			}
		}
		var err error
		logged := captureStderr(t, func() { err = single.prepare(testNode(single, home)) })
		if err != nil {
			t.Fatal(err)
		}
		bin := filepath.Join(home, "cosmovisor", testUpgradeDir, "bin", "exrpd")
		if got := currentTarget(t, home); got != testUpgradeDir {
			t.Errorf("resync %v: current -> %q, want %q", resync, got, testUpgradeDir)
		}
		if got := readFile(t, bin); got != "v12.0.1 bytes" {
			t.Errorf("resync %v: %s = %q, want the single bundle exrpd", resync, bin, got)
		}
		if !strings.Contains(logged, "exrpd: replaced "+bin+" ") {
			t.Errorf("resync %v: prepare logged %q, want the replaced %s", resync, logged, bin)
		}
	}
}

// A refused command must leave the home as it was: no binary replaced, `current`, its
// record and the data file unchanged.
func TestPrepareRefusalWritesNothing(t *testing.T) {
	v121 := upgradeInfoJSON("v12.1.0")
	type refusalCase struct {
		name    string
		setup   homeSetup
		wantErr string
	}
	twoBinary := []refusalCase{
		// x/upgrade writes the data file without fsync. The node data may be at or past
		// the plan, so only the operator can restore it.
		{"data file is not JSON", homeSetup{current: genesisName, data: `{"name":"v12.0`}, "is corrupted"},
		{"data file is blank", homeSetup{current: genesisName, data: " "}, "is corrupted"},
		{"data file is null", homeSetup{current: genesisName, data: "null"}, "is corrupted"},
		{"data file has no name", homeSetup{current: genesisName, data: `{"height":40}`}, "is corrupted"},
		{"data file has no height", homeSetup{current: genesisName, data: `{"name":"v12.0.0"}`}, "is corrupted"},
		// exrpd and Cosmovisor read the whole plan and reject these after start would
		// already have changed files.
		{"data file has an invalid time", homeSetup{current: genesisName, data: `{"name":"v12.0.0","height":40,"time":"garbage"}`}, "is corrupted"},
		{"data file has a time", homeSetup{current: genesisName, data: `{"name":"v12.0.0","height":40,"time":"2026-01-01T00:00:00Z"}`}, "is corrupted"},
		{"data file has a client state", homeSetup{current: genesisName, data: `{"name":"v12.0.0","height":40,"upgraded_client_state":{}}`}, "is corrupted"},
		{"data file has a height string", homeSetup{current: genesisName, data: `{"name":"v12.0.0","height":"40"}`}, "is corrupted"},

		// Cosmovisor writes the data file before the upgrade height from a batch file.
		{"batch file", homeSetup{current: genesisName, batch: "[]"}, "is a Cosmovisor batch upgrade file"},
		{"batch file with the data file", homeSetup{current: genesisName, data: upgradeInfoJSON(testPlan), batch: "[]"}, "is a Cosmovisor batch upgrade file"},

		// A binary runs only the data of its own plan.
		{"later plan", homeSetup{current: testUpgradeDir, data: v121}, "names plan v12.1.0"},
		{"later plan on a later upgrade", homeSetup{current: "upgrades/v12.1.0", record: v121, data: v121}, "names plan v12.1.0"},
		{"plan between genesis and upgrade", homeSetup{current: testUpgradeDir, data: upgradeInfoJSON("v11.4.0")}, "names plan v11.4.0"},
		// x/upgrade writes the file only when the running binary halts at a plan, so a node
		// that applied v11.3.0 with a binary swapped in early keeps a v11.2.0 file.
		{"plan before the genesis plan", homeSetup{current: genesisName, data: upgradeInfoJSON("v11.2.0")}, "may be stale"},
		{"plan with another name scheme", homeSetup{current: testUpgradeDir, data: upgradeInfoJSON("v12")}, "names plan v12"},

		// A chain only moves forward: an unknown directory most likely belongs to a
		// later plan, and genesis is the oldest binary of the bundle.
		{"deleted unknown dir", homeSetup{current: "upgrades/v12.1.0", dangling: true}, "does not exist and is not part of this exrpd. This exrpd runs plans v11.3.0, v12.0.0. Point it at upgrades/<plan> of the plan this node's data needs, or delete it only if the data needs plan v11.3.0:"},
		{"foreign dir recording a later plan", homeSetup{current: "custom", record: v121}, "names plan v12.1.0"},
		{"foreign dir recording an earlier plan", homeSetup{current: "custom", record: upgradeInfoJSON("v11.1.0")}, "names plan v11.1.0"},
		{"foreign dir with a broken record", homeSetup{current: "custom", record: "{"}, "does not name a plan"},
		{"foreign dir without a record", homeSetup{current: "custom"}, "has no upgrade-info.json"},
		{"later plan dir without a record", homeSetup{current: "upgrades/v12.1.0"}, "has no upgrade-info.json"},
		// Only Cosmovisor creates `current`, as a symlink. A copy that follows symlinks
		// turns it into a directory, which start cannot replace.
		{"current is a directory", homeSetup{realDir: true}, "is not a symlink"},
		{"current is a directory with the data file", homeSetup{realDir: true, data: upgradeInfoJSON(testPlan)}, "is not a symlink"},
	}
	// A single bundle runs only its genesis plan, testPlan. The data of an earlier plan
	// needs the two-binary bundle of this release.
	singleRuns := "This exrpd is a single bundle: it runs only plan v12.0.0, the data after the v12.0.0 upgrade. A node that has not reached that upgrade needs the two-binary bundle exrpd-bundle_<genesis>-12.0.1 of this release, with <genesis> the release for plan v11.3.0."
	single := []refusalCase{
		{"single bundle with an earlier plan", homeSetup{current: testGenesisDir, record: upgradeInfoJSON(testGenesisPlan), data: upgradeInfoJSON(testGenesisPlan)}, singleRuns + " The file may be stale"},
		{"single bundle on an earlier plan dir", homeSetup{current: testGenesisDir, record: upgradeInfoJSON(testGenesisPlan)}, "names plan v11.3.0. " + singleRuns},
		{"single bundle with a later plan", homeSetup{current: testUpgradeDir, data: v121}, "names plan v12.1.0. This exrpd runs plans v12.0.0."},
		{"single bundle on a later plan dir", homeSetup{current: "custom", record: v121}, "names plan v12.1.0. This exrpd runs plans v12.0.0."},
	}
	for _, set := range []struct {
		bundle func(*testing.T) bundle
		cases  []refusalCase
	}{{testBundle, twoBinary}, {testSingleBundle, single}} {
		for _, tc := range set.cases {
			t.Run(tc.name, func(t *testing.T) {
				b := set.bundle(t)
				home := tc.setup.build(t)
				record := filepath.Join(home, "cosmovisor", tc.setup.current, "upgrade-info.json")
				genesisBin := filepath.Join(home, "cosmovisor/genesis/bin/exrpd")
				writeFile(t, genesisBin, "operator bytes", 0o755)

				var err error
				logged := captureStderr(t, func() { err = b.prepare(testNode(b, home)) })
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("prepare() error = %v, want one containing %q", err, tc.wantErr)
				}
				if want := "exrpd: home " + home + " (from --home)\n"; logged != want {
					t.Errorf("a refused start logged %q, want only the home %q", logged, want)
				}
				if got := readFile(t, genesisBin); got != "operator bytes" {
					t.Errorf("genesis binary = %q, want it unchanged", got)
				}
				if _, err := os.Stat(filepath.Join(home, "cosmovisor", testUpgradeDir, "bin")); !os.IsNotExist(err) {
					t.Errorf("upgrade binary installed by a refused command (stat error %v)", err)
				}
				if tc.setup.realDir {
					if info, err := os.Lstat(filepath.Join(home, "cosmovisor", "current")); err != nil || !info.IsDir() {
						t.Errorf("current is no longer a directory (lstat error %v)", err)
					}
				} else if got := currentTarget(t, home); got != tc.setup.current {
					t.Errorf("current -> %q, want it unchanged at %q", got, tc.setup.current)
				}
				if got := readFile(t, record); got != tc.setup.record {
					t.Errorf("%s/upgrade-info.json = %q, want %q", tc.setup.current, got, tc.setup.record)
				}
				if got := readFile(t, filepath.Join(home, "data", "upgrade-info.json")); got != tc.setup.data {
					t.Errorf("data/upgrade-info.json = %q, want it unchanged", got)
				}
			})
		}
	}
}

// os.WriteFile truncates before it writes, so a command at the upgrade height can see the
// data file empty while x/upgrade writes it. It must wait for the content, and still
// refuse a file that stays empty.
func TestDataFileBeingWritten(t *testing.T) {
	b := testBundle(t)
	home := homeSetup{current: genesisName}.build(t)
	path := filepath.Join(home, "data", "upgrade-info.json")
	writeFile(t, path, "", 0o600)
	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(path, []byte(upgradeInfoJSON(testPlan)), 0o600)
	}()
	d, err := b.decide(newLayout(home, b.m))
	<-done
	if err != nil || d.release.dir() != testUpgradeDir {
		t.Fatalf("decide() = %+v, %v; want %s once the file is written", d, err, testUpgradeDir)
	}

	writeFile(t, path, "", 0o600)
	saved := dataFileRetry
	dataFileRetry.interval, dataFileRetry.attempts = time.Millisecond, 3
	defer func() { dataFileRetry = saved }()
	if err := b.prepare(testNode(b, home)); err == nil || !strings.Contains(err.Error(), "is corrupted") {
		t.Fatalf("prepare() error = %v for a file that stays empty, want a refusal", err)
	}
}

// Two directories of the bundle that are one directory, through a symlink made by hand,
// would get both binaries in turn and run the wrong one.
func TestPrepareRefusesAliasedDirectories(t *testing.T) {
	tests := []struct {
		name  string
		dir   string            // directory that exists, relative to cosmovisor, "" for none
		links map[string]string // symlink to its target, both relative to cosmovisor
	}{
		{
			name:  "plan directory links to the other plan directory",
			dir:   testUpgradeDir,
			links: map[string]string{testGenesisDir: testUpgradeDir},
		},
		{
			name:  "both bin directories link to one directory",
			dir:   "shared",
			links: map[string]string{testGenesisDir + "/bin": "shared", testUpgradeDir + "/bin": "shared"},
		},
		{
			name:  "plan directory links to the other plan directory, not there yet",
			links: map[string]string{testUpgradeDir: testGenesisDir},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBundle(t)
			n := testNode(b, t.TempDir())
			if tt.dir != "" {
				if err := os.MkdirAll(filepath.Join(n.Root, tt.dir), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			for link, target := range tt.links {
				if err := os.MkdirAll(filepath.Dir(filepath.Join(n.Root, link)), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(n.Root, target), filepath.Join(n.Root, link)); err != nil {
					t.Fatal(err)
				}
			}

			err := b.prepare(n)
			if err == nil || !strings.Contains(err.Error(), "are the same directory") {
				t.Fatalf("prepare() error = %v, want a refusal", err)
			}
			for _, dir := range []string{testGenesisDir, testUpgradeDir} {
				if _, err := os.Stat(n.bin(dir)); err == nil {
					t.Errorf("a refused start wrote %s", n.bin(dir))
				}
			}
		})
	}
}

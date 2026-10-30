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
	record   string // upgrade-info.json in the directory current points at
	data     string // data/upgrade-info.json, "" for none
	batch    string // data/upgrade-info.json.batch, "" for none
}

func (s homeSetup) build(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	switch {
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
		writeFile(t, dataPath(home, "upgrade-info.json.batch"), s.batch, 0o644)
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
	for _, tc := range []struct {
		name        string
		setup       homeSetup
		wantCurrent string // "" for none
		wantRecord  string // upgrade-info.json under wantCurrent
	}{
		{"new home leaves current to Cosmovisor", homeSetup{}, "", ""},
		{"no data keeps genesis", homeSetup{current: genesisName}, genesisName, ""},

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
		{"deleted genesis dir is created again", homeSetup{current: genesisName, dangling: true}, genesisName, ""},

		// Another Cosmovisor setup: its record names the plan, or its upgrades/<plan> name.
		{"foreign dir recording the upgrade", homeSetup{current: "custom", record: v12}, testUpgradeDir, v12},
		{"foreign dir recording the genesis plan", homeSetup{current: "custom", record: v11}, testGenesisDir, v11},
		{"foreign dir with a name-only record", homeSetup{current: "custom", record: name(testPlan)}, testUpgradeDir, name(testPlan)},
		{"foreign plan dir without a record", homeSetup{current: "upgrades/V12.0.0"}, testUpgradeDir, name(testPlan)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := tc.setup.build(t)

			if _, err := prepare(files, home); err != nil {
				t.Fatal(err)
			}
			if got := currentTarget(t, home); got != tc.wantCurrent {
				t.Fatalf("current -> %q, want %q", got, tc.wantCurrent)
			}
			if tc.wantCurrent != "" {
				if got := readFile(t, filepath.Join(home, "cosmovisor", tc.wantCurrent, "upgrade-info.json")); got != tc.wantRecord {
					t.Errorf("%s/upgrade-info.json = %q, want %q", tc.wantCurrent, got, tc.wantRecord)
				}
				if _, err := os.Stat(filepath.Join(home, "cosmovisor", "current", "bin", "exrpd")); err != nil {
					t.Errorf("current does not lead to an exrpd: %v", err)
				}
			}
			if got := readFile(t, filepath.Join(home, "cosmovisor", genesisName, "upgrade-info.json")); got != "" {
				t.Errorf("genesis/upgrade-info.json = %q, want none: Cosmovisor would compare it with a later plan", got)
			}
			if got := readFile(t, dataPath(home, "upgrade-info.json")); got != tc.setup.data {
				t.Errorf("data/upgrade-info.json = %q, want it unchanged", got)
			}
			if tmp, _ := filepath.Glob(filepath.Join(home, "cosmovisor/.exrpd-upgrade-*")); len(tmp) != 0 {
				t.Errorf("temporary files left behind: %v", tmp)
			}
		})
	}
}

// An absolute link to a directory of the bundle is that directory, and it is rewritten
// in the form Cosmovisor writes.
func TestPrepareCurrentAbsoluteLink(t *testing.T) {
	files, _ := testPayload(t)
	home := t.TempDir()
	danglingCurrent(t, home, filepath.Join(home, "cosmovisor", testUpgradeDir))

	if _, err := prepare(files, home); err != nil {
		t.Fatal(err)
	}
	if got := currentTarget(t, home); got != testUpgradeDir {
		t.Errorf("current -> %q, want %q", got, testUpgradeDir)
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
		{"unchanged genesis", homeSetup{current: genesisName}, genesisName, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := tc.setup.build(t)
			var err error
			got := captureStderr(t, func() { _, err = prepare(files, home) })
			// A home without cosmovisor/ also logs the first install.
			got = strings.Replace(got, "exrpd: installed the bundle in "+filepath.Join(home, "cosmovisor")+"\n", "", 1)
			if err != nil {
				t.Fatal(err)
			}
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

// A refused command must leave the home as it was: no binary replaced, `current`, its
// record and the data file unchanged.
func TestPrepareRefusalWritesNothing(t *testing.T) {
	v121 := upgradeInfoJSON("v12.1.0")
	for _, tc := range []struct {
		name    string
		setup   homeSetup
		wantErr string
	}{
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
		{"deleted unknown dir", homeSetup{current: "upgrades/v12.1.0", dangling: true}, "does not exist and is not part of this exrpd"},
		{"foreign dir recording a later plan", homeSetup{current: "custom", record: v121}, "names plan v12.1.0"},
		{"foreign dir recording an earlier plan", homeSetup{current: "custom", record: upgradeInfoJSON("v11.1.0")}, "names plan v11.1.0"},
		{"foreign dir with a broken record", homeSetup{current: "custom", record: "{"}, "does not name a plan"},
		{"foreign dir without a record", homeSetup{current: "custom"}, "has no upgrade-info.json"},
		{"later plan dir without a record", homeSetup{current: "upgrades/v12.1.0"}, "has no upgrade-info.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := tc.setup.build(t)
			record := filepath.Join(home, "cosmovisor", tc.setup.current, "upgrade-info.json")
			genesisBin := filepath.Join(home, "cosmovisor/genesis/bin/exrpd")
			writeFile(t, genesisBin, "operator bytes", 0o755)

			_, err := prepare(files, home)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("prepare() error = %v, want one containing %q", err, tc.wantErr)
			}
			if got := readFile(t, genesisBin); got != "operator bytes" {
				t.Errorf("genesis binary = %q, want it unchanged", got)
			}
			if _, err := os.Stat(filepath.Join(home, "cosmovisor", testUpgradeDir, "bin")); !os.IsNotExist(err) {
				t.Errorf("upgrade binary installed by a refused command (stat error %v)", err)
			}
			if got := currentTarget(t, home); got != tc.setup.current {
				t.Errorf("current -> %q, want it unchanged at %q", got, tc.setup.current)
			}
			if got := readFile(t, record); got != tc.setup.record {
				t.Errorf("%s/upgrade-info.json = %q, want %q", tc.setup.current, got, tc.setup.record)
			}
			if got := readFile(t, dataPath(home, "upgrade-info.json")); got != tc.setup.data {
				t.Errorf("data/upgrade-info.json = %q, want it unchanged", got)
			}
		})
	}
}

// os.WriteFile truncates before it writes, so a command at the upgrade height can see the
// data file empty while x/upgrade writes it. It must wait for the content, and still
// refuse a file that stays empty.
func TestDataFileBeingWritten(t *testing.T) {
	files, m := testPayload(t)
	home := homeSetup{current: genesisName}.build(t)
	path := dataPath(home, "upgrade-info.json")
	writeFile(t, path, "", 0o600)
	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(100 * time.Millisecond)
		_ = os.WriteFile(path, []byte(upgradeInfoJSON(testPlan)), 0o600)
	}()
	d, err := decide(home, newLayout(home, m), m)
	<-done
	if err != nil || d.dir != testUpgradeDir {
		t.Fatalf("decide() = %+v, %v; want %s once the file is written", d, err, testUpgradeDir)
	}

	writeFile(t, path, "", 0o600)
	saved := dataFileRetry
	dataFileRetry.interval, dataFileRetry.attempts = time.Millisecond, 3
	defer func() { dataFileRetry = saved }()
	if _, err := prepare(files, home); err == nil || !strings.Contains(err.Error(), "is corrupted") {
		t.Fatalf("prepare() error = %v for a file that stays empty, want a refusal", err)
	}
}

// Two directories of the bundle that are one directory, through a symlink made by hand,
// would get both binaries in turn and run the wrong one.
func TestPrepareRefusesAliasedDirectories(t *testing.T) {
	files, _ := testPayload(t)
	home := homeSetup{current: genesisName}.build(t)
	root := filepath.Join(home, "cosmovisor")
	if err := os.Remove(filepath.Join(root, genesisName)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, testUpgradeDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(testUpgradeDir, filepath.Join(root, genesisName)); err != nil {
		t.Fatal(err)
	}

	_, err := prepare(files, home)
	if err == nil || !strings.Contains(err.Error(), "are the same directory") {
		t.Fatalf("prepare() error = %v, want a refusal", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, testUpgradeDir)); len(entries) != 0 {
		t.Errorf("a refused start wrote %v", entries)
	}
}

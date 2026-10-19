package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// prepare must leave `current` on the bundle binary that runs the data, with the plan
// recorded where Cosmovisor compares it with data/upgrade-info.json. A wrong target runs
// the wrong exrpd. A missing record makes Cosmovisor switch to a binary it does not have.
func TestPrepareCurrent(t *testing.T) {
	v11 := upgradeInfoJSON("v11.2.0")
	v12 := upgradeInfoJSON("V12.0.0") // plan names match case-insensitively
	for _, tc := range []struct {
		name        string
		current     string // "" for none
		data        string // data/upgrade-info.json, "" for none
		selection   string
		wantCurrent string
		wantRecord  string // upgrade-info.json under wantCurrent, "" for none
		wantData    string // data/upgrade-info.json afterwards
	}{
		{"no data keeps a missing current", "", "", "", "", "", ""},
		{"no data keeps current", "upgrades/v11.2.0", "", "", "upgrades/v11.2.0", "", ""},

		{"earlier plan on a new home", "", v11, "", genesisName, v11, v11},
		{"earlier plan on genesis", genesisName, v11, "", genesisName, v11, v11},
		{"earlier plan takes over another Cosmovisor setup", "upgrades/v11.2.0", v11, "", genesisName, v11, v11},
		{"earlier plan moves current off the upgrade", testUpgradeDir, v11, "", genesisName, v11, v11},

		{"plan reached on a new home", "", v12, "", testUpgradeDir, v12, v12},
		{"plan reached on genesis", genesisName, v12, "", testUpgradeDir, v12, v12},
		{"plan reached on upgrade", testUpgradeDir, v12, "", testUpgradeDir, v12, v12},
		{"plan reached wins over --select genesis", genesisName, v12, genesisName, testUpgradeDir, v12, v12},

		{"--select plan without data", "", "", testPlan, testUpgradeDir, `{"name":"v12.0.0"}`, ""},
		{"--select plan deletes an earlier plan", genesisName, v11, "V12.0.0", testUpgradeDir, `{"name":"v12.0.0"}`, ""},
		{"--select plan keeps the reached plan", "", v12, testPlan, testUpgradeDir, v12, v12},
		{"--select genesis without data", testUpgradeDir, "", genesisName, genesisName, "", ""},
		{"--select genesis with an earlier plan", testUpgradeDir, v11, genesisName, genesisName, v11, v11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := t.TempDir()
			if tc.current != "" {
				linkCurrent(t, home, tc.current)
			}
			if tc.data != "" {
				writeDataUpgradeInfo(t, home, tc.data)
			}

			if _, err := prepare(files, home, tc.selection); err != nil {
				t.Fatal(err)
			}
			if got := currentTarget(t, home); got != tc.wantCurrent {
				t.Errorf("current -> %q, want %q", got, tc.wantCurrent)
			}
			if tc.wantCurrent != "" {
				if got := readFile(t, filepath.Join(home, "cosmovisor", tc.wantCurrent, "upgrade-info.json")); got != tc.wantRecord {
					t.Errorf("%s/upgrade-info.json = %q, want %q", tc.wantCurrent, got, tc.wantRecord)
				}
			}
			if got := readFile(t, dataPath(home, "upgrade-info.json")); got != tc.wantData {
				t.Errorf("data/upgrade-info.json = %q, want %q", got, tc.wantData)
			}
			if tmp, _ := filepath.Glob(filepath.Join(home, "cosmovisor/.exrpd-upgrade-*")); len(tmp) != 0 {
				t.Errorf("temporary files left behind: %v", tmp)
			}
		})
	}
}

// A refused command must leave the home as it was: no binary replaced, `current` and
// the data file unchanged.
func TestPrepareRefusalWritesNothing(t *testing.T) {
	for _, tc := range []struct {
		name, data, selection, wantErr string
	}{
		{"unknown --select", "", "v12.0.1", "can select"},
		{"data file is not JSON", `{"name":"v12.0`, "", "delete the file"},
		{"data file is empty", " ", "", "delete the file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files, _ := testPayload(t)
			home := t.TempDir()
			linkCurrent(t, home, genesisName)
			genesisBin := filepath.Join(home, "cosmovisor/genesis/bin/exrpd")
			writeFile(t, genesisBin, "operator bytes", 0o755)
			if tc.data != "" {
				writeDataUpgradeInfo(t, home, tc.data)
			}

			_, err := prepare(files, home, tc.selection)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("prepare() error = %v, want one containing %q", err, tc.wantErr)
			}
			if got := readFile(t, genesisBin); got != "operator bytes" {
				t.Errorf("genesis binary = %q, want it unchanged", got)
			}
			if _, err := os.Stat(filepath.Join(home, "cosmovisor", testUpgradeDir)); !os.IsNotExist(err) {
				t.Errorf("upgrade directory created by a refused command (stat error %v)", err)
			}
			if got := currentTarget(t, home); got != genesisName {
				t.Errorf("current -> %q, want it unchanged", got)
			}
		})
	}
}

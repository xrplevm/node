package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// At startup, Cosmovisor compares the plan named in <home>/data/upgrade-info.json with
// the plan recorded in current/upgrade-info.json. When they differ, it switches to
// upgrades/<data plan>. x/upgrade writes the data file when the chain reaches a plan
// height, so it names the last plan the chain reached. reconcileCurrent points `current`
// at the binary of this bundle that runs that data, and records the plan for it:
//
//   - The embedded plan: the chain is at or past the upgrade height. `current` moves to
//     the upgrade binary.
//   - Any other plan (v11.2.0 on homes that ran that upgrade): `current` moves to the
//     genesis binary. Without the record, Cosmovisor would look for upgrades/<plan>,
//     which this bundle does not carry.
//   - No file: `current` stays on the genesis or upgrade binary of this bundle. When it
//     points anywhere else, for example at the upgrade of an earlier bundle, it moves to
//     genesis, and the plan recorded there moves with it, like the other-plan case. A
//     missing `current` stays missing, because Cosmovisor links it to
//     genesis.
//
// --select <plan> is for data restored past the upgrade from a snapshot or state sync.
// A data file kept from an earlier plan would make Cosmovisor switch back, so it is
// deleted. --select genesis is for data from before the upgrade. A data file that names
// the embedded plan still wins over it.

// upgradeInfo is the content of an upgrade-info.json file.
type upgradeInfo struct {
	plan string // "" when the file is missing
	raw  []byte
}

func (u upgradeInfo) names(plan string) bool {
	return u.plan != "" && strings.EqualFold(u.plan, plan)
}

// parseSelection returns "", "genesis" or plan, matching plan case-insensitively like
// Cosmovisor does.
func parseSelection(selection, plan string) (string, error) {
	switch {
	case selection == "" || selection == genesisName:
		return selection, nil
	case strings.EqualFold(selection, plan):
		return plan, nil
	default:
		return "", fmt.Errorf("--select %q: this bundle can select %q or %q", selection, genesisName, plan)
	}
}

// reconcileCurrent points `current` at the binary the node data needs. selection is the
// value from parseSelection and info the content of data/upgrade-info.json.
func reconcileCurrent(home string, l Layout, plan, selection string, info upgradeInfo) error {
	if selection == plan {
		if info.plan != "" && !info.names(plan) {
			path := dataPath(home, "upgrade-info.json")
			if err := os.Remove(path); err != nil {
				return err
			}
			logf("deleted %s, which names plan %s: --select %s is for data past that plan", path, info.plan, plan)
			info = upgradeInfo{}
		}
		return selectUpgrade(l, plan, info, "--select "+plan)
	}
	if selection == genesisName && !info.names(plan) {
		if err := pointCurrent(l, l.GenesisDir, "--select genesis"); err != nil {
			return err
		}
	}
	switch {
	case info.names(plan):
		return selectUpgrade(l, plan, info, "data/upgrade-info.json names "+info.plan)
	case info.plan != "":
		if err := writeUpgradeInfo(filepath.Join(l.Root, l.GenesisDir), info); err != nil {
			return err
		}
		return pointCurrent(l, l.GenesisDir, "data/upgrade-info.json names "+info.plan)
	}
	return keepBundledCurrent(l)
}

// keepBundledCurrent moves `current` to genesis when it points at neither binary of this
// bundle, and records the plan it ran there for genesis.
func keepBundledCurrent(l Layout) error {
	if _, err := os.Lstat(l.Current()); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if current, err := os.Stat(l.Current()); err == nil {
		for _, dir := range []string{l.GenesisDir, l.UpgradeDir} {
			if bundled, err := os.Stat(filepath.Join(l.Root, dir)); err == nil && os.SameFile(current, bundled) {
				return nil
			}
		}
	}
	info, err := readUpgradeInfo(filepath.Join(l.Current(), "upgrade-info.json"))
	if err != nil {
		return err
	}
	if info.plan != "" {
		if err := writeUpgradeInfo(filepath.Join(l.Root, l.GenesisDir), info); err != nil {
			return err
		}
	}
	return pointCurrent(l, l.GenesisDir, "points at neither binary of this bundle")
}

// selectUpgrade points `current` at the upgrade binary and records the plan for it, the
// way Cosmovisor does when it performs the switch itself.
func selectUpgrade(l Layout, plan string, info upgradeInfo, reason string) error {
	if !info.names(plan) {
		info = upgradeInfo{plan: plan, raw: []byte(`{"name":"` + plan + `"}`)}
	}
	if err := writeUpgradeInfo(filepath.Join(l.Root, l.UpgradeDir), info); err != nil {
		return err
	}
	return pointCurrent(l, l.UpgradeDir, reason)
}

func pointCurrent(l Layout, target, reason string) error {
	old, err := os.Readlink(l.Current())
	if err == nil && old == target {
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		old = "(missing)"
	} else if err != nil {
		old = "(not a readable symlink)"
	}
	if err := symlinkAtomic(target, l.Current()); err != nil {
		return err
	}
	logf("moved %s from %s to %s: %s", l.Current(), old, target, reason)
	return nil
}

// readUpgradeInfo reads an upgrade-info.json file. A missing file, or one without a plan
// name, counts as absent.
func readUpgradeInfo(path string) (upgradeInfo, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return upgradeInfo{}, nil
	}
	if err != nil {
		return upgradeInfo{}, err
	}
	var plan struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &plan); err != nil {
		return upgradeInfo{}, fmt.Errorf("%s is not valid JSON (%w): stop the node, delete the file and start again", path, err)
	}
	return upgradeInfo{plan: plan.Name, raw: raw}, nil
}

// writeUpgradeInfo writes info to <dir>/upgrade-info.json unless that file already
// names the same plan.
func writeUpgradeInfo(dir string, info upgradeInfo) error {
	path := filepath.Join(dir, "upgrade-info.json")
	if existing, err := readUpgradeInfo(path); err == nil && existing.names(info.plan) {
		return nil
	}
	return writeFileAtomic(path, bytes.NewReader(info.raw), 0o644, nil)
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// At startup, Cosmovisor compares the plan named in <home>/data/upgrade-info.json with
// the plan recorded in current/upgrade-info.json. When they differ, it switches to
// upgrades/<data plan>. x/upgrade writes the data file when the running binary halts at
// a plan height, so it names the last plan the node halted at, and a binary runs only the
// data of its own plan. A node that applied a plan with a binary swapped in before the
// plan height keeps an older file. decide chooses the directory of the bundle whose exrpd
// runs the data, before any file changes:
//
//   - data/upgrade-info.json.batch exists: refuse. With that file, Cosmovisor writes the
//     data file before the upgrade height, so the data file cannot tell the plan.
//   - The data file names a plan of the bundle: upgrades/<plan>. The genesis plan has its
//     directory too, so Cosmovisor never looks for a directory the bundle lacks.
//   - The data file names another plan, or is corrupted: refuse.
//   - No data file, for example on a new home or after a resync wiped data/: `current`
//     stays where it is when it names a directory of the bundle, also when that
//     directory was deleted, because install creates it again. A missing `current` stays
//     missing and Cosmovisor links it to genesis. A `current` on any other directory
//     moves to the plan its record names, or to the plan in its name upgrades/<plan>,
//     when the bundle runs that plan. Anything else is refused: a chain only moves
//     forward, so an unknown directory most likely belongs to a later plan.

// upgradeInfo is the content of an upgrade-info.json file.
type upgradeInfo struct {
	plan string // "" when the file is missing
	raw  []byte
}

func (u upgradeInfo) names(plan string) bool {
	return u.plan != "" && strings.EqualFold(u.plan, plan)
}

// recordOf is the record Cosmovisor writes for plan when it switches to its directory.
func recordOf(plan string) upgradeInfo {
	raw, _ := json.Marshal(struct {
		Name string `json:"name"`
	}{plan})
	return upgradeInfo{plan: plan, raw: raw}
}

// decision is where `current` must point for the node data.
type decision struct {
	dir    string      // directory relative to Root whose exrpd runs the data
	link   bool        // point `current` at dir. false leaves a missing `current` missing
	record upgradeInfo // written to dir/upgrade-info.json when it names a plan
	reason string      // why dir runs the data, logged when `current` moves
}

// decide reads the home and returns the directory whose exrpd runs its data. It writes
// nothing, so every command can call it, also while the node runs.
func decide(home string, l Layout, m Manifest) (decision, error) {
	if _, err := os.Lstat(dataPath(home, "upgrade-info.json.batch")); err == nil {
		return decision{}, errors.New("data/upgrade-info.json.batch is a Cosmovisor batch upgrade file. This exrpd carries its upgrades itself, and that file makes Cosmovisor write data/upgrade-info.json before the upgrade height. Delete it. If the node is below the upgrade height, also delete data/upgrade-info.json when it names that upgrade: the batch watcher wrote it early")
	} else if !errors.Is(err, fs.ErrNotExist) {
		return decision{}, err
	}
	path := dataPath(home, "upgrade-info.json")
	data, err := readDataFile(path, m)
	if err != nil {
		return decision{}, err
	}
	if data.plan != "" {
		r, ok := m.release(data.plan)
		if !ok && planBefore(data.plan, m.Genesis.Plan) {
			return decision{}, fmt.Errorf("%s names plan %s, older than the plans this exrpd runs (%s). The file may be stale: x/upgrade writes it only when the running binary halts at a plan, so a node that applied a later plan with a binary swapped in before the plan height keeps the old file. If this node already runs %s or later, write the file again with the last plan it crossed and its height, from the governance proposal. Otherwise start the node with the release for plan %s",
				path, data.plan, m.plans(), m.Genesis.Plan, data.plan)
		}
		if !ok {
			return decision{}, fmt.Errorf("%s names plan %s. This exrpd runs plans %s. Start the node with the release for plan %s",
				path, data.plan, m.plans(), data.plan)
		}
		return decision{dir: planDir(r.Plan), link: true, record: data, reason: "data/upgrade-info.json names " + data.plan}, nil
	}

	target, err := os.Readlink(l.Current())
	if errors.Is(err, fs.ErrNotExist) {
		return decision{dir: genesisName, reason: "no data/upgrade-info.json and no current: genesis"}, nil
	}
	if err != nil {
		return decision{}, fmt.Errorf("%s is not a symlink: %w", l.Current(), err)
	}
	if dir, ok := bundleDir(l, m, target); ok {
		return keep(m, dir), nil
	}
	abs := target
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(l.Root, abs)
	}
	if _, err := os.Stat(abs); errors.Is(err, fs.ErrNotExist) {
		return decision{}, fmt.Errorf("%s points at %s, which does not exist and is not part of this exrpd. Point it at the right directory, or delete it to start from genesis",
			l.Current(), target)
	} else if err != nil {
		return decision{}, err
	}
	for _, dir := range binDirs(m) {
		if sameDir(abs, filepath.Join(l.Root, dir)) {
			return keep(m, dir), nil
		}
	}

	record, err := readRecord(filepath.Join(abs, "upgrade-info.json"))
	if err != nil {
		return decision{}, fmt.Errorf("%s points at %s: %w", l.Current(), target, err)
	}
	if record.plan != "" {
		r, ok := m.release(record.plan)
		if !ok {
			return decision{}, fmt.Errorf("%s points at %s, whose upgrade-info.json names plan %s. This exrpd runs plans %s. Start the node with the release for plan %s",
				l.Current(), target, record.plan, m.plans(), record.plan)
		}
		return decision{dir: planDir(r.Plan), link: true, record: record, reason: target + " records plan " + record.plan}, nil
	}
	if filepath.Base(filepath.Dir(filepath.Clean(abs))) == "upgrades" {
		if r, ok := m.release(filepath.Base(abs)); ok {
			return decision{dir: planDir(r.Plan), link: true, record: recordOf(r.Plan), reason: target + " is the directory of plan " + r.Plan}, nil
		}
	}
	return decision{}, fmt.Errorf("%s points at %s, which has no upgrade-info.json and is not part of this exrpd. Point it at the directory of the plan this node's data needs, or delete it to start from genesis",
		l.Current(), target)
}

// keep leaves `current` on dir, a directory of the bundle, and makes sure a plan
// directory has its record.
func keep(m Manifest, dir string) decision {
	d := decision{dir: dir, link: true, reason: "current points at " + dir + " of this exrpd"}
	for _, r := range m.releases() {
		if dir == planDir(r.Plan) {
			d.record = recordOf(r.Plan)
		}
	}
	return d
}

// bundleDir returns the directory of the bundle that the text of the `current` link
// names, also when that directory does not exist.
func bundleDir(l Layout, m Manifest, target string) (string, bool) {
	rel := filepath.Clean(target)
	if filepath.IsAbs(rel) {
		r, err := filepath.Rel(l.Root, rel)
		if err != nil {
			return "", false
		}
		rel = r
	}
	if rel == genesisName {
		return rel, true
	}
	for _, r := range m.releases() {
		if rel == planDir(r.Plan) {
			return rel, true
		}
	}
	return "", false
}

// sameDir reports whether both paths resolve to the same existing file.
func sameDir(a, b string) bool {
	ai, aErr := os.Stat(a)
	bi, bErr := os.Stat(b)
	return aErr == nil && bErr == nil && os.SameFile(ai, bi)
}

// apply carries out d after install: it writes the record of the plan and points
// `current` at its directory, then checks that `current` resolves.
func apply(l Layout, d decision) error {
	if d.record.plan != "" {
		if err := writeUpgradeInfo(filepath.Join(l.Root, d.dir), d.record); err != nil {
			return err
		}
	}
	if !d.link {
		return nil
	}
	if err := pointCurrent(l, d.dir, d.reason); err != nil {
		return err
	}
	if _, err := os.Stat(l.bin("current")); err != nil {
		return fmt.Errorf("%s does not lead to an exrpd: %w", l.Current(), err)
	}
	return nil
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

// dataFileRetry bounds the wait for a data file that x/upgrade is writing: os.WriteFile
// truncates the file before it writes, so a read at that moment sees it empty.
var dataFileRetry = struct {
	interval time.Duration
	attempts int
}{50 * time.Millisecond, 20}

// readDataFile reads data/upgrade-info.json. A missing file is absent. An empty file is
// read again for a moment, because x/upgrade may be writing it. A file that x/upgrade and
// Cosmovisor would reject is corrupted: power loss or a disk fault can leave it empty or
// cut short. The checks follow upgradetypes.Plan and its ValidateBasic: a name, a height
// above 0, no time and no upgraded client state.
func readDataFile(path string, m Manifest) (upgradeInfo, error) {
	raw, err := os.ReadFile(path)
	for i := 0; err == nil && len(raw) == 0 && i < dataFileRetry.attempts; i++ {
		time.Sleep(dataFileRetry.interval)
		raw, err = os.ReadFile(path)
	}
	if errors.Is(err, fs.ErrNotExist) {
		return upgradeInfo{}, nil
	}
	if err != nil {
		return upgradeInfo{}, err
	}
	var plan struct {
		Name                string          `json:"name"`
		Time                time.Time       `json:"time"`
		Height              int64           `json:"height"`
		Info                string          `json:"info"`
		UpgradedClientState json.RawMessage `json:"upgraded_client_state"`
	}
	err = json.Unmarshal(raw, &plan)
	if err != nil || plan.Name == "" || plan.Height <= 0 || !plan.Time.IsZero() ||
		(len(plan.UpgradedClientState) > 0 && string(plan.UpgradedClientState) != "null") {
		example := m.Upgrades[len(m.Upgrades)-1].Plan
		return upgradeInfo{}, fmt.Errorf(`%s is corrupted (empty or not valid). No files were changed.

Fix it by hand, then start again:
  1. Stop the node.
  2. Find the last upgrade this chain reached: plan name and height, from the governance
     proposal (block explorer or announcement). Do not use this node for that.
  3. Write the file with exactly those values, for example:
       {"name":"%s","height":4521000}
Do not just delete the file: the node may already run past that upgrade.
If you are not sure, contact the XRPL EVM team before you change it`, path, example)
	}
	return upgradeInfo{plan: plan.Name, raw: raw}, nil
}

// planBefore reports whether plan names a release before the genesis plan.
func planBefore(plan, genesisPlan string) bool {
	p, ok := planParts(strings.ToLower(plan))
	g, gok := planParts(genesisPlan)
	return ok && gok && slices.Compare(p, g) < 0
}

// readRecord reads the upgrade-info.json that Cosmovisor writes in a binary directory.
// It names the plan that directory runs and has no height. A missing file is absent.
func readRecord(path string) (upgradeInfo, error) {
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
	if err := json.Unmarshal(raw, &plan); err != nil || plan.Name == "" {
		return upgradeInfo{}, fmt.Errorf("%s does not name a plan: fix or delete it", path)
	}
	return upgradeInfo{plan: plan.Name, raw: raw}, nil
}

// writeUpgradeInfo writes info to <dir>/upgrade-info.json unless that file already
// names the same plan.
func writeUpgradeInfo(dir string, info upgradeInfo) error {
	path := filepath.Join(dir, "upgrade-info.json")
	if existing, err := readRecord(path); err == nil && existing.names(info.plan) {
		return nil
	}
	return writeFileAtomic(path, bytes.NewReader(info.raw), 0o644, nil)
}

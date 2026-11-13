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
	"syscall"
	"time"
)

// genesisName is the directory that stock Cosmovisor links `current` to first. The
// bundle installs nothing there: a `current` on it moves to upgrades/<genesis plan>.
const genesisName = "genesis"

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
	release Release     // the release whose exrpd runs the data, from its directory
	record  upgradeInfo // written to upgrade-info.json in that directory
	reason  string      // why that exrpd runs the data, logged when `current` moves
}

// decide selects a binary without changing the home, so commands can call it while the node runs.
func (b bundle) decide(l Layout) (decision, error) {
	if err := refuseBatchFile(l); err != nil {
		return decision{}, err
	}

	target, hasCurrent, err := readCurrent(l)
	if err != nil {
		return decision{}, err
	}

	path := l.data("upgrade-info.json")
	data, err := readDataFile(path)
	if errors.Is(err, errBadDataFile) {
		return decision{}, errCorruptedDataFile(path)
	}
	if err != nil {
		return decision{}, err
	}

	if data.plan != "" {
		return b.fromDataFile(path, data)
	}

	return b.fromCurrent(l, target, hasCurrent)
}

// refuseBatchFile refuses a home with data/upgrade-info.json.batch.
func refuseBatchFile(l Layout) error {
	_, err := os.Lstat(l.data("upgrade-info.json.batch"))
	if err == nil {
		return errBatchFile()
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// readCurrent returns the text of the `current` link. hasCurrent is false when it is
// missing. A `current` that is not a symlink is an error.
func readCurrent(l Layout) (target string, hasCurrent bool, err error) {
	target, err = os.Readlink(l.Current())
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "", false, nil
	case errors.Is(err, syscall.EINVAL):
		return "", false, fmt.Errorf("%s is not a symlink: %w", l.Current(), err)
	case err != nil:
		return "", false, err
	}
	return target, true, nil
}

// fromDataFile runs the plan the data file at path names, when the bundle runs it.
func (b bundle) fromDataFile(path string, data upgradeInfo) (decision, error) {
	r, ok := b.releaseFor(data.plan)
	switch {
	case ok:
		return decision{release: r, record: data, reason: "data/upgrade-info.json names " + data.plan}, nil
	case b.planBeforeGenesis(data.plan) && b.single():
		return decision{}, errSingleBundlePlan(path, data.plan, b.m)
	case b.planBeforeGenesis(data.plan):
		return decision{}, errStalePlan(path, data.plan, b)
	default:
		return decision{}, errUnknownPlan(path, data.plan, b)
	}
}

// fromCurrent chooses the directory from `current` when there is no data file.
func (b bundle) fromCurrent(l Layout, target string, hasCurrent bool) (decision, error) {
	m := b.m
	if !hasCurrent {
		return onGenesisPlan(m, "no data/upgrade-info.json and no current"), nil
	}

	abs := target
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(l.Root, abs)
	}

	// The link text relative to Root. Both paths are absolute, so Rel cannot fail.
	rel, _ := filepath.Rel(l.Root, abs)
	if r, ok := b.releaseIn(rel); ok {
		return keep(r), nil
	}

	for _, r := range b.releases {
		if sameFile(abs, filepath.Join(l.Root, r.dir())) {
			return keep(r), nil
		}
	}

	root, err := filepath.EvalSymlinks(l.Root)
	if err != nil {
		return decision{}, err
	}

	resolved, err := resolveExistingPrefix(abs)
	if err != nil {
		return decision{}, err
	}

	rel, _ = filepath.Rel(root, resolved)
	if r, ok := b.releaseIn(rel); ok {
		return keep(r), nil
	}

	// A link to a deleted plan directory is that plan, like a `current` on it.
	dangling, err := resolveDangling(resolved)
	if err != nil {
		return decision{}, err
	}
	danglingRel, _ := filepath.Rel(root, dangling)
	if r, ok := b.releaseIn(danglingRel); ok {
		return keep(r), nil
	}

	// genesis that is not a plan directory under another name: stock Cosmovisor's own,
	// named by the link text or reached through another path.
	if rel == genesisName || sameFile(abs, filepath.Join(l.Root, genesisName)) {
		return onGenesisPlan(m, "current points at genesis"), nil
	}

	if _, err := os.Stat(abs); errors.Is(err, fs.ErrNotExist) {
		return decision{}, errCurrentMissing(l, target, b)
	} else if err != nil {
		return decision{}, err
	}

	return b.fromForeignDir(l, target, abs)
}

// fromForeignDir moves `current` off an existing directory outside the bundle, to the
// plan its record names or else to the plan in its name upgrades/<plan>.
func (b bundle) fromForeignDir(l Layout, target, abs string) (decision, error) {
	record, err := readRecord(filepath.Join(abs, "upgrade-info.json"))
	if err != nil {
		return decision{}, fmt.Errorf("%s points at %s: %w", l.Current(), target, err)
	}

	if record.plan != "" {
		r, ok := b.releaseFor(record.plan)
		switch {
		case !ok && b.single() && b.planBeforeGenesis(record.plan):
			return decision{}, errSingleBundleCurrent(l, target, record.plan, b.m)
		case !ok:
			return decision{}, errCurrentUnknownPlan(l, target, record.plan, b)
		}
		return decision{release: r, record: record, reason: target + " records plan " + record.plan}, nil
	}

	if filepath.Base(filepath.Dir(filepath.Clean(abs))) == "upgrades" {
		if r, ok := b.releaseFor(filepath.Base(abs)); ok {
			return decision{release: r, record: recordOf(r.Plan), reason: target + " is the directory of plan " + r.Plan}, nil
		}
	}

	return decision{}, errCurrentNotInBundle(l, target, b)
}

// keep leaves `current` on the directory of r, a plan directory of the bundle, and makes
// sure it has its record.
func keep(r Release) decision {
	return decision{release: r, record: recordOf(r.Plan), reason: "current points at " + r.dir() + " of this exrpd"}
}

// onGenesisPlan runs the genesis exrpd from the directory of the genesis plan. Its
// record names the genesis plan, so Cosmovisor switches at the next plan as usual.
func onGenesisPlan(m Manifest, why string) decision {
	return decision{release: m.Genesis, record: recordOf(m.Genesis.Plan), reason: why + ": genesis plan " + m.Genesis.Plan}
}

// sameFile reports whether both paths resolve to the same existing file.
func sameFile(a, b string) bool {
	ai, aErr := os.Stat(a)
	bi, bErr := os.Stat(b)
	return aErr == nil && bErr == nil && os.SameFile(ai, bi)
}

// resolveExistingPrefix resolves the symlinks of the longest existing prefix of path and
// appends the rest of path unchanged, so a path whose end does not exist yet resolves too.
func resolveExistingPrefix(path string) (string, error) {
	path = filepath.Clean(path)
	for prefix := path; ; prefix = filepath.Dir(prefix) {
		resolved, err := filepath.EvalSymlinks(prefix)
		if err == nil {
			return filepath.Join(resolved, path[len(prefix):]), nil
		}
		if !errors.Is(err, fs.ErrNotExist) || filepath.Dir(prefix) == prefix {
			return "", err
		}
	}
}

// resolveDangling resolves path like resolveExistingPrefix, and also follows a symlink
// whose target does not exist, so a link to a deleted directory resolves to that directory.
func resolveDangling(path string) (string, error) {
	// EvalSymlinks does not see a loop that passes through a name that does not exist.
	for range 32 {
		resolved, err := resolveExistingPrefix(path)
		if err != nil {
			return "", err
		}
		// The first missing element of resolved is a symlink only when it dangles.
		link := resolved
		target, err := os.Readlink(link)
		for err != nil {
			if _, statErr := os.Lstat(link); statErr == nil || filepath.Dir(link) == link {
				return resolved, nil
			}
			link = filepath.Dir(link)
			target, err = os.Readlink(link)
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(link), target)
		}
		path = filepath.Join(target, resolved[len(link):])
	}
	return "", fmt.Errorf("%s: too many symlinks", path)
}

// apply carries out d after install: it writes the record of the plan and points
// `current` at its directory, then checks that `current` resolves.
func apply(l Layout, d decision) error {
	if err := writeUpgradeInfo(filepath.Join(l.Root, d.release.dir()), d.record); err != nil {
		return err
	}
	if err := pointCurrent(l, d.release.dir(), d.reason); err != nil {
		return err
	}
	if _, err := os.Stat(l.bin("current")); err != nil {
		return fmt.Errorf("%s does not lead to an exrpd: %w", l.Current(), err)
	}
	return nil
}

// pointCurrent points `current` at target, a directory relative to Root, and logs the
// move with reason. It writes nothing when `current` already points there.
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

// errBadDataFile is returned by readDataFile for a data file that x/upgrade and
// Cosmovisor would reject. decide turns it into errCorruptedDataFile.
var errBadDataFile = errors.New("data/upgrade-info.json is corrupted")

// readDataFile reads data/upgrade-info.json. A missing file is absent. An empty file is
// read again for a moment, because x/upgrade may be writing it. A file that x/upgrade and
// Cosmovisor would reject returns errBadDataFile: power loss or a disk fault can leave it
// empty or cut short. The checks follow upgradetypes.Plan and its ValidateBasic: a name,
// a height above 0, no time and no upgraded client state.
func readDataFile(path string) (upgradeInfo, error) {
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
		return upgradeInfo{}, errBadDataFile
	}

	return upgradeInfo{plan: plan.Name, raw: raw}, nil
}

// planBeforeGenesis reports whether plan names a release before the genesis plan.
func (b bundle) planBeforeGenesis(plan string) bool {
	p, ok := planParts(strings.ToLower(plan))
	return ok && slices.Compare(p[:], b.genesisParts[:]) < 0
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

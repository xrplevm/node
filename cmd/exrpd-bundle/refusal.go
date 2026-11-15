package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// homeStatus is what the commands of a home act on, read once.
type homeStatus struct {
	decision decision     // where `current` must point, unless refusal is refusedHome
	marker   installState // the bundle that the marker names, and nothing else
	refusal  *homeRefusal // why commands do not run the chosen exrpd, nil when they do
}

// inspect reads the home of n for the commands that act on it. It writes nothing. The
// error is a failure to read the marker or to stat the chosen exrpd, never a refusal.
func (b bundle) inspect(n node) (homeStatus, error) {
	d, decideErr := b.decide(n.Layout)
	marker, err := b.installStateOf(n.Layout)
	if err != nil {
		return homeStatus{}, err
	}

	st := homeStatus{decision: d, marker: marker}
	if decideErr != nil {
		st.refusal = &homeRefusal{kind: refusedHome, err: decideErr}
	} else {
		bin := n.bin(d.release.dir())
		switch marker {
		case installMissing:
			st.refusal = newHomeRefusal(n, refusedNoMarker, bin)
		case installDiffers:
			st.refusal = newHomeRefusal(n, refusedOtherBundle, bin)
		default:
			_, err := os.Stat(bin)
			if errors.Is(err, fs.ErrNotExist) {
				st.refusal = newHomeRefusal(n, refusedNoExrpd, bin)
			} else if err != nil {
				return homeStatus{}, err
			}
		}
	}

	return st, nil
}

// homeRefusal is why a command other than an installer does not run: decide refuses the
// home, or the home does not have this bundle installed, or not the exrpd that runs its
// data.
type homeRefusal struct {
	kind refusalKind
	err  error
}

func (r *homeRefusal) Error() string {
	return r.err.Error()
}

func (r *homeRefusal) Unwrap() error {
	return r.err
}

// refusalKind is what keeps commands from the exrpd of a home.
type refusalKind int

const (
	refusedHome        refusalKind = iota // decide refuses the home: it chose no exrpd
	refusedNoMarker                       // no marker: nothing installed, an older bundle, or an interrupted install
	refusedOtherBundle                    // the marker names another bundle
	refusedNoExrpd                        // the marker names this bundle, but the exrpd is missing
)

// lineLabel is what the version line shows instead of the installed version.
func (k refusalKind) lineLabel() string {
	switch k {
	case refusedNoMarker, refusedNoExrpd:
		return "not-installed"
	case refusedOtherBundle:
		return "restart-needed"
	default:
		return ""
	}
}

// newHomeRefusal names the home of n, where it comes from, and the next step for kind,
// one of the kinds after refusedHome. bin is the exrpd that decide chose. Without a
// marker, a home without config/genesis.json and data/ has no node, so the operator
// either creates one or passes the right home.
func newHomeRefusal(n node, kind refusalKind, bin string) *homeRefusal {
	home, source := n.Home, n.Source
	var msg string
	switch {
	case kind == refusedOtherBundle:
		msg = fmt.Sprintf("The bundle installed in %s (from %s) differs from this exrpd. Restart the node service, or run `exrpd bundle setup --home %s` while the node is stopped.", home, source, home)
	case kind == refusedNoExrpd:
		msg = fmt.Sprintf("The bundle is installed in %s (from %s), but %s is missing. Restart the node service, or run `exrpd bundle setup --home %s` while the node is stopped.", home, source, bin, home)
	case !exists(filepath.Join(home, "config", "genesis.json")) && !exists(n.data()):
		msg = fmt.Sprintf("No exrpd is installed in %s (from %s). New node: run `exrpd init <moniker> --chain-id <id> --home %s`. Existing node: this is probably not its home, pass --home.", home, source, home)
	default:
		msg = fmt.Sprintf("The bundle is not installed in %s (from %s). Restart the node service, or run `exrpd bundle setup --home %s` while the node is stopped.", home, source, home)
	}
	return &homeRefusal{kind: kind, err: errors.New(msg)}
}

// The errors below are the refusals of decide, which inspect wraps as refusedHome.

// errBatchFile refuses a Cosmovisor batch upgrade file.
func errBatchFile() error {
	return errors.New("data/upgrade-info.json.batch is a Cosmovisor batch upgrade file. This exrpd carries its upgrades itself, and that file makes Cosmovisor write data/upgrade-info.json before the upgrade height. Delete it. If the node is below the upgrade height, also delete data/upgrade-info.json when it names that upgrade: the batch watcher wrote it early")
}

// errCorruptedDataFile tells how to fix a corrupted data file by hand.
func errCorruptedDataFile(path string) error {
	return fmt.Errorf(`%s is corrupted (empty or not valid). No files were changed.

Fix it by hand, then start again:
  1. Stop the node.
  2. Check this node's last local block height and upgrade logs. Find the applicable
     plan name and height in the block explorer or governance proposal.
     Use the upgrade this node reached, not simply the network's latest upgrade.
     If the node halted at upgrade height H, use that plan and height H,
     even if its last committed block is H-1.
  3. Write the file using this template:
       {"name":"<plan>","height":<height>}
     Replace <plan> and <height> with the exact plan name and numeric height.
     The plan name can differ from the binary version.
Do not just delete the file: the node may already run past that upgrade.
If you are not sure, contact the XRPL EVM team before you change it`, path)
}

// errStalePlan refuses a data file that names a plan older than the genesis plan.
func errStalePlan(path, plan string, b bundle) error {
	return fmt.Errorf("%s names plan %s, older than the plans this exrpd runs (%s). The file may be stale: x/upgrade writes it only when the running binary halts at a plan, so a node that applied a later plan with a binary swapped in before the plan height keeps the old file. If this node already runs %s or later, write the file again with the last plan it crossed and its height, from the governance proposal. Otherwise start the node with the release for plan %s",
		path, plan, b.plans(), b.m.Genesis.Plan, plan)
}

// errUnknownPlan refuses a data file that names a plan the bundle does not run.
func errUnknownPlan(path, plan string, b bundle) error {
	return fmt.Errorf("%s names plan %s. This exrpd runs plans %s. Start the node with the release for plan %s",
		path, plan, b.plans(), plan)
}

// errSingleBundlePlan refuses, in a single bundle, a data file that names a plan older
// than the genesis plan.
func errSingleBundlePlan(path, plan string, m Manifest) error {
	return fmt.Errorf("%s names plan %s. %s The file may be stale: x/upgrade writes it only when the running binary halts at a plan, so a node that applied a later plan with a binary swapped in before the plan height keeps the old file. If this node already runs %s or later, write the file again with the last plan it crossed and its height, from the governance proposal",
		path, plan, singleBundleRuns(plan, m), m.Genesis.Plan)
}

// errSingleBundleCurrent refuses, in a single bundle, a `current` whose record names a
// plan older than the genesis plan.
func errSingleBundleCurrent(l Layout, target, plan string, m Manifest) error {
	return fmt.Errorf("%s points at %s, whose upgrade-info.json names plan %s. %s If this node already runs %s or later, write %s with the last plan it crossed and its height, from the governance proposal",
		l.Current(), target, plan, singleBundleRuns(plan, m), m.Genesis.Plan, l.data("upgrade-info.json"))
}

// singleBundleRuns tells that a single bundle runs only its genesis plan, and which
// bundle of this release runs the data of plan, a plan before it.
func singleBundleRuns(plan string, m Manifest) string {
	return fmt.Sprintf("This exrpd is a single bundle: it runs only plan %s, the data after the %s upgrade. A node that has not reached that upgrade needs the two-binary bundle exrpd-bundle_<genesis>-%s of this release, with <genesis> the release for plan %s.",
		m.Genesis.Plan, m.Genesis.Plan, m.Genesis.Version, plan)
}

// currentChoices tells where to point `current`, and when to delete it.
func currentChoices(b bundle) string {
	return fmt.Sprintf("This exrpd runs plans %s. Point it at upgrades/<plan> of the plan this node's data needs, or delete it only if the data needs plan %s: without it, this exrpd runs that plan",
		b.plans(), b.m.Genesis.Plan)
}

// errCurrentMissing refuses a `current` on a missing directory outside the bundle.
func errCurrentMissing(l Layout, target string, b bundle) error {
	return fmt.Errorf("%s points at %s, which does not exist and is not part of this exrpd. %s",
		l.Current(), target, currentChoices(b))
}

// errCurrentUnknownPlan refuses a `current` whose record names a plan the bundle does
// not run.
func errCurrentUnknownPlan(l Layout, target, plan string, b bundle) error {
	return fmt.Errorf("%s points at %s, whose upgrade-info.json names plan %s. This exrpd runs plans %s. Start the node with the release for plan %s",
		l.Current(), target, plan, b.plans(), plan)
}

// errCurrentNotInBundle refuses a `current` on a directory that names no plan of the
// bundle.
func errCurrentNotInBundle(l Layout, target string, b bundle) error {
	return fmt.Errorf("%s points at %s, which has no upgrade-info.json and is not part of this exrpd. %s",
		l.Current(), target, currentChoices(b))
}

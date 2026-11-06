package main

import (
	"path/filepath"
	"strings"
)

// genesisName is the directory that stock Cosmovisor links `current` to first. The
// bundle installs nothing there: a `current` on it moves to upgrades/<genesis plan>.
const genesisName = "genesis"

// Layout is where the binaries live under <home>/cosmovisor, in the stock Cosmovisor
// layout without genesis/. Every plan of the bundle has its directory upgrades/<plan>,
// the genesis plan too, so the plan in data/upgrade-info.json always names a directory
// of the bundle and `current` always has a record. The supervisor directory is named
// after the Cosmovisor hash, so a later bundle with a different Cosmovisor never
// collides with this one.
type Layout struct {
	Home       string // the node home
	Root       string // <home>/cosmovisor, Cosmovisor's working directory
	Cosmovisor string
}

func newLayout(home string, m Manifest) Layout {
	root := filepath.Join(home, "cosmovisor")
	return Layout{
		Home:       home,
		Root:       root,
		Cosmovisor: filepath.Join(root, "supervisor", m.Cosmovisor.SHA256[:16], "cosmovisor"),
	}
}

// node is the home a command acts on and where its path comes from.
type node struct {
	Layout
	Source string
}

// bin is the exrpd of dir, a directory relative to Root.
func (l Layout) bin(dir string) string {
	return filepath.Join(l.Root, dir, "bin", "exrpd")
}

// Current is the symlink Cosmovisor follows to find the binary to run.
func (l Layout) Current() string {
	return filepath.Join(l.Root, "current")
}

// dir is the directory of r's binary, upgrades/<plan>, relative to Root, as Cosmovisor
// names it.
func (r Release) dir() string {
	return filepath.Join("upgrades", r.Plan)
}

// releaseIn returns the release whose directory is dir, relative to Root.
func (b bundle) releaseIn(dir string) (Release, bool) {
	for _, r := range b.releases {
		if r.dir() == dir {
			return r, true
		}
	}
	return Release{}, false
}

// releaseFor returns the release of plan. Plan names match case-insensitively, like
// Cosmovisor matches them.
func (b bundle) releaseFor(plan string) (Release, bool) {
	for _, r := range b.releases {
		if strings.EqualFold(r.Plan, plan) {
			return r, true
		}
	}
	return Release{}, false
}

// dirs lists every directory of the bundle: the binary directories and the supervisor
// directory.
func (l Layout) dirs(b bundle) []string {
	d := make([]string, 0, len(b.releases)+1)
	for _, r := range b.releases {
		d = append(d, filepath.Join(l.Root, r.dir()))
	}
	return append(d, filepath.Dir(l.Cosmovisor))
}

// data returns a path under the node's data directory.
func (l Layout) data(elem ...string) string {
	return filepath.Join(append([]string{l.Home, "data"}, elem...)...)
}

package main

import "path/filepath"

// genesisName names the Cosmovisor directory of the binary that a home without
// data/upgrade-info.json runs. Cosmovisor links `current` to it when `current` is
// missing.
const genesisName = "genesis"

// Layout is where the binaries live under <home>/cosmovisor, in the stock Cosmovisor
// layout. Every plan of the bundle has its directory upgrades/<plan>, the genesis plan
// too, so the plan in data/upgrade-info.json always names a directory of the bundle and
// genesis never needs a record. The supervisor directory is named after the Cosmovisor
// hash, so a later bundle with a different Cosmovisor never collides with this one.
type Layout struct {
	Root       string // <home>/cosmovisor, Cosmovisor's working directory
	Cosmovisor string
}

func newLayout(home string, m Manifest) Layout {
	root := filepath.Join(home, "cosmovisor")
	return Layout{
		Root:       root,
		Cosmovisor: filepath.Join(root, supervisorPath(m)),
	}
}

// planDir is the directory of plan's binary, relative to Root, as Cosmovisor names it.
func planDir(plan string) string {
	return filepath.Join("upgrades", plan)
}

// bin is the exrpd of dir, a directory relative to Root.
func (l Layout) bin(dir string) string {
	return filepath.Join(l.Root, dir, "bin", "exrpd")
}

// Current is the symlink Cosmovisor follows to find the binary to run.
func (l Layout) Current() string {
	return filepath.Join(l.Root, "current")
}

// binDirs lists the binary directories of the bundle relative to Root: genesis and the
// plan directories.
func binDirs(m Manifest) []string {
	d := []string{genesisName}
	for _, r := range m.releases() {
		d = append(d, planDir(r.Plan))
	}
	return d
}

// dirs lists every directory of the bundle relative to Root: the binary directories and
// the supervisor directory.
func dirs(m Manifest) []string {
	return append(binDirs(m), filepath.Dir(supervisorPath(m)))
}

// supervisorPath is the embedded Cosmovisor, relative to Root.
func supervisorPath(m Manifest) string {
	return filepath.Join("supervisor", m.Cosmovisor.SHA256[:16], "cosmovisor")
}

// dataPath returns a path under the node's data directory.
func dataPath(home string, elem ...string) string {
	return filepath.Join(append([]string{home, "data"}, elem...)...)
}

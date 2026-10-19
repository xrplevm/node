package main

import "path/filepath"

// genesisName names the Cosmovisor directory of the binary the chain runs before the
// upgrade. stage --select takes the same name.
const genesisName = "genesis"

// Layout is where the binaries live under <home>/cosmovisor. Genesis and Upgrade follow
// the stock Cosmovisor layout. The supervisor directory is named after the Cosmovisor
// hash, so a later bundle with a different Cosmovisor never collides with this one.
type Layout struct {
	Root       string // <home>/cosmovisor, Cosmovisor's working directory
	GenesisDir string // directory `current` names before the upgrade, relative to Root
	UpgradeDir string // directory `current` names after the upgrade, relative to Root
	Genesis    string
	Upgrade    string
	Cosmovisor string
}

func newLayout(home string, m Manifest) Layout {
	root := filepath.Join(home, "cosmovisor")
	l := Layout{
		Root:       root,
		GenesisDir: genesisName,
		UpgradeDir: filepath.Join("upgrades", m.Plan),
		Cosmovisor: filepath.Join(root, "supervisor", m.Cosmovisor.SHA256[:16], "cosmovisor"),
	}
	l.Genesis = filepath.Join(root, l.GenesisDir, "bin", "exrpd")
	l.Upgrade = filepath.Join(root, l.UpgradeDir, "bin", "exrpd")
	return l
}

// Current is the symlink Cosmovisor follows to find the binary to run.
func (l Layout) Current() string {
	return filepath.Join(l.Root, "current")
}

// dataPath returns a path under the node's data directory.
func dataPath(home string, elem ...string) string {
	return filepath.Join(append([]string{home, "data"}, elem...)...)
}

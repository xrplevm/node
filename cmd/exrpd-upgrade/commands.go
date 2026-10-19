package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

// runStart prepares the home and replaces this process with Cosmovisor. Exec keeps the
// PID, so the service manager supervises Cosmovisor directly. args are passed through
// to "exrpd start".
func runStart(args []string) error {
	home, args, err := resolveHome(args, os.Getenv, os.UserHomeDir)
	if err != nil {
		return err
	}
	if err := refuseCosmovisorConfig(args); err != nil {
		return err
	}
	env, err := cosmovisorEnv(os.Environ(), home)
	if err != nil {
		return err
	}
	layout, err := prepare(embeddedPayload(), home, "")
	if err != nil {
		return err
	}
	argv := append([]string{"cosmovisor", "run", "start", "--home", home}, args...)
	return syscall.Exec(layout.Cosmovisor, argv, env) //nolint:gosec // the Cosmovisor that prepare staged and verified
}

// runStage prepares the home without starting the node, optionally with an explicit
// --select of the binary `current` points at.
func runStage(args []string) error {
	home, args, err := resolveHome(args, os.Getenv, os.UserHomeDir)
	if err != nil {
		return err
	}
	selection, args, _, err := takeFlag(args, "--select")
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("stage: unexpected arguments %q", args)
	}
	layout, err := prepare(embeddedPayload(), home, selection)
	if err != nil {
		return err
	}
	fmt.Printf("genesis:    %s\nupgrade:    %s\ncosmovisor: %s\n", layout.Genesis, layout.Upgrade, layout.Cosmovisor)
	if target, err := os.Readlink(layout.Current()); err == nil {
		fmt.Printf("current:    %s\n", target)
	}
	return nil
}

// runInfo prints the embedded manifest.
func runInfo() error {
	m, err := loadManifest(embeddedPayload())
	if err != nil {
		return err
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// prepare installs the embedded binaries under home and points `current` at the binary
// the node data needs. selection is the operator's --select value, "" for none. Every
// check runs before the first file is written.
func prepare(files fs.FS, home, selection string) (Layout, error) {
	m, err := loadManifest(files)
	if err != nil {
		return Layout{}, err
	}
	selection, err = parseSelection(selection, m.Plan)
	if err != nil {
		return Layout{}, err
	}
	info, err := readUpgradeInfo(dataPath(home, "upgrade-info.json"))
	if err != nil {
		return Layout{}, err
	}
	layout := newLayout(home, m)
	if err := install(files, m, layout); err != nil {
		return Layout{}, err
	}
	return layout, reconcileCurrent(home, layout, m.Plan, selection, info)
}

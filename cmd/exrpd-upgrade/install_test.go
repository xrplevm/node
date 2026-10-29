package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// The genesis binary goes to genesis/, for a home without a data file, and to the
// directory of its plan, which data/upgrade-info.json names on homes that ran it. The
// first install into a home is logged, so an install into an unexpected home shows.
func TestInstallWritesStockLayout(t *testing.T) {
	files, m := testPayload(t)
	home := t.TempDir()
	l := newLayout(home, m)
	var err error
	logged := captureStderr(t, func() { err = install(files, m, l) })
	if err != nil {
		t.Fatal(err)
	}
	if want := "exrpd: installed the bundle in " + l.Root + "\n"; logged != want {
		t.Errorf("install logged %q, want %q", logged, want)
	}

	for path, want := range map[string]string{
		filepath.Join(home, "cosmovisor/genesis/bin/exrpd"):                                  "v11 bytes",
		filepath.Join(home, "cosmovisor/upgrades/v11.3.0/bin/exrpd"):                         "v11 bytes",
		filepath.Join(home, "cosmovisor/upgrades/v12.0.0/bin/exrpd"):                         "v12 bytes",
		filepath.Join(home, "cosmovisor/supervisor", m.Cosmovisor.SHA256[:16], "cosmovisor"): "cosmovisor bytes",
	} {
		if got := readFile(t, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("%s mode = %v, want 0755", path, info.Mode().Perm())
		}
	}
	if target := currentTarget(t, home); target != "" {
		t.Errorf("install created current -> %q, but Cosmovisor owns its first link", target)
	}
	logged = captureStderr(t, func() { err = install(files, m, l) })
	if err != nil {
		t.Fatalf("installing the same files again: %v", err)
	}
	if logged != "" {
		t.Errorf("installing the same files again logged %q, want no output", logged)
	}
}

// A bundle with a patch release replaces the binaries an earlier bundle staged, and
// leaves no temporary file next to them.
func TestInstallReplacesDifferentBinary(t *testing.T) {
	files, m := testPayload(t)
	l := newLayout(t.TempDir(), m)
	genesis, upgrade := l.bin(genesisName), l.bin(testUpgradeDir)
	writeFile(t, genesis, "v11 earlier patch", 0o644)
	writeFile(t, upgrade, "v12 earlier patch", 0o644)

	if err := install(files, m, l); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{genesis: "v11 bytes", upgrade: "v12 bytes"} {
		if got := readFile(t, path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("%s mode = %v, want 0755", path, info.Mode().Perm())
		}
		if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
			t.Errorf("left temporary files next to %s: %v", path, entries)
		}
	}
}

// A matching binary copied or restored without execute permission must still run.
func TestInstallMakesExistingBinaryExecutable(t *testing.T) {
	files, m := testPayload(t)
	l := newLayout(t.TempDir(), m)
	upgrade := l.bin(testUpgradeDir)
	writeFile(t, upgrade, "v12 bytes", 0o644)

	if err := install(files, m, l); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(upgrade)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("%s mode = %v, want 0755", upgrade, info.Mode().Perm())
	}
}

// Concurrent starts move `current` at the same time and must not fail on each other's
// temporary link.
func TestSymlinkAtomicConcurrent(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "current")
	errs := make(chan error)
	for _, target := range []string{"genesis", "upgrades/v12.0.0", "genesis", "upgrades/v12.0.0"} {
		go func() {
			var err error
			for range 200 {
				if err = symlinkAtomic(target, link); err != nil {
					break
				}
			}
			errs <- err
		}()
	}
	for range 4 {
		if err := <-errs; err != nil {
			t.Error(err)
		}
	}
	if tmp, _ := filepath.Glob(filepath.Join(dir, ".exrpd-upgrade-*")); len(tmp) != 0 {
		t.Errorf("temporary links left behind: %v", tmp)
	}
}

func TestInstallRejectsCorruptPayload(t *testing.T) {
	files, m := testPayload(t)
	files[m.Upgrades[0].File] = &fstest.MapFile{Data: gzipped(t, "v12 bytez")}
	home := t.TempDir()
	l := newLayout(home, m)

	err := install(files, m, l)
	if err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Fatalf("install() error = %v, want a corrupt payload error", err)
	}
	entries, err := os.ReadDir(filepath.Dir(l.bin(testUpgradeDir)))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("left files behind: %v", entries)
	}
}

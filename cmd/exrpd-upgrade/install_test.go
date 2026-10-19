package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestInstallWritesStockLayout(t *testing.T) {
	files, m := testPayload(t)
	home := t.TempDir()
	l := newLayout(home, m)
	if err := install(files, m, l); err != nil {
		t.Fatal(err)
	}

	for path, want := range map[string]string{
		filepath.Join(home, "cosmovisor/genesis/bin/exrpd"):                                  "v11 bytes",
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
		t.Errorf("install created current -> %q; Cosmovisor owns its first link", target)
	}
	if err := install(files, m, l); err != nil {
		t.Fatalf("installing the same files again: %v", err)
	}
}

// A bundle with a patch release replaces the binaries an earlier bundle staged, and
// leaves no temporary file next to them.
func TestInstallReplacesDifferentBinary(t *testing.T) {
	files, m := testPayload(t)
	l := newLayout(t.TempDir(), m)
	writeFile(t, l.Genesis, "v11 earlier patch", 0o644)
	writeFile(t, l.Upgrade, "v12 earlier patch", 0o644)

	if err := install(files, m, l); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{l.Genesis: "v11 bytes", l.Upgrade: "v12 bytes"} {
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
	writeFile(t, l.Upgrade, "v12 bytes", 0o644)

	if err := install(files, m, l); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(l.Upgrade)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("%s mode = %v, want 0755", l.Upgrade, info.Mode().Perm())
	}
}

// Concurrent stage runs move `current` at the same time and must not fail on each
// other's temporary link.
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
	files[m.Upgrade.File] = &fstest.MapFile{Data: gzipped(t, "v12 bytez")}
	home := t.TempDir()
	l := newLayout(home, m)

	err := install(files, m, l)
	if err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Fatalf("install() error = %v, want a corrupt payload error", err)
	}
	entries, err := os.ReadDir(filepath.Dir(l.Upgrade))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("left files behind: %v", entries)
	}
}

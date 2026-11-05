package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMkdirAllSync(t *testing.T) {
	root := t.TempDir()

	dir := filepath.Join(root, "upgrades", testPlan, "bin")
	if err := mkdirAllSync(dir); err != nil {
		t.Fatal(err)
	}
	for d := dir; d != root; d = filepath.Dir(d) {
		info, err := os.Stat(d)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Errorf("%s: mode %v, want a directory", d, info.Mode())
		}
	}
	if err := mkdirAllSync(dir); err != nil {
		t.Errorf("existing directory: %v", err)
	}

	writeFile(t, filepath.Join(root, "file"), "x", 0o644)
	err := mkdirAllSync(filepath.Join(root, "file", "sub"))
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("ancestor is a file: got %v, want a not a directory error", err)
	}
}

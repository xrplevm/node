package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadManifest(t *testing.T) {
	files, want := testPayload(t)
	got, err := loadManifest(files)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("loadManifest() = %+v, want %+v", got, want)
	}
}

func TestLoadManifestRejectsInvalid(t *testing.T) {
	for name, corrupt := range map[string]func(*Manifest){
		"other platform":      func(m *Manifest) { m.Platform = "plan9/mips" },
		"plan with separator": func(m *Manifest) { m.Plan = "../v12" },
		"file with separator": func(m *Manifest) { m.Upgrade.File = "../x" },
		"short sha256":        func(m *Manifest) { m.Genesis.SHA256 = "abcd" },
		"uppercase sha256":    func(m *Manifest) { m.Genesis.SHA256 = strings.ToUpper(m.Genesis.SHA256) },
		"no version":          func(m *Manifest) { m.Cosmovisor.Version = "" },
		"no size":             func(m *Manifest) { m.Upgrade.Size = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			files, m := testPayload(t)
			corrupt(&m)
			setManifest(t, files, m)
			if _, err := loadManifest(files); err == nil {
				t.Fatal("loadManifest() accepted an invalid manifest")
			}
		})
	}
}

func TestLoadManifestWithoutPayload(t *testing.T) {
	_, err := loadManifest(fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "without a payload") {
		t.Fatalf("loadManifest() error = %v, want a missing payload error", err)
	}
}

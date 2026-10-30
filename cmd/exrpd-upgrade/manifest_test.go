package main

import (
	"reflect"
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
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loadManifest() = %+v, want %+v", got, want)
	}
}

func TestLoadManifestRejectsInvalid(t *testing.T) {
	for name, corrupt := range map[string]func(*Manifest){
		"other platform":          func(m *Manifest) { m.Platform = "plan9/mips" },
		"no upgrade":              func(m *Manifest) { m.Upgrades = nil },
		"plan with separator":     func(m *Manifest) { m.Upgrades[0].Plan = "../v12" },
		"upper-case plan":         func(m *Manifest) { m.Upgrades[0].Plan = "V12.0.0" },
		"plan not vX.Y.Z":         func(m *Manifest) { m.Genesis.Plan = "v11" },
		"upgrade before genesis":  func(m *Manifest) { m.Upgrades[0].Plan = "v11.2.0" },
		"upgrade same as genesis": func(m *Manifest) { m.Upgrades[0].Plan = m.Genesis.Plan },
		"upgrades out of order": func(m *Manifest) {
			m.Upgrades = append(m.Upgrades, Release{Plan: "v11.4.0", Binary: m.Upgrades[0].Binary})
			m.Upgrades[1].File = "other.gz"
		},
		"file listed twice":   func(m *Manifest) { m.Upgrades[0].File = m.Genesis.File },
		"file with separator": func(m *Manifest) { m.Upgrades[0].File = "../x" },
		"short sha256":        func(m *Manifest) { m.Genesis.SHA256 = "abcd" },
		"uppercase sha256":    func(m *Manifest) { m.Genesis.SHA256 = strings.ToUpper(m.Genesis.SHA256) },
		"no version":          func(m *Manifest) { m.Cosmovisor.Version = "" },
		"no size":             func(m *Manifest) { m.Upgrades[0].Size = 0 },
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

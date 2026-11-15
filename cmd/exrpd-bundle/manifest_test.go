package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadBundle(t *testing.T) {
	for name, payload := range map[string]func(*testing.T) (fstest.MapFS, Manifest){
		"two-binary": testPayload,
		"single":     testSinglePayload,
		"multi-plan": testMultiPayload,
	} {
		t.Run(name, func(t *testing.T) {
			files, want := payload(t)
			b, err := loadBundle(files)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(b.m, want) {
				t.Fatalf("loadBundle() manifest = %+v, want %+v", b.m, want)
			}
		})
	}
}

func TestLoadBundleRejectsInvalid(t *testing.T) {
	// extra is a valid second upgrade for plan, with its own file.
	extra := func(m *Manifest, plan string) Release {
		return Release{Plan: plan, Binary: Binary{Version: "11.4.0", File: "exrpd-upgrade-1.gz", SHA256: m.Genesis.SHA256, Size: 1}}
	}
	for name, tc := range map[string]struct {
		corrupt func(*Manifest)
		want    string
	}{
		"other platform":                 {func(m *Manifest) { m.Platform = "plan9/mips" }, "payload is for plan9/mips"},
		"no launcher commit":             {func(m *Manifest) { m.Launcher.Commit = "" }, "no launcher.commit"},
		"empty upgrade":                  {func(m *Manifest) { m.Upgrades[0] = Release{} }, `invalid plan name ""`},
		"plan with separator":            {func(m *Manifest) { m.Upgrades[0].Plan = "../v12" }, `invalid plan name "../v12"`},
		"upper-case plan":                {func(m *Manifest) { m.Upgrades[0].Plan = "V12.0.0" }, `invalid plan name "V12.0.0"`},
		"plan not vX.Y.Z":                {func(m *Manifest) { m.Genesis.Plan = "v11" }, `invalid plan name "v11"`},
		"upgrade before genesis":         {func(m *Manifest) { m.Upgrades[0].Plan = "v11.2.0" }, "plan v11.2.0 does not follow the genesis plan"},
		"upgrade same as genesis":        {func(m *Manifest) { m.Upgrades[0].Plan = m.Genesis.Plan }, "plan v11.3.0 does not follow the genesis plan"},
		"file listed twice":              {func(m *Manifest) { m.Upgrades[0].File = m.Genesis.File }, "manifest lists exrpd-genesis.gz twice"},
		"file with separator":            {func(m *Manifest) { m.Upgrades[0].File = "../x" }, "invalid manifest entry"},
		"short sha256":                   {func(m *Manifest) { m.Genesis.SHA256 = "abcd" }, "invalid sha256"},
		"uppercase sha256":               {func(m *Manifest) { m.Genesis.SHA256 = strings.ToUpper(m.Genesis.SHA256) }, "invalid sha256"},
		"no version":                     {func(m *Manifest) { m.Cosmovisor.Version = "" }, "invalid manifest entry"},
		"no size":                        {func(m *Manifest) { m.Upgrades[0].Size = 0 }, "invalid manifest entry"},
		"second upgrade before genesis":  {func(m *Manifest) { m.Upgrades = append(m.Upgrades, extra(m, "v11.2.0")) }, "plan v11.2.0 does not follow the genesis plan"},
		"upgrade plan listed twice":      {func(m *Manifest) { m.Upgrades = append(m.Upgrades, extra(m, testPlan)) }, "manifest lists plan v12.0.0 twice"},
		"second upgrade plan not vX.Y.Z": {func(m *Manifest) { m.Upgrades = append(m.Upgrades, extra(m, "v11.4")) }, `invalid plan name "v11.4"`},
		"second upgrade file listed twice": {func(m *Manifest) {
			m.Upgrades = append(m.Upgrades, extra(m, testMidPlan))
			m.Upgrades[1].File = m.Upgrades[0].File
		}, "manifest lists exrpd-upgrade.gz twice"},
	} {
		t.Run(name, func(t *testing.T) {
			files, m := testPayload(t)
			tc.corrupt(&m)
			setManifest(t, files, m)
			if _, err := loadBundle(files); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("loadBundle() error = %v, want %q", err, tc.want)
			}
		})
	}
}

// Upgrades may come in any order. The bundle sorts them by plan number, not by text:
// v12.0.0 comes after v3.0.0. Messages and the version line must list the plans in that
// order.
func TestLoadBundleOrdersReleasesByPlanNumber(t *testing.T) {
	files, m := testPayload(t)
	delete(files, m.Upgrades[0].File)
	m.Genesis.Plan = "v2.0.0"
	m.Upgrades = []Release{
		{Plan: "v12.0.0", Binary: addBinary(t, files, "exrpd-upgrade-0", "12.0.0", "v12 bytes")},
		{Plan: "v3.0.0", Binary: addBinary(t, files, "exrpd-upgrade-1", "3.0.0", "v3 bytes")},
	}
	setManifest(t, files, m)
	files[manifestFile].Data = append([]byte(" \n\t"), files[manifestFile].Data...)
	rawManifest := files[manifestFile].Data
	b := mustLoadBundle(t, files)

	if got := b.plans(); got != "v2.0.0, v3.0.0, v12.0.0" {
		t.Errorf("plans() = %q, want genesis then the upgrades by plan number", got)
	}
	if got := formatVersion(b, "x"); got != "x - genesis 11.3.0 - upgrade 3.0.0 12.0.0 - cosmovisor 1.7.3" {
		t.Errorf("formatVersion() = %q, want the upgrades by plan number", got)
	}

	n := testNode(b, t.TempDir())
	sum := sha256.Sum256(rawManifest)
	writeFile(t, n.Marker(), hex.EncodeToString(sum[:])+"\n", 0o644)
	rawReport, err := b.bundleReport(n)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Manifest        Manifest
		InstalledBundle string `json:"installed_bundle"`
		Dirs            []string
	}
	if err := json.Unmarshal([]byte(rawReport), &report); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Manifest, m) {
		t.Errorf("version --bundle manifest = %+v, want original unsorted manifest %+v", report.Manifest, m)
	}
	if report.InstalledBundle != "this" {
		t.Errorf("version --bundle installed_bundle = %q, want this for SHA-256 of original manifest bytes", report.InstalledBundle)
	}
	wantDirs := []string{
		filepath.Join(n.Root, "upgrades", "v2.0.0"),
		filepath.Join(n.Root, "upgrades", "v3.0.0"),
		filepath.Join(n.Root, "upgrades", "v12.0.0"),
		filepath.Dir(n.Cosmovisor),
	}
	if !reflect.DeepEqual(report.Dirs, wantDirs) {
		t.Errorf("version --bundle dirs = %q, want %q", report.Dirs, wantDirs)
	}
}

func TestLoadBundleWithoutPayload(t *testing.T) {
	_, err := loadBundle(fstest.MapFS{})
	if err == nil || !strings.Contains(err.Error(), "without a payload") {
		t.Fatalf("loadBundle() error = %v, want a missing payload error", err)
	}
}

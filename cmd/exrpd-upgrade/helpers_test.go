package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"testing/fstest"
)

const (
	testGenesisPlan = "v11.3.0"
	testPlan        = "v12.0.0"
	testGenesisDir  = "upgrades/" + testGenesisPlan // `current` target of the genesis plan
	testUpgradeDir  = "upgrades/" + testPlan        // `current` target of the upgrade binary
)

// testPayload returns a payload for the running platform whose binaries contain the
// given text, and its manifest.
func testPayload(t *testing.T) (fstest.MapFS, Manifest) {
	t.Helper()
	files := fstest.MapFS{}
	add := func(name, version, content string) Binary {
		sum := sha256.Sum256([]byte(content))
		files[name+".gz"] = &fstest.MapFile{Data: gzipped(t, content)}
		return Binary{Version: version, File: name + ".gz", SHA256: hex.EncodeToString(sum[:]), Size: int64(len(content)), Source: "test"}
	}
	m := Manifest{
		Platform:   runtime.GOOS + "/" + runtime.GOARCH,
		Genesis:    Release{Plan: testGenesisPlan, Binary: add("exrpd-genesis", "11.3.0", "v11 bytes")},
		Upgrades:   []Release{{Plan: testPlan, Binary: add("exrpd-upgrade", "12.0.0", "v12 bytes")}},
		Cosmovisor: add("cosmovisor", "v1.7.3", "cosmovisor bytes"),
	}
	setManifest(t, files, m)
	return files, m
}

func setManifest(t *testing.T, files fstest.MapFS, m Manifest) {
	t.Helper()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	files[manifestFile] = &fstest.MapFile{Data: raw}
}

func gzipped(t *testing.T, content string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// writeFile creates path with content, and its parent directories.
func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

// readFile returns the content of path, or "" if it does not exist.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// upgradeInfoJSON is an upgrade-info.json as x/upgrade writes it.
func upgradeInfoJSON(plan string) string {
	return upgradeInfoAt(plan, 40)
}

// upgradeInfoAt is upgradeInfoJSON for a plan at height.
func upgradeInfoAt(plan string, height int) string {
	return `{"name":"` + plan + `","time":"0001-01-01T00:00:00Z","height":` + strconv.Itoa(height) + `}`
}

func writeDataUpgradeInfo(t *testing.T, home, content string) {
	t.Helper()
	writeFile(t, filepath.Join(home, "data", "upgrade-info.json"), content, 0o600)
}

// linkCurrent points <home>/cosmovisor/current at target, as Cosmovisor would.
func linkCurrent(t *testing.T, home, target string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, "cosmovisor", target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(home, "cosmovisor", "current")); err != nil {
		t.Fatal(err)
	}
}

// currentTarget returns where <home>/cosmovisor/current points, or "" if it is missing.
func currentTarget(t *testing.T, home string) string {
	t.Helper()
	target, err := os.Readlink(filepath.Join(home, "cosmovisor", "current"))
	if errors.Is(err, fs.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatal(err)
	}
	return target
}

// danglingCurrent points <home>/cosmovisor/current at target without creating target.
func danglingCurrent(t *testing.T, home, target string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(home, "cosmovisor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(home, "cosmovisor", "current")); err != nil {
		t.Fatal(err)
	}
}

// captureStderr runs f with os.Stderr sent to a file and returns what f wrote there.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stderr")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	stderr := os.Stderr
	os.Stderr = file
	defer func() {
		os.Stderr = stderr
		file.Close()
	}()
	f()
	return readFile(t, path)
}

// testUserHomeDir is os.UserHomeDir for tests.
func testUserHomeDir() (string, error) { return "/users/val", nil }

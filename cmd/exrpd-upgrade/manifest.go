package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
)

// payload holds manifest.json and the gzip-compressed binaries it lists, for exactly
// one GOOS/GOARCH. scripts/upgrade-binary/build.sh fills it in before "go build". The
// committed directory is empty.
//
//go:embed all:payload
var payload embed.FS

const manifestFile = "manifest.json"

// Manifest describes the payload. Plan is the on-chain upgrade plan that switches the
// node from Genesis to Upgrade.
type Manifest struct {
	Platform   string `json:"platform"`
	Plan       string `json:"plan"`
	Genesis    Binary `json:"genesis"`
	Upgrade    Binary `json:"upgrade"`
	Cosmovisor Binary `json:"cosmovisor"`
}

// Binary is one embedded executable. File is its gzip-compressed copy in the payload.
// SHA256 and Size describe the executable after decompression.
type Binary struct {
	Version string `json:"version"`
	File    string `json:"file"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
	Source  string `json:"source"`
}

func embeddedPayload() fs.FS {
	sub, err := fs.Sub(payload, "payload")
	if err != nil {
		panic(err) // "payload" is a valid path: fs.Sub cannot fail here.
	}
	return sub
}

// loadManifest reads and validates the manifest of p. Names in it become path elements
// under the node home, so they must not contain separators.
func loadManifest(p fs.FS) (Manifest, error) {
	raw, err := fs.ReadFile(p, manifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		return Manifest{}, errors.New("this exrpd-upgrade was built without a payload. Build it with scripts/upgrade-binary/build.sh")
	}
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse %s: %w", manifestFile, err)
	}
	return m, m.validate()
}

func (m Manifest) validate() error {
	if running := runtime.GOOS + "/" + runtime.GOARCH; m.Platform != running {
		return fmt.Errorf("payload is for %s, running on %s", m.Platform, running)
	}
	if !isPathElement(m.Plan) {
		return fmt.Errorf("invalid upgrade plan name %q", m.Plan)
	}
	for _, b := range m.binaries() {
		if err := b.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (m Manifest) binaries() []Binary {
	return []Binary{m.Genesis, m.Upgrade, m.Cosmovisor}
}

func (b Binary) validate() error {
	if !isPathElement(b.File) || b.Version == "" || b.Size <= 0 {
		return fmt.Errorf("invalid manifest entry %+v", b)
	}
	// Lowercase only: installBinary compares it with hex.EncodeToString output.
	if sum, err := hex.DecodeString(b.SHA256); err != nil || len(sum) != sha256.Size || hex.EncodeToString(sum) != b.SHA256 {
		return fmt.Errorf("invalid sha256 %q for %s: want 64 lowercase hex digits", b.SHA256, b.File)
	}
	return nil
}

func isPathElement(s string) bool {
	return s != "" && s != "." && s != ".." && filepath.Base(s) == s
}

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
	"slices"
	"strconv"
	"strings"
)

// payload holds manifest.json and the gzip-compressed binaries it lists, for exactly
// one GOOS/GOARCH. scripts/upgrade-binary/build.sh fills it in before "go build". The
// committed directory is empty.
//
//go:embed all:payload
var payload embed.FS

const manifestFile = "manifest.json"

// Manifest describes the payload. Genesis runs the data of its own plan and the data of
// a home that has no data/upgrade-info.json yet. Each upgrade runs the data of its plan.
// Upgrades are in plan order, after the genesis plan.
type Manifest struct {
	Platform   string    `json:"platform"`
	Genesis    Release   `json:"genesis"`
	Upgrades   []Release `json:"upgrades"`
	Cosmovisor Binary    `json:"cosmovisor"`
}

// Release is an exrpd binary and the on-chain plan whose data it runs. The plans of this
// chain are named v<release> after the release that adds their handler, and patch
// releases keep the plan of their release.
type Release struct {
	Plan string `json:"plan"`
	Binary
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

// loadManifest reads and validates the manifest of p.
func loadManifest(p fs.FS) (Manifest, error) {
	raw, err := fs.ReadFile(p, manifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		return Manifest{}, errors.New("this exrpd was built without a payload. Build it with scripts/upgrade-binary/build.sh")
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

// validate checks what the launcher relies on. Plans and files become path elements under
// the node home. Plans are lower case because Cosmovisor lowercases the plan it reads
// from data/upgrade-info.json and switches to upgrades/<that name>.
func (m Manifest) validate() error {
	if running := runtime.GOOS + "/" + runtime.GOARCH; m.Platform != running {
		return fmt.Errorf("payload is for %s, running on %s", m.Platform, running)
	}
	if len(m.Upgrades) == 0 {
		return errors.New("manifest has no upgrade")
	}
	releases := m.releases()
	for i, r := range releases {
		if !isPathElement(r.Plan) || r.Plan != strings.ToLower(r.Plan) {
			return fmt.Errorf("invalid plan name %q: want a lower-case directory name", r.Plan)
		}
		parts, ok := planParts(r.Plan)
		if !ok {
			return fmt.Errorf("invalid plan name %q: want vX.Y.Z", r.Plan)
		}
		if i > 0 {
			if prev, _ := planParts(releases[i-1].Plan); slices.Compare(prev, parts) >= 0 {
				return fmt.Errorf("plan %s does not follow plan %s", r.Plan, releases[i-1].Plan)
			}
		}
	}
	files := map[string]bool{}
	for _, b := range m.binaries() {
		if err := b.validate(); err != nil {
			return err
		}
		if files[b.File] {
			return fmt.Errorf("manifest lists %s twice", b.File)
		}
		files[b.File] = true
	}
	return nil
}

// releases returns the genesis release and the upgrades, in plan order.
func (m Manifest) releases() []Release {
	return append([]Release{m.Genesis}, m.Upgrades...)
}

// release returns the release whose plan is plan. Plan names match case-insensitively,
// like Cosmovisor matches them.
func (m Manifest) release(plan string) (Release, bool) {
	for _, r := range m.releases() {
		if strings.EqualFold(r.Plan, plan) {
			return r, true
		}
	}
	return Release{}, false
}

// plans lists the plans this bundle runs, for messages.
func (m Manifest) plans() string {
	names := make([]string, 0, len(m.Upgrades)+1)
	for _, r := range m.releases() {
		names = append(names, r.Plan)
	}
	return strings.Join(names, ", ")
}

func (m Manifest) binaries() []Binary {
	bins := []Binary{m.Genesis.Binary}
	for _, r := range m.Upgrades {
		bins = append(bins, r.Binary)
	}
	return append(bins, m.Cosmovisor)
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

// planParts parses a plan name "vX.Y.Z" into its three numbers.
func planParts(plan string) ([]int, bool) {
	fields := strings.Split(strings.TrimPrefix(plan, "v"), ".")
	if !strings.HasPrefix(plan, "v") || len(fields) != 3 {
		return nil, false
	}
	parts := make([]int, 0, len(fields))
	for _, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil || n < 0 || f != strconv.Itoa(n) {
			return nil, false
		}
		parts = append(parts, n)
	}
	return parts, true
}

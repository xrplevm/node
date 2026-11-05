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
// one GOOS/GOARCH. scripts/bundle/build.sh fills it in before "go build". The
// committed directory is empty.
//
//go:embed all:payload
var payload embed.FS

func embeddedPayload() fs.FS {
	sub, err := fs.Sub(payload, "payload")
	if err != nil {
		panic(err) // "payload" is a valid path: fs.Sub cannot fail here.
	}
	return sub
}

const manifestFile = "manifest.json"

// Manifest describes the payload. Genesis runs the data of its own plan and the data of
// a home that has no data/upgrade-info.json yet. Each of Upgrades runs the data of its
// plan, which follows the genesis plan. Cosmovisor picks the upgrade by the plan name in
// data/upgrade-info.json, so the order of Upgrades does not matter.
type Manifest struct {
	Platform   string    `json:"platform"`
	Launcher   Launcher  `json:"launcher"`
	Genesis    Release   `json:"genesis"`
	Upgrades   []Release `json:"upgrades,omitempty"`
	Cosmovisor Binary    `json:"cosmovisor"`
}

// Launcher is the build of this launcher: the git commit it was built from, with the
// suffix -dirty when it had uncommitted changes.
type Launcher struct {
	Commit string `json:"commit"`
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

// Release is an exrpd binary and the on-chain plan whose data it runs. The plans of this
// chain are named v<release> after the release that adds their handler, and patch
// releases keep the plan of their release.
type Release struct {
	Plan string `json:"plan"`
	Binary
}

// dir is the directory of r's binary, upgrades/<plan>, relative to Root, as Cosmovisor
// names it.
func (r Release) dir() string {
	return filepath.Join("upgrades", r.Plan)
}

// bundle is the embedded payload: its files, its manifest and the digest that the marker records.
type bundle struct {
	files        fs.FS
	m            Manifest
	digest       string
	releases     []Release
	genesisParts [3]int
}

// loadBundle reads and validates the manifest of files. The digest is the lowercase hex
// SHA-256 of manifest.json, the content of the marker without its newline.
func loadBundle(files fs.FS) (bundle, error) {
	raw, err := fs.ReadFile(files, manifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		return bundle{}, errors.New("this exrpd was built without a payload. Build it with scripts/bundle/build.sh")
	}
	if err != nil {
		return bundle{}, err
	}

	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return bundle{}, fmt.Errorf("parse %s: %w", manifestFile, err)
	}

	b := bundle{files: files, m: m}
	if err := b.validate(); err != nil {
		return bundle{}, err
	}

	sum := sha256.Sum256(raw)
	b.digest = hex.EncodeToString(sum[:])
	return b, nil
}

// validate checks what the launcher relies on. Plans and files become path elements under
// the node home. Plans are lower case because Cosmovisor lowercases the plan it reads
// from data/upgrade-info.json and switches to upgrades/<that name>.
func (b *bundle) validate() error {
	m := b.m
	if running := runtime.GOOS + "/" + runtime.GOARCH; m.Platform != running {
		return fmt.Errorf("payload is for %s, running on %s", m.Platform, running)
	}
	if m.Launcher.Commit == "" {
		return errors.New("manifest has no launcher.commit")
	}

	type plannedRelease struct {
		release Release
		parts   [3]int
	}
	planned := make([]plannedRelease, 1, len(m.Upgrades)+1)
	planned[0].release = m.Genesis
	for _, r := range m.Upgrades {
		planned = append(planned, plannedRelease{release: r})
	}

	for i, r := range planned {
		if !isPathElement(r.release.Plan) || r.release.Plan != strings.ToLower(r.release.Plan) {
			return fmt.Errorf("invalid plan name %q: want a lower-case directory name", r.release.Plan)
		}
		parts, ok := planParts(r.release.Plan)
		if !ok {
			return fmt.Errorf("invalid plan name %q: want vX.Y.Z", r.release.Plan)
		}
		planned[i].parts = parts
	}

	slices.SortStableFunc(planned[1:], func(a, b plannedRelease) int {
		return slices.Compare(a.parts[:], b.parts[:])
	})

	for i := 1; i < len(planned); i++ {
		if slices.Compare(planned[i-1].parts[:], planned[i].parts[:]) < 0 {
			continue
		}
		if i == 1 {
			return fmt.Errorf("plan %s does not follow the genesis plan %s", planned[i].release.Plan, m.Genesis.Plan)
		}
		return fmt.Errorf("manifest lists plan %s twice", planned[i].release.Plan)
	}

	files := map[string]bool{}
	for i := 0; i <= len(planned); i++ {
		binary := m.Cosmovisor
		if i < len(planned) {
			binary = planned[i].release.Binary
		}
		if err := binary.validate(); err != nil {
			return err
		}
		if files[binary.File] {
			return fmt.Errorf("manifest lists %s twice", binary.File)
		}
		files[binary.File] = true
	}

	b.releases = make([]Release, len(planned))
	for i, r := range planned {
		b.releases[i] = r.release
	}
	b.genesisParts = planned[0].parts
	return nil
}

// single reports whether the bundle runs only its genesis release.
func (b bundle) single() bool {
	return len(b.releases) == 1
}

// plans lists the plans this bundle runs, for messages.
func (b bundle) plans() string {
	names := make([]string, 0, len(b.releases))
	for _, r := range b.releases {
		names = append(names, r.Plan)
	}
	return strings.Join(names, ", ")
}

// releaseIn returns the release whose directory is dir, relative to Root.
func (b bundle) releaseIn(dir string) (Release, bool) {
	for _, r := range b.releases {
		if r.dir() == dir {
			return r, true
		}
	}
	return Release{}, false
}

// releaseFor returns the release of plan. Plan names match case-insensitively, like
// Cosmovisor matches them.
func (b bundle) releaseFor(plan string) (Release, bool) {
	for _, r := range b.releases {
		if strings.EqualFold(r.Plan, plan) {
			return r, true
		}
	}
	return Release{}, false
}

func isPathElement(s string) bool {
	return s != "" && s != "." && s != ".." && filepath.Base(s) == s
}

// planParts parses a plan name "vX.Y.Z" into its three numbers.
func planParts(plan string) ([3]int, bool) {
	fields := strings.Split(strings.TrimPrefix(plan, "v"), ".")
	if !strings.HasPrefix(plan, "v") || len(fields) != 3 {
		return [3]int{}, false
	}

	var parts [3]int
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil || n < 0 || f != strconv.Itoa(n) {
			return [3]int{}, false
		}
		parts[i] = n
	}

	return parts, true
}

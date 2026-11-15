package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// The marker records which bundle installed the binaries of a home. It holds the
// SHA-256 of the embedded manifest.json, which lists every binary with its hash, so it
// names the whole bundle. An installer deletes it before the first binary write and
// writes it after the last one, so an interrupted install leaves no marker. Commands
// other than the installers run only when the marker names this bundle.

// installState is the bundle that the marker of a home names: this one, none, or another
// one. The values are what version --bundle shows as installed_bundle.
type installState string

const (
	installMatch   installState = "this"  // the marker names this bundle
	installMissing installState = "none"  // no marker: nothing installed, an older bundle, or an interrupted install
	installDiffers installState = "other" // the marker names another bundle
)

// Marker is the file that names the bundle whose binaries are installed.
func (l Layout) Marker() string {
	return filepath.Join(l.Root, "bundle.sha256")
}

// readMarker returns the digest in the marker of l, or "" when there is no marker.
func readMarker(l Layout) (string, error) {
	raw, err := os.ReadFile(l.Marker())
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// writeMarker records digest as the bundle installed in l.
func writeMarker(l Layout, digest string) error {
	return writeFileAtomic(l.Marker(), strings.NewReader(digest+"\n"), 0o644, nil)
}

// removeMarker deletes the marker of l and syncs its directory, so the deletion is on
// disk before the first binary changes. A missing marker is not an error.
func removeMarker(l Layout) error {
	err := os.Remove(l.Marker())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return syncDir(l.Root)
}

// installStateOf compares the marker of l with the digest of b.
func (b bundle) installStateOf(l Layout) (installState, error) {
	marker, err := readMarker(l)
	switch {
	case err != nil:
		return "", err
	case marker == "":
		return installMissing, nil
	case marker == b.digest:
		return installMatch, nil
	default:
		return installDiffers, nil
	}
}

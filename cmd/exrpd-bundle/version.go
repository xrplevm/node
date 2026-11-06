package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"strings"
)

// runVersion prints the version of the exrpd that runs the node data and of the bundle.
// It never installs.
func runVersion(b bundle, n node) error {
	line, err := b.versionLine(n)
	if err != nil {
		return err
	}
	fmt.Println(line)
	return nil
}

// runVersionBundle prints the manifest, the directories the bundle uses, the home, the
// exrpd that commands run and the install state. It never installs.
func runVersionBundle(b bundle, n node) error {
	out, err := b.bundleReport(n)
	if out != "" {
		fmt.Println(out)
	}
	return err
}

// versionLine returns the version line. It starts with the version of the installed
// exrpd when commands run it, else with "not-installed" or "restart-needed".
func (b bundle) versionLine(n node) (string, error) {
	st, err := b.inspect(n)
	if err != nil {
		return "", err
	}
	switch {
	case st.refusal == nil:
	case st.refusal.kind == refusedHome:
		return "", st.refusal
	default:
		return formatVersion(b, st.refusal.kind.lineLabel()), nil
	}
	installed, err := installedVersion(n.Layout, st.decision.release)
	if err != nil {
		return "", err
	}
	return formatVersion(b, installed), nil
}

// bundleReport returns the JSON of version --bundle. It tells an operator which home is
// used and why, and whether this bundle is installed there, also when nothing is
// installed or decide refuses the home: then it has the refusal, and the error is that
// refusal. It never runs a chain binary.
func (b bundle) bundleReport(n node) (string, error) {
	l := n.Layout
	st, err := b.inspect(n)
	if err != nil {
		return "", err
	}
	version := "unknown"
	var run *willRun
	var refused *string
	var homeErr error
	if st.refusal != nil && st.refusal.kind == refusedHome {
		msg := st.refusal.Error()
		refused, homeErr = &msg, st.refusal
	} else {
		if run, err = willRunOf(l, st.decision); err != nil {
			return "", err
		}
		version = run.Version
	}
	dirPaths := l.dirs(b)
	out, err := json.MarshalIndent(struct {
		Version         string   `json:"version"`
		Home            string   `json:"home"`
		HomeSource      string   `json:"home_source"`
		InstalledBundle string   `json:"installed_bundle"`
		WillRun         *willRun `json:"will_run"`
		Refused         *string  `json:"refused"`
		Dirs            []string `json:"dirs"`
		Manifest        Manifest `json:"manifest"`
	}{formatVersion(b, version), n.Home, n.Source, string(st.marker), run, refused, dirPaths, b.m}, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), homeErr
}

// formatVersion is the version line of the bundle with installed as the version of the
// installed exrpd. A single bundle has no upgrade part. A bundle with more than one
// upgrade lists them in plan order, for example "upgrade 8.0.2 9.0.3".
func formatVersion(b bundle, installed string) string {
	upgrade := ""
	if !b.single() {
		upgrade = " - upgrade"
		for _, r := range b.releases[1:] {
			upgrade += " " + r.Version
		}
	}
	return fmt.Sprintf("%s - genesis %s%s - cosmovisor %s",
		installed, b.m.Genesis.Version, upgrade, strings.TrimPrefix(b.m.Cosmovisor.Version, "v"))
}

// willRun is the exrpd that commands of the home run, and that start points `current` at.
type willRun struct {
	Dir     string `json:"dir"` // relative to <home>/cosmovisor
	Binary  string `json:"binary"`
	Version string `json:"version"` // "unknown" for a binary that is not the bundled one
	Exists  bool   `json:"exists"`
	Reason  string `json:"reason"`
}

// willRunOf describes the exrpd of d without running it, so version --bundle works also
// when that binary is broken or is this bundle. A missing exrpd gets the bundled
// version, the one that start installs there.
func willRunOf(l Layout, d decision) (*willRun, error) {
	matches, err := bundledMatch(l, d.release)
	missing := errors.Is(err, fs.ErrNotExist)
	if err != nil && !missing {
		return nil, err
	}
	run := &willRun{Dir: d.release.dir(), Binary: l.bin(d.release.dir()), Version: "unknown", Exists: !missing, Reason: d.reason}
	if missing || matches {
		run.Version = d.release.Version
	}
	return run, nil
}

// bundledMatch reports whether the exrpd of r has the hash of the bundled binary. It
// never runs a binary. A missing exrpd is an error that wraps fs.ErrNotExist.
func bundledMatch(l Layout, r Release) (bool, error) {
	sum, err := fileSHA256(l.bin(r.dir()))
	if err != nil {
		return false, err
	}
	return sum == r.SHA256, nil
}

// installedVersion returns the version of the exrpd of r, the one commands run. The
// caller has checked that the marker names this bundle. It is the bundled version when
// that binary has the bundled hash. A binary replaced by hand tells its version itself.
func installedVersion(l Layout, r Release) (string, error) {
	matches, err := bundledMatch(l, r)
	if err != nil {
		return "", err
	}
	if matches {
		return r.Version, nil
	}
	bin := l.bin(r.dir())
	if err := refuseSelf(bin); err != nil {
		return "", err
	}
	out, err := exec.Command(bin, "--home", l.Home, "version").Output()
	if exitErr := (*exec.ExitError)(nil); errors.As(err, &exitErr) && len(bytes.TrimSpace(exitErr.Stderr)) > 0 {
		return "", fmt.Errorf("%s version: %w: %s", bin, err, bytes.TrimSpace(exitErr.Stderr))
	}
	if err != nil {
		return "", fmt.Errorf("%s version: %w", bin, err)
	}
	return strings.TrimSpace(string(out)), nil
}

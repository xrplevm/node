package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

// runStart prepares the home and replaces this process with Cosmovisor. Exec keeps the
// PID, so the service manager supervises Cosmovisor directly. args are passed through
// to "exrpd start". The log names the home, so a service that runs on another home than
// the operator expects shows it.
func runStart(home, source string, args []string) error {
	if err := refuseCosmovisorConfig(args); err != nil {
		return err
	}
	env, err := cosmovisorEnv(os.Environ(), home)
	if err != nil {
		return err
	}
	logf("home %s (from %s)", home, source)
	l, err := prepare(embeddedPayload(), home)
	if err != nil {
		return err
	}
	argv := append([]string{"cosmovisor", "run", "start", "--home", home}, args...)
	return syscall.Exec(l.Cosmovisor, argv, env) //nolint:gosec // the Cosmovisor that prepare staged and verified
}

// prepare installs the embedded binaries under home and points `current` at the binary
// the node data needs. decide checks the home before the first file is written.
func prepare(files fs.FS, home string) (Layout, error) {
	m, err := loadManifest(files)
	if err != nil {
		return Layout{}, err
	}
	l := newLayout(home, m)
	d, err := decide(home, l, m)
	if err != nil {
		return Layout{}, err
	}
	if err := install(files, m, l); err != nil {
		return Layout{}, err
	}
	return l, apply(l, d)
}

// runExrpd runs any other command with the exrpd that runs the node data, without
// changing `current`: the node may be running. It installs the binaries only when that
// exrpd is missing, for example before the first start. --home comes first, so exrpd uses
// the home the launcher resolved and no flag of the command can take it as its value.
func runExrpd(home string, args []string) error {
	bin, err := exrpdFor(embeddedPayload(), home, true)
	if err != nil {
		return err
	}
	return syscall.Exec(bin, append([]string{"exrpd", "--home", home}, args...), os.Environ())
}

// runCompletion runs a shell completion request with the exrpd for home, when that exrpd
// is installed. It never installs: a tab key on a machine without a node home must not
// write the bundle. Without the exrpd there is nothing to suggest. A refusal of the home
// goes to stderr, which completion scripts discard. args start with the __complete word.
// --home follows it, so exrpd completes from the same home, also when an alias put
// --home before __complete.
func runCompletion(home string, args []string) error {
	bin, err := exrpdFor(embeddedPayload(), home, false)
	if err != nil {
		return err
	}
	if bin == "" {
		return nil
	}
	argv := append([]string{"exrpd", args[0], "--home", home}, args[1:]...)
	return syscall.Exec(bin, argv, os.Environ())
}

// runHelp runs a help request with the exrpd for home when that exrpd is installed. It
// never installs: help on a machine without a node home must not write the bundle.
func runHelp(home string, args []string) error {
	bin, err := exrpdFor(embeddedPayload(), home, false)
	if err != nil {
		return err
	}
	if bin != "" {
		return syscall.Exec(bin, append([]string{"exrpd", "--home", home}, args...), os.Environ())
	}
	fmt.Printf("exrpd: no exrpd is installed in %s yet.\n"+
		"  New node: run `exrpd init <moniker> --chain-id <chain id>` to create this home.\n"+
		"  Existing node: this is probably not your node's home. Pass it with --home, or set it the same way your service does. `exrpd version --bundle` shows which home is used and why.\n",
		home)
	return nil
}

// exrpdFor returns the exrpd that runs the data of home. When it is missing, it installs
// the binaries if installMissing is set, and otherwise returns "".
func exrpdFor(files fs.FS, home string, installMissing bool) (string, error) {
	m, err := loadManifest(files)
	if err != nil {
		return "", err
	}
	l := newLayout(home, m)
	d, err := decide(home, l, m)
	if err != nil {
		return "", err
	}
	bin := l.bin(d.dir)
	if _, err := os.Stat(bin); errors.Is(err, fs.ErrNotExist) {
		if !installMissing {
			return "", nil
		}
		if err := install(files, m, l); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	return bin, refuseSelf(bin)
}

// refuseSelf refuses to run bin when it is this bundle. With stock Cosmovisor,
// bin/exrpd is often a symlink to /usr/local/bin/exrpd, where the bundle is installed
// too: running it would run the launcher again, without end. `exrpd start` replaces it
// with the chain binary.
func refuseSelf(bin string) error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find this exrpd: %w", err)
	}
	if !sameDir(self, bin) {
		return nil
	}
	return fmt.Errorf("%s is this exrpd bundle, not a chain binary: run exrpd start once to install the chain binary there", bin)
}

// runVersion prints the version of the exrpd that runs the node data and of the bundle.
// --bundle prints the manifest, the directories the bundle uses, the home and the exrpd
// that commands run. Any other flag, such as --long, goes to the version command of that
// exrpd.
func runVersion(home, source string, args []string) error {
	bundle, err := versionMode(args)
	if err != nil {
		return err
	}
	if !bundle && len(args) > 0 {
		return runExrpd(home, append([]string{"version"}, args...))
	}
	out, err := versionOutput(embeddedPayload(), home, source, bundle)
	if out != "" {
		fmt.Println(out)
	}
	return err
}

// versionMode reports whether version prints the bundle JSON. The bare word --bundle is the
// launcher's flag anywhere after version, and it takes no other flag.
func versionMode(args []string) (bundle bool, err error) {
	if !slices.Contains(args, "--bundle") {
		return false, nil
	}
	for _, a := range args {
		if a != "--bundle" {
			return false, fmt.Errorf("version --bundle takes no other flag, got %s", a)
		}
	}
	return true, nil
}

// versionOutput returns the version line, or the bundle JSON when bundle is set. The
// JSON tells an operator which home is used and why, also when decide refuses it: then
// it has the refusal, and the error is that refusal.
func versionOutput(files fs.FS, home, source string, bundle bool) (string, error) {
	m, err := loadManifest(files)
	if err != nil {
		return "", err
	}
	l := newLayout(home, m)
	d, refusal := decide(home, l, m)
	if !bundle {
		if refusal != nil {
			return "", refusal
		}
		installed, err := installedVersion(l, m, d.dir)
		if err != nil {
			return "", err
		}
		return versionLine(m, installed), nil
	}
	version := "unknown"
	var run *willRun
	var refused *string
	if refusal != nil {
		msg := refusal.Error()
		refused = &msg
	} else {
		if run, err = willRunOf(l, m, d); err != nil {
			return "", err
		}
		version = run.Version
	}
	keep := make([]string, 0, len(m.Upgrades)+3)
	for _, dir := range dirs(m) {
		keep = append(keep, filepath.Join(l.Root, dir))
	}
	out, err := json.MarshalIndent(struct {
		Version    string   `json:"version"`
		Home       string   `json:"home"`
		HomeSource string   `json:"home_source"`
		WillRun    *willRun `json:"will_run"`
		Refused    *string  `json:"refused"`
		Dirs       []string `json:"dirs"`
		Manifest   Manifest `json:"manifest"`
	}{versionLine(m, version), home, source, run, refused, keep, m}, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), refusal
}

func versionLine(m Manifest, installed string) string {
	return fmt.Sprintf("%s - genesis %s - upgrade %s - cosmovisor %s",
		installed, m.Genesis.Version, upgradeVersions(m), strings.TrimPrefix(m.Cosmovisor.Version, "v"))
}

// willRun is the exrpd that commands of the home run, and that start points `current` at.
type willRun struct {
	Dir       string `json:"dir"` // relative to <home>/cosmovisor
	Binary    string `json:"binary"`
	Version   string `json:"version"` // "unknown" for a binary that is not the bundled one
	Installed bool   `json:"installed"`
	Reason    string `json:"reason"`
}

// willRunOf describes the exrpd of d without running it, so version --bundle works also
// when that binary is broken or is this bundle.
func willRunOf(l Layout, m Manifest, d decision) (*willRun, error) {
	bin := l.bin(d.dir)
	sum, err := fileSHA256(bin)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	run := &willRun{Dir: d.dir, Binary: bin, Version: "unknown", Installed: err == nil, Reason: d.reason}
	if b := bundledBinary(m, d.dir); !run.Installed || sum == b.SHA256 {
		run.Version = b.Version
	}
	return run, nil
}

// bundledBinary returns the binary of the bundle for dir, a binary directory.
func bundledBinary(m Manifest, dir string) Binary {
	for _, r := range m.releases() {
		if dir == planDir(r.Plan) {
			return r.Binary
		}
	}
	return m.Genesis.Binary
}

// installedVersion returns the version of the exrpd in dir, the one commands run. It is
// the bundled version when that binary is not installed yet or has the bundled hash.
// Another binary is there when this bundle was installed but not started yet, for
// example a patch bundle over 12.0.0, or when another setup installed it: then that
// exrpd tells its version itself.
func installedVersion(l Layout, m Manifest, dir string) (string, error) {
	bundled := bundledBinary(m, dir)
	bin := l.bin(dir)
	sum, err := fileSHA256(bin)
	if errors.Is(err, fs.ErrNotExist) || (err == nil && sum == bundled.SHA256) {
		return bundled.Version, nil
	}
	if err != nil {
		return "", err
	}
	if err := refuseSelf(bin); err != nil {
		return "", err
	}
	out, err := exec.Command(bin, "--home", filepath.Dir(l.Root), "version").Output()
	if err != nil {
		return "", fmt.Errorf("%s version: %w", bin, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// upgradeVersions lists the upgrade releases of the bundle, in plan order.
func upgradeVersions(m Manifest) string {
	versions := make([]string, 0, len(m.Upgrades))
	for _, r := range m.Upgrades {
		versions = append(versions, r.Version)
	}
	return strings.Join(versions, ", ")
}

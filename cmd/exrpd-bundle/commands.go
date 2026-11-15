package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"syscall"
)

// bundleUsage is the help of "exrpd bundle".
const bundleUsage = `Usage: exrpd bundle setup [--home DIR]

setup installs the binaries of this exrpd in <home>/cosmovisor and records them in
<home>/cosmovisor/bundle.sha256. It never changes which binary runs the node data and
starts nothing. Run it while the node is stopped, for example before rollback or export
when a command says the bundle is not installed or differs.`

// runStart prepares the home and replaces this process with Cosmovisor. Exec keeps the
// PID, so the service manager supervises Cosmovisor directly. args are passed through
// to "exrpd start".
func runStart(b bundle, n node, args []string) error {
	if err := refuseSelectedSlot(); err != nil {
		return err
	}
	if err := refuseCosmovisorConfig(args); err != nil {
		return err
	}
	env, err := cosmovisorEnv(os.Environ(), n.Home)
	if err != nil {
		return err
	}
	if err := b.prepare(n); err != nil {
		return err
	}
	argv := append([]string{"cosmovisor", "run", "start", "--home", n.Home}, args...)
	return execBin(n.Cosmovisor, argv, env)
}

// prepare installs the embedded binaries in the home of n and points `current` at the
// binary the node data needs.
func (b bundle) prepare(n node) error {
	d, err := b.stage(n)
	if err != nil {
		return err
	}
	return apply(n.Layout, d)
}

// stage installs the embedded binaries in the home of n and writes the marker, without
// touching `current`. It logs the home first, also before a refusal, so a service that
// runs on another home than the operator expects shows it. decide checks the home
// before the first file is written. It returns the decision for the node data. bundle
// setup runs only stage, so the next start, or the node service, picks the binary for
// the node data.
func (b bundle) stage(n node) (decision, error) {
	logf("home %s (from %s)", n.Home, n.Source)
	d, err := b.decide(n.Layout)
	if err != nil {
		return decision{}, err
	}
	if err := removeMarker(n.Layout); err != nil {
		return decision{}, err
	}
	if err := b.install(n.Layout); err != nil {
		return decision{}, err
	}
	return d, writeMarker(n.Layout, b.digest)
}

// runInit runs "exrpd init" with the exrpd that initExrpd returns.
func runInit(b bundle, n node, args []string) error {
	bin, err := b.initExrpd(n)
	if err != nil {
		return err
	}
	return execBin(bin, append([]string{"exrpd", "--home", n.Home, "init"}, args...), os.Environ())
}

// initExrpd returns the exrpd that runs "exrpd init" for the home of n. On a new home,
// one without config/genesis.json where commands refuse because there is no marker or
// no exrpd, it installs the bundle first, like bundle setup. A home with
// config/genesis.json is an existing one: native init refuses it without --overwrite
// too. config/ alone does not count, client commands of the standalone exrpd create
// config/client.toml.
func (b bundle) initExrpd(n node) (string, error) {
	bin, err := b.exrpdFor(n)
	var refusal *homeRefusal
	switch {
	case !errors.As(err, &refusal), refusal.kind == refusedHome:
		return bin, err
	case refusal.kind == refusedOtherBundle, exists(filepath.Join(n.Home, "config", "genesis.json")):
		return "", fmt.Errorf("%s (from %s) already has config/genesis.json or a different bundle installed: restart the node service, or run `exrpd bundle setup --home %s` while the node is stopped", n.Home, n.Source, n.Home)
	}
	if _, err := b.stage(n); err != nil {
		return "", err
	}
	return b.exrpdFor(n)
}

// runExrpd runs any other command with the exrpd that runs the node data, without
// changing `current`: the node may be running. It never installs, and refuses unless
// the marker names this bundle. A help request runs here too, so it never installs and
// fails on a refusal like any other command. --home comes first, so exrpd uses the home
// the launcher resolved and no flag of the command can take it as its value.
func runExrpd(b bundle, n node, args []string) error {
	bin, err := b.exrpdFor(n)
	if err != nil {
		return err
	}
	return execBin(bin, append([]string{"exrpd", "--home", n.Home}, args...), os.Environ())
}

// runCompletion runs a shell completion request with the exrpd of the home, when the
// marker names this bundle. It never installs: a tab key must not write the bundle.
// On a refusal there is nothing to suggest. args hold the __complete word, after the
// flags of an alias, if any. --home follows that word, so exrpd completes from the same
// home, also when an alias put --home before __complete.
func runCompletion(b bundle, n node, args []string) error {
	bin, err := b.exrpdFor(n)
	var refusal *homeRefusal
	if errors.As(err, &refusal) {
		return noCompletion()
	}
	if err != nil {
		return err
	}
	at := slices.IndexFunc(args, isCompleteWord)
	return execBin(bin, slices.Concat([]string{"exrpd"}, args[:at+1], []string{"--home", n.Home}, args[at+1:]), os.Environ())
}

// noCompletion answers a completion request with no suggestions and cobra's
// ShellCompDirectiveNoFileComp (4), so the shell does not suggest file names either.
func noCompletion() error {
	fmt.Println(":4")
	return nil
}

// execBin replaces this process with bin, run as argv with env. A failure names bin.
func execBin(bin string, argv, env []string) error {
	if err := syscall.Exec(bin, argv, env); err != nil {
		return fmt.Errorf("run %s: %w", bin, err)
	}
	return nil
}

// exrpdFor returns the exrpd that runs the data of the home of n. It never installs. It
// returns that exrpd only when the marker names this bundle and the exrpd exists, and
// otherwise a *homeRefusal that names the home and its source.
func (b bundle) exrpdFor(n node) (string, error) {
	st, err := b.inspect(n)
	if err != nil {
		return "", err
	}
	if st.refusal != nil {
		return "", st.refusal
	}
	bin := n.bin(st.decision.release.dir())
	return bin, refuseSelf(bin)
}

// exists reports whether path exists. An error other than a missing file counts as
// existing, so a home that cannot be read is never taken for a new one.
func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
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
	if !sameFile(self, bin) {
		return nil
	}
	return fmt.Errorf("%s is this exrpd bundle, not a chain binary: run exrpd start once to install the chain binary there", bin)
}

// refuseSelectedSlot refuses a start invoked as a binary of current/bin under the working
// directory when that binary is this bundle. Stock Cosmovisor runs exactly that, from its
// root: started so, the bundle would install itself and move `current` under the
// Cosmovisor that runs it. The path is os.Args[0], as invoked, since os.Executable
// resolves the symlinks that tell the slot apart from the installed exrpd.
func refuseSelectedSlot() error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find this exrpd: %w", err)
	}
	invoked := os.Args[0]
	if !sameFile(filepath.Dir(invoked), filepath.Join("current", "bin")) || !sameFile(invoked, self) {
		return nil
	}
	return fmt.Errorf("%s runs from current/bin of the working directory, the binary that Cosmovisor selects: an external Cosmovisor seems to run this exrpd bundle, which runs its own. Stop the external Cosmovisor and run the installed exrpd start directly", invoked)
}

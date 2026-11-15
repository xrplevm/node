package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// mountedTwiceEnv carries "<home>:<dir>" to the run of the test in a user namespace:
// dir is the directory under each plan directory that it mounts, "." for the plan directory.
const mountedTwiceEnv = "EXRPD_TEST_MOUNTED_TWICE"

// Two plan directories that are one directory through a second mount, without a symlink,
// would get both binaries in turn and run the wrong one. Only root may mount, so the test
// runs itself again as root of a user namespace, and skips where the system forbids that.
func TestPrepareRefusesDirectoriesMountedTwice(t *testing.T) {
	if setup := os.Getenv(mountedTwiceEnv); setup != "" {
		home, dir, _ := strings.Cut(setup, ":")
		prepareMountedTwice(t, home, dir)
		return
	}

	for _, dir := range []string{".", "bin"} {
		t.Run(dir, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestPrepareRefusesDirectoriesMountedTwice$", "-test.v") //nolint:gosec // the test binary itself
			cmd.Env = append(os.Environ(), mountedTwiceEnv+"="+t.TempDir()+":"+dir)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Cloneflags:  syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
				UidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
				GidMappings: []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
			}

			out, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			switch {
			case errors.As(err, &exit):
				t.Fatalf("%s", out)
			case err != nil:
				t.Skipf("no user namespace: %v", err)
			case strings.Contains(string(out), "--- SKIP"):
				t.Skipf("%s", out)
			}
		})
	}
}

// prepareMountedTwice mounts dir of the upgrade plan directory on dir of the genesis plan
// directory in home, and checks that prepare refuses and writes no binary.
func prepareMountedTwice(t *testing.T, home, dir string) {
	b := testBundle(t)
	n := testNode(b, home)
	genesis := filepath.Join(n.Root, testGenesisDir, dir)
	upgrade := filepath.Join(n.Root, testUpgradeDir, dir)
	for _, d := range []string{genesis, upgrade} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := syscall.Mount(upgrade, genesis, "", syscall.MS_BIND, ""); err != nil {
		t.Skipf("no mount in the user namespace: %v", err)
	}

	err := b.prepare(n)
	if err == nil || !strings.Contains(err.Error(), "are the same directory") {
		t.Fatalf("prepare() error = %v, want a refusal", err)
	}
	for _, d := range []string{testGenesisDir, testUpgradeDir} {
		if _, err := os.Stat(n.bin(d)); err == nil {
			t.Errorf("a refused start wrote %s", n.bin(d))
		}
	}
}

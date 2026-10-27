package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
)

// The helpers below change files so that a crash leaves either the old or the new
// version on disk, never a partial one: write a temporary file next to the target,
// sync it, rename it to the target, then sync the directory so the change is durable.
// Missing parent directories are created and synced the same way.

// writeFileAtomic writes src to dest with mode, replacing dest. check, if not nil, runs
// after src is fully written and before dest is replaced. An error from it leaves dest
// untouched.
func writeFileAtomic(dest string, src io.Reader, mode os.FileMode, check func(size int64) error) error {
	dir := filepath.Dir(dest)
	if err := mkdirAllSync(dir); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".exrpd-upgrade-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once the rename moved it
	defer tmp.Close()

	size, err := io.Copy(tmp, src)
	if err != nil {
		return err
	}
	if check != nil {
		if err := check(size); err != nil {
			return err
		}
	}
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), dest); err != nil {
		return err
	}
	return syncDir(dir)
}

// symlinkAtomic points the symlink link at target, replacing any existing link.
func symlinkAtomic(target, link string) error {
	if current, err := os.Readlink(link); err == nil && current == target {
		return nil
	}
	dir := filepath.Dir(link)
	tmp, err := tempSymlink(target, dir)
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, link); err != nil {
		os.Remove(tmp)
		return err
	}
	return syncDir(dir)
}

// tempSymlink creates a symlink to target under a new random name in dir, so concurrent
// callers never share a temporary path.
func tempSymlink(target, dir string) (string, error) {
	for range 10000 {
		tmp := filepath.Join(dir, ".exrpd-upgrade-"+strconv.FormatUint(rand.Uint64(), 36)) //nolint:gosec // a unique name, not a secret
		if err := os.Symlink(target, tmp); !errors.Is(err, fs.ErrExist) {
			return tmp, err
		}
	}
	return "", fmt.Errorf("create a temporary symlink in %s: too many name collisions", dir)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// mkdirAllSync creates dir and any missing parents with mode 0o755, like os.MkdirAll,
// then syncs the parent of each new directory so the new entries survive a crash.
func mkdirAllSync(dir string) error {
	var missing []string // deepest first
	for d := filepath.Clean(dir); ; d = filepath.Dir(d) {
		info, err := os.Stat(d)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", d)
			}
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		missing = append(missing, d)
		if filepath.Dir(d) == d {
			break
		}
	}
	for i := len(missing) - 1; i >= 0; i-- {
		if err := os.Mkdir(missing[i], 0o755); err != nil {
			if !errors.Is(err, fs.ErrExist) {
				return err
			}
			// Another process created it first. That is fine if it is a directory.
			info, err := os.Stat(missing[i])
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", missing[i])
			}
		}
	}
	for _, d := range missing {
		if err := syncDir(filepath.Dir(d)); err != nil {
			return err
		}
	}
	return nil
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

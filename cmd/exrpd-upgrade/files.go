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

// writeFileAtomic writes src to dest with mode, replacing dest. check, if not nil, runs
// after src is fully written and before dest is replaced. An error from it leaves dest
// untouched.
func writeFileAtomic(dest string, src io.Reader, mode os.FileMode, check func(size int64) error) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
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

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

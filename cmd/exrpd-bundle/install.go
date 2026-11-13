package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// install writes the embedded binaries to their place in the layout: each release,
// genesis included, to the directory of its plan,
// and Cosmovisor. A binary already there with the embedded hash is kept. A different one
// is replaced and the replacement is logged: a bundle with a patch release replaces the
// binaries of the earlier bundle, and an older bundle started on the home puts its own
// back. install never touches `current`. It refuses, before it writes, when two binary
// directories of the bundle are the same directory, for example through a symlink made
// by hand: each binary would overwrite the other. The first install into a home is
// logged, so an install into a home the operator did not expect shows.
func (b bundle) install(l Layout) error {
	if err := distinctDirs(l, b.releases); err != nil {
		return err
	}

	_, err := os.Lstat(l.Root)
	first := errors.Is(err, fs.ErrNotExist)

	for _, r := range b.releases {
		if err := installBinary(b.files, r.Binary, l.bin(r.dir())); err != nil {
			return err
		}
	}

	if err := installBinary(b.files, b.m.Cosmovisor, l.Cosmovisor); err != nil {
		return err
	}

	if first {
		logf("installed the bundle in %s", l.Root)
	}

	return nil
}

// distinctDirs refuses when two releases have the same bin directory. It compares the
// paths after symlinks are followed, so a symlink to a directory that does not exist yet
// counts, and the identity of the directories that exist, so a second mount of one
// directory counts too.
func distinctDirs(l Layout, releases []Release) error {
	resolved := make([]string, len(releases))
	for i, r := range releases {
		dir := filepath.Dir(l.bin(r.dir()))
		var err error
		if resolved[i], err = resolveDangling(dir); err != nil {
			return err
		}
		for j, o := range releases[:i] {
			other := filepath.Dir(l.bin(o.dir()))
			if resolved[j] == resolved[i] || sameFile(other, dir) || sameFile(filepath.Dir(other), filepath.Dir(dir)) {
				return fmt.Errorf("%s and %s are the same directory. Each needs its own binary: give each a real directory of its own",
					other, dir)
			}
		}
	}
	return nil
}

func installBinary(files fs.FS, b Binary, dest string) error {
	existing, err := fileSHA256(dest)
	switch {
	case err == nil && existing == b.SHA256:
		return ensureExecutable(dest)
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return err
	}

	f, err := files.Open(b.File)
	if err != nil {
		return err
	}
	defer f.Close()

	unzipped, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("embedded %s: %w", b.File, err)
	}

	hash := sha256.New()
	err = writeFileAtomic(dest, io.TeeReader(unzipped, hash), 0o755, func(size int64) error {
		if got := hex.EncodeToString(hash.Sum(nil)); got != b.SHA256 || size != b.Size {
			return fmt.Errorf("embedded %s is corrupt: sha256 %s size %d, manifest sha256 %s size %d", b.File, got, size, b.SHA256, b.Size)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if existing != "" {
		logf("replaced %s (sha256 %s) with version %s (sha256 %s)", dest, existing, b.Version, b.SHA256)
	}

	return nil
}

// ensureExecutable gives a verified binary that its owner cannot execute mode 0755
// again, so a copy restored without permissions still runs.
func ensureExecutable(dest string) error {
	info, err := os.Stat(dest)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0o100 != 0 {
		return nil
	}
	if err := os.Chmod(dest, 0o755); err != nil {
		return fmt.Errorf("%s is not executable: %w", dest, err)
	}
	return nil
}

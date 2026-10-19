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
)

// install writes the three embedded binaries to their place in the layout. A binary
// already there with the embedded hash is kept. A different one is replaced and the
// replacement is logged: a bundle with a patch release replaces the binaries of the
// earlier bundle, and an older exrpd-upgrade started on the home puts its own back.
// install never touches `current`.
func install(files fs.FS, m Manifest, l Layout) error {
	for _, target := range []struct {
		bin  Binary
		dest string
	}{
		{m.Genesis, l.Genesis},
		{m.Upgrade, l.Upgrade},
		{m.Cosmovisor, l.Cosmovisor},
	} {
		if err := installBinary(files, target.bin, target.dest); err != nil {
			return err
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

package main

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

// An install that stops before the last binary leaves no marker, also when an earlier
// bundle had written one, so commands refuse until start or bundle setup completes. The
// next start writes it.
func TestInterruptedInstallLeavesNoMarker(t *testing.T) {
	files, m := testPayload(t)
	good := files[m.Upgrades[0].File]
	files[m.Upgrades[0].File] = &fstest.MapFile{Data: gzipped(t, "v12 bytez")}
	b := mustLoadBundle(t, files)
	n := testNode(b, t.TempDir())
	writeFile(t, n.Marker(), "earlier bundle\n", 0o644)

	if err := b.prepare(n); err == nil || !strings.Contains(err.Error(), "corrupt") {
		t.Fatalf("prepare() error = %v, want a corrupt payload error", err)
	}
	if got := readFile(t, n.Marker()); got != "" {
		t.Errorf("marker = %q after an interrupted install, want none", got)
	}
	var refusal *homeRefusal
	if _, err := b.exrpdFor(n); !errors.As(err, &refusal) || refusal.kind != refusedNoMarker {
		t.Errorf("exrpdFor() error = %v after an interrupted install, want a refusal", err)
	}

	files[m.Upgrades[0].File] = good
	if err := b.prepare(n); err != nil {
		t.Fatal(err)
	}
	if _, err := b.exrpdFor(n); err != nil {
		t.Errorf("exrpdFor() error = %v after start, want the marker of this bundle", err)
	}
}

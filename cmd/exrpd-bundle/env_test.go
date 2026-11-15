package main

import (
	"slices"
	"testing"
)

func TestCosmovisorEnv(t *testing.T) {
	env, err := cosmovisorEnv([]string{"PATH=/bin", "DAEMON_HOME=/other", "DAEMON_RESTART_AFTER_UPGRADE=1"}, "/h")
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(env)
	want := []string{
		"DAEMON_ALLOW_DOWNLOAD_BINARIES=false",
		"DAEMON_HOME=/h",
		"DAEMON_NAME=exrpd",
		"DAEMON_RESTART_AFTER_UPGRADE=true",
		"DAEMON_SHUTDOWN_GRACE=30s",
		"PATH=/bin",
		"UNSAFE_SKIP_BACKUP=true",
	}
	if !slices.Equal(env, want) {
		t.Fatalf("cosmovisorEnv() = %q, want %q", env, want)
	}
}

func TestCosmovisorEnvKeepsOperatorDefaults(t *testing.T) {
	env, err := cosmovisorEnv([]string{"DAEMON_SHUTDOWN_GRACE=5s", "UNSAFE_SKIP_BACKUP=false"}, "/h")
	if err != nil {
		t.Fatal(err)
	}
	for _, kv := range []string{"DAEMON_SHUTDOWN_GRACE=5s", "UNSAFE_SKIP_BACKUP=false"} {
		if !slices.Contains(env, kv) {
			t.Errorf("cosmovisorEnv() dropped %s: %q", kv, env)
		}
	}
	for _, kv := range []string{"DAEMON_SHUTDOWN_GRACE=30s", "UNSAFE_SKIP_BACKUP=true"} {
		if slices.Contains(env, kv) {
			t.Errorf("cosmovisorEnv() added %s over the operator value: %q", kv, env)
		}
	}
}

// Cosmovisor reads the first value of a repeated key, so an empty operator value must
// be replaced by the default rather than shadow it.
func TestCosmovisorEnvTreatsEmptyAsUnset(t *testing.T) {
	env, err := cosmovisorEnv([]string{"UNSAFE_SKIP_BACKUP="}, "/h")
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(env, "UNSAFE_SKIP_BACKUP=") || !slices.Contains(env, "UNSAFE_SKIP_BACKUP=true") {
		t.Fatalf("cosmovisorEnv() = %q, want only UNSAFE_SKIP_BACKUP=true", env)
	}
}

func TestCosmovisorEnvRejectsConflicts(t *testing.T) {
	for _, kv := range []string{
		"DAEMON_ALLOW_DOWNLOAD_BINARIES=true",
		"DAEMON_RESTART_AFTER_UPGRADE=false",
		"DAEMON_NAME=simd",
	} {
		if _, err := cosmovisorEnv([]string{kv}, "/h"); err == nil {
			t.Errorf("cosmovisorEnv() accepted %s", kv)
		}
	}
}

// Cosmovisor matches its config flag in any case and with one dash, and a config file
// overrides the environment that cosmovisorEnv pins.
func TestRefuseCosmovisorConfig(t *testing.T) {
	for _, args := range [][]string{
		{"--cosmovisor-config", "/etc/cv.toml"},
		{"--x", "--", "-COSMOVISOR-CONFIG", "/etc/cv.toml"},
	} {
		if err := refuseCosmovisorConfig(args); err == nil {
			t.Errorf("refuseCosmovisorConfig(%q) accepted a Cosmovisor config file", args)
		}
	}
	if err := refuseCosmovisorConfig([]string{"--chain-id", "c_1-1"}); err != nil {
		t.Errorf("refuseCosmovisorConfig() error = %v for exrpd start flags", err)
	}
}

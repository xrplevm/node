package main

import (
	"fmt"
	"strconv"
	"strings"
)

type envVar struct{ key, value string }

// requiredEnv are the Cosmovisor settings this bundle depends on. An operator value that
// means something else is an error, not silently replaced.
var requiredEnv = []envVar{
	{"DAEMON_NAME", "exrpd"},
	// Both binaries are embedded: nothing may be fetched at the upgrade height.
	{"DAEMON_ALLOW_DOWNLOAD_BINARIES", "false"},
	// Cosmovisor starts the upgraded exrpd itself after the switch.
	{"DAEMON_RESTART_AFTER_UPGRADE", "true"},
}

// defaultEnv are Cosmovisor settings with a different default here than upstream. An
// operator value wins.
var defaultEnv = []envVar{
	{"DAEMON_SHUTDOWN_GRACE", "30s"},
	// Upstream copies all of data/ at the upgrade height while the validator is down.
	{"UNSAFE_SKIP_BACKUP", "true"},
}

// cosmovisorEnv returns environ with DAEMON_HOME set to home, the required settings
// pinned and the defaults filled in where the operator left them unset or empty.
// Cosmovisor reads the first value of a repeated key, so every key appears once.
func cosmovisorEnv(environ []string, home string) ([]string, error) {
	pinned := append([]envVar{{"DAEMON_HOME", home}}, requiredEnv...)

	env := make([]string, 0, len(environ)+len(pinned)+len(defaultEnv))
	operatorSet := map[string]bool{}
	for _, kv := range environ {
		key, value, _ := strings.Cut(kv, "=")
		if want, ok := lookup(pinned, key); ok {
			if key != "DAEMON_HOME" && !sameSetting(value, want) {
				return nil, fmt.Errorf("%s=%s is not supported by exrpd-upgrade (it requires %s)", key, value, want)
			}
			continue
		}
		if _, ok := lookup(defaultEnv, key); ok && value == "" {
			continue
		}
		operatorSet[key] = true
		env = append(env, kv)
	}
	for _, v := range pinned {
		env = append(env, v.key+"="+v.value)
	}
	for _, v := range defaultEnv {
		if !operatorSet[v.key] {
			env = append(env, v.key+"="+v.value)
		}
	}
	return env, nil
}

// refuseCosmovisorConfig fails when args name a Cosmovisor config file. Cosmovisor takes
// "--cosmovisor-config FILE" out of the app arguments, in any letter case and also with
// one dash, and then reads its settings from that file. A setting the file omits gets
// its zero value, also when it is set in the environment, so the settings that
// cosmovisorEnv pins would not apply.
func refuseCosmovisorConfig(args []string) error {
	for _, arg := range args {
		if strings.EqualFold(arg, "--cosmovisor-config") || strings.EqualFold(arg, "-cosmovisor-config") {
			return fmt.Errorf("%s is not supported by exrpd-upgrade: it sets the Cosmovisor settings in the environment", arg)
		}
	}
	return nil
}

func lookup(vars []envVar, key string) (string, bool) {
	for _, v := range vars {
		if v.key == key {
			return v.value, true
		}
	}
	return "", false
}

// sameSetting reports whether an operator value is unset or has the meaning of want,
// for example "1" or "TRUE" for "true". The caller replaces it with want, so Cosmovisor,
// which accepts only "true" and "false", never reads the operator spelling.
func sameSetting(got, want string) bool {
	if got == "" || got == want {
		return true
	}
	g, errGot := strconv.ParseBool(got)
	w, errWant := strconv.ParseBool(want)
	return errGot == nil && errWant == nil && g == w
}

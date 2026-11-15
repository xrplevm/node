#!/usr/bin/env bash
# Smoke test of the start guard against nesting: stock Cosmovisor that runs the bundle as
# its selected exrpd gets a refusal before the bundle writes a file, and the installed
# bundle started directly from an old stock home still installs and moves `current`. Every
# command is start --help, so without the guard it ends after the help instead of running
# a node. It runs no chain and opens no port. Each case has its own home and $HOME.
# Usage: nesting-smoke.sh BUNDLE STOCK_COSMOVISOR
set -euo pipefail

usage="usage: $0 BUNDLE STOCK_COSMOVISOR"
bundle=$(realpath -e "${1:?$usage}") cv=$(realpath -e "${2:?$usage}")

# The guard diagnostic contains this phrase. The rest of its wording may change.
phrase='external Cosmovisor'

case=setup
tmp=$(mktemp -d "${TMPDIR:-/tmp}/nesting-smoke.XXXXXX")
log=$tmp/$case.log
# The bundle may install read-only files.
trap 'chmod -R u+w "$tmp" 2>/dev/null; rm -rf "$tmp"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

fail() {
	echo "nesting-smoke: $case: $*" >&2
	tail -n 40 "$log" >&2 2>/dev/null
	exit 1
}

# start_case NAME selects the directory and log of a case.
start_case() {
	case=$1 dir=$tmp/$1 log=$tmp/$1.log
	mkdir -p "$dir/user"
}

# listing prints every path under dir with its type, symlink target and file size.
listing() { find "$1" \( -type f -printf '%p f %s\n' \) -o -printf '%p %y %l\n' | LC_ALL=C sort; }

# stock_home HOME NAME [TARGET] makes HOME an old stock Cosmovisor home with `cosmovisor
# init`: genesis/bin/NAME is a copy of the bundle, or with TARGET a symlink to TARGET,
# and current -> genesis.
stock_home() {
	mkdir -p "$1/data"
	timeout -k 5 30 env -i PATH="$PATH" HOME="$dir/user" DAEMON_HOME="$1" DAEMON_NAME="$2" "$cv" init "$bundle" >>"$log" 2>&1 ||
		fail "stock Cosmovisor init"
	[[ -z ${3:-} ]] || ln -sfn "$3" "$1/cosmovisor/genesis/bin/$2"
}

# refused CREATED CMD... runs CMD, stock Cosmovisor running the bundle as its selected
# exrpd. The bundle must refuse with the guard diagnostic and leave every file of the case
# as it was: the node home and the default home under $HOME. CREATED is the listing line of
# a `current` that stock Cosmovisor creates itself before it starts the bundle, or "".
refused() {
	local created=$1 before after expected status=0
	shift
	before=$(listing "$dir")
	timeout -k 5 30 "$@" >>"$log" 2>&1 || status=$?
	((status != 0)) || fail "stock Cosmovisor ran the bundle start"
	((status != 124 && status != 137)) || fail "timed out"
	grep -iqF "$phrase" "$log" || fail "no refusal naming the $phrase"
	after=$(listing "$dir")
	expected=$(printf '%s\n' "$before" ${created:+"$created"} | LC_ALL=C sort)
	diff <(echo "$expected") <(echo "$after") >&2 || fail "the bundle changed files"
	echo "nesting-smoke: $case: refused"
}

info=$(env -i PATH="$PATH" HOME="$tmp" "$bundle" version --bundle --home "$tmp/no-home") || fail "version --bundle"
plan=$(jq -er '.manifest.genesis.plan' <<<"$info") || fail "no genesis plan in the manifest: $info"
version=$(jq -er '.manifest.genesis.version' <<<"$info") || fail "no genesis version in the manifest: $info"

# Configuration in the environment, the bundle copied as genesis/bin/exrpd.
start_case env-copy
stock_home "$dir/home" exrpd
refused '' env -i PATH="$PATH" HOME="$dir/user" DAEMON_HOME="$dir/home" DAEMON_NAME=exrpd \
	"$cv" run start --help

# Configuration only in a config file, on a home the bundle cannot find from its
# environment, with another binary name, a symlink to the bundle.
start_case config-link
stock_home "$dir/home" validator-exrpd "$bundle"
refused '' env -i PATH="$PATH" HOME="$dir/user" \
	"$cv" run --cosmovisor-config "$dir/home/cosmovisor/config.toml" start --help

# No current: stock Cosmovisor links it to genesis before it starts the bundle.
start_case missing-current
stock_home "$dir/home" exrpd "$bundle"
rm "$dir/home/cosmovisor/current"
refused "$dir/home/cosmovisor/current l genesis" env -i PATH="$PATH" HOME="$dir/user" DAEMON_HOME="$dir/home" DAEMON_NAME=exrpd \
	"$cv" run start --help

# The installed bundle started directly from the root of an old stock home whose genesis
# exrpd is a symlink to it: it installs its exrpd and moves current to the genesis plan.
start_case direct
mkdir -p "$dir/bin"
cp "$bundle" "$dir/bin/exrpd"
stock_home "$dir/home" exrpd "$dir/bin/exrpd"
(cd "$dir/home/cosmovisor" && timeout -k 5 60 env -i PATH="$PATH" HOME="$dir/user" "$dir/bin/exrpd" start --home "$dir/home" --help) >>"$log" 2>&1 ||
	fail "direct start --help failed"
! grep -iqF "$phrase" "$log" || fail "direct start was refused as nested"
grep -qF 'Usage:' "$log" || fail "exrpd start printed no help"
test "$(readlink "$dir/home/cosmovisor/current")" = "upgrades/$plan" || fail "current is not upgrades/$plan"
native=$dir/home/cosmovisor/upgrades/$plan/bin/exrpd
[[ -f $native && ! $native -ef $dir/bin/exrpd ]] || fail "$native is not the native exrpd"
test "$(env -i PATH="$PATH" HOME="$dir/user" "$native" version --home "$tmp/no-home" 2>>"$log")" = "$version" ||
	fail "$native is not exrpd $version"
[[ ! -e $dir/user/.exrpd ]] || fail "direct start installed in the default home"
echo "nesting-smoke: $case: installed exrpd $version as upgrades/$plan"

echo "nesting-smoke: passed"

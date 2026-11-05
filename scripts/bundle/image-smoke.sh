#!/usr/bin/env bash
# Smoke test of a bundle image before the push: exrpd init on a new home runs the genesis
# exrpd and installs every exrpd and Cosmovisor, then each of them must run. The image runs
# commands through its entrypoint, /bin/sh -ec.
# Usage: image-smoke.sh IMAGE GENESIS_VERSION VERSION   (GENESIS_VERSION is VERSION for a single bundle)
set -euo pipefail

image=${1:?usage: $0 IMAGE GENESIS_VERSION VERSION} genesis=${2:?} version=${3:?}

fail() { echo "image-smoke: $image: $*" >&2; exit 1; }

info=$(docker run --rm "$image" 'exrpd version --bundle --home /tmp/no-home')
jq -e --arg g "$genesis" --arg v "$version" '.manifest.platform == "linux/amd64" and .manifest.genesis.version == $g
	and if $g == $v then .manifest | has("upgrades") | not else (.manifest.upgrades | length) == 1 and .manifest.upgrades[0].version == $v end' <<<"$info" >/dev/null ||
	fail "manifest is not genesis $genesis, version $version for linux/amd64: $info"

want=$(jq -r '[.manifest.genesis] + (.manifest.upgrades // []) | .[] | "exrpd \(.plan) \(.version)"' <<<"$info" | sort)

# shellcheck disable=SC2016 # the script expands in the container's shell
got=$(docker run --rm "$image" '
h=/tmp/smoke
exrpd init smoke --chain-id xrplevm_1449999-1 --home "$h" >/tmp/init.log 2>&1 || { cat /tmp/init.log >&2; exit 1; }
for b in "$h"/cosmovisor/upgrades/*/bin/exrpd; do
	dir=${b%/bin/exrpd}
	version=$("$b" version --home /tmp/no-home) || exit 1
	echo "exrpd ${dir##*/} $version"
done
for c in "$h"/cosmovisor/supervisor/*/cosmovisor; do
	version=$("$c" version --cosmovisor-only 2>&1) || exit 1
	echo "cosmovisor $version"
done')

echo "$got"
test "$(grep '^exrpd ' <<<"$got" | sort)" = "$want" || fail "installed exrpd versions are not those of the manifest: $want"
test "$(grep -c '^cosmovisor cosmovisor version: ' <<<"$got")" = 1 || fail "want one Cosmovisor that runs"

echo "image-smoke: $image passed"

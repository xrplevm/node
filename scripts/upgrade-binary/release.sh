#!/usr/bin/env bash
# Build exrpd-upgrade_<version>_<Os>_<arch> for each platform from published release archives,
# plus exrpd-upgrade_<version>_LICENSE, because the executables embed Apache-2.0 code.
#
# Usage: release.sh --version X.Y.Z --standalone-dir DIR --genesis-dir DIR --out-dir DIR
#   --standalone-dir  node_<version>_<Os>_<arch>.tar.gz + checksums.txt of this release
#   --genesis-dir     node_<GENESIS_VERSION>_<Os>_<arch>.tar.gz + checksums.txt of the genesis release
# PLATFORMS (env) defaults to "linux/amd64 linux/arm64 darwin/amd64 darwin/arm64".
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/bundle.env
source "$here/bundle.env"

version='' standalone='' genesis='' out=''
while [[ $# -gt 0 ]]; do
	case $1 in
	--version) version=$2 ;;
	--standalone-dir) standalone=$2 ;;
	--genesis-dir) genesis=$2 ;;
	--out-dir) out=$2 ;;
	*) echo "unknown argument $1" >&2; exit 2 ;;
	esac
	shift 2
done
for v in version standalone genesis out; do
	[[ -n ${!v} ]] || { echo "missing --$v" >&2; exit 2; }
done
plan_version=${UPGRADE_PLAN#v} base_version=${version%%-*}
if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ || ${base_version%.*} != "${plan_version%.*}" ]]; then
	echo "release $version does not have the major.minor of the bundle plan $UPGRADE_PLAN: bump scripts/upgrade-binary/bundle.env (see docs/upgrade-binary-dev.md, Releases and bundles)" >&2
	exit 1
fi

sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1; }

mkdir -p "$out"
out=$(cd "$out" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
sums=$out/exrpd-upgrade_${version}_checksums.txt
: >"$sums"

for platform in ${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64}; do
	goos=${platform%/*} goarch=${platform#*/}
	os_title=Linux
	[[ $goos == darwin ]] && os_title=Darwin

	archive=node_${version}_${os_title}_${goarch}.tar.gz
	want=$(awk -v f="$archive" '$2 == f { print $1 }' "$standalone/checksums.txt")
	got=$(sha256 "$standalone/$archive")
	[[ -n $want && $got == "$want" ]] || { echo "$archive sha256 $got does not match checksums.txt ($want)" >&2; exit 1; }
	rm -rf "$work/standalone" && mkdir -p "$work/standalone"
	tar -xzf "$standalone/$archive" -C "$work/standalone" bin/exrpd

	name=exrpd-upgrade_${version}_${os_title}_${goarch}
	"$here/build.sh" \
		--genesis-archive "$genesis/node_${GENESIS_VERSION}_${os_title}_${goarch}.tar.gz" \
		--genesis-checksums "$genesis/checksums.txt" \
		--upgrade-exrpd "$work/standalone/bin/exrpd" --upgrade-version "$version" \
		--goos "$goos" --goarch "$goarch" --out "$out/$name"
	echo "$(sha256 "$out/$name")  $name" >>"$sums"
done
cp "$here/../../LICENSE" "$out/exrpd-upgrade_${version}_LICENSE"
cat "$sums"

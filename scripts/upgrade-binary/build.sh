#!/usr/bin/env bash
# Build one exrpd-upgrade binary for one platform.
#
# Usage: build.sh --genesis-archive node_<GENESIS_VERSION>_<Os>_<arch>.tar.gz --genesis-checksums checksums.txt \
#                 --upgrade-exrpd PATH --upgrade-version X.Y.Z --goos GOOS --goarch GOARCH --out PATH \
#                 [--cosmovisor PATH --cosmovisor-source TEXT]
#
# The genesis exrpd is taken byte for byte from its release archive after checking the
# archive against the release checksums.txt. Without --cosmovisor, Cosmovisor is built
# from the pinned module source plus the committed patch (cosmovisor.sh).
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)
# shellcheck source=SCRIPTDIR/bundle.env
source "$here/bundle.env"

genesis_archive='' genesis_checksums='' upgrade_exrpd='' upgrade_version='' goos='' goarch='' out='' cosmovisor='' cosmovisor_source=''
while [[ $# -gt 0 ]]; do
	case $1 in
	--genesis-archive) genesis_archive=$2 ;;
	--genesis-checksums) genesis_checksums=$2 ;;
	--upgrade-exrpd) upgrade_exrpd=$2 ;;
	--upgrade-version) upgrade_version=$2 ;;
	--goos) goos=$2 ;;
	--goarch) goarch=$2 ;;
	--out) out=$2 ;;
	--cosmovisor) cosmovisor=$2 ;;
	--cosmovisor-source) cosmovisor_source=$2 ;;
	*) echo "unknown argument $1" >&2; exit 2 ;;
	esac
	shift 2
done
for v in genesis_archive genesis_checksums upgrade_exrpd upgrade_version goos goarch out; do
	[[ -n ${!v} ]] || { echo "missing --${v//_/-}" >&2; exit 2; }
done
case $goos/$goarch in
linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64) ;;
*) echo "unsupported platform $goos/$goarch" >&2; exit 2 ;;
esac
if [[ -n $cosmovisor && -z $cosmovisor_source ]]; then
	echo "--cosmovisor requires --cosmovisor-source" >&2
	exit 2
fi

# The staged upgrade binary lives in upgrades/$UPGRADE_PLAN. Major and minor releases need
# their own on-chain plan, so only releases with the plan's major.minor fit there: the
# plan's release, its patch releases and their pre-releases.
plan_version=${UPGRADE_PLAN#v} base_version=${upgrade_version%%-*}
if [[ ! $upgrade_version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ || ${base_version%.*} != "${plan_version%.*}" ]]; then
	echo "--upgrade-version $upgrade_version does not have the major.minor of plan $UPGRADE_PLAN in scripts/upgrade-binary/bundle.env: bump bundle.env for a new plan" >&2
	exit 1
fi

sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1; }
size() { wc -c <"$1" | tr -d ' '; }

os_title=Linux
[[ $goos == darwin ]] && os_title=Darwin
archive_name="node_${GENESIS_VERSION}_${os_title}_${goarch}.tar.gz"
[[ $(basename "$genesis_archive") == "$archive_name" ]] || { echo "expected genesis archive $archive_name" >&2; exit 1; }
want=$(awk -v f="$archive_name" '$2 == f { print $1 }' "$genesis_checksums")
got=$(sha256 "$genesis_archive")
[[ -n $want && $got == "$want" ]] || { echo "$archive_name sha256 $got does not match checksums.txt ($want)" >&2; exit 1; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cp -R "$repo/cmd/exrpd-upgrade" "$work/mod"
payload=$work/mod/payload raw=$work/raw
rm -rf "$payload" && mkdir -p "$payload" "$raw"

tar -xzf "$genesis_archive" -C "$work" bin/exrpd
mv "$work/bin/exrpd" "$raw/exrpd-genesis"
cp "$upgrade_exrpd" "$raw/exrpd-upgrade"
if [[ -n $cosmovisor ]]; then
	cp "$cosmovisor" "$raw/cosmovisor"
else
	"$here/cosmovisor.sh" "$goos" "$goarch" "$raw/cosmovisor"
	cosmovisor_source="$COSMOVISOR_MODULE@$COSMOVISOR_VERSION $COSMOVISOR_SUM + $COSMOVISOR_PATCH sha256:$(sha256 "$here/$COSMOVISOR_PATCH")"
fi

# Payload files are gzip-compressed. sha256 and size are those of the executable itself.
for f in exrpd-genesis exrpd-upgrade cosmovisor; do
	gzip -n -9 -c "$raw/$f" >"$payload/$f.gz"
done
entry() { # version name source
	jq -cn --arg version "$1" --arg file "$2.gz" --arg sha256 "$(sha256 "$raw/$2")" \
		--argjson size "$(size "$raw/$2")" --arg source "$3" \
		'{version: $version, file: $file, sha256: $sha256, size: $size, source: $source}'
}
{
	printf '{"platform":"%s","plan":"%s",' "$goos/$goarch" "$UPGRADE_PLAN"
	printf '"genesis":%s,' "$(entry "$GENESIS_VERSION" exrpd-genesis "github.com/xrplevm/node release v$GENESIS_VERSION $archive_name sha256:$got")"
	printf '"upgrade":%s,' "$(entry "$upgrade_version" exrpd-upgrade "exrpd $upgrade_version standalone release binary")"
	printf '"cosmovisor":%s}\n' "$(entry "$COSMOVISOR_VERSION" cosmovisor "$cosmovisor_source")"
} >"$payload/manifest.json"
jq -e . "$payload/manifest.json" >/dev/null

mkdir -p "$(dirname "$out")"
out=$(cd "$(dirname "$out")" && pwd)/$(basename "$out")
(cd "$work/mod" && GOWORK=off GOFLAGS=-mod=readonly CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
	go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$out" .)
echo "built $out ($(size "$out") bytes) for $goos/$goarch"

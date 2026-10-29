#!/usr/bin/env bash
# Build one bundle executable, exrpd, for one platform.
#
# Usage: build.sh --genesis-version X.Y.Z --genesis-archive node_<X.Y.Z>_<Os>_<arch>.tar.gz --genesis-checksums checksums.txt \
#                 --upgrade-exrpd PATH --upgrade-version X.Y.Z --goos GOOS --goarch GOARCH --out PATH \
#                 [--genesis-plan vX.Y.Z] [--cosmovisor PATH --cosmovisor-source TEXT]
#
# The upgrade plan is the upgrade handler of this checkout (plan.sh). The genesis plan
# defaults to the newest plan in the genesis release tag (genesis-plan.sh). The genesis
# exrpd is taken byte for byte from its release archive after checking the archive
# against the release checksums.txt. Without --cosmovisor, Cosmovisor is built from the
# pinned module source plus the committed patch (cosmovisor.sh).
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)
# shellcheck source=SCRIPTDIR/cosmovisor/cosmovisor.env
source "$here/cosmovisor/cosmovisor.env"

genesis_version='' genesis_plan='' genesis_archive='' genesis_checksums='' upgrade_exrpd='' upgrade_version='' goos='' goarch='' out='' cosmovisor='' cosmovisor_source=''
while [[ $# -gt 0 ]]; do
	case $1 in
	--genesis-version) genesis_version=$2 ;;
	--genesis-archive) genesis_archive=$2 ;;
	--genesis-plan) genesis_plan=$2 ;;
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
for v in genesis_version genesis_archive genesis_checksums upgrade_exrpd upgrade_version goos goarch out; do
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

# The staged upgrade binary lives in upgrades/<plan>. Major and minor releases need their
# own on-chain plan, so only releases with the plan's major.minor fit there: the plan's
# release, its patch releases and their pre-releases.
plan=$("$here/plan.sh")
plan_version=${plan#v} base_version=${upgrade_version%%-*}
if [[ ! $upgrade_version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ || ${base_version%.*} != "${plan_version%.*}" ]]; then
	echo "--upgrade-version $upgrade_version does not have the major.minor of plan $plan in app/upgrades: a new major or minor release needs its upgrade handler" >&2
	exit 1
fi
[[ -n $genesis_plan ]] || genesis_plan=$("$here/genesis-plan.sh" "$genesis_version")
# The launcher accepts only plans named vX.Y.Z without leading zeros: any other name would
# make every command of the bundle fail.
for p in "$plan" "$genesis_plan"; do
	if [[ ! $p =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
		echo "plan $p is not named vX.Y.Z, so this bundle cannot carry it" >&2
		exit 1
	fi
done
# The launcher requires plans in order: the genesis runs the data before the upgrade plan.
if [[ $genesis_plan == "$plan" || $(printf '%s\n%s\n' "$genesis_plan" "$plan" | sort -V | head -n 1) != "$genesis_plan" ]]; then
	echo "genesis $genesis_version runs plan $genesis_plan, which does not come before plan $plan of this bundle: choose a genesis release from before plan $plan" >&2
	exit 1
fi

sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1; }
size() { wc -c <"$1" | tr -d ' '; }

os_title=Linux
[[ $goos == darwin ]] && os_title=Darwin
archive_name="node_${genesis_version}_${os_title}_${goarch}.tar.gz"
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
	cosmovisor_source="$COSMOVISOR_MODULE@$COSMOVISOR_VERSION $COSMOVISOR_SUM + $COSMOVISOR_PATCH sha256:$(sha256 "$here/cosmovisor/$COSMOVISOR_PATCH")"
fi

# Payload files are gzip-compressed. sha256 and size are those of the executable itself.
for f in exrpd-genesis exrpd-upgrade cosmovisor; do
	gzip -n -9 -c "$raw/$f" >"$payload/$f.gz"
done
entry() { # plan version name source
	jq -cn --arg plan "$1" --arg version "$2" --arg file "$3.gz" --arg sha256 "$(sha256 "$raw/$3")" \
		--argjson size "$(size "$raw/$3")" --arg source "$4" \
		'{version: $version, file: $file, sha256: $sha256, size: $size, source: $source} + if $plan == "" then {} else {plan: $plan} end'
}
{
	printf '{"platform":"%s",' "$goos/$goarch"
	printf '"genesis":%s,' "$(entry "$genesis_plan" "$genesis_version" exrpd-genesis "github.com/xrplevm/node release v$genesis_version $archive_name sha256:$got")"
	printf '"upgrades":[%s],' "$(entry "$plan" "$upgrade_version" exrpd-upgrade "exrpd $upgrade_version standalone release binary")"
	printf '"cosmovisor":%s}\n' "$(entry "" "$COSMOVISOR_VERSION" cosmovisor "$cosmovisor_source")"
} >"$payload/manifest.json"
jq -e . "$payload/manifest.json" >/dev/null

mkdir -p "$(dirname "$out")"
out=$(cd "$(dirname "$out")" && pwd)/$(basename "$out")
(cd "$work/mod" && GOWORK=off GOFLAGS=-mod=readonly CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
	go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$out" .)
echo "built $out ($(size "$out") bytes) for $goos/$goarch"

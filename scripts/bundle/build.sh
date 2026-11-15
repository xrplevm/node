#!/usr/bin/env bash
# Build one bundle executable for one platform.
#
# Usage: build.sh --genesis-version X.Y.Z [--genesis-plan vX.Y.Z] --genesis-archive node_X.Y.Z_<Os>_<arch>.tar.gz \
#                 --genesis-checksums checksums.txt --goos GOOS --goarch GOARCH --out PATH \
#                 [--upgrade-exrpd PATH --upgrade-version X.Y.Z --plan vX.Y.Z] [--cosmovisor PATH --cosmovisor-source TEXT]
#
# With --upgrade-exrpd it builds a two-binary bundle. Without it, a single bundle.
# The genesis plan defaults to vX.Y.0. Without --cosmovisor, Cosmovisor is built from the pinned source plus the patch.
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)

# shellcheck source=SCRIPTDIR/lib.sh
source "$here/lib.sh"
# shellcheck source=SCRIPTDIR/cosmovisor/cosmovisor.env
source "$here/cosmovisor/cosmovisor.env"

genesis_version='' genesis_plan='' genesis_archive='' genesis_checksums='' plan='' upgrade_exrpd='' upgrade_version='' goos='' goarch='' out='' cosmovisor='' cosmovisor_source=''

while [[ $# -gt 0 ]]; do
	case $1 in
	--genesis-version) genesis_version=$2 ;;
	--genesis-archive) genesis_archive=$2 ;;
	--genesis-plan) genesis_plan=$2 ;;
	--genesis-checksums) genesis_checksums=$2 ;;
	--plan) plan=$2 ;;
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

for v in genesis_version genesis_archive genesis_checksums goos goarch out; do
	[[ -n ${!v} ]] || { echo "missing --${v//_/-}" >&2; exit 2; }
done

if [[ -n $upgrade_exrpd$upgrade_version ]]; then
	for v in upgrade_exrpd upgrade_version plan; do
		[[ -n ${!v} ]] || { echo "missing --${v//_/-}: a two-binary bundle needs --upgrade-exrpd, --upgrade-version and --plan" >&2; exit 2; }
	done
elif [[ -n $plan ]]; then
	echo "--plan is the plan of the upgrade: it needs --upgrade-exrpd and --upgrade-version" >&2
	exit 2
fi

check_platform "$goos" "$goarch"

if [[ -n $cosmovisor && -z $cosmovisor_source ]]; then
	echo "--cosmovisor requires --cosmovisor-source" >&2
	exit 2
fi

if [[ -n $upgrade_exrpd ]]; then
	plans=$(resolve_plans --upgrade-version "$upgrade_version" "$plan" "$genesis_version" "$genesis_plan")
	read -r plan genesis_plan <<<"$plans"
else
	genesis_plan=$(genesis_plan_for "$genesis_version" "$genesis_plan")
fi

archive_name=node_${genesis_version}_$(os_title "$goos")_${goarch}.tar.gz
[[ $(basename "$genesis_archive") == "$archive_name" ]] || { echo "expected genesis archive $archive_name" >&2; exit 1; }
genesis_sha=$(verify_archive "$genesis_checksums" "$genesis_archive")

# The launcher commit, with -dirty when cmd/exrpd-bundle has uncommitted changes.
commit=$(git -C "$repo" rev-parse HEAD)
changes=$(git -C "$repo" status --porcelain -- cmd/exrpd-bundle)
commit=$commit${changes:+-dirty}

# The launcher is built from a copy, so the payload never lands in the checkout.
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cp -R "$repo/cmd/exrpd-bundle" "$work/mod"
payload=$work/mod/payload raw=$work/raw
rm -rf "$payload" && mkdir -p "$payload" "$raw"

tar -xzf "$genesis_archive" -C "$work" bin/exrpd
mv "$work/bin/exrpd" "$raw/exrpd-genesis"

files=(exrpd-genesis cosmovisor)
if [[ -n $upgrade_exrpd ]]; then
	cp "$upgrade_exrpd" "$raw/exrpd-upgrade"
	files+=(exrpd-upgrade)
fi

if [[ -n $cosmovisor ]]; then
	cp "$cosmovisor" "$raw/cosmovisor"
else
	"$here/cosmovisor.sh" "$goos" "$goarch" "$raw/cosmovisor"
	cosmovisor_source="$COSMOVISOR_MODULE@$COSMOVISOR_VERSION $COSMOVISOR_SUM + $COSMOVISOR_PATCH sha256:$(sha256 "$here/cosmovisor/$COSMOVISOR_PATCH")"
fi

# Payload files are gzip-compressed. The manifest has the sha256 and size of each executable itself.
size() { wc -c <"$1" | tr -d ' '; }

for f in "${files[@]}"; do
	gzip -n -9 -c "$raw/$f" >"$payload/$f.gz"
done

entry() { # PLAN VERSION FILE SOURCE
	jq -cn --arg plan "$1" --arg version "$2" --arg file "$3.gz" --arg sha256 "$(sha256 "$raw/$3")" \
		--argjson size "$(size "$raw/$3")" --arg source "$4" \
		'{version: $version, file: $file, sha256: $sha256, size: $size, source: $source} + if $plan == "" then {} else {plan: $plan} end'
}

{
	printf '{"platform":"%s",' "$goos/$goarch"
	printf '"launcher":%s,' "$(jq -cn --arg commit "$commit" '{commit: $commit}')"
	printf '"genesis":%s,' "$(entry "$genesis_plan" "$genesis_version" exrpd-genesis "github.com/xrplevm/node release v$genesis_version $archive_name sha256:$genesis_sha")"
	if [[ -n $upgrade_exrpd ]]; then
		printf '"upgrades":[%s],' "$(entry "$plan" "$upgrade_version" exrpd-upgrade "exrpd $upgrade_version standalone release binary")"
	fi
	printf '"cosmovisor":%s}\n' "$(entry "" "$COSMOVISOR_VERSION" cosmovisor "$cosmovisor_source")"
} >"$payload/manifest.json"

jq -e . "$payload/manifest.json" >/dev/null

mkdir -p "$(dirname "$out")"
out=$(cd "$(dirname "$out")" && pwd)/$(basename "$out")
(cd "$work/mod" && GOWORK=off GOFLAGS=-mod=readonly CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
	go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$out" .)

echo "built $out ($(size "$out") bytes) for $goos/$goarch"

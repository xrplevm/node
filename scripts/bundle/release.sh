#!/usr/bin/env bash
# Build one bundle set from published release archives: exrpd-bundle_<set>_<Os>_<arch>.tar.gz
# for each platform (the bundle exrpd and LICENSE) and exrpd-bundle_<set>_checksums.txt.
#
# Usage: release.sh --version X.Y.Z [--plan vX.Y.Z] --standalone-dir DIR --out-dir DIR \
#                   [--genesis-version X.Y.Z [--genesis-plan vX.Y.Z] --genesis-dir DIR]
#   Without --genesis-version: the single set <version>, release --version alone.
#   With it: the two-binary set <genesis-version>-<version>, the genesis then --version staged for --plan.
#   --standalone-dir and --genesis-dir hold node_<version>_<Os>_<arch>.tar.gz and checksums.txt of each release.
#   --out-dir must be absent or empty, so that no old asset stays without a checksum line.
#   Plans default to the plan of each release tag. PLATFORMS (env) defaults to all four.
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/lib.sh
source "$here/lib.sh"

version='' plan='' genesis_version='' genesis_plan='' standalone_dir='' genesis_dir='' out_dir=''
while [[ $# -gt 0 ]]; do
	case $1 in
	--version) version=$2 ;;
	--plan) plan=$2 ;;
	--genesis-version) genesis_version=$2 ;;
	--genesis-plan) genesis_plan=$2 ;;
	--standalone-dir) standalone_dir=$2 ;;
	--genesis-dir) genesis_dir=$2 ;;
	--out-dir) out_dir=$2 ;;
	*) echo "unknown argument $1" >&2; exit 2 ;;
	esac
	shift 2
done
required=(version standalone_dir out_dir)
if [[ -n $genesis_version ]]; then
	required+=(genesis_dir)
elif [[ -n $genesis_plan$genesis_dir ]]; then
	echo "--genesis-plan and --genesis-dir need --genesis-version, the two-binary set" >&2
	exit 2
fi
for v in "${required[@]}"; do
	[[ -n ${!v} ]] || { echo "missing --${v//_/-}" >&2; exit 2; }
done
if [[ -e $out_dir && ( ! -d $out_dir || -n $(ls -A "$out_dir") ) ]]; then
	echo "--out-dir $out_dir is not an empty directory: use a new or empty directory, so that no old asset stays without a checksum entry" >&2
	exit 2
fi
# Resolve the plans once, before any build. The single set's genesis is the release itself.
if [[ -n $genesis_version ]]; then
	plans=$(resolve_plans release "$version" "$plan" "$genesis_version" "$genesis_plan")
	read -r plan genesis_plan <<<"$plans"
	set_name=${genesis_version}-${version}
else
	genesis_plan=$(release_plan "$version" "$plan")
	genesis_version=$version genesis_dir=$standalone_dir plan=''
	set_name=$version
fi

mkdir -p "$out_dir"
out_dir=$(cd "$out_dir" && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
sums=$out_dir/exrpd-bundle_${set_name}_checksums.txt
: >"$sums"

for platform in ${PLATFORMS:-linux/amd64 linux/arm64 darwin/amd64 darwin/arm64}; do
	goos=${platform%/*} goarch=${platform#*/}
	check_platform "$goos" "$goarch"
	os=$(os_title "$goos")
	args=(--genesis-version "$genesis_version" --genesis-plan "$genesis_plan" --goos "$goos" --goarch "$goarch"
		--genesis-archive "$genesis_dir/node_${genesis_version}_${os}_${goarch}.tar.gz" --genesis-checksums "$genesis_dir/checksums.txt")
	if [[ -n $plan ]]; then
		archive=node_${version}_${os}_${goarch}.tar.gz
		verify_archive "$standalone_dir/checksums.txt" "$standalone_dir/$archive" >/dev/null
		rm -rf "$work/standalone" && mkdir -p "$work/standalone"
		tar -xzf "$standalone_dir/$archive" -C "$work/standalone" bin/exrpd
		args+=(--upgrade-exrpd "$work/standalone/bin/exrpd" --upgrade-version "$version" --plan "$plan")
	fi

	name=exrpd-bundle_${set_name}_${os}_${goarch}
	rm -rf "${work:?}/$name" && mkdir -p "$work/$name"
	"$here/build.sh" "${args[@]}" --out "$work/$name/exrpd"
	cp "$here/../../LICENSE" "$work/$name/LICENSE"
	tar -C "$work/$name" --sort=name --owner=0 --group=0 --numeric-owner --mtime=@0 -cf - exrpd LICENSE |
		gzip -n -9 >"$out_dir/$name.tar.gz"
	echo "$(sha256 "$out_dir/$name.tar.gz")  $name.tar.gz" >>"$sums"
done
cat "$sums"

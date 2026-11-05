#!/usr/bin/env bash
# Build one bundle set from published release archives: an exrpd-bundle archive for each
# platform and its checksums file. Each archive holds the bundle executable exrpd and the
# LICENSE, because the executable embeds Apache-2.0 code.
#
# Usage: release.sh --version X.Y.Z [--plan vX.Y.Z] --standalone-dir DIR --out-dir DIR \
#                   [--genesis-version X.Y.Z [--genesis-plan vX.Y.Z] --genesis-dir DIR]
#   --out-dir          must be absent or empty, so that no old asset stays without a checksum entry
#   --plan             the plan of --version, default the newest plan of tag v<version> (release-plan.sh)
#   --standalone-dir   node_<version>_<Os>_<arch>.tar.gz + checksums.txt of this release
#   --genesis-version  build the two-binary set: the genesis release, then --version staged for --plan.
#                      Without it, the single set: release --version alone, which runs --plan
#   --genesis-plan     the plan of --genesis-version, default the newest plan of its tag
#   --genesis-dir      node_<genesis-version>_<Os>_<arch>.tar.gz + checksums.txt of the genesis release
# The single set is exrpd-bundle_<version>_<Os>_<arch>.tar.gz and exrpd-bundle_<version>_checksums.txt,
# the two-binary set exrpd-bundle_<genesis-version>-<version>_<Os>_<arch>.tar.gz and
# exrpd-bundle_<genesis-version>-<version>_checksums.txt.
# PLATFORMS (env) defaults to "linux/amd64 linux/arm64 darwin/amd64 darwin/arm64".
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
# The plans are the same for every platform, so they are checked once. The single set's
# genesis is the release itself.
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

	upgrade_exrpd=''
	if [[ -n $plan ]]; then
		archive=node_${version}_${os}_${goarch}.tar.gz
		verify_archive "$standalone_dir/checksums.txt" "$standalone_dir/$archive" >/dev/null
		rm -rf "$work/standalone" && mkdir -p "$work/standalone"
		tar -xzf "$standalone_dir/$archive" -C "$work/standalone" bin/exrpd
		upgrade_exrpd=$work/standalone/bin/exrpd
	fi

	name=exrpd-bundle_${set_name}_${os}_${goarch}
	rm -rf "${work:?}/$name" && mkdir -p "$work/$name"
	build_one "$work/$name/exrpd" "$goos" "$goarch" "$genesis_version" "$genesis_plan" \
		"$genesis_dir/node_${genesis_version}_${os}_${goarch}.tar.gz" "$genesis_dir/checksums.txt" \
		"$plan" "$upgrade_exrpd" "${upgrade_exrpd:+$version}"
	cp "$here/../../LICENSE" "$work/$name/LICENSE"
	tar -C "$work/$name" --sort=name --owner=0 --group=0 --numeric-owner --mtime=@0 -cf - exrpd LICENSE |
		gzip -n -9 >"$out_dir/$name.tar.gz"
	echo "$(sha256 "$out_dir/$name.tar.gz")  $name.tar.gz" >>"$sums"
done
cat "$sums"

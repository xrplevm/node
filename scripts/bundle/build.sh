#!/usr/bin/env bash
# Build one bundle executable, exrpd, for one platform.
#
# Usage: build.sh --genesis-version X.Y.Z --genesis-archive node_<X.Y.Z>_<Os>_<arch>.tar.gz --genesis-checksums checksums.txt \
#                 [--upgrade-exrpd PATH --upgrade-version X.Y.Z [--plan vX.Y.Z]] --goos GOOS --goarch GOARCH --out PATH \
#                 [--genesis-plan vX.Y.Z] [--cosmovisor PATH --cosmovisor-source TEXT]
#
# With --upgrade-exrpd and --upgrade-version it builds a two-binary bundle: the upgrade
# exrpd is staged for --plan, by default the upgrade handler of this checkout (plan.sh).
# Without both, it builds a single bundle: the genesis release and Cosmovisor, no
# upgrade. The genesis plan defaults to the newest plan in the genesis release tag
# (release-plan.sh). The genesis exrpd is taken byte for byte from its release archive
# after checking the archive against the release checksums.txt. Without --cosmovisor,
# Cosmovisor is built from the pinned module source plus the committed patch
# (cosmovisor.sh). The steps are release_plan, resolve_plans and build_one of lib.sh,
# which release.sh runs too.
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/lib.sh
source "$here/lib.sh"

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
	for v in upgrade_exrpd upgrade_version; do
		[[ -n ${!v} ]] || { echo "missing --${v//_/-}: a two-binary bundle needs --upgrade-exrpd and --upgrade-version" >&2; exit 2; }
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
	plan=${plan:-$("$here/plan.sh")}
	plans=$(resolve_plans --upgrade-version "$upgrade_version" "$plan" "$genesis_version" "$genesis_plan")
	read -r plan genesis_plan <<<"$plans"
else
	genesis_plan=$(release_plan "$genesis_version" "$genesis_plan")
fi
build_one "$out" "$goos" "$goarch" "$genesis_version" "$genesis_plan" "$genesis_archive" "$genesis_checksums" \
	"$plan" "$upgrade_exrpd" "$upgrade_version" "$cosmovisor" "$cosmovisor_source"

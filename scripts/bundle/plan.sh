#!/usr/bin/env bash
# Print an upgrade plan or a release version.
#
# Usage: plan.sh checkout        the UpgradeName of the one handler in app/upgrades of this checkout
#        plan.sh release X.Y.Z   the plan that release X.Y.Z runs: the newest vX.Y.Z UpgradeName in
#                                app/upgrades of its tag. Fetches the tag if the checkout lacks it
#        plan.sh genesis PLAN    the newest published vX.Y.Z release below the major.minor of PLAN,
#                                the release that validators run at the PLAN halt. Needs gh
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)
# shellcheck source=SCRIPTDIR/lib.sh
source "$here/lib.sh"

upgrade_names() { sed -n 's/^[[:space:]]*UpgradeName[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$@"; }

checkout() {
	local plans
	plans=$(upgrade_names "$repo"/app/upgrades/*/constants.go)
	if [[ -z $plans || $plans == *$'\n'* ]]; then
		echo "want one UpgradeName in app/upgrades/*/constants.go, found: ${plans//$'\n'/ }" >&2
		exit 1
	fi
	echo "$plans"
}

# A tag can keep older handlers, and a minor release can have no plan of its own (v10.2.x
# runs v10.1.0), so the plan is read from the tag, never derived from the version.
release() {
	local tag=v$1 depth=() names plan name core
	if ! git -C "$repo" rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
		# CI checks out one commit. --depth would make a full clone shallow.
		[[ $(git -C "$repo" rev-parse --is-shallow-repository) == true ]] && depth=(--depth=1)
		git -C "$repo" fetch -q ${depth[@]+"${depth[@]}"} origin "refs/tags/$tag:refs/tags/$tag"
	fi
	names=$(git -C "$repo" ls-tree --name-only "$tag" app/upgrades/ | while read -r dir; do
		git -C "$repo" show "$tag:$dir/constants.go" 2>/dev/null || true
	done | upgrade_names)
	plan=$(grep -E "$PLAN_RE" <<<"$names" | sort -V | tail -n 1 || true)
	[[ -n $plan ]] || { echo "no vX.Y.Z UpgradeName in app/upgrades of $tag" >&2; exit 1; }
	# A newer plan with another name, such as v10.2.0-rc1 after v10.1.0, is the plan the
	# release runs. Skipping it would pick an older binary.
	while read -r name; do
		if [[ -z $name ]] || is_plan "$name"; then continue; fi
		core=$(grep -oE '^v[0-9]+\.[0-9]+\.[0-9]+' <<<"$name" || true)
		if [[ -z $core || ($core != "$plan" && $(printf '%s\n%s\n' "$core" "$plan" | sort -V | tail -n 1) == "$core") ]]; then
			echo "$tag has plan $name, which a bundle cannot carry (plans are vX.Y.Z), and which may be newer than $plan" >&2
			exit 1
		fi
	done <<<"$names"
	echo "$plan"
}

genesis() {
	local plan_major plan_minor version
	IFS=. read -r plan_major plan_minor _ <<<"${1#v}"
	# gh reads the repository from the checkout, or from GH_REPO.
	version=$(cd "$here" && gh release list --exclude-drafts --exclude-pre-releases --limit 1000 --json tagName --jq '.[].tagName' |
		awk -F. -v major="$plan_major" -v minor="$plan_minor" '
			/^v[0-9]+\.[0-9]+\.[0-9]+$/ { sub(/^v/, ""); if ($1 < major || ($1 == major && $2 < minor)) print }' |
		sort -V | tail -n 1)
	[[ -n $version ]] || { echo "no published release below plan $1" >&2; exit 1; }
	echo "$version"
}

usage() { echo "usage: $0 checkout | release X.Y.Z | genesis PLAN" >&2; exit 2; }
case ${1:-} in
checkout) [[ $# -eq 1 ]] || usage; checkout ;;
release) [[ $# -eq 2 ]] || usage; release "$2" ;;
genesis) [[ $# -eq 2 ]] || usage; genesis "$2" ;;
*) usage ;;
esac

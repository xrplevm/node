#!/usr/bin/env bash
# Print the plan of a genesis release: the newest vX.Y.Z UpgradeName in app/upgrades of
# its tag. A release runs the data of that plan, and its patch releases keep it. A tag can
# register older handlers too, and a minor release can have no plan of its own (v10.2.x
# runs v10.1.0), so the plan is read from the tag, never derived from the version.
# Fetches the tag when the checkout lacks it, as in CI, which checks out one commit.
# Usage: genesis-plan.sh X.Y.Z
set -euo pipefail

version=${1:?usage: $0 X.Y.Z}
repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
tag=v$version
if ! git -C "$repo" rev-parse -q --verify "refs/tags/$tag" >/dev/null; then
	# --depth on a full clone would make it shallow, so it is only for a shallow checkout.
	depth=()
	[[ $(git -C "$repo" rev-parse --is-shallow-repository) == true ]] && depth=(--depth=1)
	git -C "$repo" fetch -q ${depth[@]+"${depth[@]}"} origin "refs/tags/$tag:refs/tags/$tag"
fi
names=$(git -C "$repo" ls-tree --name-only "$tag" app/upgrades/ | while read -r dir; do
	git -C "$repo" show "$tag:$dir/constants.go" 2>/dev/null || true
done | sed -n 's/^[[:space:]]*UpgradeName[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p')
plain='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
plan=$(grep -E "$plain" <<<"$names" | sort -V | tail -n 1 || true)
[[ -n $plan ]] || { echo "no vX.Y.Z UpgradeName in app/upgrades of $tag" >&2; exit 1; }
# A bundle carries only vX.Y.Z plans. A newer plan with another name, such as v10.2.0-rc1
# after v10.1.0, is the plan the release runs, so skipping it would pick an older binary.
while read -r name; do
	[[ -z $name || $name =~ $plain ]] && continue
	core=$(grep -oE '^v[0-9]+\.[0-9]+\.[0-9]+' <<<"$name" || true)
	if [[ -z $core || ($core != "$plan" && $(printf '%s\n%s\n' "$core" "$plan" | sort -V | tail -n 1) == "$core") ]]; then
		echo "$tag has plan $name, which a bundle cannot carry (plans are vX.Y.Z), and which may be newer than $plan" >&2
		exit 1
	fi
done <<<"$names"
echo "$plan"

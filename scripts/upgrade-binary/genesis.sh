#!/usr/bin/env bash
# Print the default genesis version for PLAN: the newest published vX.Y.Z release, drafts
# and pre-releases excluded, whose major.minor is below the plan's. Validators run it at
# the plan's halt. Needs gh with access to the repository (GH_TOKEN in CI). gh reads the
# repository from the checkout of this script, or from GH_REPO when it is set.
# Usage: genesis.sh PLAN
set -euo pipefail

plan=${1:?usage: $0 PLAN}
cd "$(dirname "${BASH_SOURCE[0]}")"
IFS=. read -r plan_major plan_minor _ <<<"${plan#v}"
version=$(gh release list --exclude-drafts --exclude-pre-releases --limit 1000 --json tagName --jq '.[].tagName' |
	awk -F. -v major="$plan_major" -v minor="$plan_minor" '
		/^v[0-9]+\.[0-9]+\.[0-9]+$/ { sub(/^v/, ""); if ($1 < major || ($1 == major && $2 < minor)) print }' |
	sort -V | tail -n 1)
[[ -n $version ]] || { echo "no published release below plan $plan" >&2; exit 1; }
echo "$version"

#!/usr/bin/env bash
# Print the upgrade plan of this checkout: the UpgradeName of the one upgrade handler in
# app/upgrades. The bundle stages its upgrade binary under upgrades/<plan>.
set -euo pipefail

repo=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
plans=$(sed -n 's/^[[:space:]]*UpgradeName[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$repo"/app/upgrades/*/constants.go)
if [[ -z $plans || $plans == *$'\n'* ]]; then
	echo "want one UpgradeName in app/upgrades/*/constants.go, found: ${plans//$'\n'/ }" >&2
	exit 1
fi
echo "$plans"

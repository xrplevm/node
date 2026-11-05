#!/usr/bin/env bash
# Run the checks of the bundle, as CI does: go vet, go test and golangci-lint of the
# launcher module cmd/exrpd-bundle, which the root commands skip, then the tests of the
# patched Cosmovisor (cosmovisor.sh --test). golangci-lint must be on PATH with the version
# in .golangci-lint-version.
# Usage: check.sh
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)

want=$(tr -d '[:space:]' <"$repo/.golangci-lint-version")
if ! command -v golangci-lint >/dev/null; then
	echo "golangci-lint not found on PATH: install $want (.golangci-lint-version), for example with go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$want" >&2
	exit 1
fi
got=$(golangci-lint --version)
version_re='^golangci-lint has version v?([^ ]+)'
if [[ ! $got =~ $version_re || ${BASH_REMATCH[1]} != "${want#v}" ]]; then
	echo "$(command -v golangci-lint) is not $want (.golangci-lint-version): $got" >&2
	exit 1
fi

cd "$repo/cmd/exrpd-bundle"
GOWORK=off go vet ./...
GOWORK=off go test -count=1 ./...
GOWORK=off golangci-lint run --config ../../.golangci.yml --timeout=10m
"$here/cosmovisor.sh" --test

#!/usr/bin/env bash
# The bundle checks of CI: go vet, go test and golangci-lint of the launcher module
# cmd/exrpd-bundle (the root commands skip it), then the patched Cosmovisor tests.
# Needs golangci-lint on PATH at the version in .golangci-lint-version.
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

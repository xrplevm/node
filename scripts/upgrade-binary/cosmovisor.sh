#!/usr/bin/env bash
# Build Cosmovisor from the pinned upstream module source plus the committed patch.
# Usage: cosmovisor.sh GOOS GOARCH OUTPUT
#        cosmovisor.sh --test    run the patched module's own tests (go test ./...)
set -euo pipefail

if [[ ${1:-} == --test && $# -eq 1 ]]; then
	mode="test"
elif [[ $# -eq 3 ]]; then
	mode=build goos=$1 goarch=$2 out=$3
else
	echo "usage: $0 GOOS GOARCH OUTPUT | $0 --test" >&2
	exit 2
fi
here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/bundle.env
source "$here/bundle.env"

work=$(mktemp -d)
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT

# A throwaway module that requires Cosmovisor lets go mod verify check the extracted
# source in the module cache. go mod download checks only the zip, and the build uses
# the extracted directory.
printf 'module verify\n\ngo 1.23\n\nrequire %s %s\n' "$COSMOVISOR_MODULE" "$COSMOVISOR_VERSION" >"$work/go.mod"
if ! info=$(cd "$work" && GOWORK=off GOFLAGS=-mod=mod go mod download -json "$COSMOVISOR_MODULE@$COSMOVISOR_VERSION"); then
	echo "go mod download $COSMOVISOR_MODULE@$COSMOVISOR_VERSION: $(jq -r '.Error // empty' <<<"$info" 2>/dev/null)" >&2
	exit 1
fi
sum=$(jq -r .Sum <<<"$info")
[[ $sum == "$COSMOVISOR_SUM" ]] || { echo "cosmovisor module sum $sum, pinned $COSMOVISOR_SUM" >&2; exit 1; }
(cd "$work" && GOWORK=off go mod verify)

cp -R "$(jq -r .Dir <<<"$info")" "$work/src"
chmod -R u+w "$work/src"
# Inside another git work tree, git apply skips paths outside the current directory and
# still exits 0. The ceiling keeps git from finding a repository above $work.
(cd "$work/src" && export GIT_CEILING_DIRECTORIES=$work &&
	git apply --check "$here/$COSMOVISOR_PATCH" && git apply "$here/$COSMOVISOR_PATCH")
# The build and the tests read the dependencies from the module cache too.
(cd "$work/src" && GOWORK=off GOFLAGS=-mod=readonly go mod download && GOWORK=off go mod verify)

if [[ $mode == test ]]; then
	(cd "$work/src" && GOWORK=off GOFLAGS=-mod=readonly go test -count=1 ./...)
	exit
fi
mkdir -p "$(dirname "$out")"
out=$(cd "$(dirname "$out")" && pwd)/$(basename "$out")
(cd "$work/src" && GOWORK=off GOFLAGS=-mod=readonly CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
	go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$out" ./cmd/cosmovisor)

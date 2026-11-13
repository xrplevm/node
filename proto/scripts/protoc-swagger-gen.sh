#!/usr/bin/env sh

# Generates docs/static/openapi.yml from the query services whose REST routes
# the app registers. docs/swagger-config.json lists them: each url names the
# swagger file that buf generates from one proto file. The Makefile passes the
# proto root of each dependency, as pinned in go.mod.
# TestSwaggerConfig_MatchesServedQueryServices in app/ checks the list against
# the services that the app serves.

set -eu

: "${SDK_PROTO:?}" "${EVM_PROTO:?}" "${IBC_PROTO:?}" "${RATELIMIT_PROTO:?}"

out=tmp-swagger-gen
rm -rf "$out"
trap 'rm -rf "$out"' EXIT

protos=$(node -e '
for (const { url } of require("./docs/swagger-config.json").apis) {
  const m = url.match(/^\.\/tmp-swagger-gen\/(.+)\.swagger\.json$/);
  if (!m) throw new Error(`unexpected url: ${url}`);
  console.log(`${m[1]}.proto`);
}')

for file in $protos; do
  root=
  for candidate in proto "$SDK_PROTO" "$IBC_PROTO" "$RATELIMIT_PROTO" "$EVM_PROTO"; do
    [ -f "$candidate/$file" ] || continue
    if [ -n "$root" ]; then
      echo "$file is in $root and $candidate" >&2
      exit 1
    fi
    root=$candidate
  done
  if [ -z "$root" ]; then
    echo "$file is in no proto root" >&2
    exit 1
  fi
  buf generate "$root" --template proto/buf.gen.swagger.yaml --path "$root/$file" -o .
done

swagger-combine docs/swagger-config.json -o docs/static/openapi.yml -f yaml \
  --continueOnConflictingPaths true --includeDefinitions true

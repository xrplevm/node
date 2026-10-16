#!/usr/bin/env sh

# Generates docs/static/openapi.yml from the query services whose REST routes
# the app registers. The Makefile passes the proto root of each dependency, as
# pinned in go.mod. docs/swagger-config.json lists the generated files that
# are combined: keep both lists in sync with the module manager in app/app.go.

set -eu

: "${SDK_PROTO:?}" "${EVM_PROTO:?}" "${IBC_PROTO:?}" "${RATELIMIT_PROTO:?}"

out=tmp-swagger-gen
rm -rf "$out"
trap 'rm -rf "$out"' EXIT

gen() {
  root=$1
  shift
  for file in "$@"; do
    buf generate "$root" --template proto/buf.gen.swagger.yaml --path "$root/$file" -o .
  done
}

gen "$SDK_PROTO" \
  cosmos/auth/v1beta1/query.proto \
  cosmos/bank/v1beta1/query.proto \
  cosmos/base/node/v1beta1/query.proto \
  cosmos/base/tendermint/v1beta1/query.proto \
  cosmos/consensus/v1/query.proto \
  cosmos/distribution/v1beta1/query.proto \
  cosmos/evidence/v1beta1/query.proto \
  cosmos/gov/v1/query.proto \
  cosmos/gov/v1beta1/query.proto \
  cosmos/slashing/v1beta1/query.proto \
  cosmos/staking/v1beta1/query.proto \
  cosmos/tx/v1beta1/service.proto \
  cosmos/upgrade/v1beta1/query.proto

gen "$IBC_PROTO" \
  ibc/applications/transfer/v1/query.proto \
  ibc/core/channel/v1/query.proto \
  ibc/core/channel/v2/query.proto \
  ibc/core/client/v1/query.proto \
  ibc/core/connection/v1/query.proto

gen "$RATELIMIT_PROTO" ratelimit/v1/query.proto

gen "$EVM_PROTO" \
  cosmos/evm/erc20/v1/query.proto \
  cosmos/evm/feemarket/v1/query.proto \
  cosmos/evm/vm/v1/query.proto

gen proto poa/query.proto

swagger-combine docs/swagger-config.json -o docs/static/openapi.yml -f yaml \
  --continueOnConflictingPaths true --includeDefinitions true

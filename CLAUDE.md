# CLAUDE.md

## Project Overview

XRPL EVM Sidechain node (`exrpd`) — a Cosmos SDK-based blockchain node that runs an EVM-compatible sidechain for the XRP Ledger. Uses Proof-of-Authority consensus with authority-gated validator management.

## Build & Test

```bash
make build                # Build the exrpd binary
make test                 # Run unit tests
make lint                 # Run linter
go vet ./...              # Vet all packages
```

The root commands do not cover `cmd/exrpd-upgrade/`, a separate Go module. Run its CI checks from the repo root with the `golangci-lint` version in `.golangci-lint-version`:

```bash
(cd cmd/exrpd-upgrade && GOWORK=off go vet ./...)
(cd cmd/exrpd-upgrade && GOWORK=off go test -count=1 ./...)
(cd cmd/exrpd-upgrade && golangci-lint run --config ../../.golangci.yml --timeout=10m)
scripts/upgrade-binary/cosmovisor.sh --test
```

`scripts/upgrade-binary/e2e.sh` needs release assets supplied via `CACHE`. See [docs/upgrade-binary-dev.md](docs/upgrade-binary-dev.md#end-to-end-test).

## Project Structure

```
cmd/exrpd/              # CLI entry point and node configuration
cmd/exrpd-upgrade/      # Upgrade binary launcher (nested Go module)
scripts/upgrade-binary/ # Bundle build, release, patched Cosmovisor, e2e
app/                    # Cosmos SDK app wiring, ante handlers, upgrade handlers
app/ante/               # Transaction validation (routes EVM vs Cosmos txs)
x/poa/                  # Proof-of-Authority module (validator add/remove)
proto/                  # Protobuf definitions (gRPC API surface)
types/                  # Core types (address prefixes, denominations)
testutil/               # Test helpers
tests/                  # Integration tests
security/               # Security audit scope and threat model
```

## Security Audits

This repo is configured for monthly AI-assisted security audits. Before performing any security-related review, read:

1. **`.claude/security/SECURITY_SCOPE.md`**: audit runbook with components, priority tiers, trust boundaries, focus areas, and the reporting workflow.
2. **`.claude/security/THREAT_MODEL.md`**: the only threat model, with system overview, trust levels, invariants, attack surfaces, out of scope, and severity. Anthropic OSS Scanner reads it too.

### Running an audit

Spawn one `security-reviewer` agent per focus area defined in SECURITY_SCOPE.md. Findings are reported as issues in the **private** `peersyst/security` repo with `repo:xrplevm/node` label. Before reporting, query existing issues to avoid duplicates.

```
Read .claude/security/SECURITY_SCOPE.md and .claude/security/THREAT_MODEL.md. Then perform a security audit focused on: [AREA]. Check the components listed in the scope. Report findings to peersyst/security following the reporting workflow in SECURITY_SCOPE.md.
```

### Key security invariants

The invariants that must always hold are listed in "Invariants" in `.claude/security/THREAT_MODEL.md`. Verify them in every audit.

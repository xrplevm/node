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

## Project Structure

```
cmd/exrpd/          # CLI entry point and node configuration
app/                # Cosmos SDK app wiring, ante handlers, upgrade handlers
app/ante/           # Transaction validation (routes EVM vs Cosmos txs)
x/poa/              # Proof-of-Authority module (validator add/remove)
proto/              # Protobuf definitions (gRPC API surface)
types/              # Core types (address prefixes, denominations)
testutil/           # Test helpers
tests/              # Integration tests
security/           # Security audit scope and threat model
```

## Security Audits

This repo is configured for monthly AI-assisted security audits. Before performing any security-related review, read:

1. **`.claude/security/SECURITY_SCOPE.md`** — Defines what to audit, trust boundaries, priority tiers, and focus areas.
2. **`.claude/security/THREAT_MODEL.md`** — Assets, adversary profiles, attack surfaces, and Cosmos SDK-specific vulnerability patterns.

### Running an audit

Spawn one `security-reviewer` agent per focus area defined in SECURITY_SCOPE.md. Findings are reported as issues in the **private** `peersyst/security` repo with `repo:xrplevm/node` label. Before reporting, query existing issues to avoid duplicates.

```
Read .claude/security/SECURITY_SCOPE.md and .claude/security/THREAT_MODEL.md. Then perform a security audit focused on: [AREA]. Check the components listed in the scope. Report findings to peersyst/security following the reporting workflow in SECURITY_SCOPE.md.
```

### Key security invariants

These must always hold and should be verified in every audit:

- **Authority gate**: `x/poa` message handlers must check `msg.Authority == k.authority` before any validator set mutation.
- **Ante handler routing**: EVM and Cosmos transactions must be routed to their respective handler chains with no bypass path.
- **Extension option rejection**: Unknown extension options must be rejected by the ante handler.
- **No non-determinism in state machine**: No maps iterated for state, no goroutines, no `time.Now()` in consensus code.
- **IBC rate limiting**: The rate limiting module must be in the IBC middleware stack.

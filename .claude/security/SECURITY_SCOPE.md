# Security Audit Scope

This document defines the scope and priorities for AI-assisted security audits of the XRPL EVM Sidechain node (`exrpd`).

## Project Context

This is a **Cosmos SDK-based blockchain node** that runs an EVM-compatible sidechain for the XRP Ledger. It uses Proof-of-Authority (POA) consensus with authority-gated validator management. The node handles real financial assets (`axrp`/`xrp`) and executes smart contracts via the EVM.

**Security criticality: HIGH** — vulnerabilities can lead to loss of funds, chain halts, or consensus failures.

## In-Scope Components

### Critical

| Component | Path | Why |
|---|---|---|
| Ante handlers | `app/ante/` | Transaction validation gate — bypasses here affect all txs |
| POA module | `x/poa/` | Authority checks for validator management |
| App wiring | `app/app.go` | Module registration, keeper initialization, upgrade handlers |
| Msg filter | `app/msg_filter.go` | Msg methods that the app does not route. A removed entry reopens a msg path for users and gov |
| Upgrade handlers | `app/upgrades/` | State migrations run during chain upgrades — bugs here can corrupt state, brick the chain, or open exploit windows post-upgrade |
| CLI entry point | `cmd/exrpd/` | Node configuration, default values, TLS setup |

### High Priority

| Component | Path | Why |
|---|---|---|
| Protocol buffers | `proto/` | API surface definition, message validation |
| Type definitions | `types/` | Address prefixes, coin denominations, constants |
| App configuration | `app/config.go` | Chain constants that affect consensus |
| Go module dependencies | `go.mod` | Supply chain risk from third-party modules |
| Encoding config | `app/encoding.go` | Custom signers, legacy type registration, EIP-712 setup — incorrect encoding can cause signature bypass or tx deserialization issues |

### Infrastructure

| Component | Path | Why |
|---|---|---|
| CI workflows | `.github/workflows/` | Secret exposure in env/steps, untrusted input in PR triggers, permission scoping |
| Release pipeline | `.goreleaser.yml`, `.github/workflows/goreleaser.yml`, `.github/workflows/release.yml` | Binary integrity — compromised release artifacts affect all node operators |
| Dockerfiles | `Dockerfile`, `.oss-scanner/Dockerfile` | Secret leakage (SSH keys in build), unpinned base images, multi-stage hygiene |
| Makefile | `Makefile` | Build flags, CGO settings, embedded variables |
| Branch protection | `.github/workflows/branch.yml`, `.github/workflows/pull-request.yml` | Enforce review requirements, prevent force-push to main |
| CODEOWNERS | `.github/CODEOWNERS` | Ensure security-sensitive paths require appropriate reviewers |

### Lower Priority

| Component | Path | Why |
|---|---|---|
| Test utilities | `testutil/` | Only matters if test helpers mask real bugs |
| Integration tests | `tests/` | Verify test coverage gaps |
| Linter/coverage config | `.golangci.yml`, `.testcoverage.yml` | Verify security linters are enabled (e.g., `gosec`) |

## Out of Scope

See "Out of Scope" in `.claude/security/THREAT_MODEL.md`.

## Trust Boundaries

These are the points where untrusted input enters the system. Every audit should verify these boundaries are intact.

```
                         +-------------------------+
   External RPCs ------> | gRPC / REST / JSON-RPC   |
                         +------------+------------+
                                      |
                                      v
                         +------------+------------+
                         |      Ante Handlers       |  <-- Signature, fee, extension validation
                         +------------+------------+
                                      |
                    +-----------------+-----------------+
                    |                                   |
                    v                                   v
       +------------+------------+         +------------+------------+
       |   Cosmos SDK Modules    |         |     EVM Execution       |
       |                         |         |                         |
       |  auth, bank, staking,   |         |  Smart contract calls   |
       |  gov, slashing,         |         |  via go-ethereum        |
       |  distribution,          |         +------------+------------+
       |  consensus, upgrade     |                      |
       +------------+------------+                      |
                    |                                   |
                    +------- POA (validator mgmt) ------+
                    |        Authority gate             |
                    |                                   |
                    v                                   v
       +------------+------------+         +------------+------------+
       |   State Store (IAVL)    |  <---   |  EVM State (accounts,   |
       |   Consensus-critical    |         |  contracts, storage)    |
       +-------------------------+         +-------------------------+
                    ^
                    |
       +------------+------------+
       |     IBC Middleware       |  <-- IBC relayer messages
       |  transfer,              |
       |  rate-limiting          |
       +-------------------------+
```

1. **Network boundary**: gRPC, REST (gRPC-gateway), JSON-RPC endpoints accept external requests.
2. **Transaction validation**: Ante handlers validate signatures, fees, and extension options before execution.
3. **Authority gate**: POA message handlers check `msg.Authority == k.authority` before allowing validator changes.
4. **IBC boundary**: Inter-blockchain messages arrive via IBC with rate limiting (`ibc-apps/modules/rate-limiting`).
5. **EVM boundary**: User-submitted bytecode executes in the EVM sandbox via `go-ethereum`.
6. **Cosmos SDK module boundaries**: The node wires the following upstream SDK modules, each with its own trust assumptions. While their internals are out of scope, the way they are configured, initialized, and composed in `app/app.go` is in scope:

   | Module | Keeper | Security relevance |
   |---|---|---|
   | `auth` | `AccountKeeper` | Account creation, sequence numbers, signature verification |
   | `bank` | `BankKeeper` | Token minting, burning, transfers — blocked addresses list |
   | `staking` | `StakingKeeper` | Validator bonding (interacts with POA — verify no bypass) |
   | `slashing` | `SlashingKeeper` | Validator penalties — misconfiguration can brick validators |
   | `distribution` | `DistrKeeper` | Reward distribution — incorrect wiring can leak or lock funds |
   | `gov` | `GovKeeper` | On-chain governance proposals — can trigger arbitrary module changes |
   | `evidence` | `EvidenceKeeper` | Double-sign evidence handling |
   | `upgrade` | `UpgradeKeeper` | Chain upgrade coordination — ties into `app/upgrades/` |
   | `consensus` | `ConsensusParamsKeeper` | Block size, gas limits, validator set changes |
   | `transfer` | `TransferKeeper` | IBC token transfers |
   | `rate-limiting` | `RateLimitKeeper` | IBC rate limiting — misconfiguration can allow or block transfers |

## Audit Focus Areas

Each audit should spawn a separate agent per focus area below. All areas are reviewed every audit cycle — run them in parallel for efficiency. Each area maps to one or more in-scope components.

| Area | Components | What to look for |
|---|---|---|
| **POA module** | POA module | Authority check correctness (`msg.Authority == k.authority`), keeper initialization, genesis state validation, alternative code paths to modify the validator set, Bech32 address validation. Cross-module interactions: token minting/burning via BankKeeper (correct module accounts and pool names), staking state mutations (delegation removal, unbonding slashing with correct slash factor, validator token removal, status-based pool routing for Bonded vs Unbonding/Unbonded), staking hooks invocation order (BeforeValidatorModified, BeforeValidatorSlashed), and edge cases around existing delegations/unbondings during add/remove flows |
| **Authorization & access control** | App wiring, Msg filter | Missing permission checks, privilege escalation via `gov`, keeper exposure across module boundaries, reintroduction of `authz` (removed in v12), a removed entry in `unroutedMsgMethods` (`app/msg_filter.go`), or a module that registers a Msg service outside the filter |
| **Transaction validation** | Ante handlers, App wiring | Ante handler routing between EVM/Cosmos, extension option validation, signature bypass, fee circumvention, a fee granter accepted by `DeductFeeDecorator` (x/feegrant removed in v12) |
| **EVM integration** | Ante handlers, App wiring | EVM tx routing, gas metering between Cosmos and EVM layers, precompile registration, reintroduction of the staking or distribution precompile (removed in v12), or any precompile that calls a msg server for a method in `unroutedMsgMethods` (the gov precompile `0x805` calls `voteWeighted`) |
| **IBC & cross-chain** | App wiring | Rate limiting config, IBC message validation, channel security, reintroduction of ICA host (removed in v12) or any IBC app that executes messages |
| **Upgrade handlers** | Upgrade handlers | State migration correctness, non-determinism, missing or stale upgrade handlers |
| **Configuration & secrets** | CLI entry point, App configuration, Encoding config | Hardcoded secrets, insecure defaults, TLS config, env var handling |
| **Consensus & state machine** | App wiring, POA module, Type definitions | Non-determinism (map iteration, goroutines, `time.Now()`), module account permissions (Minter/Burner scoping), genesis state validation, address prefix correctness |
| **API surface** | Protocol buffers, Ante handlers | Protobuf message validation, unexpected fields, gRPC/REST endpoint exposure, query handler authorization |
| **Dependency audit** | Go module dependencies | Known CVEs, pinned vs floating versions, dependency confusion, a new Msg method in a dependency bump that reopens a dropped msg path |
| **Infrastructure & CI** | CI workflows, Release pipeline, Dockerfiles, Makefile, Branch protection, CODEOWNERS | Secret exposure, unpinned base images, build reproducibility, CI permission scoping |
| **Test coverage** | Test utilities, Integration tests, Linter/coverage config | Test coverage gaps, security linters enabled (e.g., `gosec`), test helpers that mask real bugs |

### How to run

Spawn one `security-reviewer` agent per area. Example:

```
claude "Read .claude/security/SECURITY_SCOPE.md and .claude/security/THREAT_MODEL.md. Then perform a security audit focused on: [AREA]. Check the components listed in the scope. Report findings to peersyst/security following the reporting workflow below."
```

Or run all areas in parallel by spawning multiple agents in a single prompt.

## Reporting Workflow

All findings are reported as GitHub Issues in the **private** `peersyst/security` repository.

### Labels

Each issue must have labels from these categories:

| Category | Labels | Purpose |
|---|---|---|
| Repo | `repo:xrplevm/node`, `repo:xrplevm/evm`, etc. | Filter findings by source repo |
| Severity | `critical`, `high`, `medium`, `low`, `info` | Triage priority |
| Status | `pending`, `accepted`, `discarded` | Triage state (closed = done) |

### Issue format

```
Title: [SEVERITY] Short description of the finding

Body:
## Area

<Focus area from the audit: e.g., POA module, Authorization & access control, EVM integration, etc.>

## Finding

<Description of the vulnerability or concern>

## Location

<File path and line numbers>

## Impact

<What could go wrong and under what conditions>

## Recommendation

<Suggested fix or mitigation>

## Evidence

<Code snippets, reasoning, or proof of concept>
```

### Creating a finding

```bash
gh issue create -R peersyst/security \
  --title "[HIGH] Description of finding" \
  --label "repo:xrplevm/node,high,pending" \
  --body "$(cat <<'EOF'
## Area
POA module

## Finding
...
## Location
...
## Impact
...
## Recommendation
...
## Evidence
...
EOF
)"
```

### Lifecycle

1. **AI creates issue** with `pending` label
2. **Human triages**: changes to `accepted` (valid, needs fix) or `discarded` (with comment explaining why)
3. **Developer fixes**: closes the issue with a reference to the fixing commit/PR
4. **Next audit**: AI queries existing issues before reporting to avoid duplicates:

```bash
gh issue list -R peersyst/security --label "repo:xrplevm/node" --state all --json number,title,labels,state
```

### Deduplication rules

- **`pending` or `accepted`**: Do not re-raise. The finding is already tracked.
- **`discarded`**: Do not re-raise the same finding. Read the discard reason — only re-raise if circumstances have materially changed.
- **Closed (done)**: Verify the fix is still in place. If the vulnerability has regressed, open a **new** issue referencing the original.

## Severity Classification

Use the severity scale in `.claude/security/THREAT_MODEL.md` (Critical, High, Medium, Low, Info).

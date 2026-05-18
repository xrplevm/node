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
| Dockerfiles | `Dockerfile`, `tools/cosmovisor/Dockerfile` | Secret leakage (SSH keys in build), unpinned base images, multi-stage hygiene |
| Makefile | `Makefile` | Build flags, CGO settings, embedded variables |
| Branch protection | `.github/workflows/branch.yml`, `.github/workflows/pull-request.yml` | Enforce review requirements, prevent force-push to main |
| CODEOWNERS | `.github/CODEOWNERS` | Ensure security-sensitive paths require appropriate reviewers |

### Lower Priority

| Component | Path | Why |
|---|---|---|
| Test utilities | `testutil/` | Only matters if test helpers mask real bugs |
| Integration tests | `tests/` | Verify test coverage gaps |
| Cosmovisor setup | `.github/workflows/cosmovisor.yml`, `tools/cosmovisor/` | Upgrade orchestration — misconfiguration can brick nodes during upgrades |
| Linter/coverage config | `.golangci.yml`, `.testcoverage.yml` | Verify security linters are enabled (e.g., `gosec`) |

## Out of Scope

- **Upstream Cosmos SDK** — audited separately by the Cosmos ecosystem. Treat as a dependency: check for known CVEs in `go.mod`, but do not audit internal SDK code. Flag if we override or patch upstream behavior.
- **Upstream CometBFT** — same as above.
- **Upstream `cosmos/evm`** — same as above. Treat as a dependency like Cosmos SDK. Check for version pinning and known vulnerabilities, but do not audit its internals.
- **Smart contracts deployed on the EVM** — application-layer, not node-layer.

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
       |  gov, authz, slashing,  |         |  via go-ethereum        |
       |  distribution, crisis,  |         +------------+------------+
       |  feegrant, params,      |                      |
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
       |  transfer, ica/host,    |
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
   | `crisis` | `CrisisKeeper` | Invariant checking — verify invariants are registered |
   | `authz` | `AuthzKeeper` | Grant-based authorization — overly broad grants can escalate privileges |
   | `feegrant` | `FeeGrantKeeper` | Fee delegation — potential for fee exhaustion attacks |
   | `evidence` | `EvidenceKeeper` | Double-sign evidence handling |
   | `upgrade` | `UpgradeKeeper` | Chain upgrade coordination — ties into `app/upgrades/` |
   | `params` | `ParamsKeeper` | Legacy parameter changes — verify migration to consensus params |
   | `consensus` | `ConsensusParamsKeeper` | Block size, gas limits, validator set changes |
   | `transfer` | `TransferKeeper` | IBC token transfers |
   | `ica/host` | `ICAHostKeeper` | Interchain accounts — remote chains can execute messages on this chain |
   | `rate-limiting` | `RateLimitKeeper` | IBC rate limiting — misconfiguration can allow or block transfers |

## Audit Focus Areas

Each audit should spawn a separate agent per focus area below. All areas are reviewed every audit cycle — run them in parallel for efficiency. Each area maps to one or more in-scope components.

| Area | Components | What to look for |
|---|---|---|
| **POA module** | POA module | Authority check correctness (`msg.Authority == k.authority`), keeper initialization, genesis state validation, alternative code paths to modify the validator set, Bech32 address validation. Cross-module interactions: token minting/burning via BankKeeper (correct module accounts and pool names), staking state mutations (delegation removal, unbonding slashing with correct slash factor, validator token removal, status-based pool routing for Bonded vs Unbonding/Unbonded), staking hooks invocation order (BeforeValidatorModified, BeforeValidatorSlashed), and edge cases around existing delegations/unbondings during add/remove flows |
| **Authorization & access control** | App wiring | Missing permission checks, privilege escalation via `authz`/`gov`, keeper exposure across module boundaries |
| **Transaction validation** | Ante handlers, App wiring | Ante handler routing between EVM/Cosmos, extension option validation, signature bypass, fee circumvention |
| **EVM integration** | Ante handlers, App wiring | EVM tx routing, gas metering between Cosmos and EVM layers, precompile registration |
| **IBC & cross-chain** | App wiring | Rate limiting config, IBC message validation, channel security, ICA host permissions |
| **Upgrade handlers** | Upgrade handlers | State migration correctness, non-determinism, missing or stale upgrade handlers |
| **Configuration & secrets** | CLI entry point, App configuration, Encoding config | Hardcoded secrets, insecure defaults, TLS config, env var handling |
| **Consensus & state machine** | App wiring, POA module, Type definitions | Non-determinism (map iteration, goroutines, `time.Now()`), module account permissions (Minter/Burner scoping), genesis state validation, address prefix correctness |
| **API surface** | Protocol buffers, Ante handlers | Protobuf message validation, unexpected fields, gRPC/REST endpoint exposure, query handler authorization |
| **Dependency audit** | Go module dependencies | Known CVEs, pinned vs floating versions, dependency confusion |
| **Infrastructure & CI** | CI workflows, Release pipeline, Dockerfiles, Makefile, Branch protection, CODEOWNERS | Secret exposure, unpinned base images, build reproducibility, CI permission scoping |
| **Test coverage** | Test utilities, Integration tests, Linter/coverage config, Cosmovisor setup | Test coverage gaps, security linters enabled (e.g., `gosec`), test helpers that mask real bugs |

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

- **CRITICAL**: Exploitable now, loss of funds or chain halt possible
- **HIGH**: Exploitable with effort, significant impact
- **MEDIUM**: Requires specific conditions, moderate impact
- **LOW**: Minor issue, defense-in-depth concern
- **INFO**: Observation, no direct security impact but worth noting

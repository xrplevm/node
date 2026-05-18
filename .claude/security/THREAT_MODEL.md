# Threat Model

Threat model for the XRPL EVM Sidechain node. This document identifies assets, adversaries, and attack surfaces to guide security audits.

## Assets

| Asset | Description | Impact if compromised |
|---|---|---|
| Validator set | POA-managed list of block producers | Chain halt, censorship, double-spend |
| User funds | `axrp`/`xrp` balances in accounts | Direct financial loss |
| Consensus state | IAVL state tree, block history | Fork, state corruption, rollback |
| Authority key | The address authorized to manage validators | Full control of validator set |
| Node operator keys | Keys used to sign blocks | Impersonation, invalid blocks |
| IBC channels | Cross-chain communication paths | Cross-chain theft, replay attacks |
| EVM state | Smart contract storage and code | Contract manipulation, fund theft |

## Adversary Profiles

| Adversary | Capability | Goal |
|---|---|---|
| External attacker | Sends crafted transactions via RPC | Steal funds, halt chain, bypass fees |
| Malicious validator | Controls one validator node | Censor txs, manipulate ordering |
| Supply chain attacker | Compromises a Go module dependency | Backdoor node binary |
| Rogue insider | Has access to authority key or CI | Add malicious validators, push bad builds |
| Cross-chain attacker | Controls an IBC-connected chain | Forge IBC messages, drain bridge |
| ICA controller | Controls an interchain account on a remote chain | Execute arbitrary messages on this chain via ICA host |

## Attack Surfaces

### 1. Transaction Processing

**Entry**: gRPC, REST, JSON-RPC endpoints

**Threats**:
- Malformed transactions that bypass ante handler validation
- Extension option confusion (EVM tx routed as Cosmos tx or vice versa)
- Fee bypass or gas manipulation
- Signature validation flaws

**Key files**: `app/ante/ante.go`

**What to check**:
- All transaction types are routed to the correct handler chain
- Unknown extension options are rejected (verify `rejectExtensionOption` handler)
- No path exists to skip signature verification
- Fee deduction cannot be circumvented

### 2. Validator Management (POA)

**Entry**: `MsgAddValidator`, `MsgRemoveValidator` transactions

**Threats**:
- Authority bypass allowing unauthorized validator changes
- Authority address manipulation at initialization
- Validator set corruption leading to consensus failure
- Incorrect token minting/burning during add/remove flows leading to inflation or locked funds
- Wrong pool routing (Bonded vs NotBonded) during validator removal causing state inconsistency
- Incomplete cleanup of delegations/unbondings leaving orphaned state
- Staking hooks (`BeforeValidatorModified`, `BeforeValidatorSlashed`) called in wrong order or swallowed silently
- Edge cases: adding a validator that already has delegations to other validators, removing a validator mid-unbonding

**Key files**: `x/poa/keeper/keeper.go`, `x/poa/keeper/msg_server_add_validator.go`, `x/poa/keeper/msg_server_remove_validator.go`, `x/poa/types/expected_keepers.go`

**What to check**:
- `msg.Authority == k.authority` check is present and correct in both handlers
- Authority address is validated (Bech32) at keeper construction
- No alternative code path to modify the validator set
- Genesis state cannot inject unexpected authority
- `ExecuteAddValidator`: coins minted to POA module account and sent to validator, all pre-existing balance/delegation/unbonding checks are exhaustive
- `ExecuteRemoveValidator`: correct pool name used per validator status (BondedPoolName vs NotBondedPoolName), all unbonding delegations slashed before token removal, self-delegation unbonded last
- BankKeeper interactions use correct module names (not arbitrary strings)
- Slash factor of `math.LegacyOneDec()` (100%) is intentional for POA removal

### 3. EVM Execution

**Entry**: Ethereum-format transactions via JSON-RPC or Cosmos tx wrapping

**Threats**:
- EVM sandbox escape
- Incorrect gas accounting between Cosmos and EVM layers
- State inconsistency between EVM and Cosmos views
- Precompile vulnerabilities

**Key files**: Upstream `cosmos/evm` module (out of scope for deep audit, but check integration points in `app/app.go`)

**What to check**:
- EVM module is wired with correct ante handlers
- Custom precompiles (if any) are registered correctly
- Gas limit configurations are sane

### 4. IBC (Inter-Blockchain Communication)

**Entry**: IBC relayer messages

**Threats**:
- IBC token transfer manipulation
- Rate limiting bypass
- Channel/port hijacking
- Timeout exploitation

**Key files**: IBC module wiring in `app/app.go`, rate limiting configuration

**What to check**:
- Rate limiting module is correctly positioned in the IBC stack
- Transfer channels have appropriate limits configured
- No unprotected IBC ports

### 5. Configuration & Initialization

**Entry**: Config files, CLI flags, genesis state

**Threats**:
- Insecure default values (e.g., TLS disabled, permissive CORS)
- Hardcoded keys or secrets
- Genesis state manipulation
- Environment variable injection

**Key files**: `cmd/exrpd/cmd/config.go`, `cmd/exrpd/cmd/root.go`, `app/config.go`

**What to check**:
- Default JSON-RPC/gRPC bindings are not `0.0.0.0` without TLS
- No secrets in source code or default configs
- Custom app config template does not expose sensitive fields
- Address prefixes and coin types match intended values

### 6. Supply Chain & Infrastructure

**Entry**: `go.mod`, `Dockerfile`, CI pipelines, GitHub Actions workflows

**Threats**:
- Compromised Go module dependency
- Dependency confusion attacks
- Build-time secret exposure
- Unpinned base images in Dockerfile
- CI workflow injection via untrusted PR input (`pull_request_target` misuse, expression injection in `${{ }}`)
- Secret leakage in workflow logs or artifacts
- Overly permissive workflow permissions (`contents: write`, `id-token: write`)
- Compromised release pipeline producing tampered binaries
- Missing or insufficient CODEOWNERS coverage on security-sensitive paths

**Key files**: `go.mod`, `go.sum`, `Dockerfile`, `tools/cosmovisor/Dockerfile`, `Makefile`, `.goreleaser.yml`, `.github/workflows/`, `.github/CODEOWNERS`

**What to check**:
- All dependencies are pinned to specific versions (not branches)
- `go.sum` is committed and checked
- Dockerfile does not leak SSH keys into the final image (multi-stage build)
- Base images are pinned to digest, not just tag
- GitHub Actions workflows use least-privilege permissions
- No use of `pull_request_target` with checkout of PR code
- Secrets are not interpolated into shell commands (use environment variables instead)
- GoReleaser config produces reproducible builds
- CODEOWNERS requires review for `x/poa/`, `app/ante/`, `app/upgrades/`, `.github/workflows/`

### 7. Upgrade Handlers

**Entry**: Chain upgrade proposals triggering state migrations

**Threats**:
- Malicious or buggy migration logic corrupting state during upgrade
- Non-deterministic migration code causing consensus failure across nodes
- Missing upgrade handler for a scheduled upgrade bricking the chain
- Stale upgrade handlers from previous versions executing unexpectedly
- Upgrade handler accessing or modifying state outside its intended scope

**Key files**: `app/upgrades/v5/` through `app/upgrades/v10/`, upgrade registration in `app/app.go`

**What to check**:
- Each upgrade handler is registered exactly once with the correct version name
- Migration logic is deterministic (no maps iterated for state, no goroutines, no time-dependent logic)
- Upgrade handlers only modify state relevant to their version
- No upgrade handler can be re-executed after completion
- Keeper references in upgrade handlers (`app/upgrades/v9/keepers.go`, `app/upgrades/v10/keepers.go`) only expose necessary keepers

### 8. Governance

**Entry**: On-chain governance proposals

**Threats**:
- Governance proposal executing arbitrary module messages (parameter changes, software upgrades, community spend)
- Parameter changes that put the chain in an unsafe state (e.g., unbonding time to 0, max validators to 0)
- Governance bypass via `authz` grants to the gov module account
- POA module exposes `GovKeeper.SubmitProposal` — verify it cannot be abused to self-authorize

**Key files**: Gov module wiring in `app/app.go`, `x/poa/types/expected_keepers.go` (GovKeeper interface)

**What to check**:
- Gov module message allowlist is configured (if applicable) to restrict dangerous message types
- Parameter boundaries prevent unsafe values
- `authz` grants cannot escalate to gov-level authority
- POA's GovKeeper interface is minimal and cannot be used to submit arbitrary proposals

## Cosmos SDK-Specific Concerns

These are common vulnerability patterns in Cosmos SDK chains that should be checked each audit:

- **Module account permissions**: Check that module accounts have only the permissions they need. POA module has `Minter` and `Burner` — verify no other module has unexpected mint/burn access. Verify `BondedPoolName` and `NotBondedPoolName` are the only accounts used for staking token burns.
- **Keeper exposure**: Verify keepers are not passed beyond their intended module boundaries. Check `app/app.go` for keepers passed to modules that shouldn't have them. POA receives BankKeeper and StakingKeeper — verify these are the correct scoped interfaces (`x/poa/types/expected_keepers.go`), not the full keeper implementations.
- **Parameter changes**: Check if on-chain parameter changes can lead to unsafe states (e.g., `MaxValidators` to 0 in staking params, unbonding time to 0, bond denom change breaking POA).
- **Non-determinism**: Any use of maps iterated for state, goroutines, or `time.Now()` in state machine code causes consensus failures. Check all code paths in `x/poa/`, `app/ante/`, and `app/upgrades/`.
- **ICA host allow list**: Verify which message types the ICA host module is allowed to execute. An open allow list lets remote chains execute arbitrary messages on this chain.

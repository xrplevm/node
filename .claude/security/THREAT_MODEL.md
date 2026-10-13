# Threat Model: XRPL EVM Sidechain node (`exrpd`)

This is the only threat model for this repository. Internal AI-assisted audits and external scanners (Anthropic OSS Scanner) both use it. The audit runbook (component tables, focus areas, internal reporting workflow) is `.claude/security/SECURITY_SCOPE.md`.

## System Overview

`exrpd` is the node software of the XRPL EVM Sidechain. It is a Cosmos SDK chain with an EVM (`cosmos/evm`). Validators are managed by Proof of Authority (`x/poa`), not by open staking. The chain holds real value: the native `axrp` token, ERC-20 tokens, and IBC assets.

- `app/app.go`: wires the modules. Standard SDK modules (auth, bank, staking, slashing, distribution, gov, feegrant, upgrade, evidence, consensus, params, crisis), IBC (transfer, rate limiting), and cosmos/evm (`vm`, `feemarket`, `erc20`, static precompiles).
- `app/ante/`: the ante handler. It routes each transaction by its first extension option:
  - `/cosmos.evm.vm.v1.ExtensionOptionsEthereumTx` -> EVM ante chain (`evm_handler.go`)
  - `/cosmos.evm.ante.v1.ExtensionOptionDynamicFeeTx` -> Cosmos ante chain (`cosmos_handler.go`)
  - any other extension option -> rejected
  - no extension option -> Cosmos ante chain
- `x/poa/`: our own module. It adds and removes validators.
  - `MsgAddValidator` and `MsgRemoveValidator`: the signer must be the module authority (the gov module account). The add flow mints the bond tokens to the validator. The remove flow slashes 100%, unbonds, and burns.
  - `MsgSelfRemoveValidator`: a validator removes itself. The signer is the validator account.
  - `x/poa/ante/poa.go`: blocks `MsgDelegate`, `MsgUndelegate`, `MsgBeginRedelegate`, `MsgCancelUnbondingDelegation` and `MsgCreateValidator` after height 0. Users must not change the validator set or stake through staking messages.
- `app/upgrades/`: the state migration for the current chain upgrade (`v12`). The release tags contain the earlier handlers.
- `cmd/exrpd/`: CLI, node config defaults, and `in-place-testnet`.

## Untrusted Input and Trust Levels

Treat all of these as attacker-controlled:

- Any transaction from any account, Cosmos or Ethereum format, sent through CometBFT RPC, gRPC, REST, or Ethereum JSON-RPC. This includes EVM calls to precompiles.
- EVM bytecode and calldata, including contracts that call the static precompiles (for example distribution, bank, erc20, ics20, gov, slashing).
- IBC packets and acknowledgements from a counterparty chain, which may be malicious.
- Queries to the public gRPC, REST and JSON-RPC endpoints.

Trusted (do not report attacks that need these without another bug):

- The authority (gov module account) and anyone who can pass a governance proposal.
- A node operator's own config files, CLI flags, keys and genesis file.
- 2/3+ of the validators acting together.

A single validator is trusted only for its own node. Report it if one validator can affect other validators, the validator set, or other users' funds.

## Invariants

These must always hold. Verify them in every audit.

1. Only the authority can add or remove another validator. `x/poa` handlers check `msg.Authority == k.authority` before any validator set change. `MsgSelfRemoveValidator` can remove only the signer's own validator.
2. No account can change the validator set or bonded stake by any path except `x/poa`. Check every path that changes stake or the validator set.
3. Every transaction goes through the ante chain for its route. No path skips the signature, fee and sequence checks of that route. Unknown extension options are rejected.
4. Tokens are not created or destroyed except by the intended mint and burn paths. Watch the `x/poa` add and remove flows, the bond pool routing (Bonded vs NotBonded), the EVM module (Minter and Burner), erc20 conversion, and IBC transfer escrow.
5. The state machine is deterministic. No map iteration that writes state, no goroutines, no `time.Now()`, no floats in consensus code.
6. Upgrade handlers run once, at their height, and change only the state they are meant to change.
7. The rate limiting middleware stays in the IBC transfer stack. ICA host is removed in v12: the IBC router does not route the `icahost` port on any network.

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

**Entry**: `MsgAddValidator`, `MsgRemoveValidator`, `MsgSelfRemoveValidator` transactions

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
- Precompiles that call Cosmos msg servers directly and skip the POA ante handler, for example the staking precompile (`0x...0800`)

**Key files**: Upstream `cosmos/evm` module (out of scope for deep audit, but check integration points in `app/app.go`)

**What to check**:
- EVM module is wired with correct ante handlers
- Custom precompiles (if any) are registered correctly
- The staking precompile is not registered (`app/app.go`). Activating its address in the EVM params must only make calls to it fail, and no other registered precompile may create validators or change delegations
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

**Key files**: `go.mod`, `go.sum`, `Dockerfile`, `.oss-scanner/Dockerfile`, `Makefile`, `.goreleaser.yml`, `.github/workflows/`, `.github/CODEOWNERS`

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

**Key files**: `app/upgrades/v<version>/` and upgrade registration in `app/upgrades.go`. The repository keeps only the current upgrade. The release tags contain the earlier handlers.

**What to check**:
- Each upgrade handler is registered exactly once with the correct version name
- Migration logic is deterministic (no maps iterated for state, no goroutines, no time-dependent logic)
- Upgrade handlers only modify state relevant to their version
- No upgrade handler can be re-executed after completion
- Keeper references in upgrade handlers (`app/upgrades/v<version>/keepers.go`, when present) only expose necessary keepers

### 8. Governance

**Entry**: On-chain governance proposals

**Threats**:
- Governance proposal executing arbitrary module messages (parameter changes, software upgrades, community spend)
- Parameter changes that put the chain in an unsafe state (e.g., unbonding time to 0, max validators to 0)
- Delegation of gov authority to an account, for example through a reintroduced `authz` module (removed in v12)
- POA module exposes `GovKeeper.SubmitProposal` — verify it cannot be abused to self-authorize

**Key files**: Gov module wiring in `app/app.go`, `x/poa/types/expected_keepers.go` (GovKeeper interface)

**What to check**:
- Gov module message allowlist is configured (if applicable) to restrict dangerous message types
- Parameter boundaries prevent unsafe values
- `authz` has no keeper or message handler: `MsgGrant` and `MsgExec` decode only so that historical txs stay readable (`app/encoding.go`). They must fail at execution, and gov must reject them at proposal submission
- POA's GovKeeper interface is minimal and cannot be used to submit arbitrary proposals

## Cosmos SDK-Specific Concerns

These are common vulnerability patterns in Cosmos SDK chains that should be checked each audit:

- **Module account permissions**: Check that module accounts have only the permissions they need. POA module has `Minter` and `Burner` — verify no other module has unexpected mint/burn access. Verify `BondedPoolName` and `NotBondedPoolName` are the only accounts used for staking token burns.
- **Keeper exposure**: Verify keepers are not passed beyond their intended module boundaries. Check `app/app.go` for keepers passed to modules that shouldn't have them. POA receives BankKeeper and StakingKeeper — verify these are the correct scoped interfaces (`x/poa/types/expected_keepers.go`), not the full keeper implementations.
- **Parameter changes**: Check if on-chain parameter changes can lead to unsafe states (e.g., `MaxValidators` to 0 in staking params, unbonding time to 0, bond denom change breaking POA).
- **Non-determinism**: Any use of maps iterated for state, goroutines, or `time.Now()` in state machine code causes consensus failures. Check all code paths in `x/poa/`, `app/ante/`, and `app/upgrades/`.
- **ICA host**: removed in v12. The IBC router must not route the `icahost` port, and gov must reject `/ibc.applications.interchain_accounts.host.v1.MsgUpdateParams` at proposal submission. ICS-27 types decode only so that historical txs and existing interchain accounts stay readable (`app/encoding.go`). Reintroducing ICA host, or any IBC app that executes messages, lets remote chains call x/staking and bypass the POA ante handler.

## Out of Scope

- Bugs inside dependencies: cosmos-sdk, cometbft, ibc-go, ibc-apps, go-ethereum, and cosmos/evm. `go.mod` replaces cosmos-sdk with `github.com/xrplevm/cosmos-sdk` and cosmos/evm with `github.com/xrplevm/evm`. These forks are separate repositories. A bug in them is in scope only when code in this repository causes it or makes it reachable, for example through module wiring, configuration, params, or a missing guard in our ante chain.
- Known CVEs in dependency versions, unless this repository's configuration makes them exploitable. Exception: the internal "Dependency audit" focus area in `SECURITY_SCOPE.md` still checks known CVEs, version pinning, and places where we override or patch upstream behavior (including the `xrplevm/*` forks).
- Smart contracts that users deploy on the chain.
- `docs/` and Markdown files.
- `cmd/exrpd/cmd/in-place-testnet.go` input. Only an operator runs it, on a local copy of state.
- Live CI runs and GitHub repository settings. Review `.github/workflows/`, `.github/CODEOWNERS`, `.goreleaser.yml`, Dockerfiles and the `Makefile` as static files only.

## How to Exercise the Code

In the OSS Scanner image (`.oss-scanner/Dockerfile`) there is no network. All Go modules, `golangci-lint` and `mockgen` are installed, and all test binaries are already compiled.

- Build: `make build` -> `bin/exrpd`. The scanner image builds it without stripping.
- PoA unit tests: `make test-poa`. Ante decorator tests: `go test ./x/poa/ante/`.
- Integration tests (in-process multi-validator network): `make test-integration`. The suite is in `tests/integration/`. Put a proof of concept here when it needs real blocks, staking and slashing.
- Simulation (slow, 100 blocks): `make test-sim-full-app-fast`.
- Local single-node chain: `./local-node.sh` (runs `make build`, uses `bin/exrpd`, stores state in `.exrpd/`).
- Lint: `golangci-lint run --timeout=10m`.

## Severity

- **Critical**: an unprivileged account can steal or mint funds, change the validator set, halt the chain, or cause a consensus failure (state divergence between honest nodes).
- **High**: the same impact, but it needs a precondition that a normal user can reach (a specific param value, a particular IBC counterparty, a single malicious validator). Also: a permanent lock of user funds, or a bypass of the IBC rate limit.
- **Medium**: crash or resource exhaustion of one node through a public endpoint, without a chain-wide effect. Fee or gas accounting errors with small, bounded loss. CI or release pipeline weaknesses that need a second step to exploit.
- **Low**: defense in depth, unsafe defaults that operators normally override, unclear error handling with no shown impact.
- **Info**: an observation with no direct security impact.

If a finding needs governance or the authority to act maliciously, it is at most Low, unless it gives more power than the authority already has.

## Reports from External Scanners

The reporting workflow in `SECURITY_SCOPE.md` is for internal audits only. External scanners follow these rules:

- Give one root cause per report. Merge findings with the same root cause into one report, even if they reach it through different message types.
- Include the exact file and function, the attacker's starting position (which trust level), and the shortest call path from an entry point.
- A proof of concept is most useful as a Go test in the package of the bug, or in `tests/integration/`, that fails before the fix.
- Patches: keep them minimal. Fix the root cause in the owning package and add a regression test. Do not reformat unrelated code.

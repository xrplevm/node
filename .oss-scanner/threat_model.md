# Threat model: XRPL EVM Sidechain node (`exrpd`)

## What this project does

`exrpd` is the node software of the XRPL EVM Sidechain. It is a Cosmos SDK chain with an EVM (`cosmos/evm`). Validators are managed by Proof of Authority (`x/poa`), not by open staking. The chain holds real value: the native `axrp` token, ERC-20 tokens, and IBC assets.

Main parts:

- `app/app.go`: wires the modules. Standard SDK modules (auth, bank, staking, slashing, distribution, gov, authz, feegrant, upgrade, evidence, consensus, params, crisis), IBC (transfer, ICA host, rate limiting), and cosmos/evm (`vm`, `feemarket`, `erc20`, static precompiles).
- `app/ante/`: the ante handler. It routes each transaction by its first extension option:
  - `/cosmos.evm.vm.v1.ExtensionOptionsEthereumTx` -> EVM ante chain (`evm_handler.go`)
  - `/cosmos.evm.ante.v1.ExtensionOptionDynamicFeeTx` -> Cosmos ante chain (`cosmos_handler.go`)
  - any other extension option -> rejected
  - no extension option -> Cosmos ante chain
- `x/poa/`: our own module. It adds and removes validators.
  - `MsgAddValidator` and `MsgRemoveValidator`: the signer must be the module authority (the gov module account). The add flow mints the bond tokens to the validator. The remove flow slashes 100%, unbonds, and burns.
  - `MsgSelfRemoveValidator`: a validator removes itself. The signer is the validator account.
  - `x/poa/ante/poa.go`: blocks `MsgDelegate`, `MsgUndelegate`, `MsgBeginRedelegate`, `MsgCancelUnbondingDelegation` and `MsgCreateValidator` after height 0. Users must not change the validator set or stake through staking messages.
- `app/upgrades/`: state migrations for each chain upgrade (`v5.0.0` to `v11.2.0`). `v112` changes staking params, disables the ICA host, and moves escrowed IBC funds.
- `cmd/exrpd/`: CLI, node config defaults, and `in-place-testnet`.

## Where untrusted input enters

Treat all of these as attacker-controlled:

- Any transaction from any account, Cosmos or Ethereum format, sent through CometBFT RPC, gRPC, REST, or Ethereum JSON-RPC. This includes `authz.MsgExec` with nested messages and EVM calls to precompiles.
- EVM bytecode and calldata, including contracts that call the static precompiles (for example staking, distribution, bank, erc20, ics20, gov, slashing).
- IBC packets and acknowledgements from a counterparty chain, which may be malicious.
- Queries to the public gRPC, REST and JSON-RPC endpoints.

Trusted (do not report attacks that need these without another bug):

- The authority (gov module account) and anyone who can pass a governance proposal.
- A node operator's own config files, CLI flags, keys and genesis file.
- 2/3+ of the validators acting together.

A single validator is trusted only for its own node. Report it if one validator can affect other validators, the validator set, or other users' funds.

## Invariants that matter most

1. Only the authority can add or remove another validator. `MsgSelfRemoveValidator` can remove only the signer's own validator.
2. No account can change the validator set or bonded stake by any path except `x/poa`. Check every path that changes stake or the validator set.
3. Every transaction goes through the ante chain for its route. No path skips the signature, fee and sequence checks of that route.
4. Tokens are not created or destroyed except by the intended mint and burn paths. Watch the `x/poa` add and remove flows, the bond pool routing (Bonded vs NotBonded), the EVM module (Minter and Burner), erc20 conversion, and IBC transfer escrow.
5. The state machine is deterministic. No map iteration that writes state, no goroutines, no wall-clock time, no floats in consensus code.
6. Upgrade handlers run once, at their height, and change only the state they are meant to change.
7. The rate limiting middleware stays in the IBC transfer stack. The ICA host stays disabled on chains that ran `v11.2.0`, and its default params on new chains must not allow arbitrary messages.

## Components by priority

| Priority | Paths |
|---|---|
| Critical | `x/poa/`, `app/ante/`, `app/app.go`, `app/upgrades/` |
| High | `app/encoding.go`, `app/config.go`, `app/genesis.go`, `app/export.go`, `types/`, `proto/`, `cmd/exrpd/` |
| Medium | `.github/workflows/`, `.github/CODEOWNERS`, `.goreleaser.yml`, `Dockerfile`, `Makefile`. Review them as static files: workflow injection, unsafe `pull_request_target`, too-wide token permissions, unpinned actions or images, secrets in commands, release integrity. Live CI runs and GitHub repo settings are not visible in this image. Do not report findings that depend only on them. |
| Low | `testutil/`, `tests/`, `x/poa/testutil/`, `scripts/`, `local-node.sh`. Report only when a test helper hides a real bug in production code. |

## Out of scope

- Bugs inside dependencies: cosmos-sdk, cometbft, ibc-go, ibc-apps, go-ethereum, and cosmos/evm. `go.mod` replaces cosmos-sdk with `github.com/xrplevm/cosmos-sdk` and cosmos/evm with `github.com/xrplevm/evm`. These forks are separate repositories and are not the target of this scan. A bug in them is in scope only when code in this repository causes it or makes it reachable, for example through module wiring, configuration, params, or a missing guard in our ante chain.
- Known CVEs in dependency versions, unless this repository's configuration makes them exploitable.
- Smart contracts that users deploy on the chain.
- `docs/`, `.claude/`, and Markdown files.
- `cmd/exrpd/cmd/in-place-testnet.go` input. Only an operator runs it, on a local copy of state.

## How to exercise the code

The image is offline. All Go modules, `golangci-lint` and `mockgen` are installed. All test binaries are already compiled.

- Build: `make build` -> `/src/bin/exrpd` (not stripped).
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

If a finding needs governance or the authority to act maliciously, it is at most Low, unless it gives more power than the authority already has.

## Reports

- Give one root cause per report. Merge findings with the same root cause into one report, even if they reach it through different message types.
- Include the exact file and function, the attacker's starting position (which trust level), and the shortest call path from an entry point.
- A proof of concept is most useful as a Go test in the package of the bug, or in `tests/integration/`, that fails before the fix.
- Patches: keep them minimal. Fix the root cause in the owning package and add a regression test. Do not reformat unrelated code.

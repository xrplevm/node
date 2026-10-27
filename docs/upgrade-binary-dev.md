# exrpd upgrade binary: developer notes

Operator details, internals, build, CI and release process of `exrpd-upgrade`. For the
short install and run steps, see the [operator guide](upgrade-binary.md).

| Path | Content |
| --- | --- |
| `cmd/exrpd-upgrade/` | the launcher, a separate Go module |
| `scripts/upgrade-binary/plan.sh` | prints the plan: the `UpgradeName` of the one handler in `app/upgrades` |
| `scripts/upgrade-binary/genesis.sh` | prints the default genesis: the newest published release below the plan |
| `scripts/upgrade-binary/cosmovisor/` | Cosmovisor pins (`cosmovisor.env`) and the patch |
| `scripts/upgrade-binary/build.sh` | builds one `exrpd-upgrade` |
| `scripts/upgrade-binary/cosmovisor.sh` | builds or tests the patched Cosmovisor |
| `scripts/upgrade-binary/release.sh` | builds the release assets from published archives |
| `scripts/upgrade-binary/e2e.sh` | localnet chain upgrade on linux/amd64 |
| `.github/workflows/upgrade-binary.yml` | release workflow |

## Operator details

### Standalone binary or upgrade binary

| | Standalone `exrpd` | Upgrade binary `exrpd-upgrade` |
| --- | --- | --- |
| Release asset | `node_<version>_<Os>_<arch>.tar.gz` (`.zip` on Windows) | `exrpd-upgrade_<version>_<Os>_<arch>`, the executable without an archive |
| Checksums | `checksums.txt` | `exrpd-upgrade_<version>_checksums.txt`. The release also has `exrpd-upgrade_<version>_LICENSE` |
| Contents | one chain version | 11.3.0 + 12.0.0 + Cosmovisor |
| At the upgrade height | halts, then you swap the binary and restart | switches to 12.0.0 by itself |
| Platforms | linux/darwin amd64/arm64, windows/amd64 | linux/darwin amd64/arm64 |

Both use the same node home, keys and config. You can move between them at any time
while the node is stopped. The binary you start must match the chain height: 11.3.0
before the upgrade height, 12.0.0 after. If you move to the upgrade binary on data that
you restored or state-synced past the upgrade height, run `stage --select v12.0.0` first.

### Commands

- Home: `--home`/`--home=`, else `$DAEMON_HOME`, else `$HOME/.exrpd`. `start` passes the
  same home to `exrpd start`.
- `start` passes every other argument to `exrpd start` unchanged. It refuses
  `--cosmovisor-config`, because Cosmovisor would then read its settings from that file
  and ignore the environment that `start` sets.
- `start` checks the environment, stages the files, then replaces itself with Cosmovisor
  (same PID). Cosmovisor runs `exrpd start`. Under systemd, `MainPID` is Cosmovisor.
- `stage` stages the files and prints the paths without starting the node. You can run it
  right after install.
- `info` prints the manifest: versions, plan, SHA-256 and Cosmovisor source.
- Cosmovisor reads the chain height with `exrpd status --home <home>`, without `--node`.
  This is why the CometBFT RPC server (`[rpc] laddr` in `config/config.toml`) must stay on
  and `node` in `<home>/config/client.toml` must point at it.
- If the node halts at the upgrade height and Cosmovisor does not switch, a restart works
  because `start` sees the plan in `data/upgrade-info.json` and points `current` at the
  upgrade binary.
- Use `<home>/cosmovisor/current/bin/exrpd` for CLI commands so the client always matches
  the running chain version.

### macOS (launchd)

Use a LaunchDaemon with `ProgramArguments` = `/usr/local/bin/exrpd-upgrade`, `start`,
`--home`, `<dir>` and a `KeepAlive` dictionary with `SuccessfulExit` = `false`. Give the
absolute path, because launchd does not search `/usr/local/bin`. launchd sends SIGTERM to
the job's PID (Cosmovisor), which forwards it to `exrpd`. When the job exits, launchd
kills the remaining processes of its process group unless `AbandonProcessGroup` is
`true`. Keep the default `false`.

### Without a service manager

An `exrpd` can outlive Cosmovisor when Cosmovisor dies without forwarding a stop:
SIGKILL, SIGHUP or another signal it does not handle, an out-of-memory kill, or an
upstream panic while parsing a malformed non-empty `data/upgrade-info.json`. Under systemd
with the default `KillMode=control-group`, and under launchd, the service manager stops
the rest of the unit when Cosmovisor exits, so nothing keeps running. Without a service
manager, start the node in its own session and stop it by process group:

```sh
setsid exrpd-upgrade start --home <home> &
ps -o pgid= -p <cosmovisor pid>          # find the process group
kill -TERM -- -<pgid>                    # stop the node
```

Do not signal the group of a node that you did not start with `setsid`. That group can
contain your shell. Check for a leftover `exrpd` before you start again. A second node on
the same home fails on the database lock instead of double signing.

### Staged files

All staged files are `0755` and written atomically. On every `start` and `stage`:

- A file with the same SHA-256 is left alone. If its owner cannot execute it (for
  example, a copy restored with mode `0644`), its mode is set to `0755` again. So a
  reinstall of the same `exrpd-upgrade` restages identical files.
- A file with different content is replaced atomically. `exrpd-upgrade` writes a line to
  stderr that names the file and both SHA-256 values. A patch release bundle replaces the
  binaries of an earlier bundle this way.
- An older `exrpd-upgrade` started on the home (for example, one still named in a service
  file) puts its own binaries back. Check the service log for `exrpd-upgrade: replaced`.
- A home that stock Cosmovisor or an earlier bundle set up is taken over. Its `genesis`
  and `upgrades/v12.0.0` binaries are replaced, and `current` moves by the rule in
  [How the launcher chooses `current`](#how-the-launcher-chooses-current). Other
  directories under `upgrades/` stay on disk, but `current` does not point at them after
  the takeover.

### Restart and restore details

- Killed at the halt before Cosmovisor switched (power loss, SIGKILL, OOM), or
  `<home>/cosmovisor` deleted after the upgrade on a home whose `data/upgrade-info.json`
  names `v12.0.0`: the next start runs 12.0.0. Cosmovisor's backup,
  `COSMOVISOR_CUSTOM_PREUPGRADE` and the upgrade's optional `exrpd pre-upgrade` hook are
  not run in this case (12.0.0 does not implement the hook).
- `<home>/cosmovisor` deleted on a home past the upgrade height without that file (state
  sync, or a restore that used `stage --select v12.0.0`): only `current` recorded the
  selection. Run `stage --select v12.0.0` again before you start. Otherwise Cosmovisor
  links `current` to `genesis` and 11.3.0 fails on every start.
- Restored data usually has no `data/upgrade-info.json` for the plan it is past. So you
  select the binary for the height of the data yourself, with the node stopped, as the
  node user.
- `--select` accepts only `genesis` and `v12.0.0`.
- `--select v12.0.0` deletes a `data/upgrade-info.json` that names another plan, such as
  `v11.3.0` kept in the home from before the restore, and logs it. Cosmovisor would
  otherwise try to switch back to that plan's binary.
- A `data/upgrade-info.json` that names `v12.0.0` wins over `--select genesis`. If you
  restored data from before the upgrade height and kept that file from the newer data,
  delete it first.
- If you start without selecting, data past the upgrade height on `genesis` exits with a
  store error such as `version of store icahost mismatch` on every start.
- A `data/upgrade-info.json` that is not valid JSON, for example truncated by a crash,
  makes `start` and `stage` stop with an error before they change any file.
- Run the bundle for the chain's next plan before its height. A bundle started on data
  past a later plan points `current` at its genesis binary, which cannot run that data and
  exits.

### Settings details

- An empty value counts as unset.
- Cosmovisor refuses a `DAEMON_SHUTDOWN_GRACE` of `0`, so the upstream default (kill
  exrpd at once) is not available.
- All other Cosmovisor variables keep their upstream meaning, notably
  `DAEMON_POLL_INTERVAL`, `DAEMON_RESTART_DELAY`, `COSMOVISOR_COLOR_LOGS`,
  `COSMOVISOR_TIMEFORMAT_LOGS`.
- For batch upgrades, a failed `COSMOVISOR_CUSTOM_PREUPGRADE` and a stop during the grace
  wait, see [Operational caveats](#operational-caveats).

### Cosmovisor backup

With the default `UNSAFE_SKIP_BACKUP=true`, nothing copies `data/`. The upgrade migrates
the stores at the upgrade height, and 11.3.0 cannot open them after that. Pointing
`current` back at `genesis` or installing another binary does not restore the data.

To keep the upstream backup, set `UNSAFE_SKIP_BACKUP=false`, optionally with
`DAEMON_DATA_BACKUP_DIR`:

- The copy needs free disk equal to the size of `data/`.
- The validator does not sign until the copy finishes.
- Do not stop the node during the copy. Cosmovisor exits at once and leaves a partial
  `data-backup-<date>` directory.
- After a stop, a failed copy or a failed `COSMOVISOR_CUSTOM_PREUPGRADE`, the next
  `start` points `current` at 12.0.0 itself. Cosmovisor does not run the backup or
  `COSMOVISOR_CUSTOM_PREUPGRADE` again. Delete the partial backup and take your own
  snapshot before you start again if you need one.

If you restore a backup, keep the newer `data/priv_validator_state.json`. An older copy
lets the validator sign heights it already signed, which is double signing.

## Payload

- The genesis `exrpd` is the exact bytes of the GitHub release `v11.3.0` archive. The build
  checks it against that release's `checksums.txt`.
- The upgrade `exrpd` is the same bytes as the standalone 12.0.0 release binary.
- Cosmovisor is built from `cosmossdk.io/tools/cosmovisor@v1.7.3` plus one patch. See
  [Cosmovisor patch](#cosmovisor-patch).

The three executables are stored gzip-compressed. This is why the binary is about 100 MB
on linux/amd64 although it carries two full `exrpd` builds and Cosmovisor. The SHA-256
and size in the manifest are those of the uncompressed executables, and staging checks
them. Writing the files never changes `current`. The launcher sets it in a separate step,
see the next section.

## How the launcher chooses `current`

At startup, Cosmovisor compares the plan in `<home>/data/upgrade-info.json` with the plan
recorded in `current/upgrade-info.json`. It switches when they differ. x/upgrade writes the
data file when the chain reaches a plan height, so it names the last plan the chain
reached. Before Cosmovisor runs, `exrpd-upgrade` points `current` at the embedded binary
for that plan and records the plan there:

| `data/upgrade-info.json` | `current` |
| --- | --- |
| names `v12.0.0` | `upgrades/v12.0.0` |
| names any other plan (`v11.3.0` on homes that ran that upgrade) | `genesis`. Without the record, Cosmovisor would try to switch to `upgrades/v11.3.0` and stop, because it cannot download that binary |
| missing | `genesis` and `upgrades/v12.0.0` stay. Any other `current` moves to `genesis`, and the plan recorded there moves with it, like the other-plan case. A missing `current` stays missing, and Cosmovisor links it to `genesis` |
| not valid JSON | `start` and `stage` stop with an error before they change any file |

`stage --select` is for restored data that has no data file for the plan it is past.
`--select v12.0.0` deletes a data file that names another plan, because Cosmovisor would
switch back to it. A data file that names `v12.0.0` wins over `--select genesis`.

After a `--select v12.0.0` without a data file, only `current` records the selection. If
`<home>/cosmovisor` is deleted, Cosmovisor links `current` to `genesis` again.

Each move of `current` writes one line to stderr, so the service log shows why the node
runs a binary. For example:
`exrpd-upgrade: moved <home>/cosmovisor/current from upgrades/v12.0.0 to genesis: data/upgrade-info.json names v12.1.0`.

## Operational caveats

These follow from the rule above and from Cosmovisor itself:

- **Stop during the grace wait at the upgrade height.** Cosmovisor waits up to about 1 s
  before it forwards the stop. If exrpd has exited by then, Cosmovisor exits 0 and does not
  start 12.0.0. Otherwise Cosmovisor starts 12.0.0 and the service manager kills the unit
  at its stop timeout (`TimeoutStopSec`). In both cases, the next start runs 12.0.0.
- **Failed `COSMOVISOR_CUSTOM_PREUPGRADE`.** Upstream, it blocks the upgrade. Here the next
  `start` sees the data file that names `v12.0.0` and selects 12.0.0 without running the
  hook. The same applies after a stop or a failed copy during the Cosmovisor backup.
- **Batch upgrades.** Cosmovisor batch upgrades (`data/upgrade-info.json.batch`) and
  `cosmovisor add-upgrade --upgrade-height` both write `data/upgrade-info.json` before the
  plan height. `start` reads a file that names `v12.0.0` as an upgrade the chain already
  reached, so 12.0.0 would start too early. The governance upgrade needs neither.

## Cosmovisor patch

Cosmovisor v1.7.3 registers a SIGTERM/SIGQUIT handler on every run and never removes it
after an upgrade. After the in-process switch, a SIGTERM sent only to Cosmovisor's PID
also reaches the stale handler, which cannot signal the old app and calls `os.Exit(1)`.
Cosmovisor then exits while 12.0.0 keeps running outside the service manager's view.

`scripts/upgrade-binary/cosmovisor/cosmovisor-v1.7.3-signal.patch` is the fix proposed upstream in
[cosmos/cosmos-sdk#26879](https://github.com/cosmos/cosmos-sdk/pull/26879):

- It changes only `Launcher.Run`. The handler is registered before the app starts and is
  stopped on every return path. `TestLaunchProcessSignalAfterUpgrade` fails on unpatched
  v1.7.3.
- A second commit in the same PR, `4513d133`, also exits 0 when a stop arrives after the
  app has already exited, as with the default systemd stop. It still exits at once during
  the backup or the pre-upgrade, instead of swallowing the stop and relaunching the
  upgraded app. `TestLaunchProcessStop` covers both.

Remove the patch when a Cosmovisor release contains the fix.

The build fetches the module at v1.7.3 and checks its `go.sum` hash (from `cosmovisor/cosmovisor.env`). It
runs `go mod verify`, so the extracted module source and its dependencies match their
hashes. It applies the patch with `git apply --check` first and builds with Cosmovisor's
own `go.mod` (`CGO_ENABLED=0`).

## Building

```sh
mkdir -p release12 && tar -xzf node_12.0.0_Linux_amd64.tar.gz -C release12 bin/exrpd
scripts/upgrade-binary/build.sh \
  --genesis-version 11.3.0 --genesis-archive node_11.3.0_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
  --upgrade-exrpd release12/bin/exrpd --upgrade-version 12.0.0 \
  --goos linux --goarch amd64 --out dist/exrpd-upgrade
scripts/upgrade-binary/cosmovisor.sh --test # patched Cosmovisor test suite
CACHE=<dir> GENESIS_VERSION=11.3.0 scripts/upgrade-binary/e2e.sh   # localnet proof on linux/amd64
CACHE=<dir> GENESIS_VERSION=11.3.0 scripts/upgrade-binary/e2e.sh --unpatched   # same run with the unpatched v1.7.3 release binary, must show the defect
```

`--upgrade-exrpd` is the `bin/exrpd` of the standalone 12.0.0 release archive, not the
`bin/exrpd` that `make build` writes. A local build has other bytes. A home that stages it
logs a replacement whenever it moves between that build and a released bundle.

### End-to-end test

`e2e.sh` does not download release assets. Put these files in a directory and pass it as
`CACHE`:

| File | Content |
| --- | --- |
| `node_<GENESIS_VERSION>_Linux_amd64.tar.gz`, `checksums.txt` | assets of the genesis GitHub release, for example `v11.3.0` |
| `exrpd-<version>-linux-amd64` | standalone upgrade `exrpd` for linux/amd64, executable. `<version>` is `UPGRADE_VERSION`, by default the plan version (12.0.0) |
| `cosmovisor-v1.7.3-linux-amd64.tar.gz`, `SHA256SUMS-cosmovisor-v1.7.3.txt` | `--unpatched` only: the upstream Cosmovisor v1.7.3 release archive and a checksum list that contains its line |

| Input | Effect |
| --- | --- |
| `GENESIS_VERSION=11.3.0` | required: the genesis release in `CACHE`. `genesis.sh <plan>` prints the default |
| `UPGRADE_VERSION=12.0.1` | tests a patch bundle |
| `EXRPD_UPGRADE=<path>` | tests an `exrpd-upgrade` that was already built, for example a release asset. The run checks that its manifest names the plan and the SHA-256 of both `exrpd` files in `CACHE`, and does not build one |
| `E2E_DIR=<dir>` | scratch directory |

- The script stops with an error that names the first missing file.
- The build still needs Go modules. Without `--unpatched`, `cosmovisor.sh` runs
  `go mod download` for the pinned Cosmovisor module and its dependencies. The launcher
  module uses only the standard library. Run it with network access or with a Go module
  cache that already holds these modules.
- Two runs cannot use the same `E2E_DIR` at the same time. The script holds the lock
  `<E2E_DIR>.lock` and stops when another run holds it.
- The script deletes `E2E_DIR` at the start. It accepts only a directory that is missing,
  empty, or marked with `.e2e-scratch` by an earlier run.
- With `--unpatched`, the run passes only when the SIGTERM after the in-process switch
  shows the defect: an orphaned `exrpd`, the stale handler's `process already finished`
  log, or Cosmovisor exit status 1. With the patched build, every stop must show none of
  these. Which signal handler runs first is not fixed, so a run often misses it. Run it
  again in that case. `TestLaunchProcessSignalAfterUpgrade` is the deterministic check.

## CI

Pull requests (`.github/workflows/pull-request.yml`):

| Job | Runs |
| --- | --- |
| `upgrade-binary` | every pull request and every push to `main`: launcher `go vet`, `go test` and `golangci-lint` in `cmd/exrpd-upgrade`, then `cosmovisor.sh --test` |
| `upgrade-binary-e2e` | pull requests that change `cmd/exrpd-upgrade/`, `scripts/upgrade-binary/` or `app/upgrades/`: `e2e.sh` with an `exrpd` built from the pull request (`make build VERSION=<plan version>`) and the genesis release that `genesis.sh` picks |

Releases (`.github/workflows/upgrade-binary.yml`) run after GoReleaser:

1. Read the plan from the tagged code (`plan.sh`). Take the genesis from the `genesis_tag`
   input, or from `genesis.sh` when it is empty.
2. Run the launcher unit tests and `cosmovisor.sh --test`.
3. Download the published standalone archives and the genesis release archives.
4. Run `scripts/upgrade-binary/release.sh`.
5. Run `e2e.sh` with the linux/amd64 asset (`EXRPD_UPGRADE`).
6. Upload the `exrpd-upgrade_*` assets, only when every check passes.

The standalone release is already published at that point. The workflow checks out the
tag, so what to do after a failure depends on the cause:

| Cause | Recovery |
| --- | --- |
| transient failure (network, runner) | run the workflow again for the same tag |
| upload stopped part way | delete the `exrpd-upgrade_*` assets of that release, then run the workflow again. It never replaces an asset that is already there |
| defect in the tagged code or a missing upgrade handler | a rerun builds the same commit and fails again. Fix it on the release branch and publish a new release (for example the next patch), which gets its own standalone and upgrade binary assets |
| wrong genesis | the `genesis_tag` input or the default picked the wrong release. If the run failed before the upload, run the workflow again with the right `genesis_tag`. After an upload, delete the `exrpd-upgrade_*` assets first |

## Releases and bundles

Every release, pre-releases included, gets an upgrade binary. A bundle has three inputs:

| Input | Source |
| --- | --- |
| Plan | `UpgradeName` of the one handler in `app/upgrades` of the tagged code (`plan.sh`) |
| Genesis | `genesis_tag` input of the release workflow. Default: the newest published release, drafts and pre-releases excluded, whose major.minor is below the plan's (`genesis.sh`) |
| Cosmovisor | `scripts/upgrade-binary/cosmovisor/cosmovisor.env` and the patch next to it |

The release version must have the major.minor of the plan (today 12.0). `build.sh` and
`release.sh` stop with an error otherwise, so a major or minor release without its
upgrade handler fails the release workflow. Tags without `plan.sh` (releases older than
the upgrade binary) ship standalone only. The manifest records the plan and the genesis
version and SHA-256 (`exrpd-upgrade info`).

| Release | Chain | Upgrade binary |
| --- | --- | --- |
| Major or minor (`X.0.0`, `X.Y.0`) | changes state or consensus, needs a governance plan `vX.Y.0` with an upgrade handler | new bundle: the genesis is the patch release validators run at the halt, the plan is the new handler |
| Patch (`X.Y.Z`, `Z > 0`) | no new upgrade handler, no state or consensus change: a drop-in replacement | bundle with the same plan and the patch as the upgrade binary |

A patch bundle reaches validators with a reinstall and a restart, before or after the
switch: `start` replaces `upgrades/v12.0.0/bin/exrpd` and logs it. The standalone patch
binary stays an alternative: stop the service and run `exrpd start --home <home>`.
Nothing checks that a patch is a drop-in replacement. A patch that changes state or
consensus must be a minor release with its own plan.

**Patch of the genesis line.** While an upgrade bundle is out but before its halt (for
example 12.0.3 while the v12.1.0 bundle carries genesis 12.0.2):

1. Release the standalone 12.0.3.
2. Release a patch of the upgrade line (12.1.1). Its default genesis is now 12.0.3, or pass
   `genesis_tag: v12.0.3`. Its `start` replaces `genesis/bin/exrpd`.

On the 12.0 branch, the handler is still `v12.0.0`, so the 12.0.3 tag also publishes
a v12.0.0 bundle for homes that still run that one.

**Next upgrade bundle.** The next bundle, for example v12.1.0, moves `current` to its
genesis binary (12.0.x) when `data/upgrade-info.json` names `v12.0.0`. Without that file, a
`current` that points to the previous bundle's upgrade directory also moves to genesis.
The bundle copies the plan record of that directory, if there is one, to
`genesis/upgrade-info.json`. No manual step is needed for either case. Data restored past
the next bundle's plan still needs `stage --select <plan>` once.

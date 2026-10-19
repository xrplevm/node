# exrpd upgrade binary

`exrpd-upgrade` is a single executable for validators who want the node to switch to the
next chain version on its own at the governance upgrade height. It contains:

| Payload | Origin |
| --- | --- |
| genesis `exrpd` 11.2.0 | exact bytes of the GitHub release `v11.2.0` archive, checked against that release's `checksums.txt` at build time and by SHA-256 when staged |
| upgrade `exrpd` 12.0.0 | the same bytes as the standalone 12.0.0 release binary |
| Cosmovisor v1.7.3 | built from `cosmossdk.io/tools/cosmovisor@v1.7.3` plus one patch, see [Cosmovisor patch](#cosmovisor-patch) |

The three executables are stored gzip-compressed. The SHA-256 and size in the manifest are
those of the uncompressed executables, and staging checks them. A patch release of 12.0
(for example 12.0.1) ships its own `exrpd-upgrade` with the same plan, see
[Releases and bundles](#releases-and-bundles).

`exrpd-upgrade start` writes these files into the standard Cosmovisor layout and then
replaces itself with Cosmovisor (same PID). Cosmovisor runs `exrpd start`. When the chain
halts at the `v12.0.0` plan height, it stops 11.2.0, points `current` at the 12.0.0 binary
and starts it again in the same process. Nothing is downloaded.

## Standalone binary or upgrade binary

| | Standalone `exrpd` | Upgrade binary `exrpd-upgrade` |
| --- | --- | --- |
| Release asset | `node_<version>_<Os>_<arch>.tar.gz` (`.zip` on Windows) | `node_upgrade_<version>_<Os>_<arch>.tar.gz` |
| Checksums | `checksums.txt` | `node_upgrade_<version>_checksums.txt` |
| Contents | one chain version | 11.2.0 + 12.0.0 + Cosmovisor |
| At the upgrade height | halts, then you swap the binary and restart | switches to 12.0.0 by itself |
| Platforms | linux/darwin amd64/arm64, windows/amd64 | linux/darwin amd64/arm64 |

Both use the same node home, keys and config. You can move between them at any time
while the node is stopped, as long as the binary you start matches the chain height
(11.2.0 before the upgrade height, 12.0.0 after).

## Install

```sh
os=$(uname -s)                     # Linux or Darwin
arch=$(uname -m)
[ "$arch" = x86_64 ] && arch=amd64
[ "$arch" = aarch64 ] && arch=arm64
archive=node_upgrade_12.0.0_${os}_${arch}.tar.gz
sha256=sha256sum
[ "$os" = Darwin ] && sha256='shasum -a 256'
grep " $archive\$" node_upgrade_12.0.0_checksums.txt | $sha256 -c - &&
  tar -xzf "$archive" bin/exrpd-upgrade &&
  sudo install -m 0755 bin/exrpd-upgrade /usr/local/bin/exrpd-upgrade
exrpd-upgrade info                      # manifest: versions, plan, SHA-256, Cosmovisor source
exrpd-upgrade stage --home ~/.exrpd     # optional: stage now and print the paths
```

Download the archive for your platform and `node_upgrade_12.0.0_checksums.txt` into the
same directory. Verify the archive before you extract it. The `&&` chain stops at a
checksum mismatch or when the checksum file has no line for the archive, so nothing is
extracted or installed.

## Commands

```text
exrpd-upgrade start [--home DIR] [exrpd start flags...]
exrpd-upgrade stage [--home DIR] [--select genesis|v12.0.0]
exrpd-upgrade info
```

The home is `--home`/`--home=`, else `$DAEMON_HOME`, else `$HOME/.exrpd`. The same home is
passed to `exrpd start`. Every other argument after `start` is passed through unchanged,
except `--cosmovisor-config`: `start` refuses it, because Cosmovisor would then read its
settings from that file and ignore the [environment](#environment) that `start` sets.

Before it stages anything, `start` checks the [environment](#environment).

Staged files (all `0755`, written atomically):

```text
<home>/cosmovisor/genesis/bin/exrpd             11.2.0
<home>/cosmovisor/upgrades/v12.0.0/bin/exrpd    12.0.0
<home>/cosmovisor/supervisor/<sha256[:16]>/cosmovisor
```

A file that already exists with the same SHA-256 is left alone. If its owner cannot
execute it (for example, a copy restored with mode `0644`), its mode is set to `0755`
again. A file with different content is replaced atomically, and `exrpd-upgrade` writes a
line to stderr that names the file and both SHA-256 values. This is how a patch release
bundle replaces the binaries of an earlier bundle. It also means that an older
`exrpd-upgrade` started on the home (for example, one still named in a service file) puts
its own binaries back: check the service log for `exrpd-upgrade: replaced`.
`current` is managed as described under [Restart and reinstall](#restart-and-reinstall).

Use the binary in `<home>/cosmovisor/current/bin/exrpd` for CLI commands (keys, queries,
transactions) so the client always matches the running chain version.

## Chain height

Cosmovisor reads the chain height with `exrpd status --home <home>`, without `--node`.
Keep the CometBFT RPC server on (`[rpc] laddr` in `config/config.toml`) and keep `node` in
`<home>/config/client.toml` pointed at it. The defaults do this. If the node halts at the
upgrade height and Cosmovisor does not switch, restart the service: `start` sees the plan
in `data/upgrade-info.json` and points `current` at the upgrade binary.

## systemd

```ini
[Unit]
Description=XRPL EVM node (exrpd upgrade binary)
After=network-online.target
Wants=network-online.target

[Service]
User=exrpd
ExecStart=/usr/local/bin/exrpd-upgrade start --home /var/lib/exrpd
Restart=on-failure
RestartSec=5
LimitNOFILE=65535
# Default KillMode=control-group: on stop, and whenever Cosmovisor exits on its own, every
# process left in the unit's cgroup is stopped too, so no exrpd outlives the service.
TimeoutStopSec=120

[Install]
WantedBy=multi-user.target
```

Because `start` execs Cosmovisor, `MainPID` is Cosmovisor itself. Keep the default
`KillMode=control-group`. Do not use `KillMode=process` or `none`.

## macOS (launchd)

Use a LaunchDaemon with `ProgramArguments` = `/usr/local/bin/exrpd-upgrade`, `start`,
`--home`, `<dir>` (launchd does not search `/usr/local/bin`, so give the absolute path)
and a `KeepAlive` dictionary with `SuccessfulExit` = `false`. launchd sends SIGTERM to the
job's PID (Cosmovisor), which forwards it to `exrpd`. When the job exits, launchd kills the
remaining processes of its process group unless `AbandonProcessGroup` is `true`. Keep the
default `false`. The darwin binaries are not signed or notarized. Remove the quarantine
attribute (`xattr -d com.apple.quarantine`) after download.

## Environment

`exrpd-upgrade start` sets these Cosmovisor variables and refuses to start if you set a
conflicting value:

| Variable | Value |
| --- | --- |
| `DAEMON_NAME` | `exrpd` |
| `DAEMON_ALLOW_DOWNLOAD_BINARIES` | `false` |
| `DAEMON_RESTART_AFTER_UPGRADE` | `true` |

It also sets `DAEMON_HOME` to the resolved home. A `$DAEMON_HOME` that differs from
`--home` is replaced, not refused.

Defaults that differ from upstream Cosmovisor (set the variable to override):

| Variable | Default | Why |
| --- | --- | --- |
| `DAEMON_SHUTDOWN_GRACE` | `30s` | time exrpd gets to exit after SIGINT at the upgrade height before it is killed |
| `UNSAFE_SKIP_BACKUP` | `true` | upstream copies all of `<home>/data` at the upgrade height while the validator is down |

An empty value counts as unset. Cosmovisor refuses a `DAEMON_SHUTDOWN_GRACE` of `0`, so
the upstream default (kill exrpd at once) is not available.

To keep the upstream backup, set `UNSAFE_SKIP_BACKUP=false` (optionally with
`DAEMON_DATA_BACKUP_DIR`). The copy needs free disk equal to the size of `data/`, and the
validator does not sign until it finishes. Do not stop the node during the copy:
Cosmovisor exits at once and leaves a partial `data-backup-<date>` directory. After a stop
or a failed copy, the next `start` points `current` at 12.0.0 itself (see below), so
Cosmovisor does not run the backup or `COSMOVISOR_CUSTOM_PREUPGRADE` again. Delete the
partial backup and take your own snapshot before you start again if you need one.

Keep a rollback point before the upgrade height. With the default
`UNSAFE_SKIP_BACKUP=true`, nothing copies `data/`. The upgrade migrates the stores at the
upgrade height, and 11.2.0 cannot open them after that. Pointing `current` back at
`genesis` or installing another binary does not restore the data. Enable the Cosmovisor
backup, or take a snapshot of `data/` with the node stopped and test that it restores.
If you restore a backup, keep the newer `data/priv_validator_state.json`: an older copy
lets the validator sign heights it already signed, which is double signing.

All other Cosmovisor variables keep their upstream meaning, notably
`DAEMON_POLL_INTERVAL`, `DAEMON_RESTART_DELAY`, `COSMOVISOR_COLOR_LOGS`,
`COSMOVISOR_TIMEFORMAT_LOGS`. Do not use Cosmovisor batch upgrades
(`data/upgrade-info.json.batch`) or `cosmovisor add-upgrade --upgrade-height`: both write
`data/upgrade-info.json` before the plan height, and `start` reads a file that names
`v12.0.0` as an upgrade the chain already reached, so 12.0.0 would start too early. The
governance upgrade needs neither.

## Restart and reinstall

At startup, Cosmovisor compares the plan in `<home>/data/upgrade-info.json` with the plan
recorded in `current/upgrade-info.json` and switches when they differ. x/upgrade writes the
data file when the chain reaches a plan height, so it names the last plan the chain
reached. Before Cosmovisor runs, `exrpd-upgrade` points `current` at the embedded binary
for that plan and records the plan there:

| `data/upgrade-info.json` | `current` |
| --- | --- |
| names `v12.0.0` | `upgrades/v12.0.0` |
| names any other plan (`v11.2.0` on homes that ran that upgrade) | `genesis`. Without the record, Cosmovisor would try to switch to `upgrades/v11.2.0` and stop, because it cannot download that binary |
| missing (fresh home) | unchanged. Cosmovisor links a missing `current` to `genesis` |
| not valid JSON, for example truncated by a crash | `start` and `stage` stop with an error before they change any file. Stop the node, delete the file and start again |

So:

- Restart before the upgrade height: 11.2.0 runs.
- Restart after the switch: 12.0.0 runs.
- Killed at the halt before Cosmovisor switched (power loss, SIGKILL, OOM), or
  `<home>/cosmovisor` deleted after the upgrade: the next start runs 12.0.0. The upgrade's
  optional `exrpd pre-upgrade` hook is not run in this case (12.0.0 does not implement it).
- Reinstalling `exrpd-upgrade` restages identical files. A patch release bundle replaces
  the binaries that differ and logs each replacement.
- A home that stock Cosmovisor or an earlier bundle set up is taken over: its `genesis` and
  `upgrades/v12.0.0` binaries are replaced, and `current` moves by the table above. Other
  directories under `upgrades/` stay on disk and are no longer used.
- To leave the upgrade binary, stop the service and run the standalone `exrpd` that matches
  `current`.

## Restoring from a snapshot or state sync

Restored data usually has no `data/upgrade-info.json` for the plan it is past, so select
the binary for the height of the data yourself, with the node stopped:

```sh
# data at or after the upgrade height (state sync today, or a snapshot taken after it)
exrpd-upgrade stage --home <home> --select v12.0.0
# data from before the upgrade height
exrpd-upgrade stage --home <home> --select genesis
```

`--select` accepts only `genesis` and `v12.0.0`. `--select v12.0.0` deletes a
`data/upgrade-info.json` that names another plan, such as `v11.2.0` kept in the home from
before the restore, and logs it. Cosmovisor would otherwise try to switch back to that
plan's binary. A `data/upgrade-info.json` that names `v12.0.0` wins over
`--select genesis`: if you restored data from before the upgrade height and kept that file
from the newer data, delete it first.

If you start without selecting, data past the upgrade height on `genesis` exits with a
store error such as `version of store icahost mismatch` on every start.

## Cosmovisor patch

Cosmovisor v1.7.3 registers a SIGTERM/SIGQUIT handler on every run and never removes it
after an upgrade. After the in-process switch, a SIGTERM sent only to Cosmovisor's PID
also reaches the stale handler, which cannot signal the old app and calls `os.Exit(1)`.
Cosmovisor then exits while 12.0.0 keeps running outside the service manager's view.

`scripts/upgrade-binary/cosmovisor-v1.7.3-signal.patch` is the fix proposed upstream in
[cosmos/cosmos-sdk#26879](https://github.com/cosmos/cosmos-sdk/pull/26879). It changes only
`Launcher.Run`: the handler is registered before the app starts and is stopped on every
return path. It adds `TestLaunchProcessSignalAfterUpgrade`, which fails on unpatched
v1.7.3. Remove the patch when a Cosmovisor release contains the fix.

The build fetches the module at v1.7.3, checks its `go.sum` hash
(`scripts/upgrade-binary/bundle.env`), runs `go mod verify` so that the extracted module
source and its dependencies match their hashes, applies the patch with `git apply --check`
first, and builds with Cosmovisor's own `go.mod` (`CGO_ENABLED=0`).
`scripts/upgrade-binary/cosmovisor.sh --test` runs the patched module's whole test suite.
CI runs it on every pull request.

## Building

```sh
mkdir -p release12 && tar -xzf node_12.0.0_Linux_amd64.tar.gz -C release12 bin/exrpd
scripts/upgrade-binary/build.sh \
  --genesis-archive node_11.2.0_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
  --upgrade-exrpd release12/bin/exrpd --upgrade-version 12.0.0 \
  --goos linux --goarch amd64 --out dist/exrpd-upgrade
scripts/upgrade-binary/cosmovisor.sh --test # patched Cosmovisor test suite
CACHE=<dir> scripts/upgrade-binary/e2e.sh               # localnet proof on linux/amd64
CACHE=<dir> scripts/upgrade-binary/e2e.sh --unpatched   # same run with the unpatched v1.7.3 release binary, must show the defect
```

`--upgrade-exrpd` is the `bin/exrpd` of the standalone 12.0.0 release archive, not the
`bin/exrpd` that `make build` writes. A local build has other bytes: a home that stages it
logs a replacement whenever it moves between that build and a released bundle.

`e2e.sh` does not download release assets. Put these files in a directory and pass it as
`CACHE`:

| File | Content |
| --- | --- |
| `node_11.2.0_Linux_amd64.tar.gz`, `checksums.txt` | assets of the GitHub release `v11.2.0` |
| `exrpd-<version>-linux-amd64` | standalone upgrade `exrpd` for linux/amd64, executable. `<version>` is `UPGRADE_VERSION`, by default the plan version (12.0.0) |
| `cosmovisor-v1.7.3-linux-amd64.tar.gz`, `SHA256SUMS-cosmovisor-v1.7.3.txt` | `--unpatched` only: the upstream Cosmovisor v1.7.3 release archive and a checksum list that contains its line |

For a patch bundle, set `UPGRADE_VERSION` to the patch version (for example
`UPGRADE_VERSION=12.0.1`). To test an `exrpd-upgrade` that was already built, for example
a release asset, pass it as `EXRPD_UPGRADE=<path>`. The run then checks that its manifest
names the plan and the SHA-256 of both `exrpd` files in `CACHE`, and does not build one.

The script stops with an error that names the first missing file. The build still needs
Go modules: without `--unpatched`, `cosmovisor.sh` runs `go mod download` for the pinned
Cosmovisor module and its dependencies, and the launcher build needs its dependencies
too. Run it with network access or with a Go module cache that already holds these
modules. Two runs cannot use the same `E2E_DIR` at the same time: the script holds the
lock `<E2E_DIR>.lock` and stops when another run holds it. The script deletes `E2E_DIR` at
the start, so it accepts only a directory that is missing, empty, or marked with
`.e2e-scratch` by an earlier run.

With `--unpatched`, the run passes only when the SIGTERM after the in-process switch shows
the defect (an orphaned `exrpd`, the stale handler's `process already finished` log, or
Cosmovisor exit status 1). With the patched build, every stop must show none of these.
Which signal handler runs first is not fixed, so a run often misses it. Run it again in
that case. `TestLaunchProcessSignalAfterUpgrade` is the deterministic check.

Pull requests that change `cmd/exrpd-upgrade/` or `scripts/upgrade-binary/` run `e2e.sh`
with an `exrpd` built from the pull request (`make build VERSION=<plan version>`) and the
published genesis release.

Releases: `.github/workflows/upgrade-binary.yml` runs after GoReleaser, downloads the
published standalone archives and the genesis release archives and runs
`scripts/upgrade-binary/release.sh`. It then runs `e2e.sh` with the linux/amd64 asset
(`EXRPD_UPGRADE`) and uploads the `node_upgrade_*` assets only when every check passes.
The standalone release is already published at that point. The workflow checks out the
tag, so what to do after a failure depends on the cause:

- A transient failure (network, runner): run the workflow again for the same tag.
- An upload that stopped part way: delete the `node_upgrade_*` assets of that release,
  then run the workflow again. It never replaces an asset that is already there.
- A defect in the tagged code or a wrong `bundle.env`: a rerun builds the same commit and
  fails again. Fix it on the release branch and publish a new release (for example the
  next patch), which gets its own standalone and upgrade binary assets.

## Releases and bundles

All pins live in `scripts/upgrade-binary/bundle.env`. Every release, pre-releases
included, gets an upgrade binary. Its version must have the major.minor of `UPGRADE_PLAN`
(today 12.0): `build.sh` and `release.sh` stop with an error otherwise, so the release
workflow fails until `bundle.env` is bumped. Tags without `bundle.env` (releases older
than the upgrade binary) ship standalone only.

| Release | Chain | Upgrade binary |
| --- | --- | --- |
| Major or minor (`X.0.0`, `X.Y.0`) | changes state or consensus, needs a governance plan `vX.Y.0` with an upgrade handler | new bundle: `GENESIS_VERSION` is the patch release validators run at the halt, `UPGRADE_PLAN` is the plan |
| Patch (`X.Y.Z`, `Z > 0`) | no upgrade handler, no state or consensus change: a drop-in replacement | bundle with the same plan and the patch as the upgrade binary |

A patch bundle reaches validators with a reinstall and a restart, before or after the
switch: `start` replaces `upgrades/v12.0.0/bin/exrpd` and logs it. The standalone patch
binary stays an alternative: stop the service and run `exrpd start --home <home>`.
Nothing checks that a patch is a drop-in replacement. A patch that changes state or
consensus must be a minor release with its own plan.

Patches of the genesis line, while an upgrade bundle is out but before its halt (for
example 12.0.3 while the v12.1.0 bundle carries genesis 12.0.2): release the standalone
12.0.3 and a patch of the upgrade line (12.1.1) with `GENESIS_VERSION=12.0.3`. Its `start`
replaces `genesis/bin/exrpd`. On the 12.0 branch, `bundle.env` still names `v12.0.0`, so
the 12.0.3 tag also publishes a v12.0.0 bundle for homes that still run that one.

The next upgrade bundle (for example v12.1.0) installs over a home that this bundle
staged without manual steps: `data/upgrade-info.json` names `v12.0.0`, so `current` moves
to its genesis binary (12.0.x).

## Known limitations

- The binary is about 100 MB on linux/amd64: it carries two full `exrpd` builds and
  Cosmovisor, compressed. Staging writes about 300 MB under `<home>/cosmovisor`.
- No Windows upgrade binary. Use the standalone `exrpd` there.
- One upgrade per bundle: each later plan needs its own bundle.
- Run the bundle for the chain's next plan before its height. A bundle started on data
  past a later plan points `current` at its genesis binary, which cannot run that data
  and exits.
- An `exrpd` can outlive Cosmovisor when Cosmovisor itself dies without forwarding a stop:
  SIGKILL, SIGHUP or another signal it does not handle, an out-of-memory kill, or an
  upstream panic while parsing a malformed non-empty `data/upgrade-info.json`. Under
  systemd with the default `KillMode=control-group` (and under launchd) the service
  manager stops the rest of the unit when Cosmovisor exits, so nothing keeps running.
  Without a service manager, start the node in its own session, for example
  `setsid exrpd-upgrade start --home <home> &`. Find its process group with
  `ps -o pgid= -p <cosmovisor pid>` and stop the node with `kill -TERM -- -<pgid>`. Do not
  signal the group of a node that you did not start with `setsid`: that group can contain
  your shell. Check for a leftover `exrpd` before starting again: a second node on the
  same home fails on the database lock instead of double signing.
- The supervisor is a patched Cosmovisor until the fix is available upstream.

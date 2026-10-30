# exrpd upgrade bundle: developer notes

Operator details, internals, build, CI and release process of the upgrade bundle. For
the short install and run steps, see the [operator guide](upgrade-binary.md).

| Path | Content |
| --- | --- |
| `cmd/exrpd-upgrade/` | the launcher, a separate Go module. It builds the bundle executable `exrpd` |
| `scripts/upgrade-binary/plan.sh` | prints the upgrade plan: the `UpgradeName` of the one handler in `app/upgrades` |
| `scripts/upgrade-binary/genesis.sh` | prints the default genesis: the newest published release below the plan |
| `scripts/upgrade-binary/genesis-plan.sh` | prints the plan of a genesis release: the newest plan in `app/upgrades` of its tag |
| `scripts/upgrade-binary/cosmovisor/` | Cosmovisor pins (`cosmovisor.env`) and the patch |
| `scripts/upgrade-binary/build.sh` | builds one bundle executable |
| `scripts/upgrade-binary/cosmovisor.sh` | builds or tests the patched Cosmovisor |
| `scripts/upgrade-binary/release.sh` | builds the release assets from published archives |
| `scripts/upgrade-binary/e2e.sh` | localnet chain upgrade on linux/amd64 |
| `.github/workflows/upgrade-binary.yml` | "Upgrade binary" workflow: builds and uploads the bundles of a published release |

## Operator details

### Who uses the bundle

The bundle is for validators and other nodes that must switch binaries at the upgrade
height together with the rest of the network. Machines that only run client commands
(keys, transactions, queries) against a node use the standalone `exrpd` and update it by
hand. A home without `data/upgrade-info.json` runs the genesis binary, so on a client
machine the bundle would keep the old client after the chain upgrades.

### Standalone binary or bundle

| | Standalone `exrpd` | Bundle `exrpd` |
| --- | --- | --- |
| Release asset | `node_<version>_<Os>_<arch>.tar.gz` (`.zip` on Windows) | `exrpd-upgrade_<genesis>-<version>_<Os>_<arch>.tar.gz`, with `exrpd` and `LICENSE` |
| Checksums | `checksums.txt` | `exrpd-upgrade_<genesis>-<version>_checksums.txt` |
| Contents | one chain version | 11.3.0 + 12.0.0 + Cosmovisor |
| At the upgrade height | halts, then you swap the binary and restart | switches to 12.0.0 by itself |
| Platforms | linux/darwin amd64/arm64, windows/amd64 | linux/darwin amd64/arm64 |

Both use the same node home, keys, config and service command
(`exrpd start --home <home>`). You can move between them while the node is stopped. Do
not install both on one machine: it is then not clear which `exrpd` runs.

**Forward only.** A chain only moves forward. Never start an older release or bundle on a
home that a newer one has run, and never put back an older
`data/priv_validator_state.json`. The bundle cannot detect every such case.

### Commands

- Home: `--home`/`--home=`, else `$DAEMON_HOME`, else `$EXRPD_HOME`, else `$NODE_HOME`,
  else `$HOME/.exrpd`. The standalone exrpd opens the node's config and data on
  `$EXRPD_HOME`, else on `$NODE_HOME`, so the bundle picks the same home for them. One
  difference: with `$EXRPD_HOME` set, the standalone keeps the keyring in the default
  home, the bundle in `$EXRPD_HOME`. `--home` gives one home for everything. The launcher
  resolves the home once and passes it as `--home` first (`exrpd --home <home> ...`) to
  every command it runs, so exrpd and Cosmovisor always use the same home and no flag of
  the command can take it as its value.
- A TAB key sends a completion request (`__complete`). Its words go to exrpd exactly as
  typed, and the launcher passes `--home` to exrpd, as for every command. It reads
  `--home` from the words only to pick the binary, and it never installs: without an
  installed exrpd there is nothing to suggest. It also detects a TAB request behind a
  shell alias that adds `--home`, for example `alias exrpd='exrpd --home /var/lib/exrpd'`.
  `exrpd completion` (bash) and `exrpd completion --zsh` write a script and are ordinary
  commands.
- Help (`exrpd` alone, `exrpd help ...`, or `-h`/`--help` in any command but `start`)
  never installs. When nothing is installed in the home yet, it prints a hint and exits
  0: for a new node, run `exrpd init <moniker> --chain-id <chain id>`. For an existing
  node, the home is probably wrong: pass it with `--home`, and see `exrpd version --bundle`
  for the home in use and why. `exrpd start --help` goes through Cosmovisor like every
  `start`.
- Every `exrpd start` writes the home and where it comes from to stderr before it
  installs, for example `exrpd: home /var/lib/exrpd (from --home)`. The first install into
  a home, when `<home>/cosmovisor` did not exist before, also writes
  `exrpd: installed the bundle in <home>/cosmovisor`.
- `exrpd start` checks the home, installs the files, points `current` at the binary for
  the data, then replaces itself with Cosmovisor (same PID). Cosmovisor runs
  `exrpd start`. Under systemd, `MainPID` is Cosmovisor. `start` passes every other
  argument to `exrpd start`. It refuses `--cosmovisor-config`, because Cosmovisor would
  then read its settings from that file and ignore the environment that `start` sets.
- Every other command runs with the exrpd that runs the node data, chosen by the same rule
  as `start` (see [How the launcher chooses the binary](#how-the-launcher-chooses-the-binary)).
  It never moves `current`, because the node may be running. It installs the files only
  when that exrpd is missing, for example for `exrpd init` on a new home. Example: after a
  kill at the halt, `current` still names the genesis plan, but `exrpd rollback` runs
  12.0.0 because the data file names `v12.0.0`.
- `exrpd version` prints one line: the version of the exrpd installed for the node data,
  then the bundle. The first word is the same as the standalone `exrpd version` prints:
  `12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3`. After you install a
  patch bundle and before you restart, the first word is still the earlier version. It
  reads the installed binary, not the running process. To ask the running node, use
  `curl -s localhost:26657/abci_info | jq -r .result.response.version`. The
  `node_info.version` of `exrpd status` is the CometBFT version.
- `exrpd version --bundle` prints JSON. It only reads files and never runs a chain binary.
  Its fields:
  - `version`: a line in the form of `exrpd version`, with `will_run.version` as its first
    word, for example `12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3`.
  - `home`: the absolute home.
  - `home_source`: where the home comes from: `--home`, `$DAEMON_HOME`, `$EXRPD_HOME`,
    `$NODE_HOME` or `default`.
  - `will_run`: the binary that the next command runs, by the rule in
    [How the launcher chooses the binary](#how-the-launcher-chooses-the-binary): `dir`
    (for example `upgrades/v12.0.0`), `binary` (absolute path), `version`, `installed`
    (`true` when the file exists) and `reason` (why that directory). `version` is the
    bundled version when the file is missing or has the bundled SHA-256, else `unknown`.
  - `refused`: `null`, or the error that stops every command on this home. Then
    `will_run` is `null`, and the command prints the JSON and exits 1.
  - `dirs`: the directories this bundle uses under `<home>/cosmovisor`.
  - `manifest`: versions, plans, SHA-256 and Cosmovisor source.

  The earlier top-level `installed` field is gone: use `will_run.version`. `--bundle`
  must come after `version` as a bare word, and it takes no other flag. `version --long`
  and other flags go to the exrpd that runs the data.
- `start` after a flag, for example `exrpd --log_level info start`, is refused. The
  launcher cannot tell which flag takes a value, and without Cosmovisor the node would not
  upgrade. Write `exrpd start [flags]`.
- A bare `--` is refused. The launcher cannot tell whether `--` is the value of the flag
  before it or the end of the flags. If it guessed wrong, Cosmovisor and exrpd could use
  different homes. To pass `--` as a value, write it with `=`, for example `--moniker=--`.
- `--home` directly after another flag that has no `=` is refused, for example
  `--moniker --home=/srv/name`. exrpd can read the second token as the value of the first
  flag, and the launcher does not know which flags take a value. Write the other flag with
  `=`, for example `--moniker=--home=/srv/name` or `--yes=true`, or put `--home` first.
- Cosmovisor reads the chain height with `exrpd status --home <home>`, without `--node`.
  This is why the CometBFT RPC server (`[rpc] laddr` in `config/config.toml`) must stay on
  and `node` in `<home>/config/client.toml` must point at it.
- If the node halts at the upgrade height and Cosmovisor does not switch, a restart works
  because `start` sees the plan in `data/upgrade-info.json` and points `current` at the
  upgrade binary.

### macOS (launchd)

Use a LaunchDaemon with `ProgramArguments` = `/usr/local/bin/exrpd`, `start`, `--home`,
`<dir>` and a `KeepAlive` dictionary with `SuccessfulExit` = `false`. Give the absolute
path, because launchd does not search `/usr/local/bin`. launchd sends SIGTERM to the
job's PID (Cosmovisor), which forwards it to `exrpd`. When the job exits, launchd kills
the remaining processes of its process group unless `AbandonProcessGroup` is `true`.
Keep the default `false`.

### Without a service manager

An `exrpd` can outlive Cosmovisor when Cosmovisor dies without forwarding a stop:
SIGKILL, SIGHUP or another signal it does not handle, an out-of-memory kill, or an
upstream panic while parsing a malformed non-empty `data/upgrade-info.json`. Under systemd
with the default `KillMode=control-group`, and under launchd, the service manager stops
the rest of the unit when Cosmovisor exits, so nothing keeps running. Without a service
manager, start the node in its own session and stop it by process group:

```sh
setsid exrpd start --home <home> &
ps -o pgid= -p <cosmovisor pid>          # find the process group
kill -TERM -- -<pgid>                    # stop the node
```

Do not signal the group of a node that you did not start with `setsid`. That group can
contain your shell. Check for a leftover `exrpd` before you start again. A second node on
the same home fails on the database lock instead of double signing.

### Installed files

Installed binaries are `0755` and plan records `0644`. All are written atomically. On
every `start`, and on every other command whose exrpd is missing:

- A file with the same SHA-256 is left alone. If its owner cannot execute it (for
  example, a copy restored with mode `0644`), its mode is set to `0755` again. So a
  reinstall of the same bundle installs identical files.
- A file with different content is replaced atomically. The launcher writes a line to
  stderr that names the file and both SHA-256 values. A patch release bundle replaces the
  binaries of an earlier bundle this way.
- An older bundle started on the home (for example, one still named in a service file)
  puts its own binaries back. Check the service log for `exrpd: replaced`.
- A home that stock Cosmovisor set up can be taken over as it is: stop Cosmovisor, then
  start the bundle with the same home. The operator guide recommends a clean move
  instead, see [If you use or used Cosmovisor](upgrade-binary.md#if-you-use-or-used-cosmovisor).
  On a takeover, the `genesis`, `upgrades/v11.3.0` and `upgrades/v12.0.0` binaries are
  replaced (an operator's own build of 12.0.0 too), and `current` moves by the rule in
  [How the launcher chooses the binary](#how-the-launcher-chooses-the-binary). A
  `cosmovisor/config.toml` from `cosmovisor init` is ignored: Cosmovisor reads it only
  with `--cosmovisor-config`, which the bundle refuses. Remove
  `DAEMON_ALLOW_DOWNLOAD_BINARIES=true` from an old service file: `start` refuses it. A
  `data/upgrade-info.json` that `cosmovisor add-upgrade --upgrade-height` wrote ahead of
  the upgrade height makes the bundle start 12.0.0 too early, and exrpd stops: delete it
  while the chain is below that height. Every command refuses while
  `data/upgrade-info.json.batch` exists: the batch watcher writes `data/upgrade-info.json`
  before the height, and after a restart the bundle would trust that file. Delete the
  batch file, and also a data file the watcher wrote early, see
  [Operational caveats](#operational-caveats). Other directories under
  `upgrades/` and `supervisor/` stay on disk. See [Removing old files](#removing-old-files).
- The directories under `<home>/cosmovisor` are managed by the bundle. Do not replace
  them or their `bin/` with symlinks. When two binary directories of the bundle are one
  directory, for example `genesis` made a symlink to `upgrades/v12.0.0`, `exrpd start`,
  and any command that installs, stops before it writes. Other symlinks there are not
  checked.
- Do not copy the bundle `exrpd` into a directory under `cosmovisor/`: it is not a chain
  binary. With stock Cosmovisor, `bin/exrpd` is often a symlink to
  `/usr/local/bin/exrpd`, which is the bundle once it is installed there. A command that
  would run it stops with `... is this exrpd bundle, not a chain binary`. `exrpd start`
  replaces it with the chain binary.

### Removing old files

The launcher never deletes a directory. Each bundle adds about 100 MB or more under
`<home>/cosmovisor`. After you install a new bundle, remove what it does not use:

1. Start the new bundle once and check that the node runs, then stop it. This moves
   `current` off the directories of the old bundle.
2. As the node user, run `exrpd version --bundle --home <home>`. Its `dirs` list names
   every directory this bundle uses. If the command fails, delete nothing.
3. Delete only the subdirectories of `<home>/cosmovisor/upgrades/` and
   `<home>/cosmovisor/supervisor/` that the list does not name. Never delete the directory
   that `current` points at. Do not touch anything else in `<home>/cosmovisor`:
   `genesis/`, `current`, and your own files, such as a `COSMOVISOR_CUSTOM_PREUPGRADE`
   script.
4. Start the node.

### Restart, resync and restore details

- Killed at the halt before Cosmovisor switched (power loss, SIGKILL, OOM), or
  `<home>/cosmovisor` deleted after the upgrade on a home whose `data/upgrade-info.json`
  names `v12.0.0`: the next start runs 12.0.0. Cosmovisor's backup,
  `COSMOVISOR_CUSTOM_PREUPGRADE` and the upgrade's optional `exrpd pre-upgrade` hook are
  not run in this case (12.0.0 does not implement the hook).
- **Resync after the upgrade.** `exrpd comet unsafe-reset-all` deletes all of `data/`,
  `data/upgrade-info.json` and the signing state included. Keep both, see
  [Resync after the upgrade](#resync-after-the-upgrade).
- **New node after the upgrade.** A node that joins by state sync or with a snapshot from
  after the upgrade height has no `data/upgrade-info.json`, so the bundle would run 11.3.0
  on 12.0 data. Run the standalone 12.0.x instead. Before the height of the next plan,
  install the next bundle: its genesis is 12.0.x, so it runs that data with no extra step.
  This also works on a home that an earlier bundle used.
- **Restore after the upgrade on an existing node.** The same as a new node: run the
  standalone 12.0.x until the next bundle, unless `current` still points at
  `upgrades/v12.0.0` (then the bundle keeps it, as after a resync).
- **Rollback to a snapshot from before the upgrade height.** With the node stopped, delete
  `<home>/cosmovisor/current` and a `data/upgrade-info.json` that names `v12.0.0`, if the
  restore kept one. The next start then runs the genesis plan's binary for the data.
- **Corrupted `data/upgrade-info.json`.** x/upgrade writes this file without fsync, so a
  crash or a disk fault can leave it empty or cut short. This risk exists with every
  setup: exrpd and stock Cosmovisor also stop on such a file. Every command stops with an
  error before it changes any file. The data can be at the upgrade height or past it, so
  do not just delete the file: write it again with the exact plan name and height of the
  last upgrade the chain reached, from the governance proposal, for example
  `{"name":"v12.0.0","height":4521000}`. exrpd uses both values to apply its store changes
  at that height.
- Install the bundle for the chain's next plan before its height, and install bundles in
  order: start a bundle only on data that reached its genesis plan. A bundle started on
  data whose `data/upgrade-info.json` names a plan it does not carry stops with an error
  before it changes any file.
- **Stale `data/upgrade-info.json`.** x/upgrade writes the file only when the running
  binary halts at a plan. A node that applied a plan with a binary swapped in before the
  plan height (for example with `--halt-height`) keeps the file of an earlier plan. A file
  that names a plan older than the bundle's genesis plan stops every command with a
  message that says so. When you move a standalone home to the bundle, check that the
  file names the last plan this node crossed. If it does not, write it again with that
  plan and its height: a stale file that names a plan the bundle carries (for example
  `v11.3.0` on a node that already swapped to 12.0.x) selects the older binary, and exrpd
  stops on the newer data.

### Resync after the upgrade

Prepare the resync as you normally do: configure state sync in `config/config.toml`, or
get a snapshot. The bundle adds one thing: keep `data/upgrade-info.json` together with
`priv_validator_state.json`. The data file makes the next start pick 12.0.0 even when
`current` was not moved yet, for example after a crash at the halt. After the state sync,
exrpd ignores the height in it, because the node starts far past that height.

Run the commands one by one. If a command fails, stop there and do not start the node.
`exrpd` in `sudo -u exrpd` and `systemctl ... exrpd` is the user and unit of the systemd
unit in the operator guide: use yours if they differ.

0. Find your node's home. It is the folder your service runs with. Look in your service
   file (`systemctl cat exrpd`). If the `ExecStart` line has `--home`, that is it. If an
   `Environment=` line sets `DAEMON_HOME`, `EXRPD_HOME` or `NODE_HOME`, that is it. If
   neither is there, it is `.exrpd` in the home folder of the user the service runs as,
   for example `/home/exrpd/.exrpd`. Set it and check it. Do not write `~` in the
   commands below: your shell expands it to your own home, not the service user's.

   ```sh
   home=/var/lib/exrpd
   sudo -u exrpd ls "$home/data/priv_validator_state.json"
   ```

1. Stop the node: `sudo systemctl stop exrpd`
2. Create the backup directory: `sudo -u exrpd mkdir "$home/resync-backup"`. If it exists
   already, stop: an earlier resync did not finish. Do not overwrite or delete a file in
   `data/` or in `resync-backup/` until you know what that resync did. The signing state
   holds a height, a round, a step and the last signature, so two files with the same
   height are not the same.
3. Save the signing state:
   `sudo -u exrpd cp "$home/data/priv_validator_state.json" "$home/resync-backup/"`
4. Save the data file:
   `sudo -u exrpd cp "$home/data/upgrade-info.json" "$home/resync-backup/"`. If it fails
   with "No such file", the node has no data file: skip step 8.
5. Reset the data:
   `sudo -u exrpd /usr/local/bin/exrpd comet unsafe-reset-all --home "$home" --keep-addr-book`
6. If you use a snapshot, unpack it into the home now, so `data/` is filled. The snapshot
   must be taken after the upgrade height. With an older snapshot, the node must run
   11.3.0 up to the upgrade: skip step 8 and delete `$home/cosmovisor/current`, as in the
   Rollback row of the operator guide, else 12.0.0 stops at once on the older data. With
   state sync, skip this step: the node fetches the state when it starts.
7. Put the signing state back:
   `sudo -u exrpd cp "$home/resync-backup/priv_validator_state.json" "$home/data/"`
8. Put the data file back, only if step 4 saved it:
   `sudo -u exrpd cp "$home/resync-backup/upgrade-info.json" "$home/data/"`
9. Start the node: `sudo systemctl start exrpd`
10. When the node runs and signs again, delete the backup:
    `sudo -u exrpd rm -r "$home/resync-backup"`

The backup is in the home, outside `data/`, because the reset deletes `data/`. The
commands run as the node user, because `data/` is `0700`. The copied
`priv_validator_state.json` keeps its `0600` mode, so another user cannot read it.

### Settings details

- An empty value counts as unset for the variables that this bundle pins or gives a
  default. All other variables go to Cosmovisor unchanged. For example, an empty
  `COSMOVISOR_TIMEFORMAT_LOGS` removes the log timestamps. Unset it to keep the default.
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
`current` back or installing another binary does not restore the data.

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

The manifest lists the genesis release with its plan, the upgrades with their plans, in
plan order, and Cosmovisor. The genesis plan is the newest `UpgradeName` in
`app/upgrades` of the genesis release tag (`genesis-plan.sh`), not a value derived from
the version: v10.2.x runs plan `v10.1.0`. Plans are lower-case `vX.Y.Z` names, because
Cosmovisor lowercases the plan it reads from the data file and switches to
`upgrades/<that name>`. `build.sh` refuses any other name. `genesis-plan.sh` refuses a
genesis tag that also has a plan with another name that is newer than its `vX.Y.Z` plan,
such as `v10.2.0-rc1` next to `v10.1.0`, or that has no `vX.Y.Z` part: skipping it could
pick the binary of an older plan. A name that is not newer, such as `v8.0.0-rc.1` next to
`v8.0.0`, is ignored.

The three executables are stored gzip-compressed. This is why the bundle is about 100 MB
on linux/amd64 although it carries two full `exrpd` builds and Cosmovisor. The SHA-256
and size in the manifest are those of the uncompressed executables, and installing checks
them. The genesis binary is installed twice: in `genesis/`, for a home without a data
file, and in `upgrades/<genesis plan>/`. Writing the files never changes `current`. The
launcher sets it in a separate step, see the next section.

## How the launcher chooses the binary

At startup, Cosmovisor compares the plan in `<home>/data/upgrade-info.json` with the plan
recorded in `current/upgrade-info.json`. It switches to `upgrades/<data plan>` when they
differ. x/upgrade writes the data file when the running binary halts at a plan height,
so it names the last plan the node halted at (see the stale file in
[Restart, resync and restore details](#restart-resync-and-restore-details)). A binary
runs only the data of its own plan: patch releases keep the plan of their release.

Every plan of the bundle has its directory `upgrades/<plan>`, the genesis plan included,
so the data file always names a directory of the bundle and `genesis/` never needs a
record. Before anything changes, the launcher decides:

| Home | Binary | `start` |
| --- | --- | --- |
| `data/upgrade-info.json.batch` exists, any content. Checked before every other row | none | stops with an error, nothing changes. See batch upgrades in [Operational caveats](#operational-caveats) |
| `data/upgrade-info.json` names a plan of the bundle (`v11.3.0` or `v12.0.0`, any case) | `upgrades/<plan>` | points `current` there and copies the data file as its record |
| `data/upgrade-info.json` names another plan (for example `v12.1.0`) | none | stops with an error, nothing changes. A plan older than the genesis plan (for example `v11.2.0`) gets a message that the file may be stale |
| `data/upgrade-info.json` is corrupted: not JSON, no name, no height above 0, a time, an upgraded client state, or a field of the wrong type, as exrpd and Cosmovisor check it | none | stops with an error, nothing changes. An empty file is read again for up to 1 s first, because x/upgrade truncates the file before it writes it |
| no data file, `current` names a directory of the bundle, also when it was deleted | that directory | keeps `current`. Install creates a deleted directory again, with its record |
| no data file, no `current` | `genesis` | leaves `current` missing. Cosmovisor links it to `genesis` |
| no data file, `current` on another directory whose record names a plan of the bundle | `upgrades/<plan>` | moves `current` there and copies the record |
| no data file, `current` on another directory named `upgrades/<plan>` without a record, plan of the bundle | `upgrades/<plan>` | moves `current` there and writes the record |
| no data file, `current` on any other directory, deleted or not | none | stops with an error, nothing changes |

The last row stops because a chain only moves forward: an unknown directory most likely
belongs to a later plan, and genesis is the oldest binary of the bundle. The data file
must have a height, as x/upgrade and Cosmovisor require. A record in a binary directory
needs only a plan name, as Cosmovisor writes it.

Each move of `current` writes one line to stderr, so the service log shows why the node
runs a binary. For example:
`exrpd: moved <home>/cosmovisor/current from genesis to upgrades/v12.0.0: data/upgrade-info.json names v12.0.0`.

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
  reached, so 12.0.0 would start too early. The governance upgrade needs neither. Every
  command refuses while `data/upgrade-info.json.batch` exists, any content, `[]` too: under
  the bundle, Cosmovisor still runs the batch watcher, which writes the data file before
  the height, and after a restart the bundle would trust that file. Delete the batch file.
  If the node is below the upgrade height, also delete a `data/upgrade-info.json` that
  names that upgrade: the batch watcher wrote it early. A file from
  `add-upgrade --upgrade-height` looks the same as a file the chain wrote, so the launcher
  cannot refuse it.
- **Sync from genesis.** A node that replays the chain from block 1 needs the binary of
  every plan. The bundle carries two, so it does not support this. Archive nodes set up
  stock Cosmovisor with the standalone releases.

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
  --goos linux --goarch amd64 --out dist/exrpd
scripts/upgrade-binary/cosmovisor.sh --test # patched Cosmovisor test suite
CACHE=<dir> GENESIS_VERSION=11.3.0 scripts/upgrade-binary/e2e.sh   # localnet proof on linux/amd64
CACHE=<dir> GENESIS_VERSION=11.3.0 scripts/upgrade-binary/e2e.sh --unpatched   # same run with the unpatched v1.7.3 release binary, must show the defect
```

`--upgrade-exrpd` is the `bin/exrpd` of the standalone 12.0.0 release archive, not the
`bin/exrpd` that `make build` writes. A local build has other bytes. A home that installs
it logs a replacement whenever it moves between that build and a released bundle.

`build.sh` reads the genesis plan from the genesis release tag (`genesis-plan.sh`) and
fetches the tag when the checkout lacks it. `--genesis-plan vX.Y.Z` sets it instead.

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
| `EXRPD_UPGRADE=<path>` | tests a bundle `exrpd` that was already built, for example extracted from a release asset. The run checks that its manifest names the plans and the SHA-256 of both `exrpd` files in `CACHE`, and does not build one |
| `E2E_DIR=<dir>` | scratch directory |

- The script stops with an error that names the first missing file.
- The build still needs Go modules. Without `--unpatched`, `cosmovisor.sh` runs
  `go mod download` for the pinned Cosmovisor module and its dependencies. The launcher
  module uses only the standard library. Run it with network access or with a Go module
  cache that already holds these modules.
- The `stock` scenario sets up and runs a home with stock Cosmovisor, built with
  `go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.7.3`, then takes it over
  with the bundle before the upgrade height and checks the switch.
- Two runs cannot use the same `E2E_DIR` at the same time. The script holds the lock
  `<E2E_DIR>.lock` and stops when another run holds it.
- The script deletes `E2E_DIR` at the start. It accepts only a directory that is missing,
  empty, or marked with `.e2e-scratch` by an earlier run.
- With `--unpatched`, the run passes only when the SIGTERM after the in-process switch
  shows the defect. Both conditions must be true: a defect signature (an orphaned `exrpd`
  that continues to produce blocks, or the stale handler's `process already finished` log)
  and Cosmovisor exit status 1. Exit status 1 without a signature fails. A signature with
  another exit status also fails. With the patched build, every stop must show no orphaned
  `exrpd`, no stale handler log, and exit status 0. Which signal handler runs first is not
  fixed, so a run often misses the defect. Run it again in that case.
  `TestLaunchProcessSignalAfterUpgrade` is the deterministic check.

## CI

Pull requests (`.github/workflows/pull-request.yml`):

| Job | Runs |
| --- | --- |
| `upgrade-binary` | every pull request and every push to `main`: launcher `go vet`, `go test` and `golangci-lint` in `cmd/exrpd-upgrade`, then `cosmovisor.sh --test` |
| `upgrade-binary-e2e` | pull requests that change `cmd/exrpd/`, `cmd/exrpd-upgrade/`, `scripts/upgrade-binary/`, `app/upgrades/`, `app/upgrades.go`, `app/app.go`, the root `go.mod` or `go.sum`, or `.github/workflows/pull-request.yml`: `e2e.sh` with an `exrpd` built from the pull request (`make build VERSION=<plan version>`) and the genesis release that `genesis.sh` picks |

The "Release" workflow (`.github/workflows/release.yml`) runs GoReleaser, then the "Upgrade
binary" workflow (`.github/workflows/upgrade-binary.yml`). "Upgrade binary" can also be
dispatched alone for a published tag, with `tag` and `genesis_tag`. It does:

1. Read the plan from the tagged code (`plan.sh`). Take the genesis from the `genesis_tag`
   input, or from `genesis.sh` when it is empty.
2. Run the launcher unit tests and `cosmovisor.sh --test`.
3. Download the published standalone archives and the genesis release archives.
4. Run `scripts/upgrade-binary/release.sh` with a new or empty `--out-dir`. It reads the
   genesis plan from the genesis tag and fetches that tag.
5. Run `e2e.sh` with the `exrpd` from the linux/amd64 asset (`EXRPD_UPGRADE`).
6. Upload the `exrpd-upgrade_*` assets, only when every check passes.

The standalone release is already published at that point. The workflow checks out the
tag, so what to do after a failure depends on the cause:

| Cause | Recovery |
| --- | --- |
| transient failure (network, runner) | dispatch "Upgrade binary" again for the same tag. Do not run "Release" again: it builds and pushes the Docker image, the GitHub release and the GoReleaser assets again |
| upload stopped part way | delete the `exrpd-upgrade_<genesis>-<version>_*` assets of the failed run, then dispatch "Upgrade binary" again. Keep the assets of other runs, for example another network's genesis. It never replaces an asset that is already there |
| defect in the tagged code or a missing upgrade handler | a rerun builds the same commit and fails again. Fix it on the release branch and publish a new release (for example the next patch), which gets its own standalone and bundle assets |
| wrong genesis | the `genesis_tag` input or the default picked the wrong release. If the run failed before the upload, dispatch "Upgrade binary" again with the right `genesis_tag`. After an upload, delete the `exrpd-upgrade_<wrong genesis>-<version>_*` assets first |

## Releases and bundles

Every release, pre-releases included, gets a bundle. A bundle has three inputs:

| Input | Source |
| --- | --- |
| Plan | `UpgradeName` of the one handler in `app/upgrades` of the tagged code (`plan.sh`) |
| Genesis | `genesis_tag` input of "Release" or "Upgrade binary". Default: the newest published release, drafts and pre-releases excluded, whose major.minor is below the plan's (`genesis.sh`). Its plan comes from its tag (`genesis-plan.sh`) and must come before the bundle's plan: `build.sh` refuses, for example, a 12.0.x genesis for plan `v12.0.0` |
| Cosmovisor | `scripts/upgrade-binary/cosmovisor/cosmovisor.env` and the patch next to it |

The release version must have the major.minor of the plan (today 12.0). `build.sh` and
`release.sh` stop with an error otherwise, so a major or minor release without its
upgrade handler fails the "Upgrade binary" job. Tags without `plan.sh` (releases older than
the bundle) ship standalone only. The asset name holds the genesis version and the
release version, for example `exrpd-upgrade_11.3.0-12.0.1_Linux_amd64.tar.gz`, so a
patch bundle never has the name of an earlier one. `exrpd version --bundle` prints the
manifest.

| Release | Chain | Bundle |
| --- | --- | --- |
| Major or minor (`X.0.0`, `X.Y.0`) | changes state or consensus, needs a governance plan `vX.Y.0` with an upgrade handler | new bundle: the genesis is the patch release validators run at the halt, the plan is the new handler |
| Patch (`X.Y.Z`, `Z > 0`) | no new upgrade handler, no state or consensus change: a drop-in replacement | bundle with the same plan and the patch as the upgrade binary |

A patch bundle reaches validators with a reinstall and a restart, before or after the
switch: `start` replaces `upgrades/v12.0.0/bin/exrpd` and logs it. The standalone patch
binary stays an alternative. Nothing checks that a patch is a drop-in replacement. A
patch that changes state or consensus must be a minor release with its own plan.

**Patch of the genesis line.** While a bundle is out but before its halt (for example
12.0.3 while the v12.1.0 bundle carries genesis 12.0.2):

1. Release the standalone 12.0.3.
2. Release a patch of the upgrade line (12.1.1). Its default genesis is now 12.0.3, or pass
   `genesis_tag: v12.0.3`. Its `start` replaces `genesis/bin/exrpd` and
   `upgrades/v12.0.0/bin/exrpd`.

On the 12.0 branch, the handler is still `v12.0.0`, so the 12.0.3 tag also publishes
a v12.0.0 bundle for homes that still run that one.

**Next bundle.** The next bundle, for example v12.1.0, has genesis 12.0.x and genesis
plan `v12.0.0`. On a home whose data file names `v12.0.0`, `current` already points at
`upgrades/v12.0.0`: the bundle replaces the binary there with 12.0.x and adds
`upgrades/v12.1.0`. Without a data file, `current` on `upgrades/v12.0.0` is a directory
of the next bundle too, so it stays. No manual step is needed in either case.

**Different genesis per network.** Usually every network runs the same release before an
upgrade, so the one bundle that "Release" publishes serves all of them. Sometimes a
network runs another release, for example testnet on 11.1.x and mainnet on 11.2.x. Then,
after "Release" has finished, dispatch "Upgrade binary" for the same `tag` with that
network's `genesis_tag`, once for each such network. Do not run "Release" again: it builds
and pushes the Docker image, the GitHub release and the standalone assets again, and the
bundles already published would no longer match the standalone files. The asset names
hold the genesis version, for example `exrpd-upgrade_11.1.3-12.0.0_Linux_amd64.tar.gz` and
`exrpd-upgrade_11.2.0-12.0.0_Linux_amd64.tar.gz`, so both sets go to the same release with
their own checksums. Tell the validators of each network which set is theirs. The upgrade
handler must accept the state of each genesis: the bundle does not check that.

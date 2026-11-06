# exrpd bundle: developer notes

> [!NOTE]
> The versions in this document are examples from the time of writing. 11.3.0
> is the version that the chain runs before the upgrade, and 12.0.0 is the version that
> the chain upgrades to. For another bundle, take the versions from its archive name, for
> example `exrpd-bundle_<current>-<upgrade>_<Os>_<arch>.tar.gz`. The plans `v11.3.0` and
> `v12.0.0` do not always match the versions. Take them from `manifest.genesis.plan` and
> `manifest.upgrades[0].plan` in `exrpd version --bundle`.

This document explains in detail how the bundle works, and how to build, test and
release it. The first sections follow the order of the [operator guide](bundle.md):
each one adds the details of the guide section with the same name. The sections from
[Payload](#payload) on are for developers only.

## How the bundle works

### Standalone binary or bundle

| | Standalone `exrpd` | Two-binary bundle | Single bundle |
| --- | --- | --- | --- |
| Release asset | `node_<version>_<Os>_<arch>.tar.gz` (`.zip` on Windows) | `exrpd-bundle_<genesis>-<version>_<Os>_<arch>.tar.gz`, with `exrpd` and `LICENSE` | `exrpd-bundle_<version>_<Os>_<arch>.tar.gz`, with `exrpd` and `LICENSE` |
| Checksums | `checksums.txt` | `exrpd-bundle_<genesis>-<version>_checksums.txt` | `exrpd-bundle_<version>_checksums.txt` |
| Docker image | `peersyst/exrp:v<version>` | `peersyst/exrp:bundle-<genesis>-<version>` | `peersyst/exrp:bundle-<version>` |
| Contents | one chain version | 11.3.0 + 12.0.0 + Cosmovisor | 12.0.0 + Cosmovisor |
| Runs | the data of its version | data before and after the upgrade height | data after the upgrade height |
| At the upgrade height | halts, then you swap the binary and restart | switches to 12.0.0 by itself | at the next upgrade height, halts as the standalone. Install the next two-binary bundle before it |
| Platforms | linux/darwin amd64/arm64, windows/amd64 | linux/darwin amd64/arm64 | linux/darwin amd64/arm64 |

All of them use the same node home, keys, config and service command
(`exrpd start --home <node home>`). You can move between them while the node is stopped.
Do not install two of them on one machine, because then you cannot know which `exrpd`
runs.

A single bundle is a bundle without an upgrade: its genesis release is the release
itself. A node after the upgrade height runs it until it installs the two-binary bundle
of the next upgrade, for example `exrpd-bundle_12.0.x-12.1.0`. So a node can always run a
bundle, also a new node after the upgrade.

Machines that only run client commands (keys, transactions, queries) use the standalone
`exrpd`, and you update it by hand. On a home without `data/upgrade-info.json`, the bundle
runs the genesis binary. So on a client machine, a two-binary bundle would keep the old
client after the chain upgrades.

> [!WARNING]
> A chain only moves forward. Never start an older release or bundle on a home that a
> newer one has run. Never put back an older `data/priv_validator_state.json`. The bundle
> cannot detect every such case.

### Home

The launcher picks the node home in this order:

1. `--home` or `--home=`
2. `$DAEMON_HOME`
3. `$EXRPD_HOME`
4. `$NODE_HOME`
5. `$HOME/.exrpd`

The standalone `exrpd` opens the node's config and data on `$EXRPD_HOME`, else on
`$NODE_HOME`, so the bundle picks the same home for them. There is one difference. With
`$EXRPD_HOME` set, the standalone keeps the keyring in the default home, and the bundle
keeps it in `$EXRPD_HOME`. Use `--home` to get one home for everything.

The launcher resolves the home once. It passes the home as the first argument,
`exrpd --home <node home> ...`, to every command that it runs. So exrpd and Cosmovisor
always use the same home, and no flag of the command can take the home as its value.

### Commands

#### Commands that install

Only `start`, `init` on a new home, and `bundle setup` install the files. Each one:

1. checks the home, and refuses before it writes anything,
2. deletes `<node home>/cosmovisor/bundle.sha256`,
3. installs the files,
4. writes the marker again.

See [Installed files](#installed-files).

#### Other commands

Every other command checks that the marker names this exact bundle, and that the selected
`exrpd` exists. Then it runs that `exrpd`, which the rule in
[How the launcher chooses the binary](#how-the-launcher-chooses-the-binary) selects, as
for `start`. It never installs and never moves `current`, because the node may be running.
For example, after a kill at the halt, `current` still points at the genesis plan. But
`exrpd rollback` runs 12.0.0, because the data file has the plan `v12.0.0`.

If a check fails, the command writes nothing and refuses with one of these messages. Each
message names the home and where it comes from, for example `(from --home)`. Help refuses
in the same way. `version` and TAB completion answer without an error, see below.

| Home | Message |
| --- | --- |
| no `config/genesis.json` and no `data/` | ``No exrpd is installed in <node home> (from <source>). New node: run `exrpd init <moniker> --chain-id <id> --home <node home>`. Existing node: this is probably not its home, pass --home.`` |
| `config/genesis.json` or `data/` exists, no marker | ``The bundle is not installed in <node home> (from <source>). Restart the node service, or run `exrpd bundle setup --home <node home>` while the node is stopped.`` |
| marker of another bundle | ``The bundle installed in <node home> (from <source>) differs from this exrpd. Restart the node service, or run `exrpd bundle setup --home <node home>` while the node is stopped.`` |
| marker of this bundle, the selected exrpd missing | ``The bundle is installed in <node home> (from <source>), but <exrpd path> is missing. Restart the node service, or run `exrpd bundle setup --home <node home>` while the node is stopped.`` |

After you install a new bundle file, restart the service right away. Until the restart,
client commands (`keys`, `tx`, `query`, `status`, ...) refuse with the second or third
message. A bundle built before the marker existed writes no marker. So after the first
update from such a bundle, commands refuse until the first restart.

#### init

`exrpd init ...`, with `init` as the first word, works like this:

- With a matching marker, it runs `exrpd init` of the selected binary.
- On a new home, with no marker and no `config/genesis.json`, it first installs the files
  like `bundle setup`. Then it runs `init` with the genesis binary. `config/` alone does
  not make a home existing, because client commands of the standalone `exrpd` create
  `config/client.toml`.
- Otherwise it refuses: ``<node home> already has config/genesis.json or a different
  bundle installed: restart the node service, or run `exrpd bundle setup --home <node
  home>` while the node is stopped``.

exrpd `init` itself also refuses an existing `config/genesis.json` without `--overwrite`.
For `init --overwrite` on a home without a marker, run
`exrpd bundle setup --home <node home>` first, then run `init` again.

`init --help` and `init -h` are help requests and never install. Combined short flags that
contain `h`, such as `-oh`, are refused before any install. Write `-o -h` instead. A real
help flag anywhere also prevents the install, also when a later `--help=false` turns it
off in exrpd.

#### bundle setup

`exrpd bundle setup [--home <node home>]` checks the home, installs the files and writes
the marker. It never moves `current`, and it starts nothing. With a matching marker, it
only repairs files whose SHA-256 differs. Use it with the node stopped, before maintenance
commands such as `rollback`, `export` or `comet unsafe-reset-all`, on a home whose marker
is missing or differs.

`exrpd bundle`, `exrpd bundle --help` and `exrpd bundle setup --help` print a short usage
and write nothing. `--home` can come anywhere in these commands. Another subcommand or
another argument is an error.

An `exrpd` built without a payload (plain `go build`, for development only) fails every
command with `this exrpd was built without a payload`, these too.

#### start

`exrpd start` does these steps:

1. It writes the home and where it comes from to stderr, for example
   `exrpd: home <node home> (from --home)`. On the first install into a home, when
   `<node home>/cosmovisor` did not exist, it also writes
   `exrpd: installed the bundle in <node home>/cosmovisor`.
2. It checks the home, installs the files and writes the marker.
3. It points `current` at the binary for the data.
4. It replaces itself with Cosmovisor, with the same PID. Cosmovisor then runs
   `exrpd start`.

`start` passes every other argument to `exrpd start`. It refuses `--cosmovisor-config`,
because Cosmovisor would then read its settings from that file and ignore the environment
that `start` sets.

If the node halts at the upgrade height and Cosmovisor does not switch, a restart works.
`start` sees the plan in `data/upgrade-info.json` and points `current` at the upgrade
binary.

#### Help

Help (`exrpd` alone, `exrpd help ...`, or `-h`/`--help` in any command but `start`) never
installs. With a matching marker and an existing selected `exrpd`, it runs the help of
that `exrpd`. Otherwise it refuses like every other command: it writes the message from
the table above to stderr and exits 1. `exrpd version --bundle` shows the home in use and
why.

`exrpd start --help` goes through Cosmovisor like every `start`. It needs the bundle
installed, so it installs the files first. `exrpd help start` does not install.

#### TAB completion

A TAB key sends a completion request (`__complete`). The launcher passes its words to
exrpd exactly as typed, and it adds `--home`, as for every command. It reads `--home` from
the words only to pick the binary, and it never installs. When a command would refuse the
home, with one of the messages above or because the binary rule refuses it, the launcher
answers with no suggestions and the "no file completion" directive (`:4`), and exits 0.
So the shell does not suggest file names.

The launcher also detects a TAB request behind a shell alias that adds `--home` or other
flags, for example `alias exrpd='exrpd --home /var/lib/exrpd'` or
`alias exrpd='exrpd --log_level info'`. Every word before `__complete` must be a flag or
the value of a flag. After a command word, `__complete` is an ordinary argument:
`exrpd keys show __complete` refuses like any command.

`exrpd completion` (bash) and `exrpd completion --zsh` write a script. They are ordinary
commands.

#### version

`exrpd version` prints one line: the version of the `exrpd` installed for the node data,
then the bundle. For example: `12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3`,
or for a single bundle `12.0.1 - genesis 12.0.1 - cosmovisor 1.7.3`.
The first word is the same that the standalone `exrpd version` prints.

The command never installs. It runs the installed binary only when the marker matches and
that binary exists. Otherwise the first word is one of these, and the rest of the line
stays the same:

- `not-installed`: no marker, or the selected `exrpd` is missing.
- `restart-needed`: the marker of another bundle, for example after you install a patch
  bundle and before you restart.

The command reads the installed binary, not the running process. To ask the running node,
use `curl -s localhost:26657/abci_info | jq -r .result.response.version`. The
`node_info.version` of `exrpd status` is the CometBFT version.

#### version --bundle

`exrpd version --bundle` prints JSON. It only reads files and never runs a chain binary,
so it also works on a home where nothing is installed. It never installs. Its fields:

- `version`: a line in the form of `exrpd version`, with `will_run.version` as its first
  word, for example `12.0.0 - genesis 11.3.0 - upgrade 12.0.0 - cosmovisor 1.7.3`.
- `home`: the absolute home.
- `home_source`: where the home comes from: `--home`, `$DAEMON_HOME`, `$EXRPD_HOME`,
  `$NODE_HOME` or `default`.
- `installed_bundle`: the bundle that the marker names: `"this"`, `"none"` (no marker) or
  `"other"`. It reads the marker only. With a marker of this bundle and the selected
  `exrpd` missing, it is `"this"`, `will_run.exists` is `false`, and commands refuse.
- `will_run`: the binary that the next command runs, by the rule in
  [How the launcher chooses the binary](#how-the-launcher-chooses-the-binary). Its fields
  are `dir` (for example `upgrades/v12.0.0`), `binary` (absolute path), `version`,
  `exists` (`true` when the file exists) and `reason` (why that directory). `version` is
  the bundled version when the file is missing or has the bundled SHA-256, else `unknown`.
- `refused`: `null`, or the error that stops every command on this home. Then `will_run`
  is `null`, and the command prints the JSON and exits 1.
- `dirs`: the directories that this bundle uses under `<node home>/cosmovisor`:
  `upgrades/<plan>` of each plan, and the supervisor directory.
- `manifest`: versions, plans, SHA-256, Cosmovisor source and `launcher.commit`. A single
  bundle has no `upgrades` key.

Earlier builds had a top-level `installed` field and `will_run.installed`. Use
`installed_bundle` and `will_run.exists` instead, and `will_run.version` for the version
that runs.

`--bundle` must come after `version` as a bare word. With it, only `--home` is allowed, in
any order. Any other flag, `--help` too, fails with `version --bundle takes no other flag`.
`version --long` and other flags without `--bundle` go to the `exrpd` that runs the data,
like every other command.

#### Argument rules

The launcher does not know which flags of exrpd take a value. So it refuses these forms:

- `start` or `init` after a flag, for example `exrpd --log_level info start`. The error is
  `write the command first: exrpd init|start [flags]`. The launcher cannot tell that the
  command installs, and without Cosmovisor the node would not upgrade.
- A bare `--`. The launcher cannot tell if `--` is the value of the flag before it or the
  end of the flags. With a wrong guess, Cosmovisor and exrpd could use different homes. To
  pass `--` as a value, write it with `=`, for example `--moniker=--`.
- `--home` right after another flag without `=`, for example `--moniker --home=/srv/name`.
  exrpd can read the second word as the value of the first flag. Write the other flag with
  `=`, for example `--moniker=--home=/srv/name` or `--yes=true`, or put `--home` first.

### Installed files

Installed binaries have mode `0755`, and plan records have mode `0644`. The launcher
writes all of them atomically. On every `start`, on `init` on a new home and on every
`bundle setup`:

- A file with the same SHA-256 stays as it is. If its owner cannot execute it, for
  example a copy restored with mode `0644`, the launcher sets mode `0755` again. So a
  reinstall of the same bundle gives identical files.
- A file with different content is replaced atomically. The launcher writes a line to
  stderr with the file and both SHA-256 values. A patch release bundle replaces the
  binaries of an earlier bundle in this way.
- An older bundle started on the home, for example one that a service file still names,
  puts its own binaries back, and its own marker if it writes one. Check the service log
  for `exrpd: replaced`.

#### The marker

`<node home>/cosmovisor/bundle.sha256` marks the installed bundle. It has the SHA-256 of
the bundle's manifest in lowercase hex, and a newline. The manifest lists every binary
with its SHA-256 and size. So any change in the manifest counts as another bundle, also
when the binary that a command runs did not change. Each install writes in this order:

1. Delete the marker, before the first binary is written.
2. Install all files as above.
3. Write the marker atomically, only after every file is installed.

An install that stops part way leaves no marker. Then the commands that do not install
refuse until `start` or `bundle setup` completes. These commands compare the marker with
the manifest of the bundle `exrpd` that you run, and they check that the selected binary
exists. They do not hash the binaries: when the marker matches, they trust the files. A
binary replaced by hand is the operator's responsibility. `start` and `bundle setup`
always check every SHA-256 and repair.

#### Directories and symlinks

- The bundle manages the directories under `<node home>/cosmovisor`. Do not replace them
  or their `bin/` with symlinks. When two plan directories of the bundle are one
  directory, for example `upgrades/v11.3.0` made a symlink to `upgrades/v12.0.0`,
  `exrpd start`, `init` and `bundle setup` stop before they write. The launcher does not
  check other symlinks there.
- Do not copy the bundle `exrpd` into a directory under `cosmovisor/`. It is not a chain
  binary. With stock Cosmovisor, `bin/exrpd` is often a symlink to `/usr/local/bin/exrpd`,
  which is the bundle after you install it there. A command that would run it stops with
  `... is this exrpd bundle, not a chain binary`. `exrpd start` and `exrpd bundle setup`
  replace it with the chain binary.

#### Take over a stock Cosmovisor home

The bundle can take over a home that stock Cosmovisor set up, as it is: stop Cosmovisor,
then start the bundle with the same home. The operator guide recommends a clean move
instead, see [Move from Cosmovisor](bundle.md#move-from-cosmovisor). On a takeover:

- The bundle replaces the `upgrades/v11.3.0` and `upgrades/v12.0.0` binaries, also an
  operator's own build of 12.0.0. `current` moves by the rule in
  [How the launcher chooses the binary](#how-the-launcher-chooses-the-binary).
- The bundle ignores a `cosmovisor/config.toml` from `cosmovisor init`. Cosmovisor reads
  it only with `--cosmovisor-config`, which the bundle refuses.
- Remove `DAEMON_ALLOW_DOWNLOAD_BINARIES=true` from an old service file. `start` refuses
  it.
- A `data/upgrade-info.json` that `cosmovisor add-upgrade --upgrade-height` wrote before
  the upgrade height makes the bundle start 12.0.0 too early, and exrpd stops. Delete it
  while the chain is below that height.
- Every command refuses while `data/upgrade-info.json.batch` exists. See
  [Batch upgrades](#batch-upgrades).
- The bundle does not use `genesis/`. It stays on disk, as do other directories under
  `upgrades/` and `supervisor/`. See [Removing old files](#removing-old-files).

#### Partial temporary files

An install that is stopped or killed while it writes a file (SIGTERM, SIGKILL, OOM, power
loss) leaves a partial temporary file named `.exrpd-bundle-<random>` next to that file.
The installed files stay complete, but the launcher never deletes these leftovers. With
the node stopped and no other `exrpd` command running, delete every `.exrpd-bundle-*`
file under `<node home>/cosmovisor`.

### How the launcher chooses the binary

The operator guide explains this rule in short in [Restart](bundle.md#restart).

At startup, Cosmovisor compares the plan in `<node home>/data/upgrade-info.json` with the
plan in `current/upgrade-info.json`. When they differ, it switches to
`upgrades/<data plan>`. x/upgrade writes the data file when the running binary halts at a
plan height, so the file has the last plan that the node halted at (see
[Stale data file](#stale-data-file)). A binary runs only the data of its own plan. Patch
releases keep the plan of their release.

Every plan of the bundle has its directory `upgrades/<plan>`, the genesis plan too. So the
data file always names a directory of the bundle, and every directory of the bundle is a
plan directory with a record. A single bundle has one plan, its genesis plan. Before
anything changes, the launcher decides:

| Home | Binary | `start` |
| --- | --- | --- |
| `data/upgrade-info.json.batch` exists, any content. Checked before every other row | none | stops with an error, nothing changes. See batch upgrades in [Batch upgrades](#batch-upgrades) |
| `data/upgrade-info.json` names a plan of the bundle (`v11.3.0` or `v12.0.0`, any case) | `upgrades/<plan>` | points `current` there and copies the data file as its record |
| `data/upgrade-info.json` names another plan (for example `v12.1.0`) | none | stops with an error, nothing changes. A plan older than the genesis plan (for example `v11.2.0`) gets a message that the file may be stale. In a single bundle, the message also names the two-binary bundle, see [Single bundle on data before its plan](#single-bundle-on-data-before-its-plan) |
| `data/upgrade-info.json` is corrupted: not JSON, no name, no height above 0, a time, an upgraded client state, or a field of the wrong type, as exrpd and Cosmovisor check it | none | stops with an error, nothing changes. An empty file is read again for up to 1 s first, because x/upgrade truncates the file before it writes it |
| no data file, no `current` | `upgrades/<genesis plan>` | links `current` there and writes the record `{"name":"<genesis plan>"}` |
| no data file, `current` names a directory of the bundle, also when it was deleted | that directory | keeps `current`. Install creates a deleted directory again, with its record |
| no data file, `current` on `genesis`, as stock Cosmovisor links it first | the plan directory of the bundle that `genesis` resolves to, else `upgrades/<genesis plan>` | moves `current` there and writes its record. An old `genesis/` stays on disk |
| no data file, `current` on another directory whose record names a plan of the bundle | `upgrades/<plan>` | moves `current` there and copies the record |
| no data file, `current` on another directory named `upgrades/<plan>` without a record, plan of the bundle | `upgrades/<plan>` | moves `current` there and writes the record |
| no data file, `current` on any other directory, deleted or not | none | stops with an error, nothing changes |

The last row stops, because a chain only moves forward. An unknown directory most likely
belongs to a later plan, and genesis is the oldest binary of the bundle. The data file
must have a height, as x/upgrade and Cosmovisor require. A record in a binary directory
needs only a plan name, as Cosmovisor writes it.

The launcher also recognizes missing bundle directories through a symlinked home, or
through an absolute target that uses a home alias. It resolves the existing prefixes on
both sides before it compares the remaining path.

Each move of `current` writes one line to stderr, so the service log shows why the node
runs a binary. For example:

```text
exrpd: moved <node home>/cosmovisor/current from upgrades/v11.3.0 to upgrades/v12.0.0: data/upgrade-info.json names v12.0.0
exrpd: moved <node home>/cosmovisor/current from genesis to upgrades/v11.3.0: current points at genesis: genesis plan v11.3.0
```

The second line is from a stock Cosmovisor home without a data file.

## Service

### systemd

`exrpd start` replaces itself with Cosmovisor, so the service of the
[systemd unit](bundle.md#systemd-unit) has two processes: Cosmovisor as
`MainPID`, and the node `exrpd` as its child. Cosmovisor runs each `exrpd` version as a
separate child process, so the PID of `exrpd` changes at the switch.

An `exrpd` can outlive Cosmovisor when Cosmovisor exits without forwarding a stop:

- SIGKILL, SIGHUP or another signal that Cosmovisor does not handle,
- an out-of-memory kill,
- an upstream panic while Cosmovisor parses a malformed non-empty
  `data/upgrade-info.json`.

With the default `KillMode=control-group`, systemd stops the rest of the unit when
Cosmovisor exits, so nothing keeps running.

### macOS (launchd)

Use a LaunchDaemon with these settings:

- `ProgramArguments`: `/usr/local/bin/exrpd`, `start`, `--home`, `<node home>`. Give the
  absolute path, because launchd does not search `/usr/local/bin`.
- `KeepAlive`: a dictionary with `SuccessfulExit` = `false`.

launchd sends SIGTERM to the PID of the job, which is Cosmovisor, and Cosmovisor forwards
it to `exrpd`. When the job exits, launchd kills the remaining processes of its process
group, unless `AbandonProcessGroup` is `true`. Keep the default `false`.

### Without a service manager

Without a service manager, nothing stops an `exrpd` that outlives Cosmovisor (see
[systemd](#systemd) for the causes). Start the node in its own session, and stop it by
process group:

```sh
setsid exrpd start --home <node home> &
ps -o pgid= -p <cosmovisor pid>          # find the process group
kill -TERM -- -<pgid>                    # stop the node
```

Do not signal the group of a node that you did not start with `setsid`. That group can
contain your shell. Check for a leftover `exrpd` before you start again. A second node on
the same home fails on the database lock, so it does not double sign.

## Node settings

### RPC server

At the upgrade height, Cosmovisor reads the chain height with
`exrpd status --home <node home>`, without `--node`. So the CometBFT RPC server
(`[rpc] laddr` in `config/config.toml`) must stay on, and `node` in
`<node home>/config/client.toml` must point at it. If `exrpd status` fails, Cosmovisor
treats it as "not yet at the height" and does not switch.

### Working directory

Cosmovisor changes its working directory to `<node home>/cosmovisor` and starts exrpd
there. File paths in `config/config.toml` resolve against the home, so they are not
affected. These settings resolve against the working directory:

- `start` flags such as `--trace-store`, and the same settings from `EXRPD_*` variables
  or `config/app.toml`
- `[tls] certificate-path` and `key-path` in `config/app.toml`
- `unix://` addresses in `config/config.toml`, such as `priv_validator_laddr` and
  `[rpc] laddr`

A relative value that worked with the standalone exrpd has another meaning under the
bundle:

- A path whose directory does not exist under `<node home>/cosmovisor` stops `start`, for
  example `open profiles/node.kv: no such file or directory` or
  `listen unix sockets/signer.sock: bind: no such file or directory`.
- A bare file name such as `--trace-store node.kv` gives no error, and exrpd uses
  `<node home>/cosmovisor/node.kv`.
- A relative TLS path only logs `failed to start HTTP server for WS`, and the node runs
  without the EVM WebSocket endpoint.

Use absolute paths.

## At the upgrade height

### Stop during the grace wait

At the upgrade height, Cosmovisor waits up to about 1 s before it forwards a stop. If a
stop arrives then, it exits 0 and does not start 12.0.0, also if exrpd has exited by
then. The next start runs 12.0.0.

### Failed pre-upgrade script

Upstream, a failed `COSMOVISOR_CUSTOM_PREUPGRADE` blocks the upgrade. Here, the next
`start` sees the data file with the plan `v12.0.0` and selects 12.0.0 without the script.
The same happens after a stop or a failed copy during the Cosmovisor backup.

### Cosmovisor backup

With the default `UNSAFE_SKIP_BACKUP=true`, nothing copies `data/`. The upgrade migrates
the stores at the upgrade height, and 11.3.0 cannot open them after that. To point
`current` back or to install another binary does not restore the data.

To keep the upstream backup, set `UNSAFE_SKIP_BACKUP=false`, and optionally
`DAEMON_DATA_BACKUP_DIR`. Then:

- The copy needs free disk equal to the size of `data/`.
- The validator does not sign until the copy finishes.
- Do not stop the node during the copy. Cosmovisor exits at once and leaves a partial
  `data-backup-<date>` directory.
- After a stop, a failed copy or a failed `COSMOVISOR_CUSTOM_PREUPGRADE`, the next `start`
  points `current` at 12.0.0 itself. Cosmovisor does not run the backup or
  `COSMOVISOR_CUSTOM_PREUPGRADE` again. Delete the partial backup, and take your own
  snapshot before you start again if you need one.

> [!WARNING]
> If you restore a backup, keep the newer `data/priv_validator_state.json`. An older copy
> lets the validator sign heights that it already signed. That is double signing.

## Restart, reinstall and restore

### Restart

The node can be killed at the halt before Cosmovisor switched, for example by power loss,
SIGKILL or OOM. Or `<node home>/cosmovisor` can be deleted after the upgrade, on a home
whose `data/upgrade-info.json` has the plan `v12.0.0`. In both cases, the next start runs
12.0.0. It does not run the Cosmovisor backup, `COSMOVISOR_CUSTOM_PREUPGRADE`, or the
optional `exrpd pre-upgrade` hook of the upgrade (12.0.0 does not implement the hook).

### Install a new bundle file

A patch bundle reaches validators with a reinstall and a restart, before or after the
switch. `start` replaces `upgrades/v12.0.0/bin/exrpd` and logs it. Restart right after the
reinstall, because until then the other commands refuse: the installed bundle differs.
See [Releases and bundles](#releases-and-bundles) for how patch bundles are made.

### Removing old files

The launcher never deletes a directory. Each bundle adds about 100 MB or more under
`<node home>/cosmovisor`. After you install a new bundle, remove what it does not use:

1. Start the new bundle once, check that the node runs, then stop it. This moves
   `current` off the directories of the old bundle.
2. As the node user, run `exrpd version --bundle --home <node home>`. Its `dirs` list
   names every directory that this bundle uses. If the command fails, delete nothing.
3. Delete only these:
   - the subdirectories of `<node home>/cosmovisor/upgrades/` and
     `<node home>/cosmovisor/supervisor/` that the list does not name,
   - an old `<node home>/cosmovisor/genesis/`, because the bundle does not use it.

   Never delete the directory that `current` points at. Do not touch anything else in
   `<node home>/cosmovisor`: `current`, and your own files, such as a
   `COSMOVISOR_CUSTOM_PREUPGRADE` script.
4. Start the node.

### Resync after the upgrade

Prepare the resync as you normally do: configure state sync in `config/config.toml`, or
get a snapshot. The bundle adds one thing: keep `data/upgrade-info.json` together with
`priv_validator_state.json`. The data file makes the next start pick 12.0.0, also when
`current` did not move yet, for example after a crash at the halt. After the state sync,
exrpd ignores the height in the file, because the node starts far past that height.

Run the commands one by one. If a command fails, stop there and do not start the node.

0. Find the node home. It is the directory that your service runs with. Look in your
   service file (`systemctl cat <unit>`):
   - If the `ExecStart` line has `--home`, that is the home.
   - If an `Environment=` line sets `DAEMON_HOME`, `EXRPD_HOME` or `NODE_HOME`, that is
     the home.
   - Otherwise it is `.exrpd` in the home directory of the user that the service runs as,
     for example `/home/<node user>/.exrpd`.

   Set it and check it. Do not write `~` in the commands below, because your shell
   expands it to your own home directory, not the one of the node user.

   ```sh
   home=<node home>
   sudo -u <node user> ls "$home/data/priv_validator_state.json"
   ```

1. Stop the node: `sudo systemctl stop <unit>`
2. Create the backup directory: `sudo -u <node user> mkdir "$home/resync-backup"`. If it
   exists already, stop: an earlier resync did not finish. Do not overwrite or delete a
   file in `data/` or in `resync-backup/` until you know what that resync did. The
   signing state has a height, a round, a step and the last signature, so two files with
   the same height are not the same.
3. Save the signing state:
   `sudo -u <node user> cp "$home/data/priv_validator_state.json" "$home/resync-backup/"`
4. Save the data file:
   `sudo -u <node user> cp "$home/data/upgrade-info.json" "$home/resync-backup/"`. If it
   fails with "No such file", the node has no data file: skip step 8.
5. Check the bundle in the home:
   `sudo -u <node user> /usr/local/bin/exrpd version --home "$home"`. If the line starts
   with `restart-needed` or `not-installed`, for example because you installed a new
   bundle file since the node last started, install it now, with the node stopped:
   `sudo -u <node user> /usr/local/bin/exrpd bundle setup --home "$home"`. Then reset the
   data:
   `sudo -u <node user> /usr/local/bin/exrpd comet unsafe-reset-all --home "$home" --keep-addr-book`
6. If you use a snapshot, unpack it into the home now, so that `data/` is filled. The
   snapshot must be from after the upgrade height. With an older snapshot, the node must
   run 11.3.0 up to the upgrade: skip step 8, and delete `$home/cosmovisor/current`, as in
   [Restore a snapshot from before the upgrade](bundle.md#restore-a-snapshot-from-before-the-upgrade).
   Otherwise 12.0.0 stops at once on the older data. With state sync, skip this step: the
   node gets the state when it starts.
7. Put the signing state back:
   `sudo -u <node user> cp "$home/resync-backup/priv_validator_state.json" "$home/data/"`
8. Put the data file back, only if step 4 saved it:
   `sudo -u <node user> cp "$home/resync-backup/upgrade-info.json" "$home/data/"`
9. Start the node: `sudo systemctl start <unit>`
10. When the node runs and signs again, delete the backup:
    `sudo -u <node user> rm -r "$home/resync-backup"`

The backup is in the home, outside `data/`, because the reset deletes `data/`. The
commands run as the node user, because `data/` has mode `0700`. The copied
`priv_validator_state.json` keeps its mode `0600`, so another user cannot read it.

### Join after the upgrade

A node that joins with a state sync from after the upgrade height has no
`data/upgrade-info.json`, and a snapshot can have one or not. Without the file, a two-binary
bundle on a new home would run 11.3.0 on 12.0 data. The
single bundle `exrpd-bundle_12.0.x` runs its only plan, `v12.0.0`, on a home without a
data file, so a new node uses it. Before the height of the next plan, install the next
two-binary bundle. Its genesis is 12.0.x, so it runs that data with no extra step. This
also works on a home that an earlier bundle used.

The same applies when an existing node restores a snapshot from after the upgrade: use
the single bundle. A two-binary bundle also works on a home where `current` still points
at `upgrades/v12.0.0`, because it keeps that link, as after a resync.

A single bundle cannot detect a snapshot from **before** the upgrade height on a home
without a data file and without `current`. It runs 12.0.x on that data, and exrpd stops.

#### Single bundle on data before its plan

A single bundle refuses a home that is clearly from before its plan, and changes nothing:

- a data file with a plan older than its genesis plan, for example `v11.3.0`,
- `current` on a directory whose record has such a plan, for example `upgrades/v11.3.0`.

The message says that this exrpd is a single bundle, that it runs only plan `v12.0.0`,
and that a node that has not reached that upgrade needs the two-binary bundle
`exrpd-bundle_<genesis>-<version>` of the release. Because a data file can also be stale
(see [Stale data file](#stale-data-file)), it also says what to write if the node already
runs `v12.0.0` or later. A data file with a newer plan, for example `v12.1.0`, gets the
unknown-plan refusal of every bundle.

### Restore a snapshot from before the upgrade

Before you restore, obtain the two-binary bundle for the required patch. If it is missing,
ask the maintainers to publish it. They can dispatch [Bundle](#bundle) on `main`, for
example with `tag: v12.0.1` and `genesis_tag: v11.3.0` for
`exrpd-bundle_11.3.0-12.0.1`. The workflow keeps the existing single bundle.

With the node stopped, delete `<node home>/cosmovisor/current`, and delete
`data/upgrade-info.json` if the restore kept one with the plan `v12.0.0`. The next start
then runs the binary of the genesis plan for the data. A single bundle has no binary for
that data, so install the two-binary bundle first.

### Problems with the data file

#### Corrupted data file

x/upgrade writes `data/upgrade-info.json` without fsync, so a crash or a disk fault can
leave it empty or cut short. This risk exists with every setup: exrpd and stock Cosmovisor
also stop on such a file. Every command stops with an error before it changes any file.

The data can be at the upgrade height or past it, so do not only delete the file. Write
it again with the exact plan name and height of the last upgrade that the chain reached,
from the governance proposal, for example `{"name":"v12.0.0","height":4521000}`. exrpd
uses both values to apply its store changes at that height.

#### Stale data file

x/upgrade writes the file only when the running binary halts at a plan. A node that
applied a plan with a binary swapped in before the plan height (for example with
`--halt-height`) keeps the file of an earlier plan.

A file with a plan older than the genesis plan of the bundle stops every command with a
message that says so. When you move a standalone home to the bundle, check that the file
has the last plan that this node passed. If it does not, write it again with that plan
and its height. A stale file with a plan that the bundle carries, for example `v11.3.0` on
a node that already swapped to 12.0.x, selects the older binary, and exrpd stops on the
newer data.

#### Plan that the bundle does not carry

Install the bundle for the next plan of the chain before its height, and install bundles
in order: start a bundle only on data that reached its genesis plan. A bundle started on
data whose `data/upgrade-info.json` has a plan that the bundle does not carry stops with
an error before it changes any file.

#### Batch upgrades

Cosmovisor batch upgrades (`data/upgrade-info.json.batch`) and
`cosmovisor add-upgrade --upgrade-height` both write `data/upgrade-info.json` before the
plan height. `start` reads a file with the plan `v12.0.0` as an upgrade that the chain
already reached, so 12.0.0 would start too early. The governance upgrade needs neither.

Every command refuses while `data/upgrade-info.json.batch` exists, with any content, `[]`
too. Under the bundle, Cosmovisor still runs the batch watcher, which writes the data file
before the height, and after a restart the bundle would trust that file. Delete the batch
file. If the node is below the upgrade height, also delete a `data/upgrade-info.json` with
the plan of that upgrade, because the batch watcher wrote it early.

A file from `add-upgrade --upgrade-height` looks the same as a file that the chain wrote,
so the launcher cannot refuse it.

## Settings

- An empty value counts as unset for the variables that this bundle pins or gives a
  default. All other variables go to Cosmovisor unchanged. For example, an empty
  `COSMOVISOR_TIMEFORMAT_LOGS` removes the log timestamps. Unset it to keep the default.
- Cosmovisor refuses a `DAEMON_SHUTDOWN_GRACE` of `0`, so the upstream default (kill
  exrpd at once) is not available.
- All other Cosmovisor variables keep their upstream meaning, for example
  `DAEMON_POLL_INTERVAL`, `DAEMON_RESTART_DELAY`, `COSMOVISOR_COLOR_LOGS` and
  `COSMOVISOR_TIMEFORMAT_LOGS`.
- For batch upgrades, see [Batch upgrades](#batch-upgrades). For a failed
  `COSMOVISOR_CUSTOM_PREUPGRADE` and a stop during the grace wait, see
  [At the upgrade height](#at-the-upgrade-height).

## Known limitations

A node that replays the chain from block 1 needs the binary of every plan. A released
bundle carries two, so it does not support a sync from genesis. Archive nodes set up stock
Cosmovisor with the standalone releases, or use a bundle built by hand with more
upgrades.

### Sync from genesis with a bundle built by hand

The `upgrades` list of the manifest can hold more than one upgrade, in any order. Each
upgrade plan must come after the genesis plan, and no plan can appear twice. Such a bundle
installs one `upgrades/<plan>/` for each plan, and Cosmovisor switches at each plan height
as usual. A restart picks the binary of the plan in `data/upgrade-info.json`, as for a
two-binary bundle. The scripts do not build these bundles. Build one by hand: put the
gzip-compressed binaries and a `manifest.json` with all the upgrades in
`cmd/exrpd-bundle/payload/`, as `build.sh` does for one upgrade, then run
`cd cmd/exrpd-bundle && GOWORK=off CGO_ENABLED=0 go build -o <out> .`.

- Build one bundle for each network. Each plan needs the exact release that the network
  ran at that height, and the networks ran different releases, for example mainnet
  v8.0.2 and testnet v8.0.0. Take the releases from the
  [Networks](https://docs.xrplevm.org/pages/operators/resources/networks) page, and the plan
  names from `UpgradeName` in `app/upgrades` of each release tag.
- Nobody can change the config at an automatic switch. If a release needs a config change,
  end the bundle before its plan. The node then halts at that plan: Cosmovisor finds no
  binary for it and exits with an error, and each restart stops on the unknown plan.
  Make the change, then install the next bundle, whose genesis plan is that plan. For
  example: bundle A carries v7.0.0 to v9.0.3, you set `evm-chain-id` at the `v10.0.0`
  halt, and bundle B carries v10.0.2 to the current release.


## Payload

A two-binary bundle `exrpd` carries three executables:

- The genesis `exrpd` has the exact bytes of the GitHub release `v11.3.0` archive. The
  build checks it against the `checksums.txt` of that release.
- The upgrade `exrpd` has the same bytes as the standalone 12.0.0 release binary.
- Cosmovisor is built from `cosmossdk.io/tools/cosmovisor@v1.7.3` with one patch. See
  [Cosmovisor patch](#cosmovisor-patch).

A single bundle carries two: its release as the genesis `exrpd`, with the exact bytes of
that release's archive, and Cosmovisor.

The bundle stores the executables gzip-compressed. So a two-binary bundle is about 100 MB
on linux/amd64, although it carries two full `exrpd` builds and Cosmovisor. The SHA-256 and
size in the manifest are those of the uncompressed executables, and the install checks
them. The genesis binary is installed once, in `upgrades/<genesis plan>/`. Writing the
files never changes `current`. The launcher sets it in a separate step, see
[How the launcher chooses the binary](#how-the-launcher-chooses-the-binary).

### Manifest and plans

The manifest lists the genesis release with its plan (`genesis`), the upgrades with their
plans (`upgrades`), Cosmovisor (`cosmovisor`), and the commit of the code that built the
launcher (`launcher.commit`, with `-dirty` when `cmd/exrpd-bundle/` had uncommitted
changes). A released bundle has one upgrade. A single bundle has no `upgrades` key. Every
upgrade plan comes after the genesis plan, in any order. At load time, the launcher
validates the plans and orders releases once: genesis first, then upgrades by number.
Messages and the version line use that order, so `v12.0.0` comes after `v2.0.0`.
`exrpd version --bundle` prints the original manifest, including its upgrade order and
the commit of the launcher code. The bundle digest hashes the original manifest bytes,
not the ordered releases.

The genesis plan is the newest `UpgradeName` in `app/upgrades` of the genesis release tag
(`plan.sh release`). It does not come from the version: v10.2.x runs plan `v10.1.0`.

Plans are lower-case `vX.Y.Z` names, because Cosmovisor lowercases the plan that it reads
from the data file and switches to `upgrades/<that name>`. `build.sh` and `release.sh`
refuse any other name.

`plan.sh release` refuses a genesis tag in two cases, because a skip could pick the binary
of an older plan:

- The tag also has a plan with another name that is newer than its `vX.Y.Z` plan, such as
  `v10.2.0-rc1` next to `v10.1.0`.
- The tag has no `vX.Y.Z` plan.

It ignores a name that is not newer, such as `v8.0.0-rc.1` next to `v8.0.0`.

## Cosmovisor patch

### The defect

Cosmovisor v1.7.3 registers a SIGTERM/SIGQUIT handler on every run, and it never removes
the handler after an upgrade. After the in-process switch, a SIGTERM sent only to the PID
of Cosmovisor also reaches the old handler. That handler cannot signal the old app, and it
calls `os.Exit(1)`. Cosmovisor then exits, and 12.0.0 keeps running outside the view of
the service manager.

### The fix

`scripts/bundle/cosmovisor/cosmovisor-v1.7.3-signal.patch` is the fix proposed
upstream in [cosmos/cosmos-sdk#26879](https://github.com/cosmos/cosmos-sdk/pull/26879).
It has three parts from that PR:

1. It changes only `Launcher.Run`. The handler is registered before the app starts, and
   it stops on every return path. `TestLaunchProcessSignalAfterUpgrade` fails on
   unpatched v1.7.3.
2. Commit `4513d133` also exits 0 when a stop arrives after the app has already exited,
   as with the default systemd stop. During the backup or the pre-upgrade, it still exits
   at once. It does not ignore the stop and start the upgraded app. `TestLaunchProcessStop`
   covers both cases.
3. Commit `2418d78c` honors a stop that arrives while the app shuts down at the upgrade
   height (the `DAEMON_SHUTDOWN_GRACE` wait). Cosmovisor exits 0 without the backup, the
   pre-upgrade or the binary switch. So `current` stays on the old binary, and the next
   start applies the upgrade. `TestLaunchProcessStop/upgrade-halt` covers it. The patch
   also makes the test fixture trap `INT`, because v1.7.3 sends `os.Interrupt` at the
   grace wait, and upstream `main` sends `SIGTERM`.

Remove the patch when a Cosmovisor release contains the fix.

### How the build applies it

1. It fetches the module at v1.7.3 and checks its `go.sum` hash, from
   `cosmovisor/cosmovisor.env`.
2. It runs `go mod verify`, so the extracted module source and its dependencies match
   their hashes.
3. It checks the patch with `git apply --check`, then applies it.
4. It builds with the own `go.mod` of Cosmovisor (`CGO_ENABLED=0`).

## Building

These files build, test and release the bundle:

| Path | Content |
| --- | --- |
| `cmd/exrpd-bundle/` | the launcher, a separate Go module. It builds the bundle executable `exrpd` |
| `scripts/bundle/plan.sh` | prints a plan or a version: `checkout` the plan of the one handler in `app/upgrades`, `release X.Y.Z` the newest plan in `app/upgrades` of the release tag (it fetches the tag if needed), `genesis PLAN` the newest published release below the plan |
| `scripts/bundle/lib.sh` | helpers sourced by the scripts: archive checks and the plan checks (`release_plan`, `resolve_plans`) |
| `scripts/bundle/publication.sh` | remote-state checks shared by the workflow's planning and publishing jobs |
| `scripts/bundle/check.sh` | runs the checks of CI: launcher `go vet`, `go test` and `golangci-lint`, then `cosmovisor.sh --test` |
| `scripts/bundle/cosmovisor/` | Cosmovisor pins (`cosmovisor.env`) and the patch |
| `scripts/bundle/build.sh` | builds one bundle executable, two-binary or single |
| `scripts/bundle/cosmovisor.sh` | builds or tests the patched Cosmovisor |
| `scripts/bundle/release.sh` | builds one bundle set (every platform and the checksums file) from published archives |
| `scripts/bundle/e2e.sh`, `scripts/bundle/e2e/` | localnet chain upgrade on linux/amd64, with the two-binary and the single bundle. `e2e.sh` sets up the run, `e2e/` holds the helpers and one file for each scenario |
| `scripts/bundle/image-smoke.sh` | smoke test of a bundle image |
| `Dockerfile.bundle` | the bundle image. The release image builds do not read it |
| `.github/workflows/bundle.yml` | "Bundle" workflow: builds the missing bundle sets and images of a published release |

### Build and check a bundle

```sh
mkdir -p release12 && tar -xzf node_12.0.0_Linux_amd64.tar.gz -C release12 bin/exrpd
# two-binary bundle 11.3.0 + 12.0.0
scripts/bundle/build.sh \
  --genesis-version 11.3.0 --genesis-archive node_11.3.0_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
  --upgrade-exrpd release12/bin/exrpd --upgrade-version 12.0.0 --plan v12.0.0 \
  --goos linux --goarch amd64 --out dist/exrpd
# single bundle 12.0.0: only the genesis flags
scripts/bundle/build.sh \
  --genesis-version 12.0.0 --genesis-archive node_12.0.0_Linux_amd64.tar.gz --genesis-checksums checksums-12.0.0.txt \
  --goos linux --goarch amd64 --out dist/exrpd-single
scripts/bundle/check.sh   # launcher vet, test and lint, then the patched Cosmovisor test suite
```

- `--upgrade-exrpd` is the `bin/exrpd` of the standalone 12.0.0 release archive, not the
  `bin/exrpd` that `make build` writes. A local build has other bytes. A home that installs
  it logs a replacement each time it moves between that build and a released bundle.
- With `--upgrade-exrpd` and `--upgrade-version`, `build.sh` builds a two-binary bundle.
  `--plan` is its upgrade plan, by default the handler of this checkout (`plan.sh checkout`).
  Without both flags, it builds a single bundle. Only one of the two flags, or `--plan`
  in single mode, is an error.
- The genesis plan comes from the genesis release tag (`plan.sh release`).
  `--genesis-plan vX.Y.Z` sets it instead.
- The manifest gets `launcher.commit` from `git rev-parse HEAD`, with `-dirty` when
  `cmd/exrpd-bundle/` has uncommitted changes.
- `release.sh` builds one set for each call: without `--genesis-version` the single set
  `exrpd-bundle_<version>_*`, with it the two-binary set
  `exrpd-bundle_<genesis>-<version>_*`. Both plans default to `plan.sh release` of their
  tags. `--out-dir` must be absent or empty, so that no old asset stays without a
  checksum entry.

### End-to-end test

`e2e.sh` runs a real chain upgrade on a localnet, on linux/amd64:

```sh
# the normal run
scripts/bundle/e2e.sh --genesis-archive node_11.3.0_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
  --upgrade-exrpd release12/bin/exrpd
# the same run with the unpatched v1.7.3 release binary, must show the defect
scripts/bundle/e2e.sh --genesis-archive node_11.3.0_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
  --upgrade-exrpd release12/bin/exrpd --unpatched --cosmovisor-archive cosmovisor-v1.7.3-linux-amd64.tar.gz \
  --cosmovisor-checksums SHA256SUMS-cosmovisor-v1.7.3.txt
```

`e2e.sh` does not download release assets. Give their paths with these flags:

| Flag | Content |
| --- | --- |
| `--genesis-archive` | required: `node_<genesis version>_Linux_amd64.tar.gz` of the genesis GitHub release, for example `v11.3.0`. Its name gives the genesis version. `plan.sh genesis <plan>` prints the default genesis |
| `--genesis-checksums` | required: `checksums.txt` of the genesis release |
| `--upgrade-exrpd` | required: standalone upgrade `exrpd` for linux/amd64, executable |
| `--upgrade-version 12.0.1` | the version of `--upgrade-exrpd`, by default the plan version (12.0.0). Set it to test a patch bundle |
| `--plan v12.0.0` | the upgrade plan, by default the handler of this checkout (`plan.sh checkout`). The "Bundle" workflow passes the plan of the release tag |
| `--bundle <path>` | tests a two-binary bundle `exrpd` that was already built, for example extracted from a release asset. The run checks that its manifest names the plans and the SHA-256 of both `exrpd` files, and does not build the tested bundle. It still builds the earlier patch bundle, see below. Not with `--unpatched` |
| `--single-bundle <path>` | tests a single bundle `exrpd` that was already built. Its genesis must be the `--upgrade-exrpd` bytes, and its plan `--plan`. Not with `--unpatched` |
| `--unpatched` | embeds the upstream Cosmovisor v1.7.3 release binary instead of the patched build, see below |
| `--cosmovisor-archive`, `--cosmovisor-checksums` | required with `--unpatched`, and only with it: the upstream Cosmovisor release archive `cosmovisor-v1.7.3-linux-amd64.tar.gz` and a checksum list that contains its line, such as `SHA256SUMS-cosmovisor-v1.7.3.txt` |

`E2E_DIR=<dir>` sets the scratch directory.

#### Before the run

- The script stops with an error that names the first missing file.
- The build still needs Go modules. Without `--unpatched`, `cosmovisor.sh` runs
  `go mod download` for the pinned Cosmovisor module and its dependencies. The launcher
  module uses only the standard library. Run the script with network access, or with a Go
  module cache that already has these modules.
- Two runs cannot use the same `E2E_DIR` at the same time. The script holds the lock
  `<E2E_DIR>.lock`, and it stops when another run holds it.
- The script deletes `E2E_DIR` at the start. It accepts only a directory that is missing,
  empty, or marked with `.e2e-scratch` by an earlier run.

#### What the run checks

The homes run sequentially on one shared port range. The pre-upgrade copies use the
`PRE_H` checkpoint; resync copies `main` at `MAIN_H`, and the single-bundle copies
use `main` at `SINGLE_H`. Each `use_home NAME BUNDLE` selects the home, log, and launcher
together. The tested (`TB`), earlier-patch (`PB`), and single (`SB`) bundle files stay
unchanged. Only the reinstall and single-bundle replacement scenarios replace a separate
executable at `bin/exrpd`; later scenarios select `TB` or `SB` directly.

- **Main home.** The run creates it with `exrpd init` through the bundle, which installs
  the files and writes the marker. Before that, it checks that help, client commands, TAB
  completion and `version` install nothing on the empty home. The first start, without a
  data file, links `current` to the genesis plan.
- **Patch bundle.** The run always builds an earlier patch bundle from this checkout with
  `build.sh`. It has the same genesis `exrpd` and the installed Cosmovisor of the tested
  bundle, and an upgrade `exrpd` with text appended, so that its manifest differs. After
  the upgrade, the run installs this bundle on the main home again. Then it installs the
  tested bundle over it while the node runs. Until the restart, client commands must
  refuse with "differs", and `version` must start with `restart-needed`. The earlier
  bundle sets up the resync home, and `bundle setup` of the tested bundle must install it
  without a move of `current`.
- **`stock` scenario.** The run sets up and runs a home with stock Cosmovisor, built with
  `go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.7.3`. Then the bundle takes
  it over before the upgrade height, and the run checks the switch. Before the takeover,
  client commands and `init` refuse, `version` starts with `not-installed`, and
  `version --bundle` reports `refused: null` and `installed_bundle: "none"`.
- **`snapshot` scenario.** The run makes a copy of the home from before the upgrade
  height, as a stock home restored from a snapshot: no data file, and `current` linked to
  `genesis`. `start` moves `current` to the genesis plan, and the node switches at the
  upgrade height.
- **Single bundle.** Without `--single-bundle`, the run builds a single bundle with
  `build.sh` from the upgrade `exrpd`, packed as a release archive, and the tested
  Cosmovisor. It checks that the manifest has no `upgrades`, the right genesis plan and
  SHA-256, and a `launcher.commit`. Then:
  - It installs the single bundle over the two-binary bundle while the node runs. Until
    the restart, client commands refuse with "differs", and `version` starts with
    `restart-needed`. After the restart, `current` stays on `upgrades/<plan>`, the node
    runs the standalone upgrade bytes, and a signed bank transaction succeeds.
  - A data file with the genesis plan of the two-binary bundle (`v11.3.0`) makes a
    client command and `start` refuse with the single-bundle message, and the files do
    not change.
  - A new home with post-upgrade state, no data file and no `cosmovisor/` runs the
    upgrade `exrpd`, and the chain advances.

#### With `--unpatched`

The run passes only when the SIGTERM after the in-process switch shows the defect. Both
conditions must be true:

- a defect signature: an orphaned `exrpd` that still produces blocks, or the log
  `process already finished` of the old handler,
- Cosmovisor exit status 1.

Exit status 1 without a signature fails. A signature with another exit status also
fails. With the patched build, every stop must show no orphaned `exrpd`, no log of the
old handler, and exit status 0.

The order of the signal handlers is not fixed, so a run often misses the defect. Run it
again in that case. `TestLaunchProcessSignalAfterUpgrade` is the deterministic check.

## CI

### Pull requests

`.github/workflows/pull-request.yml` has two jobs for the bundle:

| Job (check name) | Runs |
| --- | --- |
| `bundle` ("Bundle: unit tests") | every pull request and every push to `main`: `check.sh`, which runs launcher `go vet`, `go test` and `golangci-lint` in `cmd/exrpd-bundle`, then `cosmovisor.sh --test` |
| `bundle-e2e` ("Bundle: E2E") | pull requests and pushes to `main` that change `cmd/exrpd/`, `cmd/exrpd-bundle/`, `scripts/bundle/`, `app/upgrades/`, `app/upgrades.go`, `app/app.go`, the root `go.mod` or `go.sum`, or `.github/workflows/pull-request.yml`, and every push without a base commit to compare with: `e2e.sh` with an `exrpd` built from the commit (`make build VERSION=<plan version>`), the plan of the commit (`plan.sh checkout`) and the genesis release that `plan.sh genesis` picks |

### Bundle

"Release" (`.github/workflows/release.yml`) does not build bundles. It publishes the
standalone release, the Docker image and the GoReleaser assets, as before the bundle.
After that, dispatch "Bundle" (`.github/workflows/bundle.yml`) for the published release.
It reads only what "Release" published and changes nothing of it.

| Input | Effect |
| --- | --- |
| `tag` (required) | the published release, for example `v12.0.1` |
| `genesis_tag` (optional) | also build the two-binary set of this genesis, for example `v11.3.0`. There is no default |

"Bundle" runs only on `main`. Its first job fails on another branch, and the other jobs do
not run. The reason: the launcher inside every bundle is built from the branch that the
workflow runs on, not from the release tag, and this workflow never replaces published
assets or images. Runs for the same tag wait for each other (`concurrency`), so two runs
cannot race on the checks below. This does not prevent another publisher from changing
an image tag outside this workflow.

The jobs:

1. **`plan`** checks that the tags are `vX.Y.Z` releases that are published, not drafts.
   It reads both plans from the tags (`plan.sh release`). For the E2E test it needs a
   genesis also when only the single set is built: `genesis_tag`, else
   `plan.sh genesis`. Then it checks each set on the release. A set has 5 files: the
   archives for Linux and Darwin, amd64 and arm64, and the checksums file.
   - No file of the set: build it.
   - All 5 files: skip it.
   - 1 to 4 files: fail, and list the files to delete before a rerun.

   It also checks each image tag with `docker manifest inspect`: missing means build.
2. **`assets`** runs only when a set must be built:
   1. `check.sh`, with no Go or lint cache, because the job publishes assets.
   2. Download the standalone archives of the release and of the genesis.
   3. `release.sh` once for each set to build.
   4. One `e2e.sh` run with the linux/amd64 assets that it built (`--bundle`,
      `--single-bundle`) and `--plan` from the tag.
   5. Check each set again. Skip complete sets and refuse partial sets. Upload missing
      sets: the archives first, the checksums file last. So a checksums file on the
      release means that the set is complete.
3. **`image`** runs for each missing image tag, after `assets` succeeded or was skipped:
   1. Download the published linux/amd64 archive and the checksums file, verify the
      archive, and extract `exrpd` and `LICENSE`.
   2. Build `Dockerfile.bundle` locally (`load: true`), with the files
      as the named build context `bundle-files`. That name must differ from every stage
      name: a named context with the name of a stage replaces that stage.
   3. `image-smoke.sh`: the manifest has the expected versions; `exrpd init` on a new
      home runs the genesis `exrpd`; every installed `exrpd` prints its version; and
      Cosmovisor runs.
   4. Check the image tag again. Skip an existing tag, or run `docker push` for a missing
      tag, only after the smoke test passed. If the registry check fails for another
      reason, stop without pushing.

### After a failed run

Publishing jobs check the remote state again before each set uploads or each image
pushes. "Re-run failed jobs" can repeat builds and tests, but skips publication of
complete sets and existing images. Partial asset sets still need the cleanup below.

| Cause | Recovery |
| --- | --- |
| transient failure (network, runner) | "Re-run failed jobs" in GitHub, or dispatch "Bundle" again with the same inputs |
| upload stopped part way | the next run fails on the partial set and lists its files. Delete them from the release, then dispatch again |
| image job failed after the upload | dispatch again with the same inputs: the assets are complete, so the run only builds the missing image |
| defect in the release itself | a rerun fails again. Publish a new release, for example the next patch, and bundle that one |
| defect in the launcher, scripts or workflow, before affected assets or images are published | fix it on `main`, then dispatch again |
| defect that requires changing a published bundle or image | fix it on `main`. Publish a new release and bundle it. A same-tag rerun keeps published assets and images |
| wrong `genesis_tag` | delete the `exrpd-bundle_<wrong genesis>-<version>_*` assets and the image tag, then dispatch with the right one |

## Releases and bundles

### What a bundle is made of

| Input | Source |
| --- | --- |
| Plan | `plan.sh release` of the release tag: the newest `vX.Y.Z` `UpgradeName` in its `app/upgrades` |
| Genesis | single set: the release itself. Two-binary set: the `genesis_tag` input. Its plan comes from its tag and must come before the plan of the release: `build.sh` and `release.sh` refuse, for example, a 12.0.x genesis for plan `v12.0.0` |
| Launcher | `cmd/exrpd-bundle/` on `main`, recorded as `launcher.commit` in the manifest |
| Cosmovisor | `scripts/bundle/cosmovisor/cosmovisor.env` and the patch next to it |

The release version must have the major.minor of its plan (today 12.0). Otherwise
`build.sh` and `release.sh` stop with an error, so a major or minor release without its
upgrade handler fails. A tag needs no bundle script, because the plan comes from
`app/upgrades` of the tag. So releases from before the bundle can get bundles too.

The asset names have the genesis version and the release version, for example
`exrpd-bundle_11.3.0-12.0.1_Linux_amd64.tar.gz` and `exrpd-bundle_12.0.1_Linux_amd64.tar.gz`,
so a patch bundle never has the name of an earlier one.

### Which bundles a release gets

| Situation | Dispatch | Sets |
| --- | --- | --- |
| v12.0.0 released, the halt is ahead | `tag: v12.0.0`, `genesis_tag: v11.3.0` | `bundle-11.3.0-12.0.0`, `bundle-12.0.0` |
| another network runs another genesis | `tag: v12.0.0`, `genesis_tag: v11.2.0` | `bundle-11.2.0-12.0.0` (`bundle-12.0.0` exists) |
| v12.0.1 before the halt | `tag: v12.0.1`, `genesis_tag: v11.3.0` | `bundle-11.3.0-12.0.1`, `bundle-12.0.1` |
| v12.0.1 after the halt | `tag: v12.0.1` | `bundle-12.0.1` |

Each set gets its image with the same name: `peersyst/exrp:bundle-<set>`.

### Release types

| Release | Chain | Bundles |
| --- | --- | --- |
| Major or minor (`X.0.0`, `X.Y.0`) | changes state or consensus, needs a governance plan `vX.Y.0` with an upgrade handler | the two-binary set: the genesis is the release that validators run at the halt, the plan is the new handler. And the single set, for nodes after the halt |
| Patch (`X.Y.Z`, `Z > 0`) | no new upgrade handler, no state or consensus change: a drop-in replacement | the two-binary set with the same plan and the patch as the upgrade binary, for nodes before the halt. And the single set, for nodes after it |

A patch bundle reaches validators with a reinstall and a restart. `start` replaces
`upgrades/v12.0.0/bin/exrpd` and logs it. Restart right after the reinstall, because
until then the other commands refuse: the installed bundle differs. The standalone patch
binary stays an alternative.

Nothing checks that a patch is a drop-in replacement. A patch that changes state or
consensus must be a minor release with its own plan.

### Patch of the genesis line

This is a patch while a bundle is out but before its halt, for example 12.0.3 while the
v12.1.0 bundle carries genesis 12.0.2:

1. Release the standalone 12.0.3, and dispatch "Bundle" for it. This gives the single
   set `bundle-12.0.3` for homes that still run plan `v12.0.0`.
2. Release a patch of the upgrade line (12.1.1), and dispatch "Bundle" with
   `genesis_tag: v12.0.3`. Its `start` replaces `upgrades/v12.0.0/bin/exrpd`.

### Next bundle

The next bundle, for example `bundle-12.0.x-12.1.0`, has genesis 12.0.x and genesis plan
`v12.0.0`. A node on the single bundle `bundle-12.0.x`, or on the two-binary bundle after
its halt, installs it with no manual step:

- On a home whose data file has the plan `v12.0.0`, `current` already points at
  `upgrades/v12.0.0`. The bundle replaces the binary there with 12.0.x and adds
  `upgrades/v12.1.0`.
- Without a data file, `current` on `upgrades/v12.0.0` is a directory of the next bundle
  too, so it stays.

### Different genesis per network

Usually every network runs the same release before an upgrade, so one two-binary set
serves all of them. Sometimes a network runs another release, for example testnet on
11.1.x and mainnet on 11.2.x. Then dispatch "Bundle" for the same `tag` once for each
genesis. The single set does not depend on the genesis, so only the first run builds it.

The asset names and image tags have the genesis version, for example
`exrpd-bundle_11.1.3-12.0.0_Linux_amd64.tar.gz` and
`exrpd-bundle_11.2.0-12.0.0_Linux_amd64.tar.gz`, so both sets go to the same release with
their own checksums. Tell the validators of each network which set is theirs. The upgrade
handler must accept the state of each genesis, and the bundle does not check that.

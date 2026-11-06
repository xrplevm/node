# exrpd bundle

> [!NOTE]
> The versions in this document are examples from the time of writing. 11.3.0
> is the version that the chain runs before the upgrade, and 12.0.0 is the version that
> the chain upgrades to. For another bundle, take the versions from its archive name, for
> example `exrpd-bundle_<current>-<upgrade>_<Os>_<arch>.tar.gz`. The plans `v11.3.0` and
> `v12.0.0` do not always match the versions. Take them from `manifest.genesis.plan` and
> `manifest.upgrades[0].plan` in `exrpd version --bundle`.

The bundle is an `exrpd` for validators and other nodes that must switch binaries at the
upgrade height together with the rest of the network. It runs 11.3.0, and at the `v12.0.0`
plan height it stops 11.3.0 and starts 12.0.0 by itself, so the chain upgrades fast.
Nothing is downloaded. Machines that only run client commands (keys, transactions,
queries) against a node do not need it: use the standalone `exrpd` there, and update it
by hand after the upgrade. For details and internals, see the
[developer notes](bundle-dev.md).

## How the bundle works

There are two kinds of bundles:

| Bundle | Contains | For |
| --- | --- | --- |
| two-binary, for example `exrpd-bundle_11.3.0-12.0.0` | the 11.3.0 `exrpd`, the 12.0.0 `exrpd` and Cosmovisor | nodes before the upgrade height |
| single, for example `exrpd-bundle_12.0.0` | the 12.0.0 `exrpd` and Cosmovisor | nodes after the upgrade height, also new nodes |

Every release has a single bundle. A two-binary bundle exists for a release that comes
out before its upgrade height, for example 12.0.0, or a patch such as 12.0.1 before the
halt. A patch after the halt normally has only the single bundle.

A node after the upgrade then installs the two-binary bundle of the next upgrade, for
example `exrpd-bundle_12.0.x-12.1.0`, before its height. So the node always runs a
bundle, and each switch is automatic.

To run the node, the bundle `exrpd` first installs its programs in the node home. The
home is the directory with the node's `config/` and `data/`. Set it with `--home`. The
default is `$HOME/.exrpd` (see [Home](bundle-dev.md#home) for other ways to set it).

The files go under `<node home>/cosmovisor/`: one directory for each `exrpd` version,
Cosmovisor, and the `current` link to the `exrpd` that matches the chain data (11.3.0
before the upgrade height, 12.0.0 after it). Nothing changes in `config/`, `data/` or
your keys. Files that are already correct stay as they are, so a second install of the
same bundle does no harm.

### Commands

Run every command as the node user, with `--home <node home>`:

| Command | What it does |
| --- | --- |
| `exrpd start` | Installs the files, then runs the node through Cosmovisor. The service runs this. |
| `exrpd init <moniker> --chain-id <id>` | On a new home: installs the files, then runs `init` with 11.3.0. |
| `exrpd bundle setup` | Installs the files and starts nothing. Use it with the node stopped, for example before `rollback`, `export` or a resync. |
| `exrpd version` | Shows the installed bundle: versions, plans and Cosmovisor. |
| Any other command (`keys`, `tx`, `query`, `status`, ...) | Runs with the `exrpd` that matches the chain data. |

Only `start`, `init` and `bundle setup` copy the files into the home. The other commands
use the files that are already in the home, so they first check that the home has the
files of this same bundle.

After you replace `/usr/local/bin/exrpd` with a new bundle, for example a patch release,
the home still has the files of the old bundle. The service installs the new files when
it restarts. Until then, `keys`, `tx`, `query` and the other commands stop with a message
that tells you to restart the service. So restart the service right away after you
replace the file.

`exrpd version` does not stop. Its first word is the version of the `exrpd` in the home,
or one of these:

- `restart-needed`: the home has the files of another bundle. Restart the service.
- `not-installed`: the home has no bundle files yet, for example the home of a
  standalone `exrpd`.

Run every command as the user that runs the service:

```sh
sudo -u <node user> /usr/local/bin/exrpd keys list --home <node home>
```

With `sudo`, write the full path of `exrpd`, because some systems (RHEL, Rocky) leave
`/usr/local/bin` out of the `PATH` of `sudo`. If the service runs as root, run the
commands as root, without `sudo -u`.

> [!WARNING]
> If another user, for example root, creates files in the home, the service cannot change
> them later. If root owns `cosmovisor/`, the next `exrpd start` fails. If root owns the
> directory of a later upgrade in `cosmovisor/upgrades/`, `start` still works, and
> Cosmovisor fails at the height of that upgrade.

For flags, help and macOS, see [Commands](bundle-dev.md#commands) and
[launchd](bundle-dev.md#macos-launchd).

## Prepare

### Download and verify

> [!NOTE]
> This guide uses `/usr/local/bin/exrpd`, the usual path. If your node runs the
> standalone `exrpd` from another path, use that path in every command, and replace that
> file, so the machine has only one `exrpd`. The `ExecStart` line of your service shows
> the path: `systemctl cat <unit> | grep ExecStart`. Do not trust `which exrpd` for this:
> it shows the `exrpd` of your shell, which can be a different file.

Take the bundle for the state of your node (see [How the bundle works](#how-the-bundle-works)):
before the upgrade height `exrpd-bundle_11.3.0-12.0.0`, after it `exrpd-bundle_12.0.0`.
From the GitHub release, download its archive for your platform and its checksums file,
for example `exrpd-bundle_11.3.0-12.0.0_Linux_amd64.tar.gz` and
`exrpd-bundle_11.3.0-12.0.0_checksums.txt`. Then verify and unpack it:

```sh
sha256sum -c --ignore-missing exrpd-bundle_11.3.0-12.0.0_checksums.txt  # must print OK
tar -xzf exrpd-bundle_11.3.0-12.0.0_Linux_amd64.tar.gz exrpd
```

On macOS, use `shasum -a 256 -c` instead of `sha256sum -c`. The binaries are not signed,
so remove the quarantine attribute (`xattr -d com.apple.quarantine`) after download.

Each path below installs this `exrpd` at its own step.

### systemd unit

The unit is a file in `/etc/systemd/system/`. This guide names it `exrpd.service`. You can
use another name: then use that name in the `systemctl` commands. To create a new unit,
create the file as root, for example with `sudoedit /etc/systemd/system/exrpd.service`,
with this content:

```ini
[Unit]
Description=XRPL EVM node
After=network-online.target
Wants=network-online.target

[Service]
User=<node user>
ExecStart=/usr/local/bin/exrpd start --home <node home>
Restart=on-failure
RestartSec=5
LimitNOFILE=65535
KillMode=control-group
TimeoutStopSec=120

[Install]
WantedBy=multi-user.target
```

`KillMode=control-group` is the systemd default. The unit sets it so that the intent is
clear: when the service stops, or when Cosmovisor exits on its own, systemd stops the node
`exrpd` too. Do not use `KillMode=process` or `none`: then the node `exrpd` can keep
running without Cosmovisor.

To change an existing unit, keep your own settings, for example extra `start` flags or
limits. Two settings must stay as in the unit above: `ExecStart` runs the bundle `exrpd`
with `start --home <node home>` (your flags can follow), and `KillMode=control-group`.
Edit the unit with `sudo systemctl edit --full <unit>`. This command saves the unit in
`/etc/systemd/system/`, so a package update does not undo the change, and it reloads
systemd. Then check the output of `systemctl cat <unit>`. It shows the unit file first,
then any drop-in files (`<unit>.service.d/*.conf`). A drop-in overrides the unit file, so
remove its `ExecStart`, `KillMode` and `DAEMON_*` lines, and run
`sudo systemctl daemon-reload`.

### Check the service

After a path starts the service, check that the service runs the bundle. Run these
commands, with your unit name:

```sh
sudo journalctl -u <unit> --since "10 min ago" | grep 'exrpd: home'
systemctl status <unit>
sudo -u <node user> /usr/local/bin/exrpd version --home <node home>
```

- The log has the line `exrpd: home <node home> (from --home)`. The standalone `exrpd`
  never writes this line, so it shows that the bundle started the node.
- The status shows a `cosmovisor` process with an `exrpd` process below it.
- The version line starts with a version, for example `11.3.0`. If it starts with
  `not-installed` or `restart-needed`, the home does not have the files of this bundle
  yet. Restart the service.

### Node settings

- At the upgrade height, Cosmovisor runs `exrpd status` to check the height. This command
  uses the node's RPC server, so keep `[rpc] laddr` in `config/config.toml` on, and `node`
  in `config/client.toml` pointed at it. The defaults do this. If `exrpd status` fails,
  the node halts at the upgrade height and does not switch until you restart the service.
- Use absolute paths in settings that exrpd resolves against the working directory,
  because Cosmovisor runs exrpd in `<node home>/cosmovisor`. A relative path there can
  fail at start, or work without an error and use a file in `<node home>/cosmovisor`.
  Check:
  - `start` flags, for example `--trace-store`
  - `EXRPD_*` variables
  - `config/app.toml`, for example `[tls] certificate-path` and `key-path`
  - `unix://` addresses in `config/config.toml`, for example `priv_validator_laddr` and
    `[rpc] laddr`

  File paths in `config/config.toml`, such as `genesis_file`, resolve against the home
  and can stay relative. See
  [Working directory](bundle-dev.md#working-directory).

## Choose your path

> [!WARNING]
> Move a running node at least one day before the upgrade, or after the upgrade. With
> less than one day left, keep your current setup and move after the upgrade. A problem
> during a move close to the upgrade height can leave the node down at that height.
> After the upgrade, the bundle finds the data file that the chain wrote at the upgrade
> height, and runs 12.0.0.

Follow the one section that matches your machine:

- [Set up a new node](#set-up-a-new-node): this machine runs no node yet.
- [Move from the standalone exrpd](#move-from-the-standalone-exrpd): the node runs the
  standalone `exrpd` now.
- [Move from Cosmovisor](#move-from-cosmovisor): the node runs Cosmovisor now, or ran it
  before.
- [Run the node in a container](#run-the-node-in-a-container): the node runs the
  `peersyst/exrp` image with Docker, Docker Compose or Kubernetes.
- The node already runs a bundle, and you install a new one, for example a patch release:
  see [Install a new bundle file](#install-a-new-bundle-file).

## Set up a new node

A node that joins after the upgrade height, with a state sync or a snapshot from after
the upgrade, uses `exrpd-bundle_12.0.0`. See [Join after the upgrade](#join-after-the-upgrade).

1. Install the bundle that you downloaded, and check it:

   ```sh
   sudo install -m 0755 ./exrpd /usr/local/bin/exrpd
   exrpd version --bundle
   ```

2. As the node user, run `exrpd init <moniker> --chain-id <id> --home <node home>`.
3. Add keys, the genesis file, peers and config as you do for the standalone `exrpd`.
4. Check that your config follows [Node settings](#node-settings).
5. Create the unit as in [systemd unit](#systemd-unit), then start it:

   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable --now exrpd
   ```

6. [Check the service](#check-the-service).

## Move from the standalone exrpd

1. Do not install the bundle that you downloaded in `/usr/local/bin` yet. First put a
   copy where the node user can run it (not in `/tmp`, which hardened hosts mount
   `noexec`):

   ```sh
   sudo install -d -m 0755 /usr/local/lib/exrpd-check
   sudo install -m 0755 ./exrpd /usr/local/lib/exrpd-check/exrpd
   ```

2. Check that the bundle accepts this home. Run the copy as the node user, with the home
   of your service:

   ```sh
   sudo -u <node user> /usr/local/lib/exrpd-check/exrpd version --bundle --home <node home>
   ```

   This command only reads files. Go on only if all these checks pass:

   - `refused` is `null`. If it is not, stop the current running node, fix what it says,
     and run the check again. For example, it can name a stale `data/upgrade-info.json`.
   - `will_run.dir` matches the state of the node:
     - Before the upgrade height, with `exrpd-bundle_11.3.0-12.0.0`: `upgrades/v11.3.0`.
       If it is `upgrades/v12.0.0`, `data/upgrade-info.json` already names `v12.0.0`.
       Someone staged that file, or an old one stayed. Stop the current running node,
       delete the file and run the check again.
     - After the upgrade height: `upgrades/v12.0.0`. Any other value means that you have
       the two-binary bundle on a home without a `v12.0.0` data file, for example after a
       state sync. It would run 11.3.0 on 12.0 data. Use `exrpd-bundle_12.0.0` instead.
   - The settings listed in [Node settings](#node-settings) use absolute paths.

   If you stop the move here, run `sudo rm -r /usr/local/lib/exrpd-check`.
   `/usr/local/bin/exrpd` is still the standalone until step 3, so the node runs and
   restarts as before. If you stopped the node for a fix, start it again.
3. Install the bundle:
   `sudo install -m 0755 ./exrpd /usr/local/bin/exrpd && sudo rm -r /usr/local/lib/exrpd-check`.
   A running node keeps going. Until the restart in step 4 installs the files, other
   commands of the bundle (`keys`, `tx`, `query`, `status`, ...) refuse with
   `The bundle is not installed in <node home>`. Run them after step 4.
4. Switch the service to the bundle. Use one of these ways:
   - Keep your current unit. Change it as in [systemd unit](#systemd-unit), then run:

     ```sh
     sudo systemctl enable <unit>
     sudo systemctl restart <unit>
     ```

   - Use a new unit. Stop and disable the old unit first:
     `sudo systemctl disable --now <old unit>`. An old unit that stays enabled starts
     again at the next boot, and then two nodes run on the same home. Then create the new
     unit as in [systemd unit](#systemd-unit) and run:

     ```sh
     sudo systemctl daemon-reload
     sudo systemctl enable --now exrpd
     ```

5. [Check the service](#check-the-service).

## Move from Cosmovisor

The bundle manages `<node home>/cosmovisor` itself. You no longer copy binaries there or
run `cosmovisor add-upgrade`. Follow these steps for a home that runs Cosmovisor now, or
ran it before.

1. Stop and disable the old service. If you followed the official docs, its name is
   `cosmovisor-exrpd.service`:

   ```sh
   sudo systemctl disable --now <old unit>
   ```

   An old unit that stays enabled starts again at the next boot, and then two nodes run
   on the same home. If you keep the old unit for the bundle, step 5 changes its file.
2. Read the data file with the node stopped. No process writes it now.

   ```sh
   sudo -u <node user> cat <node home>/data/upgrade-info.json
   ```

   If `<node home>/data/upgrade-info.json.batch` exists, delete it. Cosmovisor batch
   upgrades create this file, and every command of the bundle stops while it exists.
   Then find your case:

   - If you move before the upgrade height, delete the data file, if it exists. The chain
     writes a `v12.0.0` file only at the upgrade height. So a file there names an earlier
     plan, or a plan that you staged with `cosmovisor add-upgrade --upgrade-height` or
     `cosmovisor add-batch-upgrade`. Without the file, the bundle runs 11.3.0. At the
     upgrade height, the chain writes a new file, and the bundle switches to 12.0.0.
   - If you move after the upgrade height, you use `exrpd-bundle_12.0.0`. Keep the file
     if it names `v12.0.0`. Without a file, for example after a state sync, this bundle
     runs 12.0.0 too. A file with an older plan comes from a binary that you swapped
     before the upgrade height. The check in step 4 then refuses and says how to fix it.

   Then remove the binaries of your old Cosmovisor setup. The bundle installs its own
   binaries in `upgrades/` and sets `current` itself. It does not use `genesis/`.

   ```sh
   sudo rm -rf <node home>/cosmovisor/genesis <node home>/cosmovisor/upgrades \
     <node home>/cosmovisor/current
   ```

   Keep the other files in `<node home>/cosmovisor/`.
3. Install the bundle that you downloaded:

   ```sh
   sudo install -m 0755 ./exrpd /usr/local/bin/exrpd
   ```

   Until step 5 starts the service, the other commands of the bundle refuse with
   `The bundle is not installed in <node home>`.
4. Check the setup before you start it. Run this with the home of your service:

   ```sh
   sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <node home>
   ```

   Check these fields:

   - `home` is the home of your node.
   - `will_run.dir` matches the state of the node:
     - Before the upgrade height, with `exrpd-bundle_11.3.0-12.0.0`: `upgrades/v11.3.0`.
     - After the upgrade height: `upgrades/v12.0.0`. Any other value means that you have
       the two-binary bundle on a home without a `v12.0.0` data file, for example after a
       state sync. It would run 11.3.0 on 12.0 data. Use `exrpd-bundle_12.0.0` instead.
   - `refused` is `null`. If it is not, fix what it says, and run the check again.
   - `installed_bundle` is `none` and `will_run.exists` is `false`. This is correct here,
     because the first start installs the files.

   If the node ran without Cosmovisor before, also check that the settings listed in
   [Node settings](#node-settings) use absolute paths. Under Cosmovisor, relative paths
   already resolve in `<node home>/cosmovisor`, so they keep working.
5. Create the unit as in [systemd unit](#systemd-unit), or change the old unit. Do not
   copy the Cosmovisor `ExecStart` or the old `DAEMON_*` settings. `exrpd start` sets the
   Cosmovisor variables that it needs (see [Settings](#settings)), and it refuses
   `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`. Then start the service:

   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable --now <unit>
   ```

6. [Check the service](#check-the-service).

> [!WARNING]
> Never copy the bundle into `<node home>/cosmovisor`, and never replace a directory
> there with a symlink. The bundle manages `upgrades/`, `current` and `supervisor/` there.

## Run the node in a container

The bundle images are drop-in replacements for the standalone image `peersyst/exrp`.
They have the bundle `exrpd` at the same path, with the same entrypoint and user. Take
the image for the state of your node, as for the archives:

| Node | Image |
| --- | --- |
| before the upgrade height | `peersyst/exrp:bundle-11.3.0-12.0.0` |
| after the upgrade height, also a new node | `peersyst/exrp:bundle-12.0.0` |

Change only the image, for example in Docker Compose:

```yaml
  node:
    image: peersyst/exrp:bundle-11.3.0-12.0.0   # was peersyst/exrp:v11.3.0
    entrypoint: exrpd
    command: ["start"]
    volumes:
      - /root/.exrpd:/root/.exrpd
```

Check these settings of your container:

- Mount the entire node home, for example `/root/.exrpd`, with a bind mount or a named
  volume. Keep its config, data, keys and `cosmovisor/` on that mount. Do not use an
  anonymous volume: the new container may not reuse it.
- Start `exrpd` as the container command, for example `entrypoint: exrpd` and
  `command: ["start"]`. Cosmovisor then gets the stop signal directly.
- Give the node time to stop cleanly: `stop_grace_period: 120s` in Docker Compose, or
  `terminationGracePeriodSeconds: 120` in Kubernetes. The defaults are 10 s and 30 s,
  and the systemd unit uses 120 s.
- Run other commands in the same container, for example
  `docker exec <container> exrpd keys list`. Then they run as the same user and on the
  same home.

To move a running container to a bundle image, follow the same timing as for a host (see
the warning in [Choose your path](#choose-your-path)). Before the upgrade, check the home
first with
`docker run --rm -v <volume>:/root/.exrpd --entrypoint exrpd <image> version --bundle`,
as in step 2 of [Move from the standalone exrpd](#move-from-the-standalone-exrpd).

After you change the image in the Compose file, recreate the container. A restart does
not apply the new image. Before removing the container, verify that its entire node home
uses the bind mount or named volume above.

> [!WARNING]
> Do not add `-v` to `docker compose rm` or run `docker compose down -v`.
> Do not delete the bind-mounted directory or the named volume.

For the Compose service `node` above, run:

```sh
docker compose stop node
docker compose rm -f node
docker compose up -d node
```

## At the upgrade height

You do not need to do anything, if the service runs the bundle: after you install it,
restart the service once (see [Choose your path](#choose-your-path)). The chain halts at
the `v12.0.0` plan height, and Cosmovisor stops 11.3.0 and starts 12.0.0 under the same
Cosmovisor process. Each `exrpd` version runs as a separate child process, so its PID
changes. If the node halts and does not switch, restart the service.

> [!WARNING]
> On restart, the bundle selects 12.0.0 because `data/upgrade-info.json`
> names `v12.0.0`. It does not check that the Cosmovisor backup
> (`UNSAFE_SKIP_BACKUP=false`) or `COSMOVISOR_CUSTOM_PREUPGRADE` completed, and it does not
> run them again. The restart starts the migration, and you cannot undo it. If you enabled
> the backup or a custom pre-upgrade script, find why it failed and complete the backup
> and checks manually before you restart. See
> [At the upgrade height](bundle-dev.md#at-the-upgrade-height).

## Restart, reinstall and restore

> [!WARNING]
> A chain only moves forward. Never start an older release or bundle on a home that a
> newer one has run. Never put back an older `data/priv_validator_state.json`.

### Restart

A restart needs no extra step. On every start, the bundle picks the `exrpd` for the chain
data in this order:

1. If `data/upgrade-info.json` exists, the bundle runs the plan that it names. `v11.3.0`
   runs 11.3.0, and `v12.0.0` runs 12.0.0. The chain writes this file when it halts at an
   upgrade height, so after the upgrade the file names `v12.0.0`. A file with another
   plan stops the start with an error, see
   [Problems with the data file](#problems-with-the-data-file).
2. If there is no data file, the bundle follows `<node home>/cosmovisor/current`. This
   link points at the `exrpd` that ran the data last. If the bundle cannot tell which
   plan that `exrpd` runs, the start stops with an error. See the full rule in
   [How the launcher chooses the binary](bundle-dev.md#how-the-launcher-chooses-the-binary).
3. If there is no data file and no `current`, for example on a new home, the bundle runs
   its first version: 11.3.0 in `exrpd-bundle_11.3.0-12.0.0`, 12.0.0 in
   `exrpd-bundle_12.0.0`.

Then `current` points at the chosen `exrpd`. When the link moves, the service log has a
line that says why, for example
`exrpd: moved <node home>/cosmovisor/current from upgrades/v11.3.0 to upgrades/v12.0.0: data/upgrade-info.json names v12.0.0`.

### Install a new bundle file

Use this for a reinstall, or for a patch release such as 12.0.1. Take the bundle for the
state of your node: before the upgrade height `exrpd-bundle_11.3.0-12.0.1`, after it
`exrpd-bundle_12.0.1`. Download and verify it as in
[Download and verify](#download-and-verify). Then install it:

```sh
sudo install -m 0755 ./exrpd /usr/local/bin/exrpd
```

Then put the new files in the home. Until then, `keys`, `tx`, `query` and the other
commands refuse, because the home still has the files of the old bundle.

- If the node runs, restart the service right away: `sudo systemctl restart <unit>`. On
  start, the bundle replaces the binaries in the home that changed, and it writes one log
  line for each.
- If the node must stay stopped, for example for `rollback`, `export` or a resync, do not
  restart it. Run
  `sudo -u <node user> /usr/local/bin/exrpd bundle setup --home <node home>`. It installs
  the files, does not move `current`, and does not start the node.

In a container, change the image tag instead, for example to
`peersyst/exrp:bundle-12.0.1`. Then follow the
[container recreation steps](#run-the-node-in-a-container).

### Resync after the upgrade

Use this when you reset the node with `unsafe-reset-all` after the upgrade, and then get
the data again with a state sync or a snapshot from after the upgrade height. The reset
deletes all of `data/`, and two files there must survive it:

- `data/priv_validator_state.json` holds the last height that the validator signed. The
  reset writes a new file at height 0, and with it the validator can sign a height twice.
- `data/upgrade-info.json` has the plan `v12.0.0`. If the bundle already ran 12.0.0 on
  this home, you do not need the file: the reset does not touch `cosmovisor/`, so
  `current` still points at 12.0.0. The file matters only when `current` still points at
  11.3.0, for example after a crash at the upgrade height before the switch. Keep it in
  every case, so you do not have to check.

Copy both files out of `data/` before the reset. Put them back just before you start the
node, after you unpack the snapshot: a snapshot can contain its own, older
`priv_validator_state.json`. [The steps](bundle-dev.md#resync-after-the-upgrade)
give each command.

### Join after the upgrade

A new node that joins after the upgrade height, with a state sync or a snapshot from after
the upgrade, uses `exrpd-bundle_12.0.0`, or the newest patch such as
`exrpd-bundle_12.0.1`. This bundle runs 12.0.x on a home without `data/upgrade-info.json`.
Before the next upgrade, install the bundle of the next plan, for example
`exrpd-bundle_12.0.x-12.1.0`.

A state sync leaves no `data/upgrade-info.json`. A snapshot can contain one. If it names
the current chain version, keep it. If it names an older plan, the start stops: write the
file again with the last plan and its height, as the message says. Do not only delete it.

> [!WARNING]
> Do not use this bundle with a snapshot from before the upgrade height. It would run
> 12.0.x on data that needs 11.3.0, and exrpd stops.

### Restore a snapshot from before the upgrade

This is for a node that goes back to a snapshot of `data/` from before the upgrade height.
Before you restore the snapshot, check that the matching two-binary bundle
`exrpd-bundle_11.3.0-12.0.x` is available. If it is missing, contact the maintainers to
publish it before you restore.

With the node stopped:

1. Restore the snapshot. Keep the newer `data/priv_validator_state.json`, as the warning
   above says. With it, the validator does not sign again the heights that it already
   signed.
2. Delete `<node home>/cosmovisor/current`. It points at 12.0.0, and 12.0.0 cannot run
   data from before the upgrade.
3. Delete `data/upgrade-info.json` if it has the plan `v12.0.0`. With this file, the
   bundle would run 12.0.0 too.
4. If the node runs `exrpd-bundle_12.0.x`, install `exrpd-bundle_11.3.0-12.0.x` instead,
   as in [Install a new bundle file](#install-a-new-bundle-file). Only this bundle has
   11.3.0.
5. Start the service. With no data file and no `current`, the bundle runs 11.3.0.

The node then catches up as the network did. It gets the missing blocks from its peers
and applies them with 11.3.0. At the upgrade height, 11.3.0 halts, and Cosmovisor
switches to 12.0.0, as in [At the upgrade height](#at-the-upgrade-height). 12.0.0 then
runs the upgrade and catches up to the latest block.

### Problems with the data file

If `data/upgrade-info.json` is corrupted, every command stops with an error that explains
the fix. Write the file again with the exact plan name and height of the last upgrade,
from the governance proposal. Do not only delete it. The bundle then runs the binary for
that plan.

If the file names a plan older than the plans of this bundle, for example `v11.2.0`,
check how this node passed the later plans. If you swapped in their binaries before the
plan height, the chain did not write the file again, so the file is old. Write it again
with the last plan that the node passed and its height. Otherwise, run the release for
the plan in the file.

### Leave the bundle

Stop the service, and install the standalone `exrpd` of the version that the node runs.

For details and edge cases, see
[Restart, reinstall and restore](bundle-dev.md#restart-reinstall-and-restore).

## Settings

`exrpd start` sets these Cosmovisor variables. It refuses to start if you set a
conflicting value:

| Variable | Value |
| --- | --- |
| `DAEMON_NAME` | `exrpd` |
| `DAEMON_ALLOW_DOWNLOAD_BINARIES` | `false` |
| `DAEMON_RESTART_AFTER_UPGRADE` | `true` |
| `DAEMON_HOME` | the resolved home. A different `$DAEMON_HOME` is replaced, not refused |

Defaults that differ from upstream Cosmovisor. Set the variable to override:

| Variable | Default | Meaning |
| --- | --- | --- |
| `DAEMON_SHUTDOWN_GRACE` | `30s` | time exrpd gets to exit after SIGINT at the upgrade height before it is killed. Cosmovisor refuses `0` |
| `UNSAFE_SKIP_BACKUP` | `true` | upstream copies all of `<node home>/data` at the upgrade height while the validator is down |

All other Cosmovisor variables keep their upstream meaning. Do not use Cosmovisor batch
upgrades or `cosmovisor add-upgrade --upgrade-height`. Every command stops while
`data/upgrade-info.json.batch` exists. See
[Batch upgrades](bundle-dev.md#batch-upgrades) and [Settings](bundle-dev.md#settings).

## Rollback

Keep a rollback point before the upgrade height. By default nothing copies `data/`, and
11.3.0 cannot open the data after the upgrade. Take a snapshot of `data/` with the node
stopped and test that it restores. You can also enable the Cosmovisor backup, see
[Cosmovisor backup](bundle-dev.md#cosmovisor-backup).

> [!WARNING]
> If you restore a backup, keep the newer `data/priv_validator_state.json`.
> An older copy lets the validator sign heights it already signed. That is double
> signing.

## Known limitations

- About 100 MB to download (linux/amd64). Installing writes about 300 MB. Files of
  earlier bundles stay on disk. After each new bundle, remove them as described in
  [Removing old files](bundle-dev.md#removing-old-files).
- No Windows bundle. Use the standalone `exrpd` there.
- One upgrade per bundle. Install the bundle for the next plan before its height.
- For nodes only. On a client machine, a home without node data keeps the genesis binary
  after the upgrade. Use the standalone `exrpd` there.
- No sync from genesis with a released bundle: it carries two plans, not the whole chain
  history. See
  [Sync from genesis with a bundle built by hand](bundle-dev.md#sync-from-genesis-with-a-bundle-built-by-hand).
- Without a service manager, a leftover `exrpd` can outlive Cosmovisor. See
  [Without a service manager](bundle-dev.md#without-a-service-manager).
- The supervisor is a patched Cosmovisor until the fix is available upstream. See
  [Cosmovisor patch](bundle-dev.md#cosmovisor-patch).

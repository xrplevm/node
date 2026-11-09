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
by hand after the upgrade.

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
default is `$HOME/.exrpd`. Without `--home`, the bundle uses the first set variable in
this order: `DAEMON_HOME`, `EXRPD_HOME`, then `NODE_HOME`. With `EXRPD_HOME`,
the bundle keeps the keyring there; the standalone keeps it in the default home.
Use `--home` to select one home for everything.

The files go under `<node home>/cosmovisor/`: one `upgrades/<plan>/` directory for each
plan, Cosmovisor in `supervisor/`, and the `current` link to the `exrpd` that matches the
chain data (11.3.0 before the upgrade height, 12.0.0 after it). `start` sets `current`.
`init` and `bundle setup` install only the binaries. Nothing changes in `config/`,
`data/` or your keys. Files that are already correct stay as they are, so a second
install of the same bundle does no harm.

### Commands

Run every command as the node user, with `--home <node home>`:

| Command | What it does |
| --- | --- |
| `exrpd start` | Installs the files, then runs the node through Cosmovisor. The service runs this. |
| `exrpd init <moniker> --chain-id <id>` | On a new home: installs the files, then runs `init` with 11.3.0. |
| `exrpd bundle setup` | Installs the files, does not change `current`, and starts nothing. Use it with the node stopped, for example before `rollback`, `export` or a resync. |
| `exrpd version` | Shows the installed bundle: versions and Cosmovisor. `exrpd version --bundle` also shows the plans and the home. |
| Any other command (`keys`, `tx`, `query`, `status`, ...) | Runs with the `exrpd` that matches the chain data. |

Only `start`, `init` and `bundle setup` copy the files into the home. The other commands
use the files that are already in the home, so they first check that the home has the
files of this same bundle.

After you replace `/usr/local/bin/exrpd` with a new bundle, for example a patch release,
the home still has the files of the old bundle. The service installs the new files when
it restarts. Until then, `keys`, `tx`, `query` and the other commands stop with a message
that tells you to restart the service. So restart the service right away after you
replace the file.

`exrpd version` does not stop when the bundle is missing or outdated in the home. Its
first word is the version of the `exrpd` in the home, or one of these:

- `restart-needed`: the home has the files of another bundle. Restart the service.
- `not-installed`: the home has no bundle files yet, for example the home of a
  standalone `exrpd`.

If the home itself is refused, for example because of a corrupted data file, `version`
stops with the same error as the other commands.

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

Write `start` or `init` before their flags. The bundle refuses a bare `--`, and refuses
`--home` immediately after another flag without `=`. Put `--home` first or write the
other flag with `=`. Use `exrpd help start` to see start flags without installing files.
It works once the home has this bundle installed, and otherwise shows the same refusal as
other commands. `exrpd start --help` installs the files.

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

`<node user>` must own the node home. When you replace a unit, use the `User=` of the old
unit, or first run `sudo chown -R <node user>: <node home>`.

To change an existing unit, keep your own settings, for example extra `start` flags or
limits. Two settings must stay as in the unit above: `ExecStart` runs the bundle `exrpd`
with `start --home <node home>` (your flags can follow), and `KillMode=control-group`.
Before retaining `WorkingDirectory`, check the
[working directory restriction](#working-directory-restriction), especially for a
versioned installation that uses a `current` symlink.
Edit the unit with `sudo systemctl edit --full <unit>`. This command saves the unit in
`/etc/systemd/system/`, so a package update does not undo the change, and it reloads
systemd. Then check the output of `systemctl cat <unit>`. It shows the unit file first,
then any drop-in files (`<unit>.service.d/*.conf`). A drop-in overrides the unit file, so
remove its `ExecStart` and `KillMode` lines and the `DAEMON_*` lines that `exrpd start`
sets (see [Settings](#settings)), and run `sudo systemctl daemon-reload`.

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
  and can stay relative.

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

2. Run `sudo -u <node user> /usr/local/bin/exrpd init <moniker> --chain-id <id> --home <node home>`.
3. Add keys, the genesis file, peers and config as you do for the standalone `exrpd`.
4. Get the chain data before you start the service. The bundle cannot sync from genesis,
   because it does not carry the binaries of the earlier plans (see
   [Known limitations](#known-limitations)).
   - Before the upgrade height, with `exrpd-bundle_11.3.0-12.0.0`: use state sync only if
     the chain is well before the upgrade height. Otherwise wait, and use
     `exrpd-bundle_12.0.0`. For a snapshot, unpack one from before the upgrade height,
     whose data requires 11.3.0, and inspect its `data/upgrade-info.json` as in step 3 of
     [Restore a snapshot from before the upgrade](#restore-a-snapshot-from-before-the-upgrade).
     Then run
     `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <node home>`, and
     go on only if `refused` is `null` and `will_run.dir` is `upgrades/v11.3.0`.
   - After the upgrade height, with `exrpd-bundle_12.0.0`: configure state sync, or
     unpack a snapshot from after the upgrade, as in
     [Join after the upgrade](#join-after-the-upgrade).
5. Check that your config follows [Node settings](#node-settings).
6. Create the unit as in [systemd unit](#systemd-unit), then start it:

   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable --now exrpd
   ```

7. [Check the service](#check-the-service).

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
   Cosmovisor resolves `COSMOVISOR_CUSTOM_PREUPGRADE` under `<node home>/cosmovisor/`. If
   that script is in `genesis/`, `upgrades/` or `current/`, move it to
   `<node home>/cosmovisor/` first, and set the variable to its path relative to that
   directory, for example `COSMOVISOR_CUSTOM_PREUPGRADE=pre-upgrade.sh`. An absolute path
   does not work: Cosmovisor appends it to that directory.

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
   copy the Cosmovisor `ExecStart`. Drop `DAEMON_NAME`, `DAEMON_HOME`,
   `DAEMON_ALLOW_DOWNLOAD_BINARIES` and `DAEMON_RESTART_AFTER_UPGRADE`: `exrpd start` sets
   them (see [Settings](#settings)). Keep the other Cosmovisor settings that you use, for
   example `COSMOVISOR_CUSTOM_PREUPGRADE`, or `DAEMON_DATA_BACKUP_DIR` with
   `UNSAFE_SKIP_BACKUP=false` on a non-validator node. Validators should drop
   `UNSAFE_SKIP_BACKUP=false` (see [Backups during upgrades](#backups-during-upgrades)).
   Then start the service:

   ```sh
   sudo systemctl daemon-reload
   sudo systemctl enable --now <unit>
   ```

6. [Check the service](#check-the-service).

> [!WARNING]
> Never copy the bundle into `<node home>/cosmovisor`, and never replace a directory
> there with a symlink. The bundle manages `upgrades/`, `current` and `supervisor/` there.

### Working directory restriction

If stock Cosmovisor launches the bundle as its selected binary, `start` refuses before
installing files or changing `current`. Stop the external Cosmovisor and run the installed
bundle directly, for example `/usr/local/bin/exrpd start --home <node home>`.

The check compares paths, not the parent process. It refuses when the directory of the
invoked executable is the same directory as `current/bin` under the working directory,
and the invoked file is this bundle. Symlinks are followed. The check does not require
the working directory to be named `cosmovisor` or inspect a Cosmovisor configuration.

This also rejects a direct start from an unrelated versioned installation. For example,
with `/opt/exrpd/current -> releases/1`, starting `/opt/exrpd/current/bin/exrpd` from
`/opt/exrpd` is refused. Using `/opt/exrpd/releases/1/bin/exrpd` from the same directory
does not avoid the check.

After stopping any external Cosmovisor, run the bundle from a directory without
`current/bin`. Use absolute paths for the executable, `--home`, and other path-valued
flags. For example, if `/current/bin` does not exist:

```sh
cd /
/opt/exrpd/current/bin/exrpd start --home <node home>
```

For systemd, under the same condition, set `WorkingDirectory=/` in `[Service]` and keep
an absolute `ExecStart` path. Check `systemctl cat <unit>` for drop-ins that override
`WorkingDirectory`, remove or update those overrides, and run `sudo systemctl daemon-reload`
before restarting the service.

A manual start through the selected binary path from the Cosmovisor working directory
is also refused. Starting through an installed path outside that directory's
`current/bin` still works. Custom wrappers that change the invocation path or working
directory can bypass this check. Do not use this workaround to nest the bundle under
an external Cosmovisor.

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
    restart: unless-stopped
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
- Set a restart policy, for example `restart: unless-stopped`. The node needs a restart to
  recover from some failures at the upgrade height.
- Run other commands in the same container, for example
  `docker exec <container> exrpd keys list`. Then they run as the same user and, with the
  default home or a home set by an environment variable, on the same home. If the service
  passes `--home`, pass the same `--home`.

To move a running container to a bundle image, follow the same timing as for a host (see
the warning in [Choose your path](#choose-your-path)). Before you recreate the container,
change the image in the Compose file and check the home with the volumes and environment
of the service:

```sh
docker compose run --rm --no-deps --entrypoint exrpd node version --bundle
```

Add `--home` if the service uses one. Check `refused` and `will_run.dir` as in step 2 of
[Move from the standalone exrpd](#move-from-the-standalone-exrpd).

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
> A failed `COSMOVISOR_CUSTOM_PREUPGRADE` script does not block the next start.
> The recommended systemd service restarts automatically after a failure, and the bundle
> starts the upgrade binary without retrying the script. For backup failures, see
> [Backups during upgrades](#backups-during-upgrades).

## Restart, reinstall and restore

> [!WARNING]
> A chain only moves forward. Never start an older release or bundle on a home that a
> newer one has run. Never put back an older validator signing-state file.

### Restart

The bundle refuses every command if `data/upgrade-info.json.batch` exists, even if it
contains `[]`. With the node stopped, delete that file. If the node is below the upgrade
height and `data/upgrade-info.json` names a plan that the chain has not reached, delete
that file too. A batch upgrade or `cosmovisor add-upgrade --upgrade-height` stages such a
file early.

If `cosmovisor/current` exists, it must be a symbolic link. A regular file or directory
there stops every command before the bundle reads the data file.

A restart needs no extra step. On every start, the bundle picks the `exrpd` for the chain
data in this order:

1. If `data/upgrade-info.json` exists, the bundle runs the plan that it names. `v11.3.0`
   runs 11.3.0, and `v12.0.0` runs 12.0.0. The chain writes this file when it halts at an
   upgrade height, so after the upgrade the file names `v12.0.0`. A file with another
   plan stops the start with an error, see
   [Problems with the data file](#problems-with-the-data-file).
2. If there is no data file, the bundle follows `<node home>/cosmovisor/current`. If it
   points at a bundle plan directory, the bundle uses that plan, even if the directory
   was deleted. For stock Cosmovisor's `genesis`, it uses the plan that `genesis`
   resolves to, or the bundle's genesis plan if it cannot match it. For another directory,
   it uses its plan record or a matching `upgrades/<plan>` name. If none matches a bundle
   plan, the command stops with an error.
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

In a container, change the image tag instead: `peersyst/exrp:bundle-11.3.0-12.0.1` before
the upgrade height, `peersyst/exrp:bundle-12.0.1` after it. Then follow the
[container recreation steps](#run-the-node-in-a-container).

### Resync after the upgrade

Use state sync or a snapshot from after the upgrade height. Run commands one at a time
as an administrator. If a command fails, stop and do not start the node.

The reset clears the configured signing state, even outside `data/`. Restore the saved
state before starting a validator, or it can sign the same height twice.

1. Stop the node and create a backup directory. Set `home` to the absolute path from
   the service configuration, not `~`:

   ```sh
   home="<node home>"
   sudo systemctl stop <unit>
   sudo -u <node user> mkdir "$home/resync-backup"
   ```

   If `resync-backup` exists, stop. Do not overwrite either signing-state copy until
   you resolve the earlier resync. States at the same height can still differ.

2. Back up the signing state. Read `priv_validator_state_file` in
   `$home/config/config.toml` and resolve relative paths under `$home`:

   ```sh
   state_file="<absolute signing-state path>"
   sudo -u <node user> cp "$state_file" "$home/resync-backup/priv_validator_state.json"
   ```

   If `data/upgrade-info.json` exists, save it too. It preserves the upgrade plan if
   `current` still points at 11.3.0:

   ```sh
   sudo -u <node user> cp "$home/data/upgrade-info.json" "$home/resync-backup/"
   ```

3. Reset the databases. First run
   `sudo -u <node user> /usr/local/bin/exrpd version --home "$home"`.
   If it reports `restart-needed` or `not-installed`, run
   `sudo -u <node user> /usr/local/bin/exrpd bundle setup --home "$home"`.

   > **Custom paths:** Check `db_dir` in `$home/config/config.toml`. Resolve relative
   > paths under `$home`. The default is `data`; the application database always
   > stays in `$home/data`.
   >
   > If the signing-state file is in a subdirectory of `db_dir`, the reset deletes
   > its parent and cannot save the reset state. A file directly in `db_dir` is
   > unaffected by this problem, including the default `data/priv_validator_state.json`.
   > After the backup, move the configured signing-state file outside both `db_dir`
   > and `$home/data`, into an existing directory writable by the node user.
   > Update `priv_validator_state_file` and `state_file` to the new absolute path.
   > Do not use `resync-backup` as the new location or change its saved copy.

   ```sh
   sudo -u <node user> /usr/local/bin/exrpd comet unsafe-reset-all --home "$home" --keep-addr-book
   ```

   Check the reset output for `Error removing`: deletion failures do not make the
   command fail. Run `sudo -u <node user> ls -A "$home/data"`. If `db_dir` is `data`,
   only the configured signing-state file may remain. If other files remain, stop
   and fix the file ownership.

   If `db_dir` points elsewhere, check that directory is empty too. The reset left
   the application database in `$home/data`; clear it before restoring. The files
   needed below are saved in `resync-backup`. `-H` follows a symlink at `$home/data`:

   ```sh
   sudo -u <node user> find -H "$home/data" -mindepth 1 -delete
   ```

4. Unpack the snapshot, if used, then restore the saved files. A snapshot can contain
   an older signing state. For state sync, restore these files before starting:

   ```sh
   sudo -u <node user> cp "$home/resync-backup/priv_validator_state.json" "$state_file"
   # Run only if the upgrade data file was saved:
   sudo -u <node user> cp "$home/resync-backup/upgrade-info.json" "$home/data/"
   ```

5. Check `data/upgrade-info.json`, even if you did not save one. A snapshot's stale
   `v11.3.0` record overrides `current` and selects the wrong binary in the two-binary
   bundle. Keep a `v12.0.0` record. Correct any other plan to `v12.0.0` and its upgrade
   height, as in [Problems with the data file](#problems-with-the-data-file).

   Run `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home "$home"`.
   Check that `refused` is `null` and `will_run.dir` is `upgrades/v12.0.0`.

6. Start the node:

   ```sh
   sudo systemctl start <unit>
   ```

   After the node runs and signs again, remove the backup:
   `sudo -u <node user> rm -r "$home/resync-backup"`.

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

1. Before restoring the snapshot, save the latest file configured by `priv_validator_state_file`,
   using the path resolution and backup steps in [Resync after the upgrade](#resync-after-the-upgrade).
   Restore the snapshot, then restore the saved signing-state file. With it, the validator
   does not sign again the heights that it already signed.
2. Delete `<node home>/cosmovisor/current`. It points at 12.0.0, and 12.0.0 cannot run
   data from before the upgrade.
3. Confirm that the restored snapshot requires 11.3.0. Being before the 12.0.0
   upgrade is not enough: an older snapshot may require an earlier binary.
   If you cannot confirm this, stop.
   Inspect `data/upgrade-info.json` if it exists:
   - If it names `v11.3.0`, keep it.
   - If it names `v12.0.0`, delete it. Otherwise, the bundle selects 12.0.0.
   - If it names an older plan, such as `v11.2.0`, confirm that the snapshot
     already passed the v11.3.0 upgrade. Then correct the record to v11.3.0
     and its upgrade height, following [Problems with the data file](#problems-with-the-data-file).
     Do not simply delete an older record.
   - For any other plan or an unreadable file, stop and investigate.
4. If the node runs `exrpd-bundle_12.0.x`, install `exrpd-bundle_11.3.0-12.0.x` instead,
   as in [Install a new bundle file](#install-a-new-bundle-file). Only this bundle has
   11.3.0.
5. Run `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <node home>`. Go
   on only if `refused` is `null` and `will_run.dir` is `upgrades/v11.3.0`.
6. Start the service. The bundle runs 11.3.0 when the data file names `v11.3.0`,
   or when both the data file and `current` are absent.

The node then catches up as the network did. It gets the missing blocks from its peers
and applies them with 11.3.0. At the upgrade height, 11.3.0 halts, and Cosmovisor
switches to 12.0.0, as in [At the upgrade height](#at-the-upgrade-height). 12.0.0 then
runs the upgrade and catches up to the latest block.

After the node runs and signs again, remove the backup:
`sudo -u <node user> rm -r <node home>/resync-backup`.

### Problems with the data file

If `data/upgrade-info.json` is corrupted, every command stops with an error that explains
the fix. Stop the node. Check its last local block height and upgrade logs. Use the block
explorer or governance proposal to find the exact plan name and height for the upgrade
this node reached, not simply the network's latest upgrade. If the node halted at upgrade
height H, use that plan and height H even if its last committed block is H-1.

Write the file again with those values. Do not only delete it. The bundle then runs the
binary for that plan:

```text
{"name":"<plan>","height":<height>}
```

Replace `<plan>` with the exact plan name and `<height>` with the numeric height. The
plan name can differ from the binary version.

If the file names a plan older than the plans of this bundle, for example `v11.2.0`,
check how this node passed the later plans. If you swapped in their binaries before the
plan height, the chain did not write the file again, so the file is old. Write it again
with the last plan that the node passed and its height. Otherwise, run the release for
the plan in the file.

### Leave the bundle

1. Run `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <node home>`, and
   note `will_run.version`. After the next step, no bundle command can tell you this.
2. Stop the service, and install the standalone `exrpd` of that version at the same path.
   In a container, change the image to `peersyst/exrp:v<version>`.
3. Start the service. Before the upgrade height, you must now swap to the next release by
   hand at the upgrade height.
4. Optional: delete `upgrades/`, `supervisor/`, `current` and `bundle.sha256` in
   `<node home>/cosmovisor/`. Keep your own files there.

### Remove old files

After installing a new bundle, start it once and check that the node runs. Then stop the
service. Run `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <node home>`
and inspect its `dirs` list. If the command fails, delete nothing.

Delete only subdirectories of `cosmovisor/upgrades/` and `cosmovisor/supervisor/` that
the list does not name, and the old `cosmovisor/genesis/` if present. Never delete the
directory that `current` points at. Keep `current` and your own files, including any
custom pre-upgrade script. Start the service again.

With the node stopped and no other `exrpd` command running, you can also delete leftover
`.exrpd-bundle-*` temporary files under `cosmovisor/`.

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
| `UNSAFE_SKIP_BACKUP` | `true` | disables automatic backups. Validators should keep this default. See [Backups during upgrades](#backups-during-upgrades) |

All other Cosmovisor variables keep their upstream meaning. Do not use Cosmovisor batch
upgrades or `cosmovisor add-upgrade --upgrade-height`. See [Restart](#restart) for the
batch-file refusal and recovery.

## Backups during upgrades

The bundle disables automatic backups by default with `UNSAFE_SKIP_BACKUP=true`.
Validators should keep this setting so they can complete the upgrade and resume consensus
as soon as possible, without waiting for a database copy.

Enable backups on non-validator nodes with `UNSAFE_SKIP_BACKUP=false`. Optionally set
`DAEMON_DATA_BACKUP_DIR`. The copy needs free disk space equal to the size of `data/`,
and the upgraded binary waits for the copy to finish. Do not stop the node during the
copy, and do not use a partial backup for recovery. Operators can also use external
backup methods.

If a backup fails, the recommended systemd service automatically restarts the bundle.
The bundle then starts the new binary without retrying the backup. Operators must check
their nodes and backups, investigate failures, and confirm that backups are complete and
usable.

This behavior is intentional. Validators prioritize the upgrade, while multiple
non-validator nodes provide recovery backups. Recovery does not depend on one node: a
failed backup on one node does not, by itself, compromise chain integrity or stop the
network upgrade. A verified backup from another node can provide recovery data.

## Rollback

After the upgrade, 11.3.0 cannot open the migrated data. Recovery needs a verified backup
from before the upgrade, produced by a non-validator node or an external backup method.
Test that the backup restores before relying on it. To restore it on a bundle home, follow
[Restore a snapshot from before the upgrade](#restore-a-snapshot-from-before-the-upgrade).

> [!WARNING]
> If you restore a backup, preserve your own node and validator keys. Keep your
> validator's latest file configured by `priv_validator_state_file`, not the copy from
> the backup. An older signing state lets the validator sign heights it already signed.
> That is double signing.

## Known limitations

- About 100 MB to download (linux/amd64). Installing writes about 300 MB. Files of
  earlier bundles stay on disk. See [Remove old files](#remove-old-files).
- No Windows bundle. Use the standalone `exrpd` there.
- One upgrade per bundle. Install the bundle for the next plan before its height.
- For nodes only. On a client machine, a home without node data keeps the genesis binary
  after the upgrade. Use the standalone `exrpd` there.
- No sync from genesis with a released bundle: it carries two plans, not the whole chain
  history.
- Run the node under systemd or in a container. Without a service manager, a leftover
  `exrpd` can outlive Cosmovisor.
- The supervisor is a patched Cosmovisor until the fix is available upstream.

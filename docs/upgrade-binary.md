# exrpd upgrade bundle

The upgrade bundle is an `exrpd` for validators and other nodes that must switch binaries
at the upgrade height together with the rest of the network. It runs 11.3.0, and at the
`v12.0.0` plan height it stops 11.3.0 and starts 12.0.0 by itself, so the chain upgrades
fast. Nothing is downloaded. Machines that only run client commands (keys, transactions,
queries) against a node do not need it: use the standalone `exrpd` there, and update it
by hand after the upgrade. For details and internals, see the
[developer notes](upgrade-binary-dev.md).

| Inside | Origin |
| --- | --- |
| genesis `exrpd` 11.3.0, plan `v11.3.0` | exact bytes of the GitHub release `v11.3.0` archive |
| upgrade `exrpd` 12.0.0, plan `v12.0.0` | the same bytes as the standalone 12.0.0 release binary |
| Cosmovisor v1.7.3 | `cosmossdk.io/tools/cosmovisor@v1.7.3` plus one signal-handling patch |

What it writes in the node home:

```text
<home>/cosmovisor/genesis/bin/exrpd                    11.3.0, for a home without data/upgrade-info.json
<home>/cosmovisor/upgrades/v11.3.0/bin/exrpd           11.3.0
<home>/cosmovisor/upgrades/v12.0.0/bin/exrpd           12.0.0
<home>/cosmovisor/supervisor/<sha256[:16]>/cosmovisor  Cosmovisor
<home>/cosmovisor/current -> the directory of the binary that runs the data
```

It uses the same home, keys, config and service command as the standalone `exrpd`. See
[Standalone binary or bundle](upgrade-binary-dev.md#standalone-binary-or-bundle). Install
one or the other on a machine, not both.

## Install

Download `exrpd-upgrade_11.3.0-12.0.0_<Os>_<arch>.tar.gz` for your platform and
`exrpd-upgrade_11.3.0-12.0.0_checksums.txt` into the same directory. Then:

```sh
os=$(uname -s)                     # Linux or Darwin
arch=$(uname -m)
[ "$arch" = x86_64 ] && arch=amd64
[ "$arch" = aarch64 ] && arch=arm64
archive=exrpd-upgrade_11.3.0-12.0.0_${os}_${arch}.tar.gz
sha256=sha256sum
[ "$os" = Darwin ] && sha256='shasum -a 256'
grep " $archive\$" exrpd-upgrade_11.3.0-12.0.0_checksums.txt | $sha256 -c - &&
  tar -xzf "$archive" exrpd LICENSE &&
  sudo install -m 0755 exrpd /usr/local/bin/exrpd &&
  exrpd version --bundle               # versions, plans, SHA-256, Cosmovisor source
```

The `&&` chain stops at a checksum mismatch or a missing checksum line, so nothing is
installed. On macOS, the binaries are not signed or notarized. Remove the quarantine
attribute (`xattr -d com.apple.quarantine`) after download. The file names hold the genesis
and upgrade versions. If your network runs another genesis release, its announcement
gives the file names to use.

Installing the file does not change a node that runs now. To move a running node to the
bundle, follow [If you run the standalone exrpd](#if-you-run-the-standalone-exrpd) or
[If you use or used Cosmovisor](#if-you-use-or-used-cosmovisor).

## Run it

```text
exrpd start [--home DIR] [exrpd start flags...]   # install files, then run Cosmovisor
exrpd version [--home DIR]                         # installed version, genesis, upgrade, Cosmovisor
exrpd <any other command> [--home DIR] ...         # runs with the exrpd that runs the node data
```

- Run every command as the node user, not as root. Cosmovisor fails at the upgrade height
  when root owns `cosmovisor/`. Write the full path, because some systems (RHEL, Rocky)
  leave `/usr/local/bin` out of the `PATH` of `sudo`. Example for the unit below:
  `sudo -u exrpd /usr/local/bin/exrpd keys list --home /var/lib/exrpd`.
- A new home: `exrpd init`, keys and the other set-up commands run the 11.3.0 genesis
  binary. After the upgrade, the same commands run 12.0.0.
- Keep the CometBFT RPC server on and `node` in `config/client.toml` pointed at it. The
  defaults do this.
- Home, flags and macOS: see [Commands](upgrade-binary-dev.md#commands) and
  [launchd](upgrade-binary-dev.md#macos-launchd).

systemd unit:

```ini
[Unit]
Description=XRPL EVM node (exrpd upgrade bundle)
After=network-online.target
Wants=network-online.target

[Service]
User=exrpd
ExecStart=/usr/local/bin/exrpd start --home /var/lib/exrpd
Restart=on-failure
RestartSec=5
LimitNOFILE=65535
# Default KillMode=control-group: on stop, and whenever Cosmovisor exits on its own, every
# process left in the unit's cgroup is stopped too, so no exrpd outlives the service.
TimeoutStopSec=120

[Install]
WantedBy=multi-user.target
```

Keep the default `KillMode=control-group`. Do not use `KillMode=process` or `none`.

## If you run the standalone exrpd

A standalone `exrpd` that runs now keeps running from memory after you install the bundle
over it. At the upgrade height it halts and stays halted, because no Cosmovisor watches it.
Restart it on the bundle before the upgrade:

0. Do this at least one day before the expected upgrade time, or after the upgrade. To
   know which, compare the upgrade height from the governance proposal with this node's
   height: `curl -s localhost:26657/status | jq -r .result.sync_info.latest_block_height`.
1. Download and check the files as in [Install](#install), but do not install yet: run
   all lines of that snippet before `grep`, then check and unpack only. Then put a
   copy where the node user can run it (not in `/tmp`, which hardened hosts mount
   `noexec`):

   ```sh
   grep " $archive\$" exrpd-upgrade_11.3.0-12.0.0_checksums.txt | $sha256 -c - &&
     tar -xzf "$archive" exrpd LICENSE &&
     sudo install -d -m 0755 /usr/local/lib/exrpd-check &&
     sudo install -m 0755 exrpd /usr/local/lib/exrpd-check/exrpd
   ```

2. Check the setup with that copy and the same home as your service:
   `sudo -u <node user> /usr/local/lib/exrpd-check/exrpd version --bundle --home <home>`.
   It only reads files. Go on only if both are true:
   - `refused` is `null`. If it is not, fix what it says first, for example a stale
     `data/upgrade-info.json`.
   - `will_run.dir` is `genesis` or `upgrades/<genesis plan>` before the upgrade, and
     `upgrades/v12.0.0` after it. `<genesis plan>` is `manifest.genesis.plan` in the same
     output (`v11.3.0` for this bundle). If it shows anything else:
     - Before the upgrade, `upgrades/v12.0.0` means a staged or stale file that names
       `v12.0.0`. Delete `data/upgrade-info.json` and check again.
     - After the upgrade, any other value means this home has no file that names
       `v12.0.0` (for example after a state sync). Do not move: keep the standalone
       12.0.x, and move to the bundle of the next plan before its height.

   Until step 3, `/usr/local/bin/exrpd` is still the standalone: if you stop here, run
   `sudo rm -r /usr/local/lib/exrpd-check`, and the standalone keeps running and
   restarting as before.
3. Install the bundle:
   `sudo install -m 0755 exrpd /usr/local/bin/exrpd && sudo rm -r /usr/local/lib/exrpd-check`.
4. Switch the service to the bundle with the systemd unit above, with your user and home
   (`ExecStart=/usr/local/bin/exrpd start --home <home>`):
   - Your unit is `exrpd.service`: replace its file, then
     `sudo systemctl daemon-reload && sudo systemctl enable exrpd && sudo systemctl restart exrpd`.
   - Your unit has another name: stop and disable it first,
     `sudo systemctl disable --now <old unit>`. A unit that stays enabled starts again at
     the next boot, next to the bundle, on the same home. Then install the unit above as
     `exrpd.service` and run `sudo systemctl daemon-reload && sudo systemctl enable --now exrpd`.
5. Check that the bundle runs: the service log shows `exrpd: home <home> (from --home)`,
   and `systemctl status exrpd` shows `cosmovisor` with an `exrpd` child.

## If you use or used Cosmovisor

The bundle manages `<home>/cosmovisor` itself, so you do not stage binaries or run
`cosmovisor add-upgrade` any more. Move a home that runs, or once ran, Cosmovisor like
this:

0. Do this at least one day before the expected upgrade time. If less than a day is left,
   keep your current setup for this upgrade and move after it: at the upgrade height the
   chain writes the data file that tells the bundle to run 12.0.0. Also check this while
   the old node still runs: if the node is past the upgrade height from the governance
   proposal (see its height with
   `curl -s localhost:26657/status | jq -r .result.sync_info.latest_block_height`) and
   `sudo -u <node user> cat <home>/data/upgrade-info.json` does not name `v12.0.0` (after a
   state sync, or a binary swapped before the height), do not move to this bundle. Keep
   your current setup, and move to the bundle of the next plan before its height.
1. Stop and disable the old service: `sudo systemctl disable --now <old unit>`, often
   `cosmovisor.service`. A unit that stays enabled starts again at the next boot, next to
   the bundle, on the same home. If the old unit is `exrpd.service`, replace that file in
   step 5.
2. With the node stopped, read the data file as the node user:
   `sudo -u <node user> cat <home>/data/upgrade-info.json`. Nothing writes it now. Delete
   `<home>/data/upgrade-info.json.batch` if it exists (Cosmovisor batch upgrades): every
   command of the bundle stops while it exists. Then find your case:
   - You move before the upgrade (at least a day before it, step 0): delete the data file
     if there is one. Before the upgrade height the chain has not written a `v12.0.0` file,
     so a file there names an earlier plan, or a plan you staged with
     `cosmovisor add-upgrade --upgrade-height` or `cosmovisor add-batch-upgrade`. The
     genesis binary runs this data, and the chain writes the file again at the upgrade
     height.
   - You move after the upgrade: the file names `v12.0.0` (step 0 stops you otherwise).
     Keep it, also if you once staged that upgrade: the chain wrote it again at the
     upgrade height, and it tells the bundle to run 12.0.0.

   In every case, remove `<home>/cosmovisor/genesis/`, `<home>/cosmovisor/upgrades/` and
   `<home>/cosmovisor/current`. Other files in `<home>/cosmovisor/` stay.
3. [Install](#install) the bundle.
4. Check the setup before you start it, with the same home as your service:
   `sudo -u <node user> /usr/local/bin/exrpd version --bundle --home <home>`. Check that
   `home` is your node's home and that `will_run.dir` names the binary for your case:
   `upgrades/v12.0.0` when the data file names `v12.0.0`, else `genesis`. `refused` must
   be `null`. `will_run.installed` is `false` here: the first start installs it.
5. Install the systemd unit above with your user and home, then
   `sudo systemctl daemon-reload && sudo systemctl enable --now exrpd`. Remove old
   `DAEMON_*` settings: `exrpd start` refuses `DAEMON_ALLOW_DOWNLOAD_BINARIES=true`.

Never copy the bundle into `<home>/cosmovisor`, and never replace a directory there with a
symlink: the bundle manages `genesis/`, `upgrades/`, `current` and `supervisor/` there.

## At the upgrade height

You do not need to do anything, if the service runs the bundle: after you install it,
restart the service once (see
[If you run the standalone exrpd](#if-you-run-the-standalone-exrpd)). The chain halts at
the `v12.0.0` plan height, and Cosmovisor stops 11.3.0 and starts 12.0.0 under the same
Cosmovisor process. Each `exrpd` version runs as a separate child process, so its PID
changes. If the node halts and does not switch, restart the service.

> **Warning:** On restart, the bundle selects 12.0.0 because `data/upgrade-info.json`
> names `v12.0.0`. It does not check that the Cosmovisor backup
> (`UNSAFE_SKIP_BACKUP=false`) or `COSMOVISOR_CUSTOM_PREUPGRADE` completed, and it does not
> run them again. The restart starts the migration, and you cannot undo it. If you enabled
> the backup or a custom pre-upgrade script, find why it failed and complete the backup
> and checks manually before you restart. See
> [Operational caveats](upgrade-binary-dev.md#operational-caveats).

## Restart, reinstall and restore

A chain only moves forward. Never start an older release or bundle on a home that a newer
one has run, and never put back an older `data/priv_validator_state.json`.

| Case | Do this | Then runs |
| --- | --- | --- |
| Restart before the upgrade height | nothing | 11.3.0 |
| Restart after the upgrade, or stopped or killed at the halt | nothing | 12.0.0 |
| Reinstall, or install a 12.0 patch release (for example 12.0.1) | restart. Changed binaries are replaced and logged | the installed binaries |
| Resync after the upgrade (`unsafe-reset-all`, then state sync or a snapshot taken after the upgrade height) | keep `priv_validator_state.json` and `data/upgrade-info.json`, see [the steps](upgrade-binary-dev.md#resync-after-the-upgrade) | 12.0.0 |
| New node joining after the upgrade height (state sync, or a snapshot from after the upgrade) | use the standalone 12.0.x, not this bundle. Install the bundle of the next plan before its height | the standalone 12.0.x. The next bundle then starts on its genesis with no extra step |
| Rollback to a snapshot from before the upgrade height | with the node stopped, delete `<home>/cosmovisor/current` and a `data/upgrade-info.json` that names `v12.0.0` | 11.3.0 |
| `data/upgrade-info.json` is corrupted | every command stops with an error that explains the fix: write the file again with the exact plan name and height of the last upgrade, from the governance proposal. Do not just delete it | the binary for that plan |
| `data/upgrade-info.json` names a plan older than this bundle's (for example `v11.2.0`) | if this node applied a later plan with a binary swapped in before the plan height, the file is stale: write it again with the last plan the node crossed and its height. Otherwise use the release for that plan | the binary for that plan |
| Leave the bundle | stop the service, install the standalone `exrpd` that matches the running version | the standalone `exrpd` |

Details and edge cases:
[Restart, resync and restore details](upgrade-binary-dev.md#restart-resync-and-restore-details).

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
| `UNSAFE_SKIP_BACKUP` | `true` | upstream copies all of `<home>/data` at the upgrade height while the validator is down |

All other Cosmovisor variables keep their upstream meaning. Do not use Cosmovisor batch
upgrades or `cosmovisor add-upgrade --upgrade-height`. Every command stops while
`data/upgrade-info.json.batch` exists. See
[Settings details](upgrade-binary-dev.md#settings-details).

## Rollback

Keep a rollback point before the upgrade height. By default nothing copies `data/`, and
11.3.0 cannot open the data after the upgrade. Take a snapshot of `data/` with the node
stopped and test that it restores. You can also enable the Cosmovisor backup, see
[Cosmovisor backup](upgrade-binary-dev.md#cosmovisor-backup).

> **Warning:** If you restore a backup, keep the newer `data/priv_validator_state.json`.
> An older copy lets the validator sign heights it already signed. That is double
> signing.

## Known limitations

- About 100 MB to download (linux/amd64). Installing writes about 400 MB, because the
  genesis binary is stored twice. Files of earlier bundles stay on disk. After each new
  bundle, remove them as described in
  [Removing old files](upgrade-binary-dev.md#removing-old-files).
- No Windows bundle. Use the standalone `exrpd` there.
- One upgrade per bundle. Install the bundle for the next plan before its height.
- For nodes only. On a client machine, a home without node data keeps the genesis binary
  after the upgrade. Use the standalone `exrpd` there.
- No sync from genesis: the bundle carries two plans, not the whole chain history.
- Without a service manager, a leftover `exrpd` can outlive Cosmovisor. See
  [Without a service manager](upgrade-binary-dev.md#without-a-service-manager).
- The supervisor is a patched Cosmovisor until the fix is available upstream. See
  [Cosmovisor patch](upgrade-binary-dev.md#cosmovisor-patch).

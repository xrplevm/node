# exrpd upgrade binary

`exrpd-upgrade` is one executable that runs your node across the `v12.0.0` chain upgrade.
At the upgrade height it stops 11.3.0 and starts 12.0.0 by itself. Nothing is downloaded.
For details and internals, see the [developer notes](upgrade-binary-dev.md).

| Inside | Origin |
| --- | --- |
| genesis `exrpd` 11.3.0 | exact bytes of the GitHub release `v11.3.0` archive |
| upgrade `exrpd` 12.0.0 | the same bytes as the standalone 12.0.0 release binary |
| Cosmovisor v1.7.3 | `cosmossdk.io/tools/cosmovisor@v1.7.3` plus one signal-handling patch |

What it writes in the node home:

```text
<home>/cosmovisor/genesis/bin/exrpd                    11.3.0
<home>/cosmovisor/upgrades/v12.0.0/bin/exrpd           12.0.0
<home>/cosmovisor/supervisor/<sha256[:16]>/cosmovisor  Cosmovisor
<home>/cosmovisor/current -> genesis or upgrades/v12.0.0
```

It uses the same home, keys and config as the standalone `exrpd`. See
[Standalone binary or upgrade binary](upgrade-binary-dev.md#standalone-binary-or-upgrade-binary).

## Install

Download `exrpd-upgrade_12.0.0_<Os>_<arch>` for your platform and
`exrpd-upgrade_12.0.0_checksums.txt` into the same directory. Then:

```sh
os=$(uname -s)                     # Linux or Darwin
arch=$(uname -m)
[ "$arch" = x86_64 ] && arch=amd64
[ "$arch" = aarch64 ] && arch=arm64
binary=exrpd-upgrade_12.0.0_${os}_${arch}
sha256=sha256sum
[ "$os" = Darwin ] && sha256='shasum -a 256'
grep " $binary\$" exrpd-upgrade_12.0.0_checksums.txt | $sha256 -c - &&
  sudo install -m 0755 "$binary" /usr/local/bin/exrpd-upgrade
exrpd-upgrade info                      # manifest: versions, plan, SHA-256, Cosmovisor source
```

The `&&` chain stops at a checksum mismatch or a missing checksum line, so nothing is
installed. On macOS, the binaries are not signed or notarized. Remove the quarantine
attribute (`xattr -d com.apple.quarantine`) after download.

## Run it

```text
exrpd-upgrade start [--home DIR] [exrpd start flags...]   # stage files, then run Cosmovisor
exrpd-upgrade stage [--home DIR] [--select genesis|v12.0.0]
exrpd-upgrade info
```

- Run `start` and `stage` as the node user, not as root. Cosmovisor fails at the upgrade
  height when root owns `cosmovisor/`. Example for the unit below:
  `sudo -u exrpd exrpd-upgrade stage --home /var/lib/exrpd`.
- For CLI commands (keys, queries, transactions), use `<home>/cosmovisor/current/bin/exrpd`.
- Keep the CometBFT RPC server on and `node` in `config/client.toml` pointed at it. The
  defaults do this.
- Home, flags and macOS: see [Commands](upgrade-binary-dev.md#commands) and
  [launchd](upgrade-binary-dev.md#macos-launchd).

systemd unit:

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

Keep the default `KillMode=control-group`. Do not use `KillMode=process` or `none`.

## At the upgrade height

You do not need to do anything. The chain halts at the `v12.0.0` plan height, and
Cosmovisor stops 11.3.0 and starts 12.0.0 in the same process. If the node halts and does
not switch, restart the service.

## Restart, reinstall and restore

Stop the node before you run `stage`.

| Case | Do this | Then runs |
| --- | --- | --- |
| Restart before the upgrade height | nothing | 11.3.0 |
| Restart after the upgrade, or stopped or killed at the halt | nothing | 12.0.0 |
| Reinstall, or install a 12.0 patch release (for example 12.0.1) | restart. Changed binaries are replaced and logged | the installed binaries |
| Data restored or state-synced at or after the upgrade height | `exrpd-upgrade stage --home <home> --select v12.0.0` | 12.0.0. Without it, 11.3.0 can fail with `version of store icahost mismatch` on every start |
| Data restored from before the upgrade height | delete a kept `data/upgrade-info.json` that names `v12.0.0` (it wins over `--select genesis`), then `exrpd-upgrade stage --home <home> --select genesis` | 11.3.0 |
| `<home>/cosmovisor` deleted on data restored or state-synced past the upgrade height | run the `--select v12.0.0` command again | 12.0.0 |
| `data/upgrade-info.json` is not valid JSON | stop the node, delete the file, start again | `start` and `stage` stop with an error until then |
| Leave the upgrade binary | stop the service, run the standalone `exrpd` that matches `current` | the standalone `exrpd` |

Details and edge cases:
[Restart and restore details](upgrade-binary-dev.md#restart-and-restore-details).

## Settings

`exrpd-upgrade start` sets these Cosmovisor variables. It refuses to start if you set a
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
upgrades or `cosmovisor add-upgrade --upgrade-height`. See
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

- About 100 MB to download (linux/amd64). Staging writes about 300 MB.
- No Windows upgrade binary. Use the standalone `exrpd` there.
- One upgrade per bundle. Install the bundle for the next plan before its height.
- Without a service manager, a leftover `exrpd` can outlive Cosmovisor. See
  [Without a service manager](upgrade-binary-dev.md#without-a-service-manager).
- The supervisor is a patched Cosmovisor until the fix is available upstream. See
  [Cosmovisor patch](upgrade-binary-dev.md#cosmovisor-patch).

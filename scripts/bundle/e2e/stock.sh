# shellcheck shell=bash
# Copy of main before the upgrade, set up by stock Cosmovisor v1.7.3: init, add-upgrade of the
# genesis plan, an older plan and the operator's own build of the upgrade. The validator stops
# it before the upgrade height and starts the bundle. Commands and init refuse until then. The
# bundle replaces the operator's binary, keeps the rest and switches at the upgrade height.

use_home stock "$TB"
cv=$H/cosmovisor
rm -rf "$cv"
GOBIN=$E2E_DIR/stock-cv GOWORK=off go install "$COSMOVISOR_MODULE/cmd/cosmovisor@$COSMOVISOR_VERSION" || die "build stock Cosmovisor"
SCV=$E2E_DIR/stock-cv/cosmovisor
stock_env=(DAEMON_HOME="$H" DAEMON_NAME=exrpd DAEMON_ALLOW_DOWNLOAD_BINARIES=false DAEMON_RESTART_AFTER_UPGRADE=true
	UNSAFE_SKIP_BACKUP=true DAEMON_GRPC_ADDRESS=127.0.0.1:$((P + 90)))
# OLD_PLAN is a dummy older plan, far below every real plan of the chain.
OLD_PLAN=v0.1.0
printf '#!/bin/sh\necho old\n' >"$E2E_DIR/old-exrpd" && chmod 0755 "$E2E_DIR/old-exrpd"
# The operator's own build of the upgrade: runnable, other bytes than the bundle's.
{ cat "$UPGRADE_BIN" && echo operator-build; } >"$E2E_DIR/operator-exrpd" && chmod 0755 "$E2E_DIR/operator-exrpd"
env "${stock_env[@]}" "$SCV" init "$GENESIS_BIN" >>"$LOG" 2>&1 || die "cosmovisor init"
env "${stock_env[@]}" "$SCV" add-upgrade "$GENESIS_PLAN" "$GENESIS_BIN" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $GENESIS_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$OLD_PLAN" "$E2E_DIR/old-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $OLD_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$UPGRADE_PLAN" "$E2E_DIR/operator-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $UPGRADE_PLAN"
config_sum=$(sha256sum "$cv/config.toml" | cut -d' ' -f1) || die "cosmovisor init wrote no config.toml"
env --default-signal=INT,QUIT "${stock_env[@]}" "$SCV" run start --home "$H" --chain-id "$CHAIN_ID" >>"$LOG" 2>&1 &
sup=$!
log "stock Cosmovisor on the stock home from main at height $PRE_H, pid $sup"
for ((i = 0; i < 120; i++)); do
	current_is "upgrades/$GENESIS_PLAN" && break
	sleep 0.5
done
wait_height $((PRE_H + 3)) 90 || die "stock Cosmovisor did not run the home"
check "stock Cosmovisor switched to upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "stock Cosmovisor recorded $GENESIS_PLAN" marker_names "$cv/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
stock_h=$(height)
# Unpatched Cosmovisor can mishandle a stop after its in-process switch, so the whole
# group is stopped, as systemd does.
kill_home TERM
[[ -z $(children) ]] || die "exrpd of the stock home still running"
((stock_h < UH)) || die "stock Cosmovisor reached the upgrade height. Raise the proposal margin"
# The bundle never installed this home: commands refuse and init does not take it over.
check "client command on the stock home refuses: bundle not installed" refuses missing "$UB" --home "$H" keys list --keyring-backend test
check "init on the stock home refuses" refuses init "$UB" init localnet --chain-id "$CHAIN_ID" --home "$H"
check "version on the stock home is not-installed" version_is not-installed
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "version --bundle accepts the stock home (refused null, installed_bundle none)" bash -c '"$1" version --bundle --home "$2" | jq -e ".refused == null and .installed_bundle == \"none\"" >/dev/null' _ "$UB" "$H"

start_node
log "bundle on the stock home, supervisor pid $sup"
wait_height $((stock_h + 1)) 60 || die "chain did not start on the stock home"
(($(height) < UH)) || die "the stock home passed the upgrade height before the takeover checks"
check "takeover keeps current -> upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "takeover does not move current" not grep -q "exrpd: moved" "$LOG"
check "takeover replaces the operator's $UPGRADE_PLAN binary and logs it" grep -q "exrpd: replaced $cv/$UPGRADE_DIR/bin/exrpd" "$LOG"
check "takeover keeps identical binaries" not grep -q "exrpd: replaced $cv/upgrades/$GENESIS_PLAN/bin/exrpd" "$LOG"
check "takeover leaves other plan directories and config.toml" test "$("$cv/upgrades/$OLD_PLAN/bin/exrpd")" = old -a "$(sha256sum "$cv/config.toml" | cut -d' ' -f1)" = "$config_sum"
check "takeover runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "takeover writes the marker" bundle_installed
check "after the takeover the chain passes H=$UH" wait_height $((UH + 5)) 180
check "current -> $UPGRADE_DIR after the takeover" current_is "$UPGRADE_DIR"
check "the bundle's upgrade exrpd runs, not the operator's" cmp -s "$UPGRADE_BIN" "$(child_exe)"
check "upgrade applied at H after the takeover" applied_at "$UH"
stop_check "SIGTERM to supervisor after the takeover leaves no exrpd" TERM switched

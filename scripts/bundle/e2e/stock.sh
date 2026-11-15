# shellcheck shell=bash

check "stock Cosmovisor nesting the bundle is refused, direct start installs" \
	env TMPDIR="$E2E_DIR" "$here/nesting-smoke.sh" "$tested_bundle" "$STOCK_COSMOVISOR"

use_home stock "$tested_bundle"
cv=$node_home/cosmovisor
rm -rf "$cv"

stock_env=(DAEMON_HOME="$node_home" DAEMON_NAME=exrpd DAEMON_ALLOW_DOWNLOAD_BINARIES=false DAEMON_RESTART_AFTER_UPGRADE=true
	UNSAFE_SKIP_BACKUP=true DAEMON_GRPC_ADDRESS=127.0.0.1:$((port_base + 90)))
timeout 30 env "${stock_env[@]}" "$STOCK_COSMOVISOR" init "$GENESIS_BIN" >>"$node_log" 2>&1 || die "stock Cosmovisor init"
timeout 30 env "${stock_env[@]}" "$STOCK_COSMOVISOR" add-upgrade "$UPGRADE_PLAN" "$UPGRADE_BIN" >>"$node_log" 2>&1 || die "stock Cosmovisor add-upgrade $UPGRADE_PLAN"
config_sum=$(sha256 "$cv/config.toml") || die "stock Cosmovisor wrote no config.toml"

env --default-signal=INT,QUIT "${stock_env[@]}" "$STOCK_COSMOVISOR" run start --home "$node_home" --chain-id "$CHAIN_ID" >>"$node_log" 2>&1 &
supervisor_pid=$!
log "stock Cosmovisor at pre-upgrade height $pre_upgrade_height, pid $supervisor_pid"
wait_height $((pre_upgrade_height + 3)) 90 || die "stock Cosmovisor did not run the validator"

check "stock Cosmovisor runs the standalone genesis exrpd" running genesis

stop_check "stock validator stops before migration without orphans" TERM
stock_height=$stopped_height
((stock_height < upgrade_height)) || die "stock validator reached the upgrade height; raise the proposal margin"

start_node
log "bundle takes over the stock validator, supervisor pid $supervisor_pid"
wait_height $((stock_height + 1)) 60 || die "bundle did not start the stock validator"
(($(height) < upgrade_height)) || die "stock validator passed the upgrade height before migration checks"

check "migration selects the genesis plan" current_is "upgrades/$GENESIS_PLAN"
check "migration runs the genesis exrpd" running "upgrades/$GENESIS_PLAN"
check "migration preserves the stock config" test "$(sha256 "$cv/config.toml")" = "$config_sum"

check_upgrade "stock validator migration"
stop_check "migrated validator stops after the switch without orphans" TERM

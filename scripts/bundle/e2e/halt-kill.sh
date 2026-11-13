# shellcheck shell=bash

use_home halt-kill "$tested_bundle"
start_node DAEMON_POLL_INTERVAL=1h
log "halt recovery from pre-upgrade height $pre_upgrade_height, supervisor pid $supervisor_pid"
wait_upgrade_halt || die "node did not halt at $UPGRADE_PLAN"

kill_home KILL
check "forced stop happens before the switch" current_is "upgrades/$GENESIS_PLAN"
check "forced stop preserves the planned upgrade marker" marker_names "$node_home/data/upgrade-info.json" "$UPGRADE_PLAN"

start_node
wait_height $((upgrade_height + 5)) 120 || die "restart after the forced stop did not pass H=$upgrade_height"

check "halt recovery selects the upgrade plan" current_is "$UPGRADE_DIR"
check "halt recovery runs the upgrade exrpd" running "$UPGRADE_DIR"
check "halt recovery applies the upgrade at H" applied_at "$upgrade_height"

stop_check "halt recovery stops cleanly without orphans" TERM

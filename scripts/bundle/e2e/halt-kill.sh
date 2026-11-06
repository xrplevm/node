# shellcheck shell=bash
# Copy of main before the upgrade. The supervisor and exrpd are killed at the halt, before
# Cosmovisor switched (a long DAEMON_POLL_INTERVAL holds the switch). Commands and the next
# start must follow the data file to the upgrade, not the stale current.

use_home halt-kill "$TB"
start_node DAEMON_POLL_INTERVAL=1h
log "halt-kill home from main at height $PRE_H, supervisor pid $sup"
for ((i = 0; i < 240; i++)); do
	grep -q "UPGRADE \"$UPGRADE_PLAN\" NEEDED" "$LOG" && break
	sleep 0.5
done
kill_home KILL
sleep 1
log "killed at the halt: current -> $(readlink "$H/cosmovisor/current"), data/upgrade-info.json: $(cat "$H/data/upgrade-info.json" 2>/dev/null)"
check "killed between halt and switch (current -> upgrades/$GENESIS_PLAN, marker names $UPGRADE_PLAN)" \
	halted_before_switch "$UPGRADE_PLAN"
check "commands follow the data file, not the stale current" ran_version_is "$UPGRADE_VERSION"
start_node
log "restarted after the kill, supervisor pid $sup"
check "restart after the kill switches and passes H=$UH" wait_height $((UH + 5)) 120
check "current -> $UPGRADE_DIR after the kill" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "running upgrade exrpd after the kill" running "$UPGRADE_DIR"
check "upgrade applied at H after the kill" applied_at "$UH"
stop_check "SIGTERM to supervisor after the kill leaves no exrpd" TERM

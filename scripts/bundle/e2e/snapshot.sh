# shellcheck shell=bash disable=SC2015 # A && B || die: die exits
# Copy of main before the upgrade, rebuilt as a stock Cosmovisor home restored from a
# snapshot: no data/upgrade-info.json and current -> genesis. Start moves current to the
# genesis plan and writes its record, and Cosmovisor switches at the upgrade height.

use_home snapshot "$TB"
rm "$H/data/upgrade-info.json" "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json"
mkdir -p "$H/cosmovisor/genesis/bin" && cp "$GENESIS_BIN" "$H/cosmovisor/genesis/bin/exrpd" && ln -sfn genesis "$H/cosmovisor/current" ||
	die "set up the stock genesis link"
start_node
log "snapshot home from main at height $PRE_H on genesis, supervisor pid $sup"
wait_height $((PRE_H + 1)) 60 || die "chain did not start on the snapshot home"
check "genesis link moves to upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "the move is logged" grep -q "exrpd: moved $H/cosmovisor/current from genesis to upgrades/$GENESIS_PLAN: current points at genesis" "$LOG"
check "the genesis plan gets its record" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
check "snapshot home runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "snapshot home passes H=$UH" wait_height $((UH + 5)) 180
check "current -> $UPGRADE_DIR on the snapshot home" current_is "$UPGRADE_DIR"
check "running upgrade exrpd on the snapshot home" running "$UPGRADE_DIR"
check "upgrade applied at H on the snapshot home" applied_at "$UH"
stop_check "SIGTERM to supervisor on the snapshot home leaves no exrpd" TERM switched

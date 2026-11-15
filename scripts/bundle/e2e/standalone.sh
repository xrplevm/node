# shellcheck shell=bash disable=SC2015

use_home main "$single_bundle_bin" "$GENESIS_BIN"
init_chain "$GENESIS_BIN" || die "init standalone chain"
start_node
wait_height 3 90 || die "standalone genesis chain did not start"

propose_upgrade
wait_upgrade_halt || die "standalone genesis did not halt at $UPGRADE_PLAN"
kill_home TERM
check "standalone genesis records the planned upgrade" marker_names "$node_home/data/upgrade-info.json" "$UPGRADE_PLAN"

mkdir -p "$E2E_DIR/upgrade-release/bin" && cp "$UPGRADE_BIN" "$E2E_DIR/upgrade-release/bin/exrpd" || die "copy standalone upgrade"
standalone_exrpd=$E2E_DIR/upgrade-release/bin/exrpd

start_node
wait_height $((upgrade_height + 5)) 180 || die "standalone upgrade did not pass the upgrade height"
check "standalone upgrade applied at H" applied_at "$upgrade_height"

stop_check "standalone state preparation stops without orphans" TERM
main_height=$stopped_height

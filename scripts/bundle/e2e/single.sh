# shellcheck shell=bash

# Main stays stopped and unchanged; this journey consumes a copy of its upgraded state.
cp -a "$E2E_DIR/main" "$E2E_DIR/single-home" || die "copy post-upgrade state for the single bundle"
use_home single-home "$single_bundle_bin"
start_node
log "single bundle on post-upgrade state, supervisor pid $supervisor_pid"
wait_height $((main_height + 5)) 90 || die "single bundle did not advance on post-upgrade state"

check "single bundle selects the upgrade plan" current_is "$UPGRADE_DIR"
check "single bundle runs the upgrade exrpd" running "$UPGRADE_DIR"
check "single bundle preserves the applied upgrade height" applied_at "$upgrade_height"
check "single bundle submits a signed bank transaction" send_tx

restart_check "single bundle"

# shellcheck shell=bash

use_home main "$tested_bundle"
init_chain "$active_bundle" || die "init two-binary chain"
start_node
log "two-binary genesis, supervisor pid $supervisor_pid"
wait_height 3 90 || die "two-binary chain did not start"

check "supervisor execs cosmovisor in place" sup_exe_has /cosmovisor/supervisor/
check "fresh start selects the genesis plan" current_is "upgrades/$GENESIS_PLAN"
check "fresh start runs the genesis exrpd" running "upgrades/$GENESIS_PLAN"

propose_upgrade
stop_check "two-binary stops before upgrade without orphans" TERM
((stopped_height < upgrade_height)) || die "stopped at or past the upgrade height; raise the proposal margin"
pre_upgrade_height=$stopped_height
cp -a "$E2E_DIR/main" "$E2E_DIR/halt-kill" || die "save halt recovery home"
cp -a "$E2E_DIR/main" "$E2E_DIR/stock" || die "save stock migration home"

start_node
wait_height $((pre_upgrade_height + 1)) 60 || die "two-binary chain did not restart before upgrade"
(($(height) < upgrade_height)) || die "upgrade height passed before the switch could be observed"

check_upgrade "two-binary upgrade"
check "same supervisor survives the in-process switch" test -d "/proc/$supervisor_pid"
check "no download attempted" not grep -Eiq 'beginning to download|downloading binary' "$node_log"

restart_check "two-binary bundle"
main_height=$stopped_height

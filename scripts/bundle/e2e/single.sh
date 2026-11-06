# shellcheck shell=bash disable=SC2015 # A && B || die: die exits
# Main home after stock: the single bundle of the release is installed over the two-binary
# bundle while the node runs (client commands refuse until the restart), and the restart
# keeps the plan. Copies: a data file that names the genesis plan is refused, and a home
# without a data file and cosmovisor/ (set up from post-upgrade state) runs the upgrade exrpd.

use_home main "$E2E_DIR/bin/exrpd"
cp "$TB" "$UB" || die "copy the tested bundle to $UB for replacement"
start_node
log "main home at height $MAIN_H on the two-binary bundle, supervisor pid $sup"
wait_height $((MAIN_H + 3)) 90 || die "chain did not restart on the main home"
mv "$UB" "$UB.two-binary" && cp "$SB" "$UB" || die "install the single bundle as $UB"
check "client command before the restart refuses: the single bundle differs" refuses differs "$UB" --home "$H" --node "$NODE" status
check "version before the restart is restart-needed" version_is restart-needed
stop_check "SIGTERM to supervisor before the single bundle restart leaves no exrpd" TERM
start_node
log "restarted on the single bundle, supervisor pid $sup"
check "restart on the single bundle keeps $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "current -> $UPGRADE_DIR on the single bundle" current_is "$UPGRADE_DIR"
check "running upgrade exrpd on the single bundle" running "$UPGRADE_DIR"
check "the single bundle runs the standalone $UPGRADE_VERSION bytes" cmp -s "$UPGRADE_BIN" "$(child_exe)"
check "restart writes the single bundle's marker" bundle_installed
check "version on the single bundle is $UPGRADE_VERSION" version_is "$UPGRADE_VERSION"
check "signed bank tx on the single bundle" send_tx
stop_check "SIGTERM to supervisor on the single bundle leaves no exrpd" TERM
SINGLE_H=$last

# A data file that names the genesis plan of the two-binary bundle belongs to a node
# before the upgrade, or is stale. The single bundle runs only its plan, so it refuses
# and changes nothing.
cp -a "$E2E_DIR/main" "$E2E_DIR/single-stale"
use_home single-stale "$SB"
log "single-stale home from main at height $SINGLE_H on the single bundle"
echo "{\"name\":\"$GENESIS_PLAN\",\"time\":\"0001-01-01T00:00:00Z\",\"height\":1}" >"$H/data/upgrade-info.json"
data_before=$(sha256 "$H/data/upgrade-info.json")
check "client command with a $GENESIS_PLAN data file refuses: single bundle" refuses single "$UB" --home "$H" keys list --keyring-backend test
check "start with a $GENESIS_PLAN data file refuses: single bundle" refuses single "$UB" start --home "$H" --chain-id "$CHAIN_ID"
check "the refusals keep the data file" test "$(sha256 "$H/data/upgrade-info.json")" = "$data_before"
# A start that did not refuse leaves its exrpd behind the timeout.
for p in $(children); do kill -KILL "$p" 2>/dev/null; done

# A node set up from post-upgrade state with the single bundle: no data file and no
# cosmovisor/. Start links current to the plan and runs the upgrade exrpd.
cp -a "$E2E_DIR/main" "$E2E_DIR/single-fresh"
use_home single-fresh "$SB"
rm -rf "$H/cosmovisor" "$H/data/upgrade-info.json"
start_node
log "post-upgrade home from main at height $SINGLE_H without a data file on the single bundle, supervisor pid $sup"
check "the single bundle on post-upgrade state without a data file advances the chain" wait_height $((SINGLE_H + 5)) 90
check "current -> $UPGRADE_DIR on the new home" current_is "$UPGRADE_DIR"
check "$UPGRADE_PLAN recorded in its directory" marker_names "$H/cosmovisor/$UPGRADE_DIR/upgrade-info.json" "$UPGRADE_PLAN"
check "running upgrade exrpd on the new home" running "$UPGRADE_DIR"
check "the single bundle installs no other plan and no genesis/" test ! -e "$H/cosmovisor/genesis" -a ! -e "$H/cosmovisor/upgrades/$GENESIS_PLAN"
check "start on the new home writes the marker" bundle_installed
check "upgrade applied at H on the new home" applied_at "$UH"
stop_check "SIGTERM to supervisor on the new home leaves no exrpd" TERM

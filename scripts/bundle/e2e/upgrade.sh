# shellcheck shell=bash
# Main home. With a data file that names the genesis plan, as on validators that ran it, the
# node keeps the genesis exrpd. A governance upgrade to the plan then switches in process.
# Copies of the home before the upgrade height feed halt-kill, stock and snapshot.

echo "{\"name\":\"$GENESIS_PLAN\",\"time\":\"0001-01-01T00:00:00Z\",\"height\":1}" >"$H/data/upgrade-info.json"
start_node
log "supervisor pid $sup"
wait_height 3 90 || die "chain did not start"
check "supervisor execs cosmovisor in place (pid $sup)" sup_exe_has /cosmovisor/supervisor/
check "downloads disabled in supervisor env" sup_env_has DAEMON_ALLOW_DOWNLOAD_BINARIES=false
check "backups off by default in supervisor env" sup_env_has UNSAFE_SKIP_BACKUP=true
check "current -> upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "start writes the marker" bundle_installed
check "$GENESIS_PLAN recorded in its directory" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
# shellcheck disable=SC2016 # $1 to $3 expand in the inner shell
check "version --bundle reports upgrades/$GENESIS_PLAN from --home" bash -c '"$1" version --bundle --home "$2" | jq -e --arg d "$3" ".will_run.dir == \$d and .home_source == \"--home\"" >/dev/null' _ "$UB" "$H" "upgrades/$GENESIS_PLAN"
h0=$(height)
check "v$GENESIS_VERSION keeps producing blocks with a $GENESIS_PLAN marker" wait_height $((h0 + 15)) 60
check "same supervisor still running v$GENESIS_VERSION" sup_running "upgrades/$GENESIS_PLAN"
check "Cosmovisor did not switch" not grep -q "daemon shutting down in an attempt to restart" "$LOG"
check "signed bank tx before upgrade" send_tx

# Governance upgrade.
UH=$(($(height) + 60))
log "proposing $UPGRADE_PLAN at height $UH"
hash=$(cli tx upgrade software-upgrade "$UPGRADE_PLAN" --upgrade-height "$UH" --title e2e --summary e2e \
	--deposit 1axrp --no-validate --from alice --fees "$FEES" --gas 500000 -y -o json | jq -r .txhash) || die "submit proposal"
tx=$(wait_tx "$hash") || die "proposal tx $hash not included"
[[ $(jq -r .code <<<"$tx") == 0 ]] || die "proposal tx $hash failed: $(jq -r .raw_log <<<"$tx")"
pid=$(jq -r '[.events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value][0]' <<<"$tx")
[[ $pid =~ ^[0-9]+$ ]] || die "no proposal id in tx $hash"
cli tx gov vote "$pid" yes --from alice --fees "$FEES" --gas 300000 -y -o json >/dev/null || die "vote"
status=
deadline=$((SECONDS + 60))
while ((SECONDS < deadline)); do
	status=$(timeout 5 "$(cur)" --home "$H" --node "$NODE" query gov proposal "$pid" -o json 2>/dev/null | jq -r '.proposal.status // .status')
	[[ $status == PROPOSAL_STATUS_PASSED || $status == 3 ]] && break
	sleep 1
done
check "proposal $pid passed ($status)" test "$status" = PROPOSAL_STATUS_PASSED -o "$status" = 3

# Stop before the upgrade height and keep copies of the home.
stop_check "SIGTERM to supervisor before upgrade leaves no exrpd" TERM
((last < UH)) || die "stopped at or past the upgrade height. Raise the proposal margin"
PRE_H=$last
cp -a "$E2E_DIR/main" "$E2E_DIR/halt-kill"
cp -a "$E2E_DIR/main" "$E2E_DIR/stock"
cp -a "$E2E_DIR/main" "$E2E_DIR/snapshot"
start_node
log "restarted before the upgrade, supervisor pid $sup"
wait_height $((PRE_H + 1)) 60 || die "chain did not restart"
(($(height) < UH)) || die "upgrade height passed before the switch could be observed"

# In-process switch at the upgrade height.
wait_height $((UH + 5)) 180 || die "chain did not pass upgrade height"
check "blocks advance past H=$UH (now $(height))" wait_height $((UH + 5)) 1
check "current -> $UPGRADE_DIR" current_is "$UPGRADE_DIR"
check "same supervisor still running after switch (one start, pid $sup)" test -d "/proc/$sup"
check "running upgrade exrpd after switch" running "$UPGRADE_DIR"
check "no download attempted" not grep -Eiq 'beginning to download|downloading binary' "$LOG"
check "upgrade applied at H" applied_at "$UH"
check "signed bank tx after upgrade" send_tx
check "version after the switch is $UPGRADE_VERSION" version_is "$UPGRADE_VERSION"

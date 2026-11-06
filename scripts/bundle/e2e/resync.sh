# shellcheck shell=bash
# Copy of main after the upgrade, without data/upgrade-info.json, as after a resync. current
# is then the only evidence of the plan, and start must keep it. The earlier bundle installed
# this home, so commands refuse until `bundle setup`. A corrupted data file is refused.

cp -a "$E2E_DIR/main" "$E2E_DIR/resync"
use_home resync "$TB"
rm "$H/data/upgrade-info.json"
"$PB" bundle setup --home "$H" >>"$LOG" 2>&1 || die "bundle setup of the earlier patch bundle"
check "client command on a home of the earlier bundle refuses" refuses differs "$UB" --home "$H" keys list --keyring-backend test
check "bundle setup installs, keeps current and starts nothing" setup_installs
check "commands without a data file run the upgrade exrpd" ran_version_is "$UPGRADE_VERSION"
start_node
log "resync home from main at height $MAIN_H, supervisor pid $sup"
check "without a data file start keeps $UPGRADE_PLAN and chain advances" wait_height $((MAIN_H + 5)) 90
check "current -> $UPGRADE_DIR without a data file" current_is "$UPGRADE_DIR"
check "running upgrade exrpd without a data file" running "$UPGRADE_DIR"
stop_check "SIGTERM to supervisor on the resync home leaves no exrpd" TERM
echo '{"name":"v12.0' >"$H/data/upgrade-info.json"
bins_before=$(sha256sum "$H"/cosmovisor/upgrades/*/bin/exrpd)
check "corrupted data file is refused" fails_quietly "$UB" version --home "$H"
check "start refuses a corrupted data file" start_refused
check "current and binaries unchanged after the refusal" \
	test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR" -a "$(sha256sum "$H"/cosmovisor/upgrades/*/bin/exrpd)" = "$bins_before"

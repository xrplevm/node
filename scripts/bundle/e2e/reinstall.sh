# shellcheck shell=bash disable=SC2015 # A && B || die: die exits
# Main home after the upgrade. Stops and restarts keep the plan. The home is then reinstalled
# from an earlier patch bundle, and this bundle is installed over it while the node runs:
# client commands refuse until the restart. Also builds the single bundle for single.sh.

stop_check "SIGTERM to supervisor after upgrade leaves no exrpd" TERM switched

start_node
log "restarted, supervisor pid $sup"
check "restart keeps $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 60
check "running upgrade exrpd after restart" running "$UPGRADE_DIR"
stop_check "SIGTERM to supervisor and exrpd after restart leaves no exrpd" TERM group

# The builds below reuse the tested bundle's Cosmovisor.
TESTED_CV=$H/cosmovisor/supervisor/$(manifest cosmovisor.sha256 | cut -c1-16)/cosmovisor
TESTED_CV_SOURCE=$(manifest cosmovisor.source)
# An earlier patch release of the bundle: the same genesis exrpd and Cosmovisor, another
# build of the upgrade exrpd, so another manifest.
PREV_UPGRADE_BIN=$E2E_DIR/prev-upgrade-exrpd
PB=$E2E_DIR/prev/exrpd
{ cat "$UPGRADE_BIN" && echo earlier-patch; } >"$PREV_UPGRADE_BIN" && chmod 0755 "$PREV_UPGRADE_BIN" || die "write $PREV_UPGRADE_BIN"
"$here/build.sh" "${build_args[@]}" --upgrade-exrpd "$PREV_UPGRADE_BIN" --out "$PB" \
	--cosmovisor "$TESTED_CV" --cosmovisor-source "$TESTED_CV_SOURCE" || die "build the earlier patch bundle"

# The single bundle of the upgrade release: its genesis is the upgrade exrpd, packed as
# that release's archive, and it runs the upgrade plan.
SB=$E2E_DIR/single/exrpd
if [[ -n $single_bundle ]]; then
	mkdir -p "$(dirname "$SB")" && cp "$single_bundle" "$SB" || die "copy $single_bundle"
else
	sr=$E2E_DIR/single-release sa=node_${UPGRADE_VERSION}_Linux_amd64.tar.gz
	mkdir -p "$sr/bin" && cp "$UPGRADE_BIN" "$sr/bin/exrpd" && tar -C "$sr" -czf "$sr/$sa" bin/exrpd &&
		echo "$(sha256 "$sr/$sa")  $sa" >"$sr/checksums.txt" || die "pack the upgrade exrpd as release $UPGRADE_VERSION"
	"$here/build.sh" --genesis-version "$UPGRADE_VERSION" --genesis-plan "$UPGRADE_PLAN" --genesis-archive "$sr/$sa" \
		--genesis-checksums "$sr/checksums.txt" --goos linux --goarch amd64 --out "$SB" \
		--cosmovisor "$TESTED_CV" --cosmovisor-source "$TESTED_CV_SOURCE" || die "build the single bundle"
fi
check "single manifest has no upgrade" not has_upgrade "$SB"
check "single manifest genesis plan is $UPGRADE_PLAN" test "$(manifest genesis.plan "$SB")" = "$UPGRADE_PLAN"
check "single manifest genesis is the standalone $UPGRADE_VERSION exrpd" test "$(manifest genesis.sha256 "$SB")" = "$(sha256 "$UPGRADE_BIN")"
check "single manifest launcher commit is set" launcher_commit_is_set "$SB"

# Reinstall from the earlier patch bundle on a home without cosmovisor/.
UB=$E2E_DIR/bin/exrpd
cp "$TB" "$UB" || die "copy the tested bundle to $UB for replacement"
rm -rf "$H/cosmovisor"
mv "$UB" "$UB.patch" && cp "$PB" "$UB" || die "install the earlier patch bundle as $UB"
start_node
log "reinstalled the earlier patch bundle (deleted $H/cosmovisor), supervisor pid $sup"
check "reinstall after H selects $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "current -> $UPGRADE_DIR after reinstall" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "running upgrade exrpd after reinstall" running "$UPGRADE_DIR"
check "reinstall writes the marker" bundle_installed
check "the earlier patch bundle installed its upgrade exrpd" cmp -s "$PREV_UPGRADE_BIN" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"
# The operator installs this bundle while the node runs. Until the restart, the files of
# the home belong to the earlier bundle.
mv "$UB.patch" "$UB" || die "install the patch bundle as $UB"
check "client command before the restart refuses: installed bundle differs" refuses differs "$UB" --home "$H" --node "$NODE" status
check "version before the restart is restart-needed" version_is restart-needed
check "version --bundle before the restart reports installed_bundle other" installed_is other
stop_check "SIGTERM to supervisor after reinstall leaves no exrpd" TERM
start_node
log "restarted on the patch bundle, supervisor pid $sup"
check "restart on the patch bundle keeps $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "restart runs the patch bundle's upgrade exrpd" cmp -s "$UPGRADE_BIN" "$(child_exe)"
check "restart writes the patch bundle's marker" bundle_installed
check "client command after the restart runs" bundle_status
check "version after the restart is $UPGRADE_VERSION" version_is "$UPGRADE_VERSION"
stop_check "SIGTERM to supervisor after the patch restart leaves no exrpd" TERM
MAIN_H=$last

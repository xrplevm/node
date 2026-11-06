# shellcheck shell=bash
# Main home. Before `exrpd init`, commands install nothing and name the next step. init
# installs the bundle, and the first start without a data file links current to the genesis plan.

check "help before init refuses and writes nothing" refuses none "$UB" --home "$H" --help
check "client command before init refuses and writes nothing" refuses none "$UB" --home "$H" keys list --keyring-backend test
check "TAB completion before init suggests nothing" completes_nothing
check "version before init is not-installed" version_is not-installed
check "version --bundle before init reports installed_bundle none" installed_is none
check "nothing written before init" test ! -e "$H"
init_chain "$UB" || die "init chain"
check "init installs the bundle and writes the marker" bundle_installed
# init installs the binaries without creating `current`. Before the first start, every
# command runs the genesis exrpd.
check "command before the first start runs genesis" ran_version_is "$GENESIS_VERSION"
# exrpd's completion also reads $HOME/config.toml, so the check uses a scratch HOME: a
# config.toml of another tool in the developer's home must not fail it.
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "completion passes the typed words to exrpd" env HOME="$E2E_DIR/genesis-release" bash -c '"$1" __completeNoDesc --home "$2" ke | grep -qx keys' _ "$UB" "$H"
check "no current before the first start" test ! -e "$H/cosmovisor/current" -a ! -L "$H/cosmovisor/current"
check "nothing installed in genesis/" test ! -e "$H/cosmovisor/genesis"
check "installed $GENESIS_PLAN exrpd = release bytes" cmp -s "$GENESIS_BIN" "$H/cosmovisor/upgrades/$GENESIS_PLAN/bin/exrpd"
check "installed upgrade exrpd = standalone $UPGRADE_VERSION bytes" cmp -s "$UPGRADE_BIN" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"
check "version before the first start is $GENESIS_VERSION" version_is "$GENESIS_VERSION"
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "help runs the installed exrpd" bash -c '"$1" --home "$2" --help | grep "Available Commands" >/dev/null' _ "$UB" "$H"

# First start: no data/upgrade-info.json and no current. Start links current to the genesis
# plan and writes its record, so the bundle needs no genesis/ directory.
start_node
log "first start without a data file, supervisor pid $sup"
wait_height 3 90 || die "chain did not start"
check "first start links current -> upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "first start records $GENESIS_PLAN" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
check "first start runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "first start creates no genesis/" test ! -e "$H/cosmovisor/genesis"
stop_check "SIGTERM to supervisor after the first start leaves no exrpd" TERM

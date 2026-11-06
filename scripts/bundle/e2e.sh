#!/usr/bin/env bash
# Localnet E2E of the exrpd bundle on linux/amd64. One single-validator chain upgrades from the
# genesis release to the upgrade exrpd through the bundle. The scenarios are in e2e/ and run in
# the order below, each one on the state that the one before left.
#
# Usage: e2e.sh --genesis-archive node_X.Y.Z_Linux_amd64.tar.gz --genesis-checksums checksums.txt --upgrade-exrpd PATH
#               [--upgrade-version X.Y.Z] [--plan vX.Y.Z] [--bundle PATH] [--single-bundle PATH]
#               [--unpatched --cosmovisor-archive PATH --cosmovisor-checksums PATH]
#   --upgrade-version  version of --upgrade-exrpd, default the plan without "v". Set it to test a patch release
#   --plan             the upgrade plan, default the handler of this checkout (plan.sh checkout)
#   --bundle           test this two-binary bundle instead of building one, for example a release asset
#   --single-bundle    test this single bundle instead of building one. Its genesis must be --upgrade-exrpd
#   --unpatched        embed the upstream Cosmovisor release (--cosmovisor-archive, checked against
#                      --cosmovisor-checksums). Then the stop after the switch must show the defect
# Env: E2E_DIR scratch directory, deleted at start (default .e2e). PORT_BASE first of 100 ports (default 29000).
# Every run also builds an earlier patch bundle, and the single bundle without --single-bundle.
# shellcheck source-path=SCRIPTDIR disable=SC2015 # A && B || die: die exits, so it runs only when A or B fails.
set -uo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$here/lib.sh"

mode=patched genesis_archive='' genesis_checksums='' upgrade_exrpd='' upgrade_version='' plan='' bundle='' single_bundle='' cosmovisor_archive='' cosmovisor_checksums=''
while [[ $# -gt 0 ]]; do
	case $1 in
	--unpatched) mode=unpatched; shift; continue ;;
	--genesis-archive) genesis_archive=$2 ;;
	--genesis-checksums) genesis_checksums=$2 ;;
	--upgrade-exrpd) upgrade_exrpd=$2 ;;
	--upgrade-version) upgrade_version=$2 ;;
	--plan) plan=$2 ;;
	--bundle) bundle=$2 ;;
	--single-bundle) single_bundle=$2 ;;
	--cosmovisor-archive) cosmovisor_archive=$2 ;;
	--cosmovisor-checksums) cosmovisor_checksums=$2 ;;
	*) echo "unknown argument $1" >&2; exit 2 ;;
	esac
	shift 2
done
required=(genesis_archive genesis_checksums upgrade_exrpd)
[[ $mode == unpatched ]] && required+=(cosmovisor_archive cosmovisor_checksums)
for v in "${required[@]}"; do
	[[ -n ${!v} ]] || { echo "missing --${v//_/-}" >&2; exit 2; }
done
if [[ $mode == unpatched ]]; then
	[[ -z $bundle$single_bundle ]] || { echo "--bundle and --single-bundle cannot be used with --unpatched" >&2; exit 2; }
	[[ $(basename "$cosmovisor_archive") == "cosmovisor-$COSMOVISOR_VERSION-linux-amd64.tar.gz" ]] ||
		{ echo "--cosmovisor-archive must be cosmovisor-$COSMOVISOR_VERSION-linux-amd64.tar.gz, the pinned release" >&2; exit 2; }
elif [[ -n $cosmovisor_archive$cosmovisor_checksums ]]; then
	echo "--cosmovisor-archive and --cosmovisor-checksums need --unpatched" >&2
	exit 2
fi
[[ $(basename "$genesis_archive") =~ ^node_(.+)_Linux_amd64\.tar\.gz$ ]] ||
	{ echo "--genesis-archive must be named node_<X.Y.Z>_Linux_amd64.tar.gz" >&2; exit 2; }
GENESIS_VERSION=${BASH_REMATCH[1]}
UPGRADE_PLAN=${plan:-$("$here/plan.sh" checkout)} || exit 1
is_plan "$UPGRADE_PLAN" || { echo "--plan $UPGRADE_PLAN is not vX.Y.Z" >&2; exit 2; }
# The bundle passes the home to Cosmovisor and exrpd as a cleaned absolute path
# (filepath.Abs). The pgrep patterns in e2e/lib.sh must see the same string.
E2E_DIR=$(realpath -ms "${E2E_DIR:-$here/../../.e2e}")
P=${PORT_BASE:-29000}

CHAIN_ID=xrplevm_1449999-1
MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"
FEES=1000000000000000axrp
NODE=tcp://127.0.0.1:$((P + 57))
TB=$E2E_DIR/tested/exrpd
UPGRADE_VERSION=${upgrade_version:-${UPGRADE_PLAN#v}}
# An absolute path, so that a bare file name runs this file and not a command on PATH.
UPGRADE_BIN=$(realpath -ms "$upgrade_exrpd")
UPGRADE_DIR=upgrades/$UPGRADE_PLAN
fails=0
sup=

source "$here/e2e/lib.sh"
use_home main "$TB"
# Without these, a signal ends bash without the EXIT trap and leaves the node running.
trap 'exit 130' INT
trap 'exit 143' TERM

# ---------------------------------------------------------------- setup
log "mode: $mode, scratch: $E2E_DIR, ports: $P-$((P + 99))"
inputs=("$genesis_archive" "$genesis_checksums" "$UPGRADE_BIN" ${bundle:+"$bundle"} ${single_bundle:+"$single_bundle"})
[[ $mode == unpatched ]] && inputs+=("$cosmovisor_archive" "$cosmovisor_checksums")
for f in "${inputs[@]}"; do
	[[ -f $f ]] || die "missing input $f"
done
# The lock file is next to $E2E_DIR because the run deletes and recreates $E2E_DIR.
mkdir -p "$(dirname "$E2E_DIR")" || die "create the parent of $E2E_DIR"
exec {lock_fd}>"$E2E_DIR.lock" || die "open $E2E_DIR.lock"
flock -n "$lock_fd" || die "another e2e.sh run uses $E2E_DIR (lock $E2E_DIR.lock)"
for p in $(seq "$P" $((P + 99))); do
	(echo >"/dev/tcp/127.0.0.1/$p") 2>/dev/null && die "port $p already in use"
done
# The run deletes $E2E_DIR and cleanup kills every exrpd under it, so it only takes a
# directory that is missing, empty or marked as scratch by an earlier run.
scratch_marker=$E2E_DIR/.e2e-scratch
if [[ -e $E2E_DIR && ! -e $scratch_marker && -n $(ls -A "$E2E_DIR" 2>/dev/null) ]]; then
	die "$E2E_DIR is not empty and has no $(basename "$scratch_marker"): set E2E_DIR to a scratch directory, or delete it"
fi
trap cleanup EXIT
rm -rf "$E2E_DIR" && mkdir -p "$E2E_DIR/bin" "$E2E_DIR/tested" "$E2E_DIR/genesis-release" && touch "$scratch_marker" || die "create $E2E_DIR"

# The scenarios run in this order on one chain and share these variables: H, LOG, UB
# (use_home), sup (the running supervisor), last (height at the last stop), TB, SB, PB (the
# tested, single and earlier patch bundles), GENESIS_PLAN, GENESIS_BIN, UH (upgrade height),
# PRE_H and MAIN_H. A scenario that is missing or does not parse stops the run. Each
# scenario ends with a command that returns 0, so a nonzero status means it did not run.
source "$here/e2e/bundles.sh" || die "e2e/bundles.sh did not run"
source "$here/e2e/new-home.sh" || die "e2e/new-home.sh did not run"
source "$here/e2e/upgrade.sh" || die "e2e/upgrade.sh did not run"
source "$here/e2e/reinstall.sh" || die "e2e/reinstall.sh did not run"
source "$here/e2e/resync.sh" || die "e2e/resync.sh did not run"
source "$here/e2e/halt-kill.sh" || die "e2e/halt-kill.sh did not run"
source "$here/e2e/snapshot.sh" || die "e2e/snapshot.sh did not run"
source "$here/e2e/stock.sh" || die "e2e/stock.sh did not run"
source "$here/e2e/single.sh" || die "e2e/single.sh did not run"

echo
if ((fails > 0)); then
	echo "E2E ($mode): $fails check(s) FAILED"
	exit 1
fi
echo "E2E ($mode): all checks PASSED"

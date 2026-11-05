#!/usr/bin/env bash
# Usage: e2e.sh --genesis-archive node_X.Y.Z_Linux_amd64.tar.gz --genesis-checksums checksums.txt
#               --upgrade-exrpd PATH --upgrade-version X.Y.Z --plan vX.Y.Z
#               [--bundle PATH --stock-cosmovisor PATH] [--single-bundle PATH]
# At least one bundle is required. All executables are supplied artifacts; nothing is built.
# Env: E2E_DIR scratch directory, deleted at start (default .e2e). PORT_BASE first of 100 ports (default 29000).
# shellcheck source-path=SCRIPTDIR disable=SC2015 # A && B || die: die exits, so it runs only when A or B fails.
set -uo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$here/lib.sh"
source "$here/cosmovisor/cosmovisor.env"

genesis_archive='' genesis_checksums='' upgrade_exrpd='' upgrade_version='' plan='' bundle='' single_bundle='' stock_cosmovisor=''

while [[ $# -gt 0 ]]; do
	case $1 in
	--genesis-archive | --genesis-checksums | --upgrade-exrpd | --upgrade-version | --plan | --bundle | --single-bundle | --stock-cosmovisor)
		[[ $# -ge 2 && -n $2 && $2 != --* ]] || { echo "missing value for $1" >&2; exit 2; }
		;;
	*) echo "unknown argument $1" >&2; exit 2 ;;
	esac

	case $1 in
	--genesis-archive) genesis_archive=$2 ;;
	--genesis-checksums) genesis_checksums=$2 ;;
	--upgrade-exrpd) upgrade_exrpd=$2 ;;
	--upgrade-version) upgrade_version=$2 ;;
	--plan) plan=$2 ;;
	--bundle) bundle=$2 ;;
	--single-bundle) single_bundle=$2 ;;
	--stock-cosmovisor) stock_cosmovisor=$2 ;;
	esac
	shift 2
done

required=(genesis_archive genesis_checksums upgrade_exrpd upgrade_version plan)
[[ -n $bundle$single_bundle ]] || { echo "supply --bundle or --single-bundle" >&2; exit 2; }
[[ -n $bundle ]] && required+=(stock_cosmovisor)

for v in "${required[@]}"; do
	[[ -n ${!v} ]] || { echo "missing --${v//_/-}" >&2; exit 2; }
done

[[ $(basename "$genesis_archive") =~ ^node_(.+)_Linux_amd64\.tar\.gz$ ]] ||
	{ echo "--genesis-archive must be named node_<X.Y.Z>_Linux_amd64.tar.gz" >&2; exit 2; }
GENESIS_VERSION=${BASH_REMATCH[1]}
UPGRADE_PLAN=$plan
is_plan "$UPGRADE_PLAN" || { echo "--plan $UPGRADE_PLAN is not vX.Y.Z" >&2; exit 2; }

# The bundle passes the home to Cosmovisor and exrpd as a cleaned absolute path
# (filepath.Abs). The pgrep patterns in e2e/lib.sh must see the same string.
E2E_DIR=$(realpath -ms "${E2E_DIR:-$here/../../.e2e}")
port_base=${PORT_BASE:-29000}

CHAIN_ID=xrplevm_1449999-1
MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"
FEES=1000000000000000axrp
node_rpc=tcp://127.0.0.1:$((port_base + 57))
tested_bundle=$E2E_DIR/tested/exrpd
UPGRADE_VERSION=$upgrade_version

# An absolute path, so that a bare file name runs this file and not a command on PATH.
UPGRADE_BIN=$(realpath -ms "$upgrade_exrpd")
UPGRADE_DIR=upgrades/$UPGRADE_PLAN
fails=0
supervisor_pid=

source "$here/e2e/lib.sh"
use_home main "$tested_bundle"

# Without these, a signal ends bash without the EXIT trap and leaves the node running.
trap 'exit 130' INT
trap 'exit 143' TERM

log "scratch: $E2E_DIR, ports: $port_base-$((port_base + 99))"
inputs=("$genesis_archive" "$genesis_checksums" "$UPGRADE_BIN" ${bundle:+"$bundle"} ${single_bundle:+"$single_bundle"} ${stock_cosmovisor:+"$stock_cosmovisor"})
for f in "${inputs[@]}"; do
	[[ -f $f ]] || die "missing input $f"
done

for f in "$UPGRADE_BIN" ${bundle:+"$bundle"} ${single_bundle:+"$single_bundle"} ${stock_cosmovisor:+"$stock_cosmovisor"}; do
	[[ -x $f ]] || die "input is not executable: $f"
done
[[ -z $stock_cosmovisor ]] || STOCK_COSMOVISOR=$(realpath -ms "$stock_cosmovisor")

# The lock file is next to $E2E_DIR because the run deletes and recreates $E2E_DIR.
mkdir -p "$(dirname "$E2E_DIR")" || die "create the parent of $E2E_DIR"
exec {lock_fd}>"$E2E_DIR.lock" || die "open $E2E_DIR.lock"
flock -n "$lock_fd" || die "another e2e.sh run uses $E2E_DIR (lock $E2E_DIR.lock)"

for p in $(seq "$port_base" $((port_base + 99))); do
	(echo >"/dev/tcp/127.0.0.1/$p") 2>/dev/null && die "port $p already in use"
done

# The run deletes $E2E_DIR and cleanup kills every exrpd under it, so it only takes a
# directory that is missing, empty or marked as scratch by an earlier run.
scratch_marker=$E2E_DIR/.e2e-scratch
if [[ -e $E2E_DIR && ! -e $scratch_marker && -n $(ls -A "$E2E_DIR" 2>/dev/null) ]]; then
	die "$E2E_DIR is not empty and has no $(basename "$scratch_marker"): set E2E_DIR to a scratch directory, or delete it"
fi

trap cleanup EXIT
rm -rf "$E2E_DIR" && mkdir -p "$E2E_DIR/tested" "$E2E_DIR/genesis-release" && touch "$scratch_marker" || die "create $E2E_DIR"

# Scenarios must return zero after reporting check failures so the remaining scenarios run.
source "$here/e2e/bundles.sh" || die "e2e/bundles.sh did not run"

if [[ -n $bundle ]]; then
	source "$here/e2e/upgrade.sh" || die "e2e/upgrade.sh did not run"
	source "$here/e2e/stock.sh" || die "e2e/stock.sh did not run"
	source "$here/e2e/halt-kill.sh" || die "e2e/halt-kill.sh did not run"
else
	source "$here/e2e/standalone.sh" || die "e2e/standalone.sh did not run"
fi

if [[ -n $single_bundle ]]; then
	source "$here/e2e/single.sh" || die "e2e/single.sh did not run"
fi

echo
if ((fails > 0)); then
	echo "E2E: $fails check(s) FAILED"
	exit 1
fi

echo "E2E: all checks PASSED"

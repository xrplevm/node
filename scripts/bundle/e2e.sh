#!/usr/bin/env bash
# Localnet E2E for the exrpd bundle (linux/amd64).
#
# Usage: e2e.sh --genesis-archive node_<X.Y.Z>_Linux_amd64.tar.gz --genesis-checksums checksums.txt \
#               --upgrade-exrpd PATH [--upgrade-version X.Y.Z] [--plan vX.Y.Z] [--single-bundle PATH] \
#               [--bundle PATH | --unpatched --cosmovisor-archive PATH --cosmovisor-checksums PATH]
#   --genesis-archive       genesis GitHub release asset. Its name gives the genesis version
#   --genesis-checksums     checksums.txt of the genesis release
#   --upgrade-exrpd         standalone upgrade exrpd (executable)
#   --upgrade-version       the upgrade exrpd version, default the plan without "v", or a
#                           patch release of the plan for a patch bundle
#   --plan                  the upgrade plan, default the upgrade handler of this checkout
#                           (plan.sh)
#   --bundle                a bundle exrpd built from the same inputs to test instead of
#                           building one, for example the release asset
#   --single-bundle         a single bundle exrpd of the upgrade release to test instead of
#                           building one: its genesis must be the --upgrade-exrpd bytes and
#                           its plan --plan, for example the release asset
#   --unpatched             embed the unpatched upstream Cosmovisor release binary instead of
#                           the patched build. The run passes only when the SIGTERM after the
#                           in-process upgrade shows the defect (see stop_check)
#   --cosmovisor-archive    with --unpatched: the upstream Cosmovisor release asset
#                           cosmovisor-<COSMOVISOR_VERSION>-linux-amd64.tar.gz
#   --cosmovisor-checksums  with --unpatched: a checksum list that holds its line, such as
#                           SHA256SUMS-cosmovisor-<COSMOVISOR_VERSION>.txt of the release
#
# Env: E2E_DIR (scratch, deleted at start: it must be missing, empty or an earlier run's
# scratch), PORT_BASE (100 ports from here are used). The run always builds an earlier
# patch bundle from this checkout with build.sh, reusing the tested bundle's Cosmovisor.
# Without --single-bundle it builds the single bundle the same way, with --upgrade-exrpd
# as its genesis.
#
# Scenarios on one single-validator chain (homes run one at a time on the same ports):
#   main      genesis home created by `exrpd init`. The first start without a data file
#             links current to the genesis plan. Then a data/upgrade-info.json from the
#             genesis plan, as on the validators that ran it. Before init, commands write
#             nothing and name the next step. Commands before the first start run the
#             genesis exrpd. A gov upgrade to the plan (plan.sh) switches in process, then
#             restart, reinstall from an earlier patch bundle and the install of this
#             bundle over it (client commands refuse until the restart), SIGTERM.
#   resync    post-upgrade copy without data/upgrade-info.json, as after a resync, that
#             the earlier bundle installed: commands refuse until `bundle setup`, and
#             start keeps the upgrade. A corrupted data file is refused.
#   halt-kill pre-upgrade copy whose supervisor and app are SIGKILLed at the upgrade halt,
#             before Cosmovisor switched: commands and the next start use the upgrade.
#   snapshot  pre-upgrade copy rebuilt as a stock home restored from a snapshot: no data
#             file and current -> genesis. Start moves current to the genesis plan, and
#             the node switches at the upgrade height.
#   stock     pre-upgrade copy rebuilt as a stock Cosmovisor home: commands and init
#             refuse, then the bundle takes it over, replaces the operator's binaries and
#             switches at the upgrade height.
#   single    the main home after the stock scenario: the single bundle of the upgrade
#             release installed over the two-binary bundle while the node runs (client
#             commands refuse until the restart), then the restart keeps the plan.
#             Copies of it: a data file naming the genesis plan is refused, and without
#             a data file and cosmovisor/ (a node set up from post-upgrade state) start
#             runs the upgrade exrpd.
# shellcheck disable=SC2015,SC2001 # A && B || die: die exits, so it runs only when A or B fails. sed prefixes each line of multi-line text.
set -uo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/lib.sh
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
UPGRADE_PLAN=${plan:-$("$here/plan.sh")} || exit 1
is_plan "$UPGRADE_PLAN" || { echo "--plan $UPGRADE_PLAN is not vX.Y.Z" >&2; exit 2; }
# The bundle passes the home to Cosmovisor and exrpd as a cleaned absolute path
# (filepath.Abs). The pgrep patterns below must see the same string.
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

log() { echo "--- $*"; }
pass() { echo "PASS: $*"; }
fail() { echo "FAIL: $*"; fails=$((fails + 1)); }
check() { # name, command...
	local name=$1
	shift
	if "$@"; then pass "$name"; else fail "$name"; fi
}
die() { echo "FATAL: $*"; tail -n 40 "$LOG" 2>/dev/null; exit 1; }

# use_home NAME BUNDLE selects the node home, log and launcher every helper below works on.
use_home() { H=$E2E_DIR/$1 LOG=$E2E_DIR/$1.log UB=$2; }
use_home main "$TB"

# re_escape prints $1 with ERE metacharacters escaped, so pgrep -f matches it literally.
re_escape() { sed 's/[][\.*^$+?(){}|]/\\&/g' <<<"$1"; }
children() { pgrep -f "bin/exrpd start --home $(re_escape "$H")( |\$)" || true; }
# cleanup kills every exrpd of a home under $E2E_DIR, also children that the supervisor
# left behind. It is armed only after this run holds the scratch lock and the ports are
# free, so it cannot kill the nodes of another run that uses the same $E2E_DIR.
cleanup() {
	[[ -n $sup ]] && kill -KILL "$sup" 2>/dev/null
	local p
	for p in $(pgrep -f "bin/exrpd start --home $(re_escape "$E2E_DIR/")"); do kill -KILL "$p" 2>/dev/null; done
}
# Without these, a signal ends bash without the EXIT trap and leaves the node running.
trap 'exit 130' INT
trap 'exit 143' TERM

cur() { echo "$H/cosmovisor/current/bin/exrpd"; }
# RPC calls are bounded so a node that holds its RPC socket open without answering cannot
# hang the waits or stop_check. The waits use a wall-clock deadline.
height() { timeout 5 "$(cur)" status --home "$H" --node "$NODE" 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height // 0'; }
wait_height() { # target, timeout seconds
	local deadline=$((SECONDS + $2))
	while ((SECONDS < deadline)); do
		(($(height || echo 0) >= $1)) && return 0
		sleep 1
	done
	return 1
}
cli() { timeout 30 "$(cur)" --home "$H" --node "$NODE" --chain-id "$CHAIN_ID" --keyring-backend test "$@"; }
addr() { "$(cur)" --home "$H" keys show "$1" -a --keyring-backend test; }
balance() { cli query bank balance "$(addr bob)" axrp -o json | jq -r '.balance.amount'; }

# wait_tx HASH prints the transaction as JSON once a block includes it, within 30s.
wait_tx() {
	local deadline=$((SECONDS + 30)) out
	while ((SECONDS < deadline)); do
		out=$(timeout 5 "$(cur)" --home "$H" --node "$NODE" query tx "$1" -o json 2>/dev/null) && [[ -n $out ]] && echo "$out" && return 0
		sleep 1
	done
	return 1
}

# send_tx: bank send alice -> bob and confirm the balance moved and the tx code is 0.
send_tx() {
	local before after hash code
	before=$(balance)
	hash=$(cli tx bank send alice "$(addr bob)" 1000axrp --fees "$FEES" --gas 300000 -y -o json | jq -r .txhash) || return 1
	code=$(wait_tx "$hash" | jq -r .code)
	after=$(balance)
	echo "    tx $hash code=$code bob $before -> $after"
	[[ $code == 0 && $after -eq $((before + 1000)) ]]
}

# start_node runs the upgrade binary in the background like a service manager would:
# bash starts background jobs with SIGINT and SIGQUIT ignored, env restores the defaults
# so the signal checks exercise the supervisor's own handling. Extra args go to env.
start_node() {
	env --default-signal=INT,QUIT DAEMON_GRPC_ADDRESS=127.0.0.1:$((P + 90)) "$@" \
		"$UB" start --home "$H" --chain-id "$CHAIN_ID" >>"$LOG" 2>&1 &
	sup=$!
}

# child_exe prints the resolved executable of the supervisor's exrpd child.
child_exe() {
	local c
	c=$(pgrep -P "$sup" | head -n1) && [[ -n $c ]] && readlink -f "/proc/$c/exe"
}
running() { test "$(child_exe)" = "$(readlink -f "$H/cosmovisor/$1/bin/exrpd")"; }
current_is() { test "$(readlink "$H/cosmovisor/current")" = "$1" -o "$(readlink "$H/cosmovisor/current")" = "$H/cosmovisor/$1"; }
applied_at() { test "$(cli query upgrade applied "$UPGRADE_PLAN" -o json | jq -r '.header.height // .height')" = "$1"; }
not() { ! "$@"; }
# fails_quietly runs a command that must fail, hiding its stderr.
fails_quietly() { ! "$@" 2>/dev/null; }
marker_names() { jq -e --arg n "$2" '.name == $n' "$1" >/dev/null; } # file, plan
# bundle_info [EXRPD] prints the version --bundle JSON of the bundle EXRPD, default $UB,
# for a home that does not exist, so it reads nothing of a real home and writes nothing.
bundle_info() { "${1:-$UB}" version --bundle --home "$E2E_DIR/no-home"; }
cosmovisor_source() { bundle_info | jq -e '.manifest.cosmovisor.source' | grep "$1" >/dev/null; }
# version_is VERSION: exrpd version on this home names VERSION first, as the standalone does.
version_is() { test "$("$UB" version --home "$H" | cut -d' ' -f1)" = "$1"; }
# ran_version_is VERSION: a command passed through to exrpd ran the binary of VERSION.
ran_version_is() { "$UB" --home "$H" version --long | grep -qx "version: $1"; }
# start_refused: the launcher refuses start with its "is corrupted" message, so the
# refusal is its own and not a later Cosmovisor or exrpd failure on the same file. Exit
# status 124 means it was still running after the timeout, so it did not refuse.
start_refused() {
	local out=$E2E_DIR/start-refused.out rc
	timeout 30 "$UB" start --home "$H" >"$out" 2>&1
	rc=$?
	cat "$out" >>"$LOG"
	((rc != 0 && rc != 124)) && grep -q "^exrpd: .* is corrupted" "$out"
}
sup_exe_has() { readlink "/proc/$sup/exe" | grep "$1" >/dev/null; }
sup_env_has() { tr '\0' '\n' <"/proc/$sup/environ" | grep -x "$1" >/dev/null; }
sup_running() { kill -0 "$sup" && running "$1"; }
halted_before_switch() { current_is "upgrades/$GENESIS_PLAN" && marker_names "$H/data/upgrade-info.json" "$1"; }

# home_files lists the files of $H/cosmovisor with type, size, mtime and link target, so a
# check can tell whether a command wrote any of them. It prints "absent" without $H.
home_files() {
	[[ -e $H ]] || { echo absent; return; }
	find "$H/cosmovisor" -printf '%P %y %s %T@ %l\n' 2>/dev/null | sort
}
# says STATE COMMAND...: the launcher answers COMMAND with its message for STATE about $H
# from --home and writes nothing in $H/cosmovisor (or creates no $H). STATE is none (no
# genesis.json and no data/), missing (no marker), differs (marker of another bundle) or
# init (init on an existing home). It sets $said to the exit status of COMMAND.
says() {
	local want out before
	case $1 in
	none) want="No exrpd is installed in $H (from --home)." ;;
	missing) want="The bundle is not installed in $H (from --home)." ;;
	differs) want="The bundle installed in $H (from --home) differs from this exrpd." ;;
	init) want="$H (from --home) already has config/genesis.json or a different bundle installed" ;;
	single) want="is a single bundle: it runs only plan" ;;
	*) echo "    says: unknown state $1"; return 1 ;;
	esac
	shift
	before=$(home_files)
	out=$(timeout 30 "$@" 2>&1)
	said=$?
	echo "$out" >>"$LOG"
	grep -qF "$want" <<<"$out" || { echo "    expected: $want"; sed 's/^/    got: /' <<<"$out"; return 1; }
	test "$(home_files)" = "$before" || { echo "    the command wrote in $H"; return 1; }
}
# refuses STATE COMMAND...: COMMAND fails with the message for STATE and writes nothing.
# Exit status 124 means it was still running after the timeout, so it did not refuse.
refuses() { says "$@" && ((said != 0 && said != 124)); }
installed_is() { "$UB" version --bundle --home "$H" | jq -e --arg s "$1" '.installed_bundle == $s' >/dev/null; }
# bundle_installed: the marker holds a SHA-256 and matches the bundle at $UB.
bundle_installed() { grep -Eqx '[0-9a-f]{64}' "$H/cosmovisor/bundle.sha256" && installed_is this; }
# completes_nothing: TAB completion on a home without this bundle answers with no
# suggestions and the no-file-completion directive, exits 0.
completes_nothing() {
	local out
	out=$(env HOME="$E2E_DIR/genesis-release" "$UB" __completeNoDesc --home "$H" ke) && test "$out" = ":4"
}
# bundle_status: exrpd status through the bundle answers with the node's sync info.
bundle_status() { timeout 5 "$UB" --home "$H" --node "$NODE" status | jq -e '.sync_info // .SyncInfo' >/dev/null; }
# setup_installs: bundle setup installs this bundle and writes the marker, keeps current
# and starts no exrpd.
setup_installs() {
	local cur_before
	cur_before=$(readlink "$H/cosmovisor/current")
	"$UB" bundle setup --home "$H" >>"$LOG" 2>&1 || return
	bundle_installed && test "$(readlink "$H/cosmovisor/current")" = "$cur_before" -a -z "$(children)" &&
		cmp -s "$UPGRADE_BIN" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"
}

# stop_check NAME SIGNAL [switched|group] sends SIGNAL to the supervisor PID only (not
# the group), or with "group" to this home's exrpd as well, the way systemd's default
# KillMode=control-group does, and waits up to 15s for the supervisor and every exrpd of
# this home to exit. It records the last height in $last. The signal-handler defect shows
# as an exrpd left running, a "process already finished" log line from the handler of an
# earlier run, or the supervisor exit status 1 that handler causes. The patched Cosmovisor
# must show none of them. With --unpatched, a stop after an in-process switch (third
# argument "switched") must show the defect: the stale handler log line or an orphan exrpd
# that still produces blocks, and the supervisor exit status 1. Any other status fails.
# Afterwards it stops the supervisor and every exrpd of this home that is left, and aborts
# the run if one of them survives SIGKILL.
stop_check() {
	local name=$1 sig=$2 how=${3:-supervisor} expect=clean left status=running stale mark deadline defect='' signature='' h1='' h2='' p
	[[ $how == switched && $mode == unpatched ]] && expect=defect
	last=$(height || echo 0)
	mark=$(wc -l <"$LOG")
	# The supervisor is signalled first. With "group", the app may exit before the handler
	# forwards the signal, or it may still be shutting down. Either way the supervisor must
	# exit 0. The group case of the patch's TestLaunchProcessStop also exercises this.
	# Signalling the app first can let it exit before the supervisor gets the signal. A
	# signal that arrives after Launcher.Run has unregistered its handler kills Cosmovisor
	# with status 143.
	kill "-$sig" "$sup"
	if [[ $how == group ]]; then
		for p in $(children); do kill "-$sig" "$p" 2>/dev/null; done
	fi
	deadline=$((SECONDS + 15))
	while :; do
		if [[ $status == running ]] && ! kill -0 "$sup" 2>/dev/null; then
			wait "$sup"
			status=$?
		fi
		left=$(children)
		[[ $status != running && -z $left ]] && break
		((SECONDS >= deadline)) && break
		sleep 1
	done
	stale=$(tail -n "+$((mark + 1))" "$LOG" | grep 'process already finished')
	echo "    SIG$sig -> supervisor $sup status=$status; exrpd left: ${left:-none}"
	[[ -n $stale ]] && sed 's/^/    log: /' <<<"$stale"
	if [[ -n $left ]]; then
		h1=$(height)
		sleep 3
		h2=$(height)
		echo "    orphan exrpd pid(s) $left, ppid $(ps -o ppid= -p "${left%%$'\n'*}" | tr -d ' '), height $h1 -> $h2"
		((${h2:-0} > ${h1:-0})) && signature=1
	fi
	[[ -n $stale ]] && signature=1
	[[ -n $left || -n $stale || ($status != running && $status != 0) ]] && defect=1
	if [[ $status == running ]]; then
		fail "$name (supervisor still running)"
	elif [[ $expect == defect ]]; then
		if [[ -n $signature && $status == 1 ]]; then
			pass "SIG$sig after the in-process switch: unpatched Cosmovisor shows the defect"
		elif [[ -n $signature ]]; then
			fail "SIG$sig after the in-process switch: defect signature seen, but supervisor exit status $status is not the 1 of the stale handler"
		else
			fail "SIG$sig after the in-process switch: the unpatched defect did not show in this run, supervisor exit status $status (it depends on which signal handler runs first)"
		fi
	elif [[ -n $defect ]]; then
		fail "$name"
	else
		pass "$name"
	fi
	# Stop what is left so the next scenario starts clean.
	if [[ $status == running ]]; then
		kill -KILL "$sup" 2>/dev/null
		deadline=$((SECONDS + 5))
		while kill -0 "$sup" 2>/dev/null && ((SECONDS < deadline)); do sleep 1; done
		kill -0 "$sup" 2>/dev/null && die "supervisor $sup still running after SIGKILL"
		wait "$sup" 2>/dev/null
	fi
	left=$(children)
	if [[ -n $left ]]; then
		for p in $left; do kill -TERM "$p" 2>/dev/null; done
		deadline=$((SECONDS + 10))
		while [[ -n $(children) ]] && ((SECONDS < deadline)); do sleep 1; done
		for p in $(children); do kill -KILL "$p" 2>/dev/null; done
		deadline=$((SECONDS + 5))
		while [[ -n $(children) ]] && ((SECONDS < deadline)); do sleep 1; done
		left=$(children)
		[[ -n $left ]] && die "exrpd pid(s) $left of $H still running after SIGKILL"
	fi
	sup=
}

set_ports() {
	local c=$H/config/config.toml a=$H/config/app.toml
	sed -i \
		-e "s|^proxy_app = .*|proxy_app = \"tcp://127.0.0.1:$((P + 58))\"|" \
		-e "/^\[rpc\]/,/^\[/ s|^laddr = .*|laddr = \"tcp://127.0.0.1:$((P + 57))\"|" \
		-e "/^\[p2p\]/,/^\[/ s|^laddr = .*|laddr = \"tcp://127.0.0.1:$((P + 56))\"|" \
		-e "s|^pprof_laddr = .*|pprof_laddr = \"127.0.0.1:$((P + 60))\"|" \
		-e "s|^prometheus_listen_addr = .*|prometheus_listen_addr = \"127.0.0.1:$((P + 66))\"|" \
		-e 's|^timeout_commit = .*|timeout_commit = "1s"|' "$c" || return
	sed -i \
		-e "/^\[api\]/,/^\[/ s|^address = .*|address = \"tcp://127.0.0.1:$((P + 17))\"|" \
		-e "/^\[grpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((P + 90))\"|" \
		-e "s|^geth-metrics-address = .*|geth-metrics-address = \"127.0.0.1:$((P + 81))\"|" \
		-e "/^\[json-rpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((P + 45))\"|" \
		-e "s|^ws-address = .*|ws-address = \"127.0.0.1:$((P + 46))\"|" \
		-e "s|^metrics-address = .*|metrics-address = \"127.0.0.1:$((P + 65))\"|" "$a"
}

# init_chain EXRPD creates the home with `EXRPD init`, the first command on a new home,
# and sets up the chain with EXRPD.
init_chain() {
	local exrpd=$1 g=$H/config/genesis.json addr
	"$exrpd" init localnet --chain-id "$CHAIN_ID" --home "$H" >>"$LOG" 2>&1 || return
	"$exrpd" --home "$H" config set client chain-id "$CHAIN_ID" >/dev/null || return
	"$exrpd" --home "$H" config set client keyring-backend test >/dev/null || return
	"$exrpd" --home "$H" config set client node "$NODE" >/dev/null || return
	echo "$MNEMONIC" | "$exrpd" --home "$H" keys add alice --recover --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return
	"$exrpd" --home "$H" keys add bob --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return
	jq '.consensus.params.block.max_gas="10500000"
		| .app_state.evm.params.evm_denom="axrp"
		| .app_state.gov.params.min_deposit[0].denom="axrp"
		| .app_state.gov.params.min_deposit[0].amount="1"
		| .app_state.gov.params.expedited_min_deposit[0].denom="axrp"
		| .app_state.gov.params.expedited_min_deposit[0].amount="2"
		| .app_state.gov.params.voting_period="20s"
		| .app_state.gov.params.expedited_voting_period="10s"
		| .app_state.staking.params.bond_denom="apoa"
		| .app_state.feemarket.params.base_fee="0"
		| .app_state.feemarket.params.no_base_fee=true
		| .app_state.bank.denom_metadata=[{"description":"XRP is the gas token","denom_units":[{"denom":"axrp"},{"denom":"xrp","exponent":18}],"base":"axrp","display":"xrp","name":"XRP","symbol":"XRP"}]
		| .app_state.erc20.native_precompiles=["0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"]
		| .app_state.erc20.token_pairs=[{contract_owner:1,erc20_address:"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",denom:"axrp",enabled:true,"owner_address":"ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2"}]' \
		"$g" >"$g.tmp" && mv "$g.tmp" "$g" || return
	addr=$("$exrpd" --home "$H" keys show alice -a --keyring-backend test) || return
	"$exrpd" --home "$H" genesis add-genesis-account "$addr" 1000000apoa,1000000000000000000000000000axrp --keyring-backend test || return
	"$exrpd" --home "$H" genesis gentx alice 1000000apoa --gas-prices 0axrp --keyring-backend test --chain-id "$CHAIN_ID" >/dev/null 2>&1 || return
	"$exrpd" --home "$H" genesis collect-gentxs >/dev/null 2>&1 || return
	"$exrpd" --home "$H" genesis validate >/dev/null || return
	set_ports
}


# ---------------------------------------------------------------- build
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
build_args=(--genesis-version "$GENESIS_VERSION" --genesis-archive "$genesis_archive" --genesis-checksums "$genesis_checksums"
	--upgrade-exrpd "$UPGRADE_BIN" --upgrade-version "$UPGRADE_VERSION" --plan "$UPGRADE_PLAN" --goos linux --goarch amd64 --out "$UB")
if [[ $mode == unpatched ]]; then
	cv=$(basename "$cosmovisor_archive")
	verify_archive "$cosmovisor_checksums" "$cosmovisor_archive" >/dev/null || die "cosmovisor release checksum"
	mkdir -p "$E2E_DIR/cv" && tar -xzf "$cosmovisor_archive" -C "$E2E_DIR/cv"
	build_args+=(--cosmovisor "$E2E_DIR/cv/cosmovisor" --cosmovisor-source "unpatched upstream release $cv")
fi
IFS=. read -r plan_major plan_minor _ <<<"${UPGRADE_PLAN#v}"
next_minor=$plan_major.$((plan_minor + 1)).0
# build.sh must refuse the version with its major.minor diagnostic. Other build errors,
# for example a missing Go toolchain, do not prove the version check.
build_refuses_next_minor() {
	local out
	if out=$("$here/build.sh" "${build_args[@]}" --upgrade-version "$next_minor" 2>&1); then
		echo "build.sh accepted --upgrade-version $next_minor"
		return 1
	fi
	grep -qF "does not have the major.minor of plan" <<<"$out" && return 0
	echo "build.sh failed without the major.minor diagnostic. Output:"
	echo "$out"
	return 1
}
check "build refuses $next_minor, a minor release that needs its own plan" build_refuses_next_minor
if [[ -n $bundle ]]; then
	cp "$bundle" "$UB" || die "copy $bundle"
else
	"$here/build.sh" "${build_args[@]}" || die "build"
fi
check "manifest cosmovisor source ($mode)" cosmovisor_source "$([[ $mode == patched ]] && echo "$COSMOVISOR_PATCH" || echo unpatched)"

tar -xzf "$genesis_archive" -C "$E2E_DIR/genesis-release" bin/exrpd
GENESIS_BIN=$E2E_DIR/genesis-release/bin/exrpd
check "genesis tool is release v$GENESIS_VERSION" test "$("$GENESIS_BIN" version --home "$E2E_DIR/genesis-release")" = "$GENESIS_VERSION"
# manifest PATH [EXRPD]: the manifest value at PATH of the bundle EXRPD, default $UB.
manifest() { bundle_info "${2:-$UB}" | jq -r ".manifest.$1"; }
# launcher_commit_is_set EXRPD: the manifest names the commit that built the launcher.
launcher_commit_is_set() { manifest launcher.commit "$1" | grep -Eqx '[0-9a-f]{40}(-dirty)?'; }
GENESIS_PLAN=$(manifest genesis.plan)
check "manifest genesis plan $GENESIS_PLAN is the newest plan of v$GENESIS_VERSION" test "$GENESIS_PLAN" = "$("$here/release-plan.sh" "$GENESIS_VERSION")"
check "manifest has one upgrade" test "$(manifest 'upgrades | length')" = 1
check "manifest upgrade plan is $UPGRADE_PLAN" test "$(manifest 'upgrades[0].plan')" = "$UPGRADE_PLAN"
check "manifest genesis is the release v$GENESIS_VERSION exrpd" test "$(manifest genesis.sha256)" = "$(sha256 "$GENESIS_BIN")"
check "manifest upgrade is the standalone $UPGRADE_VERSION exrpd" test "$(manifest 'upgrades[0].sha256')" = "$(sha256 "$UPGRADE_BIN")"
check "manifest launcher commit is set" launcher_commit_is_set "$UB"

# ---------------------------------------------------------------- chain
# Before `exrpd init` the home does not exist. Help, TAB completion, version and client
# commands install nothing, and help and client commands name the next step.
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

# ---------------------------------------------------------------- first start without a data file
# A new node has no data/upgrade-info.json and no current. Start links current to the
# genesis plan and writes its record, so the bundle needs no genesis/ directory.
start_node
log "first start without a data file, supervisor pid $sup"
wait_height 3 90 || die "chain did not start"
check "first start links current -> upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "first start records $GENESIS_PLAN" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
check "first start runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "first start creates no genesis/" test ! -e "$H/cosmovisor/genesis"
stop_check "SIGTERM to supervisor after the first start leaves no exrpd" TERM

# ---------------------------------------------------------------- genesis plan marker
# Validator homes that ran the genesis plan carry data/upgrade-info.json naming it.
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

# ---------------------------------------------------------------- governance upgrade
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

# Stop before the upgrade height and keep copies of the home for the halt-kill and
# stock Cosmovisor scenarios.
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

# ---------------------------------------------------------------- in-process switch
wait_height $((UH + 5)) 180 || die "chain did not pass upgrade height"
check "blocks advance past H=$UH (now $(height))" wait_height $((UH + 5)) 1
check "current -> $UPGRADE_DIR" current_is "$UPGRADE_DIR"
check "same supervisor still running after switch (one start, pid $sup)" test -d "/proc/$sup"
check "running upgrade exrpd after switch" running "$UPGRADE_DIR"
check "no download attempted" not grep -Eiq 'beginning to download|downloading binary' "$LOG"
check "upgrade applied at H" applied_at "$UH"
check "signed bank tx after upgrade" send_tx
check "version after the switch is $UPGRADE_VERSION" version_is "$UPGRADE_VERSION"

# ---------------------------------------------------------------- signals, restart, reinstall
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
has_upgrade() { bundle_info "$1" | jq -e '.manifest | has("upgrades")' >/dev/null; }
check "single manifest has no upgrade" not has_upgrade "$SB"
check "single manifest genesis plan is $UPGRADE_PLAN" test "$(manifest genesis.plan "$SB")" = "$UPGRADE_PLAN"
check "single manifest genesis is the standalone $UPGRADE_VERSION exrpd" test "$(manifest genesis.sha256 "$SB")" = "$(sha256 "$UPGRADE_BIN")"
check "single manifest launcher commit is set" launcher_commit_is_set "$SB"

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

# ---------------------------------------------------------------- resync past H
# A resync (unsafe-reset-all, then state sync) deletes data/upgrade-info.json. `current`
# is then the only evidence of the plan, and start must keep it. The node is stopped for
# the resync, and the earlier bundle installed this home: commands such as
# unsafe-reset-all refuse until `bundle setup` installs this bundle.
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

# ---------------------------------------------------------------- killed at the halt
# A long DAEMON_POLL_INTERVAL keeps Cosmovisor from switching before the kill. The halted
# app exits on its own, so the checks rely on the state on disk: current still on the
# genesis plan and the data file naming the upgrade.
use_home halt-kill "$TB"
start_node DAEMON_POLL_INTERVAL=1h
log "halt-kill home from main at height $PRE_H, supervisor pid $sup"
for ((i = 0; i < 240; i++)); do
	grep -q "UPGRADE \"$UPGRADE_PLAN\" NEEDED" "$LOG" && break
	sleep 0.5
done
kill -KILL "$sup"
for p in $(children); do kill -KILL "$p"; done
wait "$sup" 2>/dev/null
sup=
sleep 1
log "killed at the halt: current -> $(readlink "$H/cosmovisor/current"), data/upgrade-info.json: $(cat "$H/data/upgrade-info.json" 2>/dev/null)"
check "killed between halt and switch (current -> upgrades/$GENESIS_PLAN, marker names $UPGRADE_PLAN)" \
	halted_before_switch "$UPGRADE_PLAN"
check "commands follow the data file, not the stale current" ran_version_is "$UPGRADE_VERSION"
start_node
log "restarted after the kill, supervisor pid $sup"
check "restart after the kill switches and passes H=$UH" wait_height $((UH + 5)) 120
check "current -> $UPGRADE_DIR after the kill" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "running upgrade exrpd after the kill" running "$UPGRADE_DIR"
check "upgrade applied at H after the kill" applied_at "$UH"
stop_check "SIGTERM to supervisor after the kill leaves no exrpd" TERM

# ---------------------------------------------------------------- stock genesis link without a data file
# A node restored from a snapshot under stock Cosmovisor has no data/upgrade-info.json and
# current -> genesis. Start moves current to the genesis plan and writes its record, and
# Cosmovisor switches at the upgrade height from that record.
use_home snapshot "$TB"
rm "$H/data/upgrade-info.json" "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json"
mkdir -p "$H/cosmovisor/genesis/bin" && cp "$GENESIS_BIN" "$H/cosmovisor/genesis/bin/exrpd" && ln -sfn genesis "$H/cosmovisor/current" ||
	die "set up the stock genesis link"
start_node
log "snapshot home from main at height $PRE_H on genesis, supervisor pid $sup"
wait_height $((PRE_H + 1)) 60 || die "chain did not start on the snapshot home"
check "genesis link moves to upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "the move is logged" grep -q "exrpd: moved $H/cosmovisor/current from genesis to upgrades/$GENESIS_PLAN: current points at genesis" "$LOG"
check "the genesis plan gets its record" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
check "snapshot home runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "snapshot home passes H=$UH" wait_height $((UH + 5)) 180
check "current -> $UPGRADE_DIR on the snapshot home" current_is "$UPGRADE_DIR"
check "running upgrade exrpd on the snapshot home" running "$UPGRADE_DIR"
check "upgrade applied at H on the snapshot home" applied_at "$UH"
stop_check "SIGTERM to supervisor on the snapshot home leaves no exrpd" TERM switched

# ---------------------------------------------------------------- takeover of a stock Cosmovisor home
# A validator runs stock Cosmovisor v1.7.3, stops it before the upgrade height and starts
# the bundle on the same home. Stock Cosmovisor sets up the home itself: `init` writes
# genesis/ and config.toml, `add-upgrade` stages the genesis plan, an older plan and the
# operator's own build of the upgrade, and the run switches to upgrades/<genesis plan>
# because data/upgrade-info.json names it, writing its record there.
use_home stock "$TB"
cv=$H/cosmovisor
rm -rf "$cv"
GOBIN=$E2E_DIR/stock-cv GOWORK=off go install "$COSMOVISOR_MODULE/cmd/cosmovisor@$COSMOVISOR_VERSION" || die "build stock Cosmovisor"
SCV=$E2E_DIR/stock-cv/cosmovisor
stock_env=(DAEMON_HOME="$H" DAEMON_NAME=exrpd DAEMON_ALLOW_DOWNLOAD_BINARIES=false DAEMON_RESTART_AFTER_UPGRADE=true
	UNSAFE_SKIP_BACKUP=true DAEMON_GRPC_ADDRESS=127.0.0.1:$((P + 90)))
# OLD_PLAN is a dummy older plan, far below every real plan of the chain.
OLD_PLAN=v0.1.0
printf '#!/bin/sh\necho old\n' >"$E2E_DIR/old-exrpd" && chmod 0755 "$E2E_DIR/old-exrpd"
# The operator's own build of the upgrade: runnable, other bytes than the bundle's.
{ cat "$UPGRADE_BIN" && echo operator-build; } >"$E2E_DIR/operator-exrpd" && chmod 0755 "$E2E_DIR/operator-exrpd"
env "${stock_env[@]}" "$SCV" init "$GENESIS_BIN" >>"$LOG" 2>&1 || die "cosmovisor init"
env "${stock_env[@]}" "$SCV" add-upgrade "$GENESIS_PLAN" "$GENESIS_BIN" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $GENESIS_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$OLD_PLAN" "$E2E_DIR/old-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $OLD_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$UPGRADE_PLAN" "$E2E_DIR/operator-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $UPGRADE_PLAN"
config_sum=$(sha256sum "$cv/config.toml" | cut -d' ' -f1) || die "cosmovisor init wrote no config.toml"
env --default-signal=INT,QUIT "${stock_env[@]}" "$SCV" run start --home "$H" --chain-id "$CHAIN_ID" >>"$LOG" 2>&1 &
sup=$!
log "stock Cosmovisor on the stock home from main at height $PRE_H, pid $sup"
for ((i = 0; i < 120; i++)); do
	current_is "upgrades/$GENESIS_PLAN" && break
	sleep 0.5
done
wait_height $((PRE_H + 3)) 90 || die "stock Cosmovisor did not run the home"
check "stock Cosmovisor switched to upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "stock Cosmovisor recorded $GENESIS_PLAN" marker_names "$cv/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
stock_h=$(height)
# Unpatched Cosmovisor can mishandle a stop after its in-process switch, so the whole
# group is stopped, as systemd does.
kill -TERM "$sup" 2>/dev/null
for p in $(children); do kill -TERM "$p" 2>/dev/null; done
for ((i = 0; i < 30; i++)); do
	[[ -z $(children) ]] && ! kill -0 "$sup" 2>/dev/null && break
	sleep 0.5
done
kill -KILL "$sup" 2>/dev/null
for p in $(children); do kill -KILL "$p" 2>/dev/null; done
wait "$sup" 2>/dev/null
sup=
[[ -z $(children) ]] || die "exrpd of the stock home still running"
((stock_h < UH)) || die "stock Cosmovisor reached the upgrade height. Raise the proposal margin"
# The bundle never installed this home: commands refuse and init does not take it over.
check "client command on the stock home refuses: bundle not installed" refuses missing "$UB" --home "$H" keys list --keyring-backend test
check "init on the stock home refuses" refuses init "$UB" init localnet --chain-id "$CHAIN_ID" --home "$H"
check "version on the stock home is not-installed" version_is not-installed
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "version --bundle accepts the stock home (refused null, installed_bundle none)" bash -c '"$1" version --bundle --home "$2" | jq -e ".refused == null and .installed_bundle == \"none\"" >/dev/null' _ "$UB" "$H"

start_node
log "bundle on the stock home, supervisor pid $sup"
wait_height $((stock_h + 1)) 60 || die "chain did not start on the stock home"
(($(height) < UH)) || die "the stock home passed the upgrade height before the takeover checks"
check "takeover keeps current -> upgrades/$GENESIS_PLAN" current_is "upgrades/$GENESIS_PLAN"
check "takeover does not move current" not grep -q "exrpd: moved" "$LOG"
check "takeover replaces the operator's $UPGRADE_PLAN binary and logs it" grep -q "exrpd: replaced $cv/$UPGRADE_DIR/bin/exrpd" "$LOG"
check "takeover keeps identical binaries" not grep -q "exrpd: replaced $cv/upgrades/$GENESIS_PLAN/bin/exrpd" "$LOG"
check "takeover leaves other plan directories and config.toml" test "$("$cv/upgrades/$OLD_PLAN/bin/exrpd")" = old -a "$(sha256sum "$cv/config.toml" | cut -d' ' -f1)" = "$config_sum"
check "takeover runs the $GENESIS_PLAN exrpd" sup_running "upgrades/$GENESIS_PLAN"
check "takeover writes the marker" bundle_installed
check "after the takeover the chain passes H=$UH" wait_height $((UH + 5)) 180
check "current -> $UPGRADE_DIR after the takeover" current_is "$UPGRADE_DIR"
check "the bundle's upgrade exrpd runs, not the operator's" cmp -s "$UPGRADE_BIN" "$(child_exe)"
check "upgrade applied at H after the takeover" applied_at "$UH"
stop_check "SIGTERM to supervisor after the takeover leaves no exrpd" TERM switched

# ---------------------------------------------------------------- single bundle over the two-binary bundle
# The operator of an upgraded node replaces the two-binary bundle with the single bundle
# of the release while the node runs. Until the restart, the files of the home belong to
# the two-binary bundle.
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

check "standalone upgrade exrpd version is $UPGRADE_VERSION" test "$("$UPGRADE_BIN" version --home "$E2E_DIR/genesis-release")" = "$UPGRADE_VERSION"

echo
if ((fails > 0)); then
	echo "E2E ($mode): $fails check(s) FAILED"
	exit 1
fi
echo "E2E ($mode): all checks PASSED"

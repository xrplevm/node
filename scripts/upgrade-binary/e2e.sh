#!/usr/bin/env bash
# Localnet E2E for the exrpd bundle (linux/amd64).
#
# Usage: e2e.sh [--unpatched]
#   --unpatched  embed the unpatched upstream Cosmovisor release binary instead of the
#                patched build. The run passes only when the SIGTERM after the in-process
#                upgrade shows the defect (see stop_check).
#
# Env: CACHE (inputs, see below), GENESIS_VERSION (the genesis release, for example 11.3.0),
# E2E_DIR (scratch, deleted at start: it must be missing, empty or an earlier run's
# scratch), PORT_BASE (100 ports from here are used), UPGRADE_VERSION (the upgrade exrpd
# version, default the plan without "v", or a patch release of the plan for a patch
# bundle), EXRPD_UPGRADE (a bundle exrpd built from the
# same inputs to test instead of building one, for example the release asset. It cannot
# be used with --unpatched).
#
# CACHE must hold:
#   node_<GENESIS_VERSION>_Linux_amd64.tar.gz, checksums.txt   genesis GitHub release assets
#   exrpd-<UPGRADE_VERSION>-linux-amd64                         standalone upgrade exrpd (executable)
# and for --unpatched also:
#   cosmovisor-<COSMOVISOR_VERSION>-linux-amd64.tar.gz, SHA256SUMS-cosmovisor-<COSMOVISOR_VERSION>.txt
#
# Scenarios on one single-validator chain (homes run one at a time on the same ports):
#   main      genesis home with a data/upgrade-info.json from the genesis plan, as on the
#             validators that ran it. Commands before the first start run the genesis
#             exrpd. A gov upgrade to the plan (plan.sh) switches in process, then
#             restart, reinstall, SIGTERM.
#   resync    post-upgrade copy without data/upgrade-info.json, as after a resync: start
#             keeps the upgrade. A corrupted data file is refused.
#   halt-kill pre-upgrade copy whose supervisor and app are SIGKILLed at the upgrade halt,
#             before Cosmovisor switched: commands and the next start use the upgrade.
#   stock     pre-upgrade copy rebuilt as a stock Cosmovisor home: the bundle takes it over,
#             replaces the operator's binaries and switches at the upgrade height.
set -uo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/cosmovisor/cosmovisor.env
source "$here/cosmovisor/cosmovisor.env"
GENESIS_VERSION=${GENESIS_VERSION:?set GENESIS_VERSION to the genesis release version, for example 11.3.0}
UPGRADE_PLAN=$("$here/plan.sh") || exit 1
CACHE=${CACHE:?set CACHE to the directory with the inputs listed at the top of $0}
# The bundle passes the home to Cosmovisor and exrpd as a cleaned absolute path
# (filepath.Abs). The pgrep patterns below must see the same string.
E2E_DIR=$(realpath -ms "${E2E_DIR:-$here/../../.e2e}")
P=${PORT_BASE:-29000}
mode=patched
[[ ${1:-} == --unpatched ]] && mode=unpatched

CHAIN_ID=xrplevm_1449999-1
MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"
FEES=1000000000000000axrp
NODE=tcp://127.0.0.1:$((P + 57))
UB=$E2E_DIR/bin/exrpd
UPGRADE_VERSION=${UPGRADE_VERSION:-${UPGRADE_PLAN#v}}
V12=$CACHE/exrpd-$UPGRADE_VERSION-linux-amd64
EXRPD_UPGRADE=${EXRPD_UPGRADE:-}
[[ -n $EXRPD_UPGRADE && $mode == unpatched ]] && { echo "EXRPD_UPGRADE cannot be used with --unpatched" >&2; exit 2; }
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

# use_home NAME selects the node home and log every helper below works on.
use_home() { H=$E2E_DIR/$1 LOG=$E2E_DIR/$1.log; }
use_home main

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
# bundle_info prints the version --bundle JSON for a home that does not exist, so it
# reads nothing of a real home and writes nothing.
bundle_info() { "$UB" version --bundle --home "$E2E_DIR/no-home"; }
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

init_chain() {
	local v11=$1 g=$H/config/genesis.json addr
	"$v11" --home "$H" config set client chain-id "$CHAIN_ID" >/dev/null || return
	"$v11" --home "$H" config set client keyring-backend test >/dev/null || return
	"$v11" --home "$H" config set client node "$NODE" >/dev/null || return
	echo "$MNEMONIC" | "$v11" --home "$H" keys add alice --recover --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return
	"$v11" --home "$H" keys add bob --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return
	"$v11" --home "$H" init localnet -o --chain-id "$CHAIN_ID" >/dev/null 2>&1 || return
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
	addr=$("$v11" --home "$H" keys show alice -a --keyring-backend test) || return
	"$v11" --home "$H" genesis add-genesis-account "$addr" 1000000apoa,1000000000000000000000000000axrp --keyring-backend test || return
	"$v11" --home "$H" genesis gentx alice 1000000apoa --gas-prices 0axrp --keyring-backend test --chain-id "$CHAIN_ID" >/dev/null 2>&1 || return
	"$v11" --home "$H" genesis collect-gentxs >/dev/null 2>&1 || return
	"$v11" --home "$H" genesis validate >/dev/null || return
	set_ports
}


# ---------------------------------------------------------------- build
log "mode: $mode, scratch: $E2E_DIR, ports: $P-$((P + 99))"
inputs=("$CACHE/node_${GENESIS_VERSION}_Linux_amd64.tar.gz" "$CACHE/checksums.txt" "$V12" ${EXRPD_UPGRADE:+"$EXRPD_UPGRADE"})
[[ $mode == unpatched ]] && inputs+=("$CACHE/cosmovisor-$COSMOVISOR_VERSION-linux-amd64.tar.gz" "$CACHE/SHA256SUMS-cosmovisor-$COSMOVISOR_VERSION.txt")
for f in "${inputs[@]}"; do
	[[ -f $f ]] || die "missing input $f (CACHE=$CACHE, see the list at the top of $0)"
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
rm -rf "$E2E_DIR" && mkdir -p "$E2E_DIR/bin" "$E2E_DIR/v11" && touch "$scratch_marker" || die "create $E2E_DIR"
build_args=(--genesis-version "$GENESIS_VERSION" --genesis-archive "$CACHE/node_${GENESIS_VERSION}_Linux_amd64.tar.gz" --genesis-checksums "$CACHE/checksums.txt"
	--upgrade-exrpd "$V12" --upgrade-version "$UPGRADE_VERSION" --goos linux --goarch amd64 --out "$UB")
if [[ $mode == unpatched ]]; then
	cv=cosmovisor-$COSMOVISOR_VERSION-linux-amd64.tar.gz
	(cd "$CACHE" && grep " $cv\$" "SHA256SUMS-cosmovisor-$COSMOVISOR_VERSION.txt" | sha256sum -c -) || die "cosmovisor release checksum"
	mkdir -p "$E2E_DIR/cv" && tar -xzf "$CACHE/$cv" -C "$E2E_DIR/cv"
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
if [[ -n $EXRPD_UPGRADE ]]; then
	cp "$EXRPD_UPGRADE" "$UB" || die "copy $EXRPD_UPGRADE"
else
	"$here/build.sh" "${build_args[@]}" || die "build"
fi
check "manifest cosmovisor source ($mode)" cosmovisor_source "$([[ $mode == patched ]] && echo "$COSMOVISOR_PATCH" || echo unpatched)"

tar -xzf "$CACHE/node_${GENESIS_VERSION}_Linux_amd64.tar.gz" -C "$E2E_DIR/v11" bin/exrpd
V11=$E2E_DIR/v11/bin/exrpd
check "genesis tool is release v$GENESIS_VERSION" test "$("$V11" version --home "$E2E_DIR/v11")" = "$GENESIS_VERSION"
manifest() { bundle_info | jq -r ".manifest.$1"; }
file_sha() { sha256sum "$1" | cut -d' ' -f1; }
GENESIS_PLAN=$(manifest genesis.plan)
check "manifest genesis plan $GENESIS_PLAN is the newest plan of v$GENESIS_VERSION" test "$GENESIS_PLAN" = "$("$here/genesis-plan.sh" "$GENESIS_VERSION")"
check "manifest upgrade plan is $UPGRADE_PLAN" test "$(manifest 'upgrades[0].plan')" = "$UPGRADE_PLAN"
check "manifest genesis is the release v$GENESIS_VERSION exrpd" test "$(manifest genesis.sha256)" = "$(file_sha "$V11")"
check "manifest upgrade is the standalone $UPGRADE_VERSION exrpd" test "$(manifest 'upgrades[0].sha256')" = "$(file_sha "$V12")"

# ---------------------------------------------------------------- chain
init_chain "$V11" || die "init chain"
# Help never installs: before anything is installed it only tells how to find the home.
help_writes_nothing() {
	"$UB" --home "$H" --help | grep -F "no exrpd is installed in $H yet" >/dev/null && test ! -e "$H/cosmovisor"
}
check "help before the first install writes nothing" help_writes_nothing
# Before the first start, every command runs the genesis exrpd, and installs the binaries
# without creating `current`.
check "command before the first start runs genesis" ran_version_is "$GENESIS_VERSION"
# exrpd's completion also reads $HOME/config.toml, so the check uses a scratch HOME: a
# config.toml of another tool in the developer's home must not fail it.
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "completion passes the typed words to exrpd" env HOME="$E2E_DIR/v11" bash -c '"$1" __completeNoDesc --home "$2" ke | grep -qx keys' _ "$UB" "$H"
check "no current before the first start" test ! -e "$H/cosmovisor/current" -a ! -L "$H/cosmovisor/current"
check "installed genesis exrpd = release bytes" cmp -s "$V11" "$H/cosmovisor/genesis/bin/exrpd"
check "installed $GENESIS_PLAN exrpd = release bytes" cmp -s "$V11" "$H/cosmovisor/upgrades/$GENESIS_PLAN/bin/exrpd"
check "installed upgrade exrpd = standalone v12 bytes" cmp -s "$V12" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"
check "version before the first start is $GENESIS_VERSION" version_is "$GENESIS_VERSION"
# shellcheck disable=SC2016 # $1 and $2 expand in the inner shell
check "help runs the installed exrpd" bash -c '"$1" --home "$2" --help | grep "Available Commands" >/dev/null' _ "$UB" "$H"

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
check "$GENESIS_PLAN recorded in its directory" marker_names "$H/cosmovisor/upgrades/$GENESIS_PLAN/upgrade-info.json" "$GENESIS_PLAN"
check "no record on genesis" test ! -e "$H/cosmovisor/genesis/upgrade-info.json"
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
cp -a "$H" "$E2E_DIR/halt-kill"
cp -a "$H" "$E2E_DIR/stock"
start_node
log "restarted before the upgrade, supervisor pid $sup"
wait_height $((last + 1)) 60 || die "chain did not restart"
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

rm -rf "$H/cosmovisor"
start_node
log "reinstalled (deleted $H/cosmovisor), supervisor pid $sup"
check "reinstall after H selects $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "current -> $UPGRADE_DIR after reinstall" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "running upgrade exrpd after reinstall" running "$UPGRADE_DIR"
stop_check "SIGTERM to supervisor after reinstall leaves no exrpd" TERM

# ---------------------------------------------------------------- resync past H
# A resync (unsafe-reset-all, then state sync) deletes data/upgrade-info.json. `current`
# is then the only evidence of the plan, and start must keep it.
cp -a "$H" "$E2E_DIR/resync"
use_home resync
rm "$H/data/upgrade-info.json"
start_node
log "resync home, supervisor pid $sup"
check "without a data file start keeps $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "current -> $UPGRADE_DIR without a data file" current_is "$UPGRADE_DIR"
check "running upgrade exrpd without a data file" running "$UPGRADE_DIR"
stop_check "SIGTERM to supervisor on the resync home leaves no exrpd" TERM
echo '{"name":"v12.0' >"$H/data/upgrade-info.json"
bins_before=$(sha256sum "$H"/cosmovisor/*/bin/exrpd "$H"/cosmovisor/upgrades/*/bin/exrpd)
check "corrupted data file is refused" fails_quietly "$UB" version --home "$H"
check "start refuses a corrupted data file" start_refused
check "current and binaries unchanged after the refusal" \
	test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR" -a "$(sha256sum "$H"/cosmovisor/*/bin/exrpd "$H"/cosmovisor/upgrades/*/bin/exrpd)" = "$bins_before"

# ---------------------------------------------------------------- killed at the halt
# A long DAEMON_POLL_INTERVAL keeps Cosmovisor from switching before the kill. The halted
# app exits on its own, so the checks rely on the state on disk: current still on the
# genesis plan and the data file naming the upgrade.
use_home halt-kill
start_node DAEMON_POLL_INTERVAL=1h
log "halt-kill home, supervisor pid $sup"
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

# ---------------------------------------------------------------- takeover of a stock Cosmovisor home
# A validator runs stock Cosmovisor v1.7.3, stops it before the upgrade height and starts
# the bundle on the same home. Stock Cosmovisor sets up the home itself: `init` writes
# genesis/ and config.toml, `add-upgrade` stages the genesis plan, an older plan and the
# operator's own build of the upgrade, and the run switches to upgrades/<genesis plan>
# because data/upgrade-info.json names it, writing its record there.
use_home stock
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
{ cat "$V12" && echo operator-build; } >"$E2E_DIR/operator-exrpd" && chmod 0755 "$E2E_DIR/operator-exrpd"
env "${stock_env[@]}" "$SCV" init "$V11" >>"$LOG" 2>&1 || die "cosmovisor init"
env "${stock_env[@]}" "$SCV" add-upgrade "$GENESIS_PLAN" "$V11" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $GENESIS_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$OLD_PLAN" "$E2E_DIR/old-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $OLD_PLAN"
env "${stock_env[@]}" "$SCV" add-upgrade "$UPGRADE_PLAN" "$E2E_DIR/operator-exrpd" >>"$LOG" 2>&1 || die "cosmovisor add-upgrade $UPGRADE_PLAN"
config_sum=$(sha256sum "$cv/config.toml" | cut -d' ' -f1) || die "cosmovisor init wrote no config.toml"
env --default-signal=INT,QUIT "${stock_env[@]}" "$SCV" run start --home "$H" --chain-id "$CHAIN_ID" >>"$LOG" 2>&1 &
sup=$!
log "stock Cosmovisor on the stock home, pid $sup"
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
check "after the takeover the chain passes H=$UH" wait_height $((UH + 5)) 180
check "current -> $UPGRADE_DIR after the takeover" current_is "$UPGRADE_DIR"
check "the bundle's upgrade exrpd runs, not the operator's" cmp -s "$V12" "$(child_exe)"
check "upgrade applied at H after the takeover" applied_at "$UH"
stop_check "SIGTERM to supervisor after the takeover leaves no exrpd" TERM switched

check "standalone upgrade exrpd version is $UPGRADE_VERSION" test "$("$V12" version --home "$E2E_DIR/v11")" = "$UPGRADE_VERSION"

echo
if ((fails > 0)); then
	echo "E2E ($mode): $fails check(s) FAILED"
	exit 1
fi
echo "E2E ($mode): all checks PASSED"

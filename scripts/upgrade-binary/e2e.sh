#!/usr/bin/env bash
# Localnet E2E for exrpd-upgrade (linux/amd64).
#
# Usage: e2e.sh [--unpatched]
#   --unpatched  embed the unpatched upstream Cosmovisor release binary instead of the
#                patched build. The run passes only when the SIGTERM after the in-process
#                upgrade shows the defect (see stop_check).
#
# Env: CACHE (inputs, see below), E2E_DIR (scratch, deleted at start: it must be missing,
# empty or an earlier run's scratch), PORT_BASE (100 ports from here are used),
# UPGRADE_VERSION (the upgrade exrpd version, default UPGRADE_PLAN without "v", or a patch
# release of the plan for a patch bundle), EXRPD_UPGRADE (an exrpd-upgrade built from the
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
#   main      genesis home with a data/upgrade-info.json from the plan that started the
#             genesis release line. A gov upgrade to UPGRADE_PLAN switches in process,
#             then restart, reinstall, SIGTERM.
#   restore   post-upgrade copy without cosmovisor/ and with an earlier plan's marker:
#             stage --select.
#   halt-kill pre-upgrade copy whose supervisor and app are SIGKILLed at the upgrade halt,
#             before Cosmovisor switched: the next start switches.
set -uo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/bundle.env
source "$here/bundle.env"
CACHE=${CACHE:?set CACHE to the directory with the inputs listed at the top of $0}
# exrpd-upgrade passes the home to Cosmovisor and exrpd as a cleaned absolute path
# (filepath.Abs). The pgrep patterns below must see the same string.
E2E_DIR=$(realpath -ms "${E2E_DIR:-$here/../../.e2e}")
P=${PORT_BASE:-29000}
mode=patched
[[ ${1:-} == --unpatched ]] && mode=unpatched

CHAIN_ID=xrplevm_1449999-1
MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"
FEES=1000000000000000axrp
NODE=tcp://127.0.0.1:$((P + 57))
UB=$E2E_DIR/bin/exrpd-upgrade
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
height() { timeout 5 "$(cur)" status --node "$NODE" 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height // 0'; }
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

# stop_check NAME SIGNAL [switched|group] sends SIGNAL to the supervisor PID only (not
# the group), or with "group" to this home's exrpd as well, the way systemd's default
# KillMode=control-group does, and waits up to 15s for the supervisor and every exrpd of
# this home to exit. It records the last height in $last. The signal-handler defect shows
# as an exrpd left running, a "process already finished" log line from the handler of an
# earlier run, or the supervisor exit status 1 that handler causes. The patched Cosmovisor
# must show none of them, also when the app exits before the supervisor forwards the
# signal. With --unpatched, a stop after an in-process switch (third argument
# "switched") must show the defect.
stop_check() {
	local name=$1 sig=$2 how=${3:-supervisor} expect=clean left status=running stale mark deadline defect=
	[[ $how == switched && $mode == unpatched ]] && expect=defect
	last=$(height || echo 0)
	mark=$(wc -l <"$LOG")
	# The app is signalled first, so it can exit before Cosmovisor's handler forwards the
	# signal. The handler then finds a finished process and must not exit 1.
	if [[ $how == group ]]; then
		local p
		for p in $(children); do kill "-$sig" "$p"; done
	fi
	kill "-$sig" "$sup"
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
		local h1 h2
		h1=$(height)
		sleep 3
		h2=$(height)
		echo "    orphan exrpd pid(s) $left, ppid $(ps -o ppid= -p "${left%%$'\n'*}" | tr -d ' '), height $h1 -> $h2"
		for p in $left; do kill -TERM "$p"; done
		sleep 5
	fi
	[[ -n $left || -n $stale || ($status != running && $status != 0) ]] && defect=1
	if [[ $status == running ]]; then
		fail "$name (supervisor still running)"
		kill -KILL "$sup"
	elif [[ $expect == defect ]]; then
		if [[ -n $defect ]]; then
			pass "SIG$sig after the in-process switch: unpatched Cosmovisor shows the defect"
		else
			fail "SIG$sig after the in-process switch: the unpatched defect did not show in this run (it depends on which signal handler runs first)"
		fi
	elif [[ -n $defect ]]; then
		fail "$name"
	else
		pass "$name"
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
		-e 's|^timeout_commit = .*|timeout_commit = "1s"|' "$c"
	sed -i \
		-e "/^\[api\]/,/^\[/ s|^address = .*|address = \"tcp://127.0.0.1:$((P + 17))\"|" \
		-e "/^\[grpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((P + 90))\"|" \
		-e "s|^geth-metrics-address = .*|geth-metrics-address = \"127.0.0.1:$((P + 81))\"|" \
		-e "/^\[json-rpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((P + 45))\"|" \
		-e "s|^ws-address = .*|ws-address = \"127.0.0.1:$((P + 46))\"|" \
		-e "s|^metrics-address = .*|metrics-address = \"127.0.0.1:$((P + 65))\"|" "$a"
}

init_chain() {
	local v11=$1 g=$H/config/genesis.json
	"$v11" --home "$H" config set client chain-id "$CHAIN_ID" >/dev/null
	"$v11" --home "$H" config set client keyring-backend test >/dev/null
	"$v11" --home "$H" config set client node "$NODE" >/dev/null
	echo "$MNEMONIC" | "$v11" --home "$H" keys add alice --recover --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1
	"$v11" --home "$H" keys add bob --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1
	"$v11" --home "$H" init localnet -o --chain-id "$CHAIN_ID" >/dev/null 2>&1
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
		"$g" >"$g.tmp" && mv "$g.tmp" "$g"
	"$v11" --home "$H" genesis add-genesis-account "$("$v11" --home "$H" keys show alice -a --keyring-backend test)" 1000000apoa,1000000000000000000000000000axrp --keyring-backend test
	"$v11" --home "$H" genesis gentx alice 1000000apoa --gas-prices 0axrp --keyring-backend test --chain-id "$CHAIN_ID" >/dev/null 2>&1
	"$v11" --home "$H" genesis collect-gentxs >/dev/null 2>&1
	"$v11" --home "$H" genesis validate >/dev/null
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
build_args=(--genesis-archive "$CACHE/node_${GENESIS_VERSION}_Linux_amd64.tar.gz" --genesis-checksums "$CACHE/checksums.txt"
	--upgrade-exrpd "$V12" --upgrade-version "$UPGRADE_VERSION" --goos linux --goarch amd64 --out "$UB")
if [[ $mode == unpatched ]]; then
	cv=cosmovisor-$COSMOVISOR_VERSION-linux-amd64.tar.gz
	(cd "$CACHE" && grep " $cv\$" "SHA256SUMS-cosmovisor-$COSMOVISOR_VERSION.txt" | sha256sum -c -) || die "cosmovisor release checksum"
	mkdir -p "$E2E_DIR/cv" && tar -xzf "$CACHE/$cv" -C "$E2E_DIR/cv"
	build_args+=(--cosmovisor "$E2E_DIR/cv/cosmovisor" --cosmovisor-source "unpatched upstream release $cv")
fi
IFS=. read -r plan_major plan_minor _ <<<"${UPGRADE_PLAN#v}"
next_minor=$plan_major.$((plan_minor + 1)).0
build_refuses_next_minor() { ! "$here/build.sh" "${build_args[@]}" --upgrade-version "$next_minor" 2>/dev/null; }
check "build refuses $next_minor, a minor release that needs its own plan" build_refuses_next_minor
if [[ -n $EXRPD_UPGRADE ]]; then
	cp "$EXRPD_UPGRADE" "$UB" || die "copy $EXRPD_UPGRADE"
else
	"$here/build.sh" "${build_args[@]}" || die "build"
fi
check "manifest cosmovisor source ($mode)" bash -c "'$UB' info | jq -e '.cosmovisor.source' | grep -q '$([[ $mode == patched ]] && echo "$COSMOVISOR_PATCH" || echo unpatched)'"

tar -xzf "$CACHE/node_${GENESIS_VERSION}_Linux_amd64.tar.gz" -C "$E2E_DIR/v11" bin/exrpd
V11=$E2E_DIR/v11/bin/exrpd
check "genesis tool is release v$GENESIS_VERSION" test "$("$V11" version)" = "$GENESIS_VERSION"
manifest_sha() { "$UB" info | jq -r ".$1.sha256"; }
file_sha() { sha256sum "$1" | cut -d' ' -f1; }
check "manifest plan is $UPGRADE_PLAN" test "$("$UB" info | jq -r .plan)" = "$UPGRADE_PLAN"
check "manifest genesis is the release v$GENESIS_VERSION exrpd" test "$(manifest_sha genesis)" = "$(file_sha "$V11")"
check "manifest upgrade is the standalone $UPGRADE_VERSION exrpd" test "$(manifest_sha upgrade)" = "$(file_sha "$V12")"

# ---------------------------------------------------------------- chain
init_chain "$V11" || die "init chain"
"$UB" stage --home "$H" || die "stage"
check "staged genesis exrpd = release bytes" cmp -s "$V11" "$H/cosmovisor/genesis/bin/exrpd"
check "staged upgrade exrpd = standalone v12 bytes" cmp -s "$V12" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"

# ---------------------------------------------------------------- previous plan marker
# Existing validator homes carry data/upgrade-info.json from the plan that started the
# genesis release line (v11.3.0 for genesis 11.3.0).
IFS=. read -r genesis_major genesis_minor _ <<<"$GENESIS_VERSION"
PREV_PLAN=v$genesis_major.$genesis_minor.0
echo "{\"name\":\"$PREV_PLAN\",\"time\":\"0001-01-01T00:00:00Z\",\"height\":1}" >"$H/data/upgrade-info.json"
start_node
log "supervisor pid $sup"
wait_height 3 90 || die "chain did not start"
check "supervisor execs cosmovisor in place (pid $sup)" bash -c "readlink /proc/$sup/exe | grep -q '/cosmovisor/supervisor/'"
check "downloads disabled in supervisor env" bash -c "tr '\0' '\n' </proc/$sup/environ | grep -qx DAEMON_ALLOW_DOWNLOAD_BINARIES=false"
check "backups off by default in supervisor env" bash -c "tr '\0' '\n' </proc/$sup/environ | grep -qx UNSAFE_SKIP_BACKUP=true"
check "previous plan $PREV_PLAN recorded on genesis" bash -c "jq -e '.name == \"$PREV_PLAN\"' '$H/cosmovisor/genesis/upgrade-info.json' >/dev/null"
h0=$(height)
check "v$GENESIS_VERSION keeps producing blocks with a $PREV_PLAN marker" wait_height $((h0 + 15)) 60
check "same supervisor still running v$GENESIS_VERSION" bash -c "kill -0 $sup && test '$(child_exe)' = '$(readlink -f "$H/cosmovisor/genesis/bin/exrpd")'"
check "no false switch to $PREV_PLAN" bash -c "! grep -q 'upgrades/$PREV_PLAN' '$LOG'"
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

# Stop before the upgrade height and keep a copy of the home for the halt-kill scenario.
stop_check "SIGTERM to supervisor before upgrade leaves no exrpd" TERM
((last < UH)) || die "stopped at or past the upgrade height. Raise the proposal margin"
cp -a "$H" "$E2E_DIR/halt-kill"
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
check "no download attempted" bash -c "! grep -Eiq 'beginning to download|downloading binary' '$LOG'"
check "upgrade applied at H" applied_at "$UH"
check "signed bank tx after upgrade" send_tx

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

# ---------------------------------------------------------------- restore past H with --select
# A snapshot or state-sync restore past the upgrade height has no cosmovisor/. The data
# directory keeps the marker of an earlier plan, which --select deletes.
cp -a "$H" "$E2E_DIR/restore"
use_home restore
rm -rf "$H/cosmovisor"
echo "{\"name\":\"$PREV_PLAN\",\"time\":\"0001-01-01T00:00:00Z\",\"height\":1}" >"$H/data/upgrade-info.json"
check "stage --select refuses an unknown name" bash -c "! '$UB' stage --home '$H' --select 'v$next_minor' 2>/dev/null"
check "stage --select $UPGRADE_PLAN" "$UB" stage --home "$H" --select "$UPGRADE_PLAN"
check "current -> $UPGRADE_DIR after --select" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "--select deletes the $PREV_PLAN marker" test ! -e "$H/data/upgrade-info.json"
start_node
log "restore home, supervisor pid $sup"
check "restored home past H starts $UPGRADE_PLAN and chain advances" wait_height $((last + 5)) 90
check "running upgrade exrpd on restored home" running "$UPGRADE_DIR"
stop_check "SIGTERM to supervisor on restored home leaves no exrpd" TERM

# ---------------------------------------------------------------- killed at the halt
# Cosmovisor polls upgrade-info.json every DAEMON_POLL_INTERVAL. A long interval holds the
# window between the v11 halt and the switch open so the SIGKILL lands inside it.
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
check "killed between halt and switch (current -> genesis, marker names $UPGRADE_PLAN)" \
	bash -c "test \"\$(readlink '$H/cosmovisor/current')\" = genesis && jq -e '.name == \"$UPGRADE_PLAN\"' '$H/data/upgrade-info.json' >/dev/null"
start_node
log "restarted after the kill, supervisor pid $sup"
check "restart after the kill switches and passes H=$UH" wait_height $((UH + 5)) 120
check "current -> $UPGRADE_DIR after the kill" test "$(readlink "$H/cosmovisor/current")" = "$UPGRADE_DIR"
check "running upgrade exrpd after the kill" running "$UPGRADE_DIR"
check "upgrade applied at H after the kill" applied_at "$UH"
stop_check "SIGTERM to supervisor after the kill leaves no exrpd" TERM

check "standalone upgrade exrpd version is $UPGRADE_VERSION" test "$("$V12" version)" = "$UPGRADE_VERSION"

echo
if ((fails > 0)); then
	echo "E2E ($mode): $fails check(s) FAILED"
	exit 1
fi
echo "E2E ($mode): all checks PASSED"

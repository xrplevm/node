# shellcheck shell=bash disable=SC2001,SC2015 # sed prefixes each line of multi-line text. A && B || die: die exits.
# Helpers of e2e.sh. They work on the home, log and bundle that use_home selects.

log() { echo "--- $*"; }
pass() { echo "PASS: $*"; }
fail() { echo "FAIL: $*"; fails=$((fails + 1)); }
check() { # name, command...
	local name=$1
	shift
	if "$@"; then pass "$name"; else fail "$name"; fi
}
die() { echo "FATAL: $*"; tail -n 40 "$LOG" 2>/dev/null; exit 1; }

# use_home NAME BUNDLE selects the home, its log and the launcher.
use_home() { H=$E2E_DIR/$1 LOG=$E2E_DIR/$1.log UB=$2; }

# re_escape prints $1 with ERE metacharacters escaped, so pgrep -f matches it literally.
re_escape() { sed 's/[][\.*^$+?(){}|]/\\&/g' <<<"$1"; }
children() { pgrep -f "bin/exrpd start --home $(re_escape "$H")( |\$)" || true; }
# cleanup kills every exrpd under $E2E_DIR. It is armed only after this run holds the lock,
# so it cannot kill the nodes of another run.
cleanup() {
	[[ -n $sup ]] && kill -KILL "$sup" 2>/dev/null
	local p
	for p in $(pgrep -f "bin/exrpd start --home $(re_escape "$E2E_DIR/")"); do kill -KILL "$p" 2>/dev/null; done
}

cur() { echo "$H/cosmovisor/current/bin/exrpd"; }
# RPC calls are bounded, so a node that holds its socket open cannot hang a wait.
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

# start_node runs the bundle in the background like a service manager: env restores the
# SIGINT and SIGQUIT that bash ignores in background jobs. Extra args go to env.
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
# bundle_info [EXRPD]: the version --bundle JSON of EXRPD (default $UB), for a home that does not exist.
bundle_info() { "${1:-$UB}" version --bundle --home "$E2E_DIR/no-home"; }
cosmovisor_source() { bundle_info | jq -e '.manifest.cosmovisor.source' | grep "$1" >/dev/null; }
# version_is VERSION: exrpd version on this home names VERSION first, as the standalone does.
version_is() { test "$("$UB" version --home "$H" | cut -d' ' -f1)" = "$1"; }
# ran_version_is VERSION: a command passed through to exrpd ran the binary of VERSION.
ran_version_is() { "$UB" --home "$H" version --long | grep -qx "version: $1"; }
# start_refused: start fails with the launcher's own "is corrupted" message. Status 124 is a timeout, not a refusal.
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

# home_files lists $H/cosmovisor with type, size, mtime and link target, or "absent".
home_files() {
	[[ -e $H ]] || { echo absent; return; }
	find "$H/cosmovisor" -printf '%P %y %s %T@ %l\n' 2>/dev/null | sort
}
# says STATE COMMAND...: COMMAND prints the launcher message for STATE (none, missing,
# differs, init or single) and writes nothing in $H/cosmovisor. Sets $said to its status.
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
# refuses STATE COMMAND...: says, and COMMAND fails. Status 124 is a timeout, not a refusal.
refuses() { says "$@" && ((said != 0 && said != 124)); }
installed_is() { "$UB" version --bundle --home "$H" | jq -e --arg s "$1" '.installed_bundle == $s' >/dev/null; }
# bundle_installed: the marker holds a SHA-256 and matches the bundle at $UB.
bundle_installed() { grep -Eqx '[0-9a-f]{64}' "$H/cosmovisor/bundle.sha256" && installed_is this; }
# completes_nothing: TAB completion suggests nothing and exits 0.
completes_nothing() {
	local out
	out=$(env HOME="$E2E_DIR/genesis-release" "$UB" __completeNoDesc --home "$H" ke) && test "$out" = ":4"
}
# bundle_status: exrpd status through the bundle answers with the node's sync info.
bundle_status() { timeout 5 "$UB" --home "$H" --node "$NODE" status | jq -e '.sync_info // .SyncInfo' >/dev/null; }
# setup_installs: bundle setup installs and writes the marker, keeps current and starts no exrpd.
setup_installs() {
	local cur_before
	cur_before=$(readlink "$H/cosmovisor/current")
	"$UB" bundle setup --home "$H" >>"$LOG" 2>&1 || return
	bundle_installed && test "$(readlink "$H/cosmovisor/current")" = "$cur_before" -a -z "$(children)" &&
		cmp -s "$UPGRADE_BIN" "$H/cosmovisor/$UPGRADE_DIR/bin/exrpd"
}

# stop_check NAME SIGNAL [switched|group] sends SIGNAL to the supervisor (with "group" also
# to this home's exrpd, as systemd does), waits up to 15s and records the last height in
# $last. A clean stop leaves no exrpd, no "process already finished" log line and status 0.
# With --unpatched, a stop after the switch ("switched") must show that defect with status 1.
# Then it kills what is left, and aborts the run if something survives SIGKILL.
stop_check() {
	local name=$1 sig=$2 how=${3:-supervisor} expect=clean left status=running stale mark deadline defect='' signature='' h1='' h2='' p
	[[ $how == switched && $mode == unpatched ]] && expect=defect
	last=$(height || echo 0)
	mark=$(wc -l <"$LOG")
	# The supervisor gets the signal first. If the app got it first, it could exit before
	# the supervisor, and a late signal kills Cosmovisor with status 143.
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

# init_chain EXRPD creates the home with `EXRPD init` and sets up a one-validator chain.
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

# manifest PATH [EXRPD]: the manifest value at PATH of the bundle EXRPD, default $UB.
manifest() { bundle_info "${2:-$UB}" | jq -r ".manifest.$1"; }
# launcher_commit_is_set EXRPD: the manifest names the commit that built the launcher.
launcher_commit_is_set() { manifest launcher.commit "$1" | grep -Eqx '[0-9a-f]{40}(-dirty)?'; }
has_upgrade() { bundle_info "$1" | jq -e '.manifest | has("upgrades")' >/dev/null; }

# kill_home SIGNAL signals the supervisor and this home's exrpd, waits up to 15s, then SIGKILLs.
kill_home() {
	local p i
	kill "-$1" "$sup" 2>/dev/null
	for p in $(children); do kill "-$1" "$p" 2>/dev/null; done
	for ((i = 0; i < 30; i++)); do
		[[ -z $(children) ]] && ! kill -0 "$sup" 2>/dev/null && break
		sleep 0.5
	done
	kill -KILL "$sup" 2>/dev/null
	for p in $(children); do kill -KILL "$p" 2>/dev/null; done
	wait "$sup" 2>/dev/null
	sup=
}

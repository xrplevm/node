# shellcheck shell=bash disable=SC2015 # A && B || die: die exits.
# Helpers of e2e.sh. They work on the home, log and bundle that use_home selects.

log() { echo "--- $*"; }
pass() { echo "PASS: $*"; }
fail() { echo "FAIL: $*"; fails=$((fails + 1)); }
check() { # name, command...
	local name=$1
	shift
	if "$@"; then pass "$name"; else fail "$name"; fi
}
die() { echo "FATAL: $*"; tail -n 40 "$node_log" 2>/dev/null; exit 1; }

use_home() {
	node_home=$E2E_DIR/$1
	node_log=$E2E_DIR/$1.log
	active_bundle=$2
	standalone_exrpd=${3:-}
}

# re_escape prints $1 with ERE metacharacters escaped, so pgrep -f matches it literally.
re_escape() { sed 's/[][\.*^$+?(){}|]/\\&/g' <<<"$1"; }
children() { pgrep -f "bin/exrpd start --home $(re_escape "$node_home")( |\$)" || true; }

# cleanup kills every exrpd under $E2E_DIR. It is armed only after this run holds the lock,
# so it cannot kill the nodes of another run.
cleanup() {
	[[ -n $supervisor_pid ]] && kill -KILL "$supervisor_pid" 2>/dev/null
	local p
	for p in $(pgrep -f "bin/exrpd start --home $(re_escape "$E2E_DIR/")"); do kill -KILL "$p" 2>/dev/null; done
}

cur() { echo "${standalone_exrpd:-$node_home/cosmovisor/current/bin/exrpd}"; }

# RPC calls are bounded, so a node that holds its socket open cannot hang a wait.
height() { timeout 5 "$(cur)" status --home "$node_home" --node "$node_rpc" 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height // 0'; }
wait_height() { # target, timeout seconds
	local deadline=$((SECONDS + $2))
	while ((SECONDS < deadline)); do
		(($(height || echo 0) >= $1)) && return 0
		sleep 1
	done
	return 1
}

cli() { timeout 30 "$(cur)" --home "$node_home" --node "$node_rpc" --chain-id "$CHAIN_ID" --keyring-backend test "$@"; }
addr() { "$(cur)" --home "$node_home" keys show "$1" -a --keyring-backend test; }
balance() { cli query bank balance "$(addr bob)" axrp -o json | jq -r '.balance.amount'; }

# wait_tx HASH prints the transaction as JSON once a block includes it, within 30s.
wait_tx() {
	local deadline=$((SECONDS + 30)) out
	while ((SECONDS < deadline)); do
		out=$(timeout 5 "$(cur)" --home "$node_home" --node "$node_rpc" query tx "$1" -o json 2>/dev/null) && [[ -n $out ]] && echo "$out" && return 0
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
	env --default-signal=INT,QUIT DAEMON_GRPC_ADDRESS=127.0.0.1:$((port_base + 90)) "$@" \
		"${standalone_exrpd:-$active_bundle}" start --home "$node_home" --chain-id "$CHAIN_ID" >>"$node_log" 2>&1 &
	supervisor_pid=$!
}

child_exe() {
	local c
	c=$(pgrep -P "$supervisor_pid" | head -n1) && [[ -n $c ]] && readlink -f "/proc/$c/exe"
}
running() { test "$(child_exe)" = "$(readlink -f "$node_home/cosmovisor/$1/bin/exrpd")"; }
current_is() { test "$(readlink "$node_home/cosmovisor/current")" = "$1" -o "$(readlink "$node_home/cosmovisor/current")" = "$node_home/cosmovisor/$1"; }
applied_at() { test "$(cli query upgrade applied "$UPGRADE_PLAN" -o json | jq -r '.header.height // .height')" = "$1"; }
not() { ! "$@"; }
marker_names() { jq -e --arg n "$2" '.name == $n' "$1" >/dev/null; } # file, plan
sup_exe_has() { readlink "/proc/$supervisor_pid/exe" | grep "$1" >/dev/null; }

# stop_check NAME SIGNAL waits for a clean exit, then forcibly cleans up on failure.
stop_check() {
	local name=$1 sig=$2 left status=running mark deadline
	stopped_height=$(height || echo 0)
	mark=$(wc -l <"$node_log")
	kill "-$sig" "$supervisor_pid" 2>/dev/null
	deadline=$((SECONDS + 15))
	while ((SECONDS < deadline)); do
		if [[ $status == running ]] && ! kill -0 "$supervisor_pid" 2>/dev/null; then
			wait "$supervisor_pid"
			status=$?
		fi
		left=$(children)
		[[ $status != running && -z $left ]] && break
		sleep 1
	done

	left=$(children)
	echo "    SIG$sig -> supervisor $supervisor_pid status=$status; exrpd left: ${left:-none}"
	if [[ $status == 0 && -z $left ]] &&
		! tail -n "+$((mark + 1))" "$node_log" | grep 'process already finished' >/dev/null; then
		pass "$name"
	else
		fail "$name"
	fi

	if [[ $status == running || -n $left ]]; then
		kill_home KILL
	else
		supervisor_pid=
	fi
}

check_upgrade() {
	local name=$1
	wait_height $((upgrade_height + 5)) 180 || die "$name did not pass H=$upgrade_height"
	check "$name selects $UPGRADE_DIR" current_is "$UPGRADE_DIR"
	check "$name runs the upgrade exrpd" running "$UPGRADE_DIR"
	check "$name applies the upgrade at H=$upgrade_height" applied_at "$upgrade_height"
	check "$name submits a signed bank transaction" send_tx
}

restart_check() {
	local name=$1
	stop_check "$name stops without orphans" TERM
	start_node
	wait_height $((stopped_height + 5)) 90 || die "$name did not advance after restart"

	check "$name keeps $UPGRADE_DIR after restart" current_is "$UPGRADE_DIR"
	check "$name runs the upgrade exrpd after restart" running "$UPGRADE_DIR"
	stop_check "$name stops cleanly after restart" TERM
}

set_ports() {
	local c=$node_home/config/config.toml a=$node_home/config/app.toml
	sed -i \
		-e "s|^proxy_app = .*|proxy_app = \"tcp://127.0.0.1:$((port_base + 58))\"|" \
		-e "/^\[rpc\]/,/^\[/ s|^laddr = .*|laddr = \"tcp://127.0.0.1:$((port_base + 57))\"|" \
		-e "/^\[p2p\]/,/^\[/ s|^laddr = .*|laddr = \"tcp://127.0.0.1:$((port_base + 56))\"|" \
		-e "s|^pprof_laddr = .*|pprof_laddr = \"127.0.0.1:$((port_base + 60))\"|" \
		-e "s|^prometheus_listen_addr = .*|prometheus_listen_addr = \"127.0.0.1:$((port_base + 66))\"|" \
		-e 's|^timeout_commit = .*|timeout_commit = "1s"|' "$c" || return
	sed -i \
		-e "/^\[api\]/,/^\[/ s|^address = .*|address = \"tcp://127.0.0.1:$((port_base + 17))\"|" \
		-e "/^\[grpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((port_base + 90))\"|" \
		-e "s|^geth-metrics-address = .*|geth-metrics-address = \"127.0.0.1:$((port_base + 81))\"|" \
		-e "/^\[json-rpc\]/,/^\[/ s|^address = .*|address = \"127.0.0.1:$((port_base + 45))\"|" \
		-e "s|^ws-address = .*|ws-address = \"127.0.0.1:$((port_base + 46))\"|" \
		-e "s|^metrics-address = .*|metrics-address = \"127.0.0.1:$((port_base + 65))\"|" "$a"
}

# init_chain EXRPD creates the home with `EXRPD init` and sets up a one-validator chain.
init_chain() {
	local exrpd=$1 g=$node_home/config/genesis.json addr
	"$exrpd" init localnet --chain-id "$CHAIN_ID" --home "$node_home" >>"$node_log" 2>&1 || return
	"$exrpd" --home "$node_home" config set client chain-id "$CHAIN_ID" >/dev/null || return
	"$exrpd" --home "$node_home" config set client keyring-backend test >/dev/null || return
	"$exrpd" --home "$node_home" config set client node "$node_rpc" >/dev/null || return

	echo "$MNEMONIC" | "$exrpd" --home "$node_home" keys add alice --recover --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return
	"$exrpd" --home "$node_home" keys add bob --keyring-backend test --algo eth_secp256k1 >/dev/null 2>&1 || return

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

	addr=$("$exrpd" --home "$node_home" keys show alice -a --keyring-backend test) || return
	"$exrpd" --home "$node_home" genesis add-genesis-account "$addr" 1000000apoa,1000000000000000000000000000axrp --keyring-backend test || return
	"$exrpd" --home "$node_home" genesis gentx alice 1000000apoa --gas-prices 0axrp --keyring-backend test --chain-id "$CHAIN_ID" >/dev/null 2>&1 || return
	"$exrpd" --home "$node_home" genesis collect-gentxs >/dev/null 2>&1 || return
	"$exrpd" --home "$node_home" genesis validate >/dev/null || return

	set_ports
}

wait_upgrade_halt() {
	local i
	for ((i = 0; i < 240; i++)); do
		grep -q "UPGRADE \"$UPGRADE_PLAN\" NEEDED" "$node_log" && return 0
		sleep 0.5
	done
	return 1
}

propose_upgrade() {
	local hash tx pid status deadline
	upgrade_height=$(($(height) + 60))
	log "proposing $UPGRADE_PLAN at height $upgrade_height"
	hash=$(cli tx upgrade software-upgrade "$UPGRADE_PLAN" --upgrade-height "$upgrade_height" --title e2e --summary e2e \
		--deposit 1axrp --no-validate --from alice --fees "$FEES" --gas 500000 -y -o json | jq -r .txhash) || die "submit proposal"
	tx=$(wait_tx "$hash") || die "proposal tx $hash not included"
	[[ $(jq -r .code <<<"$tx") == 0 ]] || die "proposal tx $hash failed: $(jq -r .raw_log <<<"$tx")"
	pid=$(jq -r '[.events[] | select(.type == "submit_proposal") | .attributes[] | select(.key == "proposal_id") | .value][0]' <<<"$tx")
	[[ $pid =~ ^[0-9]+$ ]] || die "no proposal id in tx $hash"

	cli tx gov vote "$pid" yes --from alice --fees "$FEES" --gas 300000 -y -o json >/dev/null || die "vote"

	status=
	deadline=$((SECONDS + 60))
	while ((SECONDS < deadline)); do
		status=$(timeout 5 "$(cur)" --home "$node_home" --node "$node_rpc" query gov proposal "$pid" -o json 2>/dev/null | jq -r '.proposal.status // .status')
		[[ $status == PROPOSAL_STATUS_PASSED || $status == 3 ]] && break
		sleep 1
	done

	check "proposal $pid passed ($status)" test "$status" = PROPOSAL_STATUS_PASSED -o "$status" = 3
}

# kill_home SIGNAL bounds both supervisor and child cleanup, including a forced stop.
kill_home() {
	local signal=$1 deadline p
	kill "-$signal" "$supervisor_pid" 2>/dev/null
	for p in $(children); do kill "-$signal" "$p" 2>/dev/null; done

	deadline=$((SECONDS + 15))
	while ((SECONDS < deadline)); do
		[[ -z $(children) ]] && ! kill -0 "$supervisor_pid" 2>/dev/null && break
		sleep 0.5
	done

	kill -KILL "$supervisor_pid" 2>/dev/null
	for p in $(children); do kill -KILL "$p" 2>/dev/null; done
	deadline=$((SECONDS + 5))
	while ((SECONDS < deadline)); do
		[[ -z $(children) ]] && ! kill -0 "$supervisor_pid" 2>/dev/null && break
		sleep 0.5
	done

	[[ -z $(children) ]] && ! kill -0 "$supervisor_pid" 2>/dev/null ||
		die "processes of $node_home survived SIGKILL"
	wait "$supervisor_pid" 2>/dev/null
	supervisor_pid=
}

CHAINID="xrplevm_1449999-1"
MONIKER="localnet"
# Remember to change to other types of keyring like 'file' in-case exposing to outside world,
# otherwise your balance will be wiped quickly
# The keyring test does not require private key to steal tokens from you
KEYRING="test"
KEYALGO="eth_secp256k1"
LOGLEVEL="info"
# Set dedicated home directory for the evmosd instance
HOMEDIR="$PWD/.exrpd"
# to trace evm
#TRACE="--trace"
TRACE=""

# feemarket params basefee
BASEFEE=0

# Path variables
CONFIG=$HOMEDIR/config/config.toml
APP_TOML=$HOMEDIR/config/app.toml
GENESIS=$HOMEDIR/config/genesis.json
TMP_GENESIS=$HOMEDIR/config/tmp_genesis.json

#Account variables
KEY_NAME="alice"
MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"

rm -rf $HOMEDIR

make build

bin/exrpd --home "$HOMEDIR" config set client chain-id "$CHAINID" --chain-id "$CHAINID"
bin/exrpd --home "$HOMEDIR" config set client keyring-backend "$KEYRING"

echo "$MNEMONIC" | bin/exrpd --home "$HOMEDIR" keys add "$KEY_NAME" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"

bin/exrpd --home "$HOMEDIR" init "$MONIKER" -o --chain-id "$CHAINID"

jq '.consensus.params["block"]["max_gas"]="10500000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="axrp"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="axrp"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="axrp"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["amount"]="1"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["voting_period"]="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["expedited_voting_period"]="5s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["bond_denom"]="apoa"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["unbonding_time"]="60s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

jq '.app_state.bank.denom_metadata=[{"description":"XRP is the gas token","denom_units":[{"denom":"axrp"},{"denom":"xrp","exponent":18}],"base":"axrp","display":"xrp","name":"XRP","symbol":"XRP"}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

#jq '.app_state["erc20"]["token_pairs"][0]["denom"]="axrp"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
#jq '.app_state["erc20"]["token_pairs"][0]["owner_address"]="ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.erc20.native_precompiles=["0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.erc20.token_pairs=[{contract_owner:1,erc20_address:"0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",denom:"axrp",enabled:true,"owner_address":"ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2"}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"


jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["slashing"]["params"]["slash_fraction_downtime"]="0"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

bin/exrpd --home "$HOMEDIR" genesis add-genesis-account "$(bin/exrpd --home "$HOMEDIR" keys show "$KEY_NAME" -a --keyring-backend "$KEYRING")" 1000000apoa,1000000000000000000000000000axrp --keyring-backend "$KEYRING"

bin/exrpd --home "$HOMEDIR" genesis add-genesis-account "ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2" 1000000000000000000000000000axrp --keyring-backend "$KEYRING"

bin/exrpd --home "$HOMEDIR" genesis gentx alice 1000000apoa --gas-prices ${BASEFEE}axrp --keyring-backend "$KEYRING" --chain-id "$CHAINID"

bin/exrpd --home "$HOMEDIR" genesis collect-gentxs

bin/exrpd --home "$HOMEDIR" genesis validate

if [[ $1 == "pending" ]]; then
		if [[ "$OSTYPE" == "darwin"* ]]; then
			sed -i '' 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i '' 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		else
			sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' "$CONFIG"
			sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' "$CONFIG"
			sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' "$CONFIG"
			sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' "$CONFIG"
			sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' "$CONFIG"
			sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' "$CONFIG"
		fi
	fi

	# enable prometheus metrics and all APIs for dev node
	if [[ "$OSTYPE" == "darwin"* ]]; then
		sed -i '' 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i '' 's/prometheus-retention-time = 0/prometheus-retention-time  = 1000000000000/g' "$APP_TOML"
		sed -i '' 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i '' 's/enable = false/enable = true/g' "$APP_TOML"
		# Don't enable memiavl by default
		grep -q -F '[memiavl]' "$APP_TOML" && sed -i '' '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
	else
		sed -i 's/prometheus = false/prometheus = true/' "$CONFIG"
		sed -i 's/prometheus-retention-time  = "0"/prometheus-retention-time  = "1000000000000"/g' "$APP_TOML"
		sed -i 's/enabled = false/enabled = true/g' "$APP_TOML"
		sed -i 's/enable = false/enable = true/g' "$APP_TOML"
		# Don't enable memiavl by default
		grep -q -F '[memiavl]' "$APP_TOML" && sed -i '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
	fi

bin/exrpd start \
	--metrics "$TRACE" \
	--log_level $LOGLEVEL \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--home "$HOMEDIR" \
	--chain-id "$CHAINID"
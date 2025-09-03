#!/bin/bash

set -ex

# TODO: remove this script and just use the local node script for it, add flag to start node in given directory

CHAINID="${CHAIN_ID:-exrp_1440002-1}"
EVMCHAINID="1440002"
MONIKER="localtestnet"
KEYRING="test"          # remember to change to other types of keyring like 'file' in-case exposing to outside world, otherwise your balance will be wiped quickly. The keyring test does not require private key to steal tokens from you
KEYALGO="eth_secp256k1" #gitleaks:allow
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""
PRUNING="default"
#PRUNING="custom"

CHAINDIR="$HOME/.tmp-exrpd-solidity-tests" # TODO: make configurable like chain id
GENESIS="$CHAINDIR/config/genesis.json"
TMP_GENESIS="$CHAINDIR/config/tmp_genesis.json"
APP_TOML="$CHAINDIR/config/app.toml"
CONFIG_TOML="$CHAINDIR/config/config.toml"
BINARY="$PWD/../../bin/exrpd"

# make sure to reset chain directory before test
rm -rf "$CHAINDIR"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# feemarket params basefee
BASEFEE=0

# Set client config
"$BINARY" config set client chain-id "$CHAINID" --home "$CHAINDIR"
"$BINARY" config set client keyring-backend "$KEYRING" --home "$CHAINDIR"

# myKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
VAL_KEY="mykey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

# user1 address 0xc6fe5d33615a1c52c08018c47e8bc53646a0e101
USER1_KEY="user1"
USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

# user2 address 0x963ebdf2e1f8db8707d05fc75bfeffba1b5bac17
USER2_KEY="user2"
USER2_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

# user3 address 0x40a0cb1C63e026A81B55EE1308586E21eec1eFa9
USER3_KEY="user3"
USER3_MNEMONIC="will wear settle write dance topic tape sea glory hotel oppose rebel client problem era video gossip glide during yard balance cancel file rose"

# user4 address 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA
USER4_KEY="user4"
USER4_MNEMONIC="doll midnight silk carpet brush boring pluck office gown inquiry duck chief aim exit gain never tennis crime fragile ship cloud surface exotic patch"

SOLIDITY_KEY="solidity"
SOLIDITY_MNEMONIC="exercise green picture marriage cause bike credit electric elephant someone march civil radio spoon sail vacant crime man fat save inject into grab drill"

# Import keys from mnemonics
echo "$VAL_MNEMONIC" | "$BINARY" keys add "$VAL_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$CHAINDIR"
echo "$USER1_MNEMONIC" | "$BINARY" keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$CHAINDIR"
echo "$USER2_MNEMONIC" | "$BINARY" keys add "$USER2_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$CHAINDIR"
echo "$USER3_MNEMONIC" | "$BINARY" keys add "$USER3_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$CHAINDIR"
echo "$USER4_MNEMONIC" | "$BINARY" keys add "$USER4_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO" --home "$CHAINDIR"
if ! "$BINARY" keys show "$SOLIDITY_KEY" --home "$CHAINDIR" >/dev/null 2>&1; then
	echo "$SOLIDITY_MNEMONIC" | "$BINARY" keys add "$SOLIDITY_KEY" --recover --algo "$KEYALGO" --home "$CHAINDIR"
fi

# Set moniker and chain-id for Cosmos EVM (Moniker can be anything, chain-id must be an integer)
"$BINARY" init "$MONIKER" --chain-id "$CHAINID" --home "$CHAINDIR"

# Change parameter token denominations to apoa
jq '.consensus.params["block"]["max_gas"]="10500000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["crisis"]["constant_fee"]["denom"]="token"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["evm_denom"]="token"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["evm"]["params"]["allow_unprotected_txs"]=true' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="token"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["min_deposit"][0]["amount"]="1"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["voting_period"]="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["gov"]["params"]["expedited_voting_period"]="5s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["bond_denom"]="apoa"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["staking"]["params"]["unbonding_time"]="60s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Enable native denomination as a token pair for STRv2
jq '.app_state.erc20.native_precompiles=["0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.erc20.token_pairs=[{contract_owner:1,erc20_address:"0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE",denom:"token",enabled:true, "owner_address":"ethm1cml96vmptgw99syqrrz8az79xer2pcgp767p9e"}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

jq '.app_state["slashing"]["params"]["slash_fraction_double_sign"]="0"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state["slashing"]["params"]["slash_fraction_downtime"]="0"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

jq '.app_state["bank"]["denom_metadata"]=[{"description":"The native token for exrp.","denom_units":[{"denom":"token","exponent":0,"aliases":["token"]},{"denom":"Token","exponent":18,"aliases":[]}],"base":"token","display":"Token","name":"Token","symbol":"TOKEN","uri":"","uri_hash":""}]' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Change proposal periods to pass within a reasonable time for local testing
sed -i.bak 's/"max_deposit_period": "172800s"/"max_deposit_period": "30s"/g' "$GENESIS"
sed -i.bak 's/"voting_period": "172800s"/"voting_period": "30s"/g' "$GENESIS"
sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "15s"/g' "$GENESIS"

# Set gas limit in genesis
jq '.consensus_params.block.max_gas="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set base fee in genesis
jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# disable produce empty block
sed -i.bak 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CONFIG_TOML"

# Allocate genesis accounts (cosmos formatted addresses)
"$BINARY" add-genesis-account "$("$BINARY" keys show "$VAL_KEY" -a --keyring-backend "$KEYRING" --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --keyring-backend "$KEYRING" --home "$CHAINDIR"
"$BINARY" add-genesis-account "$("$BINARY" keys show "$USER1_KEY" -a --keyring-backend "$KEYRING" --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --keyring-backend "$KEYRING" --home "$CHAINDIR"
"$BINARY" add-genesis-account "$("$BINARY" keys show "$USER2_KEY" -a --keyring-backend "$KEYRING" --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --keyring-backend "$KEYRING" --home "$CHAINDIR"
"$BINARY" add-genesis-account "$("$BINARY" keys show "$USER3_KEY" -a --keyring-backend "$KEYRING" --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --keyring-backend "$KEYRING" --home "$CHAINDIR"
"$BINARY" add-genesis-account "$("$BINARY" keys show "$USER4_KEY" -a --keyring-backend "$KEYRING" --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --keyring-backend "$KEYRING" --home "$CHAINDIR"
"$BINARY" add-genesis-account "$("$BINARY" keys show "$SOLIDITY_KEY" -a --home "$CHAINDIR")" 1000000apoa,1000000000000000000000000000token --home "$CHAINDIR"

# set custom pruning settings
if [ "$PRUNING" = "custom" ]; then
	sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
fi

# make sure the localhost IP is 0.0.0.0
sed -i.bak 's/localhost/0.0.0.0/g' "$CONFIG_TOML"
sed -i.bak 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"

# use timeout_commit 1s to make test faster
sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "100ms"/g' "$CONFIG_TOML"

# Sign genesis transaction
"$BINARY" gentx "$VAL_KEY" 1000000apoa --gas-prices ${BASEFEE}token --keyring-backend "$KEYRING" --chain-id "$CHAINID" --home "$CHAINDIR"
## In case you want to create multiple validators at genesis
## 1. Back to `evmd keys add` step, init more keys
## 2. Back to `evmd add-genesis-account` step, add balance for those
## 3. Clone this ~/.evmd home directory into some others, let's say `~/.clonedosd`
## 4. Run `gentx` in each of those folders
## 5. Copy the `gentx-*` folders under `~/.clonedosd/config/gentx/` folders into the original `~/.evmd/config/gentx`

# Enable the APIs for the tests to be successful
sed -i.bak 's/enable = false/enable = true/g' "$APP_TOML"

# Don't enable memiavl by default
grep -q -F '[memiavl]' "$APP_TOML" && sed -i.bak '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"

# Collect genesis tx
"$BINARY" collect-gentxs --home "$CHAINDIR"

# Run this to ensure everything worked and that the genesis file is setup correctly
"$BINARY" validate-genesis --home "$CHAINDIR"

# Start the node
"$BINARY" start "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001utest \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--chain-id "$CHAINID" \
	--home "$CHAINDIR"

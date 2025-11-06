#!/bin/bash

# ========== Helper: Read named parameters ==========
for arg in "$@"; do
  case $arg in
    --chainId=*) CHAIN_ID="${arg#*=}" ;;
    --nodeHome=*) NODE_HOME="${arg#*=}" ;;
    --exportedState=*) EXPORTED_STATE="${arg#*=}" ;;
    --version=*) VERSION="${arg#*=}" ;;
    *) echo "⚠️  Unknown argument: $arg" ;;
  esac
done

# ========== Fallback to env vars if not passed ==========
: "${CHAIN_ID:=${CHAIN_ID_ENV:-}}"
: "${NODE_HOME:=${NODE_HOME_ENV:-}}"
: "${EXPORTED_STATE:=${EXPORTED_STATE_ENV:-}}"
: "${VERSION:=${VERSION_ENV:-}}"

# ========== Validate dependencies ==========
echo "ℹ️   Checking required tools..."
for cmd in exrpd jq sponge; do
  if ! command -v "$cmd" &> /dev/null; then
    echo "❌ Error: Required tool '$cmd' not found. Please install it before running this script."
    exit 1
  fi
done
echo "✅   All tools found"

# ========== Validate required variables ==========
missing_vars=()
for var in CHAIN_ID NODE_HOME EXPORTED_STATE VERSION; do
  if [ -z "${!var}" ]; then
    missing_vars+=("$var")
  fi
done

if [ ${#missing_vars[@]} -ne 0 ]; then
  echo "❌ Error: The following required variables are missing:"
  for var in "${missing_vars[@]}"; do echo "   - $var"; done
  echo
  echo "👉 You can provide them as parameters or environment variables."
  echo
  echo "📌 Example usage:"
  echo "./test-upgrade.sh \\"
  echo "  --chainId=\"xrplevm_1440000-1\" \\"
  echo "  --nodeHome=\"./.exrpd\" \\"
  echo "  --exportedState=\"./exported.json\" \\"
  echo "  --version=\"9.0.0\" \\"
  echo
  echo "Or set them as environment variables before running:"
  echo "export CHAIN_ID=... && ./test-upgrade.sh"
  exit 1
fi

echo "ℹ️  Initializing node home in $NODE_HOME..."
exrpd config set client chain-id "$CHAIN_ID" --home "$NODE_HOME"
exrpd config set client keyring-backend test --home "$NODE_HOME"
exrpd keys add key --key-type eth_secp256k1  --keyring-backend test --home "$NODE_HOME"
exrpd init localnode --chain-id "$CHAIN_ID" --home "$NODE_HOME" > /dev/null 2>&1
echo "✅   Node home initialized "

echo "ℹ️  Configuring genesis..."
cp $EXPORTED_STATE genesis.json
echo "ℹ️  Modifying bank ibc denoms..."
for ((i=1; i< $(jq '.app_state.bank.denom_metadata | length' genesis.json); i++)); do
    jq --argjson i "$i" '.app_state.bank.denom_metadata[$i].denom_units[0].denom = .app_state.bank.denom_metadata[$i].base' genesis.json | sponge genesis.json
    jq --argjson i "$i" '.app_state.bank.denom_metadata[$i].display = .app_state.bank.denom_metadata[$i].base' genesis.json | sponge genesis.json
done

echo "ℹ️  Modifying ratelimit initial_height..."
jq '.app_state.ratelimit.hour_epoch.epoch_start_height = .initial_height' genesis.json | sponge genesis.json

echo "ℹ️  Modifying gov voting periods..."
jq '.app_state.gov.params.voting_period = "30s"' genesis.json | sponge genesis.json
jq '.app_state.gov.params.expedited_voting_period = "20s"' genesis.json | sponge genesis.json

mv genesis.json "$NODE_HOME"/config/genesis.json
echo "✅   Genesis configured "

echo "ℹ️  Configuring local validator..."
echo "ℹ️  Adding local validator genesis account..."
exrpd add-genesis-account $(exrpd keys show key -a --keyring-backend test --home "$NODE_HOME") 100000000poa,10000xrp --home "$NODE_HOME"
echo "ℹ️  Creating local validator staking transaction..."
exrpd gentx key 100000000poa --fees 80000000000000000axrp --chain-id $CHAIN_ID --commission-rate 0 \
   --commission-max-rate 0 --commission-max-change-rate 0 --timeout-height $(jq '.initial_height' $EXPORTED_STATE) \
   --keyring-backend test --home "$NODE_HOME" \
   --account-number $(jq '.app_state.auth.accounts | length' $EXPORTED_STATE) --sequence 0 --offline --gas 400000
exrpd collect-gentxs --home "$NODE_HOME" > /dev/null 2>&1
exrpd validate-genesis --home "$NODE_HOME"
echo "✅   Local validator configured "

jq '.initial_height = (.initial_height | tostring)' "$NODE_HOME"/config/genesis.json | sponge "$NODE_HOME"/config/genesis.json

echo "ℹ️  Starting the node in background process..."
exrpd start --home "$NODE_HOME" &

sleep 10

INITIAL_HEIGHT=$(jq '.initial_height' .exrpd/config/genesis.json)
INITIAL_HEIGHT=${INITIAL_HEIGHT//\"/}
PROPOSAL_HEIGHT=$(( INITIAL_HEIGHT + 15 ))
echo "ℹ️  Creating upgrade proposal for version $VERSION and height $PROPOSAL_HEIGHT..."
cat <<EOF > upgrade-proposal.json
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "ethm10d07y265gmmuvt4z0w9aw880jnsr700jpva843",
      "plan": {
        "name": "v$VERSION",
        "height": "$PROPOSAL_HEIGHT",
        "info": "{\"binaries\":{\"linux/amd64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_$VERSION_Linux_amd64.tar.gz\",\"linux/arm64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_$VERSION_Linux_arm64.tar.gz\",\"darwin/amd64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_$VERSION_Darwin_amd64.tar.gz\",\"darwin/arm64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_$VERSION_Darwin_arm64.tar.gz\"}}"
      }
    }
  ],
  "title": "Protocol upgrade to v$VERSION",
  "summary": "This proposal will execute a protocol upgrade to the version $VERSION (https://github.com/xrplevm/node/releases/tag/v$VERSION)",
  "metadata": "ipfs://QmRWbE8bibBjtacYaaJ1suRpxFiHuYCD1T4Uz6QV27GXKH",
  "deposit": "500000000000000000000axrp"
}
EOF

exrpd --home "$NODE_HOME" tx gov submit-proposal ./upgrade-proposal.json --from key --gas-prices 800000000000axrp --gas 400000 --yes
echo "✅   Proposal created"

sleep 10
PROPOSAL_ID=$(exrpd q gov proposals --home $NODE_HOME -o json --page-count-total | jq '.pagination.total')
PROPOSAL_ID=${PROPOSAL_ID//\"/}
echo "ℹ️  Voting for proposal $PROPOSAL_ID..."

exrpd --home "$NODE_HOME" tx gov vote "$PROPOSAL_ID" yes --from key --gas-prices 800000000000axrp --yes --broadcast-mode async
echo "✅   Proposal voted"

rm upgrade-proposal.json
wait

#!/bin/bash

# ==========================================
# CONFIGURE THESE VALUES BEFORE RUNNING
# ==========================================
CHAIN_ID="xrplevm_1449000-1"
NODE_HOME="$HOME/.exrpd"
VERSION="10.0.0"
KEY_NAME="scure"
CURRENT_HEIGHT=5542868   # <-- set this to the current block height
# ==========================================

set -euo pipefail

PROPOSAL_HEIGHT=$(( CURRENT_HEIGHT + 5 ))

# ========== Validate dependencies ==========
for cmd in exrpd jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "❌ Error: Required tool '$cmd' not found."
    exit 1
  fi
done

# ========== Wait for node to be ready ==========
echo "ℹ️  Waiting for node to be ready..."
for i in $(seq 1 30); do
  if exrpd status --home "$NODE_HOME" &>/dev/null; then
    echo "✅   Node is ready"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "❌ Node did not become ready in time"
    exit 1
  fi
  sleep 2
done

echo "ℹ️  Upgrade proposal height: $PROPOSAL_HEIGHT (current + 30)"

# ========== Submit upgrade proposal ==========
echo "ℹ️  Creating upgrade proposal for v$VERSION at height $PROPOSAL_HEIGHT..."
cat > /tmp/upgrade-proposal.json << EOF
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "ethm10d07y265gmmuvt4z0w9aw880jnsr700jpva843",
      "plan": {
        "name": "v$VERSION",
        "height": "$PROPOSAL_HEIGHT",
        "info": "{\"binaries\":{\"linux/amd64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_${VERSION}_Linux_amd64.tar.gz\",\"linux/arm64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_${VERSION}_Linux_arm64.tar.gz\",\"darwin/amd64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_${VERSION}_Darwin_amd64.tar.gz\",\"darwin/arm64\":\"https://github.com/xrplevm/node/releases/download/v$VERSION/node_${VERSION}_Darwin_arm64.tar.gz\"}}"
      }
    }
  ],
  "title": "Protocol upgrade to v$VERSION",
  "summary": "This proposal will execute a protocol upgrade to the version $VERSION (https://github.com/xrplevm/node/releases/tag/v$VERSION)",
  "metadata": "ipfs://QmRWbE8bibBjtacYaaJ1suRpxFiHuYCD1T4Uz6QV27GXKH",
  "deposit": "1axrp"
}
EOF

SUBMIT_RESULT=$(exrpd tx gov submit-proposal /tmp/upgrade-proposal.json \
  --home "$NODE_HOME" \
  --from "$KEY_NAME" \
  --chain-id "$CHAIN_ID" \
  --keyring-backend test \
  --gas-prices 800000000000axrp \
  --gas 400000 \
  --yes \
  -o json)

rm /tmp/upgrade-proposal.json

SUBMIT_CODE=$(echo "$SUBMIT_RESULT" | jq -r '.code')
if [ "$SUBMIT_CODE" != "0" ]; then
  echo "❌ Proposal submission failed (code $SUBMIT_CODE):"
  echo "$SUBMIT_RESULT" | jq -r '.raw_log'
  exit 1
fi
echo "✅   Proposal submitted (tx: $(echo "$SUBMIT_RESULT" | jq -r '.txhash'))"

# ========== Get proposal ID and vote ==========
echo "ℹ️  Waiting for proposal to be indexed..."
sleep 5

PROPOSAL_ID=$(exrpd q gov proposals \
  --home "$NODE_HOME" \
  --chain-id "$CHAIN_ID" \
  -o json | jq -r '[.proposals[].id | tonumber] | max')
echo "ℹ️  Voting YES on proposal $PROPOSAL_ID..."

VOTE_RESULT=$(exrpd tx gov vote "$PROPOSAL_ID" yes \
  --home "$NODE_HOME" \
  --from "$KEY_NAME" \
  --chain-id "$CHAIN_ID" \
  --keyring-backend test \
  --gas-prices 800000000000axrp \
  --yes \
  --broadcast-mode sync \
  -o json)

VOTE_CODE=$(echo "$VOTE_RESULT" | jq -r '.code')
if [ "$VOTE_CODE" != "0" ]; then
  echo "❌ Vote failed (code $VOTE_CODE):"
  echo "$VOTE_RESULT" | jq -r '.raw_log'
  exit 1
fi
echo "✅   Voted YES on proposal $PROPOSAL_ID (tx: $(echo "$VOTE_RESULT" | jq -r '.txhash'))"

echo ""
echo "ℹ️  The node will halt at block $PROPOSAL_HEIGHT"
echo "    Install v$VERSION binary and restart with: exrpd start --home $NODE_HOME"

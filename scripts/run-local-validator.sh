#!/bin/bash

# run-local-validator.sh
#
# Runs unsafe-start-local-validator with keys derived from the node home.
#
# Usage:
#   ./scripts/run-local-validator.sh --nodeHome="$HOME/.exrpd" --keyName="key"

set -euo pipefail

# ========== Read named parameters ==========
for arg in "$@"; do
  case $arg in
    --nodeHome=*) NODE_HOME="${arg#*=}" ;;
    --keyName=*) KEY_NAME="${arg#*=}" ;;
    *) echo "⚠️  Unknown argument: $arg" ;;
  esac
done

: "${NODE_HOME:=${NODE_HOME_ENV:-$HOME/.exrpd}}"
: "${KEY_NAME:=${KEY_NAME_ENV:-key}}"

# ========== Validate ==========
PRIV_VAL_KEY="$NODE_HOME/config/priv_validator_key.json"
if [ ! -f "$PRIV_VAL_KEY" ]; then
  echo "❌ priv_validator_key.json not found at $PRIV_VAL_KEY"
  exit 1
fi

# ========== Extract values ==========
# jq -r strips the surrounding quotes that plain jq adds
PUBKEY=$(jq -r '.pub_key.value' "$PRIV_VAL_KEY")
PRIVKEY=$(jq -r '.priv_key.value' "$PRIV_VAL_KEY")
VALOPER=$(exrpd keys show "$KEY_NAME" --bech val -a --keyring-backend test --home "$NODE_HOME")
ACCOUNT=$(exrpd keys show "$KEY_NAME" -a --keyring-backend test --home "$NODE_HOME")

echo "ℹ️  Starting unsafe-start-local-validator..."
echo "    Home:     $NODE_HOME"
echo "    Operator: $VALOPER"
echo "    Account:  $ACCOUNT"
echo ""

exrpd unsafe-start-local-validator \
  --home "$NODE_HOME" \
  --validator-operator="$VALOPER" \
  --validator-pubkey="$PUBKEY" \
  --validator-privkey="$PRIVKEY" \
  --accounts-to-fund="$ACCOUNT" \
  --p2p.seeds="" \
  --p2p.persistent_peers=""

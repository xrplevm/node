#!/bin/bash

MNEMONIC="birth rebuild refuse area aisle language bullet pride place clutch paddle drama"

# ========== Helper: Read named parameters ==========
for arg in "$@"; do
  case $arg in
    --chainId=*) CHAIN_ID="${arg#*=}" ;;
    --nodeHome=*) NODE_HOME="${arg#*=}" ;;
    --snapshotPath=*) SNAPSHOT_PATH="${arg#*=}" ;;
    --genesisUrl=*) GENESIS_URL="${arg#*=}" ;;
    --version=*) VERSION="${arg#*=}" ;;
    *) echo "⚠️  Unknown argument: $arg" ;;
  esac
done

# ========== Fallback to env vars if not passed ==========
: "${CHAIN_ID:=${CHAIN_ID_ENV:-}}"
: "${NODE_HOME:=${NODE_HOME_ENV:-}}"
: "${SNAPSHOT_PATH:=${SNAPSHOT_PATH_ENV:-}}"
: "${GENESIS_URL:=${GENESIS_URL_ENV:-}}"
: "${VERSION:=${VERSION_ENV:-}}"

# ========== Validate dependencies ==========
echo "ℹ️   Checking required tools..."
for cmd in exrpd jq lz4; do
  if ! command -v "$cmd" &> /dev/null; then
    echo "❌ Error: Required tool '$cmd' not found. Please install it before running this script."
    exit 1
  fi
done
echo "✅   All tools found"

# ========== Validate required variables ==========
missing_vars=()
for var in CHAIN_ID NODE_HOME SNAPSHOT_PATH GENESIS_URL VERSION; do
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
  echo "  --snapshotPath=\"./exrpd.tar.lz4\" \\"
  echo "  --genesisUrl=\"https://raw.githubusercontent.com/xrplevm/networks/refs/heads/main/testnet/genesis.json\" \\"
  echo "  --version=\"9.0.0\" \\"
  echo
  echo "Or set them as environment variables before running:"
  echo "export CHAIN_ID=... && ./test-upgrade.sh"
  exit 1
fi

echo "ℹ️  Initializing node home in $NODE_HOME..."
exrpd config set client chain-id "$CHAIN_ID" --home "$NODE_HOME" --chain-id "$CHAIN_ID"
exrpd config set client keyring-backend test --home "$NODE_HOME"
echo "$MNEMONIC" | exrpd --home "$NODE_HOME" keys add key --recover --keyring-backend test --algo eth_secp256k1
exrpd init localnode --chain-id "$CHAIN_ID" --home "$NODE_HOME" > /dev/null 2>&1
echo "✅   Node home initialized "

echo "ℹ️  Downloading genesis..."
wget -O $NODE_HOME/config/genesis.json $GENESIS_URL
echo "✅   Genesis downloaded "

echo "ℹ️  Restoring snapshot..."
lz4 -c -d $SNAPSHOT_PATH  | tar -x -C $NODE_HOME
echo "✅   Snapshot restored "

echo "ℹ️  Starting the local validator..."

exrpd unsafe-start-local-validator \
  --home "$NODE_HOME" \
  --validator-operator="$(exrpd --home "$NODE_HOME" keys show key --keyring-backend test --bech val -a)" \
  --validator-pubkey="$(jq -r '.pub_key.value' "$NODE_HOME/config/priv_validator_key.json")" \
  --validator-privkey="$(jq -r '.priv_key.value' "$NODE_HOME/config/priv_validator_key.json")" \
  --accounts-to-fund="$(exrpd --home "$NODE_HOME" keys show key --keyring-backend test --bech acc -a),ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2" \
  --xrp-owner-address="$(exrpd --home "$NODE_HOME" keys show key --keyring-backend test --bech acc -a)" \
  --upgrade-version="$VERSION"

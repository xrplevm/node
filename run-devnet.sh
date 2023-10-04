NODE_HOME=./.node
# CMD=/Users/adriacarrera/GIT/xrp-evm/packages/blockchain/build/exrpd
# CMD=/Users/adriacarrera/GIT/exrp/build/exrpd
CMD=exrpd

rm -rf $NODE_HOME

$CMD --home $NODE_HOME config chain-id exrp_1440002-1
$CMD --home $NODE_HOME keys add mykey --keyring-backend test
$CMD --home $NODE_HOME init testval --chain-id exrp_1440002-1

cp genesis.json $NODE_HOME/config/
$CMD --home $NODE_HOME validate-genesis

PEERS=`curl -sL https://raw.githubusercontent.com/Peersyst/xrp-evm-archive/main/poa-devnet/peers.txt | sort -R | head -n 10 | awk '{print $1}' | paste -s -d, -`
sed -i.bak -e "s/^persistent_peers *=.*/persistent_peers = \"$PEERS\"/" $NODE_HOME/config/config.toml


$CMD --home $NODE_HOME start

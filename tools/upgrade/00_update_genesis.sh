#!/bin/bash

DEFAULT_NODE_HOME=$HOME/.exrp

exrpd tendermint unsafe-reset-all
node src/build-genesis.js genesis-state-devnet.json $DEFAULT_NODE_HOME/config/genesis.json
docker run -it --rm --name validator -p 26657:26657 -v $DEFAULT_NODE_HOME/config:/root/.exrp/config -v $DEFAULT_NODE_HOME/keyring-test:/root/.exrp/keyring-test -v $DEFAULT_NODE_HOME/data:/root/.exrp/data cosmovisor cosmovisor run start

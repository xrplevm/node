#!/bin/bash

PROPOSAL_ID=12

exrpd tx gov submit-proposal ./upgrade-proposal.json --from alice --gas-prices 7axrp --yes --gas 300000
exrpd tx gov vote $PROPOSAL_ID yes --from alice --gas-prices 7axrp --yes
sleep 60
exrpd query gov proposals
#!/bin/sh

mkdir -p $DAEMON_HOME/cosmovisor/genesis/bin
ln -s /usr/local/bin/exrpd $DAEMON_HOME/cosmovisor/genesis/bin/exrpd
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin
cp /usr/local/bin/exrpd_v2.0.0 $DAEMON_HOME/cosmovisor/upgrades/v2.0.0/bin/exrpd
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v3.0.0/bin
cp /usr/local/bin/exrpd_v3.0.0 $DAEMON_HOME/cosmovisor/upgrades/v3.0.0/bin/exrpd

ln -s $DAEMON_HOME/cosmovisor/upgrades/v2.0.0 $DAEMON_HOME/cosmovisor/current
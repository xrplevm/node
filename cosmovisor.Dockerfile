FROM --platform=linux golang:1.20

ENV COSMOS_VERSION=v0.46.12
ENV DAEMON_HOME=/root/.exrpd
ENV DAEMON_NAME=exrpd
ENV DAEMON_ALLOW_DOWNLOAD_BINARIES=true
ENV DAEMON_RESTART_AFTER_UPGRADE=true

WORKDIR /root
RUN git clone --depth 1 --branch $COSMOS_VERSION https://github.com/cosmos/cosmos-sdk.git

WORKDIR /root/cosmos-sdk/cosmovisor
RUN make cosmovisor

RUN mv /root/cosmos-sdk/cosmovisor/cosmovisor /usr/local/bin/cosmovisor
COPY --from=peersyst/xrp-evm-blockchain:latest /usr/bin/exrpd $DAEMON_HOME/cosmovisor/genesis/bin/exrpd

RUN ls -la $DAEMON_HOME/cosmovisor/genesis/bin/
RUN cp $DAEMON_HOME/cosmovisor/genesis/bin/exrpd /usr/local/bin/exrpd
RUN ls -la /usr/local/bin/

RUN mkdir $DAEMON_HOME/data

CMD ["cosmovisor"]
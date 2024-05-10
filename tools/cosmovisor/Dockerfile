FROM golang:1.20-alpine AS build

ENV COSMOS_VERSION=v0.46.12

RUN apk update
RUN apk add --no-cache gcc libc-dev make git

WORKDIR /root
RUN git clone --depth 1 --branch $COSMOS_VERSION https://github.com/cosmos/cosmos-sdk.git
WORKDIR /root/cosmos-sdk/cosmovisor
RUN make cosmovisor


FROM alpine:3.17.3

ENV DAEMON_HOME=/root/.exrpd
ENV DAEMON_NAME=exrpd
ENV DAEMON_ALLOW_DOWNLOAD_BINARIES=true
ENV DAEMON_RESTART_AFTER_UPGRADE=true

COPY --from=build /root/cosmos-sdk/cosmovisor/cosmovisor /usr/local/bin/cosmovisor
COPY --from=peersyst/xrp-evm-blockchain:latest /usr/bin/exrpd /usr/local/bin/exrpd

ADD init.sh /usr/local/bin/initialize
RUN chmod +x /usr/local/bin/initialize

RUN apk add gcompat
RUN ln -sf ../../lib/libgcompat.so.0 /usr/lib/libresolv.so.2

CMD ["cosmovisor"]

FROM golang:1.20 AS base
USER root
RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates \
        curl

RUN curl https://get.ignite.com/cli@v0.27.2! | bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ARG WITNESSES=68f727c3cd7aeB5a04acB864B770f5aa193676Bd
ARG THRESHOLD=1
ARG BRIDGE

COPY . .

RUN ignite chain build
RUN ignite chain init --home /.exrpd
RUN exrpd --home /.exrpd add-genesis-contracts ${WITNESSES} ${THRESHOLD} ${BRIDGE}

ENTRYPOINT ["exrpd"]
CMD ["start", "--home", "/.exrpd"]
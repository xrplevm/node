FROM golang:1.22.11 AS base
USER root
RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates
WORKDIR /app
COPY . .

# Hotfix to allow download of private go module
ENV GOPRIVATE=github.com/cometbft/cometbft-sec-tachyon
RUN mkdir -p ~/.ssh
RUN --mount=type=secret,id=ssh_key,env=SSH_KEY  echo $SSH_KEY > ~/.ssh/id_rsa
RUN chmod 600 ~/.ssh/id_rsa
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."ssh://git@github.com/cometbft/cometbft-sec-tachyon".insteadOf "https://github.com/cometbft/cometbft-sec-tachyon"

RUN make install


FROM base AS build
ARG VERSION=0.0.0
RUN make build


FROM base AS integration
RUN make lint
# Unit tests
RUN make test-poa
# Integration tests
RUN make test-integration
RUN make test-upgrade
# Simulation tests
RUN make test-sim-benchmark-simulation
RUN make test-sim-full-app-fast

RUN touch /test.lock

FROM golang:1.22.11 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]
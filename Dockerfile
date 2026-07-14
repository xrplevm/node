# syntax=docker/dockerfile:1
FROM golang:1.23.8 AS base
USER root
RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates \
        openssh-client
WORKDIR /app

# Private Go module access via a forwarded SSH agent (no key is written to disk)
ENV GOPRIVATE=github.com/xrplevm/evm-priv-jul2026
RUN mkdir -p -m 0700 /root/.ssh && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts && \
    git config --global \
      url."ssh://git@github.com/xrplevm/evm-priv-jul2026".insteadOf \
      "https://github.com/xrplevm/evm-priv-jul2026"

COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY . .
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
# Simulation tests
# TODO: Restore simulation tests if possible
# RUN make test-sim-benchmark-simulation
# RUN make test-sim-full-app-fast

RUN touch /test.lock

FROM golang:1.23.8 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]

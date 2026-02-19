FROM golang:1.23.8 AS base
USER root
RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates
WORKDIR /app
COPY . .

# Hotfix to allow download of private go module
ENV GOPRIVATE=github.com/xrplevm/evm-sec-papyrus
RUN mkdir -p ~/.ssh
RUN --mount=type=secret,id=ssh_key_b64 base64 -d -i /run/secrets/ssh_key_b64 > ~/.ssh/id_rsa
RUN chmod 600 ~/.ssh/id_rsa
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."ssh://git@github.com/xrplevm/evm-sec-papyrus".insteadOf "https://github.com/xrplevm/evm-sec-papyrus"

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
RUN make test-sim-benchmark-simulation
RUN make test-sim-full-app-fast

RUN touch /test.lock

FROM golang:1.23.8 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]
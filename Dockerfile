FROM ignitehq/cli:v0.26.1 AS base
USER root
WORKDIR /go/src/github.com/Peersyst/exrp
COPY . .


FROM base AS build
RUN ignite chain build --release
RUN tar -xf /go/src/github.com/Peersyst/exrp/release/exrp_linux_amd64.tar.gz -C /go/src/github.com/Peersyst/exrp/release


FROM build AS integration
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
RUN golangci-lint run --timeout=10m
# Unit tests
RUN go test $(go list ./... | grep -v github.com/Peersyst/exrp/tests/e2e/poa)
# End to end tests
WORKDIR /go/src/github.com/Peersyst/exrp/tests/e2e/poa
RUN go test
RUN touch /test.lock

FROM golang:1.20 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /go/src/github.com/Peersyst/exrp/release/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]
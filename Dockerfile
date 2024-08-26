FROM golang:1.22.2 AS base
USER root
RUN apt update && \
    apt-get install -y \
        build-essential \
        ca-certificates
WORKDIR /app
COPY . .
RUN make install


FROM base AS build
ARG VERSION=0.0.0
RUN make build


FROM base AS integration
RUN make lint
# Unit tests
RUN go test $(go list ./... | grep -v github.com/xrplevm/node/v2/tests/e2e)
# End to end tests
RUN TEST_CLEANUP_DIR=false go test -p 1 -v -timeout 30m ./tests/e2e/...
RUN touch /test.lock

FROM golang:1.22.2 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]
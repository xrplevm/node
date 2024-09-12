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
RUN go test $(go list ./... | grep -v github.com/xrplevm/node/v3/tests/e2e | grep -v github.com/xrplevm/node/v3/app)
# End to end tests
RUN <<EOF
#!/bin/bash
retry() {
  local retries="$1"
  local command="$2"
  $command
  local exit_code=$?
  if [[ $exit_code -ne 0 && $retries -gt 0 ]]; then
    retry $(($retries - 1)) "$command"
  else
    return $exit_code
  fi
}
# TODO: Enable end to end tests when performance is improved
# retry 5 "go test -p 1 -v -timeout 30m ./tests/e2e/..."
EOF
RUN touch /test.lock

FROM golang:1.22.2 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]
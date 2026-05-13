FROM golang:1.23.8@sha256:ec5612bbd9e96d5b80a8b968cea06a4a9b985fe200ff6da784bf607063273c59 AS base
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
RUN make test-poa
# Integration tests
RUN make test-integration
# Simulation tests
# TODO: Restore simulation tests if possible
# RUN make test-sim-benchmark-simulation
# RUN make test-sim-full-app-fast

RUN touch /test.lock

FROM golang:1.23.8@sha256:ec5612bbd9e96d5b80a8b968cea06a4a9b985fe200ff6da784bf607063273c59 AS release
WORKDIR /
COPY --from=integration /test.lock /test.lock
COPY --from=build /app/bin/exrpd /usr/bin/exrpd
ENTRYPOINT ["/bin/sh", "-ec"]
CMD ["exrpd"]

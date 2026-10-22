# Build environment for Anthropic OSS Scanner. Network access exists only while this image builds.
# The scanner uses the repository root as the build context. Exclusions come from
# .oss-scanner/Dockerfile.dockerignore, not from the root .dockerignore.
FROM golang:1.23.8@sha256:ec5612bbd9e96d5b80a8b968cea06a4a9b985fe200ff6da784bf607063273c59

ENV DEBIAN_FRONTEND=noninteractive \
    GOWORK=off \
    GOFLAGS=-mod=readonly \
    GOTOOLCHAIN=local \
    COSMOS_BUILD_OPTIONS=nostrip

RUN apt-get update && apt-get install -y --no-install-recommends \
      build-essential ca-certificates git jq \
    && rm -rf /var/lib/apt/lists/*

COPY . /src
WORKDIR /src

# The threat model puts CI and release configuration in scope. Fail early if the build context dropped it.
RUN test -d .github/workflows && test -f .github/CODEOWNERS && test -f Dockerfile

RUN go mod download && go mod verify

# exrpd with symbols and DWARF info, at /src/bin/exrpd.
RUN make build

RUN go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(cat .golangci-lint-version)" \
    && go install go.uber.org/mock/mockgen@v0.6.0

# Compile every test binary so the build cache is warm and no test needs the network.
RUN go test -run '^$' ./... && go test -tags=test -run '^$' ./tests/integration

# Cheap tests prove the image works. Failures do not abort the build.
RUN make test-poa || echo "WARNING: test-poa reported failures"
RUN make test-integration || echo "WARNING: test-integration reported failures"

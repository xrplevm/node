#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BINDIR ?= $(GOPATH)/bin
APP = ./app

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BFT_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::') # grab everything after the space in "github.com/cometbft/cometbft v0.34.7"
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.0.0-rc8
BUILDDIR ?= $(CURDIR)/build
export GO111MODULE = on
ROCKSDB_VERSION = "9.8.4"


# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=exrp \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=exrpd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(BFT_VERSION)

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb grocksdb_no_link
  VERSION := $(VERSION)-rocksdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif
 
#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])

all: install lint

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/exrpd

build:
	go build $(BUILD_FLAGS) -o ./bin/exrpd ./cmd/exrpd

build-rocksdb:
	# Make sure to run this command with root permission
	CGO_ENABLED=1 CGO_CFLAGS="-I/usr/include" \
	CGO_LDFLAGS="-L/usr/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl" \
	COSMOS_BUILD_OPTIONS=rocksdb $(MAKE) build


###############################################################################
###                                Linting                                  ###
###############################################################################
golangci_lint_cmd=golangci-lint
golangci_version=v1.62.0

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --timeout=10m

lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run --fix --out-format=tab --issues-exit-code=0

###############################################################################
###                                Testing                                  ###
###############################################################################
EXCLUDED_POA_PACKAGES=$(shell go list ./x/poa/... | grep -v /x/poa/testutil | grep -v /x/poa/client | grep -v /x/poa/simulation | grep -v /x/poa/types)
EXCLUDED_UNIT_PACKAGES=$(shell go list ./... | grep -v tests | grep -v testutil | grep -v tools | grep -v app | grep -v docs | grep -v cmd | grep -v /x/poa/testutil | grep -v /x/poa/client | grep -v /x/poa/simulation | grep -v /x/poa/types)

mocks:
	@echo "--> Installing mockgen"
	go install github.com/golang/mock/mockgen@v1.6.0
	@echo "--> Generating mocks"
	@./scripts/mockgen.sh

test: test-poa test-integration test-upgrade test-sim-benchmark-simulation test-sim-full-app-fast

test-upgrade:
	@echo "--> Running upgrade testsuite"
	@go test -mod=readonly -tags=test -v ./tests/upgrade

test-integration:
	@echo "--> Running integration testsuite"
	@go test -mod=readonly -tags=test -v ./tests/integration

test-poa:
	@echo "--> Running POA tests"
	@go test $(EXCLUDED_POA_PACKAGES) 

test-sim-benchmark-simulation:
	@echo "Running simulation invariant benchmarks..."
	cd ${CURDIR}/app && go test -mod=readonly -benchmem -bench=BenchmarkSimulation -run=^$ \
	-Enabled=true -NumBlocks=100 -BlockSize=200 -Params=${CURDIR}/tests/sim/params.json \
	-Period=1 -Commit=true -v -timeout 24h

test-sim-full-app-fast:
	@echo "Running custom genesis simulation..."
	@cd ${CURDIR}/app && go test -mod=readonly -run TestFullAppSimulation \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=5 -Params=${CURDIR}/tests/sim/params.json -v -timeout 24h

###############################################################################
###                                Coverage                                 ###
###############################################################################

coverage-unit:
	@echo "--> Running unit coverage"
	@go test $(EXCLUDED_UNIT_PACKAGES) -coverprofile=coverage_unit.out > /dev/null
	@go tool cover -func=coverage_unit.out

coverage-poa:
	@echo "--> Running POA coverage"
	@go test $(EXCLUDED_POA_PACKAGES) -coverprofile=coverage_poa.out > /dev/null
	@go tool cover -func=coverage_poa.out

coverage-integration:
	@echo "--> Running integration coverage"
	@go test ./tests/integration -mod=readonly -coverprofile=coverage_integration.out > /dev/null
	@go tool cover -func=coverage_integration.out

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.13.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./proto/scripts/protocgen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./proto/scripts/protoc-swagger-gen.sh

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

proto-update-deps:
	@echo "Updating Protobuf dependencies"
	$(DOCKER) run --rm -v $(CURDIR)/proto:/workspace --workdir /workspace $(protoImageName) buf mod update

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps

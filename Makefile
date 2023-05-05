#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
BuildTime :=$(shell date '+%Y-%m-%dT%H:%M:%SZ%z')
ldflags = '-X github.com/overload-ak/cosmos-firewall.Version=$(VERSION) \
           -X github.com/overload-ak/cosmos-firewall.Commit=$(COMMIT) \
           -X github.com/overload-ak/cosmos-firewall.BuildTime=$(BuildTime) \
           -w -s'
           
BUILDDIR ?= $(CURDIR)/build

###############################################################################
###                                  Build                                  ###
###############################################################################

all: build lint test

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify
	go mod tidy
	@echo "--> Download go modules to local cache"
	go mod download

build: go.sum
	go build -mod=readonly -v -ldflags $(ldflags) -o $(BUILDDIR)/bin/firewalld ./cmd/
	@echo "--> Done building."

build-linux:
	@GOOS=linux GOARCH=amd64 $(MAKE) build

# If you are using the default builder, you need to first run 'docker buildx create --name container --driver docker-container && docker buildx use container'."
build-docker:
	@docker buildx build --build-arg https_proxy --push --platform linux/amd64,linux/arm64 -t harbor.wokoworks.com/functionx/fx-helper:6.3.0 .

INSTALL_DIR := $(shell go env GOPATH)/bin
install: build $(INSTALL_DIR)
	mv $(BUILDDIR)/bin/firewalld $(shell go env GOPATH)/bin/firewalld
	@echo "--> Run \"firewalld start\" or \"$(shell go env GOPATH)/bin/firewalld start\" to launch firewalld."

$(INSTALL_DIR):
	@echo "Folder $(INSTALL_DIR) does not exist"
	mkdir -p $@

.PHONY: build build-win install docker go.sum


###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	@echo "--> Running linter"
	@which golangci-lint > /dev/null || echo "\033[91m install golangci-lint ...\033[0m" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@which gocyclo > /dev/null || echo "\033[91m install gocyclo ...\033[0m" && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@which gofumpt > /dev/null || echo "\033[91m install gofumpt ...\033[0m" && go install mvdan.cc/gofumpt@latest
	golangci-lint run -v --go=1.19 --timeout 10m
	find . -name '*.go' -type f -not -path "./build*" -not -path "*.git*" -not -name '*.pb.*' | xargs gofumpt -d | xargs gocyclo -over 15
	find . -name '*.go' -type f -not -path "./build*" -not -path "*.git*" -not -name '*.pb.*' | xargs gofumpt -d | xargs gofumpt -d

	

format: format-goimports
	find . -name '*.go' -type f -not -path "./build*" -not -path "./contract*" -not -path "./tests/contract*" -not -name "statik.go" -not -name "*.pb.go" -not -name "*.pb.gw.go" | xargs gofumpt -w -l
	golangci-lint run --fix

format-goimports:
	@go install github.com/incu6us/goimports-reviser/v3@latest
	@find . -name '*.go' -type f -not -path './build*' -not -name 'statik.go' -not -name '*.pb.go' -not -name '*.pb.gw.go' -exec goimports-reviser -use-cache -rm-unused {} \;


.PHONY: format lint format-goimports


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test:
	@echo "--> Running tests"
	go test -mod=readonly ./...

.PHONY: test

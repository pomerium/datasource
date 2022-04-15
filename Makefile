PREFIX?=$(shell pwd)
NAME := bamboohr
PKG := github.com/pomerium/datasource/bamboohr


BUILDDIR := ${PREFIX}/dist
BINDIR := ${PREFIX}/bin

GITCOMMIT := $(shell git rev-parse --short HEAD)
BUILDMETA:=
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	BUILDMETA := dirty
endif

CTIMEVAR=-X $(PKG)/version.GitCommit=$(GITCOMMIT) \
	-X $(PKG)/internal/.BuildMeta=$(BUILDMETA) \
	-X $(PKG)/internal/version.ProjectName=$(NAME) \
	-X $(PKG)/internal/version.ProjectURL=$(PKG)

GO ?= "go"
GO_LDFLAGS=-ldflags "-s -w $(CTIMEVAR)"
GOOSARCHES = linux/amd64 darwin/amd64 windows/amd64

.PHONY: all
all: clean lint test build

.PHONY: test
test: ## test everything
	go test ./...

.PHONY: lint
lint: ## run go mod tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint --timeout=120s run ./...

.PHONY: tidy
tidy: ## run go mod tidy
	go mod tidy -compat=1.17

.PHONY: clean
clean: ## Cleanup any build binaries or packages.
	@echo "==> $@"
	$(RM) -r $(BINDIR)
	$(RM) -r $(BUILDDIR)

.PHONY: build
build: ## Build everything.
	@echo "==> $@"
	@CGO_ENABLED=0 GO111MODULE=on go build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(BINDIR)/$(NAME) ./cmd/

.PHONY: snapshot
snapshot: ## Create release snapshot
	APPARITOR_GITHUB_TOKEN=foo goreleaser release --snapshot --rm-dist


.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

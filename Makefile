BINARY     := sshcloak
CMD        := ./cmd
INSTALL    := /usr/local/bin/$(BINARY)
USER_BIN   := $(HOME)/.local/bin
USER_INSTALL := $(USER_BIN)/$(BINARY)
DOCS_REQS  := docs/requirements.txt
DOCS_VENV  := .venv
MKDOCS     := $(DOCS_VENV)/bin/mkdocs
GOBIN      := $(shell go env GOPATH)/bin
GOLANGCI_VERSION := v2.12.2
VERSION_VAR := github.com/sipuaz/sshcloak/cmd/internal/cli.Version
VERSION     := $(shell cat VERSION 2>/dev/null | tr -d '[:space:]' || git describe --tags --always)
BUILD_FLAGS := -ldflags "-X $(VERSION_VAR)=$(VERSION)"


.PHONY: all build test test-integration tidy lint lint-install release-build install uninstall clean init \
	docs-install docs-serve docs-build docs-deploy
.PHONY: install-user

all: build

## build: compile the binary into the project root
build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(BINARY) $(CMD)

## test: run unit tests for all packages
test:
	go test ./...

## test-integration: run unit tests + integration tests for all packages
test-integration:
	go test -tags integration ./...

## tidy: ensure go.mod and go.sum are normalized
tidy:
	go mod tidy

## lint: run static checks and formatting linters
lint:
	@LINT=$$(command -v golangci-lint || echo $(GOBIN)/golangci-lint); \
	if [ ! -x "$$LINT" ]; then \
		echo "golangci-lint not found; run 'make lint-install'"; \
		exit 1; \
	fi; \
	"$$LINT" run ./...

## lint-install: install the golangci-lint version used by CI
lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	@echo "installed to $(GOBIN)"

## release-build: build a platform-specific release artifact (requires GOOS/GOARCH)
release-build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "-X $(VERSION_VAR)=$(VERSION)" \
		-o $(BINARY)-$(GOOS)-$(GOARCH) $(CMD)

## install: build and install the binary to /usr/local/bin
install: build
	install -m 0755 $(BINARY) $(INSTALL)

## install-user: build and install the binary to ~/.local/bin
install-user: build
	mkdir -p $(USER_BIN)
	install -m 0755 $(BINARY) $(USER_INSTALL)

## uninstall: remove the installed binary
uninstall:
	rm -f $(INSTALL)

## clean: remove the local build artifact
clean:
	rm -f $(BINARY)

## init: build the binary and bootstrap sshcloak for the current user
init: build
	./$(BINARY) init

## docs-install: create a venv and install Python doc dependencies
docs-install:
	python3 -m venv $(DOCS_VENV)
	$(DOCS_VENV)/bin/pip install -q -r $(DOCS_REQS)

## docs-serve: live-preview the docs at http://127.0.0.1:8000
docs-serve: $(MKDOCS)
	$(MKDOCS) serve

## docs-build: build the static site into site/
docs-build: $(MKDOCS)
	$(MKDOCS) build --strict

## docs-deploy: build and push to the gh-pages branch (requires push access)
docs-deploy: $(MKDOCS)
	$(MKDOCS) gh-deploy --force

# Internal: ensure the venv exists before doc targets that need it.
$(MKDOCS):
	$(MAKE) docs-install

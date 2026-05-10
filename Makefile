BINARY     := sshcloak
CMD        := ./cmd
INSTALL    := /usr/local/bin/$(BINARY)
DOCS_REQS  := docs/requirements.txt
DOCS_VENV  := .venv
MKDOCS     := $(DOCS_VENV)/bin/mkdocs
BUILD_FLAGS := -ldflags "-X main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)"

.PHONY: all build test test-integration install uninstall clean \
        docs-install docs-serve docs-build docs-deploy

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

## install: build and install the binary to /usr/local/bin
install: build
	install -m 0755 $(BINARY) $(INSTALL)

## uninstall: remove the installed binary
uninstall:
	rm -f $(INSTALL)

## clean: remove the local build artifact
clean:
	rm -f $(BINARY)

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

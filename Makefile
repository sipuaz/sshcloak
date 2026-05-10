BINARY     := sshcloak
CMD        := ./cmd
INSTALL    := /usr/local/bin/$(BINARY)
BUILD_FLAGS := -ldflags "-X main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)"

.PHONY: all build test test-integration install uninstall clean

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

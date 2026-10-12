BINARY   := gosorter
BINDIR   := bin
INSTALLDIR := /usr/local/bin

GO       := go
GOFLAGS  := -trimpath -ldflags="-s -w"
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: all build run install uninstall clean test cover tidy fmt vet release-all

all: test build

## build: compile the binary into ./bin/gosorter
build:
	@mkdir -p $(BINDIR)
	$(GO) build $(GOFLAGS) -o $(BINDIR)/$(BINARY) .
	@echo "Built: $(BINDIR)/$(BINARY)"

## run: build and run with ARGS (e.g. make run ARGS="-v ~/Downloads")
run: build
	$(BINDIR)/$(BINARY) $(ARGS)

## test: run unit tests
test:
	$(GO) test ./...

## cover: run tests with coverage summary
cover:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -func=coverage.out
	@echo "HTML report: $(GO) tool cover -html=coverage.out"

## tidy: run go mod tidy
tidy:
	$(GO) mod tidy

## fmt / vet
fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

## install: install the binary to /usr/local/bin
install: build
	install -m 0755 $(BINDIR)/$(BINARY) $(INSTALLDIR)/$(BINARY)
	@echo "Installed: $(INSTALLDIR)/$(BINARY)"

## uninstall: remove the installed binary
uninstall:
	rm -f $(INSTALLDIR)/$(BINARY)
	@echo "Removed: $(INSTALLDIR)/$(BINARY)"

## clean: remove build artefacts
clean:
	rm -rf $(BINDIR)
	rm -f $(BINARY) gosorter-* *.tar.gz *.zip coverage.out
	@echo "Cleaned."

## Cross-compile release binaries
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-linux-amd64 .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-linux-arm64 .

build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-macos-intel .

build-macos-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-macos-apple-silicon .

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-windows-amd64.exe .

release-all: clean
	$(MAKE) build-linux build-linux-arm64 build-macos build-macos-arm64 build-windows
	tar -czf gosorter-linux-amd64.tar.gz gosorter-linux-amd64 README.md LICENSE
	tar -czf gosorter-linux-arm64.tar.gz gosorter-linux-arm64 README.md LICENSE
	tar -czf gosorter-macos-intel.tar.gz gosorter-macos-intel README.md LICENSE
	tar -czf gosorter-macos-apple-silicon.tar.gz gosorter-macos-apple-silicon README.md LICENSE
	zip gosorter-windows-amd64.zip gosorter-windows-amd64.exe README.md LICENSE
	@echo "All platforms built and archived."

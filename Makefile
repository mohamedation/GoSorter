# parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gosorter
BINARY_UNIX=$(BINARY_NAME)_unix
VERSION?=$(shell git describe --tags --always --dirty)

.PHONY: all build clean test coverage deps fmt vet lint install uninstall build-local release-all

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f gosorter-*
	rm -f *.tar.gz *.zip

test:
	$(GOTEST) -v ./...

coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

deps:
	$(GOMOD) download
	$(GOMOD) verify

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

lint:
	golangci-lint run

# Build the binary for current platform (local install)
build-local: deps fmt vet
	$(GOBUILD) -o $(BINARY_NAME) -v .

# System install - copies binary to /opt
install: build-local
	@echo "Installing $(BINARY_NAME) to /opt/$(BINARY_NAME)..."
	sudo cp $(BINARY_NAME) /opt/$(BINARY_NAME)
	@echo "Creating symlink in /usr/local/bin..."
	sudo ln -sf /opt/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete! You can now use '$(BINARY_NAME)' from anywhere."

# Uninstall - removes binary from system
uninstall:
	@echo "Removing $(BINARY_NAME) from system..."
	sudo rm -f /opt/$(BINARY_NAME)
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete."

# Cross compilation targets
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-linux-amd64 .

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-linux-arm64 .

build-macos:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-macos-intel .

build-macos-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-macos-apple-silicon .

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o gosorter-windows-amd64.exe .

# Build all platforms
release-all: clean
	@echo "Building for all platforms..."
	$(MAKE) build-linux
	$(MAKE) build-linux-arm64
	$(MAKE) build-macos
	$(MAKE) build-macos-arm64
	$(MAKE) build-windows
	@echo "Creating archives..."
	tar -czf gosorter-linux-amd64.tar.gz gosorter-linux-amd64 README.md LICENSE
	tar -czf gosorter-linux-arm64.tar.gz gosorter-linux-arm64 README.md LICENSE
	tar -czf gosorter-macos-intel.tar.gz gosorter-macos-intel README.md LICENSE
	tar -czf gosorter-macos-apple-silicon.tar.gz gosorter-macos-apple-silicon README.md LICENSE
	zip gosorter-windows-amd64.zip gosorter-windows-amd64.exe README.md LICENSE
	@echo "All platforms built and archived!"

run: build
	./$(BINARY_NAME)

# Development helpers
dev: fmt vet lint test

# Release build with optimizations
release:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-w -s -X main.Version=$(VERSION)" -o $(BINARY_NAME) -v .

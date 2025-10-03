.PHONY: build install test clean help

# Build the binary
build:
	go build -o commitgen main.go

# Install to /usr/local/bin
install: build
	./commitgen install

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f commitgen

# Install dependencies
deps:
	go mod tidy

# Development build with race detection
dev:
	go build -race -o commitgen main.go

# Release build
release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o commitgen-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o commitgen-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o commitgen-darwin-arm64 main.go

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build    - Build the commitgen binary"
	@echo "  install  - Build and install commitgen to /usr/local/bin"
	@echo "  test     - Run all tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  deps     - Install dependencies"
	@echo "  dev      - Build with race detection enabled"
	@echo "  release  - Build release binaries for multiple platforms"
	@echo "  help     - Show this help message"
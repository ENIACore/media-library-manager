# Media Library Manager - Build Configuration
BINARY_NAME=mlm
VERSION?=dev
MAIN_PATH=./cmd/media_library_manager

# Build directories
BUILD_DIR=build
DIST_DIR=dist

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GO_BUILD=go build $(LDFLAGS)

# Detect current platform for local install
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

.PHONY: all build clean install uninstall release linux macos windows help

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	@$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Install binary to /usr/local/bin
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete! Run '$(BINARY_NAME)' to use the tool."

# Uninstall binary from /usr/local/bin
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstall complete."

# Build release binaries for all platforms
release: clean
	@echo "Building release binaries for version $(VERSION)..."
	@mkdir -p $(DIST_DIR)

	@echo "Building Linux amd64..."
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

	@echo "Building Linux arm64..."
	@GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

	@echo "Building macOS amd64..."
	@GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

	@echo "Building macOS arm64..."
	@GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

	@echo "Building Windows amd64..."
	@GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

	@echo "Release build complete! Binaries in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

# Platform-specific builds
linux:
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Linux builds complete!"

macos:
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	@GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "macOS builds complete!"

windows:
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	@GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Windows build complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "Clean complete."

# Display help
help:
	@echo "Media Library Manager - Build Targets"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build      - Build binary for current platform (default)"
	@echo "  install    - Build and install to /usr/local/bin (requires sudo)"
	@echo "  uninstall  - Remove binary from /usr/local/bin (requires sudo)"
	@echo "  release    - Build binaries for all platforms"
	@echo "  linux      - Build Linux binaries (amd64, arm64)"
	@echo "  macos      - Build macOS binaries (amd64, arm64)"
	@echo "  windows    - Build Windows binary (amd64)"
	@echo "  test       - Run all tests"
	@echo "  clean      - Remove build artifacts"
	@echo "  help       - Display this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build             # Build for your platform"
	@echo "  make install           # Install mlm command"
	@echo "  make release VERSION=1.0.0  # Build v1.0.0 release"

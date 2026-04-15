# Maintenance Dashboard — Makefile

BINARY_NAME  := maintenance-dashboard
CMD_PATH     := ./cmd/server
FRONTEND_DIR := web/frontend
DIST_DIR     := internal/infra/assets/dist
RELEASES_DIR := _releases

# Version from git tag, fallback to "dev"
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

# Go env
GOCMD   := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean

.PHONY: all build build-frontend \
        build-windows build-mac build-linux \
        release clean help

# ──────────────────────────────────────────────
# Default target
# ──────────────────────────────────────────────
all: build-frontend build

# ──────────────────────────────────────────────
# Frontend
# ──────────────────────────────────────────────
build-frontend:
	@echo "→ Building frontend..."
	cd $(FRONTEND_DIR) && npm install --silent && npm run build
	@echo "✓ Frontend built to $(DIST_DIR)"

# ──────────────────────────────────────────────
# Backend (current platform)
# ──────────────────────────────────────────────
build:
	@echo "→ Building Go binary for current platform..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)
	@echo "✓ Binary: $(BINARY_NAME)"

# ──────────────────────────────────────────────
# Cross-compilation targets
# ──────────────────────────────────────────────
build-windows:
	@echo "→ Building for Windows (amd64)..."
	@mkdir -p $(RELEASES_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(RELEASES_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "✓ $(RELEASES_DIR)/$(BINARY_NAME)-windows-amd64.exe"

build-mac:
	@echo "→ Building for macOS (amd64 + arm64)..."
	@mkdir -p $(RELEASES_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(RELEASES_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) \
		-o $(RELEASES_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "✓ $(RELEASES_DIR)/$(BINARY_NAME)-darwin-amd64"
	@echo "✓ $(RELEASES_DIR)/$(BINARY_NAME)-darwin-arm64"

build-linux:
	@echo "→ Building for Linux (amd64)..."
	@mkdir -p $(RELEASES_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
		-o $(RELEASES_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "✓ $(RELEASES_DIR)/$(BINARY_NAME)-linux-amd64"

# ──────────────────────────────────────────────
# Release — build frontend + all platforms
# ──────────────────────────────────────────────
release: build-frontend build-windows build-mac build-linux
	@echo ""
	@echo "✓ All release binaries in $(RELEASES_DIR)/"

# ──────────────────────────────────────────────
# Dev
# ──────────────────────────────────────────────
run:
	$(GOCMD) run $(CMD_PATH)

run-frontend:
	cd $(FRONTEND_DIR) && npm run dev

# ──────────────────────────────────────────────
# Clean
# ──────────────────────────────────────────────
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf $(RELEASES_DIR)/$(BINARY_NAME)-*
	@echo "✓ Cleaned"

# ──────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "  all              Build frontend + binary for current platform (default)"
	@echo "  build            Build Go binary for current platform"
	@echo "  build-frontend   Build React frontend into $(DIST_DIR)"
	@echo "  build-windows    Cross-compile for Windows amd64"
	@echo "  build-mac        Cross-compile for macOS amd64 + arm64"
	@echo "  build-linux      Cross-compile for Linux amd64"
	@echo "  release          Build frontend + all platform binaries"
	@echo "  run              Run the Go server (dev)"
	@echo "  run-frontend     Run the Vite dev server"
	@echo "  clean            Remove build artifacts"
	@echo "  help             Show this message"

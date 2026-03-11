# SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
# SPDX-License-Identifier: MIT

.PHONY: all build clean test run server client install help web-install web-dev web-build web-start docs docs-dev docs-build docs-clean docs-open docs-new web-bundle web-serve fmt lint vendor deps

# Default target
all: build

# Build the binary
build:
	go build -o cheetah .

# Build with nix (reproducible)
nix-build:
	nix build

# Clean build artifacts
clean:
	rm -f cheetah
	rm -rf result

# Run tests
test:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run web/Playwright tests
test-web:
	cd web && npm run test

# Run all tests (Go + Web)
test-all: test test-web

# Run in combined mode (default)
run: build
	./cheetah

# Run server only
server: build
	./cheetah -server

# Run client only (connect to existing server)
client: build
	./cheetah -client

# Install to GOPATH/bin
install:
	go install .

# Vendor dependencies
vendor:
	go mod vendor

# Update dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Start backend in background
start-backend: build
	./scripts/start-backend.sh

# Stop backend
stop-backend:
	./scripts/stop-backend.sh

# Check backend status
status:
	./scripts/status-backend.sh

# Launch frontend against running backend
frontend: build
	./scripts/launch-frontend.sh

# Web frontend targets
web-install:
	cd web && npm install

web-dev: web-install
	cd web && npm run dev

web-build: web-install
	cd web && npm run build

web-start: build
	@echo "Starting backend in background..."
	./cheetah -server &
	@sleep 2
	@echo "Starting web frontend..."
	cd web && npm run dev

# Documentation (Hugo)
docs-dev:
	cd hugo && hugo server -D --bind 0.0.0.0

docs-build:
	cd hugo && hugo --minify

docs: docs-build
	@echo "Documentation built in hugo/public/"

docs-clean:
	rm -rf hugo/public hugo/resources web/dist/docs

docs-open:
	xdg-open http://localhost:1313 2>/dev/null || open http://localhost:1313 2>/dev/null || echo "Open http://localhost:1313 in your browser"

docs-new:
	@read -p "Enter page path (e.g., posts/my-new-post): " path; \
	cd hugo && hugo new "$$path.md"

# Combined production build (React + Hugo bundled together)
web-bundle: web-build docs-build
	@echo "Bundling Hugo docs into web/dist/docs/..."
	rm -rf web/dist/docs
	cp -r hugo/public web/dist/docs
	@echo "Production bundle complete in web/dist/"

# Run production server with bundled docs
web-serve: web-bundle build
	./cheetah web -port 8787 -dir web/dist

# Help
help:
	@echo "Cheetah - RSVP Speed Reading Application"
	@echo ""
	@echo "Build targets:"
	@echo "  make build       - Build the binary"
	@echo "  make nix-build   - Build with nix (reproducible)"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make install     - Install to GOPATH/bin"
	@echo ""
	@echo "Run targets:"
	@echo "  make run         - Run in combined mode"
	@echo "  make server      - Run backend only"
	@echo "  make client      - Run frontend only"
	@echo ""
	@echo "Backend management:"
	@echo "  make start-backend - Start backend in background"
	@echo "  make stop-backend  - Stop backend"
	@echo "  make status        - Check backend status"
	@echo "  make frontend      - Launch frontend client"
	@echo ""
	@echo "Web frontend:"
	@echo "  make web-install - Install web dependencies"
	@echo "  make web-dev     - Start web dev server"
	@echo "  make web-build   - Build web for production"
	@echo "  make web-start   - Start backend + web frontend"
	@echo "  make web-bundle  - Build React + Hugo bundled together"
	@echo "  make web-serve   - Run production server with docs"
	@echo ""
	@echo "Development:"
	@echo "  make test        - Run tests"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Lint code"
	@echo "  make vendor      - Vendor dependencies"
	@echo "  make deps        - Update dependencies"
	@echo ""
	@echo "Documentation (Hugo):"
	@echo "  make docs-dev    - Start Hugo dev server"
	@echo "  make docs-build  - Build documentation"
	@echo "  make docs        - Build documentation (alias)"
	@echo "  make docs-clean  - Remove built documentation"
	@echo "  make docs-open   - Open docs in browser"
	@echo "  make docs-new    - Create new documentation page"

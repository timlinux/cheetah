#!/usr/bin/env bash
#
# Launch Cheetah TUI frontend (connecting to running backend)
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/cheetah"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo "Binary not found. Building..."
    cd "$PROJECT_DIR"
    make build
fi

# Check if backend is running
if ! "$SCRIPT_DIR/status-backend.sh" > /dev/null 2>&1; then
    echo "Backend not running. Starting it first..."
    "$SCRIPT_DIR/start-backend.sh"
    sleep 1
fi

# Launch frontend in client mode
exec "$BINARY" -client "$@"

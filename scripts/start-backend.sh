#!/usr/bin/env bash
#
# Start Cheetah backend server in background
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

# Check if already running
PID_FILE="${XDG_RUNTIME_DIR:-/tmp}/cheetah.pid"
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if kill -0 "$PID" 2>/dev/null; then
        echo "Cheetah backend already running (PID: $PID)"
        exit 1
    else
        rm -f "$PID_FILE"
    fi
fi

# Start backend in background
echo "Starting Cheetah backend..."
"$BINARY" -server &

sleep 1

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    echo "Cheetah backend started (PID: $PID)"
else
    echo "Warning: Backend may have started but PID file not found"
fi

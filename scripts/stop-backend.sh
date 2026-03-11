#!/usr/bin/env bash
#
# Stop Cheetah backend server
#

set -e

PID_FILE="${XDG_RUNTIME_DIR:-/tmp}/cheetah.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "No PID file found. Backend may not be running."
    exit 0
fi

PID=$(cat "$PID_FILE")

if kill -0 "$PID" 2>/dev/null; then
    echo "Stopping Cheetah backend (PID: $PID)..."
    kill "$PID"
    rm -f "$PID_FILE"
    echo "Backend stopped."
else
    echo "Process $PID not running. Cleaning up PID file."
    rm -f "$PID_FILE"
fi

#!/usr/bin/env bash
#
# Check Cheetah backend status
#

PID_FILE="${XDG_RUNTIME_DIR:-/tmp}/cheetah.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "Cheetah backend: NOT RUNNING (no PID file)"
    exit 1
fi

PID=$(cat "$PID_FILE")

if kill -0 "$PID" 2>/dev/null; then
    echo "Cheetah backend: RUNNING (PID: $PID)"

    # Check health endpoint
    if curl -s http://127.0.0.1:8787/api/health > /dev/null 2>&1; then
        echo "Health check: OK"
    else
        echo "Health check: FAILED (server may be starting up)"
    fi

    exit 0
else
    echo "Cheetah backend: NOT RUNNING (stale PID file)"
    rm -f "$PID_FILE"
    exit 1
fi

#!/usr/bin/env bash
# design-review-portal.sh — serves docs/design/ for the MAPLE design review portal
#
# Usage:
#   design-review-portal.sh start <port>   start server on 127.0.0.1:<port>
#   design-review-portal.sh stop           kill the running server
#   design-review-portal.sh open           open browser to running portal
#   design-review-portal.sh status         print URL if running, exit 1 if not

set -uo pipefail

ACTION="${1:-start}"
PORT="${2:-}"
PID_FILE=".claude/state/design-portal.pid"
URL_FILE=".claude/state/design-portal.url"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PORTAL_PY="$SCRIPT_DIR/design-review-portal.py"
TOKEN_FILE=".claude/state/design-portal.token"

_stop_existing() {
  if [[ -f "$PID_FILE" ]]; then
    EXISTING_PID=$(cat "$PID_FILE" 2>/dev/null || echo "")
    if [[ -n "$EXISTING_PID" ]] && kill -0 "$EXISTING_PID" 2>/dev/null; then
      kill "$EXISTING_PID" 2>/dev/null || true
      sleep 0.2
    fi
    rm -f "$PID_FILE" "$URL_FILE"
  fi
}

case "$ACTION" in
  start)
    if [[ -z "$PORT" ]]; then
      echo "usage: design-review-portal.sh start <port>" >&2
      exit 1
    fi

    # Idempotent: return existing URL if already running on the same port
    if [[ -f "$PID_FILE" ]] && [[ -f "$URL_FILE" ]]; then
      EXISTING_PID=$(cat "$PID_FILE" 2>/dev/null || echo "")
      EXISTING_URL=$(cat "$URL_FILE" 2>/dev/null || echo "")
      if [[ -n "$EXISTING_PID" ]] && kill -0 "$EXISTING_PID" 2>/dev/null; then
        if echo "$EXISTING_URL" | grep -q ":${PORT}"; then
          echo "$EXISTING_URL"
          exit 0
        fi
      fi
    fi

    _stop_existing
    mkdir -p .claude/state

    URL="http://127.0.0.1:${PORT}"

    if [[ ! -f "$PORTAL_PY" ]]; then
      echo "design-review-portal.py not found at $PORTAL_PY" >&2
      exit 1
    fi

    if ! command -v python3 &>/dev/null; then
      echo "python3 is required to run the design review portal" >&2
      exit 1
    fi

    python3 "$PORTAL_PY" --root "$(pwd)" --port "$PORT" --token-file "$TOKEN_FILE" >/dev/null 2>&1 &
    SERVER_PID=$!

    echo "$SERVER_PID" > "$PID_FILE"
    echo "$URL" > "$URL_FILE"

    # Brief readiness wait (max 1s)
    for i in 1 2 3 4 5; do
      sleep 0.2
      if kill -0 "$SERVER_PID" 2>/dev/null; then
        break
      fi
    done

    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      rm -f "$PID_FILE" "$URL_FILE"
      echo "design portal failed to start" >&2
      exit 1
    fi

    echo "$URL"
    ;;

  stop)
    _stop_existing
    ;;

  open)
    if [[ ! -f "$URL_FILE" ]]; then
      echo "design portal is not running" >&2
      exit 1
    fi
    URL=$(cat "$URL_FILE")
    case "$(uname -s)" in
      Darwin) open "$URL" ;;
      Linux)  xdg-open "$URL" 2>/dev/null || sensible-browser "$URL" 2>/dev/null || echo "$URL" ;;
      *)      echo "$URL" ;;
    esac
    echo "$URL"
    ;;

  status)
    if [[ -f "$PID_FILE" ]] && [[ -f "$URL_FILE" ]]; then
      EXISTING_PID=$(cat "$PID_FILE" 2>/dev/null || echo "")
      if [[ -n "$EXISTING_PID" ]] && kill -0 "$EXISTING_PID" 2>/dev/null; then
        cat "$URL_FILE"
        exit 0
      fi
    fi
    exit 1
    ;;

  *)
    echo "usage: design-review-portal.sh {start <port>|stop|open|status}" >&2
    exit 1
    ;;
esac

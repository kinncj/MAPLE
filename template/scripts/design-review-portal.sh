#!/usr/bin/env bash
set -euo pipefail

ACTION="${1:-start}"
PORT="${MAPLE_DESIGN_PORT:-4173}"

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATE_DIR="$ROOT/.claude/state"
PID_FILE="$STATE_DIR/design-review-portal.pid"
LOG_FILE="$STATE_DIR/design-review-portal.log"
TOKEN_FILE="$STATE_DIR/design-review-portal.token"
SERVER_SCRIPT="$ROOT/scripts/design-review-portal.py"
if [ ! -f "$SERVER_SCRIPT" ]; then
  SERVER_SCRIPT="$SCRIPT_DIR/design-review-portal.py"
fi
URL="http://127.0.0.1:$PORT"

mkdir -p "$STATE_DIR"

ensure_token() {
  if [ ! -s "$TOKEN_FILE" ]; then
    python3 - <<'PY' > "$TOKEN_FILE"
import secrets
print(secrets.token_hex(24))
PY
    chmod 600 "$TOKEN_FILE" || true
  fi
}

is_running() {
  if [ ! -f "$PID_FILE" ]; then
    return 1
  fi
  local pid
  pid="$(cat "$PID_FILE" 2>/dev/null || true)"
  [ -z "$pid" ] && return 1
  kill -0 "$pid" 2>/dev/null
}

wait_healthy() {
  local i
  for i in $(seq 1 30); do
    if curl -sf "$URL/health" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done
  return 1
}

open_browser() {
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$URL" >/dev/null 2>&1 || true
    return 0
  fi
  if command -v open >/dev/null 2>&1; then
    open "$URL" >/dev/null 2>&1 || true
    return 0
  fi
  if command -v cmd.exe >/dev/null 2>&1; then
    cmd.exe /c start "$URL" >/dev/null 2>&1 || true
    return 0
  fi
  return 0
}

start_server() {
  if [ ! -f "$SERVER_SCRIPT" ]; then
    echo "design-review portal script missing: $SERVER_SCRIPT" >&2
    exit 1
  fi
  ensure_token
  if is_running; then
    return 0
  fi

  if [ -f "$PID_FILE" ]; then
    rm -f "$PID_FILE"
  fi

  nohup python3 "$SERVER_SCRIPT" \
    --root "$ROOT" \
    --port "$PORT" \
    --token-file "$TOKEN_FILE" \
    >>"$LOG_FILE" 2>&1 &
  local pid="$!"
  echo "$pid" > "$PID_FILE"

  if ! wait_healthy; then
    echo "design-review portal failed to start (see $LOG_FILE)" >&2
    exit 1
  fi
}

stop_server() {
  if ! is_running; then
    rm -f "$PID_FILE"
    return 0
  fi
  local pid
  pid="$(cat "$PID_FILE")"
  kill "$pid" >/dev/null 2>&1 || true
  rm -f "$PID_FILE"
}

status_server() {
  if is_running; then
    local pid
    pid="$(cat "$PID_FILE")"
    echo "design-review portal running pid=$pid url=$URL"
  else
    echo "design-review portal not running"
  fi
}

case "$ACTION" in
  start)
    start_server
    ;;
  open)
    start_server
    open_browser
    ;;
  stop)
    stop_server
    ;;
  status)
    status_server
    ;;
  *)
    echo "usage: $0 {start|open|stop|status}" >&2
    exit 2
    ;;
esac

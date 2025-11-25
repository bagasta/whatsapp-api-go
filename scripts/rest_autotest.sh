#!/usr/bin/env bash

# REST & API-OLD smoke tester for WhatsApp API Go
# Usage: BASE_URL=http://localhost:3000 USER_ID=u1 AGENT_ID=bot1 API_KEY=token PHONE=628xxxx ./scripts/rest_autotest.sh

set -euo pipefail

BASE_URL=${BASE_URL:-http://localhost:3000}
USER_ID=${USER_ID:-test_user_001}
AGENT_ID=${AGENT_ID:-test_bot_001}
AGENT_NAME=${AGENT_NAME:-"Smoke Bot"}
API_KEY=${API_KEY:-""}
PHONE=${PHONE:-""}
MESSAGE=${MESSAGE:-"Hello from rest_autotest.sh"}
QR_FILE=${QR_FILE:-qr_autotest.png}

AUTH_HEADER=()
if [[ -n "$API_KEY" ]]; then
  AUTH_HEADER=(-H "Authorization: Bearer $API_KEY" -H "X-Api-Key: $API_KEY")
fi

log() { printf "\n\033[1;34m==> %s\033[0m\n" "$*"; }
ok()  { printf "\033[0;32m✓ %s\033[0m\n" "$*"; }
fail(){ printf "\033[0;31m✗ %s\033[0m\n" "$*"; }

call() {
  local name=$1 method=$2 path=$3 body=${4:-} extra=("${@:5}")
  log "$name ($method $path)"
  local tmp resp status
  tmp=$(mktemp)
  if [[ -n "$body" ]]; then
    resp=$(curl -sS -w "\n%{http_code}" -o "$tmp" -X "$method" "${extra[@]}" "${AUTH_HEADER[@]}" -H "Content-Type: application/json" -d "$body" "$BASE_URL$path") || true
  else
    resp=$(curl -sS -w "\n%{http_code}" -o "$tmp" -X "$method" "${extra[@]}" "${AUTH_HEADER[@]}" "$BASE_URL$path") || true
  fi
  status=${resp##*$'\n'}
  cat "$tmp" | jq . 2>/dev/null || cat "$tmp"
  rm -f "$tmp"
  if [[ "$status" =~ ^2 ]]; then ok "$name (HTTP $status)"; else fail "$name (HTTP $status)"; fi
}

log "Target BASE_URL=$BASE_URL | USER_ID=$USER_ID | AGENT_ID=$AGENT_ID | PHONE=${PHONE:-<empty>} | API_KEY ${API_KEY:+set}"

call "Health" GET /health
call "Metrics" GET /metrics
call "App Devices" GET /app/devices

# Session lifecycle (API-OLD compat)
SESSION_PAYLOAD=$(jq -nc --arg uid "$USER_ID" --arg aid "$AGENT_ID" --arg an "$AGENT_NAME" --arg ap "$API_KEY" '{userId:$uid,agentId:$aid,agentName:$an,apikey:$ap}')
call "Create Session" POST /sessions "$SESSION_PAYLOAD"
sleep 1
call "Get Session" GET /sessions/"$AGENT_ID"

# QR
log "Request QR (saved to $QR_FILE if present)"
qr_resp=$(curl -sS -X POST "$BASE_URL/sessions/$AGENT_ID/qr") || true
echo "$qr_resp" | jq . 2>/dev/null || echo "$qr_resp"
if echo "$qr_resp" | jq -e '.qr.base64' >/dev/null 2>&1; then
  echo "$qr_resp" | jq -r '.qr.base64' | base64 -d > "$QR_FILE" 2>/dev/null || true
  ok "QR saved to $QR_FILE (scan with WhatsApp)"
else
  fail "No QR in response (session might already be logged in)"
fi

# REST send (modern)
if [[ -n "$PHONE" ]]; then
  REST_SEND=$(jq -nc --arg to "$PHONE" --arg msg "$MESSAGE" '{phone:$to,message:$msg}')
  call "Send Text (REST /send/message)" POST /send/message "$REST_SEND"
fi

# Agent endpoints (API-OLD compat)
if [[ -n "$PHONE" ]]; then
  AGENT_MSG=$(jq -nc --arg to "$PHONE" --arg msg "$MESSAGE" '{to:$to,message:$msg}')
  call "Agent Send Message" POST /agents/"$AGENT_ID"/messages "$AGENT_MSG"

  AGENT_RUN=$(jq -nc --arg in "$MESSAGE" --arg sid "$PHONE" '{input:$in,session_id:$sid,parameters:{max_steps:3}}')
  call "Agent Run (AI)" POST /agents/"$AGENT_ID"/run "$AGENT_RUN"
else
  log "Skip send/run because PHONE is empty"
fi

ok "Done. If QR was saved, scan it to authenticate before re-running sends."

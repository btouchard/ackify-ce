#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-or-later
# Purpose: Local Docker smoke test with clear PASS/FAIL output
set -uo pipefail

COMPOSE_FILE=${COMPOSE_FILE:-compose.local.yml}
BASE_URL=${BASE_URL:-http://localhost:8080}
DOC_ID=${DOC_ID:-demo}
USER1_EMAIL=${USER1_EMAIL:-user1@example.com}
USER2_EMAIL=${USER2_EMAIL:-user2@example.com}
KEEP_UP=${KEEP_UP:-0}
BUILD_IMAGES=${BUILD_IMAGES:-0}

usage() {
  cat <<EOF
Usage: $0 [options]

Options (CLI flags take precedence over env vars):
  --build                 Build images before starting (same as BUILD_IMAGES=1)
  --keep-up               Keep the stack running after tests (same as KEEP_UP=1)
  --compose-file <file>   Docker Compose file (default: ${COMPOSE_FILE})
  --base-url <url>        Base URL for service (default: ${BASE_URL})
  --doc-id <id>           Document ID to seed/test (default: ${DOC_ID})
  --user1 <email>         First user email (default: ${USER1_EMAIL})
  --user2 <email>         Second user email (default: ${USER2_EMAIL})
  -h, --help              Show this help

Environment overrides (if flags not set):
  COMPOSE_FILE, BASE_URL, DOC_ID, USER1_EMAIL, USER2_EMAIL, KEEP_UP, BUILD_IMAGES
EOF
}

# Parse CLI flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --build) BUILD_IMAGES=1; shift ;;
    --keep-up) KEEP_UP=1; shift ;;
    --compose-file) COMPOSE_FILE="${2:-}"; shift 2 ;;
    --base-url) BASE_URL="${2:-}"; shift 2 ;;
    --doc-id) DOC_ID="${2:-}"; shift 2 ;;
    --user1) USER1_EMAIL="${2:-}"; shift 2 ;;
    --user2) USER2_EMAIL="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

printf "\n=== Ackify-CE Docker Smoke Test ===\n"

if ! command -v docker >/dev/null 2>&1; then
  echo "[!] docker not found in PATH" >&2; exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "[!] docker compose plugin not found" >&2; exit 1
fi

if [[ ! -f ".env" ]]; then
  echo "[!] .env not found. Copy .env.example to .env and configure variables." >&2
  exit 1
fi

echo "[i] Loading .env"
set -a; source ./.env; set +a

printf "\n[1/6] Bringing up stack: %s (build=%s)\n" "${COMPOSE_FILE}" "${BUILD_IMAGES}"
COMPOSE_UP_OPTS=("-d")
if [[ "${BUILD_IMAGES}" == "1" ]]; then COMPOSE_UP_OPTS=("-d" "--build"); fi
docker compose -f "${COMPOSE_FILE}" up "${COMPOSE_UP_OPTS[@]}"

printf "[2/6] Waiting for health endpoint: %s/health\n" "${BASE_URL}"
for i in {1..60}; do
  if curl -fsS "${BASE_URL}/health" >/dev/null; then
    echo "  PASS health"; break
  fi
  sleep 1
  if [[ $i -eq 60 ]]; then echo "  FAIL health (timeout)" >&2; exit 1; fi
done

printf "\n[3/6] Seeding demo signatures into PostgreSQL (%s)\n" "${DOC_ID}"
PGPASSWORD="${POSTGRES_PASSWORD}" docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" ackify-db \
  psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -v ON_ERROR_STOP=1 -c \
  "INSERT INTO signatures (doc_id,user_sub,user_email,user_name,signed_at,payload_hash,signature,nonce,referer) VALUES 
   ('${DOC_ID}','user1','${USER1_EMAIL}','User One',now(),'ph1','sig1','n1','seed') 
   ON CONFLICT DO NOTHING;"

PGPASSWORD="${POSTGRES_PASSWORD}" docker exec -e PGPASSWORD="${POSTGRES_PASSWORD}" ackify-db \
  psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -v ON_ERROR_STOP=1 -c \
  "INSERT INTO signatures (doc_id,user_sub,user_email,user_name,signed_at,payload_hash,signature,nonce,referer) VALUES 
   ('${DOC_ID}','user2','${USER2_EMAIL}','User Two',now(),'ph2','sig2','n2','seed') 
   ON CONFLICT DO NOTHING;"

printf "\n[4/6] Endpoint checks\n"

has_cmd() { command -v "$1" >/dev/null 2>&1; }
PASS_CNT=0; FAIL_CNT=0

pass() { echo "  PASS $1"; PASS_CNT=$((PASS_CNT+1)); }
fail() { echo "  FAIL $1"; FAIL_CNT=$((FAIL_CNT+1)); }

# 4.1 JSON status for document
if out=$(curl -fsS "${BASE_URL}/status?doc=${DOC_ID}" 2>/dev/null); then
  echo "$out" | grep -q '"doc_id"\s*:\s*"' && pass "status?doc=${DOC_ID}" || fail "status content"
else
  fail "status request"
fi

# 4.2 Badge PNG
if curl -fsS "${BASE_URL}/status.png?doc=${DOC_ID}&user=${USER1_EMAIL}" -o /tmp/ackify_badge.png; then
  if has_cmd file; then
    file /tmp/ackify_badge.png | grep -qi 'PNG image' && pass "status.png badge" || fail "badge type"
  else
    [[ -s /tmp/ackify_badge.png ]] && pass "status.png badge (size>0)" || fail "badge size"
  fi
else
  fail "status.png fetch"
fi

# 4.3 oEmbed
if out=$(curl -fsS "${BASE_URL}/oembed?url=${BASE_URL}/embed?doc=${DOC_ID}" 2>/dev/null); then
  if has_cmd jq; then
    echo "$out" | jq -e -r '.html' >/dev/null 2>&1 && pass "oembed html" || fail "oembed html field"
  else
    echo "$out" | grep -qi '<html' && pass "oembed raw html" || fail "oembed raw"
  fi
else
  fail "oembed request"
fi

# 4.4 Embed view
curl -fsS "${BASE_URL}/embed?doc=${DOC_ID}" >/dev/null && pass "embed view" || fail "embed view"

# 4.5 Security headers (root)
if hdr=$(curl -fsS -D - -o /dev/null "${BASE_URL}/" 2>/dev/null); then
  echo "$hdr" | grep -qi 'Content-Security-Policy' && echo "$hdr" | grep -qi 'X-Frame-Options' \
    && echo "$hdr" | grep -qi 'Referrer-Policy' && echo "$hdr" | grep -qi 'X-Content-Type-Options' \
    && pass "security headers" || fail "security headers"
else
  fail "security headers request"
fi

# 4.6 Admin redirects to login when unauthenticated (GET to avoid HEAD 405)
if hdr=$(curl -fsS -D - -o /dev/null "${BASE_URL}/admin" 2>/dev/null); then
  echo "$hdr" | grep -qi '^location: /login' && pass "admin redirect" || fail "admin redirect"
else
  fail "admin head"
fi

printf "\n[5/6] Recent app logs (tail 80)\n"
docker logs --tail 80 ackify-ce || true

printf "\n[6/6] Summary: %s passed, %s failed\n" "${PASS_CNT}" "${FAIL_CNT}"
if [[ "${KEEP_UP}" == "0" ]]; then
  echo "[i] Bringing stack down (set KEEP_UP=1 to keep running)"
  docker compose -f "${COMPOSE_FILE}" down -v
else
  echo "[i] Stack left running as requested (KEEP_UP=1)"
fi

if [[ ${FAIL_CNT} -gt 0 ]]; then
  echo "[!] Smoke test completed with failures"; exit 1
else
  echo "[i] Docker smoke test complete (all good)"; exit 0
fi

#!/usr/bin/env bash
# 直接激活 Go 二进制版本，不使用 Docker、Compose、Node 或镜像仓库。
set -Eeuo pipefail

SHA="${1:-}"
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || { echo "STATUS=FAILED usage: activate-direct-release.sh <exact-40-character-git-sha>" >&2; exit 2; }

DIRECT_ROOT="${DIRECT_ROOT:-/opt/qiantie/v88/direct}"
RELEASE_DIR="$DIRECT_ROOT/releases/$SHA"
SERVICE_NAME="${QIANTIE_SERVICE_NAME:-qiantie.service}"
SERVICE_FILE="${QIANTIE_SERVICE_FILE:-/etc/systemd/system/$SERVICE_NAME}"
SYSTEMCTL="${SYSTEMCTL_BIN:-systemctl}"
CURL="${CURL_BIN:-curl}"
HEALTH_URL="${QIANTIE_HEALTH_URL:-http://127.0.0.1:3000/health}"
BUILD_INFO_URL="${QIANTIE_BUILD_INFO_URL:-http://127.0.0.1:3000/api/runtime-build-info}"
SERVICE_TEMPLATE="${QIANTIE_SERVICE_TEMPLATE:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/qiantie.service}"

[ -x "$RELEASE_DIR/go/qiantie" ] || { echo "STATUS=FAILED missing Go binary: $RELEASE_DIR/go/qiantie" >&2; exit 3; }
[ -s "$SERVICE_TEMPLATE" ] || { echo "STATUS=FAILED missing service template: $SERVICE_TEMPLATE" >&2; exit 4; }
mkdir -p "$DIRECT_ROOT/releases"

PREV_TARGET="$(readlink "$DIRECT_ROOT/current" 2>/dev/null || true)"
if [ -n "$PREV_TARGET" ]; then
  case "$PREV_TARGET" in
    /*) PREV_DIR="$PREV_TARGET" ;;
    *) PREV_DIR="$DIRECT_ROOT/$PREV_TARGET" ;;
  esac
else
  PREV_DIR=""
fi

MANIFEST_SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/runtime-release-manifest.sh"
[ -f "$MANIFEST_SCRIPT" ] || { echo "STATUS=FAILED missing runtime manifest generator" >&2; exit 5; }
PREV_SHA=""
if [ -n "$PREV_DIR" ] && [ -f "$PREV_DIR/runtime-release.json" ]; then
  PREV_SHA="$(python3 - "$PREV_DIR/runtime-release.json" <<'PY'
import json, re, sys
try:
    value = json.load(open(sys.argv[1], encoding='utf-8')).get('gitSha', '')
    if not re.fullmatch(r'[0-9a-f]{40}', value): raise ValueError
    print(value)
except Exception as exc:
    print(f'previous release identity invalid: {exc}', file=sys.stderr)
    sys.exit(1)
PY
)" || exit 6
fi

bash "$MANIFEST_SCRIPT" "$SHA" "$RELEASE_DIR" "$PREV_SHA"
install -D -m 0644 "$SERVICE_TEMPLATE" "$SERVICE_FILE"

wait_for_service() {
  for _ in $(seq 1 45); do
    if "$CURL" -fsS --max-time 3 "$HEALTH_URL" >/dev/null 2>&1; then return 0; fi
    sleep 2
  done
  echo "STATUS=FAILED Go health check did not become ready" >&2
  return 1
}

rollback() {
  local status=$?
  echo "STATUS=ROLLBACK restoring previous Go binary (exit=$status)" >&2
  if [ -n "$PREV_DIR" ] && [ -d "$PREV_DIR" ]; then
    ln -sfn "$PREV_DIR" "$DIRECT_ROOT/current"
    "$SYSTEMCTL" daemon-reload >/dev/null 2>&1 || true
    "$SYSTEMCTL" restart "$SERVICE_NAME" >/dev/null 2>&1 || true
  else
    rm -f "$DIRECT_ROOT/current"
    "$SYSTEMCTL" stop "$SERVICE_NAME" >/dev/null 2>&1 || true
  fi
  exit "$status"
}
trap rollback ERR

ln -sfn "$RELEASE_DIR" "$DIRECT_ROOT/current"
"$SYSTEMCTL" daemon-reload
"$SYSTEMCTL" enable "$SERVICE_NAME" >/dev/null 2>&1 || true
"$SYSTEMCTL" restart "$SERVICE_NAME"
wait_for_service

GOT_SHA="$("$CURL" -fsS --max-time 5 "$BUILD_INFO_URL" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("git_sha", ""))')"
if [ "$GOT_SHA" != "$SHA" ]; then
  echo "STATUS=FAILED runtime SHA mismatch: $GOT_SHA != $SHA" >&2
  exit 7
fi

if [ -n "$PREV_DIR" ] && [ -d "$PREV_DIR" ]; then ln -sfn "$PREV_DIR" "$DIRECT_ROOT/previous"; fi
trap - ERR
echo "STATUS=DEPLOYED_DIRECT release_sha=$SHA previous=${PREV_SHA:-none}"

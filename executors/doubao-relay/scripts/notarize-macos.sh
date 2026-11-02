#!/usr/bin/env bash
# macOS 公证(spec 13「代码签名 / macOS」):notarytool 提交 → 等结果 → stapler 装订票据。
# 必须在 scripts/sign-macos.sh 之后跑(公证要求 Developer ID 签名 + hardened runtime)。
#
# 用法:
#   scripts/notarize-macos.sh <path/to/doubao-relay_<ver>_universal.dmg>
#
# 凭据(二选一):
#   Apple ID:   APPLE_ID + APPLE_PASSWORD(app 专用密码)+ APPLE_TEAM_ID
#   API Key:    APPLE_API_KEY(key id)+ APPLE_API_ISSUER(issuer id)+ APPLE_API_KEY_PATH(.p8 路径)
#   (变量名与 Tauri 官方签名文档一致,CI secrets 可以直接复用)
#   DRY_RUN=1   只打印命令(密码打码),不执行
#
# 退出码:0 公证通过并已装订;2 参数 / 凭据缺失;1 公证被拒(会打印 notarytool log)。
set -euo pipefail

DRY_RUN="${DRY_RUN:-0}"
SECRETS=()

usage() {
  echo "用法: $0 <dmg>" >&2
  exit 2
}

run() {
  local shown=() a s masked
  for a in "$@"; do
    masked="$a"
    for s in ${SECRETS[@]+"${SECRETS[@]}"}; do
      if [ -n "$s" ] && [ "$a" = "$s" ]; then masked="***"; fi
    done
    shown+=("$masked")
  done
  printf '+'
  printf ' %q' "${shown[@]}"
  printf '\n'
  if [ "$DRY_RUN" != "1" ]; then
    "$@"
  fi
}

[ $# -eq 1 ] || usage
DMG="$1"
[ -f "$DMG" ] || {
  echo "notarize-macos: 找不到 $DMG" >&2
  exit 2
}

AUTH=()
if [ -n "${APPLE_API_KEY:-}" ] && [ -n "${APPLE_API_ISSUER:-}" ] && [ -n "${APPLE_API_KEY_PATH:-}" ]; then
  AUTH=(--key "$APPLE_API_KEY_PATH" --key-id "$APPLE_API_KEY" --issuer "$APPLE_API_ISSUER")
elif [ -n "${APPLE_ID:-}" ] && [ -n "${APPLE_PASSWORD:-}" ] && [ -n "${APPLE_TEAM_ID:-}" ]; then
  SECRETS=("$APPLE_PASSWORD")
  AUTH=(--apple-id "$APPLE_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_PASSWORD")
else
  echo "notarize-macos: 缺少公证凭据(APPLE_ID/APPLE_PASSWORD/APPLE_TEAM_ID 或 APPLE_API_KEY/APPLE_API_ISSUER/APPLE_API_KEY_PATH)" >&2
  exit 2
fi

if [ "$DRY_RUN" = "1" ]; then
  run xcrun notarytool submit "$DMG" "${AUTH[@]}" --wait --output-format json
  run xcrun stapler staple "$DMG"
  run xcrun stapler validate "$DMG"
  exit 0
fi

# notarytool 被拒(Invalid)/ 网络错误时退出非 0;不能让 set -e 在这里直接退出,
# 要落到下面的 STATUS 判断里打印 notarytool log。
OUT="$(run xcrun notarytool submit "$DMG" "${AUTH[@]}" --wait --output-format json | tee /dev/stderr | tail -n 1)" || true
STATUS="$(printf '%s' "$OUT" | /usr/bin/python3 -c 'import json,sys; print(json.load(sys.stdin).get("status",""))' 2>/dev/null || true)"
SUBMISSION_ID="$(printf '%s' "$OUT" | /usr/bin/python3 -c 'import json,sys; print(json.load(sys.stdin).get("id",""))' 2>/dev/null || true)"

if [ "$STATUS" != "Accepted" ]; then
  echo "notarize-macos: 公证未通过(status=${STATUS:-unknown})" >&2
  if [ -n "$SUBMISSION_ID" ]; then
    run xcrun notarytool log "$SUBMISSION_ID" "${AUTH[@]}" >&2 || true
  fi
  exit 1
fi

run xcrun stapler staple "$DMG"
run xcrun stapler validate "$DMG"
echo "notarize-macos: 公证通过并已装订 $DMG"

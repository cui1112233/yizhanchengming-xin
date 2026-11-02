#!/usr/bin/env bash
# macOS 代码签名(spec 13「代码签名 / macOS」):签 .app → 用签好的 .app 重建 DMG → 签 DMG。
# 公证见 scripts/notarize-macos.sh(必须在本脚本之后跑)。
#
# 用法:
#   scripts/sign-macos.sh <bundle-dir>
#     bundle-dir = cargo tauri build 的 bundle 目录,例如
#       target/universal-apple-darwin/release/bundle   (发布构建)
#       target/debug/bundle                            (cargo tauri build --debug)
#     里面要有 macos/*.app;dmg/*.dmg 存在时沿用其文件名,不存在时生成 doubao-relay_<ver>.dmg。
#
# 环境变量:
#   APPLE_SIGNING_IDENTITY       必填,"Developer ID Application: <Name> (<TEAMID>)"
#   APPLE_CERTIFICATE            可选,base64 编码的 .p12(CI 用;给了就导入到临时钥匙串)
#   APPLE_CERTIFICATE_PASSWORD   可选,.p12 密码
#   DRY_RUN=1                    只打印要执行的命令(密码 / 证书内容打码),不执行
#
# 退出码:0 成功;2 参数 / 环境缺失;其它 = 签名工具失败。
set -euo pipefail

DRY_RUN="${DRY_RUN:-0}"

usage() {
  echo "用法: APPLE_SIGNING_IDENTITY='Developer ID Application: …' $0 <bundle-dir>" >&2
  exit 2
}

# 不得出现在输出里的值(证书密码、钥匙串密码)
SECRETS=()

# 打印并执行;DRY_RUN 时只打印。SECRETS 里的值打码成 ***。
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
BUNDLE_DIR="$1"
IDENTITY="${APPLE_SIGNING_IDENTITY:-}"
[ -n "$IDENTITY" ] || {
  echo "sign-macos: 未设置 APPLE_SIGNING_IDENTITY" >&2
  exit 2
}

APP=""
for candidate in "$BUNDLE_DIR"/macos/*.app; do
  if [ -d "$candidate" ]; then
    APP="$candidate"
    break
  fi
done
[ -n "$APP" ] || {
  echo "sign-macos: $BUNDLE_DIR/macos 下没有 .app(先跑 cargo tauri build)" >&2
  exit 2
}
APP_NAME="$(basename "$APP" .app)"

DMG=""
for candidate in "$BUNDLE_DIR"/dmg/*.dmg; do
  if [ -f "$candidate" ]; then
    DMG="$candidate"
    break
  fi
done
if [ -z "$DMG" ]; then
  VERSION="$(bash "$(dirname "$0")/bump-version.sh" --check 2>/dev/null || echo 0.0.0)"
  DMG="$BUNDLE_DIR/dmg/${APP_NAME}_${VERSION}.dmg"
fi

# ---- 可选:把 base64 .p12 导入临时钥匙串(CI) ------------------------------
KEYCHAIN=""
cleanup() {
  if [ -n "$KEYCHAIN" ] && [ "$DRY_RUN" != "1" ]; then
    security delete-keychain "$KEYCHAIN" >/dev/null 2>&1 || true
  fi
  if [ -n "${CERT_PATH:-}" ] && [ "$DRY_RUN" != "1" ]; then
    rm -f "$CERT_PATH"
  fi
}
trap cleanup EXIT

if [ -n "${APPLE_CERTIFICATE:-}" ]; then
  WORK="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
  KEYCHAIN="$WORK/doubao-relay-signing.keychain-db"
  CERT_PATH="$WORK/doubao-relay-signing.p12"
  CERT_PW="${APPLE_CERTIFICATE_PASSWORD:-}"
  KEYCHAIN_PW="$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32 || true)"
  SECRETS=("$CERT_PW" "$KEYCHAIN_PW")
  echo "+ (base64 证书解码到 $CERT_PATH)"
  if [ "$DRY_RUN" != "1" ]; then
    (umask 077 && printf '%s' "$APPLE_CERTIFICATE" | base64 --decode >"$CERT_PATH")
  fi
  run security create-keychain -p "$KEYCHAIN_PW" "$KEYCHAIN"
  run security set-keychain-settings -lut 21600 "$KEYCHAIN"
  run security unlock-keychain -p "$KEYCHAIN_PW" "$KEYCHAIN"
  run security import "$CERT_PATH" -k "$KEYCHAIN" -P "$CERT_PW" -T /usr/bin/codesign -T /usr/bin/security
  run security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$KEYCHAIN_PW" "$KEYCHAIN"
  # 把临时钥匙串加到搜索列表最前面(保留原有的)
  existing=()
  while IFS= read -r line; do
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%\"}"
    line="${line#\"}"
    [ -n "$line" ] && existing+=("$line")
  done < <(security list-keychains -d user 2>/dev/null || true)
  run security list-keychains -d user -s "$KEYCHAIN" ${existing[@]+"${existing[@]}"}
fi

# ---- 签 .app ---------------------------------------------------------------
run codesign --force --deep --timestamp --options runtime --sign "$IDENTITY" "$APP"
run codesign --verify --deep --strict --verbose=2 "$APP"

# ---- 用签好的 .app 重建 DMG(Tauri 生成的 DMG 里是签名前的 .app) ---------------
# 暂存目录 = 签好的 .app + 指向 /Applications 的链接(拖拽安装)
STAGE="${RUNNER_TEMP:-${TMPDIR:-/tmp}}/doubao-relay-dmg-stage"
run rm -rf "$STAGE"
run mkdir -p "$STAGE" "$(dirname "$DMG")"
run ditto "$APP" "$STAGE/$APP_NAME.app"
run ln -s /Applications "$STAGE/Applications"
run rm -f "$DMG"
run hdiutil create -volname "$APP_NAME" -srcfolder "$STAGE" -ov -format UDZO "$DMG"
run rm -rf "$STAGE"
run codesign --force --timestamp --sign "$IDENTITY" "$DMG"
run codesign --verify --verbose=2 "$DMG"

echo "sign-macos: 已签名 $APP"
echo "sign-macos: DMG $DMG"

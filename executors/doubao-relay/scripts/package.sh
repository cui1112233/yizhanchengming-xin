#!/usr/bin/env bash
# 本机打包:macOS universal DMG + Windows x64 zip(在 macOS 上用 cargo-xwin 交叉编译,不走 CI)。
#
#   scripts/package.sh            两个都打
#   scripts/package.sh mac        只打 macOS
#   scripts/package.sh win        只打 Windows
#
# 产物在 dist/:
#   doubao-relay_<ver>_universal.dmg
#   doubao-relay_<ver>_windows_x64.zip   (内含 doubao-relay.exe,解压即用,需系统自带 WebView2)
#
# 前置:rustup target add aarch64-apple-darwin x86_64-apple-darwin x86_64-pc-windows-msvc
#       cargo install cargo-xwin;brew install llvm(提供 clang-cl / llvm-rc / llvm-lib)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
VERSION="$(bash scripts/bump-version.sh --check)"
DIST="$ROOT/dist"
mkdir -p "$DIST"
what="${1:-all}"

build_mac() {
  cargo tauri build --target universal-apple-darwin
  local dmg
  dmg="$(ls target/universal-apple-darwin/release/bundle/dmg/*.dmg | head -n1)"
  cp "$dmg" "$DIST/doubao-relay_${VERSION}_universal.dmg"
  echo "macOS: $DIST/doubao-relay_${VERSION}_universal.dmg"
}

build_win() {
  export PATH="/opt/homebrew/opt/llvm/bin:$PATH"
  cargo tauri build --runner cargo-xwin --target x86_64-pc-windows-msvc --no-bundle
  local exe=target/x86_64-pc-windows-msvc/release/doubao-relay.exe
  [[ -f "$exe" ]] || { echo "未找到 $exe" >&2; exit 1; }
  local stage
  stage="$(mktemp -d)/doubao-relay_${VERSION}_windows_x64"
  mkdir -p "$stage"
  cp "$exe" "$stage/"
  cp README.md CHANGELOG.md LICENSE "$stage/"
  local zip="$DIST/doubao-relay_${VERSION}_windows_x64.zip"
  rm -f "$zip"
  (cd "$(dirname "$stage")" && zip -qr "$zip" "$(basename "$stage")")
  echo "Windows: $zip"
}

case "$what" in
  mac) build_mac ;;
  win) build_win ;;
  all) build_mac; build_win ;;
  *) echo "用法: $0 [mac|win|all]" >&2; exit 2 ;;
esac

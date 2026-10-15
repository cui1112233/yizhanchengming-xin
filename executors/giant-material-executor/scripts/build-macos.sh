#!/bin/bash
set -euo pipefail

module_dir="$(cd "$(dirname "$0")/.." && pwd)"
version="${VERSION:-0.5.0}"
output_dir="${OUTPUT_DIR:-$module_dir/dist/macos-universal}"
public_api_url="${PUBLIC_API_URL:-}"
update_public_key="${UPDATE_PUBLIC_KEY:-}"
archive_name="GiantMaterialExecutor-macos-universal.zip"
stage_dir="$(mktemp -d)"
trap 'rm -rf "$stage_dir"' EXIT
app_dir="$stage_dir/GiantMaterialExecutor.app"
macos_dir="$app_dir/Contents/MacOS"
mkdir -p "$macos_dir" "$output_dir"

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/giant-material-go-cache}"
export CLANG_MODULE_CACHE_PATH="${CLANG_MODULE_CACHE_PATH:-${TMPDIR:-/tmp}/giant-material-clang-cache}"

for arch in arm64 amd64; do
  GOOS=darwin GOARCH="$arch" CGO_ENABLED=0 \
    go -C "$module_dir" build -trimpath -ldflags "-s -w -X main.version=$version -X main.bakedPublicAPIURL=$public_api_url -X main.bakedUpdatePublicKey=$update_public_key" \
    -o "$stage_dir/executor-$arch" ./cmd/giant-material-executor
done
lipo -create "$stage_dir/executor-arm64" "$stage_dir/executor-amd64" -output "$macos_dir/GiantMaterialExecutor"

for arch in arm64 amd64; do
  GOOS=darwin GOARCH="$arch" CGO_ENABLED=0 \
    go -C "$module_dir" build -trimpath -ldflags "-s -w" \
    -o "$stage_dir/updater-$arch" ./cmd/giant-material-executor-updater
done
lipo -create "$stage_dir/updater-arm64" "$stage_dir/updater-amd64" -output "$macos_dir/GiantMaterialExecutorUpdater"

for arch in arm64 x86_64; do
  scratch="$output_dir/swift-$arch"
  swift build --disable-sandbox --package-path "$module_dir/macos" \
    --scratch-path "$scratch" --triple "$arch-apple-macosx13.0" -c release \
    -Xswiftc -module-cache-path -Xswiftc "$output_dir/module-cache-$arch"
  bin_dir="$(swift build --disable-sandbox --package-path "$module_dir/macos" --scratch-path "$scratch" --triple "$arch-apple-macosx13.0" -c release --show-bin-path -Xswiftc -module-cache-path -Xswiftc "$output_dir/module-cache-$arch")"
  cp "$bin_dir/GiantMaterialOCRWorker" "$stage_dir/worker-$arch"
done
lipo -create "$stage_dir/worker-arm64" "$stage_dir/worker-x86_64" -output "$macos_dir/GiantMaterialOCRWorker"

cp "$module_dir/macos/Info.plist" "$app_dir/Contents/Info.plist"
cp "$module_dir/portable/README-macOS.txt" "$stage_dir/README-macOS.txt"
chmod 755 "$macos_dir/GiantMaterialExecutor" "$macos_dir/GiantMaterialExecutorUpdater" "$macos_dir/GiantMaterialOCRWorker"
codesign --force --deep --sign - "$app_dir"
ditto -c -k --sequesterRsrc --keepParent "$app_dir" "$output_dir/$archive_name"
unzip -l "$output_dir/$archive_name"
shasum -a 256 "$output_dir/$archive_name"

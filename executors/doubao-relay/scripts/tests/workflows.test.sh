#!/usr/bin/env bash
# .github/workflows 与 .gitattributes 的静态检查(没有 actionlint 依赖,只做本仓库踩过的坑):
#   - ci.yml / release.yml 存在
#   - release.yml 的每个 upload-artifact 只给单个 path(多路径会以公共祖先为根,
#     下载后安装包落进 dist/msi/、dist/nsis/ 子目录,被 files: dist/* 漏掉)
#   - GitHub Release 的 files 模式分别覆盖 dmg / msi / setup.exe,并对一个模拟的 dist/ 平铺目录逐个匹配
#   - .gitattributes 让版本号文件在 Windows checkout 时保持 LF
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
CI="$REPO/.github/workflows/ci.yml"
RELEASE="$REPO/.github/workflows/release.yml"

fails=0
pass() { printf '  ok   %s\n' "$1"; }
fail() { printf '  FAIL %s\n' "$1"; fails=$((fails + 1)); }

echo "workflows"

for f in "$CI" "$RELEASE"; do
  if [ -f "$f" ]; then pass "exists ${f#"$REPO"/}"; else fail "missing ${f#"$REPO"/}"; fi
done

if [ -f "$RELEASE" ]; then
  # upload-artifact 的 path 必须是单行(不是 `path: |` 多路径块)
  multi="$(perl -ne '
    $in = 1 if /uses:\s*actions\/upload-artifact/;
    if ($in && /^\s*path:\s*\|/) { print "line $.\n" }
    $in = 0 if $in && /^\s*-\s+(name|uses):/ && !/upload-artifact/;
  ' "$RELEASE")"
  if [ -z "$multi" ]; then pass "upload-artifact uses a single path"; else fail "upload-artifact multi-path at $multi"; fi

  # 取 action-gh-release 的 files: 块
  patterns="$(perl -ne '
    if (/^(\s*)files:\s*\|\s*$/) { $ind = length($1); $on = 1; next }
    if (/^(\s*)files:\s*(\S.*)$/) { print "$2\n"; next }
    if ($on) {
      if (/^(\s*)(\S.*)$/ && length($1) > $ind) { print "$2\n" } else { $on = 0 }
    }
  ' "$RELEASE")"
  if [ -n "$patterns" ]; then pass "release files patterns found"; else fail "release files patterns not found"; fi

  # 模拟 download-artifact merge-multiple 之后的 dist/(平铺)
  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT
  mkdir -p "$TMP/dist"
  : >"$TMP/dist/doubao-relay_0.1.0_universal.dmg"
  : >"$TMP/dist/doubao-relay_0.1.0_x64_en-US.msi"
  : >"$TMP/dist/doubao-relay_0.1.0_x64-setup.exe"
  for want in .dmg .msi -setup.exe; do
    hit=0
    while IFS= read -r p; do
      [ -n "$p" ] || continue
      # 每个模式都必须有匹配(对应 fail_on_unmatched_files)
      for f in "$TMP"/$p; do
        case "$f" in *"$want") [ -e "$f" ] && hit=1 ;; esac
      done
    done <<<"$patterns"
    if [ "$hit" -eq 1 ]; then pass "release files cover *$want"; else fail "release files do not cover *$want"; fi
  done
  while IFS= read -r p; do
    [ -n "$p" ] || continue
    set -- "$TMP"/$p
    if [ -e "$1" ]; then pass "pattern [$p] matches a file"; else fail "pattern [$p] matches nothing"; fi
  done <<<"$patterns"

  if grep -q 'fail_on_unmatched_files: true' "$RELEASE"; then
    pass "fail_on_unmatched_files"
  else
    fail "fail_on_unmatched_files not set"
  fi
fi

# .gitattributes:Windows runner 默认 core.autocrlf=true,版本号文件必须保持 LF
if git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1; then
  for f in Cargo.lock src-tauri/Cargo.toml src-tauri/tauri.conf.json package.json frontend/package.json README.md scripts/bump-version.sh; do
    eol="$(git -C "$REPO" check-attr eol -- "$f" | sed 's/.*: eol: //')"
    if [ "$eol" = "lf" ]; then pass ".gitattributes eol=lf $f"; else fail ".gitattributes eol for $f is [$eol]"; fi
  done
fi

if [ "$fails" -ne 0 ]; then
  echo "workflows: $fails failure(s)"
  exit 1
fi
echo "workflows: all passed"

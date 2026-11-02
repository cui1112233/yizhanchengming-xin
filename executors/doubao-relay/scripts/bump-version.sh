#!/usr/bin/env bash
# 版本号同步(spec 13「版本号」:Rust 与前端版本号 MUST 同步)。
#
# 用法:
#   scripts/bump-version.sh <X.Y.Z[-N]>        把版本号写进下面全部位置(预发布只允许纯数字 N,见 SEMVER_RE)
#   scripts/bump-version.sh --check [vX.Y.Z]   校验各处一致并打印版本;给了期望版本(可带 v 前缀,
#                                             即 tag 名)时还要求等于它。不一致退出 1。
#
# 版本号所在位置(spec 13 要求三处,另外三处是为了不让它们悄悄漂移):
#   src-tauri/Cargo.toml        [package] version
#   src-tauri/tauri.conf.json   顶层 "version"
#   frontend/package.json       "version"
#   package.json                "version"(根编排脚本)
#   Cargo.lock                  [[package]] name = "doubao-relay" 的 version(CI 用 --locked,不同步会失败)
#   README.md                   shields 版本徽章 badge/version-<v>-(没有徽章则跳过)
#
# 只依赖 bash + perl(macOS / GitHub runner / Git for Windows 都自带)。
# 环境变量 BUMP_ROOT 可指向别的仓库根(测试用)。
set -euo pipefail

ROOT="${BUMP_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"

# 预发布段只允许纯数字(X.Y.Z-N,N ≤ 65535):Tauri 的 MSI(WiX)目标拒绝非数字 / 超过 65535 的
# 预发布标识,0.3.0-rc.1 这种 tag 会让 release.yml 的 Windows 构建失败,所以在这里就拒绝。
SEMVER_RE='^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]{0,4}))?$'

valid_version() {
  [[ "$1" =~ $SEMVER_RE ]] || return 1
  case "$1" in
    *-*) [ "${1##*-}" -le 65535 ] ;;
  esac
}

usage() {
  echo "用法: $0 <X.Y.Z[-N]> | --check [vX.Y.Z[-N]]" >&2
  exit 2
}

die() {
  echo "bump-version: $*" >&2
  exit 1
}

need_file() {
  [ -f "$ROOT/$1" ] || die "缺少文件 $1"
}

# ---- 读 -------------------------------------------------------------------

read_cargo_toml() {
  perl -ne '
    if (/^\s*\[package\]\s*$/) { $p = 1; next }
    if (/^\s*\[/) { $p = 0 }
    if ($p && /^\s*version\s*=\s*"([^"]+)"/) { print $1; exit }
  ' "$ROOT/src-tauri/Cargo.toml"
}

# 顶层对象里第一个 "version" 键(两个 package.json 与 tauri.conf.json 的 version 都在嵌套对象之前)
read_json() {
  perl -0ne 'print $1 if /\A\s*\{[^{}]*?"version"\s*:\s*"([^"]+)"/s' "$ROOT/$1"
}

read_cargo_lock() {
  perl -0ne 'print $1 if /\[\[package\]\]\r?\nname = "doubao-relay"\r?\nversion = "([^"]+)"/' "$ROOT/Cargo.lock"
}

# shields 徽章把 '-' 写成 '--'
read_readme_badge() {
  perl -ne 'if (m{img\.shields\.io/badge/version-((?:[^-\s]|--)+)-}) { ($v = $1) =~ s/--/-/g; print $v; exit }' "$ROOT/README.md"
}

has_readme_badge() {
  [ -f "$ROOT/README.md" ] && grep -q 'img\.shields\.io/badge/version-' "$ROOT/README.md"
}

# ---- 写 -------------------------------------------------------------------

write_cargo_toml() {
  V="$1" perl -pi -e '
    if (/^\s*\[package\]\s*$/) { $p = 1 }
    elsif (/^\s*\[/) { $p = 0 }
    if ($p && !$done && s/^(\s*version\s*=\s*")[^"]+(")/$1$ENV{V}$2/) { $done = 1 }
  ' "$ROOT/src-tauri/Cargo.toml"
}

write_json() {
  V="$2" perl -0pi -e 's/\A(\s*\{[^{}]*?"version"\s*:\s*")[^"]+(")/$1$ENV{V}$2/s' "$ROOT/$1"
}

write_cargo_lock() {
  V="$1" perl -0pi -e 's/(\[\[package\]\]\r?\nname = "doubao-relay"\r?\nversion = ")[^"]+(")/$1$ENV{V}$2/' "$ROOT/Cargo.lock"
}

write_readme_badge() {
  local escaped="${1//-/--}"
  V="$escaped" perl -pi -e 's{(img\.shields\.io/badge/version-)(?:[^-\s]|--)+(-)}{$1$ENV{V}$2}' "$ROOT/README.md"
}

# ---- 主流程 ---------------------------------------------------------------

JSON_FILES=("src-tauri/tauri.conf.json" "frontend/package.json" "package.json")

collect() { # 输出 "文件<TAB>版本" 行
  need_file "src-tauri/Cargo.toml"
  printf 'src-tauri/Cargo.toml\t%s\n' "$(read_cargo_toml)"
  for f in "${JSON_FILES[@]}"; do
    need_file "$f"
    printf '%s\t%s\n' "$f" "$(read_json "$f")"
  done
  need_file "Cargo.lock"
  printf 'Cargo.lock\t%s\n' "$(read_cargo_lock)"
  if has_readme_badge; then
    printf 'README.md\t%s\n' "$(read_readme_badge)"
  fi
}

check() {
  local expected="${1:-}"
  expected="${expected#v}"
  if [ -n "$expected" ] && ! valid_version "$expected"; then
    die "期望版本不合法: '$expected'(要求 X.Y.Z 或 X.Y.Z-N,N 为 ≤65535 的数字)"
  fi
  local table first="" bad=0
  table="$(collect)"
  while IFS=$'\t' read -r file ver; do
    [ -n "$ver" ] || {
      echo "bump-version: $file 里没找到版本号" >&2
      bad=1
      continue
    }
    [ -n "$first" ] || first="$ver"
  done <<<"$table"
  local want="${expected:-$first}"
  while IFS=$'\t' read -r file ver; do
    if [ -n "$ver" ] && [ "$ver" != "$want" ]; then
      echo "bump-version: $file 的版本是 $ver,期望 $want" >&2
      bad=1
    fi
  done <<<"$table"
  [ "$bad" -eq 0 ] || exit 1
  echo "$want"
}

bump() {
  local v="$1"
  valid_version "$v" || die "版本号不合法: '$v'(要求 X.Y.Z 或 X.Y.Z-N,N 为 ≤65535 的数字,不带 v 前缀)"
  # 先确认所有文件都在、每处版本号都读得到,再动手(不留半截修改)
  local table file ver
  table="$(collect)"
  while IFS=$'\t' read -r file ver; do
    [ -n "$ver" ] || die "$file 里没找到版本号,未修改任何文件"
  done <<<"$table"
  write_cargo_toml "$v"
  for f in "${JSON_FILES[@]}"; do write_json "$f" "$v"; done
  write_cargo_lock "$v"
  if has_readme_badge; then write_readme_badge "$v"; fi
  check "$v" >/dev/null
  echo "$v"
  if [ -f "$ROOT/CHANGELOG.md" ] && ! grep -q "^## \[$v\]" "$ROOT/CHANGELOG.md"; then
    echo "bump-version: 提醒:CHANGELOG.md 还没有 ## [$v] 段落" >&2
  fi
}

[ $# -ge 1 ] || usage
case "$1" in
  --check)
    [ $# -le 2 ] || usage
    check "${2:-}"
    ;;
  -h | --help) usage ;;
  *)
    [ $# -eq 1 ] || usage
    bump "$1"
    ;;
esac

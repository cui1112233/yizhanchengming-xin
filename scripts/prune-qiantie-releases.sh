#!/bin/sh
# 清理生产版本文件，避免每次发布都无限堆积 174MB 的旧二进制。
# 保留规则：当前软链版本 + runtime-release.json 记录的回滚版本 + 最近 N 个历史版本。
# 用法：sh prune-qiantie-releases.sh [额外保留历史版本数，默认 1]
#      QIANTIE_RELEASE_ROOT=/opt/www/qiantie
set -eu

ROOT=${QIANTIE_RELEASE_ROOT:-/opt/www/qiantie}
KEEP_PREV=${1:-1}
DRY_RUN=${DRY_RUN:-0}
case "$KEEP_PREV" in
  ''|*[!0-9]*) echo "保留数量必须是数字：$KEEP_PREV" >&2; exit 2 ;;
esac

LINK="$ROOT/qiantie"
if [ ! -L "$LINK" ]; then
  echo "未找到版本软链：$LINK" >&2
  exit 1
fi
CURRENT=$(readlink -f "$LINK")
if [ -z "$CURRENT" ] || [ ! -f "$CURRENT" ]; then
  echo "无法解析当前版本：$CURRENT" >&2
  exit 1
fi

ROLLBACK=""
if [ -f "$ROOT/runtime-release.json" ]; then
  ROLLBACK=$(python3 - "$ROOT/runtime-release.json" <<'PY'
import json, os, sys
try:
    d = json.load(open(sys.argv[1], encoding="utf-8"))
except Exception:
    d = {}
name = (d.get("rollbackBinary") or "").strip()
print(os.path.join(os.path.dirname(sys.argv[1]), name) if name else "")
PY
)
fi

KEEP="$CURRENT"
[ -n "$ROLLBACK" ] && [ -f "$ROLLBACK" ] && KEEP="$KEEP
$ROLLBACK"

# 按修改时间保留最近 N 个历史版本；当前版本不重复计数。
EXTRA=$(ls -1t "$ROOT"/qiantie-* 2>/dev/null | while read -r f; do
  [ -f "$f" ] || continue
  [ "$f" = "$CURRENT" ] && continue
  echo "$f"
done | head -n "$KEEP_PREV" || true)
if [ -n "$EXTRA" ]; then
  KEEP="$KEEP
$EXTRA"
fi

deleted=0
freed=0
for f in "$ROOT"/qiantie-*; do
  [ -f "$f" ] || continue
  case "
$KEEP
" in
    *"
$f
"*) continue ;;
  esac
  size=$(stat -c %s "$f" 2>/dev/null || echo 0)
  if [ "$DRY_RUN" = "1" ]; then
    echo "would_delete=$(basename "$f")"
  else
    rm -f "$f"
  fi
  deleted=$((deleted + 1))
  freed=$((freed + size))
done

# rollback-* 备份目录同样只保留最近 1 个。
rollback_dirs=$(ls -1dt "$ROOT"/rollback-* 2>/dev/null | tail -n +2 || true)
if [ -n "$rollback_dirs" ]; then
  echo "$rollback_dirs" | while read -r d; do
    [ -d "$d" ] || continue
    du -sk "$d" 2>/dev/null | awk '{print $1}'
    if [ "$DRY_RUN" = "1" ]; then
      echo "would_delete_dir=$(basename "$d")" >&2
    else
      rm -rf "$d"
    fi
  done | awk '{s+=$1} END {if (s>0) printf "rollback_removed_kb=%d\n", s}'
fi

echo "current=$(basename "$CURRENT")"
kept_names=""
for f in $KEEP; do
  [ -n "$f" ] || continue
  kept_names="$kept_names $(basename "$f")"
done
echo "kept=$kept_names"
echo "deleted=$deleted freed_mb=$((freed / 1024 / 1024))"

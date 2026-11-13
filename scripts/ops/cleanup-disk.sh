#!/usr/bin/env bash
# 磁盘自动清理：MySQL binlog、历史发布、临时构建、日志、过期媒体。
# 设计原则：只删除可以重新生成或已过保留期的内容，不碰数据库与当前发布。
set -u

MEDIA_RETENTION_DAYS="${MEDIA_RETENTION_DAYS:-21}"
ARTIFACT_DIR="${QIANTIE_LOCAL_EXECUTOR_ARTIFACT_DIR:-/opt/www/qiantie-artifacts}"
RELEASE_DIR="/opt/www/qiantie"
BINLOG_MINUTES="${BINLOG_MINUTES:-15}"
DISK_LIMIT_PERCENT="${DISK_LIMIT_PERCENT:-80}"

log() { echo "[$(date '+%F %T')] $*"; }
usage_percent() { df --output=pcent / | tail -1 | tr -dc '0-9'; }

# 1. MySQL binlog：单实例没有副本，自动化高频写 job JSON 会把 binlog 撑到
#    每分钟上百兆。保留最近 15 分钟，足够排障，不足以撑爆磁盘。
if command -v mysql >/dev/null 2>&1; then
  mysql -e "PURGE BINARY LOGS BEFORE NOW() - INTERVAL ${BINLOG_MINUTES} MINUTE;" 2>/dev/null \
    && log "binlog purged before -${BINLOG_MINUTES}min"
  mysql -e "SET GLOBAL binlog_expire_logs_seconds=$((BINLOG_MINUTES * 60));" 2>/dev/null || true
  # 单实例没有副本/CDC 消费方。MINIMAL 只记录被修改列，避免大 JSON 行在每次
  # 更新时把修改前后两份都写进 binlog（job_json 单条约 400KB）。
  mysql -e "SET GLOBAL binlog_row_image='MINIMAL';" 2>/dev/null || true
fi

# 2. 历史发布二进制：只保留当前软链指向的版本和最近一个回滚版本。
if [ -d "$RELEASE_DIR" ]; then
  current="$(basename "$(readlink -f "$RELEASE_DIR/qiantie")")"
  previous="$(ls -1d "$RELEASE_DIR"/qiantie-* 2>/dev/null | sort -r | xargs -r -n1 basename | grep -vx "$current" | head -1)"
  for dir in "$RELEASE_DIR"/qiantie-*; do
    [ -d "$dir" ] || continue
    name="$(basename "$dir")"
    case "$name" in
      "$current"|"$previous") ;;
      *) rm -rf "$dir" && log "removed old release $name" ;;
    esac
  done
fi

# 3. 临时构建残留。
find /tmp -maxdepth 1 \( -name 'qiantie-*' -o -name 'deploy-*.sh' \) -type f -mtime +1 -delete 2>/dev/null
find /tmp -maxdepth 1 -type d -name 'qiantie-stage-*' -mtime +1 -exec rm -rf {} + 2>/dev/null

# 4. 日志体积上限，避免 journal 长期占用。
journalctl --vacuum-size=200M >/dev/null 2>&1 || true

# 5. 过期媒体（页图、缩略图、下载包）。只在磁盘超过阈值或文件超期时删除。
if [ -d "$ARTIFACT_DIR" ]; then
  freed=0
  while IFS= read -r -d '' file; do
    rm -f "$file" && freed=$((freed + 1))
  done < <(find "$ARTIFACT_DIR" -type f -mtime +"$MEDIA_RETENTION_DAYS" -print0 2>/dev/null)
  [ "$freed" -gt 0 ] && log "removed $freed media files older than ${MEDIA_RETENTION_DAYS}d"
fi

# 5b. 漫画缩略图只是派生缓存，原图在 TOS 且可随时重新生成；单独用短保留期
#     清理，避免小磁盘被首屏缩略图长期占满。
THUMB_DIR="$ARTIFACT_DIR/comic-thumbnails"
if [ -d "$THUMB_DIR" ]; then
  thumbs=0
  while IFS= read -r -d '' file; do
    rm -f "$file" && thumbs=$((thumbs + 1))
  done < <(find "$THUMB_DIR" -type f -mtime +3 -print0 2>/dev/null)
  [ "$thumbs" -gt 0 ] && log "removed $thumbs comic thumbnails older than 3d"
fi

percent="$(usage_percent)"
log "disk usage ${percent}%"
if [ "$percent" -ge "$DISK_LIMIT_PERCENT" ]; then
  # 仍然紧张时，把媒体保留期减半再清一次（最低 3 天）。
  half=$((MEDIA_RETENTION_DAYS / 2))
  [ "$half" -lt 3 ] && half=3
  while IFS= read -r -d '' file; do
    rm -f "$file"
  done < <(find "$ARTIFACT_DIR" -type f -mtime +"$half" -print0 2>/dev/null)
  log "disk still ${percent}%, trimmed media older than ${half}d"
fi

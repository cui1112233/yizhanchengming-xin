#!/usr/bin/env bash
set -Eeuo pipefail

BACKUP_ROOT="${QIANTIE_MYSQL_BACKUP_ROOT:-/opt/qiantie/backups/mysql}"
CREDENTIALS_FILE="${QIANTIE_MYSQL_CREDENTIALS_FILE:-$BACKUP_ROOT/credentials.cnf}"
LOCK_FILE="${QIANTIE_MYSQL_BACKUP_LOCK:-/run/qiantie-mysql-backup.lock}"
V78_CONTAINER="${QIANTIE_V78_MYSQL_CONTAINER:-v78-public-mysql-1}"
V88_CONTAINER="${QIANTIE_V88_MYSQL_CONTAINER:-v88-public-mysql-1}"
UNIFIED_CONTAINER="${QIANTIE_UNIFIED_MYSQL_CONTAINER:-qiantie-unified-mysql}"
HOST_V78_DB="${QIANTIE_BACKUP_V78_DB:-qiantie_v78}"
HOST_V88_DB="${QIANTIE_BACKUP_V88_DB:-qiantie_v88}"
MIN_FREE_MB="${QIANTIE_MYSQL_BACKUP_MIN_FREE_MB:-2048}"
RETENTION_DAYS="${QIANTIE_MYSQL_BACKUP_RETENTION_DAYS:-14}"
STAMP="$(date +%F)"
DAY_DIR="$BACKUP_ROOT/$STAMP"

die() { echo "mysql-backup: $*" >&2; exit 1; }
log() { echo "mysql-backup: $*"; }

[ "$(id -u)" -eq 0 ] || die "必须以 root 执行"
[ -r "$CREDENTIALS_FILE" ] || die "凭据文件不存在或不可读"
command -v gzip >/dev/null || die "gzip 不可用"
command -v sha256sum >/dev/null || die "sha256sum 不可用"
command -v flock >/dev/null || die "flock 不可用"

# V88 应用已经改为 Go 直连本机 MySQL，容器化 MySQL 不再是唯一形态。
# 有容器就走容器，否则回退到本机 mysqldump，避免备份任务因为“docker 不可用”整体停摆。
MODE=""
if command -v docker >/dev/null 2>&1 &&
  docker inspect "$UNIFIED_CONTAINER" >/dev/null 2>&1 &&
  [ "$(docker inspect -f '{{.State.Status}}' "$UNIFIED_CONTAINER")" = running ]; then
  MODE=unified
elif command -v docker >/dev/null 2>&1 &&
  docker inspect "$V78_CONTAINER" >/dev/null 2>&1 &&
  docker inspect "$V88_CONTAINER" >/dev/null 2>&1; then
  MODE=docker
else
  command -v mysqldump >/dev/null || die "docker 与 mysqldump 均不可用"
  MODE=host
fi

mkdir -p "$BACKUP_ROOT"
chmod 700 "$BACKUP_ROOT"
exec 9>"$LOCK_FILE"
flock -n 9 || { log "已有备份任务运行，跳过"; exit 0; }

free_mb="$(df -Pm "$BACKUP_ROOT" | awk 'NR==2 {print $4}')"
case "$free_mb" in ''|*[!0-9]*) die "无法读取备份目录剩余空间" ;; esac
[ "$free_mb" -ge "$MIN_FREE_MB" ] || die "备份目录剩余空间不足：${free_mb}MB，要求至少 ${MIN_FREE_MB}MB"

mkdir -p "$DAY_DIR"
chmod 700 "$DAY_DIR"
cleanup() {
  rm -f "$DAY_DIR"/*.partial "$DAY_DIR"/*.sql.gz.partial "$DAY_DIR"/SHA256SUMS.partial "$DAY_DIR"/BACKUP_OK.partial
  if command -v docker >/dev/null 2>&1; then
    docker exec "$V78_CONTAINER" rm -f /tmp/qiantie-mysql-backup.cnf >/dev/null 2>&1 || true
    docker exec "$V88_CONTAINER" rm -f /tmp/qiantie-mysql-backup.cnf >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

backup_one() {
  local suffix="$1" container="$2" database="$3" output="$4"
  docker inspect "$container" >/dev/null 2>&1 || die "容器不存在：$container"
  docker cp "$CREDENTIALS_FILE" "$container:/tmp/qiantie-mysql-backup.cnf"
  docker exec "$container" chmod 600 /tmp/qiantie-mysql-backup.cnf
  log "开始备份 $suffix（容器 $container / 库 $database）"
  docker exec "$container" mysqldump \
    --defaults-extra-file=/tmp/qiantie-mysql-backup.cnf \
    --defaults-group-suffix="-$suffix" \
    --single-transaction --routines --events --triggers --hex-blob \
    --databases "$database" \
    | gzip -c > "$output.partial"
  gzip -t "$output.partial"
  mv "$output.partial" "$output"
  [ -s "$output" ] || die "$suffix dump 为空"
  log "完成备份 $suffix"
}

backup_host() {
  local suffix="$1" database="$2" output="$3"
  mysql --defaults-extra-file="$CREDENTIALS_FILE" --defaults-group-suffix="-$suffix" \
    -N -e 'SELECT 1' >/dev/null || die "本机 MySQL 凭据不可用：$suffix"
  log "开始备份 $suffix（本机 / 库 $database）"
  mysqldump \
    --defaults-extra-file="$CREDENTIALS_FILE" \
    --defaults-group-suffix="-$suffix" \
    --single-transaction --routines --events --triggers --hex-blob \
    --databases "$database" \
    | gzip -c > "$output.partial"
  gzip -t "$output.partial"
  mv "$output.partial" "$output"
  [ -s "$output" ] || die "$suffix dump 为空"
  log "完成备份 $suffix"
}

case "$MODE" in
  unified)
    backup_one v88 "$UNIFIED_CONTAINER" qiantie_v78 "$DAY_DIR/qiantie_v78.sql.gz"
    backup_one v88 "$UNIFIED_CONTAINER" qiantie_v88 "$DAY_DIR/qiantie_v88.sql.gz"
    ;;
  docker)
    backup_one v78 "$V78_CONTAINER" qiantie "$DAY_DIR/qiantie_v78.sql.gz"
    backup_one v88 "$V88_CONTAINER" qiantie "$DAY_DIR/qiantie_v88.sql.gz"
    ;;
  host)
    backup_host v78 "$HOST_V78_DB" "$DAY_DIR/qiantie_v78.sql.gz"
    backup_host v88 "$HOST_V88_DB" "$DAY_DIR/qiantie_v88.sql.gz"
    ;;
esac

(
  cd "$DAY_DIR"
  sha256sum qiantie_v78.sql.gz qiantie_v88.sql.gz > SHA256SUMS.partial
)
mv "$DAY_DIR/SHA256SUMS.partial" "$DAY_DIR/SHA256SUMS"
(cd "$DAY_DIR" && sha256sum -c SHA256SUMS) >/dev/null
printf 'backup_date=%s\nmode=%s\n' "$STAMP" "$MODE" > "$DAY_DIR/BACKUP_OK.partial"
mv "$DAY_DIR/BACKUP_OK.partial" "$DAY_DIR/BACKUP_OK"
chmod 600 "$DAY_DIR"/qiantie_*.sql.gz "$DAY_DIR"/SHA256SUMS "$DAY_DIR"/BACKUP_OK

latest="$(find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -name '20??-??-??' -exec test -f '{}/BACKUP_OK' \; -print | sort | tail -n 1)"
find "$BACKUP_ROOT" -mindepth 1 -maxdepth 1 -type d -name '20??-??-??' -mtime "+$RETENTION_DAYS" -print0 |
  while IFS= read -r -d '' old_dir; do
    [ "$old_dir" = "$latest" ] && continue
    [ -f "$old_dir/BACKUP_OK" ] || continue
    rm -rf -- "$old_dir"
  done
log "备份成功：$DAY_DIR（模式 $MODE）"

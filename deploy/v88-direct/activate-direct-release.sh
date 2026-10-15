#!/usr/bin/env bash
# V88 Go 代码直传：在服务器上把 current 软链切到指定版本并重启 go-api。
# 由 Windows 端 deploy-v88-direct.ps1 通过 SSH 调用，也可手工执行：
#   bash activate-direct-release.sh <git-sha>
#
# 安全保证：
#   - 只重建 go-api 以及反代它的 Nginx；不碰 MySQL/浏览器工人；
#   - 切换前记录上一版软链，健康检查失败自动回滚并恢复旧版本；
#   - 不修改、不依赖任何国外镜像拉取。
set -Eeuo pipefail

SHA="${1:-}"
[[ "$SHA" =~ ^[0-9a-f]{40}$ ]] || { echo "usage: activate-direct-release.sh <exact-40-character-git-sha>" >&2; exit 2; }

DIRECT_ROOT="${DIRECT_ROOT:-/opt/qiantie/v88/direct}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/qiantie/v88/deploy/v88-public}"
BASE_COMPOSE="$DEPLOY_DIR/docker-compose.yml"
OVERRIDE="$DIRECT_ROOT/docker-compose.direct.yml"
RELEASE_DIR="$DIRECT_ROOT/releases/$SHA"

[ -x "$RELEASE_DIR/go/qiantie" ] || { echo "STATUS=FAILED missing go binary: $RELEASE_DIR/go/qiantie" >&2; exit 3; }
[ -s "$BASE_COMPOSE" ] || { echo "STATUS=FAILED missing $BASE_COMPOSE" >&2; exit 5; }
[ -s "$OVERRIDE" ] || { echo "STATUS=FAILED missing $OVERRIDE" >&2; exit 6; }
cd "$DEPLOY_DIR"

# 1) 记录回滚点（当前软链目标 + 当前 .env 中的 DIRECT_RELEASE_SHA）。
# readlink 返回的是当初写进软链的“原始字符串”：本脚本写的是绝对路径，
# 历史脚本可能写相对路径。这里统一折算成可直接判断的绝对目录 PREV_DIR。
# 不折算的话 "$DIRECT_ROOT/$PREV_TARGET" 会拼成 /opt/...//opt/... 这种假路径，
# 让 -d 判断永远为假：previous 软链永不更新、回滚还会被整个跳过。
PREV_TARGET="$(readlink "$DIRECT_ROOT/current" 2>/dev/null || true)"
PREV_SHA="$(grep -E '^DIRECT_RELEASE_SHA=' .env 2>/dev/null | tail -1 | cut -d= -f2 || true)"
if [ -n "$PREV_TARGET" ]; then
  case "$PREV_TARGET" in
    /*) PREV_DIR="$PREV_TARGET" ;;
    *)  PREV_DIR="$DIRECT_ROOT/$PREV_TARGET" ;;
  esac
else
  PREV_DIR=""
fi

# Validate provenance and atomically write runtime identity before installing the
# rollback trap: a preflight error must not recreate containers or touch current.
PREV_GIT_SHA=""
if [ -n "$PREV_DIR" ]; then
  PREV_GIT_SHA="$(python3 - "$PREV_DIR" "$PREV_SHA" <<'PY'
import json
from pathlib import Path
import re
import sys

try:
    release = Path(sys.argv[1])
    fallback_sha = sys.argv[2]
    manifest = release / "runtime-release.json"
    if manifest.exists():
        identity = json.loads(manifest.read_text(encoding="utf-8"))
        sha = identity["gitSha"]
    else:
        # Legacy direct releases may predate both manifests and RELEASE-SHA.
        # Only the already-active exact .env identity may bridge that gap; a
        # malformed existing identity remains a hard failure rather than being
        # silently hidden by the fallback.
        if re.fullmatch(r"[0-9a-f]{40}", fallback_sha):
            sha = fallback_sha
        else:
            raise ValueError("previous release has no identity and DIRECT_RELEASE_SHA is not exact")
    if not isinstance(sha, str) or not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise ValueError("previous Git identity is not an exact SHA")
    print(sha)
except (OSError, ValueError, KeyError, TypeError) as error:
    print(f"STATUS=FAILED previous release identity: {error}", file=sys.stderr)
    sys.exit(9)
PY
  )" || exit 9
fi
MANIFEST_SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/runtime-release-manifest.sh"
[ -f "$MANIFEST_SCRIPT" ] || { echo "STATUS=FAILED missing runtime manifest generator" >&2; exit 9; }
bash "$MANIFEST_SCRIPT" "$SHA" "$RELEASE_DIR" "$PREV_GIT_SHA" || exit 9
python3 - "$RELEASE_DIR" "$SHA" "$PREV_GIT_SHA" <<'PY' || exit 9
import datetime
import hashlib
import json
from pathlib import Path
import re
import sys
try:
    release = Path(sys.argv[1])
    with (release / "runtime-release.json").open(encoding="utf-8") as stream:
        manifest = json.load(stream)
    if manifest.get("gitSha") != sys.argv[2]:
        raise ValueError("generated manifest does not match requested Git SHA")
    if manifest.get("previousReleaseGitSha") != (sys.argv[3] or None):
        raise ValueError("generated manifest does not match previous Git SHA")
    timestamp = manifest.get("activatedAtUtc", "")
    datetime.datetime.strptime(timestamp, "%Y-%m-%dT%H:%M:%SZ")
    go_hash = manifest.get("goBinarySha256", "")
    if not isinstance(go_hash, str) or not re.fullmatch(r"[0-9a-f]{64}", go_hash):
        raise ValueError("generated manifest lacks an exact Go binary SHA-256")
    digest = hashlib.sha256()
    with (release / "go" / "qiantie").open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    if digest.hexdigest() != go_hash:
        raise ValueError("generated manifest does not match staged Go binary")
except (OSError, ValueError, AttributeError, TypeError) as error:
    print(f"STATUS=FAILED runtime manifest validation: {error}", file=sys.stderr)
    sys.exit(9)
PY

set_env() {
  local key="$1" value="$2" next_env
  if grep -q "^${key}=" .env; then
    next_env="$(mktemp .env.XXXXXX)"
    sed "s|^${key}=.*|${key}=${value}|" .env > "$next_env"
    mv "$next_env" .env
  else
    printf '\n%s=%s\n' "$key" "$value" >> .env
  fi
}

wait_for_go_api() {
  local go_container health
  go_container="$(docker compose -f "$BASE_COMPOSE" -f "$OVERRIDE" ps -q go-api)"
  [ -n "$go_container" ] || { echo "STATUS=FAILED go-api container was not created" >&2; return 1; }
  for _ in $(seq 1 45); do
    health="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$go_container" 2>/dev/null || true)"
    if [ "$health" = "healthy" ]; then return 0; fi
    if [ "$health" = "unhealthy" ]; then
      echo "STATUS=FAILED go-api healthcheck failed" >&2
      return 1
    fi
    sleep 2
  done
  echo "STATUS=FAILED go-api did not become healthy" >&2
  return 1
}

rollback() {
  local status=$?
  echo "STATUS=ROLLBACK release failed, restoring previous state (exit=$status)" >&2
  if [ -n "$PREV_DIR" ] && [ -d "$PREV_DIR" ]; then
    ln -sfn "$PREV_DIR" "$DIRECT_ROOT/current"
    if [ -n "$PREV_SHA" ]; then set_env DIRECT_RELEASE_SHA "$PREV_SHA"; fi
    docker compose -f "$BASE_COMPOSE" -f "$OVERRIDE" up -d --no-deps --force-recreate --pull never go-api >/dev/null 2>&1 || true
    wait_for_go_api >/dev/null 2>&1 || true
    docker compose -f "$BASE_COMPOSE" restart nginx >/dev/null 2>&1 || true
  else
    # First direct release has no prior release symlink. Restore the base
    # image runtime rather than mounting the failed, partial release again.
    rm -f "$DIRECT_ROOT/current"
    set_env DIRECT_RELEASE_SHA ""
    docker compose -f "$BASE_COMPOSE" up -d --no-deps --force-recreate --pull never go-api >/dev/null 2>&1 || true
    docker compose -f "$BASE_COMPOSE" restart nginx >/dev/null 2>&1 || true
  fi
  exit "$status"
}
trap rollback ERR

# 2) 原子切换软链与版本变量。
ln -sfn "$RELEASE_DIR" "$DIRECT_ROOT/current"
set_env DIRECT_RELEASE_SHA "$SHA"

# 3) 校验合并后的 compose 配置。
docker compose -f "$BASE_COMPOSE" -f "$OVERRIDE" config -q

# 4) 先让 Go API 通过容器内健康检查，再重载 Nginx。单一应用入口避免
# 两套服务同时响应造成读取时序错乱。
docker compose -f "$BASE_COMPOSE" -f "$OVERRIDE" up -d --no-deps --force-recreate --pull never go-api
wait_for_go_api
# Nginx 在启动时会缓存上游容器 IP。Go 重建后必须让它重载一次，否则公网会
# 继续转发到已退出的旧 IP 并表现为 502；这不改变任何业务数据或依赖容器。
docker compose -f "$BASE_COMPOSE" restart nginx

# 5) 健康检查：Go 的公开运行态接口必须回报本次 SHA（最多等 90 秒）。
# /api/build-info 属于兼容诊断接口，需要登录；切换验收必须使用不需要
# 登录的 /api/runtime-build-info，避免把登录状态误判成部署失败。
ok=0
for _ in $(seq 1 30); do
  got="$(curl -fsS --max-time 5 http://127.0.0.1:3000/api/runtime-build-info 2>/dev/null \
    | python3 -c 'import json,sys;print(json.load(sys.stdin).get("git_sha",""))' 2>/dev/null || true)"
  if [ "$got" = "$SHA" ]; then ok=1; break; fi
  sleep 3
done
if [ "$ok" != 1 ]; then
  echo "STATUS=FAILED runtime-build-info did not converge to $SHA" >&2
  exit 7
fi

# 6) 外网再验一次，防止只在内网通。
ext="$(curl -fsS --max-time 8 "http://115.190.156.223:3000/api/runtime-build-info" 2>/dev/null \
  | python3 -c 'import json,sys;print(json.load(sys.stdin).get("git_sha",""))' 2>/dev/null || true)"
if [ "$ext" != "$SHA" ]; then
  echo "STATUS=FAILED external runtime-build-info mismatch: $ext != $SHA" >&2
  exit 8
fi

# 7) Go 已通过内网健康检查和公网精确 SHA 验收后，才清理仍残留的旧
# Node 容器。配置文件和镜像仍保留为明确的回滚资料，但默认运行态只
# 保留 Go，避免两个应用同时处理同一条业务链路。
docker compose -f "$BASE_COMPOSE" -f "$OVERRIDE" rm -sf v88-node >/dev/null 2>&1 || true

# 成功：把上一版记成 previous，方便手工一键回退。
if [ -n "$PREV_DIR" ] && [ -d "$PREV_DIR" ]; then
  ln -sfn "$PREV_DIR" "$DIRECT_ROOT/previous"
fi
trap - ERR
echo "STATUS=DEPLOYED_DIRECT release_sha=$SHA previous=${PREV_DIR:-none}"

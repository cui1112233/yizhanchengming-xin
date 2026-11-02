#!/usr/bin/env bash
# Task 6.1 稳定性压测:提交大量任务,采样 RSS(ps -o rss)与 GET /tasks/{id} 延迟(curl 循环,P50/P99)。
#
# 默认(mock 模式):编译并启动 `src-tauri/examples/soak_server.rs`(真实 Storage / Scheduler / runner / HTTP,
# 假豆包后端,不连豆包、不耗配额),提交 SOAK_TASKS 个图像任务,等全部结束,再空闲 SOAK_IDLE_SECS 秒,
# 输出汇总并 kill 掉 server。
#
#   tests/soak.sh                         # 100 个任务,空闲 60s
#   SOAK_TASKS=300 SOAK_IDLE_SECS=600 tests/soak.sh
#
# 外部模式(已在运行的 relay,例如真豆包):设置 SOAK_BASE / SOAK_TOKEN / SOAK_PID。
# **默认不提交任何任务**(真豆包每个任务都耗配额),只采样 RSS 与已有任务的 GET 延迟;
# 显式 SOAK_SUBMIT_REAL=1 才提交,且最多 10 个(SOAK_ACCOUNT / SOAK_MODEL 指定账号与模型)。
#
# 依赖:bash、curl、jq、ps、cargo(mock 模式)。全程自限时(SOAK_DEADLINE_SECS,默认 1800),退出时清理子进程。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TASKS="${SOAK_TASKS:-100}"
IDLE_SECS="${SOAK_IDLE_SECS:-60}"
DEADLINE_SECS="${SOAK_DEADLINE_SECS:-1800}"
GET_SAMPLES="${SOAK_GET_SAMPLES:-1000}"
CAPTURE_MS="${SOAK_CAPTURE_MS:-200}"
PORT="${SOAK_PORT:-18787}"
OUT="${SOAK_OUT:-$(mktemp -d "${TMPDIR:-/tmp}/doubao-relay-soak.XXXXXX")}"
ACCOUNT="${SOAK_ACCOUNT:-压测账号}"
MODEL="${SOAK_MODEL:-Seedream 5.0 Flash}"

mkdir -p "$OUT"
START_TS=$(date +%s)
SERVER_PID=""
SAMPLER_PID=""

cleanup() {
  [ -n "$SAMPLER_PID" ] && kill "$SAMPLER_PID" 2>/dev/null || true
  if [ -n "$SERVER_PID" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

log() { printf '[soak %4ss] %s\n' "$(( $(date +%s) - START_TS ))" "$*"; }
over_deadline() { [ $(( $(date +%s) - START_TS )) -ge "$DEADLINE_SECS" ]; }

if [ -n "${SOAK_BASE:-}" ]; then
  MODE=external
  BASE="$SOAK_BASE"
  TOKEN="${SOAK_TOKEN:?外部模式需要 SOAK_TOKEN}"
  PID="${SOAK_PID:?外部模式需要 SOAK_PID(relay 进程号,用于 ps -o rss)}"
  if [ "${SOAK_SUBMIT_REAL:-0}" = "1" ]; then
    [ "$TASKS" -gt 10 ] && TASKS=10
  else
    TASKS=0
  fi
else
  MODE=mock
  TOKEN="${SOAK_TOKEN:-soak-token-$$}"
  BASE="http://127.0.0.1:$PORT/api/v1"
  log "编译 soak_server(release)…"
  (cd "$ROOT" && cargo build --release --example soak_server -q)
  BIN="$ROOT/target/release/examples/soak_server"
  SOAK_DIR="$OUT/data" SOAK_PORT="$PORT" SOAK_TOKEN="$TOKEN" SOAK_CAPTURE_MS="$CAPTURE_MS" \
    "$BIN" >"$OUT/server.out" 2>&1 &
  SERVER_PID=$!
  PID=$SERVER_PID
  for _ in $(seq 1 100); do
    curl -fs -m 1 "$BASE/health" >/dev/null 2>&1 && break
    sleep 0.1
  done
  curl -fs -m 2 "$BASE/health" >/dev/null || { log "soak_server 未能启动:"; cat "$OUT/server.out"; exit 1; }
fi
log "模式 $MODE,base=$BASE,pid=$PID,输出目录 $OUT"

# relay 已退出时 ps 非零退出;`|| true` 防止 set -euo pipefail 下 RSS_END=$(rss_kb) 直接中止脚本。
rss_kb() { ps -o rss= -p "$PID" 2>/dev/null | tr -d ' ' || true; }

# RSS 采样(每秒一次):epoch_s,rss_kb
( while kill -0 "$PID" 2>/dev/null; do echo "$(date +%s),$(rss_kb)"; sleep 1; done ) >"$OUT/rss.csv" &
SAMPLER_PID=$!

auth=(-H "Authorization: Bearer $TOKEN")
RSS_START=$(rss_kb)

# ---- 提交 ----
: >"$OUT/task_ids.txt"
submitted=0
rejected_full=0
while [ "$submitted" -lt "$TASKS" ]; do
  over_deadline && { log "超过总时限,停止提交"; break; }
  body=$(jq -nc --arg a "$ACCOUNT" --arg m "$MODEL" --arg p "soak #$submitted 一只在海边看日落的橘猫" \
    '{account:$a,type:"image",model_name:$m,prompt:$p,ratio:"1:1"}')
  resp=$(curl -s -m 10 -w '\n%{http_code}' "${auth[@]}" -H 'Content-Type: application/json' \
    -d "$body" "$BASE/generate")
  code=$(printf '%s' "$resp" | tail -n1)
  json=$(printf '%s' "$resp" | sed '$d')
  if [ "$code" = "200" ] || [ "$code" = "201" ] || [ "$code" = "202" ]; then
    printf '%s\n' "$(printf '%s' "$json" | jq -r .task_id)" >>"$OUT/task_ids.txt"
    submitted=$((submitted + 1))
  elif [ "$code" = "429" ]; then
    rejected_full=$((rejected_full + 1))
    sleep 0.5
  else
    log "提交失败 HTTP $code:$json"
    exit 1
  fi
done
log "已提交 $submitted 个任务(queue_full 退避 $rejected_full 次)"

# ---- 等全部结束 ----
pending_left() {
  local n=0
  while read -r id; do
    [ -z "$id" ] && continue
    st=$(curl -s -m 5 "${auth[@]}" "$BASE/tasks/$id" | jq -r .status)
    case "$st" in completed|failed|cancelled) ;; *) n=$((n + 1)) ;; esac
  done <"$OUT/task_ids.txt"
  echo "$n"
}
if [ "$submitted" -gt 0 ]; then
  while true; do
    left=$(pending_left)
    [ "$left" -eq 0 ] && break
    over_deadline && { log "超过总时限,仍有 $left 个任务未结束"; break; }
    sleep 2
  done
fi
completed=0; failed=0; other=0
: >"$OUT/final_status.txt"
while read -r id; do
  [ -z "$id" ] && continue
  t=$(curl -s -m 5 "${auth[@]}" "$BASE/tasks/$id")
  st=$(printf '%s' "$t" | jq -r .status)
  echo "$id $st $(printf '%s' "$t" | jq -c '.error // empty')" >>"$OUT/final_status.txt"
  case "$st" in completed) completed=$((completed + 1)) ;; failed) failed=$((failed + 1)) ;; *) other=$((other + 1)) ;; esac
done <"$OUT/task_ids.txt"
RSS_AFTER_TASKS=$(rss_kb)
log "任务结束:completed=$completed failed=$failed other=$other,RSS ${RSS_AFTER_TASKS}KB"

# ---- GET /tasks/{id} 延迟 ----
ids_file="$OUT/task_ids.txt"
if [ ! -s "$ids_file" ]; then
  curl -s -m 5 "${auth[@]}" "$BASE/tasks?limit=50" | jq -r '.tasks[]? | (.task_id // .id) // empty' >"$ids_file" || true
fi
: >"$OUT/get_latency_ms.txt"
if [ -s "$ids_file" ]; then
  n_ids=$(grep -c . "$ids_file")
  i=0
  while [ "$i" -lt "$GET_SAMPLES" ]; do
    id=$(sed -n "$(( i % n_ids + 1 ))p" "$ids_file")
    curl -s -o /dev/null -m 5 -w '%{time_total}\n' "${auth[@]}" "$BASE/tasks/$id" |
      awk '{ printf "%.3f\n", $1 * 1000 }' >>"$OUT/get_latency_ms.txt"
    i=$((i + 1))
  done
fi
pct() { sort -n "$OUT/get_latency_ms.txt" | awk -v p="$1" '{a[NR]=$1} END { if (NR==0) {print "n/a"; exit} i=int(NR*p/100); if (i<1) i=1; if (i>NR) i=NR; print a[i] }'; }
P50=$(pct 50); P99=$(pct 99); PMAX=$(pct 100)
log "GET /tasks/{id} × $(wc -l <"$OUT/get_latency_ms.txt" | tr -d ' '):P50=${P50}ms P99=${P99}ms max=${PMAX}ms"

# ---- 空闲观察 ----
log "空闲观察 ${IDLE_SECS}s…"
idle_end=$(( $(date +%s) + IDLE_SECS ))
while [ "$(date +%s)" -lt "$idle_end" ] && ! over_deadline; do sleep 1; done
RSS_END=$(rss_kb)
RSS_PEAK=$(cut -d, -f2 "$OUT/rss.csv" | sort -n | tail -n1)
CPU_END=$(ps -o %cpu= -p "$PID" 2>/dev/null | tr -d ' ')

errors=0
if [ "$MODE" = mock ]; then
  errors=$({ grep -h '"level":"ERROR"' "$OUT"/data/logs/* 2>/dev/null || true; } | wc -l | tr -d ' ')
fi

cat <<EOF | tee "$OUT/summary.txt"
mode=$MODE tasks_submitted=$submitted completed=$completed failed=$failed other=$other queue_full_retries=$rejected_full
rss_kb start=$RSS_START after_tasks=$RSS_AFTER_TASKS peak=$RSS_PEAK end_after_idle=$RSS_END
cpu_percent_idle_end=$CPU_END
get_task_ms p50=$P50 p99=$P99 max=$PMAX
error_log_lines=$errors
elapsed_s=$(( $(date +%s) - START_TS ))
out=$OUT
EOF

# 验收阈值(spec 14 UC-50/UC-51/NF-01):RSS < 300MB、P99 < 500ms、mock 模式全部完成
status=0
[ "${RSS_PEAK:-0}" -lt 307200 ] || { log "FAIL:RSS 峰值超过 300MB"; status=1; }
if [ "$P99" != "n/a" ]; then
  awk -v p="$P99" 'BEGIN { exit !(p < 500) }' || { log "FAIL:P99 ≥ 500ms"; status=1; }
fi
if [ "$MODE" = mock ] && [ "$completed" -ne "$submitted" ]; then
  log "FAIL:有任务未成功完成(见 $OUT/final_status.txt)"; status=1
fi
exit "$status"

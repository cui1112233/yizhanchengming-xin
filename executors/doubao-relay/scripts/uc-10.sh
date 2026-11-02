#!/usr/bin/env bash
# UC-10 ~ UC-18 真机验收(docs/spec/14-acceptance.md)。前置:relay 已在运行(`cargo tauri dev`),豆包已连接。
#
#   scripts/uc-10.sh                 只跑不消耗配额的检查(UC-11 / UC-13 / UC-17)
#   scripts/uc-10.sh --generate      另外真实生成 1 张图(UC-10 / UC-14 / UC-18),并在它运行期间
#                                    同账号再提交 1 个任务、立即取消(UC-15;UC-16 只验证排队位置)
#
# 环境变量:BASE(默认 http://127.0.0.1:8787)、TOKEN(默认读 config.toml)、ACCOUNT(默认 /accounts 的 current)、
#           OUT(下载目录,默认 mktemp)、PROMPT、RATIO(默认 16:9)、MODEL(默认 Seedream 5.0 Pro)。
# 依赖:curl、python3(解析 JSON)。macOS 没有 timeout 命令,所有 curl 都带 -m。
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8787}"
API="$BASE/api/v1"
CONFIG="${CONFIG:-$HOME/Library/Application Support/doubao-relay/config.toml}"
GENERATE=0
[[ "${1:-}" == "--generate" ]] && GENERATE=1

if [[ -z "${TOKEN:-}" ]]; then
  TOKEN="$(sed -n 's/^bearer_token *= *"\(.*\)"/\1/p' "$CONFIG" | head -n1)"
fi
[[ -n "$TOKEN" ]] || { echo "读不到 bearer_token,请设置 TOKEN" >&2; exit 2; }
OUT="${OUT:-$(mktemp -d)}"
PASS=0
FAIL=0

json() { python3 -c 'import json,sys; d=json.load(sys.stdin); print(eval(sys.argv[1], {"d": d}))' "$1"; }
ok() { echo "  PASS  $*"; PASS=$((PASS + 1)); }
bad() { echo "  FAIL  $*"; FAIL=$((FAIL + 1)); }

# req METHOD PATH [BODY] → 设置 STATUS / BODY
req() {
  local method="$1" path="$2" data="${3:-}" tmp
  tmp="$(mktemp)"
  if [[ -n "$data" ]]; then
    STATUS="$(curl -sS -m 30 -o "$tmp" -w '%{http_code}' -X "$method" -H "Authorization: Bearer $TOKEN" \
      -H 'Content-Type: application/json' --data-binary "$data" "$API$path")"
  else
    STATUS="$(curl -sS -m 30 -o "$tmp" -w '%{http_code}' -X "$method" -H "Authorization: Bearer $TOKEN" "$API$path")"
  fi
  BODY="$(cat "$tmp")"
  rm -f "$tmp"
}

expect_error() { # 名称 期望状态 期望 code 方法 路径 [body]
  local name="$1" want_status="$2" want_code="$3"
  req "$4" "$5" "${6:-}"
  local code
  code="$(echo "$BODY" | json 'd["error"]["code"]' 2>/dev/null || echo '?')"
  if [[ "$STATUS" == "$want_status" && "$code" == "$want_code" ]]; then
    ok "$name → $STATUS $code"
  else
    bad "$name → $STATUS $code(期望 $want_status $want_code)body=$BODY"
  fi
}

echo "== 前置"
req GET /doubao/status
[[ "$STATUS" == 200 ]] || { echo "relay 不可用:$STATUS $BODY" >&2; exit 2; }
echo "  doubao: $(echo "$BODY" | json '(d["running"], d["cdp_connected"], d["ui_compat_check"])')"
if [[ -z "${ACCOUNT:-}" ]]; then
  req GET /accounts
  ACCOUNT="$(echo "$BODY" | json 'd["current"]')"
fi
echo "  account: $ACCOUNT"

image_body() { # prompt ratio [model]
  python3 -c 'import json,sys; print(json.dumps({"account": sys.argv[1], "type": "image", "model_name": sys.argv[4], "prompt": sys.argv[2], "ratio": sys.argv[3]}, ensure_ascii=False))' \
    "$ACCOUNT" "$1" "$2" "${3:-${MODEL:-Seedream 5.0 Pro}}"
}

echo "== UC-17 鉴权"
s="$(curl -sS -m 10 -o /dev/null -w '%{http_code}' "$API/accounts")"
[[ "$s" == 401 ]] && ok "无 token → 401" || bad "无 token → $s"
s="$(curl -sS -m 10 -o /dev/null -w '%{http_code}' -H 'Authorization: Bearer nope' "$API/accounts")"
[[ "$s" == 401 ]] && ok "错 token → 401" || bad "错 token → $s"

echo "== UC-11 账号不存在"
expect_error "account=不存在的账号" 404 account_not_found POST /generate \
  "$(python3 -c 'import json; print(json.dumps({"account":"不存在的账号zz","type":"image","model_name":"Seedream 5.0 Pro","prompt":"x","ratio":"1:1"}, ensure_ascii=False))')"

echo "== UC-13 参数校验(全部不入队)"
before="$(req GET '/tasks?limit=100'; echo "$BODY" | json 'len(d["tasks"])')"
expect_error "prompt 空" 400 request_missing_field POST /generate "$(image_body '' '1:1')"
expect_error "model 不在白名单" 400 request_invalid_enum POST /generate "$(image_body 'x' '1:1' 'Seedream 9')"
expect_error "ratio 不在白名单" 400 request_invalid_enum POST /generate "$(image_body 'x' '5:4')"
expect_error "图像 ratio=21:9(真机图像菜单没有)" 400 request_invalid_enum POST /generate "$(image_body 'x' '21:9')"
expect_error "prompt 超 10000 字" 400 request_prompt_too_long POST /generate \
  "$(image_body "$(python3 -c 'print("字"*10001)')" '1:1')"
expect_error "参考图 > 10 张" 400 request_too_many_references POST /generate \
  "$(python3 -c 'import json,sys; print(json.dumps({"account":sys.argv[1],"type":"image","model_name":"Seedream 5.0 Pro","prompt":"x","ratio":"1:1","reference_files":["01HQZ9ABCDEFGHJKMNPQRSTV%02d"%i for i in range(11)]}, ensure_ascii=False))' "$ACCOUNT")"
after="$(req GET '/tasks?limit=100'; echo "$BODY" | json 'len(d["tasks"])')"
[[ "$before" == "$after" ]] && ok "校验失败的请求没有入队($before → $after)" || bad "任务数 $before → $after"

if [[ "$GENERATE" != 1 ]]; then
  echo "== 跳过真实生成(加 --generate 才跑 UC-10/14/15/16/18)"
  echo "结果:PASS=$PASS FAIL=$FAIL"
  [[ "$FAIL" == 0 ]]
  exit
fi

echo "== UC-10 单次生图"
PROMPT="${PROMPT:-一只蓝色小鸟站在开花的树枝上,水彩风格}"
RATIO="${RATIO:-16:9}"
t0="$(date +%s)"
req POST /generate "$(image_body "$PROMPT" "$RATIO")"
if [[ "$STATUS" != 200 && "$STATUS" != 202 ]]; then
  bad "提交 → $STATUS $BODY"
  echo "结果:PASS=$PASS FAIL=$FAIL"
  exit 1
fi
ID="$(echo "$BODY" | json 'd["task_id"]')"
ok "提交 → $STATUS task_id=$ID queue_position=$(echo "$BODY" | json 'd.get("queue_position")')"

echo "== UC-15 / UC-16 同账号第二个任务排队后立即取消(不会发给豆包)"
# 第一个任务占着账号锁(per_account_max=1),第二个必然 pending;等第一个离开 pending 再提交,避免两者同时派发。
for _ in $(seq 1 50); do
  req GET "/tasks/$ID"
  st="$(echo "$BODY" | json 'd["status"]')"
  [[ "$st" != pending ]] && break
  sleep 0.2
done
if [[ "$st" == preparing || "$st" == generating || "$st" == waiting_user_confirm ]]; then
  req POST /generate "$(image_body '排队取消测试(不应发送)' '1:1')"
  ID2="$(echo "$BODY" | json 'd["task_id"]' 2>/dev/null || echo '')"
  qp="$(echo "$BODY" | json 'd.get("queue_position")' 2>/dev/null || echo '?')"
  if [[ -n "$ID2" ]]; then
    req DELETE "/tasks/$ID2"
    from="$(echo "$BODY" | json 'd.get("cancelled_from_status")' 2>/dev/null || echo '?')"
    if [[ "$STATUS" == 200 && "$from" == pending ]]; then
      ok "UC-15 pending 取消 → 200 cancelled_from_status=pending"
    else
      bad "UC-15 取消 → $STATUS $BODY"
    fi
    [[ "$qp" -ge 1 ]] 2>/dev/null && ok "UC-16 第二个任务 queue_position=$qp(第一个在跑,串行)" || bad "UC-16 queue_position=$qp"
    req GET "/tasks/$ID2"
    [[ "$(echo "$BODY" | json 'd["status"]')" == cancelled ]] && ok "第二个任务状态 cancelled" || bad "第二个任务 $BODY"
  else
    bad "第二个任务提交失败 $BODY"
  fi
else
  bad "第一个任务没有进入运行态($st),跳过 UC-15/16 以免第二个任务被派发"
fi

echo "== UC-14 轮询(stage 变化)"
last=""
deadline=$((t0 + 300))
while :; do
  req GET "/tasks/$ID"
  line="$(echo "$BODY" | json '"%s/%s %s" % (d["status"], d.get("stage"), d.get("message") or "")')"
  if [[ "$line" != "$last" ]]; then
    echo "  +$(( $(date +%s) - t0 ))s  $line"
    last="$line"
  fi
  st="$(echo "$BODY" | json 'd["status"]')"
  case "$st" in completed | failed | cancelled) break ;; esac
  if [[ "$(date +%s)" -gt "$deadline" ]]; then
    bad "5 分钟内未结束,最后状态 $line"
    break
  fi
  sleep 2
done
echo "$BODY" | python3 -m json.tool > "$OUT/task-$ID.json"
echo "  详情已存 $OUT/task-$ID.json"
echo "  stage_history: $(echo "$BODY" | json '" → ".join(e["stage"] for e in d["stage_history"])')"
elapsed=$(( $(date +%s) - t0 ))
if [[ "$st" == completed ]]; then
  [[ "$elapsed" -le 180 ]] && ok "UC-10 ${elapsed}s 内 completed(≤3 分钟)" || bad "UC-10 completed 但用了 ${elapsed}s(>3 分钟)"
  n="$(echo "$BODY" | json 'len(d["stage_history"])')"
  [[ "$n" -ge 3 ]] && ok "UC-14 stage_history 共 $n 条" || bad "UC-14 stage_history 只有 $n 条"
else
  bad "UC-10 最终 $st:$(echo "$BODY" | json 'd.get("error")')"
  echo "结果:PASS=$PASS FAIL=$FAIL"
  exit 1
fi

echo "== UC-18 结果下载"
mime="$(echo "$BODY" | json 'd["result"]["files"][0]["mime"]')"
bytes="$(echo "$BODY" | json 'd["result"]["files"][0]["bytes"]')"
count="$(echo "$BODY" | json 'len(d["result"]["files"])')"
echo "  结果文件 $count 个;watermark_source=$(echo "$BODY" | json 'd["result"]["watermark_source"]')"
hdr="$OUT/headers.txt"
file="$OUT/task-$ID-0"
s="$(curl -sS -m 60 -D "$hdr" -o "$file" -w '%{http_code}' -H "Authorization: Bearer $TOKEN" "$API/tasks/$ID/download/0")"
ctype="$(grep -i '^content-type:' "$hdr" | tr -d '\r' | awk '{print $2}')"
clen="$(grep -i '^content-length:' "$hdr" | tr -d '\r' | awk '{print $2}')"
disp="$(grep -i '^content-disposition:' "$hdr" | tr -d '\r' | cut -d' ' -f2-)"
size="$(wc -c < "$file" | tr -d ' ')"
[[ "$s" == 200 ]] && ok "download/0 → 200" || bad "download/0 → $s"
[[ "$ctype" == "$mime" ]] && ok "Content-Type=$ctype 与 result.files[0].mime 一致" || bad "Content-Type=$ctype mime=$mime"
[[ "$clen" == "$size" && "$size" == "$bytes" ]] && ok "Content-Length=$clen = 实际字节 = result bytes" || bad "Content-Length=$clen 实际=$size bytes=$bytes"
echo "  Content-Disposition: $disp"
magic="$(head -c 8 "$file" | xxd -p)"
case "$magic" in
  89504e470d0a1a0a) kind=PNG ;;
  ffd8ff*) kind=JPEG ;;
  52494646*) kind=WEBP? ;;
  *) kind="未知($magic)" ;;
esac
echo "  文件魔数:$kind;$(file -b "$file")"
if [[ "$kind" == PNG ]]; then
  ok "UC-10 download/0 是合法 PNG"
elif [[ "$kind" == JPEG || "$kind" == WEBP? ]]; then
  ok "UC-10 download/0 是合法图片($kind,mime 与内容一致;原图格式由豆包决定)"
else
  bad "download/0 不是图片"
fi
s="$(curl -sS -m 30 -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $TOKEN" -H 'Range: bytes=0-99' "$API/tasks/$ID/download/0")"
[[ "$s" == 206 ]] && ok "Range 请求 → 206" || bad "Range 请求 → $s"
expect_error "index 越界" 404 result_index_out_of_range GET "/tasks/$ID/download/$count"

echo "TASK_ID=$ID"
echo "OUT=$OUT"
echo "结果:PASS=$PASS FAIL=$FAIL"
[[ "$FAIL" == 0 ]]

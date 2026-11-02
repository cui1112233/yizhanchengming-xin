# 04 · HTTP API 规格

## 通用约定

### 基础 URL
`http://127.0.0.1:<port>/api/v1`

默认端口 `8787`,可在 `~/.../doubao-relay/config.toml` 修改。

### 认证
所有端点 MUST 带 `Authorization: Bearer <token>` header,token 在首次启动自动生成,在设置页可见。

未带 token 或 token 错误 → `401 Unauthorized`。

### 请求格式
POST/PUT 请求体 `Content-Type: application/json; charset=utf-8`,除 `/uploads` 用 multipart/form-data。

### 响应格式
统一 JSON。字段命名用 `snake_case`。时间用 Unix 毫秒(i64)。

### 错误响应

```json
{
  "error": {
    "code": "account_not_found",
    "message": "指定的账号 \"张三\" 未在当前豆包中找到",
    "details": {
      "requested": "张三",
      "available": ["李四", "王五"]
    }
  }
}
```

HTTP 状态码:
- `400` 请求格式/参数错
- `401` 未鉴权
- `403` 鉴权通过但无权
- `404` 资源不存在
- `409` 冲突(如目标账号正在被别的任务用)
- `429` 限流(队列满)
- `500` 服务端错误
- `503` 豆包未就绪

所有错误码定义见 **[11-error-codes.md](./11-error-codes.md)**。

---

## 端点总览

| Method | Path | 说明 |
|---|---|---|
| GET | `/health` | 健康检查 |
| GET | `/accounts` | 列出所有已登录账号 |
| POST | `/accounts/switch` | 切换到指定账号 |
| POST | `/uploads` | 上传参考图/视频 |
| GET | `/uploads/{id}` | 查询上传状态 |
| POST | `/generate` | 提交生成任务 |
| GET | `/tasks` | 列出任务(分页) |
| GET | `/tasks/{id}` | 查询任务详情 |
| DELETE | `/tasks/{id}` | 取消任务 |
| POST | `/tasks/{id}/confirm` | 手动确认(manual 模式) |
| GET | `/tasks/{id}/download/{index}` | 下载结果文件 |
| GET | `/doubao/status` | 豆包连接状态 |
| POST | `/doubao/reconnect` | 触发重连豆包 |

---

## 端点详细规格

### `GET /health`

**用途**:存活检查。不需要鉴权。

**响应 200**:
```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 3600
}
```

### `GET /accounts`

**用途**:列出豆包中所有已登录账号。

**响应 200**:
```json
{
  "current": "张三",
  "accounts": [
    { "name": "张三", "subscription": "标准套餐", "label": "张三 标准套餐", "is_current": true, "is_ambiguous": false },
    { "name": "李四", "subscription": "高级套餐", "label": "李四 高级套餐", "is_current": false, "is_ambiguous": false }
  ],
  "read_at": 1697433600000
}
```

**响应 503**:豆包未就绪 → `error.code = "doubao_not_ready"`。

### `POST /accounts/switch`

**请求**:
```json
{ "target": "李四" }
```

**响应 200**:
```json
{
  "ok": true,
  "previous": "张三",
  "current": "李四",
  "switched_at": 1697433600000
}
```

**响应 409**:`error.code = "account_busy"`,目标账号有进行中任务。
**响应 404**:`error.code = "account_not_found"`。

**行为**:切换过程阻塞该账号的所有新任务提交(以避免半切换态下跑任务)。

### `POST /uploads`

**用途**:上传参考图/视频,返回 upload_id 供 `/generate` 引用。

**请求**:`multipart/form-data`
- `file`: 文件内容
- `kind`(可选): `"image"` | `"video_reference"`,不传则按 MIME 自判

**响应 200**:
```json
{
  "upload_id": "01HQZ9...",
  "mime": "image/png",
  "bytes": 2456321,
  "expires_at": 1697437200000
}
```

**响应 400**:`error.code = "upload_too_large"` / `"upload_mime_not_supported"`。

**限制**:
- 单文件 ≤ 50 MB
- 支持 MIME:`image/png`, `image/jpeg`, `image/webp`, `video/mp4`
- 过期 1 小时

### `GET /uploads/{id}`

**用途**:查询 upload 元信息(或确认未过期)。

**响应 200**:同 POST 返回。
**响应 404**:上传不存在或已过期。

### `POST /generate`

**用途**:提交生图或生视频任务。

**请求**:
```json
{
  "account": "张三",
  "type": "image",
  "model_name": "Seedream 5.0 Pro",
  "prompt": "一只蓝色小鸟...",
  "ratio": "16:9",
  "reference_files": ["01HQZ9...", "01HQZ8..."],
  "confirmation_mode": "auto",
  "auto_use_membership": true,
  "callback_url": "https://my-service/doubao-webhook"
}
```

视频任务额外:
```json
{
  "type": "video",
  "model_name": "Seedance 2.5",
  "duration": 10,
  ...
}
```

**字段**:
| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `account` | string | ✅ | 账号 name,见 `/accounts` 返回 |
| `type` | enum | ✅ | `"image"` / `"video"` |
| `model_name` | string | ✅ | 见 F3.1/F3.2 白名单 |
| `prompt` | string | ✅ | 1-10000 字符 |
| `ratio` | string | ✅ | 见 F3.1 白名单 |
| `duration` | int | video 必填 | 4-15(Seedance 2.5 支持 4-30) |
| `reference_files` | string[] | 否 | upload_id 数组,≤10(Fast 视频 ≤11) |
| `confirmation_mode` | enum | 否 | `"auto"`(默) / `"manual"` |
| `auto_use_membership` | bool | 否 | 默 `true` |
| `callback_url` | url | 否 | 任务完成/失败时 POST 回调 |

**响应 200**:
```json
{
  "task_id": "01HQZ9...",
  "status": "pending",
  "queue_position": 2,
  "created_at": 1697433600000
}
```

**响应 400**:参数错(具体见 11-error-codes.md)。
**响应 404**:`account_not_found` / `upload_not_found`。
**响应 429**:`queue_full`。

### `GET /tasks`

**用途**:列出任务。

**查询参数**:
- `status`:过滤状态(`pending|preparing|generating|completed|failed|cancelled`),逗号分隔
- `account`:过滤账号
- `limit`:默认 20,最大 100
- `cursor`:分页游标

**响应 200**:
```json
{
  "tasks": [
    { "task_id": "01HQZ9...", "account": "张三", "type": "image", "status": "completed",
      "stage": null, "created_at": 1697433600000, "completed_at": 1697433700000 }
  ],
  "next_cursor": "01HQZ8..."
}
```

### `GET /tasks/{id}`

**用途**:查询任务详细。

**响应 200**(进行中):
```json
{
  "task_id": "01HQZ9...",
  "account": "张三",
  "type": "video",
  "status": "generating",
  "stage": "quota_checking",
  "message": "豆包确认是否使用会员配额",
  "progress": null,
  "created_at": 1697433600000,
  "started_at": 1697433610000,
  "stage_history": [
    { "stage": "pending", "ts": 1697433600000 },
    { "stage": "preparing", "ts": 1697433610000 },
    { "stage": "checking", "ts": 1697433620000, "message": "豆包确认参数" },
    { "stage": "quota_checking", "ts": 1697433625000, "message": "豆包确认是否使用会员配额" }
  ],
  "request": { ... 原始请求回显 ... }
}
```

**响应 200**(完成):
```json
{
  "task_id": "01HQZ9...",
  "status": "completed",
  "completed_at": 1697433700000,
  "result": {
    "files": [
      { "index": 0, "mime": "video/mp4", "width": 1920, "height": 1080,
        "duration_ms": 5000, "bytes": 24563210,
        "download_url": "/api/v1/tasks/01HQZ9.../download/0" }
    ],
    "watermark_source": "fallback_decrypted"
  },
  "stage_history": [ ... ]
}
```

**响应 200**(失败):
```json
{
  "task_id": "01HQZ9...",
  "status": "failed",
  "failed_at": 1697433700000,
  "error": {
    "code": "doubao_refused",
    "message": "豆包拒绝生成:账号配额不足"
  },
  "stage_history": [ ... ]
}
```

### `DELETE /tasks/{id}`

**用途**:取消任务。

**响应 200**:
```json
{ "ok": true, "cancelled_from_status": "pending" }
```

**响应 409**:`task_not_cancellable`,任务已完成或正在不可中断阶段。

### `POST /tasks/{id}/confirm`

**用途**:`confirmation_mode=manual` 下,外部触发"继续生成"。

**请求**(可选 body):
```json
{ "payload": "continue" }
```

**响应 200**:
```json
{ "ok": true, "new_stage": "generating" }
```

**响应 409**:`task_not_waiting_confirm`。

### `GET /tasks/{id}/download/{index}`

**用途**:下载结果文件。

**响应 200**:
- `Content-Type`: 对应 MIME
- `Content-Length`: 文件字节数
- `Content-Disposition: attachment; filename="task-01HQZ9...-0.png"`
- 支持 `Range` 请求

**响应 404**:`result_not_available` 或 `result_expired`。

### `GET /doubao/status`

**用途**:豆包连接健康状况。

**响应 200**:
```json
{
  "running": true,
  "cdp_connected": true,
  "cdp_port": 9707,
  "doubao_version_detected": "3.5.1",
  "ui_compat_check": "ok",
  "chat_targets": 1,
  "last_check_at": 1697433600000
}
```

### `POST /doubao/reconnect`

**用途**:触发重连豆包(杀进程 + 重启 + 等待就绪)。

**响应 200**:
```json
{
  "action": "restarted",
  "cdp_ready_at": 1697433620000
}
```

**响应 500**:`error.code = "doubao_launch_failed"`。

---

## Webhook 回调规格

当任务提交时带了 `callback_url`,任务完成/失败/取消时 MUST POST 到该 URL:

```http
POST <callback_url>
Content-Type: application/json
X-Doubao-Relay-Signature: hmac-sha256=<hex>
X-Doubao-Relay-Timestamp: 1697433700000

{
  "task_id": "01HQZ9...",
  "event": "task.completed",   // "task.failed" / "task.cancelled"
  "status": "completed",
  "task": { ... 同 GET /tasks/{id} 返回体 ... }
}
```

**签名**:HMAC-SHA256(bearer_token, timestamp + body),用 `X-Doubao-Relay-Signature` 发出。

**重试**:失败时按 1s / 10s / 60s 三次重试,仍失败则放弃并在任务上记 `callback_failed`。

**超时**:30 秒。

---

## 示例完整流程

```bash
TOKEN="xxxxx"
BASE="http://127.0.0.1:8787/api/v1"

# 1. 看看有哪些账号
curl -H "Authorization: Bearer $TOKEN" $BASE/accounts

# 2. 预上传参考图
UPLOAD_ID=$(curl -H "Authorization: Bearer $TOKEN" \
  -F "file=@ref.png" $BASE/uploads | jq -r .upload_id)

# 3. 提交任务
TASK_ID=$(curl -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"account\":\"张三\",\"type\":\"image\",\"model_name\":\"Seedream 5.0 Pro\",
       \"prompt\":\"蓝色小鸟\",\"ratio\":\"16:9\",\"reference_files\":[\"$UPLOAD_ID\"]}" \
  $BASE/generate | jq -r .task_id)

# 4. 轮询
while true; do
  STATUS=$(curl -s -H "Authorization: Bearer $TOKEN" $BASE/tasks/$TASK_ID | jq -r .status)
  echo "status: $STATUS"
  [ "$STATUS" = "completed" ] && break
  [ "$STATUS" = "failed" ] && break
  sleep 3
done

# 5. 下载
curl -H "Authorization: Bearer $TOKEN" \
  $BASE/tasks/$TASK_ID/download/0 -o result.png
```

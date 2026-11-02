# 11 · 错误码表

## 结构

所有 API/IPC 错误都有:

```json
{
  "error": {
    "code": "<snake_case_code>",
    "message": "<中文可读消息>",
    "details": { <可选,额外结构化字段> }
  }
}
```

**原则**:
- `code` 稳定,是契约
- `message` 可以改,面向用户展示
- `details` 补充排查信息,优先对调试友好

## 错误码分组

| 前缀/类别 | 范围 | 典型 HTTP 状态 |
|---|---|---|
| `request_*` | 请求格式/参数问题 | 400 |
| `auth_*` | 鉴权问题 | 401 / 403 |
| `account_*` | 账号相关 | 404 / 409 |
| `upload_*` | 上传相关 | 400 / 404 |
| `task_*` | 任务相关 | 404 / 409 / 429 |
| `doubao_*` | 豆包相关 | 503 |
| `generation_*` | 生成过程中失败 | — (在任务 error 里) |
| `system_*` | 系统/内部 | 500 |
| `platform_*` | 平台适配层问题 | 500 |

## 完整错误码表

### 请求类(400)

| code | message | 典型 details |
|---|---|---|
| `request_malformed` | 请求格式错误 | parse error |
| `request_missing_field` | 必填字段缺失 | `{ field: "account" }` |
| `request_invalid_enum` | 枚举值不合法 | `{ field, value, allowed }` |
| `request_invalid_range` | 数值超出范围 | `{ field, value, min, max }` |
| `request_prompt_too_long` | 提示词超过 10000 字 | `{ length }` |
| `request_too_many_references` | 参考图数量超限 | `{ count, max }` |
| `request_invalid_callback_url` | callback_url 不合法 | `{ url }` |

### 鉴权类(401/403)

| code | message | 典型场景 |
|---|---|---|
| `auth_missing_token` | 未提供 Bearer token | 请求没带 Authorization |
| `auth_invalid_token` | Token 无效 | token 不匹配 |
| `auth_token_revoked` | Token 已被撤销 | 用户在设置页重新生成了 |

### 账号类(404/409)

| code | message | 典型 details |
|---|---|---|
| `account_not_found` | 指定账号在豆包中未找到 | `{ requested, available: [...] }` |
| `account_not_logged_in` | 该账号未登录 | `{ requested }` |
| `account_ambiguous` | 账号名在豆包中有多个 | `{ requested, matches: [...] }` |
| `account_busy` | 目标账号有进行中任务 | `{ account, running_task_ids: [...] }` |
| `account_switch_failed` | 账号切换失败 | `{ reason }` |
| `account_switch_timeout` | 账号切换超时 | `{ elapsed_ms }` |
| `account_quota_exhausted` | 账号配额已用完 | `{ account }` |

### 上传类(400/404)

| code | message | 典型 details |
|---|---|---|
| `upload_too_large` | 文件超过 50MB 上限 | `{ bytes, max_bytes }` |
| `upload_mime_not_supported` | 文件类型不支持 | `{ mime, allowed }` |
| `upload_not_found` | 上传 ID 不存在 | `{ upload_id }` |
| `upload_expired` | 上传已过期 | `{ upload_id, expired_at }` |

### 任务类(404/409/429)

| code | message | 典型 details |
|---|---|---|
| `task_not_found` | 任务不存在 | `{ task_id }` |
| `task_not_cancellable` | 任务当前状态不可取消 | `{ current_status }` |
| `task_not_waiting_confirm` | 任务不在等待确认状态 | `{ current_status }` |
| `task_already_cancelled` | 任务已被取消 | — |
| `task_already_completed` | 任务已完成 | — |
| `queue_full` | 任务队列已满 | `{ depth, max }` |
| `result_not_available` | 结果文件不存在 | `{ task_id, index }` |
| `result_expired` | 结果文件已过期清理 | `{ task_id, expired_at }` |
| `result_index_out_of_range` | 结果 index 超范围 | `{ requested, available }` |

### 豆包类(503)

| code | message | 典型 details |
|---|---|---|
| `doubao_not_running` | 豆包未运行 | — |
| `doubao_not_logged_in` | 豆包未登录 | — |
| `doubao_not_ready` | 豆包未就绪(CDP 未连接) | — |
| `doubao_cdp_disconnected` | 豆包 CDP 连接已断 | — |
| `doubao_crashed` | 豆包崩溃 | `{ last_seen_at }` |
| `doubao_launch_failed` | 启动豆包失败 | `{ reason }` |
| `doubao_launch_timeout` | 启动豆包后 20 秒 CDP 仍未就绪 | — |
| `doubao_not_found` | 未找到豆包安装 | `{ searched_paths }` |
| `doubao_version_unknown` | 无法识别豆包版本 | — |
| `doubao_ui_changed` | 豆包 UI 结构变化,工具无法适配 | `{ failed_selectors }` |
| `doubao_cdp_port_occupied` | 9707 端口被别的程序占 | `{ owning_process }` |

### 生成失败类(写入 task.error.code)

| code | message | 典型场景 |
|---|---|---|
| `generation_preparing_failed` | 准备阶段失败 | 选模型、填 prompt 等 |
| `generation_preparing_timeout` | 准备阶段超时 | 60 秒内未完成 preparing |
| `generation_submit_failed` | 发送提交失败 | 发送按钮不可点等 |
| `generation_submit_identity_missing` | 未识别到提交身份 | submissionIdentity 识别失败 |
| `generation_task_timeout` | 任务总超时 | 图 5 min / 视频 20 min |
| `generation_rendering_not_executed` | 豆包只回"正在渲染"但 30 秒未开始 | — |
| `generation_doubao_refused` | 豆包拒绝生成 | `{ reason }` |
| `generation_user_interrupted` | 用户操作导致任务终止 | `{ interruption }` |
| `generation_result_not_captured` | 任务结束但未捕获到结果 | — |
| `generation_watermark_not_removed` | 视频无水印地址未获取 | — |
| `generation_download_failed` | 结果下载失败 | `{ url, status }` |
| `generation_video_decrypt_failed` | 视频 token 解密失败 | — |
| `generation_confirm_timeout` | manual 模式等待用户确认超时 | — |
| `generation_cancelled` | 任务被取消 | — |

### 系统类(500)

| code | message | 典型场景 |
|---|---|---|
| `system_internal_error` | 内部错误 | 未分类异常 |
| `system_database_error` | 数据库错误 | SQLite 读写失败 |
| `system_disk_full` | 磁盘空间不足 | 下载时 |
| `system_io_error` | 文件 IO 错误 | 读写文件失败 |
| `system_task_panicked` | 任务线程 panic | — |
| `system_shutdown` | 服务正在关闭,拒绝新请求 | — |
| `relay_restart` | 中转程序重启导致任务作废 | — |

### 平台类(500)

| code | message |
|---|---|
| `platform_registry_error` | 注册表读写失败(Windows) |
| `platform_process_enum_failed` | 进程枚举失败 |
| `platform_launch_blocked` | 系统阻止启动(UAC/Gatekeeper) |

### Webhook 类(记录到 webhook_log)

| code | message |
|---|---|
| `webhook_timeout` | 回调请求超时(> 30s) |
| `webhook_http_error` | 回调返回非 2xx |
| `webhook_connection_failed` | 回调连接失败 |
| `webhook_signature_denied` | 对方拒绝签名(保留) |
| `webhook_max_retries_exceeded` | 重试 3 次仍失败 |

## 错误消息文案规范

- **中文**,简洁明了
- **不带标点**(前端自己加)
- **不含技术名词**(除非是公认的,如 "CDP")
- **提供下一步建议**(如果能)

**好**:`豆包尚未登录,请在豆包客户端完成登录后重试`
**不好**:`doubao_not_logged_in: user_type field is empty in login state DOM probe`

## 示例

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

```json
{
  "error": {
    "code": "doubao_ui_changed",
    "message": "豆包 UI 结构发生变化,请升级 doubao-relay 以适配新版",
    "details": {
      "failed_selectors": ["[aria-label^=avatar]", "button[data-send]"],
      "detected_version": "3.6.0"
    }
  }
}
```

```json
{
  "error": {
    "code": "generation_doubao_refused",
    "message": "豆包拒绝生成:当前账号今日视频生成次数已达上限",
    "details": {
      "stage_at_refusal": "quota_checking",
      "doubao_message": "当前账号今日视频生成次数已达上限"
    }
  }
}
```

## HTTP 状态码对照

| 错误码前缀 | HTTP 状态 |
|---|---|
| `request_*` | 400 |
| `auth_*` | 401 |
| `account_not_found` / `upload_not_found` / `task_not_found` / `result_not_available` | 404 |
| `account_busy` / `task_not_cancellable` / `task_not_waiting_confirm` | 409 |
| `queue_full` | 429 |
| `doubao_*` | 503 |
| `system_*` / `platform_*` | 500 |
| `generation_*`(写在 task.error 里,HTTP 本身是 200) | — |

## IPC(Tauri 命令)错误

统一走 `Result<T, IpcError>`,JSON 结构同 API,JS 侧 `.catch` 可以拿到 `code`/`message`/`details`。

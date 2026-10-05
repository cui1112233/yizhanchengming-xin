# 一战晟铭新主线 Production Runbook

> 适用仓库：`cui1112233/yizhanchengming-xin`
>
> 本文只使用当前新主线已经存在的事实。Task 9.4 Redis Queue / Lease / Scheduler / Worker 尚未进入 `main`，相关章节明确标记为 **PENDING TASK9**，不会假装已有运行时能力。

## 0. 排障总原则

每次故障先收集三个事实：

1. 浏览器/API 返回的 `X-Request-ID`（前端失败对象为 `APIError.requestId`）。
2. 已存在的业务 ID：`batch_project_id`、`run_id`、`book_id`、`book_run_id`、`stage_run_id`、`production_job_id`、`production_task_id`、`provider_job_id`、`merge_job_id`、`publish_intent_id`。不存在的 ID 不要人为生成。
3. 故障发生时间窗口。

结构化日志默认输出到 Go 进程 stdout。日志中按字段搜索，不要只 grep `500`：

```bash
# 例：若部署由 systemd 管理，先替换为实际 unit 名称后执行
journalctl -u <actual-go-service-unit> --since '15 minutes ago' --no-pager | grep '"request_id":"req_'

# 例：若部署由 Docker 管理，先用 docker ps 确认实际容器名
docker logs --since 15m <actual-container-name> 2>&1 | grep '"request_id":"req_'
```

仓库当前没有定义最终 ECS 的 systemd unit 名或 Docker container 名，因此本文**不虚构服务名**。Task16 部署时必须把实际名字补回本文。

### 基础 HTTP 检查

在 ECS 本机执行（默认端口只有在未设置 `QIANTIE_GO_LISTEN_ADDR` 和 `PORT` 时才是 `8080`）：

```bash
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:8080/readyz
```

正常：

- `/healthz`：HTTP 200，表示 Go 进程可处理 HTTP。
- `/readyz`：HTTP 200，表示应用初始化完成且 MySQL 可用。
- 每个响应都有 `X-Request-ID`。

`/healthz` **不会**因为 Provider、Local Executor、TOS、ffmpeg、Redis 未配置/离线而失败。

### 受保护系统诊断

管理员登录后：

```text
GET /api/v1/diagnostics
```

匿名必须是 `401`；普通用户（即使有 `batch.view`）必须是 `403`。该接口只返回安全 DTO，包括 DB pool、Go runtime 聚合、Provider 安全状态、Executor 安全状态、Video/Merge 计数；不会返回 Secret、Token、Cookie、DSN、ciphertext、nonce、环境变量、stack dump。

当前 Runtime 区域会明确返回 `pending_task9_runtime`；不是 Redis 健康结论。

---

## 1. 网站打不开

**用户现象**：浏览器超时、连接失败、502/504、页面完全打不开。

**第一检查点**：

```bash
curl -i http://127.0.0.1:8080/healthz
```

**正常结果**：`200` + `X-Request-ID`。

**异常结果**：

- 连接拒绝：Go 服务没监听、端口错、进程退出。
- 本机 200 但公网失败：继续查反向代理、安全组、监听地址、端口映射；这不是业务 Provider 故障。
- `/healthz` 200、`/readyz` 503：进程活着，但 MySQL/初始化有问题。

**下一步**：用启动日志字段 `subsystem=http/mysql/auth/bootstrap`、`operation`、`safe_error` 检查，不要把 DSN 粘进工单或聊天。

**负责 subsystem**：`http` / `mysql` / deployment。

---

## 2. 登录失败 / 登录闪退

**用户现象**：登录失败、登录后立即掉回登录页、刷新后状态消失。

**第一检查点**：从失败请求获取 `request_id`，依次查看：

```text
/api/auth/login
/api/auth/current-user
/api/auth/refresh
```

**要查的事实**：

- Cookie 是否由浏览器实际保存（不要把 Cookie 值记录/截图到公开渠道）。
- `Secure` Cookie 与实际 HTTP/HTTPS 环境是否一致。
- `current-user 401 → refresh` 链路是否符合预期。
- 请求的 `request_id` 对应服务端 `subsystem=auth/http` 日志。

**正常结果**：登录 200；后续 current-user 能恢复同一用户；刷新后仍保持 session。

**异常结果**：401、403、CSRF/same-origin 拒绝、Cookie 因 HTTPS 配置未保存。

**下一步**：Auth 业务语义问题归 Task15/Task16；本 Runbook 只负责用 request ID 定位，不修改 Auth 状态机。

**负责 subsystem**：`auth`。

---

## 3. Novel Fetch / Intake 失败

**用户现象**：某本小说获取失败、批次 partial_failed、121 数据没回来。

**第一检查点**：记录 `request_id + intake_id + book_id/external book id`。

执行接口：

```text
POST /api/v1/intakes/{id}/execute
GET  /api/v1/intakes/{id}/books
```

**正常结果**：Intake 最终 `completed`，每本状态 fetched；日志出现 `intake execute started` → `intake completed`。

**异常结果**：`partial_failed/failed`、单本 `retryable_failed`、121 upstream 错误。

**下一步**：

1. 先按 `request_id` 找 `subsystem=intake`。
2. 再确认 121 请求对应的 platform/book ID。
3. 单本失败不要误判成整个 Intake 崩溃。

HTTP 不再把底层 DB/121 internal cause 原样返回给浏览器；浏览器只拿安全 `code/message/request_id`。

**负责 subsystem**：`intake` / `novel_fetch` / `provider121`。

---

## 4. Batch Run 不动

**用户现象**：BatchProject 已创建但 Run 长期不执行。

**第一检查点**：`batch_project_id + run_id`，确认创建日志 `batch project run created` 和 MySQL 中 Run 状态/run_at。

**当前正常边界**：项目/Run 可以持久化。

**PENDING TASK9**：真正的自动化到点 Scheduler、Redis Queue、Redis Lock、Worker Claim/Lease/Recovery 尚未进入当前 main。

因此现在如果 `run_at` 已到但没有 Worker 执行，**不能由 observability 分支自己 requeue/retry**；登记 Task9 blocker。

**负责 subsystem**：`batch_project`；Task9 合并后为 `runtime/redis/worker/scheduler`。

---

## 5. BookRun 卡住 / Generation 失败

**用户现象**：某本生成一直 running，或 Script/Hook/Director/Final Prompt 失败。

**第一检查点**：`request_id + batch_project_id + book_id + book_run_id + stage_run_id`。

**正常结果**：日志有：

```text
generation book started
generation stage state
generation book completed
```

并且 Stage 状态最终为 completed/skipped/failed 的明确事实。

**异常结果**：某个 Stage terminal failed、重试仍失败、BookRun 长时间 running。

**下一步**：按 `stage_run_id` 定位具体 Stage；不要只看 BatchProject 总状态。Retry 使用已有 Generation retry API，observability detector 本身不触发 retry。

**负责 subsystem**：`generation`。

---

## 6. matchAudio / 音频失败

**用户现象**：音频时长无法测量、matchAudio 时间轴校验失败。

**第一检查点**：`request_id + batch_project_id + book_id`，查看：

```text
GET/POST /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement
```

**正常结果**：日志 `audio measurement started` → `audio measured duration`，有 `duration_ms`。

**异常结果**：`audio_probe_unavailable`、timeline validation 失败。

**下一步**：确认真实音频资产可访问以及 probe 工具链；不要用前端传入秒数替代服务端权威测量。

**负责 subsystem**：`audio` / `generation`。

---

## 7. Video 一直 pending/running

**用户现象**：视频任务长时间 queued/running，没有结果。

**第一检查点**：记录：

```text
batch_project_id
book_id
production_job_id
production_task_id
provider
model
provider_job_id
```

管理员先看 `/api/v1/diagnostics` 的 Provider/Executor/Video counts；项目用户看已有 Provider Status / Project Video Status API。

**正常结果**：日志只在真实状态变化时出现，例如：

```text
video submitted
video state transition queued -> running
video state transition running -> succeeded
```

同状态 Poll 不产生 INFO spam。

**异常结果**：Provider status `auth_failed/unavailable/unconfigured`；任务超过诊断阈值仍 queued/running；本地 executor offline。

**下一步**：

1. Provider remote：查 provider/model/status/provider_job_id。
2. `doubao_local_executor`：再查 executor heartbeat/online。
3. 有 provider_job_id 但无变化：确认上游状态，不要本任务自动 cancel/retry。

**负责 subsystem**：`video` / `video_provider`。

---

## 8. Provider auth_failed

**用户现象**：视频提交/探测提示 Provider credential rejected/auth_failed。

**第一检查点**：Provider Status 安全视图：provider、model、configured、enabled、status。

**正常结果**：`available`。

**异常结果**：`auth_failed`。

**下一步**：管理员重新验证 Provider 配置。不要在日志、截图、diagnostics 中输出 Provider Secret、Authorization 或 upstream auth response body。

**负责 subsystem**：`video_provider`。

---

## 9. Local Executor offline

**用户现象**：豆包本地执行器任务 pending，但远端 Provider 无异常。

**第一检查点**：管理员 diagnostics 的 executor：

```text
executor ID
online/offline
last_heartbeat
provider/model
capabilities
```

**正常结果**：目标 model 对应 executor `online=true`，heartbeat 新鲜。

**异常结果**：offline / last heartbeat 超过阈值。

**下一步**：恢复执行器进程与网络；正常 heartbeat 不会每次写 INFO。Token/TokenHash 永远不应从 diagnostics 获取。

**PENDING TASK9**：executor lease claim/recovery 等待 Task9.4 公共 Runtime Adapter。

**负责 subsystem**：`local_executor`。

---

## 10. Merge 失败

**用户现象**：视频生成成功但合并失败。

**第一检查点**：`request_id + merge_job_id + merge_attempt_id + book_id`。

**正常结果**：日志 `merge queued` → `merge state` 且最终 succeeded。

**异常结果**：`ffmpeg unavailable`、输入视频未 ready、执行超时、TOS upload 失败。

**下一步**：

```bash
# 在实际 ECS 上确认 ffmpeg 是否存在；不会影响 /healthz
command -v ffmpeg
ffmpeg -version

# 确认磁盘空间
df -h
```

再检查输入 artifact 是否可读、TOS 配置是否启用。不要把 TOS Secret 打到日志。

**负责 subsystem**：`merge` / `ffmpeg` / `tos`。

---

## 11. Publishing 失败

**用户现象**：无法发布、权限拒绝、PublishIntent 创建失败。

**第一检查点**：`request_id + publish_intent_id + batch_project_id + publishing_account_id + user_id`。

**正常结果**：`publish intent created`；后续 Publishing Worker/adapter（如果部署）使用同一 intent/audit 事实。

**异常结果**：`publishing_permission_denied`、credential 未配置、ownership/capability 不满足。

**下一步**：先区分 permission/ownership 与 credential/provider 错误。Publishing Credential、ciphertext、nonce、AES key 不允许进入日志和 diagnostics。

**负责 subsystem**：`publishing`。

---

## 12. MySQL 异常

**用户现象**：`/readyz` 503，大量 API 失败。

**第一检查点**：

```bash
curl -i http://127.0.0.1:8080/readyz
```

管理员 diagnostics 进一步看：

```text
ready
open_connections
in_use
idle
wait_count
wait_duration_ns
```

**正常结果**：database.ready=true。

**异常结果**：Ping 失败、连接池等待快速增长、连接耗尽。

**下一步**：查 MySQL 服务/network/pool 压力；绝不要把 `QIANTIE_MYSQL_DSN` 输出到日志或工单。

**负责 subsystem**：`mysql`。

---

## 13. Redis 异常 — PENDING TASK9

Task9.4 尚未进入当前 main，所以现在 diagnostics 明确显示：

```text
runtime.status = pending_task9_runtime
runtime.redis  = pending_task9
```

当前 observability 分支**没有**自己创建 Redis client、Queue、Lease 或 Scheduler。

Task9 合并后需要只接公共事实：reachable、queue depth、active/stale leases、recovery count、claim conflicts、retry/terminal failures、worker/scheduler 状态，并重新决定 `runtime_ready` 语义。

---

## 14. 疑似卡住任务

当前检测只报告：

- Video queued/running 超过诊断阈值。
- Merge queued/running 超过诊断阈值。

检测不会：

- retry
- requeue
- fail
- cancel

真正 recovery policy 属于 Task9 Runtime / Task14 Runtime Adapter。

---

## 15. 日志字段速查

优先字段：

```text
request_id
user_id
subsystem
operation
error_code
batch_project_id
run_id
book_id
book_run_id
stage_run_id
production_job_id
production_task_id
provider_job_id
merge_job_id
publish_intent_id
idempotency_key
generation_request_id
provider
model
status
```

严格区分：

- `request_id`：HTTP correlation。
- `idempotency_key`：业务幂等事实。
- `generation_request_id`：Generation 已有业务 RequestID。

不要相互覆盖。

---

## 16. 日志保存 / Rotation

当前 Go API 写结构化 JSON 到 stdout，不自行维护 logs table/files。

最终 ECS supervisor 尚未在仓库 deployment docs 中固定，因此 Task16 部署时必须二选一落地并回填实际参数：

### systemd/journald

建议上限：

```text
SystemMaxUse=1G
RuntimeMaxUse=512M
MaxRetentionSec=14day
```

实际值应结合 ECS 磁盘调整；修改后用 `journalctl --disk-usage` 验证。

### Docker json-file

建议 daemon/container logging policy：

```json
{"max-size":"50m","max-file":"10"}
```

不要同时再让应用写无限增长的本地日志文件。

---

## 17. 未来告警条件（本阶段不接通知渠道）

建议后续接统一告警时覆盖：

- API `/readyz` 持续 down。
- MySQL unavailable。
- **Task9 合并后** Redis unavailable / queue depth 持续上升 / stale leases。
- stale Video/Merge jobs 持续增长。
- Local Executor fleet 从 online 变 offline。
- Provider `auth_failed`。
- Merge failure spike。
- ECS disk low。
- DB pool wait_count/wait_duration 异常增长。

---

## 18. pprof / Dump 安全边界

当前不需要对公网开放 `/debug/pprof`。

浏览器 diagnostics 只返回 goroutine count、HeapAlloc、HeapInuse、Sys、uptime 等聚合值。不要加入：

- heap dump
- goroutine stack dump
- env dump
- config dump

如未来确需 pprof，只允许 localhost/internal listener 或明确的高权限 internal/admin auth。

# Batch Factory V11/V12 视频 Provider 盘点（Task 9.1.4）

本文件记录旧 v88 / 公网 Batch Factory 的视频 Provider、模型映射、浏览器接口、Node -> Go 凭据同步、Go Provider Registry 与各 Adapter 的真实运行契约。

## 1. 总结

旧 v88 的视频生产不是一个单一 provider，而是至少四类运行路径：

| Provider | 默认/对应模型 | 运行方式 | API Key |
|---|---|---|---|
| `personal_api` | `yd2.0-mini` | Yadi 云 API | 需要 |
| `doubao_local_executor` | `doubao-seedance` | 本地配对执行器 | 不需要云 Key |
| `autodl_comfyui` | `minimax-h3-video` | AutoDL ComfyUI workflow | 需要 |
| `yfai_seedance` | `seedance-2-0-official` | YFAI Seedance 云 API | 需要 |

浏览器当前生产契约实际使用 `/api/batch-factory/v12`；V12 Node 网关再把大部分 provider 请求转到 `/api/batch-factory/v11` Go 服务。

浏览器不应直接获得或提交用户 API Key。旧 v88 的可信边界是：

`Browser -> authenticated Node -> signed Node-to-Go bridge -> Go provider registry/adapter -> upstream provider`

---

## 2. Provider 常量与模型映射

旧 Go `video_provider.go`：

- `personal_api`
- `doubao_local_executor`
- `autodl_comfyui`
- `yfai_seedance`

### personal_api

- provider aliases：`personal` / `personal_api` / `yd_video` / `yadi`
- 默认模型：`yd2.0-mini`
- catalog alias：`yd2-mini-video`
- 默认创建接口：`https://ydapi.yadiai.cn/openapi/v1/video/create`
- 默认任务接口：`https://ydapi.yadiai.cn/openapi/v1/video/tasks`
- 默认结果接口：`${tasksURL}/{id}/result`

### doubao_local_executor

- aliases：`doubao` / `doubao_local` / `doubao_local_executor` / `local-doubao-executor-video`
- 默认 provider model：`doubao-seedance`
- 不保存云 API Key
- 必须存在当前账号已配对且在线的本地视频执行器

### autodl_comfyui

- aliases：`h3` / `minimax_h3` / `autodl` / `autodl_comfyui` / `autodl_comfyui_video` / `minimax-h3-video`
- 模型：`minimax-h3-video`
- 无参考图 workflow：`minimax_h3_lightx2v_no_pic`
- 有参考图 workflow：`minimax_h3_lightx2v_v5_15s`
- 创建：`https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}`
- 结果：`https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}`
- 最多 9 张参考图

### yfai_seedance

- aliases：`yfai` / `yfai_seedance` / `seedance-2-0-official`
- 默认模型：`seedance-2-0-official`
- 默认 Base URL：`https://yf.token6688.com`
- 提交：`POST /v1/media/generate`
- 查询：`GET /v1/tasks/{taskId}`
- 时长：4–15 秒
- 无参考图：`text-to-video`
- 有参考图：`reference`

### 模型 -> Provider 绑定

旧 Go 以模型选择为权威来源：

- `minimax-h3-video` -> `autodl_comfyui`
- `yd2-mini-video` / `yd2.0-mini` -> `personal_api`
- `seedance-2-0-official` -> `yfai_seedance`
- `local-doubao-executor-video` / `doubao-seedance` -> `doubao_local_executor`

后续新主线必须继续保证“所选模型”和“运行 provider”一致，不能允许 H3 模型误走 personal_api。

---

## 3. 浏览器公开接口

旧前端 wrapper 文件虽然仍名为 `batchFactoryV11.js`，实际 `BASE` 已是：

`/api/batch-factory/v12`

因此浏览器层实际调用：

### Provider 配置与状态

- `PUT /api/batch-factory/v12/video-provider/config`
- `GET /api/batch-factory/v12/video-provider/status?provider={provider}`

### 视频生产

- `POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/production`
- `POST /api/batch-factory/v12/batches/{batchId}/production`
- `POST /api/batch-factory/v12/batches/{batchId}/production/cancel`
- `GET /api/batch-factory/v12/batches/{batchId}/status`
- `DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/tasks/{taskId}`

### 本地执行器

- `GET /api/shuihuo-production/local-executors`
- `POST /api/shuihuo-production/local-executors/pairings`

### 本地产物

生产媒体可能使用：

- `/api/shuihuo-production/local-executor-artifacts/{artifactId}`
- 或短期签名的 `/api/local-executor/v1/artifacts/{artifactId}?token=...`

---

## 4. Node -> Go Provider 配置同步

旧 Node 负责可信凭据解析，不信任浏览器直接传 secret。

Go 内部真实接口：

- `PUT /api/batch-factory/v11/video-provider/config`
- `GET /api/batch-factory/v11/video-provider/status?provider={provider}`

两者都要求 signed bridge identity。

### personal_api

Node：

1. 从请求/批量设置读取选中的 `videoModelId`。
2. 优先通过账号模型目录 `resolveRuntimeModel(kind=video)` 获取 credential。
3. 老账号兼容回退到个人中心 `getVideoApiKey(config, 'yd')`。
4. 浏览器不携带 secret。
5. Node 在可信 Node -> Go hop 中写入：
   - provider
   - model
   - apiKey
   - provider URL（如需要）

缺失 Key 时返回：

- `VIDEO_API_KEY_REQUIRED`
- 文案：请先在个人中心配置视频 API Key

### AutoDL H3

Node 优先：

1. 个人中心 H3 Key
2. 服务端 `QIANTIE_AUTODL_H3_API_KEY`
3. 服务端 `QIANTIE_H3_API_KEY`

同步失败：`H3_PROVIDER_SYNC_FAILED`。

### YFAI Seedance

模型目录中的 `seedance-2-0-official` credential 作为 API Key；Base URL 默认 `https://yf.token6688.com`。

### doubao_local_executor

不向 provider registry 写云 Key；ProductionService 直接依赖 LocalExecutorVideoAdapter 和当前账号在线执行器。

---

## 5. 什么时候会触发配置同步

Node 并非每个请求都同步 secret。

旧逻辑会在以下边界同步：

- provider config 保存
- provider status 读取
- 单书 production 提交
- 批量 production 提交
- `stages/video` 执行

对于本地 Doubao provider 不执行 personal/H3 secret sync。

这说明旧架构把 Go registry 当成“运行时凭据缓存”，而不是长期凭据事实源。

---

## 6. Go Provider Registry

旧 Go 使用：

`MemoryVideoProviderRegistry`

Key 由：

`owner + provider`

组成。

Registry 保存：

- Provider
- APIKey
- Model
- CreateURL
- TasksURL
- ResultURL

接口：

- `Put`
- `Resolve`
- `View`

`View` 只返回：

- provider
- model
- configured

不会把 API Key 返回浏览器。

### 重要限制

Registry 是内存态：

- Go 重启后配置消失
- 多实例若不共享状态，各实例配置不同步
- Node 因此需要在生产/状态边界重新同步

新主线后续不能把这一点误当成持久化 provider 配置。

---

## 7. Go HTTP Provider 接口

### PUT `/api/batch-factory/v11/video-provider/config`

输入：

- `provider`
- `apiKey`
- `model`
- `createUrl`
- `tasksUrl`
- `resultUrl`

输出仅：

- `provider`
- `model`
- `configured`

### GET `/api/batch-factory/v11/video-provider/status`

query：

- `provider`

正常情况下 provider 未配置时，Go 会返回 HTTP 200 的：

- `provider`
- `configured: false`

因此“未配置”与“Go 服务不可达”应当是两个不同状态。

### 生产提交

- `POST /api/batch-factory/v11/batches/{batchId}/books/{bookId}/production`
- `POST /api/batch-factory/v11/batches/{batchId}/production`

输入至少包含：

- `requestId`
- `provider`
- 可选 `compilationId`

### 状态

- `GET /api/batch-factory/v11/batches/{batchId}/status`

### 取消

- `POST /api/batch-factory/v11/batches/{batchId}/production/cancel`

只有实际 provider 接受取消后，task 才能进入 `cancelled`；不支持取消的 provider 不能虚假标记成功取消。

---

## 8. personal_api / Yadi Adapter

实现：`YadiVideoAdapter`

### 提交

请求字段：

- `model`
- `prompt`
- `image_urls`
- `duration`
- `aspect_ratio`
- `resolution`

Authorization：

`Bearer {APIKey}`

参考图：

- 固定加入一张默认 first-frame 占位图
- 用户参考图最多 3 张

若 provider 直接返回媒体 URL，可直接标记 `succeeded`。

否则保存 `providerTaskId` 并进入轮询。

### 轮询

先：

`GET {TasksURL}/{taskId}`

成功后再：

`GET {ResultURL with {id}}`

状态映射：

- queued/submitted/pending/processing/generating -> running
- failed/error/cancelled/canceled -> failed
- success/succeeded/completed/done -> succeeded

旧实现会把 HTTP 200 body 中的 quota/moderation/parameter 错误提取出来，不能让其变成无限 running。

---

## 9. doubao_local_executor Adapter

### 可用性前置条件

必须满足：

- LocalExecutor client 已注册
- 当前 owner 有在线视频执行器
- PublicBaseURL 可用，否则后续 merge/publish 无法访问产物

### 创建任务

内部 SourceTaskID：

`bf11:{batchId}:{bookId}:{videoId}`

换开头变体会附加：

`#v{openingVariantIndex}`

提交内容：

- model
- compiled prompt
- duration
- aspect ratio
- resolution
- reference image URLs

### 本地状态映射

- queued / leased / preparing / submitting -> queued
- accepted / generating / downloading / uploading -> running
- succeeded / success / completed -> succeeded
- failed -> failed
- cancelled / canceled -> cancelled

成功后必须有 artifact ID，否则视为错误。

---

## 10. AutoDL H3 Adapter

### 提交

无参考图：

`minimax_h3_lightx2v_no_pic`

有参考图：

`minimax_h3_lightx2v_v5_15s`

最多 9 张参考图。

Authorization 使用 AutoDL 约定的 raw Authorization value，而非强制 `Bearer`。

请求包含：

- prompt
- duration
- resolution
- ref_image_0 ... ref_image_N

### 轮询

通过 `{id}` result URL 查询。

完成响应必须含可验证媒体 URL，否则不能标记 succeeded。

---

## 11. YFAI Seedance Adapter

### 提交

`POST {BaseURL}/v1/media/generate`

Authorization：

`Bearer {APIKey}`

payload：

- model
- prompt
- params.mode
- duration
- resolution
- aspect_ratio
- quality=mini
- count=1
- return_last_frame=false

参考图存在时：

- mode = `reference`
- params.images = images

否则：

- mode = `text-to-video`

时长下限自动提高到 4 秒，超过 15 秒直接拒绝。

---

## 12. ProductionService Provider 选择

旧 Go 的 `ProductionService.resolveProvider()`：

1. `doubao_local_executor` -> LocalExecutorVideoAdapter
2. `autodl_comfyui` -> Registry -> AutoDLH3VideoAdapter
3. `yfai_seedance` -> Registry -> YFAISeedanceAdapter
4. `personal_api` -> Registry -> YadiVideoAdapter
5. personal registry 不可用时允许旧 server fallback Adapter

每次提交前会核对：

`selected videoModelId` 是否与 provider adapter model 属于同一 provider 能力。

不匹配时应返回冲突，而不是偷偷改模型。

---

## 13. 已知公网历史问题与本次结论

此前公网实际排查曾出现：

`GET /api/batch-factory/v11/video-provider/status?provider=personal_api`

返回 503，并显示类似：

`Batch Factory V11 Go service unavailable`

本次源码盘点确认：

- Go status handler 对“provider 未配置”本身应该返回 HTTP 200 + `configured:false`
- Node 的 `upstreamErrorMessage()` 对非 4xx 上游失败会统一显示 `Batch Factory V11 Go service unavailable`
- personal status 读取会触发一次 Node -> Go provider config sync（允许缺失 Key）

因此旧公网 503 不能解释成单纯“用户没配置 API Key”。更可能属于：

- Go 服务不可达
- Bridge 配置/签名问题
- Node -> Go provider config sync 失败
- Go production slice/provider service 未启用

具体公网当前属于哪一种，留到 Task 9.1.5 做 ECS 真实分类；本任务不伪造运行结论。

---

## 14. 新主线迁移原则

### 必须保留

1. 浏览器永远不接触 API Key。
2. owner/account 隔离。
3. provider/model 强绑定校验。
4. provider task ID 和轮询状态持久化。
5. provider 错误必须可见，不能无限 running。
6. local executor 要有明确在线/离线状态。
7. “未配置”和“服务不可用”必须是不同诊断。
8. 取消必须以 provider/执行器实际接受为准。

### 不应原样复制

1. 不应把内存 registry 当长期事实源。
2. 不应让浏览器决定 secret/provider endpoint。
3. 不应在 Go 重启后依赖用户重新打开某页面才能恢复 provider 配置。
4. 不应把所有上游错误压成一个 503 文案。
5. 不应允许保存模型与运行 provider 静默不一致。

### 新 Go 目标

后续 Task 14 实现时，推荐把 provider 运行契约收敛到 Go：

- provider registry/credential resolver 服务端化
- credential 由账号模型配置安全解析
- provider config/status 有明确诊断码
- ProductionJob/Task 持久化 provider 与 providerTaskId
- provider adapter 统一 Submit/Poll/Cancel 能力接口
- local/cloud provider 同一状态机

---

## 15. Task 9.1.4 验收范围

本盘点已覆盖：

- provider 枚举与 aliases
- provider/model 映射
- personal_api / yd2.0-mini
- Doubao local executor
- AutoDL H3
- YFAI Seedance
- 浏览器公开接口
- Node 凭据解析与同步
- Go provider registry
- Go config/status/production 接口
- provider Submit/Poll 行为
- local executor 状态与产物
- 历史公网 personal_api 503 的源码层解释边界
- 新主线迁移安全原则

# v88 Batch Factory V11/V12 后端接口盘点

任务：Task 9.1.2 — 列出旧 v88 Batch Factory V11/V12 所有历史 Go / Node 后端接口与真实调用拓扑。

来源基线：旧仓库 `cui1112233/-` 的 `v88` 分支。

## 结论先行

旧公网的 Batch Factory 已经不是“纯 V11”。它实际是一个三层兼容架构：

```text
React 浏览器
  ↓ /api/batch-factory/v12/*
Node V12 公网网关
  ├─ 少量 V12 原生聚合/适配接口
  ├─ 账号模型/API Key/系统预设/121 会话等可信注入
  └─ 大部分请求改写成 /api/batch-factory/v11/*
        ↓ signed bridge
Go V11 服务
  ├─ Batch/Book/Settings
  ├─ Director/Hook/Stages
  ├─ Prompt Compiler
  ├─ Production/VIDEO Provider
  ├─ Merge
  ├─ Assets
  └─ External Publish

Go V12 原生：仅 H3 kernel 4 个接口
```

旧前端文件仍叫 `batchFactoryV11.js`，但其 `BASE` 已经是 `/api/batch-factory/v12`。因此新主线迁移不能按文件名判断“V11 已废弃”，也不能把旧 Node 兼容层原样长期保留；应按业务职责收敛为新 Go API，同时保证账号级密钥不返回浏览器。

---

# 1. Express 真实挂载前缀

旧 `app.js` 明确挂载：

| 前缀 | Router | 说明 |
|---|---|---|
| `/api/batch-factory/v11` | `createBatchFactoryV11Router` | Node V11 可信网关 / Go 代理 |
| `/api/batch-factory/v11` | `createBatchFactoryV11ScheduleRouter` | 旧 Schedule Router |
| `/api/batch-factory/v12` | `createBatchFactoryV12Router` | 当前浏览器生产 API 主入口 |
| `/api/batch-factory` | `createBatchFactoryIntakeRouter` | 更早期 staging intake |
| `/api/batch-factory` | `createBatchFactoryRouter` | 更早期 batch factory 主路由 |
| `/api/batch-factory` | `createBatchFactoryProductionRouter` | 更早期 production bridge |
| `/api/shuihuo-production` | 多个 Router | Local Executor / 巨量素材 / 水货视频生产等依赖 |

Node V11 Router 使用 `apiAuth`；Node → Go 使用 signed bridge。Go Router 对 `/api/batch-factory/v11/` 与 `/api/batch-factory/v12/` 都使用 BridgeAuth。

---

# 2. 浏览器当前公共契约：V12

旧前端 `frontend/src/shared/api/batchFactoryV11.js` 的模块名只是兼容名，真实 `BASE`：

```text
/api/batch-factory/v12
```

## 2.1 Batch / Intake / Source

| Method | Public path | 旧实现归属 |
|---|---|---|
| GET | `/api/batch-factory/v12/capabilities` | V12 → V11 Go capabilities |
| POST | `/api/batch-factory/v12/intakes/novel-fetch` | V12 rewrite → V11 Go |
| POST | `/api/batch-factory/v12/intakes/manual` | V12 rewrite → V11 Go |
| GET | `/api/batch-factory/v12/intakes/:intakeId` | V12 rewrite → V11 Go |
| POST | `/api/batch-factory/v12/intakes/:intakeId/batches` | V12；巨量场景 Node 原生适配，否则 rewrite → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/intakes/:intakeId/books` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/batches` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/batches/summary` | Node V12 原生聚合 → V11 `summary-index` |
| POST | `/api/batch-factory/v12/batches` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/batches/:batchId` | rewrite → V11 Go |
| DELETE | `/api/batch-factory/v12/batches/:batchId/books/:bookId` | Node V12 原生 deletion wrapper → V11 Go |
| DELETE | `/api/batch-factory/v12/batches/:batchId` | Node V12 原生 deletion wrapper → V11 Go |
| POST | `/api/batch-factory/v12/fetch-originals` | Node V12 原生，调用 121 workshop |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/fetch-original` | Node V12 原生，补正文后写回 V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/classify-fetched-metadata` | Node V12 原生，逐书分类，失败隔离 |
| PUT | `/api/batch-factory/v12/batches/:batchId/books/:bookId/source` | rewrite → V11 Go |
| PUT | `/api/batch-factory/v12/batches/:batchId/books/:bookId/metadata` | rewrite → V11 Go |

## 2.2 Settings / Config / Prompt

| Method | Public path | 旧实现归属 |
|---|---|---|
| PUT | `/api/batch-factory/v12/batches/:batchId/settings` | Node V12 校验模型后 → V11 Go |
| PUT | `/api/batch-factory/v12/batches/:batchId/books/:bookId/override` | Node V12 校验模型后 → V11 Go |
| PUT | `/api/batch-factory/v12/batches/:batchId/books/:bookId/videos/:videoId/override` | rewrite → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/change-impact` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/config-versions` | rewrite → V11 Go |
| POST | `/api/batch-factory/v12/config-versions` | rewrite → V11 Go |
| PUT | `/api/batch-factory/v12/config-versions/:versionId` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/prompts?kind=...` | rewrite → V11 Go |
| POST | `/api/batch-factory/v12/prompts` | rewrite → V11 Go |
| GET | `/api/batch-factory/v12/drafts?...` | rewrite → V11 Go |
| PUT | `/api/batch-factory/v12/drafts` | rewrite → V11 Go |

Batch Factory 还依赖非 V12 域接口：

- `GET /api/presets?module=batch-factory`
- `GET /api/script-constraint-prompts?category=...`
- `POST /api/script-constraint-prompts`

这些属于系统预设/个人约束配置，不应在新主线中被误删。

## 2.3 Assets / Images

| Method | Public path | 旧实现归属 |
|---|---|---|
| GET | `/api/batch-factory/v12/batches/:batchId/books/:bookId/assets` | rewrite → V11 Go |
| POST | same | rewrite → V11 Go |
| PATCH | `.../assets/:assetId` | rewrite → V11 Go |
| GET | `.../assets/:assetId/images` | rewrite → V11 Go |
| POST | same | rewrite → V11 Go，登记 provider 结果 |
| POST | `.../assets/:assetId/images/upload` | rewrite → V11 Go，本地图片上传 |
| PUT | `.../assets/:assetId/images/:imageId/primary` | rewrite → V11 Go |
| GET | `.../assets/:assetId/images/:imageId/content` | rewrite → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/assets/images/generate` | Node V11 gateway 拦截：解析账号图片模型凭据并调用模型，再持久化图片 |

旧前端还声明 `POST /api/batch-factory/v12/image-generation` 供历史 V11 预览页兼容使用；在本轮已检查的 Go Router 中没有对应独立 Go handler，Node 主路由中也未确认到稳定的专用注册。迁移时将其视为 **legacy compatibility / 待验证契约**，不可据此新增重复图片生成 API。

## 2.4 Director / Hook / Stage

| Method | Public path | 旧实现归属 |
|---|---|---|
| POST | `.../books/:bookId/working-front/viral` | Node 注入文本模型 → V11 Go |
| POST | `.../books/:bookId/hook` | Node 注入文本模型 → V11 Go |
| POST | `.../books/:bookId/opening-variants` | V11 Go；Node 可注入 opening preset |
| POST | `.../books/:bookId/hooks/:hookId/approve` | V11 Go |
| POST | `.../books/:bookId/director` | Node 注入文本模型/智能统一 → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/director` | Node 注入文本模型 → V11 Go 批量执行 |
| GET | `.../books/:bookId/stages` | V11 Go |
| POST | `.../books/:bookId/stages/:stage` | Node 根据 stage 注入文本/视频 provider → V11 Go |
| POST | `.../books/:bookId/stages/retry` | Node 重新注入可信运行配置 → V11 Go |
| POST | `.../books/:bookId/smart-unified/refresh` | Node-owned：调用账号文本模型并写回 V11 Go |

Node 对单书 `stages/director` 还有额外行为：Director 成功后，根据单书换开头设置自动尝试 opening variants；opening 失败不会回滚已成功的 Director。

## 2.5 H3：真正原生 V12 Go API

仅以下 4 个路径是旧 Go Router 注册的 V12 原生 H3 kernel：

| Method | Path |
|---|---|
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/h3/director` |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/h3/audio-measurement` |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/h3/compile` |
| GET | `/api/batch-factory/v12/batches/:batchId/books/:bookId/h3/trace?compilationId=...` |

其余 V12 路径主要由 Node rewrite 到 V11 Go。

## 2.6 Final Prompt Compiler

| Method | Public path | 旧实现 |
|---|---|---|
| GET | `.../videos/:videoId/effective-settings` | V11 Go PromptCompilerService |
| GET | `.../videos/:videoId/final-prompt` | V11 Go PromptCompilerService |

## 2.7 Production / VIDEO Provider

| Method | Public path | 旧实现 |
|---|---|---|
| PUT | `/api/batch-factory/v12/video-provider/config` | Node 注入账号 API Key → V11 Go registry |
| GET | `/api/batch-factory/v12/video-provider/status?provider=...` | Node 允许缺配置读取 → V11 Go |
| POST | `.../books/:bookId/production` | Node provider sync/校验 → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/production` | Node provider sync/校验 → V11 Go |
| POST | `/api/batch-factory/v12/batches/:batchId/production/cancel` | V11 Go |
| DELETE | `.../videos/:videoId/tasks/:taskId` | V11 Go，删除候选任务 |
| GET | `/api/batch-factory/v12/batches/:batchId/status` | V11 Go |
| POST multipart | `.../videos/:videoId/upload` | V11 Go → 本地暂存 → TOS → 登记任务 |

Node provider 名称：

- `personal_api`
- `doubao_local_executor`
- `autodl_comfyui`
- `yfai_seedance`

旧个人 API 默认模型：`yd2.0-mini`。

Batch Factory 还依赖：

- `GET /api/shuihuo-production/local-executors`
- `POST /api/shuihuo-production/local-executors/pairings`
- `/api/shuihuo-production/local-executor-artifacts/...`

provider 细节在 Task 9.1.4 再专项盘点。

## 2.8 Automation / Schedule

Node 自己持有自动化控制器和预设：

| Method | Public path |
|---|---|
| GET | `/api/batch-factory/v12/automation-presets` |
| POST | same |
| PUT | `/api/batch-factory/v12/automation-presets/:presetId` |
| DELETE | same |
| POST | `/api/batch-factory/v12/batches/:batchId/automation-presets` |
| GET | `/api/batch-factory/v12/batches/:batchId/automation` |
| POST | `.../automation/repair-legacy-overrides` |
| POST | `.../automation/start` |
| POST | `.../automation/pause` |
| POST | `.../automation/resume` |
| POST | `.../automation/retry` |
| POST | `.../automation/cancel` |

旧独立 Schedule Router：

| Method | Path |
|---|---|
| GET | `/api/batch-factory/v11/schedules` |
| POST | `/api/batch-factory/v11/schedules` |
| PATCH | `/api/batch-factory/v11/schedules/:id` |
| DELETE | `/api/batch-factory/v11/schedules/:id` |

注意：旧前端通过 V12 wrapper 调用 schedules，而 V12 fallback 会 rewrite 到 V11 Router；新主线后续不应继续同时保留 Schedule 文件态和新 MySQL `Run.run_at` 两套事实源。

## 2.9 Runtime Summary

| Method | Public path | 旧实现 |
|---|---|---|
| GET | `/api/batch-factory/v12/batches/:batchId/runtime-summary` | Node V12 聚合接口 |

它一次聚合：

- V11 `runtime-index`
- automation controller status
- V11 production status
- V11 merge status
- 每本书 stage summary

旧 Node 为该聚合增加约 15 秒短缓存，避免大批次页面轮询把 Go/MySQL 打满。

## 2.10 Merge

| Method | Public path | 旧实现 |
|---|---|---|
| POST | `/api/batch-factory/v12/batches/:batchId/merge` | V11 Go |
| POST | `.../books/:bookId/merge` | V11 Go |
| GET | `/api/batch-factory/v12/batches/:batchId/merge-status` | V11 Go |
| GET | `/api/batch-factory/v12/batches/:batchId/merge-media/:artifactId` | V11 Go，本地/TOS 受保护媒体 |
| GET | `/api/batch-factory/v12/batches/:batchId/merge-cover/:artifactId` | V11 Go，必要时 ffmpeg 抽帧生成封面 |
| POST | `/api/batch-factory/v12/batches/:batchId/merge-artifacts/migrate-to-tos` | V11 Go |
| POST multipart | `.../books/:bookId/merge/upload` | V11 Go，人工合成 MP4 上传 |

旧前端 wrapper 还声明 `GET .../books/:bookId/merge-status?requestId=...`。本轮检查的 Go `registerMergeRoutes` 没有找到这个单书专用 GET handler，因此标记为 **frontend-declared / handler 未确认**；后续不得盲目照抄。

## 2.11 Publish

Go external generic publish：

| Method | Path |
|---|---|
| GET | `/api/batch-factory/v12/publish/:provider/credential` |
| PUT | same |
| POST | `/api/batch-factory/v12/publish/:provider/intents` |
| POST | `/api/batch-factory/v12/publish/:provider/intents/:intentId/confirm` |
| POST | `/api/batch-factory/v12/publish/:provider/intents/:intentId/submit` |
| GET | `/api/batch-factory/v12/publish/:provider/intents/:intentId/audits` |

Node-owned 121 发布桥：

| Method | Path |
|---|---|
| GET | `/api/batch-factory/v12/publish-121/organizations` |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/classify-publish-metadata` |
| POST | `/api/batch-factory/v12/batches/:batchId/books/:bookId/publish-121` |

Node 负责使用已保存的 121 会话/Cookie、组织目录、文本/MP4 回读与提交，浏览器不直接持有这些可信凭据。

---

# 3. Go V11 实际注册路由总表

以下由旧 Go `httpapi.NewRouter` 按 slice 能力注册，并通过 BridgeAuth 保护。

## 3.1 Capability

- `GET /api/batch-factory/v11/capabilities`

能力键：`batch.read`、`batch.create`、`settings.edit`、`snapshot.read`、`override.edit`、`director.run`、`hook.review`、`working-front.viral`、`compiler.preview`、`production.submit`、`production.cancel`、`merge.run`、`publish.121`、`publish.yadi`。

## 3.2 Slice 1：Batch / Intake / Settings

- `POST /api/batch-factory/v11/intakes/manual`
- `POST /api/batch-factory/v11/intakes/novel-fetch`
- `GET /api/batch-factory/v11/intakes/{intakeId}`
- `POST /api/batch-factory/v11/intakes/{intakeId}/batches`
- `POST /api/batch-factory/v11/batches/{batchId}/intakes/{intakeId}/books`
- `GET /api/batch-factory/v11/batches`
- `GET /api/batch-factory/v11/batches/summary-index`
- `POST /api/batch-factory/v11/batches`
- `GET /api/batch-factory/v11/batches/recovery-index`
- `GET /api/batch-factory/v11/batches/{batchId}/runtime-index`
- `GET /api/batch-factory/v11/batches/{batchId}`
- `DELETE /api/batch-factory/v11/batches/{batchId}/books/{bookId}`
- `DELETE /api/batch-factory/v11/batches/{batchId}`
- `PUT /api/batch-factory/v11/batches/{batchId}/settings`
- `PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/metadata`
- `PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/source`
- `PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/override`
- `PUT /api/batch-factory/v11/batches/{batchId}/books/{bookId}/videos/{videoId}/override`
- `POST /api/batch-factory/v11/batches/{batchId}/change-impact`
- `GET /api/batch-factory/v11/config-versions`
- `POST /api/batch-factory/v11/config-versions`
- `PUT /api/batch-factory/v11/config-versions/{versionId}`
- `GET /api/batch-factory/v11/prompts`
- `POST /api/batch-factory/v11/prompts`
- `GET /api/batch-factory/v11/drafts`
- `PUT /api/batch-factory/v11/drafts`

同一 slice 还注册 Book Assets 路由，见 2.3。

## 3.3 Slice 2：Director + Stage

Director：

- `POST .../working-front/viral`
- `POST .../hook`
- `POST .../opening-variants`
- `POST .../hooks/{hookId}/approve`
- `POST .../books/{bookId}/director`
- `POST /api/batch-factory/v11/batches/{batchId}/director`

Stage：

- `GET .../books/{bookId}/stages`
- `POST .../books/{bookId}/stages/{stage}`
- `POST .../books/{bookId}/stages/retry`

## 3.4 Slice 3：Compiler + V12 H3

Compiler：

- `GET .../videos/{videoId}/effective-settings`
- `GET .../videos/{videoId}/final-prompt`

原生 V12 H3：见 2.5。

## 3.5 Slice 4：Production + Video Upload

- `PUT /api/batch-factory/v11/video-provider/config`
- `GET /api/batch-factory/v11/video-provider/status`
- `POST .../books/{bookId}/production`
- `POST .../batches/{batchId}/production`
- `POST .../batches/{batchId}/production/cancel`
- `DELETE .../videos/{videoId}/tasks/{taskId}`
- `GET .../batches/{batchId}/status`
- `POST multipart .../videos/{videoId}/upload`（仅当 MergeOutput 支持上传）
- `POST multipart .../books/{bookId}/merge/upload`（同上且 MergeService 可用）

## 3.6 Slice 5：Merge

- `POST .../merge-artifacts/migrate-to-tos`
- `POST .../batches/{batchId}/merge`
- `POST .../books/{bookId}/merge`
- `GET .../batches/{batchId}/merge-status`
- `GET .../batches/{batchId}/merge-media/{artifactId}`
- `GET .../batches/{batchId}/merge-cover/{artifactId}`

## 3.7 Slice 6：External Publish

- `GET /api/batch-factory/v11/publish/{provider}/credential`
- `PUT same`
- `POST /publish/{provider}/intents`
- `POST /publish/{provider}/intents/{intentId}/confirm`
- `POST /publish/{provider}/intents/{intentId}/submit`
- `GET /publish/{provider}/intents/{intentId}/audits`

---

# 4. 更早期 `/api/batch-factory` 历史接口

这些接口属于 V11/V12 前一代实现，仍在旧 `app.js` 中挂载，因此迁移盘点必须记录，但不代表新主线应继续保留双轨。

## 4.1 Intake staging router

- `POST /api/batch-factory/intakes/novel-fetch`
- `GET /api/batch-factory/intakes/:intakeId`

它只创建“小说获取 → 批量工厂”的交接单，不执行 AI。

## 4.2 Production staging router

- `POST /api/batch-factory/batches/:batchId/items/:itemId/generate`
- `POST /api/batch-factory/batches/:batchId/generate`

它编译旧 director storyboard 的视频 prompt 后，通过 Shuihuo bridge 提交 `/api/shuihuo-production/batch-factory/import-videos`，并逐 item 隔离失败。

新主线后续应把仍有价值的行为语义迁进统一 Go V11/V12 replacement，而不是同时维护新旧两套 Batch 数据模型。

---

# 5. Node V11 网关承担的非 CRUD 职责

这些职责解释了为什么不能简单把旧 Node Router 删除后让浏览器直接请求旧 Go：

1. **账号级模型解析**：浏览器只传 modelId；Node 读取账号可用模型和 credential。
2. **密钥隔离**：文本/图片/视频 API Key 只在 Node → Go / Node → Provider 可信跳转中出现，不回传浏览器。
3. **系统预设正文注入**：浏览器保存 preset identity，Node 在执行前解析已发布正文。
4. **智能统一视觉分析**：Node 调用账号文本模型，结果写回 Go。
5. **图片生成**：Node 解析图片模型 credential，调用 provider 后登记到 Go Book Asset。
6. **视频 provider 同步**：`personal_api` / H3 等配置在可信服务端同步到 Go registry。
7. **121 会话与发布**：Cookie、组织目录、上传、回读由服务端持有。
8. **自动化控制器**：start/pause/resume/retry/cancel 和旧文件态 schedule/recovery。
9. **V12 → V11 兼容改写**：当前公网的 V12 public contract 大量复用 V11 Go 数据。

新主线目标应是：**这些可信职责继续留在服务端，但优先直接收敛到 Go 服务内部，而不是保留 Node 作为永久生产业务层。**

---

# 6. 已发现的历史契约不一致 / 风险

## 6.1 V12 名称与 V11 实体不一致

浏览器叫 V12，绝大多数 durable CRUD 和生产状态仍来自 V11 Go。迁移时必须以“功能职责 + MySQL 数据”为准，不以 URL 版本名决定数据模型。

## 6.2 V11 mutation 已有只读升级语义

旧 Node V12 文件包含 `rejectLegacyV11Mutations`，返回：

- HTTP 410
- `BATCH_FACTORY_V11_READ_ONLY`
- upgradePath `/api/batch-factory/v12`

说明旧系统自身已经尝试停止浏览器直接写 V11。

## 6.3 `image-generation` 历史兼容契约

前端 wrapper 明确保留 `POST /api/batch-factory/v12/image-generation` 给 legacy V11 page，但当前已确认的 Go Router 没有同名独立 route。新主线应优先迁移 **book-scoped asset image generation**，不要为兼容死 UI 先新增重复接口。

## 6.4 单书 merge-status wrapper 未找到对应独立 Go handler

前端存在：

`GET /api/batch-factory/v12/batches/:batchId/books/:bookId/merge-status?requestId=...`

而已确认 Go `registerMergeRoutes` 只有 batch-level `GET .../batches/:batchId/merge-status`。该 wrapper 记为待验证历史契约。

## 6.5 自动化事实源重复

旧系统同时存在：

- Node file-backed schedule/controller
- Go Batch/Book/Stage/Production/Merge 状态
- 新主线现在已有 MySQL `Run.run_at`

新主线 Task 9.4 必须统一为 MySQL + Redis Worker，不再叠加第三套 schedule 事实源。

---

# 7. Task 9.1.2 迁移结论

新主线后续建议按职责归并成 Go 模块：

- `batch`：Batch / Book / Intake / Source / Metadata
- `settings`：Batch / Book / Video overrides、config version、change impact
- `assets`：角色/场景资产 + 图片版本
- `director`：Hook / Director / Opening / H3 director
- `stage`：阶段状态机、重试
- `compiler`：effective settings + final prompt
- `production`：provider + VIDEO job
- `merge`：merge job、受保护媒体、TOS
- `automation`：Run + Redis queue/lock/worker
- `publish`：121 / external provider
- `runtime`：聚合状态读取

浏览器继续只传非敏感 ID/参数；API Key、系统 preset body、121 Cookie 必须保持服务端隔离。

下一任务：Task 9.1.3 — 根据上述接口和旧 Store/Model，列出 Batch / Book / Run / Item / VIDEO 的状态字段与状态转换。

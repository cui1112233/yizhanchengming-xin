# Batch Factory V11/V12 状态模型盘点（Task 9.1.3）

本文件用于记录旧 v88 / 公网 Batch Factory 的 Batch、Book、Item、Run、Stage、VIDEO、Merge 状态字段，并给出新主线的迁移建议。

## 1. 结论

旧 v88 并不存在一个可以代表整个 Batch Factory 生命周期的单一 `status`。

真实状态由多层对象共同表达：

1. Batch / Book：保存业务实体、revision、settings、source、director、videos，本身不是总流程状态机。
2. staging Item：旧 Node 批量工厂页面使用一套 UI/流程状态。
3. BookStageRun：记录 assets/director/opening/visual/image/video 各阶段运行。
4. ProductionJob / ProductionTask：记录 VIDEO 生产。
5. MergeJob：记录成片合并。
6. 新主线 Run：当前只有批量项目级 pending/running/completed/failed，还没有接入上述细粒度状态。

迁移时不能把所有旧状态压成一个 `Run.status`，否则会失去单书失败、VIDEO 失败、重试、合并进度等重要信息。

---

## 2. 旧 v88：Batch

主要字段：

- `id`
- `title`
- `sourceIntakeId`
- `projectCoverJobId`
- `revision`
- `settingsState`
- `books[]`
- `createdAt`
- `updatedAt`

### 状态语义

Batch 本身没有正式 `status` 字段。

Batch 的当前运行状态需要从以下对象聚合得到：

- automation 状态
- book stage runs
- production jobs/tasks
- merge jobs
- publish 状态

因此新主线后续若增加 BatchProject 状态，应作为“聚合展示状态”，不能替代子任务事实状态。

---

## 3. 旧 v88：Book

主要字段：

- `id`
- `batchId`
- `bookId`
- `title`
- `sourceText`
- `workingFrontContent`
- `sourceTaskId`
- `platform`
- `txtText`
- `txtFileName`
- `sourceMetadata`
- `revision`
- `settingsState`
- `mode`
- `hook`
- `directorRevision`
- `assets`
- `assetRecords`
- `videos[]`

### 状态语义

Book 同样没有一个总 `status` 字段。

单书是否完成，要综合：

- 源文本是否存在
- Hook 是否完成/批准
- DirectorRevision 是否存在
- BookStageRun 最新结果
- ProductionTask 是否完成
- MergeJob 是否完成
- Publish 是否完成

新主线不能只用 `BookStatusFetched` 代表后续生产完成。

---

## 4. 旧 staging Item 状态

旧 Node `lib/batch-factory/store.js` 中 Item 使用以下状态：

| 状态 | 含义 |
|---|---|
| `pending` | 已加入待制作队列 |
| `queued_hook` | 爆款开头已进入排队 |
| `hook_generating` | 正在生成爆款开头 |
| `hook_review` | 爆款开头待审核 |
| `queued_director` | AI 导演已进入排队 |
| `director_generating` | AI 导演生成中 |
| `complete` | AI 导演完成，等待视频生产 |
| `failed` | 当前制作阶段失败 |

这些属于旧 Node staging 工作台状态，不能直接作为新 Go 核心状态机复制。

### 迁移建议

- `pending`：可映射为书级 workflow pending。
- `queued_hook / hook_generating / hook_review`：后续由 Stage/Hook 实体表达。
- `queued_director / director_generating`：后续由 BookStageRun director 表达。
- `complete`：不能直接映射为整书 completed，只能代表“导演阶段完成”。
- `failed`：必须带具体 stage + error，不能保留无上下文的总失败状态。

---

## 5. BookStageRun

阶段枚举：

| Stage | 含义 |
|---|---|
| `assets` | 资产抽取/准备 |
| `director` | 导演分镜 |
| `opening` | 换开头 |
| `visual` | 视觉提示词 |
| `image` | 图片生成 |
| `video` | 视频生成 |

运行模式：

- `missing`
- `force`

字段：

- `id`
- `batchId`
- `bookId`
- `stage`
- `status`
- `attempt`
- `requestId`
- `inputRevision`
- `errorMessage`
- `createdAt`
- `updatedAt`

StageRun 的 `status` 复用 ProductionState：

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`

### 重试语义

旧 v88 的重试不是“看到历史 failed 就重试”。

它会按 stage 取最新一次运行；如果后续 attempt 已成功，则历史失败不再视为待重试失败。

新主线后续必须保留：

- stage
- attempt
- requestId
- inputRevision
- errorMessage
- 最新失败是否已被后续成功覆盖

---

## 6. VIDEO：ProductionJob / ProductionTask

### ProductionState

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`

### ProductionJob

字段：

- `id`
- `batchId`
- `bookId`
- `requestId`
- `directorRevisionId`
- `status`
- `tasks[]`
- `createdAt`
- `updatedAt`

### ProductionTask

字段：

- `id`
- `videoId`
- `openingVariantIndex`
- `provider`
- `status`
- `attempt`
- `finalPromptHash`
- `compilationId`
- `compilationSegmentKey`
- `compileTrace`
- `referenceImageUrls`
- `downgradedAssetIds`
- `targetDurationSeconds`
- `requestedDurationSeconds`
- `actualDurationSeconds`
- `providerTaskId`
- `mediaUrl`
- `errorMessage`
- `createdAt`
- `updatedAt`

### ProviderTaskRef

Provider 回读也使用同一 ProductionState，并保留：

- `providerTaskId`
- `state`
- `mediaUrl`
- `requestedDurationSeconds`
- `actualDurationSeconds`

### 迁移要求

新主线后续 VIDEO 状态不能只放在 BatchProject/Run 上。

必须至少存在书级 ProductionJob + VIDEO 级 ProductionTask，否则无法支持：

- 同一本书多个 VIDEO 独立失败
- opening variants
- 单 VIDEO 重试
- provider task 轮询
- providerTaskId 回读
- 视频真实时长
- 多 provider
- 失败原因
- 取消

---

## 7. MergeJob

Merge 状态：

- `queued`
- `running`
- `succeeded`
- `failed`

主要字段：

- `id`
- `batchId`
- `bookId`
- `openingVariantIndex`
- `requestId`
- `timingMode`
- `speed`
- `providerTaskId`
- `status`
- `progressPhase`
- `progressCurrent`
- `progressTotal`
- `sources[]`
- `outputUrl`
- `errorMessage`
- `createdAt`
- `updatedAt`

Merge 不包含 `cancelled` 状态。

合并进度不能只靠 `status`，还依赖：

- `progressPhase`
- `progressCurrent`
- `progressTotal`

---

## 8. 新主线当前状态模型

### Intake

- `pending`
- `running`
- `completed`
- `partial_failed`
- `failed`

### Intake Book

- `pending`
- `fetched`
- `retryable_failed`

### Run

- `pending`
- `running`
- `completed`
- `failed`

当前 Run 字段：

- `id`
- `batchProjectId`
- `runAt`
- `status`
- `createdAt`
- `updatedAt`

当前 `pipeline.Create()` 只创建一个 `pending` Run。

尚未实现：

- Run -> running 的 worker 驱动
- 书级 workflow 状态
- BookStageRun
- ProductionJob
- ProductionTask
- MergeJob
- provider task polling
- publish 状态

---

## 9. 旧 -> 新建议映射

### 9.1 顶层 Run

新 `Run.status` 建议继续保留粗粒度：

- `pending`
- `running`
- `completed`
- `partial_failed`（建议新增）
- `failed`
- `cancelled`（建议新增）

Run 是聚合状态，不取代子状态。

### 9.2 Book workflow

建议后续增加书级执行记录，而不是复用 Intake Book.status。

建议聚合状态：

- `pending`
- `running`
- `completed`
- `failed`
- `blocked`
- `cancelled`

同时必须保存 current stage / last error。

### 9.3 StageRun

保留旧 v88 的：

- stage: `assets / director / opening / visual / image / video`
- status: `queued / running / succeeded / failed / cancelled`
- attempt
- requestId
- inputRevision
- errorMessage

### 9.4 VIDEO

保留 ProductionJob / ProductionTask 两层。

ProductionTask status 继续使用：

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`

### 9.5 Merge

继续独立实体：

- `queued`
- `running`
- `succeeded`
- `failed`

---

## 10. 不应迁移的设计

以下旧设计不应原样复制：

1. staging Item 的 `complete` 不得解释成整本书已完成。
2. `failed` 不得缺少 stage/error 上下文。
3. 历史 failed attempt 不得覆盖后来成功 attempt。
4. Batch 不应保存一个无法解释子任务情况的单一真假状态。
5. Intake Book.status 不应承担后续导演/视频/发布状态。
6. Browser-only 状态不能作为事实源。

---

## 11. 后续实现影响

本盘点直接约束 Task 9.4：

- 9.4.1 pending -> running
- 9.4.2 单本独立状态
- 9.4.3 一本失败不阻断其他书
- 9.4.4 错误持久化
- 9.4.5 单本重试
- 9.4.6 幂等保护
- 9.4.7 自动化到点执行
- 9.4.10 Worker 恢复

也会约束后续 Task 14 VIDEO 状态机和视频 provider 实现。

## 12. Task 9.1.3 验收结论

已完成：

- Batch 字段盘点
- Book 字段盘点
- staging Item 状态盘点
- StageRun 阶段和状态盘点
- ProductionJob / ProductionTask 状态盘点
- MergeJob 状态盘点
- 新主线 Intake / Book / Run 状态盘点
- 旧 -> 新状态映射原则
- 明确后续需要新增的细粒度执行实体

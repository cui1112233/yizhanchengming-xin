# 批量工厂链路加固设计（2026-11-02）

状态：待评审
范围：A 最小加固 + B 状态机收敛 + C 拆分与清理
目标分支：`v88`

## 1. 背景与目标

批量工厂承担「抓取原文 → 男女频/风格分类 → 漫画分镜与出图 → 合集 → 视频 → 发布到 121」的整条流水线。当前主要问题不是某个阶段算不出来，而是：

- 单本书的配置缺陷会让**整批零执行**；
- 失败与阻塞在 UI 上**没有声音**，用户不知道跑没跑、卡在哪；
- 部分阶段**不幂等**，重试会重复上传；
- 阶段与就绪规则在前后端各写一份，长期必然漂移。

本设计的目标依次是：**不因一本书卡住整批** → **任何操作都有可见反馈** → **重试安全** → **规则单一来源** → **代码与磁盘可维护**。

非目标：不改业务算法（提示词、分类策略、供应商选型），不改发布字段映射，不重做 UI 视觉。

## 2. 现状链路

| 段 | 入口 | 主要落库 | 失败后果 |
|---|---|---|---|
| 抓取原文 | `ensureAutomationSource` | `originalRaw` / `originalStatus` | 单书停；书城原文失败会整批拒绝 |
| 分类 | `finalizeFetchedBatchClassification` | `classifyStatus` | 恒为 `classified`，不阻塞（已修） |
| 漫画分镜 | `comic_assets` → `comic_bootstrap` | `batch_factory_v12_comic_pages` | 缺提示词整本停 |
| 漫画出图 | 页并发 5，页级锁 + 15min 租约 | `batch_factory_v12_comic_page_images` | 在途页被跳过 → 合集缺页 → 整本失败 |
| 合集 | `FOR UPDATE` 比对去重 | `batch_factory_v12_comic_collections` | 已修（上一轮 `driver: bad connection`） |
| 视频 | `director` → `opening` → `video` → `merge` | `batch_factory_v12_merge_jobs` | 单书 `fail` 后 `continue` |
| 发布 | `publishVideo` → `BatchFactory121Publisher.Submit` | book `sourceMetadata.websiteSubmitStatus` | 无幂等，重试重复上传 |

并发模型：每书一个 goroutine，信号量上限 4；单阶段超时 10 分钟；非漫画超时自动重试 30s × 3；漫画整本 3 次指数退避。

## 3. 问题清单

### P1 预检一票否决

`validateAutomationPreflight`（`batch_factory_automation.go:732-815`）把任意一本书的「文本模型未配置 / 书城原文失败 / 视频或图片模型未配置 / 发布配置缺失」累积成整批错误，`CreateJob:686` 直接返回，**整批零执行**。

额外缺口：定时任务在 `:236` 把 `scheduled` 改成 `running` 后直接 `launchRun:281`，**完全不跑预检**。同一套配置，立即启动被拦、定时启动放行，行为不一致。

### P2 点击无反馈

`BatchFactoryNovelList.jsx:4246` 为裸 `return`：

```js
if (!batch?.id || automationBusy) return;
```

`automationBusy` 在 `:4279` 置位、`:4298` 复位。请求悬挂期间所有点击被静默吞掉，用户看到的就是「跟没点一样」。

### P3 漫画页锁与租约

`comicPageGenerationStaleAfter = 15m`（`batch_factory_comic_generation_gate.go:15`）；`runComicPageWork:77` 对「该页正在生成」只跳过、不失败；随后合集校验发现缺页，整本以失败结束。`generateComicWithRetry` 3 次退避，最坏白等约 30 分钟才报缺页。

### P4 状态文件全量重写

`saveLocked:396-412` 每次把整个 job（全部书、全部日志）`json.Marshal` 后写 tmp + rename；`setJob:414-423` 持同一把锁全量读写。10 本 × 8 阶段 ≈ 80+ 次全量写，并发 4 本时锁争用；写失败或崩溃只会在下次启动由 `RecoverInterruptedJobs:427` 兜底成 `stopped`。

### P5 发布不幂等

`batch_factory_v11_publish_121.go:252-265` 直接 `UploadVideo` + `Connector.Submit`，无「已上传」短路。`:270-273` 回读失败即 `fail("upload")`，重试会把视频和正文**再传一遍**，可能在 121 产生重复作品。

### P6 规则重复

分类规则三份：后端 `finalizeFetchedBatchClassification`（`batch_factory_v12_single_fetch.go:236-299`）、`batchFactoryBookState.js:194-198`、`batchFactoryNovelListHelpers.js:566-570`。漫画就绪判断两份：后端 `batch_factory_automation.go:977-1019`、前端 `batchFactoryBookState.js:125-181`。

## 4. 设计 A：最小加固

### A1 预检降级为「全局致命 + 逐书跳过」

把 `validateAutomationPreflight` 拆成两层：

- `validateAutomationGlobalPreflight(ctx, owner, batch, runMode) error`：只保留**全批不可执行**的条件——文本/图片/视频执行器整体缺失、发布执行器整体缺失。返回错误时整批拒绝，理由讲清楚。
- `automationBookPreflightReason(ctx, owner, batch, book, runMode) string`：逐书条件——`isDirectFirstSourceBlocked`、`textModelId` 为空、`imageModelId`/`videoModelId` 为空、`full_submit` 下 `publishSettings` 未绑定。返回空串表示可执行。

调用点：

- `CreateJob` 只调全局预检。
- `run()` 每本书进入阶段循环前调逐书预检；有理由时写 `status=blocked, stage=preflight, message=理由`，`persist()` 后 `continue`，其余书照常执行。
- 定时任务不需要额外处理——它同样走 `run()`，逐书预检自动生效，补齐 `:236 → :281` 的缺口。

结果：9 本健康 + 1 本缺模型 → 9 本正常跑，1 本 `blocked` 且原因可见；只有全批执行器整体不可用时才拒绝启动。

### A2 所有自动化操作都有可见反馈

改 `runAutomationAction` 前置守卫：

- 无 `batch.id` → `message.error('页面批次尚未加载，请刷新后重试')`
- `automationBusy` 为真 → `message.info('上一次操作仍在执行，请稍候')`
- 增加 30 秒看门狗：记录 `automationBusyAtRef`，超过 30 秒视为僵死，复位后允许继续并提示「上一次请求超时，已重置状态」。

提交期间用 `message.loading('正在提交自动化任务…')` 常驻，`finally` 中销毁；成功/失败再替换为 `success`/`error`。这样从点击到返回全程都有可见状态。

### A3 发布幂等

在 `BatchFactory121Publisher.Submit` 增加三段状态机，复用 `book.SourceMetadata`：

| 状态 | 含义 | 重入行为 |
|---|---|---|
| `uploaded` | 已回读确认 | 直接短路返回成功，不重复上传 |
| `upload_pending_confirm` | 视频已上传、容器已提交，但回读失败 | 先 `GetSubmission`；确认成功则补写 `uploaded`，确认「不存在」才重新 `Submit` |
| （无） | 首次发布 | 正常执行 |

实现要点：

1. 入口先读 `websiteSubmitStatus`，`uploaded` 且未显式 `force` 时短路。
2. `UploadVideo` 全部成功后，**先持久化** `websiteSubmitHeads` 与 `websiteSubmitStatus=upload_pending_confirm`，再调 `Connector.Submit`。
3. `GetSubmission` 失败时返回带哨兵的 `errPublishUnconfirmed`，自动化把该书写 `status=failed, stage=upload_confirm, message=已上传待确认`，而不是「上传失败」。
4. `upload_pending_confirm` 重入时只走「回读优先」路径，绝不无条件重传。

### A4 漫画缺页单页补图

在漫画整本生成内部增加一次**补图轮**：

1. 首轮 `runComicPageWork` 结束后统计缺页。
2. 若缺页 > 0 且失败原因不是内容审核拒绝（`errShuihuoImageContentPolicy` 哨兵），等待 60 秒让在途 detached worker 释放页锁。
3. 只对缺页再跑 `runComicPageWork` 一轮，仍在整本 60 分钟预算内。
4. 补图后仍缺页才交给外层 3 次退避重试；错误信息里带「第 N 页缺失」。

这样单页锁冲突不再直接把整本打成缺页失败。

## 5. 设计 B：状态机收敛与状态存储

### B1 阶段契约单一来源

新增契约文件 `frontend/src/user/pages/shuihuo/automationStages.json`，内容为阶段 id、顺序、模式（video/comic）、用户可见文案、终态集合。放在前端目录内是为了让 Vite 可以直接 `import`，避免引入仓库根级共享目录带来的构建根限制。

- Go 侧：`batch_factory_automation_stages.go` 定义常量与顺序；新增 `TestAutomationStagesContract` 读取 `../../../frontend/src/user/pages/shuihuo/automationStages.json` 并逐项比对 Go 常量，不一致即测试失败。
- 前端侧：`batchFactoryBookState.js` 与 `batchFactoryNovelListHelpers.js` 直接 `import` 该 JSON 生成 timeline，删除各自硬编码的阶段数组与文案。
- 保留前端本地兜底：JSON 缺失时退回现有常量，保证构建不会因契约文件缺失而崩。

好处是「改一处、测一处」：任何阶段改名或新增，Go 测试先红。

### B2 状态存储改增量

把单一 `automation-state.json` 拆成：

```
data/automation/index.json          # jobId -> {owner,batchId,state,updatedAt}
data/automation/jobs/<jobId>.json   # 单个 job 全量
```

- `setJob` 只写对应 job 文件；仅当 `state` 实质变化时更新 `index.json`。
- 每个文件仍走 `tmp` + `rename` 原子写。
- `activeAutomationJob` 读 `index.json` 定位候选，再读 job 文件。
- 启动迁移：检测旧 `automation-state.json`，按 jobId 拆分写出，成功后重命名为 `automation-state.json.migrated`；迁移失败则继续沿用旧文件，不阻塞启动。
- 每本书 `logs` 数组上限 200 条，超出丢最旧，控制文件体积。

并发收益：4 本书并发时锁只覆盖「写自己的 job 文件 + 可能更新索引」，不再整份读写。

### B3 前端常显状态

工作台工具栏增加一条常驻状态行：自动化状态（运行中/已暂停/待处理/失败数）+ 当前阶段文案 + 最近错误。数据来自已有 `loadRuntimeStatus` 的 `automation` 快照，不新增接口。这样即使用户没打开「任务/日志」弹窗，也能看到「在跑 / 卡在哪 / 为什么停」。

## 6. 设计 C：拆分与清理

### C1 大文件拆分

- `backend/internal/httpapi/shuihuo_projects.go`（4557 行）按 handler 族拆为：路由与通用（保留原文件）、`shuihuo_projects_comic.go`、`shuihuo_projects_publish.go`、`shuihuo_projects_assets.go`。拆分前先导出函数清单确认边界，避免循环依赖。
- `backend/internal/httpapi/novel_fetch_workshop.go`（3614 行）按「抓取 / 分类 / 工作台操作」拆成两个文件。
- 前端 `AgentCanvas.jsx`(2661) 与 `ScriptPage.jsx`(2135) 本轮不动——它们不在本链路关键路径上，避免把风险混进来。

判定标准：拆完 `go build ./...` 与全部 `_test.go` 通过，且不改变任何导出符号。

### C2 磁盘清理

- 删除 `data/local-executor-artifacts`（3.1 GB）：先确认该目录只含已完成的执行器中间产物，且无运行时引用。
- 删除 `_local-archive-20261028/.direct-staging`（3.46 GB）：属历史 Node 发布链残留。
- 均先列目录抽样确认，再删；删除记录写入提交说明。

### C3 Git 分叉处理

当前本地 `v88` 领先 origin 1174、落后 19。

1. 建备份分支 `v88-backup-20261102`。
2. `git fetch`，逐条检查落后的 19 个提交，确认是否已包含在本地改动中（大概率是历史发布提交）。
3. 采用 `merge`（保留历史，不重写已推送提交）而非 `rebase`。
4. 冲突只允许在文档与锁文件层面，出现源码冲突立即停止并回报。
5. 完成后 `git gc`。

## 7. 兼容与迁移

- 预检行为变更只放松不收紧：原本能启动的批次照旧；原本被整批拒绝的批次现在能部分启动。
- 发布新增 `upload_pending_confirm` 状态，前端 `isBookUploadedTo121` 需同时识别 `uploaded` 与 `upload_pending_confirm`（后者显示「已上传待确认」）。
- 状态文件迁移失败不影响启动。
- 阶段契约文件缺失时前端回退现有常量。

## 8. 测试策略

Go：

- `TestAutomationBookPreflightReason`：覆盖缺文本模型、缺图片模型、缺视频模型、书城原文失败、发布配置缺失，断言只影响单书。
- `TestAutomationRunSkipsBlockedBook`：两本书，一本缺配置，断言另一本 `completed`、缺失本 `blocked` 且 `message` 非空。
- `TestPublish121IdempotentUploaded`：`websiteSubmitStatus=uploaded` 时不调用上传与提交。
- `TestPublish121PendingConfirmReadbackFirst`：回读成功补写 `uploaded`；回读不存在才重新提交。
- `TestAutomationStagesContract`：Go 常量与 JSON 一致。
- `TestAutomationStateStoreMigration`：旧文件拆分为多文件，内容等价。

前端（source test）：

- `runAutomationAction` 在 busy 时输出可见提示而非静默 `return`。
- timeline 使用契约 JSON 生成的阶段文案。

手工验收：本地 4001 用两本书的批次，故意让其中一本缺视频模型，点「立即启动」：

1. 立刻出现「正在提交自动化任务…」。
2. 一本进入 `running`，一本显示 `blocked` 且原因「视频模型未配置」。
3. 再点一次按钮出现「上一次操作仍在执行，请稍候」。

## 9. 验收标准

- [ ] 单书配置缺陷不再导致整批零执行，且原因在 UI 可见
- [ ] 定时任务与立即任务走同一套预检
- [ ] 任何自动化按钮点击都有即时可见反馈，无静默 `return`
- [ ] 对已上传的书重跑上传，不产生重复作品
- [ ] 漫画缺 1 页时能补齐，不再整本失败
- [ ] 阶段定义只有一处真相，且 Go 测试守护
- [ ] 状态文件写入不再全量重写，日志有上限
- [ ] `go build ./...`、全部 Go 测试、前端构建通过

## 10. 实施顺序与回滚

1. A1 预检 → A2 反馈 → A3 发布幂等 → A4 缺页补图
2. B1 阶段契约 → B2 状态存储 → B3 状态行
3. C1 拆分 → C2 清理 → C3 Git

每个 A 项独立提交，可单独 revert。A3 涉及发布，提交前必须有对应单测。C2 只删数据不删代码，回滚无风险。

## 11. 不做的事

- 不重写分类算法，不改提示词策略
- 不改 121 表单字段映射
- 不动 `AgentCanvas.jsx` / `ScriptPage.jsx`
- 不引入新的队列或消息中间件（现有 goroutine + 信号量够用）
- 不删除历史分支（合并阶段 7 再处理）

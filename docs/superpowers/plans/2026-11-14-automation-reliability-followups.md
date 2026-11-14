# 自动化可靠性后续优化计划

- 日期：2026-11-14
- 基线：`v88` @ `3c6653ac`（已上线 `qiantie-202611141140`）
- 前置文档：[2026-11-14-automation-reliability-hardening-design.md](../specs/2026-11-14-automation-reliability-hardening-design.md)
- 约束：每阶段必须“源码 → 测试 → 本地 4001 验证 → Go embed 发布”，不重构调度架构，不使用编译产物替代源码。

本计划只覆盖上一轮明确未纳入的项，按“先取证、再小改、可独立回滚”排序。

## 阶段 0：先补证据（0.5 天，无代码变更）

目的：把“待验证”变成结论，避免凭假设改代码。

| 验证项 | 方法 | 输出 |
|---|---|---|
| TOS 回源真实耗时边界 | 本地注入慢响应/断连，记录 `RestoreStorageRef` 从请求到返回的耗时 | 是否会在客户端断开后继续占用 goroutine |
| 回源失败后的本地/仓库状态 | 让 `MarkRehydrated` 返回错误，检查本地文件、`batch_factory_v12_media_objects.state` | 是否出现“本地有文件、仓库仍 cleaned” |
| reaper 写入频率 | 生产读取 binlog 文件字节差，观察 10 分钟 | 回收事件是否造成可测写放大 |

验收：产出一份数据记录追加到本文档“验证结果”小节；没有证据的项不进入编码阶段。

## 阶段 1：保存粒度收口（P2，1 天）

现状：
- [batch_factory_automation.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/batch_factory_automation.go#L481-L513) `ReapZombieJobs` 在回收僵尸任务时调用 `saveLocked(state)`，重写全部 job。
- [batch_factory_automation.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/batch_factory_automation.go#L1508-L1554) `RecoverInterruptedJobs` 启动时同样全量保存。
- [automation_status.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/automation_status.go#L61-L80) 状态读取路径在修复后调用全量 `Save`。

目标：所有“只改少数 job”的路径只保存变化的 job，复用已有的 `AutomationJobSaver.SaveJobs`。

改动点：
1. 抽出 `saveJobsLocked(state, changedIDs)`：变化集合为空直接返回；存储实现 `AutomationJobSaver` 时逐 job upsert，否则回退全量 `Save`。
2. `ReapZombieJobs` 使用 `reclaimed` 集合。
3. `RecoverInterruptedJobs` 记录被改写的 jobID 集合。
4. `FileAutomationStatusReader.Status` 的修复写回改为同上接口。

验收：
- 新增测试证明回收 1 个僵尸任务只 upsert 1 行。
- 现有 `automation_state_store_test.go` 全通过。
- 生产 binlog 增长不高于当前基线（4MB/2min 量级）。

风险：低。内存缓存仍是全量，只是落库粒度变细；失败回退保持原行为。

## 阶段 2：回源取消与超时边界（P1，1～2 天）

现状：
- [store.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/localartifact/store.go#L362-L383) `Open` 调用 `RestoreStorageRef(context.Background(), ...)`，HTTP 请求取消无法传播。
- TOS 客户端已配置连接 30s、请求/读写 5min、重试 3 次（[tos_store.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/productionmedia/tos_store.go#L39-L47)），因此不是“永久挂起”，但一次失败回源可能占用数分钟。

目标：回源受调用方生命周期约束，并给非请求路径一个明确上限。

改动点：
1. `Store.Open` / `Store.Path` 增加带 `context.Context` 的变体（如 `OpenContext`），内部使用调用方 ctx；旧方法保留为 `context.Background()` + 默认超时。
2. HTTP handler 与 executor 读取路径改用请求 ctx。
3. 默认回源超时设为可配置（如 `QIANTIE_ARTIFACT_RESTORE_TIMEOUT`，默认 2 分钟），超时返回可识别的错误，交由现有失败分类转人工重试。
4. 日志记录 storageRef、耗时、结果，便于定位慢回源。

验收：
- 单测：ctx 取消后 `OpenContext` 快速返回，且不落地半成品文件。
- 本地验证：客户端断开后服务端日志显示回源被取消。

风险：中。需要逐个改造调用点，避免漏传 ctx 导致行为不变；保留旧方法确保外部调用不受影响。

## 阶段 3：`MarkRehydrated` 状态协调（P2，0.5 天）

现状：[resolver.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/productionmedia/resolver.go#L39-L46) 本地恢复成功后若 `MarkRehydrated` 失败，直接返回错误，但本地文件已存在。

目标：明确“本地已恢复、仓库标记失败”的可恢复语义。

改动点：
1. `MarkRehydrated` 失败时记录日志并返回可识别错误，但不再让调用方误判为“本地没有文件”。
2. 恢复成功后短期内重复调用时，若本地文件校验通过则跳过重复下载，只重试仓库标记。
3. 补测试覆盖“下载成功 + 标记失败 + 再次打开”。

验收：重复打开不再触发第二次 TOS 下载；仓库状态最终收敛为 `verified`。

风险：低，不改变清理与校验前置条件。

## 阶段 4：发布产物来源一致性（P2，1 天）

现状：本轮为避开其他会话未提交的前端文件，使用 `-SkipFrontendBuild` 复用既有 `frontend/dist`。这能在短期保持一致，但长期存在“源码与线上页面不一致”的风险。

目标：发布二进制可追溯到唯一来源，前端产物来源可校验。

改动点：
1. `deploy-go-embedded.ps1` 增加 `-RequireCleanFrontend` 开关：构建前校验 `frontend/src` 无未提交改动，否则拒绝发布。
2. 在 `release.json` 记录 `frontend_source_state`（clean/dirty）与入口 hash。
3. 发布检查脚本比对线上 `frontend_entry` 与发布包记录。

验收：脏工作区下默认给出明确告警；干净工作区可完整重建并发布。

风险：流程约束变严，需要在其他会话并行开发时显式使用 `-SkipFrontendBuild` 并记录原因。

## 阶段 5：既有测试债（P2，1 天）

现状（非本轮引入）：
- 三个 screenplay prompt 哈希测试受 CRLF 影响。
- `TestNovelCreationRoutesExposeSixModesAndProjectRunChain` 期望“番外创作”，实际“同人创作”。
- `TestGPTImageConcurrencyIsProcessWideAcrossTasks` 期望 4，实测 2。

目标：让全量 `go test ./...` 可作为发布门禁。

改动点：逐项确认产品语义后修正断言或实现，禁止为了让测试通过而放宽真实约束。

验收：`go test ./...` 连续两次全绿；CI 或发布前脚本执行该命令。

## 阶段 6（可选）：并发预算可观测性（1～2 天）

现状：实际并发由单书 worker、任务图片并发、同时出图书数、网关自适应、Redis 额度共同决定，页面无法说明“当前被哪一级限制”。

目标：只做展示与诊断，不改并发算法。

改动点：在任务状态响应中增加 `concurrencyBudget` 字段（各层上限、当前生效值、阻塞原因），前端在任务详情展示；不参与调度决策。

验收：构造不同配置，接口返回值与实际调度一致；不改变现有吞吐基线。

## 排期与依赖

1. 阶段 0 先行，决定阶段 2、3 是否需要调整。
2. 阶段 1 独立，可随时发布。
3. 阶段 2、3 同属媒体链路，建议同批验证。
4. 阶段 4、5 属于工程与质量门禁，可与功能修复并行。
5. 阶段 6 在以上稳定后评估，不阻塞前五项。

## 每阶段完成标准

- 有对应测试或可复现验证记录。
- 本地 4001 实际运行通过。
- 提交只包含本阶段文件，记录 Git SHA。
- 生产发布记录 SHA、二进制 SHA256、回滚点。
- 未达标的项不标记完成，转入下一轮。

## 验证结果

### 2026-11-14 阶段 0 取证

1. 回源不受调用方取消影响（本地探针）
   - 调用前取消 caller ctx，`Store.Open` 仍执行完整回源：`open_elapsed_ms=1504`，`caller_ctx_err=context canceled`。
   - 结论：请求取消不会传播到回源路径；需要 `OpenContext` 与默认超时。

2. `MarkRehydrated` 失败后的本地/仓库状态（本地探针）
   - `rehydrate_fail locality=local_file_present first_remote_opens=1 second_remote_opens=2 rehydrate_calls=2`。
   - 结论：本地文件已落盘但仓库标记未更新；二次访问会重复下载。不会丢数据，但存在重复流量。

3. reaper 写入频率（生产 3 分钟采样，2026-11-14 12:2x）
   - `jobs=265 job_json_total=6684343 bytes`，`delta_bytes=664319`，`bytes_per_min=221439`。
   - `reaper_log_lines=0`、`panic_log_lines=0`、`artifact_restore_lines=0`。
   - 结论：观测窗口内没有僵尸回收事件，reaper 未产生写入；当前 binlog 增长约 216KB/分钟，远低于“全量重写 6.4MB/次”的旧行为，说明单行保存持续生效。阶段 1 仍值得做，但它属于低频优化，不是当前主要写放大来源。

### 2026-11-14 阶段 6 执行结果

任务状态接口新增 `concurrencyBudget`，只做展示，不参与调度：

```json
{
  "bookWorkers": 3,
  "bookWorkerCeiling": 4,
  "taskImageConcurrency": 8,
  "comicBookSlots": 3,
  "comicBookSlotCeiling": 8,
  "gatewayColdStart": 2,
  "gatewayCeiling": 30,
  "effectiveImageConcurrency": 2,
  "limitingLayer": "gateway_cold_start",
  "note": "实际并发取各层最小值；网关会从冷启动逐步升到天花板"
}
```

- `automationBookWorkerCeiling=4`、`shuihuoAdaptiveMaxLimit=30` 提取为具名常量，运行时代码与展示共用，避免两处漂移。
- 新增三个测试覆盖书籍上限、网关冷启动、任务级出图限制与公开视图字段。
- 前端展示暂缓：`BatchFactoryWorkbenchPage.jsx` 与 `batchFactoryV12.js` 仍被并行会话改动，直接提交会把未完成前端一并带入；待前端会话提交后再接展示。
- 验证：`go test ./...` 连续两次通过；本地 4011 `git_sha=0d2fc026`、未登录 401。
- 生产发布：`qiantie-202611141318`，二进制 SHA256 `bf231f93…`，`git_sha` 与 `runtime_manifest_sha` 均为 `0d2fc026a029c585a0b3f5e00b9a71cb74d789f4`；`verify-release.ps1` PASS。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141318-0d2fc026`（上一版本 `qiantie-202611141247`）。
- 本次发布包含并行会话已提交的 `3ae9747f`（小说抓取确定性失败转人工），全量测试覆盖通过。

#### 2026-11-14 追加：任务结束后残留“排队中”的书

- 现象：用户勾选“排队中”的书后，插队按钮仍置灰。
- 根因：任务收尾时只统计 failed/completed，未把剩余的 `queued` 书落成终态；任务已是 failed，
  插队要求 `running/paused`，于是按钮永远不可用。生产存在 4 本此类残留书（qiuwanjing、weimeixia、
  jiansen、tanglu 的失败批次各 1 本）。
- 修复：`normalizeQueuedBooksAtFinish` 在收尾时把 queued 标记为失败并写明原因，计入失败数；
  不触碰 scheduled / pending_confirmation。提交 `ed661719`，全量测试连续两次通过。
- 发布状态：**已发布**。`qiantie-202611141904`（`ed661719`），`git_sha` 与
  `runtime_manifest_sha` 均为 `ed661719bdbeebc888e6d1dd1c687a91cc52882f`，前端入口
  `/assets/user-DQpfXXLl.js`（JS 资源 200，服务 active）；`verify-release.ps1` PASS。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141904-ed661719`（上一版本 `qiantie-202611141506`）。
- 旧数据：已对生产 4 本历史残留书做一次性状态修复（只改 `queued`，不动其他状态）：
  qiuwanjing / weimeixia / jiansen / tanglu 各 1 本改为 `failed`，写明“任务已结束，这本书未开始执行；请点击重试”，
  并同步 `counts.queued-1`、`counts.failed+1`；修复前备份 `/tmp/queued-repair-backup.json`。
  复核全库已无“任务非 running/paused 却仍有 queued 书”的残留（`stale_residual=0`）。

#### 阶段 6 遗留：并发预算前端展示（已用任务日志替代上线）

- 2026-11-14 15:06 上线 `qiantie-202611141506`（`696468c9`）：书籍首次出图时在任务日志写入一行
  「并发预算：书 N/4 路 · 本任务出图 N 路 · 同时出图书 N 本 · 网关冷启动 2 路（上限 30 路）」，
  前端日志区无需改动即可看到，说明当前为什么是这个路数。
- 前端独立徽标仍未接入：`frontend/src` 仍有并行会话未提交改动；待其提交后再补。
- 本发布嵌入的前端入口与生产一致（`/assets/user-vm6z-26I.js`），未带入未完成页面改动。
- `verify-release.ps1` PASS，回滚点 `/opt/www/qiantie/rollback-qiantie-202611141506-696468c9`。

### 2026-11-14 阶段 5 执行结果

三项既有失败均已定位并修正，`go test ./...` 连续两次全绿：

| 用例 | 原因 | 处理 |
|---|---|---|
| `TestPromptDefinitionsKeepExpectedIDsAndVersions` | `agent.director-plan` 已在 `a3024dcd` 升到 v2，测试仍期待 v1 | 断言更新为 v2 |
| `TestNovelCreationRoutesExposeSixModesAndProjectRunChain` | 路由自 `1f587561` 起返回“番外创作”，测试仍期待“同人创作” | 断言改为“番外创作” |
| `TestGPTImageConcurrencyIsProcessWideAcrossTasks` | 默认单书并发已改为 4，但网关冷启动仍是 2 路，断言与自适应策略不一致 | 新增 `shuihuoAdaptiveInitialLimit=2` 常量，断言冷启动上限并校验不超过默认值 |

- 未放宽任何真实约束：并发用例仍验证进程级共享上限，只是把“冷启动 2 路”明确成常量。
- 本阶段改动为测试与常量命名，不改变线上行为；生产保持 `qiantie-202611141247`（`75c45576`）。
- 后续发布门禁：`go test ./...` 连续两次通过后再构建发布包。

### 2026-11-14 阶段 4 执行结果

- [deploy-go-embedded.ps1](file:///f:/脚本测试/一战晟铭-v88/scripts/deploy-go-embedded.ps1) 新增 `-RequireCleanFrontend`：
  - `frontend/src` 有未提交改动时拒绝构建，并列出前 8 项；
  - 同时使用 `-SkipFrontendBuild` 时，若 `frontend/dist/index.html` 早于最新源码文件，拒绝发布；
  - 无法识别入口脚本时拒绝发布。
- `release.json` 新增字段：`frontend_source_state`、`frontend_entry`、`frontend_index_sha256`、`frontend_built_at`。
- 新增 [verify-release.ps1](file:///f:/脚本测试/一战晟铭-v88/scripts/verify-release.ps1)：比对发布记录与 `/api/runtime-build-info` 的 `git_sha`、`runtime_manifest_sha`，生产未上报入口时改为校验首页是否引用发布包入口。
- 验证：
  - `-RequireCleanFrontend` 在当前 39 项未提交前端改动下正确拒绝构建；
  - 默认构建产出的 `release.json` 记录 `frontend_source_state=dirty`、`frontend_entry=/assets/user-pA6972Cj.js`；
  - `verify-release.ps1` 对 HEAD 发布记录返回 FAIL（生产仍为 `75c45576`），对匹配的 `75c45576` 发布记录返回 PASS。
- 本阶段只改脚本与文档，不改变线上二进制；生产保持 `qiantie-202611141247`（`75c45576`）。

### 2026-11-14 阶段 1 执行结果

- `75c45576`：新增 `saveJobsLocked`，reaper、dispatcher、启动恢复、加载规范化统一按 `changedIDs` 走 `SaveJobs` 单任务 upsert；状态读取修复路径同样只保存被修复任务。
- 测试：新增 `TestReapZombieJobsSavesOnlyReclaimedJobs`、`TestRecoverInterruptedJobsSavesOnlyRunningJobs`、`TestStatusRepairSavesOnlyRepairedJobs`；`httpapi` 自动化相关回归 46s 全通过。
- 本地 4011 启动成功，`git_sha=75c45576`，未登录接口返回 401。
- 生产发布：`qiantie-202611141247`，二进制 SHA256 `16907d27…`，`git_sha` 与 `runtime_manifest_sha` 均为 `75c45576`。
- 前端入口变为 `user-pA6972Cj.js`（包含并行会话 12:14 构建的未提交前端，经用户确认后随本次发布）。
- 浏览器验收：登录页完整渲染、工作台可达、控制台无错误。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141247-75c45576`（上一版本 `qiantie-202611141214`）。

### 2026-11-14 阶段 2 / 阶段 3 执行结果

- `d87146d3` 阶段 2a：`OpenContext` / `PathContext`、`ErrRestoreTimeout`、默认 2 分钟上限、回源日志、`QIANTIE_ARTIFACT_RESTORE_TIMEOUT`。
- `afc4918f` 阶段 2b 第一批：漫画预览/缩略图/打包下载、合并封面、执行器产物、水果项目图片改用请求上下文。
- `bc16b9a5` 阶段 2b 第二批：漫画参考图、素材、资产图、视频资产导入、音频与图片生成结果改用请求上下文。
- `3687f8bd` 阶段 3：本地同尺寸文件跳过重复下载，`MarkRehydrated` 失败不再阻断读取，下次访问重试标记。
- 测试：`localartifact`、`productionmedia`、`novelfetchworkshop` 全通过；`httpapi` 仅剩两个既有失败用例（`TestNovelCreationRoutesExposeSixModesAndProjectRunChain`、`TestGPTImageConcurrencyIsProcessWideAcrossTasks`），与本次改动无关。
- 本地 4011 实际启动成功，`/api/runtime-build-info` 返回 `git_sha=bc16b9a5`，媒体相关路由未登录返回 401。
- 生产发布：`qiantie-202611141214`，二进制 SHA256 `fb9c960d…`，`git_sha` 与 `runtime_manifest_sha` 均为 `bc16b9a5`，首页与 `/assets/user-BYj46LWB.js` 均 200。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141214-bc16b9a5`（上一版本 `qiantie-202611141140`）。
- 说明：本次发布包含同工作区另一会话已提交的 `01f259be`（小说抓取自动化与 Redis 队列），该提交的包测试已随上述命令通过。


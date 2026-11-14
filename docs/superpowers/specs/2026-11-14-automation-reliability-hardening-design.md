# 自动化链路可靠性加固设计（方案 A）

- 日期：2026-11-14
- 分支：`v88`
- 范围：定点修复已由源码证实的 P1 缺陷，补齐后台任务可观测性；不重构调度架构、不改并发算法、不改 TOS 状态机。

## 1. 背景与目标

自动化生产曾出现“页面显示执行中、既不失败也无法重试”的现象。前序修复已解决主因（全量 job 重写导致的 MySQL binlog 写放大、日志无限增长、心跳高频落库）。本轮针对审计中剩余的可证实缺陷做最小修复：

1. TOS 回源写入绕过磁盘写保护，且 `ENOSPC` 未被统一归一。
2. `SaveMedia` 部分写入错误未归一为 `ErrDiskPressure`。
3. 单任务保存失败前已把新状态写入内存缓存，MySQL 与内存可能分歧。
4. 执行 goroutine panic 只写任务级 `failed`，书级 `running/queued` 与 counts 未收口。
5. 媒体后台 scanner / migrator / 磁盘清理循环忽略顶层错误，故障无日志。

非目标（本轮不做）：

- 不调整任何并发上限（单书 worker、图片并发、出图书数、网关自适应、Redis 额度）。
- 不改为重启后自动续跑（防止重复提交和重复计费的安全策略保持不变）。
- 不改 `productionmedia` 的对象状态机与保留期。
- 不清理 `v88` 上其他会话的未提交文件。

## 2. 缺陷与修复

### 2.1 磁盘写保护覆盖 TOS 回源

现状：`localartifact.Store.Restore` 直接 `MkdirAll / CreateTemp / io.Copy / Sync / Rename`，既不检查 `writeGuard`，也不把 `ENOSPC` 归一为 `ErrDiskPressure`。于是磁盘超过 85% 后，新媒体写入会被拒绝，但“本地缺失 → 从 TOS 回源”的路径仍会继续写盘。

修复：

- `Restore` 入口调用 `checkWriteAllowed()`，与 `SaveImage / SaveMP4 / SaveMedia` 一致。
- 所有写入阶段统一 `diskPressureError`：`MkdirAll`、`CreateTemp`、`io.Copy`、`Sync`、`Rename`。
- 校验失败仍返回 `ErrRemoteIntegrity`，不得被磁盘错误覆盖。

影响：磁盘压力下回源会快速失败并返回 `ErrDiskPressure`，由既有的自动化失败分类转成人工基础设施失败，避免磁盘被打满后继续写入。

### 2.2 `SaveMedia` 错误归一

`os.CreateTemp`、`io.Copy`、`tmp.Close` 三处补齐 `diskPressureError`，与 `Sync`、`Rename`、`MkdirAll` 保持一致。

### 2.3 单任务保存与内存缓存顺序

现状：`setJobLocked` 先 `c.automationState = &state` 再 `SaveJobs`。且 `loadLocked` 返回的 `state.Jobs` 与缓存共享同一个顶层 map，`state.Jobs[jobID] = job` 在保存前就已经改到了缓存。

修复：

- 保存前复制顶层 `Jobs` map（浅拷贝，值仍指向各 job 文档），在副本上写入目标 job。
- 先执行 `SaveJobs`，成功后才把 `state.Jobs = next` 并替换 `c.automationState`。
- 保存失败时直接返回错误，缓存保持旧值。

影响：数据库短暂不可用时，页面与下一次读取仍看到上一次已持久化的状态，不会出现“内存已完成、MySQL 仍是 running”的分歧。

### 2.4 panic 终态收口

现状：`launchRun` 的 recover 只写 `job["state"]="failed"` 和 `lastError`，书级 `running/queued` 与 `counts` 未处理。

修复：新增 `failAutomationBooksAfterPanic(job, now, reason)`：

- 仅把 `running`、`queued` 的书改为 `failed`，写入可读原因、`updatedAt`，删除 `retryAt`，并写入 `autoContinue.willContinue=false`。
- 已完成 / 已失败 / 已停止 / 等待人工确认的书保持原状。
- 用 `automationCountsFromBooks` 重算 `counts`，再由 `normalizeAutomationJobCounts` 收敛任务状态。
- 仍受 run epoch 保护：只有当前轮次可以写终态。

### 2.5 后台循环错误日志

`Scanner.Start`、`Migrator.Start`、`CleanupSupervisor.Start` 的 `RunOnce` 错误不再丢弃，统一 `log.Printf` 记录，包含组件名，便于排查数据库不可用、TOS 校验失败、清理停滞。

## 2.6 小说获取队列的僵尸 processing 回收

现场证据（本地 Redis）：`qiantie:novel-fetch:pending = 0`，`processing = 7`，7 个 job 全部没有 lease key（`TTL = -2`），最早的是 10 月任务。消费者退出后 `BRPOPLPUSH` 已把 job 移入 processing，不会再有 `Ack/Fail`，因此这些 job 永久占用 processing 和书籍去重锁。

修复：`RedisQueue.ReapStaleProcessing` 扫描 processing 列表，以 lease key 是否存在作为消费者存活信号；lease 不存在才调用现有 `RecoverStaleJob` 清理。持有效租约的任务一律跳过。调度入口 `dispatchNovelFetchAutomationOwner` 每轮先执行一次回收，并记录回收数量与错误。

## 2.7 入队失败退避

现状：`EnqueueUniqueBooks` 失败时直接把书写成 `waiting_retry`，但没有下次重试时间。扫描器每 20 秒重跑，Redis 持续不可用时会反复改写同一批文档。

修复：新增 `markNovelFetchAutomationEnqueueRetry`，按 `automationEnqueueAttempts` 指数退避（30 秒起，上限 15 分钟），写入 `automationNextRetryAt`，并记录可读错误。`novelFetchAutomationEligible` 已有的截止时间判断会让这批书在到期前退出调度。入队成功时由 `markNovelFetchTaskQueued` 清除退避字段。

## 2.8 队列计数口径

`pending_count` 当前等于文档层 `queued + waiting_retry`，表示“待执行的书”，不等于 Redis `pending` 列表长度。两者本就不同：调度每轮最多提交 50 本，且等待重试的书有退避窗口。本轮不改前端展示，避免在无法完整验证 UI 的情况下引入回归；后端已具备在需要时补充真实 `queue_pending` 字段的位置。

## 2.9 重试预算耗尽必须转人工

对照运行正常的 Python 参考实现（`批量原文改文系统_源码`）：它用 `schedule_task_retry` 记录 `attempt` / `next_retry_at` / `manual_only`，重试预算耗尽后任务转为人工处理，而不是无限等待。

Go 版原先只做退避，没有上限；而 `novelFetchAutomationEligible` 又把 `status=failed` 视为可再次入队，导致“停止自动重试”不成立。修复：

- `novelFetchEnqueueMaxAttempts = 20`，入队失败达到上限后写 `status=failed`、`manualOnly=true`，并删除 `automationNextRetryAt`。
- `novelFetchAutomationEligible` 在入口优先检查 `manualOnly`，为真直接返回 false，避免被扫描器立刻捡回。
- 错误文案明确写“已停止自动重试”，便于用户判断需要人工介入。
- 入队成功后仍清除退避字段与计数。
- 人工重试入口 `POST /api/novel-fetch-workshop/tasks/batch-retry` 在入队前清除 `manualOnly`、失败计数和退避时间，让用户主动重试后能重新进入自动调度；否则人工标记会永久残留。

### 2.10 模型确定性失败转人工（对照参考实现）

参考实现 `Batch原文档改文` 的 `check_queue_stall_regression.py` 断言：AI 阶段确定性失败时必须写 `manual_only` 与 `manual_ai_retry_required`，并让 `task_auto_queue_allowed()` 返回 false，禁止自动队列继续重试。

Go 版补同等语义：

- `novelFetchAIManualRetryRequired` 识别长度截断、内容审核、模型拒答等确定性错误。
- 命中后写 `status=failed`、`queueState=failed`、`manualOnly=true`、`manualAiRetryRequired=true` 和可读原因，并清除 `automationNextRetryAt`。
- 这类失败不再进入 `waiting_retry` 循环，用户点批量重试时由 2.9 的复位逻辑清除标记。
- 网络抖动、限流等瞬态错误仍走原有自动重试，不改变行为。

### 2.11 原文获取重试预算

线上 131 本 `waiting_retry` 中，绝大多数是原文获取返回 `121 upstream business error 400: 获取书籍信息失败`。这类 400 属于确定性错误，但原先 `originalFetchRetryAttempts` 只增不减，会按 1 小时退避无限重试。

- 新增 `novelFetchOriginalFetchMaxAttempts = 5`。
- 达到上限写 `status=failed`、`manualOnly=true`、`manualOriginalFetchRequired=true` 和可读原因，并删除 `originalFetchNextRetryAt`。
- 未达上限继续按配置间隔退避，保留源站临时故障的自动恢复能力。
- 人工批量重试时同时清除 `originalFetchRetryAttempts`、`originalFetchNextRetryAt`、`originalFetchLastError`，让计数从零开始。

### 2.12 阶段化重试表、书号解析与平台名称回退

逐项对照参考实现 `批量原文改文系统_源码/main.py`，补齐 Go 侧仍然缺失的三类能力，并修正一处平台来源错误。

**（1）阶段化重试表 `retry_schedule`（参考 `schedule_task_retry` / `clear_task_retry` / `release_ai_manual_retry_hold` / `task_retry_due`）**

- 新增 `novel_fetch_retry_schedule.go`，按 `original_fetch`、`sensitive`、`classify`、`rewrite`、`pipeline` 五个阶段分别记录 `attempt`、`next_retry_at`、`error`、`manual_only`、`updated_at`。
- 退避序列与参考实现一致：`30s / 120s / 300s / 900s / 1800s`；模型输出格式错误（`ai_output_rejected`、`invalid_json`、`incomplete_prose`）使用 `0s` 立即重试。
- 汇总结论字段与参考实现一致：`retry_stage`、`retry_count`、`next_retry_at`、`last_retry_error`，以及人工保持字段 `manual_ai_retry_required` / `manual_ai_retry_stage` / `manual_ai_retry_error`。
- 原文抓取失败、分类待确认、敏感词失败、改文失败都会写入对应阶段；队列自身重试次数用尽后也写入退避时间，避免 `status=failed` 被扫描器立刻重新入队。
- `novelFetchAutomationEligible` 在入口检查阶段退避：未到期或存在 `manual_only` 阶段时拒绝自动入队。
- 用户主动重试时由 `resetNovelFetchAutomationRetryState` 清空整张重试表，并删除人工保持字段。
- 任务成功完成时由 `clearNovelFetchOriginalFetchRetry` 顺带清空整张重试表。

**（2）书号提取 `safe_book_id` / `book_id_from_url`**

- 新增 `novelFetchSafeBookID`，提取顺序与参考实现一致：先保留 `u_<scope>_<digits>` 复合书号；再从链接参数 `bookid` / `book_id` / `bookId` 取值；最后回退到任意 5-25 位数字。
- `parseWorkshopDocuments` 改用该函数，避免把链接里的 `chapterId` 等其他数字参数误当书号。
- 含 `book_id=` 的链接行不再被当成表头丢弃；链接行也不再从查询参数推断书名。

**（3）AI 重试延迟提示 `ai_retry_delay_hint_seconds`**

- 新增 `novelFetchAIRetryDelayHintSeconds`，从错误文本中提取“约/请 X 秒后重试”的等待时间，作为阶段退避的下限。
- 参考实现的该文案来自其 AI 熔断器；Go 侧当前没有熔断文案，因此取不到时返回 0，退避仍按标准序列执行，不引入额外状态机。

**（4）平台自动识别（修正）**

参考实现没有 `platformAutoDetected` 字段，也没有从链接推断平台；它使用的是 `web_platform_id(meta)`：优先 `platform_id`，缺失时按 `platform_name` 在平台目录中反查 ID。

Go 版原先在原文抓取、自动调度、批量重试三处直接 `firstText(platformId, platform_id, "2")`，只有平台名称的任务会被错误地发到默认书城（`2`）。现新增 `novelFetchResolvedPlatformID` 并替换上述三处：显式 ID 优先，其次按平台目录名称匹配，最后才回退 `2`。

## 3. 验收

1. `localartifact`：写保护开启时 `Restore` 返回 `ErrDiskPressure` 且不产生可见文件；校验失败仍返回 `ErrRemoteIntegrity`。
2. `httpapi`：`SaveJobs` 返回错误时，协调器缓存仍是旧 job；成功后缓存为新 job。
3. `httpapi`：panic 收口后，书级 `running/queued` 变 `failed`，`completed` 保持不变，counts 与任务状态一致。
4. `productionmedia`、`localartifact`、`httpapi` 相关包测试通过。
5. 本地 4001 构建、重启、健康检查通过后，才允许发布生产。
6. 发布记录 Git SHA、release 名称与回滚点。

## 4. 发布与回滚

- 只提交本任务涉及的文件，禁止 `git add .`。
- 使用 `scripts/deploy-go-embedded.ps1` 走 Go embed 发布，不恢复 Node 链路。
- 上一稳定版本作为回滚点，发布后校验 `/healthz` 与服务状态。
- 旧版本文件不归“制作文件保留时长”管理。发布成功后由 `scripts/prune-qiantie-releases.sh` 清理：保留当前软链版本、`runtime-release.json` 记录的回滚版本和最近 N 个历史版本（默认 N=1）。`scripts/deploy-go-embedded.ps1 -Deploy` 会自动调用，可用 `-KeepOldReleases` 关闭，或用 `-KeepPreviousReleases` 调整保留数量。
- 2026-11-14 已按该规则清理一次：`/opt/www/qiantie` 从 15 个二进制（2.7G）降到 3 个（621M），磁盘占用从 73% 降到 62%，当前版本与服务健康不受影响。

## 5. 本轮发布记录（2026-11-14）

- Git SHA：`3c6653aca5604a849b038e9245e5f41fa3045812`（`v88`）。
- 发布编号：`qiantie-202611141140`，二进制 `/opt/www/qiantie/qiantie-202611141140`。
- 二进制 SHA256：`2d100cd9eabfa5761cf91b23af288067cf45b1ef6d01e8e941b2e5cc8c74de80`。
- 发布包：`.release/go-embedded/20261114-113652-3c6653aca560`（`-SkipFrontendBuild`，沿用此前已验收的前端产物）。
- 本地验证：Go 单元/集成测试通过；本地 4011 实际启动，API 路由正常且未登录返回 401。
- 生产验证：`systemctl is-active qiantie` = active；`/api/runtime-build-info` 返回 `git_sha`、`runtime_manifest_sha` 均匹配；公网 `https://yzcm.site/api/runtime-build-info` 200。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141140-3c6653ac`，上一版本 `qiantie-202611140410`。

### 5.1 阶段化重试表发布记录（2026-11-14）

- Git SHA：`3500e244dce18672a6c7518715c1d61e096e17bd`（`v88`）。
- 发布编号：`qiantie-202611141502`，二进制 `/opt/www/qiantie/qiantie-202611141502`。
- 二进制 SHA256：`c1d0fa1e4ed2f58e1b02fe5444ba79d4ff61531f51a6875f552544333428a408`。
- 发布包：`.release/go-embedded/20261114-145846-3500e244dce1`（`-SkipFrontendBuild`，前端源码未改）。
- 本地验证：`go build ./...` 通过；`go test ./internal/httpapi/ -run "NovelFetch|Workshop|Automation"` 通过；本地 4001 重启后 `/healthz` 200，`/api/runtime-build-info` 的 `git_sha`、`runtime_manifest_sha` 均为 `3500e244`。
- 生产验证：`systemctl is-active qiantie` = active；`/healthz` 200；`/api/runtime-build-info` 的 `git_sha` 为 `3500e244dce1…`、`runtime_manifest_sha` 为 `3500e244`；公网 `https://yzcm.site/api/runtime-build-info` 200。
- 回滚点：`/opt/www/qiantie/qiantie-202611141410`（上一稳定版本，保留在原路径），`runtime-release.json` 的 `rollbackBinary` 已记录该版本。



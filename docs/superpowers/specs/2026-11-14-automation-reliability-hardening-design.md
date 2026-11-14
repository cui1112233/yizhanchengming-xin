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
- 入队成功后仍清除退避字段与计数；人工重试时需由重试入口清除 `manualOnly`。

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

## 5. 本轮发布记录（2026-11-14）

- Git SHA：`3c6653aca5604a849b038e9245e5f41fa3045812`（`v88`）。
- 发布编号：`qiantie-202611141140`，二进制 `/opt/www/qiantie/qiantie-202611141140`。
- 二进制 SHA256：`2d100cd9eabfa5761cf91b23af288067cf45b1ef6d01e8e941b2e5cc8c74de80`。
- 发布包：`.release/go-embedded/20261114-113652-3c6653aca560`（`-SkipFrontendBuild`，沿用此前已验收的前端产物）。
- 本地验证：Go 单元/集成测试通过；本地 4011 实际启动，API 路由正常且未登录返回 401。
- 生产验证：`systemctl is-active qiantie` = active；`/api/runtime-build-info` 返回 `git_sha`、`runtime_manifest_sha` 均匹配；公网 `https://yzcm.site/api/runtime-build-info` 200。
- 回滚点：`/opt/www/qiantie/rollback-qiantie-202611141140-3c6653ac`，上一版本 `qiantie-202611140410`。


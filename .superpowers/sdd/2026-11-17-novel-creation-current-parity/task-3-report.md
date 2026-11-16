# Task 3 实施与验证报告

日期：2026-11-17。基线：ebcf9d39c64eb48d22c7da0a7df2296bc7a16ce8；分支：v88。直接在现有分支实施，未创建分支/工作区，未改无关脏文件。未调用付费模型或真实上游，未重启4001，未部署、推送或运行Actions。

## 已完成范围

- 消息、会话游标/current_run、run、首事件、持久待派发记录在一次项目事务内提交；同幂等键不同输入明确冲突。确认蓝图、源状态和后继run同事务；确认后的重复请求返回原后继，内容不同冲突。
- MySQL保存执行槽、90秒租约、30秒续租与Fence。所有运行写入经受保护的项目事务，校验owner/run/project、执行记录/作品槽版本、数据库时间、取消标记；写入结束再次检查有效性。公开Service.Run也必须领取。
- app删除直接后台Run启动；HTTP/定时/批量只通知。启动worker扫描数据库，无浏览器、无Redis队列内容也能恢复。Redis通知故障不能把已保存任务改成失败。
- prepared → may_have_sent → completed/rejected 调用日志，稳定键区分提示词、审核/修订轮数和JSON修复。成功原文先落盘再展示/解析；整条流水线可重放已成功步骤。未知结果暂停，不通过普通重试再次收费。
- 旧queued无日志可领取；旧active无日志暂停，不猜测“未发送”；awaiting_confirmation保持检查点。该恢复由worker首次扫描逐项完成，不运行一次性外部迁移脚本。
- 完整bundle先存save_pending，保存失败只重存；最终提交检查原章节预期版本/前文版本且章节版本不可变，提交指纹保证重复保存不覆盖历史。检查点蓝图/状态/事件原子提交。
- 取消先持久化，再打断本地上下文、发辅助通知；跨进程每秒读取取消标记。取消释放作品槽，过期/被取消执行者不能写回。
- 小说专用单次HTTP策略覆盖OpenAI兼容和渠道provider的传输错误、读取错误、429/5xx、400参数降级、HTTP/2降级及重定向；其他模块保留原策略。明确400/401/403/404/422/429拒绝保存为rejected并给出原有可操作错误，不自动重发；连接/读取/5xx等不确定结果保守暂停。
- 排队时冻结项目写作设置，重启/重放不会因为设置被编辑改变请求指纹。

## 实际红测与修正

以下为本次工具实际输出摘要，不是预期输出：

1. `go test ./internal/novelcreation -run 'TestWorkerAtomic' -count=1`
   - `same idempotency key accepted different content`
   - `atomic request left 11 messages`（12个相同请求并发）
   - 修正后同键仅1条消息，不同内容冲突。
2. `go test ./internal/batchfactoryv11 -run TestNovelSingleAttemptHTTPFailures -count=1`
   - 400：`novel sent 2 HTTP requests`
   - 429、500：各`novel sent 3 HTTP requests`
   - 307：`novel sent 30 HTTP requests`
   - 修正后上述每种均1次POST；传输、读取、HTTP/2错误同样只尝试一次，原模块400适配和HTTP/2重试测试保留通过。
3. `go test ./internal/novelcreation -run 'TestWorkerCancelReleases|TestWorkerMemoryBundle|TestWorkerSavePending' -count=1`
   - `cancelled run still blocks project: 作品已有正在执行的任务`
   - `memory bundle did not persist terminal run like SQL`
   - 修正取消释放槽与内存终态行为后通过。
4. `go test ./internal/novelcreation -run TestWorkerQueuedSettingsAreImmutable -count=1`
   - 排队时1800字/1轮/原规则，被运行前编辑成4000字/4轮/新规则；红测发现漂移，冻结快照后通过。
5. `go test ./internal/novelcreation -run TestWorkerRedisOfflineDoesNotLoseSavedRun -count=1`
   - `Redis notification controls durable execution: failed 读取取消状态失败: redis offline`
   - MySQL取消状态作为真相、Redis仅辅助后通过。
6. `go test ./internal/novelcreation -run TestWorkerCheckpointFailure -count=1`
   - `partial blueprint escaped checkpoint transaction: found=true err=<nil>`
   - 合并蓝图/状态/事件事务后通过。
7. `go test ./internal/novelcreation -run TestWorkerKnownRate -count=1`
   - `known rejection misclassified: 上游结果待确认，已暂停自动重试`
   - 明确拒绝单独持久化为rejected后通过，重放拒绝记录也不再次调用。

## 绿色验证

工作目录为 backend。

`go test ./internal/novelcreation -count=1`

最新输出：`ok qiantie/backend/internal/novelcreation 0.080s`。

`go test ./internal/novelcreation ./internal/httpapi ./internal/app ./internal/storage ./internal/batchfactoryv11 -run 'Novel|Worker|Lease|Recovery|OpenAICompatibleProviderRetriesOnce|OpenAICompatibleProviderReportsOriginal' -count=1`

输出：

```
ok qiantie/backend/internal/novelcreation 0.065s
ok qiantie/backend/internal/httpapi 16.151s
ok qiantie/backend/internal/app 0.032s
ok qiantie/backend/internal/storage 0.021s
ok qiantie/backend/internal/batchfactoryv11 0.055s
```

最后新增快照executionState与明确拒绝分类后，另跑小说包完整测试及HTTP小说路由聚焦测试；结果追加在下方。

`go test ./internal/httpapi -run 'NovelCreation|NovelIntegration' -count=1`：`ok qiantie/backend/internal/httpapi 0.064s`。最后一轮小说包完整测试：`ok qiantie/backend/internal/novelcreation 0.080s`。指定文件的`git diff --check`退出码0（仅Windows行尾提示）。

验证覆盖：并发领取只调用一次、过期Fence禁止事件/蓝图/章节写入、租约续期/过期拒绝、未知发送暂停、prepared恢复、旧active/queued迁移、9步骤完整重放（包含修复和多轮审核，重启后请求计数仍9）、保存失败只重存、取消即时打断和取消竞态、入口原子重复、确认幂等、检查点原子回滚、最终提交幂等和历史不可变、启动扫描。

## SQL证据与限制

- SQLMock实际覆盖项目FOR UPDATE事务边界、AppendMessage嵌套保存不提前提交、后续失败整体回滚、owner/run定位调用日志恢复、novel_run_steps镜像写入、数据库时间下过期Fence在任何写入前回滚。
- 并发竞争、取消、进程重建式恢复和完整流水线重放使用MemoryStore执行，内存提交语义与MySQL事务保存一致；这不等于真实MySQL多进程实测。本次未连接真实MySQL，未安装/执行新迁移到4001运行库。
- `go test -race ...`首次返回`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`；启用CGO后返回`cgo: C compiler "gcc" not found`并build failed。未安装编译器，不能声称race通过。
- 迁移8800312已核查无编号冲突；schema测试更新迁移数量并检查新表，8800310与8800311保持。
- 上游未知结果没有实现“再次付费确认”入口；普通RetryRun明确拒绝。后续界面须展示待确认状态，不能把它计为可自动重试。
- 不声明整体一比一完成、可见浏览器验收或本地/公网运行交付。

## Task4可依赖的接口/状态

- `StartWorker(ctx, service, interval)`：数据库扫描执行器；`Service.Notify(ctx, owner, runID)`：只唤醒；现有HTTP starter、定时、批量共用此入口。
- `Service.Run`强制领取；`SendMessage`、`RetryRun`、`ConfirmRun`保留原签名并改为原子入口。
- run及session的Snapshot新增 `executionState`，不暴露调用原文、请求指纹或账户资料。
- `executionState=ready`等待领取；`leased`正在执行；`save_pending`已生成完整结果，等待保存；`paused`上游结果未知/旧active不明；`done`已结束或已到检查点。
- 未知结果：`run.status=paused`，`errorCode=novel_upstream_result_unknown`，文案“上游结果待确认，已暂停自动重试”。不能自动RetryRun。
- 保存失败：`executionState=save_pending`，run保留最后步骤状态；RetryRun返回同run，worker只保存。不能计成“正文未生成”或新建付费任务。
- 检查点：`run.status=awaiting_confirmation`且`executionState=done`；这不表示正文完成，重复ConfirmRun返回同一个后继run。
- 取消：`run.status=cancelled`，持久cancel_requested及本地取消；文案保留“已停止后续步骤，已发送请求可能仍由上游处理”。
- 明确拒绝：调用记录`rejected`，run按原有模型错误码failed；已成功调用`completed`，没有确定结果的调用`may_have_sent`。

Task4–8的界面/统计口径、自动化后续规则、归档恢复、更多端到端验收仍由后续任务完成。Task6沿用project_retention_records(module=novel_creation)，可在Store.Atomic/领取处接入归档可写性检查；本任务未另建归档体系。

## 修改文件与提交

- 新增小说worker.go、worker_store.go、worker_test.go、worker_mysql_test.go。
- 修改小说Store、MemoryStore、MySQLStore、Service、Pipeline、continuity快照和validation，以及对应已有测试。
- app.go改用StartWorker/Notify；storage/novel_creation_schema.go与测试新增8800312。
- batchfactoryv11新增single_attempt.go/director_single_attempt_test.go并修改director_provider.go；httpapi修改integration_channel_exec.go，新增novel_single_attempt_test.go。
- 本预检和本报告随Task3代码同一次提交；仅列出的本任务文件加入提交，未夹带其他工作。

提交标识将在提交后由主代理的交接消息记录（报告位于同一提交）。

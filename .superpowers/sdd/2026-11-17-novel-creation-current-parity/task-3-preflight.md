# Task3只读预检结论

来源：/root/novel_worker_design_review；未改业务代码、未跑测试或调用模型。按已批准具体规格：MySQL正式任务/待派发真相，Redis通知与协调。

## 必须解决的边界

- SendMessage当前消息、run、事件分开保存，并发幂等键会留下重复消息；创建run与更新session.current_run也不原子。
- 当前Redis RPUSH没有生产消费者；app.go starter直接Run才是真正执行者。应改为持久待派发领取，starter仅唤醒。
- Redis锁15分钟无续租，Run可45分钟；必须90秒数据库租约/30秒续租及fencing，旧执行者不能提交结果。
- MySQL SaveChapterBundle已原子保存章节/版本/记忆/终态，但没有取消/lease/fence/预期版本检查；随后Service.finish又改一次终态，通知失败会误报持久失败。
- MemoryStore.SaveChapterBundle与MySQL正式终态行为不一致，应等价。
- 仅按RunStatus作步骤键不够：foundation与chapter_plan都planning，audit/revise多轮。固定键须区分foundation、chapter_plan、draft、audit:0、revision:n、audit:n、settlement、memory、各repair。

## 持久记录与方法

- 待派发(owner,run_id)唯一，available_at/ready|leased|paused|done。
- 作品执行槽(owner,project_id)唯一，active_run/fence/lease_until；运行记录run_version/cancel_requested。
- 复用novel_run_steps表，扩展稳定步骤键/轮次/请求指纹/发送状态/原始响应/校验结果；新增原子创建、领取、续租、步骤写入、条件最终提交。
- 关键方法语义：CreateQueuedRun、ClaimNextRun、RenewRunLease、LoadRunSteps、PrepareCall、MarkCallMayHaveSent、CompleteCall、SaveRunArtifact、PersistRunEvent、PauseOrFinishRun、CancelRunAtomically、CommitChapterBundle、ConfirmBlueprintAndQueue。
- 所有步骤写入/事件/蓝图/最终提交均校验run/fence/version/未过期/未取消。
- 保存消息、会话游标/current_run、run、待派发和首事件在同一事务；已有幂等键返回同run，同键不同指纹明确冲突。内存库单互斥区保持等价。

## 发送、恢复与收费保护

- 外部调用前记录prepared，再原子改may_have_sent，然后进入provider；跨系统无法恰好一次，崩溃时必须保守暂停。
- prepared且确定未发送可恢复；may_have_sent无确定结果一律暂停待确认，不能自动重发。
- provider成功结果先落盘再展示/解析/下一步；重新执行pipeline可复用原步骤响应。
- 完整bundle先存save_pending，保存失败只重试保存。
- 取消持久标记并立即取消本地context，通知跨进程取消；下一步仍需DB条件检查。

## 底层隐藏重试必须关闭（小说作用域）

batchfactoryv11/director_provider.go Complete:265–320内有传输、读响应、429/5xx重试与400参数降级；do:356–392协议异常还会换HTTP/2重发。

新增小说专用context单次HTTP策略覆盖所有路径，其他业务默认不变；或每次HTTP尝试均写调用日志。最小方案为小说单次发送，400参数适配明确失败。所有其他resolver/provider也要满足合同，不能仅约束一个OpenAI类型。

## 入口幂等与最终保存

- RetryRun保留retry:<sourceID>唯一，save_pending只恢复保存，未知结果不能把普通重试当明确重新付费确认。
- ConfirmRun应先查询confirm:<sourceID>已有后继，再判断源状态；编辑蓝图/来源状态/后继run/队列/事件同事务。相同确认重复返回同结果，不同内容冲突。
- 最终章节提交必须幂等：同run同产物返回原成功，不改历史；同键不同产物冲突。现有ON DUPLICATE KEY UPDATE与固定Version=1不能替代此保护。
- 持久提交成功后Publish失败只记通知故障；API仍返回已保存结果，轮询可恢复。删除SaveChapterBundle后第二次finish的重复路径。

## 作用域

Task3除计划小说文件外须包含director_provider.go作用域策略及测试，以及integration_channel_exec.go中实际Complete路径核查。不要修改其他模块的默认自动重试行为。Task1上下文补齐另有实施代理负责，不能回滚它。

补充主代理核查：integration_channel_exec.go:215–276的Complete显式只有一次client.Do，没有应用层循环，但HTTP客户端重定向也可能形成额外发送；小说单次调用策略应对重定向明确返回原响应，不修改其他业务共享client实例。禁用重试必须覆盖HTTP2协议降级、400参数fallback、读响应失败、transport错误和状态码重试五种路径。测试用httptest/counter证明默认旧业务行为保留、小说一次尝试只收到一次POST。

历史运行迁移：旧planning/writing等任务没有可靠调用记录，不能自动当作未发送重新生成；应标为待确认。旧queued且能证明尚未发送才可恢复；旧awaiting_confirmation保持检查点。当前SaveChapterBundle终态事务已存在，原子提交后不要再由通知错误改成失败。

## 实施架构落点（2026-11-17）

- 在现有 Store 增加 Atomic、Clock、Get/PutExecution、Get/PutSlot、ListDispatch；MySQL Atomic 以 owner/project 行锁串行化领取、取消和写入，内部既有 AppendMessage/SaveChapterBundle 共用外层事务。内存库在互斥区复制状态，成功才整体替换，因此失败同样回滚。
- 新迁移 8800312（8800310、8800311 保留）新增 novel_run_dispatch 与 novel_project_slots。执行版本是持久 Fence，槽与执行记录必须同时匹配；判断时间来自 MySQL UTC_TIMESTAMP(6)。JSON 执行记录持有完整恢复资料，调用记录同步写入已有 novel_run_steps 表，稳定 step_id 由 run/step 键生成。
- worker 启动扫描持久记录，HTTP/定时/批量 starter 只 Notify；Service.Run 自己也必须领取。90秒/30秒租约；本地取消立刻打断，跨进程每秒读取持久取消标记。Redis不再决定任务是否已经入队或能否恢复。
- 稳定调用键使用后端提示词 ID，审核/修订带轮数，修复带 :repair；原始返回先保存，随后才展示/解析。恢复重复解析已保存原文但不重复发请求；400/401/403/404/422/429 的明确拒绝保存 rejected，其余不确定错误保守暂停。
- 保存最终bundle、提交最终章节为两次独立的受保护事务，后者合并版本、记忆、终态、终态事件与提交指纹；同产物重复提交幂等，不同产物冲突。蓝图/检查点状态/事件也同事务，取消和失败/暂停终态事件同事务。
- Task4 可直接使用快照 executionState；Task6 应把已有 project_retention_records(module=novel_creation) 可写性检查接入 Atomic/领取边界，不新建归档体系。

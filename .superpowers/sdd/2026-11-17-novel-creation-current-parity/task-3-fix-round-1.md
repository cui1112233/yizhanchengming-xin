# Task 3 fix round 1

FIX_BASE：24b066ce86cd451eb0adcf3da9e7834648801c73。分支v88；仅修R1–R4，无分支/工作区新增、真实上游调用、收费、4001重启、部署、推送或编译器安装。其他脏文件保持。

## 对原报告的修正

首轮测试没有覆盖：Build尚未完成初始化就扫描已有任务；上游返回后日志写失败；原文已保存但后续事件写失败；最早64条被繁忙作品占满。首轮报告的测试通过记录属实，但不足以证明这些边界安全。原报告顶部新增修订说明，没有删除原结果。

## R1 启动顺序

- 原代码在创建Service后立即StartWorker，而渠道配置userConfigStore和后续可能失败的准备尚未完成。
- 新增novelCreationStartup，构造时不消费任务。Build完成配置、所有可失败准备和router构造后才调用Start，统一启动worker、scheduler、batch；sync.Once防止重复启动。
- 红测：`go test ./internal/app -run TestNovelWorkerStartup -count=1`实际失败：`novel workers start before configuration/router and fallible startup preparation finish`。
- 绿色测试同时验证Build真实调用位置，以及配置延迟/后续准备延迟期间既有任务保持queued；后续准备失败不会消费任务，全部完成后才扫描，runner读取到已初始化配置。

## R2 调用记录写失败

- guarded持久边界保留本地写入错误类型，不再被模型错误包装抹掉。
- 失败处理重新读取数据库里的调用日志。只要仍有may_have_sent，就原子写paused/novel_upstream_result_unknown；不能凭内存中“看见返回了”封成failed/done。
- 数据库持续不可写时保持原leased记录；到期后仍可扫描，恢复时保守暂停。不写虚假的done。
- finishRun另加最终保护：准备写failed时若数据库仍有不确定调用，改为paused与正确错误码。
- 红测：completed和rejected记录各失败一次时，原代码均产生`uncertain journal sealed: run=failed execution=done code=novel_model_unavailable`。
- 绿色测试验证上述两种失败都暂停且请求仍1次；另验证数据库持续不可写时run非终态、execution仍leased，恢复并到期后能扫描，随后暂停且不再次调用。

## R3 已成功原文后的本地失败

- 输出事件/下一阶段事件等本地持久失败不再统一封成failed/done。已确定的completed/rejected记录保留，execution转ready（已有bundle则save_pending），释放租约后仍由同run恢复。
- 对ready且非终态任务，RetryRun只通知同run，不创建空日志的后继任务。明确业务失败后的重新生成仍走原重试入口。
- 红测：assistant.delta以及下一阶段事件写失败，原代码均出现`successful journal made unrecoverable: done`。
- 绿色测试：两种故障下RetryRun均返回原run；重新创建Service恢复后成功，六步全过程请求总数保持6，已完成步骤没有重复发送。

## R4 扫描公平性

- MySQL ListDispatch在ORDER BY/LIMIT之前以owner/project关联NOT EXISTS排除仍有有效槽租约的作品，保留领取时的最终竞争保护。
- MemoryStore同样先排除忙作品，再按创建时间排序/截取，避免测试替身与生产扫描规则相反。
- 红测：1个有效作品租约、该作品70个排队任务、另1个空闲作品任务，原代码返回`busy project occupied dispatch window: candidates=64`。
- 绿色测试：候选仅空闲作品，worker能执行；SQLMock检查owner/project、数据库当前时间和LIMIT之前的槽过滤查询。

## 验证命令与实际结果

工作目录backend。

红测合并命令：`go test ./internal/novelcreation -run TestWorkerFix -count=1`，R2两个子项、R3两个子项、R4均按上面原因FAIL。

首次修正后：`go test ./internal/novelcreation ./internal/app -run 'WorkerFix|NovelWorkerStartup' -count=1`

```
ok qiantie/backend/internal/novelcreation 0.036s
ok qiantie/backend/internal/app 0.145s
```

补持续数据库故障用例后：`go test ./internal/novelcreation -count=1`

```
ok qiantie/backend/internal/novelcreation 0.078s
```

`go test ./internal/app ./internal/httpapi -run 'NovelCreation|NovelWorkerStartup|NovelIntegration' -count=1`

```
ok qiantie/backend/internal/app 0.129s
ok qiantie/backend/internal/httpapi 0.049s
```

最后指定相关包聚焦检查结果追加在下方。没有重复全仓库套件；race限制沿用首轮已核实的CGO/gcc缺失，本轮未安装工具或重复尝试。

`go test ./internal/novelcreation ./internal/httpapi ./internal/app -run 'NovelCreation|NovelWorkerStartup|NovelIntegration|Worker|Lease|Recovery' -count=1`

```
ok qiantie/backend/internal/novelcreation 0.073s
ok qiantie/backend/internal/httpapi 6.889s
ok qiantie/backend/internal/app 0.129s
```

本轮指定文件`git diff --check`退出码0，仅有Windows行尾提示。

## 文件与边界

本轮修改app.go、service.go、worker.go、worker_store.go、worker_mysql_test.go；新增app/novel_startup.go、app/novel_startup_test.go、novelcreation/worker_fix_round1_test.go；更新首轮报告并新增本报告。

SQL证据仍为SQLMock；故障注入、并发扫描、恢复和启动延迟用MemoryStore，无真实MySQL多进程或已部署运行证据。Task4接口保持Snapshot.executionState不变；本地持久失败现在表现为ready后同run恢复，未知仍paused。Task4–8及最终浏览器/运行验收未在本轮宣称完成。

本报告与本轮代码同次提交，提交SHA在交接消息给出。

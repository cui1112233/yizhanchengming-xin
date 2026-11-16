# Task 2 修复轮次 1：会话 SQL 与批次停止失败

日期：2026-11-17。修复基线：`d7182e1c5cae319450c4f9f2ed1537aada988c60`。

实现提交：`5eb713e9b39edd6266514c351e2989c912c3d762`。仅处理只读审查 R1/R2，等待复审。

## R1：纠正会话查询投影

原实现批量编辑 SQL 时把 `s.input_snapshot_json` 误加进 `sessionColumns`。迁移只给 schedule/batch 加字段，且 `scanSession` 仅扫描九列，因此真实 MySQL 的 GetSession/ListSessions 会失败。这是本轮修复的 Critical 问题，原报告的绿测没有覆盖这条 SQL 路径，不能作为其验收证据。

先新增 `TestAutomationInputSQLSessionQueriesUseOnlySessionSchemaColumns`，独立写出九列 SQL 期望，不复用被测 `sessionColumns` 常量。覆盖 GetSession、按作品 ListSessions、全账号 ListSessions，返回九列并实际经过 `database/sql` Rows/Row 的 Scan，核对全部 Session 字段。

红测：三项全部失败，实际 SQL 中额外出现 `s.input_snapshot_json`，与独立期望的会话表列投影不符。修复仅移除误列，保留计划与批次 JSON 列及迁移。三项转绿。

## R2：批次停止保存错误可观察

先新增：

- `TestAutomationInputBatchStopPersistenceErrorIsObservable`：旧续写批次缺前文，停止写库注入失败。要求领域输入错误和持久失败均保留、扫描返回批次 ID 与错误；写库恢复后应持久失败状态及“配置需调整”，再扫描不重复处理。
- `TestAutomationInputBatchScanReturnsSuccessesAndAllFailures`：两个无效批次与一个正常批次同时扫描，要求返回所有失败批次 ID，同时正常批次仍派发。

红测分别失败于 `stop failed but combined input/save error was lost` 和 `scan swallowed per-batch failures`。

修复：`advanceBatch` 使用 `errors.Join` 返回原输入错误及停止保存错误；`AdvanceBatches` 继续扫描邻居、按批次 ID 聚合错误并返回成功派发。已有 StartBatchWorker 会输出返回错误并启动成功邻居，本次未改 worker、阈值、计数、领取或重试策略。

数据库持续不可写时，无法伪称停止已持久保存；后续扫描仍可能再次尝试保存停止，但现在每次失败均可观察，且缺输入路径不产生模型调用。写库恢复后只持久停止一次，不再扫描该批次。

## 红绿与回归证据

backend 目录执行：

```text
go test ./internal/novelcreation -run 'AutomationInputSQLSession|AutomationInputBatchStopPersistence|AutomationInputBatchScan' -count=1
```

修复前：三条会话 SQL 子用例加两条批次测试失败。修复后：`ok qiantie/backend/internal/novelcreation 0.018s`。

最终指定回归：

```text
go test ./internal/novelcreation ./internal/httpapi ./internal/storage -run 'NovelCreation|Schedule|Batch|ModeInput|AutomationInput' -count=1
ok qiantie/backend/internal/novelcreation 0.031s
ok qiantie/backend/internal/httpapi 18.479s
ok qiantie/backend/internal/storage 0.021s
```

`git diff --check` 对本轮四个文件通过。日志：

- `_local-archive-20261028/novel-parity-20261117/task2-fix1-red.log`
- `_local-archive-20261028/novel-parity-20261117/task2-fix1-green.log`
- `_local-archive-20261028/novel-parity-20261117/task2-fix1-regression.log`

修改文件：

- `backend/internal/novelcreation/mysql_store.go`
- `backend/internal/novelcreation/batch.go`
- `backend/internal/novelcreation/automation_input_mysql_test.go`
- `backend/internal/novelcreation/automation_input_test.go`

限制与范围：会话 SQL 投影和真实 Scan 路径使用 SQLMock 验证，未连接真实 MySQL、未执行迁移或重启运行包；后续本地运行验收仍须检查实际数据库。没有把原报告的包测试通过追认为真实数据库成功。保持 v88，只提交以上文件与本报告，其他脏文件未提交；没有收费调用、4001 更新或部署。

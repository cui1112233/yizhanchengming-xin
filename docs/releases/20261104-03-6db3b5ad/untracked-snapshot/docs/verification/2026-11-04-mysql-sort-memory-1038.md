# MySQL Error 1038 `Out of sort memory`（中断任务恢复扫描）— 待修记录

- 记录时间：2026-11-04 12:xx
- 状态：**已记录，未修复**（按发布要求不改 SQL、不改查询）
- 严重度：中（不影响启动与 `/health`，但"中断任务恢复"实际未执行）

## 1. 现象

每次启动带 MySQL 自动化存储的应用实例，启动日志固定出现：

```text
recover interrupted batch automation jobs: iterate automation jobs: Error 1038 (HY001): Out of sort memory, consider increasing server sort buffer size
skip resuming interrupted batch automation jobs: QIANTIE_AUTOMATION_RESUME_ON_BOOT is not enabled
```

之后进程仍正常 `listening`。

复现次数：4002 临时实例两次（11:59:45、12:05:07）均复现；4001 重启（12:07）同样走该启动路径。

## 2. 复现条件

- 触发点：`router.go` 启动段 `RecoverInterruptedJobs()`（约 L158），内部调用
  `automation_mysql_store.go:36` 的 `Load()`：
  ```sql
  SELECT id, job_json FROM batch_automation_jobs ORDER BY updated_at, id
  ```
- 必要配置：`QIANTIE_MYSQL_DSN` 指向使用 `qiantie_go_local` 的 MySQL 自动化存储。
- 表规模（本机实测）：

| 指标 | 实测值 |
| --- | --- |
| 行数 | 47 |
| `job_json` 总字节 | 844,954 B |
| 单行最大 `job_json` | 397,696 B |
| `state='running'` 行数 | 0 |
| `sort_buffer_size` | 262,144（256 KB） |
| `max_sort_length` | 1,024 |
| MySQL 版本 | 8.4.11 |

- 索引：`PRIMARY(id)`、`idx_batch_automation_due(state,scheduled_at,next_retry_at)`、
  `idx_batch_automation_owner_batch(owner_username,batch_id,updated_at)`。
  **没有** `(updated_at,id)` 前导索引，`ORDER BY updated_at, id` 必须 filesort。
- 结论：行数很少，但 `job_json` 是 `JSON` 大字段（平均约 18 KB/行、最大约 389 KB/行），
  filesort 需要把待排序行装入 sort buffer，总量约 845 KB > 256 KB，触发 1038。

## 3. 影响面

1. `RecoverInterruptedJobs()` 返回错误 → **不会把遗留的 `running` 任务标记为中断/停止**。
   若某个批次在上一进程崩溃时停在 `running`，重启后该状态会一直保留，新的立即运行可能被同批次任务挡住。
2. `loadLocked()` 在 `automationStore.Load()` 出错时直接返回 error（`batch_factory_automation.go:505-508`），
   启动阶段 `c.automationState` 保持 nil；后续任何依赖它读取任务状态/调度的路径会继续拿到同一错误，
   即**计划调度器可能无法加载任务、定时自动化不会派发**。
3. 当前数据：`state='running'` 行数为 0，所以此刻**没有任务被卡住**，属于"打开就报错"的潜在问题，
   一旦有任务在运行中重启就会真实丢恢复。

## 4. 建议方向（均未实施，待排期）

1. **调大排序缓冲**：MySQL 侧 `sort_buffer_size` 由 256 KB 提高到 1 MB+（会话级或全局），
   仅缓解，不解决"数据继续增长会再次触发"。
2. **把整表遍历改成分批 + 显式 ORDER BY 走索引**：例如按 `updated_at` 分页
   （`WHERE (updated_at,id) > (:lastUpdatedAt,:lastId) ORDER BY updated_at,id LIMIT :n`），
   并新增 `(updated_at,id)` 索引，避免一次对全部大 JSON 行做 filesort。
3. **限制返回列**：启动恢复只需要 `id,state,updated_at`（判断是否有 `running` 行），
   不要在这一步取 `job_json`；JSON 内容按需逐条读取。

## 5. 相关位置

- `backend/internal/httpapi/router.go`（约 L158-162）：启动调用与日志。
- `backend/internal/httpapi/automation_mysql_store.go:36`：失败 SQL。
- `backend/internal/httpapi/batch_factory_automation.go:499-520`：`loadLocked()` 错误处理。

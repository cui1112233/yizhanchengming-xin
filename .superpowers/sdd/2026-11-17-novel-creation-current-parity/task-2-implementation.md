# Task 2 完整自动执行输入合同实施记录

日期：2026-11-17。范围：仅已批准计划 Task 2，等待只读审查。

## 提交与边界

- 实现提交：`8fa26d6e5d97b994599acde129dd3e5191909fa6`。
- 开工时当前分支 `v88`，HEAD `efa2c6026da11cf71e0e974f0f9a55f91cae1be9`（其中其他聊天文档提交保留）。Task 1 的 `88e791a6`、`0cf83760` 及前文/重试版本保护均保留。
- 直接在 v88 修改，只提交本任务明确列出的 15 个源码/测试文件。未创建分支或工作树，未提交其他聊天的脏文件。
- 未调用收费模型、未部署、未运行迁移、未构建或替换 4001 运行包。

## 红测与修复

先新增 `TestAutomationInputFourModesPreserved`，使用 JSON 输入模拟真实合同，针对仿写长篇、仿写短篇、番外、续写分别创建定时/批量并执行。

红测命令（backend）：

```text
go test ./internal/novelcreation -run TestAutomationInputFourModesPreserved -count=1
```

真实结果：8 个子用例全部失败。定时分支报“请填写参考文本 / 请填写既有人物与世界规则 / 请填写被续写的前文”；批量分支报 `automation failed to create run with complete input`。证明自动入口丢掉附加模式输入。最终该测试同时核对运行快照及实际 Pipeline.RunContext 的字段，8 项全部通过。

实施内容：

1. 手动、定时创建/更新、批量创建/继续共用 `prepareGenerationInput`，统一校验会话与作品归属、模式必填、选定资料/备忘录/文风归属。
2. Schedule/Batch 及请求合同增加 `ModeInputs`、`KnowledgeIDs`、`NoteIDs`、`StyleProfileID`。计划保存选择，每个新运行重新校验并生成所用内容快照；已接受运行、蓝图确认后的继续及手动重试保持原快照。
3. 每个运行保存 `mode_inputs`、资料 ID 与有界正文、备忘录 ID 与内容、文风 ID 与实际正文。Pipeline 从该快照读取备注和文风，并把备注注入策划/正文/审核请求。老运行没有文风快照时保留已有项目设置回退。
4. 仿写支持参考文本、有效选定参考资料、有效文风三条入口；只选择文风可用。旧项目 `ProjectSettings.StyleProfile` 仍保留，显式选择新的文风 ID 会覆盖本轮使用的文风。
5. 经主代理确认，提前提供最小真实文风实体和账号隔离的保存/读取接口：`POST /api/novel-creation/style-profiles`、`GET /api/novel-creation/style-profiles/{styleID}`。POST 带已存在且本人拥有的 ID 可编辑；不做导入、统计、文件管理或前端界面。
6. 旧定时计划缺必填输入时 `Enabled=false`，持久保存 `configurationError=配置需调整：...`，后续扫描不再循环执行。旧批次/已失效选择在派发时失败停止并保存可读原因，继续前再次校验。有效编辑定时计划后清除配置提示。
7. MemoryStore 对计划的 map、slice、可选时间做深拷贝；资料、备忘录、计划列表补齐账号过滤，避免同名作品 ID 的账号混淆。
8. 新迁移使用 `8800311`：新增两表 `input_snapshot_json JSON NULL` 和账号隔离的 `novel_style_profiles`。`8800310` 已被 `image_capacity_schema.go` 占用，故未使用。旧四个迁移原文不变，新测试检查全局迁移版本不重复。

## 验证

最终指定命令通过：

```text
go test ./internal/novelcreation ./internal/httpapi ./internal/storage -run 'NovelCreation|Schedule|Batch|ModeInput|AutomationInput' -count=1
ok qiantie/backend/internal/novelcreation 0.030s
ok qiantie/backend/internal/httpapi 18.658s
ok qiantie/backend/internal/storage 0.019s
```

完整领域回归通过：

```text
go test ./internal/novelcreation -count=1
ok qiantie/backend/internal/novelcreation 0.029s
```

最终输出保存在 `_local-archive-20261028/novel-parity-20261117/task2-focused-tests.log`。`git diff --check` 对本任务文件通过。

测试覆盖：

- 四种附加输入模式 × 定时/批量的实际运行输入；缺必填拒绝且不保存空计划。
- 资料/备忘录/文风跨账号拒绝，其他账号无法读取或覆盖文风，作品/会话归属校验。
- 文风单独选择、旧项目默认文风快照；源资料编辑/删除后原运行和重试仍使用原内容；后续新运行拒绝失效引用。
- 批次暂停/继续选择不变，蓝图确认保持附加输入和全部内容快照（按 JSON 语义比较）。
- 旧定时缺输入只失败一次后停用，修复仍做必填校验，有效修复可恢复。
- SQLMock 检查 SQL 创建、更新、读取中的完整 JSON 和 owner 条件，兼容旧 NULL JSON、拒绝损坏合同；MemoryStore 深拷贝覆盖创建返回值、调用方及列表返回值。
- HTTP 三入口真实请求解码、定时 PUT 更新、正式运行快照、文风保存/账号隔离读取。

测试开发中曾修正两个测试数据问题：HTTP 请求混放不同入口字段被严格解码器拒绝；蓝图继续测试最初使用不完整蓝图，随后改用现有合法蓝图替身。继续快照比较最初按字节比较被 JSON 键序变化触发，改为语义比较。最终产品代码及测试以以上绿测为准。

## 文件

- `backend/internal/novelcreation/automation_input.go`
- `backend/internal/novelcreation/automation_input_test.go`
- `backend/internal/novelcreation/automation_input_mysql_test.go`
- `backend/internal/novelcreation/types.go`
- `backend/internal/novelcreation/store.go`
- `backend/internal/novelcreation/service.go`
- `backend/internal/novelcreation/pipeline.go`
- `backend/internal/novelcreation/scheduler.go`
- `backend/internal/novelcreation/batch.go`
- `backend/internal/novelcreation/memory_store.go`
- `backend/internal/novelcreation/mysql_store.go`
- `backend/internal/httpapi/novel_creation_routes.go`
- `backend/internal/httpapi/novel_creation_automation_input_test.go`
- `backend/internal/storage/novel_creation_schema.go`
- `backend/internal/storage/novel_creation_schema_test.go`

## 留给后续任务的限制

- SQL 为 SQLMock 和迁移结构测试，未在真实 MySQL 上执行本次迁移；Task 8 本地交付需正式运行迁移并验证数据库持久结果。
- Task 3 的事务入队、待派发持久化、领取/租约/取消恢复未在本任务实现。
- Task 4 的日限、失败阈值、完整状态机、跨进程轮次去重，以及立即执行不改变原定时间尚未实现；现有基于秒的定时幂等键和推进 next 时间保留。
- 新前端合同接入、完整 DTO 回传（包括定时启用切换不清空选择）归 Task 7；当前页面尚未验收这些字段。
- 文风文件导入、统计、管理界面和完整资料选章版本管理归 Task 6。此任务的最小实体保存真实内容，不伪造导入或分析能力。
- 来源轮次的完整持久追踪、Agent/提示词参数版本快照等全规格内容由后续任务补齐；本任务只完成模式及所选资料输入合同，不宣称整个小说创作链路完成。

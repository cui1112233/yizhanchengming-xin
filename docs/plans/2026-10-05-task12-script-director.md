# Task 12 剧本 / Hook / Director / Final Prompt 迁移计划

日期：2026-10-05
目标分支：`feat/task12-script-director`
范围：仅 `TASKS.md` Task 12。

## 旧 v88 审计结论

旧 v88 不是空白实现，优先复用以下 Go 业务资产与测试语义：

- `backend/internal/batchfactoryv11/director_service.go`
  - working-front 作为非破坏性生产文本；原始 source 不被覆盖。
  - viral 模式先 Hook，Hook 必须 approved 后 Director 才能继续。
  - Director 的模型格式错误属于可重试失败。
- `backend/internal/batchfactoryv11/h3_director_service.go`
  - H3 使用结构化导演文档，输入包含冻结的视频源 revision/hash 与 prompt preset revision。
  - H3 Director 与最终 VIDEO renderer 分离；Director 不负责 Task 13 的精确 matchAudio。
- `backend/internal/batchfactoryv11/stage_runs.go` / `stage_service.go`
  - Stage 每次执行保留 attempt、requestId、inputRevision、errorMessage。
  - Retry 只重跑“最新仍未恢复的失败 Stage”；后续成功后历史失败不再可重试。
  - 目标 Stage 强制重跑时复用已经完成的前置产物。
- `backend/internal/batchfactoryv11/final_prompt.go`
  - Final Prompt 在 Go 后端编译。
  - 旧 VIDEO 编译稳定顺序为：画面前缀 -> 基础设定 -> 画面约束 -> 视频提示词 -> 画面限制 -> 负面提示词。
  - 设置层按 system -> config version -> batch -> book -> video 覆盖，输出 snapshot hash。

本任务在新主线按更明确的生成域收敛：

`Book -> SCRIPT -> HOOK -> DIRECTOR -> FINAL_PROMPT`

Hook 关闭时记为 `skipped`，不是 `failed`。H3 与普通 Director 共用 StageRun，但 `director_mode` 区分行为。`matchAudio` / `audioDurationSec` 只作为 Director 请求的兼容字段保存和传递；不实现 Task 13 时长重排。

## 数据模型

新增 Goose migration：

1. `generation_prompts`
   - `prompt_key`, `version`, `content`, `enabled`, timestamps
   - `(prompt_key, version)` 唯一
2. `book_runs`
   - 每本书一次生成执行实例，关联 BatchProject / Book
   - aggregate status + timestamps + error summary
3. `stage_runs`
   - `book_run_id`, `book_id`, `stage`, `status`, `attempt`
   - `prompt_key`, `prompt_version`
   - `input_snapshot`, `output_text`, `error_message`
   - started/finished timestamps
4. 索引/唯一约束用于幂等：同一 book_run + stage + attempt 唯一；running Stage 由事务检查防止重复启动。

状态：`pending|running|completed|failed|skipped`。

## 后端包

新增 `api/internal/generation`：

- `model.go`：Stage、BookRun、StageRun、Prompt、请求/结果模型。
- `store.go`：Repository 接口。
- `mysql_store.go`：MySQL 事实源。
- `prompts.go`：Prompt resolver + 默认系统预设 + 通用剧情模式元提示词。
- `provider.go`：文本模型接口与 OpenAI-compatible HTTP 实现；密钥仅从环境读取，错误对前端脱敏。
- `compiler.go`：确定性 Final Prompt Compiler。
- `service.go`：GenerationService，负责单本/批量、Hook 开关、普通/H3 Director、Stage Retry、项目聚合。

Prompt 版本必须写入 StageRun。Final Prompt 相同输入 + 相同 prompt version 产生稳定段落顺序。

## API

按现有 `/api/v1` 风格新增：

- `GET /api/v1/batch-projects/{projectId}/generation`
- `POST /api/v1/batch-projects/{projectId}/generation`
- `GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation`
- `POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation`
- `POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry`
- `GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}`
- `GET /api/v1/generation/prompts`

HTTP 层只负责 decode/encode；业务统一委托 GenerationService。

## 前端

在 Batch Factory 项目详情增加生成状态表：

- 每本书展示 Script / Hook / Director / Final Prompt 状态。
- 单本执行、批量执行、失败 Stage 重试、查看结果/错误。
- 错误只显示脱敏摘要。
- 继续使用 React + Ant Design；状态来自 Go API，不存 LocalStorage。

## TDD 顺序

1. Prompt resolver / 剧情模式 / Final Prompt compiler 的失败测试。
2. Stage 状态、幂等、Hook skipped、单本执行的失败测试。
3. 批量隔离失败、项目聚合、Retry 只跑目标 Stage 的失败测试。
4. MySQL store / migration contract tests。
5. HTTP API tests。
6. React 状态展示与交互 tests。
7. 最小实现逐项转绿。

## 验证

必须以 CI/可执行证据为准：

- `go test ./...`
- 前台 `npm test`
- 前台 `npm run build`
- 后台 `npm run build`
- 后台 Go build（CI 当前 Go job 至少编译所有 package；补显式 build 检查）
- MySQL store contract tests（无可用 MySQL 时必须明确跳过原因，不得虚报通过）
- HTTP handler tests
- PR CI 全绿

在这些条件未全部满足前，`TASKS.md` Task 12 只标 `🟡/⚠️`，不标 `[x]`。
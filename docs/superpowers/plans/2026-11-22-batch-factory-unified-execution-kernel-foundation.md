# Batch Factory Unified Execution Kernel Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立批量工厂统一执行内核的第一阶段：一次点击只对应一个可恢复的执行记录，重复点击复用原任务，配置在启动时冻结，旧自动化任务可被新状态接口读取；本阶段继续复用现有视频/漫画执行器，不改供应商生成逻辑。

**Architecture:** 新增独立的 `batchfactoryexecution` 领域包，以 MySQL 为唯一事实来源；HTTP 层只负责鉴权和入参转换；旧 `BatchAutomationCoordinator` 通过窄接口接入，作为第一阶段的执行适配器。前端新增批量工厂模块 API 和状态适配器，先把“立即执行已选”切到 `ensure` 接口。Redis 只负责唤醒/抢占，不保存最终任务状态；旧任务只读映射，不做双写。

**Tech Stack:** Go 1.x、`net/http`、MySQL 8/InnoDB、Redis、React、Ant Design、Vite、Node test、Go embed。

## Global Constraints

- [ ] 不新建分支，直接在当前 `v88` 主开发分支逐任务提交。
- [ ] 只提交本计划涉及的文件；工作区里其他未提交内容全部保留，不清理、不覆盖、不顺手格式化。
- [ ] 当前真实目录仍是 `backend/` 与 `frontend/`。本阶段在现有目录内建立清晰模块边界，不做一次性目录搬迁；最终后端职责仍遵守“接口在 Go API、提示词只在后端”。
- [ ] MySQL 是执行状态唯一事实来源；Redis 只能用于唤醒、抢占和短时锁，不得保存唯一任务状态。
- [ ] 本阶段不修改系统提示词、不触发付费模型调用、不重做已有媒体。
- [ ] 旧活动任务继续由旧执行器跑；新入口发现旧任务时只建立引用并复用，不返回 409，也不重复提交。
- [ ] 外部上传只在读回确认后才能标记完成；本阶段只建立状态和回执结构，不伪造发布成功。
- [ ] 每个任务先写失败测试，再写最小实现，再运行聚焦测试，最后单独提交。
- [ ] 发布顺序固定：本地 4001 验收 → 推送 GitHub `main` → SSH 部署线上 → 线上读回验收。禁止 GitHub Actions 部署。

---

## File and Responsibility Map

### New backend module

- `backend/internal/batchfactoryexecution/model.go`：执行、书籍、阶段、尝试、回执状态与合法迁移。
- `backend/internal/batchfactoryexecution/fingerprint.go`：配置快照规范化和输入指纹。
- `backend/internal/batchfactoryexecution/repository.go`：事务仓储接口，不泄漏 SQL。
- `backend/internal/batchfactoryexecution/mysql_repository.go`：MySQL 实现、幂等创建、状态读回。
- `backend/internal/batchfactoryexecution/service.go`：`EnsureExecution`、重复点击复用、逐书隔离。
- `backend/internal/batchfactoryexecution/legacy.go`：旧活动任务只读映射及旧启动器窄接口。
- `backend/internal/batchfactoryexecution/outbox.go`：事务内事件与可恢复分发。
- `backend/internal/batchfactoryexecution/*_test.go`：领域、仓储、服务和恢复测试。

### Existing backend integration

- `backend/internal/storage/automation_schema.go`：追加 `8800035` 迁移，不修改任何既有迁移；`8800001`–`8800034` 已被当前仓库占用。
- `backend/internal/storage/automation_schema_test.go`：约束、索引、迁移顺序测试。
- `backend/internal/httpapi/batch_factory_execution_routes.go`：新接口的薄路由。
- `backend/internal/httpapi/batch_factory_execution_adapter.go`：把旧 coordinator 包装成 `LegacyAutomationGateway`。
- `backend/internal/httpapi/router.go`：注入新服务并注册路由。
- `backend/internal/app/app.go`：构造 MySQL 仓储、服务和 outbox dispatcher。

### Frontend module

- `frontend/src/modules/batch-factory/api/executions.js`：`ensure/get/timeline/reconcile` API。
- `frontend/src/modules/batch-factory/model/executionStatus.js`：统一状态到中文展示模型的纯函数。
- `frontend/src/modules/batch-factory/index.js`：模块出口。
- `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`：只替换“立即执行已选”的提交入口和轮询来源。
- `frontend/src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js`：入口不再直调旧 start 的防回归测试。

### Do not change in this phase

- `backend/internal/httpapi/batch_factory_automation.go` 的视频/漫画实际执行阶段。
- 图片、视频、配音、TOS、121/N8 发布供应商实现。
- `frontend/src/user/pages/shuihuo/automationStages.json` 的现有阶段顺序。
- 任何系统预设提示词正文。

---

## Task 1: Add append-only execution schema

**Files:**

- Modify: `backend/internal/storage/automation_schema.go`
- Create: `backend/internal/storage/automation_schema_test.go`
- Test: `backend/internal/storage/migrations_test.go`

**Interfaces:**

```go
func BatchExecutionStatements() []string
func AutomationMigrations() []Migration
```

- [ ] Step 1: Write a failing schema test

测试必须断言存在以下表和唯一约束：

```go
func TestAutomationMigrationsAppendUnifiedExecutionKernel(t *testing.T) {
	got := AutomationMigrations()
	if got[len(got)-1].Version != 8800035 {
		t.Fatalf("last migration version=%d", got[len(got)-1].Version)
	}
	sql := strings.Join(got[len(got)-1].SQL, "\n")
	for _, required := range []string{
		"batch_executions",
		"batch_execution_config_snapshots",
		"batch_execution_books",
		"batch_execution_stage_runs",
		"batch_execution_stage_attempts",
		"batch_execution_publication_receipts",
		"batch_execution_outbox",
		"UNIQUE KEY uq_batch_execution_request",
		"UNIQUE KEY uq_batch_execution_book",
		"UNIQUE KEY uq_batch_execution_stage_fingerprint",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
}
```

- [ ] Step 2: Run the test and prove it fails

Run: `go test ./internal/storage -run 'TestAutomationMigrationsAppendUnifiedExecutionKernel' -count=1`

Expected: FAIL because migration `8800035` and the new tables do not exist.

- [ ] Step 3: Add seven InnoDB tables as migration `8800035`

Minimum columns:

```text
batch_executions:
  id, owner_username, batch_id, mode, target, client_request_id,
  status, legacy_job_id, created_at, updated_at

batch_execution_config_snapshots:
  execution_id, config_json, config_digest, created_at

batch_execution_books:
  id, execution_id, book_id, title, status, current_stage,
  failure_code, failure_message, created_at, updated_at

batch_execution_stage_runs:
  id, execution_book_id, stage_key, input_fingerprint, status,
  next_retry_at, created_at, updated_at

batch_execution_stage_attempts:
  id, stage_run_id, attempt_no, status, provider_request_id,
  error_code, error_message, started_at, finished_at

batch_execution_publication_receipts:
  id, execution_book_id, platform, remote_work_id, remote_status,
  payload_json, confirmed_at, created_at, updated_at

batch_execution_outbox:
  id, aggregate_type, aggregate_id, event_type, payload_json,
  available_at, claimed_at, delivered_at, attempts, last_error, created_at
```

Use unique keys exactly for `(owner_username, batch_id, client_request_id)`, `(execution_id, book_id)`, and `(execution_book_id, stage_key, input_fingerprint)`.

- [ ] Step 4: Run storage tests

Run: `go test ./internal/storage -run 'TestAutomationMigrations|TestAppMigrationsIncludesV11AndV12Migrations' -count=1`

Expected: PASS.

- [ ] Step 5: Commit only schema files

```powershell
git add backend/internal/storage/automation_schema.go backend/internal/storage/automation_schema_test.go
git commit -m "feat(batch): add unified execution schema"
```

---

## Task 2: Define the domain states and legal transitions

**Files:**

- Create: `backend/internal/batchfactoryexecution/model.go`
- Create: `backend/internal/batchfactoryexecution/model_test.go`

**Interfaces:**

```go
type ExecutionStatus string
type BookStatus string
type StageStatus string

func (s ExecutionStatus) CanTransition(to ExecutionStatus) bool
func (s BookStatus) CanTransition(to BookStatus) bool
func (s StageStatus) CanTransition(to StageStatus) bool
func DeriveExecutionStatus(books []Book) ExecutionStatus
```

- [ ] Step 1: Write table-driven failing tests

Cover these rules:

- execution: `planned -> running -> completed|needs_attention|cancelled`.
- book: `pending -> running -> waiting_external|completed|retry_wait|needs_attention|cancelled`.
- stage: `pending -> queued -> running -> succeeded|reconciling|retry_wait|terminal_failed|cancelled`.
- `terminal_failed`, `succeeded`, `completed`, and `cancelled` cannot silently return to running.
- one failed book does not force other books to fail.
- execution is `completed` only when every book is completed or explicitly cancelled.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/batchfactoryexecution -run 'Test.*Transition|TestDeriveExecutionStatus' -count=1`

Expected: FAIL because package does not exist.

- [ ] Step 3: Implement concrete types and validation errors

Use exported constants, for example:

```go
const (
	ExecutionPlanned        ExecutionStatus = "planned"
	ExecutionRunning        ExecutionStatus = "running"
	ExecutionNeedsAttention ExecutionStatus = "needs_attention"
	ExecutionCompleted      ExecutionStatus = "completed"
	ExecutionCancelled      ExecutionStatus = "cancelled"
)
```

`Execution`, `Book`, `StageRun`, `StageAttempt`, and `PublicationReceipt` must use typed status fields and UTC timestamps.

- [ ] Step 4: Run package tests

Run: `go test ./internal/batchfactoryexecution -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/model.go backend/internal/batchfactoryexecution/model_test.go
git commit -m "feat(batch): define execution state machine"
```

---

## Task 3: Freeze configuration and calculate deterministic fingerprints

**Files:**

- Create: `backend/internal/batchfactoryexecution/fingerprint.go`
- Create: `backend/internal/batchfactoryexecution/fingerprint_test.go`

**Interfaces:**

```go
func CanonicalJSON(value any) ([]byte, error)
func ConfigDigest(config any) (string, error)
func StageInputFingerprint(stageKey string, configDigest string, input any) (string, error)
```

- [ ] Step 1: Write failing tests for stable canonicalization

The same map with different insertion order must produce the same SHA-256 digest. Different `stageKey`, prompt version, source text digest, model ID, or asset selection must produce a different fingerprint.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/batchfactoryexecution -run 'TestCanonicalJSON|TestConfigDigest|TestStageInputFingerprint' -count=1`

Expected: FAIL.

- [ ] Step 3: Implement recursive canonical JSON

Rules:

- map keys sorted lexicographically;
- arrays keep order;
- strings, booleans, null, and JSON numbers keep their JSON meaning;
- reject NaN/Infinity and unsupported values;
- return lowercase hex SHA-256.

- [ ] Step 4: Run tests and race detector

Run: `go test -race ./internal/batchfactoryexecution -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/fingerprint.go backend/internal/batchfactoryexecution/fingerprint_test.go
git commit -m "feat(batch): freeze execution inputs"
```

---

## Task 4: Add repository contract and transactional MySQL implementation

**Files:**

- Create: `backend/internal/batchfactoryexecution/repository.go`
- Create: `backend/internal/batchfactoryexecution/mysql_repository.go`
- Create: `backend/internal/batchfactoryexecution/mysql_repository_test.go`

**Interfaces:**

```go
type Repository interface {
	Ensure(context.Context, EnsureRecord) (EnsureResult, error)
	Get(context.Context, string, string) (Execution, error)
	GetByClientRequest(context.Context, string, string, string) (Execution, bool, error)
	ListBookTimeline(context.Context, string, string, string) ([]TimelineEvent, error)
	AttachLegacyJob(context.Context, string, string, string) error
	UpdateBookStatus(context.Context, string, string, BookStatus, Failure) error
	AppendOutbox(context.Context, OutboxEvent) error
}
```

- [ ] Step 1: Write failing SQL transaction tests

Use the repository's existing SQL test dependency pattern. Cover:

- first `Ensure` inserts execution, one snapshot, selected books, and one outbox event in one transaction;
- same owner/batch/client request returns existing execution;
- a second request with a new request ID reuses active books instead of duplicating them;
- one invalid book rolls back the entire ensure transaction;
- queries always scope by owner.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/batchfactoryexecution -run 'TestMySQLRepository' -count=1`

Expected: FAIL.

- [ ] Step 3: Implement `MySQLRepository`

Use `SELECT ... FOR UPDATE` inside `Ensure`. Translate MySQL duplicate-key races into a re-read, not an HTTP conflict. Never load all execution rows into memory.

- [ ] Step 4: Run package and race tests

Run: `go test -race ./internal/batchfactoryexecution -run 'TestMySQLRepository' -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/repository.go backend/internal/batchfactoryexecution/mysql_repository.go backend/internal/batchfactoryexecution/mysql_repository_test.go
git commit -m "feat(batch): persist unified executions"
```

---

## Task 5: Add old-task compatibility without dual writes

**Files:**

- Create: `backend/internal/batchfactoryexecution/legacy.go`
- Create: `backend/internal/batchfactoryexecution/legacy_test.go`
- Create: `backend/internal/httpapi/batch_factory_execution_adapter.go`
- Modify: `backend/internal/httpapi/automation_mysql_store.go`

**Interfaces:**

```go
type LegacyAutomationGateway interface {
	PeekActive(context.Context, string, string) (LegacyTask, bool, error)
	Start(context.Context, string, string, LegacyStartInput) (LegacyTask, error)
	Read(context.Context, string, string) (LegacyTask, bool, error)
}

type LegacyTask struct {
	JobID  string
	State  string
	Books  map[string]LegacyBook
	Since  time.Time
}
```

- [ ] Step 1: Write failing adapter tests

Cover active `running`, `paused`, and retryable jobs; completed legacy books must project to completed, active books to running, and unknown legacy states to `needs_attention` with a visible reason.

- [ ] Step 2: Add a read-only single-job query to the old store

Add a narrow optional interface instead of calling `Load()` for every request:

```go
type AutomationJobReader interface {
	FindActiveJob(context.Context, string, string) (map[string]any, bool, error)
	FindJob(context.Context, string) (map[string]any, bool, error)
}
```

The MySQL implementation must query `batch_automation_jobs` by indexed owner/batch/state columns and decode only one `job_json`.

- [ ] Step 3: Implement the coordinator adapter

The adapter may call `BatchAutomationCoordinator.Start` only with `queueIfActive=true`. It must never update new execution tables from the old coordinator; the new service owns new tables.

- [ ] Step 4: Run focused tests

Run: `go test ./internal/httpapi ./internal/batchfactoryexecution -run 'Test.*Legacy|TestMySQLAutomationStateStoreFind' -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/legacy.go backend/internal/batchfactoryexecution/legacy_test.go backend/internal/httpapi/batch_factory_execution_adapter.go backend/internal/httpapi/automation_mysql_store.go backend/internal/httpapi/automation_state_store_test.go
git commit -m "feat(batch): bridge legacy automation tasks"
```

---

## Task 6: Implement idempotent `EnsureExecution`

**Files:**

- Create: `backend/internal/batchfactoryexecution/service.go`
- Create: `backend/internal/batchfactoryexecution/service_test.go`

**Interfaces:**

```go
type EnsureCommand struct {
	Owner           string
	BatchID         string
	BookIDs         []string
	Mode            string
	Target          string
	ClientRequestID string
	Config          map[string]any
}

type EnsureResponse struct {
	ExecutionID      string   `json:"executionId"`
	Created          bool     `json:"created"`
	Attached         bool     `json:"attached"`
	ReusedBooks      []string `json:"reusedBooks"`
	QueuedBooks      []string `json:"queuedBooks"`
	ReconcilingBooks []string `json:"reconcilingBooks"`
	BlockedBooks     []string `json:"blockedBooks"`
}

func (s *Service) EnsureExecution(context.Context, EnsureCommand) (EnsureResponse, error)
```

- [ ] Step 1: Write failing service tests

Required cases:

1. First click creates one execution and queues selected books.
2. Same `clientRequestId` returns the same execution ID.
3. A different click while the same books are active reuses them and only queues missing books.
4. Existing legacy active job is attached and no second legacy start occurs.
5. One blocked book appears in `blockedBooks`; runnable books still queue.
6. Existing media is represented as reusable input; force regeneration requires an explicit command flag added in a later phase.
7. If legacy start returns an ambiguous error, books enter `reconciling`, not a second immediate submission.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/batchfactoryexecution -run 'TestServiceEnsureExecution' -count=1`

Expected: FAIL.

- [ ] Step 3: Implement the service in this order

1. validate owner, batch, mode, target, request ID, and non-empty book IDs;
2. canonicalize and freeze config;
3. read old active task once;
4. transactionally ensure execution/books/outbox;
5. attach old job or request legacy start;
6. persist only the legacy job reference and projected status;
7. return the persisted readback.

Do not hold a database transaction open while calling the legacy coordinator.

- [ ] Step 4: Run all module tests with race detector

Run: `go test -race ./internal/batchfactoryexecution -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/service.go backend/internal/batchfactoryexecution/service_test.go
git commit -m "feat(batch): ensure idempotent executions"
```

---

## Task 7: Add recoverable outbox dispatch and reconciliation

**Files:**

- Create: `backend/internal/batchfactoryexecution/outbox.go`
- Create: `backend/internal/batchfactoryexecution/outbox_test.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**

```go
type OutboxRepository interface {
	Claim(context.Context, int, time.Time) ([]OutboxEvent, error)
	MarkDelivered(context.Context, string, time.Time) error
	MarkFailed(context.Context, string, string, time.Time) error
}

type Dispatcher struct {
	Repository OutboxRepository
	Handler    func(context.Context, OutboxEvent) error
	Now        func() time.Time
}

func (d *Dispatcher) DispatchOnce(context.Context, int) error
```

- [ ] Step 1: Write failing recovery tests

Cover crash after claim, retry delay, max attempt visibility, and duplicate delivery. Delivering the same event twice must not create a second execution or legacy job.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/batchfactoryexecution -run 'TestDispatcher' -count=1`

Expected: FAIL.

- [ ] Step 3: Implement database-first outbox dispatch

Use MySQL claim timestamps and bounded batches. Redis may wake the dispatcher, but a lost Redis message must still be recoverable by polling `available_at`.

- [ ] Step 4: Wire dispatcher lifecycle in `app.go`

Start it with the server context and stop it on shutdown. Do not start a second dispatcher in tests unless explicitly constructed.

- [ ] Step 5: Run tests

Run: `go test -race ./internal/batchfactoryexecution ./internal/app -count=1`

Expected: PASS.

- [ ] Step 6: Commit

```powershell
git add backend/internal/batchfactoryexecution/outbox.go backend/internal/batchfactoryexecution/outbox_test.go backend/internal/app/app.go
git commit -m "feat(batch): recover execution dispatch"
```

---

## Task 8: Expose the new HTTP contract

**Files:**

- Create: `backend/internal/httpapi/batch_factory_execution_routes.go`
- Create: `backend/internal/httpapi/batch_factory_execution_routes_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**

```text
POST /api/batch-factory/v12/batches/{batchId}/executions/ensure
GET  /api/batch-factory/v12/executions/{executionId}
GET  /api/batch-factory/v12/executions/{executionId}/books/{bookId}/timeline
POST /api/batch-factory/v12/executions/{executionId}/reconcile
POST /api/batch-factory/v12/executions/{executionId}/cancel
```

Ensure request:

```json
{
  "bookIds": ["book-1"],
  "mode": "video",
  "target": "published_confirmed",
  "clientRequestId": "browser-generated-uuid",
  "config": {}
}
```

- [ ] Step 1: Write failing signed route tests

Assert 401 without auth, 400 for invalid mode/request ID, 201 on creation, 200 on reuse, owner isolation, and stable response fields.

- [ ] Step 2: Run and prove failure

Run: `go test ./internal/httpapi -run 'TestBatchExecutionRoutes' -count=1`

Expected: FAIL.

- [ ] Step 3: Implement thin handlers

Handlers may only decode, validate basic JSON shape, call service, and map known domain errors. No SQL, config merge, prompt content, or old task map parsing belongs in the route file.

- [ ] Step 4: Wire `RouterOptions`

Add one optional field:

```go
BatchExecutionService *batchfactoryexecution.Service
```

Register routes only when the field is non-nil, keeping existing tests and minimal routers working.

- [ ] Step 5: Run backend tests

Run: `go test ./internal/httpapi ./internal/app ./internal/batchfactoryexecution -count=1`

Expected: PASS.

- [ ] Step 6: Commit

```powershell
git add backend/internal/httpapi/batch_factory_execution_routes.go backend/internal/httpapi/batch_factory_execution_routes_test.go backend/internal/httpapi/router.go backend/internal/app/app.go
git commit -m "feat(batch): expose execution kernel api"
```

---

## Task 9: Add the frontend module API and status projection

**Files:**

- Create: `frontend/src/modules/batch-factory/api/executions.js`
- Create: `frontend/src/modules/batch-factory/api/executions.test.js`
- Create: `frontend/src/modules/batch-factory/model/executionStatus.js`
- Create: `frontend/src/modules/batch-factory/model/executionStatus.test.js`
- Create: `frontend/src/modules/batch-factory/index.js`

**Interfaces:**

```js
export function ensureBatchExecution(batchId, payload, request = apiRequest)
export function getBatchExecution(executionId, request = apiRequest)
export function getBatchExecutionTimeline(executionId, bookId, request = apiRequest)
export function reconcileBatchExecution(executionId, request = apiRequest)
export function executionViewModel(execution)
```

- [ ] Step 1: Write failing URL and status tests

Cover URL encoding, HTTP methods, request body, and Chinese labels for `running`, `reconciling`, `retry_wait`, `needs_attention`, `completed`, and `cancelled`.

- [ ] Step 2: Run and prove failure

Run: `node --test src/modules/batch-factory/api/executions.test.js src/modules/batch-factory/model/executionStatus.test.js`

Working directory: `frontend`

Expected: FAIL.

- [ ] Step 3: Implement API and pure view-model functions

Do not place system prompts, provider rules, retry rules, or state mutation logic in the browser.

- [ ] Step 4: Run tests

Run: `node --test src/modules/batch-factory/api/executions.test.js src/modules/batch-factory/model/executionStatus.test.js`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add frontend/src/modules/batch-factory
git commit -m "feat(batch): add execution frontend module"
```

---

## Task 10: Switch “立即执行已选” to the idempotent entry

**Files:**

- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactoryExecutionEntry.test.js`

**Interfaces:**

The click path must call:

```js
await ensureBatchExecution(batch.id, {
  bookIds: selectedBookIds,
  mode: batchProductionMode,
  target: 'published_confirmed',
  clientRequestId,
  config: effectiveExecutionConfig
});
```

- [ ] Step 1: Change the source test first

The test must require `ensureBatchExecution` inside `runSelectedProduction()` and reject direct calls to `runAutomationAction('start', ...)` in that function.

- [ ] Step 2: Run and prove failure

Run: `node --test src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js src/user/pages/shuihuo/batchFactoryExecutionEntry.test.js`

Working directory: `frontend`

Expected: FAIL.

- [ ] Step 3: Implement stable click idempotency

Generate one `clientRequestId` per user click and keep it until that request finishes. A network retry must reuse the same ID. A later deliberate click may use a new ID; the backend still reuses active books.

Keep existing preflight. Use its `runnableBookIds` as the submitted set, and show `blockedBooks` separately instead of aborting runnable books.

- [ ] Step 4: Poll canonical execution state

After ensure succeeds, poll `GET /executions/{id}`. Stop polling only at `completed`, `needs_attention`, or `cancelled`. `reconciling` and `retry_wait` are visible non-terminal states.

- [ ] Step 5: Run frontend tests and build

Run: `node --test src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js src/user/pages/shuihuo/batchFactoryExecutionEntry.test.js`

Run: `npm run build`

Expected: PASS.

- [ ] Step 6: Commit

```powershell
git add frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js frontend/src/user/pages/shuihuo/batchFactoryExecutionEntry.test.js
git commit -m "feat(batch): use idempotent selected execution"
```

---

## Task 11: Add observability and acceptance invariants

**Files:**

- Create: `backend/internal/batchfactoryexecution/observability.go`
- Create: `backend/internal/batchfactoryexecution/observability_test.go`
- Modify: `backend/internal/httpapi/batch_factory_execution_routes.go`

**Required fields:**

```text
execution_id, owner_hash, batch_id, book_id, stage_key,
attempt_no, input_fingerprint, legacy_job_id, status, duration_ms
```

- [ ] Step 1: Write failing redaction tests

Logs must not contain API keys, cookies, passwords, full source text, prompt bodies, or raw provider payloads.

- [ ] Step 2: Implement structured event helpers

Use hashes or IDs for correlation. Record transition rejection and reconciliation outcomes as explicit events.

- [ ] Step 3: Add invariants to service tests

Assert:

- no two active stage runs share one `(book, stage, fingerprint)`;
- completed books have a confirmed publication receipt when target is `published_confirmed`;
- ambiguous provider outcomes never become succeeded without readback;
- one failed book does not stop another book from progressing.

- [ ] Step 4: Run backend suite

Run: `go test -race ./internal/batchfactoryexecution ./internal/httpapi ./internal/storage -count=1`

Expected: PASS.

- [ ] Step 5: Commit

```powershell
git add backend/internal/batchfactoryexecution/observability.go backend/internal/batchfactoryexecution/observability_test.go backend/internal/httpapi/batch_factory_execution_routes.go
git commit -m "test(batch): enforce execution invariants"
```

---

## Task 12: Self-review against the approved design

**Files:**

- Review: `docs/superpowers/specs/2026-11-22-batch-factory-unified-execution-design.md`
- Review: every file changed by Tasks 1–11

- [ ] Step 1: Compare implementation against every first-phase design promise

Check immutable config, per-book isolation, repeated-click reuse, unknown-result reconciliation, legacy compatibility, MySQL truth, Redis boundary, and no dual-write.

- [ ] Step 2: Scan for placeholders and accidental browser rules

Run:

```powershell
rg -n "TODO|FIXME|not implemented|临时绕过|稍后实现" backend/internal/batchfactoryexecution backend/internal/httpapi/batch_factory_execution* frontend/src/modules/batch-factory
rg -n "systemPrompt|promptBody|apiKey|password" frontend/src/modules/batch-factory
```

Expected: no unexplained placeholders and no backend-only secrets/prompts in frontend.

- [ ] Step 3: Check type and contract consistency

Compare Go JSON tags, frontend field names, route tests, and status labels. Verify `published_confirmed` cannot be satisfied by a local merge artifact alone.

- [ ] Step 4: Run complete focused verification

```powershell
Set-Location backend
go test -race ./internal/batchfactoryexecution ./internal/httpapi ./internal/storage ./internal/app -count=1
Set-Location ..\frontend
node --test src/modules/batch-factory/api/executions.test.js src/modules/batch-factory/model/executionStatus.test.js src/user/pages/shuihuo/batchFactorySelectedProduction.source.test.js src/user/pages/shuihuo/batchFactoryExecutionEntry.test.js
npm run build
```

- [ ] Step 5: Inspect the exact diff

Run: `git diff --check` and `git status --short`.

Verify unrelated dirty files are neither staged nor changed by formatting.

---

## Task 13: Local 4001 acceptance

**Files:**

- Build output only; do not commit generated runtime artifacts unless the repository already tracks the required embedded files.

- [ ] Step 1: Build frontend and embed it into the Go binary using the repository's existing release script

Do not invent a second serving path. Confirm the embedded asset hash changes in `/api/runtime-build-info`.

- [ ] Step 2: Stop only the confirmed local 4001 `qiantie-local.exe`

Resolve the listener PID and executable path first. Do not stop unrelated Go or Node processes.

- [ ] Step 3: Start the new local binary and verify identity

Read back:

```text
GET http://127.0.0.1:4001/health
GET http://127.0.0.1:4001/api/runtime-build-info
```

The SHA/build identity must match the just-built source, not the previous `794f15e8` binary.

- [ ] Step 4: Perform authenticated browser acceptance

Use a disposable test batch or selected test books. Prove:

1. first click returns one execution ID;
2. browser retry with same request ID returns the same ID;
3. another click while active does not create a second legacy task;
4. runnable books continue when one book is blocked;
5. status survives browser refresh and local process restart;
6. old active task is shown through the new execution view;
7. no paid generation is triggered beyond the explicitly selected disposable scope.

- [ ] Step 5: Save acceptance evidence

Record local build SHA, execution ID, legacy job ID, relevant HTTP responses, and screenshots. A green `/health` alone is not acceptance.

---

## Task 14: Push remote only after local acceptance

- [ ] Step 1: Verify commit range and remote

Run:

```powershell
git status --short
git log --oneline --decorate -15
git remote -v
```

Remote must be `https://github.com/cui1112233/yzcm`; branch must be `main`/current `v88` main development line as configured by the repository. Do not create a branch.

- [ ] Step 2: Push the accepted commits

Run the repository's normal non-Action push command to `main`.

- [ ] Step 3: Read back remote SHA

Run `git ls-remote origin refs/heads/main` and record the exact SHA. Do not say “已推送” until local accepted commit is visible remotely.

---

## Task 15: SSH deploy and online readback

- [ ] Step 1: Inspect online disk, service, and current release before upload

Use SSH to identify the current binary, service unit/process, free disk, and rollback copy. Never delete broad directories or logs as part of this deployment.

- [ ] Step 2: Upload one Go binary using the existing release layout

Create a timestamped rollback copy, upload to a temporary filename, verify checksum, then atomically replace the service binary.

- [ ] Step 3: Restart only the confirmed production service

Read service logs immediately. If migration `8800035` fails, restore the previous binary; the append-only schema may remain because old binaries ignore the new tables.

- [ ] Step 4: Verify public identity and behavior

Read back public `/health` and `/api/runtime-build-info`, then repeat the safe authenticated acceptance:

- one click → one execution;
- repeat click → same/attached task, no 409;
- refresh → state still visible;
- legacy task projection works;
- no execution is `completed` without the required external confirmation.

- [ ] Step 5: Final evidence report

Report local SHA, remote SHA, online SHA, migration result, local execution ID, online execution ID, whether a legacy task was attached, tests run, and any books in `needs_attention`. Do not collapse build success, local success, remote push, and online acceptance into one statement.

---

## Completion Gate

This foundation phase is complete only when all statements are true:

- [ ] New execution tables exist through append-only migration `8800035`.
- [ ] Repeated clicks cannot create duplicate active execution books or duplicate legacy jobs.
- [ ] Configuration and stage inputs have persisted digests.
- [ ] Old active jobs appear in the new read model without dual-write.
- [ ] One blocked/failed book does not stop the rest.
- [ ] Ambiguous external results become `reconciling`, never false success.
- [ ] The selected-execution browser entry uses `/executions/ensure`.
- [ ] Focused backend race tests and frontend build pass.
- [ ] Local 4001 acceptance is recorded before remote push.
- [ ] Remote SHA is read back before SSH deployment.
- [ ] Online runtime SHA and authenticated behavior are read back after deployment.

The next implementation plan, after this gate passes, migrates the video stages into native handlers one stage at a time; comic stages follow only after video production and publication readback are stable.

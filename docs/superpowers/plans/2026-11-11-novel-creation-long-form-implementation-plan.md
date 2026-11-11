# 小说创作长篇原创闭环 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在一战晟铭 V88 中交付可真实使用的“小说创作 → 长篇原创”闭环，包括作品、会话、消息、规划、写章、审核、修订、结算、记忆、取消、重试、刷新恢复和 Go embed 浏览器验收。

**Architecture:** React 页面只负责交互和展示；Go `novelcreation` 领域服务拥有状态机和数据规则；MySQL 保存正式状态，Redis 保存队列、锁、取消信号和短期事件。模型调用复用现有成员模型解析器，系统提示词只从 Go `prompt/novelcreation` 读取。第一阶段建立稳定合同，后续短篇、仿写、番外、续写和定时写作在同一合同上扩展。

**Tech Stack:** Go 1.x、MySQL 8、Redis、React、Ant Design、Lucide、现有 `apiRequest`、现有 `batchfactoryv11.DirectorProvider`、Go embed。

## Global Constraints

- 只在 `v88` 维护源码；不新建分支，不覆盖或清理现有脏改动。
- 本计划不提交、不推送、不发布公网；每个任务以测试和 diff 检查点结束，提交需用户另行授权。
- 不复制参考应用源码、编译包、提示词、密钥、私有接口和专有资源。
- 系统提示词只能位于 `backend/internal/prompt/novelcreation/`；前端不得硬编码系统提示词。
- MySQL 是正式状态真相源；Redis 只保存队列、锁、取消和短期事件。
- 前端必须适配全站主题和背景，不写死白色/黑色，不另造第二套 Agent 工作区。
- 任何“完成”必须同时有领域测试、HTTP 测试、前端测试、构建、Go embed、本地运行身份和浏览器操作证据。
- 第一阶段只做长篇原创闭环；不得用空按钮提前展示未实现的短篇、仿写、番外、续写、定时写作。未实现入口显示“后续阶段”，点击不得伪造成功。

---

## 文件结构

### 新建 Go 文件

- `backend/internal/novelcreation/types.go`：领域类型、状态、请求和稳定错误码。
- `backend/internal/novelcreation/validation.go`：输入规范化、状态转换和幂等规则。
- `backend/internal/novelcreation/store.go`：存储接口。
- `backend/internal/novelcreation/memory_store.go`：领域/HTTP 测试用内存实现。
- `backend/internal/novelcreation/mysql_store.go`：MySQL 正式实现。
- `backend/internal/novelcreation/runtime.go`：队列、锁、事件、取消接口和 Redis 实现。
- `backend/internal/novelcreation/pipeline.go`：规划、写作、审核、修订、结算、记忆流水线。
- `backend/internal/novelcreation/service.go`：用例编排和状态事务。
- `backend/internal/novelcreation/*_test.go`：领域、存储、状态和流水线测试。
- `backend/internal/prompt/novelcreation/definitions.go`：长篇原创各步骤的 PromptDefinition。
- `backend/internal/prompt/novelcreation/definitions_test.go`：提示词合同测试。
- `backend/internal/storage/novel_creation_schema.go`：追加式迁移 `8800300`。
- `backend/internal/storage/novel_creation_schema_test.go`：表、索引、迁移幂等测试。
- `backend/internal/httpapi/novel_creation.go`：HTTP 和 SSE 事件路由。
- `backend/internal/httpapi/novel_creation_test.go`：鉴权、错误码、幂等、取消、事件和恢复测试。

### 新建前端文件

- `frontend/src/shared/api/novelCreation.js`：小说创作 API 与 SSE/poll 恢复入口。
- `frontend/src/shared/api/novelCreation.test.js`：请求和错误映射测试。
- `frontend/src/user/pages/NovelCreationPage.jsx`：页面装配和加载状态。
- `frontend/src/user/pages/novel-creation/NovelCreationSidebar.jsx`：六模式、作品、会话和工具入口。
- `frontend/src/user/pages/novel-creation/NovelComposer.jsx`：输入、模型、思考强度、发送/停止。
- `frontend/src/user/pages/novel-creation/NovelRunTimeline.jsx`：步骤、正文、审核、警告、失败、继续。
- `frontend/src/user/pages/novel-creation/novelCreationState.js`：纯状态归并函数。
- `frontend/src/user/pages/novel-creation/novel-creation.css`：主题适配布局。
- 对应 `*.test.js` 和 `NovelCreationPage.source.test.js`。

### 修改文件

- `backend/internal/storage/migrations.go`：装入 `NovelCreationMigrations()`。
- `backend/internal/app/app.go`：创建 store/runtime/service，传给 router。
- `backend/internal/httpapi/router.go`：增加 RouterOptions 和注册路由。
- `frontend/src/user/App.jsx`：懒加载 `/novel-creation`。
- `frontend/src/shared/layouts/UserLayout.jsx`：加入“小说创作”导航。
- `frontend/src/user/workbenchPreload.js`：预加载小说创作 chunk。

---

### Task 1: 领域类型、输入校验和状态转换

**Files:**
- Create: `backend/internal/novelcreation/types.go`
- Create: `backend/internal/novelcreation/validation.go`
- Test: `backend/internal/novelcreation/validation_test.go`

**Interfaces:**
- Produces: `NormalizeCreateProjectInput(CreateProjectInput)`, `NormalizeSendMessageInput(SendMessageInput)`, `CanTransition(RunStatus, RunStatus) bool`, `DomainError`。
- Consumes: 无。

- [ ] **Step 1: 写失败测试**

```go
func TestNormalizeSendMessageInputRejectsBlank(t *testing.T) {
	_, err := NormalizeSendMessageInput(SendMessageInput{SessionID: "session-1", Content: " \n "})
	var domainErr *DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != ErrInputRequired {
		t.Fatalf("err=%v, want %s", err, ErrInputRequired)
	}
}

func TestCanTransitionRejectsSucceededToWriting(t *testing.T) {
	if CanTransition(RunSucceeded, RunWriting) {
		t.Fatal("terminal run must not return to writing")
	}
}
```

- [ ] **Step 2: 运行失败测试**

Run: `cd backend; go test ./internal/novelcreation -run 'TestNormalize|TestCanTransition'`

Expected: FAIL，包或函数尚不存在。

- [ ] **Step 3: 实现稳定类型**

`types.go` 必须定义：

```go
type RunStatus string

const (
	RunDraft RunStatus = "draft"
	RunClarifying RunStatus = "clarifying"
	RunAwaitingConfirmation RunStatus = "awaiting_confirmation"
	RunQueued RunStatus = "queued"
	RunPlanning RunStatus = "planning"
	RunWriting RunStatus = "writing"
	RunNormalizing RunStatus = "normalizing"
	RunAuditing RunStatus = "auditing"
	RunRevising RunStatus = "revising"
	RunSettling RunStatus = "settling"
	RunIndexing RunStatus = "indexing"
	RunSucceeded RunStatus = "succeeded"
	RunSucceededWithWarnings RunStatus = "succeeded_with_warnings"
	RunFailed RunStatus = "failed"
	RunCancelled RunStatus = "cancelled"
)

const (
	ErrInputRequired = "novel_input_required"
	ErrProjectNotFound = "novel_project_not_found"
	ErrSessionNotFound = "novel_session_not_found"
	ErrRunInProgress = "novel_run_in_progress"
	ErrProjectLocked = "novel_project_locked"
	ErrInvalidTransition = "novel_invalid_transition"
	ErrModelUnavailable = "novel_model_unavailable"
	ErrGenerationTimeout = "novel_generation_timeout"
	ErrPersistenceFailed = "novel_persistence_failed"
)

type DomainError struct { Code string; Message string; Cause error }
type CreateProjectInput struct { Title, Genre, World, Protagonist, CoreConflict, ModelID string; TargetChapterWords int }
type SendMessageInput struct { SessionID, ProjectID, Content, ModelID, ThinkingLevel, IdempotencyKey string }
```

`NormalizeSendMessageInput` trim 所有字符串；空 Content、SessionID、ModelID 或 IdempotencyKey 返回对应 `DomainError`；`TargetChapterWords` 允许 800—8000，默认 2000。

- [ ] **Step 4: 运行并补齐表驱动边界测试**

覆盖空格、800、8000、799、8001、重复幂等键需要的规范化字段。

Run: `cd backend; go test ./internal/novelcreation`

Expected: PASS。

- [ ] **Step 5: 检查点**

Run: `git diff -- backend/internal/novelcreation`

确认没有参考品牌名、密钥、URL 或前端提示词。

---

### Task 2: MySQL 迁移、Store 接口和内存实现

**Files:**
- Create: `backend/internal/novelcreation/store.go`
- Create: `backend/internal/novelcreation/memory_store.go`
- Create: `backend/internal/storage/novel_creation_schema.go`
- Create: `backend/internal/storage/novel_creation_schema_test.go`
- Modify: `backend/internal/storage/migrations.go`
- Test: `backend/internal/novelcreation/memory_store_test.go`

**Interfaces:**
- Consumes: Task 1 类型。
- Produces: `Store`、`TxStore`、`NewMemoryStore()`、`storage.NovelCreationMigrations()`。

- [ ] **Step 1: 写 Store 合同测试**

```go
func TestMemoryStoreCreateProjectAndSessionAreOwnerScoped(t *testing.T) {
	store := NewMemoryStore()
	project, err := store.CreateProject(context.Background(), "alice", CreateProjectInput{Title: "迷雾记者", Genre: "都市悬疑", ModelID: "text-1"})
	if err != nil { t.Fatal(err) }
	if _, err := store.GetProject(context.Background(), "bob", project.ID); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("cross-owner read err=%v", err)
	}
}
```

- [ ] **Step 2: 定义 Store**

```go
type Store interface {
	CreateProject(context.Context, string, CreateProjectInput) (Project, error)
	GetProject(context.Context, string, string) (Project, error)
	ListProjects(context.Context, string) ([]Project, error)
	CreateSession(context.Context, string, CreateSessionInput) (Session, error)
	GetSession(context.Context, string, string) (Session, error)
	ListSessions(context.Context, string, string) ([]Session, error)
	AppendMessage(context.Context, string, Message) (Message, error)
	ListMessages(context.Context, string, string, int64) ([]Message, error)
	FindRunByIdempotencyKey(context.Context, string, string) (Run, bool, error)
	CreateRun(context.Context, string, Run) (Run, error)
	UpdateRun(context.Context, string, RunUpdate) (Run, error)
	AppendRunEvent(context.Context, string, RunEvent) (RunEvent, error)
	ListRunEvents(context.Context, string, string, int64) ([]RunEvent, error)
	SaveChapterBundle(context.Context, string, ChapterBundle) error
}
```

- [ ] **Step 3: 写迁移失败测试**

断言迁移 `8800300` 含 `novel_projects`、`novel_sessions`、`novel_messages`、`novel_volumes`、`novel_chapters`、`novel_chapter_versions`、`novel_runs`、`novel_run_steps`、`novel_run_events`、`novel_foundations`、`novel_memories`、`novel_audits`、`novel_revisions`，且所有成员数据表都有 `owner_username` 索引。

- [ ] **Step 4: 实现追加式迁移并装入总迁移**

迁移版本固定 `8800300`，checksum 固定 `v88-novel-creation-long-form-v1`。正文使用 `LONGTEXT`，模型/提示词/运行快照使用 JSON，时间统一 `DATETIME(6)`，外部 ID 使用 `VARCHAR(96)`。

- [ ] **Step 5: 运行测试**

Run: `cd backend; go test ./internal/storage ./internal/novelcreation`

Expected: PASS，重复取 `NovelCreationMigrations()` 结果稳定。

- [ ] **Step 6: 检查点**

Run: `git diff -- backend/internal/storage backend/internal/novelcreation`

确认迁移只追加，不改旧版本。

---

### Task 3: MySQL Store 实现和重启恢复

**Files:**
- Create: `backend/internal/novelcreation/mysql_store.go`
- Test: `backend/internal/novelcreation/mysql_store_test.go`

**Interfaces:**
- Consumes: Task 2 `Store`。
- Produces: `NewMySQLStore(*sql.DB) Store`。

- [ ] **Step 1: 写 SQL 行为测试**

使用项目现有 MySQL 测试辅助，覆盖：owner 隔离、幂等 run 唯一索引、事件游标、章节 bundle 事务回滚、软删除后不可读。

```go
func TestMySQLStoreIdempotentRunSurvivesReopen(t *testing.T) {
	db := openNovelCreationTestDB(t)
	store := NewMySQLStore(db)
	first, err := store.CreateRun(context.Background(), "alice", Run{ID: "run-1", IdempotencyKey: "send-1", Status: RunQueued})
	if err != nil { t.Fatal(err) }
	second, found, err := NewMySQLStore(db).FindRunByIdempotencyKey(context.Background(), "alice", "send-1")
	if err != nil || !found || second.ID != first.ID { t.Fatalf("recovered=%+v found=%v err=%v", second, found, err) }
}
```

- [ ] **Step 2: 实现查询和事务**

所有查询必须同时带 `owner_username` 和 `deleted_at IS NULL`。`SaveChapterBundle` 在一个事务保存章节版本、审核、修订、foundation、memory 和 run 终态；任一失败全回滚。

- [ ] **Step 3: 运行存储测试**

Run: `cd backend; go test ./internal/novelcreation -run MySQL`

Expected: PASS；无可用 MySQL 时测试按项目约定明确 skip，不得伪装成功。

- [ ] **Step 4: 检查点**

Run: `git diff --check`

---

### Task 4: Redis 队列、作品锁、取消和事件

**Files:**
- Create: `backend/internal/novelcreation/runtime.go`
- Test: `backend/internal/novelcreation/runtime_test.go`

**Interfaces:**
- Produces: `Runtime`、`NewRedisRuntime(prefix string, do RedisDoFunc)`、`NewMemoryRuntime()`。
- Consumes: Task 1 `RunEvent`。

- [ ] **Step 1: 写失败测试**

```go
func TestRuntimeLockRejectsSecondWriterAndCancellationPersists(t *testing.T) {
	rt := NewMemoryRuntime()
	release, err := rt.AcquireProject(context.Background(), "project-1", "run-1", time.Minute)
	if err != nil { t.Fatal(err) }
	defer release()
	if _, err := rt.AcquireProject(context.Background(), "project-1", "run-2", time.Minute); !IsCode(err, ErrProjectLocked) { t.Fatalf("err=%v", err) }
	if err := rt.Cancel(context.Background(), "run-1"); err != nil { t.Fatal(err) }
	if cancelled, _ := rt.IsCancelled(context.Background(), "run-1"); !cancelled { t.Fatal("cancel flag missing") }
}
```

- [ ] **Step 2: 定义 Runtime**

```go
type Runtime interface {
	Enqueue(context.Context, string) error
	Dequeue(context.Context) (string, error)
	AcquireProject(context.Context, string, string, time.Duration) (func(), error)
	Publish(context.Context, RunEvent) error
	Cancel(context.Context, string) error
	IsCancelled(context.Context, string) (bool, error)
}
```

Redis key 前缀固定 `qiantie:novel-creation`；锁释放使用 token 比对，不能直接 DEL 他人的锁；取消信号 TTL 24 小时；事件只是加速推送，MySQL 事件表负责断线恢复。

- [ ] **Step 3: 实现内存和 Redis 版本**

复用 `mergeworker.RedisDoFunc` 形状，不增加 Redis SDK。Redis 不可用时启动失败，不能悄悄退回仅内存生产模式。

- [ ] **Step 4: 运行测试**

Run: `cd backend; go test ./internal/novelcreation -run Runtime`

Expected: PASS。

---

### Task 5: 后端提示词定义和结构化输出合同

**Files:**
- Create: `backend/internal/prompt/novelcreation/definitions.go`
- Create: `backend/internal/prompt/novelcreation/definitions_test.go`

**Interfaces:**
- Produces: `FoundationPlan`、`ChapterPlan`、`ChapterDraft`、`ChapterAudit`、`ChapterRevision`、`TruthSettlement`、`MemoryIndex` 七个 `prompt.Definition`。
- Consumes: 现有 `prompt.Definition`。

- [ ] **Step 1: 写提示词合同测试**

```go
func TestDefinitionsKeepSystemPromptsInBackend(t *testing.T) {
	for _, def := range All() {
		if strings.TrimSpace(def.ID) == "" || strings.TrimSpace(def.SystemPrompt) == "" { t.Fatalf("invalid definition: %+v", def) }
		if !strings.Contains(def.OutputContract, "JSON") { t.Fatalf("%s missing JSON contract", def.ID) }
	}
}
```

- [ ] **Step 2: 定义稳定 JSON 输出**

规划输出含 `title/genre/world/characters/coreConflict/volumes`；章节计划含 `chapterNumber/title/goal/conflict/turn/hook/targetWords`；审核含 `score/issues/passed`；修订含 `strategy/content/resolvedIssueIDs`；结算含 `facts/characterChanges/timeline/foreshadowing`；记忆含 `summary/entities/openThreads`。

- [ ] **Step 3: 写独立中文提示词**

提示词不得出现参考品牌、参考私有方法名或反编译正文。要求模型只返回合同 JSON；正文生成合同把正文放在 `content` 字段，不用 Markdown 代码围栏。

- [ ] **Step 4: 运行测试和泄漏扫描**

Run: `cd backend; go test ./internal/prompt/novelcreation`

Run: `rg -n '洛水|小洛|api[_-]?key|Bearer ' backend/internal/prompt/novelcreation`

Expected: 测试 PASS；泄漏扫描无输出。

---

### Task 6: 模型适配器和长篇流水线

**Files:**
- Create: `backend/internal/novelcreation/pipeline.go`
- Test: `backend/internal/novelcreation/pipeline_test.go`

**Interfaces:**
- Consumes: `prompt/novelcreation` 定义、`batchfactoryv11.DirectorProvider`。
- Produces: `Pipeline.Run(context.Context, RunContext, EventSink) (ChapterBundle, error)`。

- [ ] **Step 1: 写固定模型测试**

建立 `queuedProvider`，依次返回基础设定、章节计划、正文、审核、修订、结算和记忆 JSON。断言步骤顺序为 planning → writing → normalizing → auditing → revising → settling → indexing，且每步事件先持久化再推送。

```go
func TestPipelineRevisesFailedAuditAndReturnsSettledBundle(t *testing.T) {
	provider := &queuedProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	pipeline := NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil })
	bundle, err := pipeline.Run(context.Background(), fixtureRunContext(), collectEvents(t))
	if err != nil { t.Fatal(err) }
	if bundle.Audit.Score < 85 || bundle.Chapter.Content != revisedContent { t.Fatalf("bundle=%+v", bundle) }
}
```

- [ ] **Step 2: 实现模型适配器**

每步调用 `DirectorProvider.Complete`，明确设置 `SystemPrompt`、`UserPrompt`、`MaxTokens`、`Temperature`。适配器把一次完整回复作为一个真实 `assistant.delta` 事件发出；不使用定时器伪造 token。后续支持流式的模型可发送多个 delta，但事件合同不变。

- [ ] **Step 3: 实现审核和修订边界**

默认阈值 85，最多两轮修订；分数未提升时保留最佳版本并返回 `succeeded_with_warnings`；JSON 解析失败最多进行一次“按合同修复”模型调用，仍失败则 `failed`。

- [ ] **Step 4: 实现取消检查**

每个模型调用前后检查 `Runtime.IsCancelled`。取消发生在 `SaveChapterBundle` 前返回 `ErrRunCancelled`；不得写正式章节。

- [ ] **Step 5: 运行测试**

Run: `cd backend; go test ./internal/novelcreation -run Pipeline`

Expected: 正常、审核修订、JSON错误、模型超时和取消全部 PASS。

---

### Task 7: Service 编排、幂等和失败恢复

**Files:**
- Create: `backend/internal/novelcreation/service.go`
- Test: `backend/internal/novelcreation/service_test.go`

**Interfaces:**
- Consumes: `Store`、`Runtime`、`Pipeline`。
- Produces: `Service.CreateProject`、`CreateSession`、`SendMessage`、`Run`、`CancelRun`、`RetryRun`、`GetSnapshot`。

- [ ] **Step 1: 写幂等和取消测试**

```go
func TestSendMessageReturnsExistingRunForSameIdempotencyKey(t *testing.T) {
	service := fixtureService(t)
	first, err := service.SendMessage(context.Background(), "alice", validSendInput("send-1"))
	if err != nil { t.Fatal(err) }
	second, err := service.SendMessage(context.Background(), "alice", validSendInput("send-1"))
	if err != nil || second.ID != first.ID { t.Fatalf("first=%s second=%s err=%v", first.ID, second.ID, err) }
}
```

- [ ] **Step 2: 实现用例编排**

`SendMessage` 在一个逻辑事务中保存用户消息、run 和 queued 事件，再入队；入队失败把 run 标为 failed 并保留可重试信息。`Run` 先取作品锁，再执行 pipeline，最后用 `SaveChapterBundle` 原子提交。

- [ ] **Step 3: 实现重试语义**

只有 failed/cancelled/succeeded_with_warnings 可重试；新 run 记录 `retry_of_run_id`；复用原输入快照但生成新幂等键；成功章节不重复写入。

- [ ] **Step 4: 运行测试**

Run: `cd backend; go test ./internal/novelcreation -run Service`

Expected: PASS。

---

### Task 8: HTTP、SSE 和稳定错误映射

**Files:**
- Create: `backend/internal/httpapi/novel_creation.go`
- Create: `backend/internal/httpapi/novel_creation_test.go`
- Modify: `backend/internal/httpapi/router.go`

**Interfaces:**
- Consumes: `*novelcreation.Service`。
- Produces: `/api/novel-creation/projects|sessions|runs|events` 第一阶段路由。

- [ ] **Step 1: 写 HTTP 失败测试**

覆盖：未登录 401、空输入 400 + `novel_input_required`、他人资源 404、重复幂等键返回同一 run、取消 200、事件 `after` 游标、SSE 断开不取消 run。

```go
func TestNovelCreationBlankMessageReturnsStableCode(t *testing.T) {
	mux := newNovelCreationTestRouter(t)
	response := signedJSONRequest(t, mux, fixedNow, "alice", http.MethodPost, "/api/novel-creation/sessions/session-1/messages", map[string]any{"content":" ","modelId":"text-1","idempotencyKey":"send-1"})
	assertJSONErrorCode(t, response, http.StatusBadRequest, novelcreation.ErrInputRequired)
}
```

- [ ] **Step 2: 实现第一阶段路由**

路由：

```text
GET/POST /api/novel-creation/projects
GET /api/novel-creation/projects/{id}
GET/POST /api/novel-creation/sessions
GET /api/novel-creation/sessions/{id}/messages
POST /api/novel-creation/sessions/{id}/messages
GET /api/novel-creation/runs/{id}
GET /api/novel-creation/runs/{id}/events?after={eventId}
GET /api/novel-creation/runs/{id}/stream
POST /api/novel-creation/runs/{id}/cancel
POST /api/novel-creation/runs/{id}/retry
```

SSE 每条包含 `id/event/data`，15 秒发送注释心跳；重连先从 MySQL 补发 `Last-Event-ID` 之后的事件。

- [ ] **Step 3: 错误映射**

输入错误 400、认证 401、资源隔离 404、冲突/锁 409、模型/上游 502、超时 504、持久化 500。响应统一 `{ "error": "中文提示", "code": "稳定错误码" }`。

- [ ] **Step 4: 运行测试**

Run: `cd backend; go test ./internal/httpapi -run NovelCreation`

Expected: PASS。

---

### Task 9: App、Router、MySQL 和 Redis 装配

**Files:**
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/app/app_test.go`
- Test: `backend/internal/httpapi/router_test.go`

**Interfaces:**
- Consumes: Task 3/4/7。
- Produces: 本地 Go runtime 中可用的小说创作服务。

- [ ] **Step 1: 写装配测试**

断言 RouterOptions 有 `NovelCreation *novelcreation.Service`，非 nil 时注册路由，nil 时不注册；Redis 配置错误必须让 app 启动失败并显示明确原因。

- [ ] **Step 2: 装配正式依赖**

在 `app.go` 创建 `novelcreation.NewMySQLStore(db)`；用现有 `mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:novel-creation")` 创建 runtime；模型 resolver 适配现有 `textProviderForModel`；启动一个受 app context 控制的 worker，不创建无归属 goroutine。

- [ ] **Step 3: 运行装配测试**

Run: `cd backend; go test ./internal/app ./internal/httpapi -run 'NovelCreation|Router'`

Expected: PASS。

---

### Task 10: 前端 API、事件归并和恢复

**Files:**
- Create: `frontend/src/shared/api/novelCreation.js`
- Create: `frontend/src/shared/api/novelCreation.test.js`
- Create: `frontend/src/user/pages/novel-creation/novelCreationState.js`
- Create: `frontend/src/user/pages/novel-creation/novelCreationState.test.js`

**Interfaces:**
- Produces: `listNovelProjects`、`createNovelProject`、`listNovelSessions`、`createNovelSession`、`sendNovelMessage`、`getNovelRun`、`listNovelRunEvents`、`cancelNovelRun`、`retryNovelRun`、`mergeNovelEvents`。
- Consumes: `apiRequest`。

- [ ] **Step 1: 写 API 请求测试**

```js
it('sends a stable idempotency key with a trimmed message', async () => {
  const request = vi.fn().mockResolvedValue({ run: { id: 'run-1' } });
  await sendNovelMessage('session-1', { content: '  写第一章  ', modelId: 'text-1', thinkingLevel: 'medium', idempotencyKey: 'send-1' }, request);
  expect(request).toHaveBeenCalledWith('/api/novel-creation/sessions/session-1/messages', expect.objectContaining({ method: 'POST', body: JSON.stringify({ content: '写第一章', modelId: 'text-1', thinkingLevel: 'medium', idempotencyKey: 'send-1' }) }));
});
```

- [ ] **Step 2: 实现 API 和 abort**

普通请求用 `apiRequest`；发送消息不设置 10 分钟等待，创建 run 后立即返回；事件使用 `listNovelRunEvents` 轮询作为可靠底座，每 1 秒只在运行中请求。页面隐藏时退避到 5 秒，终态停止。

- [ ] **Step 3: 写事件归并测试**

`mergeNovelEvents` 按 eventId 去重，乱序输入按 ID 排序，终态不可被旧 running 覆盖，刷新后的快照和增量事件合并结果一致。

- [ ] **Step 4: 运行测试**

Run: `cd frontend; npm test -- --run src/shared/api/novelCreation.test.js src/user/pages/novel-creation/novelCreationState.test.js`

Expected: PASS。

---

### Task 11: 页面骨架、侧栏和 Composer

**Files:**
- Create: `frontend/src/user/pages/NovelCreationPage.jsx`
- Create: `frontend/src/user/pages/novel-creation/NovelCreationSidebar.jsx`
- Create: `frontend/src/user/pages/novel-creation/NovelComposer.jsx`
- Create: `frontend/src/user/pages/novel-creation/novel-creation.css`
- Create: `frontend/src/user/pages/NovelCreationPage.source.test.js`

**Interfaces:**
- Consumes: Task 10 API；现有 `listAvailableModels('text')`。
- Produces: 与批准设计一致的工作台入口、作品/会话选择和真实发送/停止。

- [ ] **Step 1: 写源码合同测试**

断言页面包含六个稳定 mode key，但除 `book-create` 外均标记 `disabledReason: '后续阶段'`；Composer trim 空值、流式中禁用发送、Enter 发送、Shift+Enter 换行、中文输入法 composing 时不发送。

- [ ] **Step 2: 实现主题响应布局**

使用现有 CSS 变量（背景、文本、边框、主色），不写死白/黑。桌面为主侧栏 + 小说二级栏 + 工作区；窄屏二级栏为抽屉；折叠状态存 `sessionStorage`，不得影响 run。

- [ ] **Step 3: 实现真实作品/会话**

首次进入加载作品和会话；“新建会话”调用 API；选择作品后只展示其会话；新建长篇先在工作区展示题材/世界观/主角/冲突表单，提交后创建作品和绑定会话。

- [ ] **Step 4: 实现 Composer**

模型列表只取 text；没有模型时显示“尚未添加可用的文本模型”并链接 `/api-config`；思考强度为 low/medium/high；发送生成 `crypto.randomUUID()` 幂等键；运行中主按钮显示“停止生成”。

- [ ] **Step 5: 运行前端测试**

Run: `cd frontend; npm test -- --run src/user/pages/NovelCreationPage.source.test.js`

Expected: PASS。

---

### Task 12: 运行时间线、错误恢复和刷新恢复

**Files:**
- Create: `frontend/src/user/pages/novel-creation/NovelRunTimeline.jsx`
- Create: `frontend/src/user/pages/novel-creation/NovelRunTimeline.test.jsx`
- Modify: `frontend/src/user/pages/NovelCreationPage.jsx`

**Interfaces:**
- Consumes: `mergeNovelEvents` 和 run API。
- Produces: 用户可见步骤、正文、审核、警告、失败、取消、继续和重试。

- [ ] **Step 1: 写交互测试**

测试 planning/writing/auditing/revising/settling/indexing 中文状态；失败显示后端 error/code；取消保留已完成卡片；刷新根据 activeRunId 请求快照和事件；终态显示章节正文和审核摘要。

- [ ] **Step 2: 实现 Timeline**

事件卡按 eventId 排序。正文 delta 合并到当前助手消息；审核分数和问题为独立卡；`succeeded_with_warnings` 显示“正文已保存为待修订”；failed 显示“从失败处重试”；cancelled 显示“继续执行”。

- [ ] **Step 3: 实现本地草稿**

只把未发送输入存入 `sessionStorage`，key 为 `qiantie:novel-creation:draft:{sessionId}`。正式消息、运行、章节和状态不得使用 localStorage/sessionStorage 作为真相源。

- [ ] **Step 4: 运行测试**

Run: `cd frontend; npm test -- --run src/user/pages/novel-creation/NovelRunTimeline.test.jsx src/user/pages/NovelCreationPage.source.test.js`

Expected: PASS。

---

### Task 13: 路由、主导航和预加载

**Files:**
- Modify: `frontend/src/user/App.jsx`
- Modify: `frontend/src/shared/layouts/UserLayout.jsx`
- Modify: `frontend/src/user/workbenchPreload.js`
- Modify: `frontend/src/shared/layouts/UserLayout.source.test.js`
- Create: `frontend/src/user/pages/NovelCreationRoute.source.test.js`

**Interfaces:**
- Consumes: `NovelCreationPage`。
- Produces: `/novel-creation` 正式入口。

- [ ] **Step 1: 写失败测试**

断言 `App.jsx` 懒加载 `NovelCreationPage` 并注册 `/novel-creation`；主导航顺序为“小说获取 → 小说创作 → 水货生产”；页面标题返回“小说创作”；预加载映射包含该路由。

- [ ] **Step 2: 修改路由和导航**

导航使用 `BookMarked` Lucide 图标；不使用 `reload: true`，保持 React SPA；不修改现有小说获取入口。

- [ ] **Step 3: 运行聚焦测试**

Run: `cd frontend; npm test -- --run src/user/pages/NovelCreationRoute.source.test.js src/shared/layouts/UserLayout.source.test.js`

Expected: PASS。

---

### Task 14: 同输入验收、完整构建和 Go embed 本地验证

**Files:**
- Create: `backend/internal/httpapi/novel_creation_acceptance_test.go`
- Create: `frontend/src/user/pages/novel-creation/NovelCreationPage.acceptance.test.jsx`
- Modify: `docs/superpowers/specs/2026-11-11-novel-creation-clean-room-parity-spec.md`（只更新第一段状态和证据）

**Interfaces:**
- Consumes: Tasks 1—13 全部功能。
- Produces: 第一阶段完成证据。

- [ ] **Step 1: 固定同输入测试**

输入固定为：

```text
都市悬疑，女记者林夏为主角。她调查一宗旧城区失踪案，发现所有失踪者都收到同一张不存在街道的照片。目标单章约 2000 字。
```

fake provider 返回固定规划、正文、第一次 72 分审核、修订正文、第二次 91 分审核、结算和记忆。断言最终章节为 revision 2、审核 91、记忆含“照片”和“失踪案”。

- [ ] **Step 2: 边界矩阵**

分别运行空值、双击发送、模型错误、MySQL保存错误、取消、刷新恢复和同作品并发写作；每个测试断言稳定 code、run 终态和数据库行数。

- [ ] **Step 3: 运行后端完整聚焦测试**

Run: `cd backend; go test ./internal/novelcreation ./internal/httpapi ./internal/storage ./internal/app`

Expected: PASS。

- [ ] **Step 4: 运行前端聚焦和完整测试**

Run: `cd frontend; npm test -- --run src/shared/api/novelCreation.test.js src/user/pages/novel-creation src/user/pages/NovelCreationPage.source.test.js src/user/pages/NovelCreationRoute.source.test.js src/shared/layouts/UserLayout.source.test.js`

Run: `cd frontend; npm run build`

Expected: PASS；若完整测试有既存失败，单独记录并证明与本改动无关，不能把它们删除或改成 skip。

- [ ] **Step 5: 构建 Go embed 本地运行包**

使用仓库现有 Go embed 构建脚本，不把 `frontend/dist` 反向复制成源码。重启前确认当前 4001 进程路径，构建后只替换已确认的本地运行包。

- [ ] **Step 6: 运行身份验证**

Run: `Invoke-WebRequest http://127.0.0.1:4001/health`

Run: `Invoke-WebRequest http://127.0.0.1:4001/api/runtime-build-info`

Expected: health 200；build-info 指向本次本地构建和新的前端入口。

- [ ] **Step 7: 已登录浏览器逐步验收**

1. 点击主导航“小说创作”；
2. 新建作品和会话；
3. 选择文本模型和思考强度；
4. 发送固定输入；
5. 观察规划、写作、审核、修订、结算和记忆步骤；
6. 刷新后仍显示同一 run；
7. 验证章节正文、审核 91 和记忆结果；
8. 重复发送、空值、错误和取消各执行一次；
9. 截图和 API/数据库证据写入 `_local-archive-20261028/novel-creation-long-form/`。

- [ ] **Step 8: 最终检查点**

Run: `git diff --check`

Run: `git status --short`

报告：修改路径、测试命令、浏览器可见效果、未完成后续阶段；不提交、不推送、不发布。

---

## 自检结果

- 规格覆盖：第一阶段长篇原创的入口、数据、状态、提示词、模型、审核、修订、结算、记忆、取消、重试、刷新、错误和浏览器验收均有对应任务。
- 占位扫描：计划不含待补内容；短篇、仿写、番外、续写、工具入口和 Agent 深度整合被明确留给独立后续计划，不在第一阶段伪装实现。
- 类型一致性：`Store`、`Runtime`、`Pipeline`、`Service`、稳定错误码、run status、eventId 和幂等键在各任务中保持同名。
- 项目规则：没有新分支、提交、推送、Actions 或公网发布步骤；没有覆盖现有脏改动。

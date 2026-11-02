# 全项目 Go 运行时迁移 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 ECS、公网与 Windows 本地执行器的应用运行时迁移为 Go，最终停止 Node、npm、Node API bridge、Node 浏览器 Worker 和 Node 本地执行器。

**Architecture:** Go 主服务提供嵌入式 React 静态资源和全部 API；视频供应商、合成任务、视频管理系统连接器与 Windows 本地执行器均采用独立 Go 进程。迁移期间，Node 源码只读保留作行为对照和回滚证据；每一项先以 Go 影子验证，再切换正式写路径，最后才停止 Node 运行时。

**Tech Stack:** Go 1.23、`net/http`、MySQL、Redis、TOS、FFmpeg、Chrome/Chromium DevTools Protocol、React/Vite 构建产物、Nginx。

## Global Constraints

- 维护分支仅为 `v88`；所有实现先入 Git，禁止直接修改公网。
- 生产切换只能使用已批准的 V88 Direct Deploy Stage → Cutover 路径，禁止日常 Docker/GHCR 发布。
- 用户可见名称必须为“视频管理系统”，不得显示“121”。
- 在 V88 第七阶段前不得删除 Node 源码、历史分支或历史工作流；仅停止其运行时职责。
- Go 主服务、Worker、连接器和本地执行器不得调用 Node、npm、Node bridge 或 Playwright。
- React 仅保留构建时依赖；生产仅由 Go/Nginx 提供已构建静态文件。
- 不复制公网模型、TOS 或视频管理系统密钥；真实验证使用显式提供的隔离配置。
- 每项行为变更必须先写失败测试并确认失败，再写最小实现；保留现有工作区无关改动。

---

## File Structure

- `backend/cmd/qiantie/main.go`：Go HTTP 运行时组装、无 Node bridge 的路由选择、运行时证据。
- `backend/cmd/qiantie/main_test.go`：证明 API 和前端均由 Go 服务，不会路由到 Node。
- `backend/internal/runtimeaudit/`：bridge 命中计数、组件版本及健康状态。
- `backend/internal/batchfactoryv11/`：视频供应商统一接口、生产与轮询任务状态机。
- `backend/internal/batchfactoryv11/*_adapter.go`：H3、Seedance、Yadi、本地豆包的 Go HTTP 适配。
- `backend/cmd/merge-worker/` 与 `backend/internal/mergeworker/`：Redis + FFmpeg + TOS 合成 Worker。
- `backend/internal/videomanagementconnector/`：视频管理系统的会话、目录、上传、提交、回读 Go 连接器。
- `backend/cmd/video-management-connector/`：连接器 HTTP 进程。
- `backend/internal/batchfactoryv11/external/`：发布 intent、凭据加密、提交和审计的 Go 编排。
- `executor/`：新的 Go Windows 本地执行器模块与发布脚本。
- `deploy/v88-public/`：Go 服务的运行定义；最终删除 Node/Nginx Node upstream 的依赖。
- `scripts/`：本地构建、验证和精确 SHA 切换检查。

## Task 1: 建立全局 Node 依赖清单与零命中运行时审计

**Files:**
- Create: `backend/internal/runtimeaudit/audit.go`
- Create: `backend/internal/runtimeaudit/audit_test.go`
- Modify: `backend/cmd/qiantie/main.go`
- Modify: `backend/internal/httpapi/runtime_build_info.go`
- Create: `scripts/verify-no-node-runtime.ps1`
- Test: `backend/cmd/qiantie/main_test.go`

**Interfaces:**
- Produces `runtimeaudit.Registry` with `RecordBridge(path string)`, `Snapshot() Snapshot` and `AssertNoBridge() error`.
- Extends build info JSON with `runtime`, `gitSha`, `bridgeHits`, `components`.
- `verify-no-node-runtime.ps1` exits nonzero when Node/npm/node.exe, Node container or an Nginx Node upstream is detected.

- [ ] **Step 1: Write failing audit tests**

```go
func TestRegistryCountsLegacyBridgeHits(t *testing.T) {
    registry := runtimeaudit.NewRegistry()
    registry.RecordBridge("/api/script-video")
    got := registry.Snapshot()
    if got.BridgeHits != 1 || got.Paths["/api/script-video"] != 1 {
        t.Fatalf("snapshot = %#v", got)
    }
}

func TestBuildInfoReportsGoRuntimeAndBridgeCounter(t *testing.T) {
    h := runtimeBuildInfoHandler(runtimeaudit.NewRegistry(), "abc123")
    res := httptest.NewRecorder()
    h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/runtime-build-info", nil))
    if !strings.Contains(res.Body.String(), `"runtime":"go"`) { t.Fatal(res.Body.String()) }
}
```

- [ ] **Step 2: Run the focused tests and verify red**

Run: `go test ./internal/runtimeaudit ./internal/httpapi -run "TestRegistry|TestBuildInfoReportsGoRuntime" -count=1`

Expected: FAIL because `runtimeaudit` and the extended build-info contract do not exist.

- [ ] **Step 3: Implement minimal request audit and build-info output**

Implement a mutex-protected registry. Instrument every remaining legacy bridge branch in `cmd/qiantie/main.go` before it proxies. Build-info must expose only counts and route names, never secrets or full request data.

- [ ] **Step 4: Write failing process/config scanner tests**

Create a temporary compose and Nginx fixture containing `node:20`, `node app.js`, and `proxy_pass http://node`; assert the PowerShell verifier returns nonzero. Create a Go-only fixture and assert zero.

- [ ] **Step 5: Implement `verify-no-node-runtime.ps1`**

The script must inspect Windows/Linux process lists when available, configured Compose services, and Nginx upstream files passed through `-ConfigRoots`. It must output each detected Node evidence item and never stop/restart a process.

- [ ] **Step 6: Verify and commit**

Run: `go test ./internal/runtimeaudit ./internal/httpapi ./cmd/qiantie -count=1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/verify-no-node-runtime.ps1 -ConfigRoots deploy`

Expected: tests pass; verifier reports the current known Node dependencies until later tasks remove them.

Commit:

```powershell
git add backend/internal/runtimeaudit backend/internal/httpapi/runtime_build_info.go backend/cmd/qiantie/main.go backend/cmd/qiantie/main_test.go scripts/verify-no-node-runtime.ps1
git commit -m "feat: audit legacy node runtime dependencies"
```

## Task 2: Close Go video-provider parity for all official production paths

**Files:**
- Modify: `backend/internal/batchfactoryv11/production.go`
- Modify: `backend/internal/batchfactoryv11/video_provider.go`
- Create: `backend/internal/batchfactoryv11/yadi_video_adapter.go`
- Create: `backend/internal/batchfactoryv11/yadi_video_adapter_test.go`
- Modify: `backend/internal/batchfactoryv11/autodl_h3_video_adapter.go`
- Modify: `backend/internal/batchfactoryv11/yfai_seedance_adapter.go`
- Modify: `backend/internal/httpapi/script_video_local.go`
- Modify: `backend/cmd/qiantie/main.go`
- Test: `backend/internal/batchfactoryv11/*_adapter_test.go`

**Interfaces:**
- Produces `VideoProvider` with `Submit`, `Poll`, `Cancel` implemented for AutoDL H3, YFAI Seedance, Yadi and local Doubao.
- Produces normalized `ProviderTaskRef{ProviderTaskID, State, MediaURL, ErrorMessage}`.
- `/api/script-video` Go handler accepts only registered Go providers; unknown providers return a stable client error, never a Node proxy.

- [ ] **Step 1: Write failing Yadi adapter tests**

Use `httptest.NewServer` to assert Yadi create and status paths, `Authorization` handling, prompt/reference-image payload, state normalization, terminal media URL validation and cancellation.

```go
func TestYadiAdapterSubmitAndPoll(t *testing.T) {
    server := newYadiFixtureServer(t)
    adapter := &YadiAdapter{BaseURL: server.URL, APIKey: "test-key", Client: server.Client()}
    task, err := adapter.Submit(context.Background(), frozenYadiModel(), finalPrompt())
    if err != nil || task.ProviderTaskID != "yd-task-1" { t.Fatalf("task=%#v err=%v", task, err) }
    polled, err := adapter.Poll(context.Background(), frozenYadiModel(), task)
    if err != nil || polled.State != ProductionSucceeded { t.Fatalf("poll=%#v err=%v", polled, err) }
}
```

- [ ] **Step 2: Run adapter tests and verify red**

Run: `go test ./internal/batchfactoryv11 -run TestYadiAdapter -count=1`

Expected: FAIL because `YadiAdapter` does not exist.

- [ ] **Step 3: Implement the Yadi adapter and cancellation contract**

Use `net/http`, a caller-supplied client with bounded timeouts, exact provider endpoint paths, status normalization and safe error redaction. Extend the existing H3 and Seedance adapters to implement the same `Cancel` behavior and add equivalent tests.

- [ ] **Step 4: Remove script-video Node fallback only after provider coverage is green**

Add Go HTTP tests for H3, Seedance, Yadi and local Doubao requests. Then change `cmd/qiantie/main.go` so `/api/script-video` always reaches Go; unsupported model keys return `400 {"error":"unsupported video provider"}` and increment no bridge metric.

- [ ] **Step 5: Verify and commit**

Run: `go test ./internal/batchfactoryv11 ./internal/httpapi ./cmd/qiantie -count=1`

Expected: PASS; a source scan confirms `/api/script-video` is absent from legacy bridge matching.

Commit:

```powershell
git add backend/internal/batchfactoryv11 backend/internal/httpapi/script_video_local.go backend/cmd/qiantie/main.go
git commit -m "feat: run all video providers through Go"
```

## Task 3: Deployable Go merge worker and Go-owned merge API

**Files:**
- Modify: `backend/cmd/merge-worker/main.go`
- Modify: `backend/internal/mergeworker/worker.go`
- Modify: `backend/internal/mergeworker/ffmpeg.go`
- Create: `backend/internal/mergeworker/worker_integration_test.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_merge.go`
- Modify: `deploy/v88-public/docker-compose.yml`
- Create: `scripts/verify-merge-worker.ps1`
- Test: `backend/internal/mergeworker/*_test.go`

**Interfaces:**
- `merge-worker` requires Redis, TOS and `FFMPEG_BIN`; `/healthz` returns `200` only when dependencies are configured.
- API enqueues one idempotent job for an owner/batch/book/request tuple and returns persisted state.
- Worker updates `queued → running → succeeded|failed`, uploads only non-empty MP4 output, and survives an individual job failure.

- [ ] **Step 1: Write failing end-to-end worker test with fakes**

```go
func TestWorkerPersistsSucceededMergedArtifact(t *testing.T) {
    store, queue, downloader, merger, output := newWorkerFakes(t)
    job := seedQueuedJob(t, store, queue)
    worker := mergeworker.Worker{Store: store, Queue: queue, Downloader: downloader, Merger: merger, Output: output, WorkRoot: t.TempDir()}
    if err := worker.ProcessOne(context.Background(), time.Millisecond); err != nil { t.Fatal(err) }
    got, _ := store.Get(context.Background(), job.ID)
    if got.Status != mergeworker.StateSucceeded || got.OutputURL == "" { t.Fatalf("job=%#v", got) }
}
```

- [ ] **Step 2: Run the integration test and verify red**

Run: `go test ./internal/mergeworker -run TestWorkerPersistsSucceededMergedArtifact -count=1`

Expected: FAIL until worker dependencies expose deterministic fakes and the complete lifecycle is asserted.

- [ ] **Step 3: Implement explicit dependency checks and durable lifecycle**

Keep the existing failure-isolated loop, but add `FFMPEG_BIN` validation, readiness output, bounded download limits and progress updates. Do not use Node-generated concat scripts; invoke FFmpeg from Go with explicit argument arrays.

- [ ] **Step 4: Add merge-worker service to the runtime definition**

Add a Go `merge-worker` service with the same MySQL/Redis/TOS configuration source as Go API, an explicit FFmpeg path, health check and no Node image/command. Do not remove current infrastructure in this task.

- [ ] **Step 5: Verify and commit**

Run: `go test ./internal/mergeworker ./internal/httpapi -count=1`

Run: `go build ./cmd/merge-worker`

Run: `powershell -ExecutionPolicy Bypass -File scripts/verify-merge-worker.ps1 -Binary backend\merge-worker.exe`

Commit:

```powershell
git add backend/cmd/merge-worker backend/internal/mergeworker backend/internal/httpapi/batch_factory_v11_merge.go deploy/v88-public/docker-compose.yml scripts/verify-merge-worker.ps1
git commit -m "feat: run video merging with Go worker"
```

## Task 4: Build the Go 视频管理系统 browser connector

**Files:**
- Modify: `backend/go.mod`
- Create: `backend/internal/videomanagementconnector/types.go`
- Create: `backend/internal/videomanagementconnector/client.go`
- Create: `backend/internal/videomanagementconnector/session_store.go`
- Create: `backend/internal/videomanagementconnector/chrome.go`
- Create: `backend/internal/videomanagementconnector/server.go`
- Create: `backend/internal/videomanagementconnector/*_test.go`
- Create: `backend/cmd/video-management-connector/main.go`
- Create: `backend/cmd/video-management-connector/main_test.go`

**Interfaces:**
- Connector exposes `POST /v1/sessions/login`, `POST /v1/catalogs/refresh`, `POST /v1/submissions`, `GET /v1/submissions/{id}`, `GET /healthz`.
- Connector accepts an internal bearer token and rejects all unauthenticated calls.
- Session records are encrypted at rest and scoped by owner; submission responses expose provider reference/status only.
- Use a Go CDP client compatible with Go 1.23; never import Node, Playwright or execute npm.

- [ ] **Step 1: Write failing connector HTTP authorization and session tests**

```go
func TestServerRejectsMissingInternalToken(t *testing.T) {
    s := videomanagementconnector.NewServer(videomanagementconnector.Config{InternalToken: "internal-token"}, fakeService{})
    res := httptest.NewRecorder()
    s.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/v1/sessions/login", strings.NewReader(`{}`)))
    if res.Code != http.StatusUnauthorized { t.Fatalf("status=%d", res.Code) }
}

func TestSessionStoreDoesNotPersistPlainCookie(t *testing.T) {
    store := newTestSessionStore(t)
    if err := store.Save(context.Background(), "alice", Session{Cookie: "secret-cookie"}); err != nil { t.Fatal(err) }
    raw := readStoredRow(t)
    if strings.Contains(raw, "secret-cookie") { t.Fatal("plaintext cookie persisted") }
}
```

- [ ] **Step 2: Run connector tests and verify red**

Run: `go test ./internal/videomanagementconnector ./cmd/video-management-connector -count=1`

Expected: FAIL because the connector package and binary do not exist.

- [ ] **Step 3: Implement CDP session lifecycle and catalog read**

Add the selected Go CDP dependency to `go.mod`. Implement dedicated Chrome process profile, login action, encrypted cookie/session persistence, session validation, refresh and organization/style catalog extraction. All Chrome automation must have a context deadline and redact cookies/tokens from errors.

- [ ] **Step 4: Implement uploads, submit and read-back with contract tests**

Use HTTP fixtures to test pre-signed upload, multipart TXT upload, final submission and polling. Source media must be connector input supplied by the Go main service; reject arbitrary browser-provided URLs. Add a fixture asserting final provider reference and terminal status are retained.

- [ ] **Step 5: Verify and commit**

Run: `go test ./internal/videomanagementconnector ./cmd/video-management-connector -count=1`

Run: `go build ./cmd/video-management-connector`

Commit:

```powershell
git add backend/go.mod backend/go.sum backend/internal/videomanagementconnector backend/cmd/video-management-connector
git commit -m "feat: add Go video management connector"
```

## Task 5: Wire Go publish orchestration to the Go connector

**Files:**
- Create: `backend/internal/batchfactoryv11/external/video_management_provider.go`
- Create: `backend/internal/batchfactoryv11/external/video_management_provider_test.go`
- Modify: `backend/internal/batchfactoryv11/external/service.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_external.go`
- Modify: `frontend/src/user/pages/batch-factory-v11/capabilityReadiness.js`
- Test: `backend/internal/httpapi/batch_factory_v11_external_test.go`

**Interfaces:**
- `VideoManagementProvider` implements `external.SubmissionProvider` and calls only the Go connector HTTP contract.
- `QIANTIE_VIDEO_MANAGEMENT_CONNECTOR_URL` and `QIANTIE_VIDEO_MANAGEMENT_CONNECTOR_TOKEN` configure it.
- Existing `intent → confirm → submit → audit` protocol remains unchanged.
- Capability copy is `视频管理系统发布`.

- [ ] **Step 1: Write failing provider contract tests**

```go
func TestVideoManagementProviderSubmitsConfirmedIntent(t *testing.T) {
    connector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/v1/submissions" { t.Fatalf("path=%s", r.URL.Path) }
        if r.Header.Get("Authorization") != "Bearer connector-token" { t.Fatal("missing connector auth") }
        io.WriteString(w, `{"reference":"publish-42","status":"submitted"}`)
    }))
    provider := external.NewVideoManagementProvider(connector.URL, "connector-token", connector.Client())
    got, err := provider.Submit(context.Background(), credential(), confirmedIntent())
    if err != nil || got.Reference != "publish-42" { t.Fatalf("got=%#v err=%v", got, err) }
}
```

- [ ] **Step 2: Run tests and verify red**

Run: `go test ./internal/batchfactoryv11/external -run TestVideoManagementProvider -count=1`

Expected: FAIL because the provider does not exist.

- [ ] **Step 3: Implement connector provider and configuration**

Implement the HTTP client with timeouts and redacted errors. Wire it in `app.Build`; enable `Provider121` only when connector URL/token and encryption key are valid. Preserve Yadi independent gating.

- [ ] **Step 4: Change public copy and test Go-only publish path**

Replace user-facing `121 视频发布` with `视频管理系统发布`. Add HTTP tests proving intent creation, confirmation, submission and audit invoke the connector provider without an old Node endpoint.

- [ ] **Step 5: Verify and commit**

Run: `go test ./internal/batchfactoryv11/external ./internal/httpapi ./internal/app -count=1`

Run: `node --test frontend/src/user/pages/batch-factory-v11/capabilityReadiness.test.js`

Commit:

```powershell
git add backend/internal/batchfactoryv11/external backend/internal/app/app.go backend/internal/config/config.go backend/internal/httpapi/batch_factory_v11_external.go frontend/src/user/pages/batch-factory-v11/capabilityReadiness.js
git commit -m "feat: publish video management submissions from Go"
```

## Task 6: Replace the Windows Node local-executor with Go binary

**Files:**
- Create: `executor/go.mod`
- Create: `executor/cmd/qiantie-executor/main.go`
- Create: `executor/internal/protocol/client.go`
- Create: `executor/internal/protocol/client_test.go`
- Create: `executor/internal/agent/runner.go`
- Create: `executor/internal/agent/runner_test.go`
- Create: `executor/internal/update/checker.go`
- Create: `executor/internal/update/checker_test.go`
- Create: `scripts/build-go-executor.ps1`
- Modify: `backend/internal/httpapi/local_executor*.go`
- Test: `backend/internal/httpapi/local_executor*_test.go`

**Interfaces:**
- `qiantie-executor.exe` supports `pair`, `run`, `status`, `version` commands.
- It preserves current signed request, pairing, heartbeat, lease, artifact-upload and update HTTP protocol until all users are upgraded.
- Server accepts both old and Go executor protocol versions during migration; capability reporting identifies executor version.

- [ ] **Step 1: Write failing protocol fixture tests**

Capture sanitized fixture requests/responses for pair, heartbeat, lease, completed output and rejected signature. Assert the Go client produces byte-equivalent canonical signing input and handles retryable vs terminal errors.

- [ ] **Step 2: Run tests and verify red**

Run: `go test ./executor/internal/protocol ./executor/internal/agent -count=1`

Expected: FAIL because the Go executor module does not exist.

- [ ] **Step 3: Implement the minimal Go protocol client and worker loop**

Implement pairing persistence with restrictive local file permissions, periodic heartbeat, one-at-a-time lease execution, input download, controlled browser/model action execution and artifact upload. Do not support arbitrary shell commands from server payloads.

- [ ] **Step 4: Implement signed update manifest verification**

Verify a versioned manifest signed by a configured public key before replacing the executable. Keep the previous executable until new process health check succeeds.

- [ ] **Step 5: Add server compatibility and build tests**

Update Go server routes to recognize `executorVersion` and return upgrade guidance to deprecated Node clients. Cross-compile Windows binary and run fixture tests.

Run: `go test ./executor/... ./internal/httpapi -count=1`

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-executor.ps1`

- [ ] **Step 6: Commit**

```powershell
git add executor backend/internal/httpapi scripts/build-go-executor.ps1
git commit -m "feat: add Go Windows local executor"
```

## Task 7: Migrate remaining Node-only public APIs by generated contract inventory

**Files:**
- Create: `scripts/inventory-node-api-routes.ps1`
- Create: `docs/migration/node-api-contract-inventory.json`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/cmd/qiantie/main.go`
- Create: `backend/internal/httpapi/node_contract_parity_test.go`
- Modify: affected `backend/internal/httpapi/*.go` per inventory result
- Test: `backend/internal/httpapi/node_contract_parity_test.go`

**Interfaces:**
- Inventory JSON has `{method,path,owner,status,goHandler,test}` for every Node public API route.
- `status` is `migrated`, `shadow`, or `blocked`; production cutover requires no `shadow` or `blocked` record.
- Go contract tests issue every migrated request against Go fixtures and verify status and schema.

- [ ] **Step 1: Write failing inventory completeness test**

```go
func TestNodeRouteInventoryHasNoUnmigratedPublicRoute(t *testing.T) {
    routes := loadRouteInventory(t)
    for _, route := range routes {
        if route.Status != "migrated" { t.Fatalf("unmigrated route: %s %s", route.Method, route.Path) }
    }
}
```

- [ ] **Step 2: Run it and verify red**

Run: `go test ./internal/httpapi -run TestNodeRouteInventoryHasNoUnmigratedPublicRoute -count=1`

Expected: FAIL with the first Node-owned route; this creates the ordered migration backlog.

- [ ] **Step 3: Generate and review the inventory**

Implement a read-only PowerShell scanner that extracts Express route declarations and bridge matchers into JSON. Classify only externally reachable routes; do not treat internal development helpers as production APIs. Record source file/line and matching Go handler/test.

- [ ] **Step 4: Migrate in contract groups**

Complete one group at a time: authentication/account; novel fetch; novel panel; script generation; TTS; settings; asset tools. For each route group: add an HTTP fixture test, run red, implement a Go handler/service/store, run green, update JSON to `migrated`, then commit. Do not bypass groups with catch-all proxy behavior.

- [ ] **Step 5: Verify and commit each route group**

Run after each group: `go test ./internal/httpapi ./internal/app -count=1`

Final run: `go test ./... -count=1`

Commit format:

```powershell
git add backend/internal/httpapi backend/internal/app docs/migration/node-api-contract-inventory.json scripts/inventory-node-api-routes.ps1
git commit -m "feat: migrate <route-group> API to Go"
```

## Task 8: Go-only staging cutover, external acceptance and Node runtime retirement

**Files:**
- Modify: `deploy/v88-public/docker-compose.yml`
- Modify: `deploy/v88-public/nginx.conf`
- Create: `scripts/verify-go-only-runtime.ps1`
- Create: `scripts/verify-video-production-e2e.ps1`
- Create: `scripts/verify-video-management-e2e.ps1`
- Modify: `.github/workflows/v88-direct-deploy-node-stage.yml`
- Modify: `.github/workflows/v88-direct-deploy-node-cutover.yml`
- Test: `tests/v88-go-only-runtime.test.js`

**Interfaces:**
- Production service set is Go `qiantie`, Go `merge-worker`, Go `video-management-connector`, MySQL, Redis, Nginx and optional Go executor update host; it has no Node service/image/command.
- Nginx upstream points only to Go HTTP services.
- Stage verification emits exact Git SHA, runtime `go`, zero bridge hits, FFmpeg/Chromium readiness and all worker health states.

- [ ] **Step 1: Write failing compose and Nginx absence tests**

```js
test('public runtime contains no Node image, command, or upstream', () => {
  const compose = fs.readFileSync('deploy/v88-public/docker-compose.yml', 'utf8');
  const nginx = fs.readFileSync('deploy/v88-public/nginx.conf', 'utf8');
  assert.doesNotMatch(compose, /node:|npm |app\.js|browser-worker/i);
  assert.doesNotMatch(nginx, /node|browser-worker/i);
});
```

- [ ] **Step 2: Run it and verify red**

Run: `node --test tests/v88-go-only-runtime.test.js`

Expected: FAIL because current Compose and Nginx still reference Node/browser worker.

- [ ] **Step 3: Change deployment only after Tasks 1–7 show zero unmigrated routes**

Replace Node and Node browser-worker service definitions with Go binaries. Configure health checks for API, merge-worker and connector. Route Nginx to Go API. Preserve MySQL/Redis/Nginx and use exact configuration variable names established by prior tasks.

- [ ] **Step 4: Run isolated external acceptance**

With explicitly supplied non-production credentials:

1. Create a Go video task for each configured provider and poll terminal state.
2. Submit a merge of fixture media and verify TOS output is non-empty MP4.
3. Login to video management system, refresh catalog, submit one isolated artifact, and verify returned reference/status.
4. Pair a Go executor, run one harmless fixture task, upload artifact, restart executor, and verify lease recovery.

Each verification script must fail closed when credentials/configuration are missing; it must never output credentials.

- [ ] **Step 5: Stage, cut over and verify exact SHA**

Run local suite: `go test ./... -count=1`

Run frontend suite/build: `npm --prefix frontend test; npm --prefix frontend run build`

Run no-Node verifier: `powershell -ExecutionPolicy Bypass -File scripts/verify-go-only-runtime.ps1 -ConfigRoots deploy`

After Git commit and approved V88 Stage → Cutover, run the same verifier against ECS process/container configuration and call `/api/runtime-build-info`. Require exact staged SHA, `runtime=go`, `bridgeHits=0` and all Go worker health checks.

- [ ] **Step 6: Commit deployment migration**

```powershell
git add deploy/v88-public scripts .github/workflows tests/v88-go-only-runtime.test.js
git commit -m "feat: run V88 production without Node runtime"
```

## Spec Coverage Review

- Go main service/static/API/bridge audit: Tasks 1, 7, 8.
- H3/Seedance/Yadi/Doubao video execution: Task 2.
- Redis/FFmpeg/TOS video merge: Task 3.
- Go replacement for Node/Playwright 视频管理系统 worker: Tasks 4 and 5.
- Go Windows local executor: Task 6.
- Full public API Node removal: Task 7.
- Production Node shutdown, exact SHA verification, zero Node process/container/upstream: Task 8.
- Node source preservation before V88 stage 7: Global Constraints and Task 8.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-10-13-full-go-runtime-migration.md`.

Recommended execution method: **Subagent-Driven** — one fresh reviewer-gated agent per task, because the migration spans independent Go services, external contracts and deployment controls. Inline execution is possible but should preserve the task checkpoints above.

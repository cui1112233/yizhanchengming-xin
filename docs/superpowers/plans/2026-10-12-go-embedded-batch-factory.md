# Go Embedded Batch Factory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Batch Factory run from one Go binary that embeds the built React frontend, serves Batch Factory APIs directly, and calls external capabilities through independent HTTP plugins instead of Node bridges.

**Architecture:** React remains the authoring UI, but its production build is embedded under `backend/web/dist` and served by Go. The Batch Factory V12 API stays same-origin and is already registered by Go; the work removes the remaining Node-only Batch Factory bridge from the runtime path. Giant-material resolution becomes a Go-owned HTTP client contract, with the existing Go resolver implementation retained behind that contract.

**Tech Stack:** Go `net/http`, `embed`, React/Vite, MySQL, Redis, Go tests, HTTP test servers.

## Global Constraints

- `v88` is the maintained source branch; do not modify production directly.
- Do not change Batch Factory UI or business semantics in this migration.
- The production binary must embed the built frontend; Node must not serve pages or Batch Factory APIs.
- External capabilities use a versioned HTTP contract and cannot access the main database.
- Preserve existing unrelated working-tree changes.
- No public deployment in this plan; verification is local only.

---

## File Structure

- `backend/web/embed.go` embeds the built `backend/web/dist` tree and exposes an `fs.FS` limited to that directory.
- `backend/internal/web/static.go` owns safe static assets, SPA fallback, and the special admin/legacy entry pages.
- `backend/cmd/qiantie/main.go` composes the Go API and embedded web handler without a Node fallback for Batch Factory routes.
- `backend/cmd/qiantie/main_test.go` proves the production HTTP composition keeps `/api/batch-factory/` in Go and serves the SPA path locally.
- `backend/internal/giantmaterialresolver/http_plugin.go` owns the versioned HTTP plugin client and error mapping.
- `backend/internal/giantmaterialresolver/http_plugin_test.go` proves the HTTP request/response contract and failure mapping.
- `backend/internal/httpapi/router.go` injects the resolver client into the existing giant-material API handlers.
- `backend/internal/httpapi/giant_material_resolver_test.go` verifies the public Go endpoint uses the plugin client rather than a Node bridge.
- `frontend/vite.config.js` writes one deterministic production directory consumed by Go; its dev-only giant-material helper remains under the private `__local` prefix.
- `scripts/build-go-embedded-frontend.ps1` performs the repeatable local React build and copies it into the Go embed directory.

## Task 1: Make embedded-frontend input deterministic

**Files:**
- Create: `scripts/build-go-embedded-frontend.ps1`
- Modify: `frontend/vite.config.js:163-184`
- Modify: `backend/web/embed.go`
- Test: `backend/internal/web/static_test.go`

**Interfaces:**
- Produces `backend/web/dist` as the only directory compiled into the Go binary.
- Exposes `web.EmbeddedFiles() (fs.FS, error)` for main package composition.

- [ ] **Step 1: Write failing tests for the embedded filesystem**

Add a test in `backend/internal/web/static_test.go` that calls `EmbeddedFiles()`, checks that `index.html` can be read, and checks that `StaticHandler(EmbeddedFiles())` returns `200` for `/batch-factory`.

```go
func TestEmbeddedFilesServeBatchFactoryRoute(t *testing.T) {
    files, err := EmbeddedFiles()
    if err != nil { t.Fatal(err) }
    if _, err := fs.ReadFile(files, "index.html"); err != nil { t.Fatal(err) }
    res := httptest.NewRecorder()
    StaticHandler(files).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/batch-factory", nil))
    if res.Code != http.StatusOK { t.Fatalf("status = %d", res.Code) }
}
```

- [ ] **Step 2: Run the test to verify current embed wiring fails or is incomplete**

Run: `go test ./internal/web -run TestEmbeddedFilesServeBatchFactoryRoute -count=1`

Expected: FAIL until `EmbeddedFiles` and a committed non-empty `backend/web/dist` input are available.

- [ ] **Step 3: Implement deterministic assets and embed export**

Create `backend/web/embed.go` using `//go:embed dist/*` and `fs.Sub`, and add the build script:

```powershell
Push-Location frontend
npm run build
Pop-Location
Remove-Item -LiteralPath backend/web/dist -Recurse -Force -ErrorAction SilentlyContinue
Copy-Item -LiteralPath frontend/dist -Destination backend/web/dist -Recurse
```

Set Vite `build.outDir` explicitly to `dist` and keep `emptyOutDir: true` so the script always copies one clean build output.

- [ ] **Step 4: Run focused tests and build**

Run: `go test ./internal/web -count=1`

Expected: PASS.

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-embedded-frontend.ps1; go build ./cmd/qiantie`

Expected: frontend builds and Go produces one binary with embedded assets.

- [ ] **Step 5: Commit the isolated foundation**

```powershell
git add backend/web/embed.go backend/internal/web/static_test.go frontend/vite.config.js scripts/build-go-embedded-frontend.ps1
git commit -m "feat: embed built frontend in Go runtime"
```

## Task 2: Make the Go executable the Batch Factory web entrypoint

**Files:**
- Modify: `backend/cmd/qiantie/main.go:48-262`
- Modify: `backend/cmd/qiantie/main_test.go`
- Test: `backend/cmd/qiantie/main_test.go`

**Interfaces:**
- Consumes `webassets.Handler()` backed by `EmbeddedFiles()`.
- Produces a handler where `GET /batch-factory` is static React and every `/api/batch-factory/v12/*` request goes to `app.Build`.

- [ ] **Step 1: Write failing composition tests**

Extract main's long handler decision into `newRuntimeHandler(apiHandler, staticHandler, options)` and test:

```go
func TestRuntimeHandlerServesBatchFactoryFromEmbeddedWeb(t *testing.T) {
    api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
    static := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("batch factory")) })
    res := httptest.NewRecorder()
    newRuntimeHandler(api, static, runtimeOptions{}).ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/batch-factory", nil))
    if res.Code != http.StatusOK || res.Body.String() != "batch factory" { t.Fatal("route was not static") }
}
```

Add a companion test asserting `/api/batch-factory/v12/batches` reaches `api`, never a proxy handler.

- [ ] **Step 2: Run the new test**

Run: `go test ./cmd/qiantie -run TestRuntimeHandler -count=1`

Expected: FAIL because the current route decisions are embedded in `main`.

- [ ] **Step 3: Extract the handler without changing unrelated migration bridges**

Move only the handler selection code from `main` into `newRuntimeHandler`. Route all `/api/batch-factory/` paths directly to `apiHandler` before any legacy proxy checks. Preserve other module bridge rules unchanged; do not remove Node fallback for login, novel panel, or legacy novel-fetch in this task.

- [ ] **Step 4: Verify the Go-only Batch Factory entrypoint**

Run: `go test ./cmd/qiantie ./internal/web -count=1`

Expected: PASS.

Run: `go build ./cmd/qiantie`

Expected: PASS.

- [ ] **Step 5: Commit the routing seam**

```powershell
git add backend/cmd/qiantie/main.go backend/cmd/qiantie/main_test.go
git commit -m "refactor: serve batch factory directly from Go"
```

## Task 3: Define the giant-material HTTP plugin client

**Files:**
- Create: `backend/internal/giantmaterialresolver/http_plugin.go`
- Create: `backend/internal/giantmaterialresolver/http_plugin_test.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/app/build.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/httpapi/giant_material_resolver_test.go`

**Interfaces:**
- Defines `type HTTPPluginClient struct { BaseURL string; Token string; Client *http.Client }`.
- Defines `func (c *HTTPPluginClient) Resolve(ctx context.Context, request ResolveRequest) (ResolveResult, error)`.
- Uses `POST {BaseURL}/v1/giant-material/resolve` with JSON `{ "giantMaterialId": "..." }` and header `Authorization: Bearer <token>` when configured.
- Returns typed errors `ErrPluginUnavailable`, `ErrPluginTimeout`, `ErrPluginUpstream` for HTTP 503, deadline exceeded, and plugin 502 responses.

- [ ] **Step 1: Write failing client contract tests**

Use `httptest.NewServer` to assert the method, path, JSON body and token. Add table cases for success, `503`, malformed JSON, and a context deadline.

```go
if r.URL.Path != "/v1/giant-material/resolve" { t.Fatalf("path = %s", r.URL.Path) }
if got := r.Header.Get("Authorization"); got != "Bearer plugin-token" { t.Fatalf("authorization = %q", got) }
```

- [ ] **Step 2: Run the client tests to prove they fail**

Run: `go test ./internal/giantmaterialresolver -run TestHTTPPluginClient -count=1`

Expected: FAIL because `HTTPPluginClient` does not yet exist.

- [ ] **Step 3: Implement the narrow client and configuration**

Add `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` and `QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN` to the Go runtime config. Select the HTTP client only when the URL is configured; otherwise retain the existing in-process resolver so local behavior remains available during migration.

- [ ] **Step 4: Inject the configured resolver into the existing Go router**

Pass the selected resolver through `app.Build` into `httpapi.RouterOptions.GiantMaterialResolver`. Do not add any Node address, Node bridge secret, or reverse proxy dependency.

- [ ] **Step 5: Verify contract and API error mapping**

Run: `go test ./internal/giantmaterialresolver ./internal/httpapi -run "TestHTTPPluginClient|TestGiantMaterialResolver" -count=1`

Expected: PASS; public Go API preserves `INVALID_GIANT_MATERIAL_ID`, distinguishes plugin unavailable/timeout/upstream errors, and never emits a Node proxy error.

- [ ] **Step 6: Commit the plugin boundary**

```powershell
git add backend/internal/giantmaterialresolver backend/internal/config/config.go backend/internal/app/build.go backend/internal/httpapi/router.go backend/internal/httpapi/giant_material_resolver_test.go
git commit -m "feat: add giant material HTTP plugin client"
```

## Task 4: Stop Batch Factory from using Node-only API assumptions

**Files:**
- Modify: `frontend/vite.config.js:176-184`
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/shared/api/batchFactoryV11.test.js`
- Modify: `frontend/src/user/pages/giantMaterialExtractionClient.js`
- Test: `frontend/src/shared/api/batchFactoryV11.test.js`

**Interfaces:**
- All production Batch Factory calls use same-origin `/api/batch-factory/v12` and `/api/shuihuo-production/*` Go routes.
- Vite proxy settings are development-only and may point only at the Go local server.

- [ ] **Step 1: Write failing frontend source tests**

Add assertions that the Batch Factory API module contains no `QIANTIE_GO_BASE_URL`, port `3001`, `proxyV11Request`, or Node-specific endpoint origin.

```js
assert.doesNotMatch(source, /QIANTIE_GO_BASE_URL|proxyV11Request|:3001/);
assert.match(source, /const BASE = '\/api\/batch-factory\/v12'/);
```

- [ ] **Step 2: Run the test**

Run: `npm --prefix frontend test -- batchFactoryV11.test.js`

Expected: FAIL if any client-side Batch Factory path still relies on a Node bridge.

- [ ] **Step 3: Remove Batch Factory-specific Node development proxy paths**

Keep Vite only as a local UI development tool. Route the Batch Factory namespaces to `QIANTIE_DEV_API_ORIGIN` (the Go service); keep `__local/giant-material-test/*` private and unavailable in production.

- [ ] **Step 4: Run frontend tests and production build**

Run: `npm --prefix frontend test -- batchFactoryV11.test.js`

Expected: PASS.

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-embedded-frontend.ps1`

Expected: PASS and refreshes `backend/web/dist`.

- [ ] **Step 5: Commit frontend API cleanup**

```powershell
git add frontend/vite.config.js frontend/src/shared/api/batchFactoryV11.js frontend/src/shared/api/batchFactoryV11.test.js frontend/src/user/pages/giantMaterialExtractionClient.js backend/web/dist
git commit -m "refactor: use Go API for batch factory frontend"
```

## Task 5: Perform local, visible Batch Factory acceptance

**Files:**
- Modify: `docs/superpowers/specs/2026-10-12-go-embedded-batch-factory-design.md`
- Test: Go tests, frontend tests, local browser acceptance record in the design document

**Interfaces:**
- Consumes the Go binary, embedded frontend and direct V12 routes from Tasks 1-4.
- Produces evidence for proceeding to Node removal only after real UI verification.

- [ ] **Step 1: Start only the local Go runtime and required data services**

Use the existing review configuration for MySQL/Redis and start the Go binary without Node. Set `QIANTIE_DEV_API_ORIGIN` to the Go listener for Vite-only development; do not set `QIANTIE_NODE_AUTH_BASE_URL` for this acceptance.

- [ ] **Step 2: Verify the API and embedded routes**

Run: `Invoke-WebRequest http://127.0.0.1:<go-port>/health`

Expected: `200`.

Run: `Invoke-WebRequest http://127.0.0.1:<go-port>/batch-factory`

Expected: `200` HTML from the embedded React build.

- [ ] **Step 3: Run the browser acceptance flow**

Open `/batch-factory`, confirm batch list loads, open “新建批量,” add one draft book group, confirm “立即执行” unlocks, open the automation confirmation, then cancel. Do not create a persistent batch or invoke a model for this smoke test.

- [ ] **Step 4: Run all focused automated checks**

Run: `go test ./internal/web ./cmd/qiantie ./internal/giantmaterialresolver ./internal/httpapi -count=1`

Expected: PASS.

Run: `npm --prefix frontend test -- batchFactoryV11.test.js`

Expected: PASS.

- [ ] **Step 5: Record result and commit the evidence update**

Append the actual command results, effective Go port, browser outcome and any unverified external-plugin gap to the design document. Do not label plugin/provider generation as passed unless an actual configured plugin request succeeds.

```powershell
git add docs/superpowers/specs/2026-10-12-go-embedded-batch-factory-design.md
git commit -m "docs: record Go batch factory local acceptance"
```

## Plan Self-Review

- Spec coverage: Go `embed` is Task 1; direct Batch Factory routing is Task 2; HTTP plugin boundary is Task 3; frontend same-origin use is Task 4; local UI and automated acceptance are Task 5.
- Scope: this plan deliberately excludes migration of login, novel panel, script, TTS and other Node modules. They are follow-on plans after Batch Factory passes Task 5.
- Type consistency: `EmbeddedFiles`, `newRuntimeHandler`, `HTTPPluginClient.Resolve`, `RouterOptions.GiantMaterialResolver`, and the V12 same-origin base are defined before their consumers.
- Placeholder scan: no open implementation placeholders remain; exact files, route names, commands, failure expectations, and commit boundaries are stated.

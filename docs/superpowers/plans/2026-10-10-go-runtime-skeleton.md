# Go Runtime Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在当前 V88 本地建立一套可验证的 Go 主服务骨架，让 Go API、MySQL、Redis 和 React 静态资源具备统一启动入口，同时保留现有 Node 开发工具。

**Architecture:** Go 负责本地正式运行入口和 API；React 前台/后台继续用 npm 构建为静态文件；MySQL 是业务事实源，Redis 只用于运行基础设施。第一阶段不迁移旧 Node 业务路由，只把运行边界和健康检查做清楚。

**Tech Stack:** Go 1.23、`net/http`、MySQL driver、Redis、React + Ant Design、Vite、Go `embed`。

**Spec:** `docs/superpowers/specs/2026-10-10-yizhanchengming-xin-go-migration-design.md`

## Global Constraints

- 不把 React 页面重写成 Go 模板。
- 不提交密码、Cookie、Token、API Key 或真实环境文件。
- 不覆盖当前工作区已有未提交改动。
- MySQL 保存长期业务事实；Redis 只保存队列、锁和短期运行状态。
- Node/npm 只作为前端开发和打包工具，不作为新的正式业务后端。
- 所有验证必须区分本地骨架验证和真实 Provider/ECS 验收。

## Review Focus

- MySQL 不可用时，Go 健康检查必须返回明确的非 200，而不是假装服务正常；由 Task 1 的健康测试覆盖。
- Redis 不可用时，API 骨架仍能启动，但队列能力必须明确报告不可用；由 Task 2 的依赖状态测试覆盖。
- 前台或后台构建产物缺失时，Go 服务必须返回明确的静态资源错误；由 Task 3 的静态资源测试覆盖。
- 环境变量缺失时，启动错误不得打印密钥值；由 Task 1 的配置测试覆盖。
- Node 旧服务仍在运行时，不能让 Go 和 Node 抢同一个端口；由 Task 4 的启动检查覆盖。

---

### Task 1: Go 运行配置与健康检查

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/config/config_test.go`
- Test: `backend/internal/app/health_test.go`
- Test: `backend/internal/httpapi/health_test.go`

**Interfaces:**
- Consumes: `QIANTIE_GO_LISTEN_ADDR`, `QIANTIE_MYSQL_DSN`, `QIANTIE_BRIDGE_SECRET`。
- Produces: `GET /health`，返回 `{ "ok": true }` 或明确的 503 JSON；配置错误只返回变量名，不返回变量值。

- [ ] **Step 1: Write the failing tests**

  增加配置缺失测试和 MySQL 不可用健康检查测试，断言状态码与响应体，不断言具体密钥内容。

- [ ] **Step 2: Run tests to verify they fail**

  Run: `go test ./internal/config ./internal/app ./internal/httpapi`

  Expected: 新增断言在当前实现下失败。

- [ ] **Step 3: Implement the minimum configuration and health behavior**

  在现有配置读取和 Router 结构上补齐错误分类；不新增第二套配置格式，不把敏感值写入日志。

- [ ] **Step 4: Run tests to verify they pass**

  Run: `go test ./internal/config ./internal/app ./internal/httpapi`

  Expected: PASS。

- [ ] **Step 5: Commit**

  `git add backend/internal/config backend/internal/app backend/internal/httpapi && git commit -m "feat: harden Go runtime health checks"`

### Task 2: Redis 运行依赖边界

**Files:**
- Create: `backend/internal/queue/runtime.go`
- Modify: `backend/internal/app/app.go`
- Modify: `backend/internal/config/config.go`
- Test: `backend/internal/queue/runtime_test.go`

**Interfaces:**
- Consumes: `QIANTIE_REDIS_URL`（可选，未配置时不伪造队列可用）。
- Produces: `RuntimeQueueStatus`，包含 `Configured`、`Reachable` 和安全错误原因；长期任务状态仍由 MySQL 保存。

- [ ] **Step 1: Write the failing tests**

  覆盖未配置 Redis、Redis 地址格式错误和可注入 fake ping 三种情况。

- [ ] **Step 2: Run the focused test**

  Run: `go test ./internal/queue -run TestRuntimeQueue -count=1`

  Expected: FAIL，因为运行状态接口尚未存在。

- [ ] **Step 3: Implement `RuntimeQueueStatus` and dependency injection**

  使用小接口注入 ping/queue 实现，禁止在业务包里直接读取环境变量或保存长期业务状态。

- [ ] **Step 4: Run the focused test and full Go tests**

  Run: `go test ./internal/queue -run TestRuntimeQueue -count=1 && go test ./...`

  Expected: PASS。

- [ ] **Step 5: Commit**

  `git add backend/internal/queue backend/internal/app backend/internal/config && git commit -m "feat: define Redis runtime boundary"`

### Task 3: Go 提供 React 静态资源

**Files:**
- Create: `backend/internal/web/static.go`
- Create: `backend/web/dist/.gitkeep`（仅保证 Go 测试有可嵌入目录，真实文件由构建脚本生成）
- Modify: `backend/cmd/qiantie/main.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/vite.config.js`（只在需要统一输出目录时修改）
- Modify: `scripts/build-go-local.ps1`（若当前不存在则创建）
- Test: `backend/internal/web/static_test.go`

**Interfaces:**
- Consumes: `frontend/dist` 构建产物。
- Produces: Go 服务根路径和前端静态资源路由；API 路由优先于静态资源路由。

- [ ] **Step 1: Write the failing static-serving tests**

  用临时静态目录验证首页、带 hash 的资源和不存在资源的状态码；测试不依赖真实前端构建目录。

- [ ] **Step 2: Run the focused test**

  Run: `go test ./internal/web -run TestStatic -count=1`

  Expected: FAIL，因为静态服务包尚未存在。

- [ ] **Step 3: Implement static serving with an injectable filesystem**

  先实现可测试的 `http.Handler`；Go `embed` 只嵌入 `backend/web/dist`，不直接跨目录嵌入 `frontend/dist`。

- [ ] **Step 4: Build frontend, copy assets, and run Go tests**

  Run: `npm --prefix frontend run build`；`powershell -File scripts/build-go-local.ps1 -SkipGoBuild`；`go test ./internal/web ./internal/httpapi`

  Expected: 构建脚本把 `frontend/dist` 复制到 `backend/web/dist`，前端构建和静态服务测试均 PASS；生成目录不提交密钥和业务数据。

- [ ] **Step 5: Commit**

  `git add backend/internal/web backend/web/dist/.gitkeep backend/cmd/qiantie/main.go backend/internal/httpapi/router.go frontend/vite.config.js scripts/build-go-local.ps1 && git commit -m "feat: serve React assets from Go"`

### Task 4: 本地统一启动与端口验收

**Files:**
- Create: `scripts/start-local-go.ps1`
- Create: `scripts/check-local-go.ps1`
- Modify: `docs/README-local.md`（若当前文档不存在则创建）
- Test: `scripts/start-local-go.test.ps1` 或等价的 Go/PowerShell smoke check

**Interfaces:**
- Consumes: 不提交仓库的本地环境文件或已注入的环境变量。
- Produces: 一个明确的 Go API 端口、前端构建步骤、MySQL/Redis 健康检查和失败退出码。

- [ ] **Step 1: Write the failing smoke check**

  断言 Go 端口未被 Node 占用、`GET /health` 返回 200、静态首页可读取；任何一项失败都退出非 0。

- [ ] **Step 2: Run it against the current machine**

  Run: `powershell -File scripts/check-local-go.ps1`

  Expected: 当前混合环境至少暴露明确失败项，不得报告“全部正常”。

- [ ] **Step 3: Implement the start/check scripts**

  脚本只读取环境变量，不生成或提交密钥；不自动删除已有 Node/MySQL/Redis 进程。

- [ ] **Step 4: Run the complete local skeleton check**

  Run: `powershell -File scripts/start-local-go.ps1 -CheckOnly`；`powershell -File scripts/check-local-go.ps1`

  Expected: 端口、健康检查、静态首页和依赖状态都有明确结果。

- [ ] **Step 5: Commit**

  `git add scripts/start-local-go.ps1 scripts/check-local-go.ps1 docs/README-local.md && git commit -m "chore: add unified local Go runtime checks"`

### Task 5: 第一阶段验收与迁移边界记录

**Files:**
- Create: `docs/migration/local-go-runtime-baseline.md`
- Modify: `docs/superpowers/specs/2026-10-10-yizhanchengming-xin-go-migration-design.md`（只补实际验证结果）

**Interfaces:**
- Consumes: Tasks 1–4 的测试输出、端口清单和浏览器截图。
- Produces: 明确记录哪些已由 Go 接管、哪些仍由 Node 提供、哪些阻塞在凭据或基础设施。

- [ ] **Step 1: Record exact local evidence**

  记录 Git SHA、Go 端口、前端构建结果、MySQL/Redis 状态和浏览器首页/API 验收结果；不记录密钥值。

- [ ] **Step 2: Run the final skeleton verification**

  Run: `go test ./...`；`npm --prefix frontend run build`；`powershell -File scripts/check-local-go.ps1`

  Expected: 命令输出和未完成项全部写入 baseline。

- [ ] **Step 3: Commit the baseline**

  `git add docs/migration/local-go-runtime-baseline.md docs/superpowers/specs/2026-10-10-yizhanchengming-xin-go-migration-design.md && git commit -m "docs: record local Go runtime baseline"`

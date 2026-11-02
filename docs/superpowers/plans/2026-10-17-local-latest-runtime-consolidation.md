# 本地最新版运行收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `http://localhost:4001` 只能启动由当前 `v88` 源码生成的 Go 二进制和前端资源，旧构建不能被静默复用或误加载。

**Architecture:** 保持现有 React 构建、Go embed 和 Go-only 启动模式。构建脚本生成一个受 `.gitignore` 排除的运行清单；启动脚本先执行干净的前端构建和 Go 嵌入构建，再核验清单与当前 Git SHA。静态入口 HTML 禁止缓存，运行信息 API 返回前端构建标识，便于浏览器拒绝陈旧工作台。

**Tech Stack:** PowerShell、Vite、Go `embed.FS`、Go `net/http`、Go 单元测试。

## Global Constraints

- 唯一维护分支是 `v88`；不创建新的长期分支、运行入口或 Node 应用运行时。
- 只改源码和正式脚本；`frontend/dist`、`backend/web/dist`、本地 `.exe`、日志、缓存、运行数据都由脚本生成且不进入 Git。
- 旧文件安全保留，但不得作为构建输入、启动候选或浏览器路由。
- 构建失败必须停止启动；禁止复用旧二进制或旧前端页面。
- 用户可见的批量工厂页面不得回退到包含“分镜资产”列的历史工作台。

---

### Task 1: 隔离本地产物并明确唯一构建输入

**Files:**
- Modify: `.gitignore`
- Modify: `scripts/build-go-local.ps1:1-80`
- Test: `scripts/build-go-local.test.ps1`（新增）

**Interfaces:**
- Produces: `.local-review/runtime-manifest.json`，包含 `gitSha`、`frontendEntry`、`builtAt`、`binaryPath`。
- Consumes: `frontend/dist/index.html`、当前 Git `v88` SHA。

- [ ] **Step 1: 写入失败测试，要求构建清单缺失或入口不是当前构建时失败**

```powershell
$manifest = Join-Path $TestDrive 'runtime-manifest.json'
$entry = '/assets/user-current.js'
$result = Test-LocalRuntimeManifest -ManifestPath $manifest -ExpectedSha 'abc123' -ExpectedEntry $entry
$result.Valid | Should -BeFalse
```

- [ ] **Step 2: 运行测试，确认因 `Test-LocalRuntimeManifest` 尚不存在而失败**

Run: `Invoke-Pester scripts/build-go-local.test.ps1 -Output Detailed`

Expected: FAIL，指出找不到 `Test-LocalRuntimeManifest`。

- [ ] **Step 3: 在构建脚本实现清理、复制和清单写入**

```powershell
$runtimeDir = Join-Path $repo '.local-review'
$manifestPath = Join-Path $runtimeDir 'runtime-manifest.json'
$entry = [regex]::Match((Get-Content "$frontendDist/index.html" -Raw), 'src="([^"]+\.js)"').Groups[1].Value
if ([string]::IsNullOrWhiteSpace($entry)) { throw 'Frontend entry script is missing.' }
@{ gitSha = $buildSHA; frontendEntry = $entry; builtAt = (Get-Date).ToUniversalTime().ToString('o'); binaryPath = $binaryPath } |
  ConvertTo-Json | Set-Content -LiteralPath $manifestPath -Encoding utf8
```

清理 `frontend/dist` 和 `backend/web/dist` 的历史页面资源后，只复制本次 Vite 输出；继续移除 `batch-rewrite`。

- [ ] **Step 4: 运行测试，确认清单核验通过**

Run: `Invoke-Pester scripts/build-go-local.test.ps1 -Output Detailed`

Expected: PASS。

- [ ] **Step 5: 提交本任务的源码与测试**

```powershell
git add .gitignore scripts/build-go-local.ps1 scripts/build-go-local.test.ps1
git commit -m "fix: isolate local runtime build outputs"
```

### Task 2: 阻止过期构建启动，并暴露一致的运行信息

**Files:**
- Modify: `scripts/start-local-go.ps1:26-115`
- Modify: `backend/internal/httpapi/router.go:201-210`
- Modify: `backend/internal/httpapi/runtime_build_info_test.go`
- Test: `scripts/start-local-go.test.ps1`（新增）

**Interfaces:**
- Consumes: `.local-review/runtime-manifest.json`。
- Produces: `/api/runtime-build-info` 新字段 `frontend_entry` 和 `runtime_manifest_sha`。
- Produces: 启动器在清单 SHA、入口或二进制不存在/不一致时先调用 `build-go-local.ps1`，构建失败则退出。

- [ ] **Step 1: 写入失败测试，要求 API 返回前端入口和运行清单 SHA**

```go
if got := decoded["frontend_entry"]; got != "/assets/user-current.js" {
    t.Fatalf("frontend_entry = %v", got)
}
if got := decoded["runtime_manifest_sha"]; got != "abc123" {
    t.Fatalf("runtime_manifest_sha = %v", got)
}
```

- [ ] **Step 2: 运行 Go 测试，确认字段不存在导致失败**

Run: `go test ./internal/httpapi -run TestRuntimeBuildInfo -count=1`

Expected: FAIL，`frontend_entry` 或 `runtime_manifest_sha` 为空。

- [ ] **Step 3: 实现运行清单读取和启动核验**

```go
type RuntimeBuildInfo struct {
    GitSHA string `json:"gitSha"`
    FrontendEntry string `json:"frontendEntry"`
}
```

路由读取由启动器设置的 `QIANTIE_RUNTIME_MANIFEST_PATH`，只公开 `frontendEntry` 和 `gitSha`，不公开本地路径或凭证。启动器读取同一清单，若 `gitSha` 不等于当前 HEAD、入口脚本不在 `frontend/dist`、或本地二进制不存在，则运行 `build-go-local.ps1`；构建失败直接抛错。

- [ ] **Step 4: 运行 Go 和 PowerShell 测试**

Run: `go test ./internal/httpapi -run TestRuntimeBuildInfo -count=1; Invoke-Pester scripts/start-local-go.test.ps1 -Output Detailed`

Expected: PASS。

- [ ] **Step 5: 提交本任务的源码与测试**

```powershell
git add scripts/start-local-go.ps1 scripts/start-local-go.test.ps1 backend/internal/httpapi/router.go backend/internal/httpapi/runtime_build_info_test.go
git commit -m "fix: block stale local runtime startup"
```

### Task 3: 禁止浏览器继续使用历史入口页面

**Files:**
- Modify: `backend/internal/web/static.go`
- Modify: `backend/internal/web/static_test.go` 或现有静态处理测试文件
- Test: Go 静态服务测试

**Interfaces:**
- Consumes: 嵌入的 `index.html` 和哈希资源。
- Produces: `index.html`、`admin.html` 与 SPA 回退响应携带 `Cache-Control: no-cache, no-store, must-revalidate`；哈希资产仍可长期缓存。
- Produces: `/batch-rewrite/v2` 继续返回 `404`。

- [ ] **Step 1: 写入失败测试，要求入口 HTML 不可缓存且历史路由为 404**

```go
rec := httptest.NewRecorder()
handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/shuihuo-production", nil))
if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
    t.Fatalf("Cache-Control = %q", got)
}
```

- [ ] **Step 2: 运行测试，确认当前入口 HTML 没有该缓存策略时失败**

Run: `go test ./internal/web -run Test -count=1`

Expected: FAIL，缓存头为空或不同。

- [ ] **Step 3: 实现按资源类型设置缓存头**

```go
if name == "index.html" || name == "admin.html" || strings.HasSuffix(req.URL.Path, "/") {
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}
```

只对入口 HTML 与 SPA fallback 使用禁止缓存；`/assets/<hash>` 不改变现有哈希资源策略。

- [ ] **Step 4: 运行静态服务测试**

Run: `go test ./internal/web -run Test -count=1`

Expected: PASS。

- [ ] **Step 5: 提交本任务的源码与测试**

```powershell
git add backend/internal/web/static.go backend/internal/web/static_test.go
git commit -m "fix: prevent stale workbench entry caching"
```

### Task 4: 用当前源码重新构建、启动并做端到端验证

**Files:**
- Modify: 无产品源码；只生成已忽略运行产物。

**Interfaces:**
- Consumes: 任务 1-3 的构建、启动、静态资源实现。
- Produces: 4001 上经过验证的最新本地 Go 运行时。

- [ ] **Step 1: 构建最新前端和 Go 嵌入二进制**

Run: `npm.cmd run build; .\scripts\build-go-local.ps1`

Expected: 新生成的 `frontend/dist`、`backend/web/dist` 与 `.local-review/runtime-manifest.json` 使用同一入口脚本。

- [ ] **Step 2: 启动 Go-only review 运行时**

Run: `.\scripts\start-local-go-review.ps1 -UseYcmMysql -EnableBatchProduction -EnableLocalMerge`

Expected: 4001 监听，未启动 Node 应用服务。

- [ ] **Step 3: 核验 API、入口脚本和历史路由**

Run:

```powershell
Invoke-WebRequest http://localhost:4001/api/runtime-build-info -UseBasicParsing
Invoke-WebRequest http://localhost:4001/shuihuo-production -UseBasicParsing
Invoke-WebRequest http://localhost:4001/batch-rewrite/v2 -UseBasicParsing
```

Expected: build-info SHA 等于当前 `v88` HEAD；入口脚本等于清单；`batch-rewrite/v2` 为 404。

- [ ] **Step 4: 浏览器核验批量工厂表头**

打开 `http://localhost:4001/shuihuo-production`，进入用户实际批量工厂，确认：

```text
序号｜小说正文｜单书配置｜分镜提示词｜分镜视频｜操作
```

不得出现“分镜资产”。

- [ ] **Step 5: 提交运行收口相关源码改动**

```powershell
git add .gitignore scripts backend/internal/httpapi backend/internal/web
git commit -m "fix: unify local latest runtime"
```

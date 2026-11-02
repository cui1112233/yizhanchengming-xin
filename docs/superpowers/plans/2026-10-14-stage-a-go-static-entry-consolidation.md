# 阶段 A：Go 静态入口与构建产物收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 使本地 V88 的 Go 服务成为 React 静态页面的唯一运行入口，并将 `frontend/dist` 到 `backend/web/dist` 的派生关系固定、可验证、可重建，同时不修改小说获取和批量工厂链路。

**Architecture:** React/Vite 继续只在构建期生成 `frontend/dist`；`scripts/build-go-local.ps1` 清空并复制该构建结果及少量既有兼容静态覆盖物到 `backend/web/dist`；Go 通过 `go:embed` 提供所有页面和资产。本阶段只收口本地 Go 静态入口和构建链，不把生产 Nginx 从 Node 切到 Go，因为现有 Node 仍承载未迁移 API、视频管理系统 Worker 和 `shuihuo-compat` 路由。

**Tech Stack:** React 18、Vite 6、Node（仅构建期）、PowerShell 5、Go `embed.FS`、Go `net/http`、Go `testing`。

## Global Constraints

- 维护主线是 `v88`；所有变更先写入 Git 工作区，禁止直接修改 ECS 或公网。
- Node 仅可用于 React/Vite 与既有静态兼容页的构建；本阶段不启动 Node 作为本地应用运行时。
- 不修改小说获取、改文、批量工厂的 API、调度、任务状态或业务语义。
- 不修改 MySQL、Redis、FFmpeg、TOS、Docker 网络和 ECS 公网配置。
- `frontend/dist` 是 React/Vite 的唯一成品源；`backend/web/dist` 是可删除并重建的 Go embed 派生产物，禁止手工修补。
- `public/novel-panel`、`frontend/public/batch-rewrite`、`public/batch-rewrite`、`public/assets` 与 `pets` 是现存兼容静态覆盖物；本阶段只能原样复制，不得修改其业务内容。
- 不提交 `.env.v88-review`、`.local-review/`、`frontend/dist/downloads/`、日志、临时构建目录或用户已有无关未提交改动。
- 不删除 `app.js`、`server.js`、`routes/`、Node Browser Worker、`shuihuo-compat`、生产 Compose/Nginx 配置；它们属于后续阶段 D/E 的迁移对象。
- 不创建 Git commit，除非用户在实施完成后明确要求提交。

---

## 文件结构与职责

| 文件 | 本阶段职责 |
|---|---|
| `scripts/build-go-local.ps1` | 从 Vite 成品和既有静态覆盖物确定性生成 `backend/web/dist`，然后按需要编译 Go 二进制。 |
| `scripts/build-go-local.test.ps1`（新建） | 在不读取密钥、不启动服务的前提下，验证构建脚本的源、清理、复制、覆盖和可选 Go 编译契约。 |
| `scripts/start-local-go.ps1` | 检测 Go 源或静态输入更新后自动调用构建脚本，启动本地 Go 服务。 |
| `scripts/start-local-go.test.ps1`（新建） | 验证本地 Go 启动脚本仅使用 Go 二进制、包含所有静态输入根并在变更后重建。 |
| `scripts/check-local-go.ps1` | 验证 Go 本地服务的健康、SPA 页面和关键静态文件响应。 |
| `scripts/check-local-go.test.ps1`（新建） | 静态验证 Go 运行检查覆盖健康、用户端、管理端和 Vite 资产检查，不包含 Node 启动命令。 |
| `backend/internal/web/static.go` | 按 Go embed 文件系统服务根页、SPA 路由、管理端和既有嵌套工作台。 |
| `backend/internal/web/static_test.go` | 覆盖 Go 静态服务路由、资产 MIME、SPA 回退和缺失文件行为。 |
| `backend/web/assets.go` | 仅公开 `backend/web/dist` 的嵌入文件系统，不能访问 `frontend/dist`。 |
| `docs/README-local.md` | 记录唯一推荐的本地应用启动方式为 Go；Node 启动脚本只标为兼容排查工具。 |

## 不在本阶段修改的文件

- `deploy/v88-public/docker-compose.yml`
- `deploy/v88-public/nginx.conf`
- `app.js`、`server.js`、`routes/**`、`lib/**`
- `services/121-browser-worker/**`
- `local-executor/**`
- 任何小说获取、改文、批量工厂的 API/任务/调度实现

这些文件仍是当前混合运行时的一部分。将 Nginx 直接转至 Go 必须等待 Node 路由、Browser Worker 与 `shuihuo-compat` 都完成替代，不能在阶段 A 提前实施。

---

### Task 1: 为 Go 静态服务补齐入口与资产契约测试

**Files:**
- Modify: `backend/internal/web/static_test.go`
- Modify: `backend/internal/web/static.go`（仅在测试暴露实际缺口时）

**Interfaces:**
- Consumes: `func StaticHandler(files fs.FS) http.Handler`
- Produces: 对 `/`、用户 SPA 路由、`/admin/*`、`/assets/*`、`/pets/*`、`/novel-panel/workbench`、`/batch-rewrite/v2` 的稳定 HTTP 行为。

- [ ] **Step 1: 写入失败测试，覆盖静态资产类型与 SPA 回退边界**

在 `TestStaticHandlerServesIndexAndAssets` 的 MapFS 中增加一个带明确内容的哈希 CSS 文件，并加入下列断言：

```go
{
    path:            "/assets/app.js",
    want:            http.StatusOK,
    wantContentType: "text/javascript",
    wantBody:        "console.log('ok')",
},
{
    path:            "/assets/style.css",
    want:            http.StatusOK,
    wantContentType: "text/css",
    wantBody:        "body{}",
},
{
    path:            "/settings",
    want:            http.StatusOK,
    wantContentType: "text/html",
    wantBody:        "<html>home</html>",
},
{
    path: "/missing.js",
    want: http.StatusNotFound,
},
```

将表驱动结构扩展为包含 `wantContentType` 与 `wantBody`。仅对非空预期字段断言 `strings.Contains(res.Header().Get("Content-Type"), wantContentType)` 和 `strings.Contains(res.Body.String(), wantBody)`。

- [ ] **Step 2: 运行测试，确认当前行为是否缺失**

Run:

```powershell
Push-Location backend
go test ./internal/web -run 'TestStaticHandlerServesIndexAndAssets|TestStaticHandlerServesNestedWorkbenchAndLegacyV2Alias' -count=1
Pop-Location
```

Expected: 若现有 `http.FileServer` 已提供正确 MIME，测试直接 PASS；若任一断言失败，记录失败的路径与响应头，不修改其他业务路由。

- [ ] **Step 3: 仅在失败时最小化修复静态处理器**

若 CSS/JS 的 Content-Type 未正确返回，在 `StaticHandler` 内保留现有 `http.FileServer(http.FS(files))`，不要自行手写 MIME 映射；仅修正导致 `FileServer` 未被调用的路由条件。SPA 路由只能在“不存在扩展名”的路径上回退到 `index.html`，不得让缺失 `/assets/*.js`、`/assets/*.css` 回退到 HTML。

修复后应满足以下不变量：

```go
if err != nil && path.Ext(name) == "" {
    // only SPA paths may fall back to index.html
}
```

- [ ] **Step 4: 运行静态服务全量测试**

Run:

```powershell
Push-Location backend
go test ./internal/web -count=1
Pop-Location
```

Expected: PASS，且现有 `/batch-rewrite/v2`、`/admin/*`、嵌套工作台与 nil 文件系统测试继续通过。

- [ ] **Step 5: 检查变更范围**

Run:

```powershell
git diff -- backend/internal/web/static.go backend/internal/web/static_test.go
git status --short
```

Expected: 仅显示该任务新增的静态服务测试与必要的最小处理器改动；不得出现 `frontend/dist` 或无关生成文件。

---

### Task 2: 将 Go embed 产物生成固定为确定性、可检查的派生步骤

**Files:**
- Modify: `scripts/build-go-local.ps1`
- Create: `scripts/build-go-local.test.ps1`

**Interfaces:**
- Consumes: `frontend/dist/index.html`、`frontend/public/batch-rewrite/index.html`、`public/batch-rewrite/v78-novel-fetch-v2-runtime.js`、`public/novel-panel/workbench/index.html`、`public/assets/brand-logo-black.png`、`public/assets/brand-logo-white.png`、`pets/stacky/spritesheet.webp`。
- Produces: 完整替换后的 `backend/web/dist`；当未传 `-SkipGoBuild` 时生成 `backend/qiantie-local.exe`。
- Adds: `-CheckOnly` 开关，只验证源输入与现有 `backend/web/dist` 派生产物的必要入口，不写文件、不调用 Node、不执行 `go build`。

- [ ] **Step 1: 写入构建脚本契约测试**

创建 `scripts/build-go-local.test.ps1`，读取 `build-go-local.ps1` 原文，验证以下精确契约：

```powershell
$script = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'build-go-local.ps1') -Raw
$required = @(
  '[switch]$SkipGoBuild',
  '[switch]$CheckOnly',
  "Join-Path $repo 'frontend/dist'",
  "Join-Path $repo 'backend/web/dist'",
  "Join-Path $repo 'frontend/public/batch-rewrite'",
  "Join-Path $repo 'public/batch-rewrite'",
  "Join-Path $repo 'public/novel-panel'",
  "Join-Path $repo 'public/assets'",
  "Join-Path $repo 'pets'",
  "Get-ChildItem -LiteralPath $goDist -Force",
  'Copy-Item -Path (Join-Path $frontendDist',
  'go build -o qiantie-local.exe ./cmd/qiantie'
)
foreach ($needle in $required) {
  if (-not $script.Contains($needle)) { throw "missing build contract: $needle" }
}
if ($script.Contains('node server.js')) { throw 'build script must not start Node runtime' }
Write-Output 'Go embedded asset build contract passed.'
```

- [ ] **Step 2: 运行新测试，确认 `-CheckOnly` 契约先失败**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.test.ps1
```

Expected: FAIL，错误包含 `missing build contract: [switch]$CheckOnly`。

- [ ] **Step 3: 为构建脚本加入只读检查模式与显式输入校验**

在 `param` 中改为：

```powershell
param(
  [switch]$SkipGoBuild,
  [switch]$CheckOnly
)
```

将当前各个 `Test-Path` 前置检查保留，并提取为同一条执行路径。完成所有输入检查后，在第一次 `New-Item` 之前加入：

```powershell
if ($CheckOnly) {
  $requiredOutputs = @(
    (Join-Path $goDist 'index.html'),
    (Join-Path $goDist 'admin.html'),
    (Join-Path $goDist 'batch-rewrite/index.html'),
    (Join-Path $goDist 'novel-panel/workbench/index.html'),
    (Join-Path $goDist 'pets/stacky/spritesheet.webp'),
    (Join-Path $goDist 'assets/brand-logo-black.png'),
    (Join-Path $goDist 'assets/brand-logo-white.png')
  )
  $missing = @($requiredOutputs | Where-Object { -not (Test-Path -LiteralPath $_) })
  if ($missing.Count -gt 0) {
    throw ('Go embedded asset output is incomplete: ' + ($missing -join ', '))
  }
  Write-Output "Go embedded asset inputs and outputs are ready: $goDist"
  exit 0
}
```

不要在 `-CheckOnly` 中比较文件时间、哈希或内容，因为 Vite hash 文件名会变化；本阶段只校验受 Go embed 依赖的稳定入口和必要覆盖资源。

- [ ] **Step 4: 运行构建脚本测试与只读检查**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1 -CheckOnly
```

Expected: 两条命令均退出码 0；第二条不生成或删除任何文件，也不启动 Node、Go 服务。

- [ ] **Step 5: 用临时构建验证清理与重建，不提交生成结果**

Run:

```powershell
npm --prefix frontend run build
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1 -SkipGoBuild
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1 -CheckOnly
```

Expected: `backend/web/dist` 被完整刷新，`index.html`、`admin.html`、兼容工作台、宠物资源和品牌图片均存在；不执行 `go build`。检查时不得把 `frontend/dist` 或 `backend/web/dist` 的既有生成变化混入后续提交范围。

- [ ] **Step 6: 检查变更范围**

Run:

```powershell
git diff -- scripts/build-go-local.ps1 scripts/build-go-local.test.ps1
git status --short
```

Expected: 只审阅脚本和测试源码；忽略本机已有或本次构建生成的 `frontend/dist`、`backend/web/dist` 变动。

---

### Task 3: 让本地 Go 启动器以静态输入与 embed 输出共同决定重建

**Files:**
- Modify: `scripts/start-local-go.ps1`
- Create: `scripts/start-local-go.test.ps1`

**Interfaces:**
- Consumes: 当前 Go 二进制、Go 源码、阶段 A 所有静态输入根及 `backend/web/dist`。
- Produces: 当任一输入或 embed 输出缺失/过期时，调用 `scripts/build-go-local.ps1`；随后仅执行 `backend/qiantie-local.exe`。

- [ ] **Step 1: 写入启动脚本契约测试**

创建 `scripts/start-local-go.test.ps1`：

```powershell
$ErrorActionPreference = 'Stop'
$script = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'start-local-go.ps1') -Raw
$required = @(
  "Join-Path $repo 'backend/qiantie-local.exe'",
  "Join-Path $repo 'frontend/dist'",
  "Join-Path $repo 'frontend/public/batch-rewrite'",
  "Join-Path $repo 'public/batch-rewrite'",
  "Join-Path $repo 'public/novel-panel'",
  "Join-Path $repo 'public/assets'",
  "Join-Path $repo 'pets'",
  "Join-Path $repo 'backend/web/dist'",
  "build-go-local.ps1",
  '& .\qiantie-local.exe'
)
foreach ($needle in $required) {
  if (-not $script.Contains($needle)) { throw "missing local Go launch contract: $needle" }
}
if ($script.Contains('node server.js')) { throw 'local Go launcher must not start Node runtime' }
Write-Output 'Local Go launcher contract passed.'
```

- [ ] **Step 2: 运行测试，确认 embed 输出根目前未被检测**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.test.ps1
```

Expected: FAIL，错误包含 `missing local Go launch contract: Join-Path $repo 'backend/web/dist'`。

- [ ] **Step 3: 在启动脚本中加入 `backend/web/dist` 完整性条件**

在现有 `$assetRoots` 检查之前定义：

```powershell
$goDist = Join-Path $repo 'backend/web/dist'
$requiredEmbeddedOutputs = @(
  (Join-Path $goDist 'index.html'),
  (Join-Path $goDist 'admin.html'),
  (Join-Path $goDist 'batch-rewrite/index.html'),
  (Join-Path $goDist 'novel-panel/workbench/index.html'),
  (Join-Path $goDist 'pets/stacky/spritesheet.webp'),
  (Join-Path $goDist 'assets/brand-logo-black.png'),
  (Join-Path $goDist 'assets/brand-logo-white.png')
)
```

将静态输入列表补齐为：

```powershell
$assetRoots = @(
  (Join-Path $repo 'frontend/dist'),
  (Join-Path $repo 'frontend/public/batch-rewrite'),
  (Join-Path $repo 'public/novel-panel'),
  (Join-Path $repo 'public/batch-rewrite'),
  (Join-Path $repo 'public/assets'),
  (Join-Path $repo 'pets')
) | Where-Object { Test-Path $_ }
```

在计算 `$latestAsset` 前加入：

```powershell
if (@($requiredEmbeddedOutputs | Where-Object { -not (Test-Path -LiteralPath $_) }).Count -gt 0) {
  $needsBuild = $true
}
```

保留现有“任一静态输入比 Go 二进制新则重建”的行为。不要通过启动 Node 修复缺失页面；唯一修复路径必须是 `build-go-local.ps1`。

- [ ] **Step 4: 运行启动脚本静态测试和 Go 预检查**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.ps1 -CheckOnly
```

Expected: 静态测试 PASS；`-CheckOnly` 继续委托 `check-local-go.ps1`，不启动 Node。

- [ ] **Step 5: 检查变更范围**

Run:

```powershell
git diff -- scripts/start-local-go.ps1 scripts/start-local-go.test.ps1
git status --short
```

Expected: 不包含 `scripts/start-local-node.ps1`；该脚本只在后续 Node 退役阶段处理。

---

### Task 4: 扩展 Go-only 本地运行检查，覆盖页面与 embed 资产

**Files:**
- Modify: `scripts/check-local-go.ps1`
- Create: `scripts/check-local-go.test.ps1`

**Interfaces:**
- Consumes: `QIANTIE_GO_BASE_URL`（默认 `http://127.0.0.1:4001`）和可选 `QIANTIE_EXPECTED_GO_SHA`。
- Produces: 对 Go 服务的健康、运行身份、SPA 页面、管理页、嵌入资产和兼容工作台 HTTP 验收。

- [ ] **Step 1: 写入运行检查脚本契约测试**

创建 `scripts/check-local-go.test.ps1`：

```powershell
$ErrorActionPreference = 'Stop'
$script = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'check-local-go.ps1') -Raw
$required = @(
  "Check-Url '/health' 200",
  "Check-Url '/' 200",
  "Check-Url '/admin' 200",
  "Check-Url '/settings' 200",
  "Check-Url '/assets/'",
  "Check-Url '/novel-panel/workbench' 200",
  "Check-Url '/batch-rewrite/v2' 200",
  "'/api/runtime-build-info'"
)
foreach ($needle in $required) {
  if (-not $script.Contains($needle)) { throw "missing Go runtime check: $needle" }
}
if ($script.Contains('3001') -or $script.Contains('node ')) { throw 'Go runtime checker must not require Node runtime' }
Write-Output 'Local Go runtime checker contract passed.'
```

- [ ] **Step 2: 运行测试，确认资产探测机制尚未存在**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.test.ps1
```

Expected: FAIL，错误包含 `missing Go runtime check: Check-Url '/admin' 200`。

- [ ] **Step 3: 为已知 Vite 资产增加动态探测与 HTTP 校验**

在 `check-local-go.ps1` 中保留已有 `Check-Url`，增加一个仅接受成功响应的辅助函数：

```powershell
function Check-AssetFromHtml([string]$PagePath) {
  try {
    $page = Invoke-WebRequest -Uri ($BaseUrl + $PagePath) -UseBasicParsing -TimeoutSec 5
    $match = [regex]::Match($page.Content, '(/assets/[^"''<> ]+\.(js|css))')
    if (-not $match.Success) {
      Write-Output "FAIL $PagePath did not reference an embedded Vite asset"
      $script:failed = $true
      return
    }
    Check-Url $match.Groups[1].Value 200
  } catch {
    Write-Output "FAIL ${PagePath}: $($_.Exception.Message)"
    $script:failed = $true
  }
}
```

在现有页面检查后增加：

```powershell
Check-Url '/admin' 200
Check-Url '/novel-panel/workbench' 200
Check-Url '/batch-rewrite/v2' 200
Check-AssetFromHtml '/'
Check-AssetFromHtml '/admin'
```

不要写死带 hash 的具体文件名。不得向 `/api/*` 发出写请求，不得执行登录、外部模型调用或发布任务。

- [ ] **Step 4: 运行脚本测试和已启动本地 Go 服务验收**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1 -BaseUrl http://127.0.0.1:4001
```

Expected: 第一条 PASS；第二条在已有 Go 服务运行且数据库配置正常时输出各页面、资产、健康接口和 runtime build info 的 PASS。若本机服务未启动，记录连接失败，不通过启动 Node 绕过。

- [ ] **Step 5: 运行 Go 静态服务与启动脚本相关测试**

Run:

```powershell
Push-Location backend
go test ./internal/web ./cmd/qiantie -count=1
Pop-Location
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.test.ps1
```

Expected: 全部 PASS。

---

### Task 5: 固化本地入口说明，并进行阶段 A 回归审查

**Files:**
- Modify: `docs/README-local.md`
- Modify: `docs/migration/local-go-runtime-baseline.md`

**Interfaces:**
- Consumes: 阶段 A 产生的构建、启动和检查命令。
- Produces: 面向开发者的唯一推荐本地应用入口：`build-go-local.ps1` → `start-local-go.ps1` → `check-local-go.ps1`。

- [ ] **Step 1: 更新本地运行说明，不删除兼容脚本**

在 `docs/README-local.md` 中将本地操作顺序明确为：

```text
1. npm --prefix frontend run build
2. powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
3. powershell -ExecutionPolicy Bypass -File scripts/start-local-go.ps1
4. powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1
```

在同一文档增加一句：

```text
`scripts/start-local-node.ps1` 仅用于未迁移 Node 路由的兼容排查，不是推荐的本地应用启动方式；不得与 Go-only 验收混用。
```

不得宣称 Node 已完全删除，也不得要求删除 Node 项目文件。

- [ ] **Step 2: 更新迁移基线，记录阶段 A 的可验证边界**

在 `docs/migration/local-go-runtime-baseline.md` 增加“阶段 A 收口后”小节，列出：

```text
- frontend/dist 是 Vite 构建成品源；
- backend/web/dist 是 Go embed 派生产物；
- Go 本地服务提供 /、/admin、SPA 路由、/assets 和既有嵌入工作台；
- 本地 Go 验收不启动 Node；
- 生产 Nginx 仍暂时转发 Node，原因是其余 Node/兼容服务尚未迁移。
```

- [ ] **Step 3: 构建 Go 二进制并执行完整本地验收**

Run:

```powershell
npm --prefix frontend run build
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1 -CheckOnly
Push-Location backend
go test ./internal/web ./cmd/qiantie -count=1
Pop-Location
powershell -ExecutionPolicy Bypass -File scripts/start-local-go.test.ps1
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.test.ps1
```

Expected: 构建和测试全部 PASS；若要进行 HTTP 验收，使用已批准的本地 MySQL/Redis 配置启动 Go，再运行 `check-local-go.ps1`。不得启动 Node 作为替代运行时。

- [ ] **Step 4: 进行生成物与未提交文件安全检查**

Run:

```powershell
git diff --check
git status --short
git diff -- scripts/build-go-local.ps1 scripts/build-go-local.test.ps1 scripts/start-local-go.ps1 scripts/start-local-go.test.ps1 scripts/check-local-go.ps1 scripts/check-local-go.test.ps1 backend/internal/web/static.go backend/internal/web/static_test.go docs/README-local.md docs/migration/local-go-runtime-baseline.md
```

Expected: 不包含 `.env.v88-review`、`.local-review/`、`frontend/dist/downloads/`、日志、临时目录、`frontend/dist` 的 hash 资源或无关用户改动；不执行 `git clean`、`git reset`、删除历史分支或删除 Docker 资源。

- [ ] **Step 5: 准备提交内容但不执行提交**

列出建议提交范围：

```text
scripts/build-go-local.ps1
scripts/build-go-local.test.ps1
scripts/start-local-go.ps1
scripts/start-local-go.test.ps1
scripts/check-local-go.ps1
scripts/check-local-go.test.ps1
backend/internal/web/static.go（仅在 Task 1 必要时）
backend/internal/web/static_test.go
docs/README-local.md
docs/migration/local-go-runtime-baseline.md
```

建议提交信息：

```text
feat: consolidate local Go static entry
```

此步骤只供用户审核；未得到用户明确“提交”指令时，不运行 `git add` 或 `git commit`。

# Root Directory Reorganization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将仓库根目录按业务职责收敛为 Go 后端、React 前端、外部服务、执行器、部署、脚本、文档、测试和运行数据九类目录，并删除根目录无调用方的 Node 依赖与旧产物。

**Architecture:** 先建立目标目录和迁移映射，再逐项更新构建/测试/部署引用；每个阶段都保持旧路径可验证，最后再删除旧副本。Go 的最终静态输入仍统一生成到 `backend/web/dist`，页面路由不变。

**Tech Stack:** Go 1.23、React/Vite、PowerShell、Docker Compose（仅保留审计/Worker配置）、Git。

**Spec:** `docs/superpowers/specs/2026-10-15-root-directory-reorganization-design.md`

## Global Constraints

- 接口后端必须继续使用 Go，代码位于 `backend`。
- 前端代码位于 `frontend`，Go `embed` 只发布构建后的静态资源。
- 不删除 `services/121-browser-worker`、运行数据、历史审计资料和备份 tag。
- 不改变 `/batch-factory`、`/novel-fetch`、`/novel-panel` 路由和 API 合约。
- 不提交或上传真实 `.env.v88-review`，不推送远程仓库。
- 所有移动先更新引用，再删除旧路径。

## Review Focus

- Go 构建脚本仍能找到兼容静态源并生成 `backend/web/dist`。
- `services/121-browser-worker` 的 Docker 构建上下文和启动路径不变。
- Go embed 中的宠物、品牌、小说面板和批量改文资源仍可访问。
- 执行器下载、部署清单和测试路径在目录移动后仍一致。
- 根目录不再生成 `node_modules`、临时日志、临时媒体和本地密钥文件。

---

### Task 1: 建立迁移映射与路径回归测试

**Files:**
- Create: `tests/root-directory-layout.test.js`
- Modify: `docs/superpowers/specs/2026-10-15-root-directory-reorganization-design.md`（仅在发现约束冲突时）

**Interfaces:**
- Produces: 明确允许的根目录清单、旧路径禁用清单，供后续迁移任务和最终验收使用。

- [ ] **Step 1: 编写失败测试**：检查根目录允许项、禁止根 `node_modules`、检查 `backend/web/dist` 和 `services/121-browser-worker` 存在。
- [ ] **Step 2: 运行测试确认当前因布局不符合而失败。**
- [ ] **Step 3: 固化迁移映射表**：为 `local-executor`、`giant-material-executor`、静态资源和提示词定义目标路径，并把例外（`.github`、`deploy`、`data`）写入测试说明。
- [ ] **Step 4: 暂不删除任何业务文件，提交测试和映射说明。**

### Task 2: 整理执行器目录

**Files:**
- Move: `local-executor/` -> `executors/local-executor/`
- Move: `giant-material-executor/` -> `executors/giant-material-executor/`
- Modify: `scripts/`、`deploy/`、`tests/`、`test/` 中引用上述路径的文件
- Modify: `docs/` 中仍作为当前入口的路径说明

**Interfaces:**
- Consumes: Task 1 的迁移映射。
- Produces: 执行器统一入口路径，外部 API 和下载 URL 不变。

- [ ] **Step 1: 添加路径回归测试**：构建脚本、部署清单和下载测试只能引用 `executors/` 新路径。
- [ ] **Step 2: 更新脚本、清单、测试和文档引用。**
- [ ] **Step 3: 移动两个执行器目录，并确认各自 `go.mod` / `package.json` / 构建入口存在。**
- [ ] **Step 4: 运行执行器相关 Go/Node 测试和路径测试。**

### Task 3: 收敛前端静态源和 Go embed 输入

**Files:**
- Move/merge: 根 `public/`、`pets/`、`branding/`、根 `dist/` 中仍有效的资源 -> `frontend/public/` 对应业务子目录
- Modify: `scripts/build-go-local.ps1`
- Modify: `scripts/start-local-go.ps1`
- Modify: `backend/web/embed.go` 或当前静态构建入口（以实际文件为准）
- Modify: 静态资源路径测试和构建测试

**Interfaces:**
- Consumes: Task 1 映射；保留 `frontend/public/batch-rewrite` 和 `public/novel-panel` 的内容语义。
- Produces: `backend/web/dist` 为唯一 Go embed 输入，页面 URL 不变。

- [ ] **Step 1: 为 `/`, `/batch-factory`, `/novel-fetch`, `/novel-panel/workbench`, `/pets/*`, `/assets/*` 增加静态访问回归测试。**
- [ ] **Step 2: 更新构建脚本，使其只从 `frontend` 业务源复制到临时构建区，再生成 `backend/web/dist`。**
- [ ] **Step 3: 合并资源并移动目录，确认无文件覆盖和重复来源。**
- [ ] **Step 4: 运行前端构建、Go 静态测试和本地 Go 启动检查。**

### Task 4: 清理根 Node 依赖与旧静态副本

**Files:**
- Modify: `package.json`
- Delete: `package-lock.json`（根目录不再有 Node 依赖时）
- Delete: 根 `node_modules/`
- Delete: 已确认无调用方的根 `index.html`、根 `dist/` 副本
- Modify: `.gitignore`

**Interfaces:**
- Consumes: Task 2/3 已完成的引用迁移。
- Produces: 根目录不再安装或运行 Node 主服务依赖；前端依赖只存在 `frontend`，Worker 依赖只由 Worker 自己的构建配置管理。

- [ ] **Step 1: 确认 `package.json` 中只剩 Go 启动/前端代理脚本；移除 Express、TOS SDK、代理库等无调用方依赖。**
- [ ] **Step 2: 重新生成或删除根锁文件，并验证 `services/121-browser-worker` 的 Dockerfile 自己安装依赖。**
- [ ] **Step 3: 删除根 `node_modules` 和已确认的旧静态副本。**
- [ ] **Step 4: 更新 `.gitignore`，禁止根 Node 依赖、临时产物和真实本地环境文件回流。**

### Task 5: 根目录验收和提交

**Files:**
- Modify: `tests/root-directory-layout.test.js`
- Modify: `docs/CURRENT_VERSION.md` 或迁移记录（记录最终布局和提交 SHA）

- [ ] **Step 1: 运行 `go test ./...`（backend）。**
- [ ] **Step 2: 运行前端构建并检查 `backend/web/dist` 必需文件。**
- [ ] **Step 3: 运行根目录布局、静态路由、Worker 构建引用和部署拓扑测试。**
- [ ] **Step 4: 检查 `git diff --check`、`git status`，确保不包含真实密钥和用户运行数据。**
- [ ] **Step 5: 只提交本次整理涉及的文件，记录提交 SHA；不推送远程。**

# 小说获取 V78 React + Go 全功能对标实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 `/novel-fetch` 的 React 页面在 UI、交互、状态、持久化和业务结果上完整对应 V78 参考页面，并确保后端链路统一走 Go。

**Architecture:** 保留现有 Go 小说获取领域服务和 API 合约，先补齐状态/配置/任务数据，再用 React 重建 V78 工作台布局。React 只负责展示和触发窄 API；提示词、敏感词、分类、任务执行和网站提交继续由 Go 负责。旧 `frontend/public/batch-rewrite` 只用于对照，最终不作为运行入口。

**Tech Stack:** Go、Gin、MySQL、Redis、React、Ant Design、Vite、Go `embed`。

**Spec:** `docs/superpowers/specs/2026-10-16-novel-fetch-v78-react-go-parity-design.md`

## Global Constraints

- `/novel-fetch` 必须加载 React 源码编译结果，不允许 iframe 或直接加载 `frontend/public/batch-rewrite`。
- 业务接口和任务执行统一走 Go；Node 文件不得成为运行时依赖。
- 系统提示词由 Go prompt 模块提供，React 不硬编码业务提示词。
- 用户手工填写的平台、书名、风格和男女频优先级最高；缺失时按 121 元数据判断，再由 AI 兜底。
- 必须保留原文、任务日期、处理状态、失败原因、重试次数和提交回执。
- 不向公网部署；任何公网发布前必须再次得到用户确认。
- 不覆盖工作区现有未提交改动，不删除旧参考文件，直到引用和测试证明可以清理。

## Review Focus

- 页面刷新后选中的版本、改文方式和敏感词开关仍保持：由任务 1 的持久化测试覆盖。
- 原文抓取成功但分类字段缺失：由任务 2 的 121 元数据映射和 AI 兜底测试覆盖。
- 敏感词 AI 开启/关闭必须走不同真实处理路径：由任务 2 的处理流程测试覆盖。
- 任务跨日期时不能被当前日期筛选误隐藏：由任务 3 的日期和当前批次测试覆盖。
- API 返回错误或任务失败时页面不能显示“成功”：由任务 3、任务 4 的错误状态测试覆盖。

### Task 1: 固化 React 页面状态和 V78 输入配置

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Modify: `frontend/src/shared/api/novelFetch.js`
- Test: `frontend/src/user/pages/NovelFetchPage.source.test.js`
- Test: `frontend/src/user/pages/novel-fetch-v78-state.test.js`

**Interfaces:**
- Consumes: `fetchNovelContent`, `getWorkshopConfig`, `saveWorkshopConfig`, `listWorkshopTasks`。
- Produces: 可序列化的处理表单状态，包括 `platform`、`textModel`、`selectedVersions`、`rewriteMethods`、`sensitiveAiProcessEnabled` 和 `maxTxt`。

- [ ] **Step 1: 写失败测试**：验证 V78 默认选中原文和 AI1、每个 AI 改文方式可保存、敏感词开关可恢复、切换标签页不会丢失状态。
- [ ] **Step 2: 运行测试确认失败**：运行 `node --test frontend/src/user/pages/novel-fetch-v78-state.test.js frontend/src/user/pages/NovelFetchPage.source.test.js`，预期新断言失败。
- [ ] **Step 3: 实现状态持久化和输入映射**：将表单状态集中到明确的 React state/持久化函数，提交时生成后端已有 API 所需的精确 payload；不把业务提示词放入组件。
- [ ] **Step 4: 运行测试确认通过**：同上命令，预期全部通过。
- [ ] **Step 5: 提交**：`git add frontend/src/user/pages frontend/src/shared/api && git commit -m "web: 固化V78处理配置状态——切换页面不再丢失用户选择"`。

### Task 2: 校准 Go 原文、分类、敏感词和 AI 处理链路

**Files:**
- Modify: `backend/internal/novelfetchworkshop/processing.go`
- Modify: `backend/internal/novelfetchworkshop/gender.go`
- Modify: `backend/internal/novelfetchworkshop/sensitive.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Test: `backend/internal/novelfetchworkshop/gender_test.go`
- Test: `backend/internal/novelfetchworkshop/processing_test.go`
- Test: `test/novel-fetch-sensitive-fallback.test.js`

**Interfaces:**
- Consumes: React 提交的处理 payload 和现有配置/任务存储。
- Produces: 统一任务状态、分类结果、原文/AI 版本结果、敏感词处理记录和错误上下文。

- [ ] **Step 1: 写失败测试**：覆盖“手工男女频优先 → 121 `category/genre` 映射 → AI 兜底”、完整原文与 `maxTxt` 分离保存、敏感词 AI 开关两条路径。
- [ ] **Step 2: 运行 Go 测试确认失败**：运行 `go test ./backend/internal/novelfetchworkshop/...`，预期新增断言失败。
- [ ] **Step 3: 实现 Go 链路**：在服务层统一编排抓取、分类、原文保存、敏感词处理和 AI 版本生成；每次外部调用包裹操作上下文错误，不在 handler 里重复业务逻辑。
- [ ] **Step 4: 运行 Go 和 Node 合约测试**：运行 `go test ./backend/internal/novelfetchworkshop/...` 与 `node --test test/novel-fetch-sensitive-fallback.test.js`。
- [ ] **Step 5: 提交**：`git add backend/internal test && git commit -m "api: 统一小说获取处理链路——分类与敏感词由Go一次编排"`。

### Task 3: 完成 V78 任务、进度、日志和失败状态数据契约

**Files:**
- Modify: `backend/internal/novelfetchworkshop/mysql_store.go`
- Modify: `backend/internal/novelfetchworkshop/lifecycle_store.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js`
- Test: `test/novel-fetch-batch-isolation.test.js`
- Test: `test/novel-fetch-ai-reconciliation.test.js`
- Test: `test/novel-fetch-workshop-process-routes.test.js`

**Interfaces:**
- Consumes: Go 任务和持久化记录。
- Produces: 统计卡、进度条、实时日志、当前批次、历史任务、失败原因、重试和删除所需的稳定字段。

- [ ] **Step 1: 写失败测试**：验证总数/已解析/处理中/已完成/失败统计、日期筛选、跨批次隔离、失败重试和错误原因返回。
- [ ] **Step 2: 运行现有任务测试确认失败或缺口**：运行上述三个 Node 测试并记录缺失字段。
- [ ] **Step 3: 补齐 Go 响应契约**：为任务列表和任务详情返回稳定状态字段、时间字段、计数、日志和失败信息；日期筛选使用真实任务日期。
- [ ] **Step 4: 更新前端 API 封装**：补齐刷新、详情、重试、删除、批次和日志请求，统一错误转换。
- [ ] **Step 5: 运行测试确认通过并提交**：执行相关 Node/Go 测试，提交 `api: 固化小说获取任务状态契约——页面显示真实进度和失败原因`。

### Task 4: React 重建 V78 工作台布局并接入真实状态

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/novel-fetch.css`
- Modify: `frontend/src/user/App.jsx`（仅在路由入口需要修正时）
- Test: `frontend/src/user/App.performance.test.js`
- Test: `frontend/src/user/pages/novel-fetch-responsive.layout.test.js`

**Interfaces:**
- Consumes: 任务 1 的处理状态和任务 3 的数据契约。
- Produces: 与 V78 对应的侧栏、顶部标题、标签页、输入区、右侧处理状态区、进度条、实时日志、操作栏和当前批次表格。

- [ ] **Step 1: 写失败布局测试**：验证 `/novel-fetch` 仍加载 React、存在 V78 页面关键区域和所有操作入口，不存在 iframe/旧静态入口。
- [ ] **Step 2: 运行测试确认失败**：运行 `node --test frontend/src/user/App.performance.test.js frontend/src/user/pages/novel-fetch-responsive.layout.test.js`。
- [ ] **Step 3: 实现 JSX 分区**：按 V78 结构拆分处理区、状态区、操作区、当前批次区和标签面板；按钮全部绑定真实 handler，不创建展示用假数据。
- [ ] **Step 4: 实现主题和响应式样式**：复用现有主题变量，保证深色主题、窄屏、滚动和表格边界不重叠。
- [ ] **Step 5: 运行测试并提交**：通过布局和源码测试后提交 `web: 对齐V78小说获取工作台——保留真实任务交互和状态展示`。

### Task 5: 补齐配置、知识库、规则、记录和网站提交面板

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Test: `test/novel-fetch-knowledge-recovery.test.js`
- Test: `test/novel-fetch-model-resolution.test.js`
- Test: `test/novel-fetch-selected-submit-payload-guard.test.js`

**Interfaces:**
- Consumes: 任务 2、任务 3 的配置和任务接口。
- Produces: V78 配置、知识库、规则、记录和网站提交标签页的完整交互。

- [ ] **Step 1: 写失败测试**：覆盖配置保存恢复、知识库增删改、规则预览/建议、提交预览、提交确认和失败版本重试。
- [ ] **Step 2: 运行测试确认缺口**：运行对应测试并记录缺失请求或状态。
- [ ] **Step 3: 补齐前后端操作绑定**：所有按钮显示 loading、成功、失败和错误详情；提交前保留配置检查和用户确认。
- [ ] **Step 4: 运行测试确认通过并提交**：提交 `web: 补齐V78配置规则和提交面板——所有按钮接入Go状态链路`。

### Task 6: 批量工厂交接、构建和旧入口隔离

**Files:**
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryNovelFetchHandoff.js`
- Modify: `backend/internal/web/static.go`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryNovelFetchHandoff.test.js`
- Test: `backend/internal/web/static_test.go`

**Interfaces:**
- Consumes: 已完成的任务和持久化书籍信息。
- Produces: 进入批量工厂时带上已有书籍、原文、风格、男女频和处理结果；旧静态页面不再被 `/novel-fetch` 路由误加载。

- [ ] **Step 1: 写失败测试**：验证交接 payload 完整、重复书籍不混批、React 路由不会回退到 `batch-rewrite`。
- [ ] **Step 2: 实现交接和路由隔离**：复用现有 Go/React 交接接口，移除旧入口的运行时引用但保留参考文件。
- [ ] **Step 3: 运行前端和 Go 测试并提交**：提交 `web: 固化小说获取到批量工厂交接——避免旧入口和新批次混用`。

### Task 7: 本地完整链路验收和发布产物验证

**Files:**
- Modify only if required by failing verification: `scripts/build-go-local.ps1`, `backend/web/assets.go`
- Test: `test/novel-fetch-workshop-process-routes.test.js`
- Test: `backend/internal/web/static_test.go`

- [ ] **Step 1: 构建 React**：运行 `pnpm --dir frontend build`，确认生成产物来自当前 React 源码。
- [ ] **Step 2: 构建 Go**：运行 `go test ./...`，再运行本地 Go 构建脚本，确认 `embed` 资源完整。
- [ ] **Step 3: 启动本地 Go 服务并验收**：实际打开 `/novel-fetch`，完成“输入书籍 → 抓取原文 → 分类 → AI/敏感处理 → 查看任务 → 进入批量工厂”。
- [ ] **Step 4: 检查运行时依赖**：确认页面请求不进入 Node 服务、不加载 iframe、不从旧静态入口加载。
- [ ] **Step 5: 提交验收记录**：提交 `test: 验证V78小说获取React-Go完整链路——确认运行时不依赖Node`。

# Liblib 文件夹式资源分组 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Agent Canvas 项目抽屉按书籍显示可展开/收起的文件夹分组，并用固定缩略图展示图片和视频。

**Architecture:** 继续使用前端派生数据，不新增接口或改变画布数据。导航辅助函数负责把素材、节点关联和书籍分组归一化；抽屉组件负责文件夹开合状态和交互；工作区 CSS 负责固定媒体预览尺寸与紧凑列表排版。

**Tech Stack:** React、Ant Design、Lucide、现有 Agent Workspace CSS、Node `node:test`、Vite、Go embed 本地运行时。

## Global Constraints

- 只修改 Agent Canvas 项目抽屉相关源码、测试和设计文档，不改后端接口和画布数据结构。
- 保留搜索、类型筛选、拖拽素材到画布和点击添加素材功能。
- 图片/视频必须限制在固定预览框内，不能按原始尺寸撑开列表。
- 保留工作区现有脏改动，不执行 reset、clean、checkout 或无关文件回退。

### Task 1: 固化分组派生数据与测试

**Files:**
- Modify: `frontend/src/user/pages/agent-workspace-navigation.js`
- Test: `frontend/src/user/pages/agent-workspace-navigation.test.js`

**Interfaces:**
- `buildWorkspaceAssetGroupTree(assets, groups, nodes) -> Array<{ id, title, children }>`：优先使用素材自身的 `groupId/bookId`，其次使用节点配置中的资源 ID 到 `groupId` 映射；单一书籍分组时将无归属素材归入该组，否则归入“未分组”。

- [x] **Step 1: 保留并运行分组覆盖测试**

Run: `node --test frontend/src/user/pages/agent-workspace-navigation.test.js`

Expected: 分组、节点筛选、素材筛选测试通过。

- [x] **Step 2: 检查派生结果的字段优先级**

确认 `bookId`、`groupId`、节点 `config.resourceId` 等字段均按同一规则归组，缺失字段不会抛错。

### Task 2: 将资源组渲染成文件夹交互

**Files:**
- Modify: `frontend/src/user/components/AgentWorkspaceDrawer.jsx`
- Test: `frontend/src/user/pages/agent-workspace-drawer.source.test.js`

**Interfaces:**
- 每个资源组渲染 `.agent-workspace-assets__group-header`，使用按钮语义和 `aria-expanded`；组内资源只在展开状态渲染。

- [x] **Step 1: 保留当前独立组状态**

每个组使用 `collapsedAssetGroups` 的 `Set` 保存状态，点击一个组不影响其他组。

- [ ] **Step 2: 增加源代码断言**

补充断言，检查抽屉源码包含 `aria-expanded`、`collapsedAssetGroups`、文件夹组标题和子项数量。

- [ ] **Step 3: 运行抽屉测试**

Run: `node --test frontend/src/user/pages/agent-workspace-drawer.source.test.js`

Expected: 抽屉相关断言全部通过；若命中仓库既有的 preview CSS 断言差异，只修正本次缺失的样式，不回退分组逻辑。

### Task 3: 固定图片/视频缩略图布局

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- `.agent-workspace-assets__preview` 提供固定宽高、溢出裁剪、背景色和圆角；其中的 `img/video` 使用 `width: 100%`、`height: 100%`、`object-fit: cover`。

- [ ] **Step 1: 添加固定预览容器样式**

桌面抽屉中使用约 `48px × 48px` 的紧凑预览框，资源行继续使用网格布局；音频和文本只显示类型图标，不受媒体尺寸影响。

- [ ] **Step 2: 添加样式回归断言**

在抽屉源测试中检查预览类和 CSS 固定尺寸/`object-fit` 规则存在。

### Task 4: 构建与浏览器验收

**Files:**
- No source changes unless verification exposes a scoped defect.

- [ ] **Step 1: 运行聚焦测试与差异检查**

Run: `node --test frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/pages/agent-workspace-drawer.source.test.js`

Run: `git diff --check -- frontend/src/user/pages/agent-workspace-navigation.js frontend/src/user/pages/agent-workspace-navigation.test.js frontend/src/user/components/AgentWorkspaceDrawer.jsx frontend/src/user/pages/agent-workspace-drawer.source.test.js frontend/src/shared/styles/agent-workspace.css`

- [ ] **Step 2: 构建前端并准备 Go embed 资源**

Run: `npm.cmd --prefix frontend run build`

Run: `.\scripts\build-go-local.ps1`

- [ ] **Step 3: 重启并验证本地 Go 运行时**

只停止路径为 `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe` 的 4001 进程，然后运行 `.\scripts\start-local-go-review.ps1 -UseYcmMysql`。

- [ ] **Step 4: 浏览器验收**

打开 `/agent/canvas?...&view=workflow`，从项目菜单打开项目抽屉，验证：文件夹可展开/收起；图片不再撑满侧栏；音频仍是紧凑行；拖拽和点击添加仍可用；控制台无新增错误。

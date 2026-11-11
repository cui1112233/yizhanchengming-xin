# Asset Reference Visual Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 保留一战晟铭导航和数据链路，将资产参考内容区按用户参考图重做并完成本地 4001 可见验收。

**Architecture:** `AssetReferenceWorkbench.jsx` 负责页面结构与交互状态，`asset-reference-workbench.css` 负责参考图布局和主题适配，现有 `asset-reference-workbench.js` 继续负责纯业务规则。批量选择、预设和参考库只操作本工作台已有数据，不引入洛水私有接口。

**Tech Stack:** React、Ant Design、Lucide React、CSS、Node test、Vite、Go embed。

## Global Constraints

- 保留一战晟铭左侧导航、登录状态和主题系统。
- 不复制洛水源码、密钥、私有接口或专有资源。
- 不进行付费图片调用。
- 不提交、不推送、不清理用户的脏工作树。

---

### Task 1: 锁定参考图结构

**Files:**
- Modify: `frontend/src/user/pages/asset-reference-workbench.layout.test.js`
- Test: `frontend/src/user/pages/asset-reference-workbench.layout.test.js`

**Interfaces:**
- Consumes: `AssetReferenceWorkbench.jsx` 源码。
- Produces: 参考图必需控件和页面骨架的回归断言。

- [ ] **Step 1: 添加失败测试**

断言源码包含 `asset-reference-page-head`、`asset-reference-result-canvas`、`已授权`、`全选`、`下载`、`批量删除`、`预设`、`参考库`、`@ 可引用参考库卡片` 和关闭按钮。

- [ ] **Step 2: 验证测试按预期失败**

Run: `node --test src/user/pages/asset-reference-workbench.layout.test.js`

Expected: FAIL，指出上述参考图结构尚不存在。

### Task 2: 重构页面 DOM 和交互

**Files:**
- Modify: `frontend/src/user/components/AssetReferenceWorkbench.jsx`
- Test: `frontend/src/user/pages/asset-reference-workbench.layout.test.js`

**Interfaces:**
- Consumes: 现有生成结果、参考图片、设置和生成函数。
- Produces: `selectedResultIds`、批量下载/删除、参数预设、参考库弹窗和参考图一致的 DOM。

- [ ] **Step 1: 实现最小结构**

增加页面头、结果工具条、大画布、固定右栏、预设/参考库入口和底部操作区；保留现有生成链路。

- [ ] **Step 2: 实现选择和批量操作**

结果卡片点击切换选择；全选限当前搜索结果；下载逐个复用 `downloadResult`；批量删除只更新本地结果列表。

- [ ] **Step 3: 运行聚焦测试**

Run: `node --test src/user/pages/asset-reference-workbench.layout.test.js src/user/pages/pluginCatalog.layout.test.js`

Expected: PASS。

### Task 3: 对齐参考图 CSS

**Files:**
- Modify: `frontend/src/user/components/asset-reference-workbench.css`
- Test: `frontend/src/user/pages/asset-reference-workbench.layout.test.js`

**Interfaces:**
- Consumes: Task 2 的类名和 DOM。
- Produces: 白底/浅灰画布、74/26 分栏、390-420px 右栏、底部按钮和响应式布局。

- [ ] **Step 1: 添加 CSS 结构断言并确认失败**

验证右栏宽度、结果画布最小高度、底部操作定位、1250px 压缩双栏和 850px 单栏规则。

- [ ] **Step 2: 实现主题化样式**

以 CSS 变量实现明暗主题，桌面布局优先匹配 1919×1050 参考图。

- [ ] **Step 3: 验证测试通过**

Run: `node --test src/user/pages/asset-reference-workbench.layout.test.js`

Expected: PASS。

### Task 4: 构建和本地可见验收

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/web/dist/**`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: 完成的 React/CSS。
- Produces: `http://127.0.0.1:4001/plugins` 可见页面。

- [ ] **Step 1: 运行生产构建**

Run: `npm run build`

Expected: Vite exit 0。

- [ ] **Step 2: 重建并重启已核对的本地运行时**

只停止监听 4001 且路径为本仓库 `backend/qiantie-local.exe` 的进程，然后运行 `scripts/build-go-local.ps1` 和 `scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached`。

- [ ] **Step 3: 健康与运行身份验证**

检查 `/health`、`/api/runtime-build-info` 和 `/plugins`，确认新入口文件被加载。

- [ ] **Step 4: 浏览器视觉验收**

刷新页面后进入资产参考，核对参考图中的页面头、结果画布、右栏、六种比例、默认数量 5、预设/参考库和底部 AI 生图按钮；不点击生成。

# 洛水 v2.9.0 对照第一期实施计划

> **执行要求：** 使用 `superpowers:executing-plans` 逐项实施；所有生产代码使用 TDD，先观察测试因缺失能力而失败，再写最小实现。

**目标：** 在 V88 现有 Agent Canvas、分镜、导演台和生成任务基础上，完成第一期 17 项洛水功能对照，包括多主题、四工作区、统一素材引用、分镜画布联动、导演台、后台任务和首批提示词。

**架构：** 保留 `/agent` 和 `/agent/canvas` 唯一入口。React 负责工作台显示和交互，Go API 负责账号偏好、项目数据、任务和提示词组装；MySQL/现有持久层保存正式业务数据，Redis承担任务协调，TOS保存媒体。不得新增第二套画布、素材库、任务队列或 Node 业务后端。

**技术栈：** Go、React 18、Ant Design 5、Node test runner、Go `testing`、Go embed。

## 全局约束

- 直接在 `v88` 工作；不创建新分支或额外 worktree。
- 保留所有无关改动，不 reset、不 clean、不批量覆盖。
- 不调用真实付费模型，不消耗额度，不发布公网，不推送远端。
- 每个功能编号必须更新 `docs/obj/2026-11-09-agent-canvas-director-parity-matrix.md`。
- “按钮存在”“接口 200”“构建通过”不等于功能一致。
- 所有系统提示词只进入 `backend/internal/prompt/`，前端不硬编码完整系统提示词。
- 所有浮层使用 portal，不能被画布裁剪。
- 每项完成前必须有新鲜测试输出；集成完成后才进行 Go embed 和浏览器验收。

---

### Task 1：建立基线和对照台账

**文件：**

- Create: `docs/obj/2026-11-09-agent-canvas-director-parity-matrix.md`
- Verify: `frontend/src/user/App.theme.test.js`
- Verify: `frontend/src/user/pages/agent-workspace-view.test.js`
- Verify: `frontend/src/user/pages/agent-workspace-navigation.test.js`
- Verify: `backend/internal/httpapi/member_profile_test.go`

- [ ] **Step 1：记录 Git、运行和测试基线**

运行：

```powershell
git status --short
git rev-parse --short HEAD
node --test src/user/App.theme.test.js src/user/pages/agent-workspace-view.test.js src/user/pages/agent-workspace-navigation.test.js
go test ./internal/httpapi -run 'TestMemberProfile' -count=1
```

- [ ] **Step 2：为 17 个功能编号填写现状**

每项记录洛水证据、V88 文件/API、当前状态、差异和验收命令。没有完整链路证据的项标为 `partial`。

- [ ] **Step 3：提交计划和台账**

```powershell
git add docs/superpowers/plans/2026-11-09-luoshui-v29-agent-canvas-director-phase1.md docs/obj/2026-11-09-agent-canvas-director-parity-matrix.md
git commit -m "docs: plan luoshui agent parity phase one"
```

---

### Task 2：统一主题目录和语义令牌

**功能编号：** `L29-TH-001`

**文件：**

- Create: `frontend/src/shared/styles/theme-catalog.js`
- Create: `frontend/src/shared/styles/theme-catalog.test.js`
- Modify: `frontend/src/shared/styles/theme.js`
- Modify: `frontend/src/user/App.theme.test.js`

**接口：**

- `listThemeDefinitions()` 返回可选主题定义；
- `normalizeThemeId(value, systemMode)` 处理旧值、非法值和跟随系统；
- `themeDefinition(value, systemMode)` 返回完整语义令牌；
- `createAntTheme(value, systemMode)` 从同一令牌生成 Ant Design token。

- [ ] **Step 1：写失败测试**

测试要求：

- 目录包含 `system` 和 12 个洛水对照主题；
- 每个主题都有完整语义令牌；
- 旧的 `dark`、`light` 保持兼容；
- 非法值回退 `dark`；
- `system` 根据系统模式解析，但仍保留用户选择 ID；
- 深色、纯黑和浅色主题分别生成可读的 Ant Design token。

- [ ] **Step 2：运行并确认 RED**

```powershell
node --test src/shared/styles/theme-catalog.test.js src/user/App.theme.test.js
```

预期：因 `theme-catalog.js` 不存在或主题目录缺失而失败。

- [ ] **Step 3：写最小主题目录实现**

主题 ID：`system`、`dark`、`light`、`graywhite`、`mint`、`snow`、`vanilla`、`sakura`、`pureblack`、`purple`、`willow`、`ivory`、`lotus`。

- [ ] **Step 4：运行并确认 GREEN**

```powershell
node --test src/shared/styles/theme-catalog.test.js src/user/App.theme.test.js
```

---

### Task 3：把多主题投影到全局 CSS

**功能编号：** `L29-TH-001`

**文件：**

- Modify: `frontend/src/shared/styles/global.css`
- Modify: `frontend/src/shared/styles/contrast-contract.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/shared/styles/theme-css-contract.test.js`

- [ ] **Step 1：写失败的行为测试**

从真实 CSS 读取每个公开主题块，验证所有语义变量完整；验证画布、任务抽屉、Agent 输入和导演台面板使用语义变量，不固定写死深色背景。

- [ ] **Step 2：运行并确认 RED**

```powershell
node --test src/shared/styles/theme-css-contract.test.js src/shared/styles/contrast-contract.test.js
```

- [ ] **Step 3：补充主题 CSS 变量**

使用 `[data-theme='<id>']` 只定义语义变量。组件继续消费 `--surface-*`、`--text-*`、`--border-*` 和新增画布/视口变量。

- [ ] **Step 4：运行并确认 GREEN**

```powershell
node --test src/shared/styles/theme-css-contract.test.js src/shared/styles/contrast-contract.test.js
```

---

### Task 4：账号主题偏好持久化

**功能编号：** `L29-TH-001`

**文件：**

- Modify: `backend/internal/httpapi/member_profile.go`
- Modify: `backend/internal/httpapi/member_profile_test.go`
- Modify: `frontend/src/shared/api/member.js`
- Create: `frontend/src/shared/styles/theme-preference.js`
- Create: `frontend/src/shared/styles/theme-preference.test.js`

**接口：**

- `PATCH /api/member/profile` 接受 `themeId`；
- `GET /api/member/me` 的 `member.themeId` 返回已保存主题；
- `loadCachedTheme()` 和 `persistCachedTheme()` 只负责首屏缓存；
- 登录后的账号值覆盖本机缓存。

- [ ] **Step 1：写 Go 失败测试**

验证允许的主题可以保存并读回，非法主题返回 400，旧账号默认 `system`。

- [ ] **Step 2：运行 Go RED**

```powershell
go test ./internal/httpapi -run 'TestMemberProfileTheme' -count=1
```

- [ ] **Step 3：写最小后端实现并确认 GREEN**

只向现有 `member-profiles.json` 增加 `themeId` 字段，不新建数据库或配置文件。

- [ ] **Step 4：写前端缓存失败测试并实现**

```powershell
node --test src/shared/styles/theme-preference.test.js
```

验证旧缓存、非法缓存、未登录和登录覆盖语义。

---

### Task 5：主题选择器和即时切换

**功能编号：** `L29-TH-001`

**文件：**

- Create: `frontend/src/shared/components/ThemePicker.jsx`
- Create: `frontend/src/shared/components/ThemePicker.source.test.js`
- Modify: `frontend/src/shared/layouts/UserLayout.jsx`
- Modify: `frontend/src/shared/layouts/AdminLayout.jsx`
- Modify: `frontend/src/shared/layouts/UserLayout.source.test.js`
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Modify: `frontend/src/user/pages/settings-workbench-layout.source.test.js`

- [ ] **Step 1：写失败测试**

验证侧栏和设置页使用同一个 ThemePicker；色卡有名称和四个关键色；选择时先即时切换，再异步保存；保存失败保留本地显示并明确提示未同步。

- [ ] **Step 2：运行并确认 RED**

```powershell
node --test src/shared/components/ThemePicker.source.test.js src/shared/layouts/UserLayout.source.test.js src/user/pages/settings-workbench-layout.source.test.js
```

- [ ] **Step 3：实现 ThemePicker 和布局接线**

使用 Popover/Drawer 的 portal；旧开关迁移为主题按钮。切主题只更新显示，不调用画布保存或生成接口。

- [ ] **Step 4：运行并确认 GREEN**

运行上述测试以及 `src/user/App.theme.test.js`。

---

### Task 6：四工作区导航

**功能编号：** `L29-AG-002`、`L29-CV-003`、`L29-SB-003`

**文件：**

- Modify: `frontend/src/user/pages/agent-workspace-view.js`
- Modify: `frontend/src/user/pages/agent-workspace-view.test.js`
- Modify: `frontend/src/user/components/AgentWorkspaceHeader.jsx`
- Modify: `frontend/src/user/components/AgentWorkspaceHeader.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

- [ ] **Step 1：写失败测试**

要求 `workflow` 兼容迁移到“画布”，并支持 `workflow/storyboard/assets/tasks`；URL、选中节点、视口和对话面板在切换后保持规定状态。

- [ ] **Step 2：运行并确认 RED**

```powershell
node --test src/user/pages/agent-workspace-view.test.js src/user/components/AgentWorkspaceHeader.source.test.js
```

- [ ] **Step 3：实现导航状态和四视图壳层**

资产和任务视图优先投影现有 `workspaceAssets`、`executions`、`generations`，不得建立第二份数据。

- [ ] **Step 4：确认 GREEN**

运行聚焦测试和 Agent 工作区布局测试。

---

### Task 7：统一素材页和上下文引用

**功能编号：** `L29-AG-004`、`L29-CV-003`、`L29-CV-005`

**文件：**

- Modify: `frontend/src/user/pages/agent-workspace-navigation.js`
- Modify: `frontend/src/user/pages/agent-workspace-navigation.test.js`
- Modify: `frontend/src/user/components/AgentMaterialLibraryDrawer.jsx`
- Modify: `frontend/src/user/components/AgentMaterialLibraryDrawer.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/agent-node-mentions.js`
- Modify: `frontend/src/user/pages/agent-node-mentions.test.js`

- [ ] **Step 1：写失败测试**

覆盖素材唯一 ID、来源和版本显示；`@` 引用资产和节点时传稳定 ID；取消选择后不携带旧上下文。

- [ ] **Step 2：确认 RED，最小实现，再确认 GREEN**

```powershell
node --test src/user/pages/agent-workspace-navigation.test.js src/user/pages/agent-node-mentions.test.js src/user/components/AgentMaterialLibraryDrawer.source.test.js
```

---

### Task 8：分镜与画布双向关系和分镜合并

**功能编号：** `L29-SB-001`、`L29-SB-002`、`L29-SB-003`

**文件：**

- Modify: `frontend/src/user/pages/agent-storyboard-model.js`
- Modify: `frontend/src/user/pages/agent-storyboard-model.test.js`
- Create: `frontend/src/user/pages/agent-storyboard-merge.js`
- Create: `frontend/src/user/pages/agent-storyboard-merge.test.js`
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `backend/internal/agentworkspace/shot_batch.go`
- Modify: corresponding Go tests under `backend/internal/agentworkspace/`

- [ ] **Step 1：写前端和 Go 失败测试**

覆盖稳定分镜 ID、批量建节点幂等、双向定位、10/15/30 秒相邻合并、台词/提示词/资产引用合并和历史结果保留。

- [ ] **Step 2：分别确认 RED，写最小实现，再确认 GREEN**

```powershell
node --test src/user/pages/agent-storyboard-model.test.js src/user/pages/agent-storyboard-merge.test.js
go test ./internal/agentworkspace -run 'Test.*ShotBatch|Test.*StoryboardMerge' -count=1
```

---

### Task 9：导演台一致性补齐

**功能编号：** `L29-DD-001`、`L29-DD-002`、`L29-DD-003`

**文件：**

- Modify: `frontend/src/user/components/director-desk/director-scene-state.js`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.test.js`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorInspector.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: existing Director Desk source tests

- [ ] **Step 1：写失败测试**

覆盖人物/场景/道具对象、多机位、语义镜头参数、画幅、截图回画布和只同步直接连接视频节点。

- [ ] **Step 2：确认 RED，最小实现，再确认 GREEN**

```powershell
node --test src/user/components/director-desk/*.test.js src/user/components/AgentCanvas.director-desk.source.test.js
```

---

### Task 10：首批八个提示词定义

**功能编号：** `L29-PR-001`、`L29-SB-001`

**文件：**

- Create: `backend/internal/prompt/agent/definitions.go`
- Create: `backend/internal/prompt/asset/definitions.go`
- Create: `backend/internal/prompt/storyboard/definitions.go`
- Create: `backend/internal/prompt/director/definitions.go`
- Create: corresponding `_test.go` files
- Modify: existing Agent/Storyboard Go service call sites selected during implementation

- [ ] **Step 1：写失败测试**

每个 PromptDefinition 验证稳定 ID、版本、必填输入、结构化输出契约和错误语义。测试使用脱敏固定样本，不调用模型。

- [ ] **Step 2：确认 RED，最小实现，再确认 GREEN**

```powershell
go test ./internal/prompt/... -count=1
```

- [ ] **Step 3：逐个接入调用点**

每接一个调用点先写失败测试，任务记录保存 Prompt ID 和版本。不得整库复制洛水 181 条提示词。

---

### Task 11：统一生成任务状态和恢复

**功能编号：** `L29-AG-003`、`L29-CV-004`

**文件：**

- Modify: `backend/internal/agentworkspace/generation*.go`
- Modify: `backend/internal/httpapi/agent_workspace*.go`
- Modify: corresponding Go tests
- Modify: `frontend/src/user/pages/agent-generation-state.js`
- Modify: `frontend/src/user/pages/agent-generation-state.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`

- [ ] **Step 1：写失败测试**

覆盖 `queued/running/awaiting_confirmation/completed/failed/cancelled`、失败原因、临时错误重试、取消、重启恢复和多视频节点失败隔离。

- [ ] **Step 2：确认 RED，逐层实现并确认 GREEN**

```powershell
go test ./internal/agentworkspace ./internal/httpapi -run 'Test.*Generation' -count=1
node --test src/user/pages/agent-generation-state.test.js src/user/pages/agent-canvas-video-jobs.test.js
```

---

### Task 12：集成构建和真实验收

**文件：**

- Update: `docs/obj/2026-11-09-agent-canvas-director-parity-matrix.md`
- Verify: `frontend/dist/`
- Verify: `backend/web/dist/`
- Verify: `backend/qiantie-local.exe`

- [ ] **Step 1：完整聚焦测试**

```powershell
npm test
go test ./internal/agentworkspace ./internal/httpapi ./internal/prompt/... -count=1
```

既有无关失败必须单列，不得冒充本期失败或擅自修复无关模块。

- [ ] **Step 2：前端和 Go embed 构建**

使用仓库现有正式脚本；不手工复制 `dist`。记录构建命令、退出码和产物身份。

- [ ] **Step 3：本地运行验证**

只重启已确认的 `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`，验证：

- `/health` 返回 200；
- `/api/runtime-build-info` 为当前 Go embed 构建；
- 不启动第二个 4001 服务。

- [ ] **Step 4：已登录浏览器验收**

不提交付费任务，检查：

- 多主题即时切换和账号恢复；
- 画布、分镜、素材、任务四视图；
- `@` 引用和选中上下文；
- 分镜与画布双向定位；
- 导演台对象、机位、画幅、保存和截图链；
- 任务状态、失败提示和历史版本。

- [ ] **Step 5：更新一致性状态**

只有具备自动测试、本地运行和浏览器证据的功能才能标记 `matched` 或 `equivalent`。其余保持 `partial`。

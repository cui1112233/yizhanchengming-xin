# Agent Canvas 画布级生成框与快捷面板 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将现有 Agent Canvas 生成框改成画布容器内的浮动组件，并让已有模式、技能、专家、附件、找问题和连接器能力通过互斥的顶部快捷面板使用。

**Architecture:** 保留 `AgentPageV2.jsx` 中已有状态和处理函数，只整理生成框的呈现结构，把快捷入口和浮层统一放在生成框的上层区域。使用 `agent-workspace.css` 负责容器锚定、浮层翻转和窄屏适配；不改后端接口和任务提交协议。

**Tech Stack:** React、Ant Design、现有 Agent Canvas 状态/素材 API、CSS、Node 内置 source tests、Vite、Go embed 构建脚本。

## Global Constraints

- 只在 Git `v88` 上工作，保留所有无关脏文件，不 reset、clean 或覆盖。
- 不新增接口、不改数据库、不触发真实图片、视频、音频或模型生成。
- 附件能力继续支持 TXT、MD、CSV、JSON、图片、视频、音频；文本上限 200 KB，媒体上限 128 MB。
- 故事板模式隐藏专家入口；技能最多选择 3 项；连接器未配置时只显示现有提示。
- 前端改动必须通过前端构建、Go 内嵌构建和本地 Go 运行时健康检查。

---

### Task 1: 固化生成框行为契约测试

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-composer.source.test.js`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx` only if a testable semantic hook is missing

**Interfaces:**
- Consumes: `AgentPageV2.jsx` composer markup, existing `openComposerTool`, `composerPanel`, `composerMenuOpen` behavior.
- Produces: source-level assertions that later layout edits must preserve.

- [ ] **Step 1: Write failing source assertions**

  Assert the composer contains the mode, attachment, skill, expert and plus-menu entry labels; the panel branches remain keyed by `composerPanel`; the send button remains disabled for empty input and uses loading state; storyboard markup does not render the expert shortcut.

- [ ] **Step 2: Run the focused test and verify the baseline**

  Run: `node --test frontend/src/user/pages/agent-canvas-composer.source.test.js`

  Expected: PASS against the current source; if an assertion exposes a mismatch, adjust only the test to the confirmed current contract before implementation.

- [ ] **Step 3: Commit the contract test**

  Run: `git add frontend/src/user/pages/agent-canvas-composer.source.test.js; git commit -m "test(agent): cover composer capability contract"`

### Task 2: Restructure the composer markup for canvas anchoring

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx:1270-1375`

**Interfaces:**
- Consumes: Task 1 source contract; existing attachment, selected asset, skill, mode, expert, problem-search and send handlers.
- Produces: one composer root with explicit `__surface`, `__body`, `__shortcuts`, `__overlay` regions; all existing handlers and state remain unchanged.

- [ ] **Step 1: Add semantic regions without changing state logic**

  Wrap the attachment/skill/input content in a body region, move the action row into a shortcut region, and render `composerPanel` branches inside an overlay region placed before the body in DOM order so the overlay can rise above the composer. Keep the plus menu inside the shortcut region.

- [ ] **Step 2: Preserve actual capability behavior**

  Keep `openComposerTool('file')` as a direct file-picker action, keep `mode`/`expert` selections closing their panels, keep skill multi-select open, and leave connector as the existing informational message. Do not add placeholder options.

- [ ] **Step 3: Run the focused source test**

  Run: `node --test frontend/src/user/pages/agent-canvas-composer.source.test.js`

  Expected: PASS, with no real generation call.

- [ ] **Step 4: Commit the markup change**

  Run: `git add frontend/src/user/pages/AgentPageV2.jsx frontend/src/user/pages/agent-canvas-composer.source.test.js; git commit -m "refactor(agent): structure canvas composer overlays"`

### Task 3: Implement Liblib-style floating layout and responsive overlays

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css:1170-1220` and composer-specific rules near the existing storyboard composer styles

**Interfaces:**
- Consumes: Task 2 semantic regions and existing CSS variables, director split, canvas zoom classes and mobile breakpoints.
- Produces: container-anchored composer, above-composer overlays, bounded panel widths, readable minimum text sizing and narrow-screen fallback.

- [ ] **Step 1: Add desktop container anchoring**

  Make the conversation panel the positioning context, set the composer to the panel bottom with explicit horizontal margins, and give the overlay an absolute position above the composer with a bounded width and z-index above nodes/edges but below global dialogs.

- [ ] **Step 2: Add panel alignment and overflow rules**

  Align compact panels to the triggering shortcut where practical; for wide skill panels use the available composer width; constrain height and allow internal scrolling; ensure panels cannot escape the conversation/画布 container.

- [ ] **Step 3: Add narrow-screen behavior**

  At existing 860/560 breakpoints, keep the composer full width, let shortcuts scroll horizontally, make overlays width `min(100%, calc(100vw - 24px))`, and preserve 44px-equivalent click targets without changing the send contract.

- [ ] **Step 4: Verify CSS syntax and source contract**

  Run: `node --test frontend/src/user/pages/agent-canvas-composer.source.test.js`

  Expected: PASS; no CSS parser/build errors.

- [ ] **Step 5: Commit the layout change**

  Run: `git add frontend/src/shared/styles/agent-workspace.css; git commit -m "style(agent): float composer overlays within canvas"`

### Task 4: Build and run the existing canvas regression checks

**Files:**
- Modify: none unless a test reports a direct regression in the new composer behavior

**Interfaces:**
- Consumes: Tasks 1-3 implementation.
- Produces: verified frontend artifact and documented known historical residuals.

- [ ] **Step 1: Run focused interaction tests**

  Run: `node --test frontend/src/user/pages/agent-canvas-commands.test.js frontend/src/user/components/AgentCanvasInteraction.source.test.js frontend/src/user/pages/agent-canvas-composer.source.test.js`

  Expected: all focused tests pass.

- [ ] **Step 2: Build the frontend**

  Run: `npm.cmd --prefix frontend run build`

  Expected: exit code 0 and updated `frontend/dist` generated from source.

- [ ] **Step 3: Build the Go embedded runtime**

  Run: `./scripts/build-go-local.ps1`

  Expected: exit code 0 and Go binary embeds the current frontend build.

- [ ] **Step 4: Restart only the approved Go runtime and verify identity**

  Run: `./scripts/start-local-go-review.ps1 -UseYcmMysql`

  Then check `http://127.0.0.1:4001/health` is `200` and `/api/runtime-build-info` reports `runtime=go` with matching current `git_sha` and `runtime_manifest_sha`.

- [ ] **Step 5: Browser-check without generation**

  In the existing Agent Canvas workflow page, verify the composer stays inside the canvas, mode/expert/skill panels open above it, skill multi-select remains usable, attachments and dragged assets remain visible, and mobile/narrow layout does not overflow. Do not click send or start a real generation.

## Verification Notes

The existing `agent-canvas-collapsible-editor.source.test.js` historical failures remain recorded as residuals if they still fail; do not change current interaction code solely to satisfy old width, handle-visibility or obsolete copy assertions.

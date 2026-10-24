# Agent Canvas LibLib Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将工作流画布节点改成参考 LibLib 的一体化预览/输入卡片，并修正加号句柄与真实连线、空白拖放菜单之间的交互边界。

**Architecture:** 保留现有 React Flow、节点数据结构、生成回调和连接关系函数，只重排 `CanvasNode` 的 JSX，使 `AgentNodeComposer` 成为节点卡片内部编辑区。通过 CSS 让编辑区与媒体预览同宽、按选中状态展开，并让 source/target Handle 只在相关交互状态显示。连接逻辑继续集中在现有 `agent-canvas-connections.js` 和 `AgentCanvas.jsx` 的事件回调中。

**Tech Stack:** React 18, `@xyflow/react`, Ant Design, Lucide React, CSS, Node test runner, Vite/Go embed build.

## Global Constraints

- 只修改工作流画布相关前端源码和测试，不改故事板、批量工厂和后端生成接口。
- 保留现有模型、素材、生成、草稿保存、快捷键和中键拖动画布行为。
- 加号必须是 React Flow source Handle；连接到目标句柄时不弹新节点菜单。
- 不复制参考站点的项目数据、素材、网络请求或页面源码。
- 所有变更先写入 `v88` Git，保留工作区其他未提交改动。

---

### Task 1: Lock the node and connection contracts with tests

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-controls.source.test.js`
- Create: `frontend/src/user/pages/agent-canvas-liblib-layout.source.test.js`

**Interfaces:**
- Consumes: current `CanvasNode` markup, `AgentNodeComposer` callback interface, `shouldOpenConnectionDrop` and `connectionDropChoices`.
- Produces: source-level regression checks for embedded composer, handle visibility classes, and target-vs-empty connection behavior.

- [ ] **Step 1: Write failing source-contract tests**

  Add assertions that `AgentCanvas.jsx` renders `agent-node-composer` inside `agent-canvas-node-card`, that the node still has exactly one target and one source Handle, and that the source Handle keeps `aria-label="拖出连接"`. Add CSS assertions that the composer no longer uses `left: 50%` or `width: min(680px, calc(100vw - 96px))`, and that the new embedded state class exists.

- [ ] **Step 2: Run the focused tests and verify they fail**

  Run:

  ```powershell
  node --test frontend/src/user/pages/agent-canvas-liblib-layout.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js
  ```

  Expected: the new markup and CSS assertions fail before the JSX/CSS changes.

- [ ] **Step 3: Add connection branch assertions**

  In the new test, import `shouldOpenConnectionDrop`, `connectionDropChoices`, and `createTypedCanvasEdge`; assert that a completed target drop returns false, an empty drop returns true, and a valid image-to-video connection creates `data.relation === 'first_frame'`.

- [ ] **Step 4: Run the connection tests**

  ```powershell
  node --test frontend/src/user/pages/agent-canvas-connections.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js
  ```

  Expected: existing connection behavior remains green while the new layout contract remains the only failing group.

### Task 2: Embed the composer in the node card

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`

**Interfaces:**
- Consumes: existing `CanvasNode` props and `AgentNodeComposer` props; no callback signature changes.
- Produces: a node card whose preview and expanded composer share one width and whose Handle elements remain outside the card content.

- [ ] **Step 1: Move the composer JSX into `.agent-canvas-node-card`**

  Keep the current `composer` variable, but render `{composer}` immediately after the preview/text/director content and before the card footer. Keep the target Handle and source Handle as children of `.agent-canvas-node-preview-shell`, outside the card, so React Flow can calculate connection anchors without the editor changing the handle parent.

- [ ] **Step 2: Add explicit expanded state hooks**

  Add `isEditorOpen` to the node card class list as the existing state marker and add `data-editor-open={isEditorOpen ? 'true' : 'false'}` to the article for semantic styling/testing. Do not add a second editor state or duplicate prompt state.

- [ ] **Step 3: Preserve outside-click behavior**

  Keep `onMouseDown`/`onDoubleClick` stop handling on the composer. Confirm `handleNodeClick` still clears `activeEditorNodeID` only when the click target is outside `.agent-node-composer` and the preview/header selectors.

- [ ] **Step 4: Run the focused source tests**

  ```powershell
  node --test frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-canvas-liblib-layout.source.test.js
  ```

  Expected: embedded composer and editor state tests pass.

### Task 3: Restyle the card, editor and connection handles

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-controls.source.test.js`

**Interfaces:**
- Consumes: the embedded DOM from Task 2 and existing appearance variables.
- Produces: compact LibLib-inspired card visuals without changing data or interaction APIs.

- [ ] **Step 1: Replace detached composer sizing**

  Change `.agent-node-composer` from the centered viewport-sized panel to `width: 100%; margin-top: 0; border: 0; border-top: 1px solid ...; border-radius: 0 0 8px 8px; transform: none;` with compact internal padding. Add `.agent-canvas-node-card > .agent-node-composer { display: grid; }` and keep the editor hidden by React when closed.

- [ ] **Step 2: Align preview and controls**

  Make media previews use the card width, reduce shadow/contrast, and make `.agent-node-composer__toolbar` a horizontally scrollable single row at narrow widths. Keep the model picker flex basis large enough to show the configured model name; allow parameter summary to ellipsize only after the model control.

- [ ] **Step 3: Make handles contextual and attached**

  Set target/source handles to `opacity: 0` and `pointer-events: none` by default, then enable them on `.agent-canvas-node-preview-shell:hover`, `.react-flow__node.selected`, `.agent-canvas-node.is-editor-open`, `.agent-canvas.is-connecting`, and `:focus-visible`. Keep the plus glyph on the source Handle only; do not add a second absolute plus button.

- [ ] **Step 4: Tune edge and drop-menu visuals**

  Keep typed edge colors from `appearanceCssVariables`, reduce the default stroke to a readable 1.5–2px visual, and style the connection drop menu as a compact contextual menu anchored near the release point. Do not change menu labels or connection choices.

- [ ] **Step 5: Run CSS/source tests**

  ```powershell
  node --test frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-controls.source.test.js frontend/src/user/pages/agent-canvas-appearance.test.js
  ```

  Expected: no detached composer selector remains and handle/menu contracts pass.

### Task 4: Verify the real workflow canvas

**Files:**
- No new production files.
- Test evidence: `test-results/agent-canvas-liblib-layout-20261024/` (generated locally and not committed unless already tracked by project rules).

**Interfaces:**
- Consumes: completed JSX/CSS and existing local Go runtime.
- Produces: browser evidence that editing, connection and persistence behaviors work together.

- [ ] **Step 1: Run the complete focused frontend suite**

  ```powershell
  node --test frontend/src/user/pages/agent-canvas-*.test.js frontend/src/user/pages/agent-node-config.test.js frontend/src/user/pages/agent-model-picker.source.test.js
  ```

  Expected: all selected canvas/model tests pass.

- [ ] **Step 2: Build the frontend**

  ```powershell
  npm --prefix frontend run build
  ```

  Expected: production React assets build successfully.

- [ ] **Step 3: Build and start the Go embedded runtime**

  Use the existing V88 Go embed build script, then check the local health endpoint and runtime SHA. Do not overwrite production or unrelated generated assets.

- [ ] **Step 4: Browser-check the workflow canvas**

  Verify in the live local page:

  1. Clicking image/video preview opens an editor directly beneath the preview.
  2. Clicking canvas blank space hides the editor while preserving prompt, model and reference assets.
  3. Dragging source plus to an existing target creates a visible typed edge without opening the drop menu.
  4. Dragging source plus to blank space opens the two-choice continuation menu; choosing one creates and selects a connected node.
  5. Middle mouse pans, wheel zooms, edge selection plus Delete removes an edge, and the model label remains readable in a narrow viewport.

- [ ] **Step 5: Commit only the canvas implementation files**

  ```powershell
  git add frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-canvas-liblib-layout.source.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-canvas-controls.source.test.js
  git commit -m "feat(agent): align workflow canvas nodes and connections"
  ```

  Do not stage or reset unrelated dirty files.

# Agent Canvas Visual Language Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (- [ ]) syntax for tracking.

**Goal:** Rebuild the local V88 Agent canvas as a professional V88-dark creative workspace with project-specific appearance settings and animated connection lines, without losing existing projects or conversations.

**Architecture:** Extend the existing Go Canvas contract with one normalized Appearance object and persist it on the existing agent_canvases row through one append-only migration. Keep AgentCanvas.jsx as the single React Flow owner; a pure frontend appearance module supplies defaults, validation, CSS variables, and edge styles. The existing canvas save queue remains the only persistence path.

**Tech Stack:** Go, MySQL 8.4, existing Go migrations, React 18, Ant Design 5, @xyflow/react, Lucide, Node test runner, Vite, Go embed.

## Global Constraints

- Local V88 only. Do not publish publicly.
- Preserve V88 dark colors and branding. Do not copy reference CSS, names, logos, assets, or service behavior.
- Keep all current Agent projects, messages, nodes, edges, executions, canvas tools, auto-save, manual save, and revision conflict handling.
- Do not add eye-care mode.
- Default connection animation must create no network request, business task, node, or message.
- Persist only connection color, connection animation, connection thickness, background dots, and alignment guides.
- Add only migration 8800003. Do not alter applied migrations 8800001 or 8800002.
- Preserve unrelated dirty workspace changes and do not restore the Node release chain.

---

## File Structure

- backend/internal/agentworkspace/types.go: Canvas Appearance contract and Go defaults.
- backend/internal/agentworkspace/memory_store.go: applies defaults to API-test and legacy projects.
- backend/internal/agentworkspace/mysql_store.go: reads and writes appearance_json in the same revision transaction.
- backend/internal/agentworkspace/mysql_store_test.go, store_test.go: persistence and legacy-default regression tests.
- backend/internal/httpapi/agent_workspace_test.go: canvas API appearance round trip.
- backend/internal/storage/agent_workspace_schema.go: append-only migration 8800003.
- backend/internal/storage/agent_workspace_schema_test.go, migrations_mysql84_test.go: migration order and repeat-run tests.
- frontend/src/user/pages/agent-canvas-appearance.js: browser defaults, validation, CSS-variable conversion, edge styles.
- frontend/src/user/pages/agent-canvas-appearance.test.js: pure appearance tests.
- frontend/src/user/components/AgentCanvas.jsx: custom animated edge, appearance panel, background and guide behavior.
- frontend/src/user/pages/AgentPageV2.jsx: keeps the current save queue as the only write path.
- frontend/src/shared/styles/agent-workspace.css: scoped V88-dark visual redesign.
- frontend/src/user/pages/agent-workspace.layout.test.js: visible controls, responsive layout, and motion source assertions.

## Task 1: Add Appearance Persistence Without Creating a Second Data Path

**Files:**
- Modify: backend/internal/agentworkspace/types.go
- Modify: backend/internal/agentworkspace/memory_store.go
- Modify: backend/internal/agentworkspace/mysql_store.go
- Modify: backend/internal/agentworkspace/mysql_store_test.go
- Modify: backend/internal/agentworkspace/store_test.go
- Modify: backend/internal/httpapi/agent_workspace_test.go
- Modify: backend/internal/storage/agent_workspace_schema.go
- Modify: backend/internal/storage/agent_workspace_schema_test.go
- Modify: backend/internal/storage/migrations_mysql84_test.go

**Interfaces:**
- Consumes: Canvas, SaveCanvas(ctx, owner, id, canvas, expectedRevision, now), and the existing agent_canvases revision row lock.
- Produces: Canvas.Appearance using exactly this JSON contract:

~~~
type Appearance struct {
    EdgeColor       string json:"edgeColor"
    EdgeAnimated    bool   json:"edgeAnimated"
    EdgeWidth       string json:"edgeWidth"
    BackgroundDots  bool   json:"backgroundDots"
    AlignmentGuides bool   json:"alignmentGuides"
}
~~~

NormalizeAppearance must default to coral, true, default, true, true and only permit colors coral/cyan/violet/gold and widths thin/default/thick.

- [ ] **Step 1: Write the failing Go tests.**

Add a store test for EmptyCanvas defaults and extend TestAgentWorkspaceImportsLegacyAndUsesStoreForCanvasAndChat with this request value:

~~~json
"appearance":{"edgeColor":"cyan","edgeAnimated":false,"edgeWidth":"thick","backgroundDots":false,"alignmentGuides":true}
~~~

Assert PATCH and subsequent GET /canvas return the exact cyan value. Add MySQL mock expectations for reading and writing appearance_json.

- [ ] **Step 2: Run the tests and verify they fail.**

Run:

~~~powershell
go test ./backend/internal/agentworkspace ./backend/internal/httpapi ./backend/internal/storage
~~~

Expected: failure because Canvas has no Appearance field and migration 8800003 does not exist.

- [ ] **Step 3: Add the Go model and defaulting.**

Add Appearance to Canvas, create NormalizeAppearance, and call it from EmptyCanvas, normalizeProject, MemoryStore.SaveCanvas, and all MySQL read paths. Invalid or absent legacy values must normalize to defaults without affecting nodes, edges, messages, or executions.

- [ ] **Step 4: Add the append-only migration.**

Append this migration after 8800002:

~~~go
{
    Version: 8800003,
    SQL: []string{
        `ALTER TABLE agent_canvases ADD COLUMN appearance_json JSON NULL AFTER viewport_json`,
    },
    CallbackChecksum: "v88-agent-canvas-appearance-v1",
},
~~~

Update schema tests to protect the historical migration versions/checksums and assert one JSON appearance_json column after repeat migration execution.

- [ ] **Step 5: Add MySQL serialization.**

Marshal NormalizeAppearance(canvas.Appearance) next to viewport_json in project creation and SaveCanvas. Read, unmarshal, normalize, and return it from Canvas. Update SQL mocks and add an invalid-input fallback test. Use the same transaction and revision check; do not add an endpoint or local-only data store.

- [ ] **Step 6: Run verification and commit only these files.**

~~~powershell
go test ./backend/internal/agentworkspace ./backend/internal/httpapi ./backend/internal/storage
go test -tags mysql84 ./backend/internal/storage
git add backend/internal/agentworkspace backend/internal/httpapi/agent_workspace_test.go backend/internal/storage
git commit -m "feat(agent): persist canvas appearance settings"
~~~

Expected: normal tests pass; MySQL 8.4 either passes with its configured disposable DSN or cleanly skips when no DSN is configured.

## Task 2: Create a Shared Frontend Appearance Contract

**Files:**
- Create: frontend/src/user/pages/agent-canvas-appearance.js
- Create: frontend/src/user/pages/agent-canvas-appearance.test.js

**Interfaces:**
- Consumes: canvas.appearance from the current API response.
- Produces: defaultCanvasAppearance, normalizeCanvasAppearance(value), appearanceCssVariables(appearance), and edgeStyleForAppearance(appearance).

- [ ] **Step 1: Write failing Node tests.**

The new test must verify:

~~~js
assert.deepEqual(normalizeCanvasAppearance(), {
  edgeColor: 'coral', edgeAnimated: true, edgeWidth: 'default',
  backgroundDots: true, alignmentGuides: true
});
assert.equal(normalizeCanvasAppearance({ edgeColor: 'unknown' }).edgeColor, 'coral');
assert.equal(normalizeCanvasAppearance({ edgeAnimated: false }).edgeAnimated, false);
assert.deepEqual(edgeStyleForAppearance({ edgeColor: 'cyan', edgeWidth: 'thick' }), {
  stroke: 'var(--agent-edge-cyan)', strokeWidth: 3
});
~~~

- [ ] **Step 2: Run it to verify failure.**

~~~powershell
node --test src/user/pages/agent-canvas-appearance.test.js
~~~

Expected: module-not-found failure.

- [ ] **Step 3: Implement the pure module.**

Export immutable color and width presets. Return a fresh normalized object for every call. Convert width to 1/2/3 pixels and map colors to V88 CSS variables. Make dot opacity zero only when background dots are disabled. Keep animation and guide booleans as booleans, not CSS strings.

- [ ] **Step 4: Run focused tests and commit.**

~~~powershell
node --test src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-commands.test.js
git add frontend/src/user/pages/agent-canvas-appearance.js frontend/src/user/pages/agent-canvas-appearance.test.js
git commit -m "feat(agent): add canvas appearance contract"
~~~

## Task 3: Render Settings and Default Flowing Edges

**Files:**
- Modify: frontend/src/user/components/AgentCanvas.jsx
- Modify: frontend/src/user/pages/AgentPageV2.jsx
- Modify: frontend/src/user/pages/agent-workspace.layout.test.js

**Interfaces:**
- Consumes: the four exports from Task 2.
- Produces: an onAppearanceChange(nextAppearance) callback and custom CanvasEdge rendering through React Flow edgeTypes.

- [ ] **Step 1: Write failing visual-contract tests.**

Require AgentCanvas.jsx to expose settings named 连线颜色, 连线动效, 连线粗细, 背景网点, 辅助线对齐, to pass edgeTypes to ReactFlow, to include edgeAnimated, to include onAppearanceChange, and to preserve appearance in toCanvas after node movement.

- [ ] **Step 2: Run the focused tests and verify failure.**

~~~powershell
node --test src/user/pages/agent-workspace.layout.test.js src/user/pages/agent-canvas-commands.test.js
~~~

Expected: failure because no custom edge or appearance controls exist.

- [ ] **Step 3: Implement one canvas rendering path.**

In AgentCanvas.jsx:

1. Normalize currentCanvas.appearance once and include it in canvas signatures, toCanvas, and emitted saves.
2. Convert every saved edge to type canvas and carry normalized edge data.
3. Add CanvasEdge with BaseEdge and getBezierPath. Draw a static V88-colored base path and an overlay with strokeDasharray="8 10" only when edgeAnimated is true.
4. Keep a static marker arrow at the target, so direction remains visible with reduced motion.
5. Pass edgeTypes={{ canvas: CanvasEdge }} to ReactFlow.
6. Render Background with dot variant only when backgroundDots is true.
7. Set snapToGrid from alignmentGuides and show guide lines only while a node is being dragged.
8. Do not alter chat messages, selected-node context, project switching, or the existing save queue.

- [ ] **Step 4: Add the compact appearance panel.**

Add a SlidersHorizontal icon button with aria-label 画布外观. Use an Ant Design Popover containing:
- four labeled V88 color swatches;
- a Switch for 连线动效, default enabled;
- a Segmented control with 细, 默认, 粗;
- a Switch for 背景网点;
- a Switch for 辅助线对齐.

Every change normalizes the full settings object and emits the current canvas through existing onChange. In AgentPageV2.jsx, pass that emitted canvas directly to canvasSaveQueueRef.current.enqueue(nextCanvas); do not add a second API, state store, or local storage.

- [ ] **Step 5: Run focused frontend tests and commit.**

~~~powershell
node --test src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-workspace.layout.test.js
git add frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/AgentPageV2.jsx frontend/src/user/pages/agent-workspace.layout.test.js
git commit -m "feat(agent): add configurable flowing canvas edges"
~~~

## Task 4: Apply the V88-Dark Creative Canvas Redesign

**Files:**
- Modify: frontend/src/shared/styles/agent-workspace.css
- Modify: frontend/src/user/pages/agent-workspace.layout.test.js

**Interfaces:**
- Consumes: CSS variables and classes from Tasks 2–3: agent-canvas-edge-flow, agent-canvas-appearance-panel, agent-canvas-guide.
- Produces: scoped visual rules only under .agent-page.

- [ ] **Step 1: Write failing CSS assertions.**

Require:
- a low-contrast dotted canvas background;
- agent-canvas-edge-flow animation and keyframes;
- reduced-motion disabling edge animation;
- a constrained appearance panel;
- a narrow-screen rule that keeps the panel inside calc(100vw - 24px).

- [ ] **Step 2: Run the layout test and verify failure.**

~~~powershell
node --test src/user/pages/agent-workspace.layout.test.js
~~~

Expected: new selectors and reduced-motion override are absent.

- [ ] **Step 3: Implement the scoped style system.**

Update only agent canvas CSS:
- Use current dark surfaces, existing coral accent, restrained 1px borders, 8px-or-less corners, and low shadows.
- Make node cards denser with clear title, summary, state, version, selected outline, and visible failure state.
- Give left, center, and bottom toolbars the same compact floating surface and stable button sizes.
- Add agent-canvas-edge-flow with 1.1 second linear stroke-dashoffset animation.
- In prefers-reduced-motion, disable edge animation and panel transitions while retaining static arrows.
- Preserve current mobile toolbar scrolling and move the zoom bar above the conversation drawer.
- Do not add new gradients, oversized cards, reference-product green, or decorative logos.

- [ ] **Step 4: Run frontend verification and commit.**

~~~powershell
node --test src/user/pages/agent-workspace.layout.test.js
npm run build
git add frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-workspace.layout.test.js
git commit -m "feat(agent): restyle canvas workspace"
~~~

## Task 5: Verify the Local Go-Embedded Runtime

**Files:**
- Modify only regression files required by an actual defect found during verification.

**Interfaces:**
- Consumes: committed Tasks 1–4 and the existing local review scripts.
- Produces: verified local behavior at localhost:4001 only.

- [ ] **Step 1: Run final automated suites.**

~~~powershell
go test ./backend/internal/agentworkspace ./backend/internal/httpapi ./backend/internal/storage
node --test src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-commands.test.js src/user/pages/agent-workspace.layout.test.js
~~~

- [ ] **Step 2: Rebuild and restart the local Go review runtime.**

~~~powershell
./scripts/build-go-local.ps1
./scripts/start-local-go-review.ps1 -UseYcmMysql
./scripts/check-local-go.ps1
~~~

Do not use Node release scripts, Docker rebuilds, or any public deployment script.

- [ ] **Step 3: Perform authenticated browser acceptance on an existing project.**

At the existing local Agent project:
1. Verify conversations, nodes, edges, and execution records still appear.
2. Verify default animated coral flow and static direction arrow.
3. Change each of five settings, refresh, and verify persistence.
4. Open a different project and verify settings are independent.
5. Verify selection, grouping, alignment, undo, redo, zoom, save, and collapsed chat still work.
6. Verify narrow desktop/mobile layouts do not cover the message input.

Use browser screenshots as acceptance evidence only. Do not create a temporary HTML mockup.

- [ ] **Step 4: Repair any found defect test-first.**

Add the smallest failing Go or Node regression test, verify the failure, make the narrowest fix, rerun the affected suite and build, then commit only the defect files:

For a persistence defect, stage only the changed files under backend/internal/agentworkspace, backend/internal/httpapi/agent_workspace_test.go, and backend/internal/storage. For a rendering defect, stage only the changed files among frontend/src/user/components/AgentCanvas.jsx, frontend/src/user/pages/agent-canvas-appearance.js, frontend/src/user/pages/agent-workspace.layout.test.js, and frontend/src/shared/styles/agent-workspace.css. Then use:

~~~powershell
git commit -m "fix(agent): correct canvas appearance defect"
git rev-parse --short HEAD
~~~

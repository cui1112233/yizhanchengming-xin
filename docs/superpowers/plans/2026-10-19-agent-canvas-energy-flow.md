# Agent Canvas Energy Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make only animated Agent canvas connections read as moderately visible electric energy flowing toward their existing arrows.

**Architecture:** Keep the existing `CanvasEdge` base path untouched. Add a derived style for its already-present animated overlay so the overlay is one pixel wider than the static connection. Scoped CSS turns that overlay into spaced, glowing energy pulses; the current `edgeAnimated` conditional and reduced-motion rule continue to remove it when motion is disabled.

**Tech Stack:** React 18, `@xyflow/react`, CSS, Node test runner, Vite, Go embed.

## Global Constraints

- Local V88 only. Do not publish publicly.
- Modify only the custom canvas edge rendering, scoped canvas CSS, and its source-level layout test.
- Keep nodes, groups, toolbars, popovers, canvas background, APIs, migrations, and appearance data unchanged.
- Preserve the existing coral default and all user-selected connection colors and widths.
- Keep static base lines and target arrows visible when animated flow is disabled or reduced motion is requested.

---

### Task 1: Render Restrained Energy Pulses on Animated Edges

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Consumes: `CanvasEdge`, `edgeStyleForAppearance(appearance)`, `appearance.edgeAnimated`, and `.agent-canvas-edge-flow`.
- Produces: `energyStyle`, used only by the animated overlay path. It must keep the selected `stroke` and set `strokeWidth` to `Number(style.strokeWidth) + 1`.

- [ ] **Step 1: Write the failing source-level regression test.**

Add these assertions to the existing `画布外观包含轻量动效和窄屏安全规则` test:

```js
assert.match(canvasSource, /const energyStyle = \{ \.\.\.style, strokeWidth: Number\(style\.strokeWidth\) \+ 1 \}/);
assert.match(css, /\.agent-page \.agent-canvas-edge-flow\s*\{[^}]*stroke-dasharray:\s*10 26[^}]*filter:\s*drop-shadow/);
assert.match(css, /@keyframes agent-canvas-edge-flow\s*\{\s*to\s*\{\s*stroke-dashoffset:\s*-72/);
```

- [ ] **Step 2: Run the focused test and verify it fails.**

Run:

```powershell
node --test src/user/pages/agent-workspace.layout.test.js
```

Expected: the new assertions fail because `energyStyle`, the 10/26 pulse spacing, glow filter, and `-72` dash offset do not exist yet.

- [ ] **Step 3: Implement the minimal edge rendering change.**

In `CanvasEdge`, derive the animated overlay style without mutating `style`:

```jsx
const energyStyle = { ...style, strokeWidth: Number(style.strokeWidth) + 1 };
```

Keep the base path unchanged and pass `energyStyle` only to the conditional animated `<path>`:

```jsx
{appearance.edgeAnimated ? <path d={edgePath} className="agent-canvas-edge-flow" style={energyStyle} /> : null}
```

- [ ] **Step 4: Implement scoped energy-flow CSS.**

Replace the current flow properties with these values while retaining `fill`,
`stroke`, pointer safety, and the existing 1.1-second linear timing:

```css
stroke-dasharray: 10 26;
stroke-linecap: round;
filter: drop-shadow(0 0 3px var(--agent-edge-color));
opacity: 0.96;
animation: agent-canvas-edge-flow 1.1s linear infinite;

@keyframes agent-canvas-edge-flow { to { stroke-dashoffset: -72; } }
```

Do not change `.agent-canvas-edge-flow` behavior inside `prefers-reduced-motion`.

- [ ] **Step 5: Run focused checks and production build.**

Run:

```powershell
node --test src/user/pages/agent-workspace.layout.test.js src/user/pages/agent-canvas-appearance.test.js src/user/pages/agent-canvas-state.test.js src/user/pages/agent-canvas-commands.test.js
npm run build
```

Expected: all focused tests pass and Vite exits successfully.

- [ ] **Step 6: Commit only the implementation and regression files.**

```powershell
git add frontend/src/user/components/AgentCanvas.jsx frontend/src/shared/styles/agent-workspace.css frontend/src/user/pages/agent-workspace.layout.test.js
git commit -m "feat(agent): add energy flow to canvas edges"
```

### Task 2: Verify the Local Go-Embedded Page

**Files:**
- Modify only a regression file required by a verified defect.

**Interfaces:**
- Consumes: the committed Task 1 source and the existing local Go review scripts.
- Produces: visible animated energy pulses at `http://localhost:4001` without a public release.

- [ ] **Step 1: Rebuild and restart the local Go review runtime.**

Run:

```powershell
./scripts/build-go-local.ps1
./scripts/start-local-go-review.ps1 -UseYcmMysql
./scripts/check-local-go.ps1
```

If port `4001` is occupied, verify the listener path is the local
`backend/qiantie-local.exe`, stop only that exact process, and restart the
review launcher. Do not touch any public or unrelated process.

- [ ] **Step 2: Inspect the existing authenticated canvas project.**

At `http://localhost:4001/agent/canvas?taskId=agent-14c55b4999fd3e9b7854b5ab`, verify:

1. Existing nodes, messages, and execution records still appear.
2. Default coral animated edges have visible short glowing pulses moving toward their arrows.
3. Switching off `连线动效` removes the pulses but preserves static lines and arrows.
4. Switching it back on restores pulses, then leave the existing project with its original setting enabled.

- [ ] **Step 3: Repair only a demonstrated defect.**

For any defect, first add a focused failing assertion to
`agent-workspace.layout.test.js`, rerun the focused test, make the smallest
source or CSS correction, rerun Task 1 checks and build, then commit only the
affected canvas files with:

```powershell
git commit -m "fix(agent): correct canvas energy flow"
```

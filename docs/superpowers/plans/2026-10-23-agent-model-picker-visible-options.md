# Agent Model Picker Visible Options Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Agent Canvas model dropdown options fully readable without widening the compact control inside a node.

**Architecture:** `AgentModelPicker` already produces a two-line option with a 260px minimum content width. Ant Design currently constrains the popup to the narrow Select trigger, so the popup will opt out of trigger-width matching while retaining the existing choice construction and change callback.

**Tech Stack:** React, Ant Design Select, Node built-in test runner, Go embedded frontend local runtime.

## Global Constraints

- Keep the node trigger compact; only the opened popup may expand.
- Do not change provider/model synchronization, model filtering, executor bindings, or save behavior.
- Preserve the existing unavailable-model and no-model messages.
- Verify both source behavior and the served `http://localhost:4001` UI.

---

### Task 1: Allow the model popup to use its content width

**Files:**
- Modify: `frontend/src/user/components/AgentModelPicker.jsx:17-40`
- Modify: `frontend/src/user/pages/agent-model-picker.source.test.js:7-15`

**Interfaces:**
- Consumes: Ant Design `Select` prop `popupMatchSelectWidth`.
- Produces: an `AgentModelPicker` popup whose width is not constrained by the Select trigger.

- [ ] **Step 1: Write the failing test**

Add this assertion to the first model-picker source test:

```js
assert.match(source, /popupMatchSelectWidth=\{false\}/);
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```powershell
node --test frontend/src/user/pages/agent-model-picker.source.test.js
```

Expected: the first test fails because `AgentModelPicker` does not yet configure popup width behavior.

- [ ] **Step 3: Write minimal implementation**

Add the Select property alongside `showSearch`:

```jsx
popupMatchSelectWidth={false}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```powershell
node --test frontend/src/user/pages/agent-model-picker.source.test.js
```

Expected: both model-picker source tests pass.

- [ ] **Step 5: Build and verify the local embedded runtime**

Run:

```powershell
npm --prefix frontend run build
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql
.\scripts\check-local-go.ps1
```

Expected: the Go runtime serves the rebuilt frontend on port 4001.

- [ ] **Step 6: Browser acceptance**

Open the Agent Canvas video node, expand the model picker, and confirm that model names plus the provider/capability line are visible and that choosing a model still updates the node.

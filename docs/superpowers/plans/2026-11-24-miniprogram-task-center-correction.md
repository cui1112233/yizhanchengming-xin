# Mini Program Task Center Correction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct the WeChat mini-program task center so its counters, progress, retry states, icons, hero art, and compact phone layout all come from real task data and match the approved reference.

**Architecture:** Keep `taskSummary.ts` as the single task read model, but teach it the backend's real status vocabulary and hydrate batch summaries from each batch runtime summary before creating visible task entries. Keep visual selection in `taskVisuals.ts`, using explicit module and task-kind metadata before any title fallback. Render the page with task-specific class names so old shared workspace styles cannot override the four-column mobile row.

**Tech Stack:** Taro 4.2.1, React 18, TypeScript 5.7, SCSS/WXSS, Vitest 3.2, WeChat mini-program build.

## Global Constraints

- Work directly on branch `v88`; do not create another branch.
- Modify only the mini-program task center, its focused tests, generated task asset, and this plan.
- Do not upload, deploy, push, or call paid production services.
- Percentages must be calculated from real `completed / total` values; no percentage may be baked into an image.
- The top counters and filters are `进行中 / 需重试 / 已完成`.
- Reuse the approved function-page assets for task icons; never use one generic batch icon for unrelated functions.
- Preserve all unrelated dirty files.

---

### Task 1: Real Status And Runtime Summary Read Model

**Files:**
- Modify: `miniprogram/src/services/taskSummary.ts`
- Modify: `miniprogram/src/services/taskSummary.test.ts`
- Modify: `miniprogram/src/services/productionTaskSummary.test.ts`

**Interfaces:**
- Consumes: `TaskRequest(path: string): Promise<unknown>` and the existing five API namespaces.
- Produces: `TaskStatus = "running" | "retry" | "completed" | "attention" | "unknown"`, `loadTaskSummary()`, `taskCounts()`, and real progress metadata for the page.

- [ ] **Step 1: Write failing status vocabulary tests**

Add literal expectations proving `waiting_retry` and `failed` become `retry`, active queue values become `running`, completed values become `completed`, and pending configuration/login/submit states preserve a readable detail label rather than becoming a fake syncing state.

```ts
expect(taskStatusFromServer("waiting_retry")).toBe("retry");
expect(taskStatusFromServer("scheduled_waiting")).toBe("running");
expect(taskStatusFromServer("completed")).toBe("completed");
expect(taskStatusDetail("pending_submit")).toBe("待提交");
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `npm test -- --run src/services/taskSummary.test.ts src/services/productionTaskSummary.test.ts`

Expected: FAIL because `retry`, `taskStatusDetail`, and runtime hydration do not exist yet.

- [ ] **Step 3: Add runtime-summary fixtures that reject batch-library records as tasks**

Use a batch summary containing `{ id: "b", title: "Batch" }` and a runtime response containing real automation/production/stage data. Assert the visible batch entry comes from the runtime response, carries the true status and progress, and does not show an unknown library row.

- [ ] **Step 4: Implement the minimal read-model changes**

Add the backend status groups, readable detail mapping, batch runtime hydration, and optional `moduleKey`, `retryable`, and `statusDetail` fields. Calculate counts only from normalized task rows. Keep a source-level error when a batch runtime call fails instead of inventing a zero-state task.

- [ ] **Step 5: Run focused tests and verify GREEN**

Run: `npm test -- --run src/services/taskSummary.test.ts src/services/productionTaskSummary.test.ts`

Expected: PASS.

- [ ] **Step 6: Commit the read model**

```powershell
git add miniprogram/src/services/taskSummary.ts miniprogram/src/services/taskSummary.test.ts miniprogram/src/services/productionTaskSummary.test.ts
git commit -m "fix(miniprogram): use real task center states"
```

### Task 2: Function-Specific Icon Mapping And Real Right-Side States

**Files:**
- Modify: `miniprogram/src/pages/tasks/taskVisuals.ts`
- Modify: `miniprogram/src/pages/tasks/taskVisuals.test.ts`
- Modify: `miniprogram/src/pages/tasks/index.tsx`
- Modify: `miniprogram/src/pages/tasks/interaction.test.tsx`

**Interfaces:**
- Consumes: normalized `AggregatedTask` fields from Task 1 and existing function assets.
- Produces: `taskVisualAsset(entry)`, visible status copy, progress, retry/view action, and filter behavior.

- [ ] **Step 1: Write failing icon-priority tests**

Assert explicit module/source ownership wins over ambiguous title text, while a real task kind/stage can refine the icon inside that module.

```ts
expect(taskVisualAsset(task({ source: "script", moduleKey: "script", title: "批量生成" }))).toBe(scriptAsset);
expect(taskVisualAsset(task({ source: "shuihuo", kind: "storyboard" }))).toBe(storyboardAsset);
```

- [ ] **Step 2: Write failing component behavior tests**

Render one running task with `2 / 5`, one retry task, one completed task, and one `pending_submit` task. Assert `40%`, `需重试`, `已完成`, and `待提交` are rendered, and `状态同步中` is absent.

- [ ] **Step 3: Run the focused tests and verify RED**

Run: `npm test -- --run src/pages/tasks/taskVisuals.test.ts src/pages/tasks/interaction.test.tsx`

Expected: FAIL because the old page still uses `attention`, shared classes, and title-first icon guessing.

- [ ] **Step 4: Implement icon and status rendering**

Use the fixed priority `moduleKey/source -> kind/taskType -> stage -> explicit title -> module fallback`. Replace filters and counters with `需重试`. Render retry actions only when `retryable === true`; otherwise route to the owning detail page. Remove `.studio-task-record` and `.studio-task-action` from task rows.

- [ ] **Step 5: Run focused tests and verify GREEN**

Run: `npm test -- --run src/pages/tasks/taskVisuals.test.ts src/pages/tasks/interaction.test.tsx`

Expected: PASS.

- [ ] **Step 6: Commit the visible behavior**

```powershell
git add miniprogram/src/pages/tasks/taskVisuals.ts miniprogram/src/pages/tasks/taskVisuals.test.ts miniprogram/src/pages/tasks/index.tsx miniprogram/src/pages/tasks/interaction.test.tsx
git commit -m "fix(miniprogram): match task icons and states"
```

### Task 3: Correct Hero Asset And Collision-Free Phone Layout

**Files:**
- Create: `miniprogram/src/assets/tasks/task-center-runtime-hero-v2.png`
- Modify: `miniprogram/src/pages/tasks/index.tsx`
- Modify: `miniprogram/src/pages/tasks/tasks-studio.scss`
- Modify: `miniprogram/src/pages/tasks/interaction.test.tsx`

**Interfaces:**
- Consumes: the selected task-center reference and page data from Tasks 1-2.
- Produces: a text-free, number-free 3D hero illustration and responsive four-column rows.

- [ ] **Step 1: Write failing layout assertions**

Assert task rows do not carry the two conflicting shared class names, the hero progress text is data-driven, and the stylesheet uses a four-column grid with dedicated icon, copy, state, and arrow areas at 360px, 390px, and 430px.

- [ ] **Step 2: Run the interaction test and verify RED**

Run: `npm test -- --run src/pages/tasks/interaction.test.tsx`

Expected: FAIL because the current row is flex/absolute and the hero has no dynamic aggregate progress.

- [ ] **Step 3: Generate and inspect the hero asset**

Generate one transparent 3D illustration containing a task checklist card, small analytics card, circular progress ring without numbers, and a film strip. Match the cyan-blue-purple glossy style. Avoid text, numbers, logos, watermarks, characters, and backgrounds. Save the inspected final PNG to the exact project path above.

- [ ] **Step 4: Implement the compact layout**

Use separate background and hero layers. Overlay aggregate progress only when real totals exist. Replace absolute row icons with a grid, keep 13-15px phone gutters, 72-80px rows, a 42px segmented control, and a safe bottom inset. Ensure the status column cannot overlap the title.

- [ ] **Step 5: Run focused tests and verify GREEN**

Run: `npm test -- --run src/pages/tasks/interaction.test.tsx src/pages/tasks/taskVisuals.test.ts`

Expected: PASS.

- [ ] **Step 6: Commit the asset and layout**

```powershell
git add miniprogram/src/assets/tasks/task-center-runtime-hero-v2.png miniprogram/src/pages/tasks/index.tsx miniprogram/src/pages/tasks/tasks-studio.scss miniprogram/src/pages/tasks/interaction.test.tsx
git commit -m "feat(miniprogram): rebuild compact task center layout"
```

### Task 4: Build And Visible WeChat Acceptance

**Files:**
- Modify: `design-qa.md`
- Create: `miniprogram/.qa/task-center-corrected.png`

**Interfaces:**
- Consumes: completed Tasks 1-3 and the WeChat developer-tool runtime.
- Produces: focused test proof, production build proof, same-viewport screenshot, and a passing visual QA record.

- [ ] **Step 1: Run all task-center tests**

Run: `npm test -- --run src/services/taskSummary.test.ts src/services/productionTaskSummary.test.ts src/pages/tasks/taskVisuals.test.ts src/pages/tasks/interaction.test.tsx src/services/taskPageState.test.ts`

Expected: PASS with no task-center warnings.

- [ ] **Step 2: Run the complete mini-program test suite**

Run: `npm test`

Expected: PASS.

- [ ] **Step 3: Build the WeChat package**

Run: `npm run build:weapp:sandbox`

Expected: successful Taro compile with no WXSS syntax error or black screen.

- [ ] **Step 4: Verify in WeChat developer tools**

Open the tasks tab, refresh real data, test all four filters, open one task, and capture the same 390px state. Confirm counters, percentage, retry state, completed state, function-specific icons, background, and bottom safe area are visible and non-overlapping.

- [ ] **Step 5: Complete visual QA**

Compare the approved reference and the new runtime capture side by side. Write `design-qa.md` with `final result: passed`; fix every P0/P1/P2 before proceeding.

- [ ] **Step 6: Commit acceptance evidence**

```powershell
git add design-qa.md miniprogram/.qa/task-center-corrected.png
git commit -m "test(miniprogram): verify corrected task center"
```

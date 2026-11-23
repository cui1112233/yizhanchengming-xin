# Mini-program Tasks Visual Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把小程序第三项“任务”改成与 `03-tasks.png` 对齐、只展示真实跨功能任务数据的移动端任务中心。

**Architecture:** `services/taskSummary.ts` 负责归一化真实状态、统计和筛选；`pages/tasks/index.tsx` 只组装页面状态与交互；新增任务页专属 SCSS，避免继续把视觉规则塞进全局样式。现有真实接口和跳转路由保持不变。

**Tech Stack:** Taro 4、React、TypeScript、SCSS、Vitest、微信开发者工具。

## Global Constraints

- 不修改后端接口、数据库或任务生产逻辑。
- 不伪造任务、进度、时间、模型或项目名称。
- “待确认”只匹配明确人工确认或审核状态，普通排队仍归“进行中”。
- 390px 基准左右留白 15px，底部保留四项导航安全区。
- 不上传、不部署，不改无关文件。

---

### Task 1: 真实任务状态、统计和筛选

**Files:**
- Modify: `miniprogram/src/services/taskSummary.ts`
- Modify: `miniprogram/src/services/taskSummary.test.ts`

**Interfaces:**
- Produces: `TaskStatus = "running" | "attention" | "failed" | "completed" | "unknown"`
- Produces: `TaskFilter = "全部" | "进行中" | "待确认" | "已完成"`
- Produces: `taskCounts(entries)` returning `{ running, attention, completed }`

- [ ] **Step 1: Write failing tests**

Add assertions that `waiting-review`, `pending_confirmation`, `accepted_pending`, `待确认`, and `待审核` normalize to `attention`; `queued` and `pending` remain `running`; `filterTaskEntries(..., "待确认")` and `taskCounts` return matching real records.

- [ ] **Step 2: Run the focused test and confirm failure**

Run: `npm test -- src/services/taskSummary.test.ts`

Expected: FAIL because `attention`, `待确认`, and `taskCounts` do not exist.

- [ ] **Step 3: Implement the minimal normalization**

Add an explicit attention status set before the running set, extend filter mapping, and add a single-pass count helper. Do not infer attention from titles or missing data.

- [ ] **Step 4: Run the focused test**

Run: `npm test -- src/services/taskSummary.test.ts`

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add -- miniprogram/src/services/taskSummary.ts miniprogram/src/services/taskSummary.test.ts
git commit -m "feat: classify actionable miniprogram tasks"
```

### Task 2: 任务中心页面结构和交互

**Files:**
- Modify: `miniprogram/src/pages/tasks/index.tsx`
- Modify: `miniprogram/src/pages/tasks/interaction.test.tsx`
- Create: `miniprogram/src/pages/tasks/tasks-studio.scss`
- Modify: `miniprogram/src/ui/liquidGlassPages.test.ts`

**Interfaces:**
- Consumes: `taskCounts`, `filterTaskEntries`, `taskDestination`
- Produces: `.task-studio-*` 页面层级和现有真实跳转交互

- [ ] **Step 1: Write failing page tests**

Assert the page renders a three-column real summary, four filters including “待确认”, one shared list surface, status-specific row classes, conditional progress, and no per-row large glass card. Keep the missing-ID and encoded destination tests.

- [ ] **Step 2: Run focused tests and confirm failure**

Run: `npm test -- src/pages/tasks/interaction.test.tsx src/ui/liquidGlassPages.test.ts`

Expected: FAIL because the new summary and page classes are absent.

- [ ] **Step 3: Rebuild the page hierarchy**

Use the existing brand header and real summary loader. Add compact hero copy, three live counters, segmented filters, one glass list, status/progress area, refresh interaction, partial-source warning, and existing navigation destinations. Import `tasks-studio.scss` from the page.

- [ ] **Step 4: Add mobile styles**

Implement 15px gutters, a roughly 108px count card, 52px segmented control, 98px rows, restrained cyan/purple glass edges, 12–16px row typography, and safe bottom padding. Use existing `FUNCTION_VISUALS` raster assets and `AppIcon`; do not draw replacement icons with CSS.

- [ ] **Step 5: Run focused tests**

Run: `npm test -- src/services/taskSummary.test.ts src/pages/tasks/interaction.test.tsx src/ui/liquidGlassPages.test.ts`

Expected: PASS.

- [ ] **Step 6: Commit**

```powershell
git add -- miniprogram/src/pages/tasks/index.tsx miniprogram/src/pages/tasks/interaction.test.tsx miniprogram/src/pages/tasks/tasks-studio.scss miniprogram/src/ui/liquidGlassPages.test.ts
git commit -m "feat: redesign miniprogram task center"
```

### Task 3: 微信运行验收和回归

**Files:**
- Modify: `miniprogram/design-qa.md`
- Create: `miniprogram/qa-tasks-implementation.png`
- Create: `miniprogram/qa-tasks-reference.png`
- Create: `miniprogram/qa-tasks-comparison.png`

**Interfaces:**
- Consumes: `docs/design-references/miniprogram-functions-first/03-tasks.png`
- Produces: 同状态、同视口的视觉验收证据

- [ ] **Step 1: Run focused and full verification**

Run:

```powershell
npm test -- src/services/taskSummary.test.ts src/pages/tasks/interaction.test.tsx src/ui/liquidGlassPages.test.ts
npm run build:weapp
npm test
```

Expected: all focused tests pass, production build prints `Compiled successfully`, then the full suite passes.

- [ ] **Step 2: Open the real mini-program and capture the task page**

Open the existing project in WeChat Developer Tools, switch to “任务”, confirm no business console error, and capture only the simulator viewport. Do not substitute a static HTML preview.

- [ ] **Step 3: Compare equal-size images**

Normalize `03-tasks.png` and the simulator capture to the same viewport, save the side-by-side comparison, and inspect typography, gutters, card proportions, status colors, image sharpness, bottom navigation, loading/empty/content state differences.

- [ ] **Step 4: Fix all P0/P1/P2 differences and repeat capture**

Keep only P3 polish in `design-qa.md`; if the capture is blocked, mark the result blocked instead of claiming parity.

- [ ] **Step 5: Commit QA evidence**

```powershell
git add -- miniprogram/design-qa.md miniprogram/qa-tasks-implementation.png miniprogram/qa-tasks-reference.png miniprogram/qa-tasks-comparison.png
git commit -m "test: verify miniprogram task center parity"
```

### Task 4: 任务图标与真实类型对应

**Files:**
- Create: `miniprogram/src/pages/tasks/taskVisuals.ts`
- Create: `miniprogram/src/pages/tasks/taskVisuals.test.ts`
- Modify: `miniprogram/src/services/taskSummary.ts`
- Modify: `miniprogram/src/services/taskSummary.test.ts`
- Modify: `miniprogram/src/pages/tasks/index.tsx`
- Modify: `miniprogram/src/pages/tasks/interaction.test.tsx`

**Interfaces:**
- Produces: `AggregatedTask.visualHint?: string`
- Produces: `taskVisualAsset(entry: AggregatedTask): string`

- [ ] **Step 1: Write failing tests**

Assert that server `kind/type/stage/taskType` is preserved as `visualHint`, that explicit title keywords select the matching image, and that each source has a different fallback asset.

- [ ] **Step 2: Run focused tests and confirm failure**

Run: `npm test -- --run src/pages/tasks/taskVisuals.test.ts src/services/taskSummary.test.ts src/pages/tasks/interaction.test.tsx`

Expected: FAIL because `visualHint` and `taskVisualAsset` do not exist and the page still reads one source image map.

- [ ] **Step 3: Implement real visual mapping**

Preserve the server visual hint without changing status normalization. Map image, writing, novel, batch, comic/storyboard, video/export and agent categories to the existing transparent 3D assets; fall back by source only when no explicit category exists.

- [ ] **Step 4: Replace status and hero art**

Use one task hero asset plus separate running, attention and completed status assets. Keep all task numbers and labels as live UI text.

- [ ] **Step 5: Verify and commit**

Run focused tests and `npm run build:weapp:sandbox`, then commit only the task-page files and new assets.

# Mini-program Functions-First Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the four main mini-program tabs to match the approved visual references one-for-one while making 功能 the first tab and 工作台 the execution hub for the selected function.

**Architecture:** Keep module metadata and real routes centralized in `config/modules.ts`. 功能 writes the selected `ModuleKey` and switches to 工作台; 工作台 reads that key, derives a truthful current execution from the existing aggregated task service, and opens the existing real module route only when the user starts or continues. 任务 and 我的 keep their existing APIs and behavior but receive page-local presentation layers matching the approved references.

**Tech Stack:** Taro 4.2.1, React 18.3.1, TypeScript 5.7, SCSS/WXSS, Vitest, WeChat custom tab bar, existing Go-backed APIs.

## Global Constraints

- The visual contract is exactly the four files under `docs/design-references/miniprogram-functions-first/`.
- `01-functions-home.png` is the user-supplied 功能 reference; do not substitute the later “创作功能” draft.
- The only deliberate visual deviation is bottom navigation order: `功能、工作台、任务、我的` on every page.
- Use only the seven real top-level modules already present in `config/modules.ts`; each appears once on 功能.
- Lock the 功能 presentation map to `script→私人定制`, `fetch→小说创作`, `batch→批量测书`, `shuihuo→水货生产`, `agent→AI-CUT` (subtitle `Agent 工作区`), `comic→漫剧创作`, and `toolkit→会员与工具`; this preserves the approved seven-card composition without inventing an eighth route.
- 功能 selects; 工作台 executes; 任务 lists all records; 我的 manages account, assets, usage, and preferences.
- Never invent progress, counts, quota, names, dates, tasks, models, or project titles when the API does not return them; show `—`, an empty state, or a loading/error state.
- Do not change backend APIs, add frontend prompt strings, add a dependency, upload, deploy, or push.
- Preserve unrelated dirty work. Stage and commit only the files named in each task.
- Verification covers pages, states, interactions, and pixels separately; a passing build alone is not acceptance.

---

## File Structure

- `miniprogram/src/app.config.ts`: native tab order and labels.
- `miniprogram/src/ui/studioNavigation.ts`: custom navigation order and selected route.
- `miniprogram/src/custom-tab-bar/index.scss`: approved full-width dark glass navigation material.
- `miniprogram/src/config/modules.ts`: single source of module labels, groups, icons, and routes; no new modules.
- `miniprogram/src/assets/studio-v3/`: transparent 3D illustration assets derived from the approved visual language.
- `miniprogram/src/pages/functions/index.tsx`: approved 功能 composition and selection behavior.
- `miniprogram/src/pages/functions/gallery.scss`: approved 功能 geometry and responsive styling.
- `miniprogram/src/services/workbenchView.ts`: pure derivation of current module and current task display data.
- `miniprogram/src/pages/workspace/index.tsx`: selected-function execution hub.
- `miniprogram/src/pages/workspace/studio.scss`: approved 工作台 composition.
- `miniprogram/src/pages/tasks/index.tsx`: task summary, filters, and grouped list.
- `miniprogram/src/pages/tasks/index.scss`: approved 任务 styling, kept page-local.
- `miniprogram/src/pages/profile/index.tsx`: profile/usage/settings composition using real data.
- `miniprogram/src/pages/profile/index.scss`: approved 我的 styling, kept page-local.
- Existing focused test files beside those modules verify behavior and truthful fallbacks.

---

### Task 1: Make 功能 the First Native and Custom Tab

**Files:**
- Modify: `miniprogram/src/app.config.ts`
- Modify: `miniprogram/src/app.config.test.ts`
- Modify: `miniprogram/src/ui/studioNavigation.ts`
- Modify: `miniprogram/src/ui/studioNavigation.test.ts`
- Modify: `miniprogram/src/custom-tab-bar/index.scss`

**Interfaces:**
- Produces: `STUDIO_TABS` ordered as functions, workspace, tasks, profile.
- Produces: native `tabBar.list` in the same order.
- Preserves: `switchStudioTab(path: string): Promise<void>` behavior.

- [ ] **Step 1: Write failing navigation-order tests**

Add explicit order assertions:

```ts
expect(appConfig.tabBar?.list.map(item => item.text)).toEqual(["功能", "工作台", "任务", "我的"]);
expect(appConfig.tabBar?.list.map(item => item.pagePath)).toEqual([
  "pages/functions/index",
  "pages/workspace/index",
  "pages/tasks/index",
  "pages/profile/index",
]);
expect(STUDIO_TABS.map(item => item.label)).toEqual(["功能", "工作台", "任务", "我的"]);
expect(STUDIO_TABS[0].path).toBe("/pages/functions/index");
```

- [ ] **Step 2: Run tests and verify the old order fails**

Run:

```powershell
npm test -- --run src/app.config.test.ts src/ui/studioNavigation.test.ts
```

Expected: FAIL because 工作台 is currently first.

- [ ] **Step 3: Reorder both navigation sources**

Set `STUDIO_TABS` to:

```ts
[
  { path: '/pages/functions/index', label: '功能', icon: 'grid-2x2' },
  { path: '/pages/workspace/index', label: '工作台', icon: 'box' },
  { path: '/pages/tasks/index', label: '任务', icon: 'folder-clock' },
  { path: '/pages/profile/index', label: '我的', icon: 'user-round' },
]
```

Use the closest names already supported by `AppIcon`; if `grid-2x2` or `box` is absent, extend the existing `AppIconName` mapping with those two Lucide paths instead of changing icon libraries.

Reorder `app.config.ts` to the same paths and labels. Keep the existing native PNG icon paths paired with their logical tab, moving entries rather than swapping labels.

- [ ] **Step 4: Match the approved navigation material**

Update `custom-tab-bar/index.scss` so the host is edge-to-edge like the references: `padding:0`, square top corners, `min-height:72PX`, dark navy fill, 24PX icons, 12PX labels, and a cyan underline/glow only beneath the selected tab. Preserve safe-area bottom padding and 44PX minimum targets.

- [ ] **Step 5: Run navigation tests**

Run the command from Step 2. Expected: all tests PASS.

- [ ] **Step 6: Commit navigation**

```powershell
git add -- miniprogram/src/app.config.ts miniprogram/src/app.config.test.ts miniprogram/src/ui/studioNavigation.ts miniprogram/src/ui/studioNavigation.test.ts miniprogram/src/custom-tab-bar/index.scss
git commit -m "feat: make functions the first mini-program tab"
```

---

### Task 2: Build the Approved 3D Illustration Asset Set

**Files:**
- Create: `miniprogram/src/assets/studio-v3/dragon-pair.webp`
- Create: `miniprogram/src/assets/studio-v3/story-book.webp`
- Create: `miniprogram/src/assets/studio-v3/novel-book.webp`
- Create: `miniprogram/src/assets/studio-v3/batch-books.webp`
- Create: `miniprogram/src/assets/studio-v3/comic-pages.webp`
- Create: `miniprogram/src/assets/studio-v3/clapper-play.webp`
- Create: `miniprogram/src/assets/studio-v3/batch-video.webp`
- Create: `miniprogram/src/assets/studio-v3/agent-bot.webp`
- Create: `miniprogram/src/assets/studio-v3/crown-tools.webp`
- Create: `miniprogram/src/assets/studio-v3/quota-crown.webp`
- Create: `miniprogram/src/assets/studio-v3/profile-dragons.webp`
- Create: `miniprogram/src/assets/studio-v3/comic-workspace.webp`

**Interfaces:**
- Produces: transparent WebP assets imported by page components.
- Preserves: approved blue/purple/cyan 3D illustration style without embedding UI text into images.

- [ ] **Step 1: Reuse the three already-approved transparent sources**

Copy the existing `assets/functions/dragon-pair.webp`, `story-studio.webp`, and `comic-studio.webp` into the new directory as `dragon-pair.webp`, `story-book.webp`, and `comic-workspace.webp`.

- [ ] **Step 2: Generate the remaining transparent objects**

Use the built-in image generator once per asset. Every prompt must include:

```text
Transparent-background isolated 3D UI illustration for the 一战晟铭 mobile creative suite. Match the approved reference: rounded toy-like modeling, navy/cyan/purple glass, soft rim light, restrained neon cyan glow, no text, no logo, no card background, centered object, clean alpha edges.
```

Append the exact subject for each file: purple novel book with star bookmark; stacked teal/purple books; comic page panel with stylus; cyan clapperboard with play symbol; layered video cards; friendly cyan AI robot with message panel; gold crown with purple creator toolbox; crown resting on a soft cloud; paired pink and blue dragon avatar crop.

- [ ] **Step 3: Convert and size assets**

Use ffmpeg to produce lossless-alpha WebP files. Long edge targets: hero/feature assets 768px; compact icons 384px; profile avatar 512px. Do not upscale an input below its target.

- [ ] **Step 4: Verify transparency and package size**

Run:

```powershell
Get-ChildItem src/assets/studio-v3/*.webp | Select-Object Name,Length
ffprobe -v error -show_entries stream=pix_fmt,width,height -of compact src/assets/studio-v3/*.webp
```

Expected: every file uses an alpha-capable pixel format; combined size remains comfortably below the WeChat main-package limit after the development build.

- [ ] **Step 5: Visually inspect every asset**

Open each WebP and reject any asset containing baked-in words, a rectangular background, clipped edges, an extra character, or a visual style that does not match the four visual references.

- [ ] **Step 6: Commit assets**

```powershell
git add -- miniprogram/src/assets/studio-v3
git commit -m "feat: add approved mini-program illustration assets"
```

---

### Task 3: Rebuild 功能 from the User-Supplied Reference

**Files:**
- Modify: `miniprogram/src/pages/functions/index.tsx`
- Modify: `miniprogram/src/pages/functions/gallery.scss`
- Modify: `miniprogram/src/pages/functions/functions.interaction.test.tsx`
- Modify: `miniprogram/src/config/modules.ts`
- Modify: `miniprogram/src/config/modules.test.ts`

**Interfaces:**
- Consumes: `STUDIO_TABS[0]`, `ModuleKey`, `getModule`, and the illustration assets.
- Produces: `chooseModule(key: ModuleKey)` behavior that stores the key and calls `Taro.switchTab({url:'/pages/workspace/index'})`.
- Produces: seven visible cards exactly once.

- [ ] **Step 1: Write failing function-selection tests**

Mock `Taro.switchTab` and assert:

```ts
expect(labels).toEqual(expect.arrayContaining([
  '私人定制', '小说创作', '批量测书', '水货生产',
  'AI-CUT', '漫剧创作', '会员与工具',
]));
expect(new Set(renderedKeys).size).toBe(7);
clickTool('批量测书');
expect(taro.setStorageSync).toHaveBeenCalledWith('yzcm.mobile.activeModule', 'fetch');
expect(taro.switchTab).toHaveBeenCalledWith({url:'/pages/workspace/index'});
expect(taro.navigateTo).not.toHaveBeenCalled();
```

Keep the underlying keys `script`, `fetch`, `comic`, `shuihuo`, `batch`, `agent`, and `toolkit`. Add a typed `FUNCTION_PRESENTATION: Record<ModuleKey, {title:string; subtitle:string; asset:string}>` using the exact map in Global Constraints. `AI-CUT` is presentation copy for the existing `agent` route and must show `Agent 工作区` as its subtitle; no new module key or route is permitted.

- [ ] **Step 2: Run the tests and verify the current direct navigation fails**

```powershell
npm test -- --run src/pages/functions/functions.interaction.test.tsx src/config/modules.test.ts
```

Expected: FAIL because the current page calls `navigateTo(moduleRoute(key))`.

- [ ] **Step 3: Implement the exact visible hierarchy**

Rebuild the tree in this order:

1. Brand lockup.
2. `我的创作空间` title and `灵感成稿 · 创意成片` subtitle.
3. Full-width `继续创作` dragon hero, bound to the last selected module and switching to 工作台.
4. `内容创作`: one large 私人定制 card and two stacked 小说创作/批量测书 cards.
5. `生产成片`: two stacked 水货生产/AI-CUT cards and one large 漫剧创作 card. The AI-CUT card uses key `agent` and subtitle `Agent 工作区`.
6. `AI与工具`: one 会员与工具 card using key `toolkit`; do not render Agent a second time.

Render the seven locked keys exactly once and assert this exact sequence in the test: `script, fetch, batch, shuihuo, agent, comic, toolkit`. Do not create an eighth route to satisfy the decorative AI-CUT label.

- [ ] **Step 4: Match geometry in `gallery.scss`**

Use the reference image at 852px wide as the measurement source and scale to the simulator width. Preserve the relative layout: hero around 29% of reference width in height, section gaps around 7%, 1.5:0.9 asymmetric columns, 24–28PX radii, 1PX cyan/white glass rims, navy fill, cyan selected glow, and fixed full-width bottom navigation clearance.

Use the existing `studio-sunset.jpg` as the full-page aspect-fill background. Do not rasterize the entire screenshot as a page background.

- [ ] **Step 5: Run function tests and build**

```powershell
npm test -- --run src/pages/functions/functions.interaction.test.tsx src/config/modules.test.ts
npm run build:weapp:sandbox
```

Expected: tests PASS and Webpack compiles.

- [ ] **Step 6: Capture and compare 功能**

Open `/pages/functions/index` at 390 × 844. Capture the first viewport and the full scroll. Compare against `01-functions-home.png`. Record these measured checks in `miniprogram/design-qa.md`: title left edge within 4PX, hero bounds within 6PX, card bounds within 6PX, illustration occupied-area difference below 8%, text baseline difference below 4PX, corner-radius difference below 2PX, and the selected 功能 underline centered beneath the first tab. Correct any failed measurement before committing.

- [ ] **Step 7: Commit 功能**

```powershell
git add -- miniprogram/src/pages/functions/index.tsx miniprogram/src/pages/functions/gallery.scss miniprogram/src/pages/functions/functions.interaction.test.tsx miniprogram/src/config/modules.ts miniprogram/src/config/modules.test.ts
git commit -m "feat: rebuild functions as the primary mini-program home"
```

---

### Task 4: Turn 工作台 into the Selected Function Execution Hub

**Files:**
- Create: `miniprogram/src/services/workbenchView.ts`
- Create: `miniprogram/src/services/workbenchView.test.ts`
- Modify: `miniprogram/src/pages/workspace/index.tsx`
- Modify: `miniprogram/src/pages/workspace/studio.scss`
- Modify: `miniprogram/src/pages/workspace/studio.interaction.test.tsx`

**Interfaces:**
- Consumes: `activeModule: CreatorModule`, `entries: AggregatedTask[]`.
- Produces: `deriveWorkbenchView(module, entries): WorkbenchView`.
- Produces type:

```ts
export type WorkbenchView = {
  module: CreatorModule;
  current: AggregatedTask | null;
  primaryLabel: '开始配置' | '继续执行';
  completed: number | null;
  total: number | null;
  percent: number | null;
};
```

- [ ] **Step 1: Write failing pure derivation tests**

Cover: matching running entry wins; otherwise most recent matching entry; other modules are ignored; missing progress returns null rather than fabricated values; no entries produces `开始配置`.

```ts
expect(deriveWorkbenchView(comic, entries)).toMatchObject({
  current: expect.objectContaining({id:'comic-13'}),
  primaryLabel:'继续执行', completed:12, total:18, percent:67,
});
expect(deriveWorkbenchView(fetchModule, [])).toMatchObject({
  current:null, primaryLabel:'开始配置', completed:null, total:null, percent:null,
});
```

- [ ] **Step 2: Run the new test and verify it fails**

```powershell
npm test -- --run src/services/workbenchView.test.ts
```

Expected: FAIL because the module does not exist.

- [ ] **Step 3: Implement `deriveWorkbenchView`**

Match `AggregatedTask.source` to module keys using an explicit map. Compute percent only when `total > 0`, using `Math.round(completed / total * 100)`. Never infer a project title or stage from a missing field.

- [ ] **Step 4: Replace the old home/workspace catalog**

Keep existing session gating and `loadTaskSummary`. Remove `workspaceShortcuts`, the private choice sheet, duplicate recent-record catalog, and the companion hero.

Render the approved order:

1. `工作台` title and execution-center subtitle.
2. Current-workspace glass card with selected module illustration, truthful status, and one primary button.
3. Current project card only when `view.current` exists; progress and stages appear only when returned data supports them.
4. `本次执行` grouped rows for model, aspect ratio, completed count, and update time; render `—` when absent.
5. No-selection/session/error states with one recovery action.

The primary button calls `Taro.navigateTo({url: moduleRoute(activeModule.key)})`; selecting a function itself happened on 功能.

- [ ] **Step 5: Update interaction tests**

Assert current module is read on `useDidShow`, exactly one primary execution action exists, empty fields show `—`, clicking the action uses `moduleRoute`, and the old complete function catalog is absent.

- [ ] **Step 6: Run focused tests and build**

```powershell
npm test -- --run src/services/workbenchView.test.ts src/pages/workspace/studio.interaction.test.tsx
npm run build:weapp:sandbox
```

- [ ] **Step 7: Capture and compare 工作台**

Capture these four fixtures: `loading`, `entries=[]`, one matching `running` entry with `{completed:12,total:18}`, and rejected `loadTaskSummary`. Compare the running fixture to `02-workbench.png`. For the other fixtures keep the same title, 16PX page gutter, 16PX vertical section gap, 24PX card radius, and bottom-navigation clearance; render only `加载中…`, `暂无执行中的任务`, or `任务加载失败，点击重试` in the project area.

- [ ] **Step 8: Commit 工作台**

```powershell
git add -- miniprogram/src/services/workbenchView.ts miniprogram/src/services/workbenchView.test.ts miniprogram/src/pages/workspace/index.tsx miniprogram/src/pages/workspace/studio.scss miniprogram/src/pages/workspace/studio.interaction.test.tsx
git commit -m "feat: make workspace the selected function execution hub"
```

---

### Task 5: Match the Approved 任务 Center

**Files:**
- Modify: `miniprogram/src/pages/tasks/index.tsx`
- Create: `miniprogram/src/pages/tasks/index.scss`
- Modify: `miniprogram/src/pages/tasks/interaction.test.tsx`
- Modify: `miniprogram/src/services/taskSummary.ts`
- Modify: `miniprogram/src/services/taskSummary.test.ts`

**Interfaces:**
- Consumes: current `TaskSummary` and `AggregatedTask` values.
- Produces: truthful summary counts and the filters `全部、进行中、待确认、已完成`.
- Preserves: `taskDestination(entry)` for real record navigation.

- [ ] **Step 1: Write failing status/count tests**

Add normalization for known pending/confirmation raw statuses and assert unknown values remain unknown:

```ts
expect(normalizeTaskStatus('pending_confirmation')).toBe('pending');
expect(normalizeTaskStatus('accepted_pending')).toBe('pending');
expect(normalizeTaskStatus('alien_state')).toBe('unknown');
expect(taskStatusCounts(entries)).toEqual({running:3,pending:2,completed:12,failed:1});
```

Extend `TaskFilter` and `AggregatedTask['status']` with `pending` only where backed by existing server values.

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
npm test -- --run src/services/taskSummary.test.ts src/pages/tasks/interaction.test.tsx
```

- [ ] **Step 3: Implement truthful pending normalization and counts**

Do not map generic unknown status to pending. Keep failed entries visible under 全部 and provide retry only when the existing operation is safe and real; otherwise retain the current source-level reload action.

- [ ] **Step 4: Recompose the page**

Render: approved title/subtitle, summary strip, segmented filters, then one grouped glass list with separators. Each row shows source illustration; `entry.title?.trim() || '未命名任务'`; source label; `updatedAt ? formatTime(updatedAt) : '时间未知'`; normalized status; progress only when `total > 0`; and a chevron only when `entry.id` is non-empty.

- [ ] **Step 5: Add page-local SCSS**

Match `03-tasks.png`: compact 3-column summary, 4-way segmented filter, single rounded list surface, 88–104PX rows, cyan/yellow/green/red status system, and fixed navigation clearance. Use existing theme variables and sunset wall.

- [ ] **Step 6: Verify all page states**

Run focused tests and build. Capture loading, empty, partial-source-error, all-error, and populated states. The populated screenshot must match `03-tasks.png`; other states must keep the same shell and readable contrast.

- [ ] **Step 7: Commit 任务**

```powershell
git add -- miniprogram/src/pages/tasks/index.tsx miniprogram/src/pages/tasks/index.scss miniprogram/src/pages/tasks/interaction.test.tsx miniprogram/src/services/taskSummary.ts miniprogram/src/services/taskSummary.test.ts
git commit -m "feat: match the approved mini-program task center"
```

---

### Task 6: Match the Approved 我的 Space without Fake Account Data

**Files:**
- Modify: `miniprogram/src/pages/profile/index.tsx`
- Create: `miniprogram/src/pages/profile/index.scss`
- Modify: `miniprogram/src/pages/profile/index.test.tsx`
- Modify: `miniprogram/src/config/profileData.ts`
- Modify: `miniprogram/src/config/profileData.test.ts`

**Interfaces:**
- Consumes: existing member/profile/usage data and `profileFacts`.
- Produces: display-safe profile totals using `number | null`, never mock values.
- Preserves: avatar upload, profile-detail routes, appearance sheet, reduce-transparency switch, executor status, and logout.

- [ ] **Step 1: Write failing presentation tests**

Assert returned values distinguish zero from missing:

```ts
expect(profileFacts({member:{usage:{running:0,completed:28}}})).toMatchObject({
  runningTasks:0, completedTasks:28,
});
expect(profileFacts({member:{}})).toMatchObject({
  runningTasks:null, completedTasks:null,
});
```

Update render tests to require the approved visible groups and `—` for missing counts.

- [ ] **Step 2: Run tests and verify failure**

```powershell
npm test -- --run src/config/profileData.test.ts src/pages/profile/index.test.tsx
```

- [ ] **Step 3: Extend presentation data without inventing API fields**

Read only `member.usage.running`, `member.usage.completed`, and the existing quota fields already typed in `profileData.ts`. Return `null` for an absent value and `0` for an explicit zero. Render `—` for null and never add a second network request solely to fill decoration.

- [ ] **Step 4: Recompose the profile page**

Preserve the existing behavior but render the approved order: title/subtitle, dragon-avatar identity card, truthful running/completed facts, quota/usage strip, one grouped list containing assets/preferences/member/security/help, then reduce-transparency switch. Keep executor status and logout in the correct existing sections below the reference viewport.

- [ ] **Step 5: Add page-local SCSS**

Match `04-profile.png`: profile card height and avatar ratio, compact quota strip, single grouped list with separators, 3D icons, switch row, dark liquid-glass surface, and selected 我的 tab.

- [ ] **Step 6: Verify behavior and visual states**

Run focused tests and build. Capture loading, load failure, complete profile, missing optional data, appearance sheet, and executor expanded states. Compare the complete profile viewport to `04-profile.png` while preserving honest data fallbacks.

- [ ] **Step 7: Commit 我的**

```powershell
git add -- miniprogram/src/pages/profile/index.tsx miniprogram/src/pages/profile/index.scss miniprogram/src/pages/profile/index.test.tsx miniprogram/src/config/profileData.ts miniprogram/src/config/profileData.test.ts
git commit -m "feat: match the approved mini-program profile space"
```

---

### Task 7: Full Visual, Interaction, and Package Acceptance

**Files:**
- Create: `miniprogram/design-qa.md`
- Modify: only the exact Task 1–6 source file whose named assertion or measured threshold fails; add that file explicitly to the final commit command.

**Interfaces:**
- Consumes: all four implemented root pages and the four visual references.
- Produces: recorded 390 × 844 screenshots and a pass/fail matrix for pages, states, interactions, and pixels.

- [ ] **Step 1: Run the complete relevant test set**

```powershell
npm test -- --run src/app.config.test.ts src/ui/studioNavigation.test.ts src/config/modules.test.ts src/pages/functions/functions.interaction.test.tsx src/services/workbenchView.test.ts src/pages/workspace/studio.interaction.test.tsx src/services/taskSummary.test.ts src/pages/tasks/interaction.test.tsx src/config/profileData.test.ts src/pages/profile/index.test.tsx
```

Expected: all focused tests PASS.

- [ ] **Step 2: Run compatibility and production-like builds**

```powershell
npm test -- --run src/ui/wxssCompatibility.test.ts src/ui/liquidGlassPages.test.ts
npm run build:weapp:sandbox
```

Confirm Webpack succeeds and inspect the emitted main/subpackage sizes. Stop if the main package exceeds WeChat limits.

- [ ] **Step 3: Verify the actual WeChat simulator**

Open the project with the official developer-tool CLI, log into the existing authorized local session, and capture each root at 390 × 844. Do not accept a blank, loading, login-wall, or wrong-route screenshot.

- [ ] **Step 4: Verify interactions end to end**

Check: 功能 card writes the selected module and switches to 工作台; 工作台 shows that module and opens its real route; 任务 opens only records with IDs; 我的 rows preserve their existing destinations; bottom navigation order and selected state are correct on all roots.

- [ ] **Step 5: Compare pixels and fix visible drift**

For each page compare reference and implementation side by side. Record and fix deviations in background crop, title baseline, hero/card bounds, illustration scale, typography hierarchy, corner radii, dividers, bottom navigation, scroll clearance, and safe area. Repeat until no material deviation remains at the target viewport.

- [ ] **Step 6: Write `design-qa.md`**

Record each reference path, implementation screenshot path, viewport, tested states, interaction result, pixel findings, fixes applied, and final PASS/BLOCKED status. Do not mark PASS without accepted screenshots for all four pages.

- [ ] **Step 7: Final repository checks**

```powershell
git diff --check
git status --short
```

Confirm only intended mini-program files and the QA record changed. Do not stage unrelated dirty artifacts.

- [ ] **Step 8: Commit final QA fixes**

```powershell
git add -- miniprogram/design-qa.md
git commit -m "test: verify functions-first mini-program parity"
```

Add any final fixed source files explicitly to the same command; never use `git add .`.

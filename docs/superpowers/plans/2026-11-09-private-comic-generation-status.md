# Private Comic Generation Status Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add two always-visible private comic status cards that accurately show prompt and image progress, failures, completion, and retry actions across refreshes.

**Architecture:** Put all state-to-presentation rules in one pure helper so the React page only renders returned labels, tones, counts, and retry flags. Reuse existing `comicReadiness`, loaded page failure states, `action`, and generation callbacks; do not add persistence or a second task-state system.

**Tech Stack:** React, Ant Design, Node.js built-in test runner, CSS, Go `embed` local runtime.

## Global Constraints

- Always show separate `漫画提示词` and `漫画图片` cards.
- A prompt page is complete only when both `comicPrompt` and `imagePrompt` exist.
- Asset images remain optional and never count toward comic image progress.
- Running actions take immediate display priority; saved backend readiness restores state after refresh.
- Completed state wins over stale historical failure state.
- Retry actions call existing prompt/image callbacks and only refill missing work.
- Do not trigger a real provider call during acceptance.
- Work directly on the current `v88` development branch; do not create a branch, push, or deploy.

---

### Task 1: Pure status presentation resolver

**Files:**
- Create: `frontend/src/user/pages/private-studio/privateComicGenerationStatus.js`
- Create: `frontend/src/user/pages/private-studio/privateComicGenerationStatus.test.js`

**Interfaces:**
- Consumes: `{ assetCount, pagesTotal, promptReadyCount, imageReadyCount, action, taskFailed, promptFailed, imageFailed }`.
- Produces: `privateComicGenerationStatus(input)` returning `{ prompts, images }`; each item contains `key`, `label`, `detail`, `tone`, `running`, and `canRetry`.

- [ ] **Step 1: Write the failing resolver tests**

```js
assert.deepEqual(privateComicGenerationStatus({ assetCount: 5 }), {
  prompts: { key: 'prompts', label: '未生成', detail: '0 页', tone: 'idle', running: false, canRetry: false },
  images: { key: 'images', label: '等待提示词', detail: '提示词完成后可生成', tone: 'waiting', running: false, canRetry: false }
});

assert.equal(privateComicGenerationStatus({ pagesTotal: 20, promptReadyCount: 8 }).prompts.label, '部分完成');
assert.equal(privateComicGenerationStatus({ pagesTotal: 20, promptReadyCount: 20, imageReadyCount: 6, action: 'pages' }).images.label, '生成中');
assert.equal(privateComicGenerationStatus({ pagesTotal: 20, promptReadyCount: 20, imageReadyCount: 20, imageFailed: true }).images.label, '已完成');
```

- [ ] **Step 2: Run the test and verify the missing module failure**

Run: `node --test src/user/pages/private-studio/privateComicGenerationStatus.test.js`

Expected: FAIL with `ERR_MODULE_NOT_FOUND` for `privateComicGenerationStatus.js`.

- [ ] **Step 3: Implement the minimal pure resolver**

```js
export function privateComicGenerationStatus(input = {}) {
  const pagesTotal = Math.max(0, Number(input.pagesTotal || 0));
  const promptReadyCount = Math.min(pagesTotal, Math.max(0, Number(input.promptReadyCount || 0)));
  const imageReadyCount = Math.min(pagesTotal, Math.max(0, Number(input.imageReadyCount || 0)));
  const promptsComplete = pagesTotal > 0 && promptReadyCount >= pagesTotal;
  const imagesComplete = promptsComplete && imageReadyCount >= pagesTotal;
  // Return prompt and image presentation objects using the approved priority rules.
}
```

- [ ] **Step 4: Run the resolver tests and verify all states pass**

Run: `node --test src/user/pages/private-studio/privateComicGenerationStatus.test.js`

Expected: all tests PASS with `fail 0`.

- [ ] **Step 5: Commit the resolver and tests**

```bash
git add frontend/src/user/pages/private-studio/privateComicGenerationStatus.js frontend/src/user/pages/private-studio/privateComicGenerationStatus.test.js
git commit -m "feat: derive private comic generation status"
```

---

### Task 2: Render persistent prompt and image status cards

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Modify: `frontend/src/user/pages/private-studio/private-studio-workbench.css`

**Interfaces:**
- Consumes: `privateComicGenerationStatus(input)` from Task 1.
- Produces: two `role="status"` cards with stage label, state, count/detail, and conditional retry button.

- [ ] **Step 1: Wire the tested resolver to existing truthful state**

```jsx
const generationStatus = privateComicGenerationStatus({
  assetCount: assetList.length,
  pagesTotal,
  promptReadyCount,
  imageReadyCount,
  action,
  taskFailed: task?.status === 'failed',
  promptFailed: task?.status === 'failed' && promptReadyCount < pagesTotal,
  imageFailed: pageFailed
});
```

Render the two cards immediately below the toolbar. Prompt retry calls `onGeneratePrompts`; image retry calls `onGenerateAllPageImages`. Disable either retry while `busy || actionBusy`.

- [ ] **Step 2: Add responsive scoped styles**

```css
.private-studio-workbench .private-comic-generation-statuses {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

@media (max-width: 760px) {
  .private-studio-workbench .private-comic-generation-statuses { grid-template-columns: 1fr; }
}
```

Use distinct but theme-compatible idle, running, complete, and failed tones. Do not rely on color alone; always render the state label and detail text.

- [ ] **Step 3: Run focused tests**

Run: `node --test src/user/pages/private-studio/privateComicGenerationStatus.test.js src/user/pages/PrivateStudioWorkbench.layout.test.js`

Expected: all tests PASS with `fail 0`.

- [ ] **Step 4: Commit the UI and styles**

```bash
git add frontend/src/user/pages/PrivateComicWorkspace.jsx frontend/src/user/pages/private-studio/private-studio-workbench.css
git commit -m "feat: show private comic generation status"
```

---

### Task 3: Full verification and local visible acceptance

**Files:**
- Verify only; no production file should change unless a test exposes a defect.

**Interfaces:**
- Consumes: committed status resolver and UI.
- Produces: test, build, runtime identity, clean-tree, and browser evidence.

- [ ] **Step 1: Run the full frontend test suite**

Run: `npm test`

Expected: all tests PASS with `fail 0`.

- [ ] **Step 2: Build embedded frontend assets and local Go binary**

Run from repository root: `powershell -ExecutionPolicy Bypass -File .\scripts\build-go-local.ps1`

Expected: exit code `0` and embedded frontend assets produced.

- [ ] **Step 3: Restart only the confirmed local review runtime**

Run: `powershell -ExecutionPolicy Bypass -File .\scripts\start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached`

Expected: `http://localhost:4001/health` returns `200` and `/api/runtime-build-info` reports the current commit SHA.

- [ ] **Step 4: Verify the existing local private comic task without provider calls**

Open `/script`, select `漫画生成`, and assert:

- one visible `漫画提示词` status card;
- one visible `漫画图片` status card;
- current 5-asset, 0-prompt task shows `未生成` and `等待提示词`;
- the existing `生成漫画提示词` button remains enabled;
- no generation or retry button is clicked.

- [ ] **Step 5: Confirm repository cleanliness**

Run: `git status --short` and `git log -1 --oneline`.

Expected: no uncommitted output and HEAD is the status UI commit.

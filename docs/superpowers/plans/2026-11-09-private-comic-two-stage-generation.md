# Private Comic Two-Stage Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the private comic primary “开始生成” action with an explicit prompt stage followed by an image stage.

**Architecture:** Add one pure state resolver that converts current source, task, asset, prompt, image, busy, and action state into a primary action descriptor. `PrivateComicWorkspace` renders that descriptor and routes the click to the existing prompt or page-image callbacks; no backend endpoint or payload changes.

**Tech Stack:** React, Ant Design, JavaScript ES modules, Node test runner, Vite, Go embed runtime.

## Global Constraints

- Keep asset extraction as a separate left-side action.
- Asset images remain optional and never block prompt generation.
- Prompt completion requires both `comicPrompt` and `imagePrompt` on every page.
- Generate only missing page images from the primary image-stage action.
- Do not call a real text or image provider during verification.
- Update and run only the local `localhost:4001` runtime; do not deploy online.

---

### Task 1: Resolve the primary comic action from real workflow state

**Files:**
- Create: `frontend/src/user/pages/private-studio/privateComicPrimaryAction.js`
- Create: `frontend/src/user/pages/private-studio/privateComicPrimaryAction.test.js`

**Interfaces:**
- Consumes: `{ sourceReady, taskReady, assetCount, pagesTotal, promptReadyCount, imageReadyCount, busy, action }`.
- Produces: `{ key: 'prompts' | 'images' | 'complete', label: string, disabled: boolean }` from `privateComicPrimaryAction(state)`.

- [ ] **Step 1: Write the failing behavior test**

Test literal fixtures for: missing assets disables prompt generation; assets enable “生成漫画提示词”; partial prompts show “补齐漫画提示词”; all prompts enable “生成漫画图片”; all images show disabled “漫画图片已完成”; active prompt and image actions show the matching loading copy.

- [ ] **Step 2: Run test and verify RED**

Run: `node --test src/user/pages/private-studio/privateComicPrimaryAction.test.js`

Expected: FAIL because `privateComicPrimaryAction.js` does not exist.

- [ ] **Step 3: Implement the minimal resolver**

Use `pagesTotal > 0 && promptReadyCount >= pagesTotal` for the prompt-stage boundary, and require the same prompt boundary plus `imageReadyCount >= pagesTotal` for completion.

- [ ] **Step 4: Run test and verify GREEN**

Run: `node --test src/user/pages/private-studio/privateComicPrimaryAction.test.js`

Expected: all cases PASS.

### Task 2: Wire the resolver to the private comic primary button

**Files:**
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Test: `frontend/src/user/pages/private-studio/privateComicPrimaryAction.test.js`

**Interfaces:**
- Consumes: `privateComicPrimaryAction(state)` from Task 1.
- Produces: a primary button that calls `onGeneratePrompts` for `key === 'prompts'`, calls `onGenerateAllPageImages` for `key === 'images'`, and does nothing for `key === 'complete'`.

- [ ] **Step 1: Extend the failing behavior test for click routing data**

Require every non-complete resolver result to expose the literal action key consumed by the workspace (`prompts` or `images`), while the complete result exposes `complete` and stays disabled.

- [ ] **Step 2: Run focused behavior tests and verify RED**

Run: `node --test src/user/pages/private-studio/privateComicPrimaryAction.test.js`

Expected: FAIL because the resolver does not exist yet.

- [ ] **Step 3: Implement the minimal workspace wiring**

Count prompt-ready pages only when both prompt fields are present. Render `primaryAction.label`, `primaryAction.disabled`, and stage-specific loading. Keep re-generation entries in “更多操作”.

- [ ] **Step 4: Remove the unused workspace prop**

Remove `onSubmit` from `PrivateComicWorkspace` parameters and from the `ScriptPage` JSX call. Keep `submitPrivateComic()` for the existing retry path only.

- [ ] **Step 5: Run focused behavior tests and browser wiring acceptance**

Run: `node --test src/user/pages/private-studio/privateComicPrimaryAction.test.js src/user/pages/PrivateComicWorkspace.source.test.js`

Expected: all cases PASS. After rebuilding, browser acceptance verifies the real primary button displays the resolver label and the old “开始生成” button is absent. This avoids adding a source-text change detector.

### Task 3: Verify, commit, rebuild, and visually accept locally

**Files:**
- Modify through build only: `frontend/dist`, `backend/web/dist`, `backend/qiantie-local.exe` (generated/ignored runtime artifacts).

**Interfaces:**
- Consumes: committed frontend source.
- Produces: local Go runtime whose `/api/runtime-build-info` SHA equals the source commit.

- [ ] **Step 1: Run full verification**

Run the focused Node tests, `npm test`, `npm run build`, and `git diff --check`.

- [ ] **Step 2: Commit source changes**

Commit message: `feat: split private comic prompt and image actions`.

- [ ] **Step 3: Rebuild and restart local 4001**

Stop only the listener whose executable is `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`; run `scripts/build-go-local.ps1`; start with `scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached`.

- [ ] **Step 4: Verify runtime identity**

Require `/health` 200, `/script` 200, the exact listener path, and `/api/runtime-build-info.git_sha` equal to `git rev-parse --short HEAD`.

- [ ] **Step 5: Browser acceptance without provider calls**

Open a fresh local `/script` URL. With the existing local task that has assets but no prompt pages, verify the primary button says “生成漫画提示词”, is enabled, and “开始生成” is absent. Do not click the generation button. Use pure resolver tests for image-stage and complete-stage acceptance so no paid generation is triggered.

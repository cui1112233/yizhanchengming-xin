# Novel Panel V78 Workbench Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the user-facing Novel Panel as a complete React workbench backed by the existing Go compatibility APIs, with a direct Novel Fetch handoff instead of a blank project chooser.

**Architecture:** The V78.3.0.117 package and the authenticated online workbench are reference material only. The maintained implementation stays in React source and calls authenticated Go routes under `/api/novel-panel/*`; no historical iframe, static bundle, Python process, or Node runtime becomes a local runtime dependency. A persisted panel project contains the original source, style analysis, forced roster, character/relationship data, outline output, timeline output, and export settings.

**Tech Stack:** React, Ant Design, Go `net/http`, MySQL/Redis/TOS through existing Go services, Go `embed` release path.

## Global Constraints

- Maintain the `v88` source line; do not restore `/novel-panel/workbench` or a V78 iframe.
- Frontend implementation belongs under `frontend/src`; Go API behavior belongs under `backend/internal/httpapi`.
- System prompt templates remain server-owned; the browser sends user data and selected options only.
- Preserve account isolation and use the logged-in account identity for every project operation.
- Keep current persisted panel projects readable; add fields without deleting old fields.
- Validate every visible state with a test before implementation and then in the authenticated browser.

---

### Task 1: Replace the blank chooser with a direct, project-aware workbench entry

**Files:**
- Modify: `frontend/src/user/pages/NovelPanelPage.jsx`
- Modify: `frontend/src/user/pages/novel-panel/NovelPanelReactPage.jsx`
- Create: `frontend/src/user/pages/novel-panel/novelPanelProjectState.js`
- Test: `frontend/src/user/pages/novel-panel/novelPanelProjectState.test.js`

**Interfaces:**
- Consumes: URL parameter `project`, `listNovelPanelProjects()`, `getNovelPanelProject(projectId)`, `saveNovelPanelProject(payload)`.
- Produces: `createNovelPanelProjectState({ projectId, projects, project })`, returning `{ mode, activeProjectId, project }`; `mode` is always `workbench` after the initial load. Saved projects are restore choices inside the workbench, not a blocking landing page.

- [ ] **Step 1: Write the failing test**

```js
test('opens a new workbench when no project query is supplied', () => {
  assert.deepEqual(createNovelPanelProjectState({
    projectId: '',
    projects: [{ id: 'book-a', name: '书 A' }],
    project: null
  }), { mode: 'workbench', activeProjectId: '', project: null });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `node --test src/user/pages/novel-panel/novelPanelProjectState.test.js`

- [ ] **Step 3: Implement the minimal project-state helper and route selection**

```js
export function createNovelPanelProjectState({ projectId, projects = [], project = null } = {}) {
  const activeProjectId = String(projectId || '').trim();
  return { mode: 'workbench', activeProjectId, project };
}
```

- [ ] **Step 4: Run focused tests and browser entry verification**

Run: `node --test src/user/pages/novel-panel/*.test.js`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/user/pages/NovelPanelPage.jsx frontend/src/user/pages/novel-panel
git commit -m "feat: restore novel panel workbench entry"
```

### Task 2: Add the V78 source, style, and forced-character workflow

**Files:**
- Create: `frontend/src/user/pages/novel-panel/NovelSourceWorkspace.jsx`
- Create: `frontend/src/user/pages/novel-panel/novelPanelWorkspaceModel.js`
- Modify: `frontend/src/shared/api/novelPanel.js`
- Modify: `frontend/src/user/pages/novel-panel/NovelPanelReactPage.jsx`
- Test: `frontend/src/user/pages/novel-panel/novelPanelWorkspaceModel.test.js`

**Interfaces:**
- Consumes: `analyzeNovelPanel({ novel_text, character_guide_input, appearance_reference })`, `/api/novel-panel/character-core/analyze`, and project `data`.
- Produces: `normalizeNovelPanelWorkspace(data)` with `sourceText`, `style`, `formalRoster`, `characters`, and `relationships`.

- [ ] **Step 1: Write failing normalization tests**

```js
test('retains raw source and the explicit forced roster separately', () => {
  const model = normalizeNovelPanelWorkspace({ source_text: '原文', character_guide_input: '阿青\n阿城' });
  assert.equal(model.sourceText, '原文');
  assert.deepEqual(model.formalRoster, ['阿青', '阿城']);
});
```

- [ ] **Step 2: Run the new test and verify it fails**

Run: `node --test src/user/pages/novel-panel/novelPanelWorkspaceModel.test.js`

- [ ] **Step 3: Implement source editor, style analysis, roster, character cards, and relationship editor**

Use Ant Design `Input.TextArea`, `Form`, `Collapse`, `Tabs`, `Card`, and `Button`. Persist only after successful Go API responses. Show server errors in `Alert`, and never place system prompts in the React source.

- [ ] **Step 4: Run focused tests**

Run: `node --test src/user/pages/novel-panel/*.test.js`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/user/pages/novel-panel frontend/src/shared/api/novelPanel.js
git commit -m "feat: add novel panel source and character workflow"
```

### Task 3: Restore storyboard, timeline, result, and export operations

**Files:**
- Create: `frontend/src/user/pages/novel-panel/NovelStoryboardWorkspace.jsx`
- Create: `frontend/src/user/pages/novel-panel/novelPanelStoryboardModel.js`
- Modify: `frontend/src/shared/api/novelPanel.js`
- Modify: `frontend/src/user/pages/novel-panel/NovelPanelReactPage.jsx`
- Test: `frontend/src/user/pages/novel-panel/novelPanelStoryboardModel.test.js`

**Interfaces:**
- Consumes: `outlineNovelPanelScenes(payload)`, `regenerateNovelPanelSceneOutline(payload)`, `generateNovelPanelScenePrompts(payload)`.
- Produces: normalized scene rows `{ sourceKey, sourceText, selectedCharacterIds, shots, durationSeconds }` and export text from the persisted result.

- [ ] **Step 1: Write a failing scene normalization test**

```js
test('keeps one source row even when it contains several timeline shots', () => {
  const scenes = normalizeStoryboardRows([{ source_text: '第一段', outline_shots: [{ prompt: '镜头一' }, { prompt: '镜头二' }] }]);
  assert.equal(scenes.length, 1);
  assert.equal(scenes[0].shots.length, 2);
});
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `node --test src/user/pages/novel-panel/novelPanelStoryboardModel.test.js`

- [ ] **Step 3: Implement scene generation, per-row regeneration, timeline presentation, copy, and TXT export**

Use per-operation loading states so an individual scene retry does not lock the whole workbench. Use the persisted project data as the single render source after every successful request.

- [ ] **Step 4: Run focused tests and the complete panel test group**

Run: `node --test src/user/pages/novel-panel/*.test.js`

- [ ] **Step 5: Commit**

```bash
git add frontend/src/user/pages/novel-panel frontend/src/shared/api/novelPanel.js
git commit -m "feat: restore novel panel storyboard workflow"
```

### Task 4: Connect Novel Fetch versions to panel projects

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go`
- Modify: `backend/internal/httpapi/novel_panel_compat.go`
- Modify: `frontend/src/shared/api/novelFetchWorkshop.js`
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Test: `backend/internal/httpapi/novel_panel_compat_test.go`
- Test: `frontend/src/user/pages/novel-fetch-panel-handoff.test.js`

**Interfaces:**
- Consumes: selected Novel Fetch task/version and its content source (`original`, `ai1`, `sensitive`, or `processed`).
- Produces: `POST /api/novel-fetch/panel-projects` returning `{ projectId, route }`, where `route` is `/novel-panel?project=<id>`.

- [ ] **Step 1: Write a failing Go handler test**

```go
func TestNovelFetchPanelProjectUsesSelectedVersionContent(t *testing.T) {
  // seed a task with original and AI1 text, request source=ai1, and assert the saved panel project source_text is AI1.
}
```

- [ ] **Step 2: Run it to verify the missing handoff fails**

Run: `go test ./internal/httpapi -run TestNovelFetchPanelProjectUsesSelectedVersionContent -count=1`

- [ ] **Step 3: Implement the account-scoped Go handoff and React action**

The UI labels sources as `原文`, `小说获取 AI1`, `小说获取（敏感）`, and `小说获取` before opening the returned panel route.

- [ ] **Step 4: Verify handler and UI tests**

Run: `go test ./internal/httpapi -run 'TestNovelFetchPanelProject' -count=1`

Run: `node --test src/user/pages/novel-fetch-panel-handoff.test.js`

- [ ] **Step 5: Commit**

```bash
git add backend/internal/httpapi frontend/src/shared/api frontend/src/user/pages
git commit -m "feat: hand novel fetch versions to novel panel"
```

### Task 5: Embed build, full regression, and authenticated browser acceptance

**Files:**
- Modify only if required by failures discovered in this task.

- [ ] **Step 1: Run the feature suites**

Run: `node --test src/user/pages/novel-panel/*.test.js src/user/pages/novel-fetch-panel-handoff.test.js`

Run: `go test ./internal/httpapi -run 'TestNovelPanel|TestNovelFetchPanelProject' -count=1`

- [ ] **Step 2: Build the frontend and embedded Go binary**

Run: `npm run build` from `frontend`, then `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`.

- [ ] **Step 3: Start the local Go review service and inspect the authenticated browser**

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`.

Verify: direct `/novel-panel` has no blank chooser when projects exist; a Novel Fetch source opens its matching project; source, roster, character, storyboard, history, and export controls are visible and responsive.

- [ ] **Step 4: Commit only task-owned source and tests**

```bash
git add frontend/src backend/internal/httpapi docs/superpowers/plans/2026-10-20-novel-panel-v78-workbench-parity.md
git commit -m "feat: restore novel panel V78 workbench parity"
```

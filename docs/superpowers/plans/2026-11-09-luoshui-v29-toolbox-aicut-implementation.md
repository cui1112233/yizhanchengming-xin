# Luoshui v2.9 Toolbox and AI-CUT Implementation Plan

> **For Codex:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace placeholder plugin entries with evidence-mapped, working V88 workflows, beginning with Asset Reference, then add deterministic file tools, model tools, media executors, and the independent AI-CUT workbench.

**Architecture:** Keep Plugin Center as the discovery layer. Each plugin launches a typed workbench backed by existing V88 project, asset, generation, prompt, task, and local-executor services. Agent Workspace remains unchanged; Asset Reference reuses its project and image-generation contracts. AI-CUT becomes a separate persisted domain and only uses Jianying through an optional adapter.

**Tech Stack:** React + Ant Design, Go HTTP services, MySQL stores, Redis jobs/locks, TOS object storage abstraction, Go embed, Node test runner, Go tests.

---

### Task 1: Lock the evidence mapping and launch types

**Files:**
- Modify: `frontend/src/user/pages/pluginCatalog.js`
- Modify: `frontend/src/user/pages/pluginCatalog.layout.test.js`
- Reference: `F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\report\2026-11-09_功能到实现对照矩阵.md`

1. Add a failing test that Asset Reference launches a real workspace and that plugin status distinguishes available/configuration/development states.
2. Run the focused test and confirm failure.
3. Add typed launch metadata without changing the 30-item visible order.
4. Run the focused test and confirm pass.

### Task 2: Build the Asset Reference workbench shell

**Files:**
- Create: `frontend/src/user/components/AssetReferenceWorkbench.jsx`
- Create: `frontend/src/user/components/asset-reference-workbench.css`
- Create: `frontend/src/user/components/asset-reference-workbench.test.js`
- Modify: `frontend/src/user/pages/PluginCenterPage.jsx`

1. Write tests for project selection, upload gating, missing-model state, generation payload and candidate confirmation.
2. Implement the workbench state machine and accessible controls.
3. Connect the Plugin Center card to this workbench instead of the placeholder modal.
4. Run focused tests.

### Task 3: Connect real project, asset and image-generation APIs

**Files:**
- Modify: `frontend/src/shared/api/agent.js` only if a missing helper is required.
- Modify: `frontend/src/user/components/AssetReferenceWorkbench.jsx`
- Test: `frontend/src/shared/api/agent.test.js` or the closest existing generation test.

1. Reuse task list/create, asset list/upload, image generation and confirmation endpoints.
2. Send actual uploaded asset IDs in `referenceAssetIds`.
3. Poll existing task/execution state; do not invent a completed state in the browser.
4. Verify with mocked API tests. Do not call a paid model in acceptance.

### Task 4: Verify Asset Reference on local 4001

**Files:**
- Modify generated embed assets only through the existing build workflow.

1. Run focused frontend tests and the full frontend test/build suite.
2. Run relevant Go tests if API code changed.
3. Rebuild embedded frontend assets and the local Go binary using the repository scripts.
4. Restart only the confirmed local `qiantie-local.exe` on 4001.
5. Verify `/health`, `/api/runtime-build-info`, Plugin Center, project selection and a real test image upload. Do not press generate against a paid provider.

### Task 5: Implement deterministic webpage/file plugins

**Files:**
- Create typed workbenches for text formatting, document merge, subtitle merge and storyboard organization.
- Add Go handlers only where browser-only processing is unsafe or unsuitable.

For each feature: write a failing input/output test, implement parsing/export, verify download output, then mark available. Do not route deterministic tools through an LLM.

### Task 6: Correct model-tool contracts and prompts

**Files:**
- Modify: `backend/internal/prompt/...`
- Modify: plugin tool Go handlers and tests.
- Modify: typed frontend forms.

Move every system prompt into the backend prompt module. Replace the one-size-fits-all text modal with feature-specific fields and structured outputs. Test prompt selection and error behavior without real provider calls.

### Task 7: Add media executors and Jianying adapters

**Files:**
- Extend the existing local-executor contracts.
- Add persisted task records and progress UI.

Implement AI-Shot and AI-Upscale as media jobs. Implement Jianying-only tools behind explicit draft discovery and backup/restore rules. A missing executor must produce a configuration state, never fake success.

### Task 8: Build the independent AI-CUT foundation

**Files:**
- Create Go domain/store/routes under the existing backend structure.
- Create React AI-CUT project and timeline pages.

Start with persisted projects, asset pool, tracks and clips. Then add edits, ASR/SRT, smart split, audio separation, Agent commands and rendering. Keep Jianying export optional.

Current local implementation status (2026-11-09):

- [x] MySQL-backed owner-scoped projects with create/list/rename/delete APIs.
- [x] TOS/local-object-backed image, video, audio and SRT uploads up to 2 GB.
- [x] Optimistic-revision multi-track timeline persistence with clip bounds validation.
- [x] Independent `/ai-cut` React workbench with preview, add, timing edit, split, delete and save.
- [x] SRT parsing and atomic subtitle-track import, including normal/empty/duplicate/error/cancel/revision-conflict tests.
- [x] Caption correction through owner-configured text models, backend-only prompt, atomic text-only apply and request cancellation.
- [x] ASR task orchestration: dedicated `asr` model kind, OpenAI-compatible multipart transcription, MySQL job truth, Redis wake queue, progress/cancel/failure states and atomic subtitle-track write.
- [ ] Additional ASR adapters: AssemblyAI upload/poll, Deepgram, ElevenLabs, Cartesia, speaker diarization and optional local Whisper.
- [ ] Smart split, audio separation and Agent timeline commands.
- [ ] Server-side FFmpeg render/export jobs through MySQL + Redis.
- [ ] Optional Jianying adapter in the Windows local executor.

### Task 9: Repeat evidence and runtime acceptance

After every batch, update the feature matrix with source evidence, implementation path, test path, runtime proof and remaining gap. Do not commit, push or deploy unless the user separately authorizes it.

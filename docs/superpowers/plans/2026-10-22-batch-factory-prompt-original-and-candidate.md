# Batch Factory Prompt Original And Candidate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make each video storyboard retain a read-only AI original, a separately editable current prompt, and a regenerated AI candidate that only replaces either after explicit acceptance.

**Architecture:** Persist the current usable prompt in the existing video settings patch key `videoPrompt`; persist the last accepted AI output in `aiOriginalVideoPrompt`. A regenerated director result is stored as a book-scoped temporary candidate and never calls `PersistDirectorRevision`; the UI compares the selected video's current prompt with its candidate, then either applies the candidate atomically or clears it without changing formal content.

**Tech Stack:** Go, existing `batchfactoryv11` Store/settings contracts, HTTP API, React, Ant Design, Node source tests, Go focused tests.

## Global Constraints

- Work directly in the existing `v88` checkout; do not create branches, commit, reset, clean, push, deploy, or overwrite unrelated dirty work.
- Keep user text in `videoPrompt` as the only value read by final VIDEO compilation and video generation.
- Existing rows that lack `aiOriginalVideoPrompt` fall back to their current usable prompt when displayed; they must not show a legacy-warning-only experience.
- Candidate generation must not persist a `DirectorRevision`, replace a video prompt, or create a video task.
- Accepting a candidate updates both `aiOriginalVideoPrompt` and `videoPrompt` for only the selected video, using the existing revision-conflict safeguards.
- Canceling a candidate changes no accepted prompt.

---

### Task 1: Candidate Data Contract And HTTP API

**Files:**
- Modify: `backend/internal/batchfactoryv11/types.go`
- Modify: `backend/internal/batchfactoryv11/director_service.go`
- Modify: `backend/internal/httpapi/batch_factory_v11*.go` and router registration file found by the existing Batch Factory v11 routes
- Test: `backend/internal/batchfactoryv11/*_test.go`
- Test: `backend/internal/httpapi/*_test.go`

**Interfaces:**
- Produces book settings patch key `directorPromptCandidate` containing candidate ID, source book revision, created time, and candidate text keyed by video ID.
- Produces video settings patch key `aiOriginalVideoPrompt`.
- Produces endpoints to create/read candidate, accept candidate for one video, and cancel candidate.

- [ ] **Step 1: Write focused failing Go tests**
  - Candidate generation returns per-video text while leaving the stored director revision and all `videoPrompt` values unchanged.
  - Accept updates selected video's `aiOriginalVideoPrompt` and `videoPrompt` together, but leaves another video unchanged.
  - Cancel removes the temporary candidate but leaves accepted prompt data unchanged.
  - A stale expected revision returns the existing conflict response.

- [ ] **Step 2: Run the Go tests and confirm expected failure**
  - Run the new package-level test names with `go test` and verify failures are missing endpoint/service behavior, not compilation setup failures.

- [ ] **Step 3: Implement the minimal candidate flow**
  - Reuse the existing director contract/provider result parser to construct a candidate.
  - Do not call `PersistDirectorRevision` from candidate generation.
  - Persist the candidate book patch through existing settings storage.
  - On accept, persist a selected-video patch with both keys in one settings update and clear the accepted item from the temporary candidate.
  - On cancel, clear only `directorPromptCandidate`.

- [ ] **Step 4: Re-run the focused Go tests**
  - Confirm all candidate state and conflict cases pass.

### Task 2: Frontend Candidate API And Prompt Panel

**Files:**
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js` or a new focused source test that imports only this feature's helpers

**Interfaces:**
- Consumes Task 1 candidate endpoints and the two persisted settings keys.
- Produces a prompt panel where the current editable value is independent from read-only AI original and temporary candidate values.

- [ ] **Step 1: Write focused failing frontend tests**
  - The source/API layer requests a candidate instead of calling the destructive director `force` stage for “重新生成 AI”.
  - The panel labels `查看 AI 原始输出`, uses `aiOriginalVideoPrompt` with fallback to current prompt, and keeps the editable input bound to `videoPrompt`.
  - Accept and cancel use their dedicated API methods; video generation remains wired to the existing `videoPrompt` path.

- [ ] **Step 2: Run the focused frontend test and confirm expected failure**
  - Run only the new test file or exact matching test names. If the broader source suite is blocked by the known missing Novel Fetch helper, record that separately rather than fabricating it.

- [ ] **Step 3: Implement the UI**
  - Replace the user-facing “查看编译 Trace” action in this editing flow with a read-only “查看 AI 原始输出” modal.
  - Replace “重新生成导演分镜” in this flow with “重新生成 AI”; keep normal compile behavior separate.
  - Display a comparison modal: left current usable prompt, right the selected video candidate; provide “采用新结果” and “取消”.
  - After acceptance or cancellation, refresh the exact book snapshot so the open modal reflects persisted server data immediately.
  - Do not regress the existing rich editor, save action, asset mentions, visual/comic prompt behavior, or video generation action.

- [ ] **Step 4: Re-run focused frontend tests**
  - Confirm the new API names, labels, and data separation pass.

### Task 3: Integration Verification

**Files:**
- Modify only if an integration test exposes a real defect from Tasks 1-2.
- Test: focused Go and frontend tests from Tasks 1-2.

- [ ] **Step 1: Build and start the local embedded Go runtime**
  - Run `npm --prefix frontend run build`, `scripts/build-go-local.ps1`, restart only the confirmed repository-local listener using `scripts/start-local-go-review.ps1 -UseYcmMysql`, then run `scripts/check-local-go.ps1`.

- [ ] **Step 2: Verify in the authenticated browser**
  - Edit and save a current video prompt; verify the AI original remains unchanged after a reload.
  - Generate a candidate; verify current prompt and original remain unchanged before acceptance.
  - Cancel; verify neither accepted value changes after refresh.
  - Generate another candidate and accept it; verify current and original both update, survive refresh, and only the selected video changes.
  - Start video generation and verify the request uses the current saved `videoPrompt`.

- [ ] **Step 3: Record verification limits**
  - State any pre-existing unrelated test failure separately; do not modify unrelated missing files or user test data to make a full suite appear green.

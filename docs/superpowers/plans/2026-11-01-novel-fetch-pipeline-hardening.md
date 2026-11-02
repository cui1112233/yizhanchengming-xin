# Novel Fetch Pipeline Hardening Plan

> **For agentic workers:** Execute inline in the current authorized checkout. Keep changes on the existing `v88` branch; do not create a branch.

**Goal:** Make the Novel Fetch controls match their runtime behavior, prevent duplicate external submissions after uncertain outcomes, and retain enough processing provenance to resume and audit tasks.

**Architecture:** Reuse the existing Go workbench worker, MySQL document store, Redis queue, and target-site connector. Add focused behavior tests first. Implement sensitive-hit concurrency, make automatic-submit batching controls meaningful without changing manual-submit behavior, persist submission attempts before external requests, gate paid rewrite behind explicit classification policy, and attach a stable processing-settings fingerprint to source provenance.

**Tech Stack:** Go, React, existing Node/Go test suites, MySQL-backed Novel Fetch store, Redis process queue, existing video-management connector.

## Global Constraints

- Preserve unrelated dirty worktree changes and current branch.
- Do not commit, publish, deploy, fetch live novels, spend paid AI quota, or submit to the target site during verification.
- Preserve complete `OriginalRaw`; keep processing text separate.
- Do not copy sensitive match snippets or full body text into Batch Factory metadata.
- Keep system prompts and processing decisions in Go; React only exposes/presents settings.

---

### Task 1: Honor sensitive AI concurrency

**Files:** `backend/internal/httpapi/novel_fetch_workshop.go`, `backend/internal/httpapi/novel_fetch_rewrite_test.go`

- [x] Add a test that proves non-group mode runs no more than the configured number of requests at once, still repairs hits in source order, and preserves per-hit errors.
- [x] Run the test and verify it fails because the current code processes each hit serially.
- [x] Implement bounded concurrent requests while applying completed results in source order.
- [x] Run focused sensitive-repair tests.

### Task 2: Make automatic-submit batch settings effective

**Files:** `backend/internal/httpapi/novel_fetch_workshop.go`, `backend/internal/httpapi/novel_fetch_automation.go`, `backend/internal/httpapi/novel_fetch_workshop_test.go`, `frontend/src/user/pages/NovelFetchPage.jsx`

- [x] Add tests for the configured automatic-submit batch-size cap.
- [x] Run the tests and verify the settings are currently ignored.
- [x] Apply the batch-size cap only to automatic jobs; preserve immediate manual submission semantics.
- [x] Remove the unused flush-seconds control because this queue has no durable cross-job aggregation window; do not imply that a fixed delay is a real batch buffer.
- [x] Run focused automation and submission tests.

### Task 3: Persist external submission attempts before sending

**Files:** `backend/internal/httpapi/web_submit_service.go`, `backend/internal/httpapi/web_submit_submit_compat_test.go`, and existing Novel Fetch document/log storage helpers.

- [x] Add tests for a submit accepted remotely but followed by a lost response, and for task restart/retry after an uncertain attempt.
- [x] Run the tests and verify the current code resends without a durable pre-send attempt record.
- [x] Persist operation, version, and content fingerprint before calling the target. Reconcile uncertain attempts by readback before allowing another send.
- [x] Run focused submit tests with an in-memory fake connector.

### Task 4: Avoid paid rewrites while classification is unresolved

**Files:** `backend/internal/httpapi/novel_fetch_workshop.go`, `backend/internal/httpapi/novel_fetch_workshop_test.go`, `frontend/src/user/pages/NovelFetchPage.jsx` only if a policy control is required.

- [x] Add a test proving pending/low-confidence gender or style prevents automatic AI rewrite and becomes eligible again after manual confirmation.
- [x] Add a narrow gate for queued automatic processing while keeping the explicit manual generation route available.
- [x] Run focused worker/classification tests.

### Task 5: Record a stable processing-settings fingerprint

**Files:** `backend/internal/novelfetchworkshop` processing/config helpers, `backend/internal/httpapi/novel_fetch_workshop.go`, and `backend/internal/httpapi/batch_factory_v12_single_fetch.go` with focused tests.

- [x] Add tests showing equivalent normalized rule settings produce the same fingerprint, while a meaningful rule change produces a different fingerprint.
- [x] Persist the fingerprint and config revision with Novel Fetch task metadata; pass only fingerprint/version metadata into Batch Factory.
- [x] Run focused provenance and intake tests.

## Verification

- [x] Run focused `go test` packages/tests for Novel Fetch, sensitive repair, automation, and submission.
- [x] Run focused Node tests for Novel Fetch configuration controls.
- [x] Run the frontend production build.
- [ ] `go test -race` could not run because this Windows environment has no C compiler; the full package run also has two unrelated pre-existing comic automation failures.
- Do not restart the Go runtime or deploy as part of this source hardening task.

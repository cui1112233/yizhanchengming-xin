# Novel Fetch REA Workflow Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Novel Fetch stage history durable and refreshable, recover interrupted scheduled dispatches, and prevent duplicate active jobs for the same account/book intent.

**Architecture:** Keep the current Go, MySQL, and Redis boundaries. Persist concise per-book stage events in the existing document log, use stable schedule dispatch identifiers plus queue idempotency to reconcile interruptions, and use queue-level uniqueness at every enqueue entry point. Treat an uncertain external upload as pending verification, never as an automatic replay.

**Tech Stack:** Go `net/http`, existing Novel Fetch workshop `Store` and `ProcessQueue`, Redis `SET NX EX`, MySQL document/run/config persistence, React/Ant Design, Node test runner, Go tests.

## Global Constraints

- Preserve full `originalRaw`; do not modify fetch, sensitive-word, classification, rewrite, truncation, or submission business rules.
- Do not log API keys, cookies, authentication headers, or complete novel text.
- Keep the current `v88` branch; do not create a branch.
- Preserve all pre-existing dirty files and unrelated changes.
- Do not perform real fetching, paid AI calls, external submission, service restart, or deployment.
- Do not implement stage-specific retry controls or independent concurrency limits in this round.

---

## File Map

- `backend/internal/httpapi/novel_fetch_workshop.go`: stage log helpers, manual queue route, schedule dispatch/recovery, task runtime reconstruction.
- `backend/internal/httpapi/novel_fetch_automation.go`: automatic queue selection, stale-task recovery and dedupe key construction.
- `backend/internal/httpapi/novel_fetch_workshop_test.go`: API/workflow log, schedule, and duplicate-request regression tests.
- `backend/internal/httpapi/novel_fetch_automation_test.go` (if present): automatic queue dedupe and stale-task regression tests; otherwise add coverage to the existing automation test file.
- `backend/internal/novelfetchworkshop/queue.go`: in-memory queue uniqueness behavior used by tests and fallback runtime.
- `backend/internal/novelfetchworkshop/redis_queue.go`: distributed uniqueness and retry/ack cleanup behavior.
- `backend/internal/novelfetchworkshop/queue_test.go`: queue idempotency and lock lifecycle tests.
- `frontend/src/user/pages/NovelFetchPage.jsx`: render persisted task events and refresh from server only if current detail view does not already do so.
- `frontend/src/user/pages/novel-fetch-progress.js` and its test: normalize and display persisted stage events after refresh.
- `frontend/src/user/pages/novel-fetch-progress.test.js`: user-visible event order and status normalization tests.

## Task 1: Durable, Safe Per-Book Stage Events

### Step 1: Add failing backend tests

- [ ] Test that event append keeps the latest 200 events in chronological order and adds stage, status, timestamp, and safe detail fields.
- [ ] Test that appending a detail containing a credential-like value redacts it before persistence.
- [ ] Test the worker records start, success/failure, wait, recovery, and finish events without recording body text.

Run from `backend`: `go test ./internal/httpapi -run 'TestAppendWorkshopTaskLog|TestWorkshopProcessStageEvents' -count=1`.
Expected before implementation: missing safety/shape assertions fail.

### Step 2: Implement a single event writer

- [ ] Replace direct log map construction in task status updates and the workshop event helper with one helper that validates the event name, trims and redacts details, appends UTC timestamps, and caps history at 200.
- [ ] Add stage-start and stage-result events at the existing worker stage boundaries. Never include original or processed body contents.
- [ ] Preserve existing log fields consumed by compatibility routes (`time`, `event`, `detail`).

### Step 3: Verify backend event behavior

- [ ] Run `gofmt` on changed Go files.
- [ ] Run `go test ./internal/httpapi ./internal/novelfetchworkshop -count=1` from `backend`.

## Task 2: Refresh Reads Persisted Task State

### Step 1: Add failing frontend tests

- [ ] Test that persisted `logs` take precedence over absent in-memory progress after reload.
- [ ] Test chronological rendering, waiting reason visibility, and no fabricated failed/completed state when runtime data is empty.

Run from `frontend`: `node --test src/user/pages/novel-fetch-progress.test.js`.
Expected before implementation: persisted-log scenarios are not represented or assertions fail.

### Step 2: Wire persisted detail into the existing view

- [ ] Extend the existing progress normalizer to merge job progress with per-book persisted logs without duplicating identical events.
- [ ] Ensure task selection/reload requests the current task/document from the existing API and refreshes the displayed events.
- [ ] Keep the existing page layout and wording; do not add a new panel or alter other Novel Fetch views.

### Step 3: Verify frontend refresh behavior

- [ ] Run `node --test src/user/pages/novel-fetch-progress.test.js src/user/pages/novel-fetch-workbench-ui.source.test.js` from `frontend`.
- [ ] Run the frontend production build if source-level component changes are needed.

## Task 3: Recover Interrupted Scheduled Dispatch

### Step 1: Add failing schedule tests

- [ ] Test recovery when a stale `dispatching` schedule already has a durable queued/running run: mark schedule queued and do not create another job.
- [ ] Test recovery when no run/job exists and the stored dispatch intent is complete: enqueue once using the same stable identifiers.
- [ ] Test incomplete/ambiguous dispatch data becomes a visible failed/reconciliation-needed state and is not blindly re-enqueued.
- [ ] Test two dispatcher calls do not enqueue twice for one schedule.

Run from `backend`: `go test ./internal/httpapi -run 'TestRecoverNovelFetchSchedule|TestDispatchDueNovelFetchSchedules' -count=1`.
Expected before implementation: recovery cases fail.

### Step 2: Persist stable dispatch intent before enqueue

- [ ] Add named constants for the recovery timeout and stable schedule dedupe window, matching existing dispatcher cadence conventions.
- [ ] Persist `dispatching`, `dispatch_id`, `job_id`, `batch_id`, intended book IDs, and update time before queue interaction.
- [ ] Derive stable IDs from the schedule identifier and scheduled run time, not `UnixNano` generated during each retry.
- [ ] On each scan, reconcile stale `dispatching` records from persisted run state and queue dedupe result. Mark already queued/running work as queued; only enqueue when prior work is proven absent.
- [ ] Protect the read-modify-write schedule transition with the repository's existing lock primitive where available; do not add a new infrastructure service.
- [ ] Persist completed/error state after the queue action and append an audit event.

### Step 3: Verify recovery behavior

- [ ] Run `go test ./internal/httpapi ./internal/novelfetchworkshop -count=1` from `backend`.
- [ ] Confirm recovery tests do not invoke fetch, AI, or external submission services.

## Task 4: Idempotent Manual and Automatic Enqueue

### Step 1: Add failing duplicate-entry tests

- [ ] Test simultaneous same-owner/same-book manual requests produce one effective queued job and return the existing job identity to the duplicate request.
- [ ] Test automatic and scheduled entries for the same active book do not produce a second effective job.
- [ ] Test a completed book can be explicitly submitted as a new run.
- [ ] Test Redis enqueue failure clears a newly acquired dedupe key, while an already-existing key remains intact.

Run from `backend`: `go test ./internal/httpapi ./internal/novelfetchworkshop -run 'Test.*(Duplicate|Idempot|EnqueueUnique)' -count=1`.
Expected before implementation: manual/schedule duplicate cases fail.

### Step 2: Apply consistent queue idempotency

- [ ] Build normalized dedupe keys from owner, platform, source book ID, and active operation; do not use a raw concatenated batch list whose order changes the key.
- [ ] Use `ProcessDedupQueue.EnqueueUnique` for manual, automatic, and scheduled entry points; provide a locked in-memory fallback with equivalent semantics.
- [ ] On duplicate, return a conflict/accepted-existing response containing the existing task identity instead of reporting a newly created job.
- [ ] Retain dedupe ownership while jobs retry; release only on terminal ack/failure. Make Redis queue operations atomic enough that a failed enqueue cannot strand a lock or delete a different job's lock.
- [ ] Keep completed-task explicit reprocessing possible by incorporating a new run intent token into the key.

### Step 3: Verify queue uniqueness

- [ ] Run `go test ./internal/httpapi ./internal/novelfetchworkshop -count=1` from `backend`.
- [ ] Run the race detector for the focused queue/API tests: `go test -race ./internal/novelfetchworkshop ./internal/httpapi -run 'Test.*(Duplicate|Idempot|EnqueueUnique)' -count=1`.

## Task 5: Full Focused Verification and Review

- [ ] Run all Novel Fetch focused Go packages and frontend tests changed by this plan.
- [ ] Run `git diff --check` and inspect the full diff for unintended changes to the pre-existing worktree.
- [ ] Verify any schema change is append-only and compatible with already-applied migrations; prefer no schema change if existing `Logs`, `RunRecord`, and config fields suffice.
- [ ] Report exact files, tests, remaining edge cases, and confirm no service restart, external calls, or deployment occurred.

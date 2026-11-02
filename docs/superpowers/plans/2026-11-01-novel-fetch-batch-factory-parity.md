# Novel Fetch and Batch Factory Processing Parity Plan

> **For agentic workers:** Execute inline in the current authorized checkout. Keep changes on the existing main development branch; do not create a branch.

**Goal:** Make Batch Factory's direct novel-fetch paths preserve complete raw text while applying the same configured text rules and classification provenance as Novel Fetch.

**Architecture:** Reuse Novel Fetch's existing configured text processor instead of duplicating its rules. Thread the account-scoped Novel Fetch store and text provider into the direct bulk, single-book, and automatic fallback fetch paths. Keep raw text separate from the processed text and make processing stage/version visible in source metadata.

**Tech Stack:** Go, existing Novel Fetch workshop store and processors, Batch Factory V12 HTTP routes, Go tests.

## Global Constraints

- Preserve existing dirty worktree changes; edit only files required for this task.
- Keep the current branch; do not commit, publish, or deploy.
- Do not make external fetches, paid model calls, or submissions during verification.
- Preserve complete raw text separately from processed production text.

---

### Task 1: Share Novel Fetch processing with direct Batch Factory fetches

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_v12_fetch.go`
- Modify: `backend/internal/httpapi/batch_factory_v12_single_fetch.go`
- Modify: `backend/internal/httpapi/batch_factory_automation.go`
- Modify: `backend/internal/httpapi/router.go`
- Test: `backend/internal/httpapi/batch_factory_v12_fetch_test.go`

- [x] Write failing tests showing direct fetch applies configured sensitive/layout rules, preserves raw text, and records the chosen source version.
- [x] Run those tests and verify they fail because direct fetch currently returns the unprocessed processing text.
- [x] Pass account-scoped Novel Fetch settings/provider into direct fetch routes and process each fetched body through the existing configured text processor.
- [x] Keep upstream raw text unchanged in `rawData`; return processed text as `data` and mark the source metadata stage/version.
- [x] Apply the same behavior to bulk, single-book, and automatic fallback fetches.
- [x] Run focused HTTP API tests and the Novel Fetch workshop tests.

### Task 2: Verify source metadata survives the Batch Factory handoff

**Files:**
- Modify only if required: `frontend/src/user/pages/novel-fetch-batch-factory.js`
- Modify only if required: `backend/internal/batchfactoryv11/intakes.go`
- Test: existing handoff and intake tests

- [x] Add a focused regression assertion that selected content version and source metadata survive intake normalization and batch creation.
- [x] Run the focused handoff/intake tests; change production code only if the assertion exposes a loss.

### Task 3: Verify recoverable fetch and processing failures

**Files:**
- Modify only if required: the direct fetch implementation and its tests from Task 1

- [x] Add a regression test proving a text-processing failure does not discard the fetched raw source or mislabel it as a fetch failure.
- [x] Ensure metadata distinguishes fetch failure from processing/classification failure and remains retryable at the correct stage.
- [x] Run all focused tests and report remaining unverified runtime behavior separately.

## Verification Record

- Passed: `go test ./internal/novelfetchworkshop -count=1`.
- Passed: `go test ./internal/batchfactoryv11 -count=1`.
- Passed: focused Novel Fetch, Batch Factory intake, and handoff tests in `internal/httpapi`.
- Passed: `node --test frontend/src/user/pages/novel-fetch-batch-factory.test.js frontend/src/user/pages/shuihuo/batchFactoryManualFetch.test.js` (17 tests).
- The full `internal/httpapi` package run still has two failing comic automation tests (`TestComicAutomationStopsAfterWholeBookCollectionWithoutAudioOrVideo` and `TestComicAutomationHonorsBookConcurrency`); they are outside this change's novel-fetch processing path.
- No live upstream fetch, paid AI call, browser runtime acceptance, submission, or deployment was performed.

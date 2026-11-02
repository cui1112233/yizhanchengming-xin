# Novel Fetch Intake and Classification Provenance Implementation Plan

> **For agentic workers:** Execute inline in the current authorized checkout. Keep changes on the existing `v88` branch; do not create a branch.

**Goal:** Make Novel Fetch input parsing resilient, make uncertain gender/style classification explicit, and carry safe processing provenance into Batch Factory.

**Architecture:** Extend the existing input normalizer and shared Go classifier rather than replacing their call chains. Treat upstream/manual evidence as higher priority than AI, accept AI classifications only when structured output includes a valid confidence, and record the evidence and processing summary in the existing source metadata contract.

**Tech Stack:** React, Node built-in tests, Go, existing Novel Fetch and Batch Factory V12 stores/routes.

## Global Constraints

- Preserve unrelated dirty worktree files and existing edits in touched files.
- Do not change branch, commit, publish, or deploy.
- Do not perform live upstream fetches, paid AI calls, submissions, or production writes.
- Preserve complete `originalRaw`; apply `maxTxt` only to the processing text.
- Do not copy sensitive matched text snippets into Batch Factory metadata.
- Missing AI confidence is not a trusted score; classify as pending confirmation.

---

### Task 1: Improve input parsing without breaking existing formats

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch-input-format.js`
- Test: `frontend/src/user/pages/novel-fetch-input-format.test.js`
- Modify only as needed: `frontend/src/user/pages/NovelFetchPage.jsx`
- Test: `frontend/src/user/pages/novel-fetch-input-format.source.test.js`

- [x] Add tests for a header row, a two-row header, configurable column mapping, and current smart/fixed format compatibility.
- [x] Run the focused Node tests and confirm the new cases fail because headers/mapping are unsupported.
- [x] Implement header detection and safe caller-supplied column mapping while retaining duplicate/empty-ID accounting.
- [x] Run focused Node tests and the Novel Fetch source-contract test.

### Task 2: Add confidence-aware classification and pending confirmation

**Files:**
- Modify: `backend/internal/novelfetchworkshop/gender.go`
- Modify: `backend/internal/novelfetchworkshop/style.go`
- Modify: `backend/internal/novelfetchworkshop/classification.go`
- Test: `backend/internal/novelfetchworkshop/gender_test.go`
- Test: `backend/internal/novelfetchworkshop/style_test.go`
- Modify only as needed: `backend/internal/httpapi/batch_factory_v12_single_fetch.go`
- Test: focused Batch Factory fetch/classification tests in `backend/internal/httpapi`
- Modify only as needed: `frontend/src/user/pages/shuihu/BatchFactoryNovelList.jsx`
- Test: related Batch Factory source tests

- [x] Add tests proving known upstream/manual classifications remain trusted, valid structured AI results include confidence/reason, and missing/low/invalid confidence stays pending.
- [x] Run focused Go tests and confirm the new cases fail for missing confidence handling.
- [x] Parse a structured classification result with validated confidence and an explicit pending result on absent or insufficient confidence.
- [x] Persist gender/style confidence, source, reason, and an actionable classification status; keep production blocked only where the existing classification gate requires it.
- [x] Update the detail status copy so it does not claim classification failures never block production when the actual gate blocks it.
- [x] Run focused workshop and HTTP API tests plus related frontend source tests.

### Task 3: Carry safe Novel Fetch processing provenance into Batch Factory

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch-batch-factory.js`
- Test: `frontend/src/user/pages/novel-fetch-batch-factory.test.js`
- Modify only as needed: `backend/internal/httpapi/batch_factory_v12_single_fetch.go`
- Test: focused Batch Factory intake/fetch tests in `backend/internal/httpapi` and `backend/internal/batchfactoryv11`

- [x] Add tests for source task/version, processing rule version, input/raw/processed counts, replacement totals, and classification provenance in the handoff.
- [x] Assert metadata excludes sensitive match text and full body text.
- [x] Run focused tests and confirm the new metadata fields are missing before implementation.
- [x] Pass only the safe summary fields into Batch Factory and preserve the selected content version and full raw source separately.
- [x] Run focused Go and Node tests and inspect the final diff for unrelated changes.

## Verification

- Run focused Go tests from `backend` and focused Node tests from `frontend`.
- Run frontend build if touched UI/source imports warrant it.
- Do not rebuild or restart the embedded runtime unless separately authorized.
- Report source/test results separately from live browser and upstream acceptance.

# Batch Factory Multi-Opening Display and Publish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make a four-opening Batch Factory book show four complete merged videos and publish those four MP4 files plus one book-ID TXT in one Go-owned video-management submission.

**Architecture:** The Go publisher derives the required opening count from effective batch/book settings and the storyboard count, selects the newest successful merge independently for every required opening index, validates the complete set before uploading, and serializes returned video objects as one JSON array. The React clip library mirrors the same per-opening grouping for display while keeping retry attempts within an opening as history.

**Tech Stack:** Go, MySQL-backed Batch Factory repositories, Go video-management connector, React, Node test runner.

## Global Constraints

- Work only in local V88 on branch `v88`; do not publish to the public server.
- All backend selection, upload, submission, and readback logic must remain Go-only.
- React may render the Go response but must not become an upload backend.
- Do not modify, delete, or forge executed MySQL migration records.
- Do not reset, clean, overwrite, or commit unrelated dirty-worktree changes.
- Do not create a branch or commit unless the user explicitly requests it.
- Opening replacement is effective only when the book has at least two storyboards.
- Selecting four openings means opening indexes `0,1,2,3`: original opening plus three generated replacements.
- Missing any required merged video must prevent all video uploads and the final submission.

---

### Task 1: Go Multi-Opening Selection Contract

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_v11_publish_121_test.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_publish_121.go`

**Interfaces:**
- Consumes: `batchfactoryv11.Batch`, `batchfactoryv11.Book`, `[]batchfactoryv11.MergeJob`.
- Produces: `selectPublishMergeJobs(batch batchfactoryv11.Batch, book batchfactoryv11.Book, jobs []batchfactoryv11.MergeJob) ([]batchfactoryv11.MergeJob, error)`.

- [ ] **Step 1: Write failing selection tests**

Add table-driven tests with literal expectations proving that four-opening/two-storyboard books return opening indexes `0,1,2,3` in order, a retry selects only the newest successful job for its opening, a missing required opening returns an error, and a one-storyboard book returns only opening `0` even when the switch says four.

```go
selected, err := selectPublishMergeJobs(batch, book, jobs)
if err != nil { t.Fatal(err) }
got := []int{}
for _, job := range selected { got = append(got, job.OpeningVariantIndex) }
if !reflect.DeepEqual(got, []int{0, 1, 2, 3}) { t.Fatalf("indexes = %v", got) }
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd backend; go test ./internal/httpapi -run TestSelectPublishMergeJobs -count=1`

Expected: compile failure because `selectPublishMergeJobs` does not exist.

- [ ] **Step 3: Implement the smallest selector**

Resolve effective settings with `batchfactoryv11.ResolveSettings`, decode the public `openingEnabled` and `openingCount` JSON values inside the HTTP package, and require `openingEnabled == true`, `len(book.Videos) >= 2`, and a valid count from `1` through `8` before requiring multiple indexes; otherwise require only index `0`. Group successful non-empty merge outputs by `OpeningVariantIndex`, retaining the newest `UpdatedAt` per index, then return every required index in ascending order or fail before upload.

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `cd backend; go test ./internal/httpapi -run TestSelectPublishMergeJobs -count=1`

Expected: PASS.

---

### Task 2: Go One-Submission Multi-Video Upload

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_v11_publish_121_test.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_publish_121.go`

**Interfaces:**
- Consumes: `selectPublishMergeJobs`, `videomanagementconnector.VideoUploadService`.
- Produces: `marshalPublishedVideoHeads(heads []string) (string, error)` and one `Connector.Submit` call containing all heads.

- [ ] **Step 1: Write failing publisher tests**

Use a real in-memory Batch Factory store plus small connector/file-reader fakes at external boundaries. Assert literal observable results: four `UploadVideo` calls named `<bookID>-opening-0.mp4` through `<bookID>-opening-3.mp4`, one TXT submission named `<bookID>.txt`, and a `jieya_ai_head_video` field that decodes to four objects. Add a missing-opening test asserting zero uploads and zero submissions.

```go
var heads []map[string]any
if err := json.Unmarshal([]byte(connector.input.FormFields["jieya_ai_head_video"]), &heads); err != nil { t.Fatal(err) }
if len(heads) != 4 { t.Fatalf("head count = %d, want 4", len(heads)) }
```

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `cd backend; go test ./internal/httpapi -run 'TestBatchFactory121Publisher|TestMarshalPublishedVideoHeads' -count=1`

Expected: FAIL because only one latest merge is uploaded and/or the JSON-array helper is absent.

- [ ] **Step 3: Implement complete-set upload and JSON serialization**

Read all selected merge files only after selection succeeds, upload them in opening-index order, validate each returned head with `json.Unmarshal`, and marshal the collected values once with `encoding/json`. Call `Connector.Submit` exactly once after all uploads succeed. Use `<bookID>.mp4` for the compatible one-video case and `<bookID>-opening-N.mp4` for a multi-opening set.

- [ ] **Step 4: Persist an auditable multi-video history entry**

Store `videoCount`, `openingVariantIndexes`, `mergeJobIds`, and `videoFiles` in the confirmed `websiteSubmitHistory` entry. Return the same fields from `Submit` so the React page can show what was actually submitted.

- [ ] **Step 5: Run the focused tests and verify GREEN**

Run: `cd backend; go test ./internal/httpapi -run 'TestBatchFactory121Publisher|TestMarshalPublishedVideoHeads|TestPublishOrganization' -count=1`

Expected: PASS.

---

### Task 3: React Per-Opening Complete-Video Display

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`

**Interfaces:**
- Consumes: merge jobs with `openingVariantIndex`, `status`, `updatedAt`, and `outputUrl`.
- Produces: one visible current complete-video tile per opening index and history count limited to replaced attempts within an opening.

- [ ] **Step 1: Write a failing source contract test**

Assert that the list groups merges by opening index, derives all current merges, sorts by opening index, labels index `0` as `原开头` and later indexes as `换开头 N`, and computes history as total matching records minus the current record count. Remove the old assertion that expects one `latestBookMerge` upload source.

- [ ] **Step 2: Run the source test and verify RED**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: FAIL against the current single-merge rendering code.

- [ ] **Step 3: Implement grouped rendering**

Group book merge jobs by numeric `openingVariantIndex`, select the newest record from each group, render every selected record through `ProductionMediaBoundary`, and keep click-to-preview behavior. Show `原开头` / `换开头 N` and change the multi-opening marker from a single “current upload source” to `上传组成`.

- [ ] **Step 4: Run the source test and verify GREEN**

Run: `node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`

Expected: PASS.

---

### Task 4: Local Verification and Existing-Batch Acceptance

**Files:**
- No source file required unless a focused failure exposes a defect.

**Interfaces:**
- Consumes: local Go binary, MySQL state, batch `batch_27db6aa844ba1a62a504fdbbe409a0b2`.
- Produces: browser-visible four-video display and video-management readback evidence.

- [ ] **Step 1: Run focused and package tests**

Run separately:

```powershell
cd backend
go test ./internal/httpapi -run TestBatchFactory121 -count=1
go test ./internal/httpapi -count=1
go test ./internal/storage -count=1
cd ..
node --test frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js
```

Record the existing unrelated `internal/batchfactoryv11/mysql_intake_store_test.go` syntax blocker if it still prevents that package; do not edit it as part of this task.

- [ ] **Step 2: Build and restart the local Go review runtime**

Run:

```powershell
scripts\build-go-local.ps1
scripts\start-local-go-review.ps1 -UseYcmMysql -EnableBatchProduction -EnableLocalMerge
```

Verify `http://localhost:4001/shuihuo-production?run=latest` is served by the rebuilt Go binary.

- [ ] **Step 3: Verify display without regenerating videos**

Open the existing batch, inspect all four books, and confirm each shows four complete-video tiles. Query the existing task count before and after; it must remain `21`.

- [ ] **Step 4: Re-upload the existing four books and verify readback**

For each book, trigger explicit re-upload once. Confirm the local history entry reports `videoCount: 4`, four opening indexes, four merge IDs, and four filenames. Use the Go connector's target readback and confirm each video-management record contains four head-video objects plus the `<bookID>.txt` body submission.

- [ ] **Step 5: Inspect the final diff and keep local 4001 running**

Confirm only the planned source, tests, and documentation were added to the existing dirty worktree. Do not commit, push, or publish publicly.

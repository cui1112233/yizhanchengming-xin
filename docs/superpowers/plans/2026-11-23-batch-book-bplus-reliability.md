# Batch Book B+ Reliability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Batch Book processing stop permanent 121 failures immediately, keep long-running jobs under an owned Redis lease, prevent duplicate uploads, and expose safe consistency diagnostics and cleanup.

**Architecture:** Keep the existing Go/MySQL/Redis/TOS architecture. Add typed upstream failures at the provider boundary, token-owned renewable Redis leases at the queue boundary, and owner-scoped plus owner-admin aggregate diagnostics at the persistence boundary; orchestration consumes those interfaces without parsing error text or inventing frontend state.

**Tech Stack:** Go, MySQL 8 JSON columns, Redis lists/keys/Lua, React + Ant Design, Go `embed`, existing `novelfetchworkshop` and `httpapi` packages.

## Global Constraints

- Work directly on the current `v88` main development branch; do not create a branch.
- Preserve every unrelated dirty or untracked file and stage only files named by the active task.
- Several target files are already dirty before this plan (`novel_fetch_upload_batch_compat.go`, `novel_fetch_workshop_cleanup.go`, and cleanup tests). Record their pre-task diff before editing, merge around it, and use `git add -p` for those paths; never stage or rewrite their pre-existing hunks.
- Do not restore or execute `*_bak_testreset_20261121` data.
- Do not make paid model calls, real 121 calls, or real website uploads during tests.
- A 121 permanent business failure becomes manual-only on the first failure.
- Transient retries use 1, 5, 15, 30, and 60 minute delays, then become manual-only.
- Queue claim wait defaults to 30 seconds; execution lease defaults to 10 minutes and renews every 30 seconds.
- Upload completion requires a verified per-version receipt; an unknown result blocks blind resubmission.
- Orphan or unreferenced bodies must be at least 7 days old before being marked `releasable`; normal retention still applies before deletion.
- Backup tables are never scanned or cleaned automatically.
- Deploy with a single Go binary containing embedded frontend assets; do not use GitHub Actions.
- Release order is local source/test, local 4001, remote push, online deployment, then read-only online verification.
- The historical repository uses `backend/` as its Go API module and `frontend/` as its React user interface; this focused repair does not rename top-level directories.

## File Structure

- `backend/internal/novelfetchworkshop/upstream_error.go`: typed 121/HTTP failure classes and retry advice.
- `backend/internal/novelfetchworkshop/upstream.go`: construct typed failures at the upstream boundary.
- `backend/internal/novelfetchworkshop/process_lease.go`: lease interface, defaults, and lost-lease error.
- `backend/internal/novelfetchworkshop/redis_queue.go`: token-owned claim, renew, ack, fail, and stale recovery.
- `backend/internal/httpapi/novel_fetch_process_lease.go`: worker heartbeat and cancellation on lease loss.
- `backend/internal/httpapi/novel_fetch_automation.go`: permanent/config/transient retry decisions and backoff.
- `backend/internal/httpapi/novel_fetch_workshop.go`: preserve wrapped upstream errors and run work under the lease context.
- `backend/internal/novelfetchworkshop/diagnostics.go`: public diagnostic result types and lifecycle interface.
- `backend/internal/novelfetchworkshop/mysql_diagnostics.go`: owner-scoped counts and safe body reconciliation.
- `backend/internal/httpapi/novel_fetch_workshop_cleanup.go`: authenticated owner diagnostics and owner-admin aggregate route.
- `backend/internal/app/app.go`: periodic orphan-body marking before expired-body cleanup.
- Existing `*_test.go` files beside each component: focused regressions.

---

### Task 1: Typed upstream failure boundary

**Files:**
- Create: `backend/internal/novelfetchworkshop/upstream_error.go`
- Create: `backend/internal/novelfetchworkshop/upstream_error_test.go`
- Modify: `backend/internal/novelfetchworkshop/upstream.go:86-147`
- Test: `backend/internal/novelfetchworkshop/upstream_test.go`

**Interfaces:**
- Produces: `type UpstreamFailureClass string`
- Produces: `type UpstreamError struct { Class UpstreamFailureClass; Operation string; HTTPStatus int; BusinessCode int; Message string; Cause error }`
- Produces: `func ClassifyUpstreamFailure(error) (UpstreamFailureClass, bool)`
- Consumes: standard `errors.As`, `context`, and `net/http` status codes.

- [ ] **Step 1: Write failing classification tests**

```go
func TestClassifyUpstreamFailure(t *testing.T) {
	tests := []struct {
		name string
		err error
		want UpstreamFailureClass
	}{
		{"business 400", &UpstreamError{Class: UpstreamFailurePermanent, BusinessCode: 400}, UpstreamFailurePermanent},
		{"http 403", &UpstreamError{Class: UpstreamFailureConfiguration, HTTPStatus: 403}, UpstreamFailureConfiguration},
		{"http 429", &UpstreamError{Class: UpstreamFailureTransient, HTTPStatus: 429}, UpstreamFailureTransient},
		{"timeout", &UpstreamError{Class: UpstreamFailureTransient, Cause: context.DeadlineExceeded}, UpstreamFailureTransient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ClassifyUpstreamFailure(fmt.Errorf("fetch: %w", tt.err))
			if !ok || got != tt.want { t.Fatalf("got %q %v, want %q", got, ok, tt.want) }
		})
	}
}
```

- [ ] **Step 2: Run the focused test and confirm it fails**

Run: `go test ./internal/novelfetchworkshop -run TestClassifyUpstreamFailure -count=1`

Expected: FAIL because `UpstreamError` and `ClassifyUpstreamFailure` do not exist.

- [ ] **Step 3: Implement the typed error file**

```go
type UpstreamFailureClass string

const (
	UpstreamFailurePermanent     UpstreamFailureClass = "permanent"
	UpstreamFailureConfiguration UpstreamFailureClass = "configuration"
	UpstreamFailureTransient     UpstreamFailureClass = "transient"
)

type UpstreamError struct {
	Class UpstreamFailureClass
	Operation string
	HTTPStatus int
	BusinessCode int
	Message string
	Cause error
}

func (e *UpstreamError) Error() string {
	parts := []string{"121 upstream", strings.TrimSpace(e.Operation)}
	if e.HTTPStatus > 0 { parts = append(parts, fmt.Sprintf("HTTP %d", e.HTTPStatus)) }
	if e.BusinessCode > 0 { parts = append(parts, fmt.Sprintf("business %d", e.BusinessCode)) }
	if text := strings.TrimSpace(e.Message); text != "" { parts = append(parts, text) }
	return strings.Join(parts, ": ")
}
func (e *UpstreamError) Unwrap() error { return e.Cause }
func ClassifyUpstreamFailure(err error) (UpstreamFailureClass, bool) {
	var target *UpstreamError
	if !errors.As(err, &target) { return "", false }
	return target.Class, true
}
```

Construct classes in `UpstreamFetcher.Fetch` as follows: invalid input, HTTP 400/404, and business code 400 are permanent; HTTP 401/403 are configuration; timeout, connection/read failure, HTTP 408/429/5xx, invalid JSON, error pages, and empty temporary responses are transient. Preserve book and platform in the readable message but never include credentials or complete response bodies.

- [ ] **Step 4: Add HTTP and business-response construction tests**

Extend `upstream_test.go` with `httptest.Server` cases for 403, 429, 503, business code 400, invalid JSON, and an HTML error page. Assert both `errors.As(err, *UpstreamError)` and the expected class.

- [ ] **Step 5: Run the package tests**

Run: `go test ./internal/novelfetchworkshop -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 1 only**

```powershell
git add backend/internal/novelfetchworkshop/upstream_error.go backend/internal/novelfetchworkshop/upstream_error_test.go backend/internal/novelfetchworkshop/upstream.go backend/internal/novelfetchworkshop/upstream_test.go
git commit -m "fix(novel-fetch): classify upstream failures"
```

---

### Task 2: Permanent failures stop immediately and transient failures back off

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_automation.go:67-177`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go:360-390,2140-2150`
- Modify: `backend/internal/httpapi/novel_fetch_automation_retry_test.go:130-205`
- Test: `backend/internal/httpapi/novel_fetch_workshop_test.go`

**Interfaces:**
- Consumes: `novelfetchworkshop.ClassifyUpstreamFailure(error)` from Task 1.
- Produces: `func novelFetchOriginalRetryDelay(map[string]any, int) time.Duration`.
- Produces: persistent `manualOnly`, `manualOriginalFetchRequired`, `originalFetchRetryAttempts`, and `originalFetchNextRetryAt` behavior.

- [ ] **Step 1: Replace the old budget test with first-failure and backoff tests**

```go
func TestPermanentOriginalFetchFailureStopsImmediately(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(ctx, "owner", novelfetchworkshop.Document{BookID: "book-1", Meta: map[string]any{"status": "processing"}}); err != nil { t.Fatal(err) }
	err := fmt.Errorf("%w: %w", errNovelFetchOriginalFetch,
		&novelfetchworkshop.UpstreamError{Class: novelfetchworkshop.UpstreamFailurePermanent, BusinessCode: 400, Message: "获取书籍信息失败"})
	if err := markNovelFetchOriginalFetchRetry(ctx, store, "owner", "book-1", err, time.Now().UTC()); err != nil { t.Fatal(err) }
	doc, err := store.GetDocument(ctx, "owner", "book-1")
	if err != nil { t.Fatal(err) }
	if doc.Meta["status"] != "failed" || doc.Meta["manualOnly"] != true || intValue(doc.Meta["originalFetchRetryAttempts"]) != 1 { t.Fatalf("meta=%#v", doc.Meta) }
	if _, exists := doc.Meta["originalFetchNextRetryAt"]; exists { t.Fatal("permanent failure kept a retry time") }
}

func TestTransientOriginalFetchUsesBoundedBackoff(t *testing.T) {
	want := []time.Duration{time.Minute, 5*time.Minute, 15*time.Minute, 30*time.Minute, time.Hour}
	for attempt, delay := range want {
		if got := novelFetchOriginalRetryDelay(map[string]any{}, attempt+1); got != delay { t.Fatalf("attempt %d: %s", attempt+1, got) }
	}
}
```

- [ ] **Step 2: Run the focused tests and confirm failure**

Run: `go test ./internal/httpapi -run 'TestPermanentOriginalFetchFailureStopsImmediately|TestTransientOriginalFetchUsesBoundedBackoff' -count=1`

Expected: FAIL because permanent failure still waits for the fifth attempt and the backoff helper is absent.

- [ ] **Step 3: Preserve the typed cause through the workflow wrapper**

Change the fetch wrapper from `%v` to wrapping form:

```go
return fmt.Errorf("%w: %w", errNovelFetchOriginalFetch, err)
```

This keeps both `errors.Is(err, errNovelFetchOriginalFetch)` and `errors.As(err, *UpstreamError)` working.

- [ ] **Step 4: Implement the decision table**

In `markNovelFetchOriginalFetchRetry`:

```go
class, _ := novelfetchworkshop.ClassifyUpstreamFailure(fetchErr)
switch class {
case novelfetchworkshop.UpstreamFailurePermanent:
	return markNovelFetchOriginalManual(ctx, store, owner, bookID, fetchErr, now, "原文来源无效")
case novelfetchworkshop.UpstreamFailureConfiguration:
	return markNovelFetchOriginalConfiguration(ctx, store, owner, bookID, fetchErr, now)
}
// transient or untyped: increment attempt, schedule 1/5/15/30/60 minutes,
// and call markNovelFetchOriginalManual after attempt 5.
```

The manual helper clears every original-fetch and generic automatic retry timestamp. The configuration helper sets `status` and `queueState` to `waiting_configuration`, records the cause, and remains ineligible for automation until explicit manual retry.

- [ ] **Step 5: Run retry, automation, and workflow tests**

Run: `go test ./internal/httpapi -run 'NovelFetch|OriginalFetch|Automation' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 2 only**

```powershell
git add backend/internal/httpapi/novel_fetch_automation.go backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/novel_fetch_automation_retry_test.go backend/internal/httpapi/novel_fetch_workshop_test.go
git commit -m "fix(novel-fetch): stop permanent fetch retries"
```

---

### Task 3: Token-owned renewable Redis execution leases

**Files:**
- Create: `backend/internal/novelfetchworkshop/process_lease.go`
- Create: `backend/internal/httpapi/novel_fetch_process_lease.go`
- Create: `backend/internal/httpapi/novel_fetch_process_lease_test.go`
- Modify: `backend/internal/novelfetchworkshop/queue.go:10-45`
- Modify: `backend/internal/novelfetchworkshop/redis_queue.go:15-410`
- Modify: `backend/internal/novelfetchworkshop/redis_queue_test.go`
- Modify: `backend/internal/novelfetchworkshop/queue_test.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop.go:270-550`

**Interfaces:**
- Produces: `var ErrProcessLeaseLost error`.
- Produces: `type ProcessLeaseQueue interface { RenewLease(context.Context, ProcessJob, time.Duration) error; OwnsLease(context.Context, ProcessJob) (bool, error) }`.
- Extends: `ProcessJob` with `LeaseToken string ` + "`json:\"-\"`" + `.`
- Consumes: existing `ProcessQueue.Claim`, `Ack`, `Fail`, and stale reaper.

- [ ] **Step 1: Write Redis token ownership tests**

Add this command-contract test. In the same file, use table-driven fake Redis replies to cover: renew returns `1` only for the matching token; ack returns `0` for a wrong token and yields `ErrProcessLeaseLost` without any later `LREM` or `DEL`; the reaper reads a still-live matching token and returns a reaped count of zero.

```go
func TestRedisClaimUsesIndependentLeaseTTL(t *testing.T) {
	fake := &fakeRedisQueueDo{responses: []any{"job-1", `{"id":"job-1","owner":"alice","batchId":"batch-1"}`, "OK"}}
	queue := NewRedisQueueWithDo("qiantie:test:novel", fake.Do)
	job, err := queue.Claim(context.Background(), 30*time.Second)
	if err != nil { t.Fatal(err) }
	if job.LeaseToken == "" { t.Fatal("claim returned no lease token") }
	set := fake.commands[2]
	if set[0] != "SET" || set[len(set)-2] != "EX" || set[len(set)-1] != "600" { t.Fatalf("lease command=%v", set) }
}
```

`TestRedisRenewLeaseRequiresMatchingToken` must assert the renew command is a compare-and-expire Lua script. `TestRedisAckRejectsLostLease` returns zero from that Lua script and asserts `errors.Is(err, ErrProcessLeaseLost)` with no unconditional `LREM`/`DEL`. `TestRedisReaperKeepsMatchingLiveLease` returns the current token and asserts the reaped count is zero.

- [ ] **Step 2: Run the queue tests and confirm failure**

Run: `go test ./internal/novelfetchworkshop -run 'Redis.*Lease|Redis.*Reap' -count=1`

Expected: FAIL because claim still uses the wait duration as the lease TTL and leases have no token.

- [ ] **Step 3: Add the lease interface and defaults**

```go
const (
	DefaultProcessLeaseTTL = 10 * time.Minute
	DefaultProcessLeaseHeartbeat = 30 * time.Second
)

var ErrProcessLeaseLost = errors.New("novel fetch process lease lost")

type ProcessLeaseQueue interface {
	RenewLease(context.Context, ProcessJob, time.Duration) error
	OwnsLease(context.Context, ProcessJob) (bool, error)
}
```

Set `LeaseToken` with `crypto/rand` during claim. Store the token in the Redis lease key with `EX 600`. Implement renew, ack, fail, and stale recovery with Lua compare-and-act scripts; a stale worker must never delete another worker's lease or processing entry.

- [ ] **Step 4: Run the queue package tests**

Run: `go test ./internal/novelfetchworkshop -count=1`

Expected: PASS.

- [ ] **Step 5: Write failing worker heartbeat tests**

Use a fake lease queue with a buffered `renewed` channel and a controllable `renewErr`:

```go
type leaseProbeQueue struct {
	novelfetchworkshop.ProcessQueue
	renewed chan struct{}
	renewErr error
}

func (q *leaseProbeQueue) RenewLease(context.Context, novelfetchworkshop.ProcessJob, time.Duration) error {
	q.renewed <- struct{}{}
	return q.renewErr
}
```

In `TestWorkshopWorkerRenewsLeaseUntilCompletion`, block the fake rewrite provider, receive two values from `renewed`, release the provider, and assert one Ack. In `TestWorkshopWorkerCancelsWorkAfterLeaseLoss`, set `renewErr=novelfetchworkshop.ErrProcessLeaseLost`, wait for the worker context to be cancelled, and assert zero body writes, uploads, Acks, and retrying Fails.

- [ ] **Step 6: Implement the heartbeat helper**

`startWorkshopProcessLease` returns a child context and stop function. It renews every 30 seconds (tests inject a shorter interval), cancels the child context on lease loss, and reports the final lease error exactly once. `workshopProcessWorker.process` uses that child context for fetch, rewrite, persistence, and upload.

Before writing fetched or generated bodies, check `ctx.Err()`. Before upload, call `OwnsLease`; when ownership is false, return `ErrProcessLeaseLost` and do not submit.

- [ ] **Step 7: Run worker and race tests**

Run: `go test -race ./internal/novelfetchworkshop ./internal/httpapi -run 'Lease|Stale|WorkshopWorker' -count=1`

Expected: PASS with no race report.

- [ ] **Step 8: Commit Task 3 only**

```powershell
git add backend/internal/novelfetchworkshop/process_lease.go backend/internal/novelfetchworkshop/queue.go backend/internal/novelfetchworkshop/redis_queue.go backend/internal/novelfetchworkshop/redis_queue_test.go backend/internal/novelfetchworkshop/queue_test.go backend/internal/httpapi/novel_fetch_process_lease.go backend/internal/httpapi/novel_fetch_process_lease_test.go backend/internal/httpapi/novel_fetch_workshop.go
git commit -m "fix(novel-fetch): renew owned process leases"
```

---

### Task 4: Upload evidence and missing-only retry invariants

**Files:**
- Modify: `backend/internal/httpapi/novel_fetch_workflow_consistency_test.go`
- Modify: `backend/internal/httpapi/novel_fetch_schedule_partial_dispatch.go`
- Modify: `backend/internal/httpapi/novel_fetch_upload_batch_compat.go`
- Modify: `backend/internal/httpapi/novel_fetch_submit_diagnostic.go`
- Test: `backend/internal/httpapi/novel_fetch_submit_diagnostic_test.go`

**Interfaces:**
- Consumes: existing `siteSubmitStatus`, `siteSubmitDoneVersions`, `remote_receipt`, `autoSubmitReadyFingerprint`, and schedule reconciliation fields.
- Produces: one authoritative rule: `pending_verification`/`pending_confirmation` triggers readback and never direct upload.

- [ ] **Step 1: Add translated reference-invariant tests**

Use `MemoryStore` fixtures with selected versions `ai1,ai2`: (1) `ai1` has a verified receipt and unchanged fingerprint, so the upload spy receives no `ai1`; (2) `siteSubmitStatus=pending_verification`, so no upload job is created and the schedule stays `pending_verification`; (3) `ai1` is done and `ai2` failed, so the retry job contains only `ai2`; (4) `processSnapshot.auto_submit=false` with both generated versions derives `completed` without a submission receipt; (5) `force=true` explicitly resubmits an already confirmed `ai1` and appends an audit log containing the operation ID and prior receipt reference; (6) cancelling a schedule removes only its own `created_ids` and preserves pre-existing and cross-owner books.

- [ ] **Step 2: Run the four tests and record which existing guarantees already pass**

Run: `go test ./internal/httpapi -run 'ConfirmedVersion|UnknownSubmission|PartialSubmission|GenerateOnlyCompletion' -count=1`

Expected: existing covered behavior may pass; any failure identifies the exact minimal compatibility change.

- [ ] **Step 3: Make the minimal state changes**

Use the existing content fingerprint, `remote_receipt`, force flag, operation ID, manifest, and document log fields. Do not introduce a second submission ledger. Ensure schedule dispatch returns `pending_verification` when a selected version is unknown, calls readback reconciliation, and builds retry jobs only from selected versions not in `siteSubmitDoneVersions`. Forced resubmission must remain explicit and auditable; schedule cancellation must continue using `created_ids` plus owner checks.

- [ ] **Step 4: Run all submission and schedule tests**

Run: `go test ./internal/httpapi -run 'Submit|Upload|Schedule|Workflow' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 4 only**

```powershell
git add backend/internal/httpapi/novel_fetch_workflow_consistency_test.go backend/internal/httpapi/novel_fetch_schedule_partial_dispatch.go backend/internal/httpapi/novel_fetch_submit_diagnostic.go backend/internal/httpapi/novel_fetch_submit_diagnostic_test.go
git add -p backend/internal/httpapi/novel_fetch_upload_batch_compat.go
git commit -m "test(novel-fetch): enforce submission evidence"
```

---

### Task 5: Owner-scoped and owner-admin aggregate consistency diagnostics

**Files:**
- Create: `backend/internal/novelfetchworkshop/diagnostics.go`
- Create: `backend/internal/novelfetchworkshop/mysql_diagnostics.go`
- Create: `backend/internal/novelfetchworkshop/mysql_diagnostics_test.go`
- Modify: `backend/internal/novelfetchworkshop/lifecycle_store.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop_cleanup.go`
- Modify: `backend/internal/httpapi/novel_fetch_workshop_cleanup_test.go`

**Interfaces:**
- Produces: `type ConsistencyDiagnostics struct` with document, run, body, missing-ref, orphan, unreferenced, terminal-retry, and dangling-current-batch counts.
- Produces: `func (s *MySQLStore) GetConsistencyDiagnostics(context.Context, string, time.Time) (ConsistencyDiagnostics, error)`.
- Produces: `func (s *MySQLStore) GetAggregateConsistencyDiagnostics(context.Context, time.Time) (ConsistencyDiagnostics, error)` for an authenticated account with `BridgeIdentity.IsOwner=true` only.
- Produces: authenticated `GET /api/novel-fetch-workshop/diagnostics`.

- [ ] **Step 1: Write sqlmock aggregation tests**

```go
func TestMySQLConsistencyDiagnosticsCountsOwnerOnly(t *testing.T) {
	store, mock := newDiagnosticSQLMock(t)
	expectDiagnosticRows(mock, "alice", 4, 8, 12, 1, 2, 3, 1, true)
	got, err := store.GetConsistencyDiagnostics(context.Background(), "alice", time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC))
	if err != nil { t.Fatal(err) }
	if got.Documents != 4 || got.OrphanBodies != 2 || got.UnreferencedBodies != 3 || got.MissingBodyRefs != 1 || got.TerminalRetryMarkers != 1 || !got.DanglingCurrentBatch { t.Fatalf("got=%+v", got) }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatal(err) }
}
```

In `newDiagnosticSQLMock`, call `sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))`, wrap the DB in `MySQLStore`, and register `t.Cleanup(db.Close)`. In `expectDiagnosticRows`, create one `ExpectQuery` per count field and attach `WithArgs(owner)` to every owner-scoped query. Finish with `ExpectationsWereMet`, so any missing owner predicate or unexpected extra query fails the test.

Add three route tests: a normal identity gets its own counts; a normal identity requesting `?scope=all` gets HTTP 403; an identity with `IsOwner=true` gets aggregate counts. In every success body assert that counts are present but `document_json`, body content, owner names, API keys, passwords, and tokens are absent.

- [ ] **Step 2: Run diagnostics tests and confirm failure**

Run: `go test ./internal/novelfetchworkshop ./internal/httpapi -run 'ConsistencyDiagnostics|DiagnosticsRoute' -count=1`

Expected: FAIL because the store method and route do not exist.

- [ ] **Step 3: Implement the diagnostic types and queries**

```go
type ConsistencyDiagnostics struct {
	Documents int64 `json:"documents"`
	Runs int64 `json:"runs"`
	Bodies int64 `json:"bodies"`
	MissingBodyRefs int64 `json:"missingBodyRefs"`
	OrphanBodies int64 `json:"orphanBodies"`
	UnreferencedBodies int64 `json:"unreferencedBodies"`
	TerminalRetryMarkers int64 `json:"terminalRetryMarkers"`
	DanglingCurrentBatch bool `json:"danglingCurrentBatch"`
	CheckedAt string `json:"checkedAt"`
}
```

All owner SQL predicates include `owner_username=?`; aggregate SQL omits that predicate but is reachable only through the separate aggregate store method and the `IsOwner` authorization gate. Both methods target only active table names. Use JSON metadata only; never select `content_blob` or full `document_json` into the HTTP result.

- [ ] **Step 4: Register the authenticated read-only route**

Extend `LifecycleStore` with both diagnostics methods. In `registerNovelFetchCleanupRoutes`, read `BridgeIdentityFromContext`. With no scope, pass only `identity.Username` to the owner method. Accept only `scope=all`; require `identity.IsOwner`, otherwise return 403, and then call the aggregate method. Never accept an arbitrary owner username from query parameters.

- [ ] **Step 5: Run lifecycle and route tests**

Run: `go test ./internal/novelfetchworkshop ./internal/httpapi -run 'Diagnostic|Cleanup|StorageStatus' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 5 only**

```powershell
git add backend/internal/novelfetchworkshop/diagnostics.go backend/internal/novelfetchworkshop/mysql_diagnostics.go backend/internal/novelfetchworkshop/mysql_diagnostics_test.go backend/internal/novelfetchworkshop/lifecycle_store.go
git add -p backend/internal/httpapi/novel_fetch_workshop_cleanup.go backend/internal/httpapi/novel_fetch_workshop_cleanup_test.go
git commit -m "feat(novel-fetch): expose consistency diagnostics"
```

---

### Task 6: Delayed reconciliation of orphan and unreferenced bodies

**Files:**
- Modify: `backend/internal/novelfetchworkshop/diagnostics.go`
- Modify: `backend/internal/novelfetchworkshop/mysql_diagnostics.go`
- Modify: `backend/internal/novelfetchworkshop/mysql_diagnostics_test.go`
- Modify: `backend/internal/app/app.go:858-890`
- Test: `backend/internal/novelfetchworkshop/mysql_store_cleanup_test.go`

**Interfaces:**
- Produces: `func (s *MySQLStore) MarkUnreferencedBodiesReleasable(context.Context, time.Time, time.Duration, int) (BodyReconcileResult, error)`.
- Consumes: existing `CleanupExpiredBodies` and configured retention behavior.

- [ ] **Step 1: Write failing reconciliation tests**

Cover four rows: orphan older than 7 days is marked, unreferenced older than 7 days is marked, referenced body is untouched, and a 6-day-old body is untouched. Assert the update sets `state='releasable'`, `releasable_at=now`, and `expires_at=now+DefaultBodyRetentionDays`.

- [ ] **Step 2: Run the reconciliation tests and confirm failure**

Run: `go test ./internal/novelfetchworkshop -run 'MarkUnreferencedBodiesReleasable' -count=1`

Expected: FAIL because the method does not exist.

- [ ] **Step 3: Implement bounded candidate scanning**

Select at most 500 active-table body identities with `state='ready'` and `updated_at<=now-7d`. For each identity, load only the matching document metadata. Mark it releasable when the document is absent or `BodyRefs[versionID]` is absent. Never inspect tables whose names contain `_bak_`.

- [ ] **Step 4: Call reconciliation before normal expired cleanup**

In `startNovelFetchBodyCleanup`, run the marking method once per six-hour sweep, log only counts, and then call existing `CleanupExpiredBodies`. A marking failure logs an error and skips deletion for that sweep; it does not stop the service.

- [ ] **Step 5: Run storage and app tests**

Run: `go test ./internal/novelfetchworkshop ./internal/app -run 'Body|Cleanup|Reconcile' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit Task 6 only**

```powershell
git add backend/internal/novelfetchworkshop/diagnostics.go backend/internal/novelfetchworkshop/mysql_diagnostics.go backend/internal/novelfetchworkshop/mysql_diagnostics_test.go backend/internal/app/app.go backend/internal/novelfetchworkshop/mysql_store_cleanup_test.go
git commit -m "fix(novel-fetch): reconcile unreferenced bodies"
```

---

### Task 7: Full regression, embedded build, local acceptance, push, and online deployment

**Files:**
- Modify only if a test exposes a defect: files already named in Tasks 1-6.
- Verify: `frontend/dist/**` generated by the existing build pipeline and embedded by the Go binary; do not manually edit generated assets.

**Interfaces:**
- Consumes: all Tasks 1-6.
- Produces: verified local 4001 binary, pushed commit chain, deployed single online binary, and read-only acceptance evidence.

- [ ] **Step 1: Run all relevant backend tests**

Run from `backend`:

```powershell
go test ./internal/novelfetchworkshop ./internal/httpapi ./internal/app -count=1
go test -race ./internal/novelfetchworkshop ./internal/httpapi -run 'Lease|Retry|Submit|Schedule|Diagnostic|Cleanup' -count=1
```

Expected: PASS.

- [ ] **Step 2: Run frontend tests and build**

Run from `frontend` using the repository's current npm scripts:

```powershell
npm test
npm run build
```

Expected: tests pass and the production build completes without unresolved imports.

- [ ] **Step 3: Build the embedded Go binary**

From the repository root run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\build-go-local.ps1
```

This copies `frontend/dist` into `backend/web/dist`, builds `backend/qiantie-local.exe`, and writes the local runtime manifest. Do not manually edit or stage generated `dist` files.

- [ ] **Step 4: Start and verify local 4001**

First resolve the process listening on port 4001. Stop it only if `Win32_Process.ExecutablePath` is exactly `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`; otherwise stop and report the collision. Then run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-local-go.ps1 -ExpectedSHA (git rev-parse --short HEAD)
```

Also check:

```text
GET http://127.0.0.1:4001/health
GET http://127.0.0.1:4001/api/runtime-build-info
GET /api/novel-fetch-workshop/diagnostics with the existing authenticated session
```

Expected: health is true, runtime commit matches HEAD, the normal diagnostics request returns owner-scoped counts, an owner-admin `scope=all` request returns aggregate counts, and no task is enqueued.

- [ ] **Step 5: Inspect the exact commit range and push**

Confirm every commit contains only Task 1-6 files plus this approved plan/spec. Then run `git push origin v88` against `https://github.com/cui1112233/yzcm`; do not use GitHub Actions.

- [ ] **Step 6: Deploy the single binary**

The verified service target is `/opt/www/qiantie/qiantie` and the unit is `qiantie`. Before replacement, create and record a rollback copy:

```powershell
ssh root@115.190.156.223 'set -eu; ts=$(date +%Y%m%d-%H%M%S); rollback="/opt/www/qiantie/qiantie.rollback-$ts"; cp -a /opt/www/qiantie/qiantie "$rollback"; sha256sum "$rollback"; printf "ROLLBACK=%s\n" "$rollback"'
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\deploy-go-embedded.ps1 -Deploy -TargetHost 115.190.156.223 -TargetUser root -TargetPath /opt/www/qiantie/qiantie -ServiceName qiantie -KeepPreviousReleases 1
```

If build, upload, checksum, metadata sync, restart, or health verification fails, stop the release and keep the rollback file. Do not alter the backup tables.

- [ ] **Step 7: Perform read-only online acceptance**

Verify:

- `/health` succeeds;
- `/api/runtime-build-info` matches the pushed commit;
- formal novel-fetch document/body/run counts remain unchanged unless a real user independently created work;
- Redis pending, processing, job, lease, dedupe, and cancel counts are reported without mutation;
- diagnostics returns no secrets and does not trigger work;
- service logs contain no startup migration, lease, or cleanup errors.

- [ ] **Step 8: Record rollback information and final evidence**

Report the local commit, remote commit, online runtime commit, test commands, health result, diagnostics summary, and rollback binary path. Do not claim success from build output alone.

# Task 4 — fix round 1

Status: all three review findings implemented locally, ready for the original reviewer's scoped re-review.

- FIX_BASE: `9aa933d7abc9294e0942daf4df2eeacc373d4562`.
- Fix implementation commit: `2c695f64b845fe42deae1e396d1fde9e323e8058` on existing `v88`.
- Scope: R1 Important, R2 Important, R3 Minor only. No Task5–7 expansion or unrelated storage expected-count changes. No new branch/worktree, real database migration, provider calls, 4001 restart, deployment or push.

## R1 — waiting schedules blocked later eligible projects

Previous scanner selected only the earliest 20 candidates. Busy/day-cap/unknown schedules retained their old business due timestamps and therefore could permanently occupy those slots.

The scanner now captures all candidate pages before performing any settlement or new dispatch. Pages use a stable key cursor `(next_run_at, owner_username, schedule_id)` rather than an offset. MySQL uses binary owner/ID comparisons and ordering, matching the MemoryStore string comparison. A new internal `ListDueSchedulePage` method implements this in both stores; the previous `ListDueSchedules` method remains a first-page compatibility wrapper.

Capturing candidates first separates discovery from mutations caused by completion accounting or dispatch. The per-plan transaction still handles existing active rounds before attempting new creation. Fairness does not modify `NextRunAt`, cap eligibility, unknown-result state, or the one-overdue-round rule.

Behavior coverage: 21 independent projects; the first 20 separately exercise busy, daily cap, and unknown+paused active rounds. The 21st dispatches in the first scan, the second scan does not dispatch it again, and all waiting plans keep their original due timestamps. Each scenario runs with native MemoryStore paging and with MySQLStore SQLMock paging feeding the same domain fixture. The SQL variant requires a second query with all cursor arguments and the stable ordering, and verifies all expectations. This is more than a single Memory-only result assertion.

## R2 — stop missed old plans before first scan

Legacy round restoration is now shared between scanning and controls:

- Schedule: restore its old `LastRunID` to an active legacy round before pause/resume/stop logic.
- Batch: restore old `CurrentRunID` before controls, then inspect the round and follow `confirm:<sourceID>` even when the source checkpoint already has succeeded status.

Restoration, successor lookup, cancellation record, and plan state stay in the same existing project transaction. Cancellation notifications still wait until commit. A stopped legacy schedule retains the restored active round, so disabling it no longer removes it from later settlement scans. Saved terminal chapters are not cancelled; restored saved results remain countable once.

Behavior coverage: first control stop before any scan for legacy schedule and batch, both directly queued and already-confirmed blueprint successors; the actual queued successor becomes cancelled with `CancelRequested=true`. Separate saved-result tests stop old schedule/batch records, scan twice, verify exactly one saved chapter is counted, keep stopped status, and preserve the succeeded run.

## R3 — stopping an already completed batch returned a zero object

The completed early return now assigns the loaded batch to the response first. It preserves the existing ID/status/count and does not write or cancel completed work.

Behavior coverage: eight concurrent stop requests against a completed batch all return the existing ID, completed status, and saved count. A controlled completion/stop race holds the project transaction after completion is persisted; the waiting stop then returns that completed batch with its one saved chapter.

## Red evidence

Before implementation:

`go test ./internal/novelcreation -run 'TestAutomationFixRound1' -count=1`

```text
--- FAIL: TestAutomationFixRound1ScheduleWaitingDoesNotStarveLaterProject/busy
    21st eligible project starved behind busy: dispatched=0
--- FAIL: TestAutomationFixRound1ScheduleWaitingDoesNotStarveLaterProject/day_cap
    21st eligible project starved behind day_cap: dispatched=0
--- FAIL: TestAutomationFixRound1ScheduleWaitingDoesNotStarveLaterProject/unknown
    21st eligible project starved behind unknown: dispatched=0
--- FAIL: TestAutomationFixRound1LegacyStopRestoresAndCancels/schedule/confirmed=false
    legacy schedule round not restored
    first control stop missed current run: queued false
--- FAIL: TestAutomationFixRound1LegacyStopRestoresAndCancels/schedule/confirmed=true
    legacy schedule round not restored
    first control stop missed current run: queued false
--- FAIL: TestAutomationFixRound1LegacyStopRestoresAndCancels/batch/confirmed=true
    legacy batch successor not restored
    first control stop missed current run: queued false
--- FAIL: TestAutomationFixRound1StopCompletedBatchReturnsExistingValue
    completed stop returned zero/stale batch: ID:"" Status:"" CompletedChapters:0
FAIL qiantie/backend/internal/novelcreation
```

The complete output also reported the direct legacy batch's missing restored round and repeated the completed zero-value failure across its concurrent callers.

## Green and regression evidence

Focused new behavior, including the SQL page variants and saved-result tests:

`go test ./internal/novelcreation -run 'TestAutomationFixRound1' -count=1`

```text
ok qiantie/backend/internal/novelcreation 0.048s
```

Final full domain regression, also including the subsequently added controlled completion/stop race:

`go test ./internal/novelcreation -count=1`

```text
ok qiantie/backend/internal/novelcreation 0.120s
```

Focused HTTP regression:

`go test ./internal/httpapi -run 'Schedule|Batch|Automation' -count=1`

```text
ok qiantie/backend/internal/httpapi 26.228s
```

Scoped `git diff --check` passed. Existing Task1 continuity/version and Task3 durable journal/recovery tests remain in the full passing novelcreation suite. No HTTP contract or migration changed in this round.

SQL evidence is SQLMock executing MySQLStore page queries plus domain behavior through the shared test store. It does not claim live MySQL competing-process acceptance; that remains Task8. The previously documented unrelated storage aggregate expected-count mismatch remains untouched.

## Files in this round

- `backend/internal/novelcreation/automation_fix_round1_test.go`: red/green behavioral cases, SQL-backed paging, legacy stop/saved-result checks, concurrent stop and controlled completion race.
- `backend/internal/novelcreation/automation_state.go`: cursor ordering and shared legacy round restoration.
- `backend/internal/novelcreation/store.go`: internal page method.
- `backend/internal/novelcreation/memory_store.go`, `mysql_store.go`: equivalent stable keyset pagination.
- `backend/internal/novelcreation/scheduler.go`: gather complete candidate pages before settlement/dispatch; restore old round before controls.
- `backend/internal/novelcreation/batch.go`: restore old round before controls and preserve completed stop response.
- This report.

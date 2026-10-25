# Task 3 Report

Base: v88 `1767fec59257209a8c1434c49b9cb22d7ed62836`.

## Changes

- Added aspect ratio lookup, capture dimensions, and active/director camera selection with main camera fallback. Existing actor vector scale and control-to-array behavior remain covered.
- Added a serialized save queue that coalesces pending full snapshots, saves automatically, waits for in-flight and later snapshots on flush, and retains a retryable snapshot after failure.
- The brief's sample failure handler unconditionally restores the older failed snapshot, losing a newer pending snapshot. This implementation preserves the newer snapshot instead; a controlled asynchronous test covers that case.

## Red / Green

- Before implementation, the requested two-file `node --test` command failed because the queue module and aspect helper exports were missing.
- An additional `toString` unknown-ratio regression test failed against the initial lookup, returning the inherited function instead of `null`; lookup now checks own properties.
- Final targeted test command: `node --test src/user/components/director-desk/director-scene-state.test.js src/user/components/director-desk/director-save-queue.test.js` from `frontend`: 10 passed, 0 failed.

## Self Review / Boundaries

- Queue snapshots must represent the complete latest state; superseding an older failed snapshot with a newer partial patch would lose information. Callers must provide full snapshots.
- Failure remains visible as `error` and `flush()` retries; the queue does not silently retry a failed write on its own.
- These are helpers only; wiring into the Director Desk, frontend build, runtime/browser verification and deployment are outside Task 3 and were not performed.

## Review Follow-up: Error Callback Scheduling

- Red: a regression test showed that scheduling revision 2 synchronously from `onStatus('error')` left writes at `[1]` after the event loop turn, because its drain joined the failing in-flight promise.
- Fix: record the schedule version at failure; after the failed in-flight operation clears, start another drain only if a new schedule occurred after that failure. A failure without a new schedule remains pending for explicit retry.
- Green: the two targeted test files passed 12 tests, including automatic saving of the new snapshot and no automatic retries of an unchanged failed snapshot.
- Review: callers still need to submit full snapshots. A flush already waiting on the failed write still rejects; the newly scheduled snapshot is written separately after cleanup.

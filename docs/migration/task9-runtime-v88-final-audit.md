# Task 9 Runtime / v88 final behavior audit

Task 9 keeps the useful execution semantics from the v88 Batch Factory runtime, but does not copy its process-local implementation.

## v88 behavior retained

- Scheduled work becomes eligible when `run_at <= now`.
- A project Run fans out into independently tracked per-book work.
- Failed work is visible and can be retried without replaying books that already succeeded.
- Project status is derived from the latest durable state of its books.

## v88 implementation intentionally replaced

The historical v88 scheduler used JSON/local scheduler state and an in-process running set. Those mechanisms are not safe as the authoritative source after a process restart or when more than one Scheduler/Worker instance exists.

The new Go mainline uses:

- MySQL as the authoritative durable source for `Run`, `BookRun`, attempts, errors, execution owner/token and lease deadline.
- Redis only for Queue/Lease coordination.
- owner-safe Redis Lease operations with fencing-token checks instead of a dumb mutex.
- MySQL fencing CAS before Complete/Fail so a stale Worker cannot commit results.
- recovery from Redis loss, Worker crash and Go process restart using MySQL durable facts.
- bounded retries and unique database constraints for idempotency/concurrent retry.

Conclusion: this is **semantic inheritance + Go/runtime hardening**, not a line-for-line migration of the old Node/local scheduler. No second runtime table, second BookRun model, or Node production backend was introduced.

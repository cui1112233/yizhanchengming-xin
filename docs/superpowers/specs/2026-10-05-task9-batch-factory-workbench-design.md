# Task 9 Batch Factory V11 Workbench Design

## Goal

Complete Task 9 in the new mainline without reimplementing already-validated Task 9.2 work from PR #11, while keeping Go/MySQL as the business source of truth and using Redis only for queue/lock/runtime coordination.

## Scope and order

1. Finish Task 9.2 BatchProject list.
2. Finish Task 9.3 project detail / novel list.
3. Only after 9.2 and 9.3 are green, implement Task 9.4 execution infrastructure.
4. Do not implement Task 10–15.

## Existing implementation reuse

PR #11 (`task9-batch-project-source-ui`) is an implementation asset, not a branch to merge wholesale. Reuse its validated file-level behavior where still compatible with current `main`:

- `sources []string`
- `bookCount`
- `genders []string`
- `styles []string`
- latest `runStatus`
- Go/MySQL aggregate list query
- API JSON fields
- React + Ant Design Tag display
- store/API/frontend tests

Do not import unrelated divergent branch history. Port the minimal relevant diffs onto `feat/task9-batch-factory-workbench` from latest `main`.

## Task 9.2 contract

`GET /api/v1/batch-projects` returns deterministic project summaries backed directly by MySQL:

- `id`
- `intakeId`
- `name`
- `sources: []string`
- `bookCount: number`
- `genders: []string`
- `styles: []string`
- `runStatus`

Rules:

- `sources`, `genders`, `styles` are distinct deterministic collections; never take the first book as representative.
- `bookCount` is counted from real `books` rows linked by `batch_projects.intake_id -> books.intake_id`.
- latest run is selected by `run_at DESC, id DESC`.
- frontend renders collections as Ant Design `Tag` values.
- no LocalStorage, in-memory cache, or stale frontend state is a business source of truth.

### Task 9.2.9 consistency requirement

After Task 8 creates a BatchProject, the next visit to `/batch-factory` must issue `GET /api/v1/batch-projects`; the API must query MySQL and include the newly created row. The test must prove this behavior without depending on a frontend cache. The previously observed Testing Library DOM-cleanup/isolation problem is a test-harness defect, not a reason to rewrite working 9.2.4–9.2.8 business logic.

## Task 9.3 contract

Clicking a project opens the V11 workbench using real project data. The detail API/reader returns all books for the selected BatchProject and exposes at least:

- internal book row id
- Book ID / external book id
- title
- source / bookstore
- platformId
- gender
- style
- body fetch status
- persisted error message

No mock book data is allowed. Old v88 React components and interaction semantics may be referenced, but old Node gateways are not production dependencies.

## Task 9.4 contract

After 9.2 and 9.3 are stable, implement Go-only execution infrastructure:

- pending -> running transition
- per-book independent state
- one failed book does not stop the batch
- persisted real errors
- single-book retry
- idempotency / duplicate-start protection
- Scheduler for due runs
- Redis queue
- Redis distributed lock / lease
- worker crash/restart recovery

MySQL remains the durable authority for run/book-run state. Redis is transient infrastructure only. Recovery derives from durable MySQL state plus expiring Redis leases, not from Redis as a permanent ledger.

## Technical constraints

- Backend/API/Worker/Scheduler: Go
- Frontend: React + Ant Design
- Database: MySQL
- Migration: Goose
- Queue: Redis
- Distributed lock: Redis
- No new Node.js production backend
- No business logic moved to the browser
- No browser LocalStorage as business fact source

## TDD and verification

Every behavior change follows RED -> verify failure -> GREEN -> regression. Reused PR #11 production behavior must still be introduced behind tests on the formal branch; do not assume the divergent branch proves compatibility with current `main`.

Minimum final verification:

- all Go tests
- frontend tests
- frontend build
- admin build
- API contract tests
- MySQL store tests
- PR CI green

`TASKS.md` stays the migration progress source of truth. Until the repository completion rules are met, unfinished items use `🟡` or `⚠️`, not premature `[x]`.
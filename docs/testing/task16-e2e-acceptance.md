# Task 16 Browser E2E / Public Acceptance Harness

## Scope

This harness validates the real React application with Playwright. Local/PR tests may replace API responses with controlled route fixtures, but they do not mock or replace the application itself. Public smoke uses `PUBLIC_BASE_URL` and repository secrets. No public IP, password, token, provider secret, or DSN is committed.

Run locally:

```bash
cd 前台
npm install
npx playwright install chromium
npm run test:e2e
```

Run against public/ECS:

```bash
PUBLIC_BASE_URL=https://example.invalid E2E_USERNAME=... E2E_PASSWORD=... npm run test:e2e:public
```

Real 121 mutation is opt-in with `E2E_REAL_121=1` plus `E2E_121_BOOK_ID`, `E2E_121_SOURCE`, and `E2E_121_PLATFORM_ID`. Paid/real VIDEO providers remain Tier 2 and must never be required by ordinary PR CI.

## Result classes

- **GREEN**: capability exists in current `main` and must pass.
- **expected-failure**: capability exists, a real defect is proven, and the test still runs with Playwright `test.fail`. If it unexpectedly passes, CI fails so the blocker annotation can be removed.
- **pending/skip**: capability is not yet present in `main`; the test names the owning Task/subtask and must be activated after that Task merges.

## Known blockers

### TASK16-AUTH-001 First anonymous visit incorrectly reported as expired session

**Expected:** A user who has never logged in and visits `/batch-factory` sees the Login state without an expired-session warning.

**Actual:** `GET /api/auth/current-user` returns 401 → frontend attempts `POST /api/auth/refresh` → refresh returns 401 → `ycm:auth-unauthenticated` is broadcast → `AuthBoundary` renders `当前登录状态已过期，请重新登录。`.

**Reproduction:** Open a clean browser context with no session and navigate directly to `/batch-factory`.

**RED commit:** `358e8007ab4fd6b17900196caa6052009a69d9c5`.

**RED CI evidence:** workflow run `37307205455`, artifact `task16-playwright-local` / artifact ID `11343589205`.

**Screenshot:** `test-results/auth-session--auth-login-s-0eeb8-out-expired-session-warning/test-failed-1.png` in the RED artifact.

**Trace:** `test-results/auth-session--auth-login-s-0eeb8-out-expired-session-warning/trace.zip` in the RED artifact.

**Network chain:** evidence is recorded as method + sanitized URL pathname + HTTP status only. The known chain is `GET /api/auth/current-user 401` → `POST /api/auth/refresh 401`. Request/response headers and bodies are not written to Task 16 evidence.

**Console evidence:** the RED artifact contains the Playwright error context. The Task 16 evidence collector stores only redacted console errors and never records auth headers/cookies/body values.

**Owner:** Task 15 Auth business code. Task 16 must not fix it. After the Auth fix merges to `main`, sync this branch, remove `test.fail` for Case 1, and require normal GREEN.

### TASK16-AUTH-002 403 bootstrap incorrectly falls back to Login

Code inspection and acceptance test show `AuthBoundary` currently maps any `getCurrentUser()` rejection to `unauthenticated`. The required semantics are: 403 must not trigger refresh and must not automatically become Login. The E2E remains expected-failure until Task 15 Auth owns the behavior change.

## Evidence and redaction

Every Playwright test automatically attaches `task16-evidence.json` containing:

- full test title
- expected/actual Playwright status
- UTC timestamp
- current URL with query string removed from evidence
- redacted browser console errors
- redacted page errors
- failed network request method/path/failure reason
- HTTP response method/path/status for responses >=400
- secret-leak scan findings

The collector does **not** record request headers, response headers, cookies, Authorization values, passwords, request bodies, or response bodies. API response bodies are inspected in memory only for likely secret values; only the finding category is retained. Browser UI and console output are also scanned. Benign field names alone (for example `passwordRequired`) are not treated as leaked values.

## Current pending blockers

### Task 9.4 Runtime — 4 pending

1. pending → running Worker transition.
2. due `run_at` Scheduler execution.
3. queue/lock/idempotency duplicate protection.
4. Worker crash/restart recovery.

These remain `test.skip` with Task 9.4 subtask IDs until the runtime implementation merges to `main`.

### Task 14 VIDEO / Merge — 6 pending

1. provider status model and secret-safe response.
2. VIDEO Start/Poll fields and state machine.
3. VIDEO Retry/Cancel.
4. unconfigured provider resilience.
5. Merge queued → running → succeeded without re-generating successful VIDEO.
6. explicit ffmpeg-missing Merge error without API startup failure.

These remain `test.skip` with Task 14 subtask IDs until Task 14 merges to `main`.

## Acceptance matrix

| Flow | Automation | Public must test | Current classification |
| --- | --- | --- | --- |
| Login | Playwright UI | yes | GREEN except TASK16-AUTH-001 |
| Session restore | Playwright UI | yes | GREEN |
| 401 refresh | controlled browser API fixture | yes | GREEN |
| 403 semantics | controlled browser API fixture | yes | expected-failure TASK16-AUTH-002 |
| Novel fetch grouping/dedupe | Playwright UI | yes | GREEN |
| Real 121 | opt-in public smoke | yes | opt-in public |
| Shuihuo page safety | Playwright UI | yes | GREEN |
| BatchProject list/detail | Playwright UI | yes | GREEN |
| Task 10 settings Drawers | Playwright UI | yes | GREEN |
| Script | Playwright UI/API | yes | GREEN |
| Hook enabled/disabled | Playwright UI/API | yes | GREEN |
| Director | Playwright UI/API | yes | GREEN |
| H3 Director | browser API acceptance | yes | GREEN |
| Final Prompt | Playwright UI/API | yes | GREEN |
| Retry Stage | Playwright UI/API | yes | GREEN at Stage level; Runtime worker semantics pending Task 9 |
| Audio measurement | Playwright UI/API | yes | GREEN |
| matchAudio | Playwright UI/API + timeline validator | yes | GREEN |
| VIDEO | structure registered | yes | pending Task 14 |
| Merge | structure registered | yes | pending Task 14 |
| Publishing | repository/API permissions | yes | public acceptance still required |
| Permissions | auth/public API acceptance | yes | 403 bootstrap has TASK16-AUTH-002 |

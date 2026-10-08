# Task 6 User Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate authenticated V88 settings, member, profile and issue-log experiences to Go/MySQL facts without exposing credentials or weakening session, CSRF, capability, ownership or redaction boundaries.

**Architecture:** Add narrow Go read/write projections and migrations first, with one ownership/capability predicate shared by each resource. Add page-private React components and CSS that reproduce the legacy DOM hierarchy, consuming those projections through the existing request helper; retain current navigation and use the explicitly authorized minimal `RouterApp.jsx` route wiring only for `/profile` and `/member`.

**Tech Stack:** Go `net/http`, MySQL migrations, existing Cookie Session/CSRF/capability middleware, React, Ant Design, Vitest, Vite.

## Global Constraints

- Work only in `/Users/ming/Documents/Codex/yizhanchengming-xin-task6` on `main`; before every commit fetch and rebase `origin/main`.
- Do not modify `/Users/ming/Documents/ChatGPT/一战晟铭`.
- Do not modify `UserShell.jsx`, global CSS or `api.js`; `RouterApp.jsx` is authorized only to import and render `/profile` and `/member` page components.
- Never return or render keys, cookies, sessions, passwords, recovery codes, MFA secrets, pairing codes, provider raw errors, request headers or stacks.
- Every write is Cookie Session + same-origin CSRF + ownership/capability protected and audited; 401/403 behavior receives direct tests.
- Every commit is a runnable vertical slice, uses `[skip ci]`, and is pushed to `origin/main` only after Go tests, `CI=1 npm test -- --run`, and `npm run build` pass.

---

### Task 1: Safe workspace preference and runtime-status projection

**Files:**
- Create: `api/db/migrations/00018_task6_workspace_profile.sql`
- Modify: `api/internal/httpapi/workspace_settings_handlers.go`
- Modify: `api/internal/httpapi/workspace_settings_handlers_test.go`
- Modify: `前台/src/SettingsPage.jsx`
- Create: `前台/src/settings-page.css`
- Modify: `前台/src/SettingsPage.test.jsx`

**Interfaces:**
- Consumes: authenticated user ID and existing local-executor runtime service.
- Produces: `GET/PUT /api/v1/workspace/settings` with whitelisted preferences plus redacted TOS/provider/executor availability.

- [ ] Write Go tests that reject an untrusted setting key, reject a cross-origin `PUT`, and assert that unavailable status includes only a safe reason.
- [ ] Run the targeted Go test and verify it fails because the fields/projection do not exist.
- [ ] Add the migration and minimal handler projection; use `user_workspace_preferences` only for whitelisted user preferences, and derive runtime status server-side.
- [ ] Re-run the targeted Go test and verify it passes.
- [ ] Write a React test for restored preferences, unavailable runtime cards and no secret-bearing text.
- [ ] Run the React test, implement page-private legacy-derived card DOM/CSS, then re-run it green.
- [ ] Run targeted Go/React checks, fetch/rebase/test/build, commit a complete settings slice with `[skip ci]`, and push.

### Task 2: Profile facts, controlled avatar upload and authenticated writes

**Files:**
- Modify: `api/internal/httpapi/server.go`
- Create: `api/internal/httpapi/profile_handlers.go`
- Create: `api/internal/httpapi/profile_handlers_test.go`
- Create: `api/internal/profile/service.go`
- Create: `api/internal/profile/mysql_store.go`
- Create: `前台/src/ProfilePage.jsx`
- Create: `前台/src/ProfilePage.test.jsx`
- Create: `前台/src/profile-page.css`

**Interfaces:**
- Produces: read-safe `GET /api/v1/profile`, same-origin `PUT /api/v1/profile`, and MIME/size constrained `POST /api/v1/profile/avatar` returning a controlled TOS URL only.
- Consumes: session user, capability policy, object uploader and profile rows.

- [ ] Write failing Go handler/service tests for unauthenticated 401, cross-origin 403, field allowlist, unsafe avatar rejection, and redacted response fields.
- [ ] Implement storage/migration-backed profile facts, audit writes and controlled uploader boundary; represent unavailable TOS/email/MFA as safe status, never as verified.
- [ ] Write failing React tests for legacy-derived summary cards, edit modal, refresh restoration and safe unavailable states.
- [ ] Implement private profile DOM/CSS and state calls via existing `requestJSON`; only then add the authorized `/profile` RouterApp import/branch.
- [ ] Verify targeted/full tests and build, rebase, commit and push the complete profile slice.

### Task 3: Member-center and team projection

**Files:**
- Create: `api/internal/httpapi/member_handlers.go`
- Create: `api/internal/httpapi/member_handlers_test.go`
- Create: `api/internal/membercenter/service.go`
- Create: `api/internal/membercenter/mysql_store.go`
- Create: `前台/src/MemberPage.jsx`
- Create: `前台/src/MemberPage.test.jsx`
- Create: `前台/src/member-page.css`
- Modify: `前台/src/RouterApp.jsx`

**Interfaces:**
- Produces: `GET /api/v1/member-center`, capability-filtered `GET /api/v1/team/members`, and same-origin user-scoped notification read writes only where durable notification facts exist.

- [ ] Write failing Go tests for team ownership filtering, ordinary-user 403/limited views, unavailable usage/notification facts and no client-calculated totals.
- [ ] Add only factual MySQL projections; use an explicit `unavailable` state when a usage/notification source is absent.
- [ ] Write failing React tests for hero card role variants, skeleton/empty states, responsive class hierarchy and unavailable cards.
- [ ] Implement page-private legacy DOM/CSS and add only the authorized `/member` RouterApp import/branch; preserve every other router branch byte-for-byte.
- [ ] Verify, rebase, commit and push the complete member slice.

### Task 4: Issue summary, timeline and safe navigation

**Files:**
- Modify: `api/internal/httpapi/issues_handlers.go`
- Modify: `api/internal/httpapi/issues_handlers_test.go`
- Modify: `前台/src/IssuesPage.jsx`
- Modify: `前台/src/IssuesPage.test.jsx`
- Create: `前台/src/issues-page.css`

**Interfaces:**
- Extends: `GET /api/v1/issues?page=&limit=&source=&status=&q=` with summary counts from the complete, permission-filtered projection and safe navigation labels.

- [ ] Write failing handler tests proving summaries count all authorized matching rows rather than one page, request IDs/messages are sanitized, and unauthorized rows/navigation labels never leak.
- [ ] Implement one SQL predicate reused for count, summary and list; add only safe labels/navigation metadata.
- [ ] Write failing React tests for legacy-style summary cards/timeline, filters, pagination, request-ID detail and empty/loading/error states.
- [ ] Implement page-private timeline CSS/DOM using the server projection, retaining refresh and safe project navigation.
- [ ] Verify, rebase, commit and push the complete issues slice.

### Task 5: Final integration and LAN acceptance

**Files:**
- Modify only test files required by evidence.
- Create: `docs/parity/task6-lan-acceptance.md`

- [ ] Run `git fetch origin && git rebase origin/main`; resolve any Task 3 router split by retaining both route sets.
- [ ] Run focused Go tests, `go test ./...` from `api`, `CI=1 npm test -- --run`, and `npm run build` from `前台`.
- [ ] Start only the approved LAN staging environment, log in with staging-admin, save and refresh each setting/profile, exercise logout/401/cross-origin 403/ordinary-user restrictions, and capture desktop/mobile screenshots.
- [ ] Compare each screenshot to `docs/parity/task6-user-pages-matrix.md`; write factual pass/fail evidence and gaps without deploying ECS/public.
- [ ] Rebase, re-run verification, commit the acceptance evidence with `[skip ci]`, and push.

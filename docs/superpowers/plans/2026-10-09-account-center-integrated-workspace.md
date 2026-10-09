# 个人中心一体化工作区 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every personal-center route render in one independent workspace with its own navigation and a reliable return-to-workbench action.

**Architecture:** Extract account-center route and return-origin decisions into a small pure module, then render account-center routes through a dedicated layout component instead of the legacy business shell. The new layout owns the personal-center sidebar and content canvas, while `UserLayout` retains authentication, account loading, login overlay, logout, and ordinary business navigation.

**Tech Stack:** React, Ant Design, lucide-react, CSS container queries, Node built-in test runner, Vite.

**Spec:** `docs/superpowers/specs/2026-10-09-account-center-integrated-workspace-design.md`

## Global Constraints

- Work only on Git branch `v88`; do not create a branch or use GitHub Actions.
- Keep account data, permission checks, API calls, form behavior, and login/logout semantics unchanged.
- Do not hard-code system prompt text in the frontend.
- Account-center routes are exactly `/member`, `/profile`, `/security`, `/advanced-team-admin`, `/api-config`, `/usage`, `/team`, and `/accounts`.
- The return action restores the original same-origin path, query string, and hash; it must never return to another account-center route.
- If there is no valid origin, return to `/`.
- DEV-only and DEV/MANAGER-only menu entries must retain their current role gates.

## Review Focus

- Direct `/profile` visit or refresh: return action must resolve to `/`; covered by Task 1 unit tests.
- Source includes a query string and hash: return action must preserve both; covered by Task 1 unit tests.
- A stale personal-center URL is saved as origin: return action must fall back to `/`; covered by Task 1 unit tests.
- DEV, MANAGER, and MEMBER accounts: only the permitted navigation entries render; covered by Task 2 source contract tests and browser checks in Task 4.
- A 980px or narrower viewport: the personal-center navigation remains in the same layout and does not overlay content; covered by Task 3 stylesheet contract test and browser checks in Task 4.

---

## File Structure

- Create: `frontend/src/shared/layouts/accountCenterNavigation.js` — route classification, origin capture, validation, and fallback decisions with no React dependency.
- Create: `frontend/src/shared/layouts/accountCenterNavigation.test.js` — unit tests for direct entry, invalid origins, and source preservation.
- Create: `frontend/src/shared/layouts/AccountCenterLayout.jsx` — one page-level personal-center sidebar and content canvas.
- Create: `frontend/src/shared/layouts/AccountCenterLayout.source.test.js` — source contract for return action, role-gated menu entries, and no legacy shell.
- Modify: `frontend/src/shared/layouts/UserLayout.jsx` — capture origin only when leaving a business route and select `AccountCenterLayout` for account-center routes.
- Modify: `frontend/src/shared/layouts/UserLayout.source.test.js` — update the avatar route assertion and protect the new layout selection contract.
- Modify: `frontend/src/shared/styles/account-center-visual-rebuild.css` — replace fixed overlay/extra-margin rules with the integrated workspace layout and responsive behavior.
- Modify: `frontend/src/shared/layouts/AccountCenterLayout.source.test.js` — add the stylesheet source assertions after the layout and CSS exist.

### Task 1: Account-center navigation state

**Files:**
- Create: `frontend/src/shared/layouts/accountCenterNavigation.js`
- Create: `frontend/src/shared/layouts/accountCenterNavigation.test.js`

**Interfaces:**
- Produces: `ACCOUNT_CENTER_ROUTES`, `isAccountCenterPath(pathname)`, `normalizeAccountCenterReturnPath(value, origin)`, and `captureAccountCenterOrigin(currentUrl, existingOrigin, origin)`.
- Consumes: URL strings from `window.location` and the `qiantie:account-center-return-path` session-storage value.

- [ ] **Step 1: Write failing unit tests for the navigation module**

```js
test('captures a business URL once and retains its query and hash', () => {
  assert.equal(
    captureAccountCenterOrigin('/shuihuo-production?book=42#stage', null, ORIGIN),
    '/shuihuo-production?book=42#stage'
  );
});

test('returns home for direct entry, a foreign origin, or an account-center origin', () => {
  assert.equal(normalizeAccountCenterReturnPath(null, ORIGIN), '/');
  assert.equal(normalizeAccountCenterReturnPath('https://example.com/', ORIGIN), '/');
  assert.equal(normalizeAccountCenterReturnPath('/security', ORIGIN), '/');
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `node --test src/shared/layouts/accountCenterNavigation.test.js` from `frontend/`.

Expected: FAIL because the module does not exist.

- [ ] **Step 3: Implement the pure navigation helpers in `accountCenterNavigation.js`**

`captureAccountCenterOrigin(currentUrl, existingOrigin, origin)` must return the existing valid business origin when one exists; otherwise it must return the current same-origin business route including search and hash. `normalizeAccountCenterReturnPath(value, origin)` must accept only same-origin, non-account-center destinations and return `/` for every other input.

- [ ] **Step 4: Run the navigation unit tests to verify they pass**

Run: `node --test src/shared/layouts/accountCenterNavigation.test.js` from `frontend/`.

Expected: PASS.

- [ ] **Step 5: Commit the navigation state unit**

```bash
git add frontend/src/shared/layouts/accountCenterNavigation.js frontend/src/shared/layouts/accountCenterNavigation.test.js
git commit -m "feat(account): centralize center return navigation"
```

### Task 2: Dedicated personal-center workspace component

**Files:**
- Create: `frontend/src/shared/layouts/AccountCenterLayout.jsx`
- Create: `frontend/src/shared/layouts/AccountCenterLayout.source.test.js`

**Interfaces:**
- Consumes: `{ account, username, avatar, pathname, onReturn, onLogout, children }` from `UserLayout`.
- Produces: `AccountCenterLayout`, rendering a semantic `aside` personal-center sidebar and `main` content canvas with class names `account-center-workspace`, `account-center-sidebar`, and `account-center-main`.

- [ ] **Step 1: Write the failing source contract for the new layout**

```js
test('renders an integrated personal-center shell with a return action', () => {
  assert.match(source, /className="account-center-workspace"/);
  assert.match(source, /aria-label="返回工作台"/);
  assert.match(source, /onClick=\{onReturn\}/);
  assert.doesNotMatch(source, /legacy-sidebar/);
  assert.doesNotMatch(source, /account-center-popover/);
});

test('keeps current DEV and manager menu gates', () => {
  assert.match(source, /\['dev', 'manager'\]\.includes\(account\?\.role\)/);
  assert.match(source, /account\?\.role === 'dev'/);
});
```

- [ ] **Step 2: Run the source test to verify it fails**

Run: `node --test src/shared/layouts/AccountCenterLayout.source.test.js` from `frontend/`.

Expected: FAIL because the component does not exist.

- [ ] **Step 3: Implement `AccountCenterLayout` in `AccountCenterLayout.jsx`**

Reuse the exact visible labels, links, role conditions, account summary, and logout callback from the existing `account-center-popover`. Replace only the presentation boundary: use an ordinary in-layout sidebar, a “返回工作台” button wired to `onReturn`, and a sibling main content canvas. Do not render business navigation or a business top bar.

- [ ] **Step 4: Run the component source contract to verify it passes**

Run: `node --test src/shared/layouts/AccountCenterLayout.source.test.js` from `frontend/`.

Expected: PASS.

- [ ] **Step 5: Commit the workspace component**

```bash
git add frontend/src/shared/layouts/AccountCenterLayout.jsx frontend/src/shared/layouts/AccountCenterLayout.source.test.js
git commit -m "feat(account): add integrated center workspace"
```

### Task 3: Route integration and responsive layout replacement

**Files:**
- Modify: `frontend/src/shared/layouts/UserLayout.jsx`
- Modify: `frontend/src/shared/layouts/UserLayout.source.test.js`
- Modify: `frontend/src/shared/styles/account-center-visual-rebuild.css`
- Modify: `frontend/src/shared/layouts/AccountCenterLayout.source.test.js`

**Interfaces:**
- Consumes: Task 1 helper functions and Task 2 `AccountCenterLayout`.
- Produces: all account-center routes render only the integrated workspace; business routes keep `legacy-shell` unchanged.

- [ ] **Step 1: Add failing integration and stylesheet source assertions**

```js
test('routes account-center pages through the dedicated layout while the avatar records a business origin', () => {
  assert.match(userLayout, /captureAccountCenterOrigin/);
  assert.match(userLayout, /<AccountCenterLayout/);
  assert.doesNotMatch(userLayout, /accountCenterOpen/);
});

test('integrated CSS does not retain the fixed overlay or second content offset', () => {
  assert.doesNotMatch(css, /\.account-center-popover\s*\{\s*position:\s*fixed/);
  assert.doesNotMatch(css, /account-center-drawer-open .legacy-main[\s\S]*margin-left:\s*284px/);
  assert.match(css, /\.account-center-workspace\s*\{/);
});
```

- [ ] **Step 2: Run the integration test to verify it fails**

Run: `node --test src/shared/layouts/UserLayout.source.test.js src/shared/layouts/AccountCenterLayout.source.test.js` from `frontend/`.

Expected: FAIL because `UserLayout` still uses `accountCenterOpen` and the old fixed-overlay CSS remains.

- [ ] **Step 3: Replace account-center handling in `UserLayout.jsx`**

Import the Task 1 helpers and Task 2 component. When a user enters an account-center route from a business route, persist the first business URL under `qiantie:account-center-return-path`; do not replace it while navigating account-center routes. Pass a return handler that sends the validated saved origin to `window.location.assign`. Keep the avatar’s current client-side `/profile` navigation and add origin capture immediately before it. Remove only the account-center state, close handler, and fixed popover rendering; retain login, account loading, theme, and ordinary business shell code.

- [ ] **Step 4: Replace only the obsolete account-center shell CSS**

Define `account-center-workspace` as a viewport-height two-column grid or flex layout, with an in-flow 284px `account-center-sidebar` and `account-center-main` as the container-query root. Move the existing menu visual rules onto the new sidebar classes. Remove the rules that style account-center routes through `.legacy-sidebar`, `.legacy-main`, `.legacy-topbar`, `.account-center-popover`, and `margin-left: 284px`. At widths at or below 980px, keep the sidebar in flow and reduce it to `60px`; do not position it over content.

- [ ] **Step 5: Run focused source tests and the frontend test suite**

Run: `npm test` from `frontend/`.

Expected: PASS, including existing auth/route tests and the new account-center contracts.

- [ ] **Step 6: Commit the integration and stylesheet unit**

```bash
git add frontend/src/shared/layouts/UserLayout.jsx frontend/src/shared/layouts/UserLayout.source.test.js frontend/src/shared/layouts/AccountCenterLayout.source.test.js frontend/src/shared/styles/account-center-visual-rebuild.css
git commit -m "fix(account): render center as one workspace"
```

### Task 4: Build and browser acceptance

**Files:**
- Modify: no product files unless a failed acceptance check identifies a single documented defect.

**Interfaces:**
- Consumes: Tasks 1–3 completed on `v88`.
- Produces: evidence that the integrated layout and return path behavior work in the local application.

- [ ] **Step 1: Build the frontend**

Run: `npm run build` from `frontend/`.

Expected: Vite completes without an error.

- [ ] **Step 2: Verify business-origin return in the browser**

Open `/shuihuo-production`, enter personal center through the avatar, visit at least one personal-center subpage, then use “返回工作台”.

Expected: no business navigation remains inside personal center; the return action restores `/shuihuo-production` and any original query/hash.

- [ ] **Step 3: Verify direct-entry fallback and narrow layout in the browser**

Open `/profile` directly, refresh, then use “返回工作台”; repeat at a viewport no wider than 980px.

Expected: direct entry returns to `/`; personal-center navigation stays in flow and content is visible without a fixed overlay or extra blank gutter.

- [ ] **Step 4: Record the tested states and commit only if a single acceptance defect required a code change**

Run: `git status --short`.

Expected: no unintended files are staged or modified. If a targeted acceptance fix was necessary, add its failing test first, rerun `npm test` and `npm run build`, then commit it separately with an `fix(account): ...` message.

# Security / Secrets / Permission Final Audit

Status: **Stage A repository security audit complete against main `eb911e6a2249d58685e4bafe6ede52de154bb91f`; Stage B pending / blocked by Task9 Runtime merge. PR #21 must remain Draft.**

This report intentionally contains no real secret values. Secret findings, if any, are recorded only by type, path, commit/hash, severity and rotation status.

## 1. Scope and baseline

Stage A covers repository security from:

- base: `main` at `eb911e6a2249d58685e4bafe6ede52de154bb91f`;
- branch: `audit/security-permission-final`;
- sync commit: `204619604b0a4a30a93753bc3e33bedec8d49064`;
- post-cleanup code/security tree: `2abce456512b9caae91c3410123945bdd613e147`;
- subsequent changes before final exact-head verification are audit-document-only changes and do not modify production code or secret-bearing configuration.

The branch was synchronized to the Observability main without force-push. Conflict resolution briefly reintroduced several security regressions; those regressions were reproduced by tests, restored with targeted fixes, and reverified.

Stage A reviewed:

- Task15 login, refresh, logout, session cookies, capability checks and BatchProject ownership boundaries;
- Publishing account / credential / intent / audit authorization;
- Task14 provider configuration, outbound HTTP, artifact download, Local Executor identity/assignment and ffmpeg execution;
- Request IDs, structured logging/redaction, access logs, diagnostics, health/readiness, panic/500 handling and observability decorators;
- current-tree and Git-history secret exposure risk;
- repository ignore rules as preventive controls only.

Task9 Runtime is explicitly outside final Stage A closure. Redis queue/lease/fencing, Worker, Scheduler, Recovery, idempotency, retry authorization and runtime-internal paths remain Stage B work after Task9 enters main.

## 2. Findings summary

| ID | Severity | Component | Status | RED evidence | Fix commit | Residual risk |
|---|---|---|---|---|---|---|
| SEC-A-001 | HIGH | BatchProject ownership / project-scoped routes | FIXED | `ecf62bdcfc00399cd91742114d3b1f429a4402a7`: User B could access User A's project and receive HTTP 200 with project/book data | ownership boundary fixes plus post-sync restoration `cb6032f7cf1ab537a4353445a98ae244cd8a790f` | New project-scoped endpoints must reuse the same access boundary; Stage B must recheck new Task9 runtime routes |
| SEC-A-002 | HIGH | Auth backend failure semantics / logout revocation | FIXED | security RED suite and sync-regression CI showed backend failure mapped to 401 and revoke failure reported as 204 | targeted auth fixes plus sync restoration `0c0562104430a63f8dacd7f817bcb36735533866` | Future auth stores must preserve 401 vs 5xx distinction and fail closed on revoke errors |
| SEC-A-003 | MEDIUM | HTTP internal error leakage | FIXED | service error/redaction regressions | `6a0370a8...`, `0c035699...`, `0117473d...`, `8a1e2e17...` | New handlers must continue using safe error mapping rather than raw `err.Error()` |
| SEC-A-004 | MEDIUM | Audio / ffprobe path trust boundary | FIXED | `03a8b0a71c2698f9eef0da6339e9dedf39c3867a` | `59ca14d1...`, `58cfefeb...` | Any future browser-provided media path must remain constrained to server-managed assets |
| SEC-A-005 | HIGH | Provider/media SSRF | FIXED WITH DOCUMENTED BOUNDARY | provider SSRF RED tests including numeric/obfuscated loopback; final numeric-host RED `75dad55feeaa7de499f206d9360ba440b0b29702` | `e7652f17...`, `e8b8588c...`, `b4ca63db...`, `30713a93...`, `c6ff598131502b4c82f6f349e5a6e2e17ea688c5` | DNS/custom-transport changes remain security-sensitive; no claim of “100% SSRF safe” |
| SEC-A-006 | HIGH | Local Executor assignment isolation | FIXED | `31b3f449...`, `20502b51...` | `ee097bff...`, `cb758878...` | Stage B must combine this boundary with Task9 lease/fencing semantics |
| SEC-A-007 | MEDIUM | Stale Local Executor task mutation | FIXED; STAGE B RECHECK | `fdf0b9b39a48fba977406be384d6d68e033aa139` | `2fbf429eaace3f3cb6f3178d10d1d115e89013cc` | Task14 has no full executor credential revocation subsystem; Task9 fencing/lease generation must be audited in Stage B |
| SEC-A-008 | MEDIUM | Observability `slog.Any(struct/object)` redaction | FIXED | `b37029fe6a5930622d891492f53169e18e7e911c` | `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` | Future logger encoders/custom object types must preserve recursive sanitization |
| SEC-A-009 | MEDIUM | Diagnostics provider error/path disclosure | FIXED | `fd9db3cd50f9e21208f5f9f04a5c63dce93159cd` | `3bf8952df9c2b572bc96a7062353213727b6ea76` | New diagnostics fields require explicit DTO review and safe provider-error mapping |
| SEC-A-010 | MEDIUM | Observability-created resource read / IDOR risk | FIXED | `29046584ae5879ff10116d1418e8520d1f7b8401` | `67fd661ff994d9ebbfdd919b9963ebfcc1f0bea0` | Telemetry decorators must never add authorization-relevant resource reads |
| SEC-A-011 | MEDIUM | Provider secret configuration capability separation | FIXED | `8dba8006d3beefb7553f39d63353bb05e57a1926` | `35a4070b2201898f8ed843b0498f593e9183486f` | New provider configuration APIs must retain separate secret-management authorization |

## 3. Required HIGH finding: BatchProject cross-user IDOR

### RED

A regression test proved that an authenticated User B with the generic batch-view capability could request User A's BatchProject and receive HTTP 200. The response exposed the foreign project and associated book information.

RED commit: `ecf62bdcfc00399cd91742114d3b1f429a4402a7`.

### Minimal fix

Project-scoped HTTP routes were moved behind a reusable BatchProject access boundary that evaluates global/elevated access plus ownership/team authorization instead of capability-only access.

### GREEN

Cross-user / cross-team access is denied by the regression suite. List/detail and other project-scoped security tests remain in the normal Go test tree.

## 4. Observability finding: `slog.Any(struct/object)` redaction bypass

Finding ID: **SEC-A-008**.

Affected sensitive field-name examples include:

- `clientSecret`;
- `accessToken`;
- `refreshToken`.

The original sanitizer recursively handled maps and nested maps, but its default path returned unknown object types unchanged. A struct/object passed through `slog.Any(...)` could therefore bypass the map-oriented field-name redaction path.

RED commit: `b37029fe6a5930622d891492f53169e18e7e911c`.

Minimal fix: `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` extended object sanitization into a generic recursively redacted representation and avoids logging the original object if conversion cannot be made safe.

GREEN coverage includes case variants and nested values for `secret`, `Secret`, `SECRET`, `clientSecret`, `client_secret`, `accessToken`, `refreshToken`, DSN-like strings, Basic-Auth URLs, Bearer strings and query-like credentials.

## 5. Local Executor stale credential finding

Finding ID: **SEC-A-007**.

RED proved that a stale/offline Local Executor identity could still attempt terminal task mutation.

Fix commit: `2fbf429eaace3f3cb6f3178d10d1d115e89013cc`.

After the fix, stale/offline executor identity/credential state cannot `complete` or `fail` a task. Existing assignment-owner checks also prevent an old or foreign executor identity from overwriting the current assignment owner.

This control is deliberately marked for Stage B recheck: once Task9 Runtime is in main, the same guarantee must be proven together with lease generation/fencing so an old worker/executor token cannot Complete, Fail, Renew or Release after a newer owner takes the lease.

## 6. Observability security verification

### Request ID

Only one `X-Request-ID` value is accepted. It is capped at 128 characters and must use the accepted bounded character set. CR/LF, control characters, invalid punctuation, overlong IDs and multi-value headers are rejected and replaced by a server-generated request ID.

### Access log

The access logger records bounded request metadata only. It does not intentionally log query strings, request/response bodies, `Authorization`, `Cookie`, `Set-Cookie`, passwords, provider secrets, publishing credentials or executor bearer credentials.

### Diagnostics authorization

`GET /api/v1/diagnostics` requires authentication plus a global privileged role. Regression coverage pins:

- anonymous -> 401;
- ordinary authenticated user -> 403;
- `batch.view` only -> 403;
- global admin -> allowed;
- global owner -> allowed.

BatchProject ownership is stored separately from the global user role. Owning a project does not grant system diagnostics access.

### Safe diagnostics / health / panic

Diagnostics uses an explicit DTO and excludes token/token-hash, credentials, ciphertext, nonce, DSN, raw provider configuration and environment material. Provider `latest_safe_error` is sanitized instead of reflecting raw upstream bodies or sensitive paths.

`/healthz` and `/readyz` return safe status information. Panic/500 responses return a generic error plus request ID rather than stack/source path/SQL/environment/credential material.

### No observability-created IDOR

A RED test proved that a logging-only lifecycle path performed an unnecessary resource read. That telemetry-only read was removed; observability now records already-authorized operation results rather than widening the data-access surface.

## 7. Secret scan — Current Tree

Scope: the Stage A branch tree, including source, `.env`/`.env.*` paths, YAML, JSON, scripts, documentation, workflows, test fixtures, Provider configuration, Publishing configuration, Redis/MySQL/TOS/Auth-related strings and common key/token file types.

Evidence:

- the temporary audit workflow ran Gitleaks with redaction and a non-zero leak exit code policy;
- the scan used a full-history checkout (`fetch-depth: 0`), therefore it also covered the tracked tree represented by the PR merge ref at scan time;
- after that scan, the only branch changes were removal of the temporary audit workflow and audit-report edits; those diffs contain no credential values;
- the final tree contains no tracked `.env` file according to the exact-head tree listing;
- `.gitignore` prevents common local secret/key artifacts from being added, but is treated only as prevention.

Current-tree result:

**No confirmed production secret requiring rotation found**

| Finding type | Path | Commit/hash | Severity | Rotation status |
|---|---|---|---|---|
| Confirmed production secret | none | n/a | n/a | NOT REQUIRED |
| Test/example credential requiring action | none identified by the final secret scan | n/a | n/a | NOT REQUIRED |
| Unconfirmed secret-like material requiring escalation | none identified by the final secret scan | n/a | n/a | NOT REQUIRED |

## 8. Secret scan — Git History

The temporary audit workflow checked out complete repository history and executed Gitleaks with redaction:

`gitleaks git --redact --exit-code 1 .`

Evidence from the completed history job:

- **834 commits scanned**;
- approximately **2.20 MB** of Git content scanned;
- result: **no leaks found**;
- job conclusion: SUCCESS.

Git-history result:

**No confirmed production secret requiring rotation found**

`ROTATION_REQUIRED`: **No**.

If a future investigation proves that a real production credential was committed, it must be classified `ROTATION_REQUIRED`. Removing a value from Git history or adding an ignore rule is not credential remediation; the credential must be rotated/revoked at its authoritative source.

## 9. `.gitignore` prevention boundary

The current ignore rules cover common `.env` variants, secret/credential directories, private key formats, token/credential files and common key-store formats.

This is preventive control only. It does not make any previously committed credential safe and is not counted as historical-secret remediation.

## 10. SSRF control and residual boundary

Outbound provider/media controls reject loopback/private/link-local/metadata-style targets, IPv4-in-IPv6 private forms and non-canonical numeric/hex host forms; redirects are revalidated. Production transport resolution validates the resolved address before dialing an approved numeric IP.

There is an explicit local-development provider exception for loopback HTTP. Remote media download does not use that exception.

This report does **not** claim “100% SSRF safe.” DNS behavior, custom transports and future networking changes remain security-sensitive and require preservation of resolved-address and redirect validation.

## 11. ffmpeg / process execution

The merge path uses parameterized `exec.CommandContext(binary, args...)`. It does not build a shell string or invoke `sh -c` / `bash -c`.

Remote input is first downloaded to server-generated temporary names, output paths are server-controlled, and audio measurement accepts server-managed relative assets rather than arbitrary browser-supplied absolute paths. Path/argv regression tests remain in the normal Go suite.

## 12. Temporary security workflow cleanup

`.github/workflows/security-audit-temp.yml` was removed in commit:

`2abce456512b9caae91c3410123945bdd613e147`.

The temporary workflow is not part of the final Stage A tree.

Security regression tests themselves remain in normal package test files and are executed by the normal CI `go test ./...` job. Deleting the temporary workflow therefore does not remove the regression tests that protect the implemented security fixes.

## 13. Stage A final gate

Before freezing Stage A, the exact final branch head must satisfy all of the following on that same SHA:

- `go test ./...`;
- `go build ./...`;
- frontend tests;
- frontend build;
- admin build;
- Goose clean migration / reload / rollback;
- Task14 migration and restart/recovery coverage;
- Task16 E2E / Go regression;
- security regression tests through the normal Go suite;
- comparison with latest main reports `behind=0`;
- PR #21 remains Draft.

The authoritative `STAGE_A_HEAD=<sha>` is the final branch SHA after this report commit and is recorded in the Stage A completion status for PR #21. Earlier GREEN runs are supporting evidence only, not the final exact-head gate.

## 14. Stage B gate — Task9 Runtime Final Security Sweep

Stage B starts **only after Task9 Runtime is merged into main**. The audit branch must then sync latest main again before any new conclusion.

Stage B must cover:

- Redis credential/log/diagnostic safety and queue payload minimization;
- Redis as coordination only, never authorization truth;
- forged/replayed queue messages versus MySQL CAS authorization;
- Worker execution ownership/tokens;
- lease/fencing generation and stale Complete/Fail/Renew/Release denial;
- two-Scheduler duplicate-claim behavior;
- `run_at` authorization;
- idempotency-key authorization scope;
- retry authorization for project/BookRun;
- Redis wipe/replay recovery safety;
- Worker/Scheduler/internal endpoints;
- runtime diagnostics;
- Task14 Runtime Adapter interaction.

Until Stage B is complete, PR #21 must remain **Draft** and must not be marked Ready or merged.

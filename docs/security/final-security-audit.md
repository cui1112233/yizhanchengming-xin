# Security / Secrets / Permission Final Audit

Status: **Stage A complete against main `eb911e6a2249d58685e4bafe6ede52de154bb91f`; Stage B is blocked on Task9 Runtime merge. PR #21 must remain Draft.**

This report never contains real secret values. A future confirmed credential exposure must be recorded only by secret type and file/commit location, plus rotation status.

## 1. Stage A baseline and synchronization

- Initial audit baseline: `9f136e16deb2ea46bba4e2ececb1e63c829047eb`.
- Stage A Observability baseline: `eb911e6a2249d58685e4bafe6ede52de154bb91f` (`Observability / Logging / Operations Finalization`).
- Stage A sync commit: `204619604b0a4a30a93753bc3e33bedec8d49064`.
- GitHub reported conflicts for a direct main-to-audit merge. No force-push was used. A two-parent merge preserved both histories; post-sync CI then exposed security regressions from conflict resolution, which were restored with targeted fixes and retested.
- Post-cleanup code/security tree: `2abce456512b9caae91c3410123945bdd613e147`.
- At that tree, comparison with `eb911e6...` was `behind=0` / `ahead=71`.
- Normal CI run `37322619232` completed successfully: Go tests/build, frontend tests/build, admin build, Goose migration/restart/rollback coverage, and Task10 audit all passed.
- The temporary security workflow was removed before that final normal-CI evidence was recorded.
- PR #21 remains **Draft** and must not be marked Ready or merged until Stage B completes.

## 2. Stage A attack surface

Stage A reviewed current main plus Task14 Video and the Observability additions:

- Task15 login, refresh, logout, session cookies, capability checks, ownership and same-origin boundaries.
- BatchProject, Intake, Generation, unified settings and version-profile routes.
- Publishing account/credential/intent/audit authorization.
- Task14 provider configuration, outbound HTTP, artifact download, local executor identity/assignment, merge and ffmpeg execution.
- Request IDs, structured logging/redaction, access logs, diagnostics, health/readiness, panic/500 handling and lifecycle decorators.
- Current tree and full Git history for secrets, browser credential persistence and dependency vulnerabilities.

Task9 Runtime is not finalized in Stage A. Redis queue/lease/fencing, Scheduler, Worker, Recovery, idempotency and runtime-internal paths require Stage B after Task9 lands on main.

## 3. Findings and RED/GREEN evidence

| ID | Severity | Boundary | RED | Minimal fix | GREEN |
|---|---|---|---|---|---|
| SEC-A-001 | HIGH | BatchProject cross-user/cross-team IDOR | `ecf62bdcfc00399cd91742114d3b1f429a4402a7` | `fead0912...`, `0474b47b...`, `4a8232ed...`; sync restoration `cb6032f7cf1ab537a4353445a98ae244cd8a790f` | CI `37322619232` |
| SEC-A-002 | HIGH | Auth backend failure incorrectly treated as 401; logout revoke false-success | security RED tests; re-exposed by sync CI `37318185434` | `3a4be84c...`; sync restoration `0c0562104430a63f8dacd7f817bcb36735533866` | CI `37322619232` |
| SEC-A-003 | MEDIUM | Internal service error leakage | service error/redaction tests | `6a0370a8...`, `0c035699...`, `0117473d...`, `8a1e2e17...` | CI `37322619232` |
| SEC-A-004 | MEDIUM | ffprobe/local audio path traversal | `03a8b0a71c2698f9eef0da6339e9dedf39c3867a` | `59ca14d1...`, `58cfefeb...` | CI `37322619232` |
| SEC-A-005 | HIGH | Provider/media SSRF, including obfuscated numeric loopback | `b62a96e5...`; final numeric-host RED `75dad55feeaa7de499f206d9360ba440b0b29702` | `e7652f17...`, `e8b8588c...`, `b4ca63db...`, `30713a93...`, `c6ff598131502b4c82f6f349e5a6e2e17ea688c5` | CI `37322619232` |
| SEC-A-006 | HIGH | Local Executor assignment isolation | `31b3f449...`, `20502b51...` | `ee097bff...`, `cb758878...` | CI `37322619232` |
| SEC-A-007 | MEDIUM | Stale executor could complete/fail an assigned task | `fdf0b9b39a48fba977406be384d6d68e033aa139` | `2fbf429eaace3f3cb6f3178d10d1d115e89013cc` | CI `37322619232` and Task16 Go regression |
| SEC-A-008 | MEDIUM | `slog.Any` struct/object redaction bypass | `b37029fe6a5930622d891492f53169e18e7e911c` | `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` | full redaction regression set and CI `37322619232` |
| SEC-A-009 | MEDIUM | Diagnostics echoed provider upstream error/path text | `fd9db3cd50f9e21208f5f9f04a5c63dce93159cd` | `3bf8952df9c2b572bc96a7062353213727b6ea76` | CI `37322619232` |
| SEC-A-010 | MEDIUM | Observability performed a logging-only resource read | `29046584ae5879ff10116d1418e8520d1f7b8401` | `67fd661ff994d9ebbfdd919b9963ebfcc1f0bea0` | CI `37322619232` |
| SEC-A-011 | MEDIUM | Provider secret configuration capability separation | `8dba8006d3beefb7553f39d63353bb05e57a1926` | `35a4070b2201898f8ed843b0498f593e9183486f` | CI `37322619232` |

## 4. Observability security verification

### Request ID

Only one `X-Request-ID` value is accepted. It is capped at 128 characters and must match `[A-Za-z0-9._:-]`. CR/LF, tabs/control characters, invalid punctuation, overlong IDs and multi-value headers are rejected and replaced by a server-generated ID. This prevents attacker-controlled request IDs from forging multiline structured logs.

### Access log

The access logger records bounded metadata only: request ID, method, URL path, status, duration, remote-address class and authenticated user ID when present. It does not read/log query strings, request or response bodies, `Authorization`, `Cookie`, `Set-Cookie`, password, provider secret, publishing credential or executor bearer token.

### Redaction bypass coverage

Regression coverage includes `secret`, `Secret`, `SECRET`, `clientSecret`, `client_secret`, `accessToken`, `refreshToken`, nested maps/objects, `slog.Any` structs, error strings, DSN-like credentials, Basic-Auth URLs, Bearer strings and query-like credentials. Structured objects are converted to a generic JSON shape and recursively redacted; failed conversion does not write the original object.

### Diagnostics authorization

`GET /api/v1/diagnostics` requires authentication plus a **global** role:

- anonymous -> 401;
- ordinary authenticated user -> 403;
- `batch.view` only -> 403;
- global `admin` -> allowed;
- global `owner` -> allowed.

The diagnostics `owner` is `auth_users.role`. BatchProject ownership lives separately in `auth_batch_project_ownership`; being a project owner does not grant global diagnostics.

### Diagnostics DTO / health / readiness / panic

Diagnostics uses an explicit DTO and excludes token/token-hash, credential, ciphertext, nonce, DSN, environment and raw provider configuration. Provider `latest_safe_error` maps status to a fixed safe message and never returns raw upstream response text or sensitive filesystem path.

`/healthz` exposes only simple health state. `/readyz` returns safe code/message/request ID and does not return DB host, DSN, SQL error or provider/Redis/TOS configuration. Panic recovery returns a generic `INTERNAL_ERROR` plus request ID; stack/source path/SQL/env/credential material is not returned to the browser, and server-side panic text is sanitized.

### No observability-created IDOR

A RED test proved that intake lifecycle logging performed an extra `ListBooks` query solely for telemetry. That query was removed. Observability now logs already-authorized operation results instead of expanding resource reads.

## 5. Secret and history scan

The temporary Stage A workflow used a full-history checkout (`fetch-depth: 0`) and Gitleaks with redacted output and a failing exit code. It also ran Go reachable-vulnerability checks, frontend/admin production dependency audits and a browser-persistence source scan.

Result:

**No confirmed production secret requiring rotation found**

No secret value is reproduced here. If a future scan confirms that a real production credential was committed, mark it `ROTATION_REQUIRED`. Deleting a string from history or adding `.gitignore` is not sufficient; the credential must be rotated/revoked at its source.

`.gitignore` covers common `.env` variants, secret/credential files, private keys, token files and common key-store formats. These are preventive controls only, not historical-secret remediation.

## 6. SSRF controls and capability boundary

The current outbound policy:

- requires HTTPS for remote media;
- blocks literal loopback, RFC1918/private, link-local, unspecified, multicast and carrier-grade NAT addresses;
- blocks metadata-style link-local targets;
- normalizes/blocks IPv4-in-IPv6 private/loopback forms;
- rejects non-canonical all-numeric/hex-style host forms that can encode loopback;
- revalidates redirects;
- with the normal production `http.Transport`, resolves the host, rejects unsafe resolved addresses at dial time, and dials the approved numeric IP.

There is an explicit local-development provider exception for HTTP loopback (`localhost` / literal loopback). Remote media download does not use this exception.

This is **not** claimed to be “100% SSRF safe.” DNS behavior, custom injected transports and future networking changes remain security-sensitive. Any new outbound transport must preserve resolved-address validation and redirect validation.

## 7. Local Executor controls and limit

Verified:

- registration is protected by a server-side bootstrap token;
- executor credentials are random and only token hashes are persisted;
- unknown credentials cannot mutate guessed task IDs;
- executor A cannot complete/fail executor B's assignment;
- completion/failure requires a non-empty matching `ExecutorID`;
- executor failure text is sanitized;
- stale/offline executors beyond the heartbeat threshold cannot complete/fail assigned tasks.

Capability limit: Task14 currently has no first-class `revoked_at`/disabled/rotation state for executor credentials. Liveness gating blocks stale task mutation and unknown/deleted credentials are denied, but this is not a complete credential-revocation subsystem.

## 8. ffmpeg / process execution

The merge runner uses `exec.CommandContext(binary, args...)`. It does not construct a shell command or invoke `sh -c`/`bash -c`. Hostile filename content remains a literal argv value in regression tests.

Remote inputs are downloaded to server-generated temporary names before ffmpeg, output paths are server-generated, audio measurement accepts server-managed relative assets rather than arbitrary browser-supplied absolute paths, and remote downloads remain subject to SSRF and size limits.

## 9. Stage A completion checklist

- [x] sync Observability main `eb911e6...` without force-push;
- [x] resolve sync conflicts and restore security regressions with RED/GREEN evidence;
- [x] request ID, access-log, redaction, diagnostics, health/readiness, panic and observability-IDOR sweep;
- [x] full-history/current-tree secret scan;
- [x] SSRF regression including obfuscated numeric host forms;
- [x] Local Executor assignment/unknown/stale-credential regression;
- [x] ffmpeg shell/path/argument boundary review;
- [x] audit report created;
- [x] temporary security workflow removed;
- [x] main confirmed at `eb911e6...` and audit branch confirmed `behind=0` after cleanup;
- [x] normal CI `37322619232` fully green on the post-cleanup code/security tree;
- [x] PR #21 remains Draft.

## 10. Stage B gate: Task9 Runtime Final Security Sweep

Stage B starts **only after Task9 Runtime is merged to main**. Before any Stage B conclusion, sync the audit branch to that new main again.

Required Stage B checks:

### Redis

- credentials never enter logs or diagnostics;
- config/passwords are excluded from diagnostics;
- queue payload contains no unnecessary secret;
- Redis is coordination, not authorization fact;
- replay/flush/reconstruction cannot bypass MySQL authorization.

### Worker

A Redis message must never authorize execution by itself. Validate:

`Redis claim -> MySQL CAS -> current execution token/owner -> execute`

Forged or duplicated queue messages must fail unless MySQL state authorizes the claim.

### Lease / fencing

After Worker B obtains a newer lease/fencing generation, Worker A's old token must not be able to Complete, Fail, Renew or Release.

### Scheduler

Two schedulers must produce only one effective claim. User-controlled `run_at` may schedule only a resource already authorized to that user and must not authorize another project.

### Idempotency

Idempotency keys must be scoped inside the correct authorization/resource boundary. Knowing another user's key must not expose or reuse that user's Run.

### Retry

A user may retry only a project/BookRun for which current `batch.execute` authorization exists. Guessing a BookRun ID must not retry another user's task.

### Internal runtime paths

Review all Worker/Scheduler/internal endpoints for authentication, replay resistance, ownership binding, safe errors, request logging and secret handling.

Only after Stage B is RED/GREEN complete, latest main is synchronized, final CI is green, and this report is updated may PR #21 be considered for Ready / Merge.

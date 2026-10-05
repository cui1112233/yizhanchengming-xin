# Security / Secrets / Permission Final Audit

Status: **Stage A complete pending final post-cleanup CI; Stage B blocked on Task9 Runtime merge. PR #21 must remain Draft.**

This document intentionally does not contain secret values. Any future confirmed credential exposure must be recorded only by secret type and file/commit location, with rotation status tracked separately.

## 1. Audit baselines and gate

- Initial audit baseline: `9f136e16deb2ea46bba4e2ececb1e63c829047eb`.
- Stage A Observability baseline: `eb911e6a2249d58685e4bafe6ede52de154bb91f` (`Observability / Logging / Operations Finalization`).
- Stage A sync commit: `204619604b0a4a30a93753bc3e33bedec8d49064`.
- The audit branch and latest main had conflicts when GitHub attempted the direct merge. The branch was synchronized without force-push by creating a two-parent merge commit that preserved both histories. Subsequent CI exposed security regressions caused by conflict resolution; those regressions were restored with targeted security fixes and retested.
- After the Stage A sync, the branch was `behind=0` relative to `eb911e6...`. This must be rechecked immediately before declaring Stage A complete.
- PR #21 remains **Draft**. It must not be marked Ready or merged until Stage B completes after Task9 Runtime lands on main.

## 2. Attack surface reviewed in Stage A

Stage A covers the current main line plus Task14 Video and the Observability additions:

- Task15 login, refresh, logout, session cookies, capability checks, ownership checks and same-origin boundaries.
- BatchProject, Intake, Generation, unified settings and version-profile HTTP boundaries.
- Publishing account, credential, intent and audit authorization.
- Task14 Video provider configuration, provider outbound HTTP, artifact download, local executor registration/authentication/assignment, status, merge and ffmpeg execution.
- Structured logging, request correlation, access logs, diagnostics, health/readiness endpoints, panic/500 handling and service lifecycle decorators.
- Current-tree and full Git-history secrets, browser persistent storage, dependency vulnerabilities and ignore rules.

Task9 Runtime is explicitly **not** finalized in Stage A. Redis queue/lease/fencing, Scheduler, Worker, Recovery, idempotency and runtime-internal paths require Stage B after Task9 is merged to main.

## 3. Findings and remediation evidence

| Finding | Severity | Affected boundary | RED evidence | Minimal fix | GREEN evidence |
|---|---|---|---|---|---|
| SEC-A-001 BatchProject cross-user/cross-team IDOR | HIGH | `GET /api/v1/batch-projects`, `GET /api/v1/batch-projects/{id}` and project-scoped settings/generation routes | `ecf62bdcfc00399cd91742114d3b1f429a4402a7` | `fead0912e8537f90b4e2112f0a6c5f05c9a8619d`, `0474b47b198fba8a5822e8714cec854d9e913f2e`, `4a8232eda7f8fc17c982e4332119bd2aef4044e8`; post-main-sync restoration `cb6032f7cf1ab537a4353445a98ae244cd8a790f` | `go test ./...` + `go build ./...` green on post-sync CI run `37318906592`; final Stage A CI to supersede this |
| SEC-A-002 Auth backend failure mapped to 401 / logout revoke false-success | HIGH | login/refresh/auth middleware/logout | security audit RED tests on branch; behavior re-exposed by sync CI `37318185434` | `3a4be84cb4d9bd84d81bf5bd8f1fdc2f36089fdc`; post-sync restoration `0c0562104430a63f8dacd7f817bcb36735533866` | `37318906592` green; final Stage A CI pending |
| SEC-A-003 Internal service error leakage | MEDIUM | Intake, pipeline, generation, unified settings and provider-facing HTTP responses | service error/redaction tests | `6a0370a8f3d74f20d11a905e41780a5caa5fc271`, `0c0356996fbd954e8f949d69f19518c6adca9b46`, `0117473d10d6b24ff0ea4ab95246ff164c4f0360`, `8a1e2e1761f47fa018e0f659e6d7b6cf3bae9c63` | covered by branch Go suite; final Stage A CI pending |
| SEC-A-004 ffprobe local path traversal / unsafe browser path input | MEDIUM | audio measurement / ffprobe | `03a8b0a71c2698f9eef0da6339e9dedf39c3867a` | `59ca14d1d37937dc0d4e6d65ba03fd96048eb180`, `58cfefeb4c2867cc64b4d11b5ab56c87aa2f389a` | branch Go suite green after fixes; final Stage A CI pending |
| SEC-A-005 Provider/media SSRF including obfuscated numeric loopback | HIGH | provider endpoint, provider artifact download, merge media download | `b62a96e541f99c1a5e7fb642404db8ccc5032aa6`; final numeric-host RED `75dad55feeaa7de499f206d9360ba440b0b29702` | `e7652f173a09497314cd3b058f54b0208b13bd07`, `e8b8588c7637c1f5f330536db3a48abe4d9cfebf`, `b4ca63dbc7ceb390c496bd0a8b5f1af587f174a7`, `30713a939aaca4be5d1073aeaa251cae15d01e3f`, `c6ff598131502b4c82f6f349e5a6e2e17ea688c5` | final Stage A CI pending on current head |
| SEC-A-006 Local Executor assignment isolation | HIGH | executor complete/fail | `31b3f4499ee2158e44eef965fb9f65d7adb14bd5`, `20502b512f7eeb20a8040e01bda62b3c925dbbc2` | `ee097bff22e5dc90dae29f1173cdfbd8eecc8712`, `cb758878327fca921ca13b72e8884119ab120a85` | branch Go suite green; cross-executor/unknown-token regressions remain in suite |
| SEC-A-007 Stale executor credential could mutate assigned task | MEDIUM | executor complete/fail after heartbeat expiry | `fdf0b9b39a48fba977406be384d6d68e033aa139` | `2fbf429eaace3f3cb6f3178d10d1d115e89013cc` | current CI run `37322002751` pending at time of this draft |
| SEC-A-008 `slog.Any` struct/object secret redaction bypass | MEDIUM | structured logger | `b37029fe6a5930622d891492f53169e18e7e911c` | `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` | `api/internal/observability` green in subsequent runs; full regression set green on `fb135503053e5e9c4d1d8a5267a44582cf61d582` |
| SEC-A-009 Diagnostics provider error/path disclosure | MEDIUM | `GET /api/v1/diagnostics` | `fd9db3cd50f9e21208f5f9f04a5c63dce93159cd` | `3bf8952df9c2b572bc96a7062353213727b6ea76` | CI `37319968975` (`go test ./...`, `go build ./...`) green |
| SEC-A-010 Observability logging-only resource read | MEDIUM | Intake execution lifecycle logging | `29046584ae5879ff10116d1418e8520d1f7b8401` | `67fd661ff994d9ebbfdd919b9963ebfcc1f0bea0` | included in green full Go suite on `fb135503053e5e9c4d1d8a5267a44582cf61d582` |
| SEC-A-011 Provider secret capability separation | MEDIUM | provider configuration | `8dba8006d3beefb7553f39d63353bb05e57a1926` | `35a4070b2201898f8ed843b0498f593e9183486f` | covered by provider permission tests in branch Go suite |

## 4. Observability security verification

### Request ID

The HTTP observability middleware accepts only exactly one `X-Request-ID` value, with a maximum length of 128 and an allow-list of `[A-Za-z0-9._:-]`. CR/LF, tabs/control characters, invalid punctuation, overlong values and multi-value headers are rejected and replaced by a server-generated request ID. This prevents attacker-controlled request IDs from creating multiline/log-forging content in structured logs.

Regression coverage includes CRLF, control characters, overlong values and multiple header values.

### Access logs

The access logger records only bounded operational metadata: request ID, method, URL path, status, duration, remote-address class and authenticated user ID when present. It does **not** read or log request/response bodies, query strings, `Authorization`, `Cookie`, `Set-Cookie`, password fields, provider secrets, publishing credentials or executor bearer tokens.

The panic path also uses the same sanitizer and returns only a generic response plus request ID.

### Redaction bypass coverage

Redaction tests cover:

- `secret`, `Secret`, `SECRET` key normalization;
- `clientSecret`, `client_secret`;
- `accessToken`, `refreshToken`;
- nested maps and structured objects passed via `slog.Any`;
- error strings;
- DSN-like credentials;
- Basic-Auth URLs;
- Bearer strings;
- query-like credential strings.

The logger now recursively converts structured `slog.Any` structs to a generic JSON shape and applies key-based recursive redaction. Unsupported struct conversion fails closed instead of writing the original value.

### Diagnostics authorization

`GET /api/v1/diagnostics` is protected by authentication plus a **global** role check:

- anonymous: 401;
- ordinary authenticated user: 403;
- user with only `batch.view`: 403;
- global `admin`: allowed;
- global `owner`: allowed.

The `owner` used here is `auth_users.role`. BatchProject ownership is stored separately in `auth_batch_project_ownership`; owning a project does not grant global system diagnostics.

### Diagnostics DTO

Diagnostics uses an explicit DTO. Tests reject secret-like fields including token/token-hash, credentials, ciphertext, nonce, DSN and environment data. Provider `latest_safe_error` is mapped from provider status to a fixed message and never echoes the upstream error body or sensitive server path.

### `/healthz`, `/readyz`, panic/500

- `/healthz` returns only a simple health state and no database/provider/Redis/TOS configuration.
- `/readyz` returns a fixed safe code/message/request ID; database connection errors remain server-side and do not expose host, DSN, username/password or SQL error text to the browser.
- panic recovery returns generic `INTERNAL_ERROR` plus request ID. Stack/source path/SQL/env/credential data is not copied into the HTTP body; logged panic text is sanitized.

### Observability and authorization ordering

Observability must not expand the resource read surface. A RED test proved that Intake lifecycle logging performed an extra `ListBooks` read solely for logging. That read was removed. Lifecycle logging now observes already-authorized results instead of making new resource queries for telemetry.

## 5. Secrets and credential handling

### Current tree and full history

The temporary Stage A security workflow checked out full history (`fetch-depth: 0`) and ran Gitleaks over `--all` history with redacted output and a failing exit code on detection. The security workflow also ran Go reachable-vulnerability checks and frontend/admin production dependency audits.

Result:

**No confirmed production secret requiring rotation found**

No real credential value is reproduced in this report.

If a future scan confirms that a real production credential was ever committed, the finding must be marked `ROTATION_REQUIRED`. Removing a string from current Git history or adding an ignore rule does **not** make an already-exposed credential safe; the credential must be rotated/revoked at its source.

### `.gitignore`

The repository ignores common local secret material including `.env` variants, secret/credential files, private keys, token files and common key-store formats. These rules are preventive controls only; they are **not** evidence that historical secrets have been remediated.

## 6. SSRF capability and limits

The current outbound policy:

- requires HTTPS for remote media;
- blocks literal loopback, RFC1918/private, link-local, unspecified, multicast and carrier-grade NAT addresses;
- blocks metadata-style link-local targets;
- blocks IPv4-in-IPv6 private/loopback representations through IP normalization;
- rejects non-canonical all-numeric / hexadecimal-style host forms that can encode loopback such as decimal or hex IPv4 variants;
- revalidates redirects;
- for normal `http.Transport`, resolves the host and rejects unsafe resolved addresses at dial time, then dials the approved numeric address.

There is an explicit local-development exception for provider endpoints using HTTP loopback (`localhost` / literal loopback) where the code opts into that mode. Remote media download paths do not use that exception.

This is **not** documented as “100% SSRF safe.” DNS behavior, custom injected transports and future networking changes remain security-sensitive. The default production transport mitigates rebinding by validating resolved addresses at dial time and dialing the resolved IP, but every new outbound transport or resolver path must preserve that invariant.

## 7. Local Executor security and limits

Verified controls:

- registration uses a server-side bootstrap token boundary;
- executor bearer credentials are random and only hashes are persisted;
- unknown credentials cannot mutate guessed task IDs;
- executor A cannot complete/fail executor B's assignment;
- a task must have a non-empty matching `ExecutorID` before executor completion/failure;
- executor failure text is sanitized;
- stale/offline executors outside the heartbeat threshold cannot complete/fail assigned tasks after `2fbf429...`.

Capability limit: Task14 currently has no first-class `revoked_at`/disabled/rotation state for executor credentials. Liveness expiration now blocks terminal task mutation, and unknown/deleted credentials are denied, but this must not be described as a complete credential-revocation subsystem. If long-lived executor revocation is required, it should be implemented explicitly rather than inferred from heartbeat state.

## 8. ffmpeg / media process boundary

The merge runner uses `exec.CommandContext(binary, args...)`; it does not construct a shell command and does not invoke `sh -c` or `bash -c`. Tests verify hostile filename content remains an argv item rather than shell syntax.

Remote inputs are downloaded to server-generated temporary filenames before ffmpeg execution, and output paths are also server-generated. Audio measurement restricts ffprobe to server-managed local relative assets, preventing browsers from submitting arbitrary absolute server paths. Remote downloads are additionally subject to the SSRF policy and size limits.

## 9. Dependency and client-storage checks

The Stage A temporary security workflow successfully ran:

- full-history Gitleaks with redacted findings;
- `govulncheck ./...` on the patched Go toolchain;
- user frontend production dependency audit;
- admin frontend production dependency audit;
- source scan preventing browser `localStorage`, `sessionStorage` and `indexedDB` from becoming a credential/config persistence source in the audited code path.

The temporary workflow is an audit aid and must be removed before Stage A's final CI evidence is recorded.

## 10. Stage A completion checklist

Before marking Stage A complete:

- [x] sync Observability main `eb911e6...` into audit branch without force-push;
- [x] restore security regressions exposed by the post-sync `go test ./...` run;
- [x] Observability request ID, access log, redaction, diagnostics, health/readiness, panic and resource-read sweep;
- [x] full-history/current-tree secret scan;
- [x] SSRF final regression including obfuscated numeric host forms;
- [x] Local Executor assignment/unknown/stale-credential regression;
- [x] ffmpeg shell/path/argument boundary review;
- [x] initial audit report;
- [ ] remove temporary security CI;
- [ ] confirm latest main still synchronized (`behind=0`);
- [ ] normal CI fully green on the post-cleanup head;
- [ ] keep PR #21 Draft.

## 11. Stage B gate: Task9 Runtime Final Security Sweep

Stage B starts only after Task9 Runtime is merged to main. The security branch must synchronize main again before any Stage B conclusion.

Required Stage B checks:

### Redis

- Redis credential never enters logs or diagnostics.
- Redis configuration and passwords are excluded from diagnostics.
- Queue payload contains no unnecessary secret.
- Redis is coordination, not an authorization source.
- Queue replay/flush/reconstruction cannot bypass MySQL authorization state.

### Worker authorization

A Redis message must never authorize execution by itself. Validate the invariant:

`Redis claim -> MySQL CAS -> current execution token/owner -> execute`

Forged or duplicated queue messages must fail unless MySQL state authorizes the claim.

### Lease and fencing

After Worker B obtains a newer lease/fencing generation, Worker A's old token must not be able to:

- Complete;
- Fail;
- Renew;
- Release.

### Scheduler

- Two schedulers must permit only one effective claim.
- User-controlled `run_at` can schedule only resources already authorized to that user; it must never become an authorization bypass for another project.

### Idempotency

Idempotency keys must be scoped inside the correct authorization/resource boundary. Knowing another user's idempotency key must not reveal, return or reuse that user's Run.

### Retry

A user may retry only a project/BookRun for which the user currently has the required `batch.execute` authorization. Guessing a BookRun ID must not allow retrying another user's task.

### Internal runtime endpoints

Review all Worker/Scheduler/internal runtime paths for authentication, replay resistance, ownership binding, safe errors, request logging and secret handling.

Only after this Stage B sweep is RED/GREEN complete, latest main is synchronized, final CI is green and this report is updated may PR #21 be considered for Ready / Merge.

# Security / Secrets / Permission Final Audit

Status: **Stage A repository security audit complete; Stage B pending Task9 Runtime merge. PR #21 must remain Draft.**

This report contains no real secret values. Secret evidence is recorded only by finding type, path, commit/hash, severity and rotation status.

## Scope

Stage A baseline:

- `main`: `eb911e6a2249d58685e4bafe6ede52de154bb91f`
- branch: `audit/security-permission-final`
- Stage A sync commit: `204619604b0a4a30a93753bc3e33bedec8d49064`
- temporary security workflow removal: `2abce456512b9caae91c3410123945bdd613e147`
- final exact-head SHA is recorded in the PR completion status after this report commit and its exact-head CI finishes.

Stage A covers Task15 Auth/Publishing authorization, BatchProject ownership, Task14 Video Provider/Local Executor/ffmpeg boundaries, Observability/Logging/Diagnostics, current-tree secrets and full Git history secrets. Task9 Runtime is not finalized here.

## Findings Summary

| ID | Severity | Component | Status | RED evidence | Fix commit | Residual risk |
|---|---|---|---|---|---|---|
| SEC-A-001 | HIGH | BatchProject cross-user/cross-team access | FIXED | `ecf62bdcfc00399cd91742114d3b1f429a4402a7` | ownership boundary fixes; sync restoration `cb6032f7cf1ab537a4353445a98ae244cd8a790f` | New Task9 project/runtime routes must reuse the same boundary |
| SEC-A-002 | HIGH | Auth backend failure / logout revoke semantics | FIXED | security RED suite; sync regression reproduced in CI | targeted auth fixes; sync restoration `0c0562104430a63f8dacd7f817bcb36735533866` | Future auth stores must preserve 401 vs 5xx semantics |
| SEC-A-003 | MEDIUM | Internal error leakage | FIXED | handler/service redaction regressions | `6a0370a8...`, `0c035699...`, `0117473d...`, `8a1e2e17...` | New handlers must not return raw internal errors |
| SEC-A-004 | MEDIUM | Audio/ffprobe path trust boundary | FIXED | `03a8b0a71c2698f9eef0da6339e9dedf39c3867a` | `59ca14d1...`, `58cfefeb...` | Future media paths must remain server-managed |
| SEC-A-005 | HIGH | Provider/media SSRF | FIXED WITH DOCUMENTED BOUNDARY | SSRF RED suite; numeric-host RED `75dad55feeaa7de499f206d9360ba440b0b29702` | `e7652f17...`, `e8b8588c...`, `b4ca63db...`, `30713a93...`, `c6ff598131502b4c82f6f349e5a6e2e17ea688c5` | DNS/custom transport changes remain security-sensitive |
| SEC-A-006 | HIGH | Local Executor assignment isolation | FIXED | `31b3f449...`, `20502b51...` | `ee097bff...`, `cb758878...` | Recheck with Task9 lease/fencing in Stage B |
| SEC-A-007 | MEDIUM | Stale Local Executor task mutation | FIXED; STAGE B RECHECK | `fdf0b9b39a48fba977406be384d6d68e033aa139` | `2fbf429eaace3f3cb6f3178d10d1d115e89013cc` | Task9 fencing must reject old Complete/Fail/Renew/Release tokens |
| SEC-A-008 | MEDIUM | Observability `slog.Any(struct/object)` redaction | FIXED | `b37029fe6a5930622d891492f53169e18e7e911c` | `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` | Future logger encoders/custom objects must preserve recursive sanitization |
| SEC-A-009 | MEDIUM | Diagnostics provider error/path disclosure | FIXED | `fd9db3cd50f9e21208f5f9f04a5c63dce93159cd` | `3bf8952df9c2b572bc96a7062353213727b6ea76` | New diagnostic fields require explicit safe DTO review |
| SEC-A-010 | MEDIUM | Observability-created resource read / IDOR risk | FIXED | `29046584ae5879ff10116d1418e8520d1f7b8401` | `67fd661ff994d9ebbfdd919b9963ebfcc1f0bea0` | Telemetry must not add authorization-relevant reads |
| SEC-A-011 | MEDIUM | Provider secret configuration capability separation | FIXED | `8dba8006d3beefb7553f39d63353bb05e57a1926` | `35a4070b2201898f8ed843b0498f593e9183486f` | New provider config APIs must retain separate secret-management authorization |

## HIGH finding: BatchProject Cross-user IDOR

RED proved that User B could request User A's BatchProject and receive HTTP 200 with foreign project/book information.

The minimal fix introduced a reusable project ownership/access boundary instead of capability-only access. GREEN regression tests now reject cross-user/cross-team access and remain in the normal Go test tree.

## Observability finding: `slog.Any(struct/object)` sensitive-field redaction bypass

Finding ID: **SEC-A-008**.

Affected sensitive field-name examples include `clientSecret`, `accessToken`, and `refreshToken`.

The original sanitizer recursively handled maps/nested maps, but an object/struct passed through `slog.Any(...)` could take the default object path and bypass map-oriented sensitive-field filtering.

RED: `b37029fe6a5930622d891492f53169e18e7e911c`.

Fix: `0bfd17a3e2bc217fbde4d54c40c096f8fd1f8920` extended sanitization to recursively safe object representations and avoids logging the original object when safe conversion fails.

GREEN regression coverage includes case variants, nested values, DSN-like strings, Basic-Auth URLs, Bearer strings and query-like credentials.

## Local Executor stale credential finding

Finding ID: **SEC-A-007**.

RED showed stale/offline Local Executor identity could still attempt terminal task mutation.

Fix: `2fbf429eaace3f3cb6f3178d10d1d115e89013cc`.

Stale/offline executor identity/credential state can no longer `complete` or `fail` a task, and assignment-owner checks prevent an old/foreign executor identity from overwriting the current assignment owner.

This remains a Stage B recheck item because Task9 lease generation/fencing must also prove that an old worker token cannot Complete, Fail, Renew or Release after a newer owner takes the lease.

## Observability security verification

- Request ID validation rejects CR/LF, control characters, overlong values and multi-value headers; invalid values are regenerated.
- Access logs record bounded metadata only and do not intentionally log query strings, bodies, authorization/cookie material, passwords, provider secrets, publishing credentials or executor bearer credentials.
- `GET /api/v1/diagnostics`: anonymous 401; ordinary authenticated user 403; `batch.view` only 403; global admin/owner allowed. BatchProject ownership does not grant global diagnostics access.
- Diagnostics DTO excludes token/tokenHash/credential/ciphertext/nonce/DSN/environment/raw provider config material. Provider errors are sanitized rather than reflecting raw upstream bodies/paths.
- `/healthz`, `/readyz`, panic and 500 responses use safe messages plus request IDs rather than returning SQL/DSN/stack/environment/credential details.
- A logging-only resource read discovered by RED testing was removed so Observability does not widen the authorized resource-access surface.

## Secret Scan — Current Tree

Scope included tracked `.env`/`.env.*` paths, YAML, JSON, scripts, docs, workflows, test fixtures, Provider/Publishing configuration and Redis/MySQL/TOS/Auth-related strings/key material.

Evidence:

- the full-history Gitleaks job also covered the tracked PR tree at scan time;
- subsequent changes before final closure are workflow deletion and audit-document edits only;
- exact-head tree inspection found no tracked `.env` file;
- `.gitignore` blocks common `.env`, key, keystore, token and credential file patterns as prevention only.

Current-tree classification:

| Finding type | Path | Commit/hash | Severity | Rotation status |
|---|---|---|---|---|
| `TEST_FIXTURE / NON_PRODUCTION` local MySQL service credential | `.github/workflows/ci.yml` | blob `43309233b29d6303e7889bce8335ed81ecff00aa` | INFO | NOT REQUIRED |
| Confirmed production secret | none | n/a | n/a | NOT REQUIRED |
| `UNCONFIRMED` secret-like material requiring escalation | none identified | n/a | n/a | NOT REQUIRED |

The CI credential above is scoped to the ephemeral GitHub Actions MySQL service and is not treated as a production credential.

**No confirmed production secret requiring rotation found**

## Secret Scan — Git History

The temporary Stage A workflow used `actions/checkout` with `fetch-depth: 0` and ran Gitleaks with redacted output and a failing exit code policy.

History job evidence:

- 834 commits scanned;
- approximately 2.20 MB scanned;
- no leaks found;
- job conclusion SUCCESS.

Git-history result:

**No confirmed production secret requiring rotation found**

`ROTATION_REQUIRED`: **No**.

If a future finding is proven to be a real production credential, it must be marked `ROTATION_REQUIRED`; deleting history or adding `.gitignore` is not credential remediation and the source credential must be rotated/revoked.

## `.gitignore` boundary

`.gitignore` covers common `.env` variants, secret/credential directories, private-key formats, token/credential files and common keystore formats. This is preventive control only, not historical-secret remediation.

## SSRF residual boundary

Current controls reject loopback/private/link-local/metadata-style destinations, IPv4-in-IPv6 private forms and non-canonical numeric/hex host forms; redirects are revalidated and production dialing validates resolved addresses.

There is an explicit local-development loopback exception for provider HTTP; remote media download does not use it.

This report does **not** claim “100% SSRF safe.” DNS behavior, custom transports and future networking changes remain security-sensitive.

## ffmpeg / process execution

The merge path uses parameterized `exec.CommandContext(binary, args...)` and does not construct a shell command or invoke `sh -c`/`bash -c`. Server-generated temporary/output paths and media-path regressions remain in the normal Go test tree.

## Temporary Security Workflow

`.github/workflows/security-audit-temp.yml` was removed in commit `2abce456512b9caae91c3410123945bdd613e147`.

Security regression tests themselves remain as normal Go package tests. Normal CI and Task16 Go regression both execute `go test ./...`, so deleting the temporary workflow does not remove the implemented security regression coverage.

## Stage A exact-head gate

Only the final branch SHA after this report commit is authoritative as `STAGE_A_HEAD`.

That exact SHA must have:

- `go test ./...` ✅
- `go build ./...` ✅
- frontend test ✅
- frontend build ✅
- admin build ✅
- Goose migration/reload/rollback ✅
- Task14 migration/restart/recovery ✅
- Task16 E2E/Go regression ✅
- security regressions via normal `go test ./...` ✅
- latest-main comparison `behind=0` ✅
- PR #21 still Draft ✅

Earlier GREEN runs are supporting evidence only and are not the final exact-head gate.

## Stage B — blocked by Task9 Runtime merge

Stage B starts only after Task9 Runtime enters `main`, followed by a fresh main sync. Stage B must review Redis, queue payload, MySQL-CAS authorization, Worker ownership/token chain, lease/fencing, Scheduler, recovery, idempotency, retry authorization, Redis wipe/replay, internal endpoints, runtime diagnostics and the Task14 Runtime Adapter.

PR #21 must not be marked Ready or merged before Stage B is complete.

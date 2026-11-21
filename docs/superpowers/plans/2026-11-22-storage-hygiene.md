# Storage Hygiene Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent local builds, production releases, and repeated Batch Factory snapshots from filling disks again.

**Architecture:** Release cleanup is deterministic and runs after every successful Go deployment. Business snapshots remain immutable; a separate audit-first compaction design removes repeated JSON only after referenced revisions are proven safe.

**Tech Stack:** PowerShell, POSIX shell, Node test runner, Go, MySQL 8, Go embed deployment.

## Global Constraints

- Keep current binary plus one verified rollback binary.
- Never delete MySQL rows, user media, login state, or TOS objects as cache.
- `new-api` is a separate live service and is outside automatic V88 cleanup.
- All cleanup commands support a report/dry-run mode before deletion.
- Do not deploy through GitHub Actions.

---

### Task 1: Support direct-file release pruning

**Files:**
- Modify: `scripts/prune-qiantie-releases.sh`
- Modify: `scripts/deploy-go-embedded.ps1`
- Test: `tests/prune-qiantie-releases.test.js`

**Interfaces:**
- Consumes: `QIANTIE_RELEASE_ROOT`, `KeepPreviousReleases`.
- Produces: cleanup that supports both `qiantie` symlinks and direct `qiantie` files.

- [x] Write a failing test with a direct `qiantie` file and two `qiantie.rollback-*` files.
- [x] Run `node --test tests/prune-qiantie-releases.test.js` and verify failure reports “未找到版本软链”.
- [x] Update the pruner to preserve the direct current file and newest rollback.
- [x] Remove the deployment branch that skipped pruning for direct files.
- [x] Run the focused test, `sh -n`, PowerShell parsing, and `git diff --check`.
- [x] Commit as `fix(release): prune direct deployment rollbacks`.

### Task 2: Add a local build-artifact report and cleanup command

**Files:**
- Create: `scripts/maintenance/clean-local-build-artifacts.ps1`
- Test: `tests/clean-local-build-artifacts.test.ps1`

**Interfaces:**
- Consumes: repository root, `-Apply`, `-KeepPreviousBinary`.
- Produces: JSON report with candidate paths, bytes, action, and reason.

- [x] Write a Pester-free PowerShell test that creates a temporary repository fixture containing `target/`, old `qiantie-*.exe`, current binary, previous binary, and a source file.
- [x] Run the test and verify it fails because the command does not exist.
- [x] Implement report-only default; only `-Apply` removes whitelist-matched build outputs.
- [x] Verify current/previous executables and source files remain unchanged.
- [x] Run the fixture test and `git diff --check`.
- [x] Commit the script and test.

### Task 3: Add production disk reporting

**Files:**
- Create: `scripts/maintenance/report-production-disk.sh`
- Test: `tests/report-production-disk.test.js`

**Interfaces:**
- Consumes: configurable root paths and MySQL DSN from the existing environment file.
- Produces: JSON sections `filesystem`, `application`, `database`, `media`, `logs`, `otherServices`.

- [ ] Write a failing fixture test using temporary directories with literal byte sizes.
- [ ] Implement directory and file accounting without printing credentials.
- [ ] Add thresholds 75%, 82%, and 88% to the report result.
- [ ] Verify JSON output against hand-calculated fixture values.
- [ ] Commit.

### Task 4: Audit Batch Factory snapshot duplication

**Files:**
- Create: `backend/internal/batchfactoryv11/config_snapshot_audit.go`
- Create: `backend/internal/batchfactoryv11/config_snapshot_audit_test.go`
- Create: `scripts/maintenance/audit-config-snapshots.ps1`

**Interfaces:**
- Consumes: snapshot IDs referenced by `batch_factory_v12_director_revisions` and snapshot payload hashes.
- Produces: read-only counts for referenced, unreferenced, duplicate-payload, and reclaimable bytes.

- [ ] Write SQL-mock tests for referenced and duplicate snapshots.
- [ ] Verify tests fail before the audit query exists.
- [ ] Implement read-only audit queries; do not add `DELETE`.
- [ ] Run focused Go tests and execute report-only audit on production.
- [ ] Save only aggregate counts in the acceptance report; never store snapshot payloads.
- [ ] Commit.

### Task 5: Design compaction from audit evidence

**Files:**
- Create: `docs/superpowers/specs/2026-11-22-config-snapshot-compaction-design.md`

- [ ] Use Task 4 aggregates to decide between payload deduplication and safe unreferenced-row retention.
- [ ] Specify foreign-key handling, transaction size, rollback export, and dry-run output.
- [ ] Obtain explicit approval before the first production `DELETE` or table rewrite.


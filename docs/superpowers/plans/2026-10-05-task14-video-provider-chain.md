# Task 14 Video Provider Chain Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first Task 14 vertical slice from completed `FINAL_PROMPT` output through `personal_api` / `yd2.0-mini`, durable MySQL ProductionJob/ProductionTask state, provider polling, TOS artifact persistence, and VIDEO success.

**Architecture:** `api/internal/video` is an independent bounded context. MySQL owns provider config and video job/task facts; provider adapters implement protocol-specific Submit/Poll/Cancel behind one interface; artifact persistence is injected behind a video-domain `ArtifactStore` contract. No generic Redis runtime/scheduler is introduced: restart recovery is based on recoverable MySQL rows until Task 9.4 shared primitives are available.

**Tech Stack:** Go, MySQL, Goose, `net/http`, AES-256-GCM, TOS through an injected artifact-store boundary, existing React/Ant Design apps unchanged in phase 1.

**Spec:** `docs/superpowers/specs/2026-10-05-task14-video-provider-chain-design.md`

## Global Constraints

- Do not duplicate Task 9.4 generic Redis Queue/Lock/Lease/Scheduler/runtime infrastructure.
- Do not implement or modify Task 13 matchAudio, audioDurationSec algorithms, Director timeline rules, or audio measurement.
- Provider secrets use minimal AES-256-GCM with a 32-byte server-environment master key; plaintext secret is never persisted, returned, logged, or committed.
- Provider Config, ProductionJob and ProductionTask durable truth is MySQL.
- Backend/provider/poller/artifact logic is Go; no Node.js production backend.
- Phase 1 provider/model is exactly `personal_api` + `yd2.0-mini`.
- Provider success is not VIDEO success until the artifact is persisted through the TOS/artifact boundary.

## Review Focus

- Duplicate logical submissions must reuse the active durable job/task and not call the provider twice.
- Provider response bodies/auth failures must not leak API keys or unsanitized response payloads into persisted errors/views.
- A provider `succeeded` result with missing/invalid artifact URL must fail explicitly instead of marking VIDEO succeeded.
- Restart recovery must be derived from MySQL queued/running task rows, not in-memory goroutines.
- Final Prompt identity/version/revision stored on ProductionJob must exactly identify the Task 12 output consumed.

---

### Task 1: Video domain contracts, provider mapping and encrypted configuration

**Files:**
- Create: `api/internal/video/types.go`
- Create: `api/internal/video/provider.go`
- Create: `api/internal/video/crypto.go`
- Test: `api/internal/video/provider_test.go`
- Test: `api/internal/video/crypto_test.go`

**Interfaces:**
- Produces: `Provider`, `ProviderFactory`, `ProviderConfig`, `ProviderConfigView`, `SubmitRequest`, `SubmitResult`, `PollResult`, `TaskStatus`, `ErrorCode`, `ProviderForModel(model string) (string, bool)`, `EncryptSecret(masterKey []byte, plaintext string)`, `DecryptSecret(masterKey []byte, ciphertext, nonce []byte)`.

- [ ] Write failing tests for provider-unconfigured behavior, exact provider/model mapping, secret encryption round-trip, decryption failure, and JSON/view secret non-disclosure.
- [ ] Verify RED with `cd api && go test ./internal/video -run 'TestProvider|TestSecret' -count=1`.
- [ ] Implement the minimum domain types/mapping/crypto required by the tests.
- [ ] Verify GREEN with the same targeted command, then `cd api && go test ./...`.
- [ ] Commit.

### Task 2: Personal API adapter Submit/Poll protocol

**Files:**
- Create: `api/internal/video/personal_api.go`
- Test: `api/internal/video/personal_api_test.go`

**Interfaces:**
- Consumes: `Provider`, `SubmitRequest`, `SubmitResult`, `PollResult`, normalized status/error types from Task 1.
- Produces: `NewPersonalAPIProvider(config ProviderConfig, secret string, client *http.Client) (Provider, error)`.

- [ ] Write failing tests proving Bearer auth stays server-side; submit success maps task id; submit non-2xx/auth failure returns normalized sanitized errors; poll maps queued/running/succeeded/failed; succeeded extracts artifact URL; invalid success payload fails.
- [ ] Verify RED with `cd api && go test ./internal/video -run 'TestPersonalAPI' -count=1`.
- [ ] Implement only `personal_api` / `yd2.0-mini` protocol behavior equivalent to v88.
- [ ] Verify targeted GREEN and full Go suite.
- [ ] Commit.

### Task 3: MySQL schema and durable store

**Files:**
- Create: `api/db/migrations/00004_task14_video.sql`
- Create: `api/internal/video/store.go`
- Create: `api/internal/video/mysql_store.go`
- Test: `api/internal/video/mysql_store_test.go`

**Interfaces:**
- Produces: `Store` with provider-config lookup/upsert, `CreateOrGetProductionJob`, `CreateProductionTask`, `UpdateProductionTask`, `UpdateProductionJob`, `LatestTaskForJob`, and `ListRecoverableTasks(ctx, limit)`.

- [ ] Write failing store tests for durable provider config, idempotent job creation, Final Prompt revision/version fields, immutable attempt identity, and recoverable queued/running task discovery across a new store instance.
- [ ] Verify RED with `cd api && go test ./internal/video -run 'TestMySQLStore' -count=1` (CI MySQL job is authoritative if local DB is unavailable).
- [ ] Add Goose migration for `video_provider_configs`, `video_production_jobs`, `video_production_tasks` with uniqueness on provider key/model and logical idempotency key, plus recoverable-task indexes.
- [ ] Implement MySQL store methods with transactional duplicate-submit handling.
- [ ] Verify targeted tests plus Goose up/status/down in CI.
- [ ] Commit.

### Task 4: Production service, idempotent Submit and restart recovery

**Files:**
- Create: `api/internal/video/service.go`
- Test: `api/internal/video/service_test.go`

**Interfaces:**
- Consumes: `Store`, `ProviderFactory`, `ArtifactStore`, `FinalPromptSource`.
- Produces: `Service.Start(ctx, StartRequest)`, `Service.PollTask(ctx, taskID)`, `Service.RecoverPending(ctx, limit)`.
- `FinalPromptSource.ResolveFinalPrompt(ctx, batchProjectID, bookID)` returns completed output plus `StageRunID`, `PromptVersion`, `InputRevision`.
- `ArtifactStore.Persist(ctx, sourceURL, objectHint) -> Artifact` is a narrow Task-14 boundary, not a generic scheduler/runtime.

- [ ] Write failing tests for provider-unconfigured, submit success/failure, duplicate Start invoking provider once, exact Final Prompt identity/version/revision tracking, poll queued/running/failed, poll succeeded requiring artifact persistence, and RecoverPending discovering durable pending/running rows from a fresh service instance.
- [ ] Verify RED with `cd api && go test ./internal/video -run 'TestService' -count=1`.
- [ ] Implement minimal orchestration. Do not create Redis queue/lock/scheduler primitives.
- [ ] Verify targeted GREEN and full Go suite.
- [ ] Commit.

### Task 5: Task 12 Final Prompt adapter, TOS boundary and application wiring

**Files:**
- Create: `api/internal/video/final_prompt_source.go`
- Create or adapt existing object-storage integration: `api/internal/video/artifact_store.go`
- Modify: `api/internal/app/app.go`
- Modify: `api/internal/httpapi/server.go`
- Create: `api/internal/httpapi/video.go`
- Test: `api/internal/video/final_prompt_source_test.go`
- Test: `api/internal/httpapi/video_test.go`

**Interfaces:**
- Produces HTTP endpoints for provider status/config-safe view and phase-1 video start/status/poll without returning secrets.
- Consumes existing Task 12 generation store/service for completed Final Prompt only; does not invoke generation stages.

- [ ] Write failing tests that reject missing/incomplete FINAL_PROMPT, expose no secret fields, preserve distinct provider status errors, and wire Start/Status through the Go service.
- [ ] Verify RED.
- [ ] Implement adapters/wiring using existing MySQL and TOS configuration. If main lacks a reusable TOS client, keep the video `ArtifactStore` interface and provide the smallest Go TOS adapter without creating a generic runtime.
- [ ] Verify targeted tests, `cd api && go test ./...`, and `cd api && go build ./...`.
- [ ] Commit.

### Task 6: CI/migration verification and Task 14 phase-1 record

**Files:**
- Modify: `TASKS.md` only to record phase-1 repository evidence while leaving overall Task 14 `⚠️` until all later providers/merge/UI/public validation are complete.

**Interfaces:** none.

- [ ] Run/observe repository CI for the branch: Go test/build, Goose up/status/down, admin build, user tests/build.
- [ ] If any check fails, use systematic-debugging; add a reproducing RED test before production-code fixes.
- [ ] Confirm no changes touched Task 13/matchAudio files and no generic Redis runtime/scheduler was added.
- [ ] Record actual commit/CI evidence in `TASKS.md` without claiming ECS/public completion.
- [ ] Commit and re-check CI.

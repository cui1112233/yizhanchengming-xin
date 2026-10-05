# Task 14 Video Provider / Video Generation Chain Design

Date: 2026-10-05
Branch: `feat/task14-video-provider-chain`
Base main: `2223989caf4601d09f9b8344353ab17d58a155e0`

## Scope

Implement Task 14 only, under `api/internal/video` as an independent bounded context. Consume the completed Task 12 Final Prompt output; do not re-run or re-implement Script, Hook, Director, H3 Director, Final Prompt, or Task 13 matchAudio.

Primary chain:

`FINAL_PROMPT -> ProductionJob -> ProductionTask -> Provider Submit -> durable Poll -> artifact -> TOS -> optional MergeJob -> VIDEO succeeded`

Phase 1 vertical slice:

`FINAL_PROMPT -> personal_api -> yd2.0-mini -> Submit -> ProductionTask -> Poll -> artifact -> TOS -> VIDEO succeeded`

## Hard boundaries

1. Do not duplicate Task 9.4 generic runtime infrastructure. Video owns video-domain job/task state and provider polling semantics only. If the current main does not yet expose the final Task 9.4 Redis Queue/Lock/Lease/Scheduler contracts, define only narrow video-facing abstractions that can later be adapted to those shared primitives. Do not create a second global Redis runtime, generic queue, generic lock, or generic scheduler package.
2. Task 13 remains isolated. Do not implement or modify matchAudio, audioDurationSec behavior, Director timing rules, or audio measurement. VIDEO consumes already completed Task 12 output and accepts later Task 13 timing fields without interpreting them.
3. Secret encryption is intentionally minimal: AES-256-GCM using a 32-byte server environment master key. Provider secrets must never be stored plaintext, returned to clients, or written to logs. The master key is never persisted in MySQL and never committed. Decryption failures return an explicit provider configuration error. KMS and rotation are out of scope.
4. Backend/provider/poller/executor/merge logic is Go. Frontend remains React + Ant Design. Persistent facts live in MySQL with Goose migrations. Redis may only be used for short-lived coordination and, when Task 9.4 lands, through the shared runtime. TOS is the durable object store. ffmpeg remains the preferred merge engine.

## Old v88 compatibility findings

The legacy v88 implementation establishes the compatibility contract:

- `backend/internal/batchfactoryv11/video_provider.go`: provider/model normalization and the legacy in-memory provider registry. `yd2.0-mini` belongs to `personal_api`; `minimax-h3-video` belongs to `autodl_comfyui`; `seedance-2-0-official` belongs to `yfai_seedance`; `doubao_local_executor` is a distinct local-executor provider.
- `backend/internal/batchfactoryv11/yadi_video_adapter.go`: `personal_api` submit/poll behavior, Bearer auth, reference-image mapping, task/result parsing and error semantics.
- `backend/internal/batchfactoryv11/autodl_h3_video_adapter.go`: AutoDL ComfyUI H3 request schema, workflow selection, reference-image handling and result extraction.
- `backend/internal/batchfactoryv11/yfai_seedance_adapter.go`: Seedance submit/poll behavior and `seedance-2-0-official` compatibility.
- `backend/internal/batchfactoryv11/local_executor_video_adapter.go` plus `internal/localexecutor/*`: local executor availability, task identity, poll/cancel and artifact handoff semantics.
- `backend/internal/batchfactoryv11/production.go`, `production_store_mysql.go`: legacy ProductionJob/ProductionTask durability and retry/idempotency behavior.
- `backend/internal/batchfactoryv11/merge.go`, `merge_store_mysql.go`, `internal/mergeworker/ffmpeg.go`, `tos_sdk.go`, `redis.go`: merge, ffmpeg, TOS and short-lived coordination behavior.

These are behavior/protocol references only; no legacy Node production backend is migrated.

## Domain model

### ProviderConfig

MySQL is the source of truth. Minimum fields:

- id
- provider_key
- model
- endpoint/create_url/tasks_url/result_url/base_url as non-sensitive provider options where applicable
- encrypted_secret
- secret_nonce
- secret_key_id/fingerprint metadata if needed for diagnostics
- enabled/configured state
- created_at / updated_at

Client-facing views expose only non-sensitive fields and `configured: true|false`.

### ProductionJob

Represents one book-level VIDEO business job. Minimum fields:

- id
- batch_project_id
- book_id
- status: queued/running/succeeded/failed/cancelled
- input_revision
- final_prompt_stage_run_id
- final_prompt_version
- provider
- model
- idempotency_key
- created_at / started_at / completed_at / cancelled_at / updated_at

### ProductionTask

Represents one concrete provider attempt. Minimum fields:

- id
- production_job_id
- attempt
- provider
- model
- request_id
- provider_job_id
- status
- progress when supported
- error_code / error_message
- sanitized provider metadata
- artifact source URL only while being ingested
- output asset identity / TOS bucket / object key / durable URL or access reference
- created_at / started_at / completed_at / cancelled_at / updated_at

Old attempts are immutable audit history. Explicit Retry creates a new attempt instead of overwriting the failed task.

## Provider abstraction

A single Go provider interface drives business flow:

- `Submit(ctx, request) -> result`
- `Poll(ctx, providerJobID) -> result`
- `Cancel(ctx, providerJobID) -> result`

Adapters translate provider-specific request and response formats. Business services do not branch repeatedly on provider keys.

Normalized provider task states:

- queued
- running
- succeeded
- failed
- cancelled

Normalized error codes include at least:

- provider_unconfigured
- provider_unavailable
- provider_auth_failed
- provider_request_failed
- provider_invalid_response
- provider_cancel_unsupported
- provider_config_decrypt_failed

Sensitive provider payload fields are scrubbed before storage or logging.

## Phase 1 personal_api contract

Compatibility with v88:

- provider key: `personal_api`
- default model: `yd2.0-mini`
- create endpoint semantics equivalent to legacy Yadi adapter
- Bearer token auth, server-side only
- submit maps provider task/request id to `provider_job_id`
- poll maps provider states to normalized states
- reference image URLs come from backend-owned Final Prompt/asset data, not direct browser/provider calls
- provider success artifact is ingested into TOS before the business job is marked VIDEO succeeded

Provider status API distinguishes `unconfigured`, `unavailable`, `auth_failed`, and `available`; it does not collapse all failures into a generic 503 message.

## Final Prompt integration

Task 12 `stage_runs` remains authoritative for Final Prompt. VIDEO resolves a completed `FINAL_PROMPT` StageRun and records its stable identity/version into ProductionJob. No VIDEO retry invokes Task 12 stages.

Idempotency must include at least book, VIDEO stage, input revision, Final Prompt identity/version, provider, model, and retry attempt semantics. Repeated submission of the same effective request returns/reuses the active job/task and does not incur a second provider charge. Explicit Retry increments attempt and is intentionally a new provider submission.

## Durable polling and Task 9.4 integration boundary

Polling must not depend on an open browser or in-memory goroutine state. `provider_job_id`, normalized status, attempts and recovery timestamps remain in MySQL.

Until Task 9.4 shared runtime interfaces are present on main, `api/internal/video` may define only a narrow wake/schedule abstraction needed by its service, while MySQL remains sufficient to discover recoverable queued/running tasks after restart. When Task 9.4 lands, the adapter will bind this abstraction to the shared Scheduler/Queue/Lease implementation rather than introducing a competing runtime.

Redis outages must never erase durable error/job/task state.

## Artifact and TOS boundary

A succeeded provider result with a temporary URL is not considered final durable output. The backend downloads/streams the result under bounded time/size controls, validates the media source, uploads to TOS, and persists bucket/object key/access reference on ProductionTask. Only then may the task/job transition to succeeded.

TOS credentials remain server-side and are never returned to the browser.

## Later Task 14 phases

After the first vertical slice is green:

1. `yfai_seedance` + `seedance-2-0-official`
2. `autodl_comfyui` + `minimax-h3-video`
3. `doubao_local_executor` register/heartbeat/lease/renew/complete/fail/expiry recovery, using shared Task 9.4 coordination primitives when available
4. Provider Cancel and VIDEO Retry
5. MergeJob + ffmpeg + TOS output
6. React VIDEO status/progress/attempt/error/output/retry/cancel UI and provider settings with masked secrets only

## TDD gates

First RED batch:

- provider unconfigured
- provider/model mapping
- secret not leaked in API/view
- submit success
- submit failure
- poll queued
- poll running
- poll succeeded
- poll failed
- duplicate submit idempotency
- Final Prompt revision/version tracking
- restart recovery discovers pending/running poll tasks from MySQL

Phase 1 GREEN is complete only when the personal_api vertical slice passes these tests and the repository builds.

Completion verification for full Task 14 remains:

- `go test ./...`
- `go build ./...`
- frontend tests/build
- admin build
- Goose up/status/rollback when migrations change
- Redis-unavailable explicit failure/recovery tests when Redis integration is introduced

`TASKS.md` overall Task 14 remains `⚠️` until code/CI/PR are complete, and ECS/public validation remains Task 16.
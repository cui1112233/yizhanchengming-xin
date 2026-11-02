# Agent Canvas Unified Background Queue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Agent Canvas image, TTS/audio, and video tasks persist, continue in the Go backend after the page closes, and report one truthful project progress summary.

**Architecture:** `agent_generation_tasks` remains the single common state record. A new coordinator claims queued or stale-running image/audio tasks, reuses the existing video and script coordinators, and writes results before exposing `awaiting_confirmation`. The React canvas derives both per-node status and a project summary from persisted task records.

**Tech Stack:** Go, MySQL, React, existing Go-embedded frontend build, existing local executor and TTS services.

## Global Constraints

- Reuse `agent_generation_tasks`; do not create a second task table.
- Preserve `agent_video_jobs` as the provider-specific video record.
- Never infer `completed` from a missing task or node state.
- Provider credentials remain server-side.
- Keep workflow and storyboard as two views of the same project task data.
- Do not reset, clean, overwrite unrelated worktree changes, commit, push, or deploy without explicit user authorization.
- Verification must include the Go-embedded runtime and authenticated browser, not only source tests.

---

### Task 1: Persisted Work Discovery And Project Summary

**Files:**
- Modify: `backend/internal/agentworkspace/generation_tasks.go`
- Modify: `backend/internal/agentworkspace/mysql_generation_tasks.go`
- Modify: `backend/internal/agentworkspace/generation_tasks_test.go`
- Modify: `backend/internal/agentworkspace/mysql_generation_tasks_test.go`
- Create: `backend/internal/httpapi/agent_generation_summary.go`
- Create: `backend/internal/httpapi/agent_generation_summary_test.go`
- Modify: `backend/internal/httpapi/agent_generation_tasks.go`

**Consumes:** `GenerationTaskStore`, `GenerationState`, `ListGenerationTasks`.

**Produces:**
```go
type AgentGenerationSummary struct {
    Total, Queued, Running, AwaitingConfirmation int
    Completed, Failed, Cancelled int
}

func SummarizeAgentGenerations(tasks []agentworkspace.GenerationTask) AgentGenerationSummary
func (s *MemoryStore) ListRunnableGenerationTasks(ctx context.Context, modes []string, limit int) ([]GenerationTask, error)
```

- [ ] **Step 1: Write failing store tests**

Add memory and sqlmock cases with image/audio queued, image/audio running, video and script work, plus terminal tasks. Assert mode filtering, oldest-first order, limit, owner preservation, and exclusion of terminal states.

- [ ] **Step 2: Run focused store tests and confirm failure**

Run: `go test ./internal/agentworkspace -run "Test(Memory|MySQL).*RunnableGeneration" -count=1`

Expected: FAIL because mode-set discovery is not implemented.

- [ ] **Step 3: Add task discovery and summary code**

Keep the existing one-mode helper compatible, then add a mode-set form that queries:

```sql
SELECT ... FROM agent_generation_tasks
WHERE mode IN (...) AND state IN ('queued', 'running')
ORDER BY updated_at ASC, id ASC
LIMIT ?
```

Map all seven persisted states explicitly in `SummarizeAgentGenerations`; unknown values contribute only to `Total` and never `Completed`.

- [ ] **Step 4: Add summary route test and route**

Expose `GET /api/agent/tasks/{id}/generation-summary`. Return:

```json
{"summary":{"total":6,"queued":1,"running":2,"awaitingConfirmation":1,"completed":1,"failed":1,"cancelled":0}}
```

Use the authenticated owner and the existing project-scoped task listing.

- [ ] **Step 5: Run verification**

Run: `go test ./internal/agentworkspace ./internal/httpapi -run "Test.*(RunnableGeneration|GenerationSummary)" -count=1`

Expected: PASS.

### Task 2: Background Image Coordinator

**Files:**
- Create: `backend/internal/httpapi/agent_image_coordinator.go`
- Create: `backend/internal/httpapi/agent_image_coordinator_test.go`
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`

**Consumes:** `AgentImageGenerator`, `agent_generation_tasks`, `AgentAssets`, `AgentAssetObjects`, task discovery from Task 1.

**Produces:**
```go
type agentImageCoordinator struct { /* projects, assets, objects, generator, now */ }
func (s *agentImageCoordinator) Run(ctx context.Context)
func (s *agentImageCoordinator) pollPendingOnce(ctx context.Context) error
```

- [ ] **Step 1: Write a failing background-image test**

Create an image task through `POST /image-generations` with a generator blocked on a channel. Assert the HTTP response is `202` with `state=queued` before releasing the generator. Call `pollPendingOnce`, then assert assets exist and task becomes `awaiting_confirmation` without another browser request.

- [ ] **Step 2: Run the new test and confirm failure**

Run: `go test ./internal/httpapi -run TestAgentImageCoordinator -count=1`

Expected: FAIL because the route still calls the generator synchronously.

- [ ] **Step 3: Extract image execution from the HTTP handler**

The submit handler validates node/model/request, creates only the generation task, and returns `202`:

```json
{"task":{"state":"queued"}}
```

The coordinator atomically transitions `queued -> running`, calls the existing generator, saves each result using the existing asset object store, then transitions `running -> awaiting_confirmation` with result asset IDs. On error it transitions to `failed` with a user-safe message.

- [ ] **Step 4: Implement restart and cancellation rules**

Use the same in-process claim map pattern as video/script workers. A `running` image task may be resumed only when `updated_at` is older than the configured image lease. The worker re-reads task state before saving outputs; cancelled tasks discard outputs and never become awaiting confirmation.

- [ ] **Step 5: Wire runtime startup and verify**

Start the coordinator only from `app.Build` through `RouterOptions`, following the existing video worker pattern. Run:

`go test ./internal/httpapi -run "TestAgent(ImageCoordinator|NodeGeneration)" -count=1`

Expected: PASS.

### Task 3: Background TTS And Audio Coordinator

**Files:**
- Create: `backend/internal/httpapi/agent_audio_coordinator.go`
- Create: `backend/internal/httpapi/agent_audio_coordinator_test.go`
- Modify: `backend/internal/httpapi/agent_audio_jobs.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`

**Consumes:** `AgentAudioServices`, built-in TTS, configured audio generator, audio processor, task discovery from Task 1.

**Produces:**
```go
type agentAudioCoordinator struct { /* services, inFlight, now */ }
func (s *agentAudioCoordinator) Run(ctx context.Context)
func (s *agentAudioCoordinator) pollPendingOnce(ctx context.Context) error
```

- [ ] **Step 1: Write failing TTS and processor tests**

Cover one built-in TTS task and one trim task. Assert submit returns `202`/queued, no generator or processor runs until the coordinator is invoked, output assets persist, and results become `awaiting_confirmation`.

- [ ] **Step 2: Run tests and confirm failure**

Run: `go test ./internal/httpapi -run TestAgentAudioCoordinator -count=1`

Expected: FAIL because audio currently runs inside the request handler.

- [ ] **Step 3: Move the existing execution helpers behind the worker**

Retain `runAgentAudio` and `persistAgentAudioOutputs`; call them only after the worker claims the task. Store a complete request snapshot in task settings:

```go
map[string]any{
  "operation": string(input.Operation), "prompt": input.Prompt,
  "parameters": input.Settings, "advanced": input.Advanced,
  "sourceAssetId": input.SourceAssetID,
}
```

The HTTP route only validates, creates the task, and returns it.

- [ ] **Step 4: Add retry/cancel integration tests**

Use existing generic generation retry/cancel endpoints. Assert retry has `retryOfTaskId`, creates a fresh task, and does not overwrite prior output. Assert cancelling queued work prevents TTS/processor invocation; cancelling after output persistence does not mark the task complete.

- [ ] **Step 5: Run focused verification**

Run: `go test ./internal/httpapi -run "TestAgent(AudioCoordinator|Audio.*(Retry|Cancel))" -count=1`

Expected: PASS.

### Task 4: Unify Video Projection And Canvas Summary

**Files:**
- Modify: `backend/internal/httpapi/agent_video_jobs.go`
- Modify: `backend/internal/httpapi/agent_video_jobs_test.go`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/pages/agent-generation-state.js`
- Modify: `frontend/src/user/pages/agent-generation-state.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`

**Consumes:** video coordinator, generation summary route, generic generation task routes.

**Produces:** one node presentation model for image/audio/video with an explicit state, error, retry action, cancel action, and selected result.

- [ ] **Step 1: Write frontend state tests**

Create task fixtures covering all common states and a video job in `submitting`/`abandoned`. Assert no missing record renders as completed; active states have clear labels; failed/cancelled entries expose retry; queued/running entries expose cancel.

- [ ] **Step 2: Run focused frontend tests and confirm failure**

Run: `npm --prefix frontend test -- --run src/user/pages/agent-generation-state.test.js`

Expected: FAIL for video projection or project summary display.

- [ ] **Step 3: Normalize API and UI projection**

Fetch the summary with regular generation polling. Merge video job detail only where it adds provider state, while the common generation task remains the source for overall status. Add a compact project status line that shows counts without changing the canvas layout hierarchy.

- [ ] **Step 4: Wire retry and cancel commands**

Reuse generic retry/cancel for image and audio. Keep the dedicated video routes for provider cancellation/retry, then refresh the common task and summary after either action.

- [ ] **Step 5: Run frontend verification**

Run: `npm --prefix frontend test -- --run src/user/pages/agent-generation-state.test.js src/user/components/AgentCanvas.source.test.js`

Expected: PASS.

### Task 5: Full Local Embedded Runtime Acceptance

**Files:**
- Modify only if required by observed failures: `scripts/build-go-local.ps1`, `scripts/start-local-go-review.ps1`, `scripts/check-local-go.ps1`
- Test: existing backend and frontend suites

**Consumes:** Tasks 1-4.

**Produces:** a rebuilt local Go embedded runtime on port `4001` with recorded browser evidence.

- [ ] **Step 1: Run complete focused test sets**

Run:

```powershell
go test ./internal/agentworkspace ./internal/httpapi ./internal/app ./cmd/qiantie -count=1
npm --prefix frontend test -- --run
```

- [ ] **Step 2: Build the embedded frontend and Go binary**

Run:

```powershell
npm --prefix frontend run build
scripts/build-go-local.ps1
```

Expected: both commands succeed and the binary embeds the new frontend assets.

- [ ] **Step 3: Restart and health-check the local review runtime**

Run:

```powershell
scripts/start-local-go-review.ps1 -UseYcmMysql
scripts/check-local-go.ps1
```

Expected: `localhost:4001` reports healthy and runtime identity matches the rebuilt binary.

- [ ] **Step 4: Browser acceptance**

In an authenticated `http://localhost:4001/agent/canvas` project, create image, built-in TTS, and video work. Refresh or navigate away while each is pending. Verify the persisted statuses continue to change, retry/cancel work, project counts match records, workflow and storyboard agree, and no temporary nodes or edges remain.

- [ ] **Step 5: Record scope and handoff**

Report exact tests, build identity, runtime URL, browser checks, and any provider that could not be exercised because its credentials or local executor were unavailable. Do not claim a provider integration passed without a real completed task.

## Plan Self-Review

- Spec coverage: Tasks 1-4 cover shared state, restart recovery, image, TTS/audio, video projection, retry, cancellation, and summary. Task 5 covers embedded build, restart, health, and browser acceptance.
- Placeholder scan: no TODO/TBD or deferred implementation language remains.
- Type consistency: all tasks use `GenerationTask`, `GenerationState`, `AgentImageGenerator`, and `AgentAudioServices` already present in the codebase; the new coordinator names are consistent between producer and consumer tasks.

# AI-CUT AI Edit Plan Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first complete online AI editing loop: generate a validated draft from a saved AI-CUT project, let the user inspect it, apply it exactly once to the timeline, and undo the whole application as one history step.

**Architecture:** Add an `EditPlan` domain beside the existing timeline and Agent domains. The Go backend owns prompt construction, model output parsing, whitelist validation, idempotency, revision checks, and transactional timeline application; React only collects options and displays the saved draft. The first plan runs through the existing text-provider resolver, while later analysis, Redis workers, BGM, and export plans can reuse the same stored plan and operation contract.

**Tech Stack:** Go 1.x, `net/http`, MySQL, existing in-memory test stores, React, Ant Design, Node test runner, existing AI-CUT timeline history, existing model catalog/provider resolver.

## Global Constraints

- Work directly on the current `v88` development branch; do not create another branch.
- Preserve every unrelated dirty file in `F:\脚本测试\一战晟铭-v88`.
- Do not push, deploy publicly, use paid models, or consume a real provider quota during implementation or tests.
- System prompts live only under `backend/internal/prompt`; the frontend must not contain full system prompts.
- Go owns validation and application; the browser never executes model-provided shell commands, paths, SQL, URLs, or FFmpeg arguments.
- MySQL stores durable plans; Redis integration is a later plan and must not be faked in this core slice.
- A plan is a draft until the user explicitly applies it.
- Applying a plan creates exactly one timeline history revision; canceling or failing changes nothing.
- Frontend production assets must be embedded into the Go binary before browser acceptance.
- Concat AGPL source, assets, models, keys, and branding must not be copied.

## File Map

- `backend/internal/aicut/edit_plan.go`: plan types, states, operation whitelist, normalization, validation, and JSON parsing.
- `backend/internal/aicut/edit_plan_apply.go`: pure transformation from one validated plan to one new timeline.
- `backend/internal/aicut/edit_plan_store.go`: durable plan-store interface and in-memory implementation used by tests.
- `backend/internal/aicut/edit_plan_mysql.go`: MySQL plan persistence and idempotent lookup.
- `backend/internal/storage/aicut_schema.go`: plan and operation tables.
- `backend/internal/prompt/aicut_edit_plan.go`: system/user prompt contract built from the saved project snapshot.
- `backend/internal/httpapi/aicut_edit_plan_routes.go`: create, read, cancel, and apply endpoints.
- `backend/internal/httpapi/router.go`: route registration.
- `backend/internal/app/app.go`: inject the existing MySQL store and text-provider resolver.
- `frontend/src/shared/api/aicutEditPlan.js`: typed request helpers and client-side input guards.
- `frontend/src/user/components/AICutAutoEditDrawer.jsx`: three-step request, draft review, and confirmation UI.
- `frontend/src/user/pages/AICutPage.jsx`: toolbar entry, current project data, timeline refresh, and undo compatibility.
- `frontend/src/user/pages/aicut-page.css`: theme-aware drawer and operation-list styles.

---

### Task 1: Edit-plan contract and whitelist validation

**Files:**
- Create: `backend/internal/aicut/edit_plan.go`
- Create: `backend/internal/aicut/edit_plan_test.go`

**Interfaces:**
- Produces: `EditPlan`, `EditPlanRequest`, `EditPlanSummary`, `EditOperation`, `EditPlanState`, `NormalizeEditPlan(EditPlan) (EditPlan, error)`, and `ParseEditPlanModelOutput(string, EditPlanEnvelope) (EditPlan, error)`.
- Consumes: existing `Timeline`, `Track`, `Clip`, and stable editor resource IDs.

- [ ] **Step 1: Write failing validation tests**

Cover these exact cases in table tests:

```go
func TestNormalizeEditPlanRejectsUnknownOperation(t *testing.T) {
    plan := validEditPlan()
    plan.Operations[0].Type = "run_shell"
    if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
        t.Fatalf("err=%v, want ErrInvalidEditPlan", err)
    }
}

func TestNormalizeEditPlanRejectsNegativeAndOverflowTime(t *testing.T) {
    for _, operation := range []EditOperation{
        {ID: "op-1", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: -1, EndMS: 1000, Enabled: true},
        {ID: "op-2", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 1000, EndMS: 1000, Enabled: true},
    } {
        plan := validEditPlan()
        plan.Operations = []EditOperation{operation}
        if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
            t.Fatalf("operation=%#v err=%v", operation, err)
        }
    }
}
```

Also test duplicate operation IDs, unknown resource IDs, empty reason, invalid state, more than 200 operations, and model text wrapped in a Markdown code fence.

- [ ] **Step 2: Run the tests and verify failure**

Run: `go test ./internal/aicut -run 'TestNormalizeEditPlan|TestParseEditPlan'`

Expected: FAIL because the edit-plan types and functions do not exist.

- [ ] **Step 3: Implement the domain contract**

Use these states and operation names exactly:

```go
const (
    EditPlanQueued   EditPlanState = "queued"
    EditPlanPlanning EditPlanState = "planning"
    EditPlanReady    EditPlanState = "ready"
    EditPlanApplied  EditPlanState = "applied"
    EditPlanFailed   EditPlanState = "failed"
    EditPlanCanceled EditPlanState = "canceled"
    EditPlanStale    EditPlanState = "stale"

    EditTrimClip          EditOperationType = "trim_clip"
    EditRemoveRange       EditOperationType = "remove_range"
    EditMoveClip          EditOperationType = "move_clip"
    EditAddCaption        EditOperationType = "add_caption"
    EditAddTitle          EditOperationType = "add_title"
    EditSetTransition     EditOperationType = "set_transition"
    EditSetFilter         EditOperationType = "set_filter"
    EditAddEffectLayer    EditOperationType = "add_effect_layer"
    EditAddBGM            EditOperationType = "add_bgm"
    EditSetVolumeEnvelope EditOperationType = "set_volume_envelope"
    EditSetProjectFormat  EditOperationType = "set_project_format"
)
```

The parser must strip one outer Markdown JSON fence, decode with `DisallowUnknownFields`, reject trailing content, copy server-owned envelope fields (`ID`, `Owner`, `ProjectID`, `ClientRequestID`, `BaseRevision`, `ModelID`, timestamps), and then call `NormalizeEditPlan`.

- [ ] **Step 4: Run focused tests**

Run: `go test ./internal/aicut -run 'TestNormalizeEditPlan|TestParseEditPlan'`

Expected: PASS.

- [ ] **Step 5: Commit the contract**

```powershell
git add backend/internal/aicut/edit_plan.go backend/internal/aicut/edit_plan_test.go
git commit -m "feat: define AI-CUT edit plan contract"
```

### Task 2: Pure timeline application and single-step undo

**Files:**
- Create: `backend/internal/aicut/edit_plan_apply.go`
- Create: `backend/internal/aicut/edit_plan_apply_test.go`
- Modify: `backend/internal/aicut/types.go`

**Interfaces:**
- Consumes: validated `EditPlan` from Task 1 and existing `Timeline`.
- Produces: `ApplyEditPlan(Timeline, EditPlan) (Timeline, EditPlanApplySummary, error)`.

- [ ] **Step 1: Write failing application tests**

Use a two-clip video track plus one subtitle and one audio track. Assert:

```go
func TestApplyEditPlanAppliesEnabledOperationsOnly(t *testing.T) {
    result, summary, err := ApplyEditPlan(baseTimeline(), planWithTrimCaptionTransitionAndDisabledDelete())
    if err != nil { t.Fatal(err) }
    if summary.Applied != 3 || summary.Skipped != 1 { t.Fatalf("summary=%#v", summary) }
    if result.Revision != baseTimeline().Revision { t.Fatalf("pure function changed revision") }
    // Assert trim source continuity, new caption, and the transition on the second adjacent video clip.
}
```

Add tests for missing clip, locked track, source-bound overflow, overlapping move, non-adjacent transition, absent BGM asset reference, and duplicate generated IDs.

- [ ] **Step 2: Verify tests fail**

Run: `go test ./internal/aicut -run 'TestApplyEditPlan'`

Expected: FAIL because `ApplyEditPlan` is undefined.

- [ ] **Step 3: Implement deterministic operations**

Implement all Task 1 operation types without I/O. Reuse existing validation rules and preserve source timing. `remove_range` may only remove or shorten clips intersecting its range; it must not invent media. Generate new IDs from `operation.ID` with stable prefixes so retrying validation produces the same draft.

Return:

```go
type EditPlanApplySummary struct {
    Applied            int   `json:"applied"`
    Skipped            int   `json:"skipped"`
    RemovedDurationMS  int64 `json:"removedDurationMs"`
    EstimatedDurationMS int64 `json:"estimatedDurationMs"`
}
```

Call existing timeline validation before returning. Do not increment revision inside the pure function; `Store.SaveTimeline` owns the revision and history step.

- [ ] **Step 4: Run focused and existing timeline tests**

Run: `go test ./internal/aicut -run 'TestApplyEditPlan|TestValidate|TestMemoryStoreTimelineHistory'`

Expected: PASS.

- [ ] **Step 5: Commit application logic**

```powershell
git add backend/internal/aicut/edit_plan_apply.go backend/internal/aicut/edit_plan_apply_test.go backend/internal/aicut/types.go
git commit -m "feat: apply AI edit plans to timelines"
```

### Task 3: Durable plan storage and idempotency

**Files:**
- Create: `backend/internal/aicut/edit_plan_store.go`
- Create: `backend/internal/aicut/edit_plan_store_test.go`
- Create: `backend/internal/aicut/edit_plan_mysql.go`
- Create: `backend/internal/aicut/edit_plan_mysql_test.go`
- Modify: `backend/internal/aicut/mysql_store.go`
- Modify: `backend/internal/storage/aicut_schema.go`
- Modify: `backend/internal/storage/aicut_schema_test.go`

**Interfaces:**
- Produces: `EditPlanStore` with `CreateEditPlan`, `EditPlan`, `EditPlanByRequest`, `UpdateEditPlan`, and `ApplyStoredEditPlan`.
- Consumes: `EditPlan` and `ApplyEditPlan` from Tasks 1-2.

- [ ] **Step 1: Write failing memory-store and schema tests**

Test that `(owner, project_id, client_request_id)` returns the original plan, cancel is idempotent, applied plans cannot be applied twice, and a changed timeline revision marks the plan stale without modifying the timeline.

Schema assertions must require:

```text
aicut_edit_plans
aicut_edit_plan_operations
client_request_id
base_revision
state
request_json
summary_json
warning_json
applied_revision
```

- [ ] **Step 2: Verify failures**

Run: `go test ./internal/aicut ./internal/storage -run 'EditPlan|AICutSchema'`

Expected: FAIL because store methods and schema fields are absent.

- [ ] **Step 3: Implement store and migration**

Use a unique key on `(owner_username, project_id, client_request_id)`. Store request, summary, warnings, and each operation as JSON with an ordinal. `ApplyStoredEditPlan` must:

1. begin a SQL transaction;
2. lock the plan and project rows;
3. return the existing result for `applied`;
4. compare the timeline revision with `base_revision`;
5. mark mismatches `stale`;
6. run `ApplyEditPlan`;
7. save exactly one timeline history entry using the same transaction pattern as `SaveTimeline`;
8. mark the plan `applied` with `applied_revision`;
9. commit.

The in-memory store must implement the same observable behavior.

- [ ] **Step 4: Run store tests**

Run: `go test ./internal/aicut ./internal/storage -run 'EditPlan|AICutSchema'`

Expected: PASS.

- [ ] **Step 5: Commit persistence**

```powershell
git add backend/internal/aicut/edit_plan_store.go backend/internal/aicut/edit_plan_store_test.go backend/internal/aicut/edit_plan_mysql.go backend/internal/aicut/edit_plan_mysql_test.go backend/internal/aicut/mysql_store.go backend/internal/storage/aicut_schema.go backend/internal/storage/aicut_schema_test.go
git commit -m "feat: persist AI-CUT edit plans"
```

### Task 4: Backend prompt, model planning, and HTTP lifecycle

**Files:**
- Create: `backend/internal/prompt/aicut_edit_plan.go`
- Create: `backend/internal/prompt/aicut_edit_plan_test.go`
- Create: `backend/internal/httpapi/aicut_edit_plan_routes.go`
- Create: `backend/internal/httpapi/aicut_edit_plan_routes_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `backend/internal/app/app.go`

**Interfaces:**
- Consumes: existing `AICutTextProviderResolver`, `batchfactoryv11.DirectorProvider`, `EditPlanStore`, timeline store, and Tasks 1-3.
- Produces endpoints under `/api/ai-cut/projects/{id}/edit-plans`.

- [ ] **Step 1: Write prompt contract tests**

Assert the system prompt includes the exact operation whitelist, says “只输出 JSON”, forbids commands/paths/SQL/URLs, and contains no database credentials. Assert the user prompt includes only the selected project snapshot, request options, captions, track/clip IDs, time ranges, and available resource IDs.

- [ ] **Step 2: Write route tests**

Cover:

```text
POST   /api/ai-cut/projects/{id}/edit-plans
GET    /api/ai-cut/projects/{id}/edit-plans/{planId}
POST   /api/ai-cut/projects/{id}/edit-plans/{planId}/cancel
POST   /api/ai-cut/projects/{id}/edit-plans/{planId}/apply
```

Required cases: success, empty assets, missing model, invalid model JSON, duplicate `clientRequestId`, cancel, revision conflict/stale, duplicate apply, and unauthorized project.

- [ ] **Step 3: Verify failures**

Run: `go test ./internal/prompt ./internal/httpapi -run 'EditPlan'`

Expected: FAIL because the contract and routes do not exist.

- [ ] **Step 4: Implement prompt and synchronous first planner**

`POST` validates this request:

```go
type createEditPlanInput struct {
    ClientRequestID  string   `json:"clientRequestId"`
    ModelID          string   `json:"modelId"`
    ExpectedRevision int64    `json:"expectedRevision"`
    AssetIDs         []string `json:"assetIds"`
    Goal             string   `json:"goal"`
    TargetDurationMS int64    `json:"targetDurationMs"`
    AspectRatio      string   `json:"aspectRatio"`
    IncludeCaptions  bool     `json:"includeCaptions"`
    IncludeTransitions bool   `json:"includeTransitions"`
    IncludeFilters   bool     `json:"includeFilters"`
    IncludeBGM       bool     `json:"includeBgm"`
    ExtraRequirement string   `json:"extraRequirement"`
}
```

For this core slice, call the existing provider synchronously with request cancellation support, save `planning`, then `ready` or `failed`. Do not claim Redis execution. Return HTTP 201 for the first request and HTTP 200 for an idempotent repeat.

`apply` calls `ApplyStoredEditPlan`; `cancel` only transitions queued/planning/ready plans to canceled; `GET` never exposes owner or provider credentials.

- [ ] **Step 5: Run route and package tests**

Run: `go test ./internal/prompt ./internal/httpapi ./internal/aicut -run 'EditPlan'`

Expected: PASS.

- [ ] **Step 6: Commit backend lifecycle**

```powershell
git add backend/internal/prompt/aicut_edit_plan.go backend/internal/prompt/aicut_edit_plan_test.go backend/internal/httpapi/aicut_edit_plan_routes.go backend/internal/httpapi/aicut_edit_plan_routes_test.go backend/internal/httpapi/router.go backend/internal/app/app.go
git commit -m "feat: add AI-CUT edit plan API"
```

### Task 5: Frontend API and draft-review drawer

**Files:**
- Create: `frontend/src/shared/api/aicutEditPlan.js`
- Create: `frontend/src/shared/api/aicutEditPlan.test.js`
- Create: `frontend/src/user/components/AICutAutoEditDrawer.jsx`
- Create: `frontend/src/user/components/AICutAutoEditDrawer.source.test.js`
- Modify: `frontend/src/user/pages/aicut-page.css`

**Interfaces:**
- Consumes: Task 4 endpoints and the existing model/asset lists.
- Produces: `AICutAutoEditDrawer({ open, projectId, revision, assets, models, onClose, onApplied })`.

- [ ] **Step 1: Write failing API tests**

Assert safe IDs, non-empty goal/model/assets, target duration bounds `0..7_200_000`, supported ratios, `clientRequestId`, request serialization, cancel, and apply routes.

- [ ] **Step 2: Write failing UI source tests**

Require visible Chinese copy for:

```text
AI 自动剪辑
选择素材
剪辑目标
目标时长
字幕
转场
滤镜
背景音乐
生成剪辑草案
预计成片时长
操作理由
应用到时间线
取消草案
草案已过期，请基于最新工程重新生成
```

Also require that apply is disabled until a `ready` plan exists and no timeline-changing callback runs during draft generation or cancellation.

- [ ] **Step 3: Verify failures**

Run: `node --test src/shared/api/aicutEditPlan.test.js src/user/components/AICutAutoEditDrawer.source.test.js`

Expected: FAIL because the files do not exist.

- [ ] **Step 4: Implement API helpers and three-step drawer**

The drawer must have:

1. parameters and validation;
2. planning status and cancel;
3. summary plus one switch per operation and explicit apply.

When the user disables an operation, send `enabledOperationIds` with apply rather than mutating the stored plan in the browser. Display backend errors verbatim when they are already Chinese; otherwise use “AI 剪辑草案生成失败” or “剪辑草案应用失败”. Use theme variables already present in `aicut-page.css`.

- [ ] **Step 5: Run frontend focused tests**

Run: `node --test src/shared/api/aicutEditPlan.test.js src/user/components/AICutAutoEditDrawer.source.test.js`

Expected: PASS.

- [ ] **Step 6: Commit frontend feature**

```powershell
git add frontend/src/shared/api/aicutEditPlan.js frontend/src/shared/api/aicutEditPlan.test.js frontend/src/user/components/AICutAutoEditDrawer.jsx frontend/src/user/components/AICutAutoEditDrawer.source.test.js frontend/src/user/pages/aicut-page.css
git commit -m "feat: add AI auto-edit draft review"
```

### Task 6: AI-CUT page integration and one-history-step acceptance

**Files:**
- Modify: `frontend/src/user/pages/AICutPage.jsx`
- Modify: `frontend/src/user/pages/aicut-controls.contract.test.js`
- Modify: `frontend/src/user/pages/aicut-workbench.layout.test.js`
- Create: `docs/acceptance/aicut-ai-edit-plan-core.md`

**Interfaces:**
- Consumes: `AICutAutoEditDrawer` and current project/timeline/model/asset state.
- Produces: visible toolbar entry and timeline refresh after one confirmed application.

- [ ] **Step 1: Add failing page-contract tests**

Require an “AI 自动剪辑” button in the top bar, the drawer binding, `timeline.revision`, assets, models, `onApplied`, and no direct prompt string in `AICutPage.jsx`.

- [ ] **Step 2: Verify failure**

Run: `node --test src/user/pages/aicut-controls.contract.test.js src/user/pages/aicut-workbench.layout.test.js`

Expected: FAIL because the entry and drawer are absent.

- [ ] **Step 3: Integrate without changing the editor skeleton**

Add a single top-bar button. On successful apply:

1. replace local timeline with the returned saved timeline;
2. refresh history state;
3. close the drawer;
4. show `已应用 N 项 AI 剪辑操作，可使用撤销恢复`;
5. keep the current selected project and layout.

The Agent panel remains visible and unchanged in this task; connecting Agent suggestions to the same draft endpoint is a later plan.

- [ ] **Step 4: Run the complete focused frontend set**

Run:

```powershell
node --test src/shared/api/aicutEditPlan.test.js src/user/components/AICutAutoEditDrawer.source.test.js src/user/pages/aicutTimeline.test.js src/user/pages/aicutWorkbench.test.js src/user/pages/aicut-controls.contract.test.js src/user/pages/aicut-workbench.layout.test.js
```

Expected: PASS.

- [ ] **Step 5: Run backend regression and frontend production build**

Run:

```powershell
Set-Location backend
go test ./internal/aicut ./internal/prompt ./internal/httpapi ./internal/storage
Set-Location ..\frontend
npm run build
```

Expected: all selected Go packages PASS and Vite production build succeeds.

- [ ] **Step 6: Rebuild the embedded runtime**

Run from repository root:

```powershell
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached
```

Verify:

```powershell
Invoke-RestMethod http://127.0.0.1:4001/health
Invoke-RestMethod http://127.0.0.1:4001/api/runtime-build-info
git rev-parse --short HEAD
```

Expected: health is true and runtime build information identifies the current build rather than the old `c954ab5a` frontend bundle.

- [ ] **Step 7: Perform browser acceptance without a paid provider call**

Use a test provider or route test fixture to verify visibly:

1. opening “AI 自动剪辑”;
2. validation for empty material/model/goal;
3. a ready draft with enabled and disabled operations;
4. cancel leaves the timeline unchanged;
5. apply changes the timeline once;
6. a second apply does not duplicate clips;
7. undo restores the exact pre-apply timeline;
8. modifying the project before apply produces the stale message.

Record the input, expected timeline delta, actual delta, runtime SHA, screenshot paths, and test commands in `docs/acceptance/aicut-ai-edit-plan-core.md`.

- [ ] **Step 8: Commit integration and evidence**

```powershell
git add frontend/src/user/pages/AICutPage.jsx frontend/src/user/pages/aicut-controls.contract.test.js frontend/src/user/pages/aicut-workbench.layout.test.js docs/acceptance/aicut-ai-edit-plan-core.md
git commit -m "feat: integrate AI-CUT auto-edit plans"
```

## Self-Review Result

- Spec coverage for the first vertical slice: draft generation, review, cancellation, explicit apply, idempotency, revision conflict, one-step undo, Chinese failures, tests, embed build, and browser evidence are each assigned to a task.
- Deferred by design: media preprocessing, Redis plan workers, Agent reuse, BGM ducking, TOS result delivery, and advanced exports each require a separate independently testable plan after this core contract exists.
- Placeholder scan: no `TBD`, `TODO`, “implement later”, or undefined neighboring interface remains.
- Type consistency: `EditPlan`, `EditOperation`, `EditPlanStore`, `ApplyEditPlan`, endpoint names, and frontend component props use the same names throughout all tasks.

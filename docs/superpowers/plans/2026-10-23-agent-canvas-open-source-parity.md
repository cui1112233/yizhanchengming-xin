# Agent Canvas Open-Source Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make V88 Agent Canvas a truthful, production-ready visual workflow: every supported connection becomes a real model input and every long-running media operation has a durable, inspectable status.

**Architecture:** Keep V88's existing React + Go embed + MySQL + Redis + TOS architecture. Do not transplant the reference project's Tauri, SQLite, sidecar, or CLI implementation. Add one server-owned input-resolution contract used by image, video, and audio execution; keep Workflow and Storyboard as two views over the same `agent_projects`, nodes, edges, assets, and generation tasks.

**Tech Stack:** React, `@xyflow/react`, Go HTTP API, MySQL, Redis task coordination, TOS object storage, FFmpeg, existing provider adapters, Go embed binary.

## Global Constraints

- `v88` is the maintained source; never copy compiled frontend files back into source.
- Do not create a second canvas, second storyboard database, or a parallel task queue.
- Keep `agent_generation_tasks` as the shared status record; keep provider-specific video details in `agent_video_jobs`.
- Credentials stay server-side. Never test paid provider generation without the user's explicit approval.
- Treat a connection as supported only when its resolved input is included in the actual provider request and the result is persisted/read back.
- If a provider cannot accept a connection type, block that connection or mark it unavailable before generation; never silently discard it.
- Preserve the dirty checkout: do not reset, clean, overwrite unrelated files, commit, push, deploy, or alter external credentials without explicit approval.
- Each phase requires focused tests, embedded Go rebuild, local `:4001` restart, and authenticated browser acceptance before it is called complete.

---

## Delivery Order

1. **Truth layer:** connection input resolution, capability rules, and visible status.
2. **Reuse loop:** global history, library, result versions, export.
3. **Media tools:** image, audio, then video processing jobs.
4. **Advanced operations:** 3D director desk and controlled external canvas operations.

---

### Task 1: Create One Truthful Connection-Input Contract

**Files:**
- Create: `backend/internal/agentworkspace/canvas_inputs.go`
- Create: `backend/internal/agentworkspace/canvas_inputs_test.go`
- Create: `backend/internal/httpapi/agent_canvas_inputs.go`
- Create: `backend/internal/httpapi/agent_canvas_inputs_test.go`
- Modify: `frontend/src/user/pages/agent-canvas-connections.js`
- Modify: `frontend/src/user/pages/agent-canvas-node-generation.js`
- Modify: `frontend/src/user/pages/agent-canvas-video-jobs.js`
- Modify: `frontend/src/user/pages/agent-canvas-connections.test.js`

**Consumes:** canvas nodes, edges, node `assetRef`, model capability catalog, project assets.

**Produces:**

```go
type ResolvedCanvasInputs struct {
    Prompt        string
    Images        []ResolvedAsset
    Videos        []ResolvedAsset
    Audios        []ResolvedAsset
    SourceNodeIDs []string
    Warnings      []string
}

func ResolveCanvasInputs(canvas Canvas, targetNodeID string, capability ModelCapability) (ResolvedCanvasInputs, error)
```

- [ ] **Step 1: Write failing input-resolution tests.** Cover text → image, text → video, image → image, image → video, audio → video, video → video, duplicate assets, stale/unconfirmed assets, and a provider that does not accept video or audio references.
- [ ] **Step 2: Run the focused tests.** Run `go test ./internal/agentworkspace ./internal/httpapi -run CanvasInputs -count=1`; expected result before implementation is failure for missing resolver behavior.
- [ ] **Step 3: Implement the server resolver.** Compose upstream text in topological order before the node's own prompt; resolve only confirmed project assets; preserve input order; emit a typed unsupported-input error instead of dropping inputs.
- [ ] **Step 4: Replace frontend-only inference.** Keep `syncCanvasEffectiveReferences` only for display. On generate, call the resolver preview endpoint and show exact inputs: prompt sources, first-frame images, video references, audio references, and blocked items.
- [ ] **Step 5: Verify.** Add UI tests asserting a connection label and a provider request preview match. Browser-check one non-paid draft flow for each supported connection type.

### Task 2: Enforce Per-Model Connection Capability Rules

**Files:**
- Modify: `backend/internal/httpapi/model_catalog_capabilities.go`
- Modify: `backend/internal/httpapi/model_catalog.go`
- Modify: `frontend/src/user/pages/agent-model-capabilities.js`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Test: `backend/internal/httpapi/model_catalog_capabilities_test.go`
- Test: `frontend/src/user/pages/agent-model-capabilities.test.js`

**Consumes:** `ResolvedCanvasInputs` from Task 1 and existing provider/model catalog records.

**Produces:** model flags `acceptsTextPrompt`, `acceptsImageReferences`, `acceptsVideoReferences`, `acceptsAudioReferences`, and `maxReferenceAssets`.

- [ ] **Step 1: Add failing model-capability tests.** Assert every selectable model has explicit media-input flags; unknown flags default to unsupported rather than allowed.
- [ ] **Step 2: Run the capability tests.** Run the two focused backend/frontend test files and confirm the old catalog cannot describe connection support.
- [ ] **Step 3: Persist and return capability flags.** Extend public model records without exposing credentials; set values only from server-side provider adapters.
- [ ] **Step 4: Make unsupported edges honest.** Disable unsupported drop targets, show the reason in the node, and block submission if old saved edges violate the selected model.
- [ ] **Step 5: Verify provider mapping.** Test Doubao/ComfyUI/local executor/video adapters with fixtures. Do not claim any external provider executed until separately authorized.

### Task 3: Wire Resolved Inputs Into Every Generator

**Files:**
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Modify: `backend/internal/httpapi/agent_image_coordinator.go`
- Modify: `backend/internal/httpapi/agent_audio_jobs.go`
- Modify: `backend/internal/httpapi/agent_video_jobs.go`
- Modify: `backend/internal/httpapi/agent_image_generator_test.go`
- Modify: `backend/internal/httpapi/agent_audio_jobs_test.go`
- Modify: `backend/internal/httpapi/agent_video_jobs_test.go`

**Consumes:** Tasks 1-2 and existing durable generation coordinators.

**Produces:** a snapshot of `resolvedInputs` saved inside each submitted task, then passed unchanged to the chosen executor.

- [ ] **Step 1: Write request-capture tests.** Use fake image, TTS, and video executors. Assert the exact upstream prompt/media list arriving at each fake equals the stored `resolvedInputs` snapshot.
- [ ] **Step 2: Run the tests and confirm failure.** Run `go test ./internal/httpapi -run "TestAgent.*ResolvedInputs" -count=1`.
- [ ] **Step 3: Persist the resolved snapshot at submit time.** Save text/media lineage, node versions, model capability signature, and order. Coordinators must never recompute from a changed canvas after queueing.
- [ ] **Step 4: Add safety outcomes.** Mark a task `failed` with an actionable message for deleted assets, incompatible model settings, or stale node versions; never change it to completed merely because a provider response is absent.
- [ ] **Step 5: Verify cancel/retry behavior.** Retried work resolves against the current node revision and records `retryOfTaskId`; existing tasks retain their original snapshot for audit.

### Task 4: Make Node Status And Result Lineage Visible

**Files:**
- Modify: `frontend/src/user/pages/agent-generation-state.js`
- Modify: `frontend/src/user/pages/agent-result-lineage.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Test: `frontend/src/user/pages/agent-generation-state.test.js`
- Test: `frontend/src/user/pages/agent-result-lineage.test.js`

**Consumes:** common `agent_generation_tasks`, video job provider states, resolved-input snapshots.

**Produces:** one presentation model: queued, running, awaiting choice, completed, failed, cancelled; plus exact input/result lineage displayed on both Workflow and Storyboard.

- [ ] **Step 1: Add failing state tests.** Cover all seven task states, missing records, stale data, multi-result choice, retry chain, and one node with incompatible upstream input.
- [ ] **Step 2: Implement a single node-status mapper.** Remove any fallback that renders a missing/unknown task as completed.
- [ ] **Step 3: Show an in-node compact status.** Display running phase, elapsed state, error, retry/cancel, selected asset, and “input from” chips without opening a separate page.
- [ ] **Step 4: Keep Workflow and Storyboard identical in facts.** Storyboard may change layout only; it must read the same status, task IDs, input source list, selected result, and retry state.
- [ ] **Step 5: Browser acceptance.** Start an intentionally invalid local draft and confirm failure/error is visible; create and remove any temporary test edges after verification.

### Task 5: Add Global Generation History And Result Reuse

**Files:**
- Create: `backend/internal/agentworkspace/generation_history.go`
- Create: `backend/internal/agentworkspace/mysql_generation_history.go`
- Create: `backend/internal/httpapi/agent_generation_history.go`
- Create: `backend/internal/httpapi/agent_generation_history_test.go`
- Create: `frontend/src/user/pages/agent-generation-history.js`
- Create: `frontend/src/user/components/AgentGenerationHistoryDrawer.jsx`
- Modify: `frontend/src/shared/api/agent.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Consumes:** persisted generation tasks/results and project assets.

**Produces:** authenticated owner-wide history across projects with query, kind/status/model filters, result preview, delete policy, and “send to current canvas.”

- [ ] **Step 1: Write SQL/API tests.** Seed two projects for one owner and another owner. Assert owner isolation; filter by image/video/audio, state, model, date, and text query; reject cross-owner result attachment.
- [ ] **Step 2: Add an indexed history query.** Use a new append-only migration for indexes only after checking existing MySQL keys. Do not copy blob data; reference existing result asset IDs.
- [ ] **Step 3: Add send-to-canvas mutation.** Create a new media node in the target project that references the existing immutable asset and records `originProjectId`, `generationTaskId`, and result ID.
- [ ] **Step 4: Build the drawer.** Make it searchable, filterable, and explicit about source project/status. Use multi-select only after the single-item path is proved.
- [ ] **Step 5: Verify browser behavior.** Send one fixture result from project A into project B and confirm the asset resolves after reload.

### Task 6: Build A Cross-Project Folder Material Library

**Files:**
- Create: `backend/internal/agentworkspace/library.go`
- Create: `backend/internal/agentworkspace/mysql_library.go`
- Create: `backend/internal/httpapi/agent_library.go`
- Create: `frontend/src/user/components/AgentMaterialLibraryDrawer.jsx`
- Create: `frontend/src/user/pages/agent-material-library.js`
- Modify: `frontend/src/user/components/AgentWorkspaceDrawer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Consumes:** project assets and history result assets.

**Produces:** owner-scoped folders, library items, tags, source provenance, and “insert into current canvas.”

- [ ] **Step 1: Write store tests.** Cover folders, item move, duplicate prevention, owner isolation, source-project deletion policy, and preserved TOS object references.
- [ ] **Step 2: Add migrations and store methods.** Create `agent_library_folders` and `agent_library_items`; record an asset reference rather than duplicate media bytes.
- [ ] **Step 3: Add explicit save/import API.** A source node can save one selected confirmed result to a folder; importing creates a new current-project node plus a traceable linked project asset record.
- [ ] **Step 4: Build the drawer.** Support folder navigation, image/video/audio filters, search, rename, move, and insert. Do not mix it with the current-project material panel.
- [ ] **Step 5: Verify persistence.** Save a fixture image, reload, open a different project, insert it, and confirm both provenance and content URL work.

### Task 7: Implement Image Post-Processing Jobs

**Files:**
- Create: `backend/internal/agentworkspace/image_jobs.go`
- Create: `backend/internal/httpapi/agent_image_tools.go`
- Create: `backend/internal/httpapi/agent_image_tools_test.go`
- Create: `frontend/src/user/components/AgentImageToolsPopover.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-generation-state.js`

**Scope:** rotate/mirror, crop, grid split, annotation, upscale, preset-guided multi-angle, lighting, panorama.

- [ ] **Step 1: Split local and model-backed operations.** Implement rotate/mirror, crop, grid split, and annotation as deterministic processing jobs; implement upscale/multi-angle/lighting/panorama only through a model capability that expressly supports them.
- [ ] **Step 2: Write failing operation tests.** Verify every operation produces a new asset/result, preserves its source asset lineage, validates parameters, and never overwrites the original.
- [ ] **Step 3: Add one durable image-tool task mode.** Reuse `agent_generation_tasks` with modes such as `image_crop`, `image_grid_split`, and `image_upscale`; expose progress/cancel/retry through the common UI.
- [ ] **Step 4: Add compact selected-image controls.** Show only operations compatible with the selected image/model; results remain candidate versions until selected or sent to canvas.
- [ ] **Step 5: Verify one deterministic and one mocked model-backed path.** Browser-check crop/grid and use a fake provider for a model tool; do not spend provider credits without approval.

### Task 8: Finish Audio Tools And Audio-to-Video Semantics

**Files:**
- Modify: `backend/internal/httpapi/agent_audio_processor.go`
- Modify: `backend/internal/httpapi/agent_audio_jobs.go`
- Create: `backend/internal/httpapi/agent_audio_waveform.go`
- Create: `frontend/src/user/components/AgentAudioToolsPopover.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-audio-jobs.js`
- Modify: `frontend/src/user/components/AuthenticatedAgentAudio.jsx`

**Scope:** waveform preview, trim, speed, smart/custom split, TTS/music/effect, and supported audio conditioning for video.

- [ ] **Step 1: Add waveform tests.** Generate bounded waveform summary JSON from a stored asset without returning raw audio bytes; reject non-audio assets.
- [ ] **Step 2: Keep existing FFmpeg operations durable.** Verify trim/speed/split remain background tasks and each output is an independent project asset with input lineage.
- [ ] **Step 3: Define video audio behavior per model.** For models that take audio conditioning, map resolved audio to the documented request field. For models that do not, prevent audio → video edge creation and explain the limitation.
- [ ] **Step 4: Add the selected-audio UI.** Show waveform, duration, source, split outputs, selected segment, and its downstream compatibility.
- [ ] **Step 5: Verify.** Use fixture audio for waveform and trim; assert an unsupported video model cannot silently accept its audio edge.

### Task 9: Implement Video Processing And Frame Extraction Jobs

**Files:**
- Create: `backend/internal/agentworkspace/video_processing_jobs.go`
- Create: `backend/internal/httpapi/agent_video_tools.go`
- Create: `backend/internal/httpapi/agent_video_tools_test.go`
- Create: `frontend/src/user/components/AgentVideoToolsPopover.jsx`
- Modify: `frontend/src/user/components/AuthenticatedAgentVideo.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Scope:** trim, crop, extract frames, audio separation, model-backed continuation/reshoot/upscale/white-model/subtitle removal where supported.

- [ ] **Step 1: Implement deterministic first.** Ship trim, crop, extract-frames, and audio-separation with FFmpeg as queued tasks. Frame extraction must let the user select frames then create image nodes on the canvas.
- [ ] **Step 2: Write failing task/result tests.** Assert selected frames preserve timestamps/order, each output is a persisted asset, and cancellation leaves no false success state.
- [ ] **Step 3: Add capability-gated model tools.** Continuation, reshoot, upscale, white-model, and subtitle removal appear only for provider adapters with a verified request/response contract.
- [ ] **Step 4: Preserve true video references.** Enable video → video only for a model with `acceptsVideoReferences=true`; otherwise block the edge before submit.
- [ ] **Step 5: Browser-check extraction.** Upload/use a small fixture video, select two frames, send them to the canvas, reload, and confirm both image nodes remain connected to the original video lineage.

### Task 10: Add Safe Export And Project Delivery

**Files:**
- Create: `backend/internal/httpapi/agent_exports.go`
- Create: `backend/internal/httpapi/agent_exports_test.go`
- Create: `frontend/src/user/components/AgentExportDialog.jsx`
- Modify: `frontend/src/user/components/AgentStoryboardWorkspace.jsx`
- Modify: `frontend/src/shared/api/agent.js`

**Consumes:** selected confirmed assets and project history.

**Produces:** individual download, selected-result ZIP export, and a manifest describing source nodes/task IDs/model/provider/timestamps.

- [ ] **Step 1: Write export authorization tests.** Verify only the owner can export, missing/orphaned assets are reported, and ZIP entries use safe filenames.
- [ ] **Step 2: Build a streaming ZIP response.** Use server-created manifest plus asset content streams; do not create an unbounded temp archive in memory.
- [ ] **Step 3: Add export UI.** Let the user choose selected assets by canvas order, all completed media nodes, or one task result; show blocked/missing entries before starting.
- [ ] **Step 4: Add “save node output” shortcut.** Keep it a browser download, not a fake local-save action.
- [ ] **Step 5: Verify a fixture ZIP.** Read its manifest and count entries in a focused API test; browser-check a small download.

### Task 11: Add 3D Director Desk As A Separate, Connected Node Type

**Files:**
- Create: `backend/internal/agentworkspace/director_desk.go`
- Create: `backend/internal/httpapi/agent_director_desk.go`
- Create: `frontend/src/user/components/AgentDirectorDeskNode.jsx`
- Create: `frontend/src/user/pages/agent-director-desk.js`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`

**Consumes:** project assets/library characters and scenes.

**Produces:** a persisted director-desk node with scene layout, character placement, camera settings, and screenshot output that becomes a normal image asset/reference.

- [ ] **Step 1: Define the minimum usable desk contract.** Persist scene ID, placed actors, transforms, camera, light presets, and renderer version; do not start with a complex physics engine.
- [ ] **Step 2: Write serialization and screenshot-request tests.** Confirm project reload preserves placements and the screenshot result has normal asset lineage.
- [ ] **Step 3: Implement a 2.5D first release.** Use layered scene/character placement and camera framing sufficient for composition reference; label it honestly as previsualization.
- [ ] **Step 4: Generate screenshot assets.** Capture the canvas/renderer output server-side or through a controlled client upload, then save it as a project image asset and allow image/video connections.
- [ ] **Step 5: Verify.** Place one character and camera, save/reload, create a screenshot, connect it to an image/video node, and inspect resolved inputs.

### Task 12: Complete Controlled Agent And External Canvas Operations

**Files:**
- Modify: `backend/internal/agentworkspace/director_commands.go`
- Modify: `backend/internal/httpapi/agent_workspace_compat.go`
- Create: `backend/internal/httpapi/agent_canvas_manifest.go`
- Create: `backend/internal/httpapi/agent_canvas_manifest_test.go`
- Modify: `frontend/src/user/components/AgentDirectorConversation.jsx`
- Create: `docs/agent-canvas-command-contract.md`

**Consumes:** existing director command validation, resolved inputs, task state machine.

**Produces:** a restricted manifest/API contract for read canvas, create/update/connect nodes, queue generation, cancel/retry, and apply results, all audited as project executions.

- [ ] **Step 1: Write authorization and audit tests.** Reject cross-project changes, unsupported node types, direct credential changes, and unapproved provider submission; record every accepted command with before/after references.
- [ ] **Step 2: Publish a machine-readable manifest.** Return allowed node types, model capability IDs, command schemas, and task state actions for the active project.
- [ ] **Step 3: Require preview before destructive canvas changes.** Agent conversation shows proposed nodes/edges/inputs and applies only after explicit user confirmation when the command removes nodes, changes links, or starts paid generation.
- [ ] **Step 4: Do not copy the reference CLI.** If a CLI is later needed, build a V88-authenticated local client against this manifest; it must use the same server validation/audit path.
- [ ] **Step 5: Verify with browser and API fixtures.** Create then remove a temporary node/edge, confirm audit history, and ensure no external paid task starts without confirmation.

### Task 13: Regression, Embedded Runtime, And Acceptance Gates

**Files:**
- Modify: `docs/V88_PROJECT_EXECUTION_MEMORY.md` only after a verified implementation changes durable behavior
- Modify: `docs/superpowers/specs/` only when a phase's accepted design changes
- Test: focused Go/React suites created above

- [ ] **Step 1: Run focused backend tests per completed task.** Use `go test ./internal/agentworkspace ./internal/httpapi ./internal/app ./cmd/qiantie -count=1` plus each new `-run` target.
- [ ] **Step 2: Run focused frontend tests.** Use exact new test paths before broad frontend tests; separately document any pre-existing unrelated failures.
- [ ] **Step 3: Rebuild the real app.** Run `npm --prefix frontend run build`, then `scripts/build-go-local.ps1`, then restart the approved local Go review runtime on `:4001`.
- [ ] **Step 4: Perform authenticated browser acceptance.** Confirm Workflow and Storyboard share the same facts; test each completed action through visible UI and clean temporary nodes/edges afterward.
- [ ] **Step 5: Record only proven results.** Distinguish source tests, Go-embedded runtime, browser UI, and externally executed provider work. Commit or deploy only when separately authorized.

---

## Coverage Check

- Connection truth: Tasks 1-4.
- Script production chain: explicitly out of scope; do not add or change script-node behavior in this program.
- History/material reuse/export: Tasks 5, 6, and 10.
- Image/audio/video tools: Tasks 7-9.
- 3D director desk: Task 11.
- Agent/CLI-equivalent controlled operations: Task 12.
- Full build, browser, and external-provider acceptance boundary: Task 13.

## Recommended First Execution Slice

Execute **Tasks 1-4** as one release-sized milestone. It fixes the fake chains first: text → generation, audio → video, video → video, and makes their true state visible. Then execute Tasks 5-6 before any large media-tool UI expansion.

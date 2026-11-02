# Agent Canvas Production Chain Design

Date: 2026-10-22

## Goal

Turn the existing Agent Canvas into a usable short-drama production chain without replacing its current workflow view, storyboard, project conversation, account permissions, or Go-embedded release path.

The first deliverable is a complete path from a script to selected image/video/audio assets. It must make each generated asset's source, current status, selected version, and downstream use visible.

## Scope

### Included in phase one

1. Add a `script` canvas node.
   - It stores source text and a selected script-processing model.
   - It submits one server-side split task and displays queued, running, succeeded, failed, and retryable states.
   - A completed split produces structured shots: order, scene, characters, dialogue, narration, image prompt, video prompt, and TTS text.

2. Add a shot list to the storyboard workspace.
   - It is the project-level view of the script node's shots, not a second copy of data.
   - Selecting a shot locates its source script node and all assets generated for that shot.
   - Users can edit a shot before generation and save the change to the same project record.

3. Create batch actions from selected shots.
   - Create image nodes, video nodes, or TTS audio nodes only for the selected shots.
   - New nodes receive an immutable `source` record containing script node ID, shot ID, operation, and attempt number.
   - The actions use the existing image/video/audio job APIs and configured provider-model capability rules.

4. Add per-node result versions.
   - Each attempt is retained as a version with status, timestamps, error message, asset ID, and source record.
   - One version is marked active. Only the active version is exposed to downstream references and the storyboard by default.
   - A user can select an older successful version, retry a failed version, or create another attempt without overwriting history.

5. Add a project result library.
   - It lists generated and uploaded media by shot, type, source operation, and status.
   - Selecting an item can locate it in the workflow, make its version active, or add it to the storyboard.
   - It reuses the project asset storage and authenticated media rendering already used by the canvas.

6. Make state and failure truthful.
   - A submitted request is never shown as completed until its task succeeds and the result is persisted.
   - Failed tasks show the actual server/provider message, a retry action, and do not block unrelated shots.
   - A running task can be cancelled only when the provider/API supports cancellation; otherwise its UI states that it will finish in the background.

### Explicitly deferred

- 3D director desk and local desktop-only file workflow.
- Browser-only video transcoding, waveform editing, and heavyweight editing tools.
- Replacing React Flow, the current storyboard, project conversation, or account/permission rules.
- Rewriting provider integrations. Existing provider/model selection remains the single source of truth.

## Architecture

The frontend owns display and local editing state. Go owns script splitting, durable shot records, generation attempts, and project asset associations. MySQL stores project, shot, version, and task relationships. Redis continues to coordinate jobs and locks. TOS remains the media object store.

The canvas node is a view over durable server records. It may optimistically show a pending operation, but refresh/reload must reconstruct the same node status and active version from the server.

Each generated node records this lineage:

`project -> script node -> shot -> generation node -> attempt/version -> media asset`

This permits the storyboard, result library, task monitor, and Agent conversation to refer to the same asset without copying media or inventing separate status values.

## User Flow

1. User creates or opens a script node and saves source text.
2. User selects a configured text model and requests shot splitting.
3. The canvas and storyboard show the split task status; when complete, the shot list appears.
4. User adjusts selected shots and starts image, video, or TTS creation in bulk.
5. Every shot shows its node status and result versions. Failures remain visible but do not stop other shots.
6. User marks the intended result version active. The active asset is used by the storyboard and downstream nodes.
7. The result library can find any asset by shot and return to its workflow node.

## UI Boundaries

- Workflow is for dependency flow and node-level configuration.
- Storyboard is for reading, ordering, selecting, and reviewing shots.
- Result library is for browsing all uploaded/generated assets and versions.
- Project conversation can cite a script, shot, node, or version by the lineage record; it must not create detached copies.

## Error Handling

- Invalid or empty script text disables submission with a visible reason.
- Missing compatible models report the current configuration problem before a task is submitted.
- Provider failures preserve the provider's safe error message, task ID, and retry capability.
- Partial batch failures produce a finished batch summary: succeeded, failed, still running, and skipped counts.
- Retrying increments the attempt number and leaves the previous attempt intact.

## Verification

1. Focused frontend tests for shot selection, version activation, status rendering, and source-lineage labels.
2. Go API/storage tests for durable script shots, version selection, and retry behavior.
3. Real MySQL migration test using append-only Goose migrations.
4. Authenticated browser test: script -> split -> selected image/TTS batch -> failure/retry -> reload persistence -> locate from result library.
5. Build the React assets into the Go binary and verify the local `:4001` runtime, not only the development bundle.

## Delivery Order

1. Durable schema/API for script shots and generation lineage.
2. Script node and shot list, including persisted split status.
3. Selected-shot batch creation through existing generation APIs.
4. Per-node version selection and result library.
5. End-to-end reliability tests and visual acceptance.

## Acceptance Criteria

- A user can trace any active storyboard asset back to its script shot and generation attempt.
- Refreshing the page never loses shots, task state, selected versions, or source relationships.
- A provider failure is visible, retryable when supported, and never labeled as complete.
- A batch may partly succeed without leaving the project in an unknown state.
- Existing image/video/audio nodes, current storyboard materials, project conversation, and permissions continue to work unchanged.

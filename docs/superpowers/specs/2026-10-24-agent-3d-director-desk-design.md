# Agent 3D Director Desk Design

## Goal

Replace the current 2.5D director-desk modal with a usable full-screen 3D
previsualization workspace. It remains one `director` node in the existing
Agent Canvas project; it does not create a separate project, canvas, task
queue, or persistence system.

The first release must let a creator open a director node, arrange a scene and
actors, adjust the camera, save automatically, capture the composition, and
send that capture back to the current canvas as a normal image node.

## Boundaries

- Reuse and adapt the open-source director-desk React and Three.js workspace.
  Do not copy its Tauri shell, SQLite storage, local sidecar, or CLI.
- Keep V88's Go-embedded frontend, Go HTTP API, MySQL project data, existing
  asset APIs, TOS storage, and current canvas save mechanism.
- Keep Workflow and Storyboard as two presentations of the same project nodes,
  edges, assets, and director-node configuration.
- Do not start model generation during acceptance. A capture is a local scene
  snapshot uploaded through the existing project asset path.
- Do not import third-party model binaries merely because they exist in the
  reference repository. The first release uses project images and user-provided
  models only; model-library migration is a later feature with license review.
- The first release renders project-image actors as upright planes in a real
  Three.js scene with a movable 3D camera. It does not claim to include the
  reference project's bundled character models; GLB/FBX import and its model
  library are a second-stage feature.

## User Experience

1. The creator adds or opens a director node in the existing workflow.
2. The node shows a compact summary: selected scene, actor count, and last
   capture. It has one clear action: `Open 3D Director Desk`.
3. The action opens a full-screen workspace above the canvas. It owns pointer,
   wheel, context-menu, copy, paste, and undo interactions while open, so
   canvas gestures never leak into the editor.
4. The creator selects a project image as the scene or adds it as a reference,
   places available actors, and changes camera framing.
5. Each change updates the node's in-memory state immediately and is debounced
   to the existing durable canvas save path. Closing the workspace flushes the
   final pending save.
6. `Capture to Canvas` renders the current composition, uploads it as a
   project image asset, creates an image node to the right of the director
   node, and creates a director-to-image connection.

## Data Contract

The state lives in `node.data.config.directorProject` and is versioned:

```js
{
  rendererVersion: '3d-v1',
  scene: { assetId, url, name } | null,
  actors: [{ id, assetId, kind, position, rotation, scale }],
  camera: { position, target, fov },
  captures: [{ assetId, url, name, createdAt }]
}
```

- The first release accepts only assets owned by the current project.
- URLs are display values; asset IDs remain the durable references.
- Old `config.directorDesk` 2.5D data is read and converted to the initial
  `directorProject` view when possible. No existing node is deleted.
- The backend validates the `director` node type and ordinary canvas storage;
  it does not create an isolated director-desk database.

## Frontend Structure

- `AgentDirectorDeskNode`: compact canvas card and the full-screen entry point.
- `director-desk/hostAdapter`: small V88 bridge exposing read state, write
  state, project-image listing, upload, and capture delivery.
- `director-desk/App`: adapted open-source 3D workspace, lazily loaded with
  Three.js so normal canvas users do not download it.
- `AgentCanvas`: resolves the current director node, supplies the host bridge,
  receives a capture, creates the normal image node, and persists its edge.

The 3D workspace must not call V88 APIs directly. All V88-specific data access
goes through the host adapter, which makes it testable and prevents a second
asset or persistence path from appearing inside the renderer.

## Capture And Persistence Flow

```text
3D interaction
  -> hostAdapter.writeState
  -> existing canvas node-config update
  -> existing durable canvas save

Capture to Canvas
  -> render PNG in the 3D workspace
  -> hostAdapter.storeCapture
  -> existing project-asset upload
  -> AgentCanvas creates image node and edge
  -> existing durable canvas save
```

Failure rules:

- A failed asset upload leaves the desk open, shows a clear error, and creates
  no image node or connection.
- A failed canvas save marks the node as unsynchronized; the UI must not claim
  that the capture or scene layout is persisted.
- Closing the workspace waits for the final scheduled node update to flush.

## First-Release Exclusions

- Shared 3D model library, model categories, and model deletion/reference
  checks.
- Lighting presets, multi-camera presets, animation, physics, and rendering
  jobs.
- Copying the reference project's desktop shell, SQLite schema, local sidecar,
  or command-line interface.

These are separate follow-up work because they need their own asset lifecycle,
license review, and V88 server contracts.

## Verification

- Unit tests cover versioned state normalization, old 2.5D migration, actor
  transforms, camera updates, capture requests, and capture-result insertion.
- Browser acceptance uses a current project: open the desk, set scene and
  actor, change camera, close/reopen, refresh, and confirm the composition
  survives.
- Browser acceptance captures once and confirms the result is a persisted
  project image node connected to the director node.
- Build React, rebuild the Go-embedded local binary, restart `:4001`, and
  repeat browser acceptance against the embedded runtime.

## Acceptance Criteria

The first release is complete only when the user can arrange a saved 3D scene
inside one director node, reopen it after refresh without losing changes, and
use its capture as a normal image input in the existing canvas. The page must
not present the old 2.5D modal as the director desk after migration.

# V88 Agent Full Canvas Workspace Design

Date: 2026-10-20

Scope: local V88 `/agent/canvas`. This design uses interaction findings from the locally installed Shenlang AI Canvas package, but does not copy its code, branding, account system, cloud endpoints, or subscription behavior.

## 1. Goal

Turn the Agent canvas into the primary production workspace for a project. Users can arrange text, image, video, and audio results; connect references; inspect and edit one node at a time; run supported media work; and safely return results to the project.

The canvas keeps media previews on the workspace, while each generation node exposes a compact direct-input panel below its preview. The user can enter a prompt, add references, and select generation settings without first opening a modal.

## 2. Observed Reference Behavior

Static analysis of the supplied Electron package and an authenticated local UI inspection confirmed the following client behavior:

- A dot-grid workspace with project title, node count, save status, recent-canvas access, zoom control, and a visible shortcut entry point.
- Text, image, and video nodes. Media nodes expose a left reference connector and a right downstream connector.
- Selection, drag, box selection, copy, duplicate, paste, deletion, undo/redo, grouping, ungrouping, alignment, distribution, automatic layout, context menus, and viewport controls.
- Generation follows a task lifecycle: configuration and reference validation, submission, queued/running work, result download and metadata validation, asset persistence, confirmation, canvas application, retry/cancel/recovery when necessary.

Those findings are client-side observations only. Remote permissions, billing, and cloud synchronization are explicitly outside V88 scope.

## 3. V88 Page Layout

1. Header: return action, editable project name, save indicator, node count, recent project menu, and project actions.
2. Canvas: the main work surface with dot background, nodes, directed edges, group bounds, and non-intrusive connection motion.
3. Selection tools: a compact floating toolbar shown only when one or more nodes are selected.
4. View tools: bottom controls for zoom out/in, fit canvas, focus selection, and the visible shortcut reference.
5. Node input panel: image, video, and audio generation nodes show their prompt and controls directly below the preview. The panel owns reference thumbnails, a project-material `+` action, model selection, and supported settings. The modal remains only for title, full text, status, and advanced recovery actions.
6. Conversation: the Agent conversation remains available and can create, revise, or remove canvas content, but never silently deletes a user asset.

V88 keeps its existing visual palette. The reference product supplies interaction ideas only.

## 4. Node Contract

### 4.1 Shared node behavior

All nodes display type, title, concise result preview, version, status, and connection handles. Generation nodes additionally show their concise prompt input and supported settings directly beneath the preview.

Single click selects. Double click opens the inspector. The node context menu exposes only valid actions. Delete is undoable; destructive removal of persisted project assets requires an explicit confirmation and is never implied by removing a node reference.

### 4.2 Node types

- Text: title and short content preview; the modal edits full text.
- Image: rendered project asset or empty state; a direct input panel edits prompt, model, ratio, and image references.
- Video: playable project asset or empty state; a direct input panel edits prompt, model, ratio, resolution, duration, and image/video references supported by the chosen model.
- Audio: duration/preview or empty state; a direct input panel edits script, voice, and speed.
- Group: visual boundary with title, color, and members; it does not own or duplicate its member nodes.

## 5. Canvas Operations

### 5.1 Pointer controls

- Middle mouse drag always pans the viewport.
- Space plus left mouse drag pans the viewport.
- Wheel zooms around the pointer.
- Left drag on blank space makes a selection box.
- Ctrl/Cmd click adds or removes nodes from selection.
- Shift while dragging a node constrains movement to its stronger horizontal or vertical direction.
- Double click blank canvas creates a node at that canvas coordinate.

### 5.2 Keyboard controls

The shortcut dialog lists only commands that are implemented and enabled in the current selection state.

- Ctrl/Cmd+A select all nodes.
- Ctrl/Cmd+C and Ctrl/Cmd+V copy and paste selected nodes and internal edges.
- Ctrl/Cmd+D duplicates selected nodes.
- Delete/Backspace removes selected nodes or edges.
- Ctrl/Cmd+Z undo; Ctrl/Cmd+Shift+Z or Ctrl/Cmd+Y redo.
- Ctrl/Cmd+G group; Ctrl/Cmd+Shift+G ungroup.
- Shift+1 fits the canvas; Shift+2 focuses the selection.
- Ctrl/Cmd+plus/minus zooms in/out.

Typing in an input, textarea, select, or content-editable region takes priority over canvas shortcuts.

### 5.3 Context menus

Blank canvas: create text/image/video/audio, add from project assets, upload to project assets, paste, undo, redo.

Node: open inspector, copy, duplicate, delete node reference, disconnect edges, replace media asset, and reveal its project asset.

Multi-selection: copy, duplicate, delete, group/ungroup, align, distribute, and fit selection.

## 6. References, Edges, and Groups

Edges encode a relationship rather than merely a decorative line. An edge records its relation, for example `reference`, `first_frame`, `last_frame`, `audio_source`, or `sequence`.

Dragging a right-side connection handle into blank canvas opens a two-choice menu at the drop point: `文本` or `图片`. The selected node is created at that point and linked to the source. Connecting an existing image or video node to a generation node also synchronizes the eligible upstream asset into that target node's direct reference panel. Reference cards deduplicate by asset ID and show the source asset thumbnail. Unsupported reference types remain connected as canvas structure but are not presented as a valid provider reference.

The video task builder accepts only validated upstream project assets. It stores a snapshot of reference asset IDs and node version at submission time, so a later user edit cannot silently rewrite an active task.

Normal edges are light. Only an executing or result-returning relation uses limited motion; reduced-motion preference disables it.

## 7. Asset and Task Flow

1. User creates or opens a media node.
2. User chooses project assets or uploads source files. The Go API verifies project ownership and media type, writes binary data to TOS, and records asset metadata in MySQL.
3. User edits node configuration in the inspector and starts a supported task.
4. Go validates the current node, its version, model configuration, and references. Redis coordinates execution and reports progress.
5. The executor or provider returns a result. Go validates media metadata and persists the result as a project asset before it changes the canvas.
6. The task reaches `awaiting_confirmation`. The user confirms to apply the asset to the same node version, retries, cancels, or resolves a failure.
7. Confirming applies the asset reference in one transactional update. If the node changed meanwhile, the task cannot overwrite it.

Existing video jobs already establish part of this contract. Image and audio tasks must use the same state model before their UI claims they are executable.

## 8. Persistence and Safety

- MySQL stores projects, canvas revision, nodes, edges, groups, node versions, task records, and asset metadata.
- Redis is only used for queue coordination, progress, and locks.
- TOS stores media bytes.
- Canvas save uses its existing revision conflict contract. A stale browser must reload instead of overwriting another canvas state.
- Existing applied database migrations are immutable. Any new schema is an append-only, idempotent migration with tests.
- Removing a canvas node removes only the reference and edges. Deleting a project asset is a separate confirmed operation with recovery policy.

## 9. Delivery Phases

### Phase 1: Canvas Interaction Foundation

Compact nodes, middle-mouse pan, space pan, shortcut dialog, real keyboard commands, selection tools, fit/focus controls, and context-menu parity.

### Phase 2: Node Inspector and Project Assets

Unified inspector, clean image/video/audio nodes, project asset picker, upload flow, reference relation types, and recoverable node removal.

### Phase 3: Media Task Lifecycle

Consistent image/video/audio task state, queue progress, result asset ingestion, confirmation, retry, cancel, and recovery.

### Phase 4: Agent-Guided Production

The conversation can propose and perform scoped node operations. It must present generated prompts/results for review in supervised mode and execute the approved full workflow only in explicit automation mode.

## 10. Acceptance

- A real local browser user can pan with middle mouse, pan with space, zoom, select, box-select, copy, paste, duplicate, group, ungroup, align, delete, undo, redo, fit, and focus a selection.
- Shortcut help accurately reflects only working commands.
- Image, video, and audio generation nodes expose their prompt and supported settings in the direct panel below the preview; text keeps its compact preview.
- Releasing a connection in blank canvas exposes only `文本` and `图片`; connecting existing eligible media automatically updates the target reference panel.
- Reload restores canvas structure, group bounds, viewport, asset references, and safe task state.
- Task output becomes a project asset before it can become a node result.
- Local 4001 runs the rebuilt Go-embedded frontend and passes browser interaction acceptance. No public deployment is included.

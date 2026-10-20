# Agent Canvas Liblib-Style Workspace Design

Date: 2026-10-21

Status: approved design direction, implementation pending

Scope: local V88 Agent workspace and its Go APIs. This design learns from the interaction structure observed in the authenticated Liblib canvas supplied by the user. It does not copy Liblib source code, CSS, branding, private APIs, account behavior, billing, or copyrighted assets.

## 1. Objective

Turn `/agent/canvas` into a complete creative workspace rather than a node demo. The workspace must support two views over one project:

- **Workflow:** an infinite canvas for creating, connecting, grouping, generating, and processing text, image, video, and audio nodes.
- **Storyboard:** a director-oriented workspace that reuses workflow assets, generation history, and Agent conversation to organize and revise production material.

The final result should adopt the useful information hierarchy and interaction patterns of the observed Liblib canvas while retaining V88 colors, identity, Go runtime, configured model catalog, storage rules, and task infrastructure.

This is a full-product target. Delivery is phased so every phase remains usable, but completion cannot be claimed until every acceptance section in this document is verified.

## 2. Confirmed Constraints

- The maintained source remains Git branch `v88`.
- The application runtime remains Go-only. React and Ant Design are embedded into the Go binary.
- MySQL stores durable project, canvas, storyboard, task, and asset metadata.
- Redis handles queues, progress coordination, and locks.
- TOS stores media bytes.
- Model choices come from the existing API configuration and `/api/models`; canvas code must not invent a second catalog.
- Provider/channel and model settings are stored independently for each node.
- System prompts remain in the Go prompt module, not hard-coded in React.
- Existing user data and old canvas revisions must remain readable through additive normalization.
- V88 colors and branding remain. Reference-product colors, icons, CSS, copy, and logos are not copied.

## 3. Observed Reference Behaviors

The following were verified by inspecting the user's authenticated Liblib project:

- A top switch between `工作流` and `故事板`.
- A left global tool rail for add node, move, material library, character tools, generation history, shortcuts, tutorial, and director assistance.
- A resizable node drawer with canvas/assets tabs, search, type filter, rating filter, display settings, group tree, node location, and per-item actions.
- Canvas-level organize, minimap, edge visibility, grid snapping, zoom, grouped frames, and visible connections.
- Add-node choices for text, image, video, smart editing, director, frame analysis, audio, script, material library, upload, and generation-history selection.
- Image nodes with preview, replace material, reference/marker/style actions, prompt, model picker, compact parameter summary, presets, advanced settings, AutoLink, and contextual image tools.
- Video nodes with prompt, model picker, generation type, compact parameters, advanced switches, and reference-aware generation modes.
- Audio asset tools for trim, speed, smart split, custom split, and download.
- Rich model pickers that show model name, expected duration, and a short capability description.
- Generation history filters for all/current canvas, image/video/audio, rating, time order, and batch actions.
- Storyboard layout with Director conversation, categorized project materials, selected-material preview, and a node-style generation editor.

These observations define interaction goals only. They do not prove Liblib server behavior and are not used as an API contract.

## 4. Current V88 Baseline and Gaps

V88 already has useful foundations:

- React Flow canvas, nodes, edges, grouping, alignment, distribution, zoom, history, keyboard commands, middle-button pan, and save/reload.
- Text and image generation routes plus video job submission, polling, confirmation, retry, and cancellation.
- Project assets and reference synchronization from edges.
- Agent project conversation and execution records.
- Typed text/image/video/audio model catalog.
- Provider-model reconciliation helpers for personal API, Doubao local executor, AutoDL ComfyUI, and YFAI/Seedance.

The major gaps are:

- No workflow/storyboard top-level mode contract.
- No Liblib-style node/asset drawer, search, filter, group tree, minimap, grid snap, or automatic layout.
- Add-node behavior is too small and not organized as a searchable node library.
- Image configuration only persists ratio; quality, resolution, output count, provider, preset, and advanced settings are absent.
- Video parameters are separate compact selects rather than one readable summary and lack output count and audio-generation controls.
- Text nodes lack generation mode, output length, creativity, and a coherent settings surface.
- Audio nodes and audio-generation task lifecycle are not first-class Agent canvas features.
- Model selection is a plain select without provider grouping, capability explanation, expected duration, or unsupported-option filtering.
- Generation history and project material management are incomplete at workspace level.
- The current conversation panel is useful but is not integrated into storyboard material selection or supervised/automatic execution modes.

## 5. Information Architecture

### 5.1 Shared project header

The Agent project header contains:

- Back/project switcher.
- Editable project name.
- `工作流` and `故事板` tabs.
- Save state and unsaved-change indicator.
- Node/material count appropriate to the active view.
- Publish/share entry when that capability exists.
- Collapse controls for optional side panels.

Switching views does not navigate to another project or duplicate state. The URL keeps the same `taskId` and adds a stable view value, for example `view=workflow` or `view=storyboard`.

### 5.2 Workflow layout

The workflow view has four layers:

1. A narrow global tool rail on the left.
2. A resizable drawer for canvas nodes or project assets.
3. A full-bleed infinite canvas.
4. A collapsible Agent conversation panel on the right.

The drawer may be hidden to maximize the canvas. Hiding it does not remove filters or selected state. The conversation panel never overlays selected nodes; when opened, the canvas viewport resizes.

### 5.3 Storyboard layout

The storyboard view has three resizable columns:

1. Director Agent conversation and task proposals.
2. Categorized project materials.
3. Selected material preview and its type-specific editor.

Material categories are text, audio, image, and video. Supported subcategories include script/other text, character/scene/object/other image, clip/final video, and voice/music/effect audio.

Storyboard uses the same project assets and node configurations as workflow. Editing a material updates the owning node through the same revision rules; it does not create a hidden duplicate.

## 6. Workflow Canvas Interaction

### 6.1 Pointer and viewport controls

- Middle-button drag always pans.
- Space plus left-button drag pans.
- Wheel zooms around the pointer.
- Left drag on empty space box-selects.
- Ctrl/Cmd click toggles selection membership.
- Dragging selected nodes moves them together.
- Grid snapping is optional and visible in the canvas controls.
- Minimap, fit canvas, focus selection, zoom percentage, and automatic layout are available.
- Reduced-motion preference disables animated edge movement.

### 6.2 Node handles and connections

- Incoming handles sit on the left edge.
- The outgoing `+` handle is anchored to the right edge and never floats independently.
- Dragging a `+` handle creates a live connection preview.
- Dropping on an existing node creates a typed edge when valid.
- Dropping on empty canvas opens a small creation menu at the release point. The quick menu prioritizes text and image, while `更多节点` opens the full library.
- Connecting an eligible image or video node automatically adds that asset to the target node's reference row.
- Disconnecting removes only the edge-derived reference. A manually selected reference remains.
- References deduplicate by durable asset ID.

Edges store relation type, not only endpoints. Initial relation types are `reference`, `first_frame`, `last_frame`, `audio_source`, `script_source`, and `sequence`.

### 6.3 Selection operations

The selection toolbar and context menus expose only valid actions:

- Copy, paste, create duplicate, delete, undo, and redo.
- Group, ungroup, rename group, and change group color.
- Align left/center/right/top/middle/bottom.
- Distribute horizontally or vertically.
- Lock/unlock position.
- Fit or focus selection.
- Disconnect edges.
- Reveal the backing project asset.
- Save eligible result to reusable assets.
- Rate a result.

Removing a node removes its canvas reference and edges. Deleting a durable project asset is a separate confirmed operation.

### 6.4 Node drawer

The canvas tab provides:

- Search by node title and content summary.
- Filter by node type, status, and rating.
- Expandable group tree.
- Ungrouped node list.
- Node and group counts.
- Locate/focus action.
- More-actions menu consistent with canvas context actions.

The assets tab provides:

- Search, filter, category tree, rating, and batch selection.
- Create category, upload, add from generation history, and asset management.
- Drag an asset onto the canvas to create the corresponding media node.

## 7. Node Library

The node library is searchable and divided into stable sections.

### 7.1 Core nodes

- Text
- Image
- Video
- Audio
- Script
- Group

### 7.2 Processing nodes

- Smart edit
- Frame analysis
- Image tools
- Video tools
- Audio split/trim

Only tools with a real Go contract are enabled. Planned tools may be visible as unavailable with a clear explanation, but must not look executable.

### 7.3 Resources

- Upload
- Choose from project assets
- Choose from generation history
- Choose a preset

Creating a node places it at the requested canvas coordinate and immediately persists it through the existing revision flow.

## 8. Unified Node Editor

### 8.1 Open and close behavior

- A media node normally shows only its title, result preview, status, and handles.
- Clicking the preview opens the editor directly below that preview at the same width.
- Clicking another node transfers focus to that node.
- Clicking empty canvas closes the editor.
- Prompt text, references, model, settings, and unfinished input remain saved when the editor closes.
- A pending generated result may remain visible until confirmed or discarded.

### 8.2 Shared editor structure

From top to bottom:

1. Reference/media row.
2. Prompt or source-content input.
3. Bottom toolbar with model, mode/type, parameter summary, preset/tools, settings, cost/status, and generate button.
4. Optional advanced-settings area.

The bottom controls use compact icon or icon-plus-label buttons. They are not a row of full-width Ant Design selects.

### 8.3 Image editor

The image toolbar contains:

- Image model picker.
- Compact parameter summary defaults to exactly `16:9 · 标准画质 · 2K · 1张`; changing any parameter updates this same line immediately.
- Preset entry.
- Settings entry.
- Generate action.

The parameter popover supports:

- Ratio: model-supported values, initially including `16:9`, `9:16`, and `1:1`.
- Quality: low, standard, and high when supported.
- Resolution: `2K` and `4K` when supported.
- Output count: `1张`, `2张`, or `4张` when supported.

The settings panel contains the image generation channel/provider, model-specific options, AutoLink, seed when supported, and negative prompt when supported.

Contextual image actions may include HD/upscale, panorama, multi-angle, lighting, grid, element editing, layer separation, and grid slicing. Each action is enabled only after a real Go route and test exist.

### 8.4 Video editor

The video toolbar contains:

- Video model picker.
- Generation type such as text-to-video, image-to-video, first-frame, first/last-frame, or reference-video, limited by model capability.
- Compact summary, for example `16:9 · 720P · 5s · 1个`.
- Settings entry.
- Generate action.

The parameter popover supports model-allowed values for:

- Ratio.
- Resolution.
- Duration.
- Generate audio on/off.
- Output count `1个`, `2个`, or `4个`.

The settings panel contains video engine/provider, model-specific controls, material validation, AutoLink, camera movement, and other supported options.

### 8.5 Text editor

The text toolbar contains:

- Text model picker.
- Operation: generate, rewrite, or continue.
- Parameter summary such as `生成 · 标准长度 · 均衡`.
- Settings entry.
- Generate action.

Text settings include output length, creativity, optional structure/preset, and relevant context references. Text nodes do not show image channel or video engine controls.

### 8.6 Audio editor

The audio toolbar contains:

- Audio model/voice picker.
- Operation type such as speech, music, or sound effect when supported.
- Voice, speed, pitch, emotion/style, and output format settings.
- Generate action.

Existing audio assets expose trim, speed, smart split, custom split, replace, and download operations. Processing must produce a new asset version and never overwrite the source bytes silently.

## 9. Provider, Model, and Capability Contract

### 9.1 Provider filtering

Per-node provider options reuse `providerModelBinding.js`:

- `personal_api`
- `doubao_local_executor`
- `autodl_comfyui`
- `yfai_seedance`

Changing provider reconciles the selected model to the first valid model for that provider and node kind. A stale or disabled model is shown as unavailable, not silently replaced during generation.

### 9.2 Rich model picker

The picker displays:

- Model name and provider.
- Model kind.
- Short capability description.
- Expected duration when known.
- Supported reference types and generation modes.
- Availability/configuration status.

The picker is searchable and provider-filtered. Text, image, video, and audio nodes can only select models of their own kind.

### 9.3 Capability metadata

The Go model catalog must expose normalized public capability metadata without exposing credentials. Depending on kind, it can include:

- Ratios and resolutions.
- Quality levels.
- Minimum/maximum/default duration.
- Maximum output count.
- Reference image/video limits.
- Supported generation modes.
- Audio-generation support.
- Advanced option flags.

The frontend derives controls from these capabilities. It never allows a value that the selected model rejects.

## 10. Generation and Task Lifecycle

All node generation uses one visible lifecycle:

1. Draft configuration saved.
2. Client and Go validation.
3. Task submitted with node ID, node version, idempotency key, provider/model, settings, and reference asset IDs.
4. Queued/running progress shown on the node and in generation history.
5. Provider result validated and persisted as a project asset.
6. Result becomes `awaiting_confirmation` in supervised mode or is applied automatically in explicit automation mode.
7. User confirms, retries, cancels, rejects, or requests a revision.
8. Applied result updates only the submitted node version; version conflicts require review.

Image, text, video, and audio must converge on this lifecycle. Existing image/text immediate responses may be wrapped in the same visible state contract even if the provider returns synchronously.

Multiple outputs are stored as a result set. Choosing one applies it to the node; the others remain in generation history until deleted by an explicit asset action.

## 11. Generation History

Generation history supports:

- Scope: all projects or current project.
- Kind: image, video, audio, and text where useful.
- Status and rating.
- Time order.
- Search.
- Batch selection.
- Add to current canvas.
- Save to reusable assets.
- Retry supported failures.
- Download eligible results.
- Explicit delete with confirmation.

Each history item links back to its originating node, model, provider, prompt snapshot, settings snapshot, and task status.

## 12. Storyboard and Director Agent

### 12.1 Project materials

The material column reads durable project assets and text nodes. It supports category tabs, counts, search, collapse/expand, adjustable section heights, and batch selection.

Selecting a material opens its preview/editor in the right column. `添加到对话` attaches a durable reference, not raw private URLs.

### 12.2 Director conversation

The Director Agent can:

- Inspect selected nodes and project material metadata.
- Propose a workflow or storyboard arrangement.
- Create nodes and connections.
- Rewrite prompts or text.
- Start permitted generation tasks.
- Apply approved results.
- Remove nodes or edges.

It cannot silently delete durable assets. Ambiguous requests are answered with one focused question.

### 12.3 Supervised and automatic modes

- **Supervised mode:** default. Prompts, major structure changes, and generated results are shown for approval.
- **Automatic mode:** enabled only by an explicit user request such as `全自动化`. The Agent may continue through approved stages, but still stops for missing configuration, unsafe deletion, cost/limit errors, or contradictory requirements.

The current mode is visible near the conversation composer and is recorded in execution history.

## 13. Persistence Model

The persisted canvas snapshot remains backward compatible and gains normalized fields:

- `viewState`: active view, drawer states, panel sizes, canvas viewport, and display settings.
- `nodes[].data.config`: kind-specific provider, model, mode, parameters, advanced settings, manual references, and edge-derived references.
- `edges[].data.relation`: typed relation and reference ownership.
- `groups`: title, color, members, collapsed state, and bounds.
- `storyboard`: material categories, ordering, selected material, and Director layout state.
- `generationResultSets`: output IDs and selected output.

Old nodes receive in-memory defaults and are upgraded only on the next normal save. No historical migration row is edited.

## 14. Error Handling

- Missing model: disable generation and link the user to API configuration.
- Provider unavailable: keep the node configuration and report the provider-specific reason.
- Unsupported setting: reconcile visibly and require confirmation before generation if an old saved value is no longer supported.
- Reference type mismatch: keep the canvas edge but mark it ineligible for provider input.
- Version conflict: never overwrite newer node edits.
- Failed upload or asset persistence: do not expose an unpersisted provider URL as a completed node result.
- Queue interruption: preserve the task and offer retry/resume where the provider contract permits it.
- Storyboard or drawer load failure: keep the workflow canvas usable and show a contained retry state.

## 15. Delivery Program

### Phase 1: Workspace shell and navigation

- Workflow/storyboard view state.
- Shared project header.
- Left tool rail.
- Resizable node/assets drawer.
- Search, filter, group tree, minimap, grid snap, automatic layout, and shortcut help.

### Phase 2: Unified node editors and model capabilities

- Click-to-expand editor behavior.
- Rich provider/model picker.
- Image/video/text/audio parameter summaries and settings.
- Per-node persistence and backward normalization.

### Phase 3: Complete generation lifecycle

- Shared visible task state.
- Capability validation.
- Multiple outputs.
- Audio generation and processing.
- Generation history and asset actions.

### Phase 4: Storyboard and Director

- Three-column storyboard.
- Categorized project materials.
- Selected-material preview/editor.
- Supervised/automatic Agent execution.

### Phase 5: Advanced creative tools and release verification

- Enabled image/video/audio processing tools with real Go routes.
- Presets and reusable assets.
- Publish/share integration where supported.
- Full local Go-embedded acceptance and regression audit.

Each phase is implemented and verified in order. Later phases may add contracts but must not break earlier persisted projects.

## 16. Testing Strategy

### 16.1 Pure and component tests

- Node config normalization for old and new snapshots.
- Provider/model reconciliation and capability filtering.
- Parameter summary formatting.
- Edge-derived reference synchronization and manual-reference preservation.
- View switching and resizable panel persistence.
- Search/filter/group tree behavior.
- Storyboard material categorization.
- Agent action permission and supervised/automatic mode decisions.

### 16.2 Go API tests

- Public capability serialization without credentials.
- Kind/provider/model validation.
- Node-version and idempotency enforcement.
- Multiple result persistence.
- Audio task and processing routes.
- Asset ownership and typed reference validation.
- Revision conflict behavior.

### 16.3 Browser acceptance

Use an authenticated local `:4001` session and verify real interactions:

- Switch between workflow and storyboard without losing data.
- Add, locate, search, filter, connect, disconnect, group, arrange, save, reload, and restore nodes.
- Middle-button pan, shortcuts, minimap, grid snap, fit, focus, and auto-layout.
- Click a media preview to open the editor and click blank canvas to close it without losing input.
- Select provider and model per node and see only compatible settings.
- Generate text, image, video, and audio through configured test-capable providers.
- Restore connected references and selected results after reload.
- Use generation history and storyboard material selection.
- Ask Agent to propose changes, approve them in supervised mode, and run an explicitly automatic workflow.

Build output, HTTP 200, source markers, or focused tests alone are not acceptance.

## 17. Completion Criteria

The program is complete only when all of the following are true in the rebuilt Go-embedded local runtime:

- Workflow and storyboard are both usable views of the same durable project.
- The workspace navigation, drawer, canvas controls, node library, generation history, and project material flow are present and functional.
- Text, image, video, and audio nodes use the unified editor pattern and per-node provider/model/settings.
- Image summary defaults to `16:9 · 标准画质 · 2K · 1张`, updates from the current node values, and those same values drive the real request.
- Video summary truthfully represents ratio, resolution, duration, count, and audio setting.
- Connections synchronize eligible references without deleting manual references.
- Model controls are capability-driven from Go and API configuration.
- Every advertised action has a real implementation or is visibly unavailable.
- Agent supervised and explicit automatic modes behave as described.
- Save/reload restores canvas, panels, groups, references, storyboard state, and task results.
- Focused tests, frontend build, Go tests, embedded build, and authenticated browser acceptance pass.
- No public deployment is implied by local completion.

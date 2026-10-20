# Agent Canvas Collapsible Node Editor Design

Date: 2026-10-20

Scope: local V88 `/agent/canvas`. This design uses observed interaction ideas from the public Liblib read-only workflow, but does not copy Liblib code, branding, private endpoints, project data, or account behavior.

## 1. Goal

Repair the canvas nodes that currently show detached connection `+` controls and permanently expanded generation forms. Media nodes should be quiet preview cards by default. A user opens one node's editing controls by clicking its preview, edits without leaving the canvas, and closes the controls by clicking elsewhere.

This phase also adds a persisted view preference for showing or hiding existing canvas edges. It does not add image or audio execution APIs, a minimap, a storyboard, or a new project-material system.

## 2. Confirmed Problems

Runtime measurement and source inspection confirmed four separate defects:

- Image and video nodes have a transparent outer node about 620 rendered pixels wide while the visible preview is about 396 rendered pixels wide at the inspected zoom.
- Both React Flow handles are positioned against that transparent outer node. Their vertical midpoint falls in the gap between the preview and the permanently visible generation panel.
- Every node always renders a visible source `+`, even when it is neither hovered, selected, nor connecting.
- Existing `copy` nodes do not match the text-node CSS or rendering branches, so they can appear as a short line of text with a detached connection control.

Moving the current handles by a few more pixels would hide one symptom but would not repair these structural problems.

## 3. Reference Behavior

The inspected Liblib read-only workflow establishes the interaction model used here:

- A media node normally presents its preview without a permanently expanded prompt form.
- Clicking the preview selects the node and expands its editing controls below it.
- Clicking the canvas background clears selection and removes the editing controls.
- Only one media-node editor is open at a time.
- The canvas provides independent controls for showing a minimap, showing or hiding existing edges, and choosing a zoom level.

Only the collapsible editor and edge-visibility control are included in this phase.

## 4. Node Structure

### 4.1 Preview shell

Every node has a visible preview shell that owns its connection geometry.

- Image: image asset or empty image state.
- Video: video asset or empty video state.
- Audio: compact audio preview or empty audio state.
- Text: title and concise text preview.

The React Flow target and source handles live inside the preview shell, not on the outer wrapper. Expanding an editor therefore changes the node's total height without moving either connection endpoint away from the preview frame.

### 4.2 Collapsible editor

Image, video, and audio nodes render their existing prompt and setting controls only when their node ID equals the active editor node ID.

- Clicking a media preview opens that node's editor.
- Clicking another media preview switches directly to that node's editor.
- Clicking the canvas background, selecting a text node, or pressing `Escape` closes the active editor.
- Clicking or typing inside the editor does not close it and does not start node dragging or canvas panning.
- Re-clicking the already active preview keeps it open; explicit outside click or `Escape` closes it.
- A single click on an image preview no longer opens the asset picker. Adding or replacing an image uses the explicit asset button inside the expanded editor.
- Media-node double-click does not open the existing competing manual-edit modal. Text nodes retain their existing double-click text-edit behavior.
- Native controls inside an existing video or audio preview remain usable and do not toggle the editor unexpectedly.
- The active editor ID is temporary interface state. It is not written to MySQL or the saved canvas JSON.

The existing `onNodeConfigChange` path continues to save edits as the user changes them. Collapsing the panel never discards prompt, model, ratio, duration, voice, speed, asset, or reference values. Opening or closing the editor itself does not create a canvas history entry or trigger a save.

### 4.3 Canonical node types

The presentation boundary normalizes legacy `copy` nodes to the canonical `text` type. Unsupported or blank node types also fall back to `text` instead of producing a transparent, unstyled node.

Normalization happens when persisted nodes enter the React Flow presentation model, and `toCanvas` serializes the normalized value. The next normal content-changing canvas save therefore writes the canonical type; merely opening and closing an editor does not force a cleanup save. This remains inside the existing canvas JSON contract and requires no database migration.

## 5. Connection Controls

### 5.1 Source handle

- The right-side source handle is centered over the visible preview shell border.
- It is hidden and non-interactive by default.
- It becomes visible when the preview is hovered, the node is selected, or a keyboard focus is inside the preview.
- Hidden means visually transparent and pointer-inactive, not removed from the document. Keyboard focus and React Flow handle geometry therefore remain stable.
- Its icon and accessible label communicate `拖出连接`; it is not presented as a generic add button.

### 5.2 Target handle

- The left-side target handle is centered over the visible preview shell border.
- It stays hidden and non-interactive during normal viewing.
- It becomes visible and interactive while a connection drag is active.
- Valid targets use the accent highlight; invalid or duplicate targets do not advertise acceptance.

### 5.3 Existing behaviors retained

- Dropping on a valid existing node creates one edge and never opens the empty-drop menu.
- Dropping on empty canvas opens exactly `文本` and `图片`.
- Connected eligible image/video assets continue to synchronize into the target's reference list.
- The `+` inside the reference section remains an asset action, but uses the existing image-add icon and label `添加参考素材` so it cannot be confused with a connection handle.

## 6. Edge Visibility

Add one view tool with two states: `隐藏连线` and `显示连线`.

- Default is visible, preserving all existing canvases.
- Hidden mode hides persisted edges without deleting or filtering edge data.
- Users can still select nodes, edit nodes, and begin a new connection while existing edges are hidden.
- The temporary connection line remains visible during a drag.
- The preference is stored as `appearance.edgesVisible` in the existing canvas appearance JSON, with normalization defaulting missing values to `true`.
- Saving this preference uses the existing appearance/canvas save path and adds no API or schema.
- Rendering uses the React Flow edge visibility state; it never replaces the in-memory edge array with an empty array. Serialization, undo/redo, and saved edge counts therefore remain unchanged.

## 7. State and Event Flow

1. `AgentCanvas` owns `activeEditorNodeID` and connection-drag state.
2. A media preview click selects the node and sets `activeEditorNodeID`; explicit asset and native media controls stop propagation and perform only their named action.
3. `CanvasNode` renders its editor only when `isEditorOpen` is true.
4. Editor controls stop mouse and double-click propagation while keeping normal input behavior.
5. Pane click or `Escape` clears `activeEditorNodeID`.
6. Connection handles remain positioned inside the preview shell, so editor height changes do not change their intended visual anchor.
7. React Flow's node-internals refresh is called after a node opens or closes so it recalculates node dimensions and edge paths after the DOM has updated.
8. Edge visibility changes update normalized appearance and the existing rendered edges without mutating the edge collection.

## 8. Error and Safety Behavior

- A missing preview asset still produces a visible empty-state shell with correctly anchored handles.
- Unknown legacy node types render as text rather than disappearing.
- Opening or closing an editor never submits a generation task.
- Image and audio submit controls remain visibly unavailable until Agent-specific task APIs exist.
- Video keeps its existing Agent video-job path; browser acceptance must not submit a paid/provider task.
- Removing a node, edge, or project asset is outside this change.
- No public deployment is included.

## 9. Verification

Automated coverage must prove:

- Legacy `copy`, blank, and unknown types normalize to `text`.
- One active media editor renders; changing the active ID closes the previous editor.
- Pane click and `Escape` clear the editor state while editor input clicks do not.
- Image asset actions and native media controls do not accidentally toggle the editor, and media double-click does not open the legacy modal.
- Handles are descendants of the preview shell and use conditional visibility classes.
- Missing `appearance.edgesVisible` normalizes to `true`, while explicit `false` is retained.
- Hiding edges does not remove them from serialized canvas data.
- Successful connections and empty-canvas drops retain their existing deterministic behavior.

Real local browser acceptance on the rebuilt Go-embedded port 4001 must verify:

- The initial canvas contains no floating connection `+` controls.
- Hovering or selecting a preview reveals one source connection control on its border.
- Starting a drag reveals target controls on other preview borders.
- Clicking a media preview opens its editor below; clicking blank canvas and pressing `Escape` each close it.
- Adding or replacing an image works from the expanded editor's explicit asset action without collapsing the editor.
- Switching between media nodes leaves only one editor open and retains entered values.
- Existing connections can be hidden and restored without changing the saved edge count.
- A real node-to-node drag creates one edge without opening the empty-drop menu; undo restores the original edge count.
- Tests cover normal and zoomed views, and the page reports no new browser errors.

## 10. Deferred Liblib-Inspired Work

The following observed features remain separate phases:

- Optional minimap.
- Numeric zoom menu and preset zoom actions.
- Workflow/storyboard view switching.
- Searchable project-material panels grouped by audio, image, and video.
- Image categories such as character, scene, prop, and other.
- Video categories such as clip and completed production.
- Resizable/collapsible asset sections with duration, dimensions, and reference counts.

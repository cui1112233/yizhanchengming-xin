# Agent Workflow Liblib Parity Design

## Goal

Bring the V88 Agent workflow canvas to the same interaction density and visual hierarchy as the observed Liblib workflow while preserving V88 node data, generation behavior, project persistence, and Storyboard behavior.

## Scope

- Keep `@xyflow/react`, the existing canvas schema, node kinds, generation entry points, and save queue.
- Restyle workflow nodes, handles, edges, selection, hover, and connection feedback.
- Improve initial viewport recovery only when the saved viewport leaves every node outside the usable canvas.
- Keep the current left tool rail, bottom control bar, project conversation, minimap, grouping, alignment, and auto-layout commands.
- Do not implement Liblib-only publishing, credits, membership, cloud assets, or private APIs.
- Do not start paid generation during acceptance.

## Observed Differences

The local workflow currently renders mixed node widths from 176px to 340px, 14px hidden input handles, 24px coral plus output handles, and 2px coral animated arrow edges. A saved viewport can place all useful nodes outside the first visible frame. Liblib uses quieter media-first nodes, neutral connection lines, edge-mounted connection affordances, restrained selection styling, and a dense left/bottom control layout.

## Design

### Nodes

All regular nodes use a common shell with a compact 32px header. The header contains a Lucide type icon, a single-line title, an optional source/status indicator, and no decorative card nesting. Text nodes remain compact; image and video nodes use a stable 16:9 preview; audio nodes use a shorter media row. Director and pinned Batch Factory nodes keep their functional bodies but adopt the same header, border, and selection language.

Node selection uses one clear accent outline. Hover changes only the border and handle visibility. Status colors remain reserved for real queued, running, failed, and completed states.

### Handles And Connections

The target and source handles stay on the left and right midpoint of the actual node shell. Their visible dot is 8px, but their hit area remains at least 24px. Handles are faint on hover/selection and fully visible while connecting. The source handle may show a small plus only while it is actionable.

Default edges become 1.5px neutral curves without arrowheads or continuous animation. Hovered and selected edges use the canvas accent. Task activity may retain an animated overlay, but static saved references remain quiet. Connection preview and valid targets use the same visual language.

### Viewport And Layout

Existing saved viewport state remains authoritative when at least one node intersects the usable canvas. If nodes exist but none intersect after initialization, the canvas performs one local `fitView` recovery with restrained padding. This recovery does not alter node positions and does not enqueue a canvas save merely because the view was recovered.

The top project bar, left tool rail, and bottom centered controls remain fixed-size and do not overlap nodes. The minimap remains optional. Existing align, distribute, group, snap, and auto-layout commands remain unchanged.

## Error And Compatibility Handling

- Old appearance values continue to normalize safely.
- Existing saved coral/cyan/violet/gold preferences remain selectable, but the default becomes neutral and non-animated.
- Old canvases with missing viewport data still use React Flow `fitView`.
- Empty canvases retain the current conversation starter instead of trying to fit nonexistent nodes.
- Connection creation, empty-drop node creation, keyboard commands, and deletion behavior are not changed.

## Verification

- Unit-test viewport intersection and recovery decisions with literal node rectangles.
- Unit-test neutral default edge appearance and legacy appearance normalization.
- Source/layout tests verify arrow removal, stable node dimensions, 24px handle hit areas, and the one-time recovery hook.
- Rebuild the React frontend and Go embedded binary.
- In the authenticated local browser, verify node visibility on entry, node drag, connection drag, edge selection, zoom, Storyboard round trip, refresh persistence, console errors, and no paid generation.

## Acceptance Criteria

- Opening Workflow never presents an apparently empty canvas when nodes exist.
- Nodes read as one consistent system across text, image, video, audio, director, and script types.
- Handles are discoverable without dominating the node; the clickable area is not reduced.
- Static edges have no arrowhead and no permanent animation.
- Existing workflow operations and Storyboard remain functional.


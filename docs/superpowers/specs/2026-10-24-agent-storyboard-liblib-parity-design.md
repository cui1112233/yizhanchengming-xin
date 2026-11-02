# Agent Storyboard Liblib Parity Design

## Goal

Rebuild the V88 Storyboard workspace around Liblib's observed material-workbench layout while preserving V88's existing upload, category, batch, generation, detail, workflow, and add-to-conversation behavior.

## Observed Reference

- Desktop viewport: 1890 x 1070.
- Director panel: about 31% of the available workspace width.
- Material panel: about 69%.
- Material columns: about 25% text/audio and 75% image/video.
- Three independent resize controls: director/material, material columns, and one vertical split inside each material column.
- Every splitter uses the matching resize cursor and restores its default on double click.
- Text, audio, image, and video sections scroll independently.
- Section headers collapse their section and expose drag affordance for section placement.
- Text and audio use dense single-line rows; image and video use dense media grids.
- The material toolbar is a compact single row instead of a tall stacked header.

## V88 Layout

Desktop Storyboard uses a two-column workspace with a 31/69 default split. The director remains a real project conversation, not a decorative sidebar. The divider between director and materials is draggable and keyboard-accessible.

The material surface has two columns separated by an 8px drag target. The left column starts with Text above Audio at 66/34. The right column starts with Image above Video at 55/45. Each column owns its own horizontal divider so resizing Text does not move Image, and resizing Image does not move Text.

The four sections keep independent overflow containers. Empty shot data does not reserve a strip; loading, error, or real shots still render the shot strip. Selecting an item continues to open the existing overlay inspector.

On screens at or below 1000px, the director becomes the existing overlay panel and the material surface uses the full width. At or below 680px, splitters are hidden and sections stack vertically.

## Interactions

- Drag any splitter with mouse, pointer, touchpad, or touch input.
- Double-click a splitter to restore only that split to its default.
- The active splitter is visibly highlighted and text selection is disabled during drag.
- Section headers retain collapse behavior and gain a grab affordance for swapping section slots.
- Material cards retain selection, batch selection, same-kind reorder, add-to-conversation, drag-to-conversation, add-to-workflow, and the detail inspector.
- No provider generation is triggered during visual acceptance.

## Visual Density

- Material toolbar target height: about 48px when no error/status needs a second line.
- Text/audio row target height: 32px.
- Image grid target minimum width: 88-96px.
- Video grid target minimum width: 210-230px, with 16:9 previews.
- Surfaces remain flat and dark using the existing V88 tokens; no Liblib branding or assets are copied.
- The desktop pet is hidden only while Storyboard is active.

## State And Boundaries

Resize ratios and section slot order are stored in the existing normalized Storyboard view state. Values are clamped before rendering so old or malformed saved state cannot collapse a panel. Workflow layout and all backend APIs remain unchanged.

## Acceptance

- At 1890 x 1070, the measured director/material split is close to 31/69.
- The material split is close to 25/75.
- Left and right row dividers move independently and double-click reset works.
- Zero shots do not consume vertical space.
- Text/audio rows and image/video grids match the intended density without overlap.
- Upload, category filters, batch controls, selection, add-to-conversation, and inspector remain usable.
- Focused tests, frontend build, Go embed build, runtime identity check, and live browser inspection pass.


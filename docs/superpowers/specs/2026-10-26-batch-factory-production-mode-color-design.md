# Batch Factory Production Mode Color Design

## Goal

Make the active `视频生成 | 漫画生成` mode immediately visible in both existing-batch and new-batch unified configuration dialogs.

## Visual Design

- Keep the existing compact segmented-control size and title placement.
- Video mode uses a solid blue selected item with white text.
- Comic mode uses a solid coral-red selected item with white text.
- The inactive item keeps the existing neutral surface and readable text.
- Add a clear control border and a small selected-item shadow so the active state remains visible in light and dark themes.
- Keep focus-visible behavior supplied by Ant Design; color is an additional cue, not the only control state.

## Implementation Boundary

- `BatchFactoryProductionModeControl` adds an `is-video` or `is-comic` class derived from the already-normalized mode.
- Scoped CSS styles only `.batch-factory-production-mode-switch`; no global `Segmented` styles change.
- No production-mode data, save behavior, prompt selection, or layout changes.

## Verification

- A source test requires both state classes and both selected colors.
- The focused production-mode test passes.
- The frontend production build succeeds.
- In the local Go runtime, switching modes visibly changes the selected fill from blue to coral red in both unified-config entry points.

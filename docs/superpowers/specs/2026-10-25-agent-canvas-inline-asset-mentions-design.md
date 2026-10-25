# Agent Canvas Inline Asset Mentions Design

## Goal

Make Agent Canvas asset references match the approved Liblib-style interaction: selected assets remain visible as a thumbnail strip above the prompt, while each `@asset` reference is rendered as an atomic inline chip inside the prompt editor.

## Scope

- Apply to text, image, video, and audio node composers.
- Keep the existing `prompt` plain-text value and `config.referenceAssets` array as the canonical persisted data.
- Keep the existing asset listing, filtering, node update, generation, and backend API contracts.
- Reuse the proven content-editable mention behavior from `ScriptMentionEditor` without changing Script Page behavior.
- Do not add a rich-text dependency, database field, migration, or backend endpoint.
- Do not start a real generation task during acceptance.

## UI Structure

The prompt area has three layers:

1. A visible selected-asset strip above the prompt label. Every referenced image or video shows a thumbnail; audio and assets without previewable media show a type icon. Each item shows its name and has a remove button.
2. A content-editable prompt surface. Referenced assets appear as compact, non-editable inline chips containing the same thumbnail or type icon and asset name. Ordinary prompt text remains directly editable around the chips.
3. The existing candidate menu above the editor. Typing `@` opens real project assets filtered by node capability and the text typed after `@`.

The existing top action for manually adding reference material remains available. The visible strip is not hidden inside that action's popover.

## Data And Synchronization

The saved prompt remains ordinary text such as `让 @角色A 走进房间`. Inline chips are only a visual representation of matching references; no HTML is saved.

Selecting an asset from the candidate menu performs one logical update:

- Replace the active `@query` with `@asset-name` in the prompt.
- Add or retain the asset in `config.referenceAssets` through the existing callback.
- Render the asset in both the top strip and the inline prompt.
- Restore the caret immediately after the inserted chip.

Removal is bidirectional:

- Removing an inline chip removes the matching asset from the top strip and `referenceAssets`.
- Removing an asset from the top strip removes every matching inline mention from the prompt.
- Duplicate selection of the same asset does not create duplicate top-strip entries.
- Plain `@text` that does not match a selected asset remains ordinary text.

Asset identity prefers `id`, then `assetId`; display matching uses the normalized asset name only when an identifier is unavailable. Removal must avoid deleting similarly named ordinary words.

## Editing Behavior

- Backspace or Delete next to a chip removes the whole chip in one action.
- Arrow keys, mouse selection, multiline input, and IME composition remain usable.
- Pasted content is inserted as plain text; pasted HTML is never retained.
- Escape closes the asset menu. Up and Down change the active candidate. Enter selects it while the menu is open.
- Blur persists the plain-text prompt through the existing node update path.
- The editor enforces the existing 8,000-character text limit and 4,000-character media prompt limit against the plain-text value.

## Node Capability Rules

- Text nodes may reference images, videos, and audio.
- Image nodes may reference images.
- Video nodes may reference images and videos.
- Audio nodes may reference audio.

Unsupported assets never appear in the candidate menu and are not newly inserted into the node.

## Theme And Failure Handling

- All surfaces, chips, borders, names, icons, hover states, and focus states use existing Agent theme variables so light and dark themes remain readable.
- Thumbnail failures fall back to the asset type icon without breaking the chip or strip layout.
- Loading, empty, and error states remain in the existing candidate menu.
- If asset listing fails, existing prompt content and selected references remain editable and removable.
- The editor keeps a stable height and scroll behavior so chips cannot resize the outer composer unexpectedly.

## Verification

- Unit-test plain-text parsing, chip-to-asset matching, duplicate handling, and prompt cleanup after removal.
- Source/component-test the content-editable editor, candidate keyboard controls, top strip, and bidirectional removal callbacks.
- Run the existing Agent Canvas asset-input and Script mention tests to prevent regressions.
- Build the React frontend and Go embedded binary.
- Restart only `backend/qiantie-local.exe` through the approved local Go path.
- Verify `/health` is `200`, runtime is `go`, and runtime commit identifiers match the new commit.
- In the local Agent Canvas, verify image/video/audio fallbacks, `@` selection, top strip, inline chip editing, both removal directions, light/dark readability, and no real generation.

## Acceptance Criteria

- Selecting an asset visibly creates both a top-strip item and an inline chip.
- The editor continues to save and submit plain prompt text through existing contracts.
- Either removal entry point keeps the prompt and selected references synchronized.
- Chips are atomic during keyboard deletion and do not damage surrounding text.
- Text, image, video, and audio composers use one consistent interaction with their existing capability limits.
- No backend, database, or generation contract changes.

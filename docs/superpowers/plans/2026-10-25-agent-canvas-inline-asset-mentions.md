# Agent Canvas Inline Asset Mentions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add synchronized top thumbnails and inline asset chips to text, image, video, and audio node prompt editors without changing persisted prompt or backend contracts.

**Architecture:** Keep plain prompt text and `config.referenceAssets` as canonical state. Add focused pure helpers for matching and removal, build an Agent-specific content-editable prompt component from the established Script mention mechanics, and connect it to the current composer callbacks and candidate menu.

**Tech Stack:** React, Ant Design, Lucide React, CSS, Node test runner, Vite, Go embed.

## Global Constraints

- Work directly on `v88`; do not create a branch.
- Preserve every unrelated dirty or untracked file.
- Keep `prompt` as plain text and keep `config.referenceAssets` as the existing persisted reference list.
- Do not change backend APIs, database schema, model generation behavior, or Script Page behavior.
- Use existing Agent theme variables for both light and dark themes.
- Do not trigger real generation or paid media processing during acceptance.
- Commit scoped source changes before rebuilding the embedded Go runtime.

---

### Task 1: Define Mention Synchronization Rules

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-asset-input.js`
- Modify: `frontend/src/user/pages/agent-canvas-asset-input.test.js`

**Interfaces:**
- Produces: `assetReferenceKey(asset)`, `promptMentionRanges(value, references)`, and `removeAssetMention(value, asset)`.
- Consumes: plain prompt strings and existing asset objects with `id`, `assetId`, `name`, and `kind`.

- [ ] Add failing tests for matching selected assets to prompt mentions, preserving unmatched `@text`, duplicate references, similarly named assets, and removing all exact mentions for one selected asset.
- [ ] Run `node --test frontend/src/user/pages/agent-canvas-asset-input.test.js` and confirm the new assertions fail.
- [ ] Implement normalized identity, exact mention ranges, and whitespace-safe prompt removal without storing HTML.
- [ ] Run the focused helper test and confirm all assertions pass.
- [ ] Commit only the two helper files.

### Task 2: Build The Agent Rich Mention Editor

**Files:**
- Create: `frontend/src/user/components/AgentAssetMentionEditor.jsx`
- Create: `frontend/src/user/components/AgentAssetMentionEditor.source.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `value`, `references`, `placeholder`, `maxLength`, `disabled`, `ariaLabel`, `mentionOpen`, `onChange(value, caret)`, `onBlur()`, `onKeyDown(event)`, and `onRemoveReference(asset, nextValue)`.
- Produces: a content-editable prompt surface whose DOM chips always serialize back to plain text.

- [ ] Add source tests requiring `contentEditable`, non-editable mention chips, thumbnail/type fallback, IME guards, plain-text paste, caret restoration, and atomic Backspace/Delete removal.
- [ ] Run the new component test and confirm it fails because the component does not exist.
- [ ] Implement the editor by adapting the proven DOM reconciliation and selection logic from `ScriptMentionEditor`, while keeping Agent-specific asset objects and callbacks.
- [ ] Add scoped chip, focus, placeholder, scroll, light-theme, and dark-theme CSS using Agent variables.
- [ ] Run the new component test plus `frontend/src/user/components/ScriptMentionEditor.test.js`.
- [ ] Commit the new editor, its test, and only the related CSS hunk.

### Task 3: Add The Visible Reference Strip

**Files:**
- Modify: `frontend/src/user/components/AgentNodeComposerReferenceAssets.jsx`
- Create: `frontend/src/user/components/AgentNodeComposerReferenceAssets.source.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `references`, `disabled`, `onAdd(event)`, and `onRemove(asset, event)`.
- Produces: the existing compact action plus exported `AgentNodeComposerReferenceStrip` for always-visible selected assets.

- [ ] Add failing source assertions for the visible strip, image/video previews, audio/type fallback, asset name, accessible remove button, and add action compatibility.
- [ ] Run the new test and confirm it fails against the current popover-only implementation.
- [ ] Add the strip component without removing the existing compact reference action.
- [ ] Add stable thumbnail dimensions, single-line names, horizontal overflow, fallback, hover, and focus styles.
- [ ] Run the focused reference component test.
- [ ] Commit the reference component, test, and only the related CSS hunk.

### Task 4: Connect Both Views To The Node Composer

**Files:**
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/pages/agent-node-composer.source.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-asset-input.test.js`

**Interfaces:**
- Consumes: Tasks 1-3 and the existing `onMentionAsset`, `onRemoveReference`, `onConfigChange`, and `onListMentionAssets` callbacks.
- Produces: synchronized insertion and removal in text, image, video, and audio composers.

- [ ] Add failing assertions that the Ant Design textarea is replaced by `AgentAssetMentionEditor`, the visible strip is rendered, and both strip/chip removals update plain prompt text before removing the reference.
- [ ] Run the focused composer and asset-input tests and confirm the new assertions fail.
- [ ] Replace only the prompt textarea with the Agent editor; retain the current candidate menu, limits, labels, generation submission, and node capability filtering.
- [ ] Make candidate selection insert one chip and one deduplicated reference while restoring the caret.
- [ ] Make inline removal call the existing reference removal callback and persist the cleaned prompt.
- [ ] Make strip removal clean the exact prompt mention, persist it, and then call the existing reference removal callback.
- [ ] Run all focused Agent composer, asset-input, and Script mention tests.
- [ ] Commit the composer integration and focused tests.

### Task 5: Build And Local Runtime Acceptance

**Files:**
- Generated: `frontend/dist/**`
- Generated: `backend/web/dist/**`
- Generated: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: Tasks 1-4.
- Produces: verified Go-embedded local UI on `http://localhost:4001`.

- [ ] Run all focused tests and record unrelated historical failures separately.
- [ ] Run `npm.cmd --prefix frontend run build`.
- [ ] Commit all remaining scoped source changes before producing the runtime build.
- [ ] Build through the repository's approved Go embedded script and restart only the verified `backend/qiantie-local.exe` process.
- [ ] Verify `/health` returns `200`, `/api/runtime-build-info` reports `runtime: go`, and both runtime SHA fields match the scoped commit.
- [ ] In Agent Canvas, type `@`, select a real asset, verify the top strip and inline chip, test both removal directions, keyboard deletion, ordinary typing, and light/dark themes.
- [ ] Confirm no generation request was triggered and report any legacy test failures separately.

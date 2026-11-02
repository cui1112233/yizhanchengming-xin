# Batch Factory Agent Task 6 Runtime Repair

## Scope

- Local authenticated acceptance only: `http://localhost:4001`.
- Project: `agent-bf-dbba32e978f13f1909e3b928`.
- Batch: `batch_2f462bb6608219f96838d30435620386`.
- Canonical Batch Factory book: `book_4090786e93fcf091439f12369c9dd02b`.
- Legacy source book ID: `2091926811672429633`.

## Root Cause

The Batch Factory projection persisted its existing media version IDs in mapped
nodes, but their response was hydrated as Agent asset URLs. The image and video
version IDs do not exist in `agent_assets`, so the browser received `素材不存在`.
The canonical Batch Factory book ID and the legacy source book ID are distinct;
projection reads must use the canonical ID recorded by the node origin.

The remaining image was a valid primary local-executor record: it stores a
`storage_ref` instead of a provider URL. The previous bridge only redirected
to non-empty URLs, so it incorrectly reported that the existing local image
did not exist.

## Repair And Evidence

- Mapped nodes now identify `batch_factory` as their source.
- Canvas responses expose a protected node-media endpoint instead of an Agent
  asset endpoint. That endpoint checks owner, project binding and origin, then
  resolves the current Batch Factory media with the canonical book ID. It does
  not copy media or alter Batch Factory records.
- The focused Go regression covers a canonical/legacy book-ID mismatch and
  verifies the protected endpoint redirects only to the matching media record.
- The focused frontend regression verifies this protected URL still uses the
  authenticated media client.
- Local-executor image records now redirect through the existing owner-scoped
  artifact reader. The storage reference stays private, and no image is copied
  or mutated.
- Focused tests passed before local restart:
  - `go test ./internal/httpapi -run 'Test(BatchFactoryProjectionMediaReadUsesCanonicalBookID|SyncBookDoesNotPersistSignedMediaURL|BatchFactorySyncUsesMappedStateWithoutAgentGenerationTask|DeleteBoundNodeOnlyRemovesCanvasReference)' -count=1`
  - `go test ./internal/agentworkspace -run 'TestSyncBatchFactoryBook|TestResolveVideoDuration' -count=1`
  - `node --test src/user/pages/agent-canvas-media.test.js src/user/pages/agent-canvas-batch-factory.test.js`
  - `npm.cmd run build`
  - `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

## Preserved Boundaries

- The five existing node types remain unchanged.
- No Agent media copy is created.
- Removing a mapped node still removes only its canvas reference and origin.

## Runtime Status

The local Go runtime was rebuilt and restarted at SHA `3c4d8625`.

Authenticated browser acceptance on the supplied project confirmed that:

- `豆包资产图测试` loaded as a `2048 x 2048` image from the protected local
  artifact path, with no media error.
- Both mapped video nodes loaded without a `素材不存在` error.

## Final Lifecycle Source Repair

Commit: `33f89804` (`fix(agent): complete shared production lifecycle`).

- `Agent 优化` now detects a soft-deleted bound Agent project, creates a live
  replacement and atomically rebinds the owner and batch before returning it.
- A mapped video action now calls the existing single-video book-stage route
  with the mapped `videoId`, effective provider and `missing` or `force` mode.
  It no longer uses the whole-book production submit path or its default
  `personal_api` argument.
- The production router now injects an owner-scoped Batch Factory audio reader
  into projection, so persisted TTS artifacts and their duration return as
  mapped audio nodes without exposing the local storage path.

Fresh verification around the commit:

- `go test ./internal/agentworkspace ./internal/httpapi -run 'Test(MemoryBatchFactoryRebindRequiresOwnerAndLiveReplacementProject|SyncBatchFactoryBookProjectsOnlyExistingNodeTypes|BindBatchFactoryRecreatesDeletedBoundProject|RouterProjectsPersistedBatchFactoryAudioForOwner)$' -count=1`
- `node --test src/user/pages/agent-canvas-batch-factory.test.js` (`14` passed,
  `0` failed)
- `vite build --outDir C:\Users\Administrator\AppData\Local\Temp\codex-v88-lifecycle-33f89804`
  (`5514` modules transformed; build succeeded with the existing chunk-size
  warning)
- `git diff --cached --check`

Each of the three focused regressions was also mutation-checked: removing the
project recreation call, provider forwarding or audio-reader injection made
its matching test fail before the repair was restored. This commit was not
deployed or browser-accepted; the runtime evidence above still refers to the
earlier local runtime SHA.

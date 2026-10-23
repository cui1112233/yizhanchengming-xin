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

The rebuilt embedded binary is ready. Browser recheck of the supplied local
project is pending the localhost restart.

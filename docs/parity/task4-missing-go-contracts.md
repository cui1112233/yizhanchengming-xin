# Task 4 Go Contract Ledger

Updated: 2026-10-08

## Available and connected

| Page | Contract | Persistence / protection |
| --- | --- | --- |
| `/script` | project/book read, original-text save, project production settings, StageRun read/retry, storyboard card CRUD/order/recompile | MySQL; writes use same-origin API routes and project capability checks |
| `/novel-fetch` | Intake creation/read, book list, execute/retry and project creation | MySQL; list/read and execute permissions are enforced by Go |
| `/novel-fetch-workshop` | workshop snapshot/settings save, original-text restore and intake retry | `intakes.workshop_settings_json` plus persisted books; Go prompt catalog metadata is read-only in the page |
| `/novel-panel` | workspace read/save, revision conflict, history list/restore | `novel_panel_workspaces` and `novel_panel_history`; project access and same-origin write checks are registered in Go |

## Real-operation gaps (not replaced by front-end success)

1. **Workshop external upload/TOS submission lifecycle:** Task 4 has no Go endpoint returning an upload task, object key, provider receipt, polling state, remote readback, or retry contract. The Workshop page therefore has no upload-success control and does not transfer selected work through browser storage.
2. **Novel Panel provider execution:** the workspace can preserve Director request context but the actual provider execution/asset/TOS lifecycle belongs to shared generation and media services. Until a shared Go contract is wired, UI only exposes the persisted editing workspace and never claims a generated image, audio, or video.
3. **Script provider outcome:** StageRun and storyboard persistence are available. A provider failure remains a Go-reported failed stage and may only be retried through the Go stage/recompile endpoint; the UI cannot manufacture a completion.

## Explicit exclusions

- No Node request, iframe, postMessage bridge, Bearer token, `/api/chat`, localStorage, or sessionStorage carries a Task 4 business fact.
- System prompt bodies remain in Go prompt modules. The front end shows only prompt keys and versions received from Go.

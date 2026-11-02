# Agent Canvas Unified Background Queue Design

Date: 2026-10-23

## Goal

Make image, TTS/audio, and video generation continue after the canvas page is closed. A task must have one truthful state, be recoverable after a Go-process restart, support retry and cancellation, and contribute to one project-level progress summary.

## Scope

- Reuse `agent_generation_tasks` as the common task record. Do not create a second task table.
- Keep video provider task records in `agent_video_jobs`; they remain the detailed provider-side state for video.
- Keep existing assets, selected results, script shots, canvas nodes, project conversation, and model selection contracts.
- The Go embedded runtime remains the only local application runtime. No Node release path is introduced.

## State Model

All three task types expose the same state to the canvas:

`queued -> running -> awaiting_confirmation -> completed`

Terminal alternatives are `failed` and `cancelled`.

- A request writes `queued` before any external call.
- The worker changes it to `running` before it starts the provider, TTS, or processor.
- Generated assets are saved before `awaiting_confirmation`.
- A user confirms a result to move it to `completed`.
- A failed task keeps its error message and can create a new retry task linked to the original.
- A cancellation stops future work. A remote video cancellation keeps its provider-specific cancellation path.

## Background Execution

One Go coordinator scans pending tasks every two seconds. It dispatches by `kind` and `mode`:

- `image`: use the existing configured image generator, save assets, then publish candidates.
- `audio`: use the configured audio generator, built-in TTS, or audio processor, save assets, then publish candidates.
- `video`: retain the existing video coordinator and provider polling; its state is mirrored to the common task.
- `text/script_split`: retain the existing script-split coordinator.

The coordinator records work before calling external services. On restart, queued work resumes. Running image/audio work is retried only after a lease timeout; video submitting work is never submitted twice and instead becomes `abandoned` for manual review when its provider task ID was not saved.

## Retry, Cancel, and Summary

- Retry always creates a new task with `retryOfTaskId`; historical result and failure information remains visible.
- Cancel is idempotent. It prevents queued work from starting and delegates video cancellation to the provider path when applicable.
- The project summary is calculated from persisted tasks: total, queued, running, awaiting confirmation, completed, failed, and cancelled. It never infers `completed` when a record is missing.
- The canvas shows the same summary in workflow and storyboard views because they are two presentations of one project state.

## Acceptance

1. Submit one image, one built-in TTS/audio, and one video task; close or refresh the canvas page; all continue from persisted state.
2. Restart the local Go runtime while each task is pending; safe tasks resume, and uncertain provider submission is visibly marked for review rather than sent again.
3. Retry and cancel update both individual task state and project summary without changing unrelated nodes or projects.
4. Run focused backend tests, frontend tests, `npm --prefix frontend run build`, the Go embedded build, local runtime restart, health check, and an authenticated browser pass at `http://localhost:4001/agent/canvas`.

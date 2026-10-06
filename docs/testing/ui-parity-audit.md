# UI / Layout / Interaction Parity Audit

Baseline: `origin/main` at `f533e6e0ef8b68132b9d9992abf81c3e7323e778` (PR #32 merge).

This audit records repository and controlled-browser evidence only. A `PASS` requires a saved screenshot from both the old public UI and this branch; no such old-public screenshot set is available in this checkout, so this audit deliberately does not mark any row `PASS`.

## Shared parity changes in this branch

- Canonical status presentation is centralized in `前台/src/ui/statusPresentation.js`, rendered through `StatusTag.jsx`. Unknown values render as `未知状态（value）`; they never fall back to `待执行`.
- `PageState.jsx` provides the shared loading, empty, failure, and retry presentation and is used by the Shuihuo workbench error/empty states.
- Global Ant Design tokens standardize control height, radius, Card header height, table row padding, and Drawer footer spacing.
- Login now uses the copied V88 brand images in `前台/public/assets/brand-logo-black.png` and `brand-logo-white.png`, rather than a text-only Card title.
- Production Settings, Publishing Settings, and Version Profile remain in-place 760px right Drawers. `解析输入` is not a user action.

## Area status

| Area | Status | Controlled browser evidence | Remaining parity evidence / gap |
| --- | --- | --- | --- |
| Login | PARTIAL | Task16 auth fixture cases pass; logo is asserted by unit test. | No saved side-by-side public screenshot. |
| Novel Fetch | PARTIAL | Task16 grouping, dedupe, partial failure, immediate and scheduled flows pass. | No saved old/new screenshot comparison. |
| Shuihuo Production | PARTIAL | Task16 page-safety case passes; shared empty/error states applied. | No saved visual comparison. |
| BatchProject List | PARTIAL | Task16 list visibility and metadata cases pass; run tags use `StatusTag`. | No saved visual comparison. |
| Batch Factory V11 | PARTIAL | Task16 project detail and generation flows pass. | No saved visual comparison. |
| Production Settings | PARTIAL | Task16 verifies in-place Drawer save and URL retention. | No saved visual comparison. |
| Publishing Settings | PARTIAL | Task16 verifies in-place Drawer save and URL retention. | No saved visual comparison. |
| Version Profile | PARTIAL | Task16 verifies profile Drawer and sync actions. | No saved visual comparison. |
| Script | PARTIAL | Task16 generation stage acceptance passes. | No saved visual comparison. |
| Hook | PARTIAL | Task16 enabled and skipped states pass. | No saved visual comparison. |
| Director | PARTIAL | Task16 Director mode acceptance passes. | No saved visual comparison. |
| Audio | PARTIAL | Task16 measured-audio and error paths pass. | No saved visual comparison. |
| matchAudio | PARTIAL | Task16 measurement, payload, and timeline validation pass. | No saved visual comparison. |
| VIDEO | PARTIAL | Task16 status/retry/cancel and `outputUrl` link cases pass. | No saved visual comparison. |
| Merge | BLOCKED | API-level Task16 merge checks pass. | Batch Factory has no user-visible Merge panel in this branch, so no UI screenshot or visual parity claim is possible. |
| Publishing | PARTIAL | Task16 permission behavior passes. | No user-visible publishing execution panel or screenshot comparison in this branch. |

## Verification run

- `npm test` — passed locally after the shared UI changes.
- `npm run build` — passed locally after the shared UI changes.
- `npm run test:e2e` — 49 controlled-fixture Playwright cases passed locally after installing Chromium.

## Explicit limitations

The controlled Playwright suite is not public-provider or authenticated-public E2E evidence. It also does not retain screenshots for successful runs. Until screenshots are captured from the old public UI and this branch using equivalent fixture data, the visual conclusions above must remain `PARTIAL` or `BLOCKED`.

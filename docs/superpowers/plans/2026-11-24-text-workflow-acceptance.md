# Text workflow acceptance implementation plan

**Goal:** Make local text production changes, inheritance, previews and publishing states understandable without paid model calls.

**Architecture:** Retain Go text rendering and existing immutable page/collection records. Add current settings to the read model, derive actionable frontend status, and reuse existing prepare/render operations. Keep preview templates and inheritance helpers independent of the workbench.

**Tech stack:** Go, MySQL, React, Ant Design; existing embedded local 4001 runtime.

## Constraints

- Use current v88 branch; preserve unrelated edits; do not push or deploy publicly.
- Never delete source, truncate sentences, overwrite locked uploaded images, or automatically publish.
- No model calls for layout/template changes. Keep generated paper images pending selection.
- Keep one plan file; do not create parallel design documents.

## Tasks

- [x] Configuration state: return `settingsSnapshot` from text page GET; implement `textUpdateState(snapshot,collection)` with prepare/reflow/render/compose/ready stages. The UI must name affected content and let the user explicitly update it through prepare and pending-page renders.
  Files: `backend/internal/httpapi/batch_factory_text_pages.go`, `frontend/src/user/pages/shuihuo/textWorkflowStatus.js`, `TextMediaPanel.jsx`.
  Verify geometry changes require reflow, color changes require repaint, same style remains current, and locked pages are never silently replaced.
- [x] Scope: label single-page editing and explain it does not pull text from the next page; whole-book updates remain in a separate group. Preserve close/cancel/version protection.
  Files: `TextPageEditor.jsx`, `TextMediaPanel.jsx`.
- [x] Inheritance: group shared image-publish overrides with paired producer/creative and organization/name restoration. Restore only selected nested paths.
  Files: `TextProductionSettingsForm.jsx`, `textProductionConfig.js`, configuration browser test.
- [x] Preview/templates: add reading/handwriting/black-background presets with explicit application, shared text color quick actions, and a settings preview that uses the same renderer through the existing current-page preview endpoint. No silent application on opening configuration.
  Files: `textLayoutTemplates.js`, `TextProductionSettingsForm.jsx`, `TextPageEditor.jsx`.
- [x] Reading details: prohibit common Chinese closing punctuation at line start and opening punctuation at line end where a legal alternative fits. Preserve original coordinates and use identical measurement/render behavior.
  Files: `backend/internal/batchfactorytext/render.go`, renderer tests.
- [x] Collections/connection: show collection version, page count and outdated status; require explicit confirmation for historical uploads; show a review of selected target and page range. Add N8 directory refresh/status without exposing tokens or adding passwords to the UI.
  Files: `TextMediaPanel.jsx`, `textWorkflowStatus.js`, `BatchFactoryImagePublishSettingsForm.jsx`.
- [ ] Verify and deliver: execute focused Go packages and browser/state tests; build frontend, rebuild embedded Go, restart only confirmed 4001 binary, verify health/build identity. Commit only task files.

## Verification commands

```powershell
# frontend/
node --test src/user/pages/shuihuo/textWorkflowStatus.test.js src/user/pages/shuihuo/textProductionConfig.test.js src/user/pages/shuihuo/TextProductionSettingsForm.browser.test.js src/user/pages/shuihuo/TextMediaPanel.source.test.js
npm run build
# backend/
go test ./internal/batchfactorytext ./internal/renderfont -count=1
go test ./internal/httpapi -run TestBatchFactoryText -count=1
```

# Agent Canvas Phase 5: Advanced Tools and Final Release Verification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the creative workspace with capability-gated media tools, reusable presets/assets, stable sharing boundaries, performance safeguards, and full Go-embedded local acceptance.

**Architecture:** Register creative tools in a shared catalog whose enabled state comes from tested Go capabilities. Execute every enabled tool as a normal generation task that produces durable result assets. Close with migration, regression, performance, accessibility, build, and authenticated browser gates before any release is considered.

**Tech Stack:** React 18, Ant Design 5, Go, MySQL, Redis, TOS, FFmpeg/external configured connectors, `@xyflow/react`, Node and Go tests.

## Global Constraints

- Every advertised executable action has a real Go route, validation, task state, durable result, and test.
- An unavailable tool is visibly disabled with a concise reason and cannot submit a request.
- Presets contain public generation settings only; credentials and private URLs are never stored.
- Reusable project assets remain account/project scoped.
- Publish/share behavior exposes only explicitly selected material and never grants object-store credentials.
- The default edge animation remains lightweight and obeys `prefers-reduced-motion`.
- Completion requires real authenticated local browser interaction in the rebuilt Go-embedded runtime; source markers, HTTP 200, build output, and focused tests alone are insufficient.
- No public deployment or Git push occurs without a separate explicit user instruction.

---

## File Map

- `backend/internal/agentworkspace/tool_catalog.go`: tool availability/capability registry.
- `backend/internal/httpapi/agent_tools.go`: execute and inspect creative tools.
- `frontend/src/user/pages/agent-tool-catalog.js`: visible grouping and availability projection.
- `frontend/src/user/components/AgentNodeLibrary.jsx`: searchable node/tool library.
- `backend/internal/agentworkspace/presets.go`: account-scoped public presets.
- `frontend/src/user/components/AgentPresetPicker.jsx`: preset management.
- `backend/internal/httpapi/agent_share.go`: explicit read-only share manifest where configured.
- `test/agent-canvas-release-contract.test.js`: final source/release boundary guards.

### Task 1: Capability-gated creative tool catalog

**Files:**
- Create: `backend/internal/agentworkspace/tool_catalog.go`
- Create: `backend/internal/agentworkspace/tool_catalog_test.go`
- Create: `backend/internal/httpapi/agent_tools.go`
- Create: `backend/internal/httpapi/agent_tools_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Create: `frontend/src/user/pages/agent-tool-catalog.js`
- Create: `frontend/src/user/pages/agent-tool-catalog.test.js`
- Create: `frontend/src/user/components/AgentNodeLibrary.jsx`

**Interfaces:**
- Tool IDs: `image_upscale`, `image_panorama`, `image_grid_slice`, `video_frame_extract`, `audio_trim`, `audio_speed`, `audio_smart_split`, and `audio_custom_split`.
- Produces public tool entry `{ id, kind, label, enabled, disabledReason, inputKinds, settingsSchema }`.

- [ ] **Step 1: Write failing availability and disabled-submit tests**

```go
func TestToolCatalogDisablesMissingProcessorAndRejectsExecution(t *testing.T) {
    catalog := BuildToolCatalog(RuntimeCapabilities{FFmpeg: false})
    require.False(t, catalog.ByID("audio_trim").Enabled)
    require.Equal(t, http.StatusConflict, executeTool(t, "audio_trim").Code)
}
```

```js
test('node library separates enabled and unavailable tools', () => {
  const groups = groupAgentTools(catalog);
  assert.equal(groups.processing.find(x => x.id === 'audio_trim').disabledReason, '本机未配置音视频处理器');
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run TestTool -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-tool-catalog.test.js`

Expected: FAIL because the shared tool catalog and execution boundary do not exist.

- [ ] **Step 3: Implement catalog and task-backed execution**

```go
type PublicTool struct {
    ID string `json:"id"`
    Kind string `json:"kind"`
    Label string `json:"label"`
    Enabled bool `json:"enabled"`
    DisabledReason string `json:"disabledReason,omitempty"`
    InputKinds []string `json:"inputKinds"`
    SettingsSchema map[string]any `json:"settingsSchema,omitempty"`
}
```

Execution verifies tool enabled state, source asset ownership/type, setting ranges, and destination project before creating a generic generation task. The node library searches labels/descriptions and renders unavailable items disabled with the server reason.

- [ ] **Step 4: Run tool tests and frontend build**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run TestTool -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-tool-catalog.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit capability-gated tools**

```powershell
git add backend/internal/agentworkspace/tool_catalog.go backend/internal/agentworkspace/tool_catalog_test.go backend/internal/httpapi/agent_tools.go backend/internal/httpapi/agent_tools_test.go backend/internal/httpapi/router.go frontend/src/user/pages/agent-tool-catalog.js frontend/src/user/pages/agent-tool-catalog.test.js frontend/src/user/components/AgentNodeLibrary.jsx
git commit -m "feat(agent): gate creative tools by capability"
```

### Task 2: Presets and reusable assets

**Files:**
- Create: `backend/internal/agentworkspace/presets.go`
- Create: `backend/internal/agentworkspace/presets_test.go`
- Create: `backend/internal/httpapi/agent_presets.go`
- Create: `backend/internal/httpapi/agent_presets_test.go`
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Modify: `backend/internal/storage/agent_workspace_schema_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Create: `frontend/src/user/components/AgentPresetPicker.jsx`
- Create: `frontend/src/user/pages/agent-presets.test.js`
- Modify: `frontend/src/shared/api/agent.js`

**Interfaces:**
- Produces account-scoped preset CRUD with optimistic version: `{ id, name, kind, provider, modelId, mode, parameters, advanced, version }`.
- Applying a preset returns a node-config patch and passes through capability reconciliation before save.

- [ ] **Step 1: Confirm migration number and write failing redaction/scope tests**

Run: `rg -n "Version: 880000" backend/internal/storage/agent_workspace_schema.go`

Expected: after Phase 3, `8800008` is the latest Agent workspace migration and `8800009` is unused; otherwise use the next unused version consistently.

```go
func TestPresetRejectsCredentialFieldsAndStaysOwnerScoped(t *testing.T) {
    _, err := store.CreatePreset(ctx, "alice", Preset{Advanced: map[string]any{"apiKey":"secret"}})
    require.ErrorIs(t, err, ErrPresetPrivateField)
    require.Empty(t, store.ListPresets(ctx, "bob", "image"))
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run TestPreset -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-presets.test.js`

Expected: FAIL because preset persistence and picker are absent.

- [ ] **Step 3: Implement preset CRUD and picker**

```jsx
<AgentPresetPicker
  kind={kind}
  currentConfig={config}
  onApply={patch => onConfigChange(reconcileConfigWithCapabilities(patch, selectedModel))}
/>
```

Append migration version `8800009` to `AgentWorkspaceMigrations()` with a new checksum and the account-scoped preset table. Reject keys matching `token|key|secret|password|authorization|credential|url` recursively. Rename/update requires the current version; delete affects the preset only and never modifies nodes that previously applied it.

- [ ] **Step 4: Run preset and schema tests**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi ./internal/storage -run 'TestPreset|TestAgentWorkspaceSchema' -count=1; Set-Location ..; node --test frontend/src/user/pages/agent-presets.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit presets**

```powershell
git add backend/internal/agentworkspace/presets.go backend/internal/agentworkspace/presets_test.go backend/internal/httpapi/agent_presets.go backend/internal/httpapi/agent_presets_test.go backend/internal/storage/agent_workspace_schema.go backend/internal/storage/agent_workspace_schema_test.go backend/internal/httpapi/router.go frontend/src/user/components/AgentPresetPicker.jsx frontend/src/user/pages/agent-presets.test.js frontend/src/shared/api/agent.js
git commit -m "feat(agent): add reusable generation presets"
```

### Task 3: Explicit project sharing boundary

**Files:**
- Create: `backend/internal/agentworkspace/share_manifest.go`
- Create: `backend/internal/agentworkspace/share_manifest_test.go`
- Create: `backend/internal/httpapi/agent_share.go`
- Create: `backend/internal/httpapi/agent_share_test.go`
- Modify: `backend/internal/httpapi/router.go`
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`

**Interfaces:**
- Produces a read-only share manifest containing selected node/material metadata and expiring protected asset references.
- Sharing is absent/disabled when signing configuration is unavailable.

- [ ] **Step 1: Write failing explicit-selection and secret-redaction tests**

```go
func TestShareManifestContainsOnlySelectedPublicFields(t *testing.T) {
    manifest, err := BuildShareManifest(project, []string{"node-2"})
    require.NoError(t, err)
    require.Len(t, manifest.Nodes, 1)
    encoded, _ := json.Marshal(manifest)
    require.NotContains(t, string(encoded), "apiKey")
    require.NotContains(t, string(encoded), "object_key")
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run TestShare -count=1`

Expected: FAIL because no explicit share manifest exists.

- [ ] **Step 3: Implement read-only sharing**

```go
type ShareManifest struct {
    ProjectTitle string `json:"projectTitle"`
    Nodes []SharedNode `json:"nodes"`
    Assets []SharedAsset `json:"assets"`
    ExpiresAt time.Time `json:"expiresAt"`
}
```

Require the owner to select content, set an expiry, and confirm creation. Do not expose editing, conversation history, credentials, internal object keys, hidden nodes, or unselected assets.

- [ ] **Step 4: Run share tests and build**

Run: `Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi -run TestShare -count=1; Set-Location ..; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit the sharing boundary**

```powershell
git add backend/internal/agentworkspace/share_manifest.go backend/internal/agentworkspace/share_manifest_test.go backend/internal/httpapi/agent_share.go backend/internal/httpapi/agent_share_test.go backend/internal/httpapi/router.go frontend/src/user/pages/AgentPageV2.jsx
git commit -m "feat(agent): add explicit workspace sharing"
```

### Task 4: Performance, motion, keyboard, and accessibility guards

**Files:**
- Create: `frontend/src/user/pages/agent-canvas-performance.test.js`
- Create: `frontend/src/user/pages/agent-canvas-accessibility.source.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentCanvasControls.jsx`
- Modify: `frontend/src/user/pages/agent-canvas-commands.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Produces stable rendering for 200 nodes/300 edges, reduced-motion edge behavior, keyboard focus visibility, and complete button labels/tooltips.

- [ ] **Step 1: Write failing performance and source guards**

```js
test('large graph transforms remain bounded and preserve node count', () => {
  const started = performance.now();
  const result = layoutCanvas(makeNodes(200), makeEdges(300), 'LR');
  assert.equal(result.length, 200);
  assert.ok(performance.now() - started < 250);
});

test('icon controls have accessible names and reduced motion is declared', () => {
  assert.match(read('frontend/src/user/components/AgentCanvasControls.jsx'), /aria-label=/);
  assert.match(read('frontend/src/shared/styles/agent-workspace.css'), /prefers-reduced-motion/);
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-performance.test.js frontend/src/user/pages/agent-canvas-accessibility.source.test.js`

Expected: FAIL until the performance/motion/accessibility guards are complete.

- [ ] **Step 3: Implement memoization and reduced-motion styling**

```css
@media (prefers-reduced-motion: reduce) {
  .agent-canvas .react-flow__edge-path { animation: none !important; }
}

.agent-workspace button:focus-visible,
.agent-workspace [role="button"]:focus-visible {
  outline: 2px solid var(--agent-focus);
  outline-offset: 2px;
}
```

Memoize node/edge projections by stable signatures, debounce search only, virtualize long drawer/history lists, and avoid polling completed tasks. Keep canvas pan/drag updates immediate.

- [ ] **Step 4: Run frontend regression and build**

Run: `$canvasTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-canvas-*.test.js').FullName; $workspaceTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-workspace-*.test.js').FullName; $storyboardTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-storyboard*.test.js').FullName; $generationTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-generation-*.test.js').FullName; node --test $canvasTests $workspaceTests $storyboardTests $generationTests; npm --prefix frontend run build`

Expected: PASS and build succeeds without layout overflow warnings.

- [ ] **Step 5: Commit quality guards**

```powershell
git add frontend/src/user/pages/agent-canvas-performance.test.js frontend/src/user/pages/agent-canvas-accessibility.source.test.js frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/AgentCanvasControls.jsx frontend/src/user/pages/agent-canvas-commands.js frontend/src/shared/styles/agent-workspace.css
git commit -m "perf(agent): harden large canvas interactions"
```

### Task 5: Full local Go-embedded release verification

**Files:**
- Create: `test/agent-canvas-release-contract.test.js`
- Modify: `docs/V88_PROJECT_EXECUTION_MEMORY.md`
- Modify: `docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md`

**Interfaces:**
- Consumes: all five Agent canvas phases.
- Produces: exact local commit/build evidence and a verified rollback point; no public release.

- [ ] **Step 1: Add final architecture guards**

```js
test('agent workspace ships from React through Go embed only', () => {
  assert.doesNotMatch(read('frontend/src/user/pages/AgentPageV2.jsx'), /iframe/i);
  assert.match(read('scripts/deploy-go-embedded.ps1'), /go build/i);
  assert.equal(fs.existsSync('deploy/v88-direct'), false);
});
```

- [ ] **Step 2: Run complete focused regression**

Run: `$apiTests = (Get-ChildItem frontend/src/shared/api -Filter '*.test.js').FullName; $layoutTests = (Get-ChildItem frontend/src/shared/layouts -Filter '*.test.js').FullName; $agentTests = (Get-ChildItem frontend/src/user/pages -Filter 'agent-*.test.js').FullName; $contractTests = (Get-ChildItem test -Filter 'agent-*.test.js').FullName; node --test $apiTests $layoutTests $agentTests $contractTests; npm --prefix frontend run build; Set-Location backend; go test ./internal/agentworkspace ./internal/httpapi ./internal/storage -count=1; go build ./cmd/api`

Expected: all tests and builds succeed. Record unrelated suite failures separately and do not describe them as Agent acceptance.

- [ ] **Step 3: Build and start the actual local review binary**

```powershell
Set-Location F:\脚本测试\一战晟铭-v88
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql
.\scripts\check-local-go.ps1
git rev-parse HEAD
```

Expected: `localhost:4001` serves the newly embedded frontend and build info matches the recorded local HEAD/build.

- [ ] **Step 4: Execute authenticated end-to-end acceptance**

Verify in the real browser: create/open project; workflow/storyboard switch; node library and asset drag; `+` handle connection and blank-drop menu; connected references; middle-button pan; zoom/minimap/snap/layout; grouping/alignment/undo/save/reload; compact editors; exact default image summary `16:9 · 标准画质 · 2K · 1张`; independent provider/model/settings for all four node kinds; text/image/video/audio generation; multiple outputs and confirmation; cancel/retry; history; audio processing; presets; material categories; supervised Director feedback; explicit automatic execution and stop conditions; disabled unavailable tools; share selection when configured; desktop/mobile layout; reduced motion.

- [ ] **Step 5: Record evidence and commit final local completion**

Record the verified SHA, test commands/results, local URL, screenshots, provider tests actually exercised, unconfigured providers not exercised, and previous stable SHA in the execution memory. Do not claim provider end-to-end coverage for unavailable credentials or services.

```powershell
git add test/agent-canvas-release-contract.test.js docs/V88_PROJECT_EXECUTION_MEMORY.md docs/superpowers/specs/2026-10-21-agent-canvas-liblib-workspace-design.md
git commit -m "test(agent): verify complete creative workspace"
```

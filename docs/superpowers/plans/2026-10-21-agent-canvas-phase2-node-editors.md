# Agent Canvas Phase 2: Unified Node Editors and Model Capabilities Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give text, image, video, and audio nodes one compact editor pattern with independent provider/model/settings and capability-driven controls.

**Architecture:** Normalize every node into one versioned `data.config` shape. The Go model catalog publishes credential-free capabilities; frontend pure helpers filter options, reconcile legacy nodes, and format summaries. `AgentNodeComposer` renders focused kind-specific panels beneath the selected preview and sends the same normalized values shown in the summary.

**Tech Stack:** React 18, Ant Design 5, Go, MySQL JSON fields, Node test runner, Go tests.

## Global Constraints

- Each node owns its provider, model, mode, parameters, advanced settings, manual references, and edge-derived references.
- Image nodes default to the exact summary `16:9 · 标准画质 · 2K · 1张`.
- The four image summary values are not decorative: `ratio`, `quality`, `resolution`, and `count` are the values persisted and submitted.
- Text nodes never show image generation channel or video engine settings.
- Image, video, text, and audio pickers show only models of the matching kind.
- Changing provider visibly reconciles the model; generation never silently replaces a stale model.
- No API credential is serialized in model capability responses or canvas snapshots.
- Do not edit an already-applied migration; add the next append-only migration only if indexed columns are required.

---

## File Map

- `frontend/src/user/pages/agent-node-config.js`: node config defaults, backward normalization, and summary formatting.
- `frontend/src/user/pages/agent-model-capabilities.js`: provider/model filtering and value reconciliation.
- `frontend/src/user/components/AgentModelPicker.jsx`: searchable provider-filtered rich picker.
- `frontend/src/user/components/AgentNodeParameterPopover.jsx`: kind-specific compact parameters.
- `frontend/src/user/components/AgentNodeSettingsPanel.jsx`: one settings entry with provider and advanced controls.
- `frontend/src/user/components/AgentNodeComposer.jsx`: unified selected-node editor shell.
- `backend/internal/httpapi/model_catalog_capabilities.go`: credential-free capability projection and validation beside the existing catalog projection.
- `backend/internal/httpapi/model_catalog.go`: capability response.

### Task 1: Normalize per-node configuration and exact summaries

**Files:**
- Create: `frontend/src/user/pages/agent-node-config.js`
- Create: `frontend/src/user/pages/agent-node-config.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.js`
- Modify: `frontend/src/user/pages/agent-canvas-context-menu.test.js`

**Interfaces:**
- Produces: `normalizeNodeConfig(kind, value) -> NodeConfig`, `formatNodeParameterSummary(kind, config) -> string`, and `patchNodeConfig(kind, current, patch) -> NodeConfig`.
- Image parameters: `{ ratio: '16:9', quality: 'standard', resolution: '2K', count: 1 }`.
- Video parameters: `{ ratio: '16:9', resolution: '720P', durationSeconds: 5, count: 1, generateAudio: false }`.
- Text parameters: `{ operation: 'generate', length: 'standard', creativity: 'balanced' }`.

- [ ] **Step 1: Write failing default and update tests**

```js
test('image summary has the exact approved default and follows edits', () => {
  const config = normalizeNodeConfig('image', {});
  assert.equal(formatNodeParameterSummary('image', config), '16:9 · 标准画质 · 2K · 1张');
  const changed = patchNodeConfig('image', config, { parameters: { ratio: '9:16', count: 2 } });
  assert.equal(formatNodeParameterSummary('image', changed), '9:16 · 标准画质 · 2K · 2张');
});
```

- [ ] **Step 2: Run and verify failure**

Run: `node --test frontend/src/user/pages/agent-node-config.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: FAIL because normalized kind-specific config and the exact summary formatter do not exist.

- [ ] **Step 3: Implement immutable normalization**

```js
const IMAGE_DEFAULTS = Object.freeze({ ratio: '16:9', quality: 'standard', resolution: '2K', count: 1 });

export function formatNodeParameterSummary(kind, config) {
  const p = normalizeNodeConfig(kind, config).parameters;
  if (kind === 'image') return `${p.ratio} · ${qualityLabel[p.quality]} · ${p.resolution} · ${p.count}张`;
  if (kind === 'video') return `${p.ratio} · ${p.resolution} · ${p.durationSeconds}s · ${p.count}个${p.generateAudio ? ' · 含声音' : ''}`;
  return `${operationLabel[p.operation]} · ${lengthLabel[p.length]} · ${creativityLabel[p.creativity]}`;
}
```

Update all node-creation helpers to attach normalized config without changing existing node IDs or assets.

- [ ] **Step 4: Run focused tests**

Run: `node --test frontend/src/user/pages/agent-node-config.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js`

Expected: PASS.

- [ ] **Step 5: Commit the config contract**

```powershell
git add frontend/src/user/pages/agent-node-config.js frontend/src/user/pages/agent-node-config.test.js frontend/src/user/pages/agent-canvas-context-menu.js frontend/src/user/pages/agent-canvas-context-menu.test.js
git commit -m "feat(agent): normalize node generation settings"
```

### Task 2: Expose and validate public model capabilities in Go

**Files:**
- Create: `backend/internal/httpapi/model_catalog_capabilities.go`
- Create: `backend/internal/httpapi/model_catalog_capabilities_test.go`
- Modify: `backend/internal/httpapi/model_catalog.go`
- Modify: `backend/internal/httpapi/model_catalog_test.go`

**Interfaces:**
- Produces JSON `capabilities`: `{ ratios, resolutions, qualityLevels, duration, maxOutputCount, referenceKinds, generationModes, supportsAudio, advancedFlags }`.
- Produces: `ValidateModelSettings(model PublicModel, kind string, settings map[string]any) error`.

- [ ] **Step 1: Write failing serialization and rejection tests**

```go
func TestPublicModelCapabilitiesExcludeCredentialsAndValidateImageSettings(t *testing.T) {
    model := fixtureImageModel()
    public := PublicModelFromConfig(model)
    encoded, _ := json.Marshal(public)
    require.NotContains(t, string(encoded), "api_key")
    require.Equal(t, []string{"2K", "4K"}, public.Capabilities.Resolutions)
    require.Error(t, ValidateModelSettings(public, "image", map[string]any{"resolution": "8K"}))
}
```

- [ ] **Step 2: Run Go tests and verify failure**

Run: `Set-Location backend; go test ./internal/httpapi -run 'TestPublicModelCapabilities|TestModelCatalog.*Capabilities' -count=1`

Expected: FAIL because public capabilities and validation are absent.

- [ ] **Step 3: Implement capability projection and validation**

```go
type PublicCapabilities struct {
    Ratios          []string `json:"ratios,omitempty"`
    Resolutions     []string `json:"resolutions,omitempty"`
    QualityLevels   []string `json:"qualityLevels,omitempty"`
    MaxOutputCount  int      `json:"maxOutputCount,omitempty"`
    GenerationModes []string `json:"generationModes,omitempty"`
    ReferenceKinds  []string `json:"referenceKinds,omitempty"`
    SupportsAudio   bool     `json:"supportsAudio,omitempty"`
    AdvancedFlags   []string `json:"advancedFlags,omitempty"`
}
```

Reject unsupported values with stable field-specific messages; do not include configuration secrets in errors or responses.

- [ ] **Step 4: Run catalog and HTTP tests**

Run: `Set-Location backend; go test ./internal/httpapi -run 'TestPublicModelCapabilities|TestModelCatalog' -count=1`

Expected: PASS.

- [ ] **Step 5: Commit the public capability contract**

```powershell
git add backend/internal/httpapi/model_catalog_capabilities.go backend/internal/httpapi/model_catalog_capabilities_test.go backend/internal/httpapi/model_catalog.go backend/internal/httpapi/model_catalog_test.go
git commit -m "feat(agent): expose model capabilities"
```

### Task 3: Rich provider and model picker

**Files:**
- Create: `frontend/src/user/pages/agent-model-capabilities.js`
- Create: `frontend/src/user/pages/agent-model-capabilities.test.js`
- Create: `frontend/src/user/components/AgentModelPicker.jsx`
- Create: `frontend/src/user/pages/agent-model-picker.source.test.js`
- Modify: `frontend/src/shared/api/modelCatalog.js`
- Modify: `frontend/src/shared/api/modelCatalog.test.js`
- Modify: `frontend/src/user/pages/shuihuo/providerModelBinding.js`
- Modify: `frontend/src/user/pages/shuihuo/providerModelBinding.test.js`

**Interfaces:**
- Consumes: typed catalog models and `providerModelBinding.js` provider IDs.
- Produces: `compatibleModels(models, { kind, provider })`, `reconcileNodeModel(config, models)`, and `<AgentModelPicker kind config onChange />`.

- [ ] **Step 1: Write failing kind/provider/capability tests**

```js
test('an image node cannot select a video model and keeps an unavailable saved model visible', () => {
  assert.deepEqual(compatibleModels(models, { kind: 'image', provider: 'personal_api' }).map(x => x.id), ['img-1']);
  assert.deepEqual(reconcileNodeModel({ provider: 'personal_api', modelId: 'removed' }, models), {
    modelId: 'removed', available: false, reason: '当前模型不可用',
  });
});
```

- [ ] **Step 2: Run tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/shared/api/modelCatalog.test.js frontend/src/user/pages/shuihuo/providerModelBinding.test.js frontend/src/user/pages/agent-model-picker.source.test.js`

Expected: FAIL because capability projection and rich picker are missing.

- [ ] **Step 3: Implement the picker**

```jsx
<AgentModelPicker
  kind={kind}
  provider={config.provider}
  modelId={config.modelId}
  models={models}
  onChange={({ provider, modelId }) => onConfigChange({ provider, modelId })}
/>
```

Rows show model name, provider, capability description, expected duration, reference support, and unavailable reason. Search matches model name/provider/capability text. Opening the picker does not alter the saved model; only a user selection changes it.

- [ ] **Step 4: Run frontend model tests and build**

Run: `node --test frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/shared/api/modelCatalog.test.js frontend/src/user/pages/shuihuo/providerModelBinding.test.js frontend/src/user/pages/agent-model-picker.source.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit model selection**

```powershell
git add frontend/src/user/pages/agent-model-capabilities.js frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/user/components/AgentModelPicker.jsx frontend/src/user/pages/agent-model-picker.source.test.js frontend/src/shared/api/modelCatalog.js frontend/src/shared/api/modelCatalog.test.js frontend/src/user/pages/shuihuo/providerModelBinding.js frontend/src/user/pages/shuihuo/providerModelBinding.test.js
git commit -m "feat(agent): add capability driven model picker"
```

### Task 4: Compact unified editor and one settings entry

**Files:**
- Create: `frontend/src/user/components/AgentNodeParameterPopover.jsx`
- Create: `frontend/src/user/components/AgentNodeSettingsPanel.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/pages/agent-node-composer.source.test.js`
- Modify: `frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: normalized node config, compatible model capabilities, references, and generation callbacks.
- Produces: a same-width editor beneath the selected preview; `onConfigChange(patch)` saves through the canvas queue.

- [ ] **Step 1: Write failing editor behavior and label tests**

```js
test('composer renders one settings entry and the exact default image summary', () => {
  const source = read('frontend/src/user/components/AgentNodeComposer.jsx');
  assert.match(source, /AgentNodeSettingsPanel/);
  assert.match(source, /formatNodeParameterSummary/);
  assert.doesNotMatch(source, /图片生成通道[\s\S]*视频引擎/);
});
```

- [ ] **Step 2: Run source and config tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-node-config.test.js`

Expected: FAIL because the unified compact toolbar and settings panel are absent.

- [ ] **Step 3: Implement kind-specific compact controls**

```jsx
<div className="agent-node-composer__toolbar">
  <AgentModelPicker kind={kind} config={config} onChange={onConfigChange} />
  <AgentNodeParameterPopover kind={kind} config={config} capabilities={capabilities} onChange={onConfigChange} />
  <Button icon={<Settings2 size={15} />} onClick={() => setSettingsOpen(true)} aria-label="设置" />
  <Button type="primary" icon={<Sparkles size={15} />} onClick={onGenerate}>生成</Button>
</div>
```

Clicking a media preview opens the editor directly below it at the preview width; clicking blank canvas closes it. Closing never clears prompt, references, or config. The settings panel shows image channel only for image nodes, video engine only for video nodes, and provider/model-neutral settings for text/audio.

- [ ] **Step 4: Run editor tests and build**

Run: `node --test frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/user/pages/agent-node-config.test.js; npm --prefix frontend run build`

Expected: PASS and build succeeds.

- [ ] **Step 5: Commit the editor UI**

```powershell
git add frontend/src/user/components/AgentNodeParameterPopover.jsx frontend/src/user/components/AgentNodeSettingsPanel.jsx frontend/src/user/components/AgentNodeComposer.jsx frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages/agent-node-composer.source.test.js frontend/src/user/pages/agent-canvas-collapsible-editor.source.test.js frontend/src/shared/styles/agent-workspace.css
git commit -m "feat(agent): unify node editors"
```

### Task 5: Submit exactly the displayed node settings

**Files:**
- Modify: `frontend/src/user/pages/agent-canvas-node-generation.js`
- Modify: `frontend/src/user/pages/agent-canvas-node-generation.test.js`
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `backend/internal/httpapi/agent_node_generation.go`
- Modify: `backend/internal/httpapi/agent_node_generation_test.go`

**Interfaces:**
- Consumes: normalized `node.data.config` and current node version.
- Produces: request payload `{ nodeId, nodeVersion, provider, modelId, mode, parameters, advanced, referenceAssetIds, idempotencyKey }`.

- [ ] **Step 1: Write failing payload and Go validation tests**

```js
test('image request uses every value shown in the compact summary', () => {
  const payload = buildNodeGenerationRequest(imageNode);
  assert.deepEqual(payload.parameters, { ratio: '16:9', quality: 'standard', resolution: '2K', count: 1 });
  assert.equal(formatNodeParameterSummary('image', imageNode.data.config), '16:9 · 标准画质 · 2K · 1张');
});
```

```go
func TestAgentImageGenerationRejectsUnsupportedCapabilityValue(t *testing.T) {
    response := postImageGeneration(t, `{"parameters":{"ratio":"16:9","quality":"standard","resolution":"8K","count":1}}`)
    require.Equal(t, http.StatusUnprocessableEntity, response.Code)
}
```

- [ ] **Step 2: Run focused tests and verify failure**

Run: `node --test frontend/src/user/pages/agent-canvas-node-generation.test.js; Set-Location backend; go test ./internal/httpapi -run 'TestAgent(Image|Text)Generation' -count=1`

Expected: FAIL because current image submission does not carry all four summary values through capability validation.

- [ ] **Step 3: Implement one payload builder and server validation**

```js
export function buildNodeGenerationRequest(node) {
  const config = normalizeNodeConfig(node.type, node.data?.config);
  return {
    nodeId: node.id, nodeVersion: node.data?.version,
    provider: config.provider, modelId: config.modelId, mode: config.mode,
    parameters: config.parameters, advanced: config.advanced,
    referenceAssetIds: effectiveReferenceAssetIds(config),
    idempotencyKey: generationIdempotencyKey(node, config),
  };
}
```

Validate the selected public model, kind, provider, and settings before dispatch. Return stable field messages and retain the node draft on rejection.

- [ ] **Step 4: Run Phase 2 tests and builds**

Run: `node --test frontend/src/user/pages/agent-node-config.test.js frontend/src/user/pages/agent-model-capabilities.test.js frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/pages/agent-node-composer.source.test.js; npm --prefix frontend run build; Set-Location backend; go test ./internal/httpapi ./internal/agentworkspace -count=1; go build ./cmd/api`

Expected: all commands succeed.

- [ ] **Step 5: Perform local browser acceptance and commit**

Rebuild/start the Go review runtime. On separate image, video, text, and audio nodes, select different providers/models/settings, collapse editors, click blank canvas, reload, and verify each node restores independently. For an image node, verify the initial visible line is exactly `16:9 · 标准画质 · 2K · 1张`; change count to `2张`, generate, and verify the request/history records count `2`.

```powershell
git add frontend/src/user/pages/agent-canvas-node-generation.js frontend/src/user/pages/agent-canvas-node-generation.test.js frontend/src/user/components/AgentCanvas.jsx backend/internal/httpapi/agent_node_generation.go backend/internal/httpapi/agent_node_generation_test.go
git commit -m "feat(agent): submit per node generation settings"
```

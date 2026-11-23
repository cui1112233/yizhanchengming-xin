# Batch Factory Text Mode And Sensitive Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `text` as a first-class Batch Factory production mode and replace the old sensitive AI boolean with the shared `AI优化 / 系统优化 / 关闭` policy for video, comic, and text.

**Architecture:** Keep mode normalization and sparse settings inheritance in the existing V12 Batch Factory core. Add one shared sensitive-policy package used by all three outer workflows; keep each workflow responsible for deciding when to call it. Preserve old `sensitiveAIEnabled` settings as read compatibility, but persist only `sensitiveMode` for new saves.

**Tech Stack:** Go 1.26, React 18, Ant Design 5, Node test runner, Go testing

## Global Constraints

- Production modes are exactly `video`, `comic`, and `text`.
- Sensitive modes are exactly `ai`, `system`, and `off`; default is `system`.
- AI failure never silently falls back to system optimization.
- System optimization never calls a model and never overwrites the complete original.
- Unified configuration contains every whole-book production setting; book configuration uses the same fields as sparse overrides.
- Preserve unrelated dirty worktree changes and commit only files named by each task.
- Do not create a feature branch and do not push or deploy in this plan.

---

### Task 1: Make production-mode normalization accept text

**Files:**
- Modify: `backend/internal/batchfactoryv11/director_contract.go`
- Modify: `backend/internal/batchfactoryv11/settings.go`
- Test: `backend/internal/batchfactoryv11/settings_test.go`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryProductionMode.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js`

**Interfaces:**
- Produces: `normalizeProductionMode(SettingsPatch) string` and `EffectiveProductionMode(batchPatch, bookPatch SettingsPatch) string` returning `video|comic|text`.
- Produces: `batchFactoryProductionMode(value) string` and visibility `{ mode, video, comic, text }`.

- [ ] **Step 1: Add failing Go cases**

```go
func TestEffectiveProductionModeSupportsText(t *testing.T) {
    if got := normalizeProductionMode(SettingsPatch{"productionMode": raw("text")}); got != "text" {
        t.Fatalf("mode = %q, want text", got)
    }
    if got := EffectiveProductionMode(SettingsPatch{"productionMode": raw("text")}, SettingsPatch{"productionMode": raw("comic")}); got != "text" {
        t.Fatalf("batch mode = %q, want text", got)
    }
}
```

- [ ] **Step 2: Run the focused Go test and verify it fails**

Run: `go test ./internal/batchfactoryv11 -run TestEffectiveProductionModeSupportsText -count=1` from `backend`
Expected: FAIL because `text` normalizes to `video`.

- [ ] **Step 3: Extend both Go allow-lists to include text**

```go
if mode := rawString(patch, "productionMode", ""); mode == "comic" || mode == "video" || mode == "text" {
    return mode
}
```

- [ ] **Step 4: Add and pass frontend mode tests**

```js
assert.deepEqual(batchFactoryProductionModeVisibility('text'), {
  mode: 'text', video: false, comic: false, text: true
});
```

Run: `node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js` from `frontend`
Expected: PASS after adding `text` to the frontend normalizer.

- [ ] **Step 5: Commit the mode contract**

```bash
git add backend/internal/batchfactoryv11/director_contract.go backend/internal/batchfactoryv11/settings.go backend/internal/batchfactoryv11/settings_test.go frontend/src/user/pages/shuihuo/batchFactoryProductionMode.js frontend/src/user/pages/shuihuo/batchFactoryProductionMode.test.js
git commit -m "feat: recognize batch factory text mode"
```

### Task 2: Add the shared sensitive policy service

**Files:**
- Create: `backend/internal/batchfactorysensitive/policy.go`
- Create: `backend/internal/batchfactorysensitive/policy_test.go`
- Create: `backend/internal/prompt/batch_factory_sensitive.go`
- Test: `backend/internal/prompt/batch_factory_sensitive_test.go`

**Interfaces:**
- Produces: `type Mode string` with `ModeAI`, `ModeSystem`, `ModeOff`.
- Produces: `ResolveMode(settings batchfactoryv11.SettingsPatch) Mode`.
- Produces: `Processor.Process(ctx context.Context, input Input) (Result, error)`.
- Consumes: existing sensitive rules through an injected `RuleProcessor` and AI through an injected `AIProcessor`.

- [ ] **Step 1: Write failing policy tests**

```go
func TestResolveModeDefaultsAndLegacyCompatibility(t *testing.T) {
    if got := ResolveMode(nil); got != ModeSystem { t.Fatalf("default = %q", got) }
    if got := ResolveMode(patch("sensitiveMode", "off")); got != ModeOff { t.Fatalf("off = %q", got) }
    if got := ResolveMode(patch("sensitiveAIEnabled", true)); got != ModeAI { t.Fatalf("legacy = %q", got) }
}

func TestAIFailureDoesNotFallBack(t *testing.T) {
    p := Processor{AI: failingAI{}, Rules: countingRules{}}
    _, err := p.Process(context.Background(), Input{Mode: ModeAI, Text: "原文"})
    if err == nil { t.Fatal("expected AI error") }
}
```

- [ ] **Step 2: Run the package test and verify it fails to compile**

Run: `go test ./internal/batchfactorysensitive -count=1` from `backend`
Expected: FAIL because the package does not exist.

- [ ] **Step 3: Implement the explicit processor contract**

```go
type Input struct {
    Mode Mode
    Owner, BookID, Text, TextModelID string
}

type Result struct {
    OriginalText, ProductionText string
    Mode Mode
    Changes []Change
}

type Processor struct { Rules RuleProcessor; AI AIProcessor }
```

`ModeSystem` calls only `Rules`; `ModeAI` calls only `AI`; `ModeOff` returns the input unchanged. Validate the mode before dispatch.

- [ ] **Step 4: Move the AI system prompt into the prompt module**

Create `BatchFactorySensitiveSystemPrompt()` in `backend/internal/prompt/batch_factory_sensitive.go`. The function returns the approved server-side instruction; no frontend file receives the prompt body.

Run: `go test ./internal/prompt -run BatchFactorySensitive -count=1` from `backend`
Expected: PASS and prompt text is non-empty.

- [ ] **Step 5: Run all new backend tests**

Run: `go test ./internal/batchfactorysensitive ./internal/prompt -count=1` from `backend`
Expected: PASS.

- [ ] **Step 6: Commit the shared service**

```bash
git add backend/internal/batchfactorysensitive backend/internal/prompt/batch_factory_sensitive.go backend/internal/prompt/batch_factory_sensitive_test.go
git commit -m "feat: add shared sensitive processing policy"
```

### Task 3: Replace the old switch in unified and book settings

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryBookConfigRegions.js`
- Test: `frontend/src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactorySensitiveMode.test.js`

**Interfaces:**
- Consumes: `sensitiveMode: 'ai'|'system'|'off'`.
- Preserves: book-level sparse inheritance and `restoreKeys` behavior.
- Produces: helper `normalizeSensitiveMode(settings)` used by unified, book, and task presentation.

- [ ] **Step 1: Add failing source and helper tests**

```js
assert.equal(normalizeSensitiveMode({}), 'system');
assert.equal(normalizeSensitiveMode({ sensitiveAIEnabled: true }), 'ai');
assert.equal(normalizeSensitiveMode({ sensitiveMode: 'off', sensitiveAIEnabled: true }), 'off');
```

Also assert that `UNIFIED_CONFIGURATION_KEYS` and `ENGINE_MODEL_OVERRIDE_FIELDS` contain `sensitiveMode`.

- [ ] **Step 2: Run the tests and verify failure**

Run: `node --test src/user/pages/shuihuo/batchFactorySensitiveMode.test.js src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js` from `frontend`
Expected: FAIL because the helper and three-choice field are absent.

- [ ] **Step 3: Add the three-choice unified control**

```jsx
<Segmented
  block
  value={normalizeSensitiveMode(value)}
  options={[
    { value: 'ai', label: 'AI优化' },
    { value: 'system', label: '系统优化' },
    { value: 'off', label: '关闭' }
  ]}
  onChange={sensitiveMode => patch({ sensitiveMode })}
/>
```

Show repair model and failure policy only when the value is `ai`. Keep explanatory copy that system optimization is local and free.

- [ ] **Step 4: Add the same inherited field to single-book settings**

Use the same three values. The field displays “继承：系统优化” when no book override exists and writes `restoreKeys: ['sensitiveMode']` when inheritance is restored.

- [ ] **Step 5: Stop new saves from writing the legacy boolean**

Remove `sensitiveAIEnabled` from new UI writes but keep it in read compatibility. Do not delete persisted legacy fields during unrelated saves.

- [ ] **Step 6: Run focused and full frontend tests**

Run: `node --test src/user/pages/shuihuo/batchFactorySensitiveMode.test.js src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js src/user/pages/shuihuo/batchFactoryProductionMode.test.js` from `frontend`
Expected: PASS.

- [ ] **Step 7: Commit the settings UI**

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx frontend/src/user/pages/shuihuo/batchFactoryBookConfigRegions.js frontend/src/user/pages/shuihuo/BatchFactorySettingsPrecedence.source.test.js frontend/src/user/pages/shuihuo/batchFactorySensitiveMode.test.js
git commit -m "feat: unify sensitive processing settings"
```

### Task 4: Verify compatibility across all modes

**Files:**
- Modify: `backend/internal/batchfactoryv11/settings_test.go`
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`

**Interfaces:**
- Verifies: old `sensitiveAIEnabled` reads as AI, new `sensitiveMode` wins, and text mode does not expose image/video model requirements.

- [ ] **Step 1: Add a settings precedence matrix test**

Cover batch `system`, book `ai`, restored book override, legacy boolean, and explicit `off`. Assert that resolving one book never mutates the batch patch.

- [ ] **Step 2: Add UI visibility assertions**

Assert the production switch contains `文本生成`, text visibility hides image/video controls, and the text model remains visible.

- [ ] **Step 3: Run verification**

Run: `go test ./internal/batchfactoryv11 ./internal/batchfactorysensitive -count=1` from `backend`
Run: `node --test src/user/pages/shuihuo/batchFactoryProductionMode.test.js src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js src/user/pages/shuihuo/batchFactorySensitiveMode.test.js` from `frontend`
Expected: PASS.

- [ ] **Step 4: Commit compatibility coverage**

```bash
git add backend/internal/batchfactoryv11/settings_test.go frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js
git commit -m "test: cover text mode and sensitive compatibility"
```

### Task 5: Run the shared sensitive stage before every outer workflow

**Files:**
- Create: `backend/internal/httpapi/batch_factory_sensitive_stage.go`
- Create: `backend/internal/httpapi/batch_factory_sensitive_stage_test.go`
- Modify: `backend/internal/httpapi/batch_factory_automation.go`
- Modify: `backend/internal/httpapi/batch_factory_automation_preflight.go`

**Interfaces:**
- Produces: `runBatchSensitiveStage(ctx, owner, batchID, bookID string, effective batchfactoryv11.SettingsPatch) (productionText string, trace batchfactorysensitive.Result, err error)`.
- Consumes: the shared processor from Task 2 and the complete saved original.
- Produces one production-text version consumed by video, comic, or text; outer workflows never call their own sensitive implementation.

- [ ] **Step 1: Write failing cross-mode tests**

For each mode `video`, `comic`, and `text`, assert `system` calls only rules, `ai` calls only the selected text model, and `off` preserves text. Assert AI failure stops the current book before any mode-specific paid stage runs.

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/httpapi -run BatchFactorySensitiveStage -count=1` from `backend`
Expected: FAIL because the common stage is absent.

- [ ] **Step 3: Add the shared stage adapter**

Read the complete original, resolve the effective mode/settings, call `batchfactorysensitive.Processor`, persist the production text and change trace, then pass that text to the selected outer workflow. Do not overwrite the complete original field.

- [ ] **Step 4: Add preflight requirements**

Require a usable text model only when `sensitiveMode == "ai"`. `system` and `off` must pass preflight without any model.

- [ ] **Step 5: Run all sensitive tests**

Run: `go test ./internal/batchfactorysensitive ./internal/httpapi -run 'Sensitive|BatchFactorySensitiveStage' -count=1` from `backend`
Expected: PASS.

- [ ] **Step 6: Commit the workflow integration**

```bash
git add backend/internal/httpapi/batch_factory_sensitive_stage.go backend/internal/httpapi/batch_factory_sensitive_stage_test.go backend/internal/httpapi/batch_factory_automation.go backend/internal/httpapi/batch_factory_automation_preflight.go
git commit -m "feat: share sensitive processing across production modes"
```

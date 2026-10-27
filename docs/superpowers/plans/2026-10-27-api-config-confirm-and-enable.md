# API Config Confirm And Enable Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every platform preset configuration use one “确定配置” action that saves, automatically enables, locks all fields, and requires disabling before editing.

**Architecture:** Keep the existing managed-model API as the source of truth. Add per-preset saving state in `ApiConfigPage`, split “confirm and enable” from “disable”, and derive field/button locking only from refreshed server state plus the current request state. Preserve existing credentials by omitting an empty credential on edits.

**Tech Stack:** React 18, Ant Design 5, Node test runner, Go embed local runtime.

## Global Constraints

- Do not create a Git branch.
- Do not reset, clean, or overwrite unrelated dirty workspace changes.
- Only change the platform preset cards; custom model dialogs remain unchanged.
- Button copy must be `确定配置`, `保存中`, and `配置已保存`.
- Missing-key Toast must be `请输入 API Key`.
- Locked-field Toast must be `请先关闭开关，再修改配置`.
- Successful confirmation Toast must be `配置已保存`.
- Enabled cards lock API Key, Relay URL, workflow ID, and model selection.
- Disabling a card restores editing only after the server confirms the change.

---

### Task 1: Specify the platform preset state contract

**Files:**
- Modify: `frontend/src/user/pages/ApiConfigPage.source.test.js`
- Modify: `frontend/src/user/pages/api-config-preset-state.js`
- Test: `frontend/src/user/pages/api-config-preset-state.test.js`

**Interfaces:**
- Consumes: preset metadata and server model records already used by `ApiConfigPage`.
- Produces: `getPresetCardState(preset, model, saving)` returning `{ configured, active, editable, buttonText, buttonDisabled }`.

- [ ] **Step 1: Write failing tests for all four UI states**

```js
assert.deepEqual(getPresetCardState(preset, null, false), {
  configured: false, active: false, editable: true,
  buttonText: '确定配置', buttonDisabled: false
});
assert.equal(getPresetCardState(preset, configuredModel, true).buttonText, '保存中');
assert.equal(getPresetCardState(preset, enabledModel, false).editable, false);
assert.equal(getPresetCardState(preset, enabledModel, false).buttonText, '配置已保存');
```

- [ ] **Step 2: Run the state tests and verify they fail because the helper is missing**

Run: `node --test frontend/src/user/pages/api-config-preset-state.test.js`

- [ ] **Step 3: Implement the small pure state helper**

```js
export function getPresetCardState(preset, model, saving = false) {
  const configured = hasPresetConfiguration(preset, model);
  const active = model?.enabled === true && configured;
  return {
    configured,
    active,
    editable: !active && !saving,
    buttonText: saving ? '保存中' : active ? '配置已保存' : '确定配置',
    buttonDisabled: active || saving
  };
}
```

- [ ] **Step 4: Run the focused tests and verify they pass**

Run: `node --test frontend/src/user/pages/api-config-preset-state.test.js`

- [ ] **Step 5: Commit the state contract**

```powershell
git add -- frontend/src/user/pages/api-config-preset-state.js frontend/src/user/pages/api-config-preset-state.test.js frontend/src/user/pages/ApiConfigPage.source.test.js
git commit -m "test(api-config): define preset confirmation states"
```

### Task 2: Implement confirm, auto-enable, disable, locking, and Toast feedback

**Files:**
- Modify: `frontend/src/user/pages/ApiConfigPage.jsx`
- Modify: `frontend/src/user/pages/ApiConfigPage.source.test.js`
- Modify: `frontend/src/shared/styles/account-center-visual-rebuild.css`

**Interfaces:**
- Consumes: `getPresetCardState(preset, model, saving)` from Task 1 and existing `createManagedModel` / `updateManagedModel` APIs.
- Produces: `confirmPreset(preset)`, `togglePreset(preset, enabled)`, per-card saving state keyed by preset ID, and a locked-field click surface.

- [ ] **Step 1: Add failing source assertions for the exact interaction contract**

```js
assert.match(source, /message\.warning\('请输入 API Key'\)/);
assert.match(source, /message\.info\('请先关闭开关，再修改配置'\)/);
assert.match(source, /message\.success\('配置已保存'\)/);
assert.match(source, /enabled:\s*true/);
assert.match(source, /loading=\{presetSavingId === preset\.id\}/);
assert.match(source, />确定配置</);
```

- [ ] **Step 2: Run source tests and verify the new assertions fail**

Run: `node --test frontend/src/user/pages/ApiConfigPage.source.test.js`

- [ ] **Step 3: Replace the shared saving boolean for preset cards with `presetSavingId`**

Add `const [presetSavingId, setPresetSavingId] = useState('');` and only set it around the current preset request.

- [ ] **Step 4: Implement confirmation with automatic enable**

`confirmPreset(preset)` validates the first credential, builds the full existing payload with `enabled: true`, omits a blank credential when a credential already exists, submits create/update, clears the plaintext Key only on success, refreshes models, and shows `配置已保存`.

- [ ] **Step 5: Implement explicit switch behavior**

`togglePreset(preset, true)` shows `请输入 API Key` and sends no request when the card is not configured. `togglePreset(preset, false)` submits only the existing model with `enabled: false`, refreshes server state on success, and reports an error without unlocking if the request fails.

- [ ] **Step 6: Lock every field while active or saving**

Pass `disabled={!cardState.editable}` to API Key, Relay URL, workflow ID, and model Select. Wrap the locked configuration controls in a click surface that shows `请先关闭开关，再修改配置` without exposing the saved Key.

- [ ] **Step 7: Render the unified action button**

Use `cardState.buttonText`, `presetSavingId === preset.id`, and `cardState.buttonDisabled`. Remove all visible `保存配置` copy from platform preset cards.

- [ ] **Step 8: Run focused tests and build**

Run:

```powershell
node --test frontend/src/user/pages/api-config-preset-state.test.js frontend/src/user/pages/ApiConfigPage.source.test.js
Set-Location frontend
npm run build
```

- [ ] **Step 9: Commit the implementation**

```powershell
git add -- frontend/src/user/pages/ApiConfigPage.jsx frontend/src/user/pages/ApiConfigPage.source.test.js frontend/src/user/pages/api-config-preset-state.js frontend/src/user/pages/api-config-preset-state.test.js frontend/src/shared/styles/account-center-visual-rebuild.css
git commit -m "feat(api-config): confirm and enable preset models"
```

### Task 3: Embed and verify the live local page

**Files:**
- Verify: `frontend/dist/**`
- Verify: `backend/web/dist/**`
- Verify: `backend/qiantie-local.exe`

**Interfaces:**
- Consumes: built frontend from Task 2.
- Produces: verified local behavior at `http://localhost:4001/api-config`.

- [ ] **Step 1: Build the Go-embedded runtime**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

- [ ] **Step 2: Restart only the local process listening on port 4001**

Verify the listener belongs to `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`, stop it, then run `scripts/start-local-go-review.ps1`.

- [ ] **Step 3: Run service and focused regression checks**

Run:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1
node --test frontend/src/user/pages/api-config-preset-state.test.js frontend/src/user/pages/ApiConfigPage.source.test.js
```

- [ ] **Step 4: Verify browser behavior without using a real API Key**

On `http://localhost:4001/api-config`, verify the four cards show `确定配置`, clicking an unconfigured switch shows `请输入 API Key`, and an existing enabled card has all configuration controls locked and shows `配置已保存`. Do not submit a paid provider request or replace a saved credential.

- [ ] **Step 5: Report the exact verified states and any untested credential-dependent path**

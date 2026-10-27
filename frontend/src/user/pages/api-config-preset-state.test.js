import assert from 'node:assert/strict';
import test from 'node:test';
import { getPresetCardState, hasPresetConfiguration } from './api-config-preset-state.js';

const ordinaryPreset = { id: 'yd2-mini-video', credentialMode: 'apiKey' };
const h3Preset = { id: 'minimax-h3-video', credentialMode: 'apiKey' };

test('an unconfigured preset is editable and waits for confirmation', () => {
  assert.equal(hasPresetConfiguration(ordinaryPreset, null), false);
  assert.deepEqual(getPresetCardState(ordinaryPreset, null, false), {
    configured: false,
    active: false,
    editable: true,
    buttonText: '确定配置',
    buttonDisabled: false
  });
});

test('only the current preset enters the saving state', () => {
  assert.deepEqual(getPresetCardState(ordinaryPreset, null, true), {
    configured: false,
    active: false,
    editable: false,
    buttonText: '保存中',
    buttonDisabled: true
  });
});

test('an enabled configured preset locks fields and reports saved', () => {
  const model = { enabled: true, hasCredential: true };
  assert.deepEqual(getPresetCardState(ordinaryPreset, model, false), {
    configured: true,
    active: true,
    editable: false,
    buttonText: '配置已保存',
    buttonDisabled: true
  });
});

test('a disabled configured preset is editable and can be confirmed again', () => {
  const model = { enabled: false, hasCredential: true };
  assert.deepEqual(getPresetCardState(ordinaryPreset, model, false), {
    configured: true,
    active: false,
    editable: true,
    buttonText: '确定配置',
    buttonDisabled: false
  });
});

test('H3 is not configured until both its credential and workflow exist', () => {
  assert.equal(hasPresetConfiguration(h3Preset, { hasCredential: true, workflowId: '' }), false);
  assert.equal(hasPresetConfiguration(h3Preset, { hasCredential: true, workflowId: 'workflow-1' }), true);
});

import assert from 'node:assert/strict';
import test from 'node:test';
import { getPresetCardState, hasPresetConfiguration } from './api-config-preset-state.js';

const ordinaryPreset = { id: 'yd2-mini-video', credentialMode: 'apiKey' };
const h3Preset = { id: 'minimax-h3-video', credentialMode: 'apiKey' };

test('an unconfigured preset stays editable so a key can be filled in', () => {
  assert.equal(hasPresetConfiguration(ordinaryPreset, null), false);
  assert.deepEqual(getPresetCardState(ordinaryPreset, null, false), {
    configured: false,
    editable: true
  });
});

test('only the current preset enters the saving state', () => {
  assert.deepEqual(getPresetCardState(ordinaryPreset, null, true), {
    configured: false,
    editable: false
  });
});

test('a configured preset stays editable: clearing the key is the way to turn it off', () => {
  const model = { enabled: true, hasCredential: true };
  assert.deepEqual(getPresetCardState(ordinaryPreset, model, false), {
    configured: true,
    editable: true
  });
});

test('the legacy enabled flag no longer affects editability', () => {
  const model = { enabled: false, hasCredential: true };
  assert.deepEqual(getPresetCardState(ordinaryPreset, model, false), {
    configured: true,
    editable: true
  });
});

test('H3 is not configured until both its credential and workflow exist', () => {
  assert.equal(hasPresetConfiguration(h3Preset, { hasCredential: true, workflowId: '' }), false);
  assert.equal(hasPresetConfiguration(h3Preset, { hasCredential: true, workflowId: 'workflow-1' }), true);
});

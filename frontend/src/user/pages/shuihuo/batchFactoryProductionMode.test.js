import test from 'node:test';
import assert from 'node:assert/strict';
import * as modes from './batchFactoryProductionMode.js';
import { automationPresetMode, automationPresetModeLabel, automationPresetsForMode, withAutomationPresetMode } from './batchFactoryAutomationSchedule.js';
import {
  batchFactoryProductionMode,
  batchFactoryProductionModeVisibility,
  batchFactoryShouldLoadPublishEnvironment,
  batchFactoryShouldPrepareAudio,
  batchFactoryShouldRecompileVideoPrompts
} from './batchFactoryProductionMode.js';

test('defaults historical settings to video mode', () => {
  assert.equal(batchFactoryProductionMode(undefined), 'video');
  assert.deepEqual(batchFactoryProductionModeVisibility('unknown'), {
    mode: 'video',
    video: true,
    comic: false,
    text: false
  });
});

test('new text batches seed page, capture and aspect without overriding draft edits', () => {
  assert.equal(typeof modes.batchCreationSettings, 'function');
  assert.deepEqual(modes.batchCreationSettings({ productionMode: 'text' }), {
    productionMode: 'text', textPageTarget: 150, contentRange: 4000, contentCaptureCharacters: 4000, imageAspectRatio: '3:4'
  });
  const draft = { productionMode: 'text', textPageTarget: 205, contentRange: 8000, imageAspectRatio: '9:16' };
  assert.deepEqual(modes.batchCreationSettings(draft, 1000), { ...draft, contentCaptureCharacters: 8000 });
  assert.equal(modes.batchCreationSettings({ productionMode: 'text', contentCaptureCharacters: 2000 }).contentRange, 2000);
  assert.equal(modes.batchCreationSettings({ productionMode: 'text' }, 12000).contentRange, 12000);
  for (const productionMode of ['video', 'comic']) {
    assert.deepEqual(modes.batchCreationSettings({ productionMode }), { productionMode });
  }
});

test('text presets remain isolated while historical generic presets remain selectable', () => {
  const textPreset = { id: 'text', config: withAutomationPresetMode({ textPageTarget: 240 }, 'text') };
  const presets = [textPreset, { id: 'generic', config: {} }, { id: 'comic', config: { productionMode: 'comic' } }];
  assert.equal(automationPresetMode(textPreset), 'text');
  assert.equal(automationPresetModeLabel('text'), '文本');
  assert.deepEqual(automationPresetsForMode(presets, 'text').map(preset => preset.id), ['text', 'generic']);
  assert.deepEqual(automationPresetsForMode(presets, 'comic').map(preset => preset.id), ['generic', 'comic']);
});

test('video publishing environment is checked only while its tab is visible', () => {
  assert.equal(batchFactoryShouldLoadPublishEnvironment({ open: true, activeTab: 'models', productionMode: 'video' }), false);
  assert.equal(batchFactoryShouldLoadPublishEnvironment({ open: true, activeTab: 'publish', productionMode: 'comic' }), false);
  assert.equal(batchFactoryShouldLoadPublishEnvironment({ open: true, activeTab: 'publish', productionMode: 'video' }), true);
  assert.equal(batchFactoryShouldLoadPublishEnvironment({ open: false, activeTab: 'publish', productionMode: 'video' }), false);
});

test('comic mode exposes comic settings and hides video settings', () => {
  assert.deepEqual(batchFactoryProductionModeVisibility('comic'), {
    mode: 'comic',
    video: false,
    comic: true,
    text: false
  });
});

test('text mode exposes text settings and hides video and comic settings', () => {
  assert.deepEqual(batchFactoryProductionModeVisibility('text'), {
    mode: 'text',
    video: false,
    comic: false,
    text: true
  });
});

test('comic unified settings never trigger video audio measurement or video prompt recompilation', () => {
  const comicSettings = { productionMode: 'comic', audioPlanningEnabled: true };
  assert.equal(batchFactoryShouldPrepareAudio(comicSettings), false);
  assert.equal(batchFactoryShouldRecompileVideoPrompts(comicSettings), false);
  assert.equal(batchFactoryShouldPrepareAudio({ productionMode: 'video', audioPlanningEnabled: true }), true);
  assert.equal(batchFactoryShouldRecompileVideoPrompts({ productionMode: 'video' }), true);
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { automationPresetMode, automationPresetModeLabel, automationPresetSnapshot, automationPresetsForMode, formatBeijingDatetimeLocal, parseBeijingDatetimeLocal, withAutomationPresetMode } from './batchFactoryAutomationSchedule.js';

test('converts a Beijing wall-clock scheduling value to a UTC instant', () => {
  assert.equal(parseBeijingDatetimeLocal('2026-09-25T03:00'), '2026-09-24T19:00:00.000Z');
});

test('formats a persisted UTC instant for the Beijing scheduling control', () => {
  assert.equal(formatBeijingDatetimeLocal('2026-09-24T19:00:00.000Z'), '2026-09-25T03:00');
});

test('rejects malformed scheduling values instead of silently using browser local time', () => {
  assert.equal(parseBeijingDatetimeLocal('2026/09/25 03:00'), '');
  assert.equal(formatBeijingDatetimeLocal('not-an-instant'), '');
});

test('freezes the selected automation preset configuration before a giant batch starts', () => {
  const presets = [{ id: 'preset-1', name: '全自动', version: 2, config: { textModelId: 'text-1', publishSettings: { organization: 'org-a' } } }];
  const snapshot = automationPresetSnapshot(presets, 'preset-1');
  presets[0].config.publishSettings.organization = 'changed-after-selection';

  assert.deepEqual(snapshot, {
    id: 'preset-1', name: '全自动', version: 2,
    config: { textModelId: 'text-1', publishSettings: { organization: 'org-a' } }
  });
  assert.equal(automationPresetSnapshot(presets, 'missing'), null);
});

test('classifies automation presets by production mode with legacy presets as universal', () => {
  assert.equal(automationPresetMode({ config: { productionMode: 'comic' } }), 'comic');
  assert.equal(automationPresetMode({ config: { productionMode: 'video' } }), 'video');
  // 旧预设没有 productionMode：两处下拉都可见，标记为通用。
  assert.equal(automationPresetMode({ config: {} }), '');
  assert.equal(automationPresetMode({}), '');
  assert.equal(automationPresetMode({ config: { productionMode: 'other' } }), '');
  assert.equal(automationPresetModeLabel('comic'), '漫画');
  assert.equal(automationPresetModeLabel('video'), '视频');
  assert.equal(automationPresetModeLabel(''), '通用');
});

test('filters automation presets to the batch mode while keeping universal presets visible', () => {
  const presets = [
    { id: 'comic-1', config: { productionMode: 'comic' } },
    { id: 'video-1', config: { productionMode: 'video' } },
    { id: 'legacy-1', config: {} }
  ];
  assert.deepEqual(automationPresetsForMode(presets, 'video').map(item => item.id), ['video-1', 'legacy-1']);
  assert.deepEqual(automationPresetsForMode(presets, 'comic').map(item => item.id), ['comic-1', 'legacy-1']);
  // 无模式上下文（巨量素材导入等场景）返回全部预设。
  assert.deepEqual(automationPresetsForMode(presets, '').map(item => item.id), ['comic-1', 'video-1', 'legacy-1']);
});

test('stamps the current batch production mode into a preset config without touching an empty mode', () => {
  assert.deepEqual(withAutomationPresetMode({ textModelId: 'text-1' }, 'comic'), { textModelId: 'text-1', productionMode: 'comic' });
  assert.deepEqual(withAutomationPresetMode({ textModelId: 'text-1', productionMode: 'video' }, 'comic'), { textModelId: 'text-1', productionMode: 'comic' });
  // mode 为空（无模式上下文）保持原 config，不写入 productionMode。
  assert.deepEqual(withAutomationPresetMode({ textModelId: 'text-1' }, ''), { textModelId: 'text-1' });
  // 不修改原对象。
  const original = { productionMode: 'video' };
  withAutomationPresetMode(original, 'comic');
  assert.equal(original.productionMode, 'video');
});

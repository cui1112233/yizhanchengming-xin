import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ApiConfigPage.jsx'), 'utf8');

test('平台预设卡没有启用开关：填了 Key 就按生效处理', () => {
  assert.match(source, /enabled: !clearing/);
  assert.match(source, /else if \(clearing\) payload\.clearCredential = true/);
  assert.doesNotMatch(source, /togglePreset\(/);
  assert.doesNotMatch(source, /<Switch/);
});

test('预设卡不再有锁定遮罩与“先关闭开关”提示', () => {
  assert.doesNotMatch(source, /ac-platform-preset-lock/);
  assert.doesNotMatch(source, /请先关闭开关/);
});

test('未保存时输入框保持可编辑，保存中才禁用', () => {
  assert.match(source, /getPresetCardState\(preset, model, presetSavingId === preset\.id\)/);
  assert.match(source, /disabled=\{!cardState\.editable\}/);
});

test('文案明确说明清空 Key 等于停用，不依赖额外开关', () => {
  assert.match(source, /填好 API Key 后立即生效/);
  assert.match(source, /清空 Key 即从下拉移除/);
});

test('目录模型不再向三组传启用回调', () => {
  assert.doesNotMatch(source, /onToggleModelEnabled|modelTogglingId/);
});

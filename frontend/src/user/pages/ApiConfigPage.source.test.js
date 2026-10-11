import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ApiConfigPage.jsx'), 'utf8');

test('平台预设模型必须先确认配置后才能启用', () => {
  assert.match(source, /function isPresetConfigured\(preset, model\)/);
  assert.match(source, /disabled=\{!model\?\.enabled && !isPresetConfigured\(preset, model\)\}/);
  assert.match(source, /保存配置/);
  assert.match(source, /savePreset\(preset, false\)/);
});

test('H3 保存配置必须带当前 key，并和 workflow 一起提交', () => {
  assert.ok(source.includes("if (preset.id === 'minimax-h3-video' && !key)"));
  assert.ok(source.includes('工作流 ID 会和 Key 一起保存'));
  assert.ok(source.includes('payload.workflowId = workflowId || DEFAULT_H3_WORKFLOW_ID'));
});

test('H3 启用必须基于刷新后的服务端凭据和 workflow', () => {
  assert.match(source, /hasCredential\s*!==\s*true/);
  assert.ok(source.includes("Boolean(String(model.workflowId || '').trim())"));
  assert.doesNotMatch(source, /preset\.id === 'minimax-h3-video' \? '保存工作流' : '保存配置'/);
});

test('H3 保存配置必须同时启用，确保输入工作流后默认调用', () => {
  assert.match(source, /preset\.id === 'minimax-h3-video' \? savePreset\(preset, true\) : savePreset\(preset, false\)/);
  assert.match(source, /保存配置并启用/);
});

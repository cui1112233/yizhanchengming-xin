import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ApiConfigPage.jsx'), 'utf8');

test('平台预设模型统一使用确定配置并在保存成功后自动启用', () => {
  assert.match(source, /enabled:\s*true/);
  assert.match(source, /message\.success\('配置已保存'\)/);
  assert.match(source, /presetSavingId === preset\.id/);
  assert.match(source, /cardState\.buttonText/);
  assert.doesNotMatch(source, />保存配置</);
});

test('H3 确定配置必须校验 key，并和 workflow 一起提交', () => {
  assert.ok(source.includes("if (preset.id === 'minimax-h3-video' && !key && existing?.hasCredential !== true)"));
  assert.ok(source.includes('payload.workflowId = workflowId || DEFAULT_H3_WORKFLOW_ID'));
});

test('未配置直接点击开关只提示填写 API Key', () => {
  assert.match(source, /message\.warning\('请输入 API Key'\)/);
  assert.match(source, /onChange=\{enabled => togglePreset\(preset, enabled\)\}/);
});

test('启用后锁定所有预设配置并提示先关闭开关', () => {
  assert.match(source, /message\.info\('请先关闭开关，再修改配置'\)/);
  assert.match(source, /disabled=\{!cardState\.editable\}/);
  assert.match(source, /ac-platform-preset-lock/);
});

test('已保存 H3 凭据后允许不重复填写 Key 直接启用', () => {
  assert.match(source, /preset\.id === 'minimax-h3-video' && !key && existing\?\.hasCredential !== true/);
});

test('does not show a video model as enabled until its required configuration is complete', () => {
  assert.match(source, /getPresetCardState\(preset, model, presetSavingId === preset\.id\)/);
  assert.match(source, /checked=\{cardState\.active\}/);
});

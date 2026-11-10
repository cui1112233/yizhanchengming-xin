import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const file = new URL('./AICutAutoEditDrawer.jsx', import.meta.url);
const source = fs.existsSync(file) ? fs.readFileSync(file, 'utf8') : '';

test('AI自动剪辑抽屉具备完整三步审核文案', () => {
  for (const copy of ['AI 自动剪辑', '选择素材', '剪辑目标', '目标时长', '字幕', '转场', '滤镜', '背景音乐', '生成剪辑草案', '预计成片时长', '操作理由', '应用到时间线', '取消草案', '草案已过期，请基于最新工程重新生成']) assert.match(source, new RegExp(copy));
});

test('AI自动剪辑生成和取消不直接修改时间线', () => {
  assert.match(source, /createAICutEditPlan/);
  assert.match(source, /cancelAICutEditPlan/);
  assert.match(source, /applyAICutEditPlan/);
  assert.match(source, /plan\?\.state !== 'ready'/);
  assert.match(source, /enabledOperationIds/);
  assert.match(source, /onApplied\?\./);
  assert.doesNotMatch(source, /onApplied\?\.\([^)]*created/);
  assert.doesNotMatch(source, /onApplied\?\.\([^)]*cancel/);
});

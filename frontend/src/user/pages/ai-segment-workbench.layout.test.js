import test from 'node:test';
import assert from 'node:assert/strict';
import { aiSegmentModeContract, defaultAISegmentMode, visibleAISegmentModes } from './ai-segment-workbench.js';

test('AI分段四种模式使用洛水对应的小按钮名称和条件区块', () => {
  assert.deepEqual(Object.keys(aiSegmentModeContract), ['subtitle', 'main-track', 'text-track', 'ai']);
  assert.equal(aiSegmentModeContract.subtitle.label, '按字幕切割');
  assert.equal(aiSegmentModeContract['main-track'].label, '按剪映主轨道');
  assert.equal(aiSegmentModeContract['text-track'].label, '按剪映文本轨道');
  assert.equal(aiSegmentModeContract.ai.label, 'AI分段');

  assert.deepEqual(aiSegmentModeContract.subtitle.sections, ['video', 'srt', 'output', 'duration', 'quality', 'progress', 'log']);
  assert.deepEqual(aiSegmentModeContract['main-track'].sections, ['video', 'draft', 'output', 'duration', 'quality', 'progress', 'log']);
  assert.deepEqual(aiSegmentModeContract['text-track'].sections, ['video', 'draft', 'text-track', 'output', 'quality', 'progress', 'log']);
  assert.deepEqual(aiSegmentModeContract.ai.sections, ['model', 'prompt', 'video', 'draft', 'script', 'analysis', 'output', 'quality', 'progress', 'log']);
});

test('当前洛水版本隐藏旧字幕模式并默认进入剪映主轨道', () => {
  assert.equal(defaultAISegmentMode, 'main-track');
  assert.equal(aiSegmentModeContract.subtitle.hidden, true);
  assert.deepEqual(visibleAISegmentModes.map(item => item.label), ['按剪映主轨道', '按剪映文本轨道', 'AI分段']);
});

test('AI分段原版规则固定AI模式导出质量并保留其他模式设置', () => {
  assert.deepEqual(aiSegmentModeContract.ai.quality, { resolution: '1080p', fps: 30, locked: true });
  assert.deepEqual(aiSegmentModeContract.subtitle.quality, { resolution: '1080p', fps: 30, locked: false });
  assert.equal(aiSegmentModeContract.ai.primaryAction, 'AI 分析切点');
  assert.equal(aiSegmentModeContract.ai.exportAction, '开始按AI切点分段');
  assert.equal(aiSegmentModeContract['main-track'].exportAction, '开始分段');
});

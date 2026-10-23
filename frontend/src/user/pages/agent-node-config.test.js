import test from 'node:test';
import assert from 'node:assert/strict';
import { formatNodeParameterSummary, normalizeNodeConfig, patchNodeConfig } from './agent-node-config.js';

test('图片节点使用批准过的默认摘要并跟随参数变化', () => {
  const config = normalizeNodeConfig('image', {});
  assert.equal(formatNodeParameterSummary('image', config), '16:9 · 标准画质 · 2K · 1张');
  const changed = patchNodeConfig('image', config, { parameters: { ratio: '9:16', count: 2 } });
  assert.equal(formatNodeParameterSummary('image', changed), '9:16 · 标准画质 · 2K · 2张');
  assert.equal(changed.ratio, '9:16');
  assert.equal(changed.count, 2);
});

test('旧版平铺设置会升级为每个节点独立的参数但仍保留引用', () => {
  const config = normalizeNodeConfig('video', {
    model: 'video-a', ratio: '9:16', resolution: '1080p', durationSeconds: 10,
    referenceAssets: [{ id: 'image-1', kind: 'image' }]
  });
  assert.equal(config.modelId, 'video-a');
  assert.deepEqual(config.parameters, {
    ratio: '9:16', resolution: '1080P', durationSeconds: 10, count: 1, generateAudio: false
  });
  assert.deepEqual(config.referenceAssets.map(item => item.id), ['image-1']);
  assert.equal(formatNodeParameterSummary('video', config), '9:16 · 1080P · 10s · 1个');
});

test('旧版豆包本地图片模型会迁移到当前执行器和模型标识', () => {
  const config = normalizeNodeConfig('image', { modelId: 'seedream-5-flash-local', prompt: '雨夜霓虹街道' });
  assert.equal(config.provider, 'doubao_local_executor');
  assert.equal(config.modelId, 'Seedream 5.0 Flash');
  assert.equal(config.model, 'Seedream 5.0 Flash');
});

test('文字和音频节点不会获得图片或视频引擎字段', () => {
  const text = normalizeNodeConfig('text', { provider: 'personal_api' });
  const audio = normalizeNodeConfig('audio', {});
  assert.equal(formatNodeParameterSummary('text', text), '生成 · 标准长度 · 均衡');
  assert.equal(formatNodeParameterSummary('audio', audio), '语音 · 1x · MP3');
  assert.equal(Object.hasOwn(text, 'imageProvider'), false);
  assert.equal(Object.hasOwn(text, 'videoProvider'), false);
});

test('更新嵌套设置不修改原对象', () => {
  const original = normalizeNodeConfig('text', { prompt: '旧内容' });
  const changed = patchNodeConfig('text', original, { parameters: { creativity: 'creative' }, prompt: '新内容' });
  assert.equal(original.prompt, '旧内容');
  assert.equal(original.parameters.creativity, 'balanced');
  assert.equal(changed.prompt, '新内容');
  assert.equal(changed.parameters.creativity, 'creative');
});

import test from 'node:test';
import assert from 'node:assert/strict';

import {
  imageGenerationRequestSettings,
  imageNodeSettingsSummary,
  normalizeImageNodeSettings
} from './agent-canvas-image-settings.js';

test('图片节点默认摘要精确显示比例画质清晰度和数量', () => {
  const settings = normalizeImageNodeSettings({});
  assert.deepEqual(settings, {
    ratio: '16:9',
    quality: 'standard',
    resolution: '2K',
    count: 1
  });
  assert.equal(imageNodeSettingsSummary(settings), '16:9 · 标准画质 · 2K · 1张');
});

test('图片设置变化后摘要实时使用当前节点配置', () => {
  assert.equal(imageNodeSettingsSummary({
    ratio: '9:16', quality: 'high', resolution: '4K', count: 3
  }), '9:16 · 高清画质 · 4K · 3张');
});

test('真实图片生成请求携带比例画质清晰度和数量', () => {
  assert.deepEqual(imageGenerationRequestSettings({
    ratio: '1:1', quality: 'high', resolution: '4K', count: 2
  }), {
    aspectRatio: '1:1',
    quality: 'high',
    resolution: '4K',
    count: 2,
    size: '4096x4096'
  });
});

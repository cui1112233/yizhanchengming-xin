import test from 'node:test';
import assert from 'node:assert/strict';
import { keyframeEasingPresets, normalizeKeyframeEasing, sampleKeyframeEasing } from './aicutKeyframeEasing.js';

test('关键帧缓动提供线性、缓入、缓出、缓入缓出、定格和自定义曲线', () => {
  assert.deepEqual(keyframeEasingPresets.map(item => item.id), ['linear', 'ease-in', 'ease-out', 'ease-in-out', 'hold', 'custom']);
  assert.deepEqual(normalizeKeyframeEasing({ type: 'custom', x1: .2, y1: .1, x2: .8, y2: .9 }), { type: 'custom', x1: .2, y1: .1, x2: .8, y2: .9 });
  assert.deepEqual(normalizeKeyframeEasing({ type: 'custom', x1: -1, y1: 2, x2: 4, y2: -3 }), { type: 'custom', x1: 0, y1: 1, x2: 1, y2: 0 });
  assert.equal(normalizeKeyframeEasing({ type: 'unknown' }).type, 'linear');
});

test('关键帧缓动在预览与导出共用固定采样规则', () => {
  assert.equal(sampleKeyframeEasing({ type: 'linear' }, 0), 0);
  assert.equal(sampleKeyframeEasing({ type: 'linear' }, .5), .5);
  assert.equal(sampleKeyframeEasing({ type: 'linear' }, 1), 1);
  assert.equal(sampleKeyframeEasing({ type: 'hold' }, .99), 0);
  assert.equal(sampleKeyframeEasing({ type: 'hold' }, 1), 1);
  assert.ok(sampleKeyframeEasing({ type: 'ease-in' }, .5) < .5);
  assert.ok(sampleKeyframeEasing({ type: 'ease-out' }, .5) > .5);
  assert.ok(Math.abs(sampleKeyframeEasing({ type: 'ease-in-out' }, .5) - .5) < .02);
});

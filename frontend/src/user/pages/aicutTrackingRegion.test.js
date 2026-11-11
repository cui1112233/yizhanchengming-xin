import test from 'node:test';
import assert from 'node:assert/strict';
import { drawTrackingRegion, moveTrackingRegion, normalizeTrackingRegion, resizeTrackingRegion } from './aicutTrackingRegion.js';

test('跟踪框始终留在画布内并保留最小尺寸', () => {
  assert.deepEqual(normalizeTrackingRegion({ centerX: -1, centerY: 2, width: .01, height: 4 }), {
    centerX: .025,
    centerY: .5,
    width: .05,
    height: 1
  });
});

test('拖动跟踪框会按画布比例移动并在边缘停止', () => {
  const origin = { centerX: .5, centerY: .5, width: .4, height: .2 };
  assert.deepEqual(moveTrackingRegion(origin, .2, -.15), { centerX: .7, centerY: .35, width: .4, height: .2 });
  assert.deepEqual(moveTrackingRegion(origin, 2, 2), { centerX: .8, centerY: .9, width: .4, height: .2 });
});

test('从角点缩放会固定对角并限制最小尺寸', () => {
  const origin = { centerX: .5, centerY: .5, width: .4, height: .4 };
  assert.deepEqual(resizeTrackingRegion(origin, 'se', .1, .2), { centerX: .55, centerY: .6, width: .5, height: .6 });
  assert.deepEqual(resizeTrackingRegion(origin, 'nw', .5, .5), { centerX: .675, centerY: .675, width: .05, height: .05 });
});

test('空白处拖拽可直接画出目标框并兼容反向拖拽', () => {
  assert.deepEqual(drawTrackingRegion({ x: .8, y: .7 }, { x: .2, y: .1 }), {
    centerX: .5,
    centerY: .4,
    width: .6,
    height: .6
  });
});

import assert from 'node:assert/strict';
import test from 'node:test';
import { dragTransform, normalizeTransform, resizeTransform, rotateTransform } from './canvasTransform.js';

test('变换非法值会回退并限制缩放、旋转和透明度', () => {
  assert.deepEqual(normalizeTransform({ x: 'bad', y: 2, scaleX: -8, scaleY: 99, rotation: 540, opacity: -1 }), {
    x: 0, y: 1, scaleX: 0.05, scaleY: 20, rotation: 180, opacity: 0,
  });
});

test('拖动画面按真实画布尺寸换算为归一化位置', () => {
  assert.deepEqual(dragTransform({ x: 0, y: 0 }, { x: 100, y: -50 }, { width: 1000, height: 500 }), {
    x: 0.1, y: -0.1, scaleX: 1, scaleY: 1, rotation: 0, opacity: 1,
  });
});

test('四角缩放不会产生负尺寸并可保持比例', () => {
  const result = resizeTransform({ scaleX: 1, scaleY: 1 }, { x: -2000, y: -500 }, { width: 1000, height: 500 }, { lockRatio: true });
  assert.equal(result.scaleX, 0.05);
  assert.equal(result.scaleY, 0.05);
});

test('旋转以画布中心计算并规范到正负一百八十度', () => {
  const result = rotateTransform({ rotation: 170 }, { x: 0, y: 100 }, { x: 100, y: 0 }, { x: 0, y: 0 });
  assert.equal(result.rotation, 80);
});

test('拖动超出安全画布时位置被限制在可恢复范围', () => {
  const result = dragTransform({}, { x: 9000, y: -9000 }, { width: 1000, height: 500 });
  assert.equal(result.x, 1);
  assert.equal(result.y, -1);
});

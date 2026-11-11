import assert from 'node:assert/strict';
import test from 'node:test';
import { lassoGestureMoved, nextLassoAutoScrollLeft, normalizeLassoRect, selectLassoClipKeys } from './aicutLasso.js';

test('框选矩形支持从任意方向拖动并归一为左上角', () => {
  assert.deepEqual(normalizeLassoRect({ x: 120, y: 90 }, { x: 20, y: 10 }), { left: 20, top: 10, width: 100, height: 80, right: 120, bottom: 90 });
});

test('框选选择所有相交的未锁定片段并展开复合片段成员', () => {
  const items = [
    { rect: { left: 10, top: 10, right: 110, bottom: 50 }, keys: ['v::a'], locked: false },
    { rect: { left: 120, top: 10, right: 220, bottom: 50 }, keys: ['v::b', 'a::b'], locked: false },
    { rect: { left: 20, top: 60, right: 100, bottom: 100 }, keys: ['a::locked'], locked: true },
    { rect: { left: 400, top: 10, right: 450, bottom: 50 }, keys: ['v::far'], locked: false }
  ];
  assert.deepEqual(selectLassoClipKeys(items, { left: 50, top: 0, right: 180, bottom: 80 }), ['v::a', 'v::b', 'a::b']);
});

test('框选空白、边缘接触和轻微点击都有明确边界', () => {
  const item = { rect: { left: 20, top: 20, right: 40, bottom: 40 }, keys: ['v::a'], locked: false };
  assert.deepEqual(selectLassoClipKeys([item], { left: 40, top: 0, right: 60, bottom: 20 }), []);
  assert.deepEqual(selectLassoClipKeys([item], { left: 100, top: 100, right: 120, bottom: 120 }), []);
  assert.equal(lassoGestureMoved({ x: 10, y: 10 }, { x: 13, y: 12 }), false);
  assert.equal(lassoGestureMoved({ x: 10, y: 10 }, { x: 16, y: 10 }), true);
});

test('框选靠近时间线左右边缘时按距离自动滚动', () => {
  assert.equal(nextLassoAutoScrollLeft({ clientX: 1040, viewportLeft: 100, viewportRight: 1100, scrollLeft: 600, scrollWidth: 4200, viewportWidth: 1000 }), 606);
  assert.equal(nextLassoAutoScrollLeft({ clientX: 1095, viewportLeft: 100, viewportRight: 1100, scrollLeft: 600, scrollWidth: 4200, viewportWidth: 1000 }), 623);
  assert.equal(nextLassoAutoScrollLeft({ clientX: 110, viewportLeft: 100, viewportRight: 1100, scrollLeft: 600, scrollWidth: 4200, viewportWidth: 1000 }), 579);
  assert.equal(nextLassoAutoScrollLeft({ clientX: 600, viewportLeft: 100, viewportRight: 1100, scrollLeft: 600, scrollWidth: 4200, viewportWidth: 1000 }), 600);
});

test('框选自动滚动会夹取到时间线首尾并拒绝坏尺寸', () => {
  assert.equal(nextLassoAutoScrollLeft({ clientX: 0, viewportLeft: 100, viewportRight: 1100, scrollLeft: 5, scrollWidth: 4200, viewportWidth: 1000 }), 0);
  assert.equal(nextLassoAutoScrollLeft({ clientX: 1200, viewportLeft: 100, viewportRight: 1100, scrollLeft: 3195, scrollWidth: 4200, viewportWidth: 1000 }), 3200);
  assert.equal(nextLassoAutoScrollLeft({ clientX: 200, viewportLeft: 100, viewportRight: 100, scrollLeft: 80, scrollWidth: 0, viewportWidth: 0 }), 80);
});

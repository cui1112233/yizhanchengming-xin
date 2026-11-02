import test from 'node:test';
import assert from 'node:assert/strict';
import {
  STORYBOARD_SPLIT_DEFAULTS,
  storyboardSplitRatio,
  storyboardSplitRatioFromPointer
} from './agent-storyboard-split.js';

test('故事板分隔比例保持默认值，并限制为可用范围', () => {
  assert.deepEqual(STORYBOARD_SPLIT_DEFAULTS, {
    director: 31,
    column: 25,
    leftRow: 66,
    rightRow: 55
  });
  assert.equal(storyboardSplitRatio(-4, 24), 24);
  assert.equal(storyboardSplitRatio(98, 24), 76);
  assert.equal(storyboardSplitRatioFromPointer(200, { left: 0, width: 400 }, 'x'), 50);
  assert.equal(storyboardSplitRatioFromPointer(120, { top: 0, height: 400 }, 'y'), 30);
});

test('每条分隔线可以按自己的默认值复位', () => {
  assert.equal(storyboardSplitRatio(undefined, 18, STORYBOARD_SPLIT_DEFAULTS.director), 31);
  assert.equal(storyboardSplitRatio(undefined, 18, STORYBOARD_SPLIT_DEFAULTS.column), 25);
  assert.equal(storyboardSplitRatio(undefined, 18, STORYBOARD_SPLIT_DEFAULTS.leftRow), 66);
  assert.equal(storyboardSplitRatio(undefined, 18, STORYBOARD_SPLIT_DEFAULTS.rightRow), 55);
});

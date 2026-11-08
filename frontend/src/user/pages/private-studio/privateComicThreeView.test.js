import assert from 'node:assert/strict';
import test from 'node:test';

import { privateComicAssetUseThreeView } from './privateComicThreeView.js';

test('私人定制只有人物资产按开关使用三视图', () => {
  assert.equal(privateComicAssetUseThreeView('character', true), true);
  assert.equal(privateComicAssetUseThreeView('character', false), false);
  assert.equal(privateComicAssetUseThreeView('scene', true), false);
  assert.equal(privateComicAssetUseThreeView('prop', true), false);
});

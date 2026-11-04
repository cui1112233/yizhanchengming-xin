import test from 'node:test';
import assert from 'node:assert/strict';

import {
  IMAGE_ASPECT_RATIOS,
  IMAGE_ASPECT_RATIO_OPTIONS,
  isSupportedImageAspectRatio,
  isPortraitImageAspectRatio,
  normalizeImageAspectRatio,
  imageAspectRatioOrientation,
} from './imageAspectRatios.js';

test('image aspect ratio list keeps the legacy values and adds four new ratios', () => {
  assert.deepEqual(IMAGE_ASPECT_RATIOS, ['16:9', '9:16', '1:1', '4:3', '3:4', '3:2', '2:3']);
  for (const legacy of ['16:9', '9:16', '1:1']) {
    assert.equal(isSupportedImageAspectRatio(legacy), true);
  }
  for (const added of ['2:3', '3:4', '4:3', '3:2']) {
    assert.equal(isSupportedImageAspectRatio(added), true);
    assert.ok(IMAGE_ASPECT_RATIO_OPTIONS.some((option) => option.value === added));
  }
});

test('unknown stored ratios fall back without throwing', () => {
  assert.equal(isSupportedImageAspectRatio('7:5'), false);
  assert.equal(normalizeImageAspectRatio('7:5'), '9:16');
  assert.equal(normalizeImageAspectRatio(''), '9:16');
  assert.equal(normalizeImageAspectRatio('2:3'), '2:3');
});

test('portrait and landscape ratios stay distinct', () => {
  for (const ratio of ['2:3', '3:4', '9:16']) {
    assert.equal(isPortraitImageAspectRatio(ratio), true);
    assert.equal(imageAspectRatioOrientation(ratio), '竖幅构图');
  }
  for (const ratio of ['4:3', '3:2', '16:9']) {
    assert.equal(isPortraitImageAspectRatio(ratio), false);
    assert.equal(imageAspectRatioOrientation(ratio), '横幅构图');
  }
  assert.equal(imageAspectRatioOrientation('1:1'), '方形构图');
});

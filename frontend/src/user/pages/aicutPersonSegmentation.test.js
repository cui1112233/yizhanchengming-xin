import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildPersonSegmentationSignature,
  isPersonSegmentationCurrent,
  nextPersonSegmentationTimestamp,
  normalizePersonSegmentation,
  personMaskToRGBA
} from './aicutPersonSegmentation.js';

test('人物模型重复运行时始终使用严格递增的帧时间戳', () => {
  assert.equal(nextPersonSegmentationTimestamp(578001, 4000), 578002);
  assert.equal(nextPersonSegmentationTimestamp(578001, 600000), 600000);
  assert.equal(nextPersonSegmentationTimestamp(-1, 0), 0);
});

test('人物抠像设置会规范阈值、柔化和蒙版素材信息', () => {
  assert.deepEqual(normalizePersonSegmentation({ enabled: 1, matteAssetId: ' matte-1 ', threshold: 4, feather: -1, signature: ' sig ' }), {
    enabled: true,
    matteAssetId: 'matte-1',
    threshold: 1,
    feather: 0,
    signature: 'sig'
  });
});

test('人物蒙版签名随裁剪区间、时长、倍速、倒放和定格变化', () => {
  const clip = { assetId: 'video-1', sourceStartMs: 120, durationMs: 2400, playbackRate: 1.5, reversed: false };
  const signature = buildPersonSegmentationSignature(clip);
  assert.equal(isPersonSegmentationCurrent(clip, { enabled: true, matteAssetId: 'mask', signature }), true);
  assert.equal(isPersonSegmentationCurrent({ ...clip, durationMs: 2300 }, { enabled: true, matteAssetId: 'mask', signature }), false);
  assert.equal(isPersonSegmentationCurrent({ ...clip, reversed: true }, { enabled: true, matteAssetId: 'mask', signature }), false);
  assert.equal(isPersonSegmentationCurrent({ ...clip, freezeSourceMs: 300 }, { enabled: true, matteAssetId: 'mask', signature }), false);
});

test('人物置信度蒙版转透明像素并支持阈值与柔化', () => {
  const rgba = personMaskToRGBA(new Float32Array([0, .4, .5, .75, 1]), 5, 1, { threshold: .5, feather: .2 });
  assert.deepEqual([...rgba], [255, 255, 255, 0, 255, 255, 255, 0, 255, 255, 255, 128, 255, 255, 255, 255, 255, 255, 255, 255]);
  assert.throws(() => personMaskToRGBA(new Float32Array(2), 3, 1, {}), /尺寸/);
});

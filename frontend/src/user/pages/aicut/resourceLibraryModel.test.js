import assert from 'node:assert/strict';
import test from 'node:test';
import { assetDurationLabel, assetThumbnailUrl, assetToolActions } from './resourceLibraryModel.js';

test('duration distinguishes still images and unknown media from real elapsed time', () => {
  assert.equal(assetDurationLabel({ kind: 'image', durationMs: 5000 }), '图片');
  assert.equal(assetDurationLabel({ kind: 'video', durationMs: 65400 }), '01:05');
  assert.equal(assetDurationLabel({ kind: 'audio', durationMs: 3661000 }), '1:01:01');
  assert.equal(assetDurationLabel({ kind: 'video' }), '视频');
  assert.equal(assetDurationLabel({ kind: 'audio', durationMs: -20 }), '音频');
});

test('thumbnail never passes a video or audio source to an image element', () => {
  assert.equal(assetThumbnailUrl({ kind: 'image', url: '/still.png' }), '/still.png');
  assert.equal(assetThumbnailUrl({ kind: 'video', url: '/movie.mp4' }), '');
  assert.equal(assetThumbnailUrl({ kind: 'video', posterUrl: '/cover.jpg', url: '/movie.mp4' }), '/cover.jpg');
  assert.equal(assetThumbnailUrl({ kind: 'audio', url: '/music.mp3' }), '');
});

test('overflow retains each media tool and active background cancellation', () => {
  assert.deepEqual(assetToolActions({ id: 'v', kind: 'video', proxyUrl: '/proxy' }, { proxyBusyId: 'v' }), [
    { key: 'asr', label: '识别字幕' },
    { key: 'proxy', label: '生成中…', disabled: true },
  ]);
  assert.deepEqual(assetToolActions({ id: 'i', kind: 'image' }, { backgroundBusyId: 'i' }), [
    { key: 'background', label: '取消去背' },
  ]);
  assert.deepEqual(assetToolActions({ kind: 'audio' }).map(({ key }) => key), ['asr', 'analyze', 'clone']);
  assert.equal(assetToolActions({ kind: 'video', proxyUrl: '/proxy' })[1].label, '重新生成代理');
});

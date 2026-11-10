import test from 'node:test';
import assert from 'node:assert/strict';
import { buildAICutPreviewFrame, getAICutTimelineDuration, nextAICutCanvasFit, seekAICutPreviewTime } from './aicutPreview.js';

const assets = [
  { id: 'video-a', kind: 'video', name: 'A.mp4', url: '/a.mp4' },
  { id: 'image-b', kind: 'image', name: 'B.png', url: '/b.png' },
  { id: 'audio-c', kind: 'audio', name: 'C.mp3', url: '/c.mp3' }
];

const timeline = {
  tracks: [
    { id: 'v1', kind: 'video', order: 0, muted: false, clips: [
      { id: 'clip-a', assetId: 'video-a', kind: 'video', startMs: 1000, durationMs: 3000, sourceStartMs: 500, playbackRate: 2, canvasTransform: { x: .2, y: -.1, scaleX: 1.2, scaleY: .8, rotation: 12, opacity: .7 } }
    ] },
    { id: 'v2', kind: 'video', order: 1, muted: false, clips: [
      { id: 'clip-b', assetId: 'image-b', kind: 'image', startMs: 2000, durationMs: 2000, sourceStartMs: 0 }
    ] },
    { id: 's1', kind: 'subtitle', order: 2, muted: false, settings: { previewVisible: true, subtitleStyleId: '黄字大标题' }, clips: [
      { id: 'sub-a', kind: 'subtitle', startMs: 1500, durationMs: 1000, text: '字幕一' }
    ] },
    { id: 'a1', kind: 'audio', order: 3, muted: false, clips: [
      { id: 'audio-a', assetId: 'audio-c', kind: 'audio', startMs: 0, durationMs: 5000, sourceStartMs: 250, playbackRate: 1, volume: 0.4 }
    ] }
  ]
};

test('预览帧按播放头筛选所有生效轨道并计算素材内部时间', () => {
  const frame = buildAICutPreviewFrame(timeline, assets, 2250);
  assert.deepEqual(frame.visuals.map(item => item.clipId), ['clip-a', 'clip-b']);
  assert.equal(frame.visuals[0].sourceTimeMs, 3000);
  assert.equal(frame.visuals[1].asset.url, '/b.png');
  assert.deepEqual(frame.visuals[0].canvasTransform, { x: .2, y: -.1, scaleX: 1.2, scaleY: .8, rotation: 12, opacity: .7 });
  assert.deepEqual(frame.subtitles.map(item => item.text), ['字幕一']);
  assert.equal(frame.audios[0].sourceTimeMs, 2500);
  assert.equal(frame.audios[0].volume, 0.4);
});

test('预览忽略静音轨、隐藏字幕、缺失素材和片段结束边界', () => {
  const hidden = {
    tracks: [
      { id: 'muted', kind: 'video', muted: true, clips: [{ id: 'muted-clip', assetId: 'video-a', startMs: 0, durationMs: 1000 }] },
      { id: 'missing', kind: 'video', clips: [{ id: 'missing-clip', assetId: 'missing', startMs: 0, durationMs: 1000 }] },
      { id: 'sub', kind: 'subtitle', settings: { previewVisible: false }, clips: [{ id: 'hidden-sub', startMs: 0, durationMs: 1000, text: '不能显示' }] }
    ]
  };
  const during = buildAICutPreviewFrame(hidden, assets, 999);
  assert.deepEqual(during, { timeMs: 999, visuals: [], audios: [], subtitles: [] });
  assert.deepEqual(buildAICutPreviewFrame(timeline, assets, 4000).visuals, []);
});

test('时间线总时长和跳转时间会处理空值、负数和越界', () => {
  assert.equal(getAICutTimelineDuration(timeline), 5000);
  assert.equal(getAICutTimelineDuration({ tracks: [] }), 0);
  assert.equal(seekAICutPreviewTime(-10, 5000), 0);
  assert.equal(seekAICutPreviewTime(6000, 5000), 5000);
  assert.equal(seekAICutPreviewTime('bad', 5000), 0);
});

test('适应画布会在完整显示、铺满和原始大小之间循环', () => {
  assert.equal(nextAICutCanvasFit('fit'), 'fill');
  assert.equal(nextAICutCanvasFit('fill'), 'original');
  assert.equal(nextAICutCanvasFit('original'), 'fit');
  assert.equal(nextAICutCanvasFit('unexpected'), 'fit');
});

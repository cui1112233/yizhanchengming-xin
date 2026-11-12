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

test('文字预览读取字体排版并按打字机进度显示字符', () => {
  const styled = { tracks: [{ id: 's', kind: 'subtitle', settings: {
    previewVisible: true, subtitleAnimation: 'typewriter', fontFamily: 'KaiTi', fontSize: 72,
    textColor: '#ffcc00', strokeColor: '#112233', strokeWidth: 3, backgroundColor: '#00000080',
    letterSpacing: 4, lineHeight: 1.4, textAlign: 'left'
  }, clips: [{ id: 'c', kind: 'subtitle', startMs: 0, durationMs: 1000, text: '剪辑开始' }] }] };
  const item = buildAICutPreviewFrame(styled, [], 500).subtitles[0];
  assert.equal(item.visibleText, '剪辑');
  assert.deepEqual(item.textStyle, {
	fontFamily: 'KaiTi', fontSize: 72, textOpacity: 1, textCurve: 0, textColor: '#ffcc00', strokeColor: '#112233',
    strokeWidth: 3, backgroundColor: '#00000080', shadowColor: '#000000cc', shadowOffsetX: 0,
    shadowOffsetY: 0, shadowBlur: 0, letterSpacing: 4, lineHeight: 1.4, textAlign: 'left'
  });
});

test('文字模板的稳定资源编号进入播放器样式类', () => {
  const timeline = { tracks: [{ id: 't', kind: 'subtitle', settings: { textResourceId: 'text-news-bar', fontSize: 52 }, clips: [{ id: 'c', kind: 'subtitle', startMs: 0, durationMs: 1000, text: '新闻标题' }] }] };
  const frame = buildAICutPreviewFrame(timeline, [], 100);
  assert.equal(frame.subtitles[0].styleId, 'text-news-bar');
  assert.equal(frame.subtitles[0].textStyle.fontSize, 52);
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

test('预览会在线性关键帧之间插值画面变换和音量', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [{
    id: 'c', assetId: 'asset', kind: 'video', startMs: 1000, durationMs: 2000,
    keyframes: [
      { id: 'k1', offsetMs: 0, values: { x: 0, scaleX: 1, scaleY: 1, opacity: 1, volume: .5 } },
      { id: 'k2', offsetMs: 1000, values: { x: 1, scaleX: 2, scaleY: 2, opacity: .5, volume: 1.5 } }
    ]
  }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'asset', kind: 'video', url: '/a.mp4' }], 1500);
  assert.equal(frame.visuals[0].canvasTransform.x, .5);
  assert.equal(frame.visuals[0].canvasTransform.scaleX, 1.5);
  assert.equal(frame.visuals[0].canvasTransform.opacity, .75);
  assert.equal(frame.visuals[0].volume, 1);
});

test('预览使用左侧关键帧的缓动曲线插值', () => {
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [{
    id: 'c', assetId: 'asset', kind: 'video', startMs: 0, durationMs: 1000,
    keyframes: [
      { id: 'k1', offsetMs: 0, values: { x: 0 }, easing: { type: 'ease-in' } },
      { id: 'k2', offsetMs: 1000, values: { x: 1 } }
    ]
  }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'asset', kind: 'video', url: '/a.mp4' }], 500);
  assert.ok(frame.visuals[0].canvasTransform.x > 0);
  assert.ok(frame.visuals[0].canvasTransform.x < .5);
});

test('关键帧允许把画面完全透明并把音量降到零', () => {
  const muted = { tracks: [{ id: 'v', kind: 'video', clips: [{
    id: 'c', assetId: 'asset', kind: 'video', startMs: 0, durationMs: 1000, volume: 1,
    keyframes: [{ id: 'k', offsetMs: 0, values: { opacity: 0, volume: 0 } }]
  }] }] };
  const frame = buildAICutPreviewFrame(muted, [{ id: 'asset', kind: 'video', url: '/a.mp4' }], 0);
  assert.equal(frame.visuals[0].canvasTransform.opacity, 0);
  assert.equal(frame.visuals[0].volume, 0);
});

test('预览按速度曲线积分计算素材时间和当前播放倍率', () => {
  const timeline = { tracks: [{ id: 'video', kind: 'video', order: 0, clips: [{
    id: 'clip', assetId: 'video-a', kind: 'video', startMs: 0, durationMs: 1600, sourceStartMs: 200,
    speedCurve: [{ position: 0, rate: .5 }, { position: .5, rate: 2 }, { position: 1, rate: .5 }]
  }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'video-a', kind: 'video', url: '/a.mp4' }], 800);
  assert.equal(frame.visuals[0].sourceTimeMs, 1200);
  assert.equal(frame.visuals[0].playbackRate, 2);
});

test('音频预览从音量包络读取当前线性插值音量', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', order: 0, clips: [{
    id: 'c', assetId: 'audio', kind: 'audio', startMs: 0, durationMs: 1000, sourceStartMs: 0, volume: 1,
    keyframes: [{ id: 'v0', offsetMs: 0, values: { volume: 0 } }, { id: 'v1', offsetMs: 1000, values: { volume: 2 } }]
  }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'audio', kind: 'audio' }], 250);
  assert.equal(frame.audios[0].volume, .5);
});

test('音频预览把片段两端淡入淡出乘到当前音量上', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', clips: [{
    id: 'c', assetId: 'audio', kind: 'audio', startMs: 1000, durationMs: 2000, sourceStartMs: 0,
    volume: .8, fadeInMs: 400, fadeOutMs: 600
  }] }] };
  const asset = [{ id: 'audio', kind: 'audio' }];
  assert.equal(buildAICutPreviewFrame(timeline, asset, 1000).audios[0].volume, 0);
  assert.equal(buildAICutPreviewFrame(timeline, asset, 1200).audios[0].volume, .4);
  assert.equal(buildAICutPreviewFrame(timeline, asset, 1800).audios[0].volume, .8);
  assert.equal(buildAICutPreviewFrame(timeline, asset, 2700).audios[0].volume, .4);
});

test('预览音频携带降噪、均衡器和声道混音参数', () => {
  const timeline = { tracks: [{ id: 'a', kind: 'audio', order: 0, clips: [{
    id: 'c', assetId: 'asset-a', kind: 'audio', startMs: 0, durationMs: 1000,
    sourceStartMs: 0, sourceDurationMs: 1000,
    audioProcessing: { denoise: .5, eqLowDB: 2, eqMidDB: -1, eqHighDB: 3, channelMode: 'right' }
  }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'asset-a', kind: 'audio', url: '/a.mp3' }], 500);
  assert.deepEqual(frame.audios[0].audioProcessing, {
    denoise: .5, eqLowDB: 2, eqMidDB: -1, eqHighDB: 3, channelMode: 'right'
  });
});

test('画面预览携带高级调色、色键和蒙版参数', () => {
  const processing = {
    color: { exposure: .5, temperature: .2, tint: 0, highlights: .1, shadows: -.1, gamma: 1.1 },
    chromaKey: { enabled: true, color: '#00ff00', similarity: .2, blend: .08 },
    mask: { type: 'rectangle', x: .5, y: .5, width: .8, height: .6, invert: false }
  };
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', assetId: 'asset-v', kind: 'video', startMs: 0, durationMs: 1000, visualProcessing: processing }] }] };
  const frame = buildAICutPreviewFrame(timeline, [{ id: 'asset-v', kind: 'video', url: '/v.mp4' }], 500);
  assert.deepEqual(frame.visuals[0].visualProcessing, processing);
});

test('预览帧只附加与当前剪辑签名一致的人物蒙版素材', () => {
  const clip = { id: 'c', assetId: 'asset-v', kind: 'video', startMs: 0, sourceStartMs: 100, durationMs: 1000, playbackRate: 1 };
  const signature = 'asset-v|100|1000|1|0|none';
  clip.visualProcessing = { personSegmentation: { enabled: true, matteAssetId: 'matte-v', threshold: .5, feather: .1, signature } };
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [clip] }] };
  const assets = [{ id: 'asset-v', kind: 'video', url: '/v.mp4' }, { id: 'matte-v', kind: 'video', url: '/mask.webm' }];
  const current = buildAICutPreviewFrame(timeline, assets, 500).visuals[0];
  assert.equal(current.matteAsset.id, 'matte-v');
  assert.equal(current.personSegmentationCurrent, true);
  clip.durationMs = 900;
  const stale = buildAICutPreviewFrame(timeline, assets, 500).visuals[0];
  assert.equal(stale.matteAsset, null);
  assert.equal(stale.personSegmentationCurrent, false);
});

test('倒放预览反向映射素材时间，定格预览始终读取同一帧', () => {
  const reverseTimeline = { tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', assetId: 'asset-v', kind: 'video', startMs: 1000, durationMs: 2000, sourceStartMs: 500, sourceDurationMs: 5000, playbackRate: 1, reversed: true }] }] };
  const asset = [{ id: 'asset-v', kind: 'video', url: '/v.mp4' }];
  const reverseStart = buildAICutPreviewFrame(reverseTimeline, asset, 1000).visuals[0];
  const reverseLater = buildAICutPreviewFrame(reverseTimeline, asset, 1500).visuals[0];
  assert.equal(reverseStart.sourceTimeMs, 2499);
  assert.equal(reverseLater.sourceTimeMs, 1999);
  assert.equal(reverseLater.reversed, true);
  const freezeTimeline = { tracks: [{ id: 'v', kind: 'video', clips: [{ ...reverseTimeline.tracks[0].clips[0], reversed: false, freezeSourceMs: 1234 }] }] };
  const frozen = buildAICutPreviewFrame(freezeTimeline, asset, 1750).visuals[0];
  assert.equal(frozen.sourceTimeMs, 1234);
  assert.equal(frozen.frozen, true);
});

test('画面预览携带混合模式并把默认模式归一为 normal', () => {
  const assets = [{ id: 'base', kind: 'video', url: '/base.mp4' }, { id: 'top', kind: 'image', url: '/top.png' }];
  const timeline = { tracks: [
    { id: 'base-track', kind: 'video', order: 0, clips: [{ id: 'base-clip', assetId: 'base', kind: 'video', startMs: 0, durationMs: 1000 }] },
    { id: 'top-track', kind: 'video', order: 1, clips: [{ id: 'top-clip', assetId: 'top', kind: 'image', startMs: 0, durationMs: 1000, blendMode: 'screen' }] }
  ] };
  const frame = buildAICutPreviewFrame(timeline, assets, 500);
  assert.equal(frame.visuals[0].blendMode, 'normal');
  assert.equal(frame.visuals[1].blendMode, 'screen');
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { applyProjectTemplate, planProjectTemplate, projectTemplateCatalog } from './projectTemplates.js';

const assets = [
  { id: 'video', kind: 'video', name: '自有视频', durationMs: 8000, url: '/api/assets/video' },
  { id: 'image', kind: 'image', name: '自有图片', durationMs: 0, url: 'https://example.test/photo.png' },
  { id: 'audio', kind: 'audio', name: '自有配乐', durationMs: 1800, url: '/api/assets/audio' },
];
const source = () => ({ revision: 12, script: '原剧本', markers: [{ id: 'id-1', timeMs: 50 }], extra: { preserved: true }, tracks: [
  { id: 'old-track', kind: 'video', name: '锁定旧轨道', order: 14, locked: true, settings: { previewVisible: false }, clips: [
    { id: 'old-clip', kind: 'video', assetId: 'video', startMs: 0, durationMs: 8000, extra: { untouched: true } },
  ] },
] });
const options = { visualAssetId: 'video', title: '自己的标题', startMs: 1200, durationMs: 5000 };
const ids = () => { let n = 0; return () => `id-${++n}`; };

test('catalog defines four usable project layouts with explicit own-media slots', () => {
  assert.deepEqual(projectTemplateCatalog.map(item => item.id), ['template-short-drama', 'template-commentary', 'template-interview', 'template-product']);
  for (const entry of projectTemplateCatalog) {
    assert.equal(entry.category, 'template');
    assert.ok(entry.labelZh && entry.description && Array.isArray(entry.parameters));
    assert.ok(entry.requiredSlots.some(slot => slot.key === 'visualAssetId' && slot.kinds.includes('image') && slot.kinds.includes('video')));
    const result = applyProjectTemplate(source(), entry.id, assets, options, ids());
    const additions = result.timeline.tracks.slice(1);
    assert.ok(additions.some(track => track.kind === 'video' && track.clips[0].assetId === 'video'));
    assert.ok(additions.some(track => track.kind === 'subtitle' && track.clips[0].text === '自己的标题'));
    assert.equal(result.timeline.editorTemplateId, undefined);
  }
});

test('review is deterministic, references inputs and exactly describes the additive application', () => {
  const timeline = source();
  const before = structuredClone(timeline);
  const plan = planProjectTemplate(timeline, 'template-product', assets, options);
  const result = applyProjectTemplate(timeline, 'template-product', assets, options, ids());
  assert.deepEqual(plan, planProjectTemplate(timeline, 'template-product', assets, options));
  assert.equal(plan.validation.valid, true);
  assert.equal(plan.inputReferences.visualAssetId, 'video');
  assert.equal(plan.inputReferences.revision, 12);
  assert.deepEqual(plan.summary, result.summary);
  assert.equal(plan.summary.durationMs, 5000);
  assert.equal(plan.summary.startMs, 1200);
  assert.equal(plan.summary.endMs, 6200);
  assert.deepEqual(plan.summary.trackNames, result.timeline.tracks.slice(1).map(track => track.name));
  assert.equal(plan.summary.trackCount, result.timeline.tracks.length - 1);
  assert.equal(plan.summary.clipCount, result.timeline.tracks.slice(1).flatMap(track => track.clips).length);
  assert.deepEqual(plan.summary.assets, [{ id: 'video', kind: 'video', name: '自有视频' }]);
  assert.deepEqual(timeline, before);
  assert.deepEqual(result.timeline.tracks[0], before.tracks[0]);
  assert.equal(result.timeline.revision, 12);
  assert.deepEqual(result.timeline.extra, { preserved: true });
  assert.deepEqual(result.timeline.markers, before.markers);
  assert.ok(result.timeline.tracks.slice(1).every(track => track.order > 14));
  const [trackId, clipId] = result.selectionKey.split('::');
  assert.ok(result.timeline.tracks.find(track => track.id === trackId)?.clips.some(clip => clip.id === clipId));
});

test('templates produce distinct real visual and text styles without changing source references', () => {
  const outputs = projectTemplateCatalog.map(entry => applyProjectTemplate(source(), entry.id, assets, options, ids()).timeline.tracks.slice(1));
  assert.deepEqual(outputs.map(tracks => tracks.find(track => track.kind === 'video').clips[0].visualEffect.preset), ['cool-cinema', 'original', 'original', 'bright'], '每个模板应明确选择受支持的滤镜，不能依赖未知名称的静默回退');
  assert.equal(new Set(outputs.map(tracks => JSON.stringify(tracks.map(track => ({ kind: track.kind, settings: track.settings, clips: track.clips.map(clip => ({ visualEffect: clip.visualEffect, canvasTransform: clip.canvasTransform, assetId: clip.assetId })) }))))).size, 4);
  for (const tracks of outputs) {
    const visual = tracks.find(track => track.kind === 'video').clips[0];
    assert.equal(visual.assetId, 'video');
    assert.equal(visual.sourceStartMs, 0);
    assert.equal(visual.sourceDurationMs, 8000);
    assert.ok(visual.visualEffect.preset);
    assert.ok(tracks.find(track => track.kind === 'subtitle').settings.fontSize > 0);
  }
});

test('short source video clamps whole template but short optional audio only clamps its clip', () => {
  const short = assets.map(asset => asset.id === 'video' ? { ...asset, durationMs: 3100 } : asset);
  const result = applyProjectTemplate(source(), 'template-commentary', short, { ...options, audioAssetId: 'audio' }, ids());
  assert.equal(result.summary.durationMs, 3100);
  const tracks = result.timeline.tracks.slice(1);
  assert.equal(tracks.find(track => track.kind === 'video').clips[0].durationMs, 3100);
  assert.equal(tracks.find(track => track.kind === 'subtitle').clips[0].durationMs, 3100);
  const audio = tracks.find(track => track.kind === 'audio').clips[0];
  assert.equal(audio.startMs, 1200);
  assert.equal(audio.durationMs, 1800);
  assert.equal(audio.sourceDurationMs, 1800);
  assert.equal(audio.audioProcessing.loudnessEnabled, true);
  assert.ok(result.summary.warnings.length);
});

test('image can hold for requested duration and defaults are finite', () => {
  const result = applyProjectTemplate({ tracks: [] }, 'template-interview', assets, { visualAssetId: 'image' }, ids());
  assert.equal(result.summary.durationMs, 5000);
  assert.equal(result.summary.startMs, 0);
  assert.equal(result.timeline.tracks[0].kind, 'video');
  assert.equal(result.timeline.tracks[0].clips[0].kind, 'image');
  assert.equal(result.timeline.tracks[0].clips[0].durationMs, 5000);
});

test('source offsets clamp safely and never extend source content', () => {
  const result = applyProjectTemplate(source(), 'template-short-drama', assets, { ...options, sourceStartMs: 6900, audioAssetId: 'audio', audioSourceStartMs: 1000 }, ids());
  const clips = result.timeline.tracks.slice(1).flatMap(track => track.clips);
  assert.equal(result.summary.durationMs, 1100);
  const video = clips.find(clip => clip.kind === 'video');
  assert.equal(video.sourceStartMs, 6900);
  assert.ok(video.sourceStartMs + video.durationMs <= video.sourceDurationMs);
  const audio = clips.find(clip => clip.kind === 'audio');
  assert.equal(audio.durationMs, 800);
});

test('missing explicit visual choice and unknown template never silently pick a resource', () => {
  for (const id of ['', 'missing']) assert.throws(() => planProjectTemplate(source(), id, assets, options));
  for (const visualAssetId of [undefined, '', 'missing', 'audio']) assert.throws(() => planProjectTemplate(source(), 'template-product', assets, { ...options, visualAssetId }));
  assert.throws(() => planProjectTemplate(source(), 'template-product', assets, { ...options, audioAssetId: 'video' }));
});

test('invalid times, source ranges, and excessive lengths are rejected before insertion', () => {
  for (const patch of [
    { durationMs: 0 }, { durationMs: -1 }, { durationMs: Infinity }, { durationMs: NaN }, { durationMs: 600001 },
    { startMs: -1 }, { startMs: Infinity }, { startMs: 86400000 }, { sourceStartMs: -1 }, { sourceStartMs: 8000 },
    { sourceStartMs: NaN }, { audioAssetId: 'audio', audioSourceStartMs: 1800 },
  ]) assert.throws(() => planProjectTemplate(source(), 'template-product', assets, { ...options, ...patch }), JSON.stringify(patch));
});

test('unusable media URLs and invalid timed-media duration cannot enter a plan', () => {
  for (const url of ['', 'javascript:alert(1)', 'data:video/mp4,anything', '//unknown.test/video', 'file:///C:/clip.mp4']) {
    assert.throws(() => planProjectTemplate(source(), 'template-product', [{ ...assets[0], url }], options));
  }
  for (const durationMs of [0, -1, NaN, Infinity]) assert.throws(() => planProjectTemplate(source(), 'template-product', [{ ...assets[0], durationMs }], options));
});

test('new object IDs avoid source tracks, clips, markers and earlier new IDs', () => {
  const result = applyProjectTemplate(source(), 'template-product', assets, options, ids());
  const newIds = result.timeline.tracks.slice(1).flatMap(track => [track.id, ...track.clips.map(clip => clip.id)]);
  assert.equal(new Set(newIds).size, newIds.length);
  assert.ok(!newIds.includes('id-1'));
  assert.throws(() => applyProjectTemplate(source(), 'template-product', assets, options, () => 'old-clip'));
});

test('reapplication is additive and can be undone by retaining original snapshot', () => {
  const timeline = source();
  const first = applyProjectTemplate(timeline, 'template-product', assets, options, ids());
  const second = applyProjectTemplate(first.timeline, 'template-product', assets, options, ids());
  assert.equal(second.timeline.tracks.length, timeline.tracks.length + first.summary.trackCount * 2);
  assert.deepEqual(second.timeline.tracks.slice(0, first.timeline.tracks.length), first.timeline.tracks);
  assert.deepEqual(JSON.parse(JSON.stringify(second.timeline)), second.timeline);
  assert.deepEqual(timeline, source());
});

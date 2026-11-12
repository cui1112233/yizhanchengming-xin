import test from 'node:test';
import assert from 'node:assert/strict';
import { extractPersonalTemplate, planPersonalTemplate, applyPersonalTemplate } from './personalTemplates.js';
const source = () => ({ revision: 3, tracks: [{ id: 'private-track', kind: 'video', name: '画面', order: 0, settings: { previewVisible: true }, clips: [{ id: 'private-clip', assetId: 'private-asset', kind: 'video', name: '片段', startMs: 100, durationMs: 1000, sourceStartMs: 200, sourceDurationMs: 5000, playbackRate: 2, canvasTransform: { x: .2, y: 0, scaleX: 1, scaleY: 1, rotation: 0, opacity: 1 } }] }] });
const assets = [{ id: 'new-asset', kind: 'video', name: '替换', url: '/safe/video.mp4', durationMs: 3000 }];
const options = { bindings: { 'slot-1': 'new-asset' }, startMs: 500 };
test('extracts slots without source identifiers or asset addresses; applies new IDs and preserves source', () => {
 const original = source(); const before = JSON.stringify(original); const layout = extractPersonalTemplate(original);
 assert.equal(JSON.stringify(layout).includes('private-'), false);
 const target = { revision: 6, tracks: [{ id: 'existing', kind: 'video', clips: [], order: 4 }] };
 let n = 0; const result = applyPersonalTemplate(target, layout, assets, options, () => `new-${++n}`);
 assert.equal(JSON.stringify(original), before); assert.equal(target.tracks.length, 1);
 assert.equal(result.timeline.tracks.length, 2); const added = result.timeline.tracks[1];
 assert.equal(added.order, 5); assert.equal(added.clips[0].assetId, 'new-asset'); assert.equal(added.clips[0].startMs, 600); assert.equal(added.clips[0].sourceDurationMs, 3000); assert.equal(added.clips[0].id, 'new-2');
 assert.deepEqual({ ...result.timeline, tracks: result.timeline.tracks.slice(0,1) }, target);
});
test('duration check respects playback rate and rejects short or wrong replacement', () => {
 const layout = extractPersonalTemplate(source());
 assert.throws(() => planPersonalTemplate({ tracks: [] }, layout, [{ ...assets[0], durationMs: 2199 }], options), /时长不足/);
 assert.throws(() => planPersonalTemplate({ tracks: [] }, layout, assets, { bindings: {} }), /请选择|请为/);
 assert.throws(() => planPersonalTemplate({ tracks: [] }, layout, [{ ...assets[0], kind: 'audio' }], options), /请为/);
 assert.throws(() => planPersonalTemplate({ tracks: [] }, layout, assets, { ...options, startMs: 86399999 }), /超出/);
});
test('rejects private references, compound, font, matte, transition and ID collisions', () => {
 for (const patch of [{ compoundId: 'linked' }, { linkGroupId: 'live' }, { visualProcessing: { personSegmentation: { matteAssetId: 'private', enabled: true } } }, { transition: 'fade' }]) { const t=source(); Object.assign(t.tracks[0].clips[0],patch); assert.throws(()=>extractPersonalTemplate(t)); }
 const t=source();t.tracks[0].settings.fontAssetId='private';assert.throws(()=>extractPersonalTemplate(t));
 const layout=extractPersonalTemplate(source());assert.throws(()=>applyPersonalTemplate({tracks:[]},layout,assets,options,()=> 'slot-1'),/编号/);
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { executeAICutCommand } from './aicutCommands.js';
import { buildAICutPreviewFrame } from './aicutPreview.js';

const fixture = (patch = {}, trackPatch = {}) => ({ revision: 8, tracks: [
  { id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', startMs: 1000, durationMs: 2000, sourceStartMs: 500, sourceDurationMs: 10000, ...patch }, { id: 'other', kind: 'video', startMs: 5000, durationMs: 1000 }], ...trackPatch },
  { id: 'a', kind: 'audio', clips: [{ id: 'voice', startMs: 1000, durationMs: 2000, linkGroupId: 'g' }] }
] });
const trim = (timeline, edge, timeMs) => executeAICutCommand(timeline, { type: 'clip.trim_to_playhead', trackId: 'v', clipId: 'c', edge, timeMs });
const selected = timeline => timeline.tracks[0].clips[0];

test('trim to playhead rounds once, adjusts source and keeps every other clip and input intact', () => {
  const input = fixture({ playbackRate: 2, linkGroupId: 'g', keyframes: [{ id: 'k', offsetMs: 1200, values: { opacity: .5 } }] });
  const before = structuredClone(input);
  const left = trim(input, 'left', 1750.6);
  assert.equal(left.changed, true);
  assert.equal(left.selectionKey, 'v::c');
  assert.equal(selected(left.timeline).startMs, 1751);
  assert.equal(selected(left.timeline).durationMs, 1249);
  assert.equal(selected(left.timeline).sourceStartMs, 2002);
  assert.equal(selected(left.timeline).keyframes.find(frame => frame.id === 'k').offsetMs, 449);
  assert.deepEqual(left.timeline.tracks[0].clips[1], input.tracks[0].clips[1]);
  assert.deepEqual(left.timeline.tracks[1], input.tracks[1]);
  const right = trim(input, 'right', 2400.4);
  assert.equal(selected(right.timeline).durationMs, 1400);
  assert.equal(selected(right.timeline).sourceStartMs, 500);
  assert.deepEqual(input, before);
});

test('curve trims preserve source timing, animation and the existing source bounds', () => {
  const input = fixture({ speedCurve: [{ position: 0, rate: 1 }, { position: 1, rate: 3 }], playbackRate: 2, sourceDurationMs: 4500, animation: { enter: { presetId: 'fade', durationMs: 400 } } });
  const left = selected(trim(input, 'left', 2000).timeline);
  assert.equal(left.sourceStartMs, 2000);
  assert.equal(left.durationMs, 1000);
  assert.deepEqual(left.speedCurve, [{ position: 0, rate: 2 }, { position: 1, rate: 3 }]);
  assert.equal(left.playbackRate, 2.5);
  assert.deepEqual(left.animation, selected(input).animation);
  assert.ok(left.sourceStartMs + left.durationMs * left.playbackRate <= left.sourceDurationMs);
  const right = selected(trim(input, 'right', 2000).timeline);
  assert.equal(right.sourceStartMs, 500);
  assert.equal(right.playbackRate, 1.5);
});

test('old clips, subtitle and image clips can trim without NaN source offsets', () => {
  for (const kind of ['video', 'subtitle', 'image']) {
    const input = fixture({ kind, sourceStartMs: undefined, playbackRate: undefined, animation: { exit: { presetId: 'fade', durationMs: 200 } } });
    const result = selected(trim(input, 'left', 1500).timeline);
    assert.equal(result.startMs, 1500);
    assert.equal(result.durationMs, 1500);
    assert.equal(result.sourceStartMs, 500);
    assert.deepEqual(result.animation, selected(input).animation);
  }
});

test('same edge is a no-op and invalid or too-short trim does not change input', () => {
  const input = fixture();
  assert.equal(trim(input, 'left', 1000).changed, false);
  assert.equal(trim(input, 'right', 3000).changed, false);
  for (const [edge, time] of [['left', 999], ['right', 3001], ['left', 3000], ['right', 1000], ['left', 2901], ['right', 1099], ['left', NaN], ['left', Infinity], ['left', null], ['bad', 2000]]) {
    assert.throws(() => trim(input, edge, time));
  }
  assert.equal(selected(trim(input, 'left', 2900).timeline).durationMs, 100);
  assert.throws(() => trim(fixture({}, { locked: true }), 'left', 1500), /锁定/);
});

test('unsupported transition and compound bindings reject instead of leaving broken references', () => {
  for (const input of [fixture({ kind: 'transition' }), fixture({ compoundId: 'group', compoundName: '组' }), fixture({ transition: 'transition-fade', transitionMs: 200 })]) {
    assert.throws(() => trim(input, 'left', 1500), /转场|复合/);
  }
  const bound = fixture();
  bound.tracks.push({ id: 't', kind: 'transition', clips: [{ id: 't1', transitionTrackId: 'v', transitionLeftClipId: 'c', transitionRightClipId: 'other' }] });
  assert.throws(() => trim(bound, 'right', 2000), /转场/);
});

test('forward, reverse and frozen trims retain the visible source at the kept timeline times', () => {
  for (const patch of [{ playbackRate: 1.5 }, { speedCurve: [{ position: 0, rate: 1 }, { position: 1, rate: 3 }] }, { reversed: true }, { reversed: true, speedCurve: [{ position: 0, rate: 1 }, { position: 1, rate: 3 }] }, { freezeSourceMs: 900 }]) {
    const input = fixture({ assetId: 'asset', ...patch });
    const assets = [{ id: 'asset', kind: 'video', url: '/asset.mp4' }];
    for (const [edge, time] of [['left', 1500], ['right', 2500]]) {
      const output = trim(input, edge, time).timeline;
      for (const sample of [1750, 2000, 2250]) {
        const before = buildAICutPreviewFrame(input, assets, sample).visuals[0];
        const after = buildAICutPreviewFrame(output, assets, sample).visuals[0];
        assert.ok(Math.abs(before.sourceTimeMs - after.sourceTimeMs) <= 1);
      }
    }
  }
});

test('known source bounds cannot overflow after a trim', () => {
  assert.throws(() => trim(fixture({ sourceDurationMs: 1000, playbackRate: 2 }), 'left', 1500), /素材/);
  assert.throws(() => trim(fixture({ sourceDurationMs: 2499 }), 'left', 1500), /素材/);
});

test('trim preserves interpolated transform and volume including custom Bezier easing at every retained millisecond', () => {
  const assets = [{ id: 'asset', kind: 'video', url: '/asset.mp4' }];
  for (const easing of [{ type: 'linear' }, { type: 'ease-in' }, { type: 'custom', x1: .18, y1: .3, x2: .71, y2: .95 }, { type: 'hold' }]) {
    const input = fixture({ assetId: 'asset', keyframes: [
      { id: 'k0', offsetMs: 0, easing, values: { x: 0, y: -.5, scaleX: 1, scaleY: .5, rotation: -90, opacity: .2, volume: .2 } },
      { id: 'k1', offsetMs: 2000, values: { x: 1, y: .5, scaleX: 2, scaleY: 1.5, rotation: 90, opacity: .8, volume: .9 } }
    ] });
    for (const [edge, timeMs] of [['left', 1871], ['right', 2249]]) {
      const output = trim(input, edge, timeMs).timeline;
      const clip = selected(output);
      for (let time = clip.startMs; time < clip.startMs + clip.durationMs; time++) {
        const before = buildAICutPreviewFrame(input, assets, time).visuals[0];
        const after = buildAICutPreviewFrame(output, assets, time).visuals[0];
        for (const field of ['x', 'y', 'scaleX', 'scaleY', 'rotation', 'opacity']) assert.ok(Math.abs(before.canvasTransform[field] - after.canvasTransform[field]) < 1e-9, `${easing.type}/${edge}/${time}/${field}`);
        assert.ok(Math.abs(before.volume - after.volume) < 1e-9);
      }
      assert.ok(clip.keyframes.every(frame => Number.isInteger(frame.offsetMs) && frame.offsetMs >= 0 && frame.offsetMs <= clip.durationMs));
      assert.equal(new Set(clip.keyframes.map(frame => frame.id)).size, clip.keyframes.length);
    }
  }
});

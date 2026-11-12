import test from 'node:test';
import assert from 'node:assert/strict';
import { executeAICutCommand } from './aicutCommands.js';
import { clipAnimationFrame } from './aicutClipAnimation.js';

const timeline = { revision: 3, tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', assetId: 'a', kind: 'video', startMs: 0, durationMs: 2000, sourceDurationMs: 2000 }] }] };
const command = { type: 'clip.animation_set', trackId: 'v', clipId: 'c' };
const animation = { entrance: { preset: 'fade', durationMs: 800 }, exit: { preset: 'slide', durationMs: 800 }, loop: { preset: 'float', durationMs: 2000 } };

test('animation command roundtrip preserves all independent phases and old clip stays neutral', () => {
  const result = executeAICutCommand(timeline, { ...command, animation });
  assert.deepEqual(JSON.parse(JSON.stringify(result.timeline)).tracks[0].clips[0].animation, animation);
  assert.equal(timeline.tracks[0].clips[0].animation, undefined);
  assert.deepEqual(clipAnimationFrame(timeline.tracks[0].clips[0].animation, 2000, 100), { x: 0, y: 0, scale: 1, opacity: 1 });
  assert.equal(executeAICutCommand(result.timeline, { ...command, animation: null }).timeline.tracks[0].clips[0].animation, null);
});

test('command rejects invalid phases locked track and nonvisual clips', () => {
  assert.throws(() => executeAICutCommand(timeline, { ...command, animation: { loop: { preset: 'fade', durationMs: 500 } } }), /动画/);
  assert.throws(() => executeAICutCommand({ tracks: [{ ...timeline.tracks[0], locked: true }] }, { ...command, animation }), /解锁|锁定/);
  assert.throws(() => executeAICutCommand({ tracks: [{ ...timeline.tracks[0], clips: [{ ...timeline.tracks[0].clips[0], kind: 'audio' }] }] }, { ...command, animation }), /视频或图片/);
});

test('split creates independent clips with their own fitted animation boundaries', () => {
  const configured = executeAICutCommand(timeline, { ...command, animation }).timeline;
  const split = executeAICutCommand(configured, { type: 'clip.split_at', trackId: 'v', clipId: 'c', offsetMs: 1000 }, { idFactory: () => 'right' }).timeline;
  const [left, right] = split.tracks[0].clips;
  assert.deepEqual(left.animation, animation);
  assert.deepEqual(right.animation, animation);
  assert.equal(left.durationMs, 1000);
  assert.equal(right.durationMs, 1000);
  assert.equal(clipAnimationFrame(left.animation, left.durationMs, 250).opacity, .5);
  assert.equal(clipAnimationFrame(right.animation, right.durationMs, 750).x, .5);
  const changed = executeAICutCommand(split, { ...command, clipId: right.id, animation: null }).timeline;
  assert.deepEqual(changed.tracks[0].clips[0].animation, animation);
  assert.equal(changed.tracks[0].clips[1].animation, null);
});

test('split and duplicate nested animation phases have no shared mutable references', () => {
  const configured = executeAICutCommand(timeline, { ...command, animation }).timeline;
  const split = executeAICutCommand(configured, { type: 'clip.split_at', trackId: 'v', clipId: 'c', offsetMs: 1000 }, { idFactory: () => 'right' }).timeline;
  split.tracks[0].clips[1].animation.entrance.durationMs = 100;
  assert.equal(split.tracks[0].clips[0].animation.entrance.durationMs, 800);
  assert.equal(configured.tracks[0].clips[0].animation.entrance.durationMs, 800);
  const duplicate = executeAICutCommand(configured, { type: 'clip.duplicate', trackId: 'v', clipId: 'c', newId: 'copy' }).timeline;
  duplicate.tracks[0].clips.find(clip => clip.id === 'copy').animation.exit.preset = 'zoom';
  assert.equal(duplicate.tracks[0].clips.find(clip => clip.id === 'c').animation.exit.preset, 'slide');
  assert.equal(configured.tracks[0].clips[0].animation.exit.preset, 'slide');
});

import assert from 'node:assert/strict';
import test from 'node:test';
import * as timelineTools from './timelineTools.js';
import { executeAICutCommand } from '../aicutCommands.js';
const { buildTimelineAdvancedItems } = timelineTools;

const commandNames = [
  'cut', 'duplicate', 'consolidate', 'previousEdit', 'nextEdit', 'nudgeBack', 'nudgeForward',
  'addKeyframe', 'removeKeyframe', 'addMarker', 'removeMarker', 'link', 'unlink',
  'createCompound', 'expandCompound', 'openCompound', 'dissolveCompound',
];
const ready = { hasSelection: true, selectionCount: 2, hasLink: false, hasCompound: false, consolidateBusy: false };
const boundCommands = Object.fromEntries(commandNames.map((name) => [name, () => {}]));
const actions = (items) => items.flatMap((item) => item.type === 'group' ? item.children : [item]);
const byKey = (items) => Object.fromEntries(actions(items).map((item) => [item.key, item]));

test('advanced operations are grouped and retain every command callback', () => {
  const called = [];
  const commands = Object.fromEntries(commandNames.map((name) => [name, (event) => called.push([name, event])]));
  const items = buildTimelineAdvancedItems(ready, commands);
  assert.deepEqual(items.map((item) => item.label), ['片段操作', '定位与逐帧', '关键帧与标记', '联动与复合片段']);
  assert.ok(items.every((item) => item.type === 'group' && item.children.length));
  const indexed = byKey(items);
  assert.deepEqual(Object.keys(indexed), commandNames);
  const event = { key: 'original-menu-event' };
  for (const name of commandNames) {
    assert.strictEqual(indexed[name].onClick, commands[name]);
    indexed[name].onClick(event);
  }
  assert.deepEqual(called, commandNames.map((name) => [name, event]));
  assert.equal(indexed.nudgeBack.label, '向前一帧');
  assert.equal(indexed.nudgeForward.label, '向后一帧');
});

test('empty selection disables clip operations while navigation and markers stay usable', () => {
  const indexed = byKey(buildTimelineAdvancedItems({}, boundCommands));
  for (const name of ['cut', 'duplicate', 'consolidate', 'nudgeBack', 'nudgeForward', 'addKeyframe', 'removeKeyframe', 'link', 'unlink', 'createCompound', 'expandCompound', 'openCompound', 'dissolveCompound']) {
    assert.equal(indexed[name].disabled, true, name);
  }
  for (const name of ['previousEdit', 'nextEdit', 'addMarker', 'removeMarker']) {
    assert.equal(indexed[name].disabled, false, name);
  }
});

test('selection count, existing links and compounds determine the available operations', () => {
  const single = byKey(buildTimelineAdvancedItems({ ...ready, selectionCount: 1 }, boundCommands));
  assert.equal(single.consolidate.disabled, true);
  assert.equal(single.createCompound.disabled, true);
  for (const name of ['cut', 'duplicate', 'nudgeBack', 'nudgeForward', 'addKeyframe', 'removeKeyframe', 'link']) assert.equal(single[name].disabled, false, name);
  const multiple = byKey(buildTimelineAdvancedItems(ready, boundCommands));
  assert.equal(multiple.consolidate.disabled, false);
  assert.equal(multiple.createCompound.disabled, false);
  assert.equal(multiple.unlink.disabled, true);
  for (const name of ['expandCompound', 'openCompound', 'dissolveCompound']) assert.equal(multiple[name].disabled, true, name);
  const grouped = byKey(buildTimelineAdvancedItems({ ...ready, hasLink: true, hasCompound: true }, boundCommands));
  assert.equal(grouped.link.disabled, true);
  for (const name of ['unlink', 'expandCompound', 'openCompound', 'dissolveCompound']) assert.equal(grouped[name].disabled, false, name);
});

test('unsupported menu commands and empty groups are absent, and primary actions are not repeated', () => {
  assert.deepEqual(buildTimelineAdvancedItems(ready, {}), []);
  const items = buildTimelineAdvancedItems(ready, { cut: () => {}, addMarker: () => {}, nextEdit: null }, [], ['cut']);
  assert.equal(items.length, 1);
  assert.equal(items[0].key, 'keyframe-marker-actions');
  assert.deepEqual(items[0].children.map(item => item.key), ['addMarker']);
});

test('menu editing actions explain locked and busy states while consolidation cancellation stays usable', () => {
  const locked = byKey(buildTimelineAdvancedItems({ ...ready, selectionLocked: true }, boundCommands));
  for (const name of ['cut', 'nudgeBack', 'addKeyframe', 'link', 'createCompound']) {
    assert.equal(locked[name].disabled, true, name);
    assert.match(locked[name].disabledReason, /锁定/, name);
  }
  const busy = byKey(buildTimelineAdvancedItems({ ...ready, busy: true, consolidateBusy: true }, boundCommands));
  assert.equal(busy.addMarker.disabled, true);
  assert.match(busy.addMarker.disabledReason, /处理/);
  assert.equal(busy.consolidate.disabled, false);
});

test('primary descriptor exposes bound commands and invokes real clip mutation commands', () => {
  assert.equal(typeof timelineTools.buildTimelinePrimaryActions, 'function');
  let timeline = { tracks: [{ id: 'video', kind: 'video', locked: false, clips: [
    { id: 'first', startMs: 0, durationMs: 1000 },
    { id: 'second', startMs: 1000, durationMs: 1000 },
  ] }], markers: [] };
  const items = timelineTools.buildTimelinePrimaryActions({ ...ready, canSplit: true }, {
    delete: () => { timeline = executeAICutCommand(timeline, { type: 'clip.delete_many', clipKeys: ['video::first'] }).timeline; },
    addMarker: () => { timeline = executeAICutCommand(timeline, { type: 'timeline.marker_add', timeMs: 1500, label: '标记' }, { idFactory: () => 'toolbar-marker' }).timeline; },
    missing: () => { throw new Error('unsupported action'); },
    split: null,
  });
  assert.deepEqual(items.map(item => item.key), ['delete', 'addMarker']);
  items[0].onClick();
  items[1].onClick();
  assert.deepEqual(timeline.tracks[0].clips.map(clip => clip.id), ['second']);
  assert.deepEqual(timeline.markers.map(marker => [marker.timeMs, marker.label]), [[1500, '标记']]);
});

test('disabled primary callbacks cannot mutate locked, empty or busy timelines', () => {
  assert.equal(typeof timelineTools.buildTimelinePrimaryActions, 'function');
  let count = 0;
  const commands = Object.fromEntries(['undo', 'redo', 'trimLeft', 'trimRight', 'split', 'cut', 'copy', 'duplicate', 'delete', 'addKeyframe', 'addMarker'].map(key => [key, () => { count += 1; }]));
  for (const state of [{}, { ...ready, selectionLocked: true, canSplit: false }, { ...ready, busy: true, canUndo: true, canRedo: true, canSplit: true }]) {
    for (const item of timelineTools.buildTimelinePrimaryActions(state, commands).filter(item => item.disabled)) {
      assert.ok(item.disabledReason, item.key);
      item.onClick();
    }
  }
  assert.equal(count, 0);
  const unlocked = byKey(timelineTools.buildTimelinePrimaryActions({ ...ready, selectionLocked: true }, commands));
  assert.equal(unlocked.copy.disabled, false, 'copying a locked clip does not change its track');
  assert.equal(unlocked.addMarker.disabled, false, 'timeline markers do not change locked clips');
});

test('primary history, split and trim capability checks explain unavailable operations', () => {
  assert.equal(typeof timelineTools.buildTimelinePrimaryActions, 'function');
  const commands = Object.fromEntries(['undo', 'redo', 'trimLeft', 'trimRight', 'split'].map(key => [key, () => {}]));
  const items = byKey(timelineTools.buildTimelinePrimaryActions({ ...ready, canUndo: true, canRedo: false, canSplit: false, canTrimLeft: false, canTrimRight: true }, commands));
  assert.equal(items.undo.disabled, false);
  assert.equal(items.redo.disabled, true);
  assert.equal(items.trimLeft.disabled, true);
  assert.match(items.trimLeft.disabledReason, /播放头/);
  assert.equal(items.trimRight.disabled, false);
  assert.equal(items.split.disabled, true);
});

test('busy consolidation remains cancellable after selection is cleared', () => {
  const cancel = () => {};
  const indexed = byKey(buildTimelineAdvancedItems({ consolidateBusy: true }, { consolidate: cancel }));
  assert.equal(indexed.consolidate.label, '取消合并');
  assert.equal(indexed.consolidate.disabled, false);
  assert.equal(indexed.consolidate.danger, true);
  assert.strictEqual(indexed.consolidate.onClick, cancel);
});

test('track submenu preserves the original nested items and callbacks without mutation', () => {
  const onClick = () => {};
  const trackItems = Object.freeze([Object.freeze({ key: 'track-lock', label: '锁定轨道', onClick })]);
  const state = Object.freeze({ ...ready });
  const items = buildTimelineAdvancedItems(state, {}, trackItems);
  const track = items.at(-1);
  assert.equal(track.label, '轨道设置');
  assert.strictEqual(track.children, trackItems);
  assert.strictEqual(track.children[0].onClick, onClick);
  assert.equal(buildTimelineAdvancedItems(state, {}).length, items.length - 1);
});

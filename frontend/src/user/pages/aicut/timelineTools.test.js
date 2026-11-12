import assert from 'node:assert/strict';
import test from 'node:test';
import { buildTimelineAdvancedItems } from './timelineTools.js';

const commandNames = [
  'cut', 'duplicate', 'consolidate', 'previousEdit', 'nextEdit', 'nudgeBack', 'nudgeForward',
  'addKeyframe', 'removeKeyframe', 'addMarker', 'removeMarker', 'link', 'unlink',
  'createCompound', 'expandCompound', 'openCompound', 'dissolveCompound',
];
const ready = { hasSelection: true, selectionCount: 2, hasLink: false, hasCompound: false, consolidateBusy: false };
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
  const indexed = byKey(buildTimelineAdvancedItems({}, {}));
  for (const name of ['cut', 'duplicate', 'consolidate', 'nudgeBack', 'nudgeForward', 'addKeyframe', 'removeKeyframe', 'link', 'unlink', 'createCompound', 'expandCompound', 'openCompound', 'dissolveCompound']) {
    assert.equal(indexed[name].disabled, true, name);
  }
  for (const name of ['previousEdit', 'nextEdit', 'addMarker', 'removeMarker']) {
    assert.equal(indexed[name].disabled, false, name);
  }
});

test('selection count, existing links and compounds determine the available operations', () => {
  const single = byKey(buildTimelineAdvancedItems({ ...ready, selectionCount: 1 }, {}));
  assert.equal(single.consolidate.disabled, true);
  assert.equal(single.createCompound.disabled, true);
  for (const name of ['cut', 'duplicate', 'nudgeBack', 'nudgeForward', 'addKeyframe', 'removeKeyframe', 'link']) assert.equal(single[name].disabled, false, name);
  const multiple = byKey(buildTimelineAdvancedItems(ready, {}));
  assert.equal(multiple.consolidate.disabled, false);
  assert.equal(multiple.createCompound.disabled, false);
  assert.equal(multiple.unlink.disabled, true);
  for (const name of ['expandCompound', 'openCompound', 'dissolveCompound']) assert.equal(multiple[name].disabled, true, name);
  const grouped = byKey(buildTimelineAdvancedItems({ ...ready, hasLink: true, hasCompound: true }, {}));
  assert.equal(grouped.link.disabled, true);
  for (const name of ['unlink', 'expandCompound', 'openCompound', 'dissolveCompound']) assert.equal(grouped[name].disabled, false, name);
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

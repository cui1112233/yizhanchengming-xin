import test from 'node:test';
import assert from 'node:assert/strict';
import { timelineRulerTicks } from './timelineRuler.js';
import { timelineTimeToPixels } from '../aicutTimeline.js';

test('ruler marks use the same clock and position as the actual clips at each zoom', () => {
  for (const zoom of [.5, 1, 2, 4]) {
    const ticks = timelineRulerTicks(2000, zoom);
    assert.equal(ticks[0].left, 0);
    for (const tick of ticks) assert.equal(tick.left, timelineTimeToPixels(tick.timeMs, zoom));
    assert.ok(ticks.every((tick, i) => !i || tick.left - ticks[i - 1].left >= 89));
    assert.ok(ticks.at(-1).left <= 2000);
  }
});
test('long project ruler has bounded marks and labels beyond one minute', () => {
  const ticks = timelineRulerTicks(72 * 7200, 1);
  assert.ok(ticks.length <= 1001);
  assert.equal(ticks.find(tick => tick.timeMs === 60000)?.label, '01:00');
});

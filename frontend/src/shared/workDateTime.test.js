import test from 'node:test';
import assert from 'node:assert/strict';
import { formatWorkDateTime } from './workDateTime.js';

test('work dates match the requested punctuation, padding and Beijing time', () => {
  assert.equal(formatWorkDateTime('2026-11-03T06:24:59Z'), '26-11-03|14：24');
  assert.equal(formatWorkDateTime('2026-11-13T02:39:10.6381166Z'), '26-11-13|10：39');
  assert.equal(formatWorkDateTime('2026-11-03T14:24:00+08:00'), '26-11-03|14：24');
});
test('midnight and year rollover do not produce 24 oclock or the UTC date', () => {
  assert.equal(formatWorkDateTime('2026-12-31T16:00:00Z'), '27-01-01|00：00');
  assert.equal(formatWorkDateTime(new Date('2024-02-29T01:02:00Z')), '24-02-29|09：02');
});
test('missing or invalid timestamps do not show raw data or an invented current time', () => {
  for (const value of [null, undefined, '', 'invalid', '0001-01-01T00:00:00Z']) {
    assert.equal(formatWorkDateTime(value), '—');
    assert.equal(formatWorkDateTime(value, '尚未开始制作'), '尚未开始制作');
  }
});

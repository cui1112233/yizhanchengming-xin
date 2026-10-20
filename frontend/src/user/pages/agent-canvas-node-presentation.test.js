import test from 'node:test';
import assert from 'node:assert/strict';
import { isMediaCanvasNodeType, normalizeCanvasNodeType } from './agent-canvas-node-presentation.js';

test('legacy, blank, and unknown node types become text', () => {
  for (const value of ['copy', '', null, 'unexpected']) {
    assert.equal(normalizeCanvasNodeType(value), 'text');
  }
});

test('supported media types remain media types', () => {
  for (const value of ['image', 'video', 'audio']) {
    assert.equal(normalizeCanvasNodeType(value), value);
    assert.equal(isMediaCanvasNodeType(value), true);
  }
  assert.equal(isMediaCanvasNodeType('copy'), false);
});

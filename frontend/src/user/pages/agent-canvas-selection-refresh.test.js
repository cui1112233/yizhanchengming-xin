import test from 'node:test';
import assert from 'node:assert/strict';
import { retainCanvasNodeSelection } from './agent-canvas-commands.js';

test('viewport/save refresh keeps surviving selected cards and drops deleted IDs', () => {
  const source = [{ id: 'story', position: { x: 100, y: 100 } }, { id: 'scene' }];
  const result = retainCanvasNodeSelection(source, ['story', 'removed']);
  assert.deepEqual(result.filter(node => node.selected).map(node => node.id), ['story']);
  assert.equal(result[1].selected, false);
  assert.equal(source[0].selected, undefined);
  assert.deepEqual(retainCanvasNodeSelection([], ['story']), []);
});

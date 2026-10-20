import test from 'node:test';
import assert from 'node:assert/strict';
import { layoutCanvas } from './agent-canvas-layout.js';

const nodes = [
  { id: 'a', position: { x: 20, y: 40 }, data: {} },
  { id: 'b', position: { x: 20, y: 40 }, data: {} },
  { id: 'c', position: { x: 20, y: 40 }, data: {} },
  { id: 'locked', position: { x: 777, y: 333 }, data: { locked: true } }
];
const edges = [
  { source: 'a', target: 'b' },
  { source: 'b', target: 'c' },
  { source: 'a', target: 'locked' }
];

test('横向自动布局按连接层级展开且不移动锁定节点', () => {
  const result = layoutCanvas(nodes, edges, 'LR');
  assert.ok(result.find(node => node.id === 'b').position.x > result.find(node => node.id === 'a').position.x);
  assert.ok(result.find(node => node.id === 'c').position.x > result.find(node => node.id === 'b').position.x);
  assert.deepEqual(result.find(node => node.id === 'locked').position, { x: 777, y: 333 });
  assert.deepEqual(nodes[1].position, { x: 20, y: 40 });
});

test('纵向自动布局稳定处理没有连线的节点', () => {
  const result = layoutCanvas(nodes.slice(0, 3), [], 'TB');
  assert.deepEqual(result.map(node => node.position), [
    { x: 0, y: 0 },
    { x: 280, y: 0 },
    { x: 560, y: 0 }
  ]);
});

import test from 'node:test';
import assert from 'node:assert/strict';

import { insertDirectorCapture } from './director-capture.js';

const directorNode = { id: 'director-1', position: { x: 100, y: 200 }, data: { type: 'director', config: {} } };

test('导演台截图会变成图片节点并连回导演台', () => {
  const result = insertDirectorCapture({
    nodes: [directorNode],
    edges: [],
    directorNodeID: directorNode.id,
    asset: { id: 'capture-1', url: '/capture.png', name: '构图 1', kind: 'image' }
  });

  assert.equal(result.nodes.filter(node => node.data.type === 'image').length, 1);
  assert.equal(result.edges.length, 1);
  assert.equal(result.edges[0].source, directorNode.id);
  assert.equal(result.edges[0].target, result.nodes.at(-1).id);
});

test('缺少导演台或截图不是图片时，不会产生半成品画布变更', () => {
  const nodes = [directorNode];
  const edges = [];
  assert.throws(() => insertDirectorCapture({ nodes, edges, directorNodeID: 'missing', asset: { id: 'capture-1', kind: 'image' } }), /导演台/);
  assert.throws(() => insertDirectorCapture({ nodes, edges, directorNodeID: directorNode.id, asset: { id: 'capture-1', kind: 'video' } }), /图片/);
  assert.equal(nodes.length, 1);
  assert.equal(edges.length, 0);
});

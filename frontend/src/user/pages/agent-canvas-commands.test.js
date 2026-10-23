import test from 'node:test';
import assert from 'node:assert/strict';
import {
  CANVAS_SHORTCUTS,
  alignCanvasNodes,
  canvasKeyboardCommand,
  canvasGroupBounds,
  commandAvailability,
  connectCanvasNodes,
  createCanvasNode,
  distributeCanvasNodes,
  groupCanvasNodes,
  isCanvasMultiSelectionEvent,
  moveCanvasGroup,
  removeCanvasEdges,
  sameCanvasSelection,
  selectCanvasNodes,
  ungroupCanvasNodes
} from './agent-canvas-commands.js';

const nodes = [
  { id: 'a', position: { x: 20, y: 30 }, measured: { width: 40, height: 20 }, data: {} },
  { id: 'b', position: { x: 100, y: 90 }, measured: { width: 60, height: 30 }, data: {} },
  { id: 'c', position: { x: 220, y: 180 }, measured: { width: 80, height: 40 }, data: {} },
  { id: 'outside', position: { x: 500, y: 500 }, measured: { width: 50, height: 50 }, data: {} }
];

test('self and duplicate connections do not create edges', () => {
  const edges = [{ id: 'edge-a-b-reference', source: 'a', target: 'b', data: { relation: 'reference' } }];
  assert.strictEqual(connectCanvasNodes(edges, { source: 'a', target: 'a' }), edges);
  assert.strictEqual(connectCanvasNodes(edges, { source: 'a', target: 'b' }), edges);
  assert.strictEqual(connectCanvasNodes(edges, { source: '', target: 'b' }), edges);
});

test('valid connection creates one stable reference edge without mutating input', () => {
  const edges = [];
  const next = connectCanvasNodes(edges, { source: 'a', target: 'b' });
  assert.deepEqual(next, [{ id: 'edge-a-b-reference', source: 'a', target: 'b', data: { relation: 'reference' } }]);
  assert.deepEqual(edges, []);
});

test('removing a selected connection preserves every other connection', () => {
  const edges = [
    { id: 'edge-image-video', source: 'image-1', target: 'video-1', data: { relation: 'first_frame' } },
    { id: 'edge-text-image', source: 'text-1', target: 'image-1', data: { relation: 'script_source' } }
  ];

  const remaining = removeCanvasEdges(edges, ['edge-image-video']);

  assert.deepEqual(remaining, [edges[1]]);
  assert.deepEqual(edges.map(edge => edge.id), ['edge-image-video', 'edge-text-image']);
});

test('modified node clicks add and remove nodes from the current selection', () => {
  const selectedA = selectCanvasNodes(nodes, 'a', false);
  assert.deepEqual(selectedA.filter(node => node.selected).map(node => node.id), ['a']);

  const selectedAB = selectCanvasNodes(selectedA, 'b', true);
  assert.deepEqual(selectedAB.filter(node => node.selected).map(node => node.id), ['a', 'b']);

  const selectedB = selectCanvasNodes(selectedAB, 'a', true);
  assert.deepEqual(selectedB.filter(node => node.selected).map(node => node.id), ['b']);
  assert.equal(nodes.some(node => node.selected), false);
});

test('modifier click on node content is reserved for multi-selection', () => {
  assert.equal(isCanvasMultiSelectionEvent({ shiftKey: true }), true);
  assert.equal(isCanvasMultiSelectionEvent({ ctrlKey: true }), true);
  assert.equal(isCanvasMultiSelectionEvent({ metaKey: true }), true);
  assert.equal(isCanvasMultiSelectionEvent({}), false);
});

test('identical React Flow selections do not trigger another state update', () => {
  assert.equal(sameCanvasSelection(['image-1', 'video-1'], ['image-1', 'video-1']), true);
  assert.equal(sameCanvasSelection(['image-1', 'video-1'], ['video-1', 'image-1']), false);
  assert.equal(sameCanvasSelection(['image-1'], []), false);
});

test('alignment supports all six bounds and leaves unselected nodes unchanged', () => {
  const cases = [
    ['left', [{ x: 20, y: 30 }, { x: 20, y: 90 }]],
    ['centerX', [{ x: 70, y: 30 }, { x: 60, y: 90 }]],
    ['right', [{ x: 120, y: 30 }, { x: 100, y: 90 }]],
    ['top', [{ x: 20, y: 30 }, { x: 100, y: 30 }]],
    ['centerY', [{ x: 20, y: 65 }, { x: 100, y: 60 }]],
    ['bottom', [{ x: 20, y: 100 }, { x: 100, y: 90 }]]
  ];
  for (const [mode, expected] of cases) {
    const next = alignCanvasNodes(nodes, ['a', 'b'], mode);
    assert.deepEqual(next.slice(0, 2).map(node => node.position), expected, mode);
    assert.deepEqual(next[2].position, nodes[2].position, `${mode} moved an unselected node`);
  }
  assert.equal(nodes[0].position.x, 20);
});

test('horizontal distribution keeps endpoints and creates equal outer gaps', () => {
  const next = distributeCanvasNodes(nodes, ['a', 'b', 'c'], 'horizontal');
  assert.deepEqual(next.slice(0, 3).map(node => node.position.x), [20, 110, 220]);
  assert.deepEqual(next[3].position, nodes[3].position);
});

test('vertical distribution keeps endpoints and creates equal outer gaps', () => {
  const next = distributeCanvasNodes(nodes, ['a', 'b', 'c'], 'vertical');
  assert.deepEqual(next.slice(0, 3).map(node => node.position.y), [30, 100, 180]);
});

test('distribution with fewer than three selected nodes is a no-op', () => {
  assert.strictEqual(distributeCanvasNodes(nodes, ['a', 'b'], 'horizontal'), nodes);
});

test('group and ungroup only change selected members', () => {
  const grouped = groupCanvasNodes(nodes, ['a', 'b'], 'group-1');
  assert.equal(grouped[0].data.groupId, 'group-1');
  assert.equal(grouped[1].data.groupId, 'group-1');
  assert.equal(grouped[2].data.groupId, undefined);
  assert.equal(nodes[0].data.groupId, undefined);

  const ungrouped = ungroupCanvasNodes(grouped, ['a']);
  assert.equal(ungrouped[0].data.groupId, undefined);
  assert.equal(ungrouped[1].data.groupId, 'group-1');
});

test('group bounds include node dimensions and sixteen pixel padding', () => {
  const grouped = groupCanvasNodes(nodes, ['a', 'b'], 'group-1');
  assert.deepEqual(canvasGroupBounds(grouped, 'group-1'), {
    x: 4,
    y: 14,
    width: 172,
    height: 122
  });
});

test('command availability explains selection requirements', () => {
  const empty = commandAvailability(nodes, []);
  assert.deepEqual(empty, {
    canGroup: false,
    canUngroup: false,
    canAlign: false,
    canDistribute: false,
    groupReason: '至少选择 2 个节点',
    ungroupReason: '请选择已分组节点',
    alignReason: '至少选择 2 个节点',
    distributeReason: '至少选择 3 个节点'
  });

  const grouped = groupCanvasNodes(nodes, ['a', 'b'], 'group-1');
  const three = commandAvailability(grouped, ['a', 'b', 'c']);
  assert.equal(three.canGroup, true);
  assert.equal(three.canUngroup, true);
  assert.equal(three.canAlign, true);
  assert.equal(three.canDistribute, true);
});

test('keyboard command recognizes every supported canvas shortcut outside editable controls', () => {
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'a', targetTagName: 'div' }), 'select_all');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'c', targetTagName: 'div' }), 'copy');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'v', targetTagName: 'div' }), 'paste');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'd', targetTagName: 'div' }), 'duplicate');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'g', targetTagName: 'div' }), 'group');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, shiftKey: true, key: 'g', targetTagName: 'div' }), 'ungroup');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'z', targetTagName: 'div' }), 'undo');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, shiftKey: true, key: 'z', targetTagName: 'div' }), 'redo');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'y', targetTagName: 'div' }), 'redo');
  assert.equal(canvasKeyboardCommand({ key: 'Delete', targetTagName: 'div' }), 'delete');
  assert.equal(canvasKeyboardCommand({ shiftKey: true, key: '1', targetTagName: 'div' }), 'fit_canvas');
  assert.equal(canvasKeyboardCommand({ shiftKey: true, key: '2', targetTagName: 'div' }), 'focus_selection');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: '+', targetTagName: 'div' }), 'zoom_in');
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: '-', targetTagName: 'div' }), 'zoom_out');
  assert.equal(canvasKeyboardCommand({ metaKey: true, key: 'z', targetTagName: 'textarea' }), null);
  assert.equal(canvasKeyboardCommand({ ctrlKey: true, key: 'z', isContentEditable: true }), null);
});

test('快捷键帮助只展示键盘处理器已经支持的命令', () => {
  const documented = new Set(CANVAS_SHORTCUTS.map(item => item.command));
  for (const command of ['undo', 'redo', 'copy', 'paste', 'duplicate', 'delete', 'group', 'ungroup', 'fit_canvas', 'focus_selection', 'zoom_in', 'zoom_out']) {
    assert.equal(documented.has(command), true, command);
  }
  assert.equal(documented.has('auto_layout'), false);
});

test('new canvas node has stable persistent defaults at requested position', () => {
  assert.deepEqual(createCanvasNode('node-9', { x: 120, y: 240 }), {
    id: 'node-9',
    type: 'agent',
    position: { x: 120, y: 240 },
    selected: true,
    data: {
      type: 'text',
      title: '新节点',
      summary: '',
      version: 1,
      sourceNodeIds: [],
      status: 'draft',
      assetRef: null
    }
  });
});

test('moving a group shifts only members by pointer delta', () => {
  const grouped = groupCanvasNodes(nodes, ['a', 'b'], 'group-1');
  const moved = moveCanvasGroup(grouped, 'group-1', { x: 15, y: -10 });
  assert.deepEqual(moved[0].position, { x: 35, y: 20 });
  assert.deepEqual(moved[1].position, { x: 115, y: 80 });
  assert.deepEqual(moved[2].position, nodes[2].position);
});

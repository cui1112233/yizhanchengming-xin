import test from 'node:test';
import assert from 'node:assert/strict';
import {
  buildWorkspaceGroupTree,
  filterWorkspaceAssets,
  filterWorkspaceNodes,
  workspaceDrawerWidth
} from './agent-workspace-navigation.js';

const nodes = [
  { id: 'text-1', type: 'text', title: '雨夜开场', summary: '主角走进车站', status: 'done', groupId: 'group-a' },
  { id: 'image-2', type: 'image', title: '车站全景', summary: '雨夜里的空车站', status: 'draft', groupId: 'group-a', config: { rating: 5 } },
  { id: 'video-3', data: { type: 'video', title: '列车进站', summary: '清晨镜头', status: 'failed', config: { rating: 2 } } }
];

test('节点筛选会同时检查标题摘要类型状态和评分', () => {
  assert.deepEqual(
    filterWorkspaceNodes(nodes, { query: '雨夜', kinds: ['image'], statuses: ['draft'], ratings: [5] }).map(item => item.id),
    ['image-2']
  );
  assert.deepEqual(filterWorkspaceNodes(nodes, { query: '列车', kinds: [], statuses: [], ratings: [] }).map(item => item.id), ['video-3']);
});

test('筛选后的节点仍保留所属分组并单列未分组节点', () => {
  const tree = buildWorkspaceGroupTree(nodes, [{ id: 'group-a', title: '第一幕' }]);
  assert.deepEqual(tree.map(group => [group.id, group.children.map(node => node.id)]), [
    ['group-a', ['text-1', 'image-2']],
    ['ungrouped', ['video-3']]
  ]);
});

test('素材搜索支持名称类型且侧栏宽度限制在可用范围', () => {
  const assets = [
    { id: 'asset-1', kind: 'image', name: '雨夜车站.png' },
    { id: 'asset-2', kind: 'video', name: '清晨列车.mp4' }
  ];
  assert.deepEqual(filterWorkspaceAssets(assets, { query: '雨夜', kinds: ['image'] }).map(item => item.id), ['asset-1']);
  assert.equal(workspaceDrawerWidth(120), 240);
  assert.equal(workspaceDrawerWidth(620), 460);
  assert.equal(workspaceDrawerWidth(336.4), 336);
});

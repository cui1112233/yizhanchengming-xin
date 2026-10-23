import test from 'node:test';
import assert from 'node:assert/strict';
import { directorPlanPresentation, isDirectorPlanExecutable } from './agent-director-commands.js';

test('导演方案把画布命令转换成用户能看懂的确认清单', () => {
  const view = directorPlanPresentation({
    id: 'plan-1',
    summary: '建立开场镜头',
    commands: [
      { type: 'create_node', node: { id: 'image-1', type: 'image', title: '开场图' } },
      { type: 'update_node', nodeId: 'text-1', patch: { title: '开场文案' } },
      { type: 'connect_nodes', sourceNodeId: 'image-1', targetNodeId: 'text-1', relation: 'reference' },
      { type: 'remove_nodes', nodeIds: ['old-1'] }
    ]
  });
  assert.equal(view.title, '导演方案');
  assert.equal(view.summary, '建立开场镜头');
  assert.deepEqual(view.items.map(item => item.label), ['创建图片节点「开场图」', '修改节点「text-1」', '连接「image-1」到「text-1」', '移除 1 个画布节点']);
  assert.equal(view.items.at(-1).destructive, true);
});

test('包含真实任务命令的方案要交给执行接口而不是假装已完成', () => {
  assert.equal(isDirectorPlanExecutable({ commands: [{ type: 'create_node' }] }), true);
  assert.equal(isDirectorPlanExecutable({ commands: [{ type: 'start_generation', nodeId: 'image-1' }] }), false);
  assert.equal(isDirectorPlanExecutable({ commands: [{ type: 'apply_result', nodeId: 'image-1' }] }), false);
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { compatibleModels, reconcileNodeModel } from './agent-model-capabilities.js';

const models = [
  { id: 'img-1', kind: 'image', displayName: '图片 A', provider: 'personal_api', capabilities: { resolutions: ['2K'] } },
  { id: 'img-comfy', kind: 'image', displayName: 'Comfy 图片', provider: 'autodl_comfyui' },
  { id: 'video-1', kind: 'video', displayName: '视频 A', provider: 'personal_api' }
];

test('节点只显示同类型和当前渠道的模型', () => {
  assert.deepEqual(compatibleModels(models, { kind: 'image', provider: 'personal_api' }).map(item => item.id), ['img-1']);
  assert.deepEqual(compatibleModels(models, { kind: 'video', provider: 'personal_api' }).map(item => item.id), ['video-1']);
});

test('已经保存但下线的模型会明确显示不可用而不是静默替换', () => {
  assert.deepEqual(reconcileNodeModel({ provider: 'personal_api', modelId: 'removed' }, models, 'image'), {
    provider: 'personal_api', modelId: 'removed', model: null, available: false, reason: '当前模型不可用'
  });
});

test('豆包执行器模型按节点类型进入同一模型列表', () => {
  const local = compatibleModels(models, { kind: 'image', provider: 'doubao_local_executor' });
  assert.ok(local.some(item => item.id === 'Seedream 4.5' && item.provider === 'doubao_local_executor'));
  assert.equal(local.some(item => item.kind !== 'image'), false);
});

test('豆包图片执行器只公布实际可执行的参数', () => {
  const model = compatibleModels(models, { kind: 'image', provider: 'doubao_local_executor' })[0];
  assert.deepEqual(model.capabilities.qualityLevels, ['low', 'standard', 'high']);
  assert.equal(model.capabilities.resolutionManagedByModel, true);
  assert.equal(model.capabilities.maxOutputCount, 4);
});

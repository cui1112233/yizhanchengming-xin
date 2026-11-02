import test from 'node:test';
import assert from 'node:assert/strict';
import { compatibleModels, reconcileNodeModel, modelInputCapabilityError } from './agent-model-capabilities.js';

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

test('批量工厂投影的视频节点会识别专用执行器字段', () => {
  const state = reconcileNodeModel({
    videoProvider: 'doubao_local_executor',
    videoModelId: 'Seedance 2.0 Fast'
  }, models, 'video');
  assert.equal(state.provider, 'doubao_local_executor');
  assert.equal(state.modelId, 'Seedance 2.0 Fast');
  assert.equal(state.available, true);
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
	assert.equal(model.capabilities.acceptsTextPrompt, true);
	assert.equal(model.capabilities.acceptsImageReferences, true);
	assert.equal(model.capabilities.acceptsVideoReferences, false);
	assert.equal(model.capabilities.acceptsAudioReferences, false);
	assert.equal(model.capabilities.maxReferenceAssets, 10);
});

test('豆包视频执行器显式拒绝未实现的音频和视频参考', () => {
	const model = compatibleModels(models, { kind: 'video', provider: 'doubao_local_executor' })[0];
	assert.equal(model.capabilities.acceptsTextPrompt, true);
	assert.equal(model.capabilities.acceptsImageReferences, true);
	assert.equal(model.capabilities.acceptsVideoReferences, false);
	assert.equal(model.capabilities.acceptsAudioReferences, false);
	assert.equal(model.capabilities.maxReferenceAssets, 10);
});

test('模型能力会在提交前拒绝不支持的参考素材', () => {
	const model = { capabilities: { acceptsTextPrompt: true, acceptsImageReferences: true, acceptsVideoReferences: false, acceptsAudioReferences: false, maxReferenceAssets: 1 } };
	assert.equal(modelInputCapabilityError(model, { prompt: '雨夜追逐', imageReferences: 1 }), '');
	assert.match(modelInputCapabilityError(model, { prompt: '雨夜追逐', imageReferences: 2 }), /最多支持 1 个参考素材/);
	assert.match(modelInputCapabilityError(model, { prompt: '雨夜追逐', videoReferences: 1 }), /不支持视频参考/);
	assert.match(modelInputCapabilityError(model, { prompt: '雨夜追逐', audioReferences: 1 }), /不支持音频参考/);
});

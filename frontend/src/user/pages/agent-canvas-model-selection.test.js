import test from 'node:test';
import assert from 'node:assert/strict';
import { inheritedModelSelection, shareModelSelection } from './agent-canvas-model-selection.js';

test('共享视频模型时保留每个节点自己的提示词和参数', () => {
  const nodes = [
    { id: 'video-1', data: { type: 'video', config: { provider: 'personal_api', modelId: 'old-video', prompt: '镜头一', parameters: { ratio: '16:9', durationSeconds: 5 } } } },
    { id: 'video-2', data: { type: 'video', config: { provider: 'autodl_comfyui', modelId: 'old-h3', prompt: '镜头二', parameters: { ratio: '9:16', durationSeconds: 10 } } } },
    { id: 'image-1', data: { type: 'image', config: { provider: 'personal_api', modelId: 'image-a', prompt: '封面' } } }
  ];

  const result = shareModelSelection(nodes, 'video-1', { provider: 'doubao_local_executor', modelId: 'seedance-2-0', model: 'seedance-2-0' });

  assert.deepEqual(result.map(node => node.data.config.provider), ['doubao_local_executor', 'doubao_local_executor', 'personal_api']);
  assert.deepEqual(result.map(node => node.data.config.modelId), ['seedance-2-0', 'seedance-2-0', 'image-a']);
  assert.equal(result[1].data.config.prompt, '镜头二');
  assert.equal(result[1].data.config.parameters.ratio, '9:16');
  assert.equal(nodes[1].data.config.provider, 'autodl_comfyui');
});

test('新建同类节点继承画布最后选择的模型', () => {
  const selection = inheritedModelSelection([
    { id: 'video-1', data: { type: 'video', config: { provider: 'personal_api', modelId: 'seedance-1' } } },
    { id: 'video-2', data: { type: 'video', config: { provider: 'doubao_local_executor', modelId: 'seedance-2-0' } } },
    { id: 'image-1', data: { type: 'image', config: { provider: 'personal_api', modelId: 'image-a' } } }
  ], 'video');

  assert.deepEqual(selection, { provider: 'doubao_local_executor', modelId: 'seedance-2-0', model: 'seedance-2-0' });
});

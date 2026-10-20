import test from 'node:test';
import assert from 'node:assert/strict';

import { createAgentImageGeneration, createAgentTextGeneration } from '../../shared/api/agent.js';
import {
  agentCanvasModelOptions,
  composerModelSelection,
  composerShouldClose,
  nodeGenerationIdempotencyKey,
  nodeGenerationViewState,
  submitNodeGeneration
} from './agent-canvas-node-generation.js';

test('Agent canvas keeps local Doubao image and video models available first', () => {
  assert.deepEqual(agentCanvasModelOptions([], 'image'), [
    { value: 'local-doubao-executor-image', label: '豆包本地执行器（图片）' }
  ]);
  assert.deepEqual(agentCanvasModelOptions([
    { value: 'minimax-h3-video', label: 'MiniMax H3 多图生视频' },
    { value: 'local-doubao-executor-video', label: '重复项' }
  ], 'video'), [
    { value: 'local-doubao-executor-video', label: '豆包本地执行器（视频）' },
    { value: 'minimax-h3-video', label: 'MiniMax H3 多图生视频' }
  ]);
});

test('composer submit stops the canvas event before starting generation', () => {
  const calls = [];
  submitNodeGeneration({
    event: { stopPropagation: () => calls.push('stop') },
    onSubmit: (nodeID, modelID) => calls.push(['submit', nodeID, modelID]),
    nodeID: 'image-1',
    modelID: 'local-doubao-executor-image'
  });

  assert.deepEqual(calls, [
    'stop',
    ['submit', 'image-1', 'local-doubao-executor-image']
  ]);
});

test('text and image generation clients use Agent project routes', async () => {
  const calls = [];
  const request = async (path, options) => {
    calls.push({ path, options });
    return { ok: true };
  };
  const payload = { nodeId: 'node-1', nodeVersion: 2, modelId: 'model-1', prompt: '生成内容' };
  await createAgentTextGeneration('project 1', payload, request);
  await createAgentImageGeneration('project 1', payload, request);
  assert.equal(calls[0].path, '/api/agent/tasks/project%201/text-generations');
  assert.equal(calls[1].path, '/api/agent/tasks/project%201/image-generations');
  assert.equal(calls[0].options.method, 'POST');
  assert.deepEqual(JSON.parse(calls[1].options.body), payload);
});

test('composer model selection preserves a valid exact model and otherwise falls back', () => {
  const models = [
    { id: 'text-a', kind: 'text', displayName: '文字 A', enabled: true },
    { id: 'text-b', kind: 'text', displayName: '文字 B', enabled: true },
    { id: 'image-a', kind: 'image', displayName: '图片 A', enabled: true },
    { id: 'text-off', kind: 'text', enabled: false }
  ];
  assert.equal(composerModelSelection('text', { modelId: 'text-b' }, models).modelId, 'text-b');
  assert.equal(composerModelSelection('text', { modelId: 'missing' }, models).modelId, 'text-a');
  assert.equal(composerModelSelection('video', { modelId: 'text-a' }, models).disabled, true);
});

test('generation view state exposes busy and readable error states', () => {
  assert.deepEqual(nodeGenerationViewState({ kind: 'image', busy: true }), {
    busy: true, disabled: true, tone: 'running', message: '正在生成图片…'
  });
  assert.deepEqual(nodeGenerationViewState({ kind: 'text', error: new Error('模型余额不足') }), {
    busy: false, disabled: false, tone: 'error', message: '模型余额不足'
  });
});

test('generation idempotency changes with node version but remains stable for the same attempt', () => {
  const first = nodeGenerationIdempotencyKey({ taskId: 'project-1', nodeId: 'node-1', nodeVersion: 1, kind: 'image', attempt: 0 });
  const repeated = nodeGenerationIdempotencyKey({ taskId: 'project-1', nodeId: 'node-1', nodeVersion: 1, kind: 'image', attempt: 0 });
  const changed = nodeGenerationIdempotencyKey({ taskId: 'project-1', nodeId: 'node-1', nodeVersion: 2, kind: 'image', attempt: 0 });
  assert.equal(first, repeated);
  assert.notEqual(first, changed);
});

test('composer closes for canvas lifecycle events but not inside interactions', () => {
  assert.equal(composerShouldClose({ paneClick: true }), true);
  assert.equal(composerShouldClose({ escape: true }), true);
  assert.equal(composerShouldClose({ switchedNode: true }), true);
  assert.equal(composerShouldClose({ insideComposer: true, paneClick: true }), false);
  assert.equal(composerShouldClose({ insidePortal: true, escape: false }), false);
});

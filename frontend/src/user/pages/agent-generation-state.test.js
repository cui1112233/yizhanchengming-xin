import test from 'node:test';
import assert from 'node:assert/strict';

import {
	activateAgentGenerationResult,
  cancelAgentGeneration,
  confirmAgentGeneration,
  createAgentGeneration,
  getAgentGeneration,
  listAgentGenerations,
  retryAgentGeneration
} from '../../shared/api/agent.js';
import {
  activeGenerationTaskForNode,
  generationTaskActions,
  generationPresentation,
  generationSummaryText,
  latestGenerationTaskForNode,
  mergeGenerationTask,
  mergeLatestGenerationTasksByNode,
  shouldAcceptGenerationResponse,
  shouldPollGeneration
} from './agent-generation-state.js';

test('generationTaskActions exposes recovery controls without bypassing legacy video jobs', () => {
  assert.deepEqual(generationTaskActions({ state: 'running', kind: 'image' }), { canCancel: true, canRetry: false, openCanvasOnly: false });
  assert.deepEqual(generationTaskActions({ state: 'failed', kind: 'audio' }), { canCancel: false, canRetry: true, openCanvasOnly: false });
  assert.deepEqual(generationTaskActions({ state: 'failed', kind: 'video', mode: 'legacy' }), { canCancel: false, canRetry: false, openCanvasOnly: true });
  assert.deepEqual(generationTaskActions({ state: 'failed', kind: 'video', mode: 'video_image_to_video' }), { canCancel: false, canRetry: true, openCanvasOnly: false });
  assert.deepEqual(generationTaskActions({ state: 'awaiting_confirmation', kind: 'image' }), { canCancel: false, canRetry: false, openCanvasOnly: true });
});

test('画布任务汇总不把历史失败当成当前执行状态', () => {
  assert.equal(generationSummaryText({ queued: 2, running: 1, awaitingConfirmation: 3, completed: 8, failed: 21, cancelled: 1 }), '排队 2 · 生成中 1 · 待确认 3');
  assert.equal(generationSummaryText({ completed: 8, failed: 21, cancelled: 1 }), '暂无进行中任务');
});

test('重复提交时能找回同节点正在执行的任务状态', () => {
  const task = activeGenerationTaskForNode([
    { id: 'old', nodeId: 'image-1', kind: 'image', state: 'failed', updatedAt: '2026-10-22T08:00:00Z' },
    { id: 'running', nodeId: 'image-1', kind: 'image', state: 'running', updatedAt: '2026-10-22T08:02:00Z' },
    { id: 'other', nodeId: 'image-2', kind: 'image', state: 'running', updatedAt: '2026-10-22T08:03:00Z' }
  ], 'image-1', 'image');

  assert.equal(task.id, 'running');
  assert.equal(task.state, 'running');
});

test('重复提交已失败的图片时能找回真实失败任务', () => {
  const task = latestGenerationTaskForNode([
    { id: 'old', nodeId: 'image-1', kind: 'image', state: 'completed', updatedAt: '2026-10-22T08:00:00Z' },
    { id: 'failed', nodeId: 'image-1', kind: 'image', state: 'failed', errorMessage: '参考图不可用', updatedAt: '2026-10-22T08:02:00Z' },
    { id: 'other', nodeId: 'image-2', kind: 'image', state: 'failed', updatedAt: '2026-10-22T08:03:00Z' }
  ], 'image-1', 'image');

  assert.equal(task.id, 'failed');
  assert.equal(generationPresentation(task).message, '参考图不可用');
});

test('等待确认会保留全部结果和服务端已选结果', () => {
  const node = { id: 'image-1', data: { type: 'image', version: 3 } };
  const results = [
    { id: 'r1', assetId: 'asset-1', kind: 'image', asset: { id: 'asset-1', kind: 'image', url: '/one.png' } },
    { id: 'r2', assetId: 'asset-2', kind: 'image', asset: { id: 'asset-2', kind: 'image', url: '/two.png' } }
  ];
  const state = mergeGenerationTask(node, {
    id: 'generation-1', projectId: 'project-1', nodeId: 'image-1', nodeVersion: 3,
    kind: 'image', state: 'awaiting_confirmation', results, selectedResultId: 'r2'
  });

  assert.equal(state.previewAssetId, 'asset-2');
  assert.equal(state.selectedResultId, 'r2');
  assert.deepEqual(state.resultOptions.map(result => result.id), ['r1', 'r2']);
  assert.equal(generationPresentation(state.task).primaryAction, '确认使用');
});

test('本地候选选择只影响预览，不篡改服务端任务', () => {
  const task = {
    id: 'generation-1', nodeId: 'image-1', nodeVersion: 3, kind: 'image', state: 'awaiting_confirmation',
    selectedResultId: 'r1', results: [
      { id: 'r1', assetId: 'asset-1', kind: 'image' },
      { id: 'r2', assetId: 'asset-2', kind: 'image' }
    ]
  };
  const state = mergeGenerationTask({ id: 'image-1', data: { type: 'image', version: 3 } }, task, 'r2');

  assert.equal(state.selectedResultId, 'r2');
  assert.equal(state.previewAssetId, 'asset-2');
  assert.equal(task.selectedResultId, 'r1');
});

test('每个节点只保留当前项目最新的生成任务', () => {
  const merged = mergeLatestGenerationTasksByNode({
    'image-1': { id: 'old-local', projectId: 'project-1', nodeId: 'image-1', updatedAt: '2026-10-21T10:00:00Z' }
  }, [
    { id: 'other-project', projectId: 'project-2', nodeId: 'image-1', updatedAt: '2026-10-21T12:00:00Z' },
    { id: 'new-current', projectId: 'project-1', nodeId: 'image-1', state: 'completed', updatedAt: '2026-10-21T11:00:00Z' },
    { id: 'text-current', projectId: 'project-1', nodeId: 'text-1', state: 'running', updatedAt: '2026-10-21T11:30:00Z' }
  ], 'project-1');

  assert.equal(merged['image-1'].id, 'new-current');
  assert.equal(merged['text-1'].id, 'text-current');
  assert.equal(Object.values(merged).some(task => task.projectId === 'project-2'), false);
});

test('旧请求晚到的失败不会覆盖较新的运行尝试', () => {
  const merged = mergeLatestGenerationTasksByNode({}, [
    { id: 'older-failed', projectId: 'project-1', nodeId: 'script-1', state: 'failed', createdAt: '2026-10-23T09:00:00Z', updatedAt: '2026-10-23T09:10:00Z' },
    { id: 'newer-running', projectId: 'project-1', nodeId: 'script-1', state: 'running', createdAt: '2026-10-23T09:05:00Z', updatedAt: '2026-10-23T09:05:00Z' }
  ], 'project-1');

  assert.equal(merged['script-1'].id, 'newer-running');
  assert.equal(merged['script-1'].state, 'running');
});

test('同一个生成任务的实时状态会从排队更新为失败', () => {
  const merged = mergeLatestGenerationTasksByNode({
    'image-1': {
      id: 'generation-1', projectId: 'project-1', nodeId: 'image-1', state: 'queued',
      createdAt: '2026-10-26T04:40:21Z', updatedAt: '2026-10-26T04:40:21Z'
    }
  }, [{
    id: 'generation-1', projectId: 'project-1', nodeId: 'image-1', state: 'failed', errorMessage: '豆包图片任务失败',
    createdAt: '2026-10-26T04:40:21Z', updatedAt: '2026-10-26T04:40:47Z'
  }], 'project-1');

  assert.equal(merged['image-1'].state, 'failed');
  assert.equal(merged['image-1'].errorMessage, '豆包图片任务失败');
});

test('只有排队中和生成中的通用任务需要轮询', () => {
  assert.equal(shouldPollGeneration({ state: 'queued' }), true);
  assert.equal(shouldPollGeneration({ state: 'running' }), true);
  assert.equal(shouldPollGeneration({ state: 'awaiting_confirmation' }), false);
  assert.equal(shouldPollGeneration({ state: 'completed' }), false);
  assert.equal(generationPresentation({ state: 'failed', errorMessage: '余额不足' }).message, '余额不足');
});

test('切换项目或卸载后拒绝旧的任务响应', () => {
  assert.equal(shouldAcceptGenerationResponse({ requestedTaskId: 'p1', currentTaskId: 'p1', active: true }), true);
  assert.equal(shouldAcceptGenerationResponse({ requestedTaskId: 'p1', currentTaskId: 'p2', active: true }), false);
  assert.equal(shouldAcceptGenerationResponse({ requestedTaskId: 'p1', currentTaskId: 'p1', active: false }), false);
});

test('通用生成 API 使用项目级任务路径和明确动作载荷', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push({ path, options });
    return { task: { id: 'generation-1' } };
  };
  await createAgentGeneration('project 1', { nodeId: 'image-1' }, request);
  await listAgentGenerations('project 1', request);
	await getAgentGeneration('project 1', 'generation 1', request);
	await confirmAgentGeneration('project 1', 'generation 1', 'result-2', request);
	await activateAgentGenerationResult('project 1', 'generation 1', 'result-1', 3, request);
	await cancelAgentGeneration('project 1', 'generation 1', request);
  await retryAgentGeneration('project 1', 'generation 1', 'retry-key', request);

  assert.deepEqual(calls.map(call => call.path), [
    '/api/agent/tasks/project%201/generations',
    '/api/agent/tasks/project%201/generations',
    '/api/agent/tasks/project%201/generations/generation%201',
    '/api/agent/tasks/project%201/generations/generation%201/confirm',
		'/api/agent/tasks/project%201/generations/generation%201/activate',
    '/api/agent/tasks/project%201/generations/generation%201/cancel',
    '/api/agent/tasks/project%201/generations/generation%201/retry'
  ]);
  assert.equal(calls[0].options.method, 'POST');
	assert.deepEqual(JSON.parse(calls[3].options.body), { resultId: 'result-2' });
	assert.deepEqual(JSON.parse(calls[4].options.body), { resultId: 'result-1', expectedNodeVersion: 3 });
	assert.deepEqual(JSON.parse(calls[6].options.body), { idempotencyKey: 'retry-key' });
});

import test from 'node:test';
import assert from 'node:assert/strict';
import * as videoJobs from './agent-canvas-video-jobs.js';

const { canSubmitVideoFromConfig, collectVideoReferences, videoNodeViewState, videoReferenceAssetIDs, videoReferenceCapabilityError, videoResultPreview } = videoJobs;

test('video creation requires a non-empty prompt from the inline node settings', () => {
  assert.equal(canSubmitVideoFromConfig({ prompt: '' }), false);
  assert.equal(canSubmitVideoFromConfig({ prompt: '   ' }), false);
  assert.equal(canSubmitVideoFromConfig({ prompt: '夜雨中的古城追逐镜头' }), true);
});

test('collectVideoReferences keeps only directly connected confirmed image assets', () => {
  const refs = collectVideoReferences({
    videoNodeId: 'video-1',
    nodes: [
      { id: 'image-1', data: { type: 'image', status: 'confirmed', assetRef: { id: 'asset-1', kind: 'image' } } },
      { id: 'image-2', data: { type: 'image', status: 'draft', assetRef: { id: 'asset-2', kind: 'image' } } },
      { id: 'text-1', data: { type: 'text', status: 'confirmed' } }
    ],
    edges: [{ source: 'image-1', target: 'video-1' }, { source: 'image-2', target: 'video-1' }, { source: 'text-1', target: 'video-1' }]
  });
  assert.deepEqual(refs.map(item => item.id), ['asset-1']);
});

test('videoReferenceAssetIDs keeps every selected media ID so video references cannot disappear silently', () => {
  assert.deepEqual(videoReferenceAssetIDs({ referenceAssets: [
    { id: 'manual-image', kind: 'image' },
    { id: 'connected-image', kind: 'image' },
    { id: 'manual-image', kind: 'image' },
    { id: 'video-asset', kind: 'video' },
    { id: '', kind: 'image' }
  ] }), ['manual-image', 'connected-image', 'video-asset']);
});

test('video references disclose the unsupported video-to-video capability before submission', () => {
  assert.equal(typeof videoReferenceCapabilityError, 'function');
  if (typeof videoReferenceCapabilityError !== 'function') return;
  assert.equal(videoReferenceCapabilityError({ referenceAssets: [
    { id: 'image-asset', kind: 'image' }
  ] }), '');
  assert.match(videoReferenceCapabilityError({ referenceAssets: [
    { id: 'video-asset', kind: 'video', name: '上游成片.mp4' }
  ] }), /暂不支持视频参考素材/);
  assert.match(videoReferenceCapabilityError({ referenceAssets: [
    { id: 'audio-asset', kind: 'audio', name: '旁白.mp3' }
  ] }), /暂不支持音频参考素材/);
});

test('videoNodeViewState exposes confirmation instead of auto-applying result', () => {
  const state = videoNodeViewState({ state: 'awaiting_confirmation' });
  assert.equal(state.canConfirm, true);
  assert.equal(state.canGenerate, false);
  assert.equal(state.polling, false);
  assert.equal(state.label, '等待确认');
});

test('videoNodeViewState translates provider lifecycle into readable canvas states', () => {
	const submitting = videoNodeViewState({ state: 'submitting' });
	assert.equal(submitting.label, '正在提交');
	assert.equal(submitting.polling, true);
	assert.equal(submitting.canGenerate, false);
	assert.equal(submitting.canConfirm, false);
	assert.equal(submitting.canCancel, false);
	assert.equal(videoNodeViewState({ state: 'queued' }).label, '正在排队');
  assert.equal(videoNodeViewState({ state: 'running' }).label, '正在生成视频…');
  assert.equal(videoNodeViewState({ state: 'failed' }).label, '生成失败');
  assert.equal(videoNodeViewState({ state: 'applied' }).label, '已完成');
});

test('submitting ranks after preparation and before a queued job at the same timestamp', () => {
  const at = '2026-10-21T20:30:00Z';
  const merged = videoJobs.mergeLatestVideoJobsByNode({
    video: { id: 'same', projectId: 'project-1', nodeId: 'video', state: 'preparing', updatedAt: at }
  }, [{ id: 'same', projectId: 'project-1', nodeId: 'video', state: 'submitting', updatedAt: at }], 'project-1');
  assert.equal(merged.video.state, 'submitting');
  const queued = videoJobs.mergeLatestVideoJobsByNode(merged, [{ id: 'same', projectId: 'project-1', nodeId: 'video', state: 'queued', updatedAt: at }], 'project-1');
  assert.equal(queued.video.state, 'queued');
});

test('video result ingestion remains visible as an active polling state', () => {
  const state = videoNodeViewState({ state: 'ingesting' });
  assert.equal(state.polling, true);
  assert.equal(state.canGenerate, false);
  assert.equal(state.canConfirm, false);
});

test('video cancellation claim remains visible as an active state', () => {
  const state = videoNodeViewState({ state: 'canceling' });
  assert.equal(state.polling, true);
  assert.equal(state.canCancel, false);
  assert.equal(state.canGenerate, false);
});

test('videoResultPreview exposes only a playable public video reference', () => {
  assert.deepEqual(videoResultPreview({ resultAsset: {
    id: 'result-1',
    kind: 'video',
    name: '成片.mp4',
    url: ' /api/agent/tasks/project-1/assets/result-1/content ',
    objectKey: 'private/alice/project-1/result-1.mp4',
    credentials: 'never-return-this'
  } }), {
    id: 'result-1',
    kind: 'video',
    name: '成片.mp4',
    url: '/api/agent/tasks/project-1/assets/result-1/content'
  });
  assert.equal(videoResultPreview({ resultAsset: { id: 'result-1', kind: 'video' } }), null);
  assert.equal(videoResultPreview({ resultAsset: { id: 'result-1', kind: 'image', url: '/image' } }), null);
});

test('restored video jobs keep only the newest job for each node in the active project', () => {
  assert.equal(typeof videoJobs.mergeLatestVideoJobsByNode, 'function');
  if (typeof videoJobs.mergeLatestVideoJobsByNode !== 'function') return;
  const restored = videoJobs.mergeLatestVideoJobsByNode({}, [
    { id: 'new-confirm', projectId: 'project-1', nodeId: 'video-1', state: 'awaiting_confirmation', updatedAt: '2026-10-20T02:05:00Z' },
    { id: 'old-failed', projectId: 'project-1', nodeId: 'video-1', state: 'failed', updatedAt: '2026-10-20T02:00:00Z' },
    { id: 'retryable', projectId: 'project-1', nodeId: 'video-2', state: 'failed', errorMessage: 'provider failed', updatedAt: '2026-10-20T02:03:00Z' },
    { id: 'foreign', projectId: 'project-2', nodeId: 'video-3', state: 'running', updatedAt: '2026-10-20T02:06:00Z' }
  ], 'project-1');
  assert.deepEqual(Object.fromEntries(Object.entries(restored).map(([nodeID, job]) => [nodeID, job.id])), {
    'video-1': 'new-confirm',
    'video-2': 'retryable'
  });
  assert.equal(videoNodeViewState(restored['video-1']).canConfirm, true);
  assert.equal(videoNodeViewState(restored['video-2']).canRetry, true);
});

test('stale list and poll responses cannot overwrite a newer local job or another project', () => {
  assert.equal(typeof videoJobs.mergeLatestVideoJobsByNode, 'function');
  assert.equal(typeof videoJobs.shouldAcceptVideoJobResponse, 'function');
  if (typeof videoJobs.mergeLatestVideoJobsByNode !== 'function' || typeof videoJobs.shouldAcceptVideoJobResponse !== 'function') return;
  const current = { 'video-1': { id: 'local-new', projectId: 'project-1', nodeId: 'video-1', state: 'queued', updatedAt: '2026-10-20T02:10:00Z' } };
  const merged = videoJobs.mergeLatestVideoJobsByNode(current, [
    { id: 'stale-list', projectId: 'project-1', nodeId: 'video-1', state: 'failed', updatedAt: '2026-10-20T02:05:00Z' }
  ], 'project-1');
  assert.equal(merged['video-1'].id, 'local-new');
  const progressed = videoJobs.mergeLatestVideoJobsByNode(current, [
    { id: 'local-new', projectId: 'project-1', nodeId: 'video-1', state: 'awaiting_confirmation', updatedAt: '2026-10-20T02:10:00Z' }
  ], 'project-1');
  assert.equal(progressed['video-1'].state, 'awaiting_confirmation');
  assert.equal(videoJobs.shouldAcceptVideoJobResponse({ requestedTaskId: 'project-1', currentTaskId: 'project-2', active: true }), false);
  assert.equal(videoJobs.shouldAcceptVideoJobResponse({ requestedTaskId: 'project-1', currentTaskId: 'project-1', active: false }), false);
  assert.equal(videoJobs.shouldAcceptVideoJobResponse({ requestedTaskId: 'project-1', currentTaskId: 'project-1', active: true }), true);
});

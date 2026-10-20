import test from 'node:test';
import assert from 'node:assert/strict';
import { canSubmitVideoFromConfig, collectVideoReferences, videoNodeViewState, videoReferenceAssetIDs, videoResultPreview } from './agent-canvas-video-jobs.js';

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

test('videoReferenceAssetIDs sends selected existing images once and rejects videos', () => {
  assert.deepEqual(videoReferenceAssetIDs({ referenceAssets: [
    { id: 'manual-image', kind: 'image' },
    { id: 'connected-image', kind: 'image' },
    { id: 'manual-image', kind: 'image' },
    { id: 'video-asset', kind: 'video' },
    { id: '', kind: 'image' }
  ] }), ['manual-image', 'connected-image']);
});

test('videoNodeViewState exposes confirmation instead of auto-applying result', () => {
  const state = videoNodeViewState({ state: 'awaiting_confirmation' });
  assert.equal(state.canConfirm, true);
  assert.equal(state.canGenerate, false);
  assert.equal(state.polling, false);
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

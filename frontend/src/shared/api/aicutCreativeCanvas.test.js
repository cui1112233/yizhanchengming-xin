import test from 'node:test';
import assert from 'node:assert/strict';
import { ensureAICutCreativeCanvas, getAICutCreativeBinding, transferAICutCreativeAssets, transferAICutCreativeScript } from './aicut.js';

test('canvas origin lookup is read only and rejects arbitrary paths', async () => {
  const binding = await getAICutCreativeBinding('canvas-1', async (path, options) => {
    assert.equal(path, '/api/agent/tasks/canvas-1/cut-project'); assert.equal(options, undefined);
    return { binding: { taskId: 'canvas-1', projectId: 'cut-1' } };
  });
  assert.equal(binding.projectId, 'cut-1');
  assert.equal(await getAICutCreativeBinding('canvas-1', async () => ({binding:null})), null);
  await assert.rejects(getAICutCreativeBinding('../other'), /无效/);
});

test('creative canvas binding uses the current project and POST without arbitrary task id', async () => {
  const result = await ensureAICutCreativeCanvas('cut-1', async (path, options) => {
    assert.equal(path, '/api/ai-cut/projects/cut-1/creative-canvas');
    assert.equal(options.method, 'POST');
    return { taskId: 'creative-1', canvas: { revision: 2, nodes: [], edges: [] } };
  });
  assert.equal(result.taskId, 'creative-1');
  await assert.rejects(ensureAICutCreativeCanvas('../other'), /项目无效/);
});

test('creative transfer submits unique real asset ids and returns the server result', async () => {
  const result = await transferAICutCreativeAssets('cut-1', ['asset-1', 'asset-1', 'asset-2'], async (path, options) => {
    assert.equal(path, '/api/ai-cut/projects/cut-1/creative-canvas/assets');
    assert.equal(options.method, 'POST');
    assert.deepEqual(JSON.parse(options.body), { assetIds: ['asset-1', 'asset-2'] });
    return { assets: [{ id: 'import-1' }] };
  });
  assert.equal(result.assets[0].id, 'import-1');
});

test('empty and oversized transfers fail before sending a request', async () => {
  const noRequest = () => assert.fail('invalid transfer reached the server');
  await assert.rejects(transferAICutCreativeAssets('cut-1', [], noRequest), /1.*30/);
  await assert.rejects(transferAICutCreativeAssets('cut-1', Array.from({ length: 31 }, (_, i) => `asset-${i}`), noRequest), /1.*30/);
});

test('script handoff carries selected saved nodes and the current revision', async () => {
  const result = await transferAICutCreativeScript('cut-1', ['story', 'story', 'shots'], 7, async (path, options) => {
    assert.equal(path, '/api/ai-cut/projects/cut-1/creative-canvas/script');
    assert.deepEqual(JSON.parse(options.body), { nodeIds: ['story', 'shots'], expectedRevision: 7 });
    return { timeline: { revision: 8, script: '故事' } };
  });
  assert.equal(result.timeline.revision, 8);
  await assert.rejects(transferAICutCreativeScript('cut-1', [], 7), /1.*10/);
  await assert.rejects(transferAICutCreativeScript('cut-1', ['story'], -1), /版本/);
});

import test from 'node:test';
import assert from 'node:assert/strict';

import * as host from './DirectorDeskHostAdapter.js';
import { createDirectorSaveQueue } from './director-save-queue.js';
import { insertDirectorCapture } from './director-capture.js';
import { createCanvasSaveQueue } from '../../pages/agent-canvas-state.js';

const { createDirectorDeskHostAdapter } = host;

test('导演台适配器从当前导演节点读取并保存 3D 项目', async () => {
  const saves = [];
  const node = { id: 'director-1', data: { type: 'director', config: { directorDesk: { sceneAsset: { id: 'scene-1', url: '/scene.png' } } } } };
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => node,
    saveProject: async (nodeID, project) => saves.push({ nodeID, project }),
    flushCanvas: async () => {}
  });

  assert.equal(adapter.readState().scene.assetId, 'scene-1');
  const saved = await adapter.writeState({ rendererVersion: '3d-v1', scene: { assetId: 'scene-2', url: '/next.png' } });
  assert.equal(saved.rendererVersion, '3d-v2');
  assert.equal(saves[0].project.rendererVersion, '3d-v2');
  assert.equal(saves[0].nodeID, 'director-1');
  assert.equal(saves[0].project.scene.assetId, 'scene-2');
});

test('导演台适配器等待外层画布持久化并透传失败', async () => {
  const events = [];
  const node = { id: 'director-1', data: { type: 'director', config: {} } };
  const adapter = createDirectorDeskHostAdapter({
    nodeID: node.id,
    getNode: () => node,
    saveProject: async () => { events.push('save'); },
    flushCanvas: async () => { events.push('flush'); }
  });

  await adapter.writeState({ scene: { prompt: '雨夜' } });
  assert.deepEqual(events, ['save', 'flush']);

  const failed = createDirectorDeskHostAdapter({
    nodeID: node.id,
    getNode: () => node,
    saveProject: async () => {},
    flushCanvas: async () => { throw new Error('画布落库失败'); }
  });
  await assert.rejects(() => failed.writeState({}), /画布落库失败/);
});

test('截图元数据进入同一保存队列，失败内容可由关闭刷新重试', async () => {
  assert.equal(typeof host.queueDirectorCaptureMetadata, 'function');
  let fail = true;
  const writes = [];
  const queue = createDirectorSaveQueue(async project => {
    writes.push(project.captures[0]?.assetId);
    if (fail) throw new Error('外层保存失败');
  });
  const updates = [];

  await assert.rejects(() => host.queueDirectorCaptureMetadata({
    saveQueue: queue,
    project: { scene: { prompt: '雨夜' } },
    asset: { id: 'capture-1', kind: 'image' },
    updateProject: project => updates.push(project)
  }), /外层保存失败/);
  assert.equal(queue.getStatus(), 'error');
  await assert.rejects(() => queue.flush(), /外层保存失败/);

  fail = false;
  await queue.flush();
  assert.deepEqual(writes, ['capture-1', 'capture-1', 'capture-1']);
  assert.equal(updates[0].captures[0].assetId, 'capture-1');
  assert.equal(queue.getStatus(), 'saved');
});

test('截图保存中发生的新编辑保持串行，并在截图版本之后持久化', async () => {
  const writes = [];
  let activeWrites = 0;
  let releaseCapture;
  let signalCapture;
  const captureStarted = new Promise(resolve => { signalCapture = resolve; });
  const queue = createDirectorSaveQueue(async project => {
    activeWrites += 1;
    assert.equal(activeWrites, 1);
    writes.push({ prompt: project.scene.prompt, capture: project.captures[0]?.assetId });
    if (project.scene.prompt === '截图版本') {
      signalCapture();
      await new Promise(resolve => { releaseCapture = resolve; });
    }
    activeWrites -= 1;
  });
  let capturedProject;
  const persisted = host.queueDirectorCaptureMetadata({
    saveQueue: queue,
    project: { scene: { prompt: '截图版本' } },
    asset: { id: 'capture-2', kind: 'image' },
    updateProject: project => { capturedProject = project; }
  });

  await captureStarted;
  queue.schedule({
    ...capturedProject,
    scene: { ...capturedProject.scene, prompt: '稍后编辑' }
  });
  releaseCapture();
  await persisted;

  assert.deepEqual(writes, [
    { prompt: '截图版本', capture: 'capture-2' },
    { prompt: '稍后编辑', capture: 'capture-2' }
  ]);
  assert.equal(queue.getStatus(), 'saved');
});

test('节点外层保存失败时阻止关闭，重试沿用同一素材并在立即刷新后存在', async () => {
  assert.equal(typeof host.createDirectorCaptureTransaction, 'function');
  let canvas = {
    revision: 1,
    nodes: [{ id: 'director-1', position: { x: 10, y: 20 }, data: { type: 'director', config: {} } }],
    edges: []
  };
  let persistedCanvas = structuredClone(canvas);
  let failCanvasSave = false;
  let blockNodeSave = true;
  let uploads = 0;
  let closed = false;
  const canvasQueue = createCanvasSaveQueue({
    revision: canvas.revision,
    delay: 60_000,
    save: async next => {
      if (failCanvasSave) throw new Error('节点外层保存失败');
      persistedCanvas = structuredClone({ ...next, revision: persistedCanvas.revision + 1 });
      return { canvas: persistedCanvas };
    },
    reload: async () => ({ canvas: persistedCanvas })
  });
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => canvas.nodes.find(node => node.id === 'director-1'),
    saveProject: async (nodeID, project) => {
      canvas = {
        ...canvas,
        nodes: canvas.nodes.map(node => node.id === nodeID
          ? { ...node, data: { ...node.data, config: { ...node.data.config, directorDesk: project } } }
          : node)
      };
      canvasQueue.enqueue(canvas);
    },
    insertCapture: async asset => {
      const inserted = insertDirectorCapture({
        nodes: canvas.nodes,
        edges: canvas.edges,
        directorNodeID: 'director-1',
        asset
      });
      canvas = { ...canvas, nodes: inserted.nodes, edges: inserted.edges };
      failCanvasSave = blockNodeSave;
      canvasQueue.enqueue(canvas);
    },
    flushCanvas: () => canvasQueue.flushRequired(),
    uploadCapture: async () => {
      uploads += 1;
      return { id: 'capture-stable', kind: 'image', url: '/capture-stable.png' };
    },
    onClose: () => { closed = true; }
  });
  const metadataQueue = createDirectorSaveQueue(project => adapter.writeState(project));
  const transaction = host.createDirectorCaptureTransaction({
    capture: async () => new Blob(['capture'], { type: 'image/png' }),
    storeCapture: file => adapter.storeCapture(file),
    persistMetadata: asset => host.queueDirectorCaptureMetadata({
      saveQueue: metadataQueue,
      project: adapter.readState(),
      asset
    }),
    insertCapture: asset => adapter.insertCapture(asset)
  });

  await assert.rejects(() => transaction.start(), /节点外层保存失败/);
  assert.equal(persistedCanvas.nodes[0].data.config.directorDesk.captures[0].assetId, 'capture-stable');
  assert.equal(persistedCanvas.nodes.some(node => node.id === 'image-director-capture-stable'), false);

  await assert.rejects(async () => {
    await transaction.flush();
    await metadataQueue.flush();
    adapter.close();
  }, /节点外层保存失败/);
  assert.equal(closed, false);

  blockNodeSave = false;
  failCanvasSave = false;
  await transaction.start();
  assert.equal(uploads, 1);
  const refreshed = structuredClone(persistedCanvas);
  assert.equal(refreshed.nodes.filter(node => node.id === 'image-director-capture-stable').length, 1);
  assert.equal(refreshed.edges.filter(edge => edge.source === 'director-1' && edge.target === 'image-director-capture-stable').length, 1);
});

test('导演台适配器只列出图片并拒绝非图片截图结果', async () => {
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    listImages: async () => ({ assets: [{ id: 'image-1', kind: 'image' }, { id: 'video-1', kind: 'video' }] }),
    uploadCapture: async () => ({ id: 'video-1', kind: 'video' })
  });

  assert.deepEqual(await adapter.listProjectImages(), [{ id: 'image-1', kind: 'image' }]);
  await assert.rejects(() => adapter.storeCapture(new Blob(['capture'], { type: 'image/png' })), /图片/);
});

test('截图和项目图片在上传前拒绝非图片 Blob 或 File', async () => {
  let captureUploads = 0;
  let imageUploads = 0;
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    uploadCapture: async () => { captureUploads += 1; return { id: 'capture-1', kind: 'image' }; },
    uploadImage: async () => { imageUploads += 1; return { id: 'image-1', kind: 'image' }; }
  });

  await assert.rejects(() => adapter.storeCapture(new Blob(['video'], { type: 'video/mp4' })), /图片/);
  await assert.rejects(() => adapter.uploadProjectImage(new File(['text'], 'reference.txt', { type: 'text/plain' })), /图片/);
  assert.equal(captureUploads, 0);
  assert.equal(imageUploads, 0);
});

test('导演台适配器上传项目图片并拒绝非图片结果', async () => {
  const files = [];
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    uploadImage: async file => {
      files.push(file);
      return { asset: { id: 'image-1', kind: 'image', name: '参考图' } };
    }
  });
  const file = new Blob(['image'], { type: 'image/png' });

  assert.deepEqual(await adapter.uploadProjectImage(file), { id: 'image-1', kind: 'image', name: '参考图' });
  assert.deepEqual(files, [file]);

  const invalid = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    uploadImage: async () => ({ asset: { id: 'video-1', kind: 'video' } })
  });
  await assert.rejects(() => invalid.uploadProjectImage(file), /图片/);
});

test('导演台适配器将关闭通知交回画布', () => {
  let closed = '';
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => ({ id: 'director-1', data: { type: 'director', config: {} } }),
    onClose: nodeID => { closed = nodeID; }
  });

  adapter.close();
  assert.equal(closed, 'director-1');
});

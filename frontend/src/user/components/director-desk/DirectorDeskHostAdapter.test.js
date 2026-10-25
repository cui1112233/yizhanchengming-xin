import test from 'node:test';
import assert from 'node:assert/strict';

import { createDirectorDeskHostAdapter } from './DirectorDeskHostAdapter.js';

test('导演台适配器从当前导演节点读取并保存 3D 项目', async () => {
  const saves = [];
  const node = { id: 'director-1', data: { type: 'director', config: { directorDesk: { sceneAsset: { id: 'scene-1', url: '/scene.png' } } } } };
  const adapter = createDirectorDeskHostAdapter({
    nodeID: 'director-1',
    getNode: () => node,
    saveProject: async (nodeID, project) => saves.push({ nodeID, project })
  });

  assert.equal(adapter.readState().scene.assetId, 'scene-1');
  const saved = await adapter.writeState({ rendererVersion: '3d-v1', scene: { assetId: 'scene-2', url: '/next.png' } });
  assert.equal(saved.rendererVersion, '3d-v2');
  assert.equal(saves[0].project.rendererVersion, '3d-v2');
  assert.equal(saves[0].nodeID, 'director-1');
  assert.equal(saves[0].project.scene.assetId, 'scene-2');
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

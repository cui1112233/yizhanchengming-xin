import test from 'node:test';
import assert from 'node:assert/strict';

import { actorPlaneProps, cameraFromControls, updateCamera } from './director-scene-state.js';
import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';

test('图片演员会投影成稳定的 3D 立面参数', () => {
  const actor = actorPlaneProps({ id: 'actor-1', assetId: 'image-1', position: [2, 0, -1], rotation: [0, 1.2, 0], scale: 1.2 });
  assert.deepEqual(actor.position, [2, 0, -1]);
  assert.deepEqual(actor.scale, [1.2, 1.2, 1.2]);
  assert.deepEqual(actor.rotation, [0, 1.2, 0]);
});

test('v2 三轴缩放在视口中保留各轴数值', () => {
  assert.deepEqual(actorPlaneProps({ scale: [1.4, 2, 0.5] }).scale, [1.4, 2, 0.5]);
});

test('镜头取景更新会保留为可持久化的相机数据', () => {
  const next = updateCamera(normalizeDirectorProject(), { position: [0, 3, 8], target: [0, 0, 0], fov: 38 });
  assert.equal(next.camera.fov, 38);
  assert.equal(next.directorCamera.fov, 38);
  assert.deepEqual(next.camera.position, [0, 3, 8]);
});

test('轨道控制器相机能转换成普通数组数据', () => {
  const result = cameraFromControls(
    { position: { toArray: () => [1, 2, 3] }, fov: 45 },
    { toArray: () => [0, 1, 0] }
  );
  assert.deepEqual(result, { position: [1, 2, 3], target: [0, 1, 0], fov: 45 });
});

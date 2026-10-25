import test from 'node:test';
import assert from 'node:assert/strict';

import { actorPlaneProps, aspectRatioValue, cameraForView, cameraFromControls, captureViewportSize, updateCamera } from './director-scene-state.js';
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

test('画幅映射为稳定比例和截图尺寸', () => {
  assert.equal(aspectRatioValue('21:9'), 21 / 9);
  assert.equal(aspectRatioValue('adaptive'), null);
  assert.equal(aspectRatioValue('toString'), null);
  assert.deepEqual(captureViewportSize('9:16', 1280), { width: 720, height: 1280 });
  assert.deepEqual(captureViewportSize('16:9', 1280), { width: 1280, height: 720 });
  assert.deepEqual(captureViewportSize('adaptive', 1280), { width: 1280, height: 720 });
});

test('机位和导演视角分别使用活动机位和自由摄像机，失效机位回退主机位', () => {
  const project = normalizeDirectorProject({ cameras: [{ id: 'camera-other', position: [1, 2, 3] }], activeCameraId: 'camera-other' });
  assert.equal(cameraForView(project, 'camera').id, 'camera-other');
  assert.equal(cameraForView(project, 'director').kind, 'director');
  assert.equal(cameraForView({ ...project, activeCameraId: 'missing' }, 'camera').id, 'camera-main');
});

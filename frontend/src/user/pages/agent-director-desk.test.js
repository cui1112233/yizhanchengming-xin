import test from 'node:test';
import assert from 'node:assert/strict';

import {
  addDirectorActor, addDirectorCamera, updateDirectorActor, updateDirectorCamera,
  updateDirectorViewCamera, setDirectorActiveCamera, setDirectorAspectRatio,
  removeDirectorObject, directorSceneObjects,
  directorDeskConfig,
  directorDeskSceneStyle,
  directorProjectCaptureMetadata,
  migrateLegacyDirectorDesk,
  moveDirectorDeskActor,
  normalizeDirectorProject,
  updateDirectorActorTransform
} from './agent-director-desk.js';

test('导演台配置保存预演场景、演员位置和镜头参数', () => {
  const config = directorDeskConfig({
    sceneAsset: { id: 'scene-1', url: '/scene.png', name: '庭院' },
    actors: [{ id: 'actor-1', assetId: 'character-1', x: 20, y: 70, scale: 0.8 }],
    camera: { zoom: 1.2, x: 4, y: -3 }
  });
  assert.equal(config.rendererVersion, '2.5d-v1');
  assert.equal(config.sceneAsset.id, 'scene-1');
  assert.equal(config.actors[0].x, 20);
  assert.equal(config.camera.zoom, 1.2);
});

test('导演台配置会保存节点内填写的场景描述', () => {
  const config = directorDeskConfig({
    scenePrompt: '  雨夜的城市天桥，霓虹倒影  ',
    sceneAsset: { id: 'scene-1', url: '/scene.png', name: '城市天桥' }
  });

  assert.equal(config.scenePrompt, '雨夜的城市天桥，霓虹倒影');
  assert.equal(config.sceneAsset.name, '城市天桥');
});

test('导演台演员移动受预演画面边界限制', () => {
  const config = directorDeskConfig({ actors: [{ id: 'actor-1', assetId: 'character-1', x: 20, y: 70, scale: 1 }] });
  const moved = moveDirectorDeskActor(config, 'actor-1', { x: 140, y: -20 });
  assert.deepEqual(moved.actors[0], { id: 'actor-1', assetId: 'character-1', x: 100, y: 0, scale: 1, rotation: 0 });
});

test('导演台取景会映射为稳定的场景样式', () => {
  const style = directorDeskSceneStyle({ camera: { zoom: 1.5, x: 10, y: -5 } });
  assert.equal(style.transform, 'scale(1.5) translate(10%, -5%)');
});

test('旧 2.5D 导演台数据会迁移为带版本的 3D 项目', () => {
  const project = migrateLegacyDirectorDesk({
    sceneAsset: { id: 'scene-1', url: '/scene.png', name: '天桥' },
    actors: [{ id: 'actor-1', assetId: 'actor-image', x: 20, y: 70, scale: 1 }],
    camera: { zoom: 1.2, x: 4, y: -3 }
  });

  assert.equal(project.rendererVersion, '3d-v2');
  assert.equal(project.scene.assetId, 'scene-1');
  assert.equal(project.actors[0].assetId, 'actor-image');
  assert.deepEqual(project.camera.target, [0, 0, 0]);
});

test('v1 camera and scalar scale migrate without losing scene or captures', () => {
  const result = normalizeDirectorProject({ rendererVersion: '3d-v1', scene: { assetId: 's', prompt: 'rain' },
    actors: [{ id: 'a', assetId: 'img', position: [1, 0, 2], scale: 1.4 }],
    camera: { position: [0, 3, 8], target: [0, 1, 0], fov: 42 }, captures: [{ assetId: 'cap' }] });
  assert.equal(result.rendererVersion, '3d-v2');
  assert.deepEqual(result.actors[0].scale, [1.4, 1.4, 1.4]);
  assert.equal(result.cameras[0].id, 'camera-main');
  assert.equal(result.activeCameraId, 'camera-main');
  assert.deepEqual(result.directorCamera.target, [0, 1, 0]);
  assert.equal(result.scene.prompt, 'rain');
  assert.equal(result.captures[0].assetId, 'cap');
  assert.equal(directorDeskConfig(result).actors[0].scale, 1.4);
});

test('旧画布镜头参数迁移后仍能用于紧凑画布', () => {
  const project = normalizeDirectorProject({ scene: { prompt: '雨天' }, camera: { zoom: 1.2, x: 4, y: -3 } });
  const compact = directorDeskConfig(normalizeDirectorProject(project));
  assert.equal(compact.scenePrompt, '雨天');
  assert.deepEqual(compact.camera, { zoom: 1.2, x: 4, y: -3 });
});

test('object helpers preserve identity, protect main camera and switch active camera on removal', () => {
  const base = normalizeDirectorProject();
  const actor1 = addDirectorActor(base, { id: 'img', assetId: 'img' });
  const actor2 = addDirectorActor(actor1, { id: 'img', assetId: 'img' });
  assert.notEqual(actor2.actors[0].id, actor2.actors[1].id);
  assert.notEqual(addDirectorActor(base, { id: 'camera-main', assetId: 'img' }).actors[0].id, 'camera-main');
  assert.notEqual(addDirectorCamera(actor1, { id: 'img' }).cameras[1].id, 'img');
  const changed = updateDirectorActor(actor2, actor2.actors[0].id, { id: 'bad', assetId: 'bad', scale: [2, 3, 4] });
  assert.equal(changed.actors[0].assetId, 'img');
  assert.deepEqual(changed.actors[0].scale, [2, 3, 4]);
  assert.equal(actor2.actors[0].scale[0], 1);
  const camera = addDirectorCamera(changed, { id: 'camera-b', name: '侧面机位' });
  const updated = updateDirectorCamera(camera, 'camera-b', { id: 'bad', fov: 200 });
  assert.equal(updated.cameras[1].id, 'camera-b');
  assert.equal(updated.cameras[1].fov, 90);
  const active = setDirectorActiveCamera(updated, 'camera-b');
  const framed = setDirectorAspectRatio(active, '9:16');
  assert.equal(framed.activeCameraId, 'camera-b');
  assert.equal(framed.aspectRatio, '9:16');
  assert.deepEqual(directorSceneObjects(framed).map(item => item.kind), ['scene', 'actor', 'actor', 'camera', 'camera']);
  assert.equal(removeDirectorObject(framed, 'camera-main').cameras.length, 2);
  assert.equal(removeDirectorObject(framed, 'scene').scene.assetId, '');
  assert.equal(removeDirectorObject(framed, 'camera-b').activeCameraId, 'camera-main');
  assert.equal(removeDirectorObject(framed, framed.actors[0].id).actors.length, 1);
  assert.equal(setDirectorActiveCamera(framed, 'missing').activeCameraId, 'camera-b');
  assert.equal(setDirectorAspectRatio(framed, 'bad').aspectRatio, 'adaptive');
  assert.deepEqual(updateDirectorViewCamera(framed, { position: [4, 5, 6] }).directorCamera.position, [4, 5, 6]);
});

test('演员变换和截图记录会保留稳定素材 ID', () => {
  const project = normalizeDirectorProject({
    rendererVersion: '3d-v1',
    actors: [{ id: 'actor-1', assetId: 'actor-image' }]
  });
  const moved = updateDirectorActorTransform(project, 'actor-1', {
    position: [2, 0, -1], rotation: [0, 1.2, 0], scale: 1.3
  });
  const next = directorProjectCaptureMetadata(moved, {
    id: 'capture-1', url: '/capture.png', name: '构图 1', createdAt: '2026-10-24T00:00:00Z'
  });

  assert.deepEqual(moved.actors[0].position, [2, 0, -1]);
  assert.equal(next.captures[0].assetId, 'capture-1');
});

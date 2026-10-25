import test from 'node:test';
import assert from 'node:assert/strict';

import {
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

  assert.equal(project.rendererVersion, '3d-v1');
  assert.equal(project.scene.assetId, 'scene-1');
  assert.equal(project.actors[0].assetId, 'actor-image');
  assert.deepEqual(project.camera.target, [0, 0, 0]);
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

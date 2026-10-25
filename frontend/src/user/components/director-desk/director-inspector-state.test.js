import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeDirectorProject } from '../../pages/agent-director-desk.js';
import { applyDirectorInspectorChange } from './director-inspector-state.js';

test('actor axis edits preserve other axes and uniform scale changes all axes', () => {
  const base = normalizeDirectorProject({ actors: [{ id: 'a', assetId: 'image', position: [1, 2, 3], rotation: [4, 5, 6], scale: [1, 2, 3] }] });
  const moved = applyDirectorInspectorChange(base, { kind: 'actor', id: 'a' }, 'position.0', '2.5');
  const rotated = applyDirectorInspectorChange(moved, { kind: 'actor', id: 'a' }, 'rotation.1', '-3');
  const scaled = applyDirectorInspectorChange(rotated, { kind: 'actor', id: 'a' }, 'uniformScale', '1.8');
  assert.deepEqual(scaled.actors[0].position, [2.5, 2, 3]);
  assert.deepEqual(scaled.actors[0].rotation, [4, -3, 6]);
  assert.deepEqual(scaled.actors[0].scale, [1.8, 1.8, 1.8]);
  assert.deepEqual(base.actors[0].position, [1, 2, 3]);
});

test('camera FOV clamps and axis edits preserve the remaining values', () => {
  const base = normalizeDirectorProject();
  const wide = applyDirectorInspectorChange(base, { kind: 'camera', id: 'camera-main' }, 'fov', '120');
  const moved = applyDirectorInspectorChange(wide, { kind: 'camera', id: 'camera-main' }, 'target.1', '4');
  const rolled = applyDirectorInspectorChange(moved, { kind: 'camera', id: 'camera-main' }, 'rotation.2', '-0.35');
  assert.equal(rolled.cameras[0].fov, 90);
  assert.deepEqual(rolled.cameras[0].target, [0, 4, 0]);
  assert.deepEqual(rolled.cameras[0].rotation, [0, 0, -0.35]);
  assert.deepEqual(applyDirectorInspectorChange(rolled, { kind: 'camera', id: 'camera-main' }, 'rotation.0', '1'), rolled);
  assert.deepEqual(applyDirectorInspectorChange(rolled, { kind: 'camera', id: 'camera-main' }, 'rotation.1', '1'), rolled);
});

test('invalid values, unsupported paths and missing selections cannot change the project', () => {
  const base = normalizeDirectorProject({ actors: [{ id: 'a', assetId: 'image' }] });
  for (const [selection, field, value] of [
    [{ kind: 'camera', id: 'camera-main' }, 'position.1', 'abc'],
    [{ kind: 'camera', id: 'camera-main' }, 'position.1', ''],
    [{ kind: 'actor', id: 'a' }, 'position.3', '2'],
    [{ kind: 'actor', id: 'a' }, 'assetId', '2'],
    [{ kind: 'actor', id: 'missing' }, 'scale.0', '2'],
    [{ kind: 'scene', id: 'scene' }, 'fov', '2']
  ]) assert.deepEqual(applyDirectorInspectorChange(base, selection, field, value), base);
});

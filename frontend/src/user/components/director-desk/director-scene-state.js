import { normalizeDirectorProject, updateDirectorViewCamera } from '../../pages/agent-director-desk.js';

const RATIOS = Object.freeze({ '21:9': 21 / 9, '16:9': 16 / 9, '4:3': 4 / 3, '1:1': 1, '3:4': 3 / 4, '9:16': 9 / 16 });

export function aspectRatioValue(value) {
  return Object.hasOwn(RATIOS, value) ? RATIOS[value] : null;
}

export function captureViewportSize(value, longEdge = 1280) {
  const ratio = aspectRatioValue(value) || 16 / 9;
  return ratio >= 1
    ? { width: longEdge, height: Math.round(longEdge / ratio) }
    : { width: Math.round(longEdge * ratio), height: longEdge };
}

export function cameraForView(project, view) {
  const current = normalizeDirectorProject(project);
  if (view === 'director') return { ...current.directorCamera, kind: 'director' };
  return current.cameras.find(camera => camera.id === current.activeCameraId)
    || current.cameras.find(camera => camera.id === 'camera-main');
}

function vector(value, fallback) {
  if (Array.isArray(value)) return value.slice(0, fallback.length);
  if (value && typeof value.toArray === 'function') return value.toArray();
  return fallback.slice();
}

export function actorPlaneProps(actor = {}) {
  const scale = Number(actor.scale);
  const size = Number.isFinite(scale) && scale > 0 ? scale : 1;
  return {
    position: vector(actor.position, [0, 0, 0]),
    rotation: vector(actor.rotation, [0, 0, 0]),
    scale: Array.isArray(actor.scale) ? vector(actor.scale, [1, 1, 1]) : [size, size, size]
  };
}

export function updateCamera(project, camera) {
  return updateDirectorViewCamera(project, camera || {});
}

export function focusDirectorObject(project, selection) {
  const current = normalizeDirectorProject(project);
  const object = selection?.kind === 'actor'
    ? current.actors.find(actor => actor.id === selection.id)
    : selection?.kind === 'camera' ? current.cameras.find(camera => camera.id === selection.id) : null;
  if (selection?.kind !== 'scene' && !object) return current;
  const target = object ? object.position : [0, 0, 0];
  const offset = current.directorCamera.position.map((value, index) => value - current.directorCamera.target[index]);
  return updateDirectorViewCamera(current, { target, position: target.map((value, index) => value + offset[index]) });
}

export function cameraFromControls(camera, target) {
  return {
    position: vector(camera?.position, [0, 3, 8]),
    target: vector(target, [0, 0, 0]),
    fov: Number(camera?.fov) || 42
  };
}

import { updateDirectorViewCamera } from '../../pages/agent-director-desk.js';

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

export function cameraFromControls(camera, target) {
  return {
    position: vector(camera?.position, [0, 3, 8]),
    target: vector(target, [0, 0, 0]),
    fov: Number(camera?.fov) || 42
  };
}

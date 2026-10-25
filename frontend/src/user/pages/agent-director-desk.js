const DEFAULT_CAMERA = Object.freeze({ zoom: 1, x: 0, y: 0 });
const DEFAULT_3D_CAMERA = Object.freeze({ position: [0, 3, 8], target: [0, 0, 0], fov: 42 });
export const DIRECTOR_ASPECT_RATIOS = Object.freeze(['adaptive', '21:9', '16:9', '4:3', '1:1', '3:4', '9:16']);

function clamp(value, min, max, fallback) {
  const number = Number(value);
  if (!Number.isFinite(number)) return fallback;
  return Math.min(max, Math.max(min, number));
}

function text(value, limit = 1000) {
  return String(value || '').trim().slice(0, limit);
}

function numberList(value, defaults) {
  const source = Array.isArray(value) ? value : [];
  return defaults.map((fallback, index) => clamp(source[index], -10000, 10000, fallback));
}

function normalizeAsset(asset) {
  if (!asset || typeof asset !== 'object') return null;
  const assetId = text(asset.assetId || asset.id, 200);
  if (!assetId) return null;
  return { assetId, url: text(asset.url, 4000), name: text(asset.name, 500) };
}

function normalizeLegacyActor(actor = {}) {
  return {
    id: text(actor.id || actor.assetId, 200),
    assetId: text(actor.assetId || actor.id, 200),
    x: clamp(actor.x, 0, 100, 50),
    y: clamp(actor.y, 0, 100, 65),
    scale: clamp(actor.scale, 0.2, 2, 1),
    rotation: clamp(actor.rotation, -45, 45, 0)
  };
}

function normalizeActor(actor = {}) {
  if (!actor || typeof actor !== 'object') return null;
  const id = text(actor.id || actor.assetId, 200);
  const assetId = text(actor.assetId || actor.id, 200);
  if (!id || !assetId) return null;
  const legacy = !Array.isArray(actor.position);
  const position = legacy
    ? [(clamp(actor.x, 0, 100, 50) - 50) / 12, 0, (clamp(actor.y, 0, 100, 65) - 65) / 12]
    : numberList(actor.position, [0, 0, 0]);
  const rotation = Array.isArray(actor.rotation)
    ? numberList(actor.rotation, [0, 0, 0])
    : [0, 0, clamp(actor.rotation, -45, 45, 0) * Math.PI / 180];
  const scale = Array.isArray(actor.scale) ? actor.scale : [actor.scale, actor.scale, actor.scale];
  return { id, assetId, position, rotation, scale: [0, 1, 2].map(index => clamp(scale[index], 0.1, 10, 1)) };
}

function normalizeCamera(camera = {}) {
  const input = camera && typeof camera === 'object' ? camera : {};
  return {
    position: numberList(input.position, DEFAULT_3D_CAMERA.position),
    target: numberList(input.target, DEFAULT_3D_CAMERA.target),
    fov: clamp(input.fov, 20, 90, DEFAULT_3D_CAMERA.fov)
  };
}

function normalizeCameraObject(camera = {}, index = 0) {
  const input = camera && typeof camera === 'object' ? camera : {};
  return {
    id: text(input.id || (index === 0 ? 'camera-main' : `camera-${index + 1}`), 200),
    name: text(input.name || (index === 0 ? '主机位' : `机位${index + 1}`), 200),
    ...normalizeCamera(input),
    rotation: numberList(input.rotation, [0, 0, 0])
  };
}

function normalizeCapture(capture) {
  const asset = normalizeAsset(capture);
  if (!asset) return null;
  return { ...asset, createdAt: text(capture.createdAt, 100) };
}

export function normalizeDirectorProject(input = {}) {
  const source = input && typeof input === 'object' ? input : {};
  const scene = normalizeAsset(source.scene || source.sceneAsset);
  const cameras = Array.isArray(source.cameras) ? source.cameras.map(normalizeCameraObject) : [];
  if (!cameras.some(camera => camera.id === 'camera-main')) cameras.unshift(normalizeCameraObject(source.camera));
  const directorCamera = normalizeCamera(source.directorCamera || source.camera);
  return {
    rendererVersion: '3d-v2',
    scene: scene
      ? { ...scene, prompt: text(source.scenePrompt || source.scene?.prompt, 1000) }
      : { assetId: '', url: '', name: '', prompt: text(source.scenePrompt || source.scene?.prompt, 1000) },
    actors: Array.isArray(source.actors) ? source.actors.map(normalizeActor).filter(Boolean) : [],
    cameras,
    activeCameraId: cameras.some(camera => camera.id === source.activeCameraId) ? source.activeCameraId : 'camera-main',
    aspectRatio: DIRECTOR_ASPECT_RATIOS.includes(source.aspectRatio) ? source.aspectRatio : 'adaptive',
    directorCamera,
    camera: { ...directorCamera, ...(source.camera && typeof source.camera === 'object' && 'zoom' in source.camera ? {
      zoom: clamp(source.camera.zoom, 0.6, 2, 1), x: clamp(source.camera.x, -30, 30, 0), y: clamp(source.camera.y, -30, 30, 0)
    } : {}) },
    captures: Array.isArray(source.captures) ? source.captures.map(normalizeCapture).filter(Boolean) : []
  };
}

function availableId(base, objects) {
  const ids = new Set(objects.map(item => item.id));
  let id = text(base, 190) || 'object';
  const prefix = id;
  for (let index = 2; ids.has(id); index += 1) id = `${prefix}-${index}`;
  return id;
}

export function addDirectorActor(project, actor = {}) {
  const current = normalizeDirectorProject(project);
  const assetId = text(actor.assetId || actor.id, 200);
  if (!assetId) return current;
  return normalizeDirectorProject({ ...current, actors: [...current.actors, { ...actor, assetId, id: availableId(actor.id || assetId, [{ id: 'scene' }, ...current.actors, ...current.cameras]) }] });
}

export function updateDirectorActor(project, actorID, patch = {}) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({ ...current, actors: current.actors.map(actor => actor.id === actorID ? { ...actor, ...patch, id: actor.id, assetId: actor.assetId } : actor) });
}

export function addDirectorCamera(project, camera = {}) {
  const current = normalizeDirectorProject(project);
  const id = availableId(camera.id || `camera-${current.cameras.length + 1}`, [{ id: 'scene' }, ...current.actors, ...current.cameras]);
  return normalizeDirectorProject({ ...current, cameras: [...current.cameras, normalizeCameraObject({ ...camera, id }, current.cameras.length)] });
}

export function updateDirectorCamera(project, cameraID, patch = {}) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({ ...current, cameras: current.cameras.map(camera => camera.id === cameraID ? { ...camera, ...patch, id: camera.id } : camera) });
}

export function updateDirectorViewCamera(project, patch = {}) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({ ...current, directorCamera: { ...current.directorCamera, ...patch } });
}

export function setDirectorActiveCamera(project, cameraID) {
  const current = normalizeDirectorProject(project);
  return current.cameras.some(camera => camera.id === cameraID) ? { ...current, activeCameraId: cameraID } : current;
}

export function setDirectorAspectRatio(project, ratio) {
  const current = normalizeDirectorProject(project);
  return { ...current, aspectRatio: DIRECTOR_ASPECT_RATIOS.includes(ratio) ? ratio : 'adaptive' };
}

export function removeDirectorObject(project, objectID) {
  const current = normalizeDirectorProject(project);
  if (objectID === 'scene' || objectID === 'camera-main') return current;
  const cameras = current.cameras.length > 1 ? current.cameras.filter(camera => camera.id !== objectID) : current.cameras;
  return normalizeDirectorProject({ ...current, actors: current.actors.filter(actor => actor.id !== objectID), cameras,
    activeCameraId: cameras.some(camera => camera.id === current.activeCameraId) ? current.activeCameraId : cameras[0].id });
}

export function directorSceneObjects(project) {
  const current = normalizeDirectorProject(project);
  return [{ ...current.scene, id: 'scene', kind: 'scene' }, ...current.actors.map(actor => ({ ...actor, kind: 'actor' })),
    ...current.cameras.map(camera => ({ ...camera, kind: 'camera' }))];
}

export function migrateLegacyDirectorDesk(legacy = {}) {
  const source = legacy && typeof legacy === 'object' ? legacy : {};
  return normalizeDirectorProject({
    sceneAsset: source.sceneAsset,
    scenePrompt: source.scenePrompt,
    actors: source.actors,
    camera: source.camera,
    captures: source.captures
  });
}

export function updateDirectorActorTransform(project, actorID, transform = {}) {
  const normalized = normalizeDirectorProject(project);
  const id = text(actorID, 200);
  return {
    ...normalized,
    actors: normalized.actors.map(actor => actor.id === id
      ? normalizeActor({ ...actor, ...transform, id: actor.id, assetId: actor.assetId })
      : actor)
  };
}

export function directorProjectCaptureMetadata(project, asset = {}) {
  const normalized = normalizeDirectorProject(project);
  const capture = normalizeCapture(asset);
  if (!capture) return normalized;
  return { ...normalized, captures: [capture, ...normalized.captures.filter(item => item.assetId !== capture.assetId)] };
}

// This compatibility view lets the existing node remain usable while the full 3D desk lands.
export function directorDeskConfig(input = {}) {
  const project = input.rendererVersion === '3d-v1' || input.rendererVersion === '3d-v2' ? normalizeDirectorProject(input) : migrateLegacyDirectorDesk(input);
  const sourceCamera = input.camera && typeof input.camera === 'object' ? input.camera : {};
  return {
    rendererVersion: '2.5d-v1',
    scenePrompt: project.scene.prompt,
    sceneAsset: project.scene.assetId ? { id: project.scene.assetId, url: project.scene.url, name: project.scene.name } : null,
    actors: project.actors.map(actor => normalizeLegacyActor({
      id: actor.id,
      assetId: actor.assetId,
      x: actor.position[0] * 12 + 50,
      y: actor.position[2] * 12 + 65,
      scale: actor.scale[0],
      rotation: actor.rotation[2] * 180 / Math.PI
    })),
    camera: {
      zoom: clamp(sourceCamera.zoom, 0.6, 2, DEFAULT_CAMERA.zoom),
      x: clamp(sourceCamera.x, -30, 30, DEFAULT_CAMERA.x),
      y: clamp(sourceCamera.y, -30, 30, DEFAULT_CAMERA.y)
    }
  };
}

export function moveDirectorDeskActor(config, actorID, position) {
  const normalized = directorDeskConfig(config);
  return {
    ...normalized,
    actors: normalized.actors.map(actor => actor.id === actorID ? normalizeLegacyActor({ ...actor, ...position }) : actor)
  };
}

export function directorDeskSceneStyle(config) {
  const camera = directorDeskConfig(config).camera;
  return { transform: `scale(${camera.zoom}) translate(${camera.x}%, ${camera.y}%)` };
}

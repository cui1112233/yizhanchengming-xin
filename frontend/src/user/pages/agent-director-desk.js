const DEFAULT_CAMERA = Object.freeze({ zoom: 1, x: 0, y: 0 });
const DEFAULT_3D_CAMERA = Object.freeze({ position: [0, 3, 8], target: [0, 0, 0], fov: 42 });

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
  return { id, assetId, position, rotation, scale: clamp(actor.scale, 0.1, 10, 1) };
}

function normalizeCamera(camera = {}) {
  const input = camera && typeof camera === 'object' ? camera : {};
  return {
    position: numberList(input.position, DEFAULT_3D_CAMERA.position),
    target: numberList(input.target, DEFAULT_3D_CAMERA.target),
    fov: clamp(input.fov, 20, 90, DEFAULT_3D_CAMERA.fov)
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
  return {
    rendererVersion: '3d-v1',
    scene: scene
      ? { ...scene, prompt: text(source.scenePrompt || source.scene?.prompt, 1000) }
      : { assetId: '', url: '', name: '', prompt: text(source.scenePrompt, 1000) },
    actors: Array.isArray(source.actors) ? source.actors.map(normalizeActor).filter(Boolean) : [],
    camera: normalizeCamera(source.camera),
    captures: Array.isArray(source.captures) ? source.captures.map(normalizeCapture).filter(Boolean) : []
  };
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
  const project = input.rendererVersion === '3d-v1' ? normalizeDirectorProject(input) : migrateLegacyDirectorDesk(input);
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
      scale: actor.scale,
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

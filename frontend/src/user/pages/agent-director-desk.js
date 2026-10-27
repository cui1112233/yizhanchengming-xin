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

function normalizeDirectorObject(value = {}, kind = 'actor') {
  const normalized = normalizeActor(value);
  if (!normalized) return null;
  return { ...normalized, kind: kind === 'prop' ? 'prop' : 'actor' };
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
  if (!cameras.some(camera => camera.id === 'camera-main')) cameras.unshift({ ...normalizeCameraObject(source.camera), id: 'camera-main' });
  const actors = [];
  const props = [];
  const sourceObjects = Array.isArray(source.objects) ? source.objects : [];
  const legacyObjects = [
    ...(Array.isArray(source.actors) ? source.actors.map(value => ({ ...value, kind: 'actor' })) : []),
    ...(Array.isArray(source.props) ? source.props.map(value => ({ ...value, kind: 'prop' })) : [])
  ];
  const seenObjectIDs = new Set();
  const allObjects = [...sourceObjects, ...legacyObjects];
  for (const [objectIndex, value] of allObjects.entries()) {
    const object = normalizeDirectorObject(value, value.kind);
    const actor = object;
    if (!actor) continue;
    if (seenObjectIDs.has(actor.id) && sourceObjects.length > 0 && objectIndex >= sourceObjects.length) {
      const existing = [...actors, ...props].find(item => item.id === actor.id);
      if (existing) {
        const index = actors.indexOf(existing);
        if (index >= 0 && actor.kind === 'actor') actors[index] = actor;
        const propIndex = props.indexOf(existing);
        if (propIndex >= 0 && actor.kind === 'prop') props[propIndex] = actor;
      }
      continue;
    }
    const id = availableId(actor.id, [{ id: 'scene' }, ...cameras, ...actors, ...props]);
    seenObjectIDs.add(id);
    const next = { ...actor, id };
    if (actor.kind === 'prop') props.push(next);
    else actors.push(next);
  }
  const directorCamera = normalizeCamera(source.directorCamera || source.camera);
  const scenePrompt = text(source.scenePrompt || source.scene?.prompt, 2000);
  const shotPrompt = text(source.shotPrompt || source.videoPrompt || source.scene?.shotPrompt, 2000);
  return {
    rendererVersion: '3d-v2',
    scene: scene
      ? { ...scene, prompt: scenePrompt, scenePrompt, shotPrompt }
      : { assetId: '', url: '', name: '', prompt: scenePrompt, scenePrompt, shotPrompt },
    actors,
    props,
    objects: [...actors, ...props],
    cameras,
    activeCameraId: cameras.some(camera => camera.id === source.activeCameraId) ? source.activeCameraId : 'camera-main',
    aspectRatio: DIRECTOR_ASPECT_RATIOS.includes(source.aspectRatio) ? source.aspectRatio : 'adaptive',
    directorCamera,
    shotSize: text(source.shotSize, 40) || 'medium',
    focalLength: clamp(source.focalLength, 14, 200, 50),
    azimuth: clamp(source.azimuth, -180, 180, 0),
    elevation: clamp(source.elevation, -89, 89, 0),
    scenePrompt,
    shotPrompt,
    camera: { ...directorCamera, ...(source.camera && typeof source.camera === 'object' && 'zoom' in source.camera ? {
      zoom: clamp(source.camera.zoom, 0.6, 2, 1), x: clamp(source.camera.x, -30, 30, 0), y: clamp(source.camera.y, -30, 30, 0)
    } : {}) },
    captures: Array.isArray(source.captures) ? source.captures.map(normalizeCapture).filter(Boolean) : []
  };
}

export function directorObjectTransform(project, selection, patch = {}) {
  const current = normalizeDirectorProject(project);
  const id = text(selection?.id || selection, 200);
  if (!id) return current;
  const nextObjects = current.objects.map(object => object.id === id
    ? normalizeDirectorObject({ ...object, ...patch, id: object.id, assetId: object.assetId }, object.kind)
    : object).filter(Boolean);
  const actors = nextObjects.filter(object => object.kind === 'actor');
  const props = nextObjects.filter(object => object.kind === 'prop');
  return normalizeDirectorProject({ ...current, actors, props, objects: nextObjects });
}

const SHOT_SIZE_DISTANCE = Object.freeze({ extremeClose: 3.2, close: 4.2, mediumClose: 5.4, medium: 7, full: 9, wide: 12 });
const SHOT_SIZE_FOV = Object.freeze({ extremeClose: 28, close: 34, mediumClose: 38, medium: 42, full: 48, wide: 58 });

export function directorCameraSemantics(project, patch = {}) {
  const current = normalizeDirectorProject(project);
  const shotSize = text(patch.shotSize || current.shotSize, 40) || 'medium';
  const focalLength = clamp(patch.focalLength ?? current.focalLength, 14, 200, 50);
  const azimuth = clamp(patch.azimuth ?? current.azimuth, -180, 180, 0);
  const elevation = clamp(patch.elevation ?? current.elevation, -89, 89, 0);
  const target = numberList(patch.target || current.directorCamera.target, [0, 0, 0]);
  const distance = clamp(patch.distance ?? SHOT_SIZE_DISTANCE[shotSize], 1, 100, 7);
  const horizontal = distance * Math.cos(elevation * Math.PI / 180);
  const position = [
    target[0] + horizontal * Math.sin(azimuth * Math.PI / 180),
    target[1] + distance * Math.sin(elevation * Math.PI / 180),
    target[2] + horizontal * Math.cos(azimuth * Math.PI / 180)
  ];
  const fov = clamp(patch.fov ?? SHOT_SIZE_FOV[shotSize] ?? (50 / focalLength * 42), 20, 90, 42);
  return normalizeDirectorProject({ ...current, shotSize, focalLength, azimuth, elevation,
    directorCamera: { ...current.directorCamera, position, target, fov },
    camera: { ...current.camera, position, target, fov } });
}

export function directorReferenceAssetIds(project) {
  const current = normalizeDirectorProject(project);
  return [...new Set([current.scene?.assetId, ...current.objects.map(object => object.assetId)])].filter(Boolean);
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

export function addDirectorProp(project, prop = {}) {
  const current = normalizeDirectorProject(project);
  const assetId = text(prop.assetId || prop.id, 200);
  if (!assetId) return current;
  const id = availableId(prop.id || assetId, [{ id: 'scene' }, ...current.objects, ...current.cameras]);
  return normalizeDirectorProject({ ...current, props: [...current.props, { ...prop, kind: 'prop', assetId, id }], objects: [...current.objects, { ...prop, kind: 'prop', assetId, id }] });
}

export function updateDirectorActor(project, actorID, patch = {}) {
  const current = normalizeDirectorProject(project);
  const actors = current.actors.map(actor => actor.id === actorID ? { ...actor, ...patch, id: actor.id, assetId: actor.assetId } : actor);
  const props = current.props.slice();
  return normalizeDirectorProject({ ...current, actors, props, objects: [...actors, ...props] });
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

export function setDirectorSceneAsset(project, asset) {
  const current = normalizeDirectorProject(project);
  return normalizeDirectorProject({ ...current, scene: { ...current.scene, assetId: asset?.id, url: asset?.url, name: asset?.name } });
}

export function createDirectorPromptDraft(value) {
  const prompt = String(value || '');
  return { value: prompt, base: prompt, conflict: false };
}

export function syncDirectorPromptDraft(draft, external) {
  const prompt = String(external || '');
  if (draft.pending && prompt === draft.pending.normalized) {
    return draft.value === draft.pending.value ? createDirectorPromptDraft(prompt) : { value: draft.value, base: prompt, conflict: false };
  }
  if (draft.value === draft.base || draft.value === prompt) return createDirectorPromptDraft(prompt);
  return { ...draft, conflict: prompt !== draft.base };
}

export function editDirectorPromptDraft(draft, value, external) {
  return { value, base: String(external || ''), conflict: false, ...(!draft.conflict && draft.pending ? { pending: draft.pending } : {}) };
}

export function submitDirectorPromptDraft(draft, normalized) {
  return { ...draft, pending: { value: draft.value, normalized } };
}

export function discardDirectorPromptDraft(draft, external) {
  return createDirectorPromptDraft(external);
}

export function removeDirectorObject(project, objectID) {
  const current = normalizeDirectorProject(project);
  if (objectID === 'scene' || objectID === 'camera-main') return current;
  const cameras = current.cameras.length > 1 ? current.cameras.filter(camera => camera.id !== objectID) : current.cameras;
  const actors = current.actors.filter(actor => actor.id !== objectID);
  const props = current.props.filter(prop => prop.id !== objectID);
  return normalizeDirectorProject({ ...current, actors, props, objects: [...actors, ...props], cameras,
    activeCameraId: cameras.some(camera => camera.id === current.activeCameraId) ? current.activeCameraId : cameras[0].id });
}

export function directorSceneObjects(project) {
  const current = normalizeDirectorProject(project);
  return [{ ...current.scene, id: 'scene', kind: 'scene' }, ...current.actors.map(actor => ({ ...actor, kind: 'actor' })),
    ...current.props.map(prop => ({ ...prop, kind: 'prop' })),
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

export function mergeDirectorDeskConfig(project, edited) {
  const current = normalizeDirectorProject(project);
  const before = directorDeskConfig(current);
  const next = directorDeskConfig(edited);
  const actors = next.actors.map(actor => {
    const existing = current.actors.find(item => item.id === actor.id);
    const previous = before.actors.find(item => item.id === actor.id);
    if (!existing || !previous) return normalizeActor(actor);
    return {
      ...existing,
      position: [actor.x === previous.x ? existing.position[0] : (actor.x - 50) / 12,
        existing.position[1], actor.y === previous.y ? existing.position[2] : (actor.y - 65) / 12],
      rotation: [existing.rotation[0], existing.rotation[1], actor.rotation === previous.rotation ? existing.rotation[2] : actor.rotation * Math.PI / 180],
      scale: actor.scale === previous.scale ? existing.scale : [actor.scale, actor.scale, actor.scale]
    };
  });
  return normalizeDirectorProject({ ...current,
    scenePrompt: next.scenePrompt,
    shotPrompt: next.shotPrompt || current.shotPrompt,
    scene: { ...(next.sceneAsset ? { assetId: next.sceneAsset.id, url: next.sceneAsset.url, name: next.sceneAsset.name } : {}), prompt: next.scenePrompt },
    actors, camera: { ...current.camera, ...next.camera }
  });
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

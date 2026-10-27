import { directorCameraSemantics, directorObjectTransform, normalizeDirectorProject, updateDirectorActor, updateDirectorCamera } from '../../pages/agent-director-desk.js';

export function applyDirectorInspectorChange(project, selection, field, rawValue) {
  const current = normalizeDirectorProject(project);
  if ((typeof rawValue !== 'string' && typeof rawValue !== 'number') || String(rawValue).trim() === '') return current;
  const value = Number(rawValue);
  if (!Number.isFinite(value)) return current;
  const kind = selection?.kind;
  const object = kind === 'actor' ? current.actors.find(item => item.id === selection.id)
    : kind === 'prop' ? current.props.find(item => item.id === selection.id)
    : kind === 'camera' ? current.cameras.find(item => item.id === selection.id) : null;
  if (!object) return current;
  let patch;
  if (kind === 'actor' && field === 'uniformScale') patch = { scale: [value, value, value] };
  else if (kind === 'camera' && field === 'fov') patch = { fov: value };
  else {
    const match = /^(position|rotation|scale|target)\.([0-2])$/.exec(field);
    const allowed = ['actor', 'prop'].includes(kind) ? ['position', 'rotation', 'scale'] : ['position', 'rotation', 'target'];
    if (!match || !allowed.includes(match[1]) || (kind === 'camera' && match[1] === 'rotation' && match[2] !== '2')) return current;
    const vector = [...object[match[1]]];
    vector[Number(match[2])] = value;
    patch = { [match[1]]: vector };
  }
  if (kind === 'actor') return updateDirectorActor(current, object.id, patch);
  if (kind === 'prop') return directorObjectTransform(current, object.id, patch);
  return updateDirectorCamera(current, object.id, patch);
}

const storagePrefix = 'qiantie:agent-canvas-draft:';
const draftVersion = 1;

function safeString(value, fallback = '') {
  return typeof value === 'string' && value ? value : fallback;
}

function cleanElement(element) {
  if (!element || typeof element !== 'object' || Array.isArray(element)) return null;
  const { selected, ...persisted } = element;
  return persisted;
}

function cleanCanvas(canvas) {
  if (!canvas || typeof canvas !== 'object' || Array.isArray(canvas)) return null;
  if (!Array.isArray(canvas.nodes) || !Array.isArray(canvas.edges)) return null;

  const nodes = canvas.nodes.map(cleanElement).filter(Boolean);
  const edges = canvas.edges.map(cleanElement).filter(Boolean);
  if (nodes.length !== canvas.nodes.length || edges.length !== canvas.edges.length) return null;

  const result = {
    version: Number(canvas.version) || 1,
    revision: Number(canvas.revision) || 0,
    nodes,
    edges
  };
  if (canvas.viewport && typeof canvas.viewport === 'object' && !Array.isArray(canvas.viewport)) result.viewport = { ...canvas.viewport };
  if (canvas.appearance && typeof canvas.appearance === 'object' && !Array.isArray(canvas.appearance)) result.appearance = { ...canvas.appearance };
  if (canvas.viewState && typeof canvas.viewState === 'object' && !Array.isArray(canvas.viewState)) result.viewState = { ...canvas.viewState };
  return result;
}

export function canvasDraftKey(username, taskId) {
  return `${storagePrefix}${encodeURIComponent(safeString(username, 'guest'))}:${encodeURIComponent(safeString(taskId))}`;
}

export function loadCanvasDraft(storage, username, taskId) {
  if (!safeString(taskId)) return null;
  try {
    const raw = storage?.getItem(canvasDraftKey(username, taskId));
    if (!raw) return null;
    const value = JSON.parse(raw);
    if (!value || value.version !== draftVersion) return null;
    const canvas = cleanCanvas(value.canvas);
    if (!canvas) return null;
    return {
      canvas,
      baseRevision: Number(value.baseRevision) || 0,
      savedAt: safeString(value.savedAt)
    };
  } catch {
    return null;
  }
}

export function saveCanvasDraft(storage, username, taskId, canvas, baseRevision) {
  if (!safeString(taskId)) return false;
  const cleaned = cleanCanvas(canvas);
  if (!cleaned) return false;
  try {
    storage?.setItem(canvasDraftKey(username, taskId), JSON.stringify({
      version: draftVersion,
      canvas: cleaned,
      baseRevision: Number(baseRevision) || 0,
      savedAt: new Date().toISOString()
    }));
    return true;
  } catch {
    return false;
  }
}

export function clearCanvasDraft(storage, username, taskId) {
  if (!safeString(taskId)) return false;
  try {
    storage?.removeItem(canvasDraftKey(username, taskId));
    return true;
  } catch {
    return false;
  }
}

export function isDraftAtLeastAsFresh(draft, serverCanvas) {
  return Boolean(draft?.canvas && Number(draft.baseRevision) >= (Number(serverCanvas?.revision) || 0));
}

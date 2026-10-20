const supportedNodeTypes = new Set(['text', 'image', 'video', 'audio']);
const mediaNodeTypes = new Set(['image', 'video', 'audio']);

export function normalizeCanvasNodeType(value) {
  const type = String(value || '').trim().toLowerCase();
  return supportedNodeTypes.has(type) ? type : 'text';
}

export function isMediaCanvasNodeType(value) {
  return mediaNodeTypes.has(normalizeCanvasNodeType(value));
}

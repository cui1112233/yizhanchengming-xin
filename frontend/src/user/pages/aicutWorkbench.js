const layoutBounds = {
  agentWidth: [56, 520],
  mediaWidth: [220, 620]
};

function clamp(value, min, max) {
  const number = Number(value);
  if (!Number.isFinite(number)) return min;
  return Math.min(max, Math.max(min, Math.round(number)));
}

export function clampAICutLayout(layout = {}, viewport = {}) {
  const height = Math.max(480, Number(viewport.height) || 900);
  const valueOrDefault = (value, fallback) => value === undefined || value === null || value === '' ? fallback : value;
  return {
    agentWidth: clamp(valueOrDefault(layout.agentWidth, 350), ...layoutBounds.agentWidth),
    mediaWidth: clamp(valueOrDefault(layout.mediaWidth, 450), ...layoutBounds.mediaWidth),
    topHeight: clamp(valueOrDefault(layout.topHeight, 520), 260, Math.max(260, height - 220))
  };
}

export function timelineDropTime(clientX, rectLeft, scrollLeft = 0, zoom = 1) {
  const safeZoom = Math.min(4, Math.max(0.5, Number(zoom) || 1));
  const pixels = Math.max(0, (Number(clientX) || 0) - (Number(rectLeft) || 0) + (Number(scrollLeft) || 0));
  return Math.round(pixels / (72 * safeZoom) * 1000);
}

export function resolveAICutShortcut(event = {}, context = {}) {
  const key = String(event.key || '').toLowerCase();
  const modifier = Boolean(event.ctrlKey || event.metaKey);
  const typing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(String(event.targetTag || '').toUpperCase()) || Boolean(event.targetEditable);

  if (key === 'escape') return 'close';
  if (modifier && (key === '/' || key === '?')) return 'shortcuts';
  if (modifier && key === '1') return 'canvas-view';
  if (modifier && key === '2') return 'editor-view';
  if (modifier && key === 'n') return 'new-project';
  if (modifier && key === 's' && context.hasProject) return 'save';
  if (modifier && key === 'e' && context.hasProject) return 'export';
  if (modifier && key === 'z' && context.hasProject) return event.shiftKey ? 'redo' : 'undo';
  if (modifier && key === 'y' && context.hasProject) return 'redo';
  if (typing) return null;
  if ((event.code === 'Space' || key === ' ') && context.hasProject) return 'play';
  if (modifier && key === 'b' && context.hasProject && context.hasSelection) return 'split';
  if (key === 'b' && context.hasProject) return 'blade';
  if ((key === 'delete' || key === 'backspace') && context.hasProject && context.hasSelection) return 'delete';
  return null;
}

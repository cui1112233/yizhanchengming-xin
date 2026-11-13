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

export function nextTimelineFollowScrollLeft(options = {}) {
  const current = Math.max(0, Number(options.scrollLeft) || 0);
  if (options.enabled === false) return current;
  const viewportWidth = Math.max(1, Number(options.viewportWidth) || 0);
  const contentWidth = Math.max(viewportWidth, Number(options.contentWidth) || 0);
  const playheadX = Math.max(0, Number(options.playheadX) || 0);
  const maximum = Math.max(0, contentWidth - viewportWidth);
  const leftBoundary = current + viewportWidth * .2;
  const rightBoundary = current + viewportWidth * .8;
  let next = current;
  if (playheadX < leftBoundary) next = playheadX - viewportWidth * .2;
  else if (playheadX > rightBoundary) next = playheadX - viewportWidth * .25;
  return Math.min(maximum, Math.max(0, Math.round(next)));
}

export function draggedTransitionDuration(originDurationMs, deltaPixels, zoom = 1, maximumMs = 3000) {
  const safeZoom = Math.min(4, Math.max(0.5, Number(zoom) || 1));
  const maximum = Math.max(100, Math.round(Number(maximumMs) || 100));
  const origin = Math.min(maximum, Math.max(100, Math.round(Number(originDurationMs) || 500)));
  const deltaMs = Math.round((Number(deltaPixels) || 0) / (72 * safeZoom) * 1000);
  return Math.min(maximum, Math.max(100, origin - deltaMs));
}

export function resolveAICutShortcut(event = {}, context = {}) {
  const key = String(event.key || '').toLowerCase();
  const modifier = Boolean(event.ctrlKey || event.metaKey);
  const typing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(String(event.targetTag || '').toUpperCase()) || Boolean(event.targetEditable);

  if (key === 'escape') return 'close';
  if (modifier && (key === '/' || key === '?')) return 'shortcuts';
  if (modifier && key === '1') return 'canvas-view';
  if (modifier && key === 'n') return 'new-project';
  if (modifier && key === 's' && context.hasProject) return 'save';
  if (modifier && key === 'e' && context.hasProject) return 'export';
  if (modifier && key === 'z' && context.hasProject) return event.shiftKey ? 'redo' : 'undo';
  if (modifier && key === 'y' && context.hasProject) return 'redo';
  if (typing) return null;
  if (modifier && key === 'c' && context.hasProject && context.hasSelection) return 'copy';
  if (modifier && key === 'x' && context.hasProject && context.hasSelection) return 'cut';
  if (modifier && key === 'v' && context.hasProject && context.hasClipboard) return 'paste';
  if (event.altKey && key === 'arrowleft' && context.hasProject && context.hasSelection) return 'nudge-previous';
  if (event.altKey && key === 'arrowright' && context.hasProject && context.hasSelection) return 'nudge-next';
  if (modifier && key === 'arrowleft' && context.hasProject) return 'edit-previous';
  if (modifier && key === 'arrowright' && context.hasProject) return 'edit-next';
  if ((event.code === 'Space' || key === ' ') && context.hasProject) return 'play';
  if (modifier && key === 'b' && context.hasProject) return 'split';
  if (key === 'b' && context.hasProject) return 'blade';
  if ((key === 'delete' || key === 'backspace') && context.hasProject && context.hasSelection) return 'delete';
  return null;
}

export const defaultEditorLayout = Object.freeze({
  resourceWidth: 0.20,
  inspectorWidth: 0.20,
  timelineHeight: 0.34,
  resourceCollapsed: false,
  inspectorCollapsed: false,
  timelineCollapsed: false,
  playheadFollow: true,
});

function dimension(value, total, fallbackRatio, minimum, maximum) {
  const parsed = Number(value);
  const fallback = Math.round(total * fallbackRatio);
  const pixels = Number.isFinite(parsed) ? (parsed > 0 && parsed < 1 ? Math.round(total * parsed) : Math.round(parsed)) : fallback;
  return Math.min(maximum, Math.max(minimum, pixels));
}

export function normalizeEditorLayout(layout = {}, viewport = {}) {
  const width = Math.max(760, Number(viewport.width) || 1440);
  const height = Math.max(600, Number(viewport.height) || 900);
  const inspectorMode = width < 1280 ? 'drawer' : 'panel';
  const inspectorWidth = dimension(layout.inspectorWidth, width, defaultEditorLayout.inspectorWidth, 280, inspectorMode === 'drawer' ? 520 : 400);
  const requestedResourceWidth = dimension(layout.resourceWidth, width, defaultEditorLayout.resourceWidth, 260, 420);
  const resourceWidth = inspectorMode === 'panel' && !layout.inspectorCollapsed
    ? Math.min(requestedResourceWidth, Math.max(260, width - inspectorWidth - 480))
    : requestedResourceWidth;
  return {
    resourceWidth,
    inspectorWidth,
    timelineHeight: dimension(layout.timelineHeight, height, defaultEditorLayout.timelineHeight, 225, Math.round(height * 0.7)),
    resourceCollapsed: Boolean(layout.resourceCollapsed),
    inspectorCollapsed: Boolean(layout.inspectorCollapsed),
    timelineCollapsed: Boolean(layout.timelineCollapsed),
    playheadFollow: layout.playheadFollow !== false,
    inspectorMode,
  };
}

export function toggleEditorPane(layout, pane) {
  const key = { resource: 'resourceCollapsed', inspector: 'inspectorCollapsed', timeline: 'timelineCollapsed' }[pane];
  if (!key) return layout;
  return { ...layout, [key]: !layout[key] };
}

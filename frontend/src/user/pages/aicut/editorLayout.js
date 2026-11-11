export const defaultEditorLayout = Object.freeze({
  resourceWidth: 0.22,
  inspectorWidth: 0.22,
  timelineHeight: 0.4,
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
  return {
    resourceWidth: dimension(layout.resourceWidth, width, defaultEditorLayout.resourceWidth, 260, 520),
    inspectorWidth: dimension(layout.inspectorWidth, width, defaultEditorLayout.inspectorWidth, 280, 520),
    timelineHeight: dimension(layout.timelineHeight, height, defaultEditorLayout.timelineHeight, 225, Math.round(height * 0.7)),
    resourceCollapsed: Boolean(layout.resourceCollapsed),
    inspectorCollapsed: Boolean(layout.inspectorCollapsed),
    timelineCollapsed: Boolean(layout.timelineCollapsed),
    playheadFollow: layout.playheadFollow !== false,
    inspectorMode: width < 1180 ? 'drawer' : 'panel',
  };
}

export function toggleEditorPane(layout, pane) {
  const key = { resource: 'resourceCollapsed', inspector: 'inspectorCollapsed', timeline: 'timelineCollapsed' }[pane];
  if (!key) return layout;
  return { ...layout, [key]: !layout[key] };
}

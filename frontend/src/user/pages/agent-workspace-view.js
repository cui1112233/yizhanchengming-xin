export const DEFAULT_WORKSPACE_VIEW = Object.freeze({
  activeView: 'workflow',
  drawerOpen: true,
  drawerTab: 'nodes',
  drawerWidth: 288,
  conversationOpen: true,
  conversationWidth: 360,
  snapToGrid: false,
  showMinimap: true,
  showConnections: true
});

function clamp(value, minimum, maximum, fallback) {
  const number = Number(value);
  if (!Number.isFinite(number)) return fallback;
  return Math.min(maximum, Math.max(minimum, Math.round(number)));
}

export function workspaceViewQuery(value) {
  return value === 'storyboard' ? 'storyboard' : 'workflow';
}

export function normalizeWorkspaceView(value = {}) {
  const source = value && typeof value === 'object' && !Array.isArray(value) ? value : {};
  return {
    activeView: workspaceViewQuery(source.activeView),
    drawerOpen: typeof source.drawerOpen === 'boolean' ? source.drawerOpen : DEFAULT_WORKSPACE_VIEW.drawerOpen,
    drawerTab: source.drawerTab === 'assets' ? 'assets' : 'nodes',
    drawerWidth: clamp(source.drawerWidth, 240, 460, DEFAULT_WORKSPACE_VIEW.drawerWidth),
    conversationOpen: typeof source.conversationOpen === 'boolean' ? source.conversationOpen : DEFAULT_WORKSPACE_VIEW.conversationOpen,
    conversationWidth: clamp(source.conversationWidth, 320, 520, DEFAULT_WORKSPACE_VIEW.conversationWidth),
    snapToGrid: typeof source.snapToGrid === 'boolean' ? source.snapToGrid : DEFAULT_WORKSPACE_VIEW.snapToGrid,
    showMinimap: typeof source.showMinimap === 'boolean' ? source.showMinimap : DEFAULT_WORKSPACE_VIEW.showMinimap,
    showConnections: typeof source.showConnections === 'boolean' ? source.showConnections : DEFAULT_WORKSPACE_VIEW.showConnections
  };
}

export function workspaceViewPatch(current, patch) {
  return normalizeWorkspaceView({ ...normalizeWorkspaceView(current), ...(patch || {}) });
}

import { DEFAULT_STORYBOARD_STATE, normalizeStoryboardState, patchStoryboardState } from './agent-storyboard-state.js';

export const DEFAULT_WORKSPACE_VIEW = Object.freeze({
  activeView: 'workflow',
  drawerOpen: false,
  drawerTab: 'nodes',
  drawerWidth: 288,
  // The canvas should open focused on the work surface. The conversation is
  // available from the bottom circular control and must be opt-in.
  conversationOpen: false,
  conversationWidth: 360,
  snapToGrid: false,
  showMinimap: true,
  showConnections: true,
  storyboard: DEFAULT_STORYBOARD_STATE
});

function clamp(value, minimum, maximum, fallback) {
  const number = Number(value);
  if (!Number.isFinite(number)) return fallback;
  return Math.min(maximum, Math.max(minimum, Math.round(number)));
}

export function workspaceViewQuery(value) {
  return ['workflow', 'storyboard', 'assets', 'tasks'].includes(value) ? value : 'workflow';
}

export function workspaceViewForActiveView(current, activeView) {
  const nextView = workspaceViewQuery(activeView);
  return workspaceViewPatch(current, {
    activeView: nextView,
    drawerOpen: false,
    conversationOpen: nextView === 'storyboard'
  });
}

export function normalizeWorkspaceView(value = {}) {
  const source = value && typeof value === 'object' && !Array.isArray(value) ? value : {};
  return {
    activeView: workspaceViewQuery(source.activeView),
    drawerOpen: typeof source.drawerOpen === 'boolean' ? source.drawerOpen : DEFAULT_WORKSPACE_VIEW.drawerOpen,
    drawerTab: ['nodes', 'assets', 'history'].includes(source.drawerTab) ? source.drawerTab : 'nodes',
    drawerWidth: clamp(source.drawerWidth, 240, 460, DEFAULT_WORKSPACE_VIEW.drawerWidth),
    conversationOpen: typeof source.conversationOpen === 'boolean' ? source.conversationOpen : DEFAULT_WORKSPACE_VIEW.conversationOpen,
    conversationWidth: clamp(source.conversationWidth, 320, 520, DEFAULT_WORKSPACE_VIEW.conversationWidth),
    snapToGrid: typeof source.snapToGrid === 'boolean' ? source.snapToGrid : DEFAULT_WORKSPACE_VIEW.snapToGrid,
    showMinimap: typeof source.showMinimap === 'boolean' ? source.showMinimap : DEFAULT_WORKSPACE_VIEW.showMinimap,
    showConnections: typeof source.showConnections === 'boolean' ? source.showConnections : DEFAULT_WORKSPACE_VIEW.showConnections,
    storyboard: normalizeStoryboardState(source.storyboard)
  };
}

export function workspaceViewPatch(current, patch) {
  const normalized = normalizeWorkspaceView(current);
  const nextPatch = patch && typeof patch === 'object' && !Array.isArray(patch) ? patch : {};
  return normalizeWorkspaceView({
    ...normalized,
    ...nextPatch,
    storyboard: Object.hasOwn(nextPatch, 'storyboard')
      ? patchStoryboardState(normalized.storyboard, nextPatch.storyboard)
      : normalized.storyboard
  });
}

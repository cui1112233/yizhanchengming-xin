import test from 'node:test';
import assert from 'node:assert/strict';
import {
  DEFAULT_WORKSPACE_VIEW,
  normalizeWorkspaceView,
  workspaceViewPatch,
  workspaceViewQuery
} from './agent-workspace-view.js';

test('old canvases receive stable workflow workspace defaults', () => {
  assert.deepEqual(normalizeWorkspaceView(null), DEFAULT_WORKSPACE_VIEW);
  assert.deepEqual(DEFAULT_WORKSPACE_VIEW, {
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
});

test('workspace view normalization accepts known values and clamps panel widths', () => {
  assert.deepEqual(normalizeWorkspaceView({
    activeView: 'storyboard',
    drawerOpen: false,
    drawerTab: 'assets',
    drawerWidth: 900,
    conversationOpen: false,
    conversationWidth: 100,
    snapToGrid: true,
    showMinimap: false,
    showConnections: false,
    ignored: 'private'
  }), {
    activeView: 'storyboard',
    drawerOpen: false,
    drawerTab: 'assets',
    drawerWidth: 460,
    conversationOpen: false,
    conversationWidth: 320,
    snapToGrid: true,
    showMinimap: false,
    showConnections: false
  });
});

test('workspace patches preserve untouched state and URL view is normalized', () => {
  const current = normalizeWorkspaceView({ activeView: 'storyboard', drawerWidth: 340 });
  assert.deepEqual(workspaceViewPatch(current, { drawerOpen: false }), {
    ...current,
    drawerOpen: false
  });
  assert.equal(workspaceViewQuery('storyboard'), 'storyboard');
  assert.equal(workspaceViewQuery('unknown'), 'workflow');
});

import test from 'node:test';
import assert from 'node:assert/strict';
import {
  DEFAULT_WORKSPACE_VIEW,
  normalizeWorkspaceView,
  workspaceViewForActiveView,
  workspaceViewPatch,
  workspaceViewQuery
} from './agent-workspace-view.js';

test('old canvases receive stable workflow workspace defaults', () => {
  assert.deepEqual(normalizeWorkspaceView(null), DEFAULT_WORKSPACE_VIEW);
  assert.deepEqual(DEFAULT_WORKSPACE_VIEW, {
    activeView: 'workflow',
    drawerOpen: false,
    drawerTab: 'nodes',
    drawerWidth: 288,
    conversationOpen: false,
    conversationWidth: 360,
    snapToGrid: false,
    showMinimap: true,
    showConnections: true,
    storyboard: {
      directorMode: 'supervised',
      selectedKey: '',
      filters: { text: 'all', audio: 'all', image: 'all', video: 'all' },
      categories: {},
      order: [],
      libraryWidth: 340,
      sectionHeights: { text: 220, audio: 240, image: 320 },
      layout: {
        director: 31,
        column: 25,
        leftRow: 66,
        rightRow: 55,
        sectionOrder: ['text', 'audio', 'image', 'video']
      }
    }
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
    showConnections: false,
    storyboard: {
      directorMode: 'supervised',
      selectedKey: '',
      filters: { text: 'all', audio: 'all', image: 'all', video: 'all' },
      categories: {},
      order: [],
      libraryWidth: 340,
      sectionHeights: { text: 220, audio: 240, image: 320 },
      layout: {
        director: 31,
        column: 25,
        leftRow: 66,
        rightRow: 55,
        sectionOrder: ['text', 'audio', 'image', 'video']
      }
    }
  });
});

test('workspace view keeps the generation history drawer tab', () => {
  assert.equal(normalizeWorkspaceView({ drawerTab: 'history' }).drawerTab, 'history');
  assert.equal(workspaceViewPatch(DEFAULT_WORKSPACE_VIEW, { drawerTab: 'history' }).drawerTab, 'history');
});

test('workspace patch 深合并故事板状态并让分类选择随画布保存', () => {
  const current = normalizeWorkspaceView({
    storyboard: {
      selectedKey: 'asset:image-1',
      filters: { image: 'scene' },
      categories: { 'asset:image-1': 'scene' },
      order: ['asset:image-1', 'asset:image-2'],
      libraryWidth: 412,
      sectionHeights: { image: 410 }
    }
  });
  const next = workspaceViewPatch(current, {
    storyboard: { filters: { image: 'character' } }
  });

  assert.equal(next.storyboard.selectedKey, 'asset:image-1');
  assert.equal(next.storyboard.filters.image, 'character');
  assert.equal(next.storyboard.categories['asset:image-1'], 'scene');
  assert.deepEqual(next.storyboard.order, ['asset:image-1', 'asset:image-2']);
  assert.equal(next.storyboard.libraryWidth, 412);
  assert.equal(next.storyboard.sectionHeights.image, 410);
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

test('view switching gives workflow the canvas first while storyboard restores Director', () => {
  const closed = normalizeWorkspaceView({ activeView: 'workflow', drawerOpen: true, conversationOpen: false });
  assert.equal(workspaceViewForActiveView(closed, 'storyboard').conversationOpen, true);
  assert.equal(workspaceViewForActiveView(closed, 'storyboard').activeView, 'storyboard');
  const workflow = workspaceViewForActiveView({ ...closed, conversationOpen: true }, 'workflow');
  assert.equal(workflow.conversationOpen, false);
  assert.equal(workflow.drawerOpen, false);
});

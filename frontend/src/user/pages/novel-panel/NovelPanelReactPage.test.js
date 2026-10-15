import test from 'node:test';
import assert from 'node:assert/strict';
import { createNovelPanelViewState } from './novelPanelPageState.js';

test('shows a loading state while the project request is pending', () => {
  assert.deepEqual(createNovelPanelViewState({ projectId: 'p-1', loading: true }), {
    kind: 'loading',
    project: null,
    error: null,
    canFallback: true
  });
});

test('shows the project selector when project id is missing', () => {
  const state = createNovelPanelViewState({ projectId: '', loading: false });
  assert.equal(state.kind, 'select');
  assert.deepEqual(state.projects, []);
});

test('returns a project-ready state after normalization', () => {
  const state = createNovelPanelViewState({
    projectId: 'p-1',
    loading: false,
    project: { id: 'p-1', name: '测试项目', data: {} }
  });
  assert.equal(state.kind, 'ready');
  assert.equal(state.project.id, 'p-1');
  assert.equal(state.canFallback, true);
});

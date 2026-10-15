import { normalizeNovelPanelError, normalizeNovelProject } from './novelPanelContracts.js';

export function createNovelPanelViewState({ projectId, loading = false, project = null, error = null } = {}) {
  if (!String(projectId || '').trim()) {
    return {
      kind: 'error',
      project: null,
      error: { code: 400, message: '缺少小说面板项目', retryable: false },
      canFallback: true
    };
  }
  if (loading) return { kind: 'loading', project: null, error: null, canFallback: true };
  if (error) return { kind: 'error', project: null, error: normalizeNovelPanelError(error), canFallback: true };
  if (!project) {
    return {
      kind: 'error',
      project: null,
      error: { code: 404, message: '找不到小说面板项目', retryable: false },
      canFallback: true
    };
  }
  return { kind: 'ready', project: normalizeNovelProject(project), error: null, canFallback: true };
}

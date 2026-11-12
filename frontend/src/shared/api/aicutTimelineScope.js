let currentScope = { projectId: '', timelineId: 'default' };
export function setAICutTimelineRequestScope(projectId = '', timelineId = 'default') {
  currentScope = { projectId: String(projectId), timelineId: String(timelineId || 'default') };
}
export function scopedAICutOptions(path, options = {}, scope = currentScope) {
  const prefix = `/api/ai-cut/projects/${scope.projectId}`;
  if (!scope.projectId || scope.timelineId === 'default' || !(path === prefix || path.startsWith(`${prefix}/`))) return options;
  const headers = new Headers(options.headers || {});
  headers.set('X-AICut-Timeline-Id', scope.timelineId);
  return { ...options, headers };
}

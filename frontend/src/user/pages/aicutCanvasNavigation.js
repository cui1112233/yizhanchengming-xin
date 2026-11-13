function validId(value) {
  if (typeof value !== 'string' || !/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(value)) throw new Error('工程或画布编号无效');
  return value;
}
export function creativeCanvasHref(taskId, timelineId = 'default') {
  const query = new URLSearchParams({ taskId: validId(taskId), view: 'workflow' });
  if (timelineId !== 'default') query.set('cutTimeline', validId(timelineId));
  return `/agent/canvas?${query}`;
}
export function cutEditorHref(projectId, timelineId = 'default') {
  const query = new URLSearchParams({ project: validId(projectId) });
  if (timelineId !== 'default') query.set('timeline', validId(timelineId));
  return `/ai-cut?${query}`;
}
export function creativeTimelineHint(search, taskId) {
  const query = new URLSearchParams(search);
  return query.get('taskId') === taskId ? query.get('cutTimeline') || 'default' : 'default';
}
export async function persistLatestCanvasNavigation({ scope, getScope, isDirty, save }) {
  const check = () => { if (getScope() !== scope) throw new Error('工程或时间线已切换'); };
  check();
  while (isDirty()) {
    const saved = await save(); check();
    if (!saved) throw new Error('当前时间线保存失败，请重试');
  }
}
export async function openBoundCreativeCanvas({ projectId, timelineId = 'default' }, { scope, getScope, save, ensure, navigate }) {
  validId(projectId); validId(timelineId);
  const check = () => { if (getScope() !== scope) throw new Error('工程或时间线已切换'); };
  check(); await save(); check();
  const binding = await ensure(projectId); check();
  const href = creativeCanvasHref(binding?.taskId, timelineId);
  // Binding may take time; preserve edits made while that request was pending.
  await save(); check();
  navigate(href);
}

import { apiRequest } from './client.js';

function timelinePath(projectId, timelineId = '', suffix = '') {
  const valid = value => typeof value === 'string' && /^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(value);
  if (!valid(projectId) || (timelineId && !valid(timelineId))) throw new Error('工程或时间线无效');
  return `/api/ai-cut/projects/${projectId}/timelines${timelineId ? `/${timelineId}` : ''}${suffix}`;
}
export async function listProjectTimelines(projectId, request = apiRequest) { return (await request(timelinePath(projectId))).timelines || []; }
export async function createProjectTimeline(projectId, title, request = apiRequest) { return (await request(timelinePath(projectId), { method: 'POST', body: JSON.stringify({ title }) })).timeline; }
export async function renameProjectTimeline(projectId, id, title, request = apiRequest) { return (await request(timelinePath(projectId, id), { method: 'PATCH', body: JSON.stringify({ title }) })).timeline; }
export async function duplicateProjectTimeline(projectId, id, title, request = apiRequest) { return (await request(timelinePath(projectId, id, '/duplicate'), { method: 'POST', body: JSON.stringify({ title }) })).timeline; }
export async function deleteProjectTimeline(projectId, id, request = apiRequest) { return request(timelinePath(projectId, id), { method: 'DELETE' }); }
export async function reorderProjectTimelines(projectId, timelineIds, request = apiRequest) { return request(timelinePath(projectId, '', '/reorder'), { method: 'POST', body: JSON.stringify({ timelineIds }) }); }
export async function loadProjectTimeline(projectId, id, request = apiRequest) { return (await request(timelinePath(projectId, id, '/snapshot'))).timeline; }
export async function saveProjectTimeline(projectId, id, timeline, request = apiRequest) { return (await request(timelinePath(projectId, id, '/snapshot'), { method: 'PUT', body: JSON.stringify({ expectedRevision: timeline.revision, timeline }) })).timeline; }
export async function projectTimelineHistory(projectId, id, request = apiRequest) { return (await request(timelinePath(projectId, id, '/history'))).history; }
export async function moveProjectTimelineHistory(projectId, id, action, expectedRevision, request = apiRequest) {
  if (!['undo', 'redo'].includes(action)) throw new Error('历史操作无效');
  return request(timelinePath(projectId, id, `/history/${action}`), { method: 'POST', body: JSON.stringify({ expectedRevision }) });
}

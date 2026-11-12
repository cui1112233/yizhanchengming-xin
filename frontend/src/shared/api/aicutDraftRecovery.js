import { createProjectTimeline, loadProjectTimeline, saveProjectTimeline } from './aicutTimelines.js';

const content = timeline => JSON.stringify({ ...timeline, revision: 0 });
export async function saveTimelineDraftCopy(projectId, title, draft, existingId = '', api = { createProjectTimeline, loadProjectTimeline, saveProjectTimeline }) {
  let id = existingId;
  try {
    if (!id) id = (await api.createProjectTimeline(projectId, title)).id;
    const current = existingId ? await api.loadProjectTimeline(projectId, id) : { revision: 0, tracks: [] };
    if (current.revision > 0) {
      if (content(current) !== content(draft)) throw new Error('恢复副本已经发生变化，请检查该副本后重试');
      return { id, timeline: current };
    }
    const saved = await api.saveProjectTimeline(projectId, id, { ...structuredClone(draft), revision: 0 });
    return { id, timeline: saved };
  } catch (error) { error.createdTimelineId = id; throw error; }
}

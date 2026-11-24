export type CutRecord = Record<string, any>;
export type CutTimeline = CutRecord & { revision: number; tracks: CutRecord[] };
export type CutRequest = (path: string, options?: Record<string, unknown>) => Promise<any>;
export const cutID = (value: string) => {
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(value)) throw new Error('工程或任务编号无效');
  return value;
};
export const cutPath = (project: string, suffix = '') => `/api/ai-cut/projects/${cutID(project)}${suffix}`;
const scope = (id: string) => ({ 'X-AICut-Timeline-Id': cutID(id) });
const array = (value: any, key: string, label: string): CutRecord[] => {
  if (!Array.isArray(value?.[key])) throw new Error(`${label}读取失败：返回格式不正确`);
  return value[key];
};
export function createAICutClient(request: CutRequest) {
  return {
    async list() { return array(await request('/api/ai-cut/projects'), 'projects', '工程列表'); },
    async create(title: string, aspectRatio: string) {
      if (!title.trim() || !['9:16', '16:9', '1:1'].includes(aspectRatio)) throw new Error('请填写名称并选择画幅');
      const response = await request('/api/ai-cut/projects', { method: 'POST', data: { title: title.trim(), aspectRatio } });
      if (!response?.project?.id) throw new Error('未确认新工程，请刷新工程列表后核对');
      return response.project as CutRecord;
    },
    async rename(project: string, title: string) {
      if (!title.trim()) throw new Error('工程名称不能为空');
      return request(cutPath(project), { method: 'PATCH', data: { title: title.trim() } });
    },
    async timelines(project: string) { return array(await request(cutPath(project, '/timelines')), 'timelines', '时间线'); },
    async assets(project: string) { return array(await request(cutPath(project, '/assets')), 'assets', '素材'); },
    async timeline(project: string, id: string): Promise<CutTimeline> {
      const response = await request(cutPath(project, `/timelines/${cutID(id)}/snapshot`), { header: scope(id) });
      if (!response?.timeline || !Array.isArray(response.timeline.tracks) || !Number.isSafeInteger(response.timeline.revision)) throw new Error('时间线读取失败');
      return response.timeline;
    },
    async save(project: string, id: string, timeline: CutTimeline): Promise<CutTimeline> {
      const response = await request(cutPath(project, `/timelines/${cutID(id)}/snapshot`), { method: 'PUT', header: scope(id), data: { expectedRevision: timeline.revision, timeline } });
      if (!response?.timeline || !Number.isSafeInteger(response.timeline.revision) || !Array.isArray(response.timeline.tracks)) throw new Error('保存结果尚未确认，请重新读取工程核对');
      return response.timeline;
    },
    async jobs(project: string, id: string) { return array(await request(cutPath(project, '/export-jobs'), { header: scope(id) }), 'jobs', '导出任务'); },
    async export(project: string, id: string, revision: number) {
      const response = await request(cutPath(project, '/export-jobs'), { method: 'POST', header: scope(id), data: { format: 'mp4', expectedRevision: revision, resolution: '1080p', frameRate: 30, quality: 'standard', codec: 'h264', bitDepth: 8 } });
      if (!response?.job?.id) throw new Error('导出任务未确认，请刷新任务列表核对，勿重复提交');
      return response.job as CutRecord;
    },
    async cancel(project: string, id: string, job: string) { return request(cutPath(project, `/export-jobs/${cutID(job)}/cancel`), { method: 'POST', header: scope(id), data: {} }); },
  };
}

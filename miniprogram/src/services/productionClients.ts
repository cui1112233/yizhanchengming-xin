export type Row = Record<string, any>;
export type ProductionRequest = (path: string, options?: { method?: "GET" | "POST" | "PUT"; data?: unknown; timeout?: number }) => Promise<any>;
export function rows(payload: Row, key: string): Row[] {
  if (!Array.isArray(payload?.[key])) throw new Error(`服务端未返回有效的 ${key}`);
  return payload[key];
}
function required(value: string) { const text = String(value || '').trim(); if (!text) throw new Error('请先选择任务、分镜与模型并填写内容'); return text; }
export function createProductionClients(request: ProductionRequest) {
  const agentPath = (id: string) => `/api/agent/tasks/${encodeURIComponent(required(id))}`;
  const waterPath = (id: string) => `/api/shuihuo-production/projects/${encodeURIComponent(required(id))}`;
  return {
    async agentTasks() { return rows(await request('/api/agent/tasks'), 'tasks'); },
    async createAgent() { const result = await request('/api/agent/tasks', { method: 'POST' }); if (!result?.task?.id) throw new Error('服务端未返回创建的任务'); return result.task as Row; },
    async agentDetail(id: string) {
      const path = agentPath(id);
      const [detail, canvas, executions] = await Promise.all([request(path), request(`${path}/canvas`), request(`${path}/executions`)]);
      if (!detail?.task?.id || !Array.isArray(detail.task.messages) || !canvas?.canvas) throw new Error('Agent 任务或画布响应不完整');
      return { task: detail.task as Row, canvas: canvas.canvas as Row, executions: rows(executions, 'executions') };
    },
    ask(id: string, prompt: string) { return request('/api/agent/chat', { method: 'POST', data: { taskId: required(id), prompt: required(prompt), context: { page: 'Agent 工作区' }, skillIds: [] }, timeout: 180000 }); },
    async waterProjects() { return rows(await request('/api/shuihuo-production/projects'), 'projects'); },
    async waterModels() { return rows(await request('/api/shuihuo-production/models'), 'models'); },
    async waterCatalog(onProjects:(value:Row[])=>void,onModels:(value:Row[])=>void) {
      await Promise.all([
        request('/api/shuihuo-production/projects').then(value=>onProjects(rows(value,'projects'))),
        request('/api/shuihuo-production/models').then(value=>onModels(rows(value,'models'))),
      ]);
    },
    async createWater(name: string, sourceText = '', onCreated?:(project:Row)=>void) {
      const result = await request('/api/shuihuo-production/projects', { method: 'POST', data: { name: required(name) } });
      if (!result?.id) throw new Error('服务端未返回项目 ID');
      onCreated?.(result);
      if (sourceText.trim()) {
        try { await request(`${waterPath(String(result.id))}/source`, { method: 'PUT', data: { sourceText: sourceText.trim() } }); }
        catch (cause: any) { throw Object.assign(new Error(`项目 #${result.id} 已创建，但原文保存失败：${cause.message || '请刷新作品库后重试'}`),{createdProject:result}); }
      }
      return result as Row;
    },
    async waterDetail(id: string) {
      const [detail, tasks] = await Promise.all([request(waterPath(id)), request(`${waterPath(id)}/tasks`)]);
      if (!detail?.project?.id) throw new Error('水货项目响应不完整');
      return { ...detail, segments: rows(detail, 'segments'), assets: rows(detail, 'assets'), media: rows(detail, 'media'), tasks: rows(tasks, 'tasks') } as Row;
    },
    saveWaterSource(id: string, sourceText: string) { return request(`${waterPath(id)}/source`, { method: 'PUT', data: { sourceText: required(sourceText) } }); },
    async waterCandidates(id: string) { return rows(await request(`${waterPath(id)}/segmentation/paragraphs`, { method: 'POST', data: {} }), 'candidates'); },
    async waterSmartCandidates(id:string){return rows(await request(`${waterPath(id)}/segmentation/smart`,{method:'POST',data:{},timeout:180000}),'candidates');},
    waterTaskAction(id:string,action:'retry'|'cancel'){return request(`/api/shuihuo-production/tasks/${encodeURIComponent(required(id))}/${action}`,{method:action==='retry'?'POST':'PUT',data:{}});},
    confirmWaterCandidates(id: string, candidates: Row[]) {
      if (!candidates.length || candidates.some(row => !String(row.text || '').trim())) return Promise.reject(new Error('请先填写有效的候选分镜'));
      return request(`${waterPath(id)}/segmentation/confirm`, { method: 'POST', data: { candidates: candidates.map(row => ({ text: String(row.text).trim(), speaker: String(row.speaker || '旁白').trim() })) } });
    },
    saveWaterSegment(segment: Row) { return request(`/api/shuihuo-production/segments/${encodeURIComponent(required(String(segment.id || '')))}`, { method: 'PUT', data: segment }); },
    saveWaterAsset(id: string, asset: Row) {
      const data = { ...asset, name: required(asset.name), source: 'manual', manuallyEdited: true };
      return request(asset.id ? `/api/shuihuo-production/assets/${encodeURIComponent(asset.id)}` : `${waterPath(id)}/assets`, { method: asset.id ? 'PUT' : 'POST', data });
    },
    async waterTask(id: string, segmentId: string, modelId: string, kind: string) { return request(`${waterPath(id)}/tasks`, { method: 'POST', data: { segmentId: required(segmentId), modelId: required(modelId), kind }, timeout: 180000 }); },
  };
}
export function latestRequestGuard() { let sequence = 0; return { begin: () => ++sequence, current: (id: number) => id === sequence, invalidate: () => { sequence += 1; } }; }

// This endpoint resolves provider modelId, not the catalog entry id. Ambiguous channels cannot be selected safely.
export function waterImageModels(models:Row[]){
 const enabled=models.filter(item=>item.kind==='image'&&item.enabled!==false&&String(item.modelId||'').trim());
 return enabled.filter(item=>enabled.filter(other=>other.modelId===item.modelId).length===1);
}

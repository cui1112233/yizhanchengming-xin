import { apiRequest } from './client.js';
import { comicWorkflowPath } from './comicScreenplayWorkflow.js';

function jobPath(projectId, jobId) {
  const id = String(jobId || '');
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id)) throw new Error('检查任务无效');
  return comicWorkflowPath(projectId, `asset-checks/${id}`);
}
const options = body => ({method:'POST', body:JSON.stringify(body), timeoutMs:15000, suppressGlobalError:true});
export function listComicAssetChecks(projectId, request=apiRequest) {
  return request(comicWorkflowPath(projectId,'asset-checks'), {timeoutMs:15000,suppressGlobalError:true});
}
export function getComicAssetCheck(projectId, jobId, request=apiRequest) {
  return request(jobPath(projectId,jobId), {timeoutMs:15000,suppressGlobalError:true});
}
export function startComicAssetCheck(projectId, input, request=apiRequest) {
  if (input?.confirmed !== true) throw new Error('请先确认使用文字模型检查资产');
  const requestKey=String(input?.requestKey||'').trim(), model=String(input?.model||'').trim();
  if (!requestKey || requestKey.length>256) throw new Error('检查请求标识无效');
  if (!['script','original'].includes(input?.sourceMode)) throw new Error('检查来源无效');
  if (!model) throw new Error('请选择文字模型');
  if (!Array.isArray(input?.episodeNumbers) || !input.episodeNumbers.length || input.episodeNumbers.some(no=>!Number.isInteger(no)||no<1)) throw new Error('请选择有效集数');
  const body={requestKey,model,sourceMode:input.sourceMode,episodeNumbers:[...new Set(input.episodeNumbers)].sort((a,b)=>a-b),confirmed:true};
  return request(comicWorkflowPath(projectId,'asset-checks'), options(body));
}
export function cancelComicAssetCheck(projectId, jobId, request=apiRequest) {
  return request(jobPath(projectId,jobId)+'/cancel',options({}));
}
export function applyComicAssetCheck(projectId, jobId, expectedRevision, request=apiRequest) {
  if (!Number.isInteger(expectedRevision)||expectedRevision<0) throw new Error('检查结果版本无效，请重新读取');
  return request(jobPath(projectId,jobId)+'/apply',options({expectedRevision}));
}

export function discardComicAssetCheck(projectId, jobId, expectedRevision, request=apiRequest) {
  if (!Number.isInteger(expectedRevision)||expectedRevision<0) throw new Error('检查结果版本无效，请重新读取');
  return request(jobPath(projectId,jobId)+'/discard',options({expectedRevision}));
}

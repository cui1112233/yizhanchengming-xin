import { apiRequest } from './client.js';
import { comicWorkflowPath } from './comicScreenplayWorkflow.js';
export function listComicAssetLibrary({ type }, request = apiRequest) {
  if (!['character','scene','prop'].includes(type)) throw new Error('素材类型无效');
  return request(`/api/comic-screenplays/asset-library?type=${type}`,{suppressGlobalError:true}).then(data=>{if(!Array.isArray(data.entries))throw new Error('素材库返回格式不正确');return data.entries;});
}
export function saveComicAssetLibrarySource(projectId, input, { signal } = {}, request = apiRequest) {
  return request(comicWorkflowPath(projectId,'asset-library'),{method:'POST',body:JSON.stringify(input),signal,suppressGlobalError:true}).then(data=>{if(!data.entry?.id)throw new Error('保存普通库未返回素材编号');return data.entry;});
}
export function stageComicAssetLibraryEntry(projectId, entryId, input, { signal } = {}, request = apiRequest) {
  if (!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(String(entryId||''))) throw new Error('素材编号无效');
  return request(comicWorkflowPath(projectId,`asset-library/${entryId}/stage`),{method:'POST',body:JSON.stringify(input),signal,suppressGlobalError:true}).then(data=>{if(!data.item?.id)throw new Error('普通库图片暂存未返回媒体编号');return data.item;});
}

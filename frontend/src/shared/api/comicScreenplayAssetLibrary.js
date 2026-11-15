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

function libraryEntryPath(id) {
  if(!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(String(id||'')))throw new Error('素材编号无效');
  return `/api/comic-screenplays/asset-library/${id}`;
}
function metadata(input) { return Object.fromEntries(['name','description','category','prompt','voice'].map(key=>[key,String(input[key]||'')])); }
function publicEntry(data) { if(!data?.entry?.id)throw new Error('普通库未返回素材编号，请重新读取核对');return data.entry; }
export function uploadComicAssetLibraryFile(input,{signal}={},request=apiRequest) {
  if(!['character','scene','prop'].includes(input.type))throw new Error('素材类型无效');
  if(!(input.file instanceof Blob)||!input.file.name)throw new Error('请选择图片文件');
  if(input.file.size>40*1024*1024)throw new Error('图片不能超过40MB');
  if(!input.file.type.startsWith('image/'))throw new Error('请选择图片文件');
  const values=metadata(input);if(!values.name.trim()){if(input.type==='character')throw new Error('请输入角色名称');values.name=input.file.name;}
  if(!input.requestKey)throw new Error('上传请求编号缺失');
  const form=new FormData();form.append('file',input.file);form.append('type',input.type);form.append('requestKey',input.requestKey);for(const[key,value]of Object.entries(values))form.append(key,value);
  return request('/api/comic-screenplays/asset-library',{method:'POST',body:form,signal,suppressGlobalError:true}).then(publicEntry);
}
export function updateComicAssetLibraryEntry(id,input,{signal}={},request=apiRequest) {
  const path=libraryEntryPath(id);
  if(!Number.isInteger(input.expectedRevision)||input.expectedRevision<1)throw new Error('素材版本无效，请重新读取');
  if(!String(input.name||'').trim())throw new Error('请输入素材名称');
  return request(path,{method:'PATCH',body:JSON.stringify({expectedRevision:input.expectedRevision,...metadata(input)}),signal,suppressGlobalError:true}).then(publicEntry);
}
export function deleteComicAssetLibraryEntry(id,expectedRevision,{signal}={},request=apiRequest) {
  const path=libraryEntryPath(id);
  if(!Number.isInteger(expectedRevision)||expectedRevision<1)throw new Error('素材版本无效，请重新读取');
  return request(`${path}?expectedRevision=${expectedRevision}`,{method:'DELETE',signal,suppressGlobalError:true}).then(data=>{if(data?.deleted!==true)throw new Error('删除结果未确认，请重新读取核对');return data;});
}

import { apiRequest } from './client.js';
import { comicWorkflowPath } from './comicScreenplayWorkflow.js';
const categoryPath='/api/comic-screenplays/asset-library/categories';
function categoryState(data) {
  if(!Array.isArray(data?.builtin_categories)||!Array.isArray(data?.custom_categories)||![...data.builtin_categories,...data.custom_categories].every(name=>typeof name==='string')||!Number.isSafeInteger(data.revision)||data.revision<0)throw new Error('分类返回格式不正确，请重新读取');
  return data;
}
export function listComicAssetLibraryCategories({signal}={},request=apiRequest) {
  return request(categoryPath,{signal,suppressGlobalError:true}).then(categoryState);
}
export function changeComicAssetLibraryCategory(input,{signal}={},request=apiRequest) {
  if(!['add','delete'].includes(input.action))throw new Error('分类操作无效');
  if(typeof input.name!=='string'||!input.name.trim()||Array.from(input.name).length>191||/[\u0000-\u001f\u007f-\u009f]/.test(input.name))throw new Error('分类名称不能为空，最多191字且不能包含控制字符');
  if(!Number.isSafeInteger(input.expectedRevision)||input.expectedRevision<0)throw new Error('分类版本无效，请重新读取');
  return request(categoryPath,{method:'POST',body:JSON.stringify({action:input.action,name:input.name,expectedRevision:input.expectedRevision}),signal,suppressGlobalError:true}).then(categoryState);
}
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
function metadata(input) {
  const values=Object.fromEntries(['name','description','category','prompt','voice'].map(key=>[key,String(input[key]||'')]));
  if(Object.hasOwn(input,'sora2Info')) {
    if(input.type && input.type!=='character')throw new Error('仅角色支持主体存档');
    if(typeof input.sora2Info!=='string')throw new Error('主体存档必须为文字');
    if(new TextEncoder().encode(input.sora2Info).length>16*1024)throw new Error('主体存档不能超过16KiB');
    values.sora2Info=input.sora2Info;
  }
  if(Object.hasOwn(input,'subjectUploadKey')) {
    if(input.type && input.type!=='character')throw new Error('仅角色支持真人上传引用');
    if(input.subjectUploadKey!==''&&!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(input.subjectUploadKey))throw new Error('真人上传请求编号无效');
    values.subjectUploadKey=input.subjectUploadKey;
  }
  return values;
}
function publicEntry(data) { if(!data?.entry?.id)throw new Error('普通库未返回素材编号，请重新读取核对');return data.entry; }
function libraryImageForm(input,requireImage=true) {
  if(!['character','scene','prop'].includes(input.type))throw new Error('素材类型无效');
  if(requireImage||input.file){if(!(input.file instanceof Blob)||!input.file.name)throw new Error('请选择图片文件');
  if(input.file.size>40*1024*1024)throw new Error('图片不能超过40MB');
  if(!input.file.type.startsWith('image/'))throw new Error('请选择图片文件');}
  if(input.audioFile){if(!(input.audioFile instanceof Blob)||!input.audioFile.name||input.type!=='character')throw new Error('请选择角色参考音频');if(input.audioFile.size===0||input.audioFile.size>32*1024*1024)throw new Error('参考音频不能为空且最大32MB');}
  const values=metadata(input);if(!values.name.trim()){if(input.type==='character')throw new Error('请输入角色名称');values.name=input.file.name;}
  if(!input.requestKey)throw new Error('上传请求编号缺失');
  const form=new FormData();if(input.file)form.append('file',input.file);if(input.audioFile)form.append('audio',input.audioFile);form.append('type',input.type);form.append('requestKey',input.requestKey);for(const[key,value]of Object.entries(values))form.append(key,value);
  return form;
}
export function uploadComicAssetLibraryFile(input,{signal}={},request=apiRequest) {
  return request('/api/comic-screenplays/asset-library',{method:'POST',body:libraryImageForm(input),signal,suppressGlobalError:true}).then(publicEntry);
}
export function replaceComicAssetLibraryImage(id,input,{signal}={},request=apiRequest) {
  const path=libraryEntryPath(id);
  if(!Number.isInteger(input.expectedRevision)||input.expectedRevision<1)throw new Error('素材版本无效，请重新读取');
  return request(`${path}/image?expectedRevision=${input.expectedRevision}`,{method:'PUT',body:libraryImageForm(input),signal,suppressGlobalError:true}).then(publicEntry);
}
export function updateComicAssetLibraryAudio(id,input,{signal}={},request=apiRequest) {
  const path=libraryEntryPath(id);
  if(!Number.isInteger(input.expectedRevision)||input.expectedRevision<1)throw new Error('素材版本无效，请重新读取');
  if(!(input.audioFile instanceof Blob))throw new Error('请选择参考音频文件');
  return request(`${path}/audio?expectedRevision=${input.expectedRevision}`,{method:'PUT',body:libraryImageForm(input,false),signal,suppressGlobalError:true}).then(publicEntry);
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

const subjectPath='/api/comic-screenplays/asset-library/subject-uploads';
function subjectKey(key){if(!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(String(key||'')))throw new Error('真人上传请求编号无效');return key;}
function subjectOperation(data){
 const op=data?.operation;
 if(!op?.requestKey||!['running','paused','needs_review','failed','succeeded'].includes(op.status))throw new Error('上传结果格式不正确，请核对本次上传');
 if(op.status==='succeeded'&&(!op.result?.assetId||typeof op.result.assetId!=='string'))throw new Error('上传未返回主体编号，请核对本次上传');
 return op;
}
export function getComicSubjectUploadCapability({signal}={},request=apiRequest){
 return request('/api/comic-screenplays/asset-library/subject-upload-capability',{signal,suppressGlobalError:true}).then(data=>{
 if(typeof data?.available!=='boolean'||typeof data.label!=='string')throw new Error('无法确认真人素材上传配置');return data;});
}
export function createComicSubjectUpload(input,{signal}={},request=apiRequest){
 subjectKey(input.requestKey);if(!String(input.name||'').trim())throw new Error('请输入角色名称');
 const form=new FormData();for(const key of ['requestKey','name','description','groupId'])form.append(key,String(input[key]||''));
 if(input.file){if(!(input.file instanceof Blob)||!input.file.type.startsWith('image/')||input.file.size===0||input.file.size>40*1024*1024)throw new Error('请选择不超过40MB的图片');form.append('file',input.file);}
 else {if(!input.entryId||!Number.isInteger(input.expectedRevision)||input.expectedRevision<1)throw new Error('请选择角色图片');form.append('entryId',input.entryId);form.append('expectedRevision',String(input.expectedRevision));}
 return request(subjectPath,{method:'POST',body:form,signal,timeoutMs:190000,suppressGlobalError:true}).then(subjectOperation);
}
export function getComicSubjectUpload(key,{signal}={},request=apiRequest){return request(subjectPath+'/'+subjectKey(key)+'/status',{signal,suppressGlobalError:true}).then(subjectOperation);}

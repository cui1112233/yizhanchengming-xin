import { apiRequest, LONG_RUNNING_API_TIMEOUT_MS } from './client.js';

function identifier(value,label='项目') {
  const id=String(value||'');
  if(!/^[A-Za-z0-9][A-Za-z0-9_-]{0,95}$/.test(id))throw new Error(`${label}无效`);
  return id;
}
export function comicWorkflowPath(projectId,suffix='') { return `/api/comic-screenplays/projects/${identifier(projectId)}${suffix?'/'+suffix:''}`; }
const read=(id,path,request)=>request(comicWorkflowPath(id,path));
const write=(id,path,body,request)=>request(comicWorkflowPath(id,path),{method:'POST',body:JSON.stringify(body||{}),timeoutMs:LONG_RUNNING_API_TIMEOUT_MS});
const put=(id,path,body,request)=>request(comicWorkflowPath(id,path),{method:'PUT',body:JSON.stringify(body||{})});
export const getComicProduction=(id,request=apiRequest)=>read(id,'production/status',request);
export const startComicProduction=(id,request=apiRequest)=>write(id,'production/start',{},request);
export function rebuildComicProduction(id,input,request=apiRequest){
  if(input?.confirmed!==true)return Promise.reject(new Error('请先确认重建变更镜头可能产生新的供应商费用'));
  return write(id,'production/start',{rebuild:true},request);
}
export function scheduleComicProduction(id,input,request=apiRequest){
  if(input?.confirmed!==true)return Promise.reject(new Error('请先确认服务器到点自动制作'));
  const at=new Date(input?.run_at||'');if(!Number.isFinite(at.getTime())||at.getTime()<=Date.now())return Promise.reject(new Error('请选择未来的定时制作时间'));
  return write(id,'production/schedule',{run_at:at.toISOString(),...(input.rebuild===true?{rebuild:true}:{})},request);
}
export const cancelComicProductionSchedule=(id,request=apiRequest)=>write(id,'production/cancel-schedule',{},request);
export const pauseComicProduction=(id,request=apiRequest)=>write(id,'production/pause',{},request);
export const getComicChapters=(id,request=apiRequest)=>read(id,'chapters',request);
export const saveComicChapters=(id,input,request=apiRequest)=>put(id,'chapters',input,request);
export const planComicChapters=(id,input,request=apiRequest)=>write(id,'chapters/plan',input,request);
export const generateComicChapter=(id,chapterId,input,request=apiRequest)=>write(id,`chapters/${chapterIdentifier(chapterId)}/script`,input,request);
export const getComicAssistant=(id,request=apiRequest)=>read(id,'assistant/chat',request);
export const sendComicAssistant=(id,input,request=apiRequest)=>write(id,'assistant/chat',input,request);
export function applyComicProposal(id,input,request=apiRequest){
  if(input?.confirmed!==true)return Promise.reject(new Error('请先确认应用这些修改'));
  return write(id,'assistant/apply',input,request);
}
export const getComicDisplay=(id,request=apiRequest)=>read(id,'display',request);
export const saveComicDisplay=(id,input,request=apiRequest)=>put(id,'display',input,request);
export const listArchivedComics=(request=apiRequest)=>request('/api/comic-screenplays/archived');
export const listSharedComics=(request=apiRequest)=>request('/api/comic-screenplays/shared');
export const archiveComic=(id,input,request=apiRequest)=>write(id,'archive',input,request);
export const restoreComic=(id,input,request=apiRequest)=>write(id,'restore',input,request);
export const getComicTeam=(id,request=apiRequest)=>read(id,'team',request);
export const saveComicTeam=(id,input,request=apiRequest)=>put(id,'team',input,request);
export function batchEditComicShots(id,engine,episode,input,request=apiRequest){
  if(!Number.isInteger(Number(episode))||Number(episode)<1)return Promise.reject(new Error('集数无效'));
  return write(id,`storyboards/${identifier(engine,'引擎')}/${Number(episode)}/shots/batch`,input,request);
}
export const getComicMedia=(id,request=apiRequest)=>read(id,'media',request);
export const actComicMedia=(id,input,request=apiRequest)=>write(id,'media/actions',input,request);
export function uploadComicMedia(id,file,fields,request=apiRequest){
  if(!(file instanceof Blob))return Promise.reject(new Error('请选择文件'));
  const body=new FormData();body.append('file',file,file.name||'media');
  Object.entries(fields||{}).forEach(([key,value])=>{if(value!==undefined&&value!==null)body.append(key,String(value));});
  return request(comicWorkflowPath(id,'media/upload'),{method:'POST',body});
}

export const downloadComicMediaPackage=(id,format='media',request=apiRequest)=>request(comicWorkflowPath(id,'media/package')+(format==='jianying'?'?format=jianying':''),{responseType:'blob'});

function chapterIdentifier(value){const id=String(value||'');if(!id||id.length>96||/[/\\?#\x00-\x20]/.test(id)||id==='.'||id==='..')throw new Error('章节无效');return encodeURIComponent(id);}

export function importComicMediaPackage(id,file,expectedRevision,request=apiRequest){
  if(!(file instanceof Blob))return Promise.reject(new Error('请选择本系统导出的ZIP资料包'));
  const body=new FormData();body.append('file',file,file.name||'media.zip');body.append('expectedRevision',String(expectedRevision));
  return request(comicWorkflowPath(id,'media/import'),{method:'POST',body,timeoutMs:LONG_RUNNING_API_TIMEOUT_MS});
}
export const listComicMediaVoices=(id,request=apiRequest)=>read(id,'media/voices',request);
export const downloadComicMediaResult=(id,jobId,request=apiRequest)=>request(comicWorkflowPath(id,`media/exports/${identifier(jobId,'导出任务')}`),{responseType:'blob'});


export const getComicPositioning = (id, request = apiRequest) => read(id, 'positioning', request);
export const saveComicPositioning = (id, input, request = apiRequest) => put(id, 'positioning', input, request);
export function generateComicPositioningImage(id, itemId, input, request = apiRequest) {
  const { signal, ...body } = input || {};
  return request(comicWorkflowPath(id, `positioning/${identifier(itemId, '站位条目')}/image`), { method: 'POST', body: JSON.stringify(body), signal, timeoutMs: 0, suppressGlobalError: true });
}
export function analyzeComicPositioning(id, input, signal, request = apiRequest) {
  return request(comicWorkflowPath(id, 'positioning/analyze'), { method: 'POST', body: JSON.stringify(input), timeoutMs: 0, signal, suppressGlobalError: true });
}

export function bindComicPositioningImage(id, itemId, input, request = apiRequest) {
  return request(comicWorkflowPath(id, `positioning/${identifier(itemId, '站位条目')}/image/upload`), { method: 'POST', body: JSON.stringify(input) });
}
export function editComicPositioningImage(id, itemId, input, request = apiRequest) {
  const { signal, ...body } = input || {};
  return request(comicWorkflowPath(id, `positioning/${identifier(itemId, '站位条目')}/image/edit`), { method: 'POST', body: JSON.stringify(body), signal, timeoutMs: 0, suppressGlobalError: true });
}

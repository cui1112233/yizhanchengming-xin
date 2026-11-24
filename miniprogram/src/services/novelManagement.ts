import {novelID,type NovelRequest} from './novelCreation';
import type {NovelRecord} from './novelCreationInputs';
export type NovelAutomation='schedules'|'batches';
export type NovelResource='materials'|'notes'|'documents'|'styles';
const required=(value:any,label:string)=>{const text=String(value??'').trim();if(!text)throw Object.assign(Error(`请填写${label}`),{beforeRequest:true});return text;};
const integer=(n:any,min:number,max:number,label:string)=>{const value=Number(n);if(!Number.isInteger(value)||value<min||value>max)throw Object.assign(Error(`${label}应为${min}至${max}的整数`),{beforeRequest:true});return value;};
export function findNovelAutomationReceipt(kind:NovelAutomation,key:string,rows:NovelRecord[]){return rows.find(row=>row.creation?.idempotencyKey===(kind==='schedules'?`scheduler:${key}`:key));}
export function createNovelManagement(request:NovelRequest){
 const call=(project:string,suffix:string,data?:NovelRecord,method='POST')=>request(`/api/novel-creation/projects/${novelID(project)}${suffix}`,data===undefined?{}:{method,data,timeout:180000});
 return {
  call,
  async list(project:string,kind:NovelResource|NovelAutomation|'sessions'){const r=await call(project,`/${kind}`);if(!Array.isArray(r?.[kind]))throw Error('列表读取失败，请重试');return r[kind] as NovelRecord[];},
  saveResource(project:string,kind:'materials'|'notes',value:NovelRecord){const data={title:required(value.title,'名称'),content:required(value.content,'正文')};if(value.id&&kind!=='notes')throw Error('资料不支持直接改写，可新建后替换引用');return call(project,`/${kind}${value.id?`/${novelID(value.id)}`:''}`,data,value.id?'PUT':'POST');},
  removeResource:(project:string,kind:NovelResource,value:string)=>call(project,`/${kind}/${novelID(value)}`,{},'DELETE'),
  pasteDocument:(project:string,title:string,content:string)=>call(project,'/documents/paste',{title:required(title,'文档名称'),content:required(content,'文档正文')}),
  selectDocument:(project:string,document:string,version:number,chapterIds:string[])=>call(project,`/documents/${novelID(document)}/selection`,{expectedVersion:version,chapterIds}),
  editDocument:(project:string,document:string,version:number,chapters:NovelRecord[])=>call(project,`/documents/${novelID(document)}/chapters`,{expectedVersion:version,chapters},'PUT'),
  retryDocument:(project:string,document:string)=>call(project,`/documents/${novelID(document)}/retry`,{}),
  createStyle:(project:string,documentId:string,title:string,description:string)=>call(project,'/styles',{documentId:novelID(documentId),title:required(title,'文风名称'),description}),
  maintenance:(project:string,kind:'continuity_refresh'|'style_analysis',sessionId:string,key:string,styleId='')=>call(project,'/maintenance',{kind,sessionId:required(sessionId,'创作会话'),idempotencyKey:required(key,'原请求编号'),...(styleId?{styleId}:{})}),
  async startSchedule(project:string,values:NovelRecord){return call(project,'/scheduler/start',{sessionId:required(values.sessionId,'创作会话'),modelId:required(values.modelId,'文本模型'),intervalSeconds:integer(values.minutes,1,1440,'写作间隔')*60,dailyChapterCap:integer(values.cap,1,100,'每日上限'),maxConsecutiveFailures:integer(values.failures,1,10,'连续失败上限'),idempotencyKey:required(values.key,'原请求编号')});},
  async startBatch(project:string,values:NovelRecord){return call(project,'/batches',{sessionId:required(values.sessionId,'创作会话'),modelId:required(values.modelId,'文本模型'),totalChapters:integer(values.total,1,50,'总章数'),prompt:required(values.prompt,'写作要求'),idempotencyKey:required(values.key,'原请求编号'),maxConsecutiveFailures:integer(values.failures??3,1,10,'连续失败上限'),thinkingLevel:values.thinkingLevel||'standard',knowledgeIds:values.knowledgeIds||[],noteIds:values.noteIds||[],styleProfileId:values.styleProfileId||'',modeInputs:values.modeInputs||{}});},
  async control(project:string,kind:NovelAutomation,id:string,action:'pause'|'resume'|'stop'|'run'){if(!['pause','resume','stop',...(kind==='schedules'?['run']:[])].includes(action))throw Error('不支持的任务操作');return call(project,`/${kind}/${novelID(id)}/${action}`,{});},
  async logs(project:string,kind:NovelAutomation,id:string){const r=await call(project,`/${kind}/${novelID(id)}/logs`);if(!Array.isArray(r?.rounds))throw Error('执行日志读取失败');return r.rounds as NovelRecord[];},
  rename:(project:string,title:string)=>call(project,'',{title:required(title,'作品名称')},'PATCH'),
  retain:(project:string,restore=false)=>call(project,restore?'/restore':'/archive',{}),
  manageSession:(project:string,session:string,action:string,title:string)=>call(project,`/sessions/${novelID(session)}`,{action,title},'PATCH'),
 };
}

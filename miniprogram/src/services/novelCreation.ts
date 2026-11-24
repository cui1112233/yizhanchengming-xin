import type { NovelRecord,NovelMode } from './novelCreationInputs';
export type NovelRequest=(path:string,options?:Record<string,unknown>)=>Promise<any>;
export const novelID=(id:string)=>{if(!/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(id))throw Error('作品或会话编号无效');return encodeURIComponent(id);};
const array=(value:any,key:string,label:string):NovelRecord[]=>{if(!Array.isArray(value?.[key]))throw Error(`${label}返回格式不正确，请重试`);return value[key];};
export function createNovelClient(request:NovelRequest){
  const call=(path:string,data?:any,method='POST')=>request(`/api/novel-creation${path}`,data===undefined?{}:{method,data,timeout:180000});
  return {
    call,
    async projects(){return array(await call('/projects'),'projects','作品列表');},
    async studios(){return array(await call('/studio-sessions'),'sessions','讨论列表');},
    async sessions(project:string){return array(await call(`/projects/${novelID(project)}/sessions`),'sessions','创作会话');},
    async chapters(project:string){return array(await call(`/projects/${novelID(project)}/chapters`),'chapters','正式章节');},
    async styles(){return array(await call('/style-profiles'),'styles','文风目录');},
    async memos(){return array(await call('/long-imitation/memos'),'memos','备忘录目录');},
    async createStudio(mode:NovelMode,key:string){return call('/studio-sessions',{mode,idempotencyKey:key});},
    async studio(id:string){return call(`/studio-sessions/${novelID(id)}`);},
    async sendStudio(id:string,input:NovelRecord){return call(`/studio-sessions/${novelID(id)}/messages`,input);},
    async studioAction(id:string,actionId:string,modelId:string,thinkingLevel:string,cancel=false){return call(`/studio-sessions/${novelID(id)}/actions/${cancel?'cancel':'confirm'}`,cancel?{actionId}:{actionId,modelId,thinkingLevel});},
    async snapshot(id:string){return call(`/sessions/${novelID(id)}/snapshot`);},
    async createSession(project:string,mode:string){return call(`/projects/${novelID(project)}/sessions`,{mode,title:'手机创作会话',contextBudget:120000});},
    async send(id:string,input:NovelRecord){return call(`/sessions/${novelID(id)}/messages`,input);},
    async runAction(id:string,action:'confirm'|'retry'|'cancel',data:NovelRecord={}){return call(`/runs/${novelID(id)}/${action}`,data);},
    async saveChapter(project:string,chapter:NovelRecord){return request(`/api/novel-creation/projects/${novelID(project)}/chapters/${novelID(chapter.id)}`,{method:'PUT',data:{expectedVersion:chapter.version,title:chapter.title,content:chapter.content}});},
    async versions(project:string,id:string){return call(`/projects/${novelID(project)}/chapters/${novelID(id)}/versions`);},
    async restoreChapter(project:string,id:string,expectedVersion:number,version:number){return call(`/projects/${novelID(project)}/chapters/${novelID(id)}/restore`,{expectedVersion,version});},
  };
}

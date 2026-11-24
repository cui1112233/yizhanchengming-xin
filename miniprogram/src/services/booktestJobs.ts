import type {FetchProcessInput} from './fetchWorkspace';
export type BooktestJob=Record<string,any>;
type Request=(path:string,options?:any)=>Promise<any>;
const root='/api/novel-fetch-workshop';
const jobID=(id:string)=>{if(!/^[\w-]{1,150}$/.test(id))throw Error('测书作业编号无效');return encodeURIComponent(id);};
export const booktestJobEnded=(job:BooktestJob)=>['done','failed','stopped','cancelled','canceled','completed'].includes(String(job.status||'').toLowerCase());
export function booktestInput(input:FetchProcessInput,modelId:string){
 if(!input.input.trim()||!input.platformId)throw Error('请先填写书单并选择平台');if(!input.selectedVersions.length)throw Error('请至少选择一个文案版本');
 if(!Number.isInteger(input.maxTxt)||input.maxTxt<1||input.maxTxt>200000)throw Error('处理字数应为1至200000的整数');
 if(input.selectedVersions.some(version=>version!=='original')&&!modelId)throw Error('请选择可用文本模型');
 return {inputText:input.input.trim(),platform_id:input.platformId,maxTxt:input.maxTxt,selected_versions:input.selectedVersions,rewrite_methods:input.aiSlotMethods,sensitive_ai_enabled:input.sensitiveAiEnabled,text_model_id:modelId};
}
export function createBooktestJobs(request:Request){return {
 start:(data:any)=>request(`${root}/process/start`,{method:'POST',data,timeout:180000}),
 read:(id:string)=>request(`${root}/process/jobs/${jobID(id)}`),
 latest:()=>request(`${root}/process/jobs/latest?limit=10`),
 stop:(id:string)=>request(`${root}/process/stop`,{method:'POST',data:{job_id:id,pause_auto_queue:false}}),
};}
export const BOOKTEST_METHODS=[{id:'auto',name:'自动轮换'},{id:'high_imitation',name:'高仿文章库'},{id:'opening_instruction',name:'爆款开头词库'},{id:'instruction',name:'批量改文指令库'}];
export function fetchVersionTexts(payload:Record<string,any>){
 const source=Array.isArray(payload.ai_texts)?payload.ai_texts:Array.isArray(payload.aiVersions)?payload.aiVersions:Object.entries(payload.versions||{}).filter(([id])=>/^ai[1-5]$/.test(id)).map(([id,value]:[string,any])=>typeof value==='string'?{id,name:id.toUpperCase(),text:value}:{...value,id,name:id.toUpperCase()});
 return source.flatMap((item:any,index:number)=>{const content=typeof item.text==='string'?item.text:typeof item.content==='string'?item.content:'';return content?[{key:String(item.id||item.version||`ai-${index}`),label:String(item.name||item.label||item.version||`AI ${index+1}`),content}]:[];});
}

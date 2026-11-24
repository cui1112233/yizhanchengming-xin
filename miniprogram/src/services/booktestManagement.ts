import type {apiRequest} from './api';
export type FetchRequest=typeof apiRequest;
export type FetchAccount={username:string;displayName?:string;role?:string;boundTo?:string};
export function fetchAccountOptions(self:FetchAccount,team:FetchAccount[]=[],global:FetchAccount[]=[]){
 const seen=new Set<string>();return [self,...(self.role==='dev'?global:self.role==='manager'?team.filter(a=>a.role==='member'&&a.boundTo===self.username):[])].filter(a=>a.username&&!seen.has(a.username)&&Boolean(seen.add(a.username)));
}
export function scopedFetchRequest(raw:FetchRequest,account:string):FetchRequest{return ((path:string,options:any={})=>raw(path,/^\/api\/(novel-fetch-workshop|novel-fetch|batch-rewrite|web-submit)(?:[/?]|$)/.test(path)&&account?{...options,header:{...options.header,'X-Qiantie-Novel-Fetch-Account':account}}:options)) as FetchRequest;}
export const KNOWLEDGE_KINDS=[['high_imitation','高仿文章库'],['opening_phrases','爆款开头词库'],['rewrite_templates','批量改文指令库']];
export function knowledgeFields(kind:string,collection=''){
 if(kind==='opening_phrases')return [['name','名称'],['text','开头词内容'],['style','适用风格'],['analysis','分析说明']];
 if(collection==='references')return [['name','名称'],['content','参考文章'],['analysis','分析说明']];
 if(collection==='temporary_instructions')return [['name','名称'],['content','临时指令']];
 return [['name','名称'],['system_prompt','系统要求'],['user_prompt_template','改文指令'],['analysis','说明']];
}
export function createBooktestManagement(request:(path:string,options?:any)=>Promise<any>){
 const root='/api/novel-fetch-workshop',id=encodeURIComponent;
 const kind=(v:string)=>{if(!KNOWLEDGE_KINDS.some(row=>row[0]===v))throw Error('资料库类型无效');return v;};
 const checked=async(p:string,o?:any)=>{const r=await request(p,o);if(r?.ok===false||r?.error)throw Error(String(r.error||'操作失败'));return r;};
 return {
 current:()=>checked(`${root}/batches/current`),
 queue:(batch:string,date:string)=>checked(`${root}/realtime/status${date?`?date=${id(date)}`:batch?`?batch_id=${id(batch)}`:''}`),
 automation:(enabled:boolean)=>checked(`${root}/automation/${enabled?'start':'pause'}`,{method:'POST',data:{}}),
 knowledge:(k:string)=>checked(`${root}/knowledge/${kind(k)}`),
 saveKnowledge:(k:string,item:Record<string,any>,configRevision:any)=>{if(configRevision===undefined)throw Error('配置版本未读取，请刷新资料库');return checked(`${root}/knowledge/${kind(k)}`,{method:'POST',data:{item,configRevision}});},
 removeKnowledge:(k:string,item:string,revision:any)=>{if(revision===undefined)throw Error('配置版本未读取，请刷新资料库');return checked(`${root}/knowledge/${kind(k)}/${id(item)}?configRevision=${id(String(revision))}`,{method:'DELETE'});},
 optimize:(k:string,item:string,revision:any)=>checked(`${root}/knowledge/${kind(k)}/${id(item)}/optimize`,{method:'POST',data:{configRevision:revision},timeout:180000}),
 };
}

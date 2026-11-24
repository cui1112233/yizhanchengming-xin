export type CmRow=Record<string,any>;
type Request=(path:string,options?:any)=>Promise<any>;
const required=(v:any)=>{const s=String(v??'').trim();if(!s)throw Error('缺少项目、节点或模型编号');return s;};
const path=(id:string)=>`/api/agent/tasks/${encodeURIComponent(required(id))}`;
const list=(r:CmRow,key:string)=>{if(!Array.isArray(r?.[key]))throw Error(`服务端未返回有效${key}`);return r[key] as CmRow[];};
export function createCmClient(request:Request){return{
 projects:async()=>list(await request('/api/agent/tasks'),'tasks'),
 create:async(projectType:string)=>{if(!['comic','video'].includes(projectType))throw Error('请选择作品类型');const r=await request('/api/agent/tasks',{method:'POST',data:{projectType}});if(!r?.task?.id)throw Error('创建结果未知，请刷新项目列表核对');return r.task as CmRow;},
 project:async(id:string)=>{const r=await request(path(id));if(String(r?.task?.id)!==id)throw Error('项目编号不匹配');return r.task as CmRow;},
 canvas:async(id:string)=>{const r=await request(`${path(id)}/canvas`);if(!r?.canvas||!Array.isArray(r.canvas.nodes)||!Array.isArray(r.canvas.edges)||!Number.isFinite(r.canvas.revision))throw Error('画布响应不完整');return r.canvas as CmRow;},
 save:async(id:string,canvas:CmRow)=>{if(!Array.isArray(canvas.nodes)||!Array.isArray(canvas.edges)||!Number.isFinite(canvas.revision))throw Error('画布版本不完整，不能保存');const r=await request(`${path(id)}/canvas`,{method:'PATCH',data:{canvas,revision:canvas.revision}});if(!r?.canvas||!Array.isArray(r.canvas.nodes))throw Error('保存结果未知，请保留草稿并刷新核对');return r.canvas as CmRow;},
 collection:async(id:string,kind:'assets'|'executions'|'generations'|'video-jobs')=>list(await request(`${path(id)}/${kind}`),kind==='generations'?'tasks':kind==='video-jobs'?'jobs':kind),
 models:async(kind:string)=>list(await request(`/api/models?kind=${encodeURIComponent(kind)}`),'models').filter(m=>m.enabled!==false),
 skills:async()=>list(await request('/api/agent/skills'),'skills'),
 chat:(id:string,prompt:string,skillIds:string[]=[])=>request('/api/agent/chat',{method:'POST',data:{taskId:required(id),prompt:required(prompt),context:{page:'CM 画布'},skillIds},timeout:180000}),
 rename:(id:string,title:string)=>request(path(id),{method:'PATCH',data:{title:required(title)}}),
 remove:(id:string)=>request(path(id),{method:'DELETE'}),
 generate:async(id:string,kind:string,data:CmRow)=>{const endpoint=({text:'text-generations',image:'image-generations',audio:'audio-generations',video:'video-jobs'} as CmRow)[kind];if(!endpoint)throw Error('此节点类型不支持媒体生成');const result=await request(`${path(id)}/${endpoint}`,{method:'POST',data,timeout:180000});if(!result?.generation?.id&&!result?.execution?.id&&!result?.job?.id&&!result?.task?.id)throw Error('生成提交结果未确认，请刷新任务核对，勿重复提交');return result;},
 action:(id:string,generation:string,action:'retry'|'cancel'|'confirm'|'activate',video=false,data:CmRow={})=>request(`${path(id)}/${video?'video-jobs':'generations'}/${encodeURIComponent(required(generation))}/${action}`,{method:'POST',data}),
};}
export function patchCmNode(node:CmRow,edit:CmRow):CmRow{return{...node,title:String(edit.title??node.title),version:Number(node.version||1)+1,config:{...node.config,...(edit.config||{}),...(edit.prompt!==undefined?{prompt:edit.prompt}:{}),...(edit.sourceText!==undefined?{sourceText:edit.sourceText}:{})}};}
export function cmMediaPath(id:string,asset:CmRow){const expected=`${path(id)}/assets/${encodeURIComponent(required(asset.id||asset.assetId))}/content`;const url=String(asset.url||'');const projection=new RegExp(`^${path(id).replace(/[.*+?^${}()|[\]\\]/g,'\\$&')}/batch-factory/nodes/[^/?#]+/media$`);if(url&&url!==expected&&!projection.test(url))throw Error('素材地址与当前项目不匹配');return url||expected;}
export function buildCmGeneration(node:CmRow,model:CmRow,key:string){
 if(!node.id||!['text','image','video','audio'].includes(node.type)||model.kind!==node.type||model.enabled===false||!model.id)throw Error('请选择与节点类型对应的可用模型');
 if(node.config?.source==='batch_factory')throw Error('此节点绑定批量作品，请使用对应作品工作区执行');
 const c=node.config||{},p={...(c.parameters||{})},kind=node.type;
 const defaults=({text:{operation:'generate',length:'standard',creativity:'balanced'},image:{ratio:'16:9',quality:'standard',resolution:'2K',count:1},video:{ratio:'16:9',resolution:'720P',durationSeconds:5,count:1,generateAudio:false},audio:{operation:'speech',speed:1,pitch:0,emotion:'neutral',format:'mp3'}} as CmRow)[kind];
 const parameters={...defaults,...p};for(const k of Object.keys(defaults))if(c[k]!==undefined&&p[k]===undefined)parameters[k]=c[k];
 const prompt=[...(c.edgePromptSources||[]).map((v:CmRow)=>String(v.text||'').trim()),String(c.prompt||'').trim()].filter(Boolean).join('\n');if(!prompt)throw Error('请先保存节点的生成要求');
 const referenceAssetIds=[...new Set((c.referenceAssets||[]).map((a:CmRow)=>String(a.id||a.assetId||'')).filter(Boolean))];
 const common={nodeId:required(node.id),nodeVersion:Number(node.version),provider:String(model.provider||c.provider||''),modelId:String(model.id),prompt,advanced:{...(c.advanced||{})},idempotencyKey:required(key)};
 if(!Number.isInteger(common.nodeVersion)||common.nodeVersion<1)throw Error('节点版本无效，请刷新画布');
 if(kind==='audio')return{...common,operation:parameters.operation,settings:parameters,sourceAssetId:String(node.assetRef?.id||'' )};
 return{...common,mode:String(c.mode|| (kind==='image'?(referenceAssetIds.length?'image_to_image':'text_to_image'):parameters.operation||'')),parameters,referenceAssetIds};
}

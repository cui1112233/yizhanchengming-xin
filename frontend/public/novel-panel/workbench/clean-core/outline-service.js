/* V78 Clean Core Phase 15: Outline Generation Service Facade + Generator Authority integration.
 * Scope: one stable orchestration entry for whole-outline and single-scene regeneration.
 * It coordinates existing Task / Submission / V77 native generator / Response / Writeback / Refresh layers.
 * It orchestrates the production transaction while Phase15 Native Generator Service owns the promoted network request path; V77 remains fallback.
 */
(function installV78OutlineGenerationService(){
  if(globalThis.__V78_OUTLINE_GENERATION_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const txt=v=>String(v??'').trim();
  const arr=v=>Array.isArray(v)?v:[];
  const now=()=>new Date().toISOString();
  let seq=0,enabled=true,currentId='';
  const active=new Map(),history=[];
  const MAX_HISTORY=80;
  const stats={outline_runs:0,regeneration_runs:0,success:0,error:0,native_invocations:0,prevented_duplicate_native_invocations:0,v78_primary_transactions:0,v77_fallback_transactions:0,post_send_protected_errors:0,last_duration_ms:0};
  function makeId(kind){seq+=1;return `v78-${kind}-${Date.now().toString(36)}-${seq.toString(36)}-${Math.random().toString(36).slice(2,7)}`;}
  function characterRevision(){try{return Number(globalThis.__characterCoreV2Api?.state?.()?.character_revision||0);}catch(_){return 0;}}
  function sourceSnapshot(){try{return clone(globalThis.__characterCoreV2Api?.sourceSnapshot?.()||{});}catch(_){return {};}}
  function stateCounts(){
    try{
      const project=typeof globalThis.getProjectData==='function'?(globalThis.getProjectData()||{}):{};
      const s=globalThis.state||globalThis.appState||project;
      return {scene_count:arr(s.scenes||project.scenes).length,shot_count:arr(s.outlineShots||s.outline_shots||project.outlineShots||project.outline_shots).length,segment_count:arr(s.segments||project.segments).length};
    }catch(_){return {scene_count:0,shot_count:0,segment_count:0};}
  }
  function authoritySnapshot(){
    return {
      submission:clone(ns.sceneContext?.submissionAuthoritySnapshot?.()||null),
      writeback:clone(ns.outlineResponse?.authoritySnapshot?.()||null),
      casting:clone(ns.casting?.authoritySnapshot?.()||null),
      generator_contract:clone(ns.generatorContract?.snapshot?.()||null),
      generator_protocol:clone(ns.generatorProtocol?.authoritySnapshot?.()||null),
      generator:clone(ns.generator?.authoritySnapshot?.()||null),
      generator_service:clone(ns.generatorService?.authoritySnapshot?.()||null),
      production:clone(ns.production?.snapshot?.()||null),
    };
  }
  function taskSnapshot(){try{return clone(globalThis.__V77_AI_TASK_MANAGER__?.snapshot?.()||null);}catch(_){return null;}}
  function traceFor(kind){const endpoint=kind==='regeneration'?'/api/regenerate-scene-outline':'/api/outline-scenes';return txt(globalThis.__V77_LAST_TRACE_BY_ENDPOINT__?.[endpoint]);}
  function rowPublic(row){if(!row)return null;const out=clone(row);delete out._native;return out;}
  function pushStage(row,stage,data={}){
    if(!row)return null;
    const item={stage:txt(stage)||'unknown',at:now(),data:clone(data||{})};
    row.stage=item.stage;row.timeline.push(item);row.updated_at=item.at;
    return item;
  }
  function markStage(stage,data={}){
    const mode=txt(data?.mode),sceneId=txt(data?.scene_id);let row=null;
    if(mode==='regeneration'||sceneId){row=[...active.values()].reverse().find(x=>x.kind==='regeneration'&&(!sceneId||txt(x.scene_id)===sceneId))||null;}
    else if(mode==='outline'){row=[...active.values()].reverse().find(x=>x.kind==='outline')||null;}
    if(!row)row=active.get(currentId)||[...active.values()].slice(-1)[0]||null;
    return pushStage(row,stage,data);
  }
  function begin(kind,meta={}){
    const source=sourceSnapshot(),id=makeId(kind),started=Date.now();
    const row={transaction_id:id,kind,status:'running',started_at:new Date(started).toISOString(),updated_at:new Date(started).toISOString(),source_hash:txt(source.source_hash),source_revision:Number(source.source_revision||0),character_revision:characterRevision(),scene_id:txt(meta.scene_id),source_key:txt(meta.source_key),timeline:[],native_invocation_count:0,network_task_before:taskSnapshot(),authority_before:authoritySnapshot(),counts_before:stateCounts(),_native:null};
    active.set(id,row);currentId=id;pushStage(row,'entry',{enabled,source_hash:row.source_hash,source_revision:row.source_revision,character_revision:row.character_revision,scene_id:row.scene_id,source_key:row.source_key});
    return row;
  }
  async function invokeNativeOnce(row,native){
    if(typeof native!=='function')throw new Error('V77原生分镜执行器缺失，Phase15 Facade无法继续。');
    if(row.native_invocation_count>=1){stats.prevented_duplicate_native_invocations+=1;throw new Error('Phase15已阻止同一编排事务重复调用原生生成器。');}
    row.native_invocation_count+=1;stats.native_invocations+=1;pushStage(row,'native_pipeline_start',{native_executor:row.kind==='regeneration'?'v78_phase15_regeneration_compatibility_host':'v78_phase15_outline_compatibility_host'});
    const value=await native();
    pushStage(row,'native_pipeline_return',{result_type:value===null?'null':Array.isArray(value)?'array':typeof value});
    return value;
  }
  function finish(row,status,error=''){
    if(!row)return;
    const ended=Date.now();row.status=status;row.finished_at=new Date(ended).toISOString();row.duration_ms=Math.max(0,ended-Date.parse(row.started_at));row.error=txt(error).slice(0,1000);row.trace_id=traceFor(row.kind);row.network_task_after=taskSnapshot();row.authority_after=authoritySnapshot();row.counts_after=stateCounts();
    pushStage(row,status==='success'?'completed':status==='handled_error'?'handled_error':'failed',{trace_id:row.trace_id,duration_ms:row.duration_ms,counts:row.counts_after,error:row.error});
    active.delete(row.transaction_id);if(currentId===row.transaction_id)currentId='';history.push(rowPublic(row));while(history.length>MAX_HISTORY)history.shift();stats.last_duration_ms=row.duration_ms;if(status==='success')stats.success+=1;else stats.error+=1;
  }
  async function run(kind,meta,native){
    if(!enabled)return native();
    const row=begin(kind,meta);if(kind==='outline')stats.outline_runs+=1;else stats.regeneration_runs+=1;
    try{
      pushStage(row,'preflight',{submission_authority:row.authority_before.submission,writeback_authority:row.authority_before.writeback,task_manager:row.network_task_before?.version||''});
      const result=await invokeNativeOnce(row,native);
      const productionStages=row.timeline.filter(x=>x?.data?.executor&&String(x.data.executor).includes('phase15'));
      const usedV78=productionStages.some(x=>String(x.data.executor).includes('v78_'));
      const usedFallback=productionStages.some(x=>String(x.data.executor).includes('v77_')||String(x.data.executor).includes('fallback'));
      if(usedV78)stats.v78_primary_transactions+=1;if(usedFallback)stats.v77_fallback_transactions+=1;
      const handled=row.timeline.slice().reverse().find(x=>x.stage==='native_error_handled');
      if(handled&&row.timeline.some(x=>x.stage==='native_generator_service_error'&&x.data?.no_retry))stats.post_send_protected_errors+=1;
      finish(row,handled?'handled_error':'success',handled?.data?.message||'');return result;
    }catch(error){finish(row,'error',error?.message||error);throw error;}
  }
  async function runOutline(options={}){
    const native=options.executeNative||globalThis.__V78_OUTLINE_NATIVE_EXECUTORS__?.outline;
    return run('outline',{},native);
  }
  async function runRegeneration(sceneId,options={}){
    let source_key='';try{const s=(globalThis.appState?.scenes||globalThis.state?.scenes||[]).find(x=>txt(x.id)===txt(sceneId));source_key=txt(s?.source_key);}catch(_){}
    const native=options.executeNative||(()=>globalThis.__V78_OUTLINE_NATIVE_EXECUTORS__?.regenerate?.(sceneId));
    return run('regeneration',{scene_id:sceneId,source_key},native);
  }
  function snapshot(){return {service:'outline_generation_facade_v78_phase15',phase:'v78_phase15',enabled,current_transaction:rowPublic(active.get(currentId)),active:[...active.values()].map(rowPublic),history:history.slice(-30),stats:clone(stats),native_generator:'phase15_v78_default_production_mainline_with_v77_realtime_fallback',generator_service:ns.generatorService?.version||'',generator_contract:ns.generatorContract?.version||'',generator_protocol:ns.generatorProtocol?.version||'',generator_adapter:ns.generator?.version||'',hidden_retry:false,submission_service:ns.sceneContext?.version||'',response_adapter:ns.outlineResponse?.version||''};}
  function setEnabled(value){enabled=Boolean(value);return snapshot();}
  function resetHistory(){history.length=0;Object.keys(stats).forEach(k=>{if(typeof stats[k]==='number')stats[k]=0;});return snapshot();}
  const api={version:'outline_generation_facade_v78_phase15',phase:'v78_phase15',runOutline,runRegeneration,markStage,snapshot,setEnabled,resetHistory};
  ns.outlineService=api;globalThis.__V78_OUTLINE_GENERATION_SERVICE__=api;
  // Update executor ownership after CharacterCore has installed its legacy-compatible handlers.
  try{
    const reg=globalThis.__V77_EXECUTOR_REGISTRY__;if(reg?.executors){
      if(reg.executors.outline)Object.assign(reg.executors.outline,{id:'v78-phase15-outline-production-mainline',owner:'V78ProductionMainline',function_name:'outlineService.runOutline → Phase15 Native Generator Service → single-request executor'});
      if(reg.executors.regenerate)Object.assign(reg.executors.regenerate,{id:'v78-phase15-regenerate-production-mainline',owner:'V78ProductionMainline',function_name:'outlineService.runRegeneration → Phase15 Native Generator Service → single-request executor'});
    }
  }catch(_){}
  console.info('[V78 Clean Core] Phase15 outline generation facade ready',{version:api.version,generator_service:ns.generatorService?.version||'not_loaded',generator_authority:ns.generator?.version||'not_loaded'});
})();

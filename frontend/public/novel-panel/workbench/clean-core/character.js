/* V78 Clean Core Phase 2: read-only CharacterCore state/health/relationship service.
 * Production generation remains owned by CharacterCore 2.0. This module only reads,
 * normalizes and diagnoses current data. All public payloads are deep clones.
 */
(function installV78CharacterStateService(){
  if(globalThis.__V78_CHARACTER_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const arr=(v)=>Array.isArray(v)?v:[];
  const obj=(v)=>v&&typeof v==="object"&&!Array.isArray(v)?v:{};
  const txt=(v)=>String(v??"").trim();
  const clone=(v)=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const now=()=>new Date().toISOString();
  function legacy(){return globalThis.__characterCoreV2Api||null;}
  function liveState(){
    const api=legacy();
    try{if(typeof api?.state==="function")return api.state()||{};}catch(_){}
    const root=(typeof state!=="undefined"?state:globalThis.state)||{};
    return root.characterCoreV2||root.character_core_v2||{};
  }
  function stateSnapshot(){return clone(liveState())||{};}
  function people(){return arr(liveState()?.people).map(clone);}
  function slots(options={}){const rows=arr(liveState()?.slots);return rows.filter(row=>options.includeDisabled||!row?.disabled).map(clone);}
  function findSlot(slotId){const id=txt(slotId);return clone(arr(liveState()?.slots).find(row=>txt(row?.slot_id)===id)||null);}
  function findPerson(personId){const id=txt(personId);return clone(arr(liveState()?.people).find(row=>txt(row?.person_id)===id)||null);}
  function relationships(options={}){return arr(liveState()?.relationships).filter(row=>options.includeDisabled||!row?.disabled).map(clone);}
  function aliases(options={}){return arr(liveState()?.aliases).filter(row=>options.includeDisabled||!row?.disabled).map(clone);}
  function sceneCasting(){return clone(obj(liveState()?.scene_casting));}
  function temporaryContinuity(){const api=legacy();try{if(typeof api?.temporaryContinuity==="function")return clone(api.temporaryContinuity()||{});}catch(_){}return clone(obj(liveState()?.temporary_continuity_cache));}
  function temporaryContinuityMirrors(){const api=legacy();try{if(typeof api?.temporaryContinuityMirrors==="function")return clone(api.temporaryContinuityMirrors()||{});}catch(_){}const c=liveState();return {active:clone(obj(c?.temporary_continuity_cache)),character_core_validator:clone(obj(c?.temporary_continuity_cache_character_core_validator)),v78_primary:clone(obj(c?.temporary_continuity_cache_v78_primary)),authority:clone(obj(c?.temporary_continuity_authority_v78))};}
  function stageSignature(ref){const api=legacy(),slot=typeof ref==="string"?arr(liveState()?.slots).find(row=>txt(row?.slot_id)===txt(ref)):ref;try{if(slot&&typeof api?.stageSignature==="function")return txt(api.stageSignature(slot));}catch(_){}const stage=txt(slot?.age?.visual_age_stage||slot?.visual_age_stage||slot?.current_values?.visual_age_stage||"未定阶段");const timeline=txt(slot?.age?.timeline_stage||slot?.current_values?.timeline_stage||"当前时间线");return `${stage}@@${timeline}`;}
  function relationshipConflicts(){const api=legacy();try{if(typeof api?.relationshipConflicts==="function")return clone(api.relationshipConflicts()||[]);}catch(_){}return clone(arr(liveState()?.relationship_conflicts));}
  function slotHealth(slot){const api=legacy();try{if(typeof api?.characterHealth==="function")return clone(api.characterHealth(slot));}catch(_){}const issues=[],warnings=[];if(!txt(slot?.gender)||txt(slot?.gender)==="待确认"||txt(slot?.gender)==="未定")issues.push("性别待确认");const stage=txt(slot?.age?.visual_age_stage||slot?.current_values?.visual_age_stage||slot?.visual_age_stage);if(!stage||/待确认|未定/.test(stage))issues.push("年龄阶段待确认");if(!txt(slot?.appearance))issues.push("外形未生成");else if(slot?.appearance_stage_stale)issues.push("阶段变化后外形待重构");return {level:issues.length?"error":warnings.length?"warn":"ok",issues,warnings,summary:issues.join("；")||warnings.join("；")||"人物资料正常"};}
  function health(){
    const api=legacy();
    try{if(typeof api?.coreParityHealth==="function"){const raw=clone(api.coreParityHealth()||{});const rows=arr(raw.character_health);return {...raw,service:"character_state_service_v78_phase2",generated_at:now(),summary:{total:rows.length,ok:rows.filter(x=>x.level==="ok").length,warn:rows.filter(x=>x.level==="warn").length,error:rows.filter(x=>x.level==="error").length}};}}catch(_){}
    const rows=arr(liveState()?.slots).filter(x=>!x?.disabled).map(slot=>({slot_id:txt(slot?.slot_id),display_name:txt(slot?.display_name||slot?.canonical_name),...slotHealth(slot)}));
    return {service:"character_state_service_v78_phase2",generated_at:now(),character_health:rows,summary:{total:rows.length,ok:rows.filter(x=>x.level==="ok").length,warn:rows.filter(x=>x.level==="warn").length,error:rows.filter(x=>x.level==="error").length}};
  }
  function diagnostics(){
    const c=liveState(),issues=[],personIds=new Set(arr(c?.people).map(x=>txt(x?.person_id)).filter(Boolean)),slotIds=arr(c?.slots).map(x=>txt(x?.slot_id)).filter(Boolean);
    const push=(level,code,message,detail={})=>issues.push({level,code,message,...clone(detail)});
    if(slotIds.length!==new Set(slotIds).size)push("error","DUPLICATE_SLOT_ID","存在重复 slot_id");
    arr(c?.slots).filter(x=>!x?.disabled).forEach(slot=>{const pid=txt(slot?.person_id);if(pid&&!personIds.has(pid))push("error","DANGLING_SLOT_PERSON","人物阶段卡指向不存在的 Person",{slot_id:txt(slot?.slot_id),person_id:pid});});
    arr(c?.relationships).filter(x=>!x?.disabled).forEach(rel=>{for(const [field,label] of [["source_person_id","关系起点"],["target_person_id","关系终点"]]){const pid=txt(rel?.[field]);if(pid&&!personIds.has(pid))push("error","DANGLING_RELATION_PERSON",`${label}指向不存在的 Person`,{relation_id:txt(rel?.relation_id),person_id:pid});}});
    arr(c?.aliases).filter(x=>!x?.disabled).forEach(binding=>{const pid=txt(binding?.target_person_id);if(pid&&!personIds.has(pid))push("error","DANGLING_ALIAS_PERSON","别称绑定指向不存在的 Person",{binding_id:txt(binding?.binding_id),alias:txt(binding?.alias),person_id:pid});});
    const conflicts=relationshipConflicts();conflicts.forEach(conflict=>push("warn",txt(conflict?.type||"RELATION_CONFLICT").toUpperCase(),txt(conflict?.summary||"人物关系/别称存在冲突"),{conflict}));
    const h=health(),hs=h?.summary||{};if(Number(hs.error||0)>0)push("warn","CHARACTER_HEALTH_ERRORS",`${hs.error} 张人物卡存在待处理健康项`);if(Number(hs.warn||0)>0)push("warn","CHARACTER_HEALTH_WARNINGS",`${hs.warn} 张人物卡存在需复核项`);
    const summary={error:issues.filter(x=>x.level==="error").length,warn:issues.filter(x=>x.level==="warn").length,total:issues.length,relationship_conflicts:conflicts.length};
    return {ok:summary.error===0,service:"character_diagnostics_v78_phase2",generated_at:now(),summary,issues,relationship_conflicts:conflicts};
  }
  function snapshot(){return {service:"character_state_service_v78_phase2",phase:"v78_phase2",generated_at:now(),state:stateSnapshot(),health:health(),diagnostics:diagnostics(),temporary_continuity:temporaryContinuity(),temporary_continuity_mirrors:temporaryContinuityMirrors(),scene_casting:sceneCasting()};}
  const service={version:"character_state_service_v78_phase2",phase:"v78_phase2",read_only:true,state:stateSnapshot,people,slots,findSlot,findPerson,relationships,aliases,sceneCasting,temporaryContinuity,temporaryContinuityMirrors,stageSignature,relationshipConflicts,slotHealth,health,diagnostics,snapshot,legacy};
  ns.character=service;globalThis.__V78_CHARACTER_SERVICE__=service;
  console.info("[V78 Clean Core] character state service ready",{version:service.version,read_only:true});
})();

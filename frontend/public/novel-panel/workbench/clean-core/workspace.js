/* V78 Clean Core Phase 1: stable workspace snapshot/store boundary. */
(function installV78WorkspaceStore(){
  if(globalThis.__V78_WORKSPACE_STORE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const SCHEMA_VERSION=40,SCHEMA_NAME="video_prompt_workspace_v40";
  const baseGet=globalThis.getProjectData,baseApply=globalThis.applyProjectData;
  const clone=(v)=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const obj=(v)=>v&&typeof v==="object"&&!Array.isArray(v)?v:{};
  const arr=(v)=>Array.isArray(v)?v:[];
  const txt=(v)=>String(v??"").trim();
  function normalize(input={}){
    const out=clone(obj(input))||{};const from=Number(out.workspace_schema_version||1);
    out.workspace_schema_version=SCHEMA_VERSION;out.workspace_schema=SCHEMA_NAME;out.workspace_migrated_from=from;
    out.clean_core_meta={...(obj(out.clean_core_meta)),phase:"v78_1_reference_entities",transport:"clean_transport_v78_phase1",workspace:"workspace_store_v78_phase1",diagnostics:"clean_diagnostics_v78_phase15",compatibility:"v77_hotfix53"};
    if(!out.scene_memory_v37&&arr(out.scenes).length){const anchors={},by_source_key={};arr(out.scenes).forEach(scene=>{const id=txt(scene.scene_anchor_id||scene.temporary_scene_group);if(id&&scene.scene_memory){anchors[id]=clone(scene.scene_memory);if(scene.source_key)by_source_key[scene.source_key]=id;}});out.scene_memory_v37={version:"scene_memory_v37",anchors,by_source_key,updated_at:new Date().toISOString()};}
    if(!out.reference_bindings_v781||typeof out.reference_bindings_v781!=="object")out.reference_bindings_v781={version:"reference_bindings_v781",source_hash:"",by_source_key:{},scene_anchor_assets:{},temporary_assets:{},prop_entities:{},updated_at:""};
    return out;
  }
  function validate(input={}){
    const data=normalize(input),issues=[];const scenes=arr(data.scenes),slots=arr(data.characterCoreV2?.slots||data.character_core_v2?.slots);
    const keys=scenes.map(s=>txt(s.source_key)).filter(Boolean);if(keys.length&&new Set(keys).size!==keys.length)issues.push({level:"error",code:"DUPLICATE_SOURCE_KEY",message:"存在重复 source_key"});
    const slotIds=slots.map(s=>txt(s.slot_id)).filter(Boolean);if(slotIds.length&&new Set(slotIds).size!==slotIds.length)issues.push({level:"error",code:"DUPLICATE_SLOT_ID",message:"存在重复 slot_id"});
    return {ok:!issues.some(x=>x.level==="error"),schema_version:SCHEMA_VERSION,issues};
  }
  function decorate(input={}){const out=normalize(input);out.workspace_saved_at=new Date().toISOString();out.workspace_capabilities=Array.from(new Set([...(arr(out.workspace_capabilities)),"clean_transport_v78_phase1","workspace_store_v78_phase1","clean_diagnostics_v78_phase15","clean_core_bridge_v78_phase15","scene_context_service_v78_phase8","outline_response_adapter_v78_phase9","outline_generation_facade_v78_phase13","outline_generator_authority_v78_phase13","native_generator_request_contract_v78_phase13","native_generator_protocol_service_v78_phase13","native_generator_service_v78_phase15","reference_entity_resolver_v78_2"]));return out;}
  function snapshot(){if(typeof baseGet!=="function")return decorate({});return decorate(baseGet()||{});}
  function restore(projectId,name,data={}){if(typeof baseApply!=="function")throw new Error("Clean Core Workspace 未找到兼容 applyProjectData");const normalized=normalize(data);const result=baseApply(projectId,name,normalized);try{globalThis.__V77_DIRTY_RENDER__?.invalidate?.("*");}catch(_){ }globalThis.__V78_WORKSPACE_MIGRATION__={from:Number(data?.workspace_schema_version||1),to:SCHEMA_VERSION,at:new Date().toISOString(),compatible:true};return result;}
  const store={version:"workspace_store_v78_phase1",schema_version:SCHEMA_VERSION,schema:SCHEMA_NAME,normalize,decorate,validate,snapshot,restore,migrationReport:()=>clone(globalThis.__V78_WORKSPACE_MIGRATION__||{})};
  ns.workspace=store;globalThis.__V78_WORKSPACE_STORE__=store;globalThis.__V77_WORKSPACE_SCHEMA__=store;
  if(typeof baseGet==="function"){const wrapped=function v78WorkspaceSnapshotCompat(...args){return decorate(baseGet.apply(this,args)||{});};wrapped.__v78WorkspaceStore=true;wrapped.__v78Base=baseGet;globalThis.getProjectData=wrapped;try{getProjectData=wrapped;}catch(_){}}
  if(typeof baseApply==="function"){const wrapped=function v78WorkspaceRestoreCompat(projectId,name,data={}){return restore(projectId,name,data);};wrapped.__v78WorkspaceStore=true;wrapped.__v78Base=baseApply;globalThis.applyProjectData=wrapped;try{applyProjectData=wrapped;}catch(_){}}
  console.info("[V78 Clean Core] workspace store ready",{schema:SCHEMA_VERSION});
})();

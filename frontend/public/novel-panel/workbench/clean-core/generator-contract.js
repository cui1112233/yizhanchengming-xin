/* V78 Clean Core Phase 13: Native Generator Request Contract Model.
 * Defines the versioned request schema boundary without changing business content.
 * Contract metadata and revisions are INTERNAL ONLY and never injected into the API payload.
 * The contract is immutable after creation; toApiPayload() always returns a detached clone.
 */
(function installV78GeneratorContractService(){
  if(globalThis.__V78_GENERATOR_CONTRACT_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const CONTRACT_ID='video_prompt.outline_generator_request';
  const CONTRACT_VERSION='1.0.0';
  const PHASE='v78_phase13';
  const txt=v=>String(v??'').trim();
  const arr=v=>Array.isArray(v)?v:[];
  function clone(v){
    if(v===null||typeof v!=='object')return v;
    if(Array.isArray(v))return v.map(clone);
    const out={};for(const k of Object.keys(v))out[k]=clone(v[k]);return out;
  }
  function stable(v){
    if(v===null||typeof v!=='object')return v;
    if(Array.isArray(v))return v.map(stable);
    const out={};Object.keys(v).sort().forEach(k=>{if(typeof v[k]!=='function'&&typeof v[k]!=='symbol'&&v[k]!==undefined)out[k]=stable(v[k]);});return out;
  }
  function deepFreeze(v){if(!v||typeof v!=='object'||Object.isFrozen(v))return v;Object.freeze(v);Object.keys(v).forEach(k=>deepFreeze(v[k]));return v;}
  function hash(text=''){let h=2166136261;const s=String(text);for(let i=0;i<s.length;i++){h^=s.charCodeAt(i);h=Math.imul(h,16777619);}return (h>>>0).toString(16).padStart(8,'0');}
  function typeOf(v){return Array.isArray(v)?'array':v===null?'null':typeof v;}
  const F=(required,type,visibility='ai',mutability='readonly_after_freeze')=>({required,types:Array.isArray(type)?type:[type],visibility,mutability,send:true});
  const COMMON={
    novel_text:F(true,'string'),generation_rules:F(true,'string'),characters:F(true,'array'),
    temporary_visible_cast:F(false,'array'),must_cover_details:F(false,'string'),shot_rhythm_requirements:F(false,'string'),
    scene_memory:F(false,'object'),scene_context:F(false,'object'),manual_added_slot_ids:F(false,'array'),manual_excluded_slot_ids:F(false,'array')
  };
  const CONTRACTS={
    outline:{...COMMON,genre:F(true,'string'),trailer_style:F(true,'string'),camera:F(true,'string'),audio_total_seconds:F(false,['number','string'])},
    regeneration:{...COMMON,scene:F(true,'object'),outline_shots:F(true,'array')}
  };
  const INTERNAL_ONLY=new Set(['contract_id','contract_version','contract_kind','source_hash','source_revision','character_revision','expected_sources','trace_id','task_id','authority','diagnostics','internal_meta','__internal','_internal']);
  const READONLY_PATHS={
    outline:['characters[*].slot_id','characters[*].person_id','temporary_visible_cast[*].continuity_id'],
    regeneration:['scene.source_key','scene.source_index','characters[*].slot_id','characters[*].person_id','outline_shots[*].source_key','temporary_visible_cast[*].continuity_id']
  };
  function descriptor(kind,key){
    const d=CONTRACTS[kind]?.[key];
    return d?{...d}:{required:false,types:['any'],visibility:'ai_legacy_passthrough',mutability:'readonly_after_freeze',send:true,legacy:true};
  }
  function identityProjection(kind,payload={}){
    const chars=arr(payload.characters).map(x=>({slot_id:txt(x?.slot_id||x?.id),person_id:txt(x?.person_id),name:txt(x?.name)}));
    const temps=arr(payload.temporary_visible_cast).map(x=>({continuity_id:txt(x?.continuity_id),entity_id:txt(x?.entity_id),label:txt(x?.label)}));
    const scene=payload.scene||{};
    return {kind,scene_source_key:txt(scene.source_key),scene_source_index:Number(scene.source_index||0),characters:chars,temporary:temps,outline_source_keys:arr(payload.outline_shots).map(x=>txt(x?.source_key)).filter(Boolean)};
  }
  function internalMeta(meta={}){
    return {
      source_hash:txt(meta.source_hash),source_revision:Number(meta.source_revision||0),character_revision:Number(meta.character_revision||0),
      source_key:txt(meta.source_key),source_index:Number(meta.source_index||0),expected_sources:clone(arr(meta.expected_sources)),
    };
  }
  function manifest(kind,payload={}){
    const keys=[...new Set([...Object.keys(CONTRACTS[kind]||{}),...Object.keys(payload||{})])].sort();
    return keys.map(name=>{const d=descriptor(kind,name),present=Object.prototype.hasOwnProperty.call(payload,name);return {name,present,required:Boolean(d.required),types:[...d.types],actual_type:present?typeOf(payload[name]):'missing',visibility:d.visibility,mutability:d.mutability,send:Boolean(d.send),legacy:Boolean(d.legacy)};});
  }
  function create(kind,payload={},meta={}){
    if(!CONTRACTS[kind])throw new Error(`未知Generator Contract kind: ${kind}`);
    const apiPayload=clone(payload||{}),internal=internalMeta(meta),identity=identityProjection(kind,apiPayload);
    const model={contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,contract_kind:kind,phase:PHASE,created_at:new Date().toISOString(),api_payload:apiPayload,internal,manifest:manifest(kind,apiPayload),readonly_paths:[...(READONLY_PATHS[kind]||[])],identity_fingerprint:hash(JSON.stringify(stable(identity))),api_fingerprint:hash(JSON.stringify(stable(apiPayload)))};
    return deepFreeze(model);
  }
  function validateModel(model){
    const errors=[],warnings=[];
    if(!model||typeof model!=='object')return {ok:false,errors:['Contract model不存在'],warnings:[]};
    const kind=txt(model.contract_kind),payload=model.api_payload||{},spec=CONTRACTS[kind];
    if(model.contract_id!==CONTRACT_ID)errors.push('contract_id不匹配');
    if(model.contract_version!==CONTRACT_VERSION)errors.push('contract_version不匹配');
    if(!spec)errors.push(`未知contract_kind ${kind}`);
    if(spec){
      for(const [key,d] of Object.entries(spec)){
        const present=Object.prototype.hasOwnProperty.call(payload,key),v=payload[key];
        if(d.required&&(!present||v===undefined||v===null||(d.types.includes('string')&&!txt(v))))errors.push(`缺少必需字段 ${key}`);
        if(present&&v!==undefined&&v!==null&&!d.types.includes('any')&&!d.types.includes(typeOf(v)))errors.push(`${key}类型应为${d.types.join('|')}，实际${typeOf(v)}`);
      }
      Object.keys(payload).forEach(key=>{if(!spec[key]&&!INTERNAL_ONLY.has(key))warnings.push(`兼容透传字段 ${key} 尚未登记到Contract v${CONTRACT_VERSION}`);});
    }
    for(const key of INTERNAL_ONLY){if(Object.prototype.hasOwnProperty.call(payload,key))errors.push(`内部字段 ${key} 禁止进入API Payload`);}
    if(kind==='regeneration'){
      const expected=txt(model.internal?.source_key),actual=txt(payload?.scene?.source_key);
      if(expected&&actual&&expected!==actual)errors.push(`只读身份 scene.source_key 被改变：${actual} != ${expected}`);
    }
    const frozen=Object.isFrozen(model)&&Object.isFrozen(model.api_payload)&&Object.isFrozen(model.internal);if(!frozen)errors.push('Contract model未完全冻结');
    return {ok:errors.length===0,errors,warnings,contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,kind,identity_fingerprint:model.identity_fingerprint,api_fingerprint:model.api_fingerprint,unknown_fields:warnings.filter(x=>x.startsWith('兼容透传字段')).length};
  }
  function toApiPayload(model){return clone(model?.api_payload||{});}
  function toAiVisiblePayload(model){
    const payload=model?.api_payload||{},kind=txt(model?.contract_kind),out={};
    Object.keys(payload).forEach(key=>{const d=descriptor(kind,key);if(d.send&&(d.visibility==='ai'||d.visibility==='ai_legacy_passthrough'))out[key]=clone(payload[key]);});
    return out;
  }
  function compareModels(a,b){
    const reasons=[];if(!a||!b)return {ok:false,reasons:['Contract model缺失']};
    if(a.contract_id!==b.contract_id)reasons.push('contract_id不同');
    if(a.contract_version!==b.contract_version)reasons.push('contract_version不同');
    if(a.contract_kind!==b.contract_kind)reasons.push('contract_kind不同');
    if(a.identity_fingerprint!==b.identity_fingerprint)reasons.push('只读身份指纹不同');
    if(a.api_fingerprint!==b.api_fingerprint)reasons.push('API Payload语义指纹不同');
    const aa=hash(JSON.stringify(stable(toAiVisiblePayload(a)))),bb=hash(JSON.stringify(stable(toAiVisiblePayload(b))));if(aa!==bb)reasons.push('AI可见字段指纹不同');
    return {ok:reasons.length===0,reasons,contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,v77:{identity:a.identity_fingerprint,api:a.api_fingerprint,ai_visible:aa},v78:{identity:b.identity_fingerprint,api:b.api_fingerprint,ai_visible:bb}};
  }
  function schema(kind){return clone({contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,kind,fields:CONTRACTS[kind]||{},readonly_paths:READONLY_PATHS[kind]||[],internal_only:[...INTERNAL_ONLY]});}
  function snapshot(){return {service:'native_generator_request_contract_v78_phase13',phase:PHASE,contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,kinds:Object.keys(CONTRACTS),outline_field_count:Object.keys(CONTRACTS.outline).length,regeneration_field_count:Object.keys(CONTRACTS.regeneration).length,internal_only:[...INTERNAL_ONLY],readonly_paths:clone(READONLY_PATHS),network_calls:0,mutates_payload:false,unknown_field_policy:'warn_and_legacy_passthrough',internal_meta_policy:'never_send'};}
  const api={version:'native_generator_request_contract_v78_phase13',phase:PHASE,contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,createOutline:(payload,meta={})=>create('outline',payload,meta),createRegeneration:(payload,meta={})=>create('regeneration',payload,meta),create,validateModel,toApiPayload,toAiVisiblePayload,compareModels,schema,snapshot};
  ns.generatorContract=api;globalThis.__V78_GENERATOR_CONTRACT_SERVICE__=api;
  console.info('[V78 Clean Core] Phase13 Generator Request Contract ready',{contract_id:CONTRACT_ID,contract_version:CONTRACT_VERSION,network_calls:0});
})();

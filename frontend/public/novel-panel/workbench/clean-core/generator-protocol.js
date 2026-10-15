/* V78 Clean Core Phase 13: Native Generator Protocol Service.
 * Owns final request-protocol construction for whole-outline and single-scene regeneration.
 * V77 sanitizeApiRequestBody remains an independent validator/fallback builder.
 * Zero network calls: this module only builds, validates, compares and selects request payloads.
 * Three consecutive safe semantic-parity cycles are required before V78 protocol becomes primary.
 */
(function installV78GeneratorProtocolService(){
  if(globalThis.__V78_GENERATOR_PROTOCOL_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const THRESHOLD=3,MAX_HISTORY=100;
  const contractService=()=>ns.generatorContract||globalThis.__V78_GENERATOR_CONTRACT_SERVICE__||null;
  const txt=v=>String(v??'').trim();
  const arr=v=>Array.isArray(v)?v:[];
  const now=()=>new Date().toISOString();
  function clone(v){
    if(v===null||typeof v!=='object')return v;
    if(Array.isArray(v))return v.map(clone);
    const out={};Reflect.ownKeys(v).forEach(k=>{if(typeof k==='string')out[k]=clone(v[k]);});return out;
  }
  function stable(v){
    if(v===null||typeof v!=='object')return v;
    if(Array.isArray(v))return v.map(stable);
    const out={};Object.keys(v).sort().forEach(k=>{if(typeof v[k]!=='function'&&typeof v[k]!=='symbol')out[k]=stable(v[k]);});return out;
  }
  function hash(text=''){let h=2166136261;const s=String(text);for(let i=0;i<s.length;i++){h^=s.charCodeAt(i);h=Math.imul(h,16777619);}return (h>>>0).toString(16).padStart(8,'0');}
  function normalizeRole(value='',gender='未定'){
    const source=txt(value);if(!source)return '普通配角';if(source.length<=12)return source;
    if(/女主|女主角|核心女主/.test(source))return '女主';
    if(/男主|男主角|核心男主/.test(source))return '男主';
    if(/主角|主人公|核心人物/.test(source))return gender==='女'?'女主':gender==='男'?'男主':'重要配角';
    if(/重要配角|关键配角|核心配角|主要配角|反派|白月光|死对头|家人|母亲|父亲|妈妈|爸爸/.test(source))return '重要配角';
    if(/普通配角|配角|路人|邻居|同学|同事|医生|护士|老师|秘书|助理/.test(source))return '普通配角';
    return source.slice(0,12);
  }
  function sanitize(value,parentKey=''){
    if(Array.isArray(value))return value.map(item=>sanitize(item,parentKey));
    if(!value||typeof value!=='object')return value;
    const next={};Object.entries(value).forEach(([key,item])=>{next[key]=sanitize(item,key);});
    const isCharacter=Object.prototype.hasOwnProperty.call(next,'role')&&(Object.prototype.hasOwnProperty.call(next,'name')||Object.prototype.hasOwnProperty.call(next,'gender')||Object.prototype.hasOwnProperty.call(next,'appearance')||parentKey==='characters');
    if(isCharacter)next.role=normalizeRole(next.role,next.gender);
    return next;
  }
  function build(kind,draft={},meta={}){
    let out=sanitize(clone(draft||{}));
    // Test hook is only for deterministic regression tests. Production has no hook installed.
    try{const hook=globalThis.__V78_GENERATOR_PROTOCOL_TEST_HOOK__;if(typeof hook==='function')out=hook(clone(out),kind,clone(meta))||out;}catch(_){}
    return out;
  }
  function fieldType(v){return Array.isArray(v)?'array':v===null?'null':typeof v;}
  function summary(kind,payload={},meta={}){
    const scene=payload?.scene||{};
    return {
      kind,
      keys:Object.keys(payload||{}).sort(),
      types:Object.fromEntries(Object.keys(payload||{}).sort().map(k=>[k,fieldType(payload[k])])),
      novel_hash:hash(payload?.novel_text||''),
      rules_hash:hash(payload?.generation_rules||''),
      genre_hash:hash(payload?.genre||''),
      style_hash:hash(payload?.trailer_style||''),
      camera_hash:hash(payload?.camera||''),
      character_ids:arr(payload?.characters).map(x=>txt(x?.slot_id||x?.id||x?.name)).filter(Boolean),
      character_roles:arr(payload?.characters).map(x=>txt(x?.role)),
      temporary_ids:arr(payload?.temporary_visible_cast).map(x=>txt(x?.continuity_id||x?.entity_id||x?.label)).filter(Boolean),
      manual_added:arr(payload?.manual_added_slot_ids).map(txt).filter(Boolean),
      manual_excluded:arr(payload?.manual_excluded_slot_ids).map(txt).filter(Boolean),
      scene_source_key:txt(scene?.source_key||meta?.source_key),
      scene_source_index:Number(scene?.source_index||meta?.source_index||0),
      outline_shot_count:arr(payload?.outline_shots).length,
      scene_memory_hash:hash(JSON.stringify(stable(payload?.scene_memory||payload?.scene_context?.scene_memory||{}))),
      payload_hash:hash(JSON.stringify(stable(payload||{}))),
      payload_chars:JSON.stringify(payload||{}).length,
    };
  }
  function requiredFields(kind){return kind==='outline'?['novel_text','generation_rules','genre','trailer_style','camera','characters']:['novel_text','generation_rules','scene','outline_shots','characters'];}
  function validate(kind,payload={},meta={}){
    const errors=[],warnings=[],required=requiredFields(kind);
    required.forEach(k=>{const v=payload?.[k];if(v===undefined||v===null||(typeof v==='string'&&!txt(v)))errors.push(`缺少必需字段 ${k}`);});
    if(payload?.characters!==undefined&&!Array.isArray(payload.characters))errors.push('characters必须为数组');
    const charIds=arr(payload?.characters).map(x=>txt(x?.slot_id||x?.id||x?.name)).filter(Boolean);if(new Set(charIds).size!==charIds.length)warnings.push('characters存在重复人物标识');
    const temp=arr(payload?.temporary_visible_cast);temp.forEach((x,i)=>{if(!txt(x?.continuity_id||x?.entity_id||x?.label))warnings.push(`临时人物${i+1}缺少continuity/entity标识`);});
    if(kind==='regeneration'){
      const expected=txt(meta?.source_key),actual=txt(payload?.scene?.source_key);if(expected&&actual&&expected!==actual)errors.push(`scene.source_key不一致 ${actual} != ${expected}`);
      if(expected&&!actual)warnings.push('单镜scene缺少source_key，将由旧链source_key保护层兜底');
    }
    // Chinese camera language is a quality contract. Warn only: never mutate the user's current prompt/protocol here.
    const cameraText=`${txt(payload?.camera)}\n${txt(payload?.generation_rules)}`;
    const englishCamera=(cameraText.match(/\b(?:medium shot|close[- ]?up|wide shot|dolly(?: in| out)?|truck(?:ing)?|pan(?:ning)?|tilt(?:ing)?|zoom(?:ing)?|handheld|over[- ]the[- ]shoulder|bird'?s eye|worm'?s eye)\b/ig)||[]);
    if(englishCamera.length)warnings.push(`检测到英文镜头术语 ${Array.from(new Set(englishCamera.map(x=>x.toLowerCase()))).slice(0,6).join(', ')}；最终输出仍由中文镜头规则约束`);
    return {ok:errors.length===0,errors,warnings,kind,summary:summary(kind,payload,meta)};
  }
  function compare(kind,v77Payload={},v78Payload={},meta={}){
    const a=summary(kind,v77Payload,meta),b=summary(kind,v78Payload,meta),reasons=[];
    if(JSON.stringify(a.keys)!==JSON.stringify(b.keys))reasons.push('字段集合不同');
    if(JSON.stringify(a.types)!==JSON.stringify(b.types))reasons.push('字段类型不同');
    if(a.novel_hash!==b.novel_hash)reasons.push('novel_text不同');
    if(a.rules_hash!==b.rules_hash)reasons.push('generation_rules不同');
    if(a.genre_hash!==b.genre_hash)reasons.push('genre不同');
    if(a.style_hash!==b.style_hash)reasons.push('trailer_style不同');
    if(a.camera_hash!==b.camera_hash)reasons.push('camera不同');
    if(JSON.stringify(a.character_ids)!==JSON.stringify(b.character_ids))reasons.push('正式人物顺序/ID不同');
    if(JSON.stringify(a.character_roles)!==JSON.stringify(b.character_roles))reasons.push('人物role规范化不同');
    if(JSON.stringify(a.temporary_ids)!==JSON.stringify(b.temporary_ids))reasons.push('临时人物ID不同');
    if(JSON.stringify(a.manual_added)!==JSON.stringify(b.manual_added))reasons.push('人工增加人物不同');
    if(JSON.stringify(a.manual_excluded)!==JSON.stringify(b.manual_excluded))reasons.push('人工排除人物不同');
    if(a.scene_source_key!==b.scene_source_key)reasons.push('scene.source_key不同');
    if(a.scene_source_index!==b.scene_source_index)reasons.push('scene.source_index不同');
    if(a.outline_shot_count!==b.outline_shot_count)reasons.push('outline_shots数量不同');
    if(a.scene_memory_hash!==b.scene_memory_hash)reasons.push('Scene Memory不同');
    if(a.payload_hash!==b.payload_hash)reasons.push('完整协议语义指纹不同');
    return {ok:reasons.length===0,reasons,v77:a,v78:b};
  }

  function phase15Production(stage,safe,meta={}){const p=globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__;return p?.decide?.(stage,{safe,kind:meta.kind||'',source_key:meta.source_key||'',reasons:meta.reasons||[]})||null;}
  let enabled=true;
  const outline={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_protocol_warmup'};
  const regeneration=new Map();
  const history=[];
  const stats={prepare_calls:0,outline_calls:0,regeneration_calls:0,v77_selected:0,v78_selected:0,promotions:0,demotions:0,parity_fallbacks:0,validation_fallbacks:0,contract_fallbacks:0,network_calls:0};
  let lastOutline=null,lastRegeneration=null;
  function stateFor(kind,meta,sig){
    const key=txt(meta?.source_key)||'__unknown__';let target=kind==='outline'?outline:regeneration.get(key);
    if(!target){target={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_protocol_warmup'};regeneration.set(key,target);}
    if(target.signature!==sig){target.signature=sig;target.safe_streak=0;target.promoted=false;target.last_executor='v77_protocol_context_reset';target.last_reason='context_changed';}
    return target;
  }
  function authSignature(kind,meta,v77Summary){return hash(JSON.stringify({kind,source_hash:txt(meta?.source_hash),source_revision:Number(meta?.source_revision||0),character_revision:Number(meta?.character_revision||0),source_key:txt(meta?.source_key),source_index:Number(meta?.source_index||0),v77_payload_hash:v77Summary?.payload_hash||''}));}
  function demote(target,reason){if(target.promoted){target.promoted=false;target.demotion_count+=1;stats.demotions+=1;}target.safe_streak=0;target.last_reason=reason;}
  function publicAuth(x){return clone({signature:x.signature,safe_streak:x.safe_streak,promoted:x.promoted,promotion_count:x.promotion_count,demotion_count:x.demotion_count,last_executor:x.last_executor,last_reason:x.last_reason,last_at:x.last_at,threshold:THRESHOLD});}
  function record(row){history.push(clone(row));while(history.length>MAX_HISTORY)history.shift();if(row.kind==='outline')lastOutline=clone(row);else lastRegeneration=clone(row);}
  function prepare(kind,{draft={},v77Payload={},meta={}}={}){
    stats.prepare_calls+=1;if(kind==='outline')stats.outline_calls+=1;else stats.regeneration_calls+=1;
    const legacy=clone(v77Payload||{}),candidate=build(kind,draft,meta),parity=compare(kind,legacy,candidate,meta),v77Validation=validate(kind,legacy,meta),v78Validation=validate(kind,candidate,meta);
    const contract=contractService();
    const v77Contract=contract?.create?.(kind,legacy,meta)||null,v78Contract=contract?.create?.(kind,candidate,meta)||null;
    const v77ContractValidation=v77Contract?contract.validateModel(v77Contract):{ok:false,errors:['Generator Contract Service未加载'],warnings:[]};
    const v78ContractValidation=v78Contract?contract.validateModel(v78Contract):{ok:false,errors:['Generator Contract Service未加载'],warnings:[]};
    const contractParity=(v77Contract&&v78Contract)?contract.compareModels(v77Contract,v78Contract):{ok:false,reasons:['Generator Contract model缺失']};
    const sig=authSignature(kind,meta,parity.v77),authority=stateFor(kind,meta,sig),wasPromoted=authority.promoted;
    const contractOk=Boolean(v77ContractValidation.ok&&v78ContractValidation.ok);
    const safe=parity.ok&&v77Validation.ok&&v78Validation.ok&&contractOk;let executor='v77_protocol_warmup',takeover=false,promotedNow=false,demotedNow=false;
    const production=phase15Production(`protocol_${kind}`,Boolean(enabled&&safe),{kind,source_key:txt(meta?.source_key),reasons:[...parity.reasons,...v77Validation.errors,...v78Validation.errors,...v77ContractValidation.errors,...v78ContractValidation.errors]});
    if(!enabled){executor='v77_protocol_authority_disabled';}
    else if(!safe){
      if(authority.promoted)demotedNow=true;
      const contractFailure=!contractOk;demote(authority,contractFailure?'contract_validation_failed':!parity.ok?'protocol_parity_mismatch':'protocol_validation_failed');
      executor=contractFailure?'v77_protocol_contract_fallback':!parity.ok?'v77_protocol_parity_fallback':'v77_protocol_validation_fallback';
      if(contractFailure)stats.contract_fallbacks+=1;else if(!parity.ok)stats.parity_fallbacks+=1;else stats.validation_fallbacks+=1;
    }else if(production){
      if(production.use_v78){takeover=true;executor='v78_native_protocol_primary_phase15';authority.safe_streak=Math.max(THRESHOLD,Number(authority.safe_streak||0)+1);if(!authority.promoted){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;promotedNow=true;}authority.last_reason='phase15_default_primary_safe';}
      else{demote(authority,'phase15_production_authority_disabled');executor='v77_protocol_realtime_fallback_phase15';}
    }else if(authority.promoted){executor='v78_native_protocol_primary_phase13';takeover=true;}
    else{
      authority.safe_streak+=1;authority.last_reason='stable_protocol_parity';
      if(authority.safe_streak>=THRESHOLD){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;promotedNow=true;}
      executor='v77_protocol_warmup';
    }
    authority.last_executor=executor;authority.last_at=now();
    const selected=takeover?candidate:legacy;if(takeover)stats.v78_selected+=1;else stats.v77_selected+=1;
    const row={at:authority.last_at,kind,source_key:txt(meta?.source_key),executor,takeover,threshold:THRESHOLD,safe_streak:authority.safe_streak,promoted_before:wasPromoted,promoted_after:authority.promoted,promoted_now:promotedNow,demoted_now:demotedNow,parity_ok:parity.ok,parity_reasons:clone(parity.reasons),v77_validation:clone(v77Validation),v78_validation:clone(v78Validation),contract_id:v77Contract?.contract_id||'',contract_version:v77Contract?.contract_version||'',contract_ok:contractOk,contract_parity:clone(contractParity),v77_contract_validation:clone(v77ContractValidation),v78_contract_validation:clone(v78ContractValidation),contract_identity_fingerprint:v78Contract?.identity_fingerprint||'',selected_payload_hash:summary(kind,selected,meta).payload_hash,network_calls:0};
    record(row);
    return {kind,v77_payload:legacy,v78_payload:candidate,selected_payload:clone(selected),decision:clone(row),parity,validation:{v77:v77Validation,v78:v78Validation},contracts:{v77:v77Contract,v78:v78Contract,parity:contractParity,validation:{v77:v77ContractValidation,v78:v78ContractValidation}}};
  }
  function prepareOutline(options={}){return prepare('outline',options);}
  function prepareRegeneration(options={}){return prepare('regeneration',options);}
  function authoritySnapshot(){return {enabled,threshold:THRESHOLD,outline:publicAuth(outline),regeneration:Object.fromEntries([...regeneration.entries()].map(([k,v])=>[k,publicAuth(v)])),stats:clone(stats)};}
  function snapshot(){return {service:'native_generator_protocol_service_v78_phase13',phase:'v78_phase13',enabled,threshold:THRESHOLD,network_calls:0,zero_extra_ai:true,contract_service:contractService()?.version||'',contract_version:contractService()?.contract_version||'',v77_validator_builder_retained:true,outline:publicAuth(outline),regeneration:Object.fromEntries([...regeneration.entries()].map(([k,v])=>[k,publicAuth(v)])),last_outline:clone(lastOutline),last_regeneration:clone(lastRegeneration),history:history.slice(-40),stats:clone(stats)};}
  function setEnabled(v){enabled=Boolean(v);if(!enabled){outline.promoted=false;outline.safe_streak=0;regeneration.forEach(x=>{x.promoted=false;x.safe_streak=0;});}return snapshot();}
  function resetAuthority(){outline.signature='';outline.safe_streak=0;outline.promoted=false;outline.last_reason='manual_reset';regeneration.clear();return snapshot();}
  const api={version:'native_generator_protocol_service_v78_phase13',phase:'v78_phase13',buildOutline:(draft,meta={})=>build('outline',draft,meta),buildRegeneration:(draft,meta={})=>build('regeneration',draft,meta),validateOutline:(payload,meta={})=>validate('outline',payload,meta),validateRegeneration:(payload,meta={})=>validate('regeneration',payload,meta),compareOutline:(v77,v78,meta={})=>compare('outline',v77,v78,meta),compareRegeneration:(v77,v78,meta={})=>compare('regeneration',v77,v78,meta),prepareOutline,prepareRegeneration,authoritySnapshot,snapshot,setEnabled,resetAuthority,threshold:THRESHOLD};
  ns.generatorProtocol=api;globalThis.__V78_GENERATOR_PROTOCOL_SERVICE__=api;
  console.info('[V78 Clean Core] Phase13 Native Generator Protocol Service ready',{threshold:THRESHOLD,network_calls:0});
})();

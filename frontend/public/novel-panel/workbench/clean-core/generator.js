/* V78 Clean Core Phase 13: Outline Generator Authority + Native Protocol integration.
 * Zero-extra-AI shadow validation: V78 and V77 request protocols are compared before send,
 * but exactly ONE network request is issued. V77 warms the authority for 3 safe cycles;
 * V78 becomes the primary request executor only after stable protocol + response coverage.
 * Any pre-send protocol mismatch falls back to V77 immediately. Invalid AI response demotes
 * V78 immediately and is handed to the existing Phase9/legacy writeback safeguards; no hidden retry.
 */
(function installV78OutlineGeneratorAuthority(){
  if(globalThis.__V78_OUTLINE_GENERATOR_ADAPTER__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const txt=v=>String(v??'').trim();
  const arr=v=>Array.isArray(v)?v:[];
  const THRESHOLD=3,MAX_HISTORY=100;
  let enabled=true;
  const outline={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_generator_warmup'};
  const regen=new Map();
  const history=[];
  const stats={requests:0,outline_requests:0,regeneration_requests:0,v77_exec:0,v78_exec:0,promotions:0,demotions:0,protocol_fallbacks:0,response_demotions:0,network_errors:0,hidden_retry_count:0};
  let lastOutline=null,lastRegeneration=null;

  function now(){return new Date().toISOString();}
  function hash(text=''){let h=2166136261;const s=String(text);for(let i=0;i<s.length;i++){h^=s.charCodeAt(i);h=Math.imul(h,16777619);}return (h>>>0).toString(16).padStart(8,'0');}
  function normalizeText(v=''){return txt(v).replace(/\s+/g,'').replace(/[，。！？；：、“”‘’（）()\[\]【】,.!?;:'"`~—…·]/g,'').toLowerCase();}
  function stable(v){
    if(v===null||typeof v!=='object')return v;
    if(Array.isArray(v))return v.map(stable);
    const o={};Object.keys(v).sort().forEach(k=>{if(typeof v[k]!=='undefined'&&typeof v[k]!=='function')o[k]=stable(v[k]);});return o;
  }
  function canonicalizePayload(payload={}){
    const clean=(v)=>{
      if(v===undefined||typeof v==='function'||typeof v==='symbol')return undefined;
      if(v===null||typeof v!=='object')return v;
      if(Array.isArray(v))return v.map(clean).filter(x=>x!==undefined);
      const out={};Object.keys(v).forEach(k=>{const cv=clean(v[k]);if(cv!==undefined)out[k]=cv;});return out;
    };
    let out=clean(payload)||{};
    try{const hook=globalThis.__V78_GENERATOR_TEST_HOOK__;if(typeof hook==='function')out=hook(clone(out))||out;}catch(_){}
    return out;
  }
  function fieldType(v){return Array.isArray(v)?'array':v===null?'null':typeof v;}
  function protocolSummary(kind,payload={},meta={}){
    const chars=arr(payload.characters).map(x=>txt(x?.slot_id||x?.id||x?.name)).filter(Boolean);
    const temp=arr(payload.temporary_visible_cast).map(x=>txt(x?.continuity_id||x?.entity_id||x?.label)).filter(Boolean);
    const scene=payload.scene||{};
    const required=kind==='outline'?['novel_text','generation_rules','genre','trailer_style','camera','characters']:['novel_text','generation_rules','scene','outline_shots','characters'];
    const missing=required.filter(k=>payload[k]===undefined||payload[k]===null||(typeof payload[k]==='string'&&!txt(payload[k])));
    return {
      kind,keys:Object.keys(payload).sort(),types:Object.fromEntries(Object.keys(payload).sort().map(k=>[k,fieldType(payload[k])])),missing,
      novel_hash:hash(payload.novel_text||''),generation_rules_hash:hash(payload.generation_rules||''),camera_hash:hash(payload.camera||''),
      character_ids:chars,temporary_ids:temp,scene_source_key:txt(scene.source_key||meta.source_key),scene_source_index:Number(scene.source_index||meta.source_index||0),
      outline_shot_count:arr(payload.outline_shots).length,payload_hash:hash(JSON.stringify(stable(payload))),payload_chars:JSON.stringify(payload).length,
    };
  }
  function compareProtocol(nativePayload,candidate,kind,meta={}){
    const a=protocolSummary(kind,nativePayload,meta),b=protocolSummary(kind,candidate,meta),reasons=[];
    if(JSON.stringify(a.keys)!==JSON.stringify(b.keys))reasons.push('字段集合不同');
    if(JSON.stringify(a.types)!==JSON.stringify(b.types))reasons.push('字段类型不同');
    if(a.missing.length||b.missing.length)reasons.push(`必需字段缺失 V77=${a.missing.join(',')||'无'} V78=${b.missing.join(',')||'无'}`);
    if(a.novel_hash!==b.novel_hash)reasons.push('novel_text不同');
    if(a.generation_rules_hash!==b.generation_rules_hash)reasons.push('generation_rules不同');
    if(a.camera_hash!==b.camera_hash)reasons.push('camera不同');
    if(JSON.stringify(a.character_ids)!==JSON.stringify(b.character_ids))reasons.push('正式人物顺序/ID不同');
    if(JSON.stringify(a.temporary_ids)!==JSON.stringify(b.temporary_ids))reasons.push('临时人物ID不同');
    if(a.scene_source_key!==b.scene_source_key)reasons.push('scene.source_key不同');
    if(a.outline_shot_count!==b.outline_shot_count)reasons.push('当前镜头上下文数量不同');
    if(a.payload_hash!==b.payload_hash)reasons.push('请求语义指纹不同');
    return {ok:reasons.length===0,reasons,native:a,candidate:b};
  }
  function expectedSources(meta={}){
    return arr(meta.expected_sources).map((x,i)=>({source_key:txt(x?.source_key),source_index:Number(x?.source_index||i+1),source_text:txt(x?.source_text||x?.text)}));
  }
  function validateOutlineResponse(data={},meta={}){
    const shots=arr(data?.outline_shots),scenes=arr(data?.scenes),expected=expectedSources(meta),errors=[],warnings=[];
    if(!shots.length)errors.push('AI未返回outline_shots');
    const returned=[...shots,...scenes];
    const coverage=expected.map((src,idx)=>{
      const sk=txt(src.source_key),si=Number(src.source_index||idx+1),st=normalizeText(src.source_text);
      let matched=returned.some(item=>sk&&txt(item?.source_key)===sk);
      if(!matched)matched=returned.some(item=>Number(item?.source_index||0)===si&&si>0);
      if(!matched&&st)matched=returned.some(item=>normalizeText(item?.source_basis||item?.source_text||item?.text)===st);
      // Preserve V16's safe strict-order fallback when the AI returned exactly one shot per source line.
      if(!matched&&shots.length===expected.length&&shots[idx])matched=true;
      return {source_key:sk,source_index:si,matched};
    });
    const missing=coverage.filter(x=>!x.matched);if(missing.length)errors.push(`缺少原文覆盖 ${missing.length}/${coverage.length}`);
    const explicitKeys=returned.map(x=>txt(x?.source_key)).filter(Boolean);const expectedKeys=new Set(expected.map(x=>x.source_key).filter(Boolean));
    const foreign=expectedKeys.size?explicitKeys.filter(k=>!expectedKeys.has(k)):[];if(foreign.length)warnings.push(`AI返回外来source_key ${[...new Set(foreign)].slice(0,5).join(',')}`);
    return {ok:errors.length===0,errors,warnings,scene_count:scenes.length,shot_count:shots.length,expected_count:expected.length,covered_count:coverage.filter(x=>x.matched).length,coverage};
  }
  function validateRegenerationResponse(data={},meta={}){
    const shots=arr(data?.outline_shots),expected=txt(meta.source_key),errors=[],warnings=[];
    if(!shots.length)errors.push('AI未返回outline_shots');
    if(expected){const explicit=shots.map(x=>txt(x?.source_key)).filter(Boolean),foreign=explicit.filter(k=>k!==expected);if(foreign.length)errors.push(`单镜返回外来source_key ${[...new Set(foreign)].slice(0,5).join(',')}`);}
    shots.forEach((shot,i)=>{if(!txt(shot?.prompt)&&!arr(shot?.timeline_segments).some(seg=>txt(seg?.prompt)))warnings.push(`shot ${i+1} 缺少可见prompt`);});
    return {ok:errors.length===0,errors,warnings,shot_count:shots.length,source_key:expected};
  }
  function signature(kind,meta,protocol){
    const base={kind,source_hash:txt(meta.source_hash),source_revision:Number(meta.source_revision||0),character_revision:Number(meta.character_revision||0),source_key:txt(meta.source_key),expected_keys:expectedSources(meta).map(x=>x.source_key||x.source_index),protocol_hash:protocol?.native?.payload_hash||''};
    return hash(JSON.stringify(base));
  }
  function stateFor(kind,meta,sig){
    const target=kind==='outline'?outline:(regen.get(txt(meta.source_key)||'__unknown__')||{signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_generator_warmup'});
    if(kind!=='outline'&&!regen.has(txt(meta.source_key)||'__unknown__'))regen.set(txt(meta.source_key)||'__unknown__',target);
    if(target.signature!==sig){target.signature=sig;target.safe_streak=0;target.promoted=false;target.last_executor='v77_generator_context_reset';}
    return target;
  }
  function demote(target,reason){if(target.promoted){target.promoted=false;target.demotion_count+=1;stats.demotions+=1;}target.safe_streak=0;target.last_reason=txt(reason);}
  function record(row){history.push(clone(row));while(history.length>MAX_HISTORY)history.shift();if(row.kind==='outline')lastOutline=clone(row);else lastRegeneration=clone(row);}
  function facadeStage(stage,data){try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.(stage,data);}catch(_){}}

  function phase15Production(stage,safe,meta={}){const p=globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__;return p?.decide?.(stage,{safe,kind:meta.kind||'',source_key:meta.source_key||'',reasons:meta.reasons||[]})||null;}
  async function execute(kind,options={}){
    const nativePayload=clone(options.nativePayload||{}),candidate=clone(options.candidatePayload||canonicalizePayload(nativePayload)),meta={...(options.meta||{})};
    const protocolDecision=options.protocolDecision||null;
    const protocol=compareProtocol(nativePayload,candidate,kind,meta),sig=signature(kind,meta,protocol),authority=stateFor(kind,meta,sig);
    const wasPromoted=authority.promoted;let executor='v77_generator_warmup',takeover=false,preSendFallback=false;
    const upstreamSafe=Boolean(!protocolDecision||protocolDecision.takeover===true),production=phase15Production(`generator_${kind}`,Boolean(enabled&&upstreamSafe&&protocol.ok),{kind,source_key:txt(meta.source_key),reasons:[...protocol.reasons,...(upstreamSafe?[]:['上游Protocol未选择V78'])]});
    if(!enabled){executor='v77_generator_authority_disabled';preSendFallback=true;}
    else if(protocolDecision&&protocolDecision.takeover!==true){executor=`v77_generator_${txt(protocolDecision.executor)||'protocol_fallback'}`;preSendFallback=true;}
    else if(!protocol.ok){if(authority.promoted)demote(authority,'protocol_mismatch');else authority.safe_streak=0;executor='v77_generator_protocol_fallback';preSendFallback=true;stats.protocol_fallbacks+=1;}
    else if(production){
      if(production.use_v78){executor='v78_generator_primary_phase15';takeover=true;authority.safe_streak=Math.max(THRESHOLD,Number(authority.safe_streak||0)+1);if(!authority.promoted){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;}authority.last_reason='phase15_default_primary_safe';}
      else{demote(authority,'phase15_production_authority_disabled');executor='v77_generator_realtime_fallback_phase15';preSendFallback=true;}
    }else if(authority.promoted){executor='v78_generator_primary_phase13';takeover=true;}else executor='v77_generator_warmup';
    const row={at:now(),kind,source_key:txt(meta.source_key),signature:sig,executor,takeover,pre_send_protocol_ok:protocol.ok,protocol_reasons:protocol.reasons,protocol_service_executor:txt(protocolDecision?.executor),protocol_service_takeover:Boolean(protocolDecision?.takeover),protocol_service_streak:Number(protocolDecision?.safe_streak||0),contract_id:txt(protocolDecision?.contract_id),contract_version:txt(protocolDecision?.contract_version),contract_ok:protocolDecision?.contract_ok!==false,safe_streak_before:authority.safe_streak,promoted_before:wasPromoted,threshold:THRESHOLD,response:null,promoted_now:false,demoted_now:false,network_calls:1};
    stats.requests+=1;if(kind==='outline')stats.outline_requests+=1;else stats.regeneration_requests+=1;
    facadeStage('generator_authority',{mode:kind,source_key:row.source_key,executor,takeover,protocol_ok:protocol.ok,protocol_service_executor:row.protocol_service_executor,protocol_service_takeover:row.protocol_service_takeover,safe_streak:authority.safe_streak,threshold:THRESHOLD});
    const exec=takeover?options.executeV78:options.executeV77;
    if(typeof exec!=='function')throw new Error(`Phase13 ${takeover?'V78':'V77'} Generator执行器缺失。`);
    if(takeover)stats.v78_exec+=1;else stats.v77_exec+=1;
    let data;
    try{data=await exec(takeover?candidate:nativePayload);}catch(error){stats.network_errors+=1;demote(authority,'network_error');row.demoted_now=wasPromoted;row.error=txt(error?.message||error).slice(0,500);row.safe_streak_after=authority.safe_streak;row.promoted_after=authority.promoted;record(row);try{globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__?.recordPostSendFailure?.(`generator_${kind}`,{kind,source_key:row.source_key,error:row.error});}catch(_){}facadeStage('generator_network_error',{mode:kind,source_key:row.source_key,executor,error:row.error,no_retry:true});throw error;}
    const response=kind==='outline'?validateOutlineResponse(data,meta):validateRegenerationResponse(data,meta);row.response=clone(response);
    const safe=protocol.ok&&response.ok;
    if(safe){
      if(!authority.promoted){authority.safe_streak+=1;if(authority.safe_streak>=THRESHOLD){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;row.promoted_now=true;authority.last_reason='stable_protocol_and_response';}}
    }else{
      const promotedAtStart=wasPromoted||authority.promoted;if(promotedAtStart)row.demoted_now=true;
      demote(authority,response.ok?'protocol_mismatch':'response_contract_failed');
      if(!response.ok)stats.response_demotions+=1;
    }
    authority.last_executor=executor;authority.last_at=row.at;authority.last_response_ok=response.ok;authority.last_protocol_ok=protocol.ok;
    row.safe_streak_after=authority.safe_streak;row.promoted_after=authority.promoted;row.response_fallback_deferred_to_writeback=Boolean(takeover&&!response.ok);record(row);
    facadeStage('generator_response_contract',{mode:kind,source_key:row.source_key,executor,response_ok:response.ok,covered_count:response.covered_count,expected_count:response.expected_count,shot_count:response.shot_count,promoted_now:row.promoted_now,demoted_now:row.demoted_now,safe_streak:authority.safe_streak,threshold:THRESHOLD,no_hidden_retry:true});
    return data;
  }
  function requestOutline(options={}){return execute('outline',options);}
  function requestRegeneration(options={}){return execute('regeneration',options);}
  function authPublic(x){return clone({signature:x.signature,safe_streak:x.safe_streak,promoted:x.promoted,promotion_count:x.promotion_count,demotion_count:x.demotion_count,last_executor:x.last_executor,last_reason:x.last_reason,last_at:x.last_at,last_response_ok:x.last_response_ok,last_protocol_ok:x.last_protocol_ok,threshold:THRESHOLD});}
  function authoritySnapshot(){return {enabled,threshold:THRESHOLD,outline:authPublic(outline),regeneration:Object.fromEntries([...regen.entries()].map(([k,v])=>[k,authPublic(v)])),stats:clone(stats)};}
  function snapshot(){return {service:'outline_generator_authority_v78_phase13',phase:'v78_phase13',enabled,threshold:THRESHOLD,zero_extra_ai_shadow_validation:true,hidden_retry:false,native_v77_validator_retained:true,outline:authPublic(outline),regeneration:Object.fromEntries([...regen.entries()].map(([k,v])=>[k,authPublic(v)])),last_outline:clone(lastOutline),last_regeneration:clone(lastRegeneration),history:history.slice(-40),stats:clone(stats)};}
  function setEnabled(v){enabled=Boolean(v);if(!enabled){outline.promoted=false;outline.safe_streak=0;regen.forEach(x=>{x.promoted=false;x.safe_streak=0;});}return snapshot();}
  function resetAuthority(){outline.signature='';outline.safe_streak=0;outline.promoted=false;outline.last_reason='manual_reset';regen.clear();return snapshot();}
  const api={version:'outline_generator_authority_v78_phase13',phase:'v78_phase13',requestOutline,requestRegeneration,compareProtocol,validateOutlineResponse,validateRegenerationResponse,canonicalizePayload,authoritySnapshot,snapshot,setEnabled,resetAuthority,threshold:THRESHOLD};
  ns.generator=api;globalThis.__V78_OUTLINE_GENERATOR_ADAPTER__=api;
  console.info('[V78 Clean Core] Phase13 Outline Generator Authority ready',{threshold:THRESHOLD,hidden_retry:false,zero_extra_ai:true});
})();

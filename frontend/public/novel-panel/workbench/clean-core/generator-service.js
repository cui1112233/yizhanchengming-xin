/* V78 Clean Core Phase 14: Native Generator Service.
 * Owns the real outline/regeneration request service while retaining the V77 postJSON executor
 * as an independent validator/fallback during promotion. Protocol and Generator authorities remain
 * separate gates. Exactly ONE network request is permitted per service transaction; a failed native
 * request demotes V78 for the next request and is NEVER retried through V77 in the same transaction.
 */
(function installV78NativeGeneratorService(){
  if(globalThis.__V78_NATIVE_GENERATOR_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const THRESHOLD=3,MAX_HISTORY=100;
  const txt=v=>String(v??'').trim();
  const arr=v=>Array.isArray(v)?v:[];
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const now=()=>new Date().toISOString();
  function stable(v){if(v===null||typeof v!=='object')return v;if(Array.isArray(v))return v.map(stable);const o={};Object.keys(v).sort().forEach(k=>{if(v[k]!==undefined&&typeof v[k]!=='function'&&typeof v[k]!=='symbol')o[k]=stable(v[k]);});return o;}
  function hash(text=''){let h=2166136261;for(const ch of String(text)){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(16).padStart(8,'0');}
  function compact(value=''){return String(value||'').replace(/\r/g,'').replace(/[ \t]+\n/g,'\n').replace(/\n{3,}/g,'\n\n').replace(/[ \t]{2,}/g,' ').trim();}
  function normalizeRole(value='',gender='未定'){
    const raw=txt(value),g=txt(gender);if(!raw){if(g==='男')return '男';if(g==='女')return '女';return '未定';}
    if(['男','女','未定','无性别'].includes(raw))return raw;if(/女/.test(raw))return '女';if(/男/.test(raw))return '男';if(/系统|AI|人工智能|机器人|精灵|非人|无性别/.test(raw))return '无性别';return g==='男'?'男':g==='女'?'女':'未定';
  }
  function nativeSanitize(value,parentKey=''){
    if(Array.isArray(value))return value.map(item=>nativeSanitize(item,parentKey));
    if(!value||typeof value!=='object')return value;
    const out={};Object.entries(value).forEach(([k,v])=>{if(v!==undefined&&typeof v!=='function'&&typeof v!=='symbol')out[k]=nativeSanitize(v,k);});
    const isCharacter=parentKey==='characters'||Object.prototype.hasOwnProperty.call(out,'role')||Object.prototype.hasOwnProperty.call(out,'gender')||Object.prototype.hasOwnProperty.call(out,'appearance');
    if(isCharacter)out.role=normalizeRole(out.role,out.gender);
    return out;
  }
  function prepareNativePayload(payload={}){
    let out=nativeSanitize(clone(payload||{}));
    if(typeof out.must_cover_details==='string')out.must_cover_details=compact(out.must_cover_details);
    if(typeof out.shot_rhythm_requirements==='string')out.shot_rhythm_requirements=compact(out.shot_rhythm_requirements);
    try{const hook=globalThis.__V78_GENERATOR_SERVICE_TEST_HOOK__;if(typeof hook==='function')out=hook(clone(out))||out;}catch(_){}
    return out;
  }
  function host(){return globalThis.__V78_GENERATOR_HOST_BRIDGE__||{};}
  function timeoutSeconds(){const h=host();const n=Number(h.currentAiTimeoutSeconds?.()||400);return Math.max(30,Math.min(400,Number.isFinite(n)?n:400));}
  function nativePlan(endpoint,payload,meta={}){
    const seconds=timeoutSeconds(),body=prepareNativePayload(payload);
    return {endpoint:txt(endpoint),method:'POST',body,timeout_seconds:seconds,timeout_ms:seconds*1000+15000,content_type:'application/json',source_key:txt(meta.source_key),body_hash:hash(JSON.stringify(stable(body))),body_chars:JSON.stringify(body).length};
  }
  function legacyPlan(endpoint,payload,meta={}){
    const h=host();
    if(typeof h.describeLegacyRequest==='function'){
      const p=clone(h.describeLegacyRequest(endpoint,payload)||{});p.source_key=txt(meta.source_key);p.body_hash=p.body_hash||hash(JSON.stringify(stable(p.body||{})));p.body_chars=Number(p.body_chars||JSON.stringify(p.body||{}).length);return p;
    }
    return nativePlan(endpoint,payload,meta);
  }
  function comparePlans(a,b){
    const reasons=[];
    if(txt(a.endpoint)!==txt(b.endpoint))reasons.push('endpoint不同');
    if(txt(a.method).toUpperCase()!==txt(b.method).toUpperCase())reasons.push('HTTP method不同');
    if(Number(a.timeout_seconds||0)!==Number(b.timeout_seconds||0))reasons.push('AI超时秒数不同');
    if(txt(a.content_type)!==txt(b.content_type))reasons.push('Content-Type语义不同');
    if(txt(a.body_hash)!==txt(b.body_hash))reasons.push('最终发送Payload指纹不同');
    return {ok:reasons.length===0,reasons,native:{endpoint:a.endpoint,method:a.method,timeout_seconds:a.timeout_seconds,body_hash:a.body_hash,body_chars:a.body_chars},legacy:{endpoint:b.endpoint,method:b.method,timeout_seconds:b.timeout_seconds,body_hash:b.body_hash,body_chars:b.body_chars}};
  }
  let enabled=true,seq=0;
  const outline={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_network_service_warmup'};
  const regen=new Map(),history=[];
  const stats={requests:0,outline_requests:0,regeneration_requests:0,legacy_exec:0,native_exec:0,promotions:0,demotions:0,plan_fallbacks:0,native_network_errors:0,legacy_network_errors:0,prevented_duplicate_network_calls:0,total_network_calls:0,hidden_retry_count:0};
  let lastOutline=null,lastRegeneration=null;
  function signature(kind,meta,plan){return hash(JSON.stringify({kind,endpoint:plan.endpoint,source_hash:txt(meta.source_hash),source_revision:Number(meta.source_revision||0),character_revision:Number(meta.character_revision||0),source_key:txt(meta.source_key),source_index:Number(meta.source_index||0),body_hash:plan.body_hash,timeout_seconds:plan.timeout_seconds}));}
  function stateFor(kind,meta,sig){const key=txt(meta.source_key)||'__unknown__';let s=kind==='outline'?outline:regen.get(key);if(!s){s={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_executor:'v77_network_service_warmup'};regen.set(key,s);}if(s.signature!==sig){s.signature=sig;s.safe_streak=0;s.promoted=false;s.last_executor='v77_network_service_context_reset';s.last_reason='context_changed';}return s;}
  function demote(s,reason){if(s.promoted){s.promoted=false;s.demotion_count+=1;stats.demotions+=1;}s.safe_streak=0;s.last_reason=txt(reason);}
  function publicState(s){return clone({signature:s.signature,safe_streak:s.safe_streak,promoted:s.promoted,promotion_count:s.promotion_count,demotion_count:s.demotion_count,last_executor:s.last_executor,last_reason:s.last_reason,last_at:s.last_at,threshold:THRESHOLD});}
  function record(row){history.push(clone(row));while(history.length>MAX_HISTORY)history.shift();if(row.kind==='outline')lastOutline=clone(row);else lastRegeneration=clone(row);}
  function facadeStage(stage,data){try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.(stage,data);}catch(_){} }
  async function nativeSend(plan,tx){
    const h=host(),transport=ns.transport||globalThis.__V78_CLEAN_TRANSPORT__;
    if(!transport?.request)throw new Error('V78 Native Generator Service未找到Clean Transport。');
    try{h.commitLatestEditableUiState?.(`before_native_generator:${plan.endpoint}`);}catch(_){}
    await h.persistSettingsBeforeRequest?.(plan.endpoint);
    const raw=await transport.request(plan.endpoint,{method:'POST',body:JSON.stringify(plan.body),timeoutMs:plan.timeout_ms,timeoutLabelSeconds:plan.timeout_seconds,headers:{'Content-Type':'application/json','X-V78-Native-Generator-Service':'phase14','X-V78-Generator-Service-Transaction':tx}});
    return typeof h.normalizeResponse==='function'?h.normalizeResponse(raw):raw;
  }
  async function legacySend(endpoint,payload,options={}){
    const legacy=options.executeLegacy||host().executeLegacy;
    if(typeof legacy!=='function')throw new Error('V77 Generator网络Fallback执行器缺失。');
    return legacy(endpoint,payload);
  }
  function phase15Production(stage,safe,meta={}){const p=globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__;return p?.decide?.(stage,{safe,kind:meta.kind||'',source_key:meta.source_key||'',reasons:meta.reasons||[]})||null;}
  async function executeNetwork(kind,{endpoint,payload,meta={},generator_executor='',executeLegacy}={}){
    stats.requests+=1;if(kind==='outline')stats.outline_requests+=1;else stats.regeneration_requests+=1;
    const nPlan=nativePlan(endpoint,payload,meta),lPlan=legacyPlan(endpoint,payload,meta),parity=comparePlans(nPlan,lPlan),sig=signature(kind,meta,lPlan),authority=stateFor(kind,meta,sig),wasPromoted=authority.promoted;
    let executor='v77_network_service_warmup',takeover=false,demotedNow=false,promotedNow=false;
    const upstreamV78=txt(generator_executor).startsWith('v78_'),production=phase15Production(`network_${kind}`,Boolean(enabled&&parity.ok&&upstreamV78),{kind,source_key:txt(meta.source_key),reasons:[...parity.reasons,...(upstreamV78?[]:['上游Generator未选择V78'])]});
    if(!enabled)executor='v77_network_service_authority_disabled';
    else if(!parity.ok){if(authority.promoted)demotedNow=true;demote(authority,'request_plan_mismatch');stats.plan_fallbacks+=1;executor='v77_network_service_plan_fallback';}
    else if(production){
      if(production.use_v78){executor='v78_native_generator_service_primary_phase15';takeover=true;authority.safe_streak=Math.max(THRESHOLD,Number(authority.safe_streak||0)+1);if(!authority.promoted){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;promotedNow=true;}authority.last_reason='phase15_default_primary_safe';}
      else{demote(authority,upstreamV78?'phase15_production_authority_disabled':'phase15_upstream_v77');executor='v77_network_service_realtime_fallback_phase15';}
    }else if(authority.promoted){executor='v78_native_generator_service_primary_phase14';takeover=true;}
    const tx=`gen14-${Date.now().toString(36)}-${(++seq).toString(36)}`;let networkCalls=0;
    const row={at:now(),transaction_id:tx,kind,source_key:txt(meta.source_key),generator_executor:txt(generator_executor),executor,takeover,plan_parity_ok:parity.ok,plan_reasons:clone(parity.reasons),plan:clone(parity),safe_streak_before:authority.safe_streak,promoted_before:wasPromoted,threshold:THRESHOLD,promoted_now:false,demoted_now:demotedNow,network_calls:0,hidden_retry:false};
    facadeStage('native_generator_service_authority',{mode:kind,source_key:row.source_key,executor,takeover,plan_ok:parity.ok,safe_streak:authority.safe_streak,threshold:THRESHOLD,generator_executor:row.generator_executor});
    const one=async(fn)=>{if(networkCalls>=1){stats.prevented_duplicate_network_calls+=1;throw new Error('Phase14已阻止同一Native Generator Service事务重复发起网络请求。');}networkCalls+=1;stats.total_network_calls+=1;row.network_calls=networkCalls;return fn();};
    try{
      let data;
      if(takeover){stats.native_exec+=1;data=await one(()=>nativeSend(nPlan,tx));}
      else{stats.legacy_exec+=1;data=await one(()=>legacySend(endpoint,payload,{executeLegacy}));}
      if(parity.ok&&!takeover&&enabled&&!globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__){authority.safe_streak+=1;authority.last_reason='stable_request_service_parity';if(authority.safe_streak>=THRESHOLD){authority.promoted=true;authority.promotion_count+=1;stats.promotions+=1;promotedNow=true;}}
      authority.last_executor=executor;authority.last_at=row.at;row.safe_streak_after=authority.safe_streak;row.promoted_after=authority.promoted;row.promoted_now=promotedNow;record(row);
      facadeStage('native_generator_service_return',{mode:kind,source_key:row.source_key,executor,network_calls:networkCalls,promoted_now:promotedNow,safe_streak:authority.safe_streak,threshold:THRESHOLD});
      return data;
    }catch(error){
      if(takeover){stats.native_network_errors+=1;if(authority.promoted)demotedNow=true;demote(authority,'native_network_error_no_retry');try{globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__?.recordPostSendFailure?.(`network_${kind}`,{kind,source_key:row.source_key,error:txt(error?.message||error)});}catch(_){}}
      else stats.legacy_network_errors+=1;
      authority.last_executor=executor;authority.last_at=row.at;row.safe_streak_after=authority.safe_streak;row.promoted_after=authority.promoted;row.demoted_now=demotedNow;row.error=txt(error?.message||error).slice(0,700);record(row);
      facadeStage('native_generator_service_error',{mode:kind,source_key:row.source_key,executor,network_calls:networkCalls,error:row.error,no_retry:true});
      throw error;
    }
  }
  async function request(kind,options={}){
    const protocol=ns.generatorProtocol||globalThis.__V78_GENERATOR_PROTOCOL_SERVICE__,generator=ns.generator||globalThis.__V78_OUTLINE_GENERATOR_ADAPTER__;
    const draft=clone(options.draft||{}),v77Payload=clone(options.v77Payload||{}),meta={...(options.meta||{})};
    const bundle=kind==='outline'?protocol?.prepareOutline?.({draft,v77Payload,meta}):protocol?.prepareRegeneration?.({draft,v77Payload,meta});
    const nativePayload=bundle?.v77_payload||v77Payload,candidate=bundle?.v78_payload||v77Payload,protocolDecision=bundle?.decision||null;
    const execV77=(payload)=>executeNetwork(kind,{endpoint:options.endpoint,payload,meta,generator_executor:'v77_generator',executeLegacy:options.executeLegacy});
    const execV78=(payload)=>executeNetwork(kind,{endpoint:options.endpoint,payload,meta,generator_executor:'v78_generator',executeLegacy:options.executeLegacy});
    if(generator){return kind==='outline'?generator.requestOutline({nativePayload,candidatePayload:candidate,protocolDecision,meta,executeV77:execV77,executeV78:execV78}):generator.requestRegeneration({nativePayload,candidatePayload:candidate,protocolDecision,meta,executeV77:execV77,executeV78:execV78});}
    return executeNetwork(kind,{endpoint:options.endpoint,payload:bundle?.selected_payload||v77Payload,meta,generator_executor:'generator_authority_unavailable',executeLegacy:options.executeLegacy});
  }
  function requestOutline(options={}){return request('outline',options);}
  function requestRegeneration(options={}){return request('regeneration',options);}
  function authoritySnapshot(){return {enabled,threshold:THRESHOLD,outline:publicState(outline),regeneration:Object.fromEntries([...regen.entries()].map(([k,v])=>[k,publicState(v)])),stats:clone(stats)};}
  function snapshot(){return {service:'native_generator_service_v78_phase14',phase:'v78_phase14',enabled,threshold:THRESHOLD,v77_network_validator_retained:true,clean_transport_primary_after_promotion:true,phase15_production_default_primary:Boolean(globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__?.isEnabled?.()),v77_realtime_fallback:true,exactly_one_network_request:true,hidden_retry:false,protocol_service:ns.generatorProtocol?.version||'',generator_authority:ns.generator?.version||'',outline:publicState(outline),regeneration:Object.fromEntries([...regen.entries()].map(([k,v])=>[k,publicState(v)])),last_outline:clone(lastOutline),last_regeneration:clone(lastRegeneration),history:history.slice(-40),stats:clone(stats)};}
  function setEnabled(v){enabled=Boolean(v);if(!enabled){outline.promoted=false;outline.safe_streak=0;regen.forEach(s=>{s.promoted=false;s.safe_streak=0;});}return snapshot();}
  function resetAuthority(){outline.signature='';outline.safe_streak=0;outline.promoted=false;outline.last_reason='manual_reset';regen.clear();return snapshot();}
  const api={version:'native_generator_service_v78_phase14',phase:'v78_phase14',requestOutline,requestRegeneration,prepareNativePayload,nativePlan,legacyPlan,comparePlans,authoritySnapshot,snapshot,setEnabled,resetAuthority,threshold:THRESHOLD};
  ns.generatorService=api;globalThis.__V78_NATIVE_GENERATOR_SERVICE__=api;
  console.info('[V78 Clean Core] Phase14 Native Generator Service ready',{threshold:THRESHOLD,hidden_retry:false,exactly_one_network_request:true});
})();

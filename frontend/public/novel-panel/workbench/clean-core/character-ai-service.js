/* V78 Clean Core Phase 16: unified Style / Character Facts / Relations / Appearance AI service.
 * Business semantics and writeback stay in CharacterCore 2.0. This module owns the
 * request contract + network service boundary. V77 remains a pre-send validator/fallback.
 * One service invocation performs exactly one network request; no hidden retry is added.
 */
(function installV78CharacterAIService(){
  if(globalThis.__V78_CHARACTER_AI_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const txt=v=>String(v??'').trim();
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const stable=v=>{
    if(Array.isArray(v)) return '['+v.map(stable).join(',')+']';
    if(v&&typeof v==='object') return '{'+Object.keys(v).sort().map(k=>JSON.stringify(k)+':'+stable(v[k])).join(',')+'}';
    return JSON.stringify(v);
  };
  const now=()=>new Date().toISOString();
  const MAX_HISTORY=100;
  let enabled=true,seq=0;
  const stats={
    version:'character_ai_service_v78_phase16',total:0,v78_primary:0,v77_fallback:0,success:0,error:0,cancelled:0,
    protocol_mismatch:0,hidden_retry_suppressed:0,post_send_errors:0,last_stage:'',last_executor:'',last_request_id:'',history:[],by_stage:{}
  };
  function host(){return globalThis.__V78_CHARACTER_AI_HOST_BRIDGE__||null;}
  function stageOf(input={}){
    const h=host();
    return txt(input.request_stage||h?.markerStage?.(input.marker)||'unknown')||'unknown';
  }
  function stageRow(stage){
    const key=txt(stage)||'unknown';
    if(!stats.by_stage[key]) stats.by_stage[key]={total:0,v78_primary:0,v77_fallback:0,success:0,error:0,cancelled:0,last_executor:'',last_at:''};
    return stats.by_stage[key];
  }
  function buildNativePayload(input={}){
    const h=host(),instance=txt(input.instance_id||h?.currentInstance?.());
    const marker=txt(input.marker),meta=input.meta&&typeof input.meta==='object'?input.meta:{};
    const appearanceMarker=txt(h?.markers?.appearance);
    return {
      novel_text:String(input.prompt??''),
      generation_rules:`${marker}\ncharacter_core_version=2\ninstance_id=${instance}${marker===appearanceMarker?'\nplain_text_card=true':''}`,
      request_stage:stageOf(input),
      instance_id:instance,
      slot_id:txt(meta.slot_id),
      slot_token:txt(meta.slot_token),
    };
  }
  function parity(nativePayload,legacyPayload){
    const a=stable(nativePayload),b=stable(legacyPayload||{}),ok=a===b;
    const reasons=[];
    if(!ok){
      const keys=Array.from(new Set([...Object.keys(nativePayload||{}),...Object.keys(legacyPayload||{})])).sort();
      keys.forEach(k=>{if(stable(nativePayload?.[k])!==stable(legacyPayload?.[k]))reasons.push(`field:${k}`);});
    }
    return {ok,reasons,native_fingerprint:a.length+':'+hashString(a),legacy_fingerprint:b.length+':'+hashString(b)};
  }
  function hashString(value=''){
    let h=2166136261>>>0;for(let i=0;i<value.length;i++){h^=value.charCodeAt(i);h=Math.imul(h,16777619)>>>0;}return h.toString(16).padStart(8,'0');
  }
  function remember(row){stats.history.push(clone(row));while(stats.history.length>MAX_HISTORY)stats.history.shift();}
  async function timeoutSeconds(){
    try{const v=Number(await host()?.resolveTimeoutSeconds?.());if(Number.isFinite(v)&&v>=30)return Math.min(400,Math.max(30,v));}catch(_){}
    return 400;
  }
  async function request(input={}){
    const h=host();if(!h?.buildLegacyPayload||!h?.legacyRawRequest)throw new Error('V78 Character AI Service 未找到 CharacterCore Host Bridge。');
    const stage=stageOf(input),row=stageRow(stage),marker=txt(input.marker),retriesRequested=Math.max(0,Number(input.retries_requested)||0);
    const nativePayload=buildNativePayload(input),legacyPayload=h.buildLegacyPayload(String(input.prompt??''),marker,input.meta||{});
    const check=parity(nativePayload,legacyPayload);
    const transport=ns.transport||globalThis.__V78_CLEAN_TRANSPORT__;
    const transportReady=Boolean(transport?.request);
    if(!transportReady)check.reasons=[...check.reasons,'clean_transport_missing'];
    const useV78=Boolean(enabled&&check.ok&&transportReady);
    const executor=useV78?'v78_character_ai_primary_phase16':'v77_character_ai_realtime_fallback_phase16';
    const requestId=`charai16-${Date.now().toString(36)}-${(++seq).toString(36)}`;
    stats.total++;row.total++;stats.last_stage=stage;stats.last_executor=executor;stats.last_request_id=requestId;
    if(retriesRequested>0)stats.hidden_retry_suppressed+=retriesRequested;
    if(!check.ok)stats.protocol_mismatch++;
    if(useV78){stats.v78_primary++;row.v78_primary++;}else{stats.v77_fallback++;row.v77_fallback++;}
    const baseLog={at:now(),request_id:requestId,stage,executor,protocol_ok:check.ok,protocol_reasons:check.reasons,retries_requested:retriesRequested,hidden_retry:false,network_calls:1,slot_id:txt(input.meta?.slot_id)};
    try{
      let data;
      if(useV78){
        const timeout=await timeoutSeconds();
        data=await transport.request('/api/character-core/analyze',{
          method:'POST',
          body:JSON.stringify(nativePayload),
          headers:{'Content-Type':'application/json','X-VideoPromptTool-Session':txt(nativePayload.instance_id),'X-V78-Character-AI':'phase16'},
          timeoutMs:timeout*1000+15000,
          timeoutLabelSeconds:timeout,
        });
      }else{
        data=await h.legacyRawRequest({prompt:String(input.prompt??''),marker,meta:input.meta||{}});
      }
      stats.success++;row.success++;row.last_executor=executor;row.last_at=now();remember({...baseLog,status:'success'});return data;
    }catch(error){
      const cancelled=error?.name==='AbortError'||/cancel|abort|取消/i.test(txt(error?.message));
      stats[cancelled?'cancelled':'error']++;row[cancelled?'cancelled':'error']++;row.last_executor=executor;row.last_at=now();
      if(useV78){stats.post_send_errors++;}
      remember({...baseLog,status:cancelled?'cancelled':'error',error:txt(error?.message).slice(0,500),post_send_no_retry:useV78});
      // Critical invariant: once the selected request path reached transport, do not
      // send a second request through the other executor in the same service call.
      throw error;
    }
  }
  function snapshot(){return clone({...stats,enabled,service:'character_ai_service_v78_phase16',phase:'v78_phase16',default_route:enabled?'v78_primary':'v77_fallback',authority:'independent_character_ai_authority_phase16',v77_validator_retained:true,exactly_one_network_request_per_service_call:true,hidden_retry:false,explicit_business_repairs_preserved:true,contract:'video_prompt.character_ai_stage_request@1.0.0'});}
  function setEnabled(value){enabled=Boolean(value);return snapshot();}
  function reset(){stats.total=stats.v78_primary=stats.v77_fallback=stats.success=stats.error=stats.cancelled=stats.protocol_mismatch=stats.hidden_retry_suppressed=stats.post_send_errors=0;stats.last_stage=stats.last_executor=stats.last_request_id='';stats.history.length=0;stats.by_stage={};return snapshot();}
  const api={version:'character_ai_service_v78_phase16',phase:'v78_phase16',contract_id:'video_prompt.character_ai_stage_request',contract_version:'1.0.0',request,buildNativePayload,parity,snapshot,setEnabled,reset,isEnabled:()=>enabled};
  ns.characterAI=api;globalThis.__V78_CHARACTER_AI_SERVICE__=api;
  console.info('[V78 Clean Core] Phase16 Character/Style AI Service ready',{default_route:'v78_primary',v77_fallback:true,hidden_retry:false});
})();

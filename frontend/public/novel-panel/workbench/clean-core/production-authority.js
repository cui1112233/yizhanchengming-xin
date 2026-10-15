/* V78 Clean Core Phase 15: Outline Production Mainline Authority.
 * Default route is V78 from the FIRST safe transaction. V77 remains an independent
 * validator / pre-send / pre-writeback fallback. A request that has already reached
 * the network is NEVER repeated through V77 in the same transaction.
 */
(function installV78OutlineProductionAuthority(){
  if(globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const txt=v=>String(v??'').trim();
  const now=()=>new Date().toISOString();
  const MAX=120;
  let enabled=true,mode='v78_primary_with_v77_realtime_fallback',seq=0;
  const stages={},history=[];
  const stats={decisions:0,v78_primary:0,v77_fallback:0,preflight_fallbacks:0,authority_disabled_fallbacks:0,post_send_errors:0,manual_disable_count:0,manual_enable_count:0};
  function stageRow(name){
    const key=txt(name)||'unknown';
    if(!stages[key])stages[key]={stage:key,v78_primary:0,v77_fallback:0,last_decision:null,last_changed_at:null};
    return stages[key];
  }
  function decide(stage,context={}){
    const row=stageRow(stage),safe=context.safe!==false,reasons=Array.isArray(context.reasons)?context.reasons.filter(Boolean).map(String):[];
    const useV78=Boolean(enabled&&safe);
    const reason=!enabled?'production_authority_disabled':safe?'safe_v78_primary':'validator_or_parity_failed';
    const decision={decision_id:`prod15-${Date.now().toString(36)}-${(++seq).toString(36)}`,at:now(),stage:row.stage,kind:txt(context.kind),source_key:txt(context.source_key),safe,use_v78:useV78,executor:useV78?'v78_production_primary_phase15':'v77_realtime_fallback_phase15',reason,reasons:clone(reasons),network_started:Boolean(context.network_started),writeback_started:Boolean(context.writeback_started)};
    stats.decisions+=1;
    if(useV78){stats.v78_primary+=1;row.v78_primary+=1;}else{stats.v77_fallback+=1;row.v77_fallback+=1;if(!enabled)stats.authority_disabled_fallbacks+=1;else stats.preflight_fallbacks+=1;}
    row.last_decision=clone(decision);row.last_changed_at=decision.at;history.push(clone(decision));while(history.length>MAX)history.shift();
    return decision;
  }
  function recordPostSendFailure(stage,context={}){
    const row=stageRow(stage);stats.post_send_errors+=1;
    const item={decision_id:`prod15-post-${Date.now().toString(36)}-${(++seq).toString(36)}`,at:now(),stage:row.stage,kind:txt(context.kind),source_key:txt(context.source_key),safe:false,use_v78:false,executor:'v78_post_send_error_no_retry_phase15',reason:'post_send_error_no_second_ai_request',reasons:[txt(context.error)].filter(Boolean),network_started:true,no_retry:true};
    row.last_decision=clone(item);row.last_changed_at=item.at;history.push(clone(item));while(history.length>MAX)history.shift();return item;
  }
  function setEnabled(value,reason='manual'){
    const next=Boolean(value);if(next!==enabled){if(next)stats.manual_enable_count+=1;else stats.manual_disable_count+=1;}enabled=next;
    history.push({decision_id:`prod15-toggle-${Date.now().toString(36)}-${(++seq).toString(36)}`,at:now(),stage:'production_authority',safe:enabled,use_v78:enabled,executor:enabled?'v78_production_primary_phase15':'v77_realtime_fallback_phase15',reason:txt(reason)||'manual'});while(history.length>MAX)history.shift();return snapshot();
  }
  function snapshot(){return {service:'outline_production_authority_v78_phase15',phase:'v78_phase15',enabled,mode,default_route:enabled?'v78_primary':'v77_fallback',first_safe_request_v78:true,v77_validator_retained:true,pre_send_fallback:true,pre_writeback_fallback:true,post_send_retry:false,exactly_one_ai_request:true,stages:clone(stages),history:history.slice(-40),stats:clone(stats)};}
  function resetStats(){Object.keys(stats).forEach(k=>stats[k]=0);Object.keys(stages).forEach(k=>delete stages[k]);history.length=0;return snapshot();}
  const api={version:'outline_production_authority_v78_phase15',phase:'v78_phase15',decide,recordPostSendFailure,setEnabled,snapshot,resetStats,isEnabled:()=>enabled,mode:()=>mode};
  ns.production=api;globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__=api;
  console.info('[V78 Clean Core] Phase15 production authority ready',{default_route:'v78_primary',v77_fallback:true,no_second_ai_request:true});
})();

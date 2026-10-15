/* V78 Clean Core Phase 9: Outline AI Response / Writeback Adapter.
 * Scope: response envelope recovery -> deterministic source_key alignment -> writeback candidate -> parity -> safe apply.
 * The outline/regeneration AI generators remain V77 native. V77 canonical/writeback remains validator/fallback.
 */
(function installV78OutlineResponseAdapter(){
  if(globalThis.__V78_OUTLINE_RESPONSE_ADAPTER__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const arr=v=>Array.isArray(v)?v:[];
  const obj=v=>v&&typeof v==='object'&&!Array.isArray(v)?v:{};
  const txt=v=>String(v??'').trim();
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const uniq=v=>[...new Set(arr(v).map(txt).filter(Boolean))];
  const now=()=>new Date().toISOString();
  const PROMOTION_STREAK=3;
  function stable(v){if(Array.isArray(v))return '['+v.map(stable).join(',')+']';if(v&&typeof v==='object')return '{'+Object.keys(v).sort().map(k=>JSON.stringify(k)+':'+stable(v[k])).join(',')+'}';return JSON.stringify(v??null);}
  function hash(v){let h=2166136261;for(const ch of stable(v)){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(36);}
  function normText(v){return txt(v).replace(/\s+/g,'').replace(/[，。！？；：“”‘’、,.!?;:'"\-—_（）()【】\[\]{}]/g,'').toLowerCase();}
  function parseMaybeJson(value){
    if(value&&typeof value==='object') return clone(value);
    const raw=txt(value);if(!raw)return {};
    const attempts=[raw];const fence=raw.match(/```(?:json)?\s*([\s\S]*?)```/i);if(fence)attempts.push(fence[1]);
    const a=raw.indexOf('{'),b=raw.lastIndexOf('}');if(a>=0&&b>a)attempts.push(raw.slice(a,b+1));
    for(const item of attempts){try{const parsed=JSON.parse(item);if(typeof parsed==='string')return parseMaybeJson(parsed);return obj(parsed);}catch(_){}}
    return {};
  }
  function normalizeEnvelope(raw){
    let data=parseMaybeJson(raw);const visited=new Set();
    for(let i=0;i<5;i++){
      if(!data||typeof data!=='object'||visited.has(data))break;visited.add(data);
      if(Array.isArray(data.scenes)||Array.isArray(data.outline_shots))break;
      let next=null;
      for(const key of ['data','result','output','response','payload']){const v=data[key];if(v&&typeof v==='object'){next=v;break;}if(typeof v==='string'){const parsed=parseMaybeJson(v);if(Object.keys(parsed).length){next=parsed;break;}}}
      if(!next)break;data=next;
    }
    return {...obj(data),scenes:arr(data?.scenes).map(clone),outline_shots:arr(data?.outline_shots).map(clone)};
  }
  function simpleSources(novelText=''){
    const lines=String(novelText||'').split(/\r?\n/).map(txt).filter(Boolean);
    return lines.map((source_text,i)=>({source_text,source_index:i+1}));
  }
  function expectedSources(novelText='',helpers={}){
    let rows=[];try{rows=arr(helpers.splitSource?.(novelText));}catch(_){}
    if(!rows.length)rows=simpleSources(novelText);
    return rows.map((row,i)=>{const source_text=txt(row.source_text||row.text||row.original_text);let source_key='';try{source_key=txt(helpers.buildSourceKey?.(source_text,i+1));}catch(_){}if(!source_key)source_key=`${i+1}_${hash(source_text)}`;return {source_index:i+1,source_key,source_text};});
  }
  function chooseTarget(record={},index=0,expected=[],sceneLink=null){
    const byKey=new Map(expected.map((x,i)=>[txt(x.source_key),i]));const directKey=txt(record.source_key);
    if(directKey&&byKey.has(directKey))return {index:byKey.get(directKey),strategy:'source_key'};
    if(sceneLink&&Number.isInteger(sceneLink.index))return {index:sceneLink.index,strategy:'parent_scene'};
    const basis=normText(record.source_basis||record.source_text||record.text||record.original_text);
    if(basis){const exact=expected.findIndex(x=>normText(x.source_text)===basis);if(exact>=0)return {index:exact,strategy:'source_text_exact'};}
    const sourceIndex=Number(record.source_index||0);if(sourceIndex>=1&&sourceIndex<=expected.length)return {index:sourceIndex-1,strategy:'source_index'};
    return {index:Math.min(Math.max(0,index),Math.max(0,expected.length-1)),strategy:'ordered_fallback'};
  }
  function alignResponse(raw,novelText='',helpers={}){
    const data=normalizeEnvelope(raw),expected=expectedSources(novelText,helpers),strategies={},issues=[],sceneLinks=new Map();
    const alignedScenes=arr(data.scenes).map((scene,i)=>{
      const chosen=chooseTarget(scene,i,expected),target=expected[chosen.index];strategies[chosen.strategy]=(strategies[chosen.strategy]||0)+1;
      if(!target){issues.push({code:'scene_unmapped',index:i});return clone(scene);}
      const out={...clone(scene),source_index:target.source_index,source_key:target.source_key,source_text:target.source_text,text:target.source_text};
      const id=txt(scene.id||scene.scene_id);if(id)sceneLinks.set(id,{index:chosen.index,target});return out;
    });
    const alignedShots=arr(data.outline_shots).map((shot,i)=>{
      const parent=txt(shot.parent_scene_id||arr(shot.parent_scene_ids)[0]);const link=parent?sceneLinks.get(parent):null,chosen=chooseTarget(shot,i,expected,link),target=expected[chosen.index];strategies[chosen.strategy]=(strategies[chosen.strategy]||0)+1;
      if(!target){issues.push({code:'shot_unmapped',index:i});return clone(shot);}
      return {...clone(shot),source_index:target.source_index,source_key:target.source_key,source_basis:target.source_text};
    });
    const out={...data,scenes:alignedScenes,outline_shots:alignedShots};
    return {data:out,expected,strategies,issues,source_hash:hash(expected.map(x=>[x.source_key,x.source_text]))};
  }
  function shotPrompt(shot={}){const segments=arr(shot.timeline_segments);return txt(shot.prompt)||segments.map(x=>txt(x.prompt)).filter(Boolean).join('|');}
  function candidateSemantic(candidate={}){
    const scenes=arr(candidate.scenes).map(s=>({source_key:txt(s.source_key),source_index:Number(s.source_index||0),source_text:txt(s.source_text||s.text),character_slot_ids:arr(s.character_slot_ids).map(txt),manual_added_slot_ids:uniq(s.manual_added_slot_ids).sort(),manual_excluded_slot_ids:uniq(s.manual_excluded_slot_ids).sort()}));
    const byScene=new Map(scenes.map(x=>[x.source_key,[]]));
    const sceneIdToKey=new Map(arr(candidate.scenes).map(s=>[txt(s.id),txt(s.source_key)]));
    arr(candidate.outline_shots).forEach(shot=>{const parent=txt(shot.parent_scene_id||arr(shot.parent_scene_ids)[0]),key=txt(shot.source_key)||sceneIdToKey.get(parent)||'';if(!byScene.has(key))byScene.set(key,[]);byScene.get(key).push({source_basis:txt(shot.source_basis),prompt:shotPrompt(shot),timeline:arr(shot.timeline_segments).map(x=>({start:Number(x.start_offset||0),end:Number(x.end_offset||0),prompt:txt(x.prompt),shot_size:txt(x.shot_size),shot_angle:txt(x.shot_angle),movement:txt(x.movement),transition:txt(x.transition)})),temporary:arr(shot.temporary_characters).map(t=>({continuity_id:txt(t.continuity_id||t.entity_id),label:txt(t.label),profile:obj(t.continuity_profile),wardrobe:obj(t.wardrobe_state)}))});});
    return {scenes,shots:[...byScene.entries()].map(([source_key,shots])=>({source_key,shots}))};
  }
  function validateCandidate(candidate={},expected=[],plans=[]){
    const errors=[],warnings=[],scenes=arr(candidate.scenes),shots=arr(candidate.outline_shots),expectedKeys=expected.map(x=>x.source_key),expectedSet=new Set(expectedKeys),keys=scenes.map(s=>txt(s.source_key));
    if(scenes.length!==expected.length)errors.push({code:'scene_count_mismatch',expected:expected.length,current:scenes.length});
    if(new Set(keys).size!==keys.length)errors.push({code:'duplicate_source_key'});
    keys.forEach(k=>{if(!expectedSet.has(k))errors.push({code:'foreign_source_key',source_key:k});});
    expectedKeys.forEach(k=>{if(!keys.includes(k))errors.push({code:'missing_source_key',source_key:k});});
    const sceneIds=new Set(scenes.map(s=>txt(s.id)).filter(Boolean)),covered=new Set();
    shots.forEach((shot,i)=>{const parent=txt(shot.parent_scene_id||arr(shot.parent_scene_ids)[0]);if(parent&&!sceneIds.has(parent))errors.push({code:'invalid_parent_scene',shot:i,parent_scene_id:parent});if(parent)covered.add(parent);if(!shotPrompt(shot))errors.push({code:'empty_prompt',shot:i});});
    scenes.forEach((scene,i)=>{if(scene.id&&!covered.has(txt(scene.id)))warnings.push({code:'scene_without_shot',source_key:txt(scene.source_key)});const plan=arr(plans)[i];if(plan&&arr(plan.character_slot_ids).length&&stable(arr(scene.character_slot_ids).map(txt))!==stable(arr(plan.character_slot_ids).map(txt)))errors.push({code:'formal_cast_changed',source_key:txt(scene.source_key)});});
    return {ok:errors.length===0,errors,warnings,scene_count:scenes.length,shot_count:shots.length,source_key_count:new Set(keys).size};
  }
  function parityCandidates(legacy={},candidate={}){
    const a=candidateSemantic(legacy),b=candidateSemantic(candidate),diffs=[];
    if(stable(a.scenes)!==stable(b.scenes))diffs.push({type:'scenes',legacy_hash:hash(a.scenes),v78_hash:hash(b.scenes)});
    if(stable(a.shots)!==stable(b.shots))diffs.push({type:'shots',legacy_hash:hash(a.shots),v78_hash:hash(b.shots)});
    return {ok:diffs.length===0,diffs,mismatch_count:diffs.length,legacy_hash:hash(a),v78_hash:hash(b),validator:'v77_writeback_semantic_validator_phase9'};
  }
  function prepareOutline(input={}){
    const helpers=obj(input.helpers),aligned=alignResponse(input.data,input.novelText,helpers);
    if(typeof helpers.canonicalize!=='function')return {ok:false,reason:'canonicalize_missing',aligned};
    const legacyBase=helpers.canonicalize(clone(normalizeEnvelope(input.data))),v78Base=helpers.canonicalize(clone(aligned.data));
    const legacy=typeof helpers.mergePrevious==='function'?helpers.mergePrevious(legacyBase):legacyBase;
    const candidate=typeof helpers.mergePrevious==='function'?helpers.mergePrevious(v78Base):v78Base;
    const validation=validateCandidate(candidate,aligned.expected,input.lineCharacterPlans);const parity=parityCandidates(legacy,candidate);
    const prepared={ok:validation.ok,aligned,legacy:clone(legacy),candidate:clone(candidate),validation,parity,created_at:now()};lastOutline=clone({validation,parity,alignment:aligned.strategies,created_at:prepared.created_at});return prepared;
  }
  function phase15Production(stage,safe,meta={}){const p=globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__;return p?.decide?.(stage,{safe,kind:meta.kind||'',source_key:meta.source_key||'',reasons:meta.reasons||[],writeback_started:false})||null;}
  let authorityEnabled=true;
  const outlineAuthority={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,takeover_count:0,warmup_count:0,fallback_count:0,last_executor:'v77_writeback_fallback',last_decision:null,last_changed_at:null};
  const regenAuthority={global_signature:'',scenes:{},promotion_count:0,demotion_count:0,takeover_count:0,warmup_count:0,fallback_count:0,last_decision:null,last_changed_at:null};
  function outlineSignature(input={},prepared={}){return hash({source_hash:prepared.aligned?.source_hash||hash(input.novelText),character_revision:Number(input.characterRevision||0),plans:arr(input.lineCharacterPlans).map(p=>({source_key:txt(p.source_key),character_slot_ids:arr(p.character_slot_ids).map(txt),manual_added:uniq(p.manual_added_slot_ids).sort(),manual_excluded:uniq(p.manual_excluded_slot_ids).sort()}))});}
  function resolveOutlineAuthority(input={},prepared={}){
    const sig=outlineSignature(input,prepared);if(outlineAuthority.signature!==sig){if(outlineAuthority.promoted)outlineAuthority.demotion_count+=1;Object.assign(outlineAuthority,{signature:sig,safe_streak:0,promoted:false,last_executor:'v77_writeback_fallback',last_changed_at:now()});}
    const was=outlineAuthority.promoted,safe=Boolean(authorityEnabled&&prepared.ok&&prepared.parity?.ok),production=phase15Production('writeback_outline',safe,{kind:'outline',reasons:[...(prepared.validation?.errors||[]).map?.(x=>x.code||x)||[],...(prepared.parity?.diffs||[]).map?.(x=>x.type||x)||[]]});
    if(production){
      if(safe&&production.use_v78){outlineAuthority.safe_streak=Math.max(PROMOTION_STREAK,Number(outlineAuthority.safe_streak||0)+1);if(!outlineAuthority.promoted){outlineAuthority.promoted=true;outlineAuthority.promotion_count+=1;outlineAuthority.last_changed_at=now();}}
      else{outlineAuthority.safe_streak=0;if(outlineAuthority.promoted){outlineAuthority.promoted=false;outlineAuthority.demotion_count+=1;outlineAuthority.last_changed_at=now();}}
    }else if(safe){outlineAuthority.safe_streak+=1;if(!outlineAuthority.promoted&&outlineAuthority.safe_streak>=PROMOTION_STREAK){outlineAuthority.promoted=true;outlineAuthority.promotion_count+=1;outlineAuthority.last_changed_at=now();}}
    else{outlineAuthority.safe_streak=0;if(outlineAuthority.promoted){outlineAuthority.promoted=false;outlineAuthority.demotion_count+=1;outlineAuthority.last_changed_at=now();}}
    const takeover=Boolean(safe&&(production?production.use_v78:outlineAuthority.promoted)),executor=takeover?(production?'v78_outline_writeback_primary_phase15':'v78_outline_writeback_primary_phase9'):safe?(production?'v77_outline_writeback_realtime_validator_phase15':'v77_outline_writeback_warmup_validator'):'v77_outline_writeback_fallback';
    if(takeover)outlineAuthority.takeover_count+=1;else if(safe)outlineAuthority.warmup_count+=1;else outlineAuthority.fallback_count+=1;
    outlineAuthority.last_executor=executor;outlineAuthority.last_decision={takeover,executor,safe_streak:outlineAuthority.safe_streak,promoted:outlineAuthority.promoted,promoted_now:!was&&outlineAuthority.promoted,demoted_now:was&&!outlineAuthority.promoted,threshold:PROMOTION_STREAK,validation:clone(prepared.validation),parity:clone(prepared.parity),alignment:clone(prepared.aligned?.strategies||{}),decided_at:now()};
    return clone(outlineAuthority.last_decision);
  }
  function applyOutline(stateRef={},candidate={}){const data=clone(candidate);stateRef.scenes=arr(data.scenes);stateRef.outlineShots=arr(data.outline_shots);return {scene_count:stateRef.scenes.length,shot_count:stateRef.outlineShots.length,applied_at:now(),executor:'v78_outline_writeback_primary_phase9'};}
  function prepareRegeneration(input={}){
    const helpers=obj(input.helpers),data=normalizeEnvelope(input.data),scene=obj(input.sceneSnapshot),sourceText=txt(input.sourceText||scene.source_text||scene.text),sceneId=txt(input.sceneId||scene.id),legacy=arr(input.legacyRefreshed).map(clone),raw=arr(data.outline_shots);
    const candidate=raw.map((shot)=>{let out=clone(shot);try{if(typeof helpers.makeShot==='function')out=helpers.makeShot(out);}catch(_){};out={...out,parent_scene_id:sceneId,parent_scene_ids:[sceneId],source_basis:sourceText,source_index:Number(scene.source_index||1),source_key:txt(scene.source_key)};try{if(typeof helpers.mergeTemporary==='function')out.temporary_characters=helpers.mergeTemporary(scene.temporary_characters||[],out.temporary_characters||[],sourceText,txt(scene.temporary_scene_group||scene.source_key||`temp_scene_${scene.source_index||1}`));}catch(_){}return out;});
    const validation={ok:candidate.length>0&&candidate.every(x=>Boolean(shotPrompt(x))),errors:[],warnings:[],shot_count:candidate.length};if(!candidate.length)validation.errors.push({code:'no_shots'});candidate.forEach((x,i)=>{if(!shotPrompt(x))validation.errors.push({code:'empty_prompt',shot:i});if(txt(x.parent_scene_id)!==sceneId)validation.errors.push({code:'wrong_parent',shot:i});});validation.ok=validation.errors.length===0;
    const sem=x=>arr(x).map(shot=>({prompt:shotPrompt(shot),timeline:arr(shot.timeline_segments).map(s=>({start:Number(s.start_offset||0),end:Number(s.end_offset||0),prompt:txt(s.prompt),shot_size:txt(s.shot_size),shot_angle:txt(s.shot_angle),movement:txt(s.movement),transition:txt(s.transition)})),temporary:arr(shot.temporary_characters).map(t=>({continuity_id:txt(t.continuity_id||t.entity_id),label:txt(t.label),profile:obj(t.continuity_profile),wardrobe:obj(t.wardrobe_state)}))}));
    const a=sem(legacy),b=sem(candidate),parity={ok:stable(a)===stable(b),diffs:stable(a)===stable(b)?[]:[{type:'regeneration_shots',legacy_hash:hash(a),v78_hash:hash(b)}],mismatch_count:stable(a)===stable(b)?0:1,validator:'v77_regeneration_writeback_semantic_validator_phase9'};
    const prepared={ok:validation.ok,legacy,candidate,validation,parity,source_key:txt(scene.source_key),created_at:now()};lastRegeneration=clone({source_key:prepared.source_key,validation,parity,created_at:prepared.created_at});return prepared;
  }
  function regenSignature(input={}){const scene=obj(input.sceneSnapshot);return hash({source_hash:txt(input.sourceHash),character_revision:Number(input.characterRevision||0),source_key:txt(scene.source_key),source_text:txt(input.sourceText||scene.source_text),character_slot_ids:arr(scene.character_slot_ids).map(txt),manual_added:uniq(scene.manual_added_slot_ids).sort(),manual_excluded:uniq(scene.manual_excluded_slot_ids).sort(),temporary:arr(scene.temporary_characters).map(t=>({id:txt(t.continuity_id||t.entity_id),profile:obj(t.continuity_profile),wardrobe:obj(t.wardrobe_state)}))});}
  function resolveRegenerationAuthority(input={},prepared={}){
    const globalSig=hash({source_hash:txt(input.sourceHash),character_revision:Number(input.characterRevision||0)});if(regenAuthority.global_signature!==globalSig){regenAuthority.global_signature=globalSig;regenAuthority.scenes={};regenAuthority.last_changed_at=now();}
    const key=txt(prepared.source_key||input.sceneSnapshot?.source_key)||'scene_unknown',sig=regenSignature(input),current=obj(regenAuthority.scenes[key]);if(current.signature!==sig){if(current.promoted)regenAuthority.demotion_count+=1;regenAuthority.scenes[key]={signature:sig,safe_streak:0,promoted:false,last_executor:'v77_regeneration_writeback_fallback',last_changed_at:now()};}
    const row=regenAuthority.scenes[key],was=row.promoted,safe=Boolean(authorityEnabled&&prepared.ok&&prepared.parity?.ok),production=phase15Production('writeback_regeneration',safe,{kind:'regeneration',source_key:key,reasons:[...(prepared.validation?.errors||[]).map?.(x=>x.code||x)||[],...(prepared.parity?.diffs||[]).map?.(x=>x.type||x)||[]]});
    if(production){
      if(safe&&production.use_v78){row.safe_streak=Math.max(PROMOTION_STREAK,Number(row.safe_streak||0)+1);if(!row.promoted){row.promoted=true;regenAuthority.promotion_count+=1;row.last_changed_at=now();}}
      else{row.safe_streak=0;if(row.promoted){row.promoted=false;regenAuthority.demotion_count+=1;row.last_changed_at=now();}}
    }else if(safe){row.safe_streak=Number(row.safe_streak||0)+1;if(!row.promoted&&row.safe_streak>=PROMOTION_STREAK){row.promoted=true;regenAuthority.promotion_count+=1;row.last_changed_at=now();}}
    else{row.safe_streak=0;if(row.promoted){row.promoted=false;regenAuthority.demotion_count+=1;row.last_changed_at=now();}}
    const takeover=Boolean(safe&&(production?production.use_v78:row.promoted)),executor=takeover?(production?'v78_regeneration_writeback_primary_phase15':'v78_regeneration_writeback_primary_phase9'):safe?(production?'v77_regeneration_writeback_realtime_validator_phase15':'v77_regeneration_writeback_warmup_validator'):'v77_regeneration_writeback_fallback';if(takeover)regenAuthority.takeover_count+=1;else if(safe)regenAuthority.warmup_count+=1;else regenAuthority.fallback_count+=1;
    Object.assign(row,{last_executor:executor,last_validation:clone(prepared.validation),last_parity:clone(prepared.parity),last_decision_at:now()});regenAuthority.last_decision={source_key:key,takeover,executor,safe_streak:row.safe_streak,promoted:row.promoted,promoted_now:!was&&row.promoted,demoted_now:was&&!row.promoted,threshold:PROMOTION_STREAK,validation:clone(prepared.validation),parity:clone(prepared.parity),decided_at:now()};return clone(regenAuthority.last_decision);
  }
  function applyRegeneration(input={}){
    const stateRef=input.stateRef||{},sceneId=txt(input.sceneId),scene=obj(input.sceneSnapshot),shots=arr(input.shots).map(clone),helpers=obj(input.helpers);
    stateRef.outlineShots=arr(stateRef.outlineShots).filter(shot=>!arr(shot.parent_scene_ids).map(txt).includes(sceneId)&&txt(shot.parent_scene_id)!==sceneId);stateRef.outlineShots.push(...shots);
    stateRef.scenes=arr(stateRef.scenes).map(item=>txt(item.id)===sceneId?{...item,source_key:txt(scene.source_key)||txt(helpers.buildSourceKey?.(input.sourceText,scene.source_index||1)),characters:[...arr(scene.characters)],character_slot_ids:[...arr(helpers.sceneEffectiveSlotIds?.(scene)||scene.character_slot_ids)],relation_selected_slot_ids:[...arr(helpers.normalizeCastIds?.(scene.relation_selected_slot_ids)||scene.relation_selected_slot_ids)],keyword_selected_slot_ids:[...arr(helpers.normalizeCastIds?.(scene.keyword_selected_slot_ids)||scene.keyword_selected_slot_ids)],manual_added_slot_ids:[...arr(helpers.normalizeCastIds?.(scene.manual_added_slot_ids)||scene.manual_added_slot_ids)],manual_excluded_slot_ids:[...arr(helpers.normalizeCastIds?.(scene.manual_excluded_slot_ids)||scene.manual_excluded_slot_ids)],character_core_cast:arr(scene.character_core_cast).map(clone),characters_mode:arr(helpers.normalizeCastIds?.(scene.manual_added_slot_ids)||scene.manual_added_slot_ids).length||arr(helpers.normalizeCastIds?.(scene.manual_excluded_slot_ids)||scene.manual_excluded_slot_ids).length?'manual':'auto'}:item);
    return {scene_id:sceneId,shot_count:shots.length,applied_at:now(),executor:'v78_regeneration_writeback_primary_phase9'};
  }
  function authoritySnapshot(){const rows=Object.values(regenAuthority.scenes);return {service:'outline_response_writeback_authority_v78_phase9',enabled:authorityEnabled,promotion_threshold:PROMOTION_STREAK,outline:clone(outlineAuthority),regeneration:{global_signature:regenAuthority.global_signature,promotion_count:regenAuthority.promotion_count,demotion_count:regenAuthority.demotion_count,takeover_count:regenAuthority.takeover_count,warmup_count:regenAuthority.warmup_count,fallback_count:regenAuthority.fallback_count,promoted_scene_count:rows.filter(r=>r.promoted).length,scenes:clone(regenAuthority.scenes),last_decision:clone(regenAuthority.last_decision),last_changed_at:regenAuthority.last_changed_at}};}
  function setAuthorityEnabled(v){authorityEnabled=Boolean(v);if(!authorityEnabled){outlineAuthority.promoted=false;outlineAuthority.safe_streak=0;outlineAuthority.last_executor='v77_writeback_fallback';Object.values(regenAuthority.scenes).forEach(r=>{r.promoted=false;r.safe_streak=0;r.last_executor='v77_regeneration_writeback_fallback';});}return authoritySnapshot();}
  function resetAuthority(){Object.assign(outlineAuthority,{signature:'',safe_streak:0,promoted:false,last_executor:'v77_writeback_fallback',last_decision:null});regenAuthority.global_signature='';regenAuthority.scenes={};return authoritySnapshot();}
  let lastOutline=null,lastRegeneration=null;
  function snapshot(){return {service:'outline_response_adapter_v78_phase9',phase:'v78_phase9',authoritative:'phase15_response_recovery_source_key_alignment_v78_default_writeback_with_v77_validator_fallback',last_outline:clone(lastOutline),last_regeneration:clone(lastRegeneration),authority:authoritySnapshot()};}
  const api={version:'outline_response_adapter_v78_phase9',phase:'v78_phase9',promotion_threshold:PROMOTION_STREAK,normalizeEnvelope,expectedSources,alignResponse,validateCandidate,parityCandidates,prepareOutline,resolveOutlineAuthority,applyOutline,prepareRegeneration,resolveRegenerationAuthority,applyRegeneration,authoritySnapshot,setAuthorityEnabled,resetAuthority,snapshot};
  ns.outlineResponse=api;globalThis.__V78_OUTLINE_RESPONSE_ADAPTER__=api;
  console.info('[V78 Clean Core] Phase9 outline response/writeback adapter ready',{version:api.version,promotion_threshold:PROMOTION_STREAK});
})();

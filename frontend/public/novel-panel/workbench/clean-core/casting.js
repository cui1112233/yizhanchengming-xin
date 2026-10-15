/* V78 Clean Core Phase 6: Scene Cast / Alias Keyword / Temporary Continuity / Evidence authority service.
 * Safety contract:
 * - This service does not write CharacterCore production state.
 * - CharacterCore 2.0 remains the independent live validator and immediate fallback.
 * - Phase6 additionally promotes Temporary Continuity Cache and Scene Cast Evidence to V78 primary state only after repeated semantic parity.
 * - CharacterCore validator mirrors are always retained; any ambiguity, manual-override uncertainty, continuity/evidence mismatch, or structural conflict falls back immediately.
 */
(function installV78CastingComputeService(){
  if(globalThis.__V78_CASTING_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const arr=(v)=>Array.isArray(v)?v:[];
  const obj=(v)=>v&&typeof v==='object'&&!Array.isArray(v)?v:{};
  const txt=(v)=>String(v??'').trim();
  const clone=(v)=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const uniq=(values)=>[...new Set(arr(values).map(txt).filter(Boolean))];
  const now=()=>new Date().toISOString();
  const CONFIDENCE=0.8, TEMP_CONFIDENCE=0.55, TEMP_TTL=5, PROMOTION_STREAK=3;
  let lastProduction=null,lastKeywordProduction=null,lastShadow=null,lastParity=null;
  let gateEnabled=true;
  const gateState={decisions:{},scene_authority:{},takeover_count:0,fallback_count:0,warmup_count:0,promotion_count:0,demotion_count:0,last_decision:null,last_reset_at:now(),authority_signature:''};
  const keywordAuthority={safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_parity:null,last_executor:'character_core_keyword_fallback',last_changed_at:now(),authority_signature:''};
  const evidenceAuthority={scene_authority:{},promotion_count:0,demotion_count:0,takeover_count:0,fallback_count:0,last_decision:null,authority_signature:''};
  const temporaryAuthority={safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,last_parity:null,last_executor:'character_core_temporary_fallback',last_changed_at:now(),authority_signature:'',primary_cache:{},validator_cache:{}};

  function core(){
    try{const service=globalThis.__V78_CHARACTER_SERVICE__;if(service?.state)return service.state()||{};}catch(_){}
    try{const api=globalThis.__characterCoreV2Api;if(api?.state)return clone(api.state()||{});}catch(_){}
    const root=(typeof state!=='undefined'?state:globalThis.state)||{};
    return clone(root.characterCoreV2||root.character_core_v2||{});
  }
  function scenes(){const root=(typeof state!=='undefined'?state:globalThis.state)||{};return clone(arr(root.scenes));}
  function fingerprint(value=''){let h=2166136261;for(const ch of String(value??'')){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(36);}
  function sourceKey(source,line){return `${Math.max(1,Number(line||1))}_${fingerprint(txt(source).replace(/\r\n?/g,'\n').trim())}`;}
  function authoritySignature(){
    const c=core();
    const compact={
      source_hash:txt(c.source_hash),character_revision:Number(c.character_revision||0),
      slots:arr(c.slots).map(x=>[txt(x.slot_id),txt(x.person_id),Boolean(x.disabled),stage(x)]),
      aliases:arr(c.aliases).map(x=>[txt(x.binding_id),txt(x.alias),txt(x.target_person_id),txt(x.target_slot_id),Boolean(x.disabled),Boolean(x.manual_locked),Number(x.confidence??0),x.valid_from_line??null,x.valid_to_line??null]),
      relationships:arr(c.relationships).map(x=>[txt(x.relation_id),txt(x.source_person_id||x.from_person_id),txt(x.target_person_id||x.to_person_id),txt(x.label||x.relation_type),Boolean(x.disabled),Boolean(x.manual_locked)])
    };
    return fingerprint(JSON.stringify(compact));
  }
  function resetAuthority(reason='authority_changed'){
    const sig=authoritySignature();
    gateState.decisions={};gateState.scene_authority={};gateState.takeover_count=0;gateState.fallback_count=0;gateState.warmup_count=0;gateState.last_decision=null;gateState.last_reset_at=now();gateState.authority_signature=sig;
    keywordAuthority.safe_streak=0;keywordAuthority.promoted=false;keywordAuthority.last_parity=null;keywordAuthority.last_executor='character_core_keyword_fallback';keywordAuthority.last_changed_at=now();keywordAuthority.authority_signature=sig;
    evidenceAuthority.scene_authority={};evidenceAuthority.last_decision=null;evidenceAuthority.authority_signature=sig;
    temporaryAuthority.safe_streak=0;temporaryAuthority.promoted=false;temporaryAuthority.last_parity=null;temporaryAuthority.last_executor='character_core_temporary_fallback';temporaryAuthority.last_changed_at=now();temporaryAuthority.authority_signature=sig;temporaryAuthority.primary_cache={};temporaryAuthority.validator_cache={};
    console.info('[V78_CAST_AUTHORITY_RESET]',{reason,authority_signature:sig});
  }
  function ensureAuthorityContext(){
    const sig=authoritySignature();
    if(!gateState.authority_signature){gateState.authority_signature=sig;keywordAuthority.authority_signature=sig;evidenceAuthority.authority_signature=sig;temporaryAuthority.authority_signature=sig;}
    else if(gateState.authority_signature!==sig||keywordAuthority.authority_signature!==sig||evidenceAuthority.authority_signature!==sig||temporaryAuthority.authority_signature!==sig)resetAuthority('character_or_source_revision_changed');
    return sig;
  }
  function stage(slot={}){return txt(slot?.age?.visual_age_stage||slot?.current_values?.visual_age_stage||slot?.visual_age_stage||slot?.stage_label);}
  function lineInScope(row={},line=null){if(!Number.isFinite(Number(line)))return true;const n=Number(line),a=row.valid_from_line,b=row.valid_to_line;return !(a!==null&&a!==undefined&&n<Number(a))&&!(b!==null&&b!==undefined&&n>Number(b));}
  function keywordPriority(row={}){const t=txt(row.type||row.keyword_type),source=txt(row.source);if(row.manual_locked||/manual/.test(t))return 100;if(/canonical|display|slot|forced_roster/.test(`${t}|${source}`))return 90;if(row.relation_id||/relation/.test(t))return 80;return 70;}
  function addEntry(target,keyword,person,meta={}){const k=txt(keyword),pid=txt(person?.person_id);if(!k||!pid)return;target.push({keyword:k,person_id:pid,slot_id:txt(meta.slot_id),keyword_id:txt(meta.keyword_id||meta.binding_id),binding_id:txt(meta.binding_id),relation_id:txt(meta.relation_id),type:txt(meta.type||'alias'),source:txt(meta.source||'character_core'),confidence:Number(meta.confidence??1),manual_locked:Boolean(meta.manual_locked),valid_from_line:meta.valid_from_line??null,valid_to_line:meta.valid_to_line??null,timeline_scope:txt(meta.timeline_scope),evidence:txt(meta.evidence)});}
  function buildKeywordIndex(){
    const c=core(),out={};
    const people=arr(c.people),slots=arr(c.slots).filter(x=>!x?.disabled),bindings=arr(c.aliases).filter(x=>!x?.disabled&&x?.keyword_enabled!==false);
    for(const person of people){
      const entries=[];addEntry(entries,person.canonical_name,person,{type:'canonical',source:'person',confidence:1});addEntry(entries,person.primary_display_name,person,{type:'display',source:'person',confidence:1});arr(person.aliases).forEach(v=>addEntry(entries,v,person,{type:'person_alias',source:'person',confidence:1}));
      slots.filter(s=>txt(s.person_id)===txt(person.person_id)).forEach(slot=>{addEntry(entries,slot.display_name,person,{slot_id:slot.slot_id,type:'slot_name',source:'forced_roster',confidence:1});addEntry(entries,slot.base_name,person,{slot_id:slot.slot_id,type:'slot_base',source:'forced_roster',confidence:1});addEntry(entries,slot.canonical_name_hint,person,{slot_id:slot.slot_id,type:'canonical_hint',source:'forced_roster',confidence:1});arr(slot.aliases).forEach(v=>addEntry(entries,v,person,{slot_id:slot.slot_id,type:'guide_alias',source:'forced_roster',confidence:1}));});
      bindings.filter(a=>txt(a.target_person_id)===txt(person.person_id)).forEach(a=>addEntry(entries,a.alias,person,{slot_id:a.target_slot_id,type:a.manual_locked?'manual_alias':a.relation_id?'relation_alias':'ai_alias',source:a.manual_locked?'manual':'relationship_graph',keyword_id:a.binding_id,binding_id:a.binding_id,relation_id:a.relation_id,valid_from_line:a.valid_from_line,valid_to_line:a.valid_to_line,timeline_scope:a.timeline_scope,evidence:a.evidence,confidence:a.manual_locked?1:Number(a.confidence??.75),manual_locked:a.manual_locked}));
      const seen=new Map();entries.sort((a,b)=>b.keyword.length-a.keyword.length||keywordPriority(b)-keywordPriority(a)).forEach(e=>{const key=`${e.keyword}|${e.person_id}|${e.slot_id}`;const old=seen.get(key);if(!old||keywordPriority(e)>keywordPriority(old))seen.set(key,e);});
      out[txt(person.person_id)]={person_id:txt(person.person_id),canonical_name:txt(person.canonical_name||person.primary_display_name),entries:[...seen.values()]};
    }
    const total=Object.values(out).reduce((n,row)=>n+arr(row.entries).length,0);
    return {service:'casting_compute_v78_phase6',generated_at:now(),total,index:out};
  }
  function detectRole(source,keyword){
    const s=txt(source),k=txt(keyword).replace(/[.*+?^${}()|[\]\\]/g,'\\$&');if(!k)return 'unknown';
    const clauses=s.split(/[，,。！？!?；;\n]+/).map(txt).filter(Boolean).filter(c=>c.includes(keyword));let roles=[];
    for(const clause of clauses){
      if(new RegExp(`(?:提到|说起|谈及|听说|得知|名单).{0,16}${k}|${k}.{0,4}(?:只是被提及|并未在场)`).test(clause))roles.push('mentioned_only');
      else if(new RegExp(`(?:照片|相片|合影|画像).{0,20}${k}|${k}.{0,8}(?:照片|相片|合影|画像)`).test(clause))roles.push('photo_visible');
      else if(new RegExp(`(?:监控|屏幕|视频|手机画面|电脑画面|直播).{0,20}${k}|${k}.{0,8}(?:监控|屏幕|视频|手机画面|电脑画面|直播)`).test(clause))roles.push('screen_visible');
      else if(new RegExp(`(?:梦里|梦中|梦境|噩梦).{0,20}${k}`).test(clause))roles.push('dream_visible');
      else if(new RegExp(`(?:回忆|记忆中|想起|往事|前世).{0,20}${k}`).test(clause))roles.push('memory_visible');
      else if(new RegExp(`(?:电话里|手机里|耳机里|广播中|录音中|画外音|门外传来).{0,20}${k}|${k}.{0,8}(?:的声音|画外音|在电话里说|从门外传来)`).test(clause))roles.push('offscreen_voice');
      else if(new RegExp(`${k}[：:]|${k}.{0,8}(?:说|问|喊|低声|开口|回答)`).test(clause))roles.push('speaker_visible');
      else roles.push('visible');
    }
    const rank={unknown:0,mentioned_only:1,offscreen_voice:2,photo_visible:3,screen_visible:3,memory_visible:3,dream_visible:3,visible:5,speaker_visible:7};return roles.sort((a,b)=>(rank[b]||0)-(rank[a]||0))[0]||'visible';
  }
  function chooseStageSlot(c,personId,source,previousStageByPerson,hitSlots=[]){
    const slots=arr(c.slots).filter(s=>!s?.disabled&&txt(s.person_id)===txt(personId));if(!slots.length)return null;
    const uniqueHits=uniq(hitSlots);if(uniqueHits.length===1){const bound=slots.find(s=>txt(s.slot_id)===uniqueHits[0]);if(bound)return bound;}
    const explicit=slots.filter(s=>stage(s)&&txt(source).includes(stage(s)));if(explicit.length===1)return explicit[0];
    const previous=slots.find(s=>txt(s.slot_id)===txt(previousStageByPerson?.[personId]));if(previous)return previous;
    const current=slots.filter(s=>/当前|主线|present|current/i.test(txt(s.age?.timeline_stage)||txt(s.inferred_values?.timeline_stage))||s.inferred_values?.is_current_timeline===true);if(current.length===1)return current[0];
    return slots.length===1?slots[0]:slots.slice().sort((a,b)=>Number(a.return_index||0)-Number(b.return_index||0))[0];
  }
  function matchKeywords(source,lineNumber=null,options={}){
    const s=txt(source),pack=options.keywordIndex||buildKeywordIndex(),all=[];
    for(const row of Object.values(pack.index||{})){for(const entry of arr(row.entries)){if(!lineInScope(entry,lineNumber)||Number(entry.confidence??1)<CONFIDENCE)continue;let from=0;while(entry.keyword&&(from=s.indexOf(entry.keyword,from))>=0){all.push({...entry,start:from,end:from+entry.keyword.length});from+=Math.max(1,entry.keyword.length);}}}
    all.sort((a,b)=>a.start-b.start||(b.end-b.start)-(a.end-a.start)||keywordPriority(b)-keywordPriority(a));
    const chosen=[],ambiguous=[];let i=0;
    while(i<all.length){const seed=all[i],same=all.filter(h=>h.start===seed.start&&h.end===seed.end&&h.keyword===seed.keyword),bestP=Math.max(...same.map(keywordPriority)),best=same.filter(h=>keywordPriority(h)===bestP),persons=uniq(best.map(h=>h.person_id));if(persons.length>1){ambiguous.push({keyword:seed.keyword,start:seed.start,end:seed.end,candidates:best.map(h=>({person_id:h.person_id,slot_id:h.slot_id,confidence:h.confidence}))});i+=same.length;continue;}let hit=best.slice().sort((a,b)=>Number(Boolean(b.slot_id))-Number(Boolean(a.slot_id)))[0];if(!chosen.some(prev=>Math.max(prev.start,hit.start)<Math.min(prev.end,hit.end)))chosen.push(hit);i+=same.length;}
    return {source:s,line_number:Number.isFinite(Number(lineNumber))?Number(lineNumber):null,hits:chosen,ambiguous};
  }
  function computeFormalCast(source,lineNumber=null,context={}){
    const c=core(),match=matchKeywords(source,lineNumber,context),byPerson=new Map(),previous=obj(context.previousStageByPerson);
    for(const hit of match.hits){const row=byPerson.get(hit.person_id)||[];row.push(hit);byPerson.set(hit.person_id,row);}
    const selected=[];byPerson.forEach((hits,personId)=>{const slot=chooseStageSlot(c,personId,source,previous,hits.map(h=>h.slot_id).filter(Boolean));if(!slot)return;const best=hits.slice().sort((a,b)=>keywordPriority(b)-keywordPriority(a)||b.keyword.length-a.keyword.length)[0];const role=detectRole(source,best.keyword),keywordEvidence=hits.map(h=>({type:'keyword',keyword:h.keyword,keyword_id:h.keyword_id,person_id:h.person_id,slot_id:txt(h.slot_id),binding_id:txt(h.binding_id),relation_id:txt(h.relation_id),source:txt(h.source),confidence:Number(h.confidence??1),reason:`V78本地关键词命中:${h.keyword}`}));selected.push({slot_id:txt(slot.slot_id),slot_token:txt(slot.slot_token),person_id:txt(personId),display_name:txt(slot.display_name||slot.base_name),scene_role:role,evidence:clone(keywordEvidence),matched_keywords:uniq(hits.map(h=>h.keyword)),keyword_evidence:clone(keywordEvidence),relationship_evidence:keywordEvidence.filter(e=>e.relation_id||e.binding_id),confidence:Math.max(...hits.map(h=>Number(h.confidence||0)),CONFIDENCE),selection_source:'v78_casting_compute'});previous[personId]=txt(slot.slot_id);});
    return {selected:selected.sort((a,b)=>Number(b.confidence)-Number(a.confidence)),ambiguous:match.ambiguous,previousStageByPerson:previous};
  }
  function normalizeTemp(raw={}){const label=txt(raw.label||raw.source_phrase||raw.identity_hint||raw.entity_type);if(!label)return null;const participation=txt(raw.participation_state||'visible_candidate');return {entity_id:txt(raw.entity_id||raw.mention_id||`temp_${fingerprint(label)}`),label,source_phrase:txt(raw.source_phrase||label),aliases:uniq([label,...arr(raw.aliases)]),entity_type:txt(raw.entity_type||'temporary_character'),participation_state:participation,valid_from_line:Number.isFinite(Number(raw.valid_from_line))?Number(raw.valid_from_line):null,valid_to_line:Number.isFinite(Number(raw.valid_to_line))?Number(raw.valid_to_line):null,count:Math.max(1,Number(raw.count||1)||1),gender_hint:txt(raw.gender_hint),age_stage_hint:txt(raw.age_stage_hint),species_hint:txt(raw.species_hint),identity_hint:txt(raw.identity_hint),evidence:txt(raw.evidence),confidence:Number(raw.confidence??.8),continuity_id:txt(raw.continuity_id||raw.entity_id||`temp_${fingerprint(label)}`),continuity_scene_key:txt(raw.continuity_scene_key||raw.scene_anchor_id),continuity_profile:txt(raw.continuity_profile),wardrobe_state:txt(raw.wardrobe_state)};}
  function tempScopeActive(e,line){if(!Number.isFinite(Number(line)))return true;const n=Number(line);return !(e.valid_from_line!==null&&n<e.valid_from_line)&&!(e.valid_to_line!==null&&n>e.valid_to_line);}
  function temporaryContinuityProfile(e={}){const identity=txt(e.identity_hint||e.label||e.source_phrase||'临时人物'),anchors=uniq([txt(e.age_stage_hint),txt(e.gender_hint),txt(e.species_hint),identity]).filter(Boolean).join('、');return `${anchors||'临时人物'}；同一continuity_id连续镜头保持同一体型、发型、服装主色与主要配饰，除非原文明示换装或时间线变化`;}
  function computeTemporary(source,lineNumber=null,context={}){
    const c=core(),s=txt(source),cache=clone(obj(context.continuityCache||c.temporary_continuity_cache)),formalNames=new Set(arr(c.slots).filter(x=>!x?.disabled).flatMap(slot=>uniq([slot.display_name,slot.base_name,slot.canonical_name_hint,...arr(slot.aliases)]))),rows=[];
    for(const raw of arr(c.mention_entities)){const e=normalizeTemp(raw);if(!e||e.confidence<TEMP_CONFIDENCE||!tempScopeActive(e,lineNumber))continue;const hit=uniq([e.label,e.source_phrase,...e.aliases]).sort((a,b)=>b.length-a.length).find(k=>k&&s.includes(k));if(!hit||formalNames.has(hit)||['mentioned_only','offscreen_voice'].includes(e.participation_state))continue;const id=e.continuity_id,n=Number.isFinite(Number(lineNumber))?Number(lineNumber):null,old=obj(cache[id]),gap=n!==null&&Number.isFinite(Number(old.last_line))?Math.abs(n-Number(old.last_line)):0,sceneChanged=Boolean(e.continuity_scene_key&&txt(old.scene_key)&&e.continuity_scene_key!==txt(old.scene_key)),expired=Boolean(old.continuity_id&&(sceneChanged||gap>TEMP_TTL)),base=expired?{}:old,profile=e.continuity_profile||txt(base.continuity_profile)||temporaryContinuityProfile(e),wardrobe=e.wardrobe_state||txt(base.wardrobe_state)||'沿用首次出镜确定的同一套简易服装、主色和主要配饰；原文明示换装时才更新';const next={continuity_id:id,scene_key:e.continuity_scene_key||txt(base.scene_key),first_line:base.first_line??n,last_line:n??base.last_line??null,appearance_count:Number(base.appearance_count||0)+1,continuity_profile:profile,wardrobe_state:wardrobe,gender_hint:e.gender_hint||txt(base.gender_hint),age_stage_hint:e.age_stage_hint||txt(base.age_stage_hint),species_hint:e.species_hint||txt(base.species_hint),identity_hint:e.identity_hint||txt(base.identity_hint)||e.label,updated_at:now()};cache[id]=next;rows.push({...e,matched_keyword:hit,must_render:true,continuity_scope:'scene_short_term',continuity_ttl_shots:TEMP_TTL,continuity_profile:profile,wardrobe_state:wardrobe,continuity_first_line:next.first_line,continuity_last_line:next.last_line});if(rows.length>=8)break;}
    const keys=Object.keys(cache);if(keys.length>80)keys.sort((a,b)=>txt(cache[a]?.updated_at).localeCompare(txt(cache[b]?.updated_at))).slice(0,keys.length-80).forEach(k=>delete cache[k]);
    return {temporary_characters:rows,continuity_cache:cache};
  }
  function applyManual(scene,automaticRows,c){const automaticIds=uniq(automaticRows.map(x=>x.slot_id)),manualAdded=uniq(scene.manual_added_slot_ids),manualExcluded=uniq(scene.manual_excluded_slot_ids);let added=manualAdded,excluded=manualExcluded;if(!added.length&&!excluded.length&&scene.character_core_manual_locked===true&&arr(scene.character_slot_ids).length){const legacyIds=uniq(scene.character_slot_ids);added=legacyIds.filter(id=>!automaticIds.includes(id));excluded=automaticIds.filter(id=>!legacyIds.includes(id));}const banned=new Set(excluded),effective=uniq([...automaticIds,...added]).filter(id=>!banned.has(id)),bySlot=new Map(automaticRows.map(r=>[r.slot_id,r]));for(const id of added){const slot=arr(c.slots).find(x=>txt(x.slot_id)===id);if(slot)bySlot.set(id,{slot_id:id,slot_token:txt(slot.slot_token),person_id:txt(slot.person_id),display_name:txt(slot.display_name||slot.base_name),scene_role:'visible',evidence:[{type:'manual_add',reason:'用户人工增加本镜人物'}],matched_keywords:[],keyword_evidence:[],relationship_evidence:[],confidence:1,selection_source:'manual_add'});}return {manual_added_slot_ids:added,manual_excluded_slot_ids:excluded,selected_slot_ids:effective,selected:effective.map(id=>bySlot.get(id)).filter(Boolean),manual_locked:Boolean(added.length||excluded.length)};}
  function computeScene(scene={},index=0,context={}){
    const c=core(),source=txt(scene.source_text||scene.text||scene.original_text),line=Number(scene.source_index||index+1),previous=obj(context.previousStageByPerson),formal=computeFormalCast(source,line,{keywordIndex:context.keywordIndex,previousStageByPerson:previous}),visible=formal.selected.filter(x=>!['offscreen_voice','mentioned_only'].includes(x.scene_role)),manual=applyManual(scene,visible,c),temp=computeTemporary(source,line,{continuityCache:context.continuityCache}),source_key=txt(scene.source_key)||sourceKey(source,line);return {record:{scene_id:txt(scene.id||`scene_${line}`),source_key,source_index:line,source_text:source,selected:clone(manual.selected),selected_slot_ids:clone(manual.selected_slot_ids),manual_added_slot_ids:clone(manual.manual_added_slot_ids),manual_excluded_slot_ids:clone(manual.manual_excluded_slot_ids),manual_locked:manual.manual_locked,keyword_evidence:manual.selected.flatMap(x=>arr(x.keyword_evidence)),relationship_evidence:manual.selected.flatMap(x=>arr(x.relationship_evidence)),ambiguous_matches:clone(formal.ambiguous),temporary_characters:clone(temp.temporary_characters),snapshot_frozen:false,compute_source:'v78_casting_compute_phase6'},previousStageByPerson:formal.previousStageByPerson,continuityCache:temp.continuity_cache};}
  function computeAll(inputScenes=null){const rows=arr(inputScenes||scenes()),keywordIndex=buildKeywordIndex(),previous={},continuityCache=clone(obj(core().temporary_continuity_cache)),records=[];let cache=continuityCache;rows.forEach((scene,index)=>{const result=computeScene(scene,index,{keywordIndex,previousStageByPerson:previous,continuityCache:cache});records.push(result.record);Object.assign(previous,result.previousStageByPerson);cache=result.continuityCache;});const snapshot={protocol:'v78_casting_compute_primary_phase6',created_at:now(),source_hash:txt(core().source_hash),keyword_total:keywordIndex.total,scenes:records,proposed_temporary_continuity_cache:cache};lastShadow=clone(snapshot);return clone(snapshot);}
  function keywordParityAgainst(productionIndex){
    const shadow=buildKeywordIndex(),prod=obj(productionIndex||{}),diffs=[];const keys=uniq([...Object.keys(prod),...Object.keys(shadow.index)]);
    for(const pid of keys){const a=uniq(arr(prod?.[pid]?.entries).map(x=>`${txt(x.keyword)}|${txt(x.slot_id)}`)).sort(),b=uniq(arr(shadow.index?.[pid]?.entries).map(x=>`${txt(x.keyword)}|${txt(x.slot_id)}`)).sort();if(JSON.stringify(a)!==JSON.stringify(b))diffs.push({person_id:pid,production:a,proposal:b});}
    return {ok:diffs.length===0,generated_at:now(),person_count:keys.length,mismatch_count:diffs.length,diffs,proposal:shadow.index,proposal_total:shadow.total};
  }
  function resolveKeywordProduction(productionIndex={}){
    ensureAuthorityContext();
    const report=keywordParityAgainst(productionIndex),wasPromoted=keywordAuthority.promoted;
    if(report.ok){keywordAuthority.safe_streak+=1;if(keywordAuthority.safe_streak>=PROMOTION_STREAK)keywordAuthority.promoted=true;}
    else{keywordAuthority.safe_streak=0;keywordAuthority.promoted=false;}
    if(!wasPromoted&&keywordAuthority.promoted){keywordAuthority.promotion_count+=1;keywordAuthority.last_changed_at=now();}
    if(wasPromoted&&!keywordAuthority.promoted){keywordAuthority.demotion_count+=1;keywordAuthority.last_changed_at=now();}
    keywordAuthority.last_parity=clone(report);
    keywordAuthority.last_executor=keywordAuthority.promoted?'v78_keyword_primary_phase6':report.ok?'character_core_keyword_warmup_validator':'character_core_keyword_fallback';
    const index=keywordAuthority.promoted?clone(report.proposal):clone(productionIndex);
    lastKeywordProduction=clone(productionIndex);
    console.info('[V78_KEYWORD_AUTHORITY]',{promoted:keywordAuthority.promoted,safe_streak:keywordAuthority.safe_streak,threshold:PROMOTION_STREAK,executor:keywordAuthority.last_executor,mismatch_count:report.mismatch_count});
    return {index,executor:keywordAuthority.last_executor,promoted:keywordAuthority.promoted,safe_streak:keywordAuthority.safe_streak,threshold:PROMOTION_STREAK,parity:clone(report),decided_at:now()};
  }
  function keywordAuthoritySnapshot(){return {service:'v78_keyword_authority_phase6',threshold:PROMOTION_STREAK,...clone(keywordAuthority)};}
  function ids(row){return uniq(row?.selected_slot_ids||arr(row?.selected).map(x=>x.slot_id)).sort();}
  function compareSnapshots(production,shadow){const p=arr(production?.scenes),s=arr(shadow?.scenes),pm=new Map(p.map(r=>[txt(r.source_key),r])),sm=new Map(s.map(r=>[txt(r.source_key),r])),keys=uniq([...pm.keys(),...sm.keys()]),diffs=[];for(const key of keys){const a=pm.get(key),b=sm.get(key),ai=ids(a),bi=ids(b),same=JSON.stringify(ai)===JSON.stringify(bi),at=uniq(arr(a?.temporary_characters).map(x=>x.continuity_id||x.entity_id||x.label)).sort(),bt=uniq(arr(b?.temporary_characters).map(x=>x.continuity_id||x.entity_id||x.label)).sort(),tempSame=JSON.stringify(at)===JSON.stringify(bt);if(!same||!tempSame||!a||!b)diffs.push({source_key:key,production_slot_ids:ai,shadow_slot_ids:bi,production_temp_ids:at,shadow_temp_ids:bt,formal_match:same,temporary_match:tempSame,missing_production:!a,missing_shadow:!b});}return {ok:diffs.length===0,service:'casting_parity_v78_phase6',generated_at:now(),scene_count:keys.length,match_count:keys.length-diffs.length,mismatch_count:diffs.length,diffs};}
  function orderedIds(row){return uniq(row?.selected_slot_ids||arr(row?.selected).map(x=>x.slot_id));}
  function sortedIds(values){return uniq(values).sort();}
  function tempIds(row){return uniq(arr(row?.temporary_characters).map(x=>txt(x.continuity_id||x.entity_id||x.label)));}
  function sameOrdered(a,b){return JSON.stringify(arr(a).map(txt))===JSON.stringify(arr(b).map(txt));}
  function sameSet(a,b){return JSON.stringify(sortedIds(a))===JSON.stringify(sortedIds(b));}

  function canonicalEvidenceTokens(record={}){
    const tokens=[];
    for(const row of arr(record.selected)){
      const sid=txt(row.slot_id),role=txt(row.scene_role||'visible');
      tokens.push(`S|${sid}|${role}`);
      for(const keyword of uniq(row.matched_keywords).sort())tokens.push(`M|${sid}|${keyword}`);
      for(const e of arr(row.keyword_evidence)){tokens.push(`K|${sid}|${txt(e.keyword)}|${txt(e.binding_id)}|${txt(e.relation_id)}`);}
      for(const e of arr(row.relationship_evidence)){tokens.push(`R|${sid}|${txt(e.keyword)}|${txt(e.binding_id)}|${txt(e.relation_id)}`);}
      for(const e of arr(row.evidence)){const type=txt(e.type);if(type==='manual_add'||type==='manual')tokens.push(`U|${sid}|${type}`);}
    }
    return uniq(tokens).sort();
  }
  function evidenceParity(production={},shadow={}){
    const a=canonicalEvidenceTokens(production),b=canonicalEvidenceTokens(shadow),ok=JSON.stringify(a)===JSON.stringify(b);
    return {ok,service:'scene_cast_evidence_parity_v78_phase6',production_tokens:a,proposal_tokens:b,mismatch_count:ok?0:1,generated_at:now()};
  }
  function resolveEvidenceProduction(sourceKeyValue,production={},shadow={},castSafe=true){
    ensureAuthorityContext();const key=txt(sourceKeyValue)||txt(production.source_key)||txt(shadow.source_key)||'scene_unknown',report=evidenceParity(production,shadow),previous=obj(evidenceAuthority.scene_authority[key]),wasPromoted=Boolean(previous.promoted);let safeStreak=Number(previous.safe_streak||0),promoted=wasPromoted;
    const safe=Boolean(gateEnabled&&castSafe&&report.ok);
    if(safe){safeStreak+=1;if(!promoted&&safeStreak>=PROMOTION_STREAK){promoted=true;evidenceAuthority.promotion_count+=1;}}
    else{safeStreak=0;if(promoted){promoted=false;evidenceAuthority.demotion_count+=1;}}
    const takeover=Boolean(safe&&promoted),executor=takeover?'v78_scene_cast_evidence_primary_phase6':safe?'character_core_evidence_warmup_validator':'character_core_evidence_fallback';
    const decision={source_key:key,safe_streak:safeStreak,promoted,takeover,executor,threshold:PROMOTION_STREAK,parity:clone(report),updated_at:now()};
    evidenceAuthority.scene_authority[key]=clone(decision);evidenceAuthority.last_decision=clone(decision);if(takeover)evidenceAuthority.takeover_count+=1;else evidenceAuthority.fallback_count+=1;
    return decision;
  }
  function evidenceAuthoritySnapshot(){const rows=Object.values(evidenceAuthority.scene_authority);return {service:'scene_cast_evidence_authority_v78_phase6',threshold:PROMOTION_STREAK,promotion_count:evidenceAuthority.promotion_count,demotion_count:evidenceAuthority.demotion_count,takeover_count:evidenceAuthority.takeover_count,fallback_count:evidenceAuthority.fallback_count,promoted_scene_count:rows.filter(x=>x.promoted).length,last_decision:clone(evidenceAuthority.last_decision),scene_authority:clone(evidenceAuthority.scene_authority),authority_signature:evidenceAuthority.authority_signature};}

  function canonicalTempRow(row={}){return {continuity_id:txt(row.continuity_id||row.entity_id||row.label),label:txt(row.label),source_phrase:txt(row.source_phrase||row.label),participation_state:txt(row.participation_state),continuity_scope:txt(row.continuity_scope),continuity_ttl_shots:Number(row.continuity_ttl_shots||TEMP_TTL),continuity_profile:txt(row.continuity_profile),wardrobe_state:txt(row.wardrobe_state),continuity_first_line:row.continuity_first_line??null,continuity_last_line:row.continuity_last_line??null,gender_hint:txt(row.gender_hint),age_stage_hint:txt(row.age_stage_hint),species_hint:txt(row.species_hint),identity_hint:txt(row.identity_hint)};}
  function temporarySceneParity(production={},shadow={}){const a=arr(production.temporary_characters).map(canonicalTempRow).sort((x,y)=>x.continuity_id.localeCompare(y.continuity_id)),b=arr(shadow.temporary_characters).map(canonicalTempRow).sort((x,y)=>x.continuity_id.localeCompare(y.continuity_id)),ok=JSON.stringify(a)===JSON.stringify(b);return {ok,service:'temporary_scene_parity_v78_phase6',production:a,proposal:b,mismatch_count:ok?0:1,generated_at:now()};}
  function canonicalContinuityCache(cache={}){const out={};for(const key of Object.keys(obj(cache)).sort()){const row=obj(cache[key]);out[key]={continuity_id:txt(row.continuity_id||key),scene_key:txt(row.scene_key),first_line:row.first_line??null,last_line:row.last_line??null,appearance_count:Number(row.appearance_count||0),continuity_profile:txt(row.continuity_profile),wardrobe_state:txt(row.wardrobe_state),gender_hint:txt(row.gender_hint),age_stage_hint:txt(row.age_stage_hint),species_hint:txt(row.species_hint),identity_hint:txt(row.identity_hint)};}return out;}
  function temporaryCacheParity(characterCoreCache={},v78Cache={}){const a=canonicalContinuityCache(characterCoreCache),b=canonicalContinuityCache(v78Cache),ok=JSON.stringify(a)===JSON.stringify(b);return {ok,service:'temporary_continuity_cache_parity_v78_phase6',mismatch_count:ok?0:1,production:a,proposal:b,generated_at:now()};}
  function resolveTemporaryCacheProduction(characterCoreCache={},v78Cache={}){
    ensureAuthorityContext();const report=temporaryCacheParity(characterCoreCache,v78Cache),wasPromoted=temporaryAuthority.promoted;
    if(gateEnabled&&report.ok){temporaryAuthority.safe_streak+=1;if(temporaryAuthority.safe_streak>=PROMOTION_STREAK)temporaryAuthority.promoted=true;}else{temporaryAuthority.safe_streak=0;temporaryAuthority.promoted=false;}
    if(!wasPromoted&&temporaryAuthority.promoted){temporaryAuthority.promotion_count+=1;temporaryAuthority.last_changed_at=now();}
    if(wasPromoted&&!temporaryAuthority.promoted){temporaryAuthority.demotion_count+=1;temporaryAuthority.last_changed_at=now();}
    temporaryAuthority.last_parity=clone(report);temporaryAuthority.validator_cache=clone(characterCoreCache);temporaryAuthority.primary_cache=clone(v78Cache);temporaryAuthority.last_executor=temporaryAuthority.promoted?'v78_temporary_continuity_primary_phase6':report.ok?'character_core_temporary_warmup_validator':'character_core_temporary_fallback';
    const cache=temporaryAuthority.promoted?clone(v78Cache):clone(characterCoreCache);
    console.info('[V78_TEMPORARY_AUTHORITY]',{promoted:temporaryAuthority.promoted,safe_streak:temporaryAuthority.safe_streak,threshold:PROMOTION_STREAK,executor:temporaryAuthority.last_executor,mismatch_count:report.mismatch_count});
    return {cache,executor:temporaryAuthority.last_executor,promoted:temporaryAuthority.promoted,safe_streak:temporaryAuthority.safe_streak,threshold:PROMOTION_STREAK,parity:clone(report),decided_at:now()};
  }
  function temporaryAuthoritySnapshot(){return {service:'temporary_continuity_authority_v78_phase6',threshold:PROMOTION_STREAK,...clone(temporaryAuthority)};}
  function trustedManual(scene={},record={}){
    const added=uniq(record.manual_added_slot_ids),excluded=uniq(record.manual_excluded_slot_ids);
    if(!added.length&&!excluded.length)return {ok:true,reason:'no_manual_override'};
    const meta=obj(scene.manual_cast_override_meta_v19),trusted=Boolean(txt(meta.upgraded_at)||arr(meta.added).length||arr(meta.excluded).length);
    if(!trusted)return {ok:false,reason:'manual_override_without_v19_evidence'};
    const metaAdded=uniq(arr(meta.added).map(x=>x?.slot_id)),metaExcluded=uniq(arr(meta.excluded).map(x=>x?.slot_id));
    if(!sameSet(added,metaAdded)||!sameSet(excluded,metaExcluded))return {ok:false,reason:'manual_override_evidence_mismatch'};
    return {ok:true,reason:'manual_override_evidence_trusted'};
  }
  function sceneConflictReasons(source,shadow={}){
    const reasons=[];
    if(arr(shadow.ambiguous_matches).length)reasons.push('ambiguous_keyword_match');
    try{
      const diag=globalThis.__V78_CHARACTER_SERVICE__?.diagnostics?.()||{};
      if(Number(diag?.summary?.error||0)>0)reasons.push('character_structure_error');
      const selectedPersons=new Set(arr(shadow.selected).map(x=>txt(x.person_id)).filter(Boolean));
      for(const conflict of arr(diag.relationship_conflicts)){
        const raw=JSON.stringify(conflict||{});
        const touchesPerson=[...selectedPersons].some(pid=>pid&&raw.includes(pid));
        const touchesAlias=txt(conflict?.alias)&&txt(source).includes(txt(conflict.alias));
        if(touchesPerson||touchesAlias){reasons.push('relationship_or_alias_conflict');break;}
      }
    }catch(_){reasons.push('character_diagnostics_unavailable');}
    return uniq(reasons);
  }
  function gateDecision(scene={},production={},shadow={}){
    const reasons=[];
    if(!gateEnabled)reasons.push('gate_disabled');
    const pKey=txt(production.source_key),sKey=txt(shadow.source_key);
    if(!pKey||!sKey||pKey!==sKey)reasons.push('source_key_mismatch');
    const pIds=orderedIds(production),sIds=orderedIds(shadow);
    // Order is intentionally part of parity: takeover must not even reorder the final character block.
    if(!sameOrdered(pIds,sIds))reasons.push('selected_slot_order_mismatch');
    if(!sameSet(production.manual_added_slot_ids,shadow.manual_added_slot_ids)||!sameSet(production.manual_excluded_slot_ids,shadow.manual_excluded_slot_ids))reasons.push('manual_override_mismatch');
    if(!sameOrdered(tempIds(production),tempIds(shadow)))reasons.push('temporary_character_mismatch');
    const manual=trustedManual(scene,production);if(!manual.ok)reasons.push(manual.reason);
    reasons.push(...sceneConflictReasons(production.source_text||scene.source_text||scene.text,shadow));
    const c=core(),validSlots=new Set(arr(c.slots).filter(x=>!x?.disabled).map(x=>txt(x.slot_id)));
    if(sIds.some(id=>!validSlots.has(id)))reasons.push('invalid_or_disabled_slot');
    const decision={takeover:reasons.length===0,executor:reasons.length===0?'v78_casting_candidate_phase6':'character_core_fallback',source_key:pKey||sKey,reasons:uniq(reasons),manual_reason:manual.reason,production_slot_ids:pIds,proposal_slot_ids:sIds,production_temp_ids:tempIds(production),proposal_temp_ids:tempIds(shadow),decided_at:now()};
    return decision;
  }
  function resolveProductionScene(scene={},index=0,context={}){
    ensureAuthorityContext();
    const production=clone(context.productionRecord||{}),computed=computeScene(scene,index,{keywordIndex:context.keywordIndex||buildKeywordIndex(),previousStageByPerson:clone(obj(context.previousStageByPerson)),continuityCache:clone(obj(context.continuityCache))}),shadow=computed.record;
    const safety=gateDecision(scene,production,shadow),key=txt(safety.source_key)||`scene_${index+1}`,previous=obj(gateState.scene_authority[key]);
    let safeStreak=Number(previous.safe_streak||0),promoted=Boolean(previous.promoted),demoted=false,promotedNow=false;
    if(safety.takeover){safeStreak+=1;if(!promoted&&safeStreak>=PROMOTION_STREAK){promoted=true;promotedNow=true;gateState.promotion_count+=1;}}
    else{if(promoted){demoted=true;gateState.demotion_count+=1;}promoted=false;safeStreak=0;}
    if(!gateEnabled){if(promoted){demoted=true;gateState.demotion_count+=1;}promoted=false;safeStreak=0;}
    const takeover=Boolean(safety.takeover&&promoted&&gateEnabled);
    const executor=takeover?'v78_casting_primary_phase6':safety.takeover?'character_core_warmup_validator':'character_core_fallback';
    const reasons=clone(safety.reasons);if(safety.takeover&&!takeover)reasons.push(`promotion_warmup_${safeStreak}_of_${PROMOTION_STREAK}`);
    const authority={source_key:key,safe_streak:safeStreak,promoted,takeover,executor,threshold:PROMOTION_STREAK,promoted_now:promotedNow,demoted_now:demoted,reasons:uniq(reasons),last_safe:Boolean(safety.takeover),updated_at:now()};
    gateState.scene_authority[key]=clone(authority);
    const evidence=resolveEvidenceProduction(key,production,shadow,Boolean(safety.takeover));
    const tempParity=temporarySceneParity(production,shadow),temporaryTakeover=Boolean(gateEnabled&&takeover&&temporaryAuthority.promoted&&tempParity.ok);
    const temporaryExecutor=temporaryTakeover?'v78_temporary_scene_primary_phase6':temporaryAuthority.promoted&&!tempParity.ok?'character_core_temporary_scene_fallback':temporaryAuthority.promoted?'character_core_temporary_scene_validator':'character_core_temporary_warmup_validator';
    const decision={...safety,...authority,takeover,executor,reasons:authority.reasons,evidence_takeover:Boolean(evidence.takeover&&takeover),evidence_executor:evidence.takeover&&takeover?evidence.executor:'character_core_evidence_fallback',evidence_authority:clone(evidence),temporary_takeover:temporaryTakeover,temporary_executor:temporaryExecutor,temporary_scene_parity:clone(tempParity),decided_at:now()};
    gateState.decisions[key]=clone(decision);gateState.last_decision=clone(decision);if(takeover)gateState.takeover_count+=1;else if(safety.takeover)gateState.warmup_count+=1;else gateState.fallback_count+=1;
    console.info('[V78_CAST_AUTHORITY]',{source_key:key,takeover,promoted,safe_streak:safeStreak,threshold:PROMOTION_STREAK,executor,reasons:decision.reasons,evidence_takeover:decision.evidence_takeover,temporary_takeover:decision.temporary_takeover});
    return { ...decision, record:clone(shadow), proposed_continuity_cache:clone(computed.continuityCache), previousStageByPerson:clone(computed.previousStageByPerson) };
  }
  function gateSnapshot(){const decisions=Object.values(gateState.decisions),authorities=Object.values(gateState.scene_authority);return {service:'casting_production_authority_v78_phase6',enabled:gateEnabled,authoritative_mode:'promoted_primary_with_live_character_core_fallback',promotion_threshold:PROMOTION_STREAK,takeover_count:gateState.takeover_count,warmup_count:gateState.warmup_count,fallback_count:gateState.fallback_count,promotion_count:gateState.promotion_count,demotion_count:gateState.demotion_count,promoted_scene_count:authorities.filter(x=>x.promoted).length,scene_count:decisions.length,last_decision:clone(gateState.last_decision),decisions:clone(decisions),scene_authority:clone(authorities),last_reset_at:gateState.last_reset_at,authority_signature:gateState.authority_signature,keyword_authority:keywordAuthoritySnapshot(),evidence_authority:evidenceAuthoritySnapshot(),temporary_authority:temporaryAuthoritySnapshot()};}
  function setGateEnabled(value){const next=Boolean(value);if(gateEnabled!==next){gateEnabled=next;if(!next){Object.values(gateState.scene_authority).forEach(row=>{row.promoted=false;row.safe_streak=0;row.takeover=false;row.executor='character_core_fallback';});Object.values(evidenceAuthority.scene_authority).forEach(row=>{row.promoted=false;row.safe_streak=0;row.takeover=false;row.executor='character_core_evidence_fallback';});keywordAuthority.promoted=false;keywordAuthority.safe_streak=0;keywordAuthority.last_executor='character_core_keyword_fallback';temporaryAuthority.promoted=false;temporaryAuthority.safe_streak=0;temporaryAuthority.last_executor='character_core_temporary_fallback';gateState.demotion_count+=1;keywordAuthority.demotion_count+=1;evidenceAuthority.demotion_count+=1;temporaryAuthority.demotion_count+=1;}gateState.last_reset_at=now();}return gateSnapshot();}
  function resetGateStats(){resetAuthority('manual_reset');return gateSnapshot();}
  function authoritySnapshot(){ensureAuthorityContext();return {service:'casting_authority_v78_phase6',promotion_threshold:PROMOTION_STREAK,gate:gateSnapshot(),keyword:keywordAuthoritySnapshot(),evidence:evidenceAuthoritySnapshot(),temporary:temporaryAuthoritySnapshot()};}
  function productionSnapshot(){if(lastProduction)return clone(lastProduction);const c=core();return clone(c.outline_cast_snapshot||{scenes:Object.values(obj(c.scene_casting))});}
  function observeProduction(snapshot,meta={}){lastProduction=clone(snapshot||{});const shadow=computeAll();lastParity=compareSnapshots(lastProduction,shadow);lastParity.production_meta=clone(meta);console.info('[V78_CAST_PARITY]',{scene_count:lastParity.scene_count,mismatch_count:lastParity.mismatch_count});return clone(lastParity);}
  function observeKeywordProduction(index){return resolveKeywordProduction(index||{}).parity;}
  function keywordParity(){return keywordParityAgainst(lastKeywordProduction||core().keyword_index_character_core_validator||core().keyword_index);}
  function parity(){const production=productionSnapshot(),shadow=computeAll(),report=compareSnapshots(production,shadow);report.keyword=keywordParity();lastParity=clone(report);return report;}
  function snapshot(){return {service:'casting_compute_service_v78_phase6',phase:'v78_phase6',authoritative:'promoted_primary_with_continuity_and_evidence',shadow_only:false,gate_enabled:gateEnabled,promotion_threshold:PROMOTION_STREAK,generated_at:now(),keyword:buildKeywordIndex(),keyword_authority:keywordAuthoritySnapshot(),evidence_authority:evidenceAuthoritySnapshot(),temporary_authority:temporaryAuthoritySnapshot(),shadow:lastShadow||computeAll(),production:productionSnapshot(),parity:lastParity||parity(),gate:gateSnapshot()};}
  const service={version:'casting_compute_service_v78_phase6',phase:'v78_phase6',authoritative:'promoted_primary_with_continuity_and_evidence',shadow_only:false,gate_mode:true,promotion_threshold:PROMOTION_STREAK,buildKeywordIndex,matchKeywords,computeFormalCast,computeTemporary,computeScene,computeAll,compareSnapshots,observeProduction,observeKeywordProduction,resolveKeywordProduction,keywordAuthoritySnapshot,evidenceParity,resolveEvidenceProduction,evidenceAuthoritySnapshot,temporarySceneParity,temporaryCacheParity,resolveTemporaryCacheProduction,temporaryAuthoritySnapshot,productionSnapshot,keywordParity,parity,gateDecision,resolveProductionScene,gateSnapshot,authoritySnapshot,setGateEnabled,resetGateStats,snapshot};
  ns.casting=service;globalThis.__V78_CASTING_SERVICE__=service;
  console.info('[V78 Clean Core] casting phase6 authority service ready',{version:service.version,authoritative:service.authoritative,promotion_threshold:PROMOTION_STREAK,gate_mode:true,continuity_primary:true,evidence_primary:true});
})();

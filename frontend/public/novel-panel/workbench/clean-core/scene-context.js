/* V78.2.7 Clean Core: text-only Scene Context / AI Submission Authority.
 * V78 builds the candidate context package; CharacterCore semantics remain an independent validator/fallback.
 * Promotion requires 3 consecutive full-semantic parity passes. Any mismatch demotes immediately.
 * This service still does NOT own the AI generator, scene writeback, or final output formatting.
 */
(function installV78SceneContextService(){
  if(globalThis.__V78_SCENE_CONTEXT_SERVICE__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const arr=v=>Array.isArray(v)?v:[];
  const obj=v=>v&&typeof v==='object'&&!Array.isArray(v)?v:{};
  const txt=v=>String(v??'').trim();
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const uniq=v=>[...new Set(arr(v).map(txt).filter(Boolean))];
  const now=()=>new Date().toISOString();
  const PROMOTION_STREAK=3;
  const clip=(v,n=1200)=>String(v??'').slice(0,n);
  function rootState(){try{return (typeof state!=='undefined'?state:globalThis.state)||{};}catch(_){return globalThis.state||{};}}
  function core(){
    const s=rootState();
    return obj(s.characterCoreV2||s.character_core_v2||globalThis.__characterCoreV2Api?.getState?.()||globalThis.__characterCoreV2Api?.state?.());
  }
  function scenes(){return arr(rootState().scenes);}
  function sourceKey(scene={},index=0){
    const direct=txt(scene.source_key);if(direct)return direct;
    const source=txt(scene.source_text||scene.text||scene.original_text);let h=2166136261;for(const ch of source){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return `${Number(scene.source_index||index+1)}_${(h>>>0).toString(36)}`;
  }
  function stage(slot={}){return txt(slot?.age?.visual_age_stage||slot?.current_values?.visual_age_stage||slot?.visual_age_stage||slot?.stage_label);}
  function slotById(id){return arr(core().slots).find(x=>txt(x.slot_id)===txt(id))||null;}
  function personById(id){return arr(core().people).find(x=>txt(x.person_id)===txt(id))||null;}
  function enrichSelected(row={}){
    const slot=slotById(row.slot_id)||{},person=personById(slot.person_id||row.person_id)||{};
    return {
      slot_id:txt(row.slot_id||slot.slot_id),slot_token:txt(row.slot_token||slot.slot_token),person_id:txt(row.person_id||slot.person_id),
      display_name:txt(row.display_name||slot.display_name||slot.base_name||person.primary_display_name||person.canonical_name),
      scene_role:txt(row.scene_role||'visible'),gender:txt(row.gender||slot.gender),visual_age_stage:txt(row.visual_age_stage||stage(slot)),
      chronological_age:txt(row.chronological_age||slot?.age?.chronological_age),species:txt(row.species||slot.species||person.species),
      appearance:clip(row.appearance||slot.appearance,1800),matched_keywords:uniq(row.matched_keywords).slice(0,10),
      keyword_evidence:clone(arr(row.keyword_evidence).slice(0,8)),relationship_evidence:clone(arr(row.relationship_evidence).slice(0,8)),
      evidence:clone(arr(row.evidence).slice(0,12)),selection_source:txt(row.selection_source)
    };
  }
  function compactTemporary(item={}){
    return {temporary_entity_id:txt(item.temporary_entity_id||item.entity_id||item.continuity_id),continuity_id:txt(item.continuity_id||item.entity_id),entity_id:txt(item.entity_id),label:txt(item.label),source_phrase:txt(item.source_phrase||item.label),entity_type:txt(item.entity_type||'temporary_character'),count:Math.max(1,Number(item.count||1)),participation_state:txt(item.participation_state||'visible_candidate'),must_render:item.must_render===true,identity_hint:txt(item.identity_hint),gender_hint:txt(item.gender_hint),age_stage_hint:txt(item.age_stage_hint),species_hint:txt(item.species_hint),continuity_profile:clone(obj(item.continuity_profile)),wardrobe_state:clone(obj(item.wardrobe_state)),scene_key:txt(item.scene_key),first_line:Number(item.first_line||0)||null,last_line:Number(item.last_line||0)||null};
  }
  const TEMP_ACTION_RE=/(走|跑|冲|站|坐|躺|蹲|跪|抬|低头|转身|回头|伸手|拿|放|递|接|推|拉|打开|关上|查看|翻|盯|看向|望向|扫视|点头|摇头|皱眉|微笑|哭|笑|喊|说|问|答|扶|抱|抓|握|敲|按|掏出|靠近|后退|停下|离开|进入|端着|弯腰|俯身|起身|抬眼|垂眼|示意|拦住|挡住|追上|避开)/g;
  function tempAliases(item={}){return uniq([item.label,item.source_phrase,item.name,...arr(item.aliases)]).filter(x=>x.length>1||!/^[我你他她它]$/.test(x)).sort((a,b)=>b.length-a.length);}
  function sourceClauseForTemp(source='',item={}){const raw=txt(source);if(!raw)return '';let pos=-1,hit='';for(const a of tempAliases(item)){const p=raw.indexOf(a);if(p>=0){pos=p;hit=a;break;}}if(pos<0)return clip(raw,220);let start=pos,end=pos+hit.length;while(start>0&&!/[。！？；\n]/.test(raw[start-1]))start--;while(end<raw.length&&!/[。！？；\n]/.test(raw[end]))end++;return clip(raw.slice(start,Math.min(raw.length,end+1)).trim(),300);}
  function compactStaticValue(v,limit=220){if(typeof v==='string')return clip(v,limit);if(v&&typeof v==='object'){try{return clip(JSON.stringify(v),limit);}catch(_){return '';}}return '';}
  function performanceForTemp(source,item={}){const evidence=sourceClauseForTemp(source,item),actions=uniq((evidence.match(TEMP_ACTION_RE)||[])).slice(0,8);let role='反应/陪衬人物';if(/说|问|答|喊|质问|回应/.test(evidence))role='对话人物';else if(actions.length)role='主动作人物';if(Math.max(1,Number(item.count||1))>1||/众|人群|家人|姐妹|兄弟|同学|宾客|士兵|侍卫/.test(txt(item.label)))role='群体人物';return {source_evidence:evidence,action_keywords:actions,performance_role:role,requires_dynamic_render:item.must_render===true};}

  function sceneMemoryFor(plan={},scene={}){
    const direct=obj(plan.scene_memory);if(Object.keys(direct).length)return clone(direct);
    const sceneDirect=obj(scene.scene_memory);if(Object.keys(sceneDirect).length)return clone(sceneDirect);
    try{const found=globalThis.__V77_SCENE_MEMORY__?.memoryForSourceKey?.(txt(plan.source_key||scene.source_key));return clone(obj(found));}catch(_){return {};}
  }
  function visualText(scene={}){
    const shots=arr(scene.outline_shots||scene.shots);let value=shots.length?shots.map(s=>txt(s.prompt||s.visual_prompt||s.description)).filter(Boolean).join(' | '):txt(scene.prompt||scene.visual_prompt||scene.visual_context||'');
    try{value=globalThis.__V78_MASTER_PROMPT_SERVICE__?.stripReferenceTokens?.(value)||value;}catch(_){}
    return clip(value,900);
  }
  function adjacencyFor(index,allScenes){
    const make=(scene,i)=>scene?{source_key:sourceKey(scene,i),source_index:Number(scene.source_index||i+1),source_text:clip(scene.source_text||scene.text||scene.original_text,700),visual_reference:visualText(scene),character_slot_ids:uniq(scene.character_slot_ids)}:null;
    return {previous:make(allScenes[index-1],index-1),next:make(allScenes[index+1],index+1)};
  }
  function relationshipContextFor(selectedRows=[]){
    const ids=new Set(arr(selectedRows).map(x=>txt(x.person_id)).filter(Boolean));
    return arr(core().relationships).filter(r=>!r?.disabled&&ids.has(txt(r.source_person_id))&&ids.has(txt(r.target_person_id))).map(r=>({relation_id:txt(r.relation_id),source_person_id:txt(r.source_person_id),target_person_id:txt(r.target_person_id),relation_type:txt(r.relation_type||r.type||r.relationship),label:txt(r.label||r.display_name),manual_locked:Boolean(r.manual_locked),evidence:clip(r.evidence,300)}));
  }
  function buildSceneRecord(plan={},index=0,allScenes=scenes()){
    const key=txt(plan.source_key)||sourceKey(plan,index),scene=allScenes.find(s=>txt(s.source_key)===key)||allScenes[index]||{};
    const selectedRaw=arr(plan.character_packages).length?arr(plan.character_packages):arr(scene.character_core_cast).filter(row=>uniq(scene.character_slot_ids).includes(txt(row.slot_id)));
    const selected=selectedRaw.map(enrichSelected).filter(x=>x.slot_id);
    const selectedIds=selected.map(x=>x.slot_id),manualAdded=uniq(plan.manual_added_slot_ids||scene.manual_added_slot_ids),manualExcluded=uniq(plan.manual_excluded_slot_ids||scene.manual_excluded_slot_ids);
    const temporary=arr(plan.temporary_characters).length?arr(plan.temporary_characters):arr(scene.temporary_characters);
    const manualTemporary=[...arr(plan.manual_temporary_characters_v7824),...arr(scene.manual_temporary_characters_v7824)].map(compactTemporary);
    const ctx=obj(plan.character_core_context);
    return {
      source_key:key,source_index:Number(plan.line_index||scene.source_index||index+1),source_text:clip(plan.source_text||scene.source_text||scene.text,1800),
      scene_anchor_id:txt(plan.scene_anchor_id||scene.scene_anchor_id||plan.temporary_scene_group||scene.temporary_scene_group),scene_memory:sceneMemoryFor(plan,scene),director_decision:clone(obj(plan.director_decision||scene.director_decision)),
      manual_locked:plan.characters_mode==='manual'||scene.characters_mode==='manual'||ctx.manual_locked===true,
      selected_characters:selected,selected_slot_ids:selectedIds,manual_added_slot_ids:manualAdded,manual_excluded_slot_ids:manualExcluded,
      offscreen_voice:arr(ctx.offscreen_voice).map(enrichSelected).filter(x=>x.slot_id),mentioned_only:arr(ctx.mentioned_only).map(enrichSelected).filter(x=>x.slot_id),
      temporary_characters:temporary.map(compactTemporary),manual_temporary_characters_v7824:manualTemporary,crowd_characters:clone(arr(plan.crowd_characters||scene.crowd_characters)),
      relationship_context:relationshipContextFor(selected),keyword_evidence:clone(arr(plan.keyword_evidence||scene.keyword_evidence).slice(0,30)),relationship_evidence:clone(arr(plan.relationship_evidence||scene.relationship_evidence).slice(0,30)),
      excluded_formal_slot_ids:uniq(plan.excluded_formal_slot_ids),adjacent_context:adjacencyFor(index,allScenes),
      cast_authority:{executor:txt(scene.cast_executor_v78||scene.cast_snapshot_source),evidence_executor:txt(scene.cast_evidence_executor_v78),temporary_executor:txt(scene.temporary_executor_v78)}
    };
  }
  function characterRegistry(records=[]){
    const ids=uniq(records.flatMap(r=>[...r.selected_slot_ids,...arr(r.offscreen_voice).map(x=>x.slot_id),...arr(r.mentioned_only).map(x=>x.slot_id)]));
    return ids.map(id=>{const slot=slotById(id);if(!slot)return null;const person=personById(slot.person_id)||{};return {slot_id:txt(slot.slot_id),slot_token:txt(slot.slot_token),person_id:txt(slot.person_id),display_name:txt(slot.display_name||slot.base_name),canonical_name:txt(person.canonical_name),gender:txt(slot.gender),visual_age_stage:stage(slot),chronological_age:txt(slot?.age?.chronological_age),species:txt(slot.species||person.species),appearance:clip(slot.appearance,1800)};}).filter(Boolean);
  }
  function validate(pkg={}){
    const issues=[],warnings=[],rows=arr(pkg.scenes),keys=new Set(),validSlots=new Set(arr(core().slots).filter(x=>!x?.disabled).map(x=>txt(x.slot_id)));
    rows.forEach((r,i)=>{
      if(!txt(r.source_key))issues.push({code:'missing_source_key',scene:i+1});
      else if(keys.has(txt(r.source_key)))issues.push({code:'duplicate_source_key',source_key:txt(r.source_key)});else keys.add(txt(r.source_key));
      arr(r.selected_slot_ids).forEach(id=>{if(!validSlots.has(txt(id)))issues.push({code:'invalid_selected_slot',source_key:r.source_key,slot_id:id});});
      const selected=new Set(arr(r.selected_slot_ids).map(txt));arr(r.manual_excluded_slot_ids).forEach(id=>{if(selected.has(txt(id)))issues.push({code:'manual_excluded_still_selected',source_key:r.source_key,slot_id:id});});
      arr(r.temporary_characters).forEach(t=>{if(t.must_render&&['offscreen_voice','mentioned_only'].includes(txt(t.participation_state)))warnings.push({code:'temporary_visibility_conflict',source_key:r.source_key,label:t.label});});
      if(r.scene_anchor_id&&!Object.keys(obj(r.scene_memory)).length)warnings.push({code:'scene_memory_missing',source_key:r.source_key,scene_anchor_id:r.scene_anchor_id});
    });
    return {ok:issues.length===0,issues,warnings,scene_count:rows.length,source_key_count:keys.size};
  }
  function buildOutlinePackage(plans=[]){
    const allScenes=scenes(),records=arr(plans).map((p,i)=>buildSceneRecord(p,i,allScenes));
    const pkg={protocol:'v78_scene_context_submission_phase8',version:'scene_context_service_v78_2_4',created_at:now(),source_hash:txt(core().source_hash),character_revision:Number(core().character_revision||0),authority:'v78_submission_promoted_primary_with_character_core_validator_fallback',character_registry:characterRegistry(records),scenes:records,requirements:{formal_characters:'selected_slot_ids are frozen before AI; AI must not add/remove/replace formal slots',temporary_characters:'AI must decide temporary people/groups directly from each source_text. Local heuristic candidates are validator-only and must not force casting. Only manual_temporary_characters_v7824 is a user-forced supplement. Temporary people are text-only performers.',scene_memory:'same scene_anchor_id preserves spatial geometry, fixed elements, light direction and color temperature unless source explicitly changes place/time',camera_language:'shot size, angle, movement and transition must remain Chinese-only',internal_metadata:'source_key, stage, evidence and continuity fields are internal; do not print field names or rule explanations'}};
    pkg.validation=validate(pkg);lastOutline=clone(pkg);return clone(pkg);
  }
  function aiSceneView(row={},options={}){
    const mode=txt(options.mode||"outline");
    const manual=arr(row.manual_temporary_characters_v7824).map(t=>({label:txt(t.label||t.source_phrase),count:Math.max(1,Number(t.count||1)),participation_state:txt(t.participation_state||"visible_candidate"),manual_added:true}));
    const priorAi=mode==="regeneration"?arr(row.temporary_characters).filter(t=>t?.manual_added_v7824!==true&&(t?.ai_discovered===true||txt(t?.selection_source_v7824)==="ai_primary")).map(t=>({label:txt(t.label||t.source_phrase),count:Math.max(1,Number(t.count||1)),participation_state:txt(t.participation_state),identity_hint:txt(t.identity_hint),gender_hint:txt(t.gender_hint),age_stage_hint:txt(t.age_stage_hint),previous_ai_result:true,current_performance:performanceForTemp(row.source_text,t)})):[];
    const compactSelected=arr(row.selected_characters).map(x=>({slot_id:txt(x.slot_id),display_name:txt(x.display_name),scene_role:txt(x.scene_role),appearance:clip(x.appearance,1200)}));
    return {source_key:row.source_key,source_index:row.source_index,source_text:row.source_text,scene_anchor_id:row.scene_anchor_id,scene_memory:row.scene_memory,director_decision:row.director_decision,selected_characters:compactSelected,manual_added_slot_ids:row.manual_added_slot_ids,manual_excluded_slot_ids:row.manual_excluded_slot_ids,manual_temporary_characters:manual,previous_ai_temporary_characters:priorAi,temporary_detection:"AI读取当前source_text自行判断本镜名单外单体/群体；仅真正可见才返回temporary_characters，纯提及/画外音分别按mentioned_only/offscreen_voice；本地词库候选不是最终人物。",relationship_context:row.relationship_context,adjacent_context:row.adjacent_context};
  }
  function aiOutlineView(pkg={}){return {scenes:arr(pkg.scenes).map(row=>aiSceneView(row,{mode:'outline'})),requirements:{formal_characters:'只使用本镜 selected_characters；不得新增正式人物',temporary_characters:'本镜临时人物由AI直接阅读source_text自行判断。不要采用本地猜测人物。用户manual_temporary_characters必须纳入。返回temporary_characters时只反馈本镜确实需要的单体/群体及其可见状态，并把可见人物的动作、准确站位、视线、反应和互动写入画面正文；不要输出内部ID或规则字段',scene_memory:'同一scene_anchor_id保持空间、固定元素、主光方向与色温连续',camera_language:'景别、机位/角度、运镜、转场全部使用中文'}};}
  function buildOutlinePrompt(plans=[]){
    const pkg=buildOutlinePackage(plans);if(!pkg.validation.ok)return {ok:false,package:pkg,prompt:'',reason:'validation_failed'};
    const json=JSON.stringify(aiOutlineView(pkg));return {ok:true,package:pkg,prompt:`【V78.2.7 分镜AI上下文｜AI主判临时人物｜内部数据不得原样输出】\n${json}\n执行：只根据原文、人物卡文字外形、Scene Memory、导演决策和相邻镜普通母版生成画面。临时人物/群体由你读取当前source_text自行判断；本地候选不参与最终人物判定，只有manual_temporary_characters是用户强制补充。可见临时人物必须落成可拍动作，不得只写名字或“在场”。最终每镜完整描述空间/站位、动作链、反应、互动、道具操作和镜头落点；禁止输出“根据原文确定子空间/人物当前站位”等占位语；统一风格只作全片成像基线，具体焦段/现场光源/构图由当前镜与Scene Memory决定。镜头字段使用V78.2.7结构：shot_size为中文景别，shot_angle为“焦段｜机位/角度｜创意构图”，movement为中文运镜，transition为中文转场；焦距仅保留mm单位。`,reason:''};
  }
  function buildRegenerationPackage(sceneInput={}){
    const all=scenes(),key=txt(sceneInput.source_key)||sourceKey(sceneInput,Number(sceneInput.source_index||1)-1),found=all.findIndex(s=>txt(s.source_key)===key||txt(s.id)===txt(sceneInput.id)),fallbackIndex=Math.max(0,Number(sceneInput.source_index||1)-1),index=found>=0?found:fallbackIndex,scene=all[index]||sceneInput;
    const pseudo={source_key:key,line_index:Number(scene.source_index||index+1),source_text:scene.source_text||scene.text,scene_anchor_id:scene.scene_anchor_id,temporary_scene_group:scene.temporary_scene_group,scene_memory:scene.scene_memory,director_decision:scene.director_decision,characters_mode:scene.characters_mode,character_packages:arr(scene.character_core_cast).filter(r=>uniq(scene.character_slot_ids).includes(txt(r.slot_id))),manual_added_slot_ids:scene.manual_added_slot_ids,manual_excluded_slot_ids:scene.manual_excluded_slot_ids,temporary_characters:scene.temporary_characters,manual_temporary_characters_v7824:scene.manual_temporary_characters_v7824,keyword_evidence:scene.keyword_evidence,relationship_evidence:scene.relationship_evidence,excluded_formal_slot_ids:arr(core().slots).map(s=>txt(s.slot_id)).filter(id=>!uniq(scene.character_slot_ids).includes(id)),character_core_context:{offscreen_voice:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='offscreen_voice'),mentioned_only:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='mentioned_only')}};
    const record=buildSceneRecord(pseudo,index,all),pkg={protocol:'v78_scene_regeneration_context_phase8',version:'scene_context_service_v78_2_4',created_at:now(),source_hash:txt(core().source_hash),character_revision:Number(core().character_revision||0),scene:record,current_visual_reference:visualText(scene),requirements:{preserve_structure:'unless user guidance explicitly requests merge/delete/change duration, keep current timeline structure and valid information',formal_cast:'selected_slot_ids are frozen; do not add unselected formal characters',continuity:'respect scene_memory and adjacent_context; current user guidance has highest semantic priority',camera_language:'all camera fields Chinese-only'}};
    pkg.validation=validate({scenes:[record]});lastRegeneration=clone(pkg);return clone(pkg);
  }
  function buildRegenerationPrompt(scene={}){const pkg=buildRegenerationPackage(scene);if(!pkg.validation.ok)return {ok:false,package:pkg,prompt:'',reason:'validation_failed'};const view={scene:aiSceneView(pkg.scene||{},{mode:'regeneration'}),current_visual_reference:visualText(scene),requirements:{preserve_structure:'除非用户明确要求改变结构，否则保留已有正确时段和信息',master_prompt:'前后镜与当前镜视觉参考只读取普通母版文字',temporary_characters:'重新读取当前source_text判断临时人物。previous_ai_temporary_characters只是上一版参考，可增删纠正；manual_temporary_characters必须保留。可见人物必须写清动作、具体站位、视线、反应和互动。',scene_style_authority:'原文事实 > Scene Memory > 当前镜导演决策 > 统一风格基线；禁止场景占位语；统一风格不得锁具体逐镜焦段/灯位/构图',camera_language:'V78.2.7：shot_size=景别；shot_angle=焦段｜机位/角度｜创意构图；movement=中文运镜；transition=中文转场；焦距仅保留mm单位'}};return {ok:true,package:pkg,prompt:`【V78.2.7 单镜重生成上下文｜AI重判临时人物｜内部数据不得原样输出】\n${JSON.stringify(view)}\n只修改当前镜；优先保持上一镜结束状态到本镜开场状态的连续承接。`,reason:''};}
  function stable(v){
    if(Array.isArray(v))return '['+v.map(stable).join(',')+']';
    if(v&&typeof v==='object')return '{'+Object.keys(v).sort().map(k=>JSON.stringify(k)+':'+stable(v[k])).join(',')+'}';
    return JSON.stringify(v??null);
  }
  function hash(v){let h=2166136261;for(const ch of stable(v)){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(36);}
  function tempSemantic(item={}){
    return {continuity_id:txt(item.continuity_id||item.entity_id),entity_id:txt(item.entity_id),label:txt(item.label),source_phrase:txt(item.source_phrase||item.label),entity_type:txt(item.entity_type||'temporary_character'),participation_state:txt(item.participation_state||'visible_candidate'),must_render:item.must_render===true,count:Math.max(1,Number(item.count||1)),identity_hint:txt(item.identity_hint),gender_hint:txt(item.gender_hint),age_stage_hint:txt(item.age_stage_hint),species_hint:txt(item.species_hint),continuity_profile:obj(item.continuity_profile),wardrobe_state:obj(item.wardrobe_state),scene_key:txt(item.scene_key)};
  }
  function sceneSourceTruth(plan={},index=0,allScenes=scenes()){
    const key=txt(plan.source_key)||sourceKey(plan,index),scene=allScenes.find(s=>txt(s.source_key)===key)||allScenes[index]||{};
    const selected=arr(plan.character_packages).length?arr(plan.character_packages):arr(scene.character_core_cast).filter(r=>uniq(scene.character_slot_ids).includes(txt(r.slot_id)));
    const temporary=arr(plan.temporary_characters).length?arr(plan.temporary_characters):arr(scene.temporary_characters),ctx=obj(plan.character_core_context);
    const memory=sceneMemoryFor(plan,scene),director=obj(plan.director_decision||scene.director_decision),adj=adjacencyFor(index,allScenes);
    return {source_key:key,selected_slot_ids:selected.map(x=>txt(x.slot_id)).filter(Boolean),manual_added_slot_ids:uniq(plan.manual_added_slot_ids||scene.manual_added_slot_ids).sort(),manual_excluded_slot_ids:uniq(plan.manual_excluded_slot_ids||scene.manual_excluded_slot_ids).sort(),offscreen_slot_ids:uniq(arr(ctx.offscreen_voice).map(x=>txt(x.slot_id))).sort(),mentioned_slot_ids:uniq(arr(ctx.mentioned_only).map(x=>txt(x.slot_id))).sort(),temporary_characters:temporary.map(tempSemantic).sort((a,b)=>stable(a).localeCompare(stable(b))),scene_anchor_id:txt(plan.scene_anchor_id||scene.scene_anchor_id||plan.temporary_scene_group||scene.temporary_scene_group),scene_memory_hash:hash(memory),director_decision_hash:hash(director),previous_source_key:txt(adj.previous?.source_key),next_source_key:txt(adj.next?.source_key)};
  }
  function packageSemantic(row={}){
    return {source_key:txt(row.source_key),selected_slot_ids:arr(row.selected_slot_ids).map(txt).filter(Boolean),manual_added_slot_ids:uniq(row.manual_added_slot_ids).sort(),manual_excluded_slot_ids:uniq(row.manual_excluded_slot_ids).sort(),offscreen_slot_ids:uniq(arr(row.offscreen_voice).map(x=>txt(x.slot_id))).sort(),mentioned_slot_ids:uniq(arr(row.mentioned_only).map(x=>txt(x.slot_id))).sort(),temporary_characters:arr(row.temporary_characters).map(tempSemantic).sort((a,b)=>stable(a).localeCompare(stable(b))),scene_anchor_id:txt(row.scene_anchor_id),scene_memory_hash:hash(obj(row.scene_memory)),director_decision_hash:hash(obj(row.director_decision)),previous_source_key:txt(row.adjacent_context?.previous?.source_key),next_source_key:txt(row.adjacent_context?.next?.source_key)};
  }
  function parityAgainstPlans(plans=[],pkg=null){
    const p=pkg||buildOutlinePackage(plans),allScenes=scenes(),diffs=[],truth=arr(plans).map((plan,i)=>sceneSourceTruth(plan,i,allScenes)),candidate=arr(p.scenes).map(packageSemantic);
    if(truth.length!==candidate.length)diffs.push({type:'scene_count',legacy:truth.length,current:candidate.length});
    const fields=['source_key','selected_slot_ids','manual_added_slot_ids','manual_excluded_slot_ids','offscreen_slot_ids','mentioned_slot_ids','temporary_characters','scene_anchor_id','scene_memory_hash','director_decision_hash','previous_source_key','next_source_key'];
    for(let i=0;i<Math.max(truth.length,candidate.length);i++){
      const a=truth[i],b=candidate[i];if(!a||!b){diffs.push({type:'missing_scene',index:i,legacy:Boolean(a),current:Boolean(b)});continue;}
      for(const field of fields)if(stable(a[field])!==stable(b[field]))diffs.push({type:field,source_key:a.source_key||b.source_key,legacy:clone(a[field]),current:clone(b[field])});
    }
    return {ok:Boolean(p.validation?.ok)&&diffs.length===0,diffs,mismatch_count:diffs.length,validation:clone(p.validation),scene_count:p.scenes.length,validator:'character_core_submission_semantic_validator_phase8'};
  }
  function parityRegenerationAgainstScene(sceneInput={},pkg=null){
    const p=pkg||buildRegenerationPackage(sceneInput),row=p.scene||{},all=scenes(),key=txt(sceneInput.source_key)||txt(row.source_key),found=all.findIndex(s=>txt(s.source_key)===key||txt(s.id)===txt(sceneInput.id)),index=found>=0?found:Math.max(0,Number(sceneInput.source_index||1)-1),scene=all[index]||sceneInput;
    const pseudo={source_key:key,line_index:Number(scene.source_index||index+1),source_text:scene.source_text||scene.text,scene_anchor_id:scene.scene_anchor_id,temporary_scene_group:scene.temporary_scene_group,scene_memory:scene.scene_memory,director_decision:scene.director_decision,characters_mode:scene.characters_mode,character_packages:arr(scene.character_core_cast).filter(r=>uniq(scene.character_slot_ids).includes(txt(r.slot_id))),manual_added_slot_ids:scene.manual_added_slot_ids,manual_excluded_slot_ids:scene.manual_excluded_slot_ids,temporary_characters:scene.temporary_characters,character_core_context:{offscreen_voice:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='offscreen_voice'),mentioned_only:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='mentioned_only')}};
    const a=sceneSourceTruth(pseudo,index,all),b=packageSemantic(row),diffs=[],fields=['source_key','selected_slot_ids','manual_added_slot_ids','manual_excluded_slot_ids','offscreen_slot_ids','mentioned_slot_ids','temporary_characters','scene_anchor_id','scene_memory_hash','director_decision_hash','previous_source_key','next_source_key'];
    fields.forEach(field=>{if(stable(a[field])!==stable(b[field]))diffs.push({type:field,source_key:a.source_key||b.source_key,legacy:clone(a[field]),current:clone(b[field])});});
    return {ok:Boolean(p.validation?.ok)&&diffs.length===0,diffs,mismatch_count:diffs.length,validation:clone(p.validation),source_key:txt(row.source_key),validator:'character_core_regeneration_semantic_validator_phase8'};
  }
  function phase15Production(stage,safe,meta={}){const p=globalThis.__V78_OUTLINE_PRODUCTION_AUTHORITY__;return p?.decide?.(stage,{safe,kind:meta.kind||'',source_key:meta.source_key||'',reasons:meta.reasons||[]})||null;}
  let authorityEnabled=true;
  const outlineAuthority={signature:'',safe_streak:0,promoted:false,promotion_count:0,demotion_count:0,takeover_count:0,warmup_count:0,fallback_count:0,last_executor:'character_core_submission_fallback',last_parity:null,last_decision:null,last_changed_at:null};
  const regenerationAuthority={signature:'',scenes:{},promotion_count:0,demotion_count:0,takeover_count:0,warmup_count:0,fallback_count:0,last_decision:null,last_changed_at:null};
  function contextSignature(plans=[]){
    const all=scenes(),rows=arr(plans).map((p,i)=>sceneSourceTruth(p,i,all));return hash({source_hash:txt(core().source_hash),character_revision:Number(core().character_revision||0),rows});
  }
  function ensureOutlineContext(plans=[]){
    const signature=contextSignature(plans);if(outlineAuthority.signature!==signature){if(outlineAuthority.promoted)outlineAuthority.demotion_count+=1;outlineAuthority.signature=signature;outlineAuthority.safe_streak=0;outlineAuthority.promoted=false;outlineAuthority.last_executor='character_core_submission_fallback';outlineAuthority.last_changed_at=now();}return signature;
  }
  function resolveOutlineAuthority(plans=[],pkg=null,parity=null){
    ensureOutlineContext(plans);const p=pkg||buildOutlinePackage(plans),report=parity||parityAgainstPlans(plans,p),was=outlineAuthority.promoted,safe=Boolean(authorityEnabled&&p.validation?.ok&&report.ok);
    const production=phase15Production('submission_outline',safe,{kind:'outline',reasons:report?.diffs?.map?.(x=>x.type)||[]});
    if(production){
      if(safe&&production.use_v78){outlineAuthority.safe_streak=Math.max(PROMOTION_STREAK,Number(outlineAuthority.safe_streak||0)+1);if(!outlineAuthority.promoted){outlineAuthority.promoted=true;outlineAuthority.promotion_count+=1;outlineAuthority.last_changed_at=now();}}
      else{outlineAuthority.safe_streak=0;if(outlineAuthority.promoted){outlineAuthority.promoted=false;outlineAuthority.demotion_count+=1;outlineAuthority.last_changed_at=now();}}
    }else if(safe){outlineAuthority.safe_streak+=1;if(!outlineAuthority.promoted&&outlineAuthority.safe_streak>=PROMOTION_STREAK){outlineAuthority.promoted=true;outlineAuthority.promotion_count+=1;outlineAuthority.last_changed_at=now();}}
    else{outlineAuthority.safe_streak=0;if(outlineAuthority.promoted){outlineAuthority.promoted=false;outlineAuthority.demotion_count+=1;outlineAuthority.last_changed_at=now();}}
    const takeover=Boolean(safe&&(production?production.use_v78:outlineAuthority.promoted)),executor=takeover?(production?'v78_submission_package_primary_phase15':'v78_submission_package_primary_phase8'):safe?(production?'character_core_submission_realtime_validator_phase15':'character_core_submission_warmup_validator'):'character_core_submission_fallback';
    if(takeover)outlineAuthority.takeover_count+=1;else if(safe)outlineAuthority.warmup_count+=1;else outlineAuthority.fallback_count+=1;
    outlineAuthority.last_executor=executor;outlineAuthority.last_parity=clone(report);outlineAuthority.last_decision={takeover,executor,safe_streak:outlineAuthority.safe_streak,promoted:outlineAuthority.promoted,promoted_now:!was&&outlineAuthority.promoted,demoted_now:was&&!outlineAuthority.promoted,threshold:PROMOTION_STREAK,parity:clone(report),validation:clone(p.validation),decided_at:now()};
    console.info('[V78_SUBMISSION_AUTHORITY]',{stage:'outline',takeover,executor,safe_streak:outlineAuthority.safe_streak,threshold:PROMOTION_STREAK,mismatch_count:report.mismatch_count||0});return clone(outlineAuthority.last_decision);
  }
  function regenSignature(sceneInput={},pkg=null){const all=scenes(),p=pkg||buildRegenerationPackage(sceneInput),key=txt(sceneInput.source_key)||txt(p.scene?.source_key),found=all.findIndex(s=>txt(s.source_key)===key||txt(s.id)===txt(sceneInput.id)),index=found>=0?found:Math.max(0,Number(sceneInput.source_index||1)-1),scene=all[index]||sceneInput,pseudo={source_key:key,line_index:Number(scene.source_index||index+1),source_text:scene.source_text||scene.text,scene_anchor_id:scene.scene_anchor_id,temporary_scene_group:scene.temporary_scene_group,scene_memory:scene.scene_memory,director_decision:scene.director_decision,characters_mode:scene.characters_mode,character_packages:arr(scene.character_core_cast).filter(r=>uniq(scene.character_slot_ids).includes(txt(r.slot_id))),manual_added_slot_ids:scene.manual_added_slot_ids,manual_excluded_slot_ids:scene.manual_excluded_slot_ids,temporary_characters:scene.temporary_characters,character_core_context:{offscreen_voice:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='offscreen_voice'),mentioned_only:arr(scene.character_core_cast).filter(r=>txt(r.scene_role)==='mentioned_only')}};return hash({source_hash:txt(core().source_hash),character_revision:Number(core().character_revision||0),semantic:sceneSourceTruth(pseudo,index,all)});}
  function ensureRegenerationContext(sceneInput={},pkg=null){
    const globalSig=hash({source_hash:txt(core().source_hash),character_revision:Number(core().character_revision||0)});if(regenerationAuthority.signature!==globalSig){regenerationAuthority.signature=globalSig;regenerationAuthority.scenes={};regenerationAuthority.last_changed_at=now();}
    const p=pkg||buildRegenerationPackage(sceneInput),key=txt(p.scene?.source_key)||txt(sceneInput.source_key)||'scene_unknown',sig=regenSignature(sceneInput,p),current=obj(regenerationAuthority.scenes[key]);if(current.signature!==sig){if(current.promoted)regenerationAuthority.demotion_count+=1;regenerationAuthority.scenes[key]={signature:sig,safe_streak:0,promoted:false,last_executor:'character_core_regeneration_submission_fallback',last_changed_at:now()};}return {key,p,row:regenerationAuthority.scenes[key]};
  }
  function resolveRegenerationAuthority(sceneInput={},pkg=null,parity=null){
    const ctx=ensureRegenerationContext(sceneInput,pkg),p=ctx.p,key=ctx.key,row=ctx.row,report=parity||parityRegenerationAgainstScene(sceneInput,p),was=Boolean(row.promoted),safe=Boolean(authorityEnabled&&p.validation?.ok&&report.ok);
    const production=phase15Production('submission_regeneration',safe,{kind:'regeneration',source_key:key,reasons:report?.diffs?.map?.(x=>x.type)||[]});
    if(production){
      if(safe&&production.use_v78){row.safe_streak=Math.max(PROMOTION_STREAK,Number(row.safe_streak||0)+1);if(!row.promoted){row.promoted=true;regenerationAuthority.promotion_count+=1;row.last_changed_at=now();}}
      else{row.safe_streak=0;if(row.promoted){row.promoted=false;regenerationAuthority.demotion_count+=1;row.last_changed_at=now();}}
    }else if(safe){row.safe_streak=Number(row.safe_streak||0)+1;if(!row.promoted&&row.safe_streak>=PROMOTION_STREAK){row.promoted=true;regenerationAuthority.promotion_count+=1;row.last_changed_at=now();}}
    else{row.safe_streak=0;if(row.promoted){row.promoted=false;regenerationAuthority.demotion_count+=1;row.last_changed_at=now();}}
    const takeover=Boolean(safe&&(production?production.use_v78:row.promoted)),executor=takeover?(production?'v78_regeneration_submission_primary_phase15':'v78_regeneration_submission_primary_phase8'):safe?(production?'character_core_regeneration_submission_realtime_validator_phase15':'character_core_regeneration_submission_warmup_validator'):'character_core_regeneration_submission_fallback';
    if(takeover)regenerationAuthority.takeover_count+=1;else if(safe)regenerationAuthority.warmup_count+=1;else regenerationAuthority.fallback_count+=1;
    Object.assign(row,{last_executor:executor,last_parity:clone(report),last_validation:clone(p.validation),last_decision_at:now()});regenerationAuthority.last_decision={source_key:key,takeover,executor,safe_streak:row.safe_streak,promoted:row.promoted,promoted_now:!was&&row.promoted,demoted_now:was&&!row.promoted,threshold:PROMOTION_STREAK,parity:clone(report),validation:clone(p.validation),decided_at:now()};
    console.info('[V78_SUBMISSION_AUTHORITY]',{stage:'regeneration',source_key:key,takeover,executor,safe_streak:row.safe_streak,threshold:PROMOTION_STREAK,mismatch_count:report.mismatch_count||0});return clone(regenerationAuthority.last_decision);
  }
  function submissionAuthoritySnapshot(){
    const regenRows=Object.values(regenerationAuthority.scenes);return {service:'ai_submission_package_authority_v78_phase8',enabled:authorityEnabled,promotion_threshold:PROMOTION_STREAK,outline:{...clone(outlineAuthority)},regeneration:{signature:regenerationAuthority.signature,promotion_count:regenerationAuthority.promotion_count,demotion_count:regenerationAuthority.demotion_count,takeover_count:regenerationAuthority.takeover_count,warmup_count:regenerationAuthority.warmup_count,fallback_count:regenerationAuthority.fallback_count,promoted_scene_count:regenRows.filter(x=>x.promoted).length,scenes:clone(regenerationAuthority.scenes),last_decision:clone(regenerationAuthority.last_decision),last_changed_at:regenerationAuthority.last_changed_at}};
  }
  function setAuthorityEnabled(value){authorityEnabled=Boolean(value);if(!authorityEnabled){if(outlineAuthority.promoted)outlineAuthority.demotion_count+=1;outlineAuthority.promoted=false;outlineAuthority.safe_streak=0;outlineAuthority.last_executor='character_core_submission_fallback';Object.values(regenerationAuthority.scenes).forEach(row=>{row.promoted=false;row.safe_streak=0;row.last_executor='character_core_regeneration_submission_fallback';});}return submissionAuthoritySnapshot();}
  function resetAuthority(){outlineAuthority.signature='';outlineAuthority.safe_streak=0;outlineAuthority.promoted=false;outlineAuthority.last_executor='character_core_submission_fallback';regenerationAuthority.signature='';regenerationAuthority.scenes={};return submissionAuthoritySnapshot();}
  function snapshot(){return {service:'scene_context_service_v78_2_4',phase:'v78_phase8',authoritative:'phase15_v78_default_primary_with_character_core_realtime_validator_fallback',last_outline:lastOutline?{scene_count:lastOutline.scenes.length,validation:lastOutline.validation,created_at:lastOutline.created_at}:null,last_regeneration:lastRegeneration?{source_key:lastRegeneration.scene?.source_key,validation:lastRegeneration.validation,created_at:lastRegeneration.created_at}:null,submission_authority:submissionAuthoritySnapshot()};}
  let lastOutline=null,lastRegeneration=null;
  const service={version:'scene_context_service_v78_2_4',phase:'v78_phase8',authoritative:'phase15_v78_default_primary_with_character_core_realtime_validator_fallback',promotion_threshold:PROMOTION_STREAK,buildSceneRecord,buildOutlinePackage,buildOutlinePrompt,buildRegenerationPackage,buildRegenerationPrompt,validate,parityAgainstPlans,parityRegenerationAgainstScene,resolveOutlineAuthority,resolveRegenerationAuthority,submissionAuthoritySnapshot,setAuthorityEnabled,resetAuthority,snapshot};
  ns.sceneContext=service;globalThis.__V78_SCENE_CONTEXT_SERVICE__=service;
  console.info('[V78 Clean Core] scene context phase8 submission authority ready',{version:service.version,authoritative:service.authoritative,promotion_threshold:PROMOTION_STREAK});
})();

/* V78.2 Clean Core - unified Reference Entity binding + lossless reference injection. */
(function installV781ReferenceEntities(){
  if(globalThis.__V78_REFERENCE_RESOLVER__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const VERSION="reference_entity_resolver_v78_2_4";
  const STORE_VERSION="reference_bindings_v781";
  const TEMP_PERFORMANCE_REPAIR_TASK="temporary_performance_repair";
  const txt=(v)=>String(v??"").trim();
  const arr=(v)=>Array.isArray(v)?v:[];
  const obj=(v)=>v&&typeof v==="object"&&!Array.isArray(v)?v:{};
  const clone=(v)=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const uniq=(v)=>[...new Set(arr(v).map(txt).filter(Boolean))];
  const now=()=>new Date().toISOString();
  function hash(value=""){let h=2166136261;for(const ch of String(value)){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(36);}
  function esc(value=""){return String(value).replace(/[.*+?^${}()|[\]\\]/g,"\\$&");}
  function htmlEsc(value=""){return String(value??"").replace(/[&<>"']/g,ch=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","\'":"&#39;"}[ch]||ch));}
  function sourceHash(){
    try{return txt(globalThis.__currentV77SourceSnapshot?.("reference_entities_v781")?.source_hash);}catch(_){ }
    try{return txt(state?.characterCoreV2?.source_hash||state?.source_hash);}catch(_){return "";}
  }
  function premium(){try{return state.outputMode==="premium";}catch(_){return false;}}
  function host(){return globalThis.__V78_PREMIUM_HOST__||{};}
  function allAssets(){
    const refs=obj(state.referenceAssets);return [...arr(refs.characters),...arr(refs.scenes),...arr(refs.props)];
  }
  function bucket(type){
    state.referenceAssets=state.referenceAssets||{characters:[],scenes:[],props:[]};
    if(type==="character"||type==="temporary_character")return state.referenceAssets.characters=arr(state.referenceAssets.characters);
    if(type==="prop")return state.referenceAssets.props=arr(state.referenceAssets.props);
    return state.referenceAssets.scenes=arr(state.referenceAssets.scenes);
  }
  function assetById(id){const key=txt(id);return allAssets().find(x=>txt(x.asset_id)===key)||null;}
  function aliasesOf(asset={}){
    const aliases=[];
    aliases.push(asset.name);
    arr(asset.aliases).forEach(x=>aliases.push(x));
    String(asset.keywords||"").split(/[\n,，、;；|｜/]+/).forEach(x=>aliases.push(x));
    return uniq(aliases).sort((a,b)=>b.length-a.length);
  }
  function sourceKey(scene={},index=0){return txt(scene.source_key)||`line_${Number(scene.source_index||index+1)}`;}
  function sceneAnchor(scene={},index=0){return txt(scene.scene_anchor_id||scene.temporary_scene_group||scene.scene_memory?.scene_anchor_id)||`scene_anchor_${sourceKey(scene,index)}`;}
  function ensureKind(entry,key){
    const current=obj(entry[key]);entry[key]={auto:uniq(current.auto),manual_added:uniq(current.manual_added),manual_excluded:uniq(current.manual_excluded),suggested:uniq(current.suggested)};return entry[key];
  }
  function ensureStore(){
    const sh=sourceHash();
    if(!state.referenceBindingsV781||typeof state.referenceBindingsV781!=="object")state.referenceBindingsV781={version:STORE_VERSION,source_hash:sh,by_source_key:{},scene_anchor_assets:{},temporary_assets:{},prop_entities:{},updated_at:now()};
    const store=state.referenceBindingsV781;store.version=STORE_VERSION;store.by_source_key=obj(store.by_source_key);store.scene_anchor_assets=obj(store.scene_anchor_assets);store.temporary_assets=obj(store.temporary_assets);store.prop_entities=obj(store.prop_entities);
    if(store.source_hash&&sh&&store.source_hash!==sh){store.by_source_key={};store.scene_anchor_assets={};store.temporary_assets={};store.prop_entities={};}
    store.source_hash=sh||store.source_hash||"";return store;
  }
  function ensureEntry(scene={},index=0){
    const store=ensureStore(),key=sourceKey(scene,index);let entry=obj(store.by_source_key[key]);
    entry.source_key=key;entry.scene_anchor_id=sceneAnchor(scene,index);entry.prop_state_variants=obj(entry.prop_state_variants);ensureKind(entry,"scene");ensureKind(entry,"prop");ensureKind(entry,"temporary");
    if(!entry.legacy_migrated){
      const auto=uniq(scene.reference_auto_asset_ids),added=uniq(scene.reference_manual_added_ids),excluded=uniq(scene.reference_manual_excluded_ids);
      for(const id of auto){const a=assetById(id);if(a?.type==="scene")entry.scene.auto.push(id);else if(a?.type==="prop")entry.prop.auto.push(id);}
      for(const id of added){const a=assetById(id);if(a?.type==="scene")entry.scene.manual_added.push(id);else if(a?.type==="prop")entry.prop.manual_added.push(id);}
      for(const id of excluded){const a=assetById(id);if(a?.type==="scene")entry.scene.manual_excluded.push(id);else if(a?.type==="prop")entry.prop.manual_excluded.push(id);}
      ensureKind(entry,"scene");ensureKind(entry,"prop");entry.legacy_migrated=true;
    }
    store.by_source_key[key]=entry;return entry;
  }
  function effective(kind={}){const set=new Set(uniq(kind.auto));uniq(kind.manual_added).forEach(x=>set.add(x));uniq(kind.manual_excluded).forEach(x=>set.delete(x));return [...set];}
  function visibility(temp={}){const s=txt(temp.participation_state||temp.scene_role||"visible_candidate");return !/(mentioned_only|offscreen_voice|offscreen|仅提及|画外音)/i.test(s);}
  function tempName(temp={}){return txt(temp.source_phrase||temp.label||temp.name||temp.display_name||"临时人物");}
  function tempEntityId(temp={},scene={},index=0){
    const existing=txt(temp.temporary_entity_id||temp.entity_id||temp.continuity_id);if(existing)return existing;
    return `temp_${hash(`${sceneAnchor(scene,index)}|${tempName(temp)}|${txt(temp.gender)}|${txt(temp.visual_age_stage||temp.age_stage)}`)}`;
  }
  function normalizeTemp(temp={},scene={},index=0){
    const out={...temp};out.temporary_entity_id=tempEntityId(out,scene,index);out.entity_type="temporary_character";out.label=tempName(out);out.source_phrase=txt(out.source_phrase)||out.label;out.continuity_id=txt(out.continuity_id)||`tc_${hash(out.temporary_entity_id)}`;
    out.continuity_profile=txt(out.continuity_profile);out.wardrobe_state=txt(out.wardrobe_state);return out;
  }
  function shotsForScene(scene={}){try{return typeof outlineShotsByScene==="function"?arr(outlineShotsByScene(scene.id)):arr(state.outlineShots).filter(s=>txt(s.parent_scene_id)===txt(scene.id)||arr(s.parent_scene_ids).map(txt).includes(txt(scene.id)));}catch(_){return [];}}
  function temporaryForScene(scene={},index=0){
    const combined=[...arr(scene.temporary_characters),...shotsForScene(scene).flatMap(s=>arr(s.temporary_characters)),...arr(scene.manual_temporary_characters_v7824)].filter(x=>x?.manual_added_v7824===true||txt(x?.selection_source_v7824)==="manual"||txt(x?.selection_source_v7824)==="ai_primary"||x?.ai_discovered===true),map=new Map();
    combined.forEach(raw=>{const item=normalizeTemp(raw,scene,index),id=item.temporary_entity_id;if(!id)return;const prior=map.get(id)||{};map.set(id,{...prior,...item,continuity_profile:txt(item.continuity_profile)||txt(prior.continuity_profile),wardrobe_state:txt(item.wardrobe_state)||txt(prior.wardrobe_state)});});
    const list=[...map.values()];
    // Persist stable IDs without changing formal CharacterCore membership.
    arr(scene.temporary_characters).forEach(raw=>{const n=normalizeTemp(raw,scene,index);raw.temporary_entity_id=n.temporary_entity_id;raw.entity_type="temporary_character";if(!raw.continuity_id)raw.continuity_id=n.continuity_id;});
    shotsForScene(scene).forEach(shot=>arr(shot.temporary_characters).forEach(raw=>{const n=normalizeTemp(raw,scene,index);raw.temporary_entity_id=n.temporary_entity_id;raw.entity_type="temporary_character";if(!raw.continuity_id)raw.continuity_id=n.continuity_id;}));
    return list;
  }
  function tempAsset(temp={}){
    const store=ensureStore(),entityId=txt(temp.temporary_entity_id);const mapped=txt(store.temporary_assets[entityId]);
    if(mapped){const a=assetById(mapped);if(a)return a;}
    const found=arr(state.referenceAssets?.characters).find(a=>txt(a.entity_type)==="temporary_character"&&txt(a.temporary_entity_id)===entityId)||null;
    if(found){store.temporary_assets[entityId]=found.asset_id;return found;}return null;
  }
  function sceneSearchText(scene={}){
    const memory=obj(scene.scene_memory);return [scene.source_text,scene.text,scene.visual_context,memory.location_hint,memory.fixed_elements,memory.reference_asset_names,shotsForScene(scene).map(s=>`${txt(s.visual_context)} ${txt(s.prompt)}`).join("\n")].flat().map(txt).filter(Boolean).join("\n").toLowerCase();
  }
  function phraseMatch(hay,phrase){const p=txt(phrase).toLowerCase();if(!p)return 0;if(hay.includes(p))return p.length>=4?0.96:p.length>=2?0.9:0.75;return 0;}
  function weakOverlap(hay,asset={}){
    const tokens=aliasesOf(asset).flatMap(x=>x.split(/[\s的与和及、，,。；;：:·\-_/]+/)).map(txt).filter(x=>x.length>=2);let hits=0;for(const t of uniq(tokens)){if(hay.includes(t.toLowerCase()))hits++;}return hits>=2?0.72:hits===1?0.55:0;
  }
  function matchAsset(scene,asset,index=0){
    const hay=sceneSearchText(scene),anchor=sceneAnchor(scene,index),store=ensureStore();let score=0,reason="";
    if(asset.type==="scene"&&txt(store.scene_anchor_assets[anchor])===txt(asset.asset_id)){score=1;reason="场景锚点稳定绑定";}
    for(const alias of aliasesOf(asset)){const s=phraseMatch(hay,alias);if(s>score){score=s;reason=alias===asset.name?`名称命中：${alias}`:`别称/关键词命中：${alias}`;}}
    if(asset.type==="scene"){
      const loc=txt(scene.scene_memory?.location_hint).toLowerCase();for(const alias of aliasesOf(asset)){const a=alias.toLowerCase();if(a&&loc&&(loc.includes(a)||a.includes(loc))&&Math.min(a.length,loc.length)>=2&&score<0.88){score=0.88;reason=`Scene Memory地点命中：${scene.scene_memory?.location_hint}`;}}
    }
    const weak=weakOverlap(hay,asset);if(weak>score){score=weak;reason="描述特征候选";}
    return {score,reason};
  }
  function refreshBindings(options={}){
    const store=ensureStore();const scenes=arr(state.scenes),sceneAssets=arr(state.referenceAssets?.scenes),propAssets=arr(state.referenceAssets?.props);
    // Discover one stable scene asset per Scene Memory anchor before filling individual source lines.
    // This lets later lines inherit the same place even when only one line explicitly names it.
    const anchorCandidates=new Map();
    scenes.forEach((scene,index)=>{const anchor=sceneAnchor(scene,index);if(store.scene_anchor_assets[anchor]||!sceneAssets.length)return;for(const asset of sceneAssets){const m=matchAsset(scene,asset,index),row=anchorCandidates.get(anchor)||[];row.push({asset_id:asset.asset_id,score:m.score});anchorCandidates.set(anchor,row);}});
    for(const [anchor,rows] of anchorCandidates){const best=rows.sort((a,b)=>b.score-a.score),top=best[0],next=best[1];if(top&&top.score>=0.82&&(!next||top.score-next.score>=0.05||top.score>=0.93))store.scene_anchor_assets[anchor]=top.asset_id;}
    scenes.forEach((scene,index)=>{
      const entry=ensureEntry(scene,index),anchor=sceneAnchor(scene,index);entry.scene.auto=[];entry.scene.suggested=[];entry.prop.auto=[];entry.prop.suggested=[];
      for(const asset of sceneAssets){const m=matchAsset(scene,asset,index);if(m.score>=0.82)entry.scene.auto.push(asset.asset_id);else if(m.score>=0.52)entry.scene.suggested.push(asset.asset_id);asset.scene_asset_id=txt(asset.scene_asset_id)||asset.asset_id;}
      for(const asset of propAssets){const m=matchAsset(scene,asset,index);if(m.score>=0.82)entry.prop.auto.push(asset.asset_id);else if(m.score>=0.52)entry.prop.suggested.push(asset.asset_id);asset.prop_entity_id=txt(asset.prop_entity_id)||`prop_${hash(asset.asset_id)}`;store.prop_entities[asset.asset_id]={prop_entity_id:asset.prop_entity_id,aliases:aliasesOf(asset),state_variant:txt(asset.state_variant)};}
      // V78.2.6: temporary people are text-only performance entities. Legacy temporary image bindings
      // remain in persisted history for compatibility, but are never auto-selected or used.
      entry.temporary.auto=[];entry.temporary.suggested=[];entry.temporary.manual_added=[];
      ensureKind(entry,"scene");ensureKind(entry,"prop");ensureKind(entry,"temporary");
      const selectedSceneId=effective(entry.scene)[0]||"",selectedSceneAsset=assetById(selectedSceneId);
      if(selectedSceneId){scene.scene_memory={...obj(scene.scene_memory),scene_anchor_id:anchor,reference_scene_asset_id:selectedSceneId,reference_scene_asset_revision:txt(selectedSceneAsset?.main_revision||selectedSceneAsset?.main_updated_at||selectedSceneAsset?.updated_at)};}
      for(const id of effective(entry.prop)){const a=assetById(id);if(a)entry.prop_state_variants[id]=txt(a.state_variant)||txt(entry.prop_state_variants[id]);}
      // Keep V77 Scene Memory compatible: only scene/prop reference IDs are mirrored to legacy fields.
      scene.reference_auto_asset_ids=uniq([...entry.scene.auto,...entry.prop.auto]);
      scene.reference_manual_added_ids=uniq([...entry.scene.manual_added,...entry.prop.manual_added]);
      scene.reference_manual_excluded_ids=uniq([...entry.scene.manual_excluded,...entry.prop.manual_excluded]);
      entry.scene_anchor_id=anchor;entry.updated_at=now();store.by_source_key[sourceKey(scene,index)]=entry;
    });
    store.updated_at=now();if(options.persist){try{scheduleDraftSave?.();}catch(_){}}
    return store;
  }
  function onAssetRemoved(assetId){
    const id=txt(assetId),store=ensureStore();for(const entry of Object.values(store.by_source_key||{})){for(const type of ["scene","prop","temporary"]){const k=ensureKind(entry,type);k.auto=k.auto.filter(x=>x!==id);k.manual_added=k.manual_added.filter(x=>x!==id);k.manual_excluded=k.manual_excluded.filter(x=>x!==id);k.suggested=k.suggested.filter(x=>x!==id);}}
    Object.keys(store.scene_anchor_assets).forEach(k=>{if(txt(store.scene_anchor_assets[k])===id)delete store.scene_anchor_assets[k];});Object.keys(store.temporary_assets).forEach(k=>{if(txt(store.temporary_assets[k])===id)delete store.temporary_assets[k];});delete store.prop_entities[id];store.updated_at=now();try{scheduleDraftSave?.();}catch(_){ }
  }
  function setBinding(scene,type,assetId,checked,index=0){
    const entry=ensureEntry(scene,index),kind=ensureKind(entry,type);const auto=new Set(kind.auto),added=new Set(kind.manual_added),excluded=new Set(kind.manual_excluded),id=txt(assetId);
    if(checked){excluded.delete(id);if(!auto.has(id))added.add(id);}else{added.delete(id);excluded.add(id);}
    kind.manual_added=[...added];kind.manual_excluded=[...excluded];
    if(type==="scene"&&checked){ensureStore().scene_anchor_assets[sceneAnchor(scene,index)]=id;}
    refreshBindings({persist:false});
    try{host().invalidateSegments?.("reference_entity_binding_v781");}catch(_){ }
    try{scheduleDraftSave?.();}catch(_){ }
  }
  function bindingState(entry,type,assetId){
    const k=ensureKind(entry,type),id=txt(assetId),eff=new Set(effective(k));
    if(k.manual_excluded.includes(id))return {checked:false,status:"人工排除",kind:"excluded"};
    if(k.manual_added.includes(id))return {checked:true,status:"人工勾选",kind:"manual"};
    if(k.auto.includes(id))return {checked:eff.has(id),status:"自动匹配",kind:"auto"};
    if(k.suggested.includes(id))return {checked:false,status:"建议匹配",kind:"suggested"};
    return {checked:false,status:"未匹配",kind:"none"};
  }
  function formalSlotIds(scene={}){
    const ids=[...arr(scene.character_slot_ids),...arr(scene.character_core_cast?.selected_characters).map(x=>x?.slot_id),...arr(scene.character_core_context?.selected_characters).map(x=>x?.slot_id)];return uniq(ids);
  }
  function slotById(id){try{return arr(state.characterCoreV2?.slots).find(s=>txt(s.slot_id)===txt(id))||null;}catch(_){return null;}}
  function slotName(slot={}){return txt(slot.display_name||slot.name||slot.canonical_name||slot.base_name||"人物");}
  function slotAppearance(slot={}){return txt(slot.appearance||slot.appearance_text||slot.character_appearance||slot.appearance_detail||slot.appearance_details||"");}
  function slotAliases(slot={}){return uniq([slotName(slot),...arr(slot.aliases),...(txt(slot.alias_text)?txt(slot.alias_text).split(/[、,，|｜/]+/):[])]);}
  function identityText(temp={}){
    return [txt(temp.gender),txt(temp.visual_age_stage||temp.age_stage||temp.life_stage),txt(temp.identity_hint||temp.role||temp.label),txt(temp.continuity_profile),txt(temp.wardrobe_state)].filter(Boolean).join("，");
  }
  function ensureTemporaryAsset(_temp,_scene,_index=0){
    // V78.2.6: legacy temporary image assets are read-only compatibility data.
    // New temporary people never create/upload/generate reference images.
    console.info("[V78.2.6 TEMP_REFERENCE] 临时人物参考图功能已退场；请使用文字动态表达。 ");
    return null;
  }
  const ACTION_RE=/(走|跑|站|坐|躺|蹲|跪|抬|低头|转身|回头|伸手|收手|拿|放|递|接|推|拉|打开|关上|查看|翻|盯|看向|望向|扫视|点头|摇头|皱眉|微笑|哭|笑|喊|说|问|答|扶|抱|抓|握|敲|按|点击|掏出|穿过|靠近|后退|停下|离开|进入|端着|推着|弯腰|俯身|起身|抬眼|垂眼|示意|拦住|挡住|追上|避开|递给|放下)/;
  const INTERACT_RE=/(看向|面对|对着|递给|接过|扶住|拦住|拉住|推开|靠近|远离|回应|质问|回答|示意|朝|向|与|对)/;
  function scenePromptText(scene={}){return shotsForScene(scene).map(s=>txt(s.prompt||s.visual_context)).filter(Boolean).join("\n");}
  function performanceHealthFor(scene={},temp={},index=0){
    if(!visibility(temp))return {status:"offscreen",level:"pass",message:"非实体出镜"};
    const label=tempName(temp),aliases=uniq([label,temp.source_phrase,temp.name]),textAll=scenePromptText(scene);let found="",context="";
    for(const alias of aliases.sort((a,b)=>b.length-a.length)){const p=textAll.indexOf(alias);if(p>=0){found=alias;context=textAll.slice(Math.max(0,p-100),Math.min(textAll.length,p+Math.max(140,alias.length+120)));break;}}
    if(!found)return {status:"missing",level:"warn",message:`${label}需要实体出镜，但画面正文没有明确落地`};
    const action=ACTION_RE.test(context),interaction=INTERACT_RE.test(context);if(!action&&!interaction)return {status:"weak",level:"warn",message:`${label}已出现，但附近缺少有效动作/交互`};
    return {status:"ok",level:"pass",message:`${label}具有当前镜动作/交互`};
  }
  function performanceHealth(){
    const issues=[],rows=[];arr(state.scenes).forEach((scene,index)=>temporaryForScene(scene,index).forEach(temp=>{const h=performanceHealthFor(scene,temp,index);rows.push({source_key:sourceKey(scene,index),temporary_entity_id:temp.temporary_entity_id,label:temp.label,...h});if(h.level!=="pass")issues.push(rows[rows.length-1]);}));
    return {ok:!issues.length,total:rows.length,issues,rows};
  }
  function sceneAssetSelection(scene,index=0){const e=ensureEntry(scene,index);return {scene:effective(e.scene),prop:effective(e.prop),temporary:effective(e.temporary)};}
  function selectedMissingForScene(scene,index=0){
    const e=ensureEntry(scene,index),out=[];for(const type of ["scene","prop"]){for(const id of effective(e[type])){const a=assetById(id);if(!a?.has_main_image)out.push({type,asset:a||{asset_id:id,name:id},asset_id:id});}}return out;
  }
  function renderThumb(asset,alt="参考图"){
    const img=document.createElement("img");img.className="v781-reference-thumb";img.alt=alt;const url=host().previewUrl?.(asset)||host().assetUrl?.(asset,"main")||"";if(url)img.src=url;else img.classList.add("is-empty");return img;
  }
  function statusBadge(textValue,kind="none") {const s=document.createElement("span");s.className=`v781-reference-status is-${kind}`;s.textContent=textValue;return s;}
  function renderReferenceCard({scene,index,type,asset,temp,slot,entry,readonly=false}){
    const card=document.createElement("div");card.className=`v781-reference-card type-${type}`;const id=asset?.asset_id||"";
    const name=type==="formal"?slotName(slot):type==="temporary"?tempName(temp):txt(asset?.name)||"未命名";
    const binding=type==="formal"?{checked:Boolean(asset?.has_main_image),status:asset?.has_main_image?"随本镜人物":"人物已选但无主图",kind:asset?.has_main_image?"auto":"missing"}:bindingState(entry,type,id);
    const check=document.createElement("input");check.type="checkbox";check.checked=Boolean(binding.checked);check.disabled=readonly||type==="formal"||!asset;
    if(type!=="formal"&&asset)check.addEventListener("change",()=>{setBinding(scene,type,id,check.checked,index);try{renderSegments?.();}catch(_){ }renderReferenceBoxes();renderLibrary();});
    const media=document.createElement("div");media.className="v781-reference-media";media.appendChild(renderThumb(asset,name));
    const meta=document.createElement("div");meta.className="v781-reference-meta";const title=document.createElement("strong");title.textContent=name;const kind=document.createElement("span");kind.className="v781-reference-kind";kind.textContent=type==="formal"?"正式人物":type==="temporary"?"临时人物":type==="scene"?"场景":"道具";meta.append(title,kind,statusBadge(binding.status,binding.kind));
    if((type==="scene"||type==="prop")&&asset){const m=matchAsset(scene,asset,index);if(m.reason){const why=document.createElement("span");why.className="v781-reference-note";why.textContent=`匹配依据：${m.reason}`;meta.appendChild(why);}}
    if(type==="temporary"){
      const ph=performanceHealthFor(scene,temp,index),p=document.createElement("span");p.className=`v781-performance-state ${ph.level}`;p.textContent=ph.level==="pass"?"动作表达正常":ph.message;meta.appendChild(p);
      const continuity=document.createElement("span");continuity.className="v781-reference-note";continuity.textContent=[txt(temp.continuity_id)&&`连续ID：${temp.continuity_id}`,txt(temp.continuity_profile),txt(temp.wardrobe_state)].filter(Boolean).join(" · ");if(continuity.textContent)meta.appendChild(continuity);
    } else if(type==="scene"){
      const note=document.createElement("span");note.className="v781-reference-note";note.textContent=`场景锚点：${sceneAnchor(scene,index)}`;meta.appendChild(note);
    } else if(type==="prop"&&txt(asset?.state_variant)){
      const note=document.createElement("span");note.className="v781-reference-note";note.textContent=`状态：${asset.state_variant}`;meta.appendChild(note);
    }
    if(asset&&!asset.has_main_image){const warn=document.createElement("span");warn.className="v781-reference-warning";warn.textContent="已绑定资产但暂无主参考图，最终不会生成 @图N；画面文字不会被删除。";meta.appendChild(warn);}
    const actions=document.createElement("div");actions.className="v781-reference-actions";
    const settings=document.createElement("button");settings.type="button";settings.className="btn secondary small-btn";settings.textContent=asset?"图片设置 / 生成":"创建参考图";
    settings.addEventListener("click",()=>{if(type==="formal")host().openAssetDialog?.("character",asset,slot?.slot_id||"");else if(type==="temporary")ensureTemporaryAsset(temp,scene,index);else host().openAssetDialog?.(type,asset);});actions.appendChild(settings);
    card.append(check,media,meta,actions);return card;
  }
  function humanTempType(temp={}){
    const stateValue=txt(temp.participation_state||temp.scene_role||"visible_candidate");
    if(/offscreen_voice|offscreen|画外音/i.test(stateValue))return "画外音";
    if(/mentioned_only|仅提及/i.test(stateValue))return "仅提及";
    if(/group|群体|众|人群|家人|姐妹|兄弟|宾客|同学|士兵|侍卫/i.test(`${stateValue} ${tempName(temp)}`)||Number(temp.count||1)>1)return "群体出镜";
    if(/background|背景/i.test(stateValue))return "背景出镜";
    return "单体出镜";
  }
  function manualTemp(temp={}){return temp?.manual_added_v7824===true||txt(temp.selection_source_v7824)==="manual";}
  function temporaryPool(currentScene={}){
    const currentKey=sourceKey(currentScene,arr(state.scenes).indexOf(currentScene)),map=new Map();
    arr(state.scenes).forEach((scene,index)=>{
      if(sourceKey(scene,index)===currentKey)return;
      temporaryForScene(scene,index).forEach(raw=>{
        const name=tempName(raw);if(!name)return;const key=name.toLowerCase();
        if(!map.has(key))map.set(key,{...clone(raw),label:name,source_phrase:name,borrowed_from_source_key:sourceKey(scene,index)});
      });
    });
    return [...map.values()].slice(0,80);
  }
  function normalizeManualParticipation(value="single"){
    return ({single:"screen_visible",group:"group_visible",background:"background_visible",offscreen:"offscreen_voice",mentioned:"mentioned_only"})[value]||"screen_visible";
  }
  function upsertManualTemporary(scene={},raw={}){
    const label=txt(raw.label||raw.source_phrase);if(!label)return null;
    const participation_state=txt(raw.participation_state)||"screen_visible";
    const item={...clone(raw),label,source_phrase:label,participation_state,count:Math.max(1,Number(raw.count||1)),entity_type:"temporary_character",manual_added_v7824:true,selection_source_v7824:"manual",ai_discovered:false};
    const list=arr(scene.manual_temporary_characters_v7824).filter(x=>tempName(x)!==label);list.push(item);scene.manual_temporary_characters_v7824=list;
    const current=arr(scene.temporary_characters).filter(x=>tempName(x)!==label);current.push(item);scene.temporary_characters=current;
    try{scheduleDraftSave?.();}catch(_){}return item;
  }
  function removeManualTemporary(scene={},label=""){
    scene.manual_temporary_characters_v7824=arr(scene.manual_temporary_characters_v7824).filter(x=>tempName(x)!==txt(label));
    scene.temporary_characters=arr(scene.temporary_characters).filter(x=>!(manualTemp(x)&&tempName(x)===txt(label)));
    shotsForScene(scene).forEach(shot=>{shot.temporary_characters=arr(shot.temporary_characters).filter(x=>!(manualTemp(x)&&tempName(x)===txt(label)));});
    try{scheduleDraftSave?.();}catch(_){}
  }
  function openTemporaryAddDialog(scene,index,card){
    document.querySelector(".v7824-temp-dialog")?.remove();
    const dialog=document.createElement("dialog");dialog.className="v7824-temp-dialog";
    const pool=temporaryPool(scene);
    const opts=pool.map((x,i)=>`<option value="${i}">${htmlEsc(tempName(x))}${humanTempType(x)?` · ${htmlEsc(humanTempType(x))}`:""}</option>`).join("");
    dialog.innerHTML=`<form method="dialog" class="v7824-temp-dialog-body"><div class="field-head"><strong>添加本镜临时人物</strong><button value="cancel" class="btn secondary small-btn">关闭</button></div><p class="hint">可从其他分镜已由AI识别的临时人物借用，也可手工新增。这里只改变当前镜；正式人物仍只来自强制名单。</p><label>从其他镜借用<select data-temp-borrow><option value="">不借用</option>${opts}</select></label><label>临时人物/群体名称<input data-temp-name placeholder="例如：我妈、司机、沈家众人"></label><label>本镜类型<select data-temp-type><option value="single">单体出镜</option><option value="group">群体出镜</option><option value="background">背景群像</option><option value="offscreen">画外音</option><option value="mentioned">仅提及</option></select></label><div class="v7824-temp-dialog-actions"><button type="button" class="btn secondary" data-temp-add>仅添加</button><button type="button" class="btn primary" data-temp-add-regen>添加并重生本镜</button></div></form>`;
    document.body.appendChild(dialog);
    const borrow=dialog.querySelector("[data-temp-borrow]"),nameInput=dialog.querySelector("[data-temp-name]"),typeInput=dialog.querySelector("[data-temp-type]");
    borrow.addEventListener("change",()=>{const row=pool[Number(borrow.value)];if(row){nameInput.value=tempName(row);const ht=humanTempType(row);typeInput.value=ht==="群体出镜"?"group":ht==="背景出镜"?"background":ht==="画外音"?"offscreen":ht==="仅提及"?"mentioned":"single";}});
    const add=(regen=false)=>{const label=txt(nameInput.value)||tempName(pool[Number(borrow.value)]||{});if(!label){nameInput.focus();return;}const base=clone(pool[Number(borrow.value)]||{});upsertManualTemporary(scene,{...base,label,source_phrase:label,participation_state:normalizeManualParticipation(typeInput.value),count:typeInput.value==="group"?Math.max(2,Number(base.count||2)):Math.max(1,Number(base.count||1))});dialog.close();dialog.remove();renderTemporaryPanels();if(regen){const guidance=card.querySelector('[data-field="guidance"]');if(guidance){const addition=`【人工补充本镜临时人物】必须把“${label}”按${humanTempType({label,participation_state:normalizeManualParticipation(typeInput.value),count:typeInput.value==="group"?2:1})}纳入当前镜；AI仍需结合当前原文重新判断其他临时人物。写清该人物与现有主体的实际站位、动作或画外音状态，不改变正式人物名单。`;guidance.value=[txt(guidance.value),addition].filter(Boolean).join("\n");guidance.dispatchEvent(new Event("input",{bubbles:true}));}card.querySelector(".generate-one")?.click();}};
    dialog.querySelector("[data-temp-add]").addEventListener("click",()=>add(false));dialog.querySelector("[data-temp-add-regen]").addEventListener("click",()=>add(true));
    dialog.addEventListener("close",()=>dialog.remove());dialog.showModal();
  }
  const BLOCKING_ANCHOR_RE=/(左|右|中央|中间|前景|中景|后景|门口|门边|窗边|床边|床侧|桌前|桌边|书案|沙发|走廊|入口|出口|身侧|身后|对面|旁边|墙边|台阶|院中|庭院|门槛|廊下)/;
  const BAD_BLOCKING_RE=/(人物当前站位|人物所在位置|人物保持站位|位于当前位置|人物在场)/;
  function blockingHealthFor(scene={},index=0){
    const prompt=shotsForScene(scene).map(s=>txt(s.prompt||s.visual_prompt||s.description)).join("\n")||txt(scene.visual_prompt||scene.prompt||scene.visual_context);
    const visibleFormal=formalSlotIds(scene).length>0,visibleTemps=temporaryForScene(scene,index).filter(visibility).length>0,hasPeople=visibleFormal||visibleTemps;
    const issues=[];if(BAD_BLOCKING_RE.test(prompt))issues.push("存在占位式人物站位");if(!hasPeople&&/(人物|站位)/.test(prompt)&&BAD_BLOCKING_RE.test(prompt))issues.push("无人镜头错误套用人物模板");if(hasPeople&&prompt&&!BLOCKING_ANCHOR_RE.test(prompt))issues.push("人物站位缺少可拍空间锚点");
    const filler=(prompt.match(/高级感|电影感|氛围感|氛围拉满|环境保持一致|人物在场/g)||[]).length;if(filler>=2)issues.push("低信息量修饰偏多");
    return {ok:issues.length===0,issues,prompt_chars:prompt.length};
  }
  function renderTemporaryTextPanel(scene,index,card,parent){
    const temps=temporaryForScene(scene,index).filter(t=>visibility(t)||manualTemp(t));
    const panel=document.createElement("section");panel.className="v783-temporary-text-panel v7824-temporary-panel";panel.dataset.sourceKey=sourceKey(scene,index);
    const head=document.createElement("div");head.className="field-head";const title=document.createElement("strong");title.textContent="本镜临时人物（AI主判）";const hint=document.createElement("span");hint.className="hint";hint.textContent="AI读取当前原文判断；本地只做正式人物去重和质量提示。可手工借用/补充后只重生本镜。";const addBtn=document.createElement("button");addBtn.type="button";addBtn.className="btn secondary small-btn";addBtn.textContent="+ 添加临时人物";addBtn.addEventListener("click",()=>openTemporaryAddDialog(scene,index,card));head.append(title,hint,addBtn);panel.appendChild(head);
    const list=document.createElement("div");list.className="v783-temporary-text-list";
    if(!temps.length){const empty=document.createElement("p");empty.className="hint";empty.textContent="AI本镜未判定需要临时人物；如剧情确实需要，可手工添加后单条重生。";list.appendChild(empty);}else temps.forEach(temp=>{const h=performanceHealthFor(scene,temp,index),row=document.createElement("div");row.className=`v783-temporary-text-row ${h.level}`;const titleEl=document.createElement("strong");titleEl.textContent=tempName(temp);const badges=document.createElement("div");badges.className="v7824-temp-badges";for(const label of [manualTemp(temp)?"人工添加":"AI主判",humanTempType(temp)]){const b=document.createElement("span");b.className="v7824-temp-badge";b.textContent=label;badges.appendChild(b);}const perf=document.createElement("span");perf.className=`v781-performance-state ${h.level}`;perf.textContent=h.level==="pass"?"动作/互动已落实":h.message;row.append(titleEl,badges,perf);if(manualTemp(temp)){const remove=document.createElement("button");remove.type="button";remove.className="btn secondary small-btn";remove.textContent="移除";remove.addEventListener("click",()=>{removeManualTemporary(scene,tempName(temp));renderTemporaryPanels();});row.appendChild(remove);}list.appendChild(row);});panel.appendChild(list);
    const issues=temps.filter(visibility).map(t=>performanceHealthFor(scene,t,index)).filter(x=>x.level!=="pass"),blocking=blockingHealthFor(scene,index);
    if(issues.length||!blocking.ok){const foot=document.createElement("div");foot.className="v781-reference-health";const messages=[];if(issues.length)messages.push(`${issues.length} 个临时人物动作/交互需要加强`);if(!blocking.ok)messages.push(`站位/画面：${blocking.issues.join("、")}`);const msg=document.createElement("span");msg.textContent=messages.join("；");const repair=document.createElement("button");repair.type="button";repair.className="btn secondary small-btn";repair.textContent="按当前人物重生本镜";repair.dataset.taskType=TEMP_PERFORMANCE_REPAIR_TASK;repair.addEventListener("click",()=>{const guidance=card.querySelector('[data-field="guidance"]');if(guidance){const addition="【站位与画面内容修复】保留当前镜正确剧情事实和人物选择；AI重新判断本镜临时人物。先判断是否有人物可见；无人则只写具体环境与可拍变化。有人的镜头用左/中/右、前/中/后景和门口/窗边/桌前等真实空间锚点写清站位、朝向、视线、移动和互动；删除人物当前站位等占位词以及低信息量空话，优先落实原文动作、道具变化和直接反应。";if(!txt(guidance.value).includes("【站位与画面内容修复】")){guidance.value=[txt(guidance.value),addition].filter(Boolean).join("\n");guidance.dispatchEvent(new Event("input",{bubbles:true}));}}card.querySelector(".generate-one")?.click();});foot.append(msg,repair);panel.appendChild(foot);}
    parent.appendChild(panel);
  }
  function renderTemporaryPanels(){
    document.querySelectorAll(".v7824-temp-panel-host").forEach(n=>n.remove());
    document.querySelectorAll("#scenes .scene-card[data-scene-id]").forEach((card,index)=>{const scene=arr(state.scenes).find(x=>txt(x.id)===txt(card.dataset.sceneId))||arr(state.scenes)[index];if(!scene)return;const mapped=card.querySelector(".mapped-outline");const panelHost=document.createElement("div");panelHost.className="v7824-temp-panel-host";(mapped?.parentNode||card).insertBefore(panelHost,mapped||null);renderTemporaryTextPanel(scene,index,card,panelHost);});
  }
  function renderReferenceBoxes(){
    if(!premium()){document.querySelectorAll(".v781-reference-box").forEach(n=>n.remove());return;}
    refreshBindings({persist:false});
    document.querySelectorAll("#scenes .scene-card[data-scene-id]").forEach((card,index)=>{
      const scene=arr(state.scenes).find(x=>txt(x.id)===txt(card.dataset.sceneId))||arr(state.scenes)[index];if(!scene)return;const entry=ensureEntry(scene,index);
      card.querySelectorAll(".scene-reference-box:not(.v781-reference-box)").forEach(n=>n.remove());
      let box=card.querySelector(".v781-reference-box");if(!box){box=document.createElement("section");box.className="scene-reference-box v781-reference-box";const mapped=card.querySelector(".mapped-outline");(mapped?.parentNode||card).insertBefore(box,mapped||null);}box.innerHTML="";
      const head=document.createElement("div");head.className="field-head";head.innerHTML='<strong>本镜参考图（正式人物 / 场景 / 道具）</strong><span class="hint">图片只用于最终导出引用；分镜AI完全不读取图片、图片URL、图片状态或 @图N</span>';box.appendChild(head);
      const grid=document.createElement("div");grid.className="v781-reference-grid";
      const formalIds=formalSlotIds(scene);formalIds.forEach(id=>{const slot=slotById(id),asset=slot?host().characterAsset?.(id):null;if(slot)grid.appendChild(renderReferenceCard({scene,index,type:"formal",asset,slot,entry,readonly:true}));});
      const sceneIds=uniq([...entry.scene.auto,...entry.scene.manual_added,...entry.scene.suggested,...arr(state.referenceAssets?.scenes).map(a=>a.asset_id)]);sceneIds.forEach(id=>{const asset=assetById(id);if(asset)grid.appendChild(renderReferenceCard({scene,index,type:"scene",asset,entry}));});
      const propIds=uniq([...entry.prop.auto,...entry.prop.manual_added,...entry.prop.suggested,...arr(state.referenceAssets?.props).map(a=>a.asset_id)]);propIds.forEach(id=>{const asset=assetById(id);if(asset)grid.appendChild(renderReferenceCard({scene,index,type:"prop",asset,entry}));});
      if(!grid.children.length){const empty=document.createElement("p");empty.className="hint";empty.textContent="当前镜暂无可匹配的正式人物/场景/道具参考图。临时人物始终使用文字动态表达。";grid.appendChild(empty);}box.appendChild(grid);
      const missing=selectedMissingForScene(scene,index);if(missing.length){const foot=document.createElement("div");foot.className="v781-reference-health";foot.textContent=`${missing.length} 个已选参考资产没有主图`;box.appendChild(foot);}
    });
  }
  function usageCount(assetId){let n=0;arr(state.scenes).forEach((scene,index)=>{const e=ensureEntry(scene,index);if(["scene","prop"].some(t=>effective(e[t]).includes(assetId)))n++;});return n;}
  function renderLibrary(){
    if(!premium())return;const box=document.querySelector("#premiumAssetLibraryList");if(!box)return;box.innerHTML="";
    const assets=[...arr(state.referenceAssets?.scenes),...arr(state.referenceAssets?.props)];
    if(!assets.length){box.innerHTML='<span class="hint">尚未添加场景/道具参考资产。临时人物不使用参考图，只通过文字动态表达。</span>';return;}
    assets.forEach(asset=>{const type=asset.type;const card=document.createElement("button");card.type="button";card.className=`v781-library-card type-${type}`;card.appendChild(renderThumb(asset,asset.name));const meta=document.createElement("span");meta.className="v781-library-meta";const title=document.createElement("strong");title.textContent=txt(asset.name)||"未命名";const sub=document.createElement("span");sub.textContent=`${type==="scene"?"场景":"道具"} · ${asset.has_main_image?"已有主图":"暂无主图"} · ${usageCount(asset.asset_id)}镜使用`;meta.append(title,sub);card.appendChild(meta);card.addEventListener("click",()=>host().openAssetDialog?.(asset.type,asset,asset.target_slot_id||""));box.appendChild(card);});
  }
  function sourceKeysForSegment(segment={}){
    const ids=[];const start=Number(segment.source_start??-1),end=Number(segment.source_end??-1);arr(state.outlineShots).forEach(shot=>{const ss=Number(shot.start_second??0),ee=Number(shot.end_second??ss);if(start>=0&&end>=0&&(ee<=start||ss>=end))return;for(const id of uniq([shot.parent_scene_id,...arr(shot.parent_scene_ids)])){const scene=arr(state.scenes).find(x=>txt(x.id)===id);if(scene){const key=sourceKey(scene,arr(state.scenes).indexOf(scene));if(key&&!ids.includes(key))ids.push(key);}}});
    if(!ids.length)uniq(segment.source_keys).forEach(k=>ids.push(k));return ids;
  }
  function scenesForSegment(segment={}){const keys=sourceKeysForSegment(segment);return keys.map(k=>arr(state.scenes).find((s,i)=>sourceKey(s,i)===k)).filter(Boolean);}
  function refsForSegment(segment={},rawText=""){
    refreshBindings({persist:false});const refs=[],slots=arr(state.characterCoreV2?.slots);const segmentSlotIds=uniq(segment.character_slot_ids);
    let formal=slots.filter(s=>segmentSlotIds.includes(txt(s.slot_id)));if(!formal.length){formal=slots.filter(s=>slotAliases(s).some(a=>a&&String(rawText||"").includes(a)));}
    formal.forEach(slot=>{const asset=host().characterAsset?.(slot.slot_id);if(asset?.has_main_image)refs.push({type:"character",entity_type:"formal_character",asset,slot,name:slotName(slot),aliases:slotAliases(slot),appearance:slotAppearance(slot)});});
    const seenAsset=new Set();
    scenesForSegment(segment).forEach(scene=>{const index=arr(state.scenes).indexOf(scene),selected=sceneAssetSelection(scene,index);
      for(const type of ["scene","prop"]){for(const id of selected[type]){if(seenAsset.has(id))continue;const asset=assetById(id);if(!asset?.has_main_image)continue;seenAsset.add(id);refs.push({type,entity_type:type,asset,name:txt(asset.name),aliases:aliasesOf(asset),state_variant:txt(asset.state_variant)});}}
    });
    const ordered=[...refs.filter(r=>r.type==="character"),...refs.filter(r=>r.type==="scene"),...refs.filter(r=>r.type==="prop")];return ordered.map((ref,index)=>({...ref,index:index+1}));
  }
  const GENERIC_PRONOUN_ALIAS=new Set(["我","你","他","她","它","我们","你们","他们","她们","男主","女主","主角"]);
  function stripReferenceTokens(source=""){return String(source||"").replace(/\(@图\d+\)/g,"");}
  function injectionCandidates(ref={}){
    const canonical=txt(ref.name);const aliases=uniq(ref.aliases).filter(x=>x!==canonical&&!GENERIC_PRONOUN_ALIAS.has(x));
    return uniq([canonical,...aliases]).filter(Boolean).sort((a,b)=>b.length-a.length);
  }
  function replaceTagFirst(result,names,index){
    let out=String(result||"");for(const name of uniq(names).sort((a,b)=>b.length-a.length)){if(!name)continue;const re=new RegExp(`${esc(name)}(?!\\(@图\\d+\\))`);const m=re.exec(out);if(m)return out.slice(0,m.index)+m[0]+`(@图${index})`+out.slice(m.index+m[0].length);}return out;
  }
  function injectRefs(source="",refs=[]){
    // Reference tokens are export annotations only. One entity is annotated once at its first
    // meaningful occurrence; later pronouns/subjects are preserved for natural Chinese continuity.
    let result=stripReferenceTokens(source);for(const ref of arr(refs))result=replaceTagFirst(result,injectionCandidates(ref),ref.index);return result;
  }
  function injectRefsLossless(source="",refs=[]){
    const master=stripReferenceTokens(source),referenced=injectRefs(master,refs),restored=stripReferenceTokens(referenced);
    return {text:restored===master?referenced:master,master,lossless:restored===master,reference_count:arr(refs).length};
  }
  function missingRefsForSegment(segment={}){
    const out=[],seen=new Set();scenesForSegment(segment).forEach(scene=>{const index=arr(state.scenes).indexOf(scene);selectedMissingForScene(scene,index).forEach(x=>{if(seen.has(x.asset_id))return;seen.add(x.asset_id);out.push(x);});});return out;
  }
  function renderSegmentPreviews(){
    if(!premium()){document.querySelectorAll(".v781-segment-reference-preview").forEach(n=>n.remove());return;}
    document.querySelectorAll("#segments .segment-card").forEach((card,index)=>{const segment=arr(state.segments)[index];if(!segment)return;card.querySelector(".v781-segment-reference-preview")?.remove();const refs=arr(segment._premium_refs).length?arr(segment._premium_refs):refsForSegment(segment,segment.current_text||segment.text||"");const missing=missingRefsForSegment(segment);const wrap=document.createElement("section");wrap.className="v781-segment-reference-preview";const title=document.createElement("div");title.className="field-head";title.innerHTML='<strong>本段使用参考图</strong><span class="hint">顺序固定：正式人物 → 场景 → 道具，与 @图N / 图文复制完全一致；临时人物不进入图片索引</span>';wrap.appendChild(title);const grid=document.createElement("div");grid.className="v781-segment-ref-grid";refs.forEach(ref=>{const item=document.createElement("div");item.className=`v781-segment-ref-item type-${ref.type}`;item.appendChild(renderThumb(ref.asset,ref.name));const t=document.createElement("span");t.textContent=`图${ref.index} · ${ref.name}`;const k=document.createElement("small");k.textContent=ref.type==="character"?"正式人物":ref.type==="scene"?"场景":"道具";item.append(t,k);grid.appendChild(item);});if(!refs.length){const e=document.createElement("span");e.className="hint";e.textContent="本段暂无可用主参考图";grid.appendChild(e);}wrap.appendChild(grid);if(missing.length){const warn=document.createElement("p");warn.className="v781-reference-warning";warn.textContent=`已选择但无主图：${missing.map(x=>x.asset?.name||x.asset_id).join("、")}。这些资产不会生成 @图N，人物/动作/场景文字仍完整保留。`;wrap.appendChild(warn);}const masterSvc=globalThis.__V78_MASTER_PROMPT_SERVICE__;const q=segment._premium_reference_quality||masterSvc?.segmentComparison?.(segment,refs);if(q){const parity=document.createElement("details");parity.className=`v782-master-parity ${q.ok===false?"warn":"pass"}`;const sum=document.createElement("summary");const mr=q.metrics?.master||{},rr=q.metrics?.referenced||{};sum.textContent=`普通母版对照 · ${q.ok===false?"已启用无损回退":"语义无损"} · 动作 ${mr.actions||0}→${rr.actions||0} · 画面密度 ${mr.richness||0}→${rr.richness||0}`;parity.appendChild(sum);const pre=document.createElement("pre");pre.className="v782-master-preview";pre.textContent=q.master||segment._premium_master_prompt||"";parity.appendChild(pre);wrap.appendChild(parity);}const textarea=card.querySelector(".segment-text");card.insertBefore(wrap,textarea||null);});
  }
  function sourceClause(source,label){const s=txt(source);if(!s)return "";const p=s.indexOf(label);if(p<0)return s.slice(0,180);let a=p,b=p;while(a>0&&!/[。！？；\n]/.test(s[a-1]))a--;while(b<s.length&&!/[。！？；\n]/.test(s[b]))b++;return s.slice(a,Math.min(s.length,b+1)).trim().slice(0,260);}
  function performanceFactsFromPlans(plans=[]){
    const rows=[];arr(plans).slice(0,120).forEach((plan,i)=>{const source=txt(plan.source_text),key=txt(plan.source_key)||`line_${i+1}`;arr(plan.temporary_characters).filter(visibility).slice(0,12).forEach(raw=>{const label=tempName(raw),evidence=sourceClause(source,label);rows.push({source_key:key,temporary_entity_id:txt(raw.temporary_entity_id||raw.entity_id||raw.continuity_id)||`temp_${hash(`${key}|${label}`)}`,label,scene_role:txt(raw.scene_role)||(/说|问|喊|回答/.test(evidence)?"对话人物":ACTION_RE.test(evidence)?"主动作人物":"反应/陪衬人物"),visual_identity:{continuity_id:txt(raw.continuity_id),continuity_profile:txt(raw.continuity_profile),wardrobe_state:txt(raw.wardrobe_state)},performance:{source_evidence:evidence,requires_action_or_interaction:true}});});});return rows;
  }
  const PERFORMANCE_RULES=`【V78.2.6 临时人物AI主判与站位规则】临时人物由AI直接读取当前source_text判断；本地识别只做诊断，不作为必达人物。只有用户人工添加项是强制补充。可见临时人物必须写成可拍动作与具体站位，不输出内部ID。`;
  function installPromptRules(){
    // V78.2.6: no local temporary-candidate facts are appended to the initial outline prompt.
    // The authoritative AI-primary rules live in app.js / Scene Context. This avoids duplicate rules and local-heuristic bias.
  }
  function renderAllReferenceUI(){renderTemporaryPanels();if(!premium())return;refreshBindings({persist:false});renderLibrary();renderReferenceBoxes();renderSegmentPreviews();}
  function installRenderHooks(){
    if(typeof globalThis.renderScenes==="function"&&!globalThis.renderScenes.__v781Refs){const base=globalThis.renderScenes;const wrapped=function v781RenderScenes(){const result=base.apply(this,arguments);queueMicrotask(()=>{renderTemporaryPanels();if(premium()){refreshBindings({persist:false});renderLibrary();renderReferenceBoxes();}});return result;};wrapped.__v781Refs=true;wrapped.__base=base;globalThis.renderScenes=wrapped;try{renderScenes=wrapped;}catch(_){}}
    if(typeof globalThis.renderSegments==="function"&&!globalThis.renderSegments.__v781Refs){const base=globalThis.renderSegments;const wrapped=function v781RenderSegments(){const result=base.apply(this,arguments);queueMicrotask(()=>{renderTemporaryPanels();if(premium()){refreshBindings({persist:false});renderReferenceBoxes();renderSegmentPreviews();renderLibrary();}});return result;};wrapped.__v781Refs=true;wrapped.__base=base;globalThis.renderSegments=wrapped;try{renderSegments=wrapped;}catch(_){}}
  }
  function installPersistence(){
    if(typeof globalThis.getProjectData==="function"&&!globalThis.getProjectData.__v781Refs){const base=globalThis.getProjectData;const wrapped=function v781GetProjectData(){const data=base.apply(this,arguments)||{};return {...data,reference_bindings_v781:clone(ensureStore())};};wrapped.__v781Refs=true;wrapped.__base=base;globalThis.getProjectData=wrapped;try{getProjectData=wrapped;}catch(_){}}
    if(typeof globalThis.applyProjectData==="function"&&!globalThis.applyProjectData.__v781Refs){const base=globalThis.applyProjectData;const wrapped=function v781ApplyProjectData(projectId,name,data={}){const result=base.apply(this,arguments);state.referenceBindingsV781=clone(obj(data.reference_bindings_v781));ensureStore();setTimeout(()=>{refreshBindings({persist:false});renderAllReferenceUI();},120);return result;};wrapped.__v781Refs=true;wrapped.__base=base;globalThis.applyProjectData=wrapped;try{applyProjectData=wrapped;}catch(_){}}
  }
  function bindingHealth(){
    refreshBindings({persist:false});const issues=[];arr(state.scenes).forEach((scene,index)=>{const key=sourceKey(scene,index),missing=selectedMissingForScene(scene,index);missing.forEach(x=>issues.push({level:"warn",code:"SELECTED_REFERENCE_WITHOUT_MAIN",source_key:key,asset_id:x.asset_id,message:`${x.asset?.name||x.asset_id} 已选择但没有主参考图`}));const perf=temporaryForScene(scene,index).filter(visibility).map(t=>({t,h:performanceHealthFor(scene,t,index)}));perf.filter(x=>x.h.level!=="pass").forEach(x=>issues.push({level:"warn",code:"TEMPORARY_PERFORMANCE_WEAK",source_key:key,temporary_entity_id:x.t.temporary_entity_id,message:x.h.message}));const blocking=blockingHealthFor(scene,index);if(!blocking.ok)issues.push({level:"warn",code:"BLOCKING_VISUAL_WEAK",source_key:key,message:blocking.issues.join("、")});});
    return {ok:!issues.some(x=>x.level==="error"),version:VERSION,issues,performance:performanceHealth(),store:clone(ensureStore())};
  }
  const resolver={version:VERSION,storeVersion:STORE_VERSION,temporaryPerformanceRepairTask:TEMP_PERFORMANCE_REPAIR_TASK,ensureStore,refreshBindings,render:renderAllReferenceUI,refsForSegment,injectRefs,injectRefsLossless,stripReferenceTokens,missingRefsForSegment,temporaryForScene,temporaryPool,upsertManualTemporary,removeManualTemporary,blockingHealthFor,performanceHealth,bindingHealth,setBinding,onAssetRemoved,snapshot:()=>({version:VERSION,store:clone(ensureStore()),health:bindingHealth(),assets:{formal:arr(state.referenceAssets?.characters).filter(a=>txt(a.entity_type)!=="temporary_character").length,legacy_temporary_hidden:arr(state.referenceAssets?.characters).filter(a=>txt(a.entity_type)==="temporary_character").length,scenes:arr(state.referenceAssets?.scenes).length,props:arr(state.referenceAssets?.props).length}})};
  ns.referenceEntities=resolver;globalThis.__V78_REFERENCE_RESOLVER__=resolver;
  installPromptRules();installRenderHooks();installPersistence();
  const boot=()=>{try{refreshBindings({persist:false});renderAllReferenceUI();}catch(error){console.warn("[V78.2.6 Reference Entities] boot warning",error);}};
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",()=>setTimeout(boot,1050),{once:true});else setTimeout(boot,1050);
  console.info("[V78.2.6 Clean Core] reference images decoupled from temporary people",{version:VERSION,principle:"formal_scene_prop_images_export_only;temporary_people_text_only"});
})();

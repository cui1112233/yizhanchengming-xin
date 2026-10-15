/* CharacterCore 2.0 v77 hotfix19 — single cast executor, stable source_key writeback, structured temporary entities. */
(function installCharacterCoreV2() {
  "use strict";
  const PROTOCOL = "character_core_2_all_genres_v77";
  const VERSION = 77;
  const BUILD_VERSION = "v78.3.0.2";
  const FACTS_MARKER = "V77_CHARACTER_CORE_FACTS";
  const APPEARANCE_MARKER = "V77_CHARACTER_CORE_APPEARANCE";
  const REVISION_CHECK_MARKER = "V77_CHARACTER_CORE_REVISION_CHECK";
  const STYLE_MARKER = "V77_CHARACTER_CORE_STYLE";
  const STAGES = ["女婴儿","男婴儿","小女孩","小男孩","少女","少年","年轻女性","年轻男性","中年女性","中年男性","老年女性","老年男性","婴幼儿阶段","儿童阶段","青少年阶段","成年阶段","中年阶段","老年阶段"];
  const DETAIL_LEVELS = ["简洁","标准","详细","极致"];
  const MAX_APPEARANCE = 6000;
  const MAX_NOTE = 5000;
  const APPEARANCE_CONCURRENCY = 2;
  const APPEARANCE_WRITEBACK_ERROR = "AI 没有返回可写入的人物外形正文。";
  const q = (s) => document.querySelector?.(s) || null;
  const qa = (s) => [...(document.querySelectorAll?.(s) || [])];
  const text = (v) => String(v ?? "").replace(/\s+/g, " ").trim();
  const object = (v) => v && typeof v === "object" && !Array.isArray(v) ? v : {};
  const array = (v) => Array.isArray(v) ? v : [];
  const unique = (xs) => [...new Set(xs.map(text).filter(Boolean))];
  const regexEscape = (value="") => String(value).replace(/[.*+?^${}()|[\]\\]/g,"\\$&");
  const deepClone = (v) => { try { return JSON.parse(JSON.stringify(v)); } catch (_) { return v; } };
  const appState = typeof state !== "undefined" ? state : (globalThis.state ||= {});
  const baseSaveProject = typeof saveProject === "function" ? saveProject : null;
  const normalizeNovelSource = (value = "") => typeof globalThis.__normalizeV77NovelSource === "function"
    ? globalThis.__normalizeV77NovelSource(value)
    : String(value ?? "").replace(/\r\n?/g,"\n").replace(/[\uFEFF\u200B\u200C\u200D\u2060]/g,"").split("\n").map(line=>line.replace(/[\t \u3000]+$/g,"")).join("\n").trim();
  const currentNovel = () => normalizeNovelSource(typeof readNovelTextRaw === "function" ? readNovelTextRaw() : q("#novelText")?.value);
  const valueOf = (selector, fallback = "") => typeof readFieldValue === "function" ? readFieldValue(selector, fallback) : (q(selector)?.value ?? fallback);
  const setValue = (selector, value) => { if (typeof writeFieldValue === "function") writeFieldValue(selector, value); else if (q(selector)) q(selector).value = value ?? ""; };
  const fingerprint = (v = "") => {
    if (typeof buildCharacterSourceFingerprint === "function") return buildCharacterSourceFingerprint(v);
    let h = 2166136261; for (const c of String(v)) { h ^= c.charCodeAt(0); h = Math.imul(h, 16777619); } return (h >>> 0).toString(36);
  };
  const currentInstance = () => {
    try { return new URLSearchParams(globalThis.location?.search || "").get("instance") || new URLSearchParams(globalThis.location?.search || "").get("session") || "instance-browser"; }
    catch (_) { return "instance-browser"; }
  };
  const formalRosterApi = () => globalThis.__V78_FORMAL_ROSTER_STORE__ || globalThis.__V78_FORMAL_ROSTER__ || null;
  function formalRosterSnapshot(reason="read", forceDom=true) {
    const api=formalRosterApi();
    if(api?.snapshot) return api.snapshot({forceDom,reason});
    const raw=String(q("#characterGuideInput")?.value ?? appState.character_guide_input ?? "");
    return Object.freeze({version:"legacy_roster_fallback",raw_text:raw,normalized_text:raw.trim(),entries:raw.split(/[\n；;，,、|｜]+/).map(x=>String(x||"").trim()).filter(Boolean),revision:0,hash:fingerprint(raw.trim()),snapshot_reason:reason});
  }
  function assertFormalRosterFresh(snapshot, stage="人物生成") {
    const api=formalRosterApi();
    if(api?.assertFresh) return api.assertFresh(snapshot,stage);
    return true;
  }
  function showFormalRosterError(message="") {
    const node=q("#formalRosterError");
    if(node){node.textContent=String(message||"");node.hidden=!message;}
    if(message){const guide=q("#characterGuideInput");try{guide?.scrollIntoView?.({behavior:"smooth",block:"center"});guide?.focus?.();}catch(_){}}
  }

  function ensureCore() {
    // Keep one stable object for the lifetime of the page. Reconstructing the
    // root object on every read makes previously captured references stale and
    // silently discards later UI edits (for example alias scope changes).
    let existing = appState.characterCoreV2;
    if (!existing || typeof existing !== "object" || Array.isArray(existing)) existing = object(appState.character_core_v2);
    if (!existing || typeof existing !== "object" || Array.isArray(existing)) existing = {};
    existing.character_core_version = 2;
    existing.source_hash = text(existing.source_hash);
    existing.character_revision = Number(existing.character_revision || 0);
    existing.formal_roster_stale = Boolean(existing.formal_roster_stale);
    existing.people = array(existing.people);
    existing.slots = array(existing.slots);
    existing.relationships = array(existing.relationships);
    existing.aliases = array(existing.aliases);
    existing.mention_entities = array(existing.mention_entities);
    existing.scene_casting = object(existing.scene_casting);
    existing.relationship_pending = array(existing.relationship_pending);
    existing.temporary_continuity_cache = object(existing.temporary_continuity_cache);
    existing.temporary_continuity_cache_character_core_validator = object(existing.temporary_continuity_cache_character_core_validator);
    existing.temporary_continuity_cache_v78_primary = object(existing.temporary_continuity_cache_v78_primary);
    existing.temporary_continuity_authority_v78 = object(existing.temporary_continuity_authority_v78);
    existing.relationship_conflicts = array(existing.relationship_conflicts);
    existing.keyword_index = object(existing.keyword_index);
    existing.analysis_state = object(existing.analysis_state);
    existing.analysis_state.protocol = PROTOCOL;
    existing.analysis_state.source_hash = text(existing.analysis_state.source_hash || existing.source_hash);
    existing.analysis_state.style_ready = Boolean(existing.analysis_state.style_ready);
    existing.analysis_state.slots_ready = Boolean(existing.analysis_state.slots_ready || existing.analysis_state.forced_roster_ready);
    existing.analysis_state.forced_roster_ready = existing.analysis_state.slots_ready;
    existing.analysis_state.facts_ready = Boolean(existing.analysis_state.facts_ready);
    // Hotfix31: old builds sometimes set relationship_graph_ready=true merely because
    // fallback人物节点 existed. Do not migrate that historical fallback flag into a
    // false "AI关系已完成" state unless there is actual relationship/alias data.
    const hadExplicitRelationshipsReady = Object.prototype.hasOwnProperty.call(existing.analysis_state, "relationships_ready");
    const legacyRelationshipDataPresent = Boolean(existing.relationships.length || existing.aliases.length);
    existing.analysis_state.relationships_ready = hadExplicitRelationshipsReady
      ? Boolean(existing.analysis_state.relationships_ready)
      : Boolean(existing.analysis_state.relationship_graph_ready && legacyRelationshipDataPresent);
    if (!existing.analysis_state.relationships_ready && existing.analysis_state.relationship_graph_ready && !legacyRelationshipDataPresent) {
      existing.analysis_state.relationship_fallback_available = true;
    }
    existing.analysis_state.relationship_graph_ready = Boolean(existing.analysis_state.relationships_ready || legacyRelationshipDataPresent);
    existing.analysis_state.appearances_ready = Boolean(existing.analysis_state.appearances_ready || existing.analysis_state.character_cards_ready);
    existing.analysis_state.character_cards_ready = existing.analysis_state.appearances_ready;
    existing.analysis_state.keyword_index_ready = Boolean(existing.analysis_state.keyword_index_ready || existing.analysis_state.casting_ready);
    existing.analysis_state.casting_ready = existing.analysis_state.keyword_index_ready;
    existing.analysis_state.scene_cast_ready = Boolean(existing.analysis_state.scene_cast_ready);
    existing.analysis_state.ready = Boolean(existing.analysis_state.ready);
    existing.analysis_state.source_revision = Math.max(1, Number(existing.analysis_state.source_revision || existing.source_snapshot?.source_revision || appState.sourceSnapshotV77?.source_revision || 1));
    existing.analysis_state.style_revision = Number(existing.analysis_state.style_revision || 0);
    existing.analysis_state.roster_revision = Number(existing.analysis_state.roster_revision || 0);
    existing.analysis_state.cards_revision = Number(existing.analysis_state.cards_revision || 0);
    existing.analysis_state.relationship_revision = Number(existing.analysis_state.relationship_revision || 0);
    existing.analysis_state.keyword_revision = Number(existing.analysis_state.keyword_revision || 0);
    existing.source_snapshot = object(existing.source_snapshot);
    existing.semantic_layer_version = BUILD_VERSION;
    existing.instance_id = currentInstance();
    // V78.3.0: appearance text is no longer allowed to promote unresolved AI facts
    // into an authoritative gender/stage. Old appearance_writeback values remain
    // visible as legacy evidence, but CharacterFactsCompleteness requires a real
    // manual/declared/AI-facts stage before downstream appearance/reference work.
    for (const slot of existing.slots) sanitizeManualIdentityValues(slot);
    appState.characterCoreV2 = existing;
    appState.character_core_v2 = existing;
    return existing;
  }
  const core = () => ensureCore();

  function captureSourceSnapshot(reason = "read") {
    const external = typeof globalThis.__captureV77SourceSnapshot === "function"
      ? globalThis.__captureV77SourceSnapshot(reason, typeof readNovelTextRaw === "function" ? readNovelTextRaw() : q("#novelText")?.value)
      : null;
    const canonical = normalizeNovelSource(external?.canonical_text ?? currentNovel());
    const sourceHash = text(external?.source_hash) || fingerprint(canonical);
    const c = core();
    const previous = object(c.source_snapshot);
    const changed = Boolean(previous.source_hash && previous.source_hash !== sourceHash);
    const revision = Math.max(1, Number(external?.source_revision || previous.source_revision || 1) + (changed && !external?.changed ? 1 : 0));
    const snapshot = {
      raw_text: String(external?.raw_text ?? (typeof readNovelTextRaw === "function" ? readNovelTextRaw() : q("#novelText")?.value) ?? ""),
      canonical_text: canonical,
      source_hash: sourceHash,
      source_revision: revision,
      changed: Boolean(external?.changed || changed),
      change_reason: text(reason),
      updated_at: new Date().toISOString(),
    };
    c.source_snapshot = snapshot;
    if(!c.source_hash)c.source_hash = sourceHash;
    if(!c.analysis_state.source_hash)c.analysis_state.source_hash = c.source_hash;
    c.analysis_state.source_revision = revision;
    appState.sourceSnapshotV77 = deepClone(snapshot);
    return snapshot;
  }
  function setStageRevision(stage, revision = captureSourceSnapshot(stage).source_revision) {
    const st = analysisState();
    const key = ({style:"style_revision",slots:"roster_revision",facts:"relationship_revision",relationships:"relationship_revision",appearances:"cards_revision",casting:"keyword_revision"})[stage];
    if(key) st[key] = Number(revision || 0);
    return st;
  }
  function revisionDiagnostics(snapshot = captureSourceSnapshot("diagnostics")) {
    const st = analysisState();
    return {
      source_revision:Number(snapshot.source_revision||0),
      style_revision:Number(st.style_revision||0),
      roster_revision:Number(st.roster_revision||0),
      cards_revision:Number(st.cards_revision||0),
      relationship_revision:Number(st.relationship_revision||0),
      keyword_revision:Number(st.keyword_revision||0),
      source_hash:text(snapshot.source_hash),
      build:BUILD_VERSION,
    };
  }

  function analysisState() { return core().analysis_state; }
  function setAnalysisStage(stage, ready, message = "") {
    const c = core(), st = c.analysis_state;
    st[`${stage}_ready`] = Boolean(ready);
    if(stage === "slots") st.forced_roster_ready = Boolean(ready);
    if(stage === "relationships") st.relationship_graph_ready = Boolean(ready);
    if(stage === "appearances") st.character_cards_ready = Boolean(ready);
    if(stage === "casting") st.keyword_index_ready = Boolean(ready);
    if(stage === "scene_cast") st.scene_cast_ready = Boolean(ready);
    st.current_stage = stage;
    st.last_message = text(message);
    st.last_updated_at = new Date().toISOString();
    const snapshot = captureSourceSnapshot(`stage_${stage}`);
    st.source_hash = snapshot.source_hash;
    st.source_revision = snapshot.source_revision;
    if (ready) setStageRevision(stage, snapshot.source_revision);
    if (!ready) st.ready = false;
    return st;
  }
  function resetAnalysisState(reason = "") {
    const c = core();
    const snapshot = captureSourceSnapshot(`reset:${reason}`);
    c.analysis_state = { protocol:PROTOCOL, source_hash:snapshot.source_hash, source_revision:snapshot.source_revision, style_revision:0, roster_revision:0, cards_revision:0, relationship_revision:0, keyword_revision:0, style_ready:false, slots_ready:false, forced_roster_ready:false, facts_ready:false, relationships_ready:false, relationship_graph_ready:false, facts_fallback_available:false, relationship_fallback_available:false, appearances_ready:false, character_cards_ready:false, keyword_index_ready:false, casting_ready:false, scene_cast_ready:false, ready:false, current_stage:"", last_message:text(reason), last_updated_at:new Date().toISOString() };
    c.keyword_index = {};
    c.scene_casting = {};
    return c.analysis_state;
  }

  function invalidateRosterStages(reason = "强制人物名单已修改") {
    const c=core(),st=analysisState(),snapshot=captureSourceSnapshot("roster_changed");
    // V78.2.8: roster editing invalidates authority, but does NOT destroy the
    // previous Person/Slot objects. They are preservation cache for the next
    // reconcile so unchanged entries retain stable ids/manual values/appearance.
    c.formal_roster_stale=true;c.scene_casting={};c.relationship_pending=[];c.keyword_index={};c.character_revision+=1;
    st.source_hash=c.source_hash||snapshot.source_hash;st.source_revision=snapshot.source_revision;
    st.slots_ready=false;st.forced_roster_ready=false;st.facts_ready=false;st.relationships_ready=false;st.relationship_graph_ready=false;st.facts_fallback_available=false;st.relationship_fallback_available=false;st.appearances_ready=false;st.character_cards_ready=false;st.keyword_index_ready=false;st.casting_ready=false;st.scene_cast_ready=false;st.ready=false;
    st.roster_revision=0;st.cards_revision=0;st.relationship_revision=0;st.keyword_revision=0;st.current_stage="roster_changed";st.last_message=text(reason);st.last_updated_at=new Date().toISOString();
    return st;
  }

  function normalizeGender(value = "") {
    const raw = text(value).toLowerCase();
    if (!raw || /待确认|未定|未知|待ai/.test(raw)) return "待确认";
    if (/无性别|不适用|无生理性别|genderless|agender/.test(raw)) return "无性别";
    if (/女性|女生|女人|女子|女孩|少女|女婴|female|woman|girl/.test(raw) || raw === "女") return "女";
    if (/男性|男生|男人|男子|男孩|少年|男婴|male|man|boy/.test(raw) || raw === "男") return "男";
    return "待确认";
  }
  function isUnresolvedGender(value = "") {
    return normalizeGender(value) === "待确认";
  }
  function isUnresolvedStage(value = "") {
    const raw = text(value).replace(/[〔〕【】\[\]（）()]/g, "");
    return !raw || /待AI|待确认|未定|未知/i.test(raw) || !normalizeStage(raw);
  }
  function sanitizeManualIdentityValues(slot) {
    const manual = object(slot.manual_values);
    if (isUnresolvedGender(manual.gender)) delete manual.gender;
    if (isUnresolvedStage(manual.visual_age_stage)) delete manual.visual_age_stage;
    slot.manual_values = manual;
    return manual;
  }
  function normalizeStage(value = "", gender = "") {
    const raw = text(value).replace(/[〔〕【】\[\]（）()]/g, "");
    if (!raw || /待AI|待确认|未定|未知/i.test(raw)) return "";
    const g = normalizeGender(gender);
    if (STAGES.includes(raw)) {
      const neutralFamily={"婴幼儿阶段":"infant","儿童阶段":"child","青少年阶段":"teen","成年阶段":"adult","中年阶段":"middle","老年阶段":"elder"}[raw];
      const table={infant:{女:"女婴儿",男:"男婴儿"},child:{女:"小女孩",男:"小男孩"},teen:{女:"少女",男:"少年"},adult:{女:"年轻女性",男:"年轻男性"},middle:{女:"中年女性",男:"中年男性"},elder:{女:"老年女性",男:"老年男性"}};
      return neutralFamily&&["男","女"].includes(g)?table[neutralFamily][g]:raw;
    }
    const map = [
      [/女婴|女宝宝|女新生儿/,"女婴儿"],[/男婴|男宝宝|男新生儿/,"男婴儿"],
      [/小女孩|女童|幼女/,"小女孩"],[/小男孩|男童|幼男/,"小男孩"],
      [/少女|未成年女性/,"少女"],[/少年|未成年男性/,"少年"],
      [/年轻女性|成年女性|青年女性/,"年轻女性"],[/年轻男性|成年男性|青年男性/,"年轻男性"],
      [/中年女性/,"中年女性"],[/中年男性/,"中年男性"],[/老年女性|年迈女性/,"老年女性"],[/老年男性|年迈男性/,"老年男性"],
    ];
    for (const [pattern, stage] of map) if (pattern.test(raw)) return stage;
    const family = /婴幼儿|婴儿/.test(raw) ? "infant" : /儿童|幼儿|孩童/.test(raw) ? "child" : /青少年|未成年/.test(raw) ? "teen" : /成年|年轻|青年/.test(raw) ? "adult" : /中年/.test(raw) ? "middle" : /老年|高龄|年迈/.test(raw) ? "elder" : "";
    const table = { infant:{女:"女婴儿",男:"男婴儿"}, child:{女:"小女孩",男:"小男孩"}, teen:{女:"少女",男:"少年"}, adult:{女:"年轻女性",男:"年轻男性"}, middle:{女:"中年女性",男:"中年男性"}, elder:{女:"老年女性",男:"老年男性"} };
    const neutral={infant:"婴幼儿阶段",child:"儿童阶段",teen:"青少年阶段",adult:"成年阶段",middle:"中年阶段",elder:"老年阶段"};
    return table[family]?.[g] || neutral[family] || "";
  }
  function genderFromStage(stage = "") {
    const s = normalizeStage(stage);
    if (["女婴儿","小女孩","少女","年轻女性","中年女性","老年女性"].includes(s)) return "女";
    if (["男婴儿","小男孩","少年","年轻男性","中年男性","老年男性"].includes(s)) return "男";
    return "待确认";
  }
  function visibleStage(slot) { return text(object(slot.age).visual_age_stage || slot.stage_label); }
  function normalizedVisibleStage(slot) { return normalizeStage(visibleStage(slot), slot.gender) || visibleStage(slot); }
  function slotDisplay(slot) {
    // Hotfix31: preserve the ordinary-mode/current-value priority (manual/current
    // display name first) and only clean aliases/stage decorations for display.
    // Premium mode must not silently replace a user's current character name with
    // an older canonical/base hint.
    const raw = text(object(slot?.manual_values).display_name || slot?.display_name || slot?.canonical_name_hint || slot?.base_name || slot?.name || slot?.slot_token);
    return raw.replace(/〔[^〕]*〕/g, "").replace(/（[^）]*）/g, "").replace(/\([^)]*\)/g, "").trim() || text(slot?.display_name || slot?.slot_token);
  }
  const APPEARANCE_ALIAS_NOISE = /^(?:我|你|他|她|它|某人|一位|这位|那位|这个人|那个人)$/;
  function splitAppearanceTokens(value = "") {
    return String(value || "").split(/[\s,，、\/|｜:：;；（）()【】\[\]<>《》"'“”‘’]+/).map(text).filter(Boolean);
  }
  function isAppearanceAliasNoise(value = "") {
    const safe = text(value);
    if (!safe) return true;
    if (APPEARANCE_ALIAS_NOISE.test(safe)) return true;
    if (/^(?:某人|一位|这位|那位|这个人|那个人)$/.test(safe)) return true;
    return false;
  }
  function dedupeAppearanceNames(values = []) {
    return unique(array(values).flatMap((value) => [text(value), ...splitAppearanceTokens(value)]).map((item) => item.replace(/^[「『【（(]+|[」』】）)]+$/g, "")).filter((item) => item.length <= 12));
  }
  function appearanceNameHints(slot) {
    const inferred = object(slot.inferred_values);
    return dedupeAppearanceNames([
      slot.base_name,
      slot.display_name,
      slot.canonical_name_hint,
      inferred.canonical_name,
      ...array(slot.aliases),
      ...array(slot.bracket_hints),
    ]).filter((value) => !isAppearanceAliasNoise(value)).slice(0, 8);
  }
  function appearanceIdentitySummary(slot) {
    const inferred = object(slot.inferred_values);
    return unique([
      ...array(slot.identity_hints),
      text(inferred.identity_role),
      text(inferred.social_position),
      text(inferred.narrative_function),
      ...array(inferred.personality_keywords),
    ]).slice(0, 8);
  }
  function clipEvidenceFragment(fragment = "", keywords = [], limit = 120) {
    const safe = text(fragment);
    if (!safe || safe.length <= limit) return safe;
    const hit = array(keywords).find((keyword) => text(keyword).length > 1 && safe.includes(keyword));
    if (!hit) return safe.slice(0, limit).trim();
    const index = safe.indexOf(hit);
    const start = Math.max(0, index - Math.floor((limit - hit.length) / 2));
    return safe.slice(start, start + limit).trim();
  }
  function buildAppearanceEvidenceFragments(slot, keywords = appearanceNameHints(slot)) {
    const novel = currentNovel();
    if (!novel || !array(keywords).some((keyword) => text(keyword).length > 1)) return [];
    const segments = novel.replace(/\r/g, "\n").split(/\n+/).flatMap((line) => line.split(/(?<=[。！？!?；;])/));
    const hits = [];
    for (const segment of segments) {
      const safe = text(segment);
      if (!safe || safe.length < 4) continue;
      if (!array(keywords).some((keyword) => text(keyword).length > 1 && safe.includes(keyword))) continue;
      hits.push(clipEvidenceFragment(safe, keywords));
      if (hits.length >= 6) break;
    }
    return unique(hits).slice(0, 6);
  }
  function buildAppearanceWorldContext() {
    return {
      content_type: text(valueOf("#genre")),
      story_era: appState.storyEraLock || appState.characterAppearanceContext?.story_era || "",
      world_setting: appState.characterAppearanceContext?.world_setting || "",
      visual_style: text(valueOf("#trailerStyle")),
    };
  }
  function buildAppearanceSiblingSummary(slot) {
    return core().slots.filter((item) => item.slot_id !== slot.slot_id).map((item) => ({
      display_name: slotDisplay(item),
      gender: text(item.gender),
      visual_age_stage: visibleStage(item),
      species: text(item.species || ""),
    })).slice(0, 12);
  }
  function buildAppearanceExcludedContext(slot) {
    const kept = appearanceNameHints(slot);
    const raw = dedupeAppearanceNames([
      slot.source_entry,
      slot.display_name,
      slot.base_name,
      slot.canonical_name_hint,
      ...array(slot.aliases),
      ...array(slot.bracket_hints),
    ]);
    return {
      removed_aliases: raw.filter((value) => isAppearanceAliasNoise(value) && !kept.includes(value)).slice(0, 12),
      omitted_raw_fields: ["source_entry", "bracket_hints", "raw_aliases", "existing_appearance", "sibling_appearance_text", "full_novel"],
      sibling_slot_count: Math.max(0, core().slots.length - 1),
    };
  }
  function appearanceCultureAnchor(slot) {
    return shouldRelaxAsianAnchor(slot) ? "" : "亚洲";
  }
  function buildAppearanceSubjectFacts(slot) {
    const inferred = object(slot.inferred_values);
    const aliases = appearanceNameHints(slot);
    const evidence_fragments = buildAppearanceEvidenceFragments(slot, aliases);
    return {
      slot_token: slot.slot_token,
      slot_id: slot.slot_id,
      display_name: slotDisplay(slot),
      canonical_name: text(slot.canonical_name_hint || inferred.canonical_name || slot.base_name || slot.display_name),
      aliases: aliases.filter((value) => value !== slotDisplay(slot)).slice(0, 6),
      gender: text(slot.gender),
      visual_age_stage: visibleStage(slot),
      chronological_age: text(slot.age?.chronological_age),
      life_stage: text(slot.age?.life_stage),
      timeline_stage: text(slot.age?.timeline_stage || "当前时间线"),
      species: text(slot.species || ""),
      culture_anchor: appearanceCultureAnchor(slot),
      identity_summary: appearanceIdentitySummary(slot),
      appearance_constraints: unique(array(inferred.appearance_constraints).map(text).filter(Boolean)).slice(0, 6),
      evidence_fragments,
    };
  }

  function inferIdentityFromAppearanceText(candidate = "") {
    const safe = text(candidate);
    const mapping = [
      ["亚洲女婴儿", { gender:"女", stage:"女婴儿" }],
      ["亚洲男婴儿", { gender:"男", stage:"男婴儿" }],
      ["亚洲小女孩", { gender:"女", stage:"小女孩" }],
      ["亚洲小男孩", { gender:"男", stage:"小男孩" }],
      ["亚洲少女", { gender:"女", stage:"少女" }],
      ["亚洲少年", { gender:"男", stage:"少年" }],
      ["亚洲年轻女性", { gender:"女", stage:"年轻女性" }],
      ["亚洲年轻男性", { gender:"男", stage:"年轻男性" }],
      ["亚洲中年女性", { gender:"女", stage:"中年女性" }],
      ["亚洲中年男性", { gender:"男", stage:"中年男性" }],
      ["亚洲老年女性", { gender:"女", stage:"老年女性" }],
      ["亚洲老年男性", { gender:"男", stage:"老年男性" }],
    ];
    for (const [prefix, payload] of mapping) {
      if (safe.startsWith(prefix)) return payload;
    }
    const first = firstAppearanceSentence(safe).slice(0, 80);
    const gender = normalizeGender(first);
    const stage = normalizeStage(first, gender);
    if (stage) return { gender: gender === "待确认" ? genderFromStage(stage) : gender, stage };
    return { gender:"", stage:"" };
  }
  function backfillSlotIdentityFromAppearance(slot, candidate = "") {
    const inferred = inferIdentityFromAppearanceText(candidate);
    const manual = sanitizeManualIdentityValues(slot);
    slot.age ||= {};
    if (!manual.gender && inferred.gender && isUnresolvedGender(slot.gender)) {
      slot.gender = inferred.gender;
      slot.current_values = { ...object(slot.current_values), gender: inferred.gender };
    }
    if (!manual.visual_age_stage && inferred.stage && isUnresolvedStage(visibleStage(slot))) {
      slot.age.visual_age_stage = inferred.stage;
      slot.age.stage_source = "appearance_writeback";
      slot.age.stage_lock_state = "ai_inferred";
      slot.current_values = { ...object(slot.current_values), visual_age_stage: inferred.stage };
    }
  }

  function appearanceDiagnosticFlags(review = {}) {
    const flags = [];
    for (const item of [...array(review.violations), ...array(review.missing)]) {
      const safe = text(item);
      if (!safe) continue;
      if (safe.includes("开头必须写")) flags.push("anchor_missing");
      else if (safe.includes("性别污染")) flags.push("gender_contamination");
      else if (safe.includes("阶段主体冲突")) flags.push("stage_identity_conflict");
      else if (safe.includes("老年特征")) flags.push("elder_contamination");
      else if (safe.includes("年轻或幼态特征")) flags.push("young_contamination");
      else if (safe.includes("同批老年角色特征")) flags.push("sibling_contamination");
      else if (safe.includes("完整穿着层次")) flags.push("clothing_layers_missing");
      else if (safe.includes("服装材质")) flags.push("material_missing");
      else if (safe.includes("服装主色")) flags.push("color_missing");
      else if (safe.includes("鞋履")) flags.push("shoes_missing");
    }
    return unique(flags);
  }
  function firstAppearanceSentence(candidate = "") {
    const safe = text(candidate);
    return safe ? safe.split(/[。！？!?]/)[0].slice(0, 80) : "";
  }

  let runtimeTimeoutCache = { value: 0, at: 0, promise: null };
  async function resolveRuntimeTimeoutSeconds() {
    try {
      if (typeof currentAiTimeoutSeconds === "function") {
        const value = Number(currentAiTimeoutSeconds());
        if (Number.isFinite(value) && value >= 30) return Math.min(400, Math.max(30, value));
      }
    } catch (_) {}
    if (runtimeTimeoutCache.value && Date.now() - runtimeTimeoutCache.at < 15000) return runtimeTimeoutCache.value;
    if (!runtimeTimeoutCache.promise) {
      runtimeTimeoutCache.promise = fetch("/api/runtime-config", {cache:"no-store", headers:{"X-VideoPromptTool-Session":currentInstance()}})
        .then((response)=>response.ok?response.json():{})
        .then((data)=>{
          const value=Number(data.ai_timeout_seconds || data.runtime_config?.ai_timeout_seconds || 400);
          runtimeTimeoutCache.value=Math.min(400,Math.max(30,Number.isFinite(value)?value:400));
          runtimeTimeoutCache.at=Date.now();
          return runtimeTimeoutCache.value;
        })
        .catch(()=>400)
        .finally(()=>{runtimeTimeoutCache.promise=null;});
    }
    return runtimeTimeoutCache.promise;
  }
  async function postCore(path, payload = {}) {
    const isAi=path==="/api/character-core/analyze";
    const timeoutSeconds=isAi?await resolveRuntimeTimeoutSeconds():90;
    const timeoutMs=isAi?timeoutSeconds*1000+15000:90000;
    const headers={"Content-Type":"application/json","X-VideoPromptTool-Session":currentInstance()};
    if (typeof requestJSON === "function") return requestJSON(path, { method:"POST", body:JSON.stringify(payload), timeoutMs, timeoutLabelSeconds:timeoutSeconds, headers });
    const controller=new AbortController();const timer=setTimeout(()=>controller.abort(),timeoutMs);
    try{
      const response=await fetch(path,{method:"POST",headers,body:JSON.stringify(payload),cache:"no-store",signal:controller.signal});
      const raw=await response.text();let data={};try{data=raw?JSON.parse(raw):{};}catch(_){data={error:raw||`HTTP ${response.status}`};}
      if(!response.ok)throw new Error(data.error||data.detail||`HTTP ${response.status}`);
      return data;
    }finally{clearTimeout(timer);}
  }
  async function projectLease(action="check", projectId=appState.projectId, force=false) {
    const id=text(projectId);if(!id)return {ok:true,acquired:true,unsaved:true};
    return postCore(`/api/character-core/project-lease?session=${encodeURIComponent(currentInstance())}`,{action,project_id:id,instance_id:currentInstance(),force:Boolean(force)});
  }
  async function ensureProjectLease({force=false,interactive=true}={}) {
    if(!text(appState.projectId))return true;
    let result=await projectLease("acquire",appState.projectId,force);
    if(result.acquired){core().project_lease={project_id:appState.projectId,instance_id:currentInstance(),acquired:true};return true;}
    if(interactive&&typeof globalThis.confirm==="function"){
      const confirmed=globalThis.confirm(`项目正在另一个窗口编辑（实例 ${result.holder_instance_id||"未知"}）。是否强制取得编辑权？`);
      if(confirmed){result=await projectLease("acquire",appState.projectId,true);if(result.acquired){core().project_lease={project_id:appState.projectId,instance_id:currentInstance(),acquired:true,forced:true};return true;}}
    }
    core().project_lease={project_id:appState.projectId,instance_id:currentInstance(),acquired:false,holder_instance_id:result.holder_instance_id||""};
    showAIStatusNotice("当前项目正在另一个窗口编辑，本窗口暂不保存以避免静默覆盖。可另存项目，或再次保存并选择强制取得编辑权。","warning",12000);
    return false;
  }
  async function saveProjectWithLease() {
    try { globalThis.__commitLatestEditableUiState?.("before_project_save_with_lease"); } catch (_) {}
    if(!baseSaveProject)return apiError("项目保存函数不可用。");
    try{
      if(!(await ensureProjectLease({interactive:true})))return null;
      const result=await baseSaveProject();
      if(appState.projectId)await projectLease("heartbeat",appState.projectId,false);
      return result;
    }catch(error){apiError(error?.message||String(error));return null;}
  }
  function releaseProjectLease() {
    const projectId=text(appState.projectId);if(!projectId)return;
    const url=`/api/character-core/project-lease?session=${encodeURIComponent(currentInstance())}`;
    const body=JSON.stringify({action:"release",project_id:projectId,instance_id:currentInstance()});
    try{if(globalThis.navigator?.sendBeacon){globalThis.navigator.sendBeacon(url,new Blob([body],{type:"application/json"}));return;}}catch(_){}
    try{fetch(url,{method:"POST",headers:{"Content-Type":"application/json","X-VideoPromptTool-Session":currentInstance()},body,keepalive:true,cache:"no-store"});}catch(_){}
  }

  function markerStage(marker="") {
    if(marker===STYLE_MARKER)return "style";
    if(marker===FACTS_MARKER)return "facts_relationships";
    if(marker===APPEARANCE_MARKER)return "appearance";
    if(marker===REVISION_CHECK_MARKER)return "revision_check";
    return "unknown";
  }
  function retryableAiError(error) {
    const message=text(error?.message||error);
    if(/AI_JSON_INVALID|V77_STAGE_PROTOCOL_ERROR|V77_PROTOCOL_INVALID/.test(message))return false;
    return /502|503|504|超时|timeout|network|fetch|连接|中转站/.test(message);
  }
  function buildCharacterAiRequestPayload(prompt, marker, meta = {}) {
    const requestStage=markerStage(marker);
    return {
      novel_text:prompt,
      generation_rules:`${marker}\ncharacter_core_version=2\ninstance_id=${currentInstance()}${marker===APPEARANCE_MARKER?"\nplain_text_card=true":""}`,
      request_stage:requestStage,
      instance_id:currentInstance(),
      slot_id:text(meta.slot_id),slot_token:text(meta.slot_token),
    };
  }
  async function legacyAiAnalyzeRaw(prompt, marker, meta = {}) {
    return postCore("/api/character-core/analyze",buildCharacterAiRequestPayload(prompt,marker,meta));
  }
  async function legacyAiAnalyze(prompt, marker, retries = 0, meta = {}) {
    let last;const maxRetries=Math.max(0,Math.min(1,Number(retries)||0));
    for(let i=0;i<=maxRetries;i++){
      try{return await legacyAiAnalyzeRaw(prompt,marker,meta);}
      catch(error){last=error;if(i>=maxRetries||!retryableAiError(error))break;await new Promise((resolve)=>setTimeout(resolve,650));}
    }
    throw last;
  }
  async function aiAnalyze(prompt, marker, retries = 0, meta = {}) {
    try { globalThis.__commitLatestEditableUiState?.(`before_character_ai:${markerStage(marker)}`); } catch (_) {}
    const service=globalThis.__V78_CHARACTER_AI_SERVICE__;
    if(service?.request){
      return service.request({
        prompt,marker,meta,
        request_stage:markerStage(marker),
        retries_requested:Math.max(0,Number(retries)||0),
        instance_id:currentInstance(),
      });
    }
    // Phase16 compatibility fallback only. Normal runtime loads the V78 service.
    return legacyAiAnalyze(prompt,marker,retries,meta);
  }

  function legacyCard(slot) {
    const age = object(slot.age);
    return {
      ...slot,
      id: slot.slot_id, name: slot.display_name, display_name: slot.display_name,
      base_name: slot.base_name, canonical_name: slot.canonical_name_hint || slot.inferred_values?.canonical_name || "",
      aliases: array(slot.aliases), identity_hint: array(slot.identity_hints).join("、"),
      gender: slot.gender === "待确认" ? "未定" : slot.gender,
      stage_label: age.visual_age_stage || "", age_stages: age.visual_age_stage || "",
      appearance: slot.appearance || "", appearance_note: slot.appearance_revision_note || "",
      person_id: slot.person_id, base_person_id: slot.person_id,
      character_core_version: 2, character_logic_version: VERSION,
      appearance_status: slot.appearance ? "ready" : "missing",
      gender_status: ["男","女","无性别"].includes(slot.gender) ? "ready" : "missing",
      stage_status: age.visual_age_stage ? "ready" : "missing",
    };
  }
  function syncLegacyState() {
    const c = core();
    appState.characters = c.slots.map(legacyCard);
    appState.forcedCharacterSlots = deepClone(appState.characters);
    appState.relationshipGraph = {
      version: "character_core_relationship_v2",
      nodes: c.people.map((person) => ({ node_id: person.person_id, person_id: person.person_id, base_name: person.canonical_name || person.primary_display_name, aliases: person.aliases, stage_cards: c.slots.filter((s) => s.person_id === person.person_id).map((s) => ({ name:s.display_name, slot_id:s.slot_id, stage:visibleStage(s) })) })),
      edges: c.relationships.map((r) => ({ ...r, source_node_id:r.source_person_id, target_node_id:r.target_person_id, relation_id:r.relation_id, relation_label:r.display_label, relation_type:r.relation_type })),
      alias_bindings: c.aliases.map((a) => ({ ...a, target_node_id:a.target_person_id, binding_id:a.binding_id })),
      mention_entities: c.mention_entities, pending: c.relationship_pending,
    };
    appState.characterRelations = appState.relationshipGraph.edges;
    appState.characterIdentityGraph = { version: 77, nodes: appState.relationshipGraph.nodes, relationships: appState.relationshipGraph.edges, alias_bindings: appState.relationshipGraph.alias_bindings };
    appState.characterStageMatrix = c.people.map((p) => ({ person_id:p.person_id, base_name:p.canonical_name || p.primary_display_name, slots:c.slots.filter((s) => s.person_id === p.person_id).map((s) => ({slot_id:s.slot_id, name:s.display_name, stage:visibleStage(s)})) }));
    appState.characterSemanticLayer = {version:BUILD_VERSION,keyword_index:deepClone(c.keyword_index),scene_casting:deepClone(c.scene_casting),analysis_state:deepClone(c.analysis_state)};
    // All compatibility fields are mirrors of CharacterCore 2.0. They are never
    // allowed to select an older人物事务协议 or older人物卡 registry again.
    appState.characterLogicVersion = VERSION;
    appState.characterAppearanceSkillVersion = VERSION;
    appState.characterIdentityGraphVersion = VERSION;
    appState.characterCoreVersion = 2;
    appState.characterRegistryEngineVersion = PROTOCOL;
    appState.analysisProtocolVersion = PROTOCOL;
  }

  function markV77RuntimeReady(source = currentNovel()) {
    const c=core(),st=analysisState(),snapshot=captureSourceSnapshot("runtime_ready");
    const sourcePresent=Boolean(snapshot.canonical_text);
    const styleReady=Boolean(st.style_ready||text(appState.characterAppearanceContext?.content_type)||text(valueOf("#genre"))||text(valueOf("#trailerStyle")));
    const slotsReady=Boolean(st.slots_ready||c.slots.length);
    const cardsReady=Boolean(st.appearances_ready||c.slots.some(slot=>text(slot.appearance))||array(appState.characters).length);
    const keywordReady=Boolean(st.keyword_index_ready||Object.values(c.keyword_index||{}).some(item=>array(item.entries).some(entry=>entry.enabled!==false)));
    st.style_ready=styleReady;
    st.slots_ready=slotsReady;st.forced_roster_ready=slotsReady;
    st.appearances_ready=cardsReady;st.character_cards_ready=cardsReady;
    st.keyword_index_ready=keywordReady;st.casting_ready=keywordReady;
    st.relationship_graph_ready=Boolean(st.relationships_ready||c.relationships.length||c.aliases.length);
    st.source_revision=snapshot.source_revision;
    // Hotfix12: revisions and source_stale are diagnostics only. They never
    // block whole-outline generation when current usable material still exists.
    st.ready=Boolean(sourcePresent);
    syncLegacyState();
    appState.analysisComplete=sourcePresent;
    appState.analysisSourceText=snapshot.canonical_text;
    appState.characterRegistryStatus=sourcePresent?"ready":"stale";
    appState.characterAnalysisResultStatus=sourcePresent?"current":text(st.last_message||"stale");
    appState.characterRegistrySourceHash=snapshot.source_hash;appState.characterGuideSourceHash=snapshot.source_hash;
    return st.ready;
  }

  function getOutlineUsableDataState(snapshot = captureSourceSnapshot("outline_usable_state")) {
    const c=core(),st=analysisState();
    const genre=text(valueOf("#genre"));
    const trailerStyle=text(valueOf("#trailerStyle"));
    const guideText=String(formalRosterSnapshot("outline_usable_state",true).raw_text||"").trim();
    const slotRows=array(c.slots).filter(slot=>!slot.disabled);
    const legacyCards=array(appState.characters).filter(Boolean);
    const keywordCount=Object.values(c.keyword_index||{}).reduce((count,item)=>count+array(item?.entries).filter(entry=>entry?.enabled!==false).length,0);
    const relationshipCount=array(c.relationships).filter(item=>!item.disabled).length;
    const aliasCount=array(c.aliases).filter(item=>!item.disabled).length;
    const manualCastCount=array(appState.scenes).filter(scene=>scene?.characters_mode==="manual"&&array(scene.character_slot_ids).length).length;
    const styleAvailable=Boolean(genre||trailerStyle||text(appState.characterAppearanceContext?.content_type)||text(appState.characterAppearanceContext?.trailer_style));
    const rosterAvailable=Boolean(guideText||slotRows.length);
    const cardsAvailable=Boolean(slotRows.length||legacyCards.length);
    const appearanceCount=slotRows.filter(slot=>text(slot.appearance)).length||legacyCards.filter(card=>text(card.appearance)).length;
    const relationshipsAvailable=Boolean(relationshipCount||aliasCount);
    const keywordsAvailable=Boolean(keywordCount);
    const manualCastAvailable=Boolean(manualCastCount);
    const sourceReplaced=Boolean(c.source_stale||(c.source_hash&&c.source_hash!==snapshot.source_hash)||(st.source_hash&&st.source_hash!==snapshot.source_hash));
    const allMaterialsGone=Boolean(!styleAvailable&&!rosterAvailable&&!cardsAvailable&&!relationshipsAvailable&&!keywordsAvailable&&!manualCastAvailable);
    const sourceReplacedAndDataCleared=Boolean(sourceReplaced&&allMaterialsGone);
    const warnings=[];
    if(!styleAvailable)warnings.push("当前没有统一风格，将按当前原文直接生成");
    if(!cardsAvailable)warnings.push("当前没有正式人物卡，将允许AI按原文生成纯场景或临时人物");
    else if(!appearanceCount)warnings.push("当前人物卡没有完整外形，将提交现有名称、性别和阶段继续生成");
    if(!relationshipsAvailable)warnings.push("人物关系图不完整，将使用现有人物卡与本地名称匹配继续生成");
    if(!keywordsAvailable)warnings.push("本地关键词索引为空，将在点击时自动重建；仍为空也不阻断");
    return {
      source_present:Boolean(snapshot.canonical_text),
      source_hash:text(snapshot.source_hash),
      source_revision:Number(snapshot.source_revision||0),
      source_replaced:sourceReplaced,
      source_stale:Boolean(c.source_stale),
      style_available:styleAvailable,
      roster_available:rosterAvailable,
      cards_available:cardsAvailable,
      appearance_count:appearanceCount,
      slot_count:slotRows.length,
      legacy_card_count:legacyCards.length,
      relationships_available:relationshipsAvailable,
      relationship_count:relationshipCount,
      alias_count:aliasCount,
      keywords_available:keywordsAvailable,
      keyword_count:keywordCount,
      manual_cast_available:manualCastAvailable,
      manual_cast_count:manualCastCount,
      all_materials_gone:allMaterialsGone,
      source_replaced_and_data_cleared:sourceReplacedAndDataCleared,
      can_generate:Boolean(snapshot.canonical_text),
      warnings,
      diagnostics:revisionDiagnostics(snapshot),
    };
  }

  function v77OutlineReadiness() {
    const snapshot=captureSourceSnapshot("outline_readiness");
    const usable=getOutlineUsableDataState(snapshot);
    // Hotfix24: the only business prerequisite for outline/regeneration is source text.
    // Missing style/cards/relations/revisions are degradation paths, never step gates.
    if(!usable.source_present)return {ok:false,message:"无法生成：当前整段原文为空。",hard_block:"source_empty",usable};
    return {ok:true,source:snapshot.canonical_text,source_hash:snapshot.source_hash,source_revision:snapshot.source_revision,slot_count:usable.slot_count,analysis_state:deepClone(analysisState()),usable,warnings:usable.warnings};
  }

  async function parseSlots({ preserve = true, rosterSnapshot = null } = {}) {
    const roster=rosterSnapshot||formalRosterSnapshot("parse_slots",true);
    const guide=String(roster.raw_text||"");
    const novel=currentNovel();
    appState.character_guide_input=guide;
    if(!guide.trim()){
      const c=core();c.people=[];c.slots=[];c.relationships=[];c.aliases=[];c.mention_entities=[];c.scene_casting={};c.keyword_index={};c.relationship_pending=[];c.formal_roster_stale=false;c.character_revision+=1;
      setAnalysisStage("slots",false,"强制人物名单为空");syncLegacyState();
      return {slot_count:0,slots:[],people:[],formal_roster_snapshot:roster};
    }
    const result = await postCore("/api/character-core/parse-slots", { guide_text:guide, novel_text:novel, source_hash:fingerprint(novel) });
    assertFormalRosterFresh(roster,"名单解析");
    const previousCore=core();const nextSourceHash=fingerprint(novel);const sameSource=previousCore.source_hash===nextSourceHash;
    const previous = new Map(previousCore.slots.map((s) => [s.source_entry, s]));
    const nextSlots = array(result.slots).map((slot) => {
      const old = preserve && sameSource ? previous.get(slot.source_entry) : null;
      if (!old) return slot;
      const manual = object(old.manual_values);
      if (isUnresolvedGender(manual.gender)) delete manual.gender;
      if (isUnresolvedStage(manual.visual_age_stage)) delete manual.visual_age_stage;
      const oldAge = object(old.age);
      const parsedAge = object(slot.age);
      const currentStage = text(manual.visual_age_stage || (!isUnresolvedStage(oldAge.visual_age_stage) ? oldAge.visual_age_stage : "") || parsedAge.visual_age_stage);
      const currentGender = text(manual.gender || (!isUnresolvedGender(old.gender) ? old.gender : "") || slot.gender || "待确认");
      return {
        ...slot,
        slot_id: old.slot_id || slot.slot_id,
        person_id: old.person_id || slot.person_id,
        display_name: text(manual.display_name || old.display_name || slot.display_name),
        gender: currentGender,
        age: {
          ...parsedAge,
          chronological_age: text(manual.chronological_age || oldAge.chronological_age || parsedAge.chronological_age),
          visual_age_stage: currentStage,
          life_stage: text(manual.life_stage || oldAge.life_stage || parsedAge.life_stage),
          timeline_stage: text(manual.timeline_stage || oldAge.timeline_stage || parsedAge.timeline_stage || "当前时间线"),
          stage_locked: Boolean(currentStage && (oldAge.stage_locked || oldAge.stage_lock_state === "manual_locked" || manual.visual_age_stage)),
          stage_source: manual.visual_age_stage ? "user_manual" : text(oldAge.stage_source || parsedAge.stage_source),
          stage_lock_state: manual.visual_age_stage ? "manual_locked" : text(oldAge.stage_lock_state || parsedAge.stage_lock_state || "ai_locked"),
        },
        species: text(manual.species || old.species || slot.species),
        appearance: old.appearance || "", appearance_revision_note: text(manual.appearance_revision_note || old.appearance_revision_note || ""),
        appearance_detail_level: old.appearance_detail_level || "详细", appearance_status: old.appearance_status || (old.appearance ? "ready" : "missing"),
        appearance_revision: Number(old.appearance_revision || 0), appearance_stage_stale: Boolean(old.appearance_stage_stale),
        manual_values: { ...object(slot.manual_values), ...manual },
        current_values: { ...object(slot.current_values), ...object(old.current_values), ...manual, gender:currentGender, visual_age_stage:currentStage },
        generated_values: object(old.generated_values), inferred_values: object(old.inferred_values),
        character_revision: Number(old.character_revision || 0),
      };
    });
    const c = core(); c.source_hash = nextSourceHash; c.analysis_state.source_hash=nextSourceHash; c.source_stale=false; c.formal_roster_stale=false; c.slots = nextSlots; c.people = array(result.people); if(!sameSource){c.relationships=[];c.aliases=[];c.mention_entities=[];c.scene_casting={};c.relationship_pending=[];c.temporary_continuity_cache={};c.temporary_continuity_cache_character_core_validator={};c.temporary_continuity_cache_v78_primary={};c.temporary_continuity_authority_v78={};c.relationship_conflicts=[];}
    else {const validPeople=new Set(c.people.map(p=>text(p.person_id)).filter(Boolean));c.relationships=array(c.relationships).filter(r=>validPeople.has(text(r.source_person_id))&&validPeople.has(text(r.target_person_id)));c.aliases=array(c.aliases).filter(a=>validPeople.has(text(a.target_person_id)));c.scene_casting={};c.relationship_pending=array(c.relationship_pending).filter(r=>!r.source_person_id||validPeople.has(text(r.source_person_id)));}
    c.character_revision += 1;
    setAnalysisStage("slots", Boolean(nextSlots.length), nextSlots.length?`已解析 ${nextSlots.length} 个人物槽位`:"强制人物名单为空");
    syncLegacyState();
    return {...result,slot_count:nextSlots.length,formal_roster_snapshot:roster};
  }

  function recordsArray(data = {}) {
    function walk(value, depth = 0) {
      if (depth > 5) return [];
      if (Array.isArray(value)) return value.map((x, i) => ({ ...(typeof x === "string" ? { appearance:x } : object(x)), __index:i+1 }));
      const source = object(value);
      for (const key of ["characters","character_cards","people","cards"]) {
        if (Array.isArray(source[key])) return walk(source[key], depth+1);
        if (source[key] && typeof source[key] === "object") return Object.entries(source[key]).map(([token, x], i) => ({ ...(typeof x === "string" ? {appearance:x} : object(x)), slot_token:token, __index:i+1 }));
      }
      if (source.character_facts && typeof source.character_facts === "object") return Object.entries(source.character_facts).map(([token,x],i)=>({...object(x),slot_token:token,__index:i+1}));
      for (const key of ["character","character_card","card","result","data","analysis","output","response"]) {
        const nested = walk(source[key], depth+1); if (nested.length) return nested;
      }
      if (source.slot_token || source.slot_id || source.gender || source.visual_age_stage || source.story_age_stage || source.appearance) return [{...source,__index:1}];
      return [];
    }
    return walk(data);
  }
  function characterFactsCompleteness(slot = {}) {
    const gender = normalizeGender(slot.gender);
    const visualStage = normalizeStage(visibleStage(slot), slot.gender);
    const timelineStage = text(slot.age?.timeline_stage || slot.current_values?.timeline_stage || "当前时间线");
    const lifeStage = text(slot.age?.life_stage || slot.current_values?.life_stage || "");
    const criticalMissing = [];
    const advisoryMissing = [];
    const stageSource=text(slot.age?.stage_source||"");
    if (gender === "待确认") criticalMissing.push("gender");
    if (!visualStage || stageSource === "appearance_writeback") criticalMissing.push("visual_age_stage");
    if (!timelineStage) advisoryMissing.push("timeline_stage");
    if (!lifeStage) advisoryMissing.push("life_stage");
    return {
      slot_id:text(slot.slot_id), slot_token:text(slot.slot_token), display_name:slotDisplay(slot),
      gender, visual_age_stage:visualStage, stage_source:stageSource, timeline_stage:timelineStage, life_stage:lifeStage,
      critical_missing:criticalMissing, advisory_missing:advisoryMissing,
      complete:criticalMissing.length===0,
    };
  }
  function characterFactsIntegrity(c = core()) {
    const rows=c.slots.filter(slot=>!slot.disabled).map(characterFactsCompleteness);
    const incomplete=rows.filter(row=>!row.complete);
    const advisory=rows.filter(row=>row.advisory_missing.length);
    return {
      total:rows.length, complete_count:rows.length-incomplete.length, incomplete_count:incomplete.length,
      rows, incomplete, advisory,
      ready:rows.length>0 && incomplete.length===0,
      summary:rows.length?`人物事实完整 ${rows.length-incomplete.length}/${rows.length}${incomplete.length?`；待补全 ${incomplete.map(row=>`${row.display_name}(${row.critical_missing.join("+")})`).join("、")}`:""}`:"当前没有正式人物槽位",
    };
  }
  function applyFactRecordToSlot(record = {}, slot = {}, { missingOnly = false } = {}) {
    if (!slot) return false;
    slot.age ||= {};
    const beforeStage=normalizeStage(visibleStage(slot),slot.gender);
    const manual=sanitizeManualIdentityValues(slot), declared=object(slot.declared_values);
    // AI 返回键名不稳定：兼容 current_gender/gender_hint/sex 与 story_age_stage/stage_label/age_stage 等变体。
    const recordGender=text(record.gender||record.current_gender||record.gender_hint||record.sex||genderFromStage(record.visual_age_stage||record.story_age_stage||record.stage_label||record.age_stage||record.age_stage_hint||""));
    const returnedStage=normalizeStage(record.visual_age_stage || record.story_age_stage || record.stage_label || record.age_stage || record.age_stage_hint, recordGender || slot.gender);
    const returnedGender=normalizeGender(recordGender || genderFromStage(returnedStage));
    const currentStageSource=text(slot.age?.stage_source||"");
    const legacyAppearanceFact=currentStageSource==="appearance_writeback";
    const canGender=!manual.gender&&!declared.gender&&returnedGender!=="待确认"&&(!missingOnly||isUnresolvedGender(slot.gender)||legacyAppearanceFact);
    const canStage=!manual.visual_age_stage&&!declared.visual_age_stage&&returnedStage&&(!missingOnly||isUnresolvedStage(visibleStage(slot))||legacyAppearanceFact);
    if(canGender) slot.gender=returnedGender;
    if(canStage){slot.age.visual_age_stage=returnedStage;slot.age.stage_source=missingOnly?"ai_fact_repair":"ai_inferred";slot.age.stage_lock_state="ai_inferred";}
    if(!manual.chronological_age&&record.chronological_age&&(!missingOnly||!text(slot.age.chronological_age)))slot.age.chronological_age=text(record.chronological_age);
    if(!manual.life_stage&&record.life_stage&&(!missingOnly||!text(slot.age.life_stage)))slot.age.life_stage=text(record.life_stage);
    if(!manual.timeline_stage&&record.timeline_stage&&(!missingOnly||!text(slot.age.timeline_stage)))slot.age.timeline_stage=text(record.timeline_stage);
    if(!text(slot.age.timeline_stage))slot.age.timeline_stage="当前时间线";
    if(!manual.species&&record.species&&(!missingOnly||!text(slot.species)))slot.species=text(record.species);
    const stageGender=genderFromStage(slot.age.visual_age_stage);
    if(!manual.gender&&stageGender!=="待确认"&&(!missingOnly||isUnresolvedGender(slot.gender)))slot.gender=stageGender;
    slot.current_values={...object(slot.current_values),gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:text(slot.age?.chronological_age),life_stage:text(slot.age?.life_stage),timeline_stage:text(slot.age?.timeline_stage||"当前时间线"),species:text(slot.species)};
    slot.inferred_values={...object(slot.inferred_values),canonical_name:text(record.canonical_name)||text(slot.inferred_values?.canonical_name),ai_aliases:array(record.aliases).length?array(record.aliases):array(slot.inferred_values?.ai_aliases),gender:returnedGender!=="待确认"?returnedGender:text(slot.inferred_values?.gender),visual_age_stage:returnedStage||text(slot.inferred_values?.visual_age_stage),chronological_age:text(record.chronological_age)||text(slot.inferred_values?.chronological_age),life_stage:text(record.life_stage)||text(slot.inferred_values?.life_stage),timeline_stage:text(record.timeline_stage)||text(slot.inferred_values?.timeline_stage),species:text(record.species)||text(slot.inferred_values?.species),identity_role:text(record.identity_role)||text(slot.inferred_values?.identity_role),social_position:text(record.social_position)||text(slot.inferred_values?.social_position),personality_keywords:array(record.personality_keywords).length?array(record.personality_keywords):array(slot.inferred_values?.personality_keywords),narrative_function:text(record.narrative_function)||text(slot.inferred_values?.narrative_function),appearance_constraints:array(record.appearance_constraints).length?array(record.appearance_constraints):array(slot.inferred_values?.appearance_constraints),is_current_timeline:Boolean(record.is_current_timeline ?? slot.inferred_values?.is_current_timeline)};
    const afterStage=normalizeStage(visibleStage(slot),slot.gender);
    if(beforeStage!==afterStage&&text(slot.appearance))slot.appearance_stage_stale=true;
    if(!missingOnly){slot.aliases=unique([...array(slot.aliases),text(record.canonical_name),...array(record.aliases)]);const person=core().people.find(p=>p.person_id===slot.person_id);if(person){if(record.canonical_name)person.canonical_name=text(record.canonical_name);person.aliases=unique([...array(person.aliases),...slot.aliases]);person.identity_hints=unique([...array(person.identity_hints),...slot.identity_hints]);person.species=slot.species||person.species;person.is_protagonist=Boolean(record.is_protagonist??person.is_protagonist);person.is_narrator=Boolean(record.is_narrator??person.is_narrator);}}
    slot.character_revision+=1;
    return true;
  }
  function buildFactsRepairPrompt(integrity = characterFactsIntegrity()) {
    const missing=integrity.incomplete.map(row=>{
      const slot=core().slots.find(item=>item.slot_id===row.slot_id);
      return {
        slot_token:slot?.slot_token,slot_id:slot?.slot_id,return_index:slot?.return_index,display_name:slotDisplay(slot||{}),
        source_entry:text(slot?.source_entry),manual_values:object(slot?.manual_values),declared_values:object(slot?.declared_values),
        current_values:{gender:text(slot?.gender),visual_age_stage:visibleStage(slot||{}),chronological_age:text(slot?.age?.chronological_age),life_stage:text(slot?.age?.life_stage),timeline_stage:text(slot?.age?.timeline_stage||"当前时间线"),species:text(slot?.species)},
        missing_fields:row.critical_missing,
      };
    });
    return [
      "【任务】CharacterCore 2.0 人物事实缺失字段补全。只补下面明确缺失的字段，不重做人物关系、不生成外形、不覆盖任何已有或人工值。",
      "【最高规则】必须逐slot_token原样返回；missing_fields里的字段必须给出明确结果。visual_age_stage必须从固定剧情年龄阶段中选择，不能返回待确认、未知、空字符串。gender必须根据全文判断，非人类允许无性别；不得用本地称谓规则猜测。",
      "【待补全槽位】",JSON.stringify(missing),
      "【完整原文开始】",currentNovel(),"【完整原文结束】",
      "【返回】只返回JSON：{characters:[{slot_token,slot_id,return_index,gender,visual_age_stage,chronological_age,life_stage,timeline_stage,species}]}。每个待补槽位必须同时返回gender与visual_age_stage以满足人物事实协议；程序只写回当前缺失字段，不覆盖已有正确值。不返回relationships、alias_bindings、appearance或解释。",
    ].join("\\n\\n");
  }
  async function repairMissingCharacterFacts(integrity = characterFactsIntegrity()) {
    if(!integrity.incomplete.length)return {attempted:false,integrity};
    const before=core(),requestState={source_hash:before.source_hash,character_revision:before.character_revision,instance_id:currentInstance()};
    const data=await aiAnalyze(buildFactsRepairPrompt(integrity),FACTS_MARKER,0,{task_type:"facts_missing_fields_repair"});
    const c=core();
    if(c.source_hash!==requestState.source_hash||c.character_revision!==requestState.character_revision||currentInstance()!==requestState.instance_id)return {attempted:true,stale:true,discarded:true,integrity:characterFactsIntegrity()};
    let mapped=0;
    for(const record of recordsArray(data)){
      const slot=recordSlot(record,c.slots);if(!slot)continue;
      const target=integrity.incomplete.find(row=>row.slot_id===slot.slot_id);if(!target)continue;
      if(applyFactRecordToSlot(record,slot,{missingOnly:true}))mapped++;
    }
    if(mapped)c.character_revision+=1;
    const after=characterFactsIntegrity(c);
    console.info("[CHARACTER_FACTS_REPAIR_TRACE]",{requested:integrity.incomplete.length,mapped,remaining:after.incomplete_count,remaining_rows:after.incomplete});
    emitCharacterCoreTrace({trace_stage:"facts_missing_fields_repair",request_id:text(data?._ai_meta?.request_id)||`facts-repair-${Date.now()}`,requested:integrity.incomplete.length,mapped,remaining:after.incomplete_count,outcome:after.ready?"complete":"partial",final_snapshot:after.rows}).catch(()=>{});
    return {attempted:true,data,mapped,integrity:after};
  }

  function recordSlot(record, slots) {
    const token = text(record.slot_token).toUpperCase();
    let slot = token ? slots.find((s) => text(s.slot_token).toUpperCase() === token) : null;
    if (slot) return slot;
    const slotId = text(record.slot_id);
    slot = slotId ? slots.find((s) => text(s.slot_id) === slotId) : null;
    if (slot) return slot;
    const personId = text(record.person_id);
    if (personId) {
      const personMatches = slots.filter((s) => text(s.person_id) === personId);
      if (personMatches.length === 1) return personMatches[0];
    }
    const names = unique([record.canonical_name, record.name, record.display_name, record.base_name, ...array(record.aliases)].map(text).filter(Boolean));
    if (names.length) {
      const nameMatches = slots.filter((s) => {
        const candidates = unique([s.display_name, s.base_name, s.canonical_name_hint, s.source_entry, ...array(s.aliases), ...array(s.bracket_hints)].map(text).filter(Boolean));
        return names.some((name) => {
          const cleanName = cleanIdentityToken(name);
          return cleanName && candidates.some((candidate) => cleanIdentityToken(candidate) === cleanName);
        });
      });
      if (nameMatches.length === 1) return nameMatches[0];
    }
    const idx = Number(record.return_index || record.__index);
    if (idx >= 1 && idx <= slots.length) return slots[idx-1];
    return slots.length === 1 ? slots[0] : null;
  }

  function buildFactsPrompt() {
    try { globalThis.__commitLatestEditableUiState?.("before_character_facts_prompt"); } catch (_) {}
    const c = core();
    const rows = c.slots.map((s) => ({
      slot_token:s.slot_token, slot_id:s.slot_id, return_index:s.return_index, person_id:s.person_id,
      source_entry:s.source_entry, base_name:s.base_name, display_name:s.display_name,
      canonical_name_hint:s.canonical_name_hint, aliases:s.aliases, identity_hints:s.identity_hints, bracket_hints:s.bracket_hints,
      declared_values:s.declared_values, manual_values:s.manual_values,
      current_gender:s.gender, chronological_age:s.age?.chronological_age || "", visual_age_stage:s.age?.visual_age_stage || "",
      life_stage:s.age?.life_stage || "", timeline_stage:s.age?.timeline_stage || "", species:s.species || "",
    }));
    return [
      "【任务】CharacterCore 2.0人物事实与AI人物关系图分析。人物卡只来自slot_rows，处理所有小说题材、种族、世界观和时间线。",
      (globalThis.__v23GetInstruction?.("global") ? `【Hotfix23全局AI语义指令｜协议锁定】\n${globalThis.__v23GetInstruction("global")}` : ""),
      (globalThis.__v23GetInstruction?.("relationship") ? `【Hotfix23当前AI人物关系自定义指令｜只改语义不改协议】\n${globalThis.__v23GetInstruction("relationship")}` : ""),
      "【通用语义原则】不得使用固定称谓词表或固定人物关系模板。必须根据本次完整原文、叙事视角、对白方向、行为连续性、时间线和强制名单，动态判断人物事实、关系、别称与关系称谓指向。",
      "【证据优先级】用户manual_values > declared_values/强制名单明确事实 > 原文明示 > AI全文语义判断。姓名只能作为弱证据，不得单独决定性别、年龄、关系或同一人物归并。",
      "【年龄模型】必须区分chronological_age、visual_age_stage、life_stage、timeline_stage；用户提供的visual_age_stage不可被改写。gender可为男、女、无性别或待确认；非人类、系统和人工智能不得被强行套成人类二元性别。",
      "【slot_rows】", JSON.stringify(rows),
      "【当前人物关系图｜当前有效值】", JSON.stringify(array(c.relationships).filter((item)=>!item.disabled)),
      "【当前称呼与别称映射｜当前有效值】", JSON.stringify(array(c.aliases).filter((item)=>!item.disabled)),
      "【当前关系值原则】当前关系图和别称映射来自用户此刻界面，是本次重新分析的输入基础；不得读取更早的AI返回或历史缓存。",
      "【当前内容类型/时代/世界观】", JSON.stringify({ content_type:text(valueOf("#genre")), story_era:appState.storyEraLock || appState.characterAppearanceContext?.story_era || "", world_setting:appState.characterAppearanceContext?.world_setting || "" }),
      "【完整原文开始】", currentNovel(), "【完整原文结束】",
      "【本地关键词契约】本地程序只做确定性关键词匹配，不自行理解关系。凡是应触发某人物出镜勾选的正式名、别称、昵称、代号、身份称呼或关系称谓，都必须放入alias_bindings并明确target_person_id；存在多阶段时尽量返回target_slot_id或timeline_scope；局部有效称呼必须返回valid_from_line与valid_to_line；低置信或歧义表达放入ambiguous_relations，不得强行绑定。",
      "【名单外通用临时人物】完整扫描原文中不在slot_rows但可能进入画面的所有人物/群体，不限制现代、古代、民国、校园、职场、医疗、刑侦、仙侠、玄幻、科幻、末世、宫廷、战争、异族或系统题材，也不依赖固定称谓词库。具名人物、关系称谓、职业/身份人物、随从、仆役、军士、修士、灵体、机械人、异族、妖兽化形、群体成员等都放入mention_entities；不得只返回正式人物。每项字段：label,aliases,entity_type,count,participation_state,valid_from_line,valid_to_line,gender_hint,age_stage_hint,species_hint,identity_hint,evidence,confidence。participation_state区分visible_candidate、visual_exposition、group_visible、offscreen_voice、mentioned_only、photo_visible、screen_visible、memory_visible、dream_visible。介绍家族/组织/队伍/班级/宗门/军阵等人物构成的句子，只要适合直接视觉化，标visual_exposition而不是mentioned_only。",
      "【返回】严格按槽位返回characters，并返回relationships、alias_bindings、mention_entities、ambiguous_relations。characters每项必须含：slot_token(原样返回)、display_name、canonical_name、aliases、gender(男/女/无性别/待确认)、visual_age_stage(固定剧情年龄阶段)、chronological_age、life_stage、timeline_stage、species、identity_role、social_position。alias_bindings每项字段：alias,target_person_id,target_slot_id,alias_type,valid_from_line,valid_to_line,timeline_scope,confidence,evidence,keyword_enabled。不得生成appearance。",
    ].join("\n\n");
  }

  async function analyzeFactsAndRelations() {
    const before=core();const requestState={source_hash:before.source_hash,character_revision:before.character_revision,instance_id:currentInstance()};
    const data = await aiAnalyze(buildFactsPrompt(), FACTS_MARKER, 1);
    const c = core();
    if(c.source_hash!==requestState.source_hash||c.character_revision!==requestState.character_revision||currentInstance()!==requestState.instance_id){
      emitCharacterCoreTrace({trace_stage:"core_parity_facts_relationships",request_id:text(data?._ai_meta?.request_id)||`core-parity-${Date.now()}`,outcome:"stale_discarded",summary:"响应到达时人物/实例状态已变化，本次返回整包丢弃",request_state:requestState,current:{source_hash:c.source_hash,character_revision:c.character_revision,instance_id:currentInstance()}}).catch(()=>{});
      return {stale:true,discarded:true};
    }
    const records=recordsArray(data);
    let mappedFacts=0;
    const unmappedFacts=[];
    for (const record of records) {
      const slot = recordSlot(record, c.slots);
      if (!slot) { unmappedFacts.push({slot_token:text(record.slot_token),slot_id:text(record.slot_id),return_index:Number(record.return_index||record.__index||0),canonical_name:text(record.canonical_name||record.name)}); continue; }
      mappedFacts += 1;
      applyFactRecordToSlot(record,slot,{missingOnly:false});
    }
    const previousState=analysisState();
    const previous = { relationships:deepClone(c.relationships), aliases:deepClone(c.aliases), mention_entities:deepClone(c.mention_entities), relationship_pending:deepClone(c.relationship_pending), relationships_ready:Boolean(previousState.relationships_ready), relationship_fallback_available:Boolean(previousState.relationship_fallback_available) };
    const sourceData=object(data),graphData=object(sourceData.relationship_graph||sourceData.relationshipGraph);
    const hasRelationshipEnvelope=["relationships","relations","alias_bindings","keyword_bindings","aliases"].some((key)=>Object.prototype.hasOwnProperty.call(sourceData,key)) || ["relationships","relations","edges","alias_bindings","keyword_bindings","aliases"].some((key)=>Object.prototype.hasOwnProperty.call(graphData,key));
    // Hotfix31: a facts-only/partially recovered response must never erase a
    // previously usable relationship graph. Replace relationship data only when
    // the AI response actually contains the relationship/alias envelope.
    if (hasRelationshipEnvelope) {
      c.relationships = normalizeRelationships(data, c, previous);
      c.aliases = normalizeAliases(data, c, previous);
    } else {
      c.relationships = previous.relationships;
      c.aliases = previous.aliases;
    }
    const hasMentionEntities=Object.prototype.hasOwnProperty.call(sourceData,"mention_entities")||Object.prototype.hasOwnProperty.call(graphData,"mention_entities");
    const hasPendingRelations=["ambiguous_relations","pending"].some((key)=>Object.prototype.hasOwnProperty.call(sourceData,key)||Object.prototype.hasOwnProperty.call(graphData,key));
    c.mention_entities = hasMentionEntities ? (array(sourceData.mention_entities).length ? array(sourceData.mention_entities) : array(graphData.mention_entities)) : previous.mention_entities;
    c.relationship_pending = hasPendingRelations ? (array(sourceData.ambiguous_relations || sourceData.pending).length ? array(sourceData.ambiguous_relations || sourceData.pending) : array(graphData.ambiguous_relations || graphData.pending)) : previous.relationship_pending;
    c.character_revision += 1;
    const activeSlots=c.slots.filter((slot)=>!slot.disabled);
    let factsIntegrity=characterFactsIntegrity(c);
    let factsRepair=null;
    if(factsIntegrity.incomplete_count){
      try{factsRepair=await repairMissingCharacterFacts(factsIntegrity);factsIntegrity=factsRepair.integrity||characterFactsIntegrity(c);}
      catch(repairError){factsRepair={attempted:true,error:text(repairError?.message||repairError),integrity:characterFactsIntegrity(c)};factsIntegrity=factsRepair.integrity;}
    }
    const factsReady=Boolean(factsIntegrity.ready);
    const relationshipsReady=hasRelationshipEnvelope ? true : Boolean(previous.relationships_ready && (previous.relationships.length || previous.aliases.length));
    const st=analysisState();st.facts_fallback_available=!factsReady;st.relationship_fallback_available=hasRelationshipEnvelope ? false : Boolean(previous.relationship_fallback_available);
    st.character_facts_integrity=deepClone(factsIntegrity);
    setAnalysisStage("facts", factsReady, `${factsIntegrity.summary}${unmappedFacts.length?`；未映射记录 ${unmappedFacts.length}`:""}${factsRepair?.attempted?`；缺失字段补全${factsRepair?.error?`失败：${factsRepair.error}`:factsReady?"完成":"仍有缺失"}`:""}`);
    setAnalysisStage("relationships", relationshipsReady, hasRelationshipEnvelope?`关系 ${c.relationships.length} 条，别称 ${c.aliases.length} 条`:(relationshipsReady?`本次未返回关系字段；已保留当前有效关系 ${c.relationships.length} 条、别称 ${c.aliases.length} 条`:`AI人物关系字段未完整返回；当前保留已有关系/别称，不伪装为已完成`));
    rebuildKeywordIndex();
    syncLegacyState(); renderAll();
    const trace={trace_stage:"core_parity_facts_relationships",request_id:text(data?._ai_meta?.request_id)||`core-parity-${Date.now()}`,facts_returned:records.length,facts_mapped:mappedFacts,slot_count:activeSlots.length,facts_integrity:factsIntegrity,unmapped_facts:unmappedFacts,relationships_returned:relationshipRows(data).length,relationships_mapped:c.relationships.filter((r)=>!r.disabled).length,aliases_returned:aliasRows(data).length,aliases_mapped:c.aliases.filter((a)=>!a.disabled).length,relationships_envelope_present:hasRelationshipEnvelope,outcome:(factsReady&&relationshipsReady)?"complete":"partial",final_snapshot:{slots:activeSlots.map(slot=>({slot_id:slot.slot_id,person_id:slot.person_id,display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:text(slot.age?.chronological_age),species:text(slot.species)})),relationships:c.relationships.filter(r=>!r.disabled).slice(0,80),aliases:c.aliases.filter(a=>!a.disabled).slice(0,120),mention_entities:array(c.mention_entities).slice(0,80)}};
    console.info("[CORE_PARITY_TRACE]",trace);emitCharacterCoreTrace(trace).catch(()=>{});
    return data;
  }

  function cleanIdentityToken(value = "") {
    return text(value).replace(/〔[^〕]*〕/g, "").replace(/（[^）]*）/g, "").replace(/\([^)]*\)/g, "").replace(/\s+/g, "").trim();
  }
  function resolvePerson(value, c = core()) {
    const v = text(value);
    if (!v) return "";
    if (c.people.some((p) => p.person_id === v)) return v;
    const byToken = c.slots.find((s) => text(s.slot_token).toUpperCase() === v.toUpperCase() || s.slot_id === v || s.person_id === v);
    if (byToken) return byToken.person_id;
    const clean = cleanIdentityToken(v);
    const personMatches = c.people.filter((p) => unique([p.canonical_name,p.primary_display_name,...array(p.aliases)]).some((name)=>cleanIdentityToken(name)===clean));
    if (personMatches.length === 1) return personMatches[0].person_id;
    const slotMatches = c.slots.filter((s)=>unique([s.display_name,s.base_name,s.canonical_name_hint,s.source_entry,...array(s.aliases),...array(s.bracket_hints)]).some((name)=>cleanIdentityToken(name)===clean));
    const personIds=unique(slotMatches.map((s)=>s.person_id).filter(Boolean));
    return personIds.length===1 ? personIds[0] : "";
  }
  function relationshipRows(data = {}) {
    const source=object(data),graph=object(source.relationship_graph||source.relationshipGraph);
    return [...array(source.relationships),...array(source.relations),...array(graph.relationships),...array(graph.relations),...array(graph.edges)];
  }
  function aliasRows(data = {}) {
    const source=object(data),graph=object(source.relationship_graph||source.relationshipGraph);
    return [...array(source.alias_bindings),...array(source.keyword_bindings),...array(source.aliases),...array(graph.alias_bindings),...array(graph.keyword_bindings),...array(graph.aliases)];
  }
  function normalizeRelationships(data, c, previous) {
    const auto = relationshipRows(data).map((r) => {
      let source = resolvePerson(r.source_person_id || r.subject_person_id || r.subject_slot || r.source_slot || r.subject || r.source || r.from || r.person_id_a || r.personA || r.character_a || r.actor_id || r.person_1, c);
      let target = resolvePerson(r.target_person_id || r.object_person_id || r.object_slot || r.target_slot || r.object || r.target || r.to || r.person_id_b || r.personB || r.character_b || r.receiver_id || r.person_2, c);
      // AI 输出键名不稳定兜底：标准键缺失时，扫描本对象中能解析到正式人物的值作为关系两端。
      if (!source || !target) {
        const resolvedIds = [];
        for (const [key, value] of Object.entries(r)) {
          if (typeof value !== "string" || resolvedIds.length >= 2) continue;
          const resolved = resolvePerson(value, c);
          if (resolved && !resolvedIds.includes(resolved)) resolvedIds.push(resolved);
        }
        if (!source) source = resolvedIds[0] || "";
        if (!target) target = resolvedIds.find((id) => id !== source) || "";
      }
      if (!source || !target || source === target) return null;
      const label = text(r.display_label || r.relation_label || r.label || r.relationship || r.relation_type || "相关");
      return { relation_id:text(r.relation_id || r.id) || `rel_${fingerprint(`${source}|${label}|${target}`)}`, source_person_id:source, target_person_id:target, relation_type:text(r.relation_type || r.relation_code || r.type || "related"), display_label:label, reverse_type:text(r.reverse_type || r.reverse_relation_type), reverse_label:text(r.reverse_label || r.reverse_relation_label), evidence_lines:array(r.evidence_lines), evidence:text(r.evidence || r.reason || r.relation_description), timeline:text(r.timeline || r.timeline_scope || "current"), worldline:text(r.worldline || r.worldline_id || "main"), confidence:Number(r.confidence || .75), origin:"ai", manual_locked:false, disabled:false };
    }).filter(Boolean);
    const manual = array(previous.relationships).filter((r)=>r.manual_locked && !r.disabled);
    return [...new Map([...auto,...manual].map((r)=>[r.relation_id,r])).values()];
  }
  function normalizeAliases(data, c, previous) {
    const sourceRows=aliasRows(data);
    const auto = sourceRows.map((a) => {
      const target = resolvePerson(a.target_person_id || a.target_slot || a.target_slot_id || a.target || a.target_name || a.person_id, c);
      const alias = text(a.alias || a.expression || a.keyword || a.trigger_text || a.name);
      if (!target || !alias) return null;
      const targetSlot=text(a.target_slot_id || a.slot_id);
      const confidence=Math.max(0,Math.min(1,Number(a.confidence ?? .75)));
      return {
        binding_id:text(a.binding_id||a.id)||`alias_${fingerprint(`${alias}|${target}|${targetSlot}`)}`,
        alias,target_person_id:target,target_slot_id:targetSlot,
        alias_type:text(a.alias_type || a.keyword_type || a.type || "alias"),
        valid_from_line:a.valid_from_line ?? null, valid_to_line:a.valid_to_line ?? null,
        valid_scope:text(a.valid_scope || a.scope || ""),timeline_scope:text(a.timeline_scope || a.timeline || ""),
        relation_id:text(a.relation_id),evidence:text(a.evidence || a.reason),confidence,
        keyword_enabled:a.keyword_enabled !== false && a.enabled !== false,
        origin:"ai",manual_locked:false,disabled:false,
      };
    }).filter(Boolean);
    const manual = array(previous.aliases).filter((a)=>a.manual_locked && !a.disabled).map(a=>({...a,confidence:1,keyword_enabled:a.keyword_enabled!==false}));
    return [...new Map([...auto,...manual].map((a)=>[a.binding_id,a])).values()];
  }

  function directAppearance(data) {
    if (typeof data === "string") return text(data);
    const source = object(data);
    for (const key of ["appearance","appearance_text","appearance_description","description","visual_description","外形","外形描述","人物外形细节描述"]) if (text(source[key])) return text(source[key]);
    for (const record of recordsArray(data)) {
      for (const key of ["appearance","appearance_text","appearance_description","description","visual_description","外形","外形描述","人物外形细节描述"]) if (text(record[key])) return text(record[key]);
    }
    return "";
  }
  const STAGE_ANCHORS = { 女婴儿:"亚洲女婴儿", 男婴儿:"亚洲男婴儿", 小女孩:"亚洲小女孩", 小男孩:"亚洲小男孩", 少女:"亚洲少女", 少年:"亚洲少年", 年轻女性:"亚洲年轻女性", 年轻男性:"亚洲年轻男性", 中年女性:"亚洲中年女性", 中年男性:"亚洲中年男性", 老年女性:"亚洲老年女性", 老年男性:"亚洲老年男性" };
  const NON_ASIAN_ANCHOR_EXEMPT = /欧美|西方|外国|异域|外星|非人形|兽形|纯机械|机械体|纯能量|触手|史莱姆|战舰|飞船|怪物/;
  const ELDER_MARKERS = /老年|年迈|苍老|皱纹|皱褶|银白|白发苍苍|老者|老翁|老人|暮气|鹤发|霜发/;
  const YOUNG_OR_CHILD_MARKERS = /少女|少年|年轻|青春|孩童|儿童|小女孩|小男孩|婴儿肥|奶膘|稚气|青涩|校服/;
  function stageFamily(stage = "") {
    const normalized = normalizeStage(stage);
    if (["女婴儿","男婴儿"].includes(normalized)) return "infant";
    if (["小女孩","小男孩"].includes(normalized)) return "child";
    if (["少女","少年"].includes(normalized)) return "teen";
    if (["年轻女性","年轻男性"].includes(normalized)) return "young";
    if (["中年女性","中年男性"].includes(normalized)) return "middle";
    if (["老年女性","老年男性"].includes(normalized)) return "elder";
    return "";
  }
  function shouldRelaxAsianAnchor(slot) {
    const raw = [text(slot.species), text(slot.age?.life_stage), text(slot.source_entry), text(slot.display_name)].join(" ");
    return NON_ASIAN_ANCHOR_EXEMPT.test(raw);
  }
  function expectedAppearanceAnchor(slot) {
    const normalized = normalizedVisibleStage(slot);
    if (!normalized || shouldRelaxAsianAnchor(slot)) return "";
    return STAGE_ANCHORS[normalized] || "";
  }
  function normalizeAppearanceIdentityPrefix(slot, candidate) {
    const safe = text(candidate).slice(0, MAX_APPEARANCE);
    const anchor = expectedAppearanceAnchor(slot);
    const stage = normalizedVisibleStage(slot);
    if (!safe || !anchor || safe.startsWith(anchor)) return safe;
    // Only fix an unambiguous missing "亚洲" prefix or reversed word order.
    // A different gender/age stage is never rewritten locally.
    if (stage && safe.startsWith(stage)) return `${anchor}${safe.slice(stage.length)}`;
    const reversed = anchor.replace(/^亚洲(老年|中年|年轻)/, "$1亚洲");
    if (reversed !== anchor && safe.startsWith(reversed)) return `${anchor}${safe.slice(reversed.length)}`;
    return safe;
  }
  function localAppearanceReview(slot, candidate, siblings = core().slots.filter((s) => s.slot_id !== slot.slot_id)) {
    const safe = text(candidate).slice(0, MAX_APPEARANCE);
    const missing = [];
    const violations = [];
    const hardFailures = [];
    const family = stageFamily(normalizedVisibleStage(slot));
    const stage = normalizedVisibleStage(slot);
    const anchor = expectedAppearanceAnchor(slot);
    const gender = normalizeGender(slot.gender || genderFromStage(stage));
    if (!safe) { missing.push("未返回外形正文"); hardFailures.push("未返回外形正文"); }
    if (anchor && !safe.startsWith(anchor)) { const item=`开头必须写“${anchor}”`; missing.push(item); hardFailures.push(item); }
    if (gender === "女" && /男性|男人|男子|男生|男士|男孩|男婴|少年人物|亚洲少年|亚洲年轻男性|亚洲中年男性|亚洲老年男性|雄性/.test(safe)) {
      const item="性别污染：女性人物正文出现男性主体特征"; violations.push(item); hardFailures.push(item);
    }
    if (gender === "男" && /女性|女人|女子|女生|女士|女孩|女婴|少女人物|亚洲少女|亚洲年轻女性|亚洲中年女性|亚洲老年女性|雌性/.test(safe)) {
      const item="性别污染：男性人物正文出现女性主体特征"; violations.push(item); hardFailures.push(item);
    }
    const firstSentence = firstAppearanceSentence(safe);
    const otherAnchors = Object.entries(STAGE_ANCHORS).filter(([otherStage])=>otherStage!==stage).map(([,value])=>value);
    const conflictingAnchor = otherAnchors.find((value)=>firstSentence.includes(value));
    if (conflictingAnchor) { const item=`阶段主体冲突：出现“${conflictingAnchor}”`; violations.push(item); hardFailures.push(item); }
    if (["infant","child","teen","young"].includes(family) && ELDER_MARKERS.test(safe)) { const item="年龄阶段污染：出现老年特征"; violations.push(item); hardFailures.push(item); }
    if (family === "elder" && YOUNG_OR_CHILD_MARKERS.test(safe)) { const item="年龄阶段污染：出现年轻或幼态特征"; violations.push(item); hardFailures.push(item); }
    if (family === "middle" && /(婴儿|小女孩|小男孩|孩童|婴儿肥|奶膘|稚气)/.test(safe)) { const item="年龄阶段污染：出现儿童或幼态特征"; violations.push(item); hardFailures.push(item); }
    const elderSibling = siblings.some((s) => stageFamily(normalizedVisibleStage(s)) === "elder");
    if (elderSibling && family !== "elder" && ELDER_MARKERS.test(safe)) { const item="疑似串用了同批老年角色特征"; violations.push(item); hardFailures.push(item); }
    const hasOnePiece = /连衣裙|长袍|旗袍|礼服|斗篷|法袍|战甲|盔甲|作战服|套装|长衫|袍服/.test(safe);
    const hasTop = /外套|大衣|风衣|西装|夹克|衬衫|上衣|T恤|卫衣|针织|毛衣|开衫|马甲|内搭|背心|披肩|斗篷|袄|衫/.test(safe);
    const hasBottom = /长裤|短裤|阔腿裤|西裤|牛仔裤|半裙|长裙|短裙|百褶裙|裙摆|裤装|下装/.test(safe);
    const hasFace = /脸|脸型|鹅蛋脸|瓜子脸|圆脸|方脸|下颌线|眉骨|眉眼|眼型|眼眸|鼻梁|鼻尖|鼻骨|唇形|嘴唇|肤色|肤质|皮肤/.test(safe);
    const hasHair = /头发|发型|发色|短发|长发|卷发|直发|马尾|丸子头|盘发|刘海|碎发/.test(safe);
    const hasClothingColor = /(?:身穿|穿着|穿|外搭|内搭|下身|披着|搭配).{0,18}(黑色|白色|灰色|银灰|米色|卡其|藏蓝|深蓝|浅蓝|天蓝|青色|墨绿|绿色|酒红|红色|粉色|裸粉|紫色|金色|棕色|咖色|驼色|橙色|黄色)/.test(safe) || /(黑色|白色|灰色|银灰|米色|卡其|藏蓝|深蓝|浅蓝|天蓝|青色|墨绿|绿色|酒红|红色|粉色|裸粉|紫色|金色|棕色|咖色|驼色|橙色|黄色).{0,12}(外套|大衣|风衣|西装|夹克|衬衫|上衣|针织|毛衣|长裤|短裤|裙|袍|靴|鞋)/.test(safe);
    const hasShoes = /皮鞋|运动鞋|短靴|长靴|高跟鞋|布鞋|凉鞋|马丁靴|靴子|鞋履|鞋面|鞋跟|脚穿|脚踩/.test(safe);
    const hasAccessory = /耳环|耳钉|项链|吊坠|手链|手镯|戒指|发簪|发饰|胸针|领带|腰带|腕表|手表|眼镜|饰品|配饰/.test(safe);
    const hasTemperament = /气质|气场|风格|神态|观感|氛围/.test(safe);
    if (!hasFace) missing.push("缺少脸部或五官信息");
    if (!hasHair) missing.push("缺少头发或发型信息");
    if (!(hasOnePiece || (hasTop && hasBottom))) missing.push("缺少完整穿着层次（上装/下装或裙袍一体）");
    if (!hasClothingColor) missing.push("缺少服装主色");
    if (!hasAccessory) missing.push("缺少装饰或配饰");
    if (!hasShoes) missing.push("缺少鞋履");
    if (!hasTemperament) missing.push("缺少整体气质");
    const passed = Boolean(safe) && !missing.length && !violations.length;
    const identity_passed = Boolean(safe) && !hardFailures.length;
    return { passed, identity_passed, hard_failures:unique(hardFailures), implemented:[], missing:unique(missing), violations:unique(violations), summary: passed ? "本地年龄阶段、性别与穿着完整性校验通过" : unique([...violations, ...missing]).join("；") };
  }
  function normalizeReview(review = {}) {
    return { passed:Boolean(review.passed), review_error:Boolean(review.review_error), implemented:unique(array(review.implemented).map(text).filter(Boolean)), missing:unique(array(review.missing).map(text).filter(Boolean)), violations:unique(array(review.violations).map(text).filter(Boolean)), summary:text(review.summary) };
  }
  function mergeAppearanceReviews(...reviews) {
    const normalized = reviews.filter(Boolean).map(normalizeReview);
    const implemented = unique(normalized.flatMap((review)=>review.implemented));
    const missing = unique(normalized.flatMap((review)=>review.missing));
    const violations = unique(normalized.flatMap((review)=>review.violations));
    const passed = !missing.length && !violations.length && normalized.every((review)=>review.review_error || review.passed);
    const review_error = normalized.some((review)=>review.review_error);
    const summary = passed ? (normalized.map((review)=>review.summary).find(Boolean) || "校验通过") : unique([...violations, ...missing]).join("；");
    return { passed, review_error, implemented, missing, violations, summary };
  }
  async function emitCharacterCoreTrace(payload = {}) {
    if (typeof fetch !== "function") return;
    try {
      await fetch(`/api/character-core/trace?session=${encodeURIComponent(currentInstance())}`, {
        method:"POST",
        headers:{ "Content-Type":"application/json", "X-VideoPromptTool-Session":currentInstance() },
        body:JSON.stringify({ ...payload, instance_id:currentInstance() }),
        cache:"no-store",
        keepalive:true,
      });
    } catch (_) {}
  }
  const detailInstruction = (level) => ({ 简洁:"简洁准确，只写静态外形核心信息。", 标准:"完整写出脸部五官、发型、服装、装饰、鞋子与气质，但不写剧情和心理。", 详细:"主角级外形密度，突出脸蛋、头发、外貌、衣服、装饰、鞋子、气质，禁止剧情化叙述。", 极致:"高密度角色设计稿级外形描写，只展开静态外形层，不写经历、动作过程、心理活动或剧情状态。" }[level] || "完整自然描写，只写静态外形。");
  function stageContract(slot) {
    return { old_visual_age_stage:text(slot.generated_values?.appearance_stage || ""), new_visual_age_stage:visibleStage(slot), chronological_age:text(slot.age?.chronological_age), life_stage:text(slot.age?.life_stage), timeline_stage:text(slot.age?.timeline_stage || "当前时间线"), preserve:["人物核心身份","连续可辨识的骨相与五官特征","标志性发色或身体特征","原文明示特征"], redesign:["面部成熟度","皮肤状态","体态","发型结构","服装剪裁材质","配饰鞋履","阶段气质"] };
  }
  function buildAppearancePromptPackage(slot, forceStage = false) {
    try { globalThis.__commitLatestEditableUiState?.("before_character_appearance_prompt"); } catch (_) {}
    const anchor = expectedAppearanceAnchor(slot);
    const subjectFacts = buildAppearanceSubjectFacts(slot);
    const worldContext = buildAppearanceWorldContext();
    const siblingSummary = buildAppearanceSiblingSummary(slot);
    const excludedContext = buildAppearanceExcludedContext(slot);
    const evidenceLines = array(subjectFacts.evidence_fragments).length ? subjectFacts.evidence_fragments.join("\n") : "无直接描写证据，按当前锁定阶段、主体事实和世界观稳定设计。";
    const prompt = [
      "【任务】CharacterCore 2.0全题材单人物外形生成。禁止使用固定现代、古代、民国、校园或其他人物模板。",
      (globalThis.__v23GetInstruction?.("global") ? `【Hotfix23全局AI语义指令｜协议锁定】\n${globalThis.__v23GetInstruction("global")}` : ""),
      (globalThis.__v23GetInstruction?.(globalThis.__v23CharacterInstructionMode || "single_character") ? `【Hotfix23当前人物外形自定义指令｜只改appearance语义】\n${globalThis.__v23GetInstruction(globalThis.__v23CharacterInstructionMode || "single_character")}` : ""),
      "【外形主体事实】", JSON.stringify(subjectFacts),
      "【阶段重构契约】", JSON.stringify(stageContract(slot)),
      "【阶段规则】用户提供的visual_age_stage是唯一硬阶段来源。chronological_age、life_stage、timeline_stage 只能补充细节，不能把当前人物改写成别的年龄阶段。" + (forceStage || slot.appearance_stage_stale ? "本次必须按新阶段重构完整外形，同时保持同一人物核心辨识连续性。" : ""),
      "【文化锚点】", anchor ? `当前人物默认文化锚点为“亚洲”。正文第一短句必须直接写“${anchor}”，不得改成其他阶段词，也不得省略“亚洲”。若别名、称谓、他人文本与当前阶段冲突，一律忽略冲突项，以 visual_age_stage 为准。` : "若原文明示非人形或非亚洲设定，可不强制使用“亚洲+阶段”开头，但仍必须严格服从当前锁定阶段。",
      "【世界观摘要】", JSON.stringify(worldContext),
      "【角色身份摘要】", JSON.stringify({ identity_summary: subjectFacts.identity_summary, appearance_constraints: subjectFacts.appearance_constraints }),
      "【人物局部证据】", evidenceLines,
      "【全局外形方向】", String(valueOf("#appearanceReference") || "").slice(0, 6000) || "无。由AI根据当前小说自由设计。",
      "【当前人物卡字段｜唯一有效值】", JSON.stringify({display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:slot.age?.chronological_age||"",life_stage:slot.age?.life_stage||"",timeline_stage:slot.age?.timeline_stage||"当前时间线",species:slot.species||"",detail_level:slot.appearance_detail_level||"详细"}),
      "【当前人物外形正文｜改写基础】", text(slot.appearance) || "当前为空，请从零生成。",
      "【当前人物修改意见｜最高优先级】", slot.appearance_revision_note || "无",
      "【当前值原则】上面当前人物卡字段、当前外形正文和修改意见均来自用户此刻界面，必须以这些值为准；不得读取或恢复上一次AI原始返回。若当前外形已有可用细节，在满足当前阶段、性别与修改意见的前提下进行完整改写。",
      "【详细程度】", `${slot.appearance_detail_level || "详细"}：${detailInstruction(slot.appearance_detail_level)}`,
      "【同批差异参考｜仅防撞型】", JSON.stringify(siblingSummary),
      "【正文范围】只写静态人物外形，不写剧情过程、奔波经历、疲惫原因、受伤落寞、防御姿态、心理活动、动作过程、对白、环境叙述或故事解释；不要写具体年龄数字，只保留当前年龄阶段。",
      "【外形必写项】必须自然写出：脸蛋/脸型与五官、肤色肤质、头发/发型发色、体态、衣服（上装+下装，或裙袍一体）、装饰/配饰、鞋子、整体气质。",
      "【表达要求】使用一整段自然中文，优先简洁清楚，不要写成小传，不要堆叠剧情形容词，不要出现“他是一位二十八九岁……因为……所以……”这类叙事句式。",
      "【污染隔离】不要参考 source_entry、原始别名称谓、existing_appearance、同批其他人物外形正文、整篇小说无关长文本；若这些信息与当前锁定阶段冲突，全部忽略。",
      "【最终返回】只返回一整段完整人物外形正文。不要JSON、标题、解释、自检或字段名。尽量控制在80到180个中文字符内，必要时可略长；核心是纯外形、可直接生图。" + (anchor ? ` 第一短句必须以“${anchor}”开头。` : ""),
    ].filter(Boolean).join("\n\n");
    return { prompt, subjectFacts, worldContext, siblingSummary, excludedContext, promptProfile:"appearance_clean", anchor };
  }
  function appearancePrompt(slot, forceStage = false) {
    return buildAppearancePromptPackage(slot, forceStage).prompt;
  }
  function revisionReviewScore(review={}) {
    return array(review.missing).length + array(review.violations).length * 2;
  }
  async function verifyAppearanceRevision(slot, candidate) {
    if (!text(slot.appearance_revision_note)) return { passed:true, implemented:[], missing:[], violations:[], summary:"未填写修改意见" };
    const prompt = [
      "【任务】只验收候选人物外形是否语义落实当前人物修改意见，不重写人物。",
      "【当前人物硬事实】", JSON.stringify({display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:slot.age?.chronological_age||"",life_stage:slot.age?.life_stage||"",timeline_stage:slot.age?.timeline_stage||"",species:slot.species||""}),
      "【用户修改意见】", slot.appearance_revision_note,
      "【候选完整外形】", candidate,
      "【返回】只返回JSON：{passed,implemented,missing,violations,summary}。理解同义表达；禁止项即使以‘不佩戴/没有’方式复述也算violations。",
    ].join("\n\n");
    try {
      const data=object(await aiAnalyze(prompt,REVISION_CHECK_MARKER,0));
      const missing=array(data.missing).map(text).filter(Boolean),violations=array(data.violations).map(text).filter(Boolean),implemented=array(data.implemented).map(text).filter(Boolean);
      return { passed:Boolean(data.passed) && !missing.length && !violations.length, implemented, missing, violations, summary:text(data.summary) };
    } catch(error) {
      return { passed:true, implemented:[], missing:[], violations:[], summary:`验收请求未完成，不阻拦写入：${error?.message||error}`, review_error:true };
    }
  }
  async function requestAppearanceText(requestPackage, retries=1) {
    let lastError=null;
    for(let attempt=0;attempt<=retries;attempt++){
      try{
        const prompt=attempt ? `${requestPackage.prompt}\n\n【补充说明】上一次没有返回可读取的完整正文。本次只直接返回一整段完整外形。` : requestPackage.prompt;
        const data=await aiAnalyze(prompt,APPEARANCE_MARKER,attempt?0:1);
        const result=directAppearance(data).slice(0,MAX_APPEARANCE);if(result)return {result,error:null,meta:object(data?._ai_meta)};
        lastError=new Error(APPEARANCE_WRITEBACK_ERROR);
      }catch(error){lastError=error instanceof Error?error:new Error(String(error||APPEARANCE_WRITEBACK_ERROR));}
    }
    return {result:"",error:lastError||new Error(APPEARANCE_WRITEBACK_ERROR),meta:{}};
  }
  async function requestAppearanceIdentityRepair(slot, requestPackage, candidate, localReview) {
    const anchor=expectedAppearanceAnchor(slot);
    const prompt=[
      requestPackage.prompt,
      "【身份纠错重写｜最高优先级】上一次候选发生人物串线、性别/年龄阶段冲突或主体锚点错误。必须从零重写完整外形，禁止只改开头，禁止保留错误候选中的性别、年龄、发型或服装。",
      "【当前唯一硬事实】", JSON.stringify({display_name:slot.display_name,gender:slot.gender,visual_age_stage:normalizedVisibleStage(slot),required_anchor:anchor,species:slot.species||""}),
      "【上次错误候选｜只用于识别并彻底排除错误】", candidate,
      "【必须修复的问题】", JSON.stringify({hard_failures:array(localReview.hard_failures),violations:array(localReview.violations),missing:array(localReview.missing)}),
      `【返回】只返回一整段重写后的完整人物外形正文。${anchor?`第一短句必须以“${anchor}”开头。`:""}必须包含完整服装层次、材质、主色与鞋履。不要解释。`,
    ].join("\n\n");
    try{
      const data=await aiAnalyze(prompt,APPEARANCE_MARKER,0,{slot_id:slot.slot_id,slot_token:slot.slot_token});
      const result=directAppearance(data).slice(0,MAX_APPEARANCE);
      return {result,error:result?null:new Error(APPEARANCE_WRITEBACK_ERROR),meta:object(data?._ai_meta)};
    }catch(error){return {result:"",error:error instanceof Error?error:new Error(String(error||APPEARANCE_WRITEBACK_ERROR)),meta:{}};}
  }
  async function generateAppearance(slot, { forceStage = false, __stageRetry = false } = {}) {
    const facts=characterFactsCompleteness(slot);
    if(!facts.complete){const error=new Error(`人物事实未完整，外形生成已跳过：${facts.critical_missing.join("、")||"剧情年龄阶段待AI确认"}`);error.code="CHARACTER_FACTS_INCOMPLETE";return {appearance:text(slot?.appearance),error,review:{passed:false,implemented:[],missing:facts.critical_missing,violations:[],summary:error.message},success:false,failed:true,repaired:false,advisory:false,slot_id:slot?.slot_id,display_name:slot?.display_name,facts_incomplete:true};}
    const requestState={source_hash:core().source_hash,slot_id:slot.slot_id,character_revision:slot.character_revision,visual_age_stage:visibleStage(slot),gender:slot.gender,instance_id:currentInstance()};
    emitCharacterCoreTrace({trace_stage:"appearance_request_snapshot",slot_id:slot.slot_id,slot_token:slot.slot_token,display_name:slot.display_name,visual_age_stage:visibleStage(slot),raw_stage:object(slot.age).visual_age_stage,stage_source:object(slot.age).stage_source,character_revision:slot.character_revision,gender:slot.gender,slots_len:core().slots.length,slot_index:core().slots.findIndex((s)=>s===slot),slots_includes:core().slots.includes(slot),people_len:core().people.length}).catch(()=>{});
    const old=text(slot.appearance);let error=null;
    const promptPackage=buildAppearancePromptPackage(slot,forceStage);
    const evaluateCandidate = async (candidate) => {
      if(!candidate)return { localReview:{passed:false,identity_passed:false,hard_failures:["未返回外形正文"],implemented:[],missing:["未返回外形正文"],violations:[],summary:error?.message||APPEARANCE_WRITEBACK_ERROR}, revisionReview:{passed:true,implemented:[],missing:[],violations:[],summary:"未填写修改意见"}, review:{passed:false,implemented:[],missing:["未返回外形正文"],violations:[],summary:error?.message||APPEARANCE_WRITEBACK_ERROR} };
      const localReview=localAppearanceReview(slot,candidate);
      const revisionReview=await verifyAppearanceRevision(slot,candidate);
      return { localReview, revisionReview, review:mergeAppearanceReviews(localReview, revisionReview) };
    };
    const candidateRank=(evaluated)=>{
      const local=object(evaluated?.localReview),merged=object(evaluated?.review);
      return array(local.hard_failures).length*1000+array(merged.violations).length*100+array(merged.missing).length*10+(merged.review_error?1:0);
    };
    const first=await requestAppearanceText(promptPackage,1);
    let best=normalizeAppearanceIdentityPrefix(slot,first.result);
    let bestMeta=object(first.meta);error=first.error;
    let evaluated=await evaluateCandidate(best);
    const initialCandidate=best;
    const initialMeta=object(bestMeta);
    const initialReview=evaluated.review;
    const traceBase={slot_id:slot.slot_id,slot_token:slot.slot_token,display_name:slot.display_name,visual_age_stage:normalizedVisibleStage(slot),subject_facts:promptPackage.subjectFacts,excluded_context:promptPackage.excludedContext,prompt_profile:promptPackage.promptProfile,anchor:promptPackage.anchor};
    if(text(bestMeta.request_id))await emitCharacterCoreTrace({trace_stage:"appearance_prompt_prepared",request_id:text(bestMeta.request_id),...traceBase});
    if(best)await emitCharacterCoreTrace({trace_stage:"appearance_local_validation",request_id:text(bestMeta.request_id),...traceBase,passed:evaluated.localReview.passed,identity_passed:evaluated.localReview.identity_passed,summary:evaluated.localReview.summary,missing:evaluated.localReview.missing,violations:evaluated.localReview.violations,hard_failures:evaluated.localReview.hard_failures,candidate_text:best,diagnostic_flags:appearanceDiagnosticFlags(evaluated.review),first_sentence:firstAppearanceSentence(best)});

    // Hard identity failures get one isolated AI rewrite. Missing clothing
    // details remain advisory and never roll back an otherwise correct card.
    if(best && !evaluated.localReview.identity_passed){
      const repaired=await requestAppearanceIdentityRepair(slot,promptPackage,best,evaluated.localReview);
      const repairedCandidate=normalizeAppearanceIdentityPrefix(slot,repaired.result);
      if(repairedCandidate){
        const repairedEvaluated=await evaluateCandidate(repairedCandidate);
        await emitCharacterCoreTrace({trace_stage:"appearance_identity_repair_validation",request_id:text(repaired.meta?.request_id),...traceBase,passed:repairedEvaluated.localReview.passed,identity_passed:repairedEvaluated.localReview.identity_passed,summary:repairedEvaluated.localReview.summary,missing:repairedEvaluated.localReview.missing,violations:repairedEvaluated.localReview.violations,hard_failures:repairedEvaluated.localReview.hard_failures,candidate_text:repairedCandidate,diagnostic_flags:appearanceDiagnosticFlags(repairedEvaluated.review),first_sentence:firstAppearanceSentence(repairedCandidate)});
        if(candidateRank(repairedEvaluated)<candidateRank(evaluated)){
          best=repairedCandidate;bestMeta=object(repaired.meta);evaluated=repairedEvaluated;error=repaired.error;
        }
      }else if(repaired.error){error=repaired.error;}
    }

    const {localReview,review}=evaluated;
    const active=core().slots.find(s=>s.slot_id===requestState.slot_id);
    // 外形写回只在身份（性别/剧情年龄阶段）或会话（原文/实例）真实变化时丢弃。
    // character_revision 会因 facts 补全、关系分析等并发操作频繁递增，若把它也作为
    // 丢弃依据，合法外形结果会被误弃（previous_appearance 永远为空），人物卡缺失外形。
    const identityAligned=Boolean(active)&&active===slot&&visibleStage(slot)===requestState.visual_age_stage&&(requestState.gender==null||normalizeGender(slot.gender)===normalizeGender(requestState.gender));
    const sessionAligned=core().source_hash===requestState.source_hash&&currentInstance()===requestState.instance_id;
    if(!active||active!==slot||!sessionAligned||!identityAligned){
      const stageMoved = active && active === slot && visibleStage(slot) !== requestState.visual_age_stage && !slot.manual_values?.visual_age_stage && !slot.age?.stage_locked;
      if (stageMoved && !__stageRetry) {
        // 请求期间剧情年龄阶段被 AI 事实补全更新（如「成年阶段」→「年轻女性」）：
        // 旧锚点的外形不再适用，按最新阶段重试一次，避免合法结果被丢弃。
        await emitCharacterCoreTrace({trace_stage:"appearance_stage_retry",request_id:text(bestMeta.request_id),...traceBase,outcome:"stage_changed_retry",summary:`阶段在请求期间由AI推断更新（${requestState.visual_age_stage||"空"} → ${visibleStage(slot)||"空"}），按新阶段重试一次外形生成。`,candidate_text:best,previous_appearance:old}).catch(()=>{});
        return generateAppearance(slot, { forceStage, __stageRetry: true });
      }
      await emitCharacterCoreTrace({trace_stage:"appearance_writeback",request_id:text(bestMeta.request_id),...traceBase,outcome:"stale_discarded",final_outcome:"stale_discarded",summary:"响应到达时人物卡已更新，本次结果已丢弃。",candidate_text:best,previous_appearance:old,diagnostic_flags:appearanceDiagnosticFlags(review),first_sentence:firstAppearanceSentence(best),stale_diagnostics:{active_exists:Boolean(active),active_same_reference:active===slot,source_hash_changed:core().source_hash!==requestState.source_hash,current_source_hash:core().source_hash,requested_source_hash:requestState.source_hash,slot_revision:slot.character_revision,requested_revision:requestState.character_revision,visible_stage_changed:visibleStage(slot)!==requestState.visual_age_stage,current_visible_stage:visibleStage(slot),requested_visible_stage:requestState.visual_age_stage,instance_changed:currentInstance()!==requestState.instance_id,current_instance:currentInstance(),requested_instance:requestState.instance_id}});
      return {appearance:old,error,review,stale:true,discarded:true,success:false,failed:true,slot_id:slot.slot_id,display_name:slot.display_name};
    }
    if(best && localReview.identity_passed){
      backfillSlotIdentityFromAppearance(slot,best);
      slot.appearance=best;slot.appearance_status="ready";slot.appearance_revision+=1;slot.appearance_stage_stale=false;
      slot.generated_values={...object(slot.generated_values),appearance:best,appearance_stage:visibleStage(slot),generated_at:new Date().toISOString(),revision_review:review};
      // AI生成/外形反推属于 generated/current 值，绝不能冒充用户 manual_values。
      // 否则“待AI判断/未定”或一次AI推断会被永久锁死，后续人物事实与关系链无法正常更新。
      const preservedManual=sanitizeManualIdentityValues(slot);
      slot.manual_values={...preservedManual};
      slot.current_values={...object(slot.current_values),display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:text(slot.age?.chronological_age),life_stage:text(slot.age?.life_stage),timeline_stage:text(slot.age?.timeline_stage||"当前时间线"),species:text(slot.species),appearance:best,appearance_revision_note:text(slot.appearance_revision_note)};
      slot.appearance_revision_review=review;slot.appearance_revision_status=review.review_error?"review_unavailable":review.passed?"passed":"needs_review";slot.character_revision+=1;
      const repaired=best!==initialCandidate;
      const outcome=review.passed?(repaired?"accepted_after_identity_repair":"accepted_initial"):(repaired?"accepted_repair_with_advisory":"accepted_with_advisory");
      const summary=review.passed?(repaired?"人物串线已自动纠正，修复稿通过并写回。":"首稿已通过并写回。"):`${review.summary}；身份与年龄阶段已确认，当前稿已写回，请复核缺失细节。`;
      await emitCharacterCoreTrace({trace_stage:"appearance_writeback",request_id:text(bestMeta.request_id||initialMeta.request_id),...traceBase,outcome,final_outcome:outcome,passed:review.passed,identity_passed:true,summary,missing:review.missing,violations:review.violations,candidate_text:best,previous_appearance:old,initial_candidate_text:initialCandidate,initial_summary:text(initialReview?.summary),diagnostic_flags:appearanceDiagnosticFlags(review),first_sentence:firstAppearanceSentence(best)});
      return {appearance:best,error:null,review,success:true,failed:false,repaired,advisory:Boolean(!review.passed),summary,slot_id:slot.slot_id,display_name:slot.display_name};
    }
    slot.appearance_status=slot.appearance?"ready":"missing";
    const hardSummary=unique([...array(localReview.hard_failures),...array(localReview.violations),...array(localReview.missing)]).join("；")||review.summary||APPEARANCE_WRITEBACK_ERROR;
    const finalError=error||new Error(`AI返回的人物身份/性别/年龄阶段与当前人物卡冲突，已保留原外形：${hardSummary}`);
    await emitCharacterCoreTrace({trace_stage:"appearance_writeback",request_id:text(bestMeta.request_id||initialMeta.request_id),...traceBase,outcome:"identity_conflict_rejected",final_outcome:"identity_conflict_rejected",passed:false,identity_passed:false,summary:finalError.message,missing:review.missing,violations:review.violations,hard_failures:localReview.hard_failures,candidate_text:best,previous_appearance:old,initial_candidate_text:initialCandidate,diagnostic_flags:appearanceDiagnosticFlags(review),first_sentence:firstAppearanceSentence(best)});
    return {appearance:old,error:finalError,review,success:false,failed:true,repaired:false,advisory:false,slot_id:slot.slot_id,display_name:slot.display_name};
  }
  async function generateAll({ onlySlotId = "", forceStage = false } = {}) {
    const c = core(); const requestedTargets = (onlySlotId ? c.slots.filter((s)=>s.slot_id===onlySlotId) : c.slots).filter(s=>!s.disabled);
    const blocked=requestedTargets.filter(slot=>!characterFactsCompleteness(slot).complete);
    const targets=requestedTargets.filter(slot=>characterFactsCompleteness(slot).complete);
    let cursor=0, complete=0;
    const results=[],successes=[],failures=blocked.map((slot,index)=>({index:index+1,name:slot.display_name,error:`人物事实未完整：${characterFactsCompleteness(slot).critical_missing.join("、")||"剧情年龄阶段待AI补全"}`,facts_incomplete:true,stale:false})),advisories=[],activeSlots=new Set();
    const progress=(msg)=>{const n=q("#characterGenerationProgress");if(n)n.textContent=msg;};
    const refreshProgress=()=>{
      if(!requestedTargets.length){progress("人物卡生成完成：成功 0，失败 0");return;} if(!targets.length){progress(`人物卡生成暂停：${blocked.length} 个人物事实未完整，未进入外形生成`);return;}
      if(complete>=targets.length){
        progress(`人物卡生成完成：成功 ${successes.length}，失败 ${failures.length}${blocked.length?`；事实未完整 ${blocked.length}`:""}${advisories.length?`；建议复核 ${advisories.length}`:""}`);
        return;
      }
      progress(`已完成 ${complete}/${targets.length}${blocked.length?`，事实待补全 ${blocked.length}`:""}${activeSlots.size?`，正在生成 ${activeSlots.size} 张`:""}`);
    };
    refreshProgress();
    const workers=Array.from({length:Math.min(APPEARANCE_CONCURRENCY,Math.max(1,targets.length))},async()=>{
      while(cursor<targets.length){
        const slot=targets[cursor++], targetIndex=cursor, activeKey=slot.slot_id||slot.display_name||String(targetIndex);
        activeSlots.add(activeKey);refreshProgress();
        let result;
        try{result=await generateAppearance(slot,{forceStage});}
        catch(error){result={appearance:text(slot.appearance),error:error instanceof Error?error:new Error(String(error||APPEARANCE_WRITEBACK_ERROR)),review:{passed:false,implemented:[],missing:["未返回外形正文"],violations:[],summary:error?.message||String(error||APPEARANCE_WRITEBACK_ERROR)},success:false,failed:true,repaired:false,advisory:false,slot_id:slot.slot_id,display_name:slot.display_name};}
        results.push(result);
        if(result?.success){
          successes.push(slot.display_name);
          if(result?.advisory)advisories.push({index:targetIndex,name:slot.display_name,summary:text(result?.summary||result?.review?.summary||"建议复核")});
        }
        else failures.push({index:targetIndex,name:slot.display_name,error:text(result?.error?.message||result?.review?.summary||APPEARANCE_WRITEBACK_ERROR),stale:Boolean(result?.stale)});
        complete++;activeSlots.delete(activeKey);syncLegacyState();renderCharacters();if(typeof renderScenes==="function")renderScenes();if(typeof scheduleDraftSave==="function")scheduleDraftSave();refreshProgress();
      }
    });
    await Promise.all(workers);refreshProgress();return {targets,requestedTargets,blocked,results,successes,failures,advisories,total:requestedTargets.length,success_count:successes.length,failure_count:failures.length,blocked_count:blocked.length,advisory_count:advisories.length};
  }

  async function runCharacters(options = {}) {
    try { globalThis.__commitLatestEditableUiState?.("before_forced_roster_generation"); } catch (_) {}
    const button=options.button||q("#optimizeAllCharactersBtn"); if(typeof setButtonBusy==="function")setButtonBusy(button,true,options.onlySlotId?"按当前人物卡重写外形中":"强制名单人物链路生成中");
    try {
      const requestSnapshot=captureSourceSnapshot("forced_roster_request"),st=analysisState();
      const onlySlotId=options.onlySlotId||"";
      const rosterTransaction=onlySlotId?null:formalRosterSnapshot("forced_roster_generation",true);
      if(!onlySlotId&&!String(rosterTransaction?.raw_text||"").trim()){
        setAnalysisStage("slots",false,"强制人物名单为空");
        showFormalRosterError("正式人物强制名单当前为空。人物卡只读取这个名单；原文、临时人物和其他输入框不会自动补正式人物。");
        const e=new Error("强制人物名单为空。");e.code="FORMAL_ROSTER_EMPTY";throw e;
      }
      showFormalRosterError("");
      const progress=q("#characterGenerationProgress");if(progress&&!onlySlotId)progress.textContent=`已冻结当前正式人物名单 revision ${rosterTransaction.revision}，本地读取 ${rosterTransaction.entries?.length||0} 个条目；正在建立正式人物槽位……`;
      const currentContentType=text(valueOf("#genre")||appState.characterAppearanceContext?.content_type||"");
      const currentUnifiedStyle=text(valueOf("#trailerStyle")||appState.characterAppearanceContext?.trailer_style||appState.characterAppearanceContext?.unified_style||"");
      const currentCameraStyle=text(valueOf("#camera")||appState.characterAppearanceContext?.camera||"");
      const currentStyleUsable=Boolean(currentContentType||currentUnifiedStyle||currentCameraStyle);
      if(!currentStyleUsable)throw new Error("当前内容类型与统一风格均为空，请先生成或手动填写后再生成人物卡。");
      // Hotfix28: 页面当前已有内容类型/统一风格就是当前有效值。原文revision或旧style revision
      // 只能作为诊断，不能要求用户重复生成风格；人物卡直接使用当前可见内容继续。
      st.style_ready=true;
      st.style_revision=Number(requestSnapshot.source_revision||st.style_revision||0);
      st.source_hash=requestSnapshot.source_hash;
      appState.characterAppearanceContext={...object(appState.characterAppearanceContext),content_type:currentContentType||text(appState.characterAppearanceContext?.content_type),trailer_style:currentUnifiedStyle||text(appState.characterAppearanceContext?.trailer_style),camera:currentCameraStyle||text(appState.characterAppearanceContext?.camera)};
      console.info("[STYLE_REUSE_TRACE]",{source_hash:requestSnapshot.source_hash,source_revision:requestSnapshot.source_revision,style_revision:st.style_revision,content_type_present:Boolean(currentContentType),unified_style_present:Boolean(currentUnifiedStyle),camera_present:Boolean(currentCameraStyle),reused_current_style:true});
      let parsed={slot_count:core().slots.length};
      if(!onlySlotId){
        parsed=await parseSlots({preserve:true,rosterSnapshot:rosterTransaction});
        assertFormalRosterFresh(rosterTransaction,"人物槽位建立");
        if(!parsed.slot_count){setAnalysisStage("slots",false,"人物名单协议异常");const e=new Error("强制人物名单已有内容，但人物槽位解析结果为空。请打开诊断中心查看 Formal Roster Authority。 ");e.code="FORMAL_ROSTER_PROTOCOL_MISMATCH";throw e;}
      }else{
        const currentSlot=core().slots.find((item)=>item.slot_id===onlySlotId);
        if(!currentSlot)throw new Error("当前人物槽位不存在，请重新按强制名单生成人物卡。");
        currentSlot.age ||= {};
        currentSlot.manual_values ||= {};
        currentSlot.current_values ||= {};
        const currentStage=visibleStage(currentSlot);
        if(currentStage&&!isUnresolvedStage(currentStage)){
          currentSlot.age.stage_locked=true;
          currentSlot.age.stage_source="user_manual";
          currentSlot.age.stage_lock_state="manual_locked";
          currentSlot.manual_values.visual_age_stage=currentStage;
          currentSlot.current_values.visual_age_stage=currentStage;
        }
        currentSlot.character_revision+=1;
        const anchor=expectedAppearanceAnchor(currentSlot);
        console.info("[APPEARANCE_CURRENT_VALUE_TRACE]",{slot_id:currentSlot.slot_id,ui_stage:currentStage,committed_stage:visibleStage(currentSlot),locked_stage:currentStage,anchor,stage_source:currentSlot.age.stage_source});
        emitCharacterCoreTrace({trace_stage:"appearance_current_value",request_id:`appearance-current-${Date.now()}`,slot_id:currentSlot.slot_id,slot_token:currentSlot.slot_token,display_name:currentSlot.display_name,gender:currentSlot.gender,visual_age_stage:currentStage,appearance_anchor:anchor,outcome:"manual_stage_locked_before_request",summary:`当前界面阶段已重新锁定为${currentStage||"待确认"}，单人重写不再重新解析强制名单。`}).catch(()=>{});
      }
      let factsWarning="";
      if(!onlySlotId){
        assertFormalRosterFresh(rosterTransaction,"人物事实分析前");
        const p=q("#characterGenerationProgress");if(p)p.textContent=`已建立 ${core().slots.filter(s=>!s.disabled).length} 个正式人物槽位；正在判断人物事实、年龄层、别称与关系……`;
        try{await analyzeFactsAndRelations();assertFormalRosterFresh(rosterTransaction,"人物事实分析");const integrity=characterFactsIntegrity();if(!integrity.ready)factsWarning=integrity.summary;}
        catch(factsError){factsWarning=factsError?.message||String(factsError);setAnalysisStage("facts",false,factsWarning);setAnalysisStage("relationships",false,factsWarning);ensureFallbackPeopleGraph();if(p)p.textContent=`人物事实请求未完成；仅已确认性别与剧情年龄阶段的人物会继续生成外形：${factsWarning}`;emitCharacterCoreTrace({trace_stage:"core_parity_facts_relationships",request_id:`core-parity-error-${Date.now()}`,outcome:"error",error:factsWarning,summary:"人物事实/关系AI请求前端处理异常，已走保底人物链路"}).catch(()=>{});}
      }
      if(!core().people.length&&core().slots.length)ensureFallbackPeopleGraph();
      globalThis.__v23CharacterInstructionMode = onlySlotId ? "single_character" : "all_characters";
      const batch=await generateAll({onlySlotId,forceStage:Boolean(options.forceStage)});
      if(!onlySlotId)assertFormalRosterFresh(rosterTransaction,"人物外形生成");
      globalThis.__v23CharacterInstructionMode = "single_character";
      const appearanceReady=Boolean(!batch.blocked_count && (batch.success_count||core().slots.filter(s=>!s.disabled).every(s=>text(s.appearance))));
      setAnalysisStage("appearances",appearanceReady,`人物外形成功 ${batch.success_count}，失败 ${batch.failure_count}${batch.blocked_count?`；人物事实未完整 ${batch.blocked_count}`:""}`);
      rebuildKeywordIndex();syncLegacyState();renderAll();refreshSceneBindings();
      const analyzedSource=currentNovel();
      const relationReady=Boolean(analysisState().relationships_ready);
      setAnalysisStage("casting",Object.values(core().keyword_index||{}).some(item=>array(item.entries).length),relationReady?"强制名单人物卡、AI关系图与本地关键词索引已建立":"强制名单人物卡与本地关键词索引已建立；AI人物关系未完整返回但不阻断选角");
      markV77RuntimeReady(analyzedSource);
      if(typeof globalThis.__syncFinalSegmentCharacterBlocksV20==="function"){globalThis.__syncFinalSegmentCharacterBlocksV20({force:false});try{if(typeof renderSegments==="function")renderSegments();}catch(_error){}}else if(typeof clearFinalSegments==="function")clearFinalSegments();if(typeof scheduleDraftSave==="function")scheduleDraftSave();
      const c=core(), missing=c.slots.filter(s=>!s.appearance).length;
      const failedNames=array(batch.failures).sort((a,b)=>Number(a.index||0)-Number(b.index||0)).map(item=>text(item.name)).filter(Boolean);
      const failedDetails=array(batch.failures).sort((a,b)=>Number(a.index||0)-Number(b.index||0)).map((item)=>`${text(item.name)}（${text(item.error||"未返回可写入外形") || "未返回可写入外形"}）`).filter(Boolean);
      const advisoryDetails=array(batch.advisories).sort((a,b)=>Number(a.index||0)-Number(b.index||0)).map((item)=>`${text(item.name)}（${text(item.summary||"建议复核当前外形") || "建议复核当前外形"}）`).filter(Boolean);
      const progressText=`人物卡生成完成：成功 ${batch.success_count}，失败 ${batch.failure_count}${batch.advisory_count?`；建议复核 ${batch.advisory_count}`:""}`;
      const progressNode=q("#characterGenerationProgress");if(progressNode)progressNode.textContent=progressText;
      showAIStatusNotice(`CharacterCore 2.0已完成：${progressText}${missing?`；当前缺外形 ${missing}`:""}${failedNames.length?`；失败角色：${failedNames.join("、")}`:""}${failedDetails.length?`；失败原因：${failedDetails.join("；")}`:""}${advisoryDetails.length?`；建议复核：${advisoryDetails.join("；")}`:""}${factsWarning?`；人物事实仍有缺失；未确认阶段的人物已隔离，其他人物继续生成：${factsWarning}`:""}。人物卡只来自强制名单；人物关系图由AI判断；分镜人物由本地项目关键词确定性勾选。`,(batch.failure_count||batch.advisory_count||missing||factsWarning)?"warning":"ready",14000);
      return c;
    }catch(error){
      const code=text(error?.code);
      if(code==="FORMAL_ROSTER_EMPTY"||code==="FORMAL_ROSTER_PROTOCOL_MISMATCH"||code==="STALE_FORMAL_ROSTER_TRANSACTION"){
        showFormalRosterError(error?.message||"正式人物名单状态异常，请按当前名单重新生成。");
        const progress=q("#characterGenerationProgress");if(progress)progress.textContent=error?.message||"正式人物名单状态异常，请按当前名单重新生成。";
        showAIStatusNotice(error?.message||"正式人物名单状态异常，请按当前名单重新生成。","warning",12000);return null;
      }
      apiError(error?.message||String(error));showAIStatusNotice(`人物卡处理异常，但已写入内容不会回滚：${error?.message||error}`,"warning",12000);return null;
    }
    finally{globalThis.__v23CharacterInstructionMode = "single_character";if(typeof setButtonBusy==="function")setButtonBusy(button,false);
      // 无论 AI 流程成功、部分成功还是中断，只要内存里已写回人物事实/外形，
      // 都立即触发草稿保存，避免用户刷新后从旧草稿恢复到空壳人物卡。
      try{if(typeof scheduleDraftSave==="function")scheduleDraftSave();}catch(_e){}
    }
  }

  function cleanStyleAnalysisAdvice(value = "") {
    return String(value || "")
      .replace(/\r/g, "\n")
      .split(/\n+/)
      .map((line) => text(line))
      .filter(Boolean)
      .filter((line) => !/(?:最终导出|负面提示|画面限制|画质约束|最终输出首行|不参与\s*AI|negative\s*prompt|quality\s*constraint|picture\s*limit|\bno\b|\bavoid\b|4\s*K|8\s*K|logo|watermark|subtitle)/i.test(line))
      .join("\n")
      .slice(0, 2400);
  }
  function cleanChineseStylePart(value = "") {
    return text(value)
      .replace(/^\s*(?:影视质感|电影质感|实拍质感|拍摄介质|颗粒纹理|滤镜|镜头|光学纹理|噪点|畸变|光晕|对比度|饱和度|光线|明暗层次|叙事构图|构图|氛围感|整体氛围)\s*[：:]\s*/i, "")
      .replace(/[\r\n]+/g, "，")
      .replace(/[；;]+$/g, "")
      .replace(/[，,。\s]+$/g, "")
      .trim();
  }
  function v7827GlobalizeStyleComponent(key = "", value = "") {
    const part=cleanChineseStylePart(value);
    if(key==="lens_language") return "自然叙事镜头体系与克制透视及稳定空间关系";
    if(key==="lighting_layers") return "动机光优先并保留柔和高光与可读暗部及自然明暗层次";
    if(key==="narrative_composition") return "强调人物关系与空间纵深及视线组织的电影级叙事构图";
    return part;
  }
  function hasLatinText(value = "") { return /[A-Za-z]/.test(String(value || "")); }
  function hasStylePollution(value = "") {
    return /(?:不要|禁止|避免|拒绝|不得|无水印|字幕|Logo|水印|负面提示|画面限制|画质约束|人物外形|人物形象精准描写|角色外形|提示词|\bno\b|\bavoid\b|\bwithout\b|\bmust\b)/i.test(String(value || ""));
  }
  function normalizeChineseContentType(value = "") {
    return text(value)
      .replace(/^内容类型\s*[：:]\s*/, "")
      .replace(/[“”"'`]/g, "")
      .replace(/[|｜/+、，,。；;：:\s]/g, "")
      .replace(/短剧短剧$/g, "短剧")
      .trim();
  }
  const STYLE_COMPONENT_FIELDS_V78 = [
    ["cinematic_quality", "影视质感"],
    ["capture_texture", "实拍质感"],
    ["grain_texture", "颗粒纹理"],
    ["filter_tone", "滤镜"],
    ["lens_language", "镜头"],
    ["optical_texture", "光学纹理"],
    ["contrast_level", "对比度"],
    ["saturation_level", "饱和度"],
    ["lighting_layers", "光线层次"],
    ["narrative_composition", "叙事构图"],
    ["atmosphere", "氛围感"],
  ];
  function normalizeStyleAnalysisPayload(data = {}) {
    const source = object(data);
    const contentType = normalizeChineseContentType(source.content_type || source.genre || source.final_genre);
    const components = {};
    const issues = [];
    for (const [key, label] of STYLE_COMPONENT_FIELDS_V78) {
      const part = v7827GlobalizeStyleComponent(key, source[key]);
      components[key] = part;
      if (!part) issues.push(`${label}缺失`);
      else if (hasLatinText(part)) issues.push(`${label}含英文`);
      else if (hasStylePollution(part)) issues.push(`${label}混入负面提示、限制项或人物要求`);
      else if (part.length > 70) issues.push(`${label}过长`);
    }
    if (!contentType) issues.push("内容类型缺失");
    else {
      if (!contentType.endsWith("短剧")) issues.push("内容类型必须以“短剧”结尾");
      if (hasLatinText(contentType)) issues.push("内容类型含英文");
      if (!/^[\u3400-\u9fff]{2,14}短剧$/.test(contentType)) issues.push("内容类型必须是简洁中文类型名，不得堆叠符号或长句");
      if (contentType.length > 16) issues.push("内容类型过长，疑似堆叠多个标签");
    }
    const suspensePrimary = source.suspense_is_primary === true || String(source.suspense_is_primary).toLowerCase() === "true";
    if (contentType.includes("悬疑") && !suspensePrimary) issues.push("悬疑不是主要叙事发动机，却写入了内容类型");
    const trailerStyle = STYLE_COMPONENT_FIELDS_V78.map(([key]) => components[key]).filter(Boolean).join("，");
    if (trailerStyle && (hasLatinText(trailerStyle) || hasStylePollution(trailerStyle))) issues.push("统一风格存在英文或跨字段污染");
    return {
      ok: !unique(issues).length,
      issues: unique(issues),
      content_type: contentType,
      trailer_style: trailerStyle,
      components,
      story_era: cleanChineseStylePart(source.story_era),
      world_setting: cleanChineseStylePart(source.world_setting),
      core_relationship: cleanChineseStylePart(source.core_relationship),
      core_conflict: cleanChineseStylePart(source.core_conflict),
      narrative_engine: cleanChineseStylePart(source.narrative_engine),
      emotional_tone: cleanChineseStylePart(source.emotional_tone),
      audience_expectation: cleanChineseStylePart(source.audience_expectation),
      suspense_is_primary: suspensePrimary,
      raw: source,
    };
  }
  function buildStyleOnlyPrompt({ repair = false, previous = null, issues = [] } = {}) {
    const advice = cleanStyleAnalysisAdvice(valueOf("#globalAnalysisAdvice") || "");
    const lines = [
      "【任务】只分析当前完整原文的内容类型、时代、世界观和统一视觉风格。禁止处理人物卡、人物外形、人物关系、负面提示词、画面限制、画质约束和最终输出首行提示。",
      (globalThis.__v23GetInstruction?.("global") ? `【Hotfix23全局AI语义指令｜协议锁定】\n${globalThis.__v23GetInstruction("global")}` : ""),
      (globalThis.__v23GetInstruction?.("analysis") ? `【Hotfix23内容类型与统一风格自定义指令｜只改语义】\n${globalThis.__v23GetInstruction("analysis")}` : ""),
      "【分析顺序】先判断时代与世界观，再判断核心关系、核心冲突、主要叙事发动机、情绪基调和观看期待；最后据此生成一个简洁内容类型，并生成十一项统一风格组件。不得先猜标签再反推原文。",
      "【内容类型规则】content_type必须是一个简洁中文“XXXX短剧”，只保留一个主要类型轴和必要的情绪/关系修饰，不得堆叠三个以上标签，不得包含人物名、地点名、场景名、英文或标点。只有调查、取证、追踪线索和推理解谜是主要叙事发动机时，suspense_is_primary才返回true，content_type才允许包含“悬疑”。",
      "【统一风格规则】十一项必须全部使用中文正向视觉描述；不得混入不要、禁止、避免、拒绝项、人物外形要求、画面限制、画质约束、提示词解释或剧情内容；不得返回英文摄影术语，全部改写为中文。每项只写一个清晰短语。镜头项只写全片镜头美学倾向，不锁死具体逐镜焦段；光线项只写动机光/高光/暗部原则，不锁死具体场景灯位或色温；叙事构图项只写全片人物关系与空间组织原则，不要求每镜同一种构图。",
      "【固定JSON字段】只返回：{content_type,story_era,world_setting,core_relationship,core_conflict,narrative_engine,emotional_tone,audience_expectation,suspense_is_primary,cinematic_quality,capture_texture,grain_texture,filter_tone,lens_language,optical_texture,contrast_level,saturation_level,lighting_layers,narrative_composition,atmosphere}。不得返回negative_prompt、picture_limit_prompt、quality_constraint_prompt、remark_prompt、trailer_style、characters或其他字段。",
      "【十一项含义】cinematic_quality=什么影视/电影质感；capture_texture=什么实拍质感；grain_texture=什么颗粒纹理；filter_tone=什么滤镜；lens_language=全片镜头性格与透视原则，不写具体逐镜焦段；optical_texture=什么噪点/畸变/光晕；contrast_level=什么对比度；saturation_level=什么饱和度；lighting_layers=全片动机光、高光与暗部层次原则，不写具体场景灯位；narrative_composition=全片人物关系、空间纵深与视线组织原则，不锁定逐镜构图；atmosphere=什么整体氛围感。",
      "【语言硬约束】所有字符串字段必须为简体中文，不得出现任何英文字母。",
      "【用户风格判断建议】", advice || "无；完全依据当前原文判断。",
      "【当前界面内容类型｜当前有效值】", text(valueOf("#genre")) || "空",
      "【当前界面统一风格｜当前有效值】", text(valueOf("#trailerStyle")) || "空",
      "【当前值原则】当前界面已经存在的内容就是本次分析的有效基础，不得回退到上一次AI返回或历史缓存；在原文证据支持时进行优化或重写。",
    ];
    if (repair) {
      lines.push("【上次结果不合格，必须从零重写】", JSON.stringify(previous || {}), "【不合格原因】", array(issues).join("；") || "协议不合格");
    }
    lines.push("【完整原文开始】", currentNovel(), "【完整原文结束】", "【返回】只返回合法JSON对象，不要Markdown、解释、候选或额外文字。");
    return lines.join("\n\n");
  }
  async function requestValidatedStyleAnalysis() {
    let data = await aiAnalyze(buildStyleOnlyPrompt(), STYLE_MARKER, 1);
    let normalized = normalizeStyleAnalysisPayload(data);
    if (!normalized.ok) {
      data = await aiAnalyze(buildStyleOnlyPrompt({repair:true, previous:data, issues:normalized.issues}), STYLE_MARKER, 0);
      normalized = normalizeStyleAnalysisPayload(data);
    }
    if (!normalized.ok) throw new Error(`内容类型与统一风格结果未通过中文纯净协议：${normalized.issues.join("；")}。本次未覆盖当前内容。`);
    return normalized;
  }
  function applyStyleAnalysisResult(result) {
    const locks = object(appState.styleLocks);
    const currentGenre = text(valueOf("#genre"));
    const currentStyle = text(valueOf("#trailerStyle"));
    if (locks.genre) {
      if (typeof updateStyleSuggestion === "function") updateStyleSuggestion("genre", result.content_type, currentGenre);
    } else {
      setValue("#genre", result.content_type);
      if (typeof clearStyleSuggestion === "function") clearStyleSuggestion("genre");
    }
    if (locks.trailer_style) {
      if (typeof updateStyleSuggestion === "function") updateStyleSuggestion("trailer_style", result.trailer_style, currentStyle);
    } else {
      setValue("#trailerStyle", result.trailer_style);
      if (typeof clearStyleSuggestion === "function") clearStyleSuggestion("trailer_style");
    }
    try { if (typeof renderStyleLockState === "function") renderStyleLockState(); } catch (_) {}
    return {
      final_genre: text(valueOf("#genre")),
      final_trailer_style: text(valueOf("#trailerStyle")),
      genre_locked: Boolean(locks.genre),
      trailer_style_locked: Boolean(locks.trailer_style),
    };
  }

  async function runStyle(button=q("#analyzeBtn"), options={}) {
    try { globalThis.__commitLatestEditableUiState?.("before_style_analysis"); } catch (_) {}
    if(!currentNovel())return apiError("请先粘贴整段原文。");if(typeof setButtonBusy==="function")setButtonBusy(button,true,"判断中");
    try{
      const c=core(),requestSnapshot=captureSourceSnapshot("style_request");
      if(c.source_hash&&c.source_hash!==requestSnapshot.source_hash){c.people=[];c.slots=[];c.relationships=[];c.aliases=[];c.mention_entities=[];c.scene_casting={};c.relationship_pending=[];c.temporary_continuity_cache={};c.temporary_continuity_cache_character_core_validator={};c.temporary_continuity_cache_v78_primary={};c.temporary_continuity_authority_v78={};c.relationship_conflicts=[];resetAnalysisState("整段原文已变化，旧人物链路已隔离");}
      c.source_hash=requestSnapshot.source_hash;c.analysis_state.source_hash=requestSnapshot.source_hash;c.source_stale=false;
      const result=await requestValidatedStyleAnalysis();
      const responseSnapshot=captureSourceSnapshot("style_response");
      if(responseSnapshot.source_hash!==requestSnapshot.source_hash)throw new Error(`内容类型与统一风格返回期间原文已变化：请求 revision ${requestSnapshot.source_revision}，当前 revision ${responseSnapshot.source_revision}。本次结果未写入。`);
      const applied=applyStyleAnalysisResult(result);
      appState.storyEraLock=result.story_era;
      appState.characterAppearanceContext={
        content_type:applied.final_genre,
        trailer_style:applied.final_trailer_style,
        story_era:result.story_era,
        world_setting:result.world_setting,
        core_relationship:result.core_relationship,
        core_conflict:result.core_conflict,
        narrative_engine:result.narrative_engine,
        emotional_tone:result.emotional_tone,
        audience_expectation:result.audience_expectation,
        suspense_is_primary:result.suspense_is_primary,
        style_components:deepClone(result.components),
        protocol:"v77_style_chinese_clean_v78",
      };
      appState.analysisComplete=false;
      appState.analysisSourceText=currentNovel();
      appState.analysisProtocolVersion=PROTOCOL;
      setAnalysisStage("style",Boolean(applied.final_genre&&applied.final_trailer_style),"内容类型与统一风格中文纯净分析完成");
      appState.characterAnalysisResultStatus="style_ready_waiting_character_core";
      const styleTraceId=text(result?._ai_meta?.request_id||result?.__v77_trace_id);
      if(styleTraceId)emitCharacterCoreTrace({trace_stage:"style_writeback",request_id:styleTraceId,outcome:"accepted",summary:"内容类型与统一风格已写入当前工作区",final_snapshot:{genre:applied.final_genre,trailer_style:applied.final_trailer_style,story_era:result.story_era,world_setting:result.world_setting}}).catch(()=>{});
      const lockNote=[applied.genre_locked?"内容类型已锁定并保留当前值":"",applied.trailer_style_locked?"统一风格已锁定并保留当前值":""].filter(Boolean).join("；");
      if(!options.silentStatus)showAIStatusNotice(`V77已完成中文内容类型与统一风格判断；负面提示词、画面限制、画质约束、最终输出首行提示均未读取、未提交、未修改。${lockNote?`${lockNote}。`:""}下一步请点击“按强制名单生成人物卡”。`,"ready",10000);
      scheduleDraftSave?.();
      return {...result,...applied};
    }catch(e){setAnalysisStage("style",false,e?.message||String(e));apiError(e?.message||String(e));return null;}finally{if(typeof setButtonBusy==="function")setButtonBusy(button,false);}
  }

  function v36StageSignature(slot={}) {
    const stage=normalizeStage(visibleStage(slot),slot.gender);
    const timeline=text(slot?.age?.timeline_stage||slot?.current_values?.timeline_stage||"当前时间线");
    return `${stage||"待AI判断"}@@${timeline}`;
  }
  function v36ReferenceAssetState(slot={}) {
    const facts=characterFactsCompleteness(slot);
    if(!facts.complete)return {state:"pending_facts",asset:null,summary:"参考图：等待AI确认性别与剧情年龄阶段"};
    const refs=object(appState.referenceAssets),assets=array(refs.characters),personId=text(slot.person_id),slotId=text(slot.slot_id),sig=v36StageSignature(slot);
    const samePerson=assets.filter(a=>text(a.person_id)===personId);
    const exact=assets.find(a=>text(a.target_slot_id)===slotId&&(!text(a.stage_signature)||text(a.stage_signature)===sig))
      || assets.find(a=>text(a.person_id)===personId&&text(a.stage_signature)===sig);
    const mismatched=samePerson.filter(a=>Boolean(a.has_main_image)&&text(a.stage_signature)&&text(a.stage_signature)!==sig);
    if(exact?.has_main_image)return {state:"ready",asset:exact,summary:`参考图：已匹配当前阶段（${stageLabel(sig)}）`};
    if(mismatched.length)return {state:"stale",asset:null,summary:`参考图：存在其他阶段主图，当前阶段需单独绑定`};
    const legacy=samePerson.find(a=>Boolean(a.has_main_image)&&!text(a.stage_signature));
    if(legacy)return {state:"legacy",asset:legacy,summary:"参考图：旧资产缺少阶段标记，建议重新保存当前阶段主图"};
    return {state:"missing",asset:null,summary:"参考图：未绑定当前阶段主图"};
  }
  function stageLabel(signature="") {return text(signature).split("@@")[0]||"当前阶段";}
  function v36ScopesOverlap(a={},b={}) {
    const a1=Number.isFinite(Number(a.valid_from_line))?Number(a.valid_from_line):-Infinity,a2=Number.isFinite(Number(a.valid_to_line))?Number(a.valid_to_line):Infinity;
    const b1=Number.isFinite(Number(b.valid_from_line))?Number(b.valid_from_line):-Infinity,b2=Number.isFinite(Number(b.valid_to_line))?Number(b.valid_to_line):Infinity;
    return Math.max(a1,b1)<=Math.min(a2,b2);
  }
  function v36RelationshipConflicts(c=core()) {
    const conflicts=[];
    const rels=array(c.relationships).filter(r=>!r.disabled);
    const byPair=new Map();
    rels.forEach(r=>{const key=[text(r.source_person_id),text(r.target_person_id),text(r.timeline||"current"),text(r.worldline||"main")].join("|");const list=byPair.get(key)||[];list.push(r);byPair.set(key,list);});
    byPair.forEach((rows,key)=>{
      const labels=unique(rows.map(r=>text(r.display_label||r.relation_type||"相关")));
      if(labels.length>1){
        const manual=rows.filter(r=>r.manual_locked||text(r.origin)==="manual"),ai=rows.filter(r=>!(r.manual_locked||text(r.origin)==="manual"));
        conflicts.push({type:manual.length&&ai.length?"manual_vs_ai_relation":"relation_disagreement",key,labels,relation_ids:rows.map(r=>text(r.relation_id)),evidence:unique(rows.map(r=>text(r.evidence))),summary:`${personLabel(rows[0].source_person_id)} → ${personLabel(rows[0].target_person_id)} 同一作用域存在不同关系：${labels.join(" / ")}`});
      }
    });
    const aliases=array(c.aliases).filter(a=>!a.disabled&&text(a.alias));
    for(let i=0;i<aliases.length;i++)for(let j=i+1;j<aliases.length;j++){
      const a=aliases[i],b=aliases[j];if(text(a.alias)!==text(b.alias)||text(a.target_person_id)===text(b.target_person_id)||!v36ScopesOverlap(a,b))continue;
      conflicts.push({type:(a.manual_locked||b.manual_locked)?"manual_vs_ai_alias":"alias_collision",alias:text(a.alias),targets:unique([personLabel(a.target_person_id),personLabel(b.target_person_id)]),binding_ids:unique([text(a.binding_id),text(b.binding_id)]),evidence:unique([text(a.evidence),text(b.evidence)]),summary:`称呼“${text(a.alias)}”在重叠作用域同时指向：${unique([personLabel(a.target_person_id),personLabel(b.target_person_id)]).join(" / ")}`});
    }
    c.relationship_conflicts=conflicts.slice(0,80);return c.relationship_conflicts;
  }
  function v36CharacterHealth(slot={}) {
    const c=core(),issues=[],warnings=[];
    if(isUnresolvedGender(slot.gender))issues.push("性别待确认");
    if(isUnresolvedStage(visibleStage(slot)))issues.push("年龄阶段待确认");
    if(!text(slot.appearance))issues.push("外形未生成");
    else if(slot.appearance_stage_stale)issues.push("阶段变化后外形待重构");
    const ref=v36ReferenceAssetState(slot);if(ref.state==="pending_facts")warnings.push("参考图等待人物事实完成");else if(ref.state==="stale")issues.push("参考图阶段不匹配");else if(ref.state==="legacy")warnings.push("参考图缺阶段标记");else if(ref.state==="missing")warnings.push("当前阶段无参考图");
    const conflicts=v36RelationshipConflicts(c).filter(x=>JSON.stringify(x).includes(text(slot.person_id))||JSON.stringify(x).includes(personLabel(slot.person_id)));if(conflicts.length)warnings.push(`关系/别称冲突 ${conflicts.length} 项`);
    const aliases=c.aliases.filter(a=>!a.disabled&&text(a.target_person_id)===text(slot.person_id));
    const rels=c.relationships.filter(r=>!r.disabled&&(text(r.source_person_id)===text(slot.person_id)||text(r.target_person_id)===text(slot.person_id)));
    const level=issues.length?"error":warnings.length?"warn":"ok";
    return {level,issues,warnings,reference:ref,relation_count:rels.length,alias_count:aliases.length,summary:issues.length?issues.join("；"):warnings.length?warnings.join("；"):"人物资料、阶段外形与当前引用状态正常"};
  }
  function renderCharacters() {
    const dm=globalThis.__V77_DIRTY_RENDER__,cc=core();
    if(dm&&!dm.shouldRender("core.characters",{revision:cc.character_revision,slots:cc.slots.map(slot=>({slot_id:slot.slot_id,person_id:slot.person_id,display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),appearance:slot.appearance,appearance_status:slot.appearance_status,current_values:slot.current_values,manual_values:slot.manual_values,status:slot.status,revisions:slot.revisions})),assets:appState.referenceAssets?.characters,mode:appState.outputMode}))return;
    const container=q("#characters"),template=q("#characterTemplateV2")||q("#characterTemplate");if(!container||!template)return;
    container.innerHTML="";const c=core();if(!c.slots.length){container.innerHTML='<div class="empty-state compact">尚未生成CharacterCore 2.0人物卡。</div>';return;}
    const frag=document.createDocumentFragment();
    c.slots.forEach((slot,index)=>{
      const node=template.content.firstElementChild.cloneNode(true);node.dataset.slotId=slot.slot_id;
      const title=node.querySelector(".character-title");if(title)title.textContent=`人物 ${index+1} · ${slotDisplay(slot)}`;
      const actions=node.querySelector(".character-actions");
      const addAction=(label,field,handler)=>{let b=document.createElement("button");b.type="button";b.className="btn secondary small-btn";b.dataset.field=field;b.textContent=label;b.addEventListener("click",handler);actions?.appendChild(b);return b;};
      const status=document.createElement("p");status.className="hint character-core-v2-status full";const revisionState=slot.appearance_revision_status==="needs_review"?"；诊断：建议复核":slot.appearance_revision_status==="passed"?"；诊断：已通过":"";status.textContent=`性别：${slot.gender}；可视阶段：${visibleStage(slot)||"待AI判断"}；实际年龄：${slot.age?.chronological_age||"未明确"}；外形：${slot.appearance?slot.appearance_stage_stale?"阶段已变更，待重构":"已写入":"待AI生成"}${revisionState}`;actions?.before(status);
      const health=v36CharacterHealth(slot),healthNode=document.createElement("p");healthNode.className=`character-health-badge full ${health.level}`;healthNode.textContent=`${health.level==="ok"?"✓ 人物资料正常":health.level==="warn"?"⚠ 人物资料需复核":"⚠ 人物资料需处理"}｜关系 ${health.relation_count}｜别称 ${health.alias_count}${health.issues.length?`｜${health.issues[0]}`:health.warnings.length?`｜${health.warnings[0]}`:""}`;healthNode.title=`${health.reference.summary}${health.issues.length?`\n问题：${health.issues.join("；")}`:""}${health.warnings.length?`\n提醒：${health.warnings.join("；")}`:""}`;actions?.before(healthNode);
      const name=node.querySelector('[data-field="name"]');if(name){name.value=slotDisplay(slot);name.addEventListener("change",()=>{slot.display_name=text(name.value)||slot.display_name;slot.manual_values={...object(slot.manual_values),display_name:slot.display_name};slot.character_revision++;syncLegacyState();renderCharacters();scheduleDraftSave?.();});}
      const gender=node.querySelector('[data-field="gender"]');if(gender){if(![...gender.options].some(o=>o.value==="无性别")){const option=document.createElement("option");option.value="无性别";option.textContent="无性别/不适用";gender.appendChild(option);}gender.value=slot.gender==="待确认"?"未定":slot.gender;gender.addEventListener("change",()=>{slot.gender=normalizeGender(gender.value);const manual={...object(slot.manual_values)};if(isUnresolvedGender(slot.gender))delete manual.gender;else manual.gender=slot.gender;slot.manual_values=manual;slot.current_values={...object(slot.current_values),gender:slot.gender};const g=genderFromStage(visibleStage(slot));if(g!=="待确认"&&g!==slot.gender){const family=/婴儿/.test(visibleStage(slot))?"婴儿":/小/.test(visibleStage(slot))?"儿童":/少女|少年|青少年/.test(visibleStage(slot))?"青少年":/成年|年轻|青年/.test(visibleStage(slot))?"成年":/中年/.test(visibleStage(slot))?"中年":"老年";slot.age.visual_age_stage=normalizeStage(family,slot.gender);}slot.appearance_stage_stale=Boolean(slot.appearance);slot.character_revision++;syncLegacyState();renderCharacters();scheduleDraftSave?.();});}
      const stage=node.querySelector('[data-field="ageStages"]');if(stage){stage.innerHTML="";["",...STAGES].forEach(v=>{const o=document.createElement("option");o.value=v;o.textContent=v||"待AI判断";stage.appendChild(o);});stage.value=visibleStage(slot);stage.addEventListener("change",()=>{const old=visibleStage(slot);slot.age.visual_age_stage=normalizeStage(stage.value,slot.gender);slot.age.stage_source=slot.age.visual_age_stage?"user_pending":"user_cleared";slot.age.stage_locked=false;slot.age.stage_lock_state=slot.age.visual_age_stage?"manual_editing":"unlocked";{const manual={...object(slot.manual_values)};if(isUnresolvedStage(slot.age.visual_age_stage))delete manual.visual_age_stage;else manual.visual_age_stage=slot.age.visual_age_stage;slot.manual_values=manual;}slot.current_values={...object(slot.current_values),visual_age_stage:slot.age.visual_age_stage};const g=genderFromStage(slot.age.visual_age_stage);if(g!=="待确认"){slot.gender=g;slot.manual_values.gender=g;slot.current_values.gender=g;}if(old!==slot.age.visual_age_stage&&slot.appearance)slot.appearance_stage_stale=true;slot.character_revision++;syncLegacyState();renderCharacters();refreshSceneBindings();scheduleDraftSave?.();});}
      const note=node.querySelector('[data-field="appearanceNote"]');if(note){note.maxLength=MAX_NOTE;note.value=slot.appearance_revision_note||"";note.placeholder="写完整修改目标或方向要求。AI会结合当前年龄阶段重写整段外形，不使用固定人物模板。";note.addEventListener("input",()=>{slot.appearance_revision_note=String(note.value||"").slice(0,MAX_NOTE);slot.manual_values={...object(slot.manual_values),appearance_revision_note:slot.appearance_revision_note};slot.character_revision++;scheduleDraftSave?.();});}
      const appearance=node.querySelector('[data-field="appearance"]');if(appearance){appearance.maxLength=MAX_APPEARANCE;appearance.value=slot.appearance||"";appearance.placeholder="AI根据当前小说、身份、种族、世界观和年龄阶段自由设计完整外形。";appearance.addEventListener("input",()=>{slot.appearance=String(appearance.value||"").slice(0,MAX_APPEARANCE);slot.manual_values={...object(slot.manual_values),appearance:slot.appearance};slot.appearance_status=slot.appearance?"ready":"missing";slot.character_revision++;syncLegacyState();scheduleDraftSave?.();});}
      const fields=node.querySelector(".character-fields");
      if(fields){
        const extras=document.createElement("div");extras.className="character-core-extra-fields full";extras.innerHTML=`<label>实际年龄/年岁<input data-core="chronologicalAge" maxlength="80" placeholder="如：17岁、千岁、无生理年龄"></label><label>物种/种族<input data-core="species" maxlength="80" placeholder="人类、狐妖、仿生人等"></label><label>物种生命阶段<input data-core="lifeStage" maxlength="100" placeholder="如：幼体、成年体、完全体"></label><label>剧情时间线<input data-core="timelineStage" maxlength="100" placeholder="当前、重生前、前世等"></label><label>外形详细程度<select data-core="detailLevel">${DETAIL_LEVELS.map(v=>`<option value="${v}">${v}</option>`).join("")}</select></label>`;fields.appendChild(extras);
        const bind=(sel,key,get,set)=>{const el=extras.querySelector(sel);if(!el)return;el.value=get();el.addEventListener("change",()=>{const before=text(get());const next=text(el.value);set(next);slot.manual_values={...object(slot.manual_values),[key]:next};if(before!==next&&slot.appearance)slot.appearance_stage_stale=true;slot.character_revision++;syncLegacyState();renderCharacters();scheduleDraftSave?.();});};
        bind('[data-core="chronologicalAge"]',"chronological_age",()=>slot.age?.chronological_age||"",v=>slot.age.chronological_age=v);
        bind('[data-core="species"]',"species",()=>slot.species||"",v=>slot.species=v);
        bind('[data-core="lifeStage"]',"life_stage",()=>slot.age?.life_stage||"",v=>slot.age.life_stage=v);
        bind('[data-core="timelineStage"]',"timeline_stage",()=>slot.age?.timeline_stage||"当前时间线",v=>slot.age.timeline_stage=v||"当前时间线");
        const detail=extras.querySelector('[data-core="detailLevel"]');if(detail){detail.value=slot.appearance_detail_level||"详细";detail.addEventListener("change",()=>{slot.appearance_detail_level=detail.value;scheduleDraftSave?.();});}
      }
      const optimize=node.querySelector('[data-field="optimizeCharacter"]');if(optimize){const fresh=optimize.cloneNode(true);optimize.replaceWith(fresh);fresh.textContent="重写外形";fresh.addEventListener("click",()=>runCharacters({onlySlotId:slot.slot_id,button:fresh}));}
      if(slot.appearance_stage_stale)addAction("按当前阶段重构外形","rebuildStage",(event)=>runCharacters({onlySlotId:slot.slot_id,forceStage:true,button:event.currentTarget}));
      addAction("基于当前值AI重写","restoreAuto",async(event)=>{try{globalThis.__commitLatestEditableUiState?.("before_current_value_ai_rewrite");}catch(_){}await runCharacters({onlySlotId:slot.slot_id,forceStage:Boolean(slot.appearance_stage_stale),button:event.currentTarget});});
      const deleteButton=node.querySelector(".delete-character");if(deleteButton){deleteButton.title="从强制名单删除该人物";deleteButton.addEventListener("click",()=>{const guide=q("#characterGuideInput");guide?.focus();showAIStatusNotice(`人物卡数量由强制名单唯一决定。请从强制名单中删除“${slot.source_entry||slot.display_name}”，再重新生成。`,"warning",9000);});}
      frag.appendChild(node);
    });
    container.appendChild(frag);
  }

  function personLabel(personId) { const p=core().people.find(x=>x.person_id===personId);return p?.canonical_name||p?.primary_display_name||personId; }
  function personOptions(selected="") { return core().people.map(p=>`<option value="${p.person_id}" ${p.person_id===selected?"selected":""}>${personLabel(p.person_id)}</option>`).join(""); }
  function renderRelations() {
    const dm=globalThis.__V77_DIRTY_RENDER__,cc=core();
    if(dm&&!dm.shouldRender("core.relations",{relationships:cc.relationships,aliases:cc.aliases,conflicts:cc.relationship_conflicts,revision:cc.character_revision}))return;
    const c=cc,status=q("#relationshipGraphStatus"),relBox=q("#relationshipGraphList"),aliasBox=q("#relationshipAliasList"),pending=q("#relationshipPendingList");
    const v36Conflicts=v36RelationshipConflicts(c);
    if(status){const st=analysisState();const phase=st.relationships_ready?"AI关系已完成":st.relationship_fallback_available?"AI关系未完成（当前使用人物节点/关键词保底）":"AI关系待分析";status.textContent=`${phase}｜基础人物 ${c.people.length} 个｜关系 ${c.relationships.filter(r=>!r.disabled).length} 条｜称呼映射 ${c.aliases.filter(a=>!a.disabled).length} 条｜待确认 ${c.relationship_pending.length} 项｜冲突 ${v36Conflicts.length} 项`;
      let conflictBox=document.querySelector("#characterCoreRelationshipConflicts");if(!conflictBox){conflictBox=document.createElement("div");conflictBox.id="characterCoreRelationshipConflicts";status.insertAdjacentElement("afterend",conflictBox);}conflictBox.className=`character-relationship-conflicts ${v36Conflicts.length?"has-conflict":"clear"}`;conflictBox.innerHTML=v36Conflicts.length?`<strong>关系/别称冲突诊断（不会自动覆盖人工值）</strong>${v36Conflicts.slice(0,12).map(x=>`<div class="relationship-conflict-row"><span>${text(x.summary)}</span>${array(x.evidence).filter(Boolean).length?`<small>证据：${array(x.evidence).filter(Boolean).join("；")}</small>`:""}</div>`).join("")}`:`<span>未检测到关系或称呼冲突。</span>`;}
    if(relBox){relBox.innerHTML="";if(!c.relationships.length)relBox.innerHTML='<div class="empty-state compact">暂无人物关系。</div>';c.relationships.filter(r=>!r.disabled).forEach(r=>{const row=document.createElement("article");row.className="relationship-row";row.innerHTML=`<select data-r="source">${personOptions(r.source_person_id)}</select><input data-r="label" value="${r.display_label||"相关"}" placeholder="正向关系"><select data-r="target">${personOptions(r.target_person_id)}</select><input data-r="reverse" value="${r.reverse_label||""}" placeholder="反向关系"><input data-r="timeline" value="${r.timeline||"current"}" placeholder="时间线"><input data-r="worldline" value="${r.worldline||"main"}" placeholder="世界线"><input data-r="evidence" value="${r.evidence||""}" placeholder="证据"><button class="btn secondary small-btn" data-r="save">保存</button><button class="btn danger small-btn" data-r="delete">删除</button>`;row.querySelectorAll('input,select').forEach((control)=>{const mark=()=>{row.dataset.liveModified="true";};control.addEventListener("input",mark);control.addEventListener("change",mark);});row.querySelector('[data-r="save"]').onclick=()=>{r.source_person_id=row.querySelector('[data-r="source"]').value;r.target_person_id=row.querySelector('[data-r="target"]').value;r.display_label=text(row.querySelector('[data-r="label"]').value)||"相关";r.reverse_label=text(row.querySelector('[data-r="reverse"]').value);r.timeline=text(row.querySelector('[data-r="timeline"]').value)||"current";r.worldline=text(row.querySelector('[data-r="worldline"]').value)||"main";r.evidence=text(row.querySelector('[data-r="evidence"]').value);r.manual_locked=true;r.origin="manual";renderRelations();refreshSceneBindings();scheduleDraftSave?.();};row.querySelector('[data-r="delete"]').onclick=()=>{r.disabled=true;r.manual_locked=true;renderRelations();refreshSceneBindings();scheduleDraftSave?.();};relBox.appendChild(row);});}
    if(aliasBox){aliasBox.innerHTML="";if(!c.aliases.length)aliasBox.innerHTML='<div class="empty-state compact">暂无称呼映射。</div>';c.aliases.filter(a=>!a.disabled).forEach(a=>{const row=document.createElement("article");row.className="relationship-row alias-row";row.innerHTML=`<input data-a="alias" value="${a.alias||""}" placeholder="称呼/别称"><select data-a="target">${personOptions(a.target_person_id)}</select><input data-a="type" value="${a.alias_type||"alias"}" placeholder="称呼类型"><input data-a="scope" value="${a.valid_from_line??""}-${a.valid_to_line??""}" placeholder="有效行范围，如1-20"><input data-a="evidence" value="${a.evidence||""}" placeholder="证据"><button class="btn secondary small-btn" data-a="save">保存</button><button class="btn danger small-btn" data-a="delete">删除</button>`;row.querySelectorAll('input,select').forEach((control)=>{const mark=()=>{row.dataset.liveModified="true";};control.addEventListener("input",mark);control.addEventListener("change",mark);});row.querySelector('[data-a="save"]').onclick=()=>{a.alias=text(row.querySelector('[data-a="alias"]').value);a.target_person_id=row.querySelector('[data-a="target"]').value;a.alias_type=text(row.querySelector('[data-a="type"]').value)||"alias";const scope=text(row.querySelector('[data-a="scope"]').value);const match=scope.match(/^(\d*)\s*[-~至]\s*(\d*)$/);a.valid_from_line=match&&match[1]?Number(match[1]):null;a.valid_to_line=match&&match[2]?Number(match[2]):null;a.evidence=text(row.querySelector('[data-a="evidence"]').value);a.manual_locked=true;a.origin="manual";renderRelations();refreshSceneBindings();scheduleDraftSave?.();};row.querySelector('[data-a="delete"]').onclick=()=>{a.disabled=true;a.manual_locked=true;renderRelations();refreshSceneBindings();scheduleDraftSave?.();};aliasBox.appendChild(row);});}
    if(pending){pending.innerHTML=c.relationship_pending.length?c.relationship_pending.map(x=>`<div class="pending-relation">${text(x.text||x.evidence||JSON.stringify(x))}</div>`).join(""):'<div class="empty-state compact">当前没有待确认项。</div>';}
  }
  function addRelation() { const c=core();if(c.people.length<2)return apiError("至少需要两个人物才能添加关系。");c.relationships.push({relation_id:`manual_rel_${Date.now()}`,source_person_id:c.people[0].person_id,target_person_id:c.people[1].person_id,relation_type:"related",display_label:"相关",timeline:"current",worldline:"main",confidence:1,origin:"manual",manual_locked:true,disabled:false});renderRelations();scheduleDraftSave?.(); }
  function addAlias() { const c=core();if(!c.people.length)return apiError("当前没有人物节点。");c.aliases.push({binding_id:`manual_alias_${Date.now()}`,alias:"新称呼",target_person_id:c.people[0].person_id,alias_type:"manual",confidence:1,origin:"manual",manual_locked:true,disabled:false});renderRelations();scheduleDraftSave?.(); }

  function ensureFallbackPeopleGraph() {
    const c = core();
    const slotRows = c.slots.filter(slot => !slot.disabled);
    if (!slotRows.length) return 0;
    const existing = new Map(array(c.people).map(person => [text(person.person_id), person]));
    const merged = new Map();
    slotRows.forEach((slot, index) => {
      const fallbackId = text(slot.person_id || slot.slot_id || `person_fallback_${index+1}`);
      slot.person_id = fallbackId;
      const current = merged.get(fallbackId) || existing.get(fallbackId) || {
        person_id: fallbackId,
        canonical_name: text(slot.canonical_name_hint || slot.base_name || slot.display_name || slot.source_entry || `人物${index+1}`),
        primary_display_name: text(slot.display_name || slot.base_name || slot.source_entry || `人物${index+1}`),
        aliases: [],
        origin: existing.has(fallbackId) ? text(existing.get(fallbackId).origin || 'merged') : 'forced_roster_fallback',
      };
      current.canonical_name = text(current.canonical_name || slot.canonical_name_hint || slot.base_name || slot.display_name || slot.source_entry || `人物${index+1}`);
      current.primary_display_name = text(current.primary_display_name || slot.display_name || slot.base_name || slot.source_entry || current.canonical_name || `人物${index+1}`);
      current.aliases = unique([
        ...array(current.aliases),
        text(slot.display_name), text(slot.base_name), text(slot.canonical_name_hint), text(slot.source_entry),
        ...array(slot.aliases), ...array(slot.bracket_hints), ...array(slot.identity_hints),
      ]);
      merged.set(fallbackId, current);
    });
    c.people = [...merged.values()];
    if (c.people.length) {
      const st = analysisState();
      if (!st.facts_ready) { st.facts_fallback_available=true; st.last_message='AI人物事实未完整返回；已按强制名单建立人物节点，但事实阶段仍标记为未完成。'; }
      if (!st.relationships_ready && !c.relationships.length && !c.aliases.length) {
        st.relationship_fallback_available=true;st.relationship_graph_ready=false;st.last_message='AI关系图未完整返回；强制名单人物节点仅用于继续生成，本地关键词仍可选角，但关系阶段不会伪装为已完成。';
      }
      st.last_updated_at=new Date().toISOString();
    }
    return c.people.length;
  }

  function explicitRosterTokens(value = "") {
    const raw = String(value || "");
    if (!raw) return [];
    const stageNoise = /^(?:女婴儿|男婴儿|小女孩|小男孩|少女|少年|年轻女性|年轻男性|中年女性|中年男性|老年女性|老年男性|婴幼儿阶段|儿童阶段|青少年阶段|成年阶段|中年阶段|老年阶段|待AI判断|待确认)$/;
    return unique(raw.replace(/[〔【\[][^〕】\]]*[〕】\]]/g, " ").split(/[（）()、，,；;｜|/\\\s]+/).map(text).filter((token) => token && !stageNoise.test(token) && (token.length >= 2 || token === "我")));
  }

  const KEYWORD_CONFIDENCE_THRESHOLD = 0.8;
  function keywordPriority(entry={}) {
    const type=text(entry.type);
    if(entry.manual_locked||type==="manual_alias"||type==="manual_keyword")return 100;
    if(["canonical","display","slot_name","slot_base","canonical_hint","guide_alias"].includes(type))return 90;
    if(type==="ai_relation_alias"||type==="relationship_title")return 80;
    return 70;
  }
  function keywordEntriesForPerson(person,c=core()) {
    const entries=[];
    const add=(keyword,type="alias",meta={})=>{
      const safe=text(keyword);if(!safe)return;
      const confidence=Number(meta.confidence ?? 1);
      if(meta.enabled===false||confidence<KEYWORD_CONFIDENCE_THRESHOLD)return;
      entries.push({keyword:safe,normalized_keyword:safe,keyword_id:text(meta.keyword_id)||`kw_${fingerprint(`${person.person_id}|${safe}|${type}|${text(meta.slot_id)}`)}`,person_id:person.person_id,type,confidence,enabled:true,...meta});
    };
    add(person.canonical_name,"canonical",{source:"forced_roster_or_ai",confidence:1});
    add(person.primary_display_name,"display",{source:"forced_roster",confidence:1});
    array(person.aliases).forEach(v=>add(v,"ai_person_alias",{source:"ai_relationship_graph",confidence:1}));
    c.slots.filter(s=>s.person_id===person.person_id&&!s.disabled).forEach(slot=>{
      add(slot.display_name,"slot_name",{slot_id:slot.slot_id,source:"forced_roster",confidence:1});
      add(slot.base_name,"slot_base",{slot_id:slot.slot_id,source:"forced_roster",confidence:1});
      add(slot.canonical_name_hint,"canonical_hint",{slot_id:slot.slot_id,source:"forced_roster",confidence:1});
      array(slot.aliases).forEach(v=>add(v,"guide_alias",{slot_id:slot.slot_id,source:"forced_roster",confidence:1}));
      [slot.source_entry,slot.display_name,slot.base_name,slot.canonical_name_hint,...array(slot.bracket_hints),...array(slot.identity_hints)].forEach(value=>{
        explicitRosterTokens(value).forEach(token=>add(token,"forced_roster_token",{slot_id:slot.slot_id,source:"forced_roster_explicit_text",confidence:1}));
      });
    });
    c.aliases.filter(a=>a.target_person_id===person.person_id&&!a.disabled&&a.keyword_enabled!==false).forEach(a=>add(a.alias,a.manual_locked?"manual_alias":(a.relation_id?"ai_relation_alias":"ai_alias"),{
      keyword_id:a.binding_id,binding_id:a.binding_id,slot_id:text(a.target_slot_id),valid_from_line:a.valid_from_line,valid_to_line:a.valid_to_line,
      valid_scope:text(a.valid_scope),timeline_scope:text(a.timeline_scope),relation_id:text(a.relation_id),evidence:text(a.evidence),confidence:a.manual_locked?1:Number(a.confidence??.75),manual_locked:Boolean(a.manual_locked),source:a.manual_locked?"manual":"ai_relationship_graph",
    }));
    const seen=new Map();
    entries.sort((a,b)=>b.keyword.length-a.keyword.length||keywordPriority(b)-keywordPriority(a)).forEach(entry=>{
      const key=`${entry.keyword}|${entry.person_id}|${text(entry.slot_id)}`;
      if(!seen.has(key)||keywordPriority(entry)>keywordPriority(seen.get(key)))seen.set(key,entry);
    });
    return [...seen.values()];
  }
  function rebuildKeywordIndex() {
    const c=core(),legacyIndex={};
    if(!c.people.length&&c.slots.length)ensureFallbackPeopleGraph();
    c.people.forEach(person=>{legacyIndex[person.person_id]={person_id:person.person_id,canonical_name:text(person.canonical_name||person.primary_display_name),entries:keywordEntriesForPerson(person,c)};});
    // Phase6 keeps the original CharacterCore index as an independent live validator.
    c.keyword_index_character_core_validator=deepClone(legacyIndex);
    let authority=null;
    try{authority=globalThis.__V78_CASTING_SERVICE__?.resolveKeywordProduction?.(deepClone(legacyIndex))||null;}catch(error){console.warn('[V78_KEYWORD_AUTHORITY] 主索引决策异常，自动回退 CharacterCore',error);}
    c.keyword_index=authority?.promoted?deepClone(authority.index):legacyIndex;
    c.keyword_executor_v78=authority?.executor||'character_core_keyword_fallback';
    c.keyword_authority_v78=authority?{promoted:Boolean(authority.promoted),safe_streak:Number(authority.safe_streak||0),threshold:Number(authority.threshold||3),executor:c.keyword_executor_v78,decided_at:authority.decided_at||new Date().toISOString(),mismatch_count:Number(authority.parity?.mismatch_count||0)}:{promoted:false,safe_streak:0,threshold:3,executor:c.keyword_executor_v78,decided_at:new Date().toISOString(),mismatch_count:0};
    const total=Object.values(c.keyword_index).reduce((n,item)=>n+array(item.entries).length,0);
    const usedFallback = Boolean(c.slots.length && c.people.length && !c.relationships.length && !c.aliases.length);
    const owner=c.keyword_authority_v78.promoted?'V78主索引':'CharacterCore校验/回退索引';
    setAnalysisStage("casting",Boolean(total),usedFallback?`AI关系图缺失时，已按强制名单保底建立 ${total} 个本地关键词映射；当前 ${owner}`:`已根据强制名单名称与AI关系图建立 ${total} 个本地关键词映射；当前 ${owner}`);
    syncLegacyState();return c.keyword_index;
  }
  function detectSceneRole(source,hits=[]) {
    const textSource=text(source),safeHits=array(hits).map(item=>text(item.keyword||item)).filter(Boolean);if(!safeHits.length)return "unknown";
    const clauses=textSource.split(/[，,。！？!?；;\n]+/).map(text).filter(Boolean),roles=[];
    const classify=(clause,hit)=>{const near=regexEscape(hit);const medium=(before,after,beforeSpan=20,afterSpan=8)=>new RegExp(`(?:${before}).{0,${beforeSpan}}(?:${near})|(?:${near}).{0,${afterSpan}}(?:${after})`).test(clause);
      if(new RegExp(`(?:提到|说起|谈及|听说|得知|资料(?:中|里)?写着|名单(?:上|里)?有).{0,16}(?:${near})|(?:${near}).{0,4}(?:只是被提及|并未在场)`).test(clause))return "mentioned_only";
      if(medium("照片|相片|遗照|合影|画像","照片|相片|遗照|合影|画像"))return "photo_visible";
      if(medium("监控|屏幕|视频|手机画面|电脑画面|直播","监控|屏幕|视频|手机画面|电脑画面|直播"))return "screen_visible";
      if(medium("梦里|梦中|梦境|噩梦","梦里|梦中|梦境|噩梦"))return "dream_visible";
      if(medium("回忆|记忆中|想起|往事|前世","回忆|记忆|往事|前世"))return "memory_visible";
      if(medium("电话里|手机里|耳机里|广播中|录音中|画外音|门外传来","的声音|画外音|在电话里说|从门外传来",16,4))return "offscreen_voice";
      if(new RegExp(`(?:对|朝|向).{0,8}(?:${near}).{0,4}(?:说|问|喊|解释|回答)`).test(clause))return "addressee_visible";
      if(new RegExp(`(?:${near})[：:]|(?:${near}).{0,8}(?:说|问|喊|低声|开口|回答)`).test(clause))return "speaker_visible";
      return "visible";};
    safeHits.forEach(hit=>clauses.filter(clause=>clause.includes(hit)).forEach(clause=>roles.push(classify(clause,hit))));
    const rank={unknown:0,mentioned_only:1,offscreen_voice:2,photo_visible:3,screen_visible:3,memory_visible:3,dream_visible:3,group_visible:4,visible:5,addressee_visible:6,speaker_visible:7};return roles.sort((a,b)=>(rank[b]||0)-(rank[a]||0))[0]||"unknown";
  }
  function aliasScopeActive(binding,lineNumber=null){if(binding.disabled||binding.enabled===false)return false;if(!Number.isFinite(Number(lineNumber)))return true;const n=Number(lineNumber),start=binding.valid_from_line,end=binding.valid_to_line;return !(start!==null&&start!==undefined&&n<Number(start))&&!(end!==null&&end!==undefined&&n>Number(end));}
  function addCastEvidence(candidate,evidence,role="") {candidate.evidence.push(evidence);if(role)candidate.roles.push(role);}
  function directCastCandidates(source,lineNumber=null) {
    const c=core(),textSource=text(source),candidates=new Map();if(!Object.keys(c.keyword_index||{}).length)rebuildKeywordIndex();
    const all=[];
    for(const person of c.people){for(const entry of array(c.keyword_index?.[person.person_id]?.entries)){if(!aliasScopeActive(entry,lineNumber)||Number(entry.confidence??1)<KEYWORD_CONFIDENCE_THRESHOLD)continue;let cursor=0;while(entry.keyword&&(cursor=textSource.indexOf(entry.keyword,cursor))>=0){all.push({...entry,start:cursor,end:cursor+entry.keyword.length,person});cursor+=Math.max(1,entry.keyword.length);}}}
    all.sort((a,b)=>a.start-b.start||(b.end-b.start)-(a.end-a.start)||keywordPriority(b)-keywordPriority(a));
    const chosen=[],ambiguous=[];
    for(let i=0;i<all.length;){const same=all.filter(hit=>hit.start===all[i].start&&hit.end===all[i].end&&hit.keyword===all[i].keyword);const bestPriority=Math.max(...same.map(keywordPriority));const best=same.filter(hit=>keywordPriority(hit)===bestPriority);const personTargets=unique(best.map(hit=>hit.person_id)),slotTargets=unique(best.map(hit=>text(hit.slot_id)).filter(Boolean));if(personTargets.length>1){ambiguous.push({keyword:all[i].keyword,start:all[i].start,end:all[i].end,candidates:best.map(hit=>({person_id:hit.person_id,slot_id:text(hit.slot_id),keyword_id:hit.keyword_id,confidence:hit.confidence}))});i+=same.length;continue;}let hit=best.slice().sort((a,b)=>Number(Boolean(b.slot_id))-Number(Boolean(a.slot_id)))[0];if(personTargets.length===1&&slotTargets.length>1)hit={...hit,slot_id:"",reason:"同一人物多阶段关键词，交由当前剧情阶段选择slot_id"};if(!chosen.some(prev=>Math.max(prev.start,hit.start)<Math.min(prev.end,hit.end)))chosen.push(hit);i+=same.length;}
    for(const hit of chosen){const candidate=candidates.get(hit.person_id)||{person:hit.person,hits:[],evidence:[],roles:[],ambiguous:[]};candidate.hits.push(hit);const evidence={type:"keyword",keyword:hit.keyword,keyword_id:hit.keyword_id,keyword_type:hit.type,start:hit.start,end:hit.end,person_id:hit.person_id,slot_id:text(hit.slot_id),binding_id:text(hit.binding_id),relation_id:text(hit.relation_id),source:text(hit.source),confidence:Number(hit.confidence??1),reason:`本地关键词命中:${hit.keyword}`};addCastEvidence(candidate,evidence,detectSceneRole(textSource,[hit.keyword]));candidates.set(hit.person_id,candidate);}
    candidates.ambiguous=ambiguous;return candidates;
  }
  function chooseStageSlot(personId,source,previousStageByPerson=new Map(),info={}) {
    const c=core(),slots=c.slots.filter(s=>s.person_id===personId&&!s.disabled);if(!slots.length)return null;
    const hitSlotIds=unique(array(info.hits).map(hit=>text(hit.slot_id)).filter(Boolean));
    if(hitSlotIds.length===1){const bound=slots.find(s=>s.slot_id===hitSlotIds[0]);if(bound)return bound;}
    const explicit=slots.filter(s=>visibleStage(s)&&text(source).includes(visibleStage(s)));if(explicit.length===1)return explicit[0];
    const previous=slots.find(s=>s.slot_id===previousStageByPerson.get(personId));if(previous)return previous;
    const current=slots.filter(s=>/当前|主线|present|current/i.test(text(s.age?.timeline_stage)||text(s.inferred_values?.timeline_stage))||s.inferred_values?.is_current_timeline===true);if(current.length===1)return current[0];
    return slots.length===1?slots[0]:slots.slice().sort((a,b)=>Number(a.return_index||0)-Number(b.return_index||0))[0];
  }
  function localCast(source,lineNumber=null,previousStageByPerson=new Map()) {
    const c=core(),textSource=text(source),candidates=directCastCandidates(textSource,lineNumber),selected=[];
    candidates.forEach((info,personId)=>{const slot=chooseStageSlot(personId,textSource,previousStageByPerson,info);if(!slot)return;const roles=unique(info.roles),role=detectSceneRole(textSource,info.hits);const resolvedRole=role!=="unknown"?role:(roles[0]||"visible");const keywordEvidence=info.evidence.filter(e=>e.type==="keyword"),relationshipEvidence=keywordEvidence.filter(e=>text(e.relation_id));selected.push({slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:personId,display_name:slot.display_name,scene_role:resolvedRole,evidence:deepClone(info.evidence),matched_keywords:keywordEvidence.map(e=>e.keyword),keyword_evidence:keywordEvidence,relationship_evidence:relationshipEvidence,confidence:Math.max(...keywordEvidence.map(e=>Number(e.confidence||0)),.8),stage_resolution_required:c.slots.filter(s=>s.person_id===personId&&!s.disabled).length>1&&!keywordEvidence.some(e=>text(e.slot_id))&&!text(source).includes(visibleStage(slot))});previousStageByPerson.set(personId,slot.slot_id);});
    selected.ambiguous=array(candidates.ambiguous);return selected.sort((a,b)=>Number(b.confidence||0)-Number(a.confidence||0));
  }

  function castRoleRank(role="") {
    return ({unknown:0,mentioned_only:1,offscreen_voice:2,photo_visible:3,screen_visible:3,memory_visible:3,dream_visible:3,group_visible:4,visible_candidate:5,visible:6,addressee_visible:7,speaker_visible:8})[text(role)]||0;
  }
  function mergeCastRows(...groups) {
    const bySlot=new Map();
    groups.flatMap(group=>array(group)).forEach(raw=>{
      if(!raw||!text(raw.slot_id))return;
      const existing=bySlot.get(raw.slot_id);
      if(!existing){bySlot.set(raw.slot_id,deepClone(raw));return;}
      const preferred=castRoleRank(raw.scene_role)>castRoleRank(existing.scene_role)?raw:existing;
      bySlot.set(raw.slot_id,{
        ...existing,...preferred,
        evidence:[...array(existing.evidence),...array(raw.evidence)],
        matched_keywords:unique([...array(existing.matched_keywords),...array(raw.matched_keywords)]),
        keyword_evidence:[...array(existing.keyword_evidence),...array(raw.keyword_evidence)],
        relationship_evidence:[...array(existing.relationship_evidence),...array(raw.relationship_evidence)],
        confidence:Math.max(Number(existing.confidence||0),Number(raw.confidence||0)),
      });
    });
    return [...bySlot.values()];
  }
  function slotKeywordFallbackCast(source,lineNumber=null,previousStageByPerson=new Map()) {
    const c=core(),sourceText=text(source),rows=[];
    if(!sourceText)return rows;
    if(!Object.keys(c.keyword_index||{}).length)rebuildKeywordIndex();
    const personHits=new Map();
    for(const slot of c.slots.filter(item=>!item.disabled)){
      const entries=[
        ...array(c.keyword_index?.[slot.person_id]?.entries).filter(entry=>!text(entry.slot_id)||text(entry.slot_id)===slot.slot_id),
        ...unique([slot.display_name,slot.base_name,slot.canonical_name_hint,slot.source_entry,...array(slot.aliases),...array(slot.bracket_hints),...array(slot.identity_hints)]).flatMap(value=>explicitRosterTokens(value).map(keyword=>({keyword,slot_id:slot.slot_id,person_id:slot.person_id,type:'slot_direct_fallback',confidence:1,source:'forced_roster'}))),
      ];
      const matches=entries.filter(entry=>aliasScopeActive(entry,lineNumber)&&Number(entry.confidence??1)>=KEYWORD_CONFIDENCE_THRESHOLD&&text(entry.keyword)&&sourceText.includes(text(entry.keyword)));
      if(!matches.length)continue;
      const current=personHits.get(slot.person_id)||{hits:[],slots:[]};
      current.hits.push(...matches);current.slots.push(slot);personHits.set(slot.person_id,current);
    }
    personHits.forEach((info,personId)=>{
      const slot=chooseStageSlot(personId,sourceText,previousStageByPerson,{hits:info.hits});if(!slot)return;
      const sorted=info.hits.slice().sort((a,b)=>text(b.keyword).length-text(a.keyword).length||keywordPriority(b)-keywordPriority(a));
      const best=sorted[0];const keyword=text(best?.keyword);if(!keyword)return;
      const relationEvidence=sorted.filter(item=>text(item.relation_id)).map(item=>({type:'relationship_keyword',keyword:text(item.keyword),relation_id:text(item.relation_id),binding_id:text(item.binding_id),person_id:personId,slot_id:slot.slot_id,confidence:Number(item.confidence??1),reason:`AI关系图确认称谓后由本地关键词命中:${text(item.keyword)}`}));
      const evidence={type:'pre_ai_local_keyword',keyword,keyword_id:text(best.keyword_id),keyword_type:text(best.type),person_id:personId,slot_id:slot.slot_id,binding_id:text(best.binding_id),relation_id:text(best.relation_id),source:text(best.source),confidence:Number(best.confidence??1),reason:`AI请求前本地关键词预选角:${keyword}`};
      rows.push({slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:personId,display_name:slot.display_name,scene_role:'visible',evidence:[evidence,...relationEvidence],matched_keywords:unique(sorted.map(item=>text(item.keyword))),keyword_evidence:[evidence],relationship_evidence:relationEvidence,confidence:1,selection_source:'pre_ai_local_keyword'});
      previousStageByPerson.set(personId,slot.slot_id);
    });
    return rows;
  }
  function preselectFormalCast(source,lineNumber=null,previousStageByPerson=new Map()) {
    const semantic=localCast(source,lineNumber,previousStageByPerson);
    const deterministic=slotKeywordFallbackCast(source,lineNumber,previousStageByPerson);
    const merged=mergeCastRows(semantic,deterministic);
    merged.ambiguous=unique([...array(semantic.ambiguous),...array(deterministic.ambiguous)]);
    return merged;
  }
  function expandCastItem(item) {const c=core(),s=c.slots.find(z=>z.slot_id===item.slot_id);if(!s)return null;return{slot_id:s.slot_id,slot_token:s.slot_token,person_id:s.person_id,display_name:s.display_name,gender:s.gender,visual_age_stage:visibleStage(s),chronological_age:s.age?.chronological_age||"",life_stage:s.age?.life_stage||"",timeline_stage:s.age?.timeline_stage||"",species:s.species||"",appearance:s.appearance||"",identity_role:text(s.inferred_values?.identity_role),social_position:text(s.inferred_values?.social_position),scene_role:item.scene_role,matched_keywords:array(item.matched_keywords),keyword_evidence:array(item.keyword_evidence),relationship_evidence:array(item.relationship_evidence),evidence:array(item.evidence),confidence:Number(item.confidence||0)};}
  function v36TemporaryContinuityProfile(entity={}) {
    const identity=text(entity.identity_hint||entity.label||entity.source_phrase||"临时人物");
    const anchors=unique([text(entity.age_stage_hint),text(entity.gender_hint),text(entity.species_hint),identity]).filter(Boolean).join("、");
    return `${anchors||"临时人物"}；同一continuity_id连续镜头保持同一体型、发型、服装主色与主要配饰，除非原文明示换装或时间线变化`;
  }
  function v36HydrateTemporaryContinuity(entity={},lineNumber=null) {
    const c=core(),cache=c.temporary_continuity_cache||(c.temporary_continuity_cache={}),id=text(entity.continuity_id||entity.entity_id||`temp_${fingerprint(entity.label||entity.source_phrase)}`),n=Number.isFinite(Number(lineNumber))?Number(lineNumber):null,sceneKey=text(entity.continuity_scene_key),old=object(cache[id]);
    const gap=n!==null&&Number.isFinite(Number(old.last_line))?Math.abs(n-Number(old.last_line)):0;
    const sceneChanged=Boolean(sceneKey&&text(old.scene_key)&&sceneKey!==text(old.scene_key));
    const expired=Boolean(old.continuity_id&&(sceneChanged||gap>5));
    const base=expired?{}:old;
    const profile=text(entity.continuity_profile)||text(base.continuity_profile)||v36TemporaryContinuityProfile(entity);
    const wardrobe=text(entity.wardrobe_state)||text(base.wardrobe_state)||"沿用首次出镜确定的同一套简易服装、主色和主要配饰；原文明示换装时才更新";
    const next={continuity_id:id,scene_key:sceneKey||text(base.scene_key),first_line:base.first_line??n,last_line:n??base.last_line??null,appearance_count:Number(base.appearance_count||0)+1,continuity_profile:profile,wardrobe_state:wardrobe,gender_hint:text(entity.gender_hint||base.gender_hint),age_stage_hint:text(entity.age_stage_hint||base.age_stage_hint),species_hint:text(entity.species_hint||base.species_hint),identity_hint:text(entity.identity_hint||base.identity_hint||entity.label),updated_at:new Date().toISOString()};
    cache[id]=next;
    // Keep the cache bounded; the continuity is intentionally short-term.
    const keys=Object.keys(cache);if(keys.length>80)keys.sort((a,b)=>text(cache[a]?.updated_at).localeCompare(text(cache[b]?.updated_at))).slice(0,keys.length-80).forEach(k=>delete cache[k]);
    return {...entity,continuity_id:id,continuity_profile:profile,wardrobe_state:wardrobe,continuity_scope:"scene_short_term",continuity_ttl_shots:5,continuity_first_line:next.first_line,continuity_last_line:next.last_line};
  }
  function normalizeTemporaryCharacter(item = {}) {
    const source = object(item);
    const label = text(source.label || source.source_phrase || source.identity_hint || source.entity_type);
    if (!label) return null;
    const participation=text(source.participation_state || "visible_candidate");
    const continuitySceneKey=text(source.continuity_scene_key || source.scene_anchor_id || "");
    return {
      entity_id:text(source.entity_id || source.mention_id || `temp_${fingerprint(label)}`),
      label,
      source_phrase:text(source.source_phrase || label),
      aliases:unique([label,...array(source.aliases)]),
      entity_type:text(source.entity_type || "temporary_character"),
      participation_state:participation,
      valid_from_line:Number.isFinite(Number(source.valid_from_line))?Number(source.valid_from_line):null,
      valid_to_line:Number.isFinite(Number(source.valid_to_line))?Number(source.valid_to_line):null,
      count:Math.max(1,Number(source.count||1)||1),
      gender_hint:text(source.gender_hint),age_stage_hint:text(source.age_stage_hint),species_hint:text(source.species_hint),identity_hint:text(source.identity_hint),
      evidence:text(source.evidence),confidence:Number(source.confidence??.8),
      must_render:source.must_render===true||['visible','visual_exposition','visible_candidate','group_visible','photo_visible','screen_visible','screen_or_memory_visible','memory_visible','dream_visible','background_visible'].includes(participation),
      continuity_scope:text(source.continuity_scope || "scene_short_term"),
      continuity_scene_key:continuitySceneKey,
      continuity_id:text(source.continuity_id || source.entity_id || `temp_${fingerprint(label)}`),
      continuity_profile:text(source.continuity_profile),
      wardrobe_state:text(source.wardrobe_state),
      continuity_ttl_shots:Number(source.continuity_ttl_shots||5),
      continuity_first_line:source.continuity_first_line??null,
      continuity_last_line:source.continuity_last_line??null,
      rendered:source.rendered===true,
    };
  }

  function temporaryEntityScopeActive(entity,lineNumber=null){
    if(!Number.isFinite(Number(lineNumber)))return true;const n=Number(lineNumber);
    return !(entity.valid_from_line!==null&&n<entity.valid_from_line)&&!(entity.valid_to_line!==null&&n>entity.valid_to_line);
  }
  function v783NormalizeRelationAlias(value=""){
    return text(value).replace(/^(?:我的?|他的?|她的?|其|咱们的?|我们(?:家)?的?)/,"").replace(/^(?:两个|两位|两名|一位|一个|那位|这位|那名|这名)/,"").trim();
  }
  function v783FormalAliasCollision(c,raw,hit="",entity={}){
    const candidates=unique([hit,entity.label,entity.source_phrase,...array(entity.aliases)]).map(v783NormalizeRelationAlias).filter(Boolean);
    return c.slots.some(slot=>{
      if(slot.person_id===text(raw.target_person_id))return true;
      const aliases=unique([slot.display_name,slot.base_name,slot.canonical_name_hint,...array(slot.aliases)]).map(v783NormalizeRelationAlias).filter(Boolean);
      return candidates.some(candidate=>aliases.some(alias=>candidate===alias||(/^(?:我|我的|他|他的|她|她的)/.test(hit)&&candidate.endsWith(alias))));
    });
  }
  function temporaryEntitiesForSource(source="",lineNumber=null){
    const c=core(),sourceText=text(source),result=[];
    for(const raw of array(c.mention_entities)){
      const entity=normalizeTemporaryCharacter(raw);if(!entity||entity.confidence<.55||!temporaryEntityScopeActive(entity,lineNumber))continue;
      const tokens=unique([entity.label,entity.source_phrase,...array(entity.aliases)]).sort((a,b)=>b.length-a.length);
      const hit=tokens.find(token=>token&&sourceText.includes(token));if(!hit)continue;
      if(['mentioned_only','offscreen_voice'].includes(entity.participation_state))continue;
      // V78.2.6: resolve relation-prefixed phrases back to existing formal slots first.
      // Example: “我未婚妻” must map to formal “未婚妻” when that slot exists, never duplicate as a temporary person.
      if(v783FormalAliasCollision(c,raw,hit,entity))continue;
      entity.matched_keyword=hit;
      entity.must_render=['visible','visual_exposition','visible_candidate','group_visible','photo_visible','screen_visible','screen_or_memory_visible','memory_visible','dream_visible','background_visible'].includes(entity.participation_state);
      result.push(v36HydrateTemporaryContinuity(entity,lineNumber));
      if(result.length>=8)break;
    }
    return result;
  }
  function mergeTemporaryCharacters(...groups){
    const merged=[];
    groups.flatMap(group=>array(group)).forEach(raw=>{
      const item=normalizeTemporaryCharacter(raw);if(!item)return;
      const key=text(item.continuity_id)||text(item.entity_id)||text(item.source_phrase)||text(item.label);
      const at=merged.findIndex(existing=>(text(existing.continuity_id)||text(existing.entity_id)||text(existing.source_phrase)||text(existing.label))===key||text(existing.source_phrase)===text(item.source_phrase)||text(existing.label)===text(item.label));
      if(at<0)merged.push(item);else merged[at]={...merged[at],...item,aliases:unique([...array(merged[at].aliases),...array(item.aliases)])};
    });
    return merged.slice(0,8);
  }

  function temporaryEntitiesForContextReadOnly(source='',lineNumber=null){
    const c=core(),svc=globalThis.__V78_CASTING_SERVICE__,authority=object(c.temporary_continuity_authority_v78||{});
    if(authority.promoted&&svc?.computeTemporary){
      try{return array(svc.computeTemporary(text(source),lineNumber,{continuityCache:deepClone(c.temporary_continuity_cache_v78_primary||c.temporary_continuity_cache||{})})?.temporary_characters);}
      catch(error){console.warn('[V78_TEMPORARY_CONTEXT] 主状态读取失败，回退 CharacterCore validator',error);}
    }
    const saved=c.temporary_continuity_cache;
    try{c.temporary_continuity_cache=deepClone(c.temporary_continuity_cache_character_core_validator||saved||{});return temporaryEntitiesForSource(text(source),lineNumber);}
    finally{c.temporary_continuity_cache=saved;}
  }
  function fallbackFormalCastFromPlan(plan={},previousStageByPerson=new Map()){
    const c=core(),result=[];
    for(const identifier of unique(array(plan.allowed))){
      const safe=text(identifier);if(!safe)continue;
      const legacy=array(appState.characters).find(card=>[card.name,card.display_name,card.slot_id,card.id].some(value=>text(value)===safe));
      let slot=legacy?c.slots.find(item=>item.slot_id===text(legacy.slot_id||legacy.id)):null;
      if(!slot){slot=c.slots.find(item=>[item.display_name,item.base_name,item.canonical_name_hint,item.source_entry,item.slot_id,item.slot_token,...array(item.aliases)].some(value=>text(value)===safe||(text(value)&&safe.includes(text(value)))));}
      if(!slot)continue;
      previousStageByPerson.set(slot.person_id,slot.slot_id);
      result.push({slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:slot.person_id,display_name:slot.display_name,scene_role:"visible",evidence:[{type:"source_semantic_fallback",keyword:safe,reason:`原分镜语义层已选中:${safe}`}],matched_keywords:[safe],keyword_evidence:[{type:"source_semantic_fallback",keyword:safe,person_id:slot.person_id,slot_id:slot.slot_id,confidence:1,reason:`原分镜语义层已选中:${safe}`}],relationship_evidence:[],confidence:1});
    }
    return result;
  }

  function characterCardKeywordFallbackCast(source,lineNumber=null,previousStageByPerson=new Map()) {
    const c=core(),sourceText=text(source),tokenOwners=new Map(),personMatches=new Map(),rows=[];
    if(!sourceText)return rows;
    for(const slot of c.slots.filter(item=>!item.disabled)){
      const rawValues=unique([slot.display_name,slot.base_name,slot.canonical_name_hint,slot.source_entry,...array(slot.aliases),...array(slot.bracket_hints),...array(slot.identity_hints)]);
      const tokens=unique(rawValues.flatMap(value=>explicitRosterTokens(value))).filter(token=>token&&sourceText.includes(token));
      for(const keyword of tokens){
        if(!tokenOwners.has(keyword))tokenOwners.set(keyword,new Set());
        tokenOwners.get(keyword).add(slot.person_id);
        const list=personMatches.get(slot.person_id)||[];
        list.push({keyword,slot_id:slot.slot_id,person_id:slot.person_id});personMatches.set(slot.person_id,list);
      }
    }
    personMatches.forEach((matches,personId)=>{
      const valid=matches.filter(item=>tokenOwners.get(item.keyword)?.size===1).sort((a,b)=>text(b.keyword).length-text(a.keyword).length);
      if(!valid.length)return;
      const slot=chooseStageSlot(personId,sourceText,previousStageByPerson,{hits:valid});if(!slot)return;
      const best=valid[0],evidence={type:'character_card_keyword_fallback',keyword:text(best.keyword),person_id:personId,slot_id:slot.slot_id,source:'character_card',confidence:1,reason:`人物卡关键词保底命中:${text(best.keyword)}`};
      rows.push({slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:personId,display_name:slot.display_name,scene_role:'visible',evidence:[evidence],matched_keywords:unique(valid.map(item=>text(item.keyword))),keyword_evidence:[evidence],relationship_evidence:[],confidence:1,selection_source:'character_card_keyword_fallback'});
      previousStageByPerson.set(personId,slot.slot_id);
    });
    return rows;
  }

  function sceneSourceKey(source="",lineNumber=1) {
    const canonical=normalizeNovelSource(source);
    const index=Math.max(1,Number(lineNumber||1));
    return `${index}_${fingerprint(canonical)}`;
  }
  function ensureOutlineSourceSceneShells() {
    if(typeof splitOutlineSourceScenes!=="function")return array(appState.scenes);
    const units=array(splitOutlineSourceScenes(currentNovel()));
    if(!units.length)return array(appState.scenes);
    const existingByKey=new Map(array(appState.scenes).map((scene,index)=>[text(scene.source_key)||sceneSourceKey(scene.source_text||scene.text||"",Number(scene.source_index||index+1)),scene]));
    appState.scenes=units.map((unit,index)=>{
      const source=text(unit.source_text||unit.text||"");
      const key=sceneSourceKey(source,index+1);
      const existing=existingByKey.get(key);
      if(existing){existing.source_key=key;existing.source_index=index+1;existing.source_text=source;existing.text=source;return existing;}
      return typeof makeScene==="function"?makeScene(source,{...unit,id:`scene_source_${index+1}_${fingerprint(source)}`,source_index:index+1,source_key:key}):{id:`scene_source_${index+1}_${fingerprint(source)}`,source_index:index+1,source_key:key,source_text:source,text:source,characters:[],character_slot_ids:[],relation_selected_slot_ids:[],keyword_selected_slot_ids:[],manual_added_slot_ids:[],manual_excluded_slot_ids:[]};
    });
    return appState.scenes;
  }
  function restoreSceneBindingsBySourceKey(options={}) {
    const c=core(),records=array(c.outline_cast_snapshot?.scenes);
    const recordByKey=new Map(records.map(record=>[text(record.source_key)||sceneSourceKey(record.source_text,record.source_index),record]));
    array(appState.scenes).forEach((scene,index)=>{
      const key=text(scene.source_key)||sceneSourceKey(scene.source_text||scene.text||"",Number(scene.source_index||index+1));
      const record=recordByKey.get(key)||c.scene_casting[key];if(!record)return;
      scene.source_key=key;
      scene.relation_selected_slot_ids=unique(array(record.relation_selected_slot_ids));
      scene.keyword_selected_slot_ids=unique(array(record.keyword_selected_slot_ids));
      scene.manual_added_slot_ids=unique(array(record.manual_added_slot_ids));
      scene.manual_excluded_slot_ids=unique(array(record.manual_excluded_slot_ids));
      scene.character_slot_ids=unique(array(record.selected_slot_ids));
      scene.character_core_cast=deepClone(array(record.selected));
      scene.characters=array(record.selected).map(item=>text(item.display_name)).filter(Boolean);
      scene.keyword_evidence=deepClone(array(record.keyword_evidence));
      scene.relationship_evidence=deepClone(array(record.relationship_evidence));
      scene.temporary_characters=deepClone(array(record.temporary_characters));
      scene.cast_executor_v78=text(record.cast_executor||scene.cast_executor_v78||'character_core_fallback');
      scene.cast_authority_v78=deepClone(object(record.cast_authority||scene.cast_authority_v78||{}));
      scene.cast_evidence_executor_v78=text(record.cast_evidence_executor||scene.cast_evidence_executor_v78||'character_core_evidence_fallback');
      scene.cast_evidence_authority_v78=deepClone(object(record.cast_evidence_authority||scene.cast_evidence_authority_v78||{}));
      scene.temporary_executor_v78=text(record.temporary_executor||scene.temporary_executor_v78||'character_core_temporary_fallback');
      scene.temporary_scene_parity_v78=deepClone(object(record.temporary_scene_parity||scene.temporary_scene_parity_v78||{}));
      scene.characters_mode=scene.manual_added_slot_ids.length||scene.manual_excluded_slot_ids.length?'manual':'auto';
      scene.characters_authoritative=true;scene.preselected_cast_frozen=false;
    });
    syncLegacyState();
    try{if(typeof syncAllSceneCharactersToOutlineShots==='function')syncAllSceneCharactersToOutlineShots();}catch(_){}
    if(options?.render!==false&&typeof renderScenes==='function')renderScenes();
    return deepClone(appState.scenes);
  }

  function refreshSceneBindings(options={}) {
    const c=core(),previousStageByPerson=new Map(),preRequest=Boolean(options?.preRequest),shouldRender=options?.render!==false;
    const castSnapshots=[];
    const activeContinuityAtStart=deepClone(c.temporary_continuity_cache||{});
    const validatorStored=object(c.temporary_continuity_cache_character_core_validator||{}),v78Stored=object(c.temporary_continuity_cache_v78_primary||{});
    let validatorContinuityCursor=deepClone(Object.keys(validatorStored).length?validatorStored:activeContinuityAtStart);
    let v78ContinuityCursor=deepClone(Object.keys(v78Stored).length?v78Stored:activeContinuityAtStart);
    array(appState.scenes).forEach((scene,index)=>{
      const source=text(scene.source_text||scene.text||scene.original_text),lineNumber=Number(scene.source_index||index+1);
      // CharacterCore remains the independent live validator even after V78 keyword promotion.
      const activeKeywordIndex=c.keyword_index,validatorKeywordIndex=object(c.keyword_index_character_core_validator||activeKeywordIndex);
      let relationshipRows=[],keywordRows=[];
      try{
        c.keyword_index=validatorKeywordIndex;
        relationshipRows=localCast(source,lineNumber,previousStageByPerson)
          .filter(item=>array(item.relationship_evidence).length||array(item.keyword_evidence).some(e=>text(e.relation_id)||text(e.binding_id)))
          .filter(item=>!['offscreen_voice','mentioned_only'].includes(item.scene_role));
        keywordRows=characterCardKeywordFallbackCast(source,lineNumber,previousStageByPerson)
          .filter(item=>!['offscreen_voice','mentioned_only'].includes(item.scene_role));
      }finally{c.keyword_index=activeKeywordIndex;}
      const relationIds=unique(relationshipRows.map(item=>item.slot_id));
      const keywordIds=unique(keywordRows.map(item=>item.slot_id));
      const automaticIds=unique([...relationIds,...keywordIds]);

      let manualAdded=unique(array(scene.manual_added_slot_ids));
      let manualExcluded=unique(array(scene.manual_excluded_slot_ids));
      if(!manualAdded.length&&!manualExcluded.length&&scene.character_core_manual_locked===true&&array(scene.character_slot_ids).length){
        const legacyIds=unique(array(scene.character_slot_ids));
        manualAdded=legacyIds.filter(id=>!automaticIds.includes(id));
        manualExcluded=automaticIds.filter(id=>!legacyIds.includes(id));
      }
      const excluded=new Set(manualExcluded);
      const effectiveIds=unique([...automaticIds,...manualAdded]).filter(id=>!excluded.has(id));
      const autoRows=mergeCastRows(relationshipRows,keywordRows);
      const rowBySlot=new Map(autoRows.map(item=>[item.slot_id,item]));
      manualAdded.forEach(id=>{
        const slot=c.slots.find(item=>item.slot_id===id);if(!slot)return;
        rowBySlot.set(id,{slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:slot.person_id,display_name:slot.display_name,scene_role:'visible',evidence:[{type:'manual_add',keyword:slot.display_name,reason:'用户人工增加本镜人物'}],matched_keywords:[],keyword_evidence:[],relationship_evidence:[],confidence:1,selection_source:'manual_add'});
      });
      const effectiveRows=effectiveIds.map(id=>rowBySlot.get(id)||(()=>{const slot=c.slots.find(item=>item.slot_id===id);return slot?{slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:slot.person_id,display_name:slot.display_name,scene_role:'visible',evidence:[{type:'manual',keyword:slot.display_name,reason:'当前本镜人物'}],matched_keywords:[],keyword_evidence:[],relationship_evidence:[],confidence:1,selection_source:'manual'}:null;})()).filter(Boolean);
      // Phase6: CharacterCore temporary continuity is calculated against its independent validator mirror.
      // The currently active V78 cache is never used as CharacterCore's own validator input.
      const savedActiveContinuity=c.temporary_continuity_cache;
      let temporary=[];
      try{
        c.temporary_continuity_cache=validatorContinuityCursor;
        temporary=mergeTemporaryCharacters(scene.temporary_characters,temporaryEntitiesForSource(source,lineNumber));
        validatorContinuityCursor=deepClone(c.temporary_continuity_cache||{});
      }finally{c.temporary_continuity_cache=savedActiveContinuity;}
      const sceneId=scene.id||`scene_${index+1}`,sourceKey=text(scene.source_key)||sceneSourceKey(source,lineNumber);
      const productionCandidate={scene_id:sceneId,source_key:sourceKey,source_index:lineNumber,source_text:source,selected:deepClone(effectiveRows),selected_slot_ids:deepClone(effectiveIds),manual_added_slot_ids:deepClone(manualAdded),manual_excluded_slot_ids:deepClone(manualExcluded),temporary_characters:deepClone(temporary)};
      let v78Gate=null;
      try{v78Gate=globalThis.__V78_CASTING_SERVICE__?.resolveProductionScene?.(scene,index,{productionRecord:productionCandidate,keywordIndex:globalThis.__V78_CASTING_SERVICE__?.buildKeywordIndex?.(),previousStageByPerson:Object.fromEntries(previousStageByPerson),continuityCache:deepClone(v78ContinuityCursor)})||null;if(v78Gate?.proposed_continuity_cache)v78ContinuityCursor=deepClone(v78Gate.proposed_continuity_cache);}catch(error){console.warn('[V78_CAST_GATE] 闸门计算异常，自动回退 CharacterCore',error);}
      const finalIds=v78Gate?.takeover?unique(array(v78Gate.record?.selected_slot_ids)):effectiveIds;
      let finalRows=finalIds.map(id=>rowBySlot.get(id)||effectiveRows.find(row=>row.slot_id===id)).filter(Boolean);
      // Evidence changes ownership only after its own 3-pass parity gate; identity/appearance rows remain CharacterCore-rich.
      if(v78Gate?.evidence_takeover){const shadowBySlot=new Map(array(v78Gate.record?.selected).map(row=>[text(row.slot_id),row]));finalRows=finalRows.map(row=>{const shadow=shadowBySlot.get(text(row.slot_id));return shadow?{...row,scene_role:text(shadow.scene_role||row.scene_role),matched_keywords:deepClone(array(shadow.matched_keywords)),keyword_evidence:deepClone(array(shadow.keyword_evidence)),relationship_evidence:deepClone(array(shadow.relationship_evidence)),evidence:deepClone(array(shadow.evidence)),selection_source:'v78_scene_cast_evidence_primary_phase6'}:row;});}
      const finalTemporary=v78Gate?.temporary_takeover?mergeTemporaryCharacters(v78Gate.record?.temporary_characters):temporary;
      finalRows.forEach(item=>previousStageByPerson.set(item.person_id,item.slot_id));

      scene.relation_selected_slot_ids=relationIds;
      scene.keyword_selected_slot_ids=keywordIds;
      scene.manual_added_slot_ids=manualAdded;
      scene.manual_excluded_slot_ids=manualExcluded;
      scene.relation_cast_rows=deepClone(relationshipRows);
      scene.keyword_cast_rows=deepClone(keywordRows);
      scene.character_core_cast=deepClone(finalRows);
      scene.character_slot_ids=finalIds;
      scene.characters=finalRows.map(item=>item.display_name);
      scene.character_casting_evidence=finalRows.flatMap(item=>array(item.evidence));
      scene.keyword_evidence=v78Gate?.evidence_takeover?finalRows.flatMap(item=>array(item.keyword_evidence)):keywordRows.flatMap(item=>array(item.keyword_evidence));
      scene.relationship_evidence=v78Gate?.evidence_takeover?finalRows.flatMap(item=>array(item.relationship_evidence)):relationshipRows.flatMap(item=>array(item.relationship_evidence));
      scene.temporary_characters=finalTemporary;
      scene.cast_evidence_executor_v78=v78Gate?.evidence_executor||'character_core_evidence_fallback';
      scene.cast_evidence_authority_v78=deepClone(v78Gate?.evidence_authority||{});
      scene.temporary_executor_v78=v78Gate?.temporary_executor||'character_core_temporary_warmup_validator';
      scene.temporary_scene_parity_v78=deepClone(v78Gate?.temporary_scene_parity||{});
      scene.characters_mode=manualAdded.length||manualExcluded.length?'manual':'auto';
      scene.characters_authoritative=true;
      scene.character_core_manual_locked=false;
      scene.preselected_cast_frozen=preRequest;
      scene.cast_snapshot_source=v78Gate?.takeover?'v78_casting_primary_phase6_promoted':'relationship_graph_plus_character_card_keyword_plus_manual_override';
      if(v78Gate?.evidence_takeover)scene.cast_snapshot_source+=' + v78_evidence_primary';
      if(v78Gate?.temporary_takeover)scene.cast_snapshot_source+=' + v78_temporary_primary';
      scene.cast_executor_v78=v78Gate?.executor||'character_core_fallback';
      scene.cast_authority_v78={promoted:Boolean(v78Gate?.promoted),safe_streak:Number(v78Gate?.safe_streak||0),threshold:Number(v78Gate?.threshold||3),takeover:Boolean(v78Gate?.takeover),executor:scene.cast_executor_v78,promoted_now:Boolean(v78Gate?.promoted_now),demoted_now:Boolean(v78Gate?.demoted_now)};
      scene.cast_gate_reasons_v78=deepClone(array(v78Gate?.reasons));
      scene.cast_gate_at_v78=v78Gate?.decided_at||new Date().toISOString();
      scene.cast_snapshot_at=new Date().toISOString();

      scene.source_key=sourceKey;
      const record={scene_id:sceneId,source_key:sourceKey,source_index:lineNumber,source_text:source,relation_selected_slot_ids:relationIds,keyword_selected_slot_ids:keywordIds,manual_added_slot_ids:manualAdded,manual_excluded_slot_ids:manualExcluded,selected:deepClone(finalRows),selected_slot_ids:finalIds,visible_slot_ids:finalIds,keyword_evidence:deepClone(scene.keyword_evidence),relationship_evidence:deepClone(scene.relationship_evidence),ambiguous_matches:[],mentioned_only:[],offscreen_voice:[],temporary_characters:deepClone(finalTemporary),manual_locked:Boolean(manualAdded.length||manualExcluded.length),ui_checkbox_synced:true,snapshot_frozen:preRequest,cast_executor:scene.cast_executor_v78,cast_authority:deepClone(scene.cast_authority_v78),cast_evidence_executor:scene.cast_evidence_executor_v78,cast_evidence_authority:deepClone(scene.cast_evidence_authority_v78),temporary_executor:scene.temporary_executor_v78,temporary_scene_parity:deepClone(scene.temporary_scene_parity_v78),cast_gate_reasons:deepClone(scene.cast_gate_reasons_v78),cast_gate_at:scene.cast_gate_at_v78};
      c.scene_casting[sourceKey]=record;castSnapshots.push(record);
      console.info('[RELATION_CAST_TRACE]',{scene_id:sceneId,source_index:lineNumber,selected_slot_ids:relationIds});
      console.info('[KEYWORD_CAST_TRACE]',{scene_id:sceneId,source_index:lineNumber,selected_slot_ids:keywordIds});
      console.info('[MANUAL_CAST_TRACE]',{scene_id:sceneId,manual_added_slot_ids:manualAdded,manual_excluded_slot_ids:manualExcluded,effective_slot_ids:finalIds});
      console.info('[CAST_UI_TRACE]',{scene_id:sceneId,ui_checkbox_synced:true,effective_slot_ids:finalIds,executor:scene.cast_executor_v78,gate_reasons:scene.cast_gate_reasons_v78});
    });
    // Phase6: keep two independent continuity mirrors and promote V78 only after full-cache parity for 3 consecutive refreshes.
    c.temporary_continuity_cache_character_core_validator=deepClone(validatorContinuityCursor);
    c.temporary_continuity_cache_v78_primary=deepClone(v78ContinuityCursor);
    let temporaryAuthority=null;
    try{temporaryAuthority=globalThis.__V78_CASTING_SERVICE__?.resolveTemporaryCacheProduction?.(deepClone(validatorContinuityCursor),deepClone(v78ContinuityCursor))||null;}catch(error){console.warn('[V78_TEMPORARY_AUTHORITY] 连续性主状态决策异常，自动回退 CharacterCore',error);}
    c.temporary_continuity_cache=temporaryAuthority?.promoted?deepClone(temporaryAuthority.cache):deepClone(validatorContinuityCursor);
    // On a cache-level mismatch, rebase the V78 candidate mirror from the independent CharacterCore validator so a later 3-pass recovery is possible.
    if(temporaryAuthority&&!temporaryAuthority.promoted&&Number(temporaryAuthority.parity?.mismatch_count||0)>0)c.temporary_continuity_cache_v78_primary=deepClone(validatorContinuityCursor);
    c.temporary_continuity_authority_v78=temporaryAuthority?{promoted:Boolean(temporaryAuthority.promoted),safe_streak:Number(temporaryAuthority.safe_streak||0),threshold:Number(temporaryAuthority.threshold||3),executor:temporaryAuthority.executor,decided_at:temporaryAuthority.decided_at||new Date().toISOString(),mismatch_count:Number(temporaryAuthority.parity?.mismatch_count||0)}:{promoted:false,safe_streak:0,threshold:3,executor:'character_core_temporary_fallback',decided_at:new Date().toISOString(),mismatch_count:0};
    c.outline_cast_snapshot={protocol:'v78_casting_primary_with_character_core_fallback_phase6',created_at:new Date().toISOString(),source_hash:c.source_hash,temporary_continuity_authority:deepClone(c.temporary_continuity_authority_v78),scenes:castSnapshots};
    setAnalysisStage('casting',Boolean(castSnapshots.length),'人物关系判断与人物卡关键词保底已分通道完成，并合并人工覆盖');
    syncLegacyState();
    try{if(typeof syncAllSceneCharactersToOutlineShots==='function')syncAllSceneCharactersToOutlineShots();}catch(_){}
    if(shouldRender&&typeof renderScenes==='function')renderScenes();
    try{globalThis.__V78_CASTING_SERVICE__?.observeProduction?.(deepClone(c.outline_cast_snapshot),{source_hash:c.source_hash,pre_request:preRequest});}catch(error){console.warn('[V78_CAST_PARITY] 本镜人物影子对比失败',error);}
    return deepClone(c.outline_cast_snapshot);
  }
  function buildSceneContext(source,manualIdentifiers=null,lineNumber=null,previousStageByPerson=new Map()) {
    const c=core();let selected;const manual=Array.isArray(manualIdentifiers);
    if(manual){const wanted=new Set(manualIdentifiers.map(text));selected=c.slots.filter(s=>[s.display_name,s.base_name,s.slot_id,s.slot_token].some(v=>wanted.has(text(v)))).map(s=>({slot_id:s.slot_id,slot_token:s.slot_token,person_id:s.person_id,display_name:s.display_name,scene_role:"visible",evidence:[{type:"manual",keyword:s.display_name,reason:"用户人工勾选人物卡"}],matched_keywords:[],keyword_evidence:[],relationship_evidence:[],confidence:1}));}
    else selected=preselectFormalCast(text(source),lineNumber,previousStageByPerson);
    const expanded=selected.map(expandCastItem).filter(Boolean),personIds=new Set(expanded.map(x=>x.person_id));
    const relationships=c.relationships.filter(r=>personIds.has(r.source_person_id)&&personIds.has(r.target_person_id)&&!r.disabled);
    const temporaryCharacters=temporaryEntitiesForContextReadOnly(text(source),lineNumber);
    return{casting_protocol:"character_core_single_executor_v77_hotfix19",manual_locked:manual,selected_characters:expanded,visible_characters:expanded.filter(x=>!['offscreen_voice','mentioned_only'].includes(x.scene_role)),offscreen_voice:expanded.filter(x=>x.scene_role==='offscreen_voice'),mentioned_only:expanded.filter(x=>x.scene_role==='mentioned_only'),temporary_characters:temporaryCharacters,keyword_evidence:expanded.flatMap(x=>x.keyword_evidence),relationship_evidence:expanded.flatMap(x=>x.relationship_evidence),relationship_context:relationships,continuity_context:{}};
  }
  function compactCharacterRegistryFromPlans(plans=[], appearanceLimit=1400) {
    const c=core(),ids=unique(array(plans).flatMap(plan=>array(plan.character_packages).map(item=>item.slot_id)));
    return ids.map(id=>{const slot=c.slots.find(item=>item.slot_id===id);return slot?{
      slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:slot.person_id,display_name:slot.display_name,
      gender:slot.gender,visual_age_stage:visibleStage(slot),chronological_age:slot.age?.chronological_age||"",
      species:slot.species||"",appearance:String(slot.appearance||"").slice(0,appearanceLimit),
    }:null;}).filter(Boolean);
  }
  function sceneCastingRowsForSubmission(plans=[], compact=false) {
    return array(plans).slice(0,150).map(plan=>({
      line_index:plan.line_index,
      source_key:text(plan.source_key)||sceneSourceKey(plan.source_text,plan.line_index),
      ...(compact?{}:{source_text:String(plan.source_text||"").slice(0,220)}),
      manual_locked:plan.characters_mode==='manual',
      selected:array(plan.character_packages).map(item=>({
        slot_id:item.slot_id,scene_role:item.scene_role,
        matched_keywords:array(item.matched_keywords).slice(0,compact?3:8),
        relationship_evidence:array(item.relationship_evidence).slice(0,compact?2:6).map(e=>compact?{relation_id:e.relation_id,keyword:e.keyword}:{relation_id:e.relation_id,keyword:e.keyword,reason:e.reason}),
      })),
      temporary_characters:array(plan.temporary_characters).map(item=>({label:text(item.label),source_phrase:text(item.source_phrase||item.label),entity_type:text(item.entity_type||"temporary_character"),count:Math.max(1,Number(item.count||1)),identity_hint:text(item.identity_hint),gender_hint:text(item.gender_hint),age_stage_hint:text(item.age_stage_hint),participation_state:text(item.participation_state||"visible_candidate"),must_render:item.must_render===true})),
      crowd_characters:array(plan.crowd_characters).map(item=>({label:text(item.label),count:Math.max(1,Number(item.count||1)),identity_hint:text(item.identity_hint),participation_state:text(item.participation_state||"background_visible")})),
      offscreen_voice:array(plan.character_core_context?.offscreen_voice).map(item=>({slot_id:item.slot_id,display_name:item.display_name,scene_role:'offscreen_voice'})),
      mentioned_only:array(plan.character_core_context?.mentioned_only).map(item=>({slot_id:item.slot_id,display_name:item.display_name,scene_role:'mentioned_only'})),
      exclude_all_other_registered_slots:true,
      ...(array(plan.excluded_formal_slot_ids).length<=12?{excluded_slot_ids:array(plan.excluded_formal_slot_ids)}:{}),
    }));
  }
  function buildV77SceneSubmissionPrompt(plans=[]) {
    try { globalThis.__commitLatestEditableUiState?.("before_outline_cast_package"); } catch (_) {}
    try {
      const svc=globalThis.__V78_SCENE_CONTEXT_SERVICE__;
      const built=svc?.buildOutlinePrompt?.(plans);
      const parity=built?.package?svc?.parityAgainstPlans?.(plans,built.package):null;
      const decision=built?.package?svc?.resolveOutlineAuthority?.(plans,built.package,parity):null;
      if(built?.ok&&parity?.ok&&decision?.takeover){
        console.info('[V78_SCENE_CONTEXT]',{stage:'outline_submission',executor:decision.executor,scene_count:built.package?.scenes?.length||0,parity:true,safe_streak:decision.safe_streak,threshold:decision.threshold});
        try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'outline',executor:decision.executor,scene_count:built.package?.scenes?.length||0,parity:true,safe_streak:decision.safe_streak,threshold:decision.threshold});}catch(_){}
        return built.prompt;
      }
      if(svc){console.warn('[V78_SCENE_CONTEXT] Phase8提交包尚未晋升或校验未通过，本次使用CharacterCore Validator/Fallback',{validation:built?.package?.validation||null,parity,authority:decision});try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'outline',executor:decision?.executor||'character_core_submission_fallback',parity:Boolean(parity?.ok),validation_ok:Boolean(built?.ok),safe_streak:Number(decision?.safe_streak||0)});}catch(_){}}
    }catch(error){console.warn('[V78_SCENE_CONTEXT] Phase8上下文构建异常，自动回退CharacterCore旧提交包',error);}
    const buildNormal=(appearanceLimit,compact)=>JSON.stringify({
      protocol:'character_core_pre_ai_keyword_cast_v77_hotfix19',
      character_registry:compactCharacterRegistryFromPlans(plans,appearanceLimit),
      scene_casting:sceneCastingRowsForSubmission(plans,compact),
    });
    let payload=buildNormal(1400,false),compactMode=false;
    if(payload.length>12000){
      compactMode=true;
      const c=core();
      const usedIds=unique(array(plans).flatMap(plan=>[
        ...array(plan.character_packages).map(item=>item.slot_id),
        ...array(plan.character_core_context?.offscreen_voice).map(item=>item.slot_id),
        ...array(plan.character_core_context?.mentioned_only).map(item=>item.slot_id),
      ]));
      const slots=usedIds.map(id=>c.slots.find(slot=>slot.slot_id===id)).filter(Boolean);
      const indexById=new Map(slots.map((slot,index)=>[slot.slot_id,index]));
      const roleCode={visible:'v',speaker_visible:'s',addressee_visible:'a',group_visible:'g',photo_visible:'p',screen_visible:'c',memory_visible:'m',dream_visible:'d',offscreen_voice:'o',mentioned_only:'n'};
      const registry=slots.map((slot,index)=>[index,slot.slot_id,slot.slot_token,slot.display_name,slot.gender,visibleStage(slot)]);
      const rows=array(plans).slice(0,150).map(plan=>[
        plan.line_index,
        plan.characters_mode==='manual'?1:0,
        array(plan.character_packages).map(item=>[indexById.get(item.slot_id),roleCode[item.scene_role]||'v',text(array(item.matched_keywords)[0]),text(array(item.relationship_evidence)[0]?.relation_id)]),
        array(plan.temporary_characters).map(item=>[text(item.source_phrase||item.label),text(item.participation_state),Math.max(1,Number(item.count||1))]),
        array(plan.crowd_characters).map(item=>[text(item.label),text(item.description)]),
        array(plan.character_core_context?.offscreen_voice).map(item=>indexById.get(item.slot_id)).filter(Number.isInteger),
        array(plan.character_core_context?.mentioned_only).map(item=>indexById.get(item.slot_id)).filter(Number.isInteger),
      ]);
      payload=JSON.stringify({p:'cc77h15',a:'完整appearance从同一请求characters数组按slot_id读取',r:registry,c:rows,k:'r=[index,slot_id,slot_token,name,gender,stage];c=[line,manual,selected[index,role,keyword,relation_id],temporary[source_phrase,participation_state,count],crowd[label,description],offscreen_indexes,mentioned_indexes];除selected外注册正式人物禁止实体出镜，但temporary/crowd必须按原文落实'});
      if(payload.length>26000){
        const leanRows=array(plans).slice(0,150).map(plan=>[
          plan.line_index,
          plan.characters_mode==='manual'?1:0,
          array(plan.character_packages).map(item=>[indexById.get(item.slot_id),roleCode[item.scene_role]||'v']),
          array(plan.temporary_characters).map(item=>text(item.source_phrase||item.label)),
          array(plan.crowd_characters).map(item=>text(item.label)),
          array(plan.character_core_context?.offscreen_voice).map(item=>indexById.get(item.slot_id)).filter(Number.isInteger),
          array(plan.character_core_context?.mentioned_only).map(item=>indexById.get(item.slot_id)).filter(Number.isInteger),
        ]);
        payload=JSON.stringify({p:'cc77h15-lean',a:'完整appearance从同一请求characters数组按slot_id读取',r:registry,c:leanRows,k:'c=[line,manual,selected[index,role],temporary_phrases,crowd_labels,offscreen,mentioned];正式人物唯一来源是强制名单slots；名单外人物只作为临时人物/群体/仅提及实体，不得要求补正式人物卡'});
      }
    }
    try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'outline',executor:'character_core_submission_legacy_fallback',compact_mode:compactMode,payload_chars:payload.length});}catch(_){}
    return [
      '【V77 Hotfix19 请求前已冻结的逐镜选角数据｜必须执行】',
      '逐镜 selected 已在请求AI前由人物关系图、人物卡关键词和人工覆盖计算完成，slot_id是正式人物唯一身份。AI只负责生成画面，不得新增、删除、替换 selected，也不得用返回内容修改正式人物勾选。temporary_characters 是名单外通用临时演员的最低必达数据，不是封闭白名单；AI还要阅读当前原文补充识别其他可见临时人物。visible/visual_exposition/visible_candidate/group_visible/background_visible必须实体进入成品画面，并沿用continuity_id/continuity_profile/wardrobe_state保持同场景短期一致；offscreen只可发声，mentioned不得实体出镜。禁止在prompt输出字段名、分类名、无人物卡或规则说明。',
      compactMode?'【紧凑传输】以下JSON保持全部分镜行，不截断任何行。':'【完整传输】以下JSON包含人物卡外形、逐镜关键词、关系证据与临时角色。',
      payload,
    ].join('\n');
  }
  function installOutlineAdapter() {
    if(typeof buildLineCharacterPlansV36==='function'&&!buildLineCharacterPlansV36.__characterCoreV2Hotfix16){
      const previous=buildLineCharacterPlansV36;
      const wrapped=function(novelText=''){
        const c=core(),plans=array(previous(novelText)),previousStageByPerson=new Map();
        return plans.map((plan,index)=>{
          const planSourceKey=text(plan.source_key)||sceneSourceKey(plan.source_text,index+1);
          const existingScene=array(appState.scenes).find(scene=>text(scene.source_key)===planSourceKey)||array(appState.scenes).find(scene=>Number(scene.source_index)===index+1)||array(appState.scenes)[index];
          const manual=existingScene?.characters_mode==='manual'||plan.characters_mode==='manual'||plan.whitelist_source==='current_scene_selection';
          const frozen=Boolean(existingScene?.preselected_cast_frozen&&array(existingScene?.character_slot_ids).length);
          let ctx,visual=[];
          if(frozen){
            const frozenRows=array(existingScene.character_core_cast).filter(item=>array(existingScene.character_slot_ids).includes(item.slot_id));
            visual=frozenRows.map(expandCastItem).filter(Boolean);
            const personIds=new Set(visual.map(item=>item.person_id));
            ctx={casting_protocol:'character_core_pre_ai_cast_snapshot_v77_hotfix19',manual_locked:manual,selected_characters:visual,visible_characters:visual.filter(item=>!['offscreen_voice','mentioned_only'].includes(item.scene_role)),offscreen_voice:visual.filter(item=>item.scene_role==='offscreen_voice'),mentioned_only:visual.filter(item=>item.scene_role==='mentioned_only'),temporary_characters:mergeTemporaryCharacters(existingScene.temporary_characters,temporaryEntitiesForSource(plan.source_text||'',index+1)),keyword_evidence:deepClone(existingScene.keyword_evidence||[]),relationship_evidence:deepClone(existingScene.relationship_evidence||[]),relationship_context:c.relationships.filter(r=>personIds.has(r.source_person_id)&&personIds.has(r.target_person_id)&&!r.disabled),continuity_context:{},snapshot_frozen:true};
          }else{
            const ids=manual?array(existingScene?.character_slot_ids):null;
            ctx=buildSceneContext(plan.source_text||'',ids,index+1,previousStageByPerson);
            visual=array(ctx.visible_characters);
          }
          if(!visual.length&&!manual){
            const fallback=preselectFormalCast(plan.source_text||'',index+1,previousStageByPerson).map(expandCastItem).filter(Boolean);
            visual=fallback;ctx={...ctx,selected_characters:fallback,visible_characters:fallback,keyword_evidence:fallback.flatMap(item=>array(item.keyword_evidence)),relationship_evidence:fallback.flatMap(item=>array(item.relationship_evidence))};
          }
          const temporary=mergeTemporaryCharacters(plan.temporary_characters,ctx.temporary_characters,existingScene?.temporary_characters);
          const allowed=visual.map(item=>item.display_name),selectedIds=new Set(visual.map(item=>item.slot_id));
          return{
            ...plan,source_key:planSourceKey,allowed,character_slot_ids:visual.map(item=>item.slot_id),
            relation_selected_slot_ids:unique(array(existingScene?.relation_selected_slot_ids)),keyword_selected_slot_ids:unique(array(existingScene?.keyword_selected_slot_ids)),manual_added_slot_ids:unique(array(existingScene?.manual_added_slot_ids)),manual_excluded_slot_ids:unique(array(existingScene?.manual_excluded_slot_ids)),
            relation_cast_rows:deepClone(array(existingScene?.relation_cast_rows)),keyword_cast_rows:deepClone(array(existingScene?.keyword_cast_rows)),
            stage_bindings:visual.map(item=>({name:item.display_name,base_name:item.display_name,stage:item.visual_age_stage,scene_identity_label:item.visual_age_stage,gender:item.gender,appearance:item.appearance,slot_id:item.slot_id,slot_token:item.slot_token,person_id:item.person_id,scene_role:item.scene_role,matched_keywords:item.matched_keywords,relationship_evidence:item.relationship_evidence})),
            character_packages:visual,temporary_characters:temporary,
            keyword_evidence:visual.flatMap(item=>array(item.keyword_evidence)),relationship_evidence:visual.flatMap(item=>array(item.relationship_evidence)),
            character_core_context:{...ctx,selected_characters:visual,visible_characters:visual,temporary_characters:temporary,snapshot_frozen:true},
            excluded_formal_characters:c.slots.filter(slot=>!selectedIds.has(slot.slot_id)).map(slot=>slot.display_name),
            excluded_formal_slot_ids:c.slots.filter(slot=>!selectedIds.has(slot.slot_id)).map(slot=>slot.slot_id),
            visual_center:allowed[0]||temporary[0]?.source_phrase||temporary[0]?.label||plan.visual_center||'',
            characters_mode:manual?'manual':'auto',whitelist_source:manual?'current_scene_selection':'pre_ai_relationship_keyword_snapshot',preselected_cast_frozen:true,
          };
        });
      };
      wrapped.__characterCoreV2=true;wrapped.__characterCoreV2Hotfix16=true;buildLineCharacterPlansV36=wrapped;
    }
    if(typeof buildLineCharacterWhitelistPromptV36==='function'&&!buildLineCharacterWhitelistPromptV36.__characterCoreV2Hotfix16){
      const previousPrompt=buildLineCharacterWhitelistPromptV36;
      const wrappedPrompt=function(plans=[]){return`${previousPrompt(plans)}\n\n【V77 Hotfix19选角说明】本镜正式人物已在AI请求前完成slot_id计算并冻结。AI只能使用逐镜selected正式人物卡。临时、群体和系统实体是结构化镜头数据，仅根据participation_state自然融入画面，不输出字段名、分类名或规则说明。`;};
      wrappedPrompt.__characterCoreV2Hotfix16=true;buildLineCharacterWhitelistPromptV36=wrappedPrompt;
    }
    if(typeof buildOutlineInstructionAppendix==='function'&&!buildOutlineInstructionAppendix.__characterCoreV2Hotfix16){
      const previousAppendix=buildOutlineInstructionAppendix;
      const wrappedAppendix=function(options={}){const v77=buildV77SceneSubmissionPrompt(array(options?.lineCharacterPlans));const base=previousAppendix(options);return`${v77}\n\n${base}`;};
      wrappedAppendix.__characterCoreV2Hotfix16=true;buildOutlineInstructionAppendix=wrappedAppendix;
    }
    if(typeof buildRegenerationInstructionAppendix==='function'&&!buildRegenerationInstructionAppendix.__characterCoreV2Hotfix16){
      const previousRegeneration=buildRegenerationInstructionAppendix;
      const wrappedRegeneration=function(scene={},guidance=''){
        const sourceKey=text(scene.source_key)||sceneSourceKey(scene.source_text||scene.text,scene.source_index||1);
        const base=previousRegeneration(scene,guidance),stored=array(appState.scenes).find(item=>text(item.source_key)===sourceKey)||array(appState.scenes).find(item=>item.id===scene.id)||scene;
        try{
          const svc=globalThis.__V78_SCENE_CONTEXT_SERVICE__,sceneInput={...stored,...scene,source_key:sourceKey},built=svc?.buildRegenerationPrompt?.(sceneInput);
          const parity=built?.package?svc?.parityRegenerationAgainstScene?.(sceneInput,built.package):null;
          const decision=built?.package?svc?.resolveRegenerationAuthority?.(sceneInput,built.package,parity):null;
          if(built?.ok&&parity?.ok&&decision?.takeover){console.info('[V78_SCENE_CONTEXT]',{stage:'regeneration_submission',executor:decision.executor,source_key:sourceKey,safe_streak:decision.safe_streak,threshold:decision.threshold});try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'regeneration',executor:decision.executor,source_key:sourceKey,parity:true,safe_streak:decision.safe_streak,threshold:decision.threshold});}catch(_){}return `${built.prompt}\n\n${base}`;}
          if(svc){console.warn('[V78_SCENE_CONTEXT] Phase8单镜提交包尚未晋升或校验未通过，本次使用CharacterCore Validator/Fallback',{validation:built?.package?.validation||null,parity,authority:decision});try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'regeneration',executor:decision?.executor||'character_core_submission_fallback',source_key:sourceKey,parity:Boolean(parity?.ok),validation_ok:Boolean(built?.ok),safe_streak:Number(decision?.safe_streak||0)});}catch(_){}}
        }catch(error){console.warn('[V78_SCENE_CONTEXT] 单镜上下文构建异常，自动回退CharacterCore旧提交包',error);}
        const ids=array(stored.character_slot_ids),ctx=buildSceneContext(scene.source_text||stored.source_text||'',ids,Number(stored.source_index||0)||null,new Map()),selectedIds=new Set(ids);
        const registry=ids.map(id=>{const slot=core().slots.find(item=>item.slot_id===id);return slot?{slot_id:slot.slot_id,slot_token:slot.slot_token,person_id:slot.person_id,display_name:slot.display_name,gender:slot.gender,visual_age_stage:visibleStage(slot),appearance:String(slot.appearance||'').slice(0,1600),scene_role:'visible'}:null;}).filter(Boolean);
        const pack={protocol:'character_core_scene_regeneration_preselected_v77_hotfix19',preselected_before_ai:true,manual_locked:stored.characters_mode==='manual',selected_characters:registry,temporary_characters:mergeTemporaryCharacters(stored.temporary_characters,ctx.temporary_characters),offscreen_voice:ctx.offscreen_voice.map(item=>({slot_id:item.slot_id,display_name:item.display_name,scene_role:'offscreen_voice'})),mentioned_only:ctx.mentioned_only.map(item=>({slot_id:item.slot_id,display_name:item.display_name,scene_role:'mentioned_only'})),relationship_context:ctx.relationship_context,keyword_evidence:stored.keyword_evidence||ctx.keyword_evidence,excluded_slot_ids:core().slots.filter(slot=>!selectedIds.has(slot.slot_id)).map(slot=>slot.slot_id)};
        try{globalThis.__V78_OUTLINE_GENERATION_SERVICE__?.markStage?.('submission_package',{mode:'regeneration',executor:'character_core_submission_legacy_fallback',source_key:sourceKey,selected_count:registry.length});}catch(_){}
        return`【V77 Hotfix19当前分镜选角快照｜最高优先级】\n${JSON.stringify(pack)}\n正式人物复选框已经在请求前完成勾选并冻结；正式人物唯一来源是强制名单slots，你不得新增、修改或要求补充名单外正式人物卡。必须使用selected_characters中的完整外形。temporary_characters是结构化镜头数据，只根据participation_state自然融入最终画面，不输出字段名、分类名、规则说明或执行过程。\n\n${base}`;
      };
      wrappedRegeneration.__characterCoreV2Hotfix16=true;buildRegenerationInstructionAppendix=wrappedRegeneration;
    }
  }

  const v77BaseGenerateOutline = typeof globalThis.__V77_NATIVE_OUTLINE_ENGINE__ === "function" ? globalThis.__V77_NATIVE_OUTLINE_ENGINE__ : (typeof generateOutline === "function" ? generateOutline : null);
  const v77BaseRegenerateScene = typeof globalThis.__V77_NATIVE_REGENERATE_ENGINE__ === "function" ? globalThis.__V77_NATIVE_REGENERATE_ENGINE__ : (typeof regenerateSceneOutline === "function" ? regenerateSceneOutline : null);
  let outlineRequestInFlight = false;
  let runtimeBuildMismatch = "";
  let outlineClickSequence = 0;

  async function runFullAnalysisV77(button=q("#analyzeBtn")) {
    // Hotfix4: legacy global analyzeNovel is intentionally style-only.
    // Character cards, AI relationship graph and local keyword casting belong
    // exclusively to the forced-roster button.
    return runStyle(button,{silentStatus:false});
  }

  async function runOutlineV77(v78Options={}) {
    const facade=globalThis.__V78_OUTLINE_GENERATION_SERVICE__;
    if(facade&&!v78Options?.__v78FacadeInternal){return facade.runOutline({executeNative:()=>runOutlineV77({__v78FacadeInternal:true})});}
    try { globalThis.__commitLatestEditableUiState?.('before_outline_entry'); } catch (_) {}
    if(outlineRequestInFlight){showAIStatusNotice?.('整段分镜请求正在执行，请等待当前请求完成。','warning',5000);return null;}
    const clickId=++outlineClickSequence,snapshot=captureSourceSnapshot('outline_click');
    if(!snapshot.canonical_text)return apiError('无法生成：当前整段原文为空。');
    try{ensureOutlineSourceSceneShells();}catch(error){console.warn('[CAST_TRACE] 原文分镜壳建立失败',error);}
    const rosterSnapshot=formalRosterSnapshot('outline_formal_roster_reconcile',true);
    const guideText=String(rosterSnapshot.raw_text||'').trim();
    // V78.2.8: every outline request reconciles from the same frozen formal-roster snapshot.
    try{await parseSlots({preserve:true,rosterSnapshot});assertFormalRosterFresh(rosterSnapshot,'分镜正式人物对账');}catch(error){console.warn('[OUTLINE_TRACE] Formal Roster对账失败；不从原文或旧状态补正式人物',error);}
    if(!core().people.length&&core().slots.length){try{ensureFallbackPeopleGraph();}catch(error){console.warn('[OUTLINE_TRACE] people保底失败但继续',error);}}
    try{rebuildKeywordIndex();}catch(error){console.warn('[OUTLINE_TRACE] 关键词重建失败但继续',error);}
    let castSnapshot=null;
    try{
      castSnapshot=refreshSceneBindings({preRequest:true,render:true});
      try{facade?.markStage?.('cast_frozen',{scene_count:array(appState.scenes).length,selected_slot_count:array(castSnapshot?.scenes).reduce((sum,item)=>sum+array(item.selected_slot_ids).length,0)});}catch(_){}
      showAIStatusNotice?.(`已先完成 ${array(appState.scenes).length} 个分镜的人物关系图+本地关键词匹配、本镜人物勾选与人物卡快照冻结，正在提交给AI。`,'ready',6000);
      await new Promise(resolve=>{if(typeof requestAnimationFrame==='function')requestAnimationFrame(()=>resolve());else setTimeout(resolve,0);});
    }catch(error){console.warn('[CAST_TRACE] 预选角失败，使用当前人工勾选继续',error);}
    try{syncLegacyState();markV77RuntimeReady(snapshot.canonical_text);}catch(error){console.warn('[OUTLINE_TRACE] 状态同步失败但继续',error);}
    const readiness=v77OutlineReadiness(),usable=readiness.usable||getOutlineUsableDataState(snapshot);
    const selectedCount=array(castSnapshot?.scenes).reduce((sum,item)=>sum+array(item.selected_slot_ids).length,0);
    const trace={stage:'outline_click',click_id:clickId,handler:'v77-hotfix19-manual-stage-roster-only',listener_count:1,build:BUILD_VERSION,source_present:usable.source_present,source_hash:usable.source_hash,style_available:usable.style_available,slot_count:usable.slot_count,appearance_count:usable.appearance_count,relationship_count:usable.relationship_count,alias_count:usable.alias_count,keyword_count:usable.keyword_count,preselected_scene_count:array(castSnapshot?.scenes).length,preselected_slot_count:selectedCount,ui_checkbox_synced:true,cast_snapshot_ready:Boolean(castSnapshot),hard_block:readiness.ok?false:(readiness.hard_block||'unknown'),request_started:Boolean(readiness.ok),warnings:usable.warnings};
    console.info('[OUTLINE_TRACE]',trace);
    emitCharacterCoreTrace({trace_stage:'outline_click',request_id:`outline-${Date.now()}-${clickId}`,outcome:readiness.ok?'request_started':'blocked',summary:readiness.ok?'本镜人物已在AI请求前预选并冻结，开始生成画面':readiness.message,subject_facts:trace}).catch(()=>{});
    if(!readiness.ok)return apiError(readiness.message);
    if(runtimeBuildMismatch)showAIStatusNotice?.(`${runtimeBuildMismatch}；已按当前页面资料继续尝试生成。`,'warning',8000);
    if(typeof v77BaseGenerateOutline!=='function')return apiError('分镜执行器缺失，无法发起AI请求。请确认软件文件完整。');
    outlineRequestInFlight=true;
    try{
      appState.analysisSourceText=snapshot.canonical_text;appState.analysisComplete=true;appState.characterRegistryStatus='ready';appState.characterAnalysisResultStatus='current';
      try{facade?.markStage?.('generator_dispatch',{generator:'v77_native_outline_generator'});}catch(_){}
      const result=await v77BaseGenerateOutline();
      try{facade?.markStage?.('generator_return',{generator:'v77_native_outline_generator'});}catch(_){}
      try{restoreSceneBindingsBySourceKey({render:true});facade?.markStage?.('post_writeback_refresh',{scene_count:array(appState.scenes).length,shot_count:array(appState.outlineShots).length,source_key_restore:true});}catch(error){console.warn('[CAST_TRACE] AI返回后source_key人物恢复失败',error);}
      try{const traceId=globalThis.__V77_LAST_TRACE_BY_ENDPOINT__?.["/api/outline-scenes"]||"";if(traceId)emitCharacterCoreTrace({trace_stage:"outline_writeback",request_id:traceId,outcome:"accepted",summary:"整段分镜已写入当前工作区并按source_key恢复人物勾选",scene_count:array(appState.scenes).length,shot_count:array(appState.outlineShots).length,source_key_count:new Set(array(appState.scenes).map(x=>text(x.source_key)).filter(Boolean)).size,final_snapshot:{scene_count:array(appState.scenes).length,outline_shot_count:array(appState.outlineShots).length,first_scenes:array(appState.scenes).slice(0,6).map(x=>({source_key:x.source_key,source_text:x.source_text,character_slot_ids:x.character_slot_ids}))}}).catch(()=>{});}catch(_){}
      return result;
    }finally{outlineRequestInFlight=false;}
  }

  async function runSceneRegenerationV77(sceneId,v78Options={}) {
    const facade=globalThis.__V78_OUTLINE_GENERATION_SERVICE__;
    if(facade&&!v78Options?.__v78FacadeInternal){return facade.runRegeneration(sceneId,{executeNative:()=>runSceneRegenerationV77(sceneId,{__v78FacadeInternal:true})});}
    try { globalThis.__commitLatestEditableUiState?.("before_scene_regeneration_entry"); } catch (_) {}
    const readiness=v77OutlineReadiness();if(!readiness.ok)return apiError(readiness.message);
    try{rebuildKeywordIndex();}catch(_){}
    let snapshot=null;
    try{snapshot=refreshSceneBindings({preRequest:true,render:true});facade?.markStage?.('cast_frozen',{scene_id:sceneId,scene_count:array(appState.scenes).length});}catch(error){console.warn('[CAST_SNAPSHOT_TRACE] 单条重生成预选角失败，继续使用当前勾选',error);}
    const scene=array(appState.scenes).find(item=>item.id===sceneId);
    console.info('[CAST_SNAPSHOT_TRACE]',{scene_id:sceneId,relation_selected_slot_ids:array(scene?.relation_selected_slot_ids),keyword_selected_slot_ids:array(scene?.keyword_selected_slot_ids),manual_added_slot_ids:array(scene?.manual_added_slot_ids),manual_excluded_slot_ids:array(scene?.manual_excluded_slot_ids),effective_slot_ids:array(scene?.character_slot_ids),snapshot_frozen:true});
    try{markV77RuntimeReady(currentNovel());}catch(_){}
    if(typeof v77BaseRegenerateScene!=="function")return apiError("V77原生单条重生成执行器缺失，请确认软件文件完整。");
    try{facade?.markStage?.('generator_dispatch',{generator:'v77_native_regeneration_generator',scene_id:sceneId,source_key:scene?.source_key||''});}catch(_){}
    const result=await v77BaseRegenerateScene(sceneId);
    try{facade?.markStage?.('generator_return',{generator:'v77_native_regeneration_generator',scene_id:sceneId});}catch(_){}
    if(scene){scene.preselected_cast_frozen=false;scene.character_core_manual_locked=false;}
    try{const traceId=globalThis.__V77_LAST_TRACE_BY_ENDPOINT__?.["/api/regenerate-scene-outline"]||"";if(traceId)emitCharacterCoreTrace({trace_stage:"regenerate_writeback",request_id:traceId,outcome:"accepted",summary:"当前分镜重生成结果已写回",final_snapshot:{scene_id:sceneId,source_key:scene?.source_key||"",character_slot_ids:array(scene?.character_slot_ids),outline_shots:array(scene?.outline_shots).slice(0,8)}}).catch(()=>{});}catch(_){}
    return result;
  }

  async function runOutlineProductionV78(){
    const facade=globalThis.__V78_OUTLINE_GENERATION_SERVICE__;
    if(facade)return facade.runOutline({executeNative:()=>runOutlineV77({__v78FacadeInternal:true})});
    return runOutlineV77({__v78FacadeInternal:true});
  }
  async function runSceneRegenerationProductionV78(sceneId){
    const facade=globalThis.__V78_OUTLINE_GENERATION_SERVICE__;
    if(facade)return facade.runRegeneration(sceneId,{executeNative:()=>runSceneRegenerationV77(sceneId,{__v78FacadeInternal:true})});
    return runSceneRegenerationV77(sceneId,{__v78FacadeInternal:true});
  }

  function installV77ExclusiveEntrypoints() {
    // V78 Stable: historical V77 engines remain compatibility-only fallbacks; all public entrypoints bind V78 production.
  globalThis.__V78_OUTLINE_NATIVE_EXECUTORS__={version:'v78_stable_v77_compatibility_host',compatibility_only:true,outline:()=>runOutlineV77({__v78FacadeInternal:true}),regenerate:(sceneId)=>runSceneRegenerationV77(sceneId,{__v78FacadeInternal:true})};
    try { if(typeof generateOutline==="function")generateOutline=runOutlineProductionV78; } catch(_) {}
    try { if(typeof regenerateSceneOutline==="function")regenerateSceneOutline=runSceneRegenerationProductionV78; } catch(_) {}
    globalThis.analyzeNovel=(button=q("#analyzeBtn"))=>runStyle(button,{silentStatus:false});
    globalThis.generateOutline=runOutlineProductionV78;
    globalThis.regenerateSceneOutline=runSceneRegenerationProductionV78;
  }

  function renderAll(){syncLegacyState();renderCharacters();renderRelations();if(typeof renderScenes==="function")renderScenes();}
  function cloneRebind(selector,handler,handlerVersion=""){const old=q(selector);if(!old)return null;const fresh=old.cloneNode(true);old.replaceWith(fresh);if(handlerVersion){fresh.dataset.handlerVersion=handlerVersion;fresh.dataset.executorId=handlerVersion;}fresh.dataset.executorOwner="CharacterCore2";fresh.addEventListener("click",handler);return fresh;}
  function bindUI(){
    cloneRebind("#analyzeBtn",()=>runStyle(q("#analyzeBtn"),{silentStatus:false}),"v78-stable-style-ai-service");
    cloneRebind("#optimizeAllCharactersBtn",()=>runCharacters({button:q("#optimizeAllCharactersBtn")}),"v78-stable-character-ai-service");
    const outlineButtonV78=cloneRebind("#outlineBtn",()=>runOutlineProductionV78(),"v78-stable-outline-production-mainline");if(outlineButtonV78)outlineButtonV78.dataset.executorOwner="V78ProductionMainline";
    cloneRebind("#saveProjectBtn",saveProjectWithLease,"v78-stable-history-runtime");
    const addButton=cloneRebind("#addCharacterBtn",()=>{const guide=q("#characterGuideInput");if(!guide)return;guide.focus();if(String(guide.value||"").trim()&&!/[\n；;，,、|｜]\s*$/.test(guide.value))guide.value=`${guide.value.trim()}\n`;guide.dispatchEvent(new Event("input",{bubbles:true}));showAIStatusNotice("请在强制名单中添加人物；人物卡只由名单槽位生成。","ready",5000);});
    if(addButton){addButton.disabled=false;addButton.textContent="+ 在强制名单添加";addButton.title="人物卡只能来自强制名单；点击后定位到名单输入框。";}
    cloneRebind("#relationshipAddBtn",addRelation);cloneRebind("#relationshipAliasAddBtn",addAlias);cloneRebind("#relationshipReanalyzeBtn",async()=>{try{await analyzeFactsAndRelations();rebuildKeywordIndex();refreshSceneBindings();renderAll();const ok=Boolean(analysisState().relationships_ready);showAIStatusNotice(ok?"已由AI重新分析人物关系和动态称呼映射，本地关键词索引已同步重建；人工锁定关系未被覆盖。":"AI人物关系请求已返回，但关系字段仍不完整；现有人物卡与本地关键词继续可用。",ok?"ready":"warning",9000);}catch(error){setAnalysisStage("relationships",false,error?.message||String(error));ensureFallbackPeopleGraph();rebuildKeywordIndex();refreshSceneBindings();renderAll();showAIStatusNotice(`人物关系AI请求未完成：${error?.message||error}。现有人物卡和本地关键词仍可继续使用。`,"warning",12000);}});cloneRebind("#relationshipSaveBtn",()=>{try{rebuildKeywordIndex();refreshSceneBindings({preRequest:false,render:true});}catch(_){}scheduleDraftSave?.();showAIStatusNotice("人物关系已保存；本地关键词索引与每个分镜的本镜人物勾选已立即更新。","ready",6000);});
    const novel=q("#novelText");if(novel&&!novel.dataset.characterCoreSourceBound){novel.dataset.characterCoreSourceBound="true";novel.addEventListener("input",()=>{const c=core(),snapshot=captureSourceSnapshot("novel_input");if(!c.source_stale&&c.source_hash&&c.source_hash!==snapshot.source_hash){c.source_stale=true;c.character_revision+=1;c.people=[];c.slots=[];c.relationships=[];c.aliases=[];c.mention_entities=[];c.scene_casting={};c.relationship_pending=[];c.temporary_continuity_cache={};c.temporary_continuity_cache_character_core_validator={};c.temporary_continuity_cache_v78_primary={};c.temporary_continuity_authority_v78={};c.relationship_conflicts=[];resetAnalysisState("整段原文实际内容已更换");appState.analysisComplete=false;appState.analysisSourceText="";appState.characterRegistryStatus="stale";appState.characterAnalysisResultStatus="stale";syncLegacyState();renderAll();refreshSceneBindings();const p=q("#characterGenerationProgress");if(p)p.textContent=`检测到整段原文实际内容已变化：当前 revision ${snapshot.source_revision}。旧人物事实、外形、关系和分镜选角已隔离；强制名单保留，请重新生成。`;}});}
    const guide=q("#characterGuideInput");if(guide){
      const fresh=guide.cloneNode(true);fresh.value=guide.value;guide.replaceWith(fresh);
      const roster=formalRosterApi();roster?.bindInput?.();roster?.commitFromDom?.("character_core_bind");
      if(roster&&!globalThis.__V78_FORMAL_ROSTER_CC_SUBSCRIBED__){
        globalThis.__V78_FORMAL_ROSTER_CC_SUBSCRIBED__=true;
        roster.subscribe((next,_previous,meta)=>{if(meta?.source!=="user_input")return;invalidateRosterStages("强制人物名单已修改；统一风格仍保留当前原文版本");appState.analysisComplete=false;appState.characterRegistryStatus="stale";appState.characterAnalysisResultStatus="stale";syncLegacyState();renderAll();const p=q("#characterGenerationProgress");if(p)p.textContent=`正式人物名单已更新为 revision ${next.revision}（${next.entries?.length||0} 个条目）；统一风格保持有效，请按当前名单重新生成人物卡。`;scheduleDraftSave?.();});
      }
    }
    const ref=q("#appearanceReference");
  }

  function installProjectAdapter(){
    if(typeof getProjectData==="function"&&!getProjectData.__characterCoreV2){const prev=getProjectData;getProjectData=function(){const roster=formalRosterSnapshot("project_save",true);const fullGuide=String(roster.raw_text||"");const fullAppearanceReference=String(valueOf("#appearanceReference",appState.appearance_reference||"")||"");appState.character_guide_input=fullGuide;appState.appearance_reference=fullAppearanceReference;return{...prev(),character_guide_input:fullGuide,formal_roster_snapshot_v7828:deepClone(roster),appearance_reference:fullAppearanceReference,character_core_v2:deepClone(core()),character_core_version:2,character_core_instance_id:currentInstance()};};getProjectData.__characterCoreV2=true;}
    if(typeof applyProjectData==="function"&&!applyProjectData.__characterCoreV2){const prev=applyProjectData;applyProjectData=function(projectId,name,data={}){prev(projectId,name,data);const fullGuide=String(data.formal_roster_snapshot_v7828?.raw_text??data.character_guide_input??"");const fullAppearanceReference=String(data.appearance_reference||"");formalRosterApi()?.restore?.(fullGuide,{source:"history_restore",reason:"apply_project"});appState.character_guide_input=fullGuide;setValue("#characterGuideInput",fullGuide);formalRosterApi()?.bindInput?.();if(fullAppearanceReference){appState.appearance_reference=fullAppearanceReference;setValue("#appearanceReference",fullAppearanceReference);}ensureProjectLease({interactive:false}).then(acquired=>{if(!acquired)showAIStatusNotice("当前工作区不使用项目编辑锁。","warning",12000);}).catch(()=>{});const incoming=object(data.character_core_v2||data.character_core);if(Number(incoming.character_core_version)>=2){appState.characterCoreV2=deepClone(incoming);appState.characterCoreV2.instance_id=currentInstance();rebuildKeywordIndex();setAnalysisStage("casting",Boolean(Object.keys(core().keyword_index||{}).length),"已从历史/当前工作区恢复人物语义层");markV77RuntimeReady(currentNovel());renderAll();refreshSceneBindings();}else{postCore("/api/character-core/migrate-project",{project:data,guide_text:fullGuide||appState.character_guide_input||"",novel_text:currentNovel()}).then(result=>{appState.characterCoreV2=result.character_core;appState.characterCoreV2.instance_id=currentInstance();rebuildKeywordIndex();setAnalysisStage("casting",Boolean(Object.keys(core().keyword_index||{}).length),"已从历史/当前工作区恢复人物语义层");markV77RuntimeReady(currentNovel());renderAll();refreshSceneBindings();if(result.migrated)showAIStatusNotice("旧历史人物数据已迁移到CharacterCore 2.0运行结构。","ready",8000);}).catch(()=>{});}};applyProjectData.__characterCoreV2=true;}
  }

  async function init(){
    ensureCore();globalThis.__activeAnalyzeProtocol=PROTOCOL;globalThis.__characterAppearanceSkillVersion=VERSION;globalThis.__VIDEO_PROMPT_TOOL_BUILD__={...(globalThis.__VIDEO_PROMPT_TOOL_BUILD__||{}),version:BUILD_VERSION,characterPipeline:PROTOCOL,instance:currentInstance(),multiInstance:true};
    // V78.2.8: take UI/data authority synchronously before any awaited health
    // request. This removes the startup race where legacy app.js handlers could
    // receive a click before CharacterCore replaced them.
    installV77ExclusiveEntrypoints();bindUI();installProjectAdapter();installOutlineAdapter();
    try{
      const response=await fetch(`/api/character-core/health?instance=${encodeURIComponent(currentInstance())}`,{cache:"no-store",headers:{"Cache-Control":"no-cache"}});
      const health=await response.json();
      if(text(health.app_version)!==BUILD_VERSION)runtimeBuildMismatch=`当前页面脚本为 ${BUILD_VERSION}，后端为 ${text(health.app_version)||"未知版本"}。请关闭旧进程并使用当前目录的 START_CLEAN.ps1 重新启动。`;
    }catch(error){runtimeBuildMismatch=`无法验证当前后端版本：${error?.message||error}`;}
    const badge=document.createElement("span");badge.id="v77NativeBuildBadge";badge.className="status-pill";badge.textContent=runtimeBuildMismatch?"版本不一致":"V78.3.0.2 · 场景锚点/事件归属/连续时间轴根治";badge.title=runtimeBuildMismatch||"单一运行时版本源；旧Hotfix模块不得再改软件版本徽标；历史记录、AI指令中心、临时人物等仅作为功能模块运行。";badge.dataset.runtimeVersion=BUILD_VERSION;badge.dataset.runtimeBuildId=(globalThis.__V78_CURRENT_RUNTIME__||globalThis.__V77_CURRENT_RUNTIME__)?.buildId||"v78.3.0.2-scene-event-canonical-timeline-20260818-r1";q("#aiStatus")?.insertAdjacentElement("afterend",badge);
    if(!runtimeBuildMismatch) (globalThis.__v78ApplyRuntimeBadge||globalThis.__v77ApplyRuntimeBadge)?.();
    if(runtimeBuildMismatch)showAIStatusNotice(runtimeBuildMismatch,"warning",20000);
    globalThis.addEventListener?.("beforeunload",releaseProjectLease,{once:true});
    globalThis.analyzeNovel=()=>runStyle(q("#analyzeBtn"),{silentStatus:false});globalThis.generateOutline=runOutlineProductionV78;globalThis.regenerateSceneOutline=runSceneRegenerationProductionV78;globalThis.optimizeAllCharacters=()=>runCharacters({button:q("#optimizeAllCharactersBtn")});globalThis.optimizeCharacterDescription=(index)=>{const slot=core().slots[index];return slot?runCharacters({onlySlotId:slot.slot_id,button:q(`[data-optimize-character="${index}"]`)}):null;};globalThis.renderCharacters=renderCharacters;
    const stableRegistry={build:BUILD_VERSION,owner:"V78ReleaseCore",release:"stable",installed_at:new Date().toISOString(),executors:{analyze:{id:"v78-stable-style-ai-service",owner:"V78CharacterAIService",function_name:"runStyle"},characters:{id:"v78-stable-character-ai-service",owner:"CharacterCore2+V78CharacterAIService",function_name:"runCharacters"},outline:{id:"v78-stable-outline-production-mainline",owner:"V78ProductionMainline",function_name:"runOutlineProductionV78"},regenerate:{id:"v78-stable-regenerate-production-mainline",owner:"V78ProductionMainline",function_name:"runSceneRegenerationProductionV78"},save:{id:"v78-stable-history-runtime",owner:"CharacterCore2",function_name:"saveProjectWithLease"}},compatibility_only:{outline:"runOutlineV77",regenerate:"runSceneRegenerationV77",protocol_markers:"V77_CHARACTER_CORE_*"}};globalThis.__V78_EXECUTOR_REGISTRY__=stableRegistry;globalThis.__V77_EXECUTOR_REGISTRY__=stableRegistry;
    const bootRoster=formalRosterSnapshot("character_core_init",true);if(!core().slots.length&&String(bootRoster.raw_text||"").trim()){try{await parseSlots({preserve:true,rosterSnapshot:bootRoster});}catch(_){} }
    renderAll();refreshSceneBindings();
  }

  function coreParityHealth() {
    const c=core(),st=analysisState(),slots=c.slots.filter((slot)=>!slot.disabled);
    return {
      build:BUILD_VERSION,
      slot_count:slots.length,
      unresolved_gender:slots.filter((slot)=>isUnresolvedGender(slot.gender)).map((slot)=>slot.slot_id),
      unresolved_stage:slots.filter((slot)=>isUnresolvedStage(visibleStage(slot))).map((slot)=>slot.slot_id),
      appearance_ready:slots.filter((slot)=>Boolean(text(slot.appearance))).length,
      facts_ready:Boolean(st.facts_ready),
      facts_fallback_available:Boolean(st.facts_fallback_available),
      facts_integrity:characterFactsIntegrity(c),
      relationships_ready:Boolean(st.relationships_ready),
      relationship_fallback_available:Boolean(st.relationship_fallback_available),
      relationship_count:c.relationships.filter((row)=>!row.disabled).length,
      alias_count:c.aliases.filter((row)=>!row.disabled).length,
      keyword_person_count:Object.keys(c.keyword_index||{}).length,
      scene_cast_count:Object.keys(c.scene_casting||{}).length,
      temporary_continuity_count:Object.keys(c.temporary_continuity_cache||{}).length,
      relationship_conflict_count:v36RelationshipConflicts(c).length,
      character_health:c.slots.filter(slot=>!slot.disabled).map(slot=>({slot_id:slot.slot_id,display_name:slot.display_name,...v36CharacterHealth(slot)})),
      output_mode:text(appState.outputMode||"normal"),
    };
  }

  globalThis.__V78_CHARACTER_AI_HOST_BRIDGE__={
    version:"character_ai_host_bridge_v78_phase16",
    markers:{style:STYLE_MARKER,facts_relationships:FACTS_MARKER,appearance:APPEARANCE_MARKER,revision_check:REVISION_CHECK_MARKER},
    markerStage,
    currentInstance,
    resolveTimeoutSeconds:resolveRuntimeTimeoutSeconds,
    buildLegacyPayload:(prompt,marker,meta={})=>deepClone(buildCharacterAiRequestPayload(prompt,marker,meta)),
    legacyRawRequest:({prompt="",marker="",meta={}}={})=>legacyAiAnalyzeRaw(prompt,marker,meta),
    legacyRequest:({prompt="",marker="",meta={},retries=0}={})=>legacyAiAnalyze(prompt,marker,retries,meta),
    retryableAiError,
  };
  globalThis.__characterCoreV2Api={protocol:PROTOCOL,version:VERSION,state:core,analysisState,parseSlots,runCharacters,runStyle,runFullAnalysis:runFullAnalysisV77,runOutline:runOutlineProductionV78,runSceneRegeneration:runSceneRegenerationProductionV78,isOutlineReady:()=>v77OutlineReadiness().ok,sourceSnapshot:()=>deepClone(captureSourceSnapshot("api_read")),revisionDiagnostics:()=>deepClone(revisionDiagnostics()),outlineUsableData:()=>deepClone(getOutlineUsableDataState()),syncRuntimeState:markV77RuntimeReady,analyzeFactsAndRelations,generateAppearance,verifyAppearanceRevision,projectLease,ensureProjectLease,saveProjectWithLease,renderCharacters,renderRelations,rebuildKeywordIndex,refreshSceneBindings,restoreSceneBindingsBySourceKey,ensureOutlineSourceSceneShells,buildSceneContext,buildV77SceneSubmissionPrompt,installOutlineAdapter,localCast,temporaryEntitiesForSource,mergeTemporaryCharacters,fallbackFormalCastFromPlan,aliasScopeActive,normalizeGender,normalizeStage,genderFromStage,coreParityHealth,characterHealth:v36CharacterHealth,characterFactsIntegrity:()=>deepClone(characterFactsIntegrity()),relationshipConflicts:()=>deepClone(v36RelationshipConflicts(core())),temporaryContinuity:()=>deepClone(core().temporary_continuity_cache||{}),temporaryContinuityMirrors:()=>({active:deepClone(core().temporary_continuity_cache||{}),character_core_validator:deepClone(core().temporary_continuity_cache_character_core_validator||{}),v78_primary:deepClone(core().temporary_continuity_cache_v78_primary||{}),authority:deepClone(core().temporary_continuity_authority_v78||{})}),stageSignature:v36StageSignature,formalRoster:()=>formalRosterSnapshot("api_read",true),formalRosterHealth:()=>formalRosterApi()?.health?.()||null};
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",init,{once:true});else init();
})();

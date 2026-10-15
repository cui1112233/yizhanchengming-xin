/* V78.3.0.31 Stable Release Authority: final public identity after final-semantics verification. */
(function installV78StableRelease(){
  const VERSION="v78.3.0.31",BUILD="v78.3.0.31-final-semantics-20260906-r1";
  if(globalThis.__V78_STABLE_RELEASE__?.version===VERSION) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const identity=Object.freeze({
    version:VERSION,
    buildId:BUILD,
    label:"V78.3.0.31 · 最终语义收口",
    title:"V20/21 导演语义审计、V22/23 事务隔离、V24-26 current-truth 清理、V27/28 单条补丁、V30 专业摄影、V31 字段边界均已纳入专用回归测试。"
  });
  function applyIdentity(){
    const prior=globalThis.__VIDEO_PROMPT_TOOL_BUILD__||{};
    globalThis.__VIDEO_PROMPT_TOOL_BUILD__={...prior,version:VERSION,buildId:BUILD,cleanCache:true,multiInstance:true};
    globalThis.__V78_CURRENT_RUNTIME__=identity;
    globalThis.__V77_CURRENT_RUNTIME__=identity;
    const applyBadge=()=>{
      const badge=document.querySelector("#v77NativeBuildBadge");
      if(!badge) return false;
      badge.textContent=identity.label;
      badge.title=identity.title;
      badge.dataset.runtimeVersion=VERSION;
      badge.dataset.runtimeBuildId=BUILD;
      return true;
    };
    globalThis.__v78ApplyRuntimeBadge=applyBadge;
    globalThis.__v77ApplyRuntimeBadge=applyBadge;
    applyBadge();
    return identity;
  }
  function checks(){
    const runtime=globalThis.__V78_CLEAN_RUNTIME__||globalThis.VideoPromptRuntime||{};
    const finalRuntime=globalThis.__V783031_RUNTIME__||{};
    const prod=ns.production?.snapshot?.()||{};
    const charAI=ns.characterAI?.snapshot?.()||{};
    const registry=globalThis.__V78_EXECUTOR_REGISTRY__||{};
    const items=[
      ["runtime_version",finalRuntime.version===VERSION||globalThis.__V78_CURRENT_RUNTIME__?.version===VERSION,`runtime=${finalRuntime.version||globalThis.__V78_CURRENT_RUNTIME__?.version||runtime.version||"?"}`],
      ["runtime_build",finalRuntime.build===BUILD||globalThis.__VIDEO_PROMPT_TOOL_BUILD__?.buildId===BUILD,`build=${finalRuntime.build||globalThis.__VIDEO_PROMPT_TOOL_BUILD__?.buildId||runtime.build||"?"}`],
      ["outline_v78_primary",prod.default_route==="v78_primary",`route=${prod.default_route||"?"}`],
      ["character_ai_v78_primary",charAI.default_route==="v78_primary",`route=${charAI.default_route||"?"}`],
      ["single_ai_request",prod.exactly_one_ai_request===true&&prod.post_send_retry===false,"post_send_retry=false"],
      ["workspace_schema",Number(ns.workspace?.schema_version||0)===40,`schema=${ns.workspace?.schema_version||0}`],
      ["formal_roster_authority",globalThis.__V78_FORMAL_ROSTER_STORE__?.version==="formal_roster_store_v78_2_8",`roster=${globalThis.__V78_FORMAL_ROSTER_STORE__?.version||"?"}`],
      ["stable_registry",registry.owner==="V78ReleaseCore",`owner=${registry.owner||"?"}`],
      ["outline_entry",typeof globalThis.generateOutline==="function","generateOutline=ready"],
      ["regenerate_entry",typeof globalThis.regenerateSceneOutline==="function","regenerateSceneOutline=ready"],
      ["v77_fallback_retained",Boolean(globalThis.__V78_OUTLINE_NATIVE_EXECUTORS__?.compatibility_only),"V77 native host=compatibility_only"],
      ["reference_entities",ns.referenceEntities?.version==="reference_entity_resolver_v78_2_4",`reference=${ns.referenceEntities?.version||"?"}`],
      ["master_prompt",ns.masterPrompt?.version==="master_prompt_service_v78_2",`master=${ns.masterPrompt?.version||"?"}`],
      ["v31_final_semantics",finalRuntime.version===VERSION,`overlay=${finalRuntime.version||"pending"}`],
    ];
    return items.map(([id,ok,detail])=>({id,status:ok?"pass":"fail",detail}));
  }
  function snapshot(){const rows=checks();return {version:VERSION,build:BUILD,channel:"stable",released_at:"2026-09-06",status:rows.every(x=>x.status==="pass")?"ready":"degraded",checks:rows,legacy:{default_executor:false,fallback_retained:true,protocol_markers_retained:true,globals_retained_as_aliases:true},workspace_schema:40};}
  const release=Object.freeze({version:VERSION,build:BUILD,channel:"stable",checks,snapshot,applyIdentity,deprecatedCompatibility:Object.freeze(["runOutlineV77","runSceneRegenerationV77","__V77_CLEAN_CORE_BRIDGE__","V77_CHARACTER_CORE_* markers"])});
  ns.release=release;
  globalThis.__V78_STABLE_RELEASE__=release;
  globalThis.VideoPromptRelease=release;
  applyIdentity();
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",()=>{applyIdentity();console.info("[V78.3.0.31 Stable] release ready",snapshot());},{once:true});else console.info("[V78.3.0.31 Stable] release ready",snapshot());
})();

/* V78.3.0.31 final-semantics runtime: loaded after the legacy stable modules, then becomes the public current-truth overlay. */
(function loadV783031CompatibilityRuntime(){
  if(globalThis.__V783031_RUNTIME__){globalThis.VideoPromptRelease?.applyIdentity?.();return;}
  if(document.querySelector('script[data-v783031-runtime="true"]')) return;
  const script=document.createElement('script');
  script.src='/novel-panel/workbench/v783031-runtime.js';
  script.async=false;
  script.dataset.v783031Runtime='true';
  script.addEventListener('load',()=>{globalThis.VideoPromptRelease?.applyIdentity?.();console.info('[V78.3.0.31] final semantics runtime active',globalThis.__V783031_RUNTIME__?.version);});
  script.addEventListener('error',()=>console.error('[V78.3.0.31] final semantics runtime failed to load'));
  document.head.appendChild(script);
})();
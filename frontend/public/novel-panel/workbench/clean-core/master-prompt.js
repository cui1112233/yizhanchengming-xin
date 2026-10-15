/* V78.2 Clean Core - Master Visual Prompt / lossless referenced view. */
(function installV782MasterPromptService(){
  if(globalThis.__V78_MASTER_PROMPT_SERVICE__)return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const txt=v=>String(v??"").trim(),arr=v=>Array.isArray(v)?v:[];
  const clone=v=>{try{return JSON.parse(JSON.stringify(v));}catch(_){return v;}};
  const ACTION_RE=/(抬|低|转|走|跑|停|坐|站|起身|俯身|抬头|低头|回头|伸手|收回|握|抓|拿|放|递|接|推|拉|打开|合上|翻开|翻页|摘下|戴上|扔|拾|举|按|敲|看向|望向|盯|扫过|靠近|后退|迈步|点头|摇头|皱眉|微笑|停顿|颤|深呼吸|擦|抱|扶|蹲|跪|转身|侧身|抬眼|垂眼)/g;
  const INTERACT_RE=/(看向|望向|对视|递给|接过|拉住|推开|扶住|抱住|挡住|靠近|远离|回应|打断|示意|跟随|面对|朝向|低声说|开口|回答)/g;
  const PROP_RE=/(拿起|放下|递出|接过|打开|合上|翻开|翻页|摘下|戴上|扔下|放到|按下|握住|松开|掏出|收起)/g;
  const CAMERA_RE=/(特写|近景|中近景|中景|中远景|全景|远景|平视|俯拍|仰拍|过肩|跟拍|推镜|推进|拉远|横移|环绕|固定|摇镜|甩镜|硬切|叠化|溶解)/g;
  function stripTokens(v=""){return String(v||"").replace(/\(@图\d+\)/g,"");}
  function count(re,s){const m=String(s||"").match(re);return m?m.length:0;}
  function metrics(v=""){
    const s=stripTokens(v),len=[...s.replace(/\s+/g,"")].length,actions=count(ACTION_RE,s),interactions=count(INTERACT_RE,s),props=count(PROP_RE,s),camera=count(CAMERA_RE,s);
    const spatial=count(/(左侧|右侧|前方|后方|窗边|门口|床边|桌边|走廊|前景|中景|背景|身侧|对面|旁边|身后|光线|灯光|窗外|阴影)/g,s);
    const richness=Math.min(10,Number((actions*0.55+interactions*0.65+props*0.45+camera*0.35+spatial*0.25+Math.min(len,220)/70).toFixed(2)));
    return {length:len,actions,interactions,prop_actions:props,camera,spatial,richness};
  }
  function hash(v=""){let h=2166136261;for(const ch of String(v)){h^=ch.charCodeAt(0);h=Math.imul(h,16777619);}return (h>>>0).toString(36);}
  function ensureSegmentMaster(segment={},baseText=""){
    const clean=stripTokens(baseText),h=hash(clean);
    if(segment.master_visual_prompt_v782_hash!==h||!txt(segment.master_visual_prompt_v782)){segment.master_visual_prompt_v782=clean;segment.master_visual_prompt_v782_hash=h;segment.master_visual_metrics_v782=metrics(clean);segment.master_visual_updated_at_v782=new Date().toISOString();}
    return segment.master_visual_prompt_v782;
  }
  function deriveReferencedPrompt(masterText="",refs=[]){
    const master=stripTokens(masterText),resolver=globalThis.__V78_REFERENCE_RESOLVER__;let result;
    try{result=resolver?.injectRefsLossless?.(master,refs);}catch(_){result=null;}
    if(!result){let text=master;try{text=resolver?.injectRefs?.(master,refs)||master;}catch(_){}result={text,master,lossless:stripTokens(text)===master,reference_count:arr(refs).length};}
    if(stripTokens(result.text)!==master)result={...result,text:master,lossless:false,fallback:"master_only"};
    const mm=metrics(master),rm=metrics(result.text),ratio=mm.length?rm.length/mm.length:1;
    const semantic={action_preserved:rm.actions>=mm.actions,interaction_preserved:rm.interactions>=mm.interactions,prop_preserved:rm.prop_actions>=mm.prop_actions,camera_preserved:rm.camera>=mm.camera,richness_ratio:mm.richness?rm.richness/mm.richness:1,length_ratio:ratio};
    const ok=result.lossless&&semantic.action_preserved&&semantic.interaction_preserved&&semantic.prop_preserved&&semantic.camera_preserved&&semantic.length_ratio>=0.92;
    return {text:result.text,master,lossless:result.lossless,ok,metrics:{master:mm,referenced:rm},semantic,reference_count:arr(refs).length,fallback:result.fallback||""};
  }
  function segmentComparison(segment={},refs=[]){const master=txt(segment.master_visual_prompt_v782||segment.current_text||segment.text),d=deriveReferencedPrompt(master,refs);return {master,referenced:d.text,...d};}
  function health(){const segs=arr((typeof state!=="undefined"&&state?.segments)||[]);const rows=segs.map((segment,index)=>{const refs=arr(segment._premium_refs);if(!refs.length)return null;const c=segmentComparison(segment,refs);return {index,ok:c.ok,lossless:c.lossless,length_ratio:c.semantic.length_ratio,richness_ratio:c.semantic.richness_ratio,actions:[c.metrics.master.actions,c.metrics.referenced.actions],interactions:[c.metrics.master.interactions,c.metrics.referenced.interactions]};}).filter(Boolean);return {ok:rows.every(x=>x.ok),segments:rows,issues:rows.filter(x=>!x.ok)};}
  const service={version:"master_prompt_service_v78_2",stripReferenceTokens:stripTokens,metrics,ensureSegmentMaster,deriveReferencedPrompt,segmentComparison,health,snapshot:()=>clone({version:"master_prompt_service_v78_2",principle:"premium_is_lossless_view_of_master",health:health()})};
  ns.masterPrompt=service;globalThis.__V78_MASTER_PROMPT_SERVICE__=service;
  console.info("[V78.2 Clean Core] Master Visual Prompt service ready",{principle:"reference_tokens_are_export_only"});
})();

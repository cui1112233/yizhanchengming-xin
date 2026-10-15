/* V78.2.8 Formal Roster Single Source of Truth.
 * The forced roster is user-authored data. No source-text inference, relation
 * vocabulary, temporary-character logic, history mirror or legacy state may
 * add/remove formal characters behind this store.
 */
(function installV7828FormalRosterStore(){
  "use strict";
  if(globalThis.__V78_FORMAL_ROSTER_STORE__) return;
  const BUILD="v78.3.0.2";
  const VERSION="formal_roster_store_v78_2_8";
  const appState=typeof state!=="undefined"?state:(globalThis.state||={});
  const listeners=new Set();
  const OPEN={"(":")","（":"）","[":"]","【":"】","〔":"〕","{":"}"};
  const CLOSE=new Set(Object.values(OPEN));
  const SEPARATORS=new Set(["\n",",","，","、",";","；","|","｜"]);
  const cleanRaw=(value="")=>String(value??"").replace(/\r\n?/g,"\n").replace(/[\uFEFF\u200B\u200C\u200D\u2060]/g,"");
  const fnv=(value="")=>{let h=2166136261;for(const c of String(value)){h^=c.charCodeAt(0);h=Math.imul(h,16777619);}return(h>>>0).toString(36);};
  function parseEntries(value=""){
    const raw=cleanRaw(value),out=[];let buf="",stack=[];
    const flush=()=>{const v=buf.replace(/^[-•·\d.、\s]+/,"").trim();if(v)out.push(v);buf="";};
    for(const ch of raw){
      if(OPEN[ch]){stack.push(OPEN[ch]);buf+=ch;continue;}
      if(CLOSE.has(ch)){if(stack.length&&stack[stack.length-1]===ch)stack.pop();buf+=ch;continue;}
      if(!stack.length&&SEPARATORS.has(ch)){flush();continue;}
      buf+=ch;
    }
    flush();
    return out;
  }
  function deepFreeze(value){if(!value||typeof value!=="object"||Object.isFrozen(value))return value;Object.freeze(value);Object.values(value).forEach(deepFreeze);return value;}
  let current={
    version:VERSION,build:BUILD,raw_text:"",normalized_text:"",entries:[],revision:0,hash:fnv(""),updated_at:"",source:"bootstrap"
  };
  function dom(){return document.querySelector?.("#characterGuideInput")||null;}
  function stateRaw(){return String(appState.character_guide_input??"");}
  function view(){return {...current,entries:[...current.entries]};}
  function render(){
    const status=document.querySelector?.("#formalRosterStatus");
    const error=document.querySelector?.("#formalRosterError");
    if(status){
      const names=current.entries.slice(0,12).map((x)=>x.replace(/\s+/g," ").trim());
      status.textContent=current.entries.length
        ? `正式人物名单：已读取 ${current.entries.length} 个条目${names.length?`｜${names.join("｜")}${current.entries.length>names.length?"｜…":""}`:""}`
        : "正式人物名单：当前为空；人物卡不会从原文、关系词或其他输入框自动补入正式人物。";
      status.dataset.count=String(current.entries.length);
      status.dataset.revision=String(current.revision);
      status.dataset.rosterHash=current.hash;
    }
    if(error&&current.entries.length) { error.hidden=true; error.textContent=""; }
  }
  function emit(previous,meta){for(const fn of [...listeners]){try{fn(view(),previous,meta||{});}catch(err){console.warn("[FORMAL_ROSTER] listener failed",err);}}}
  function commit(value,meta={}){
    const raw=cleanRaw(value),normalized=raw.trim(),entries=parseEntries(raw),hash=fnv(normalized);
    const previous=view(),changed=raw!==current.raw_text||hash!==current.hash;
    if(changed){current={version:VERSION,build:BUILD,raw_text:raw,normalized_text:normalized,entries,revision:Math.max(1,Number(current.revision||0)+1),hash,updated_at:new Date().toISOString(),source:String(meta.source||"commit")};}
    else current={...current,source:String(meta.source||current.source||"commit")};
    appState.character_guide_input=raw;
    appState.formal_roster_store_v7828={...current,entries:[...entries]};
    render();
    if(changed)emit(previous,{...meta,changed:true});
    return view();
  }
  function commitFromDom(reason="dom_commit"){
    const node=dom();return commit(node?node.value:stateRaw(),{source:"dom",reason});
  }
  function snapshot(options={}){
    const forceDom=options.forceDom!==false;
    if(forceDom)commitFromDom(options.reason||"snapshot");
    return deepFreeze({...current,entries:[...current.entries],snapshot_at:new Date().toISOString(),snapshot_reason:String(options.reason||"snapshot")});
  }
  function isFresh(snap){return Boolean(snap&&String(snap.hash||"")===current.hash&&Number(snap.revision||0)===Number(current.revision||0));}
  function assertFresh(snap,stage="transaction"){
    if(isFresh(snap))return true;
    const e=new Error(`强制人物名单已在${stage}期间发生修改；旧事务结果已丢弃，请按当前名单重新生成。`);e.code="STALE_FORMAL_ROSTER_TRANSACTION";e.snapshot_revision=Number(snap?.revision||0);e.current_revision=Number(current.revision||0);throw e;
  }
  function restore(value,meta={}){
    const raw=cleanRaw(value);const node=dom();if(node&&node.value!==raw)node.value=raw;
    return commit(raw,{source:String(meta.source||"restore"),reason:String(meta.reason||"restore")});
  }
  function bindInput(){
    const node=dom();if(!node||node.__v7828FormalRosterBound)return node;
    node.__v7828FormalRosterBound=true;
    node.addEventListener("input",()=>commit(node.value,{source:"user_input",reason:"input"}));
    node.addEventListener("change",()=>commit(node.value,{source:"user_input",reason:"change"}));
    return node;
  }
  function subscribe(fn){if(typeof fn!=="function")return()=>{};listeners.add(fn);return()=>listeners.delete(fn);}
  function health(){
    const node=dom(),domRaw=cleanRaw(node?.value??""),stateValue=cleanRaw(stateRaw());
    return {
      version:VERSION,build:BUILD,revision:current.revision,entry_count:current.entries.length,hash:current.hash,
      dom_hash:fnv(domRaw.trim()),state_hash:fnv(stateValue.trim()),dom_store_match:domRaw===current.raw_text,state_store_match:stateValue===current.raw_text,
      dom_nonempty:Boolean(domRaw.trim()),store_nonempty:Boolean(current.normalized_text),source:current.source
    };
  }
  function bootstrap(){
    const node=dom();const domValue=cleanRaw(node?.value??"");const mirror=cleanRaw(stateRaw());const restored=appState.formal_roster_store_v7828;
    let seed=domValue.trim()?domValue:(mirror.trim()?mirror:String(restored?.raw_text||""));
    if(node&&node.value!==seed)node.value=seed;
    commit(seed,{source:"bootstrap",reason:"initial_authority"});bindInput();render();
  }
  const api={version:VERSION,build:BUILD,parseEntries,commit,commitFromDom,snapshot,isFresh,assertFresh,restore,bindInput,subscribe,health,state:()=>view(),render};
  globalThis.__V78_FORMAL_ROSTER_STORE__=api;
  globalThis.__V78_FORMAL_ROSTER__=api;
  if(document.readyState==="loading")document.addEventListener("DOMContentLoaded",bootstrap,{once:true});else bootstrap();
  console.info("[v78.3.0.2] Formal Roster Single Source of Truth ready",health());
})();

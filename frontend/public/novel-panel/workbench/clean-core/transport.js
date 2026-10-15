/* V78 Clean Core Phase 1: stable frontend AI/API transport boundary. */
(function installV78CleanTransport(){
  if(globalThis.__V78_CLEAN_TRANSPORT__) return;
  const ns=globalThis.VideoPromptCleanCore=globalThis.VideoPromptCleanCore||{};
  const base=globalThis.requestJSON;
  const txt=(v)=>String(v??"").trim();
  const perf=()=>globalThis.performance?.now?.()||Date.now();
  const stats={version:"clean_transport_v78_phase1",total:0,success:0,error:0,cancelled:0,last_ms:0,last_endpoint:"",history:[]};
  function endpoint(url=""){try{return new URL(String(url),location.href).pathname;}catch(_){return String(url||"").split("?")[0];}}
  function remember(row){stats.history.push(row);while(stats.history.length>80)stats.history.shift();}
  function safeOptions(options={}){const out={...options};delete out.cleanCoreBypass;return out;}
  async function request(url,options={}){
    if(typeof base!=="function") throw new Error("Clean Core Transport 未找到兼容 requestJSON。请使用完整软件包启动。");
    if(options?.cleanCoreBypass) return base(url,safeOptions(options));
    const path=endpoint(url),started=perf();stats.total++;stats.last_endpoint=path;
    const headers={...(options?.headers||{}),"X-V78-Clean-Core":"transport_phase1"};
    try{
      const data=await base(url,{...safeOptions(options),headers});
      const ms=Math.max(0,perf()-started);stats.success++;stats.last_ms=Math.round(ms*10)/10;remember({at:new Date().toISOString(),endpoint:path,status:"success",ms:stats.last_ms});return data;
    }catch(error){
      const ms=Math.max(0,perf()-started),cancelled=error?.name==="AbortError"||/cancel|abort|取消/i.test(txt(error?.message));stats[cancelled?"cancelled":"error"]++;stats.last_ms=Math.round(ms*10)/10;remember({at:new Date().toISOString(),endpoint:path,status:cancelled?"cancelled":"error",ms:stats.last_ms,error:txt(error?.message).slice(0,300)});throw error;
    }
  }
  request.__v78CleanTransport=true;request.__v78Base=base;
  const api={
    version:"clean_transport_v78_phase1",phase:"v78_phase1",ready:typeof base==="function",request,
    get:(url,options={})=>request(url,{...options,method:"GET"}),
    post:(url,body,options={})=>request(url,{...options,method:"POST",headers:{"Content-Type":"application/json",...(options.headers||{})},body:typeof body==="string"?body:JSON.stringify(body??{})}),
    snapshot:()=>JSON.parse(JSON.stringify(stats)),
    legacy:()=>base
  };
  ns.transport=api;globalThis.__V78_CLEAN_TRANSPORT__=api;
  if(typeof base==="function"){
    globalThis.requestJSON=request;
    try{requestJSON=request;}catch(_){/* global lexical binding may be read-only in isolated harnesses */}
  }
  console.info("[V78 Clean Core] transport ready",api.snapshot());
})();

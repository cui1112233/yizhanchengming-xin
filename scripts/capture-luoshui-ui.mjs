// User-authorized CDP UI collection. No application APIs or model calls.
import fs from 'node:fs/promises';
import path from 'node:path';
import crypto from 'node:crypto';
const [command='state',...args]=process.argv.slice(2);
const out=path.resolve('docs/reference/luoshui-ui-v2.9.5-20261121');
const endpoint='http://127.0.0.1:9226';
const targets=await(await fetch(endpoint+'/json/list')).json();
const target=targets.find(t=>t.type==='page'&&/^http:\/\/127\.0\.0\.1:\d+\/index\.html/.test(t.url));
if(!target)throw Error('Original application page missing');
const socket=new WebSocket(target.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{socket.addEventListener('open',resolve,{once:true});socket.addEventListener('error',reject,{once:true});});
let sequence=0;const pending=new Map();
socket.addEventListener('message',event=>{const m=JSON.parse(event.data);if(pending.has(m.id)){const p=pending.get(m.id);pending.delete(m.id);clearTimeout(p.timer);m.error?p.reject(Error(m.error.message)):p.resolve(m.result);}});
function cdp(method,params={}){return new Promise((resolve,reject)=>{const id=++sequence,timer=setTimeout(()=>{pending.delete(id);reject(Error('CDP timeout: '+method));},15000);pending.set(id,{resolve,reject,timer});socket.send(JSON.stringify({id,method,params}));});}
async function evaluate(expression){const r=await cdp('Runtime.evaluate',{expression,returnByValue:true,awaitPromise:true});if(r.exceptionDetails)throw Error(r.exceptionDetails.exception?.description || r.exceptionDetails.text);return r.result.value;}
const stateExpression=`(()=>{
 const visible=e=>{const r=e.getBoundingClientRect(),s=getComputedStyle(e);return r.width>0&&r.height>0&&s.display!=='none'&&s.visibility!=='hidden'};
 const selectors='button,[role="button"],[role="menuitem"],[role="tab"],a,summary,input:not([type="password"]),textarea,select,[role="combobox"],.n-base-selection,.n-tabs-tab';
 const controls=[...document.querySelectorAll(selectors)].filter(visible).map((e,index)=>{const r=e.getBoundingClientRect();return {index,tag:e.tagName,role:e.getAttribute('role')||'',label:(e.getAttribute('aria-label')||e.getAttribute('title')||e.textContent||e.getAttribute('placeholder')||'').trim().replace(/\\s+/g,' ').slice(0,150),disabled:e.disabled===true,rect:{x:r.x,y:r.y,width:r.width,height:r.height}}});
 return {url:location.href,title:document.title,viewport:{width:innerWidth,height:innerHeight},text:document.body.innerText.slice(0,24000),controls};
})()`;
try{
 if(command==='restore') {const w=await cdp('Browser.getWindowForTarget',{targetId:target.id});await cdp('Browser.setWindowBounds',{windowId:w.windowId,bounds:{windowState:'normal'}});console.log('WebView bounds set normal; native host may remain minimized');}
 else if(command==='state'){const s=await evaluate(stateExpression);console.log(JSON.stringify({url:s.url,text:s.text.slice(0,2400),controls:s.controls.map(c=>({index:c.index,role:c.role,label:c.label,disabled:c.disabled}))},null,2));}
 else if(command==='click'){
  const index=Number(args[0]),expected=args[1];if(!Number.isInteger(index)||!expected)throw Error('click requires index and exact observed label');
  if(!expected.startsWith('去使用') && /删除|立即重启|退出登录|保存设置|开始仿写|开始生成|开始分析|确认生成|一键生成|发布|支付/.test(expected))throw Error('Action outside UI collection scope');
  const r=await evaluate(`(()=>{const s=${stateExpression};const c=s.controls[${index}];if(!c||c.label!==${JSON.stringify(expected)})throw Error('Control changed; re-observe');const visible=e=>{const r=e.getBoundingClientRect(),s=getComputedStyle(e);return r.width>0&&r.height>0&&s.display!=='none'&&s.visibility!=='hidden'};const e=[...document.querySelectorAll('button,[role="button"],[role="menuitem"],[role="tab"],a,summary,input:not([type="password"]),textarea,select,[role="combobox"],.n-base-selection,.n-tabs-tab')].filter(visible)[${index}];e.scrollIntoView({block:"nearest",inline:"nearest"});const r=e.getBoundingClientRect();return {clicked:c.label,x:r.x+r.width/2,y:r.y+r.height/2};})()`);await cdp("Input.dispatchMouseEvent",{type:"mousePressed",x:r.x,y:r.y,button:"left",clickCount:1});await cdp("Input.dispatchMouseEvent",{type:"mouseReleased",x:r.x,y:r.y,button:"left",clickCount:1});console.log(JSON.stringify({clicked:r.clicked}));
 }
 else if(command==='key'){
  if(args[0]!=='Escape')throw Error('Only Escape allowed');await cdp('Input.dispatchKeyEvent',{type:'keyDown',key:'Escape',code:'Escape',windowsVirtualKeyCode:27});await cdp('Input.dispatchKeyEvent',{type:'keyUp',key:'Escape',code:'Escape',windowsVirtualKeyCode:27});console.log('Escape');
 }
 else if(command==='text'){
  const expected=args[0];if(!expected||/删除|保存|开始生成|立即重启|支付/.test(expected))throw Error('Text action outside collection scope');
  const r=await evaluate(`(()=>{const all=[...document.querySelectorAll('*')].filter(e=>e.children.length===0&&e.textContent.trim()===${JSON.stringify(expected)}&&e.getBoundingClientRect().width>0&&e.getBoundingClientRect().height>0);if(all.length!==1)throw Error('Expected one text target, found '+all.length);const e=all[0];e.scrollIntoView({block:'nearest'});const r=e.getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2};})()`);
  await cdp('Input.dispatchMouseEvent',{type:'mousePressed',...r,button:'left',clickCount:1});await cdp('Input.dispatchMouseEvent',{type:'mouseReleased',...r,button:'left',clickCount:1});console.log(JSON.stringify({clickedText:expected}));
 }
 else if(command==='scrollables'){
  console.log(JSON.stringify(await evaluate(`(()=>{const all=[...document.querySelectorAll('*')].filter(e=>e.clientWidth>200&&e.clientHeight>80&&e.scrollHeight>e.clientHeight+60&&/auto|scroll/.test(getComputedStyle(e).overflowY));return all.map((e,index)=>({index,tag:e.tagName,className:String(e.className).slice(0,120),height:e.clientHeight,scrollHeight:e.scrollHeight,top:e.scrollTop}));})()`),null,2));
 }
 else if(command==='scroll'){
  console.log(JSON.stringify(await evaluate(`(()=>{const all=[...document.querySelectorAll('*')].filter(e=>e.clientWidth>200&&e.clientHeight>80&&e.scrollHeight>e.clientHeight+60&&/auto|scroll/.test(getComputedStyle(e).overflowY));const e=all[${Number(args[0])}];if(!e)throw Error('Scroller changed');e.scrollTop=${Number(args[1])};return {top:e.scrollTop};})()`)));
 }
 else if(command==='capture'){
  const [slug,title,group='其他',entry='']=args;if(!/^[a-z0-9-]+$/.test(slug||''))throw Error('Invalid slug');
  await fs.mkdir(path.join(out,'screenshots'),{recursive:true});await fs.mkdir(path.join(out,'states'),{recursive:true});
  let prior=-1, stable=0;for(let i=0;i<14&&stable<4;i++){const n=await evaluate('document.body.innerText.length');stable=n===prior?stable+1:0;prior=n;await new Promise(r=>setTimeout(r,180));}
  const state=await evaluate(stateExpression);
  // No field values are collected. Hide password/key inputs from screenshots.
  await evaluate(`(()=>{const e=document.createElement('style');e.id='codex-capture-mask';e.textContent='input[type=password],input[placeholder*=密钥],input[placeholder*=Key],input[placeholder*=Token],input[placeholder*=token],input[placeholder*=API]{-webkit-text-security:disc!important}';document.head.append(e);return true})()`);
  let shot;try{try{shot=await cdp('Page.captureScreenshot',{format:'png',captureBeyondViewport:false,fromSurface:true});}catch(e){if(!String(e.message).includes('Page.captureScreenshot'))throw e;shot=await cdp('Page.captureScreenshot',{format:'png',captureBeyondViewport:false,fromSurface:false});}}finally{await evaluate("document.getElementById('codex-capture-mask')?.remove();true");}
  const bytes=Buffer.from(shot.data,'base64');await fs.writeFile(path.join(out,'screenshots',slug+'.png'),bytes);
  await fs.writeFile(path.join(out,'states',slug+'.json'),JSON.stringify(state,null,2));
  let manifest=[];try{manifest=JSON.parse(await fs.readFile(path.join(out,'manifest.json'),'utf8'));}catch(e){if(e.code!=='ENOENT')throw e;}
  const record={id:slug,title,group,entry,version:'2.9.5',capturedAt:new Date().toISOString(),targetId:target.id,url:state.url,viewport:state.viewport,screenshot:'screenshots/'+slug+'.png',state:'states/'+slug+'.json',sha256:crypto.createHash('sha256').update(bytes).digest('hex'),controls:state.controls.length,status:'captured',businessVerified:false};
  manifest=manifest.filter(x=>x.id!==slug);manifest.push(record);await fs.writeFile(path.join(out,'manifest.json'),JSON.stringify(manifest,null,2));
  console.log(JSON.stringify({saved:record.id,title,bytes:bytes.length,controls:record.controls}));
 }
 else throw Error('Unknown command');
}finally{socket.close();setTimeout(()=>process.exit(process.exitCode??0),100);}

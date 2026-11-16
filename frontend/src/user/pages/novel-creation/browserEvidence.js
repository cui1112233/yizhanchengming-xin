// Based on the existing frontend test-browser-dom harness; adds fixed-size screenshot and fake DTO evidence.
import assert from 'node:assert/strict';
import {existsSync,readFileSync,writeFileSync} from 'node:fs';
import path from 'node:path';
import {pathToFileURL} from 'node:url';
import {spawn} from 'node:child_process';

// Attach while blank, then navigate. Loading a large React page before attachment
// can starve the first inspection and turn a passing UI into a harness timeout.
export async function runNovelBrowserDOM(browser,html,dir){
 const child=spawn(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--remote-debugging-port=0','--user-data-dir='+path.join(dir,'profile'),'about:blank'],{windowsHide:true,stdio:'ignore'});
 const protocol=[];let socket;const pending=new Map();let id=0;
 try{
  const portFile=path.join(dir,'profile','DevToolsActivePort');let port;
  for(let n=0;n<100;n++){if(existsSync(portFile)){try{port=Number(readFileSync(portFile,'utf8').split('\n')[0]);if(port)break;}catch(e){if(e.code!=='EBUSY')throw e;}}if(child.exitCode!==null)throw Error('test browser exited before inspection');await new Promise(r=>setTimeout(r,100));}
  assert.ok(port,'test browser did not expose its inspection port');
  let page;for(let n=0;n<50;n++){try{page=(await(await fetch('http://127.0.0.1:'+port+'/json/list')).json()).find(p=>p.type==='page');if(page)break;}catch{}await new Promise(r=>setTimeout(r,100));}
  assert.ok(page,'test page unavailable');socket=new WebSocket(page.webSocketDebuggerUrl);await new Promise((resolve,reject)=>{socket.addEventListener('open',resolve,{once:true});socket.addEventListener('error',reject,{once:true});});
  socket.addEventListener('message',event=>{const data=JSON.parse(event.data);protocol.push({received:data.id,method:data.method,error:data.error,exception:data.params?.exceptionDetails?.text,description:data.params?.exceptionDetails?.exception?.description,args:data.method==='Runtime.consoleAPICalled'?data.params?.args?.map(x=>x.value).slice(0,2):undefined});if(data.id&&pending.has(data.id)){const entry=pending.get(data.id);pending.delete(data.id);clearTimeout(entry.timer);data.error?entry.reject(Error(data.error.message)):entry.resolve(data.result);}});
  const send=(method,params={})=>new Promise((resolve,reject)=>{const key=++id;const timer=setTimeout(()=>{pending.delete(key);reject(Error('test page inspection timed out'));},50000);pending.set(key,{resolve,reject,timer});protocol.push({sent:key,method,expression:params.expression?.slice(0,100)});socket.send(JSON.stringify({id:key,method,params}));});
  await send('Page.enable');await send('Runtime.enable');await send('Emulation.setDeviceMetricsOverride',{width:1440,height:900,deviceScaleFactor:1,mobile:false});await send('Page.navigate',{url:pathToFileURL(html).href});let state='';const deadline=Date.now()+90000;
  while(Date.now()<deadline){const result=await send('Runtime.evaluate',{expression:'document.body?.dataset.result || ""',returnByValue:true});state=result.result?.value||'';if(state)break;await new Promise(r=>setTimeout(r,100));}
  const snapshot=await send('Runtime.evaluate',{expression:'window.__codexDOMSnapshot=document.documentElement.outerHTML;window.__codexDOMSnapshot.length',returnByValue:true});let output='';const size=snapshot.result?.value||0;for(let offset=0;offset<size;offset+=8192){const part=await send('Runtime.evaluate',{expression:'window.__codexDOMSnapshot.slice('+offset+','+(offset+8192)+')',returnByValue:true});output+=part.result?.value||'';}writeFileSync(path.join(dir,'result.html'),output);
  const detail=await send('Runtime.evaluate',{expression:'(document.body?.dataset.error || "browser did not finish").slice(0,4000)',returnByValue:true});
  await new Promise(r=>setTimeout(r,450));const shot=await send('Page.captureScreenshot',{format:'png',captureBeyondViewport:false});writeFileSync(path.join(dir,'screenshot.png'),Buffer.from(shot.data,'base64'));const calls=await send('Runtime.evaluate',{expression:'JSON.stringify(window.calls||[])',returnByValue:true});writeFileSync(path.join(dir,'requests.json'),calls.result?.value||'[]');assert.equal(state,'passed',detail.result?.value);return output;
 }finally{writeFileSync(path.join(dir,'protocol.json'),JSON.stringify(protocol,null,2));for(const entry of pending.values()){clearTimeout(entry.timer);entry.reject(Error('browser closed'));}if(socket?.readyState===1){socket.send(JSON.stringify({id:++id,method:'Browser.close'}));await new Promise(r=>setTimeout(r,100));}socket?.close();if(child.exitCode===null)child.kill();}
}

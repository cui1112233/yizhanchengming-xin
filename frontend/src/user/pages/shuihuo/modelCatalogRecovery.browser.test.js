import test from 'node:test';
import assert from 'node:assert/strict';
import {build} from 'esbuild';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
const browser='C:/Program Files/Google/Chrome/Application/chrome.exe';
test('video model catalog retries failures and removes the warning without ASR requests',{skip:!existsSync(browser)},async()=>{
 const dir=mkdtempSync(join(tmpdir(),'model-recovery-'));
 try {
  const output=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
   import React from 'react';
   import {createRoot} from 'react-dom/client';
   import {BatchFactoryEngineSettingsForm} from './BatchFactoryUnifiedSettingsModal.jsx';
   const calls=[];let retried=false;
   window.fetch=async url=>{
    const kind=new URL(url,'http://localhost').searchParams.get('kind');
    if(String(url).startsWith('/api/models?')){
     calls.push(kind);
     if(kind==='video'&&!retried)return new Response(JSON.stringify({error:'测试服务异常'}),{status:503});
     return new Response(JSON.stringify({models:[{id:kind,kind,displayName:'已恢复'+kind,provider:'personal_api'}]}),{status:200});
    }
    return new Response('{}',{status:200});
   };
   createRoot(document.getElementById('root')).render(<BatchFactoryEngineSettingsForm value={{productionMode:'video'}} onChange={()=>{}} sections={['models']}/>);
   const pause=()=>new Promise(r=>setTimeout(r,80));
   async function waitFor(fn){for(let i=0;i<70;i++){if(fn())return;await pause()}throw Error('等待界面超时')}
   async function run(){
    await waitFor(()=>document.body.textContent.includes('视频：服务暂时不可用'));
    if(calls.includes('asr'))throw Error('视频配置不应读取语音识别目录');
    retried=true;
    [...document.querySelectorAll('button')].find(b=>b.textContent==='重新读取').click();
    await waitFor(()=>!document.body.textContent.includes('服务暂时不可用'));
    if(calls.filter(k=>k==='video').length!==2||calls.filter(k=>k==='text').length!==1||calls.filter(k=>k==='image').length!==1)throw Error('重试重复读取成功目录');
    document.body.dataset.result='passed';
   }
   run().catch(e=>document.body.dataset.error=e.message);
  `},bundle:true,jsx:'automatic',format:'iife',write:false,minify:true,loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'}});
  const html=join(dir,'index.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script>'+output.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
  const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=12000','--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:45000,windowsHide:true,stdio:['ignore','pipe','pipe']});
  assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1]||'recovery not completed');
 }finally{rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150});}
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync,mkdirSync,mkdtempSync,writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath,pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';
const here=path.dirname(fileURLToPath(import.meta.url));
const browser=[process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file=>file&&existsSync(file));
const archive=path.resolve(here,'../../../../_local-archive-20261028/luoshui-screenplay-contract-20261114/ui-reconstruction');

async function runCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1200,900','--virtual-time-budget=16000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('版本择优弹窗：横排对照 + 选用此版 + 删除该版 + 当前版本受限', {skip:!browser},async()=>{
 await runCase('version-compare',`
import React from 'react';import{createRoot}from'react-dom/client';
import ComicScreenplayVersionCompare from'./ComicScreenplayVersionCompare.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
const documents=[
 {id:'doc-1',version:1,status:'superseded',created_at:'2026-11-19T01:00:00Z',updated_at:'2026-11-19T01:00:00Z'},
 {id:'doc-2',version:2,status:'active',created_at:'2026-11-19T01:10:00Z',updated_at:'2026-11-19T01:10:00Z'}
];
const contents={'doc-1':'林川走进巷口。','doc-2':'林川走进巷口，雨下起来了。'};
const calls=[];
window.fetch=async(input,opts={})=>{const url=String(input);const method=opts.method||'GET';await new Promise(resolve=>setTimeout(resolve,20));calls.push(method+' '+url);
 if(method==='GET'&&url.includes('/documents/doc-1'))return new Response(JSON.stringify({document:{...documents[0],content:contents['doc-1']}}),{status:200,headers:{'Content-Type':'application/json'}});
 if(method==='GET'&&url.includes('/documents/doc-2'))return new Response(JSON.stringify({document:{...documents[1],content:contents['doc-2']}}),{status:200,headers:{'Content-Type':'application/json'}});
 if(method==='POST'&&url.includes('/restore'))return new Response(JSON.stringify({document:{}}),{status:200,headers:{'Content-Type':'application/json'}});
 if(method==='DELETE'&&url.includes('/documents/doc-1'))return new Response('',{status:204});
 return new Response('{}',{status:404});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>(el.textContent.replace(/\\s/g,'')===text));
(async()=>{
 let changed=0,closed=false;
 const root=createRoot(document.getElementById('root'));
 root.render(<ComicScreenplayVersionCompare projectId="p" documents={documents} readOnly={false} open={true} onClose={()=>{closed=true;}} onChanged={async()=>{changed+=1;}}/>);
 for(let i=0;i<60&&!document.body.textContent.includes('版本择优');i++)await tick();
 for(let i=0;i<60&&!document.body.textContent.includes('雨下起来了');i++)await tick();
 assert(document.body.textContent.includes('按版本横排对照，选用更好的一稿'),'modal subtitle missing');
 assert(document.body.textContent.includes('版本 v1')&&document.body.textContent.includes('版本 v2'),'version columns missing');
 assert(document.body.textContent.includes('当前版本'),'active tag missing');
 assert(document.body.textContent.includes('林川走进巷口，雨下起来了。'),'doc2 content missing');
 assert(document.body.textContent.includes('7字')&&document.body.textContent.includes('13字'),'word count missing: '+document.body.textContent);
 const delBtn=button('删除该版');assert(delBtn,'delete button missing');
 delBtn.click();
 let confirm=null;
 for(let i=0;i<40&&!confirm;i++)await tick();
 confirm=[...document.querySelectorAll('.ant-popconfirm button')].find(el=>el.classList.contains('ant-btn-primary'));
 assert(confirm,'popconfirm confirm missing');confirm.click();
 for(let i=0;i<60&&!document.body.textContent.includes('已删除：版本 v1');i++)await tick();
 assert(calls.some(c=>c.startsWith('DELETE /')&&c.includes('/documents/doc-1')),'delete call missing: '+calls.join(';'));
 const adopt=[...document.querySelectorAll('.wsvp-col')].map(col=>({col,btn:[...col.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')==='选用此版')})).find(x=>x.btn);
 assert(adopt&&adopt.btn&&!adopt.btn.disabled,'adopt button for v1 missing/disabled');
 adopt.btn.click();
 for(let i=0;i<60&&!document.body.textContent.includes('已选用：版本 v1');i++)await tick();
 assert(calls.some(c=>c.startsWith('POST /')&&c.includes('/restore')),'restore call missing');
 assert(changed>=1,'onChanged not fired');
 // 当前版本列：选用/删除按钮禁用且文案为「正在使用」
 const activeCol=[...document.querySelectorAll('.wsvp-col.preferred')];assert(activeCol.length===1,'preferred col missing');
 const activeBtns=[...activeCol[0].querySelectorAll('button')];
 assert(activeBtns.some(el=>el.textContent.replace(/\\s/g,'')==='正在使用'),'active adopt label missing');
 assert(activeBtns.filter(el=>el.disabled).length>=2,'active col buttons should be disabled');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

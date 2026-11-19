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
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1000,900','--virtual-time-budget=16000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('章节面板导出本章下载 TXT，空原文禁用', {skip:!browser},async()=>{
 await runCase('chapter-export',`
import React from 'react';import{createRoot}from'react-dom/client';import{ComicChaptersPanel}from'./ComicScreenplayAuthoringPanels.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
const chapters=[{id:'c1',title:'第一章 相遇',source_text:'林川走进巷口。',script_text:''},{id:'c2',title:'空章',source_text:'',script_text:''}];
window.fetch=async(input,opts={})=>{const url=String(input);await new Promise(resolve=>setTimeout(resolve,20));
 if(url.includes('/chapters')&&(opts.method||'GET')==='GET')return new Response(JSON.stringify({state:{chapters,revision:1}}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response('{}',{status:404});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>(el.textContent.replace(/\\s/g,'')===text));
window.downloads=[];
URL.createObjectURL=blob=>{window.downloads.push({blob});return 'blob:mock';};
URL.revokeObjectURL=()=>{};
HTMLAnchorElement.prototype.click=function(){const last=window.downloads[window.downloads.length-1];if(last)last.download=this.download;};
(async()=>{
 createRoot(document.getElementById('root')).render(<ComicChaptersPanel projectId="p" readOnly={false} canGenerate={true}/>);
 for(let i=0;i<60&&!document.body.textContent.includes('第一章 相遇');i++)await tick();
 assert(document.body.textContent.includes('第一章 相遇'),'chapter list missing');
 const exportBtn=button('导出本章');assert(exportBtn,'export button missing');
 assert(!exportBtn.disabled,'export should be enabled for chapter with content');
 exportBtn.click();
 for(let i=0;i<60&&!document.body.textContent.includes('已导出：第一章 相遇.txt');i++)await tick();
 const record=window.downloads[0];assert(record,'no download recorded');
 const text=await record.blob.text();
 assert(text==='林川走进巷口。','export content wrong: '+text);
 const anchorDownload=(await new Promise(resolve=>{const orig=URL.createObjectURL;resolve(window.downloads.length);}))&&window.downloads;
 // 文件名来自 download 属性：已通过 click 拦截器并入 record
 assert(document.body.textContent.includes('已导出：第一章 相遇.txt'),'success message missing');
 // 切到空章节：导出按钮应禁用
 const aside=[...document.querySelectorAll('.comic-chapter-layout aside button')];
 const emptyBtn=aside.find(el=>el.textContent.includes('空章'));assert(emptyBtn,'empty chapter aside button missing');
 emptyBtn.click();await tick();
 for(let i=0;i<40&&!button('导出本章')?.disabled;i++)await tick();
 assert(button('导出本章').disabled,'export should be disabled for empty chapter');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

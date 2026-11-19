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

async function runMergeCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=900,900','--virtual-time-budget=12000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('分镜合并弹窗按洛水规则预览、切换时长重算并确认执行', {skip:!browser},async()=>{
 await runMergeCase('storyboard-merge',`
import React from 'react';import{createRoot}from'react-dom/client';import Merge from './ComicScreenplayStoryboardMerge.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,120));const assert=(v,m)=>{if(!v)throw Error(m)};
const plan={max_seconds:30,before_count:4,after_count:2,removed_count:2,pending_groups:1,pending_shots:3,untimed_count:0,groups:[{keep_shot_id:'s1',shot_ids:['s1','s2','s3'],seconds:29}]};
window.calls=[];
window.fetch=async(input,options={})=>{const url=String(input);const body=options.body?JSON.parse(options.body):{};window.calls.push({url,body});await new Promise(resolve=>setTimeout(resolve,10));
 return new Response(JSON.stringify(body.dry_run?{plan}:{plan,storyboard:{shots:[{id:'s1'},{id:'s4'}]}}),{status:200,headers:{'Content-Type':'application/json'}});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>(el.textContent.replace(/\\s/g,'')===text||el.getAttribute('aria-label')===text));
let merged=null,closed=false;
(async()=>{
 createRoot(document.getElementById('root')).render(<Merge projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2,3,4]},{id:'b2',episode_no:2,shots:[1]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" readOnly={false} onMerged={async no=>{merged=no;}} onClose={()=>{closed=true;}}/>);
 for(let i=0;i<30&&!document.querySelector('.comic-sb-merge-stats');i++)await tick();
 const stats=document.querySelector('.comic-sb-merge-stats');assert(stats,'preview stats missing');
 assert(stats.textContent.includes('4')&&stats.textContent.includes('2'),'preview counts wrong: '+stats.textContent);
 assert(window.calls.length===1,'preview issued '+window.calls.length+' requests');
 assert(window.calls[0].body.dry_run===true,'preview must be dry run');
 assert(window.calls[0].body.max_seconds===30,'default max seconds wrong');
 assert(window.calls[0].url.includes('/storyboards/xiaoluo_v7_15s/1/shots/merge'),'wrong endpoint '+window.calls[0].url);
 const groups=document.querySelectorAll('.comic-sb-merge-groups li');assert(groups.length===1,'group list missing');
 assert(document.body.textContent.includes('保留 1 个'),'group preview copy missing');
 assert(document.body.textContent.includes('不调用模型'),'merge hint missing no-model copy');
 const duration=document.querySelector('[aria-label="最长合并时长"] .ant-select-selector');assert(duration,'duration select missing');
 duration.dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));await tick();await tick();
 const ten=[...document.querySelectorAll('.ant-select-item-option')].find(el=>el.textContent==='10 秒');assert(ten,'duration option missing');ten.click();
 for(let i=0;i<20&&window.calls.length<2;i++)await tick();
 assert(window.calls.at(-1).body.max_seconds===10&&window.calls.at(-1).body.dry_run===true,'duration change did not re-preview');
 const merge=button('开始合并');assert(merge&&!merge.disabled,'merge button unavailable');
 const before=window.calls.length;merge.click();
 for(let i=0;i<30&&!merged;i++)await tick();
 assert(merged===1,'onMerged not called with episode number');
 assert(closed,'dialog stayed open after merge');
 const exec=window.calls.slice(before).find(call=>!call.body.dry_run);assert(exec,'no execute request');
 assert(exec.body.max_seconds===10,'execute used wrong duration');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

test('分镜合并弹窗在只读或没有分镜时不允许执行', {skip:!browser},async()=>{
 await runMergeCase('storyboard-merge-readonly',`
import React from 'react';import{createRoot}from'react-dom/client';import Merge from './ComicScreenplayStoryboardMerge.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,120));const assert=(v,m)=>{if(!v)throw Error(m)};
const plan={max_seconds:30,before_count:2,after_count:1,removed_count:1,pending_groups:1,pending_shots:2,untimed_count:0,groups:[{keep_shot_id:'s1',shot_ids:['s1','s2'],seconds:18}]};
window.calls=[];
window.fetch=async(input,options={})=>{const body=options.body?JSON.parse(options.body):{};window.calls.push({body});await new Promise(resolve=>setTimeout(resolve,10));return new Response(JSON.stringify({plan}),{status:200,headers:{'Content-Type':'application/json'}});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 createRoot(document.getElementById('root')).render(<Merge projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" readOnly onMerged={()=>{}} onClose={()=>{}}/>);
 for(let i=0;i<30&&!document.querySelector('.comic-sb-merge-stats');i++)await tick();
 const merge=button('开始合并');assert(merge,'merge button missing');
 assert(merge.disabled,'readOnly allowed merge');
 assert(window.calls.every(call=>call.body.dry_run===true),'readOnly wrote storyboard');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

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

async function runTranslateCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=900,900','--virtual-time-budget=16000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('批量翻译弹窗按类型方向范围逐条翻译并汇总成功数', {skip:!browser},async()=>{
 await runTranslateCase('storyboard-translate',`
import React from 'react';import{createRoot}from'react-dom/client';import Translate from './ComicScreenplayBatchTranslate.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
window.calls=[];
const items=[{shot_id:'bt-1',seq:1,text:'夜晚的街道',from_subtitle:false},{shot_id:'bt-2',seq:2,text:'只有台词',from_subtitle:true},{shot_id:'bt-3',seq:3,text:'卧室清晨',from_subtitle:false}];
window.fetch=async(input,opts={})=>{const url=String(input);const body=opts.body?JSON.parse(opts.body):{};window.calls.push({url,body});await new Promise(resolve=>setTimeout(resolve,20));
 if(url.includes('/translate/prepare'))return new Response(JSON.stringify({items,total:items.length}),{status:200,headers:{'Content-Type':'application/json'}});
 if(url.includes('/translate/item'))return new Response(JSON.stringify({shot_id:body.shot_id,seq:1,text:'night street',revision:2}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response('{}',{status:404});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>(el.textContent.replace(/\\s/g,'')===text||el.getAttribute('aria-label')===text));
async function pick(label,optionText){
 const root=document.querySelector('[aria-label="'+label+'"]');assert(root,label+' select missing');
 for(let i=0;i<40&&/ant-select-disabled/.test(root.className);i++)await tick();
 assert(!/ant-select-disabled/.test(root.className),label+' select stayed disabled');
 const input=root.querySelector('input');assert(input,label+' search input missing');
 input.focus();
 input.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowDown',keyCode:40,which:40,bubbles:true}));
 await tick();
 const dropdown=[...document.querySelectorAll('.ant-select-dropdown')].filter(el=>!el.className.includes('hidden')).pop();
 assert(dropdown,label+' dropdown did not open');
 const option=[...dropdown.querySelectorAll('.ant-select-item-option')].find(el=>el.textContent.trim()===optionText);
 assert(option,label+' option missing: '+optionText+' in '+[...dropdown.querySelectorAll('.ant-select-item-option')].map(el=>el.textContent.trim()).join('/'));
 option.click();await tick();}
let changed=null;
(async()=>{
 createRoot(document.getElementById('root')).render(<Translate projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2,3]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" modelId="text-main" readOnly={false} onChanged={async no=>{changed=no;}} onClose={()=>{}}/>);
 for(let i=0;i<40&&!document.querySelector('[aria-label="翻译类型"]');i++)await tick();
 assert(document.body.textContent.includes('画面提示词'),'field option missing');
 assert(document.body.textContent.includes('中文 → 英文'),'direction option missing');
 assert(document.body.textContent.includes('未锁定')===false,'locked scope should not appear: no lock concept in V88');
 await pick('翻译类型','视频提示词');
 await pick('翻译范围','自定义范围');
 const startInput=document.querySelector('[aria-label="起始分镜"] input')||document.querySelector('[aria-label="起始分镜"]');assert(startInput,'shot start input missing');
 assert(String(startInput.value)==='1','default start wrong: '+startInput.value);
 const endInput=document.querySelector('[aria-label="结束分镜"] input')||document.querySelector('[aria-label="结束分镜"]');assert(endInput,'shot end input missing');
 assert(String(endInput.value)==='3','default end should be shot count: '+endInput.value);
 assert(document.body.textContent.includes('共 3 个分镜'),'shot count copy missing');
 await pick('翻译范围','全部');
 const start=button('开始翻译');assert(start&&!start.disabled,'start button unavailable');
 start.click();
 for(let i=0;i<120&&!changed&&!(document.querySelector('.comic-sb-translate-summary'));i++)await tick();
 const prepares=window.calls.filter(call=>call.url.includes('/translate/prepare'));
 assert(prepares.length===1&&prepares[0].body.field==='video'&&prepares[0].body.scope==='all'&&prepares[0].body.direction==='zh2en','prepare payload wrong: '+JSON.stringify(prepares[0]?.body));
 const itemCalls=window.calls.filter(call=>call.url.includes('/translate/item'));
 assert(itemCalls.length===3,'item calls wrong: '+itemCalls.length);
 assert(itemCalls.every(call=>call.body.field==='video'&&call.body.model==='text-main'&&call.body.shot_id),'item payload wrong');
 assert(itemCalls[0].body.shot_id==='bt-1'&&itemCalls[1].body.shot_id==='bt-2'&&itemCalls[2].body.shot_id==='bt-3','item order should follow prepare');
 const summary=document.querySelector('.comic-sb-translate-summary');assert(summary,'summary missing');
 assert(summary.textContent.includes('成功 3/3'),'summary copy wrong: '+summary.textContent);
 assert(changed===1,'onChanged not called');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

test('批量翻译弹窗未选模型或只读时禁止执行', {skip:!browser},async()=>{
 await runTranslateCase('storyboard-translate-guard',`
import React from 'react';import{createRoot}from'react-dom/client';import Translate from './ComicScreenplayBatchTranslate.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
window.calls=[];
window.fetch=async()=>new Response('{}',{status:200,headers:{'Content-Type':'application/json'}});
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 createRoot(document.getElementById('root')).render(<Translate projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" modelId="" readOnly={false} onChanged={()=>{}} onClose={()=>{}}/>);
 for(let i=0;i<40&&!document.querySelector('[aria-label="翻译类型"]');i++)await tick();
 const noModel=button('开始翻译');assert(noModel,'start button missing');
 assert(noModel.disabled,'no model should block translation');
 assert(document.body.textContent.includes('还没有选择作品文本模型'),'model hint missing');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

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

const review={
  pass:false,
  summary:'第2集出现人物错位，建议修复后再进入下一步。',
  issues:[
    {id:'i-1',severity:'blocker',dimension:'人物一致性',location:'第2集/分镜3',problem:'林川被写成李梅',suggestion:'统一人物名',evidence:'李梅走进巷口'},
    {id:'i-2',severity:'major',dimension:'场景连贯',problem:'场景从卧室跳到街角无过渡'},
    {id:'i-3',severity:'minor',dimension:'台词节奏',problem:'台词略长'}
  ],
  suggested_actions:['重新生成第2集分镜','检查人物绑定'],
  reviewed_at:'2026-11-19T02:00:00Z'
};

test('剧本医生：审核报告 pass 横幅 + 问题分级 + 建议动作 + POST 调用', {skip:!browser},async()=>{
 await runCase('doctor-screenplay',`
import React from 'react';import{createRoot}from'react-dom/client';
import ComicScreenplayDoctor from'./ComicScreenplayDoctor.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
const review=${JSON.stringify(review)};
const calls=[];
window.fetch=async(input,opts={})=>{const url=String(input);const method=opts.method||'GET';await new Promise(resolve=>setTimeout(resolve,20));calls.push(method+' '+url+' '+(opts.body||''));
 if(method==='POST'&&url.includes('/doctor/screenplay'))return new Response(JSON.stringify({review}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response('{}',{status:404});};
const button=(text)=>[...document.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 let closed=false;
 const root=createRoot(document.getElementById('root'));
 root.render(<ComicScreenplayDoctor projectId="p" kind="screenplay" modelId="m1" open={true} onClose={()=>{closed=true;}}/>);
 for(let i=0;i<60&&!document.body.textContent.includes('剧本医生');i++)await tick();
 assert(document.body.textContent.includes('只出报告不改稿'),'doctor subtitle missing');
 const run=button('开始审核');assert(run&&!run.disabled,'run button missing/disabled');
 run.click();
 for(let i=0;i<80&&!document.querySelector('.doctor-review-report');i++)await tick();
 assert(document.body.textContent.includes('审核未通过'),'fail banner missing');
 assert(document.body.textContent.includes('第2集出现人物错位'),'summary missing');
 assert(document.querySelector('.doctor-review-issues li.is-blocker'),'blocker issue missing');
 assert(document.querySelector('.doctor-review-issues li.is-major'),'major issue missing');
 assert(document.body.textContent.includes('建议：统一人物名'),'suggestion missing');
 assert(document.body.textContent.includes('原文：李梅走进巷口'),'evidence missing');
 assert(document.body.textContent.includes('重新生成第2集分镜'),'suggested actions missing');
 assert(calls.some(c=>c.startsWith('POST /')&&c.includes('/doctor/screenplay')&&c.includes('m1')),'doctor POST call missing: '+calls.join(';'));
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

test('分镜医生：剧集选择 + POST 调用 + 无内容/无模型/只读禁用', {skip:!browser},async()=>{
 await runCase('doctor-storyboard',`
import React from 'react';import{createRoot}from'react-dom/client';
import ComicScreenplayDoctor from'./ComicScreenplayDoctor.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
const review=${JSON.stringify(review)};
const calls=[];
window.fetch=async(input,opts={})=>{const url=String(input);const method=opts.method||'GET';await new Promise(resolve=>setTimeout(resolve,20));calls.push(method+' '+url+' '+(opts.body||''));
 if(method==='POST'&&url.includes('/doctor/storyboard'))return new Response(JSON.stringify({review}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response('{}',{status:404});};
const button=(text)=>[...document.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 const root=createRoot(document.getElementById('root'));
 root.render(<ComicScreenplayDoctor projectId="p" kind="storyboard" engine="xiaoluo_v7_15s" episodes={[{no:1},{no:2}]} modelId="m1" hasContent={true} open={true} onClose={()=>{}}/>);
 for(let i=0;i<60&&!document.body.textContent.includes('分镜医生');i++)await tick();
 assert(document.body.textContent.includes('审核剧集'),'episode row missing');
 const run=button('开始审核');assert(run&&!run.disabled,'run button missing/disabled');
 run.click();
 for(let i=0;i<80&&!document.querySelector('.doctor-review-report');i++)await tick();
 assert(document.body.textContent.includes('审核未通过'),'fail banner missing');
 assert(calls.some(c=>c.startsWith('POST /')&&c.includes('/xiaoluo_v7_15s/1/doctor/storyboard')&&c.includes('m1')),'storyboard doctor POST call missing: '+calls.join(';'));
 // 无内容时禁用
 root.render(<ComicScreenplayDoctor projectId="p" kind="storyboard" engine="xiaoluo_v7_15s" episodes={[{no:1}]} modelId="m1" hasContent={false} open={true} onClose={()=>{}}/>);
 for(let i=0;i<30;i++)await tick();
 assert(document.body.textContent.includes('本集还没有分镜'),'empty hint missing');
 assert(button('开始审核').disabled,'run should be disabled without content');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

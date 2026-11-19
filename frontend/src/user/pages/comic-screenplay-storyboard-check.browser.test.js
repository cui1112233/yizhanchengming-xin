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

async function runCheckCase(name,contents,timeout=60000){
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

test('分镜检查弹窗出报告并逐条补齐缺失分镜', {skip:!browser},async()=>{
 await runCheckCase('storyboard-check',`
import React from 'react';import{createRoot}from'react-dom/client';import Check from './ComicScreenplayStoryboardCheck.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
window.calls=[];
const report={total:3,expected_total:4,min_seq:1,max_seq:4,missing:[3],extra:[],duplicates:[],invalid:[],passed:false};
window.fetch=async(input,opts={})=>{const url=String(input);const body=opts.body?JSON.parse(opts.body):{};window.calls.push({url,body});await new Promise(resolve=>setTimeout(resolve,20));
 if(url.endsWith('/check'))return new Response(JSON.stringify({report}),{status:200,headers:{'Content-Type':'application/json'}});
 if(url.endsWith('/check/fill'))return new Response(JSON.stringify({shot:{id:'new-'+body.seq,seq:body.seq,video_desc:'衔接镜头'},revision:9}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response('{}',{status:404});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>(el.textContent.replace(/\\s/g,'')===text||el.getAttribute('aria-label')===text));
let changed=null;
(async()=>{
 createRoot(document.getElementById('root')).render(<Check projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2,3]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" modelId="text-main" readOnly={false} onChanged={async no=>{changed=no;}} onClose={()=>{}}/>);
 for(let i=0;i<40&&!button('开始检查');i++)await tick();
 const start=button('开始检查');assert(start&&!start.disabled,'start button unavailable');
 start.click();
 for(let i=0;i<120&&!document.querySelector('.comic-sb-check-report');i++)await tick();
 assert(document.body.textContent.includes('storyboard_check_report.json'),'report download copy missing');
 const checkCalls=window.calls.filter(call=>call.url.endsWith('/check'));
 assert(checkCalls.length===1&&checkCalls[0].url.includes('/storyboards/xiaoluo_v7_15s/1/check'),'check payload wrong: '+checkCalls[0]?.url);
 const stats=document.querySelector('.comic-sb-check-stats');assert(stats,'stats missing');
 assert(document.body.textContent.includes('【缺失的分镜】')&&document.body.textContent.includes('缺失编号: 3'),'missing section missing');
 assert(document.body.textContent.includes('发现 1 个缺失分镜'),'missing count copy missing');
 assert(!document.body.textContent.includes('【多余'),'extra section should not render: no original-text comparison in V88');
 const fill=button('AI补齐缺失分镜');assert(fill&&!fill.disabled,'fill button unavailable');
 fill.click();
 for(let i=0;i<160&&!(document.body.textContent.includes('分镜补全完成'));i++)await tick();
 const fillCalls=window.calls.filter(call=>call.url.endsWith('/check/fill'));
 assert(fillCalls.length===1&&fillCalls[0].body.seq===3&&fillCalls[0].body.model==='text-main','fill payload wrong: '+JSON.stringify(fillCalls[0]?.body));
 assert(document.body.textContent.includes('已补全 1/1'),'fill summary copy wrong');
 assert(changed===1,'onChanged not called');
 // 补齐后弹窗会重新检查；模拟补全完成（全部通过）验证刷新。
 window.fetch=async(input,opts={})=>{const url=String(input);window.calls.push({url,body:opts.body?JSON.parse(opts.body):{}});return new Response(JSON.stringify({report:{...report,missing:[],passed:true}}),{status:200,headers:{'Content-Type':'application/json'}});};
 const recheck=button('重新检查');assert(recheck,'recheck button missing');
 recheck.click();
 for(let i=0;i<120&&!document.body.textContent.includes('检查通过：序号连续');i++)await tick();
 assert(document.body.textContent.includes('检查通过：序号连续，无缺失、重复或异常'),'passed copy missing');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

test('分镜检查弹窗未选模型或只读时禁止补齐', {skip:!browser},async()=>{
 await runCheckCase('storyboard-check-guard',`
import React from 'react';import{createRoot}from'react-dom/client';import Check from './ComicScreenplayStoryboardCheck.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
window.calls=[];
const report={total:2,expected_total:3,min_seq:1,max_seq:3,missing:[2,3],extra:[],duplicates:[],invalid:[],passed:false};
window.fetch=async(input,opts={})=>{const url=String(input);window.calls.push({url,body:opts.body?JSON.parse(opts.body):{}});return new Response(JSON.stringify({report}),{status:200,headers:{'Content-Type':'application/json'}});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 createRoot(document.getElementById('root')).render(<Check projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" modelId="" readOnly={false} onChanged={()=>{}} onClose={()=>{}}/>);
 for(let i=0;i<40&&!button('开始检查');i++)await tick();
 assert(document.body.textContent.includes('还没有选择作品文本模型'),'model hint missing');
 const start=button('开始检查');start.click();
 for(let i=0;i<120&&!document.querySelector('.comic-sb-check-report');i++)await tick();
 const fill=button('AI补齐缺失分镜');assert(fill,'fill button missing');
 assert(fill.disabled,'no model should block fill');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

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

async function runReshootCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1200,900','--virtual-time-budget=20000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('片段重拍弹窗加载 20 帧时间轴、框选后提交并支持更新', {skip:!browser},async()=>{
 await runReshootCase('reshoot',`
import React from 'react';import{createRoot}from'react-dom/client';import Reshoot from './ComicScreenplayReshoot.jsx';
const tick=(ms=120)=>new Promise(resolve=>setTimeout(resolve,ms));const assert=(v,m)=>{if(!v)throw Error(m)};
const json=(data,status=200)=>new Response(JSON.stringify(data),{status,headers:{'Content-Type':'application/json'}});
const frame='data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';
const frameList=Array.from({length:20},()=>frame);
const candidates={episode:1,candidates:[{shot_id:'shot-1',seq:1,episode:1,engine:'xiaoluo_v9_30s',duration:12,item_id:'item-1',video_url:'/api/ai-cut/projects/p/assets/a1/content'},{shot_id:'shot-2',seq:2,episode:1,engine:'xiaoluo_v9_30s',duration:8,item_id:'item-2',video_url:'/api/ai-cut/projects/p/assets/a2/content'}],capabilities:{frames:{available:true,reason:''},generate:{available:true,reason:'使用当前账号在设置中保存的视频模型'},apply:{available:true,reason:''},max_seconds:30,min_seconds:1}};
const done={shot_id:'shot-1',seq:1,episode:1,start:0,end:6,duration:6,status:'done',progress:100,provider_id:'task-1',item_id:'seg-1',segment_url:'/api/ai-cut/projects/p/assets/seg/content',source_url:'/api/ai-cut/projects/p/assets/a1/content'};
const applied={...done,status:'applied',applied_item_id:'final-1',applied_url:'/api/ai-cut/projects/p/assets/final/content'};
window.calls=[];
window.fetch=async(input,options={})=>{const url=String(input);const body=options.body?JSON.parse(options.body):{};window.calls.push({url,body,method:options.method||'GET'});await new Promise(resolve=>setTimeout(resolve,10));
 if(url.includes('/reshoot/candidates'))return json(candidates);
 if(url.includes('/reshoot/frames'))return json({shot_id:body.shot_id||'shot-1',duration:12,frames:frameList,first_frame:frame,last_frame:frame});
 if(url.includes('/reshoot/status'))return json({...done,status:'failed',error:'供应商离线'});
 if(url.includes('/reshoot/generate'))return json(done);
 if(url.includes('/reshoot/apply'))return json(applied);
 return json({error:'unexpected '+url},404);};
(async()=>{
 try{
  createRoot(document.getElementById('root')).render(<Reshoot projectId="p" boards={[{id:'b1',episode_no:1,shots:[{id:'shot-1',seq:1},{id:'shot-2',seq:2}]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v9_30s" readOnly={false} defaultEpisodeNo={1} onClose={()=>{}} onFinished={async()=>{window.finished=true;}}/>);
  for(let i=0;i<50&&document.querySelectorAll('.comic-rs-frames img').length<20;i++)await tick();
  assert(document.querySelectorAll('.comic-rs-frames img').length===20,'20 frame timeline missing: '+document.querySelectorAll('.comic-rs-frames img').length);
  assert(document.body.textContent.includes('片段重拍'),'title missing');
  assert(document.body.textContent.includes('分镜 1'),'shot label missing');
  assert(document.body.textContent.includes('框选片段'),'stage chip missing');
  assert(document.body.textContent.includes('选中片段 3.0 秒'),'initial selection missing: '+document.body.textContent.match(/选中片段[^\\n]*/));
  assert(document.body.textContent.includes('首尾帧参考'),'frame mode hint missing');
  const timeline=document.querySelector('.comic-rs-timeline');
  const rect=timeline.getBoundingClientRect();
  const endHandle=document.querySelector('.comic-rs-handle.is-end');assert(endHandle,'end handle missing');
  endHandle.dispatchEvent(new MouseEvent('mousedown',{bubbles:true,clientX:rect.left+rect.width*.25}));
  document.dispatchEvent(new MouseEvent('mousemove',{bubbles:true,clientX:rect.left+rect.width*.5}));
  document.dispatchEvent(new MouseEvent('mouseup',{bubbles:true}));
  await tick(200);
  assert(document.body.textContent.includes('选中片段 6.0 秒'),'drag selection not applied: '+document.body.textContent.match(/选中片段[^\\n]*/));
  const textarea=document.querySelector('.comic-rs-mode').parentElement.querySelector('textarea');assert(textarea,'prompt textarea missing');
  const setValue=Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype,'value').set;
  setValue.call(textarea,'把外套换成红色');
  textarea.dispatchEvent(new Event('input',{bubbles:true}));
  await tick(120);
  const submitButton=[...document.querySelectorAll('.comic-rs-footer button')].find(el=>el.textContent.replace(/\\s/g,'')==='重拍');assert(submitButton,'submit button missing');assert(!submitButton.disabled,'submit disabled');
  submitButton.click();
  for(let i=0;i<50&&!window.calls.some(call=>call.url.includes('/reshoot/generate'));i++)await tick();
  const generate=window.calls.find(call=>call.url.includes('/reshoot/generate'));assert(generate,'generate request missing');
  assert(generate.method==='POST','generate must POST');
  assert(generate.url.includes('/comic-screenplays/projects/p/reshoot/generate'),'wrong endpoint '+generate.url);
  assert(generate.body.shot_id==='shot-1','shot id not sent');
  assert(generate.body.start===0&&Math.abs(generate.body.end-6)<.01,'range not sent '+JSON.stringify(generate.body));
  assert(generate.body.duration===6,'duration not sent');
  assert(generate.body.prompt==='把外套换成红色','prompt not sent');
  assert(generate.body.mode==='frames','mode not sent');
  for(let i=0;i<50&&!document.body.textContent.includes('新片段已生成');i++)await tick();
  assert(document.body.textContent.includes('新片段已生成'),'result panel missing');
  assert(document.body.textContent.includes('预览新片段'),'result stage missing');
  const applyButton=[...document.querySelectorAll('.comic-rs-footer button')].find(el=>el.textContent.replace(/\\s/g,'')==='更新');assert(applyButton,'apply button missing');
  applyButton.click();
  for(let i=0;i<50&&!window.calls.some(call=>call.url.includes('/reshoot/apply'));i++)await tick();
  const apply=window.calls.find(call=>call.url.includes('/reshoot/apply'));assert(apply&&apply.body.shot_id==='shot-1','apply request missing');
  for(let i=0;i<50&&!window.finished;i++)await tick();
  assert(window.finished===true,'onFinished not called');
  for(let i=0;i<50&&!document.body.textContent.includes('已更新');i++)await tick();
  assert(document.body.textContent.includes('已更新'),'applied state missing');
  assert(document.querySelector('.comic-rs-applied video'),'applied video missing');
  document.body.setAttribute('data-result','passed');
 }catch(err){document.body.setAttribute('data-result','failed');document.body.setAttribute('data-error',String(err&&err.message||err));}
})();
`);
});

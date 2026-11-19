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

async function runLongVideoCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1200,900','--virtual-time-budget=15000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('超长视频弹窗按洛水规则分组、展示合并提示词并提交生成', {skip:!browser},async()=>{
 await runLongVideoCase('long-video',`
import React from 'react';import{createRoot}from'react-dom/client';import LongVideo from './ComicScreenplayLongVideo.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,120));const assert=(v,m)=>{if(!v)throw Error(m)};
const json=(data,status=200)=>new Response(JSON.stringify(data),{status,headers:{'Content-Type':'application/json'}});
const sections=[
 {index:1,label:'超长视频分镜*1',shot_ids:['s1','s2','s3'],shot_seqs:[1,2,3],total_seconds:29,prompt:'雨夜推门\\n闪电照亮走廊\\n阿岚回头',characters:'小雨、阿岚',scene:'旧宅、走廊',props:'伞',status:'idle',progress:0,shot_numbers:'第1镜、第2镜、第3镜'},
 {index:2,label:'超长视频分镜*2',shot_ids:['s4'],shot_seqs:[4],total_seconds:15,prompt:'门外雨停',characters:'阿岚',scene:'门口',props:'',status:'idle',progress:0,shot_numbers:'第4镜'}
];
const plan={episode:1,target_seconds:300,cap_seconds:30,total_seconds:44,final_status:'idle',running:false,capabilities:{generate:{available:true,reason:'使用当前账号在设置中保存的视频模型'},combine:{available:true,reason:''}},sections};
window.calls=[];
window.fetch=async(input,options={})=>{const url=String(input);const body=options.body?JSON.parse(options.body):{};window.calls.push({url,body,method:options.method||'GET'});await new Promise(resolve=>setTimeout(resolve,10));
 if(url.includes('/long-video/plan'))return json(plan);
 if(url.includes('/long-video/status'))return json(plan);
 if(url.includes('/long-video/generate'))return json({...plan,running:true,sections:sections.map(section=>({...section,status:'generating',progress:12}))});
 return json({error:'unexpected '+url},404);};
(async()=>{
 try{
  createRoot(document.getElementById('root')).render(<LongVideo projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2,3,4]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v9_30s" readOnly={false} modelName="Seedance 2.5" defaultEpisodeNo={1} onClose={()=>{}}/>);
  for(let i=0;i<40&&document.querySelectorAll('.comic-lv-row').length<2;i++)await tick();
  const rows=document.querySelectorAll('.comic-lv-row');assert(rows.length===2,'group rows='+rows.length);
  assert(document.body.textContent.includes('超长视频分镜*1')&&document.body.textContent.includes('超长视频分镜*2'),'group labels missing');
  assert(document.body.textContent.includes('第1镜、第2镜、第3镜'),'shot numbers missing');
  assert(document.body.textContent.includes('29 秒'),'group seconds missing');
  assert(document.body.textContent.includes('2 组'),'summary group count missing');
  const prompt=document.querySelector('.comic-lv-prompt textarea');assert(prompt,'prompt editor missing');
  assert(prompt.value==='雨夜推门\\n闪电照亮走廊\\n阿岚回头','merged prompt not shown: '+JSON.stringify(prompt.value));
  const button=[...document.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')==='批量生成');assert(button,'batch generate button missing');assert(!button.disabled,'batch generate disabled');
  button.click();
  for(let i=0;i<40&&!window.calls.some(call=>call.url.includes('/long-video/generate'));i++)await tick();
  const generate=window.calls.find(call=>call.url.includes('/long-video/generate'));assert(generate,'generate request missing');
  assert(generate.method==='POST','generate must POST');
  assert(generate.url.includes('/comic-screenplays/projects/p/long-video/generate'),'wrong endpoint '+generate.url);
  assert(generate.body.target_seconds===300,'target seconds not sent');
  assert(generate.body.episode===1,'episode not sent');
  assert(Array.isArray(generate.body.sections)&&generate.body.sections.length===2,'edited sections not sent');
  assert(generate.body.sections[0].prompt==='雨夜推门\\n闪电照亮走廊\\n阿岚回头','section prompt not sent');
  assert(!('indexes' in generate.body),'full batch must not send indexes');
  for(let i=0;i<40&&!document.querySelector('.comic-lv-body').textContent.includes('生成中');i++)await tick();
  assert(document.querySelector('.comic-lv-body').textContent.includes('生成中'),'generation status not shown after submit');
  const asked=window.calls.map(call=>call.url);
  assert(asked.some(url=>url.includes('/long-video/plan')),'plan preview not requested');
  assert(!asked.some(url=>url.includes('/shots/merge')),'long video must not reuse merge endpoint');
  document.body.setAttribute('data-result','passed');
 }catch(err){document.body.setAttribute('data-result','failed');document.body.setAttribute('data-error',String(err&&err.message||err));}
})();
`);
});

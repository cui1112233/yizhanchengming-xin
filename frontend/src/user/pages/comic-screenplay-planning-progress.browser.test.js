import test from 'node:test';
import assert from 'node:assert/strict';
import {existsSync,mkdirSync,mkdtempSync,writeFileSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {runBrowserDOM} from '../../../scripts/test-browser-dom.js';
import {build} from 'esbuild';
const here=path.dirname(fileURLToPath(import.meta.url));
const browser=[process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(x=>x&&existsSync(x));
const fixture=String.raw`
import React from 'react';import {createRoot} from 'react-dom/client';import Prep from './ComicScreenplayStoryboardPrep.jsx';
const root=createRoot(document.getElementById('root')),tick=()=>new Promise(r=>setTimeout(r,100)),assert=(v,m)=>{if(!v)throw Error(m)},text=()=>document.body.textContent.replace(/\s/g,''),visible=e=>e&&e.getClientRects().length>0;
const button=label=>[...document.querySelectorAll('button')].find(e=>visible(e)&&e.textContent.replace(/\s/g,'')===label),click=e=>{assert(e&&!e.disabled,'button unavailable');e.click();},wait=async fn=>{for(let i=0;i<100;i++){if(fn())return;await tick();}throw Error('UI wait timed out '+text());};
const rows={p1:{revision:1,chapters:[{id:'c',title:'本集一',script_text:'自造剧本一',planning_text:'OLD_SAVED_PLAN'}]},p2:{revision:1,chapters:[{id:'c',title:'本集二',script_text:'自造剧本二'}]}};
let posts=0,reads=0;window.frame={sequence:3,round:2,total:8,chars:35,missing:['二','三','四','五','六'],complete:false,phase:'requesting'};
const response=(data,status=200)=>new Response(JSON.stringify(data),{status,headers:{'content-type':'application/json'}});
window.fetch=async(input,options={})=>{const url=new URL(String(input),'http://local.test'),project=url.pathname.includes('/p2/')?'p2':'p1',state=rows[project],chapter=state.chapters[0];
 if(url.pathname.endsWith('/chapters'))return response({state,access:{read:true,edit:true,generate:true}});
 if(url.pathname.includes('storyboard-generation-attempts'))return response({attempts:[]});
 if(url.pathname.endsWith('/storyboard-plan')){posts++;const body=JSON.parse(options.body);chapter.planning_request_key=body.idempotencyKey;chapter.planning_status='running';state.revision++;
  return new Promise((resolve,reject)=>{window.completePlan=()=>{chapter.planning_text='已保存的新部分稿';chapter.planning_status='partial';chapter.planning_rounds=2;state.revision++;resolve(response({chapter,revision:state.revision}));};options.signal?.addEventListener('abort',()=>{chapter.planning_status='cancelled';state.revision++;reject(new DOMException('Aborted','AbortError'));},{once:true});});
 }
 if(url.pathname.endsWith('/storyboard-plan-progress')){reads++;const key=url.searchParams.get('request_key');const payload={chapter_id:'c',request_key:key,status:chapter.planning_status,available:true,progress:{type:'plan',chapter_id:'c',request_key:key,status:chapter.planning_status,...window.frame}};
  if(window.deferProgress)return new Promise(resolve=>{window.releaseProgress=()=>resolve(response({...payload,progress:{...payload.progress,chars:9999}}));});
  if(window.progressFailure)return response({error:'受控进度读取失败'},503);return response(payload);
 }
 return response({});
};
function mount(project='p1'){root.render(<Prep projectId={project} episode={{no:1,title:project==='p1'?'本集一':'本集二',chapterId:'c'}} modelId="own-model" canGenerate readOnly={false} defaultDensity="standard" onSaveDensity={async()=>true} onGenerate={async()=>true} onChanged={async()=>{}} onClose={()=>{}} engine="xiaoluo_v7_15s"/>);}
function pass(){document.body.dataset.result='passed';document.body.dataset.posts=String(posts);document.body.dataset.reads=String(reads);}
function fail(e){document.body.dataset.result='failed';document.body.dataset.error=e.message+" posts="+posts+" reads="+reads;}
`;
async function run(name,body){
 const result=await build({stdin:{contents:fixture+body,resolveDir:here,loader:'jsx'},bundle:true,minify:true,define:{'process.env.NODE_ENV':'"production"'},write:false,jsx:'automatic',platform:'browser',format:'iife',loader:{'.css':'empty'}});
 const base=path.resolve(here,'../../../../_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/planning-progress-browser-20261116');mkdirSync(base,{recursive:true});const dir=mkdtempSync(path.join(base,name+'-'));
 writeFileSync(path.join(dir,'test.js'),result.outputFiles[0].text);const html=path.join(dir,'test.html');writeFileSync(html,'<meta charset="utf-8"><style>.ant-modal,.ant-modal-content{animation:none!important;transition:none!important}.ant-modal.ant-zoom-leave{display:none!important}</style><div id="root"></div><script src="test.js"></script>');
 await runBrowserDOM(browser,html,dir);

}
test('真实准备窗显示本轮进度，读取失败不重付，取消保留旧稿',{skip:!browser},async()=>{await run('live-cancel',String.raw`
(async()=>{mount();await wait(()=>button('开始分镜')&&!button('开始分镜').disabled);click(button('只生成规划保存本章规划，不自动生成分镜JSON'));await wait(()=>button('开始规划')&&!button('开始规划').disabled);click(button('开始规划'));await wait(()=>text().includes('第2/8轮')&&text().includes('本次已收到35字'));assert(posts===1,'duplicate planner');window.progressFailure=true;await wait(()=>text().includes('暂时无法读取规划进度'));assert(posts===1,'read error re-posted planner');window.progressFailure=false;click(button('停止等待'));await wait(()=>rows.p1.chapters[0].planning_status==='cancelled'&&button('开始规划')&&!button('开始规划').disabled);assert(rows.p1.chapters[0].planning_text==='OLD_SAVED_PLAN'&&posts===1,'cancel lost old plan or re-posted');pass();})().catch(fail);
`);});
test('切换作品后丢弃迟到规划进度，不启动另一本作品生成',{skip:!browser},async()=>{await run('late-project',String.raw`
(async()=>{window.deferProgress=true;mount();await wait(()=>button('开始分镜')&&!button('开始分镜').disabled);click(button('开始分镜'));await wait(()=>window.releaseProgress);mount('p2');await wait(()=>text().includes('本集二'));window.releaseProgress();await tick();await tick();assert(!text().includes('9999')&&posts===1,'late old progress leaked or new generation started');await wait(()=>button('开始分镜')&&!button('开始分镜').disabled);pass();})().catch(fail);
`);});
test('重开运行中的规划仅读取进度，完成后恢复正式结果',{skip:!browser},async()=>{await run('restore-running',String.raw`
(async()=>{Object.assign(rows.p1.chapters[0],{planning_request_key:'persisted',planning_status:'running'});mount();await wait(()=>text().includes('第2/8轮'));assert(posts===0&&button('开始分镜').disabled,'reopen re-posted planner');Object.assign(rows.p1.chapters[0],{planning_status:'partial',planning_text:'已完成部分稿',planning_coverage:{ok:false,chars:35,rows:0,missing:['二'],wrong:[]}});await wait(()=>text().includes('规划可能仍不完整，草稿已保存'));assert(posts===0,'finish re-posted planner');pass();})().catch(fail);
`);});


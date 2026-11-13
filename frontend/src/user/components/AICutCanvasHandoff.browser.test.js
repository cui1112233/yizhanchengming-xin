import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';
const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file => file && fs.existsSync(file));

test('canonical handoff saves once, targets the named timeline, and hides unbound controls', { skip: !browser }, async () => {
  const bundle = await build({ stdin: {resolveDir:here,loader:'jsx',contents:`
    import React,{useState} from 'react';import{createRoot}from'react-dom/client';import Handoff from './AICutCanvasHandoff.jsx';
    window.calls=[];window.failSave=false;window.destinations=[];
    const currentUrl=new URL(location.href);currentUrl.search='?taskId=canvas-1&cutTimeline=version-2';history.replaceState({},'',currentUrl.href);
    function Harness(){const [task,setTask]=useState('canvas-1');window.setTask=setTask;return <Handoff taskId={task} canvas={{nodes:[{id:'n1',type:'text',summary:'故事正文'}]}} selectedNodeIds={['n1']} onFlushCanvas={async()=>{window.calls.push('save');if(window.failSave)throw Error('保存失败');return{revision:3}}} onNavigate={href=>window.destinations.push(href)}/>}
    const tick=()=>new Promise(resolve=>setTimeout(resolve,120));const assert=(value,text)=>{if(!value)throw Error(text)};
    const click=text=>{const b=[...document.querySelectorAll('button')].find(b=>b.textContent.replace(/\\s/g,'')===text.replace(/\\s/g,''));assert(b&&!b.disabled,'button unavailable: '+text);b.click()};
    (async()=>{createRoot(document.getElementById('root')).render(<Harness/>);await tick();await tick();
      assert(document.body.textContent.includes('同一张画布'),'origin missing');assert(document.body.textContent.includes('分镜版本'),'timeline hint lost');
      click('发送剧本');await tick();await tick();assert(document.body.textContent.includes('故事正文'),'saved text not shown');
      click('确认 1 项并返回剪辑');await tick();await tick();
      const sent=window.calls.find(v=>v&&v.path?.endsWith('/creative-canvas/script'));assert(sent,'no real transfer request');
      assert(sent.timeline==='version-2'&&sent.body.expectedRevision===7&&sent.body.nodeIds[0]==='n1','wrong timeline/revision/nodes');
      assert(window.destinations[0]==='/ai-cut?project=cut-one&timeline=version-2','return route lost target');
      window.setTask('independent');await tick();await tick();assert(!document.body.textContent.includes('送入剪辑'),'unbound project has dead controls');
      window.setTask('canvas-1');await tick();await tick();window.failSave=true;const before=window.calls.length;click('送入剪辑');await tick();
      assert(document.body.textContent.includes('保存失败'),'save failure hidden');assert(!window.calls.slice(before).includes('list-assets'),'failed save allowed transfer');
      click('继续创作');await tick();window.failSave=false;click('发送剧本');await tick();await tick();window.holdTransfer=true;const destinations=window.destinations.length;click('确认 1 项并返回剪辑');await tick();
      assert(typeof window.transferRelease==='function','transfer not pending');window.setTask('independent');await tick();window.transferRelease();await tick();await tick();assert(window.destinations.length===destinations,'stale transfer navigated the new project');
      const invalidUrl=new URL(location.href);invalidUrl.search='?taskId=canvas-2&cutTimeline=deleted';history.replaceState({},'',invalidUrl.href);window.setTask('canvas-2');await tick();await tick();
      assert(document.body.textContent.includes('原目标时间线已失效'),'invalid target silently fell back');assert([...document.querySelectorAll('button')].find(b=>b.textContent==='返回剪辑').disabled,'invalid target enabled handoff');
      document.body.dataset.result='passed';
    })().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
  `},plugins:[{name:'controlled-local-services',setup(plugin){
    plugin.onResolve({filter:/shared\/api\/(aicut|agent|client|aicutTimelines)\.js$/},args=>({path:args.path.split('/').pop(),namespace:'fake'}));
    plugin.onLoad({filter:/.*/,namespace:'fake'},args=>({loader:'js',contents: args.path==='aicut.js'?`
      export async function getAICutCreativeBinding(task){return task==='independent'?null:{taskId:task,projectId:'cut-one',projectTitle:'剪辑工程'}}
      export async function transferAICutCreativeAssets(project,ids,request){return request('/api/ai-cut/projects/'+project+'/creative-canvas/assets',{method:'POST',body:JSON.stringify({assetIds:ids})})}
      export async function transferAICutCreativeScript(project,ids,revision,request){return request('/api/ai-cut/projects/'+project+'/creative-canvas/script',{method:'POST',body:JSON.stringify({nodeIds:ids,expectedRevision:revision})})}
    `: args.path==='agent.js'?`
      export async function getAgentCanvas(){return{canvas:{revision:3,nodes:[{id:'n1',type:'text',title:'故事',summary:'故事正文'}]}}}
      export async function listAgentAssets(){window.calls.push('list-assets');return{assets:[]}}
    `: args.path==='aicutTimelines.js'?`
      export async function listProjectTimelines(){return[{id:'default',title:'默认时间线'},{id:'version-2',title:'分镜版本'}]}
      export async function loadProjectTimeline(project,timeline){if(timeline!=='version-2')throw Error('wrong target');return{revision:7}}
    `: `export async function apiRequest(path,options){window.calls.push({path,timeline:options.headers?.get('X-AICut-Timeline-Id'),body:JSON.parse(options.body)});if(window.holdTransfer)await new Promise(resolve=>window.transferRelease=resolve);return{timeline:{revision:8}}}` }));
  }}],bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  const archive=path.resolve(here,'../../../../_local-archive-20261028/canonical-canvas');fs.mkdirSync(archive,{recursive:true});const dir=fs.mkdtempSync(path.join(archive,'handoff-'));fs.writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);const html=path.join(dir,'test.html');fs.writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--virtual-time-budget=6000',`--user-data-dir=${path.join(dir,'profile')}`,'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,windowsHide:true,stdio:['ignore','pipe','pipe']});assert.match(output,/data-result="passed"/,output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
});

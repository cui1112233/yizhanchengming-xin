import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';

const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file => file && existsSync(file));

test('real canvas pet panel registers, submits saved selection, applies a plan, and resets by project', { skip: !browser }, async () => {
  const bundle = await build({
    stdin: { resolveDir: here, loader: 'jsx', contents: `
      import React,{useEffect,useState} from 'react'; import {createRoot} from 'react-dom/client';
      import {useCanvasPetService} from './useCanvasPetService.jsx';
      import {getPetPageService,subscribePetPageService} from '../../shared/pet/pageService.js';
      const tick=()=>new Promise(resolve=>setTimeout(resolve,60));
      const assert=(condition,message)=>{if(!condition)throw Error(message)};
      let edit=0,locate=0,saves=0; window.calls=[];
      function Harness(){
        const [task,setTask]=useState('one'),[selection,setSelection]=useState(['n1']),[canvas,setCanvas]=useState({revision:4,nodes:[{id:'n1'},{id:'n2'}],edges:[]});
        window.switchProject=()=>setTask('two');window.clearSelection=()=>setSelection([]);
        useCanvasPetService({taskId:task,canvas,selectedIds:selection,flush:async()=>{saves++;return {revision:canvas.revision}},acceptCanvas:async next=>{setCanvas(next);return true},editSelection:()=>edit++,locateSelection:()=>locate++});
        return <button id="background-edit" onClick={()=>{window.backgroundEdits=(window.backgroundEdits||0)+1}}>其他画布编辑</button>;
      }
      function Host(){const [service,setService]=useState(getPetPageService);useEffect(()=>subscribePetPageService(()=>setService(getPetPageService())),[]);return service?<div key={service.scope}>{service.render()}</div>:null}
      const click=text=>{const button=[...document.querySelectorAll('button')].find(b=>b.textContent===text);assert(button&&!button.disabled,'button unavailable: '+text);button.click()};
      const type=text=>{const el=document.querySelector('textarea');Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value').set.call(el,text);el.dispatchEvent(new Event('input',{bubbles:true}));};
      (async()=>{
        createRoot(document.getElementById('root')).render(<React.StrictMode><Host/><Harness/></React.StrictMode>);await tick();await tick();
        assert(getPetPageService()?.id==='canvas-agent','wrong service');
        assert(document.body.textContent.includes('已选 1 张'),'selection missing');
        click('编辑选中卡片');click('定位选中卡片');assert(edit===1&&locate===1,'real callbacks not run');
        type('整理当前故事');await tick();click('生成创作方案');await tick();await tick();
        assert(window.calls[0].context.canvasPet.revision===4&&window.calls[0].context.canvasPet.selectedNodeIds[0]==='n1','request lost durable context');
        assert(document.body.textContent.includes('改故事标题'),'missing real plan');
        click('确认应用到画布');await tick();await tick();
        assert(window.calls[1].revision===4&&saves===2,'approval did not save/use revision');
        assert(document.body.textContent.includes('已应用到画布')&&document.body.textContent.includes('版本 5'),'real result not accepted');
        assert(!window.backgroundEdits,'approval allowed concurrent canvas writes');
        document.querySelector('#background-edit').click();assert(window.backgroundEdits===1,'write lock was not released');
        type('未发送草稿');await tick();window.clearSelection();await tick();
        assert(document.querySelector('textarea').value==='未发送草稿','same-scope update lost draft');
        assert([...document.querySelectorAll('button')].find(b=>b.textContent==='编辑选中卡片').disabled,'selection guard missing');
        window.switchProject();await tick();await tick();assert(document.querySelector('textarea').value==='','project leaked draft');
        document.body.dataset.result='passed';
      })().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
    ` },
    plugins: [{ name:'fake-local-services', setup(plugin) {
      plugin.onResolve({filter:/shared\/api\/(agent|auth)\.js$/}, args=>({path:args.path.includes('auth')?'auth':'agent',namespace:'fake'}));
      plugin.onLoad({filter:/.*/,namespace:'fake'}, args=>({loader:'js',contents:args.path==='auth' ? 'export function getCurrentUsername(){return "alice"}' : `
        export async function getAgentTask(){return {task:{messages:[]}}}
        export async function askAgent(input){window.calls.push(input);return {task:{messages:[{role:'assistant',content:'请确认',context:{canvasPetRevision:4,directorPlan:{id:'p1',summary:'改故事标题',commands:[{type:'update_node',nodeId:'n1',patch:{title:'新故事'}}]}}}]}}}
        export async function approveAgentDirectorPlan(task,plan,revision){window.calls.push({task,plan,revision});document.querySelector('#background-edit').click();await new Promise(resolve=>setTimeout(resolve,30));return {canvas:{revision:5,nodes:[{id:'n1',title:'新故事'},{id:'n2'}],edges:[]}}}
      `}));
    }}], bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}
  });
  const archive=path.resolve(here,'../../../../_local-archive-20261028/canvas-pet-parity');mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,'panel-'));writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);const html=path.join(dir,'test.html');writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--virtual-time-budget=5000',`--user-data-dir=${path.join(dir,'profile')}`,'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,windowsHide:true,stdio:['ignore','pipe','pipe']});
  assert.match(output,/data-result="passed"/,output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
});

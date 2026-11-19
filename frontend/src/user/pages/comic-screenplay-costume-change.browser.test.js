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

async function runCostumeCase(name,contents,timeout=60000){
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

test('换装弹窗按洛水规则圈范围、筛场景、选人物并预览命中', {skip:!browser},async()=>{
 await runCostumeCase('storyboard-costume',`
import React from 'react';import{createRoot}from'react-dom/client';import Costume from './ComicScreenplayCostumeChange.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
const options={characters:['张三','张三-睡衣','小Y','李四'],scenes:['卧室','学校','客厅'],shot_count:6};
const targets=[{shot_id:'cs-1',seq:1,scene:'卧室',before:'张三、小Y',after:'张三-睡衣、小Y'},{shot_id:'cs-2',seq:2,scene:'卧室',before:'张三',after:'张三-睡衣'}];
window.calls=[];
window.fetch=async(input,opts={})=>{const url=String(input);const body=opts.body?JSON.parse(opts.body):{};window.calls.push({url,body});await new Promise(resolve=>setTimeout(resolve,20));
 if(url.includes('/costume/options'))return new Response(JSON.stringify({options}),{status:200,headers:{'Content-Type':'application/json'}});
 if(url.includes('/costume/preview'))return new Response(JSON.stringify({plan:{shot_start:body.shot_start,shot_end:body.shot_end,scene:body.scene,from:body.from,to:body.to,total:6,matched:2,targets}}),{status:200,headers:{'Content-Type':'application/json'}});
 return new Response(JSON.stringify({plan:{shot_start:1,shot_end:6,scene:'卧室',from:'张三',to:'张三-睡衣',total:6,matched:2,targets},message:'已在分镜 1-6（卧室）将「张三」换成「张三-睡衣」，共 2 个分镜'}),{status:200,headers:{'Content-Type':'application/json'}});};
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
let changed=null,closed=false;
(async()=>{
 createRoot(document.getElementById('root')).render(<Costume projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2,3,4,5,6]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" readOnly={false} onChanged={async no=>{changed=no;}} onClose={()=>{closed=true;}}/>);
 for(let i=0;i<40&&!document.querySelector('[aria-label="原人物"]');i++)await tick();
 assert(window.calls[0].url.includes('/storyboards/xiaoluo_v7_15s/1/costume/options'),'options endpoint wrong: '+window.calls[0].url);
 const startInput=document.querySelector('[aria-label="起始分镜"] input')||document.querySelector('[aria-label="起始分镜"]');assert(startInput,'shot start input missing');
 assert(String(startInput.value)==='1','default shot start wrong: '+startInput.value);
 const endInput=document.querySelector('[aria-label="结束分镜"] input')||document.querySelector('[aria-label="结束分镜"]');assert(endInput,'shot end input missing');
 assert(String(endInput.value)==='6','default shot end should be shot count: '+endInput.value);
 assert(document.body.textContent.includes('共 6 个分镜'),'shot count copy missing');
 assert(document.body.textContent.includes('不限制场景'),'default scene should be unlimited');
 await pick('原人物','张三');
 await pick('替换为','张三-睡衣');
 await pick('筛选场景','卧室');
 for(let i=0;i<40&&!document.querySelector('.comic-sb-costume-stats');i++)await tick();
 const preview=window.calls.filter(call=>call.url.includes('/costume/preview')).at(-1);assert(preview,'preview not requested');
 assert(preview.body.scene==='卧室'&&preview.body.from==='张三'&&preview.body.to==='张三-睡衣','preview payload wrong: '+JSON.stringify(preview.body));
 assert(preview.body.shot_start===1&&preview.body.shot_end===6,'preview range wrong: '+JSON.stringify(preview.body));
 const stats=document.querySelector('.comic-sb-costume-stats');assert(stats.textContent.includes('2'),'matched count missing: '+stats.textContent);
 const list=document.querySelectorAll('.comic-sb-costume-list li');assert(list.length===2,'target list wrong');
 assert(list[0].textContent.includes('张三-睡衣、小Y'),'replacement should keep character order: '+list[0].textContent);
 const apply=button('开始换装');assert(apply&&!apply.disabled,'apply button unavailable');
 apply.click();
 for(let i=0;i<40&&!changed;i++)await tick();
 assert(changed===1,'onChanged not called');
 assert(closed,'dialog stayed open');
 const exec=window.calls.filter(call=>call.url.includes('/costume/apply')).at(-1);assert(exec,'apply not requested');
 assert(exec.body.scene==='卧室'&&exec.body.from==='张三'&&exec.body.to==='张三-睡衣','apply payload wrong: '+JSON.stringify(exec.body));
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

test('换装弹窗只读时不允许执行', {skip:!browser},async()=>{
 await runCostumeCase('storyboard-costume-readonly',`
import React from 'react';import{createRoot}from'react-dom/client';import Costume from './ComicScreenplayCostumeChange.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,140));const assert=(v,m)=>{if(!v)throw Error(m)};
window.calls=[];
window.fetch=async(input,opts={})=>{const url=String(input);const body=opts.body?JSON.parse(opts.body):{};window.calls.push({url,body});await new Promise(resolve=>setTimeout(resolve,20));
 return new Response(JSON.stringify({options:{characters:['张三','张三-睡衣'],scenes:['卧室'],shot_count:2}}),{status:200,headers:{'Content-Type':'application/json'}});};
const button=(text,scope=document)=>[...scope.querySelectorAll('button')].find(el=>el.textContent.replace(/\\s/g,'')===text);
(async()=>{
 createRoot(document.getElementById('root')).render(<Costume projectId="p" boards={[{id:'b1',episode_no:1,shots:[1,2]}]} episodes={[{no:1,title:'相遇'}]} engine="xiaoluo_v7_15s" readOnly onChanged={()=>{}} onClose={()=>{}}/>);
 for(let i=0;i<40&&!document.querySelector('[aria-label="原人物"]');i++)await tick();
 const apply=button('开始换装');assert(apply,'apply button missing');
 assert(apply.disabled,'readOnly allowed costume change');
 assert(window.calls.every(call=>!call.url.includes('/costume/apply')),'readOnly wrote storyboard');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

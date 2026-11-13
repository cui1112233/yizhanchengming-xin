import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';
const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file => file && existsSync(file));
test('真实画风弹窗取消保留原选择，确认和清空才修改草稿；四用途各有选择框', { skip: !browser }, async () => {
  const bundle = await build({ stdin: { resolveDir: here, loader: 'jsx', contents: `
    import React, {useState} from 'react'; import {createRoot} from 'react-dom/client';
    import {ComicCreativeFields} from './ComicScreenplayCreationFields.jsx';
    import {emptyCreativeConfig} from './comic-screenplay-creative.js';
    const tick=()=>new Promise(resolve=>setTimeout(resolve,100));
    const assert=(condition,message)=>{if(!condition)throw Error(message)};
    function Harness(){const [value,setValue]=useState({...emptyCreativeConfig(),style_id:'a',style_name:'原画风'});return <><ComicCreativeFields value={value} onChange={patch=>setValue(current=>({...current,...patch}))} library={{styles:[{id:'a',name:'原画风',images:[]},{id:'b',name:'新画风',images:[]}],prompts:[]}}/><output id="value">{JSON.stringify(value)}</output></>}
    const click=(selector,text)=>{const button=[...document.querySelectorAll(selector)].find(b=>b.textContent.replace(/\\s/g,'')===text);assert(button&&!button.disabled,'missing '+text);button.click()};
    (async()=>{createRoot(document.getElementById('root')).render(<Harness/>);await tick();await tick();
      for(const label of ['剧本提示词','分镜提示词','角色设计提示词','场景设计提示词'])assert(document.querySelector('[aria-label="'+label+'"]'),'missing '+label);
      document.querySelector('.style-picker-trigger').click();await tick();await tick();
      click('.comic-style-library b','新画风');await tick();
      click('.screenplay-style-modal button','取消');await tick();await tick();
      assert(JSON.parse(document.querySelector('#value').textContent).style_id==='a','cancel applied draft');
      document.querySelector('.style-picker-trigger').click();await tick();await tick();
      click('.comic-style-library b','新画风');await tick();click('.screenplay-style-modal button','确认选择');await tick();await tick();
      assert(JSON.parse(document.querySelector('#value').textContent).style_id==='b','confirm did not apply selected ID');
      document.querySelector('[aria-label="清除画风选择"]').click();await tick();
      assert(JSON.parse(document.querySelector('#value').textContent).style_id===null,'clear did not preserve explicit null');
      document.body.dataset.result='passed';
    })().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
  ` }, plugins: [{ name: 'mock-model-catalog', setup(plugin) { plugin.onResolve({ filter: /modelCatalog\.js$/ }, () => ({ path: 'models', namespace: 'fake' })); plugin.onLoad({ filter: /.*/, namespace: 'fake' }, () => ({ contents: 'export async function listAvailableModels(){return []}', loader: 'js' })); } }], bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser', loader: { '.css': 'empty' } });
  const archive = path.resolve(here, '../../../../_local-archive-20261028/screenplay-creative-ui'); mkdirSync(archive, { recursive: true });
  const dir = mkdtempSync(path.join(archive, 'picker-')); writeFileSync(path.join(dir, 'test.js'), bundle.outputFiles[0].text);
  const html = path.join(dir, 'test.html'); writeFileSync(html, '<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output = execFileSync(browser, ['--headless=new', '--disable-gpu', '--no-first-run', '--disable-background-networking', '--host-resolver-rules=MAP * ~NOTFOUND', '--virtual-time-budget=6000', '--user-data-dir='+path.join(dir, 'profile'), '--dump-dom', pathToFileURL(html).href], { encoding: 'utf8', timeout: 60000, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'], maxBuffer: 4000000 });
  assert.ok(/data-result="passed"/.test(output), output.match(/data-error="([^"]*)"/)?.[1] || 'browser did not finish');
});


test('真实媒体页默认发送项目B/C模型，显式临时选择D，失效项目模型阻止提交', { skip: !browser }, async () => {
  const bundle = await build({ stdin: { resolveDir: here, loader: 'jsx', contents: `
    import React,{useState} from 'react';import {createRoot} from 'react-dom/client';
    import ComicScreenplayMediaPanel from './ComicScreenplayMediaPanel.jsx';
    const tick=()=>new Promise(resolve=>setTimeout(resolve,120));
    const assert=(condition,message)=>{if(!condition)throw Error(message)};
    window.calls=[];
    function Harness(){const [config,setConfig]=useState({image_model_id:'B',image_edit_model_id:'C'});window.invalidate=()=>setConfig({image_model_id:'missing',image_edit_model_id:'missing'});return <ComicScreenplayMediaPanel projectId="p" workspace={{project:{creative_config:config},assets:[{id:'hero',name:'主角',type:'character'}]}}/>}
    const button=text=>[...document.querySelectorAll('button')].find(b=>b.textContent.replace(/\\s/g,'')===text);
    (async()=>{createRoot(document.getElementById('root')).render(<Harness/>);await tick();await tick();await tick();
      assert(!button('生成并保存图片').disabled,'configured B did not become available');
      button('生成并保存图片').click();await tick();await tick();
      assert(window.calls[0]?.modelId==='B','directory first model overrode project B');
      document.querySelector('.comic-media-form .ant-checkbox-input').click();await tick();
      button('生成并保存图片').click();await tick();await tick();
      assert(window.calls[1]?.modelId==='C','reference image did not use project edit C');
      const selector=document.querySelector('[aria-label="本次图片模型"] .ant-select-selector');
      selector.dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));await tick();await tick();
      const option=[...document.querySelectorAll('.ant-select-item-option')].find(el=>el.title==='模型D');assert(option,'override D option missing');option.click();await tick();
      assert(document.body.textContent.includes('本次临时覆盖作品模型'),'temporary override not marked');
      button('生成并保存图片').click();await tick();await tick();assert(window.calls[2]?.modelId==='D','explicit override did not reach request');
      button('恢复作品配置').click();await tick();button('生成并保存图片').click();await tick();await tick();assert(window.calls[3]?.modelId==='C','restore did not resume project C');
      window.invalidate();await tick();assert(button('生成并保存图片').disabled,'invalid configured model did not block request');assert(document.body.textContent.includes('当前不可用'),'invalid configuration lacked error');
      document.body.dataset.result='passed';
    })().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
  ` }, plugins: [{ name: 'mock-media-services', setup(plugin) {
    plugin.onResolve({ filter: /shared\/api\/(modelCatalog|comicScreenplayWorkflow)\.js$/ }, args => ({ path: args.path.includes('modelCatalog') ? 'models' : 'media', namespace: 'fake' }));
    plugin.onLoad({ filter: /.*/, namespace: 'fake' }, args => ({ loader: 'js', contents: args.path === 'models' ? 'export async function listAvailableModels(){return ["A","B","C","D"].map(id=>({id,displayName:"模型"+id}))}' : `
      const data={items:[{id:'ref',name:'参考图片',kind:'image',status:'succeeded'}],settings:{revision:0},capabilities:{image:{available:true}}};
      export async function getComicMedia(){return data}
      export async function listComicMediaVoices(){return {voices:[]}}
      export async function actComicMedia(id,input){window.calls.push(input);return data}
      export async function downloadComicMediaPackage(){} export async function downloadComicMediaResult(){} export async function uploadComicMedia(){} export async function importComicMediaPackage(){}
    ` }));
  } }], bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser', loader: { '.css': 'empty' } });
  const archive=path.resolve(here,'../../../../_local-archive-20261028/screenplay-creative-ui');mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,'media-model-'));writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--virtual-time-budget=8000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';
const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file => file && existsSync(file));

test('普通库选择：分类搜索、类型隔离、关闭、重读、单次确认及键盘绑定均不写入', { skip: !browser }, async () => {
  const bundle = await build({ stdin: { resolveDir: here, loader: 'jsx', contents: `
    import React from 'react';import{createRoot}from'react-dom/client';import Picker from './ComicScreenplayAssetLibraryPicker.jsx';
    const root=createRoot(document.getElementById('root')),tick=()=>new Promise(r=>setTimeout(r,70)),assert=(v,m)=>{if(!v)throw Error(m)};
    let selects=[],closes=0,calls=[],requests=0;window.fetch=()=>{requests++;throw Error('network forbidden')};
    const entries=Array.from({length:65},(_,i)=>({id:String(i),type:'character',name:i===1?'Alice':'人物'+i,description:i===2?'Silver HERO':'说明',category:i%2?'A':'a',image:{url:''},revision:1}));
    const load=async arg=>{calls.push(arg);return entries},cards=()=>[...document.querySelectorAll('.comic-library-card')];
    const button=text=>[...document.querySelectorAll('button')].find(e=>e.textContent.trim()===text),click=async text=>{button(text).click();await tick()};
    const mount=async(props={})=>{root.render(<Picker key={Math.random()} type="character" loadEntries={load} onSelect={e=>selects.push(e)} onClose={()=>closes++} {...props}/>);await tick();await tick()};
    const input=async text=>{const el=document.querySelector('[aria-label="搜索普通库"]');Object.getOwnPropertyDescriptor(HTMLInputElement.prototype,'value').set.call(el,text);el.dispatchEvent(new Event('input',{bubbles:true}));await tick()};
    (async()=>{
      await mount({type:'prop',variant:'legacy-add'});assert(calls.at(-1).type==='character','legacy add must read characters');assert(cards().length===60,'legacy cap');assert(document.querySelector('select').options.length===3,'case distinct categories lost');
      await input('  aLiCe  ');assert(cards().length===1&&cards()[0].lastElementChild.textContent==='Alice','name trim/lower');await input(' HERO ');assert(cards().length===1&&cards()[0].lastElementChild.textContent==='人物2','description search');await input('');
      const category=document.querySelector('select');category.value='A';category.dispatchEvent(new Event('change',{bubbles:true}));await tick();assert(cards().length===32,'category must match exact case');cards()[0].click();await tick();assert(selects.length===0,'selection must not commit');await click('取消');assert(selects.length===0&&closes===1,'cancel selected item');assert(!document.querySelector('[role="dialog"]'),'cancel retained dialog');
      await mount({type:'scene',variant:'replace'});assert(calls.at(-1).type==='scene'&&!document.querySelector('select'),'scene replacement category/type');assert(cards().length===60,'replace cap');cards()[0].click();await tick();const confirm=button('确认更换');confirm.click();confirm.click();await tick();assert(selects.length===1&&closes===2,'duplicate confirm');
      await mount();assert(cards().length===65&&!document.querySelector('select'),'binding cap/category');const card=cards()[64];card.dispatchEvent(new KeyboardEvent('keydown',{key:' ',bubbles:true,cancelable:true}));card.click();await tick();assert(selects.length===2&&selects.at(-1).id==='64'&&closes===3,'space/duplicate binding');
      await mount();cards()[1].dispatchEvent(new KeyboardEvent('keydown',{key:'Enter',bubbles:true,cancelable:true}));await tick();assert(selects.length===3&&closes===4,'enter binding');
      let attempts=0;await mount({loadEntries:async()=>{attempts++;if(attempts===1)throw Error('读取被拒绝');return entries}});assert(document.querySelector('[role="alert"]').textContent.includes('读取被拒绝'),'failure missing');assert(!document.body.textContent.includes('暂无可绑定资产'),'failure shown as empty');await click('重新读取');assert(attempts===2&&cards().length===65,'reload failed');await click('取消');
      let resolve;await mount({loadEntries:()=>new Promise(r=>resolve=r)});assert(document.body.textContent.includes('加载中'),'loading missing');await click('取消');const before=selects.length;resolve(entries);await tick();assert(!document.querySelector('[role="dialog"]')&&selects.length===before,'late load reopened/selected');assert(requests===0,'picker made network request');document.body.dataset.result='passed';
    })().catch(e=>{document.body.dataset.result='failed';document.body.dataset.error=e.message});
  ` }, bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser', loader: { '.css': 'empty' } });
  const archive=path.resolve(here,'../../../../_local-archive-20261028/luoshui-screenplay-contract-20261114/ui-reconstruction');mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,'ordinary-library-picker-'));writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');writeFileSync(html,'<meta charset="utf-8"><style>'+readFileSync(path.join(here,'comic-screenplay-asset-library-picker.css'),'utf8')+'</style><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--virtual-time-budget=10000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
});

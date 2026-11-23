import assert from 'node:assert/strict';
import test from 'node:test';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';
import { batchFactoryBookTimeline } from './batchFactoryBookState.js';

const here = fileURLToPath(new URL('.', import.meta.url));
const bundle = await build({
  stdin: {contents: "export * from './batchFactoryNovelListHelpers.js'", resolveDir: here},
  bundle: true, write: false, format: 'esm', define: {'import.meta.env':'{}'}
});
const helpers = await import('data:text/javascript;base64,' + Buffer.from(bundle.outputFiles[0].contents).toString('base64'));
const list = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url),'utf8');
const panel = readFileSync(new URL('./TextMediaPanel.jsx', import.meta.url),'utf8');

test('文本只在 AI 优化且缺少可用文本模型时阻止启动', () => {
  assert.equal(typeof helpers.textModelPreflightReasons,'function');
  for (const settings of [{},{sensitiveMode:'system'},{sensitiveMode:'off'}, {sensitiveMode:'ai',textModelId:'model'},{sensitiveMode:'ai',sensitiveAIModels:['repair']},{sensitiveAIEnabled:true,sensitiveAIModels:'repair,backup'}]) {
    assert.deepEqual(helpers.textModelPreflightReasons([settings]),[]);
  }
  assert.deepEqual(helpers.textModelPreflightReasons([{sensitiveMode:'ai'}]),['AI 优化需要选择文本模型']);
});

test('文本日志保留六个阶段、已成功步骤和后端失败原因', () => {
  const runs = [{stage:'text_process',status:'succeeded'}, {stage:'text_render',status:'failed',errorMessage:'第 2 页：请调整字号后重试'}];
  const timeline = batchFactoryBookTimeline({id:'b'},{productionMode:'text',stageSummary:{runs}});
  assert.deepEqual(timeline.map(step=>step.key),['text_process','text_paginate','text_render','text_validate','text_collection','text_publish']);
  assert.equal(timeline[0].status,'completed');
  assert.equal(timeline[2].detail,runs[1].errorMessage);
  assert.ok(list.includes('stageSummary: stageSummaries?.[record.id]'),'日志应使用当前书的服务端阶段记录');
});

test('重试只渲染失败文字页，并按当前版本重建合集或提交原合集', async () => {
  const calls=[];
  const api={
    listTextPages:async()=>({document:{revision:7,activeVersionId:'v',paginationVersion:2},pages:[
      {id:'one',ordinal:1,revision:3,status:'ready'}, {id:'two',ordinal:2,revision:5,status:'failed'},
      {id:'three',ordinal:3,revision:4,status:'ready',locked:true}]}),
    renderTextPage:async(...args)=>calls.push(['render',...args]),
    createTextCollection:async(...args)=>calls.push(['collection',...args]),
    publishTextCollection:async(...args)=>calls.push(['publish',...args])
  };
  assert.equal(typeof helpers.retryTextStage,'function');
  await helpers.retryTextStage(api,'batch','book',{stage:'text_render',pageId:'two'});
  await helpers.retryTextStage(api,'batch','book',{stage:'text_collection'});
  await helpers.retryTextStage(api,'batch','book',{stage:'text_publish',collectionId:'c'});
  assert.deepEqual(calls,[
    ['render','batch','book','two',{expectedRevision:5}],
    ['collection','batch','book',{expectedRevision:7,versionId:'v',paginationVersion:2,pageIds:['one','two','three']}],
    ['publish','batch','book','c',{force:false}]
  ]);
  await assert.rejects(()=>helpers.retryTextStage(api,'batch','book',{stage:'text_render',pageId:'three'}));
  assert.equal(calls.length,3);
});

test('渲染失败信息里的页码可定位尚未落图的页和封面', async () => {
  const calls = [];
  await helpers.retryTextStage({
    listTextPages:async()=>({document:{revision:7},pages:[
      {id:'p1',ordinal:1,revision:3,status:'ready'}, {id:'p2',ordinal:2,revision:5,status:'pending'},
      {id:'p3',ordinal:3,revision:2,status:'pending'}]}),
    renderTextPage:async(...args)=>calls.push(args),
    renderTextCover:async(...args)=>calls.push(args)
  },'batch','book',{stage:'text_render',errorMessage:'第 2 页：失败\n封面：失败'});
  assert.deepEqual(calls,[['batch','book','p2',{expectedRevision:5}],['batch','book',{expectedRevision:7}]]);
});

test('当前上传以回执状态为准，首次上传、等待回读和再次上传分别显示', () => {
  assert.equal(typeof helpers.textUploadControl,'function');
  assert.equal(helpers.textUploadControl(null,null).disabled,true);
  assert.equal(helpers.textUploadControl({id:'c'},null).disabled,true);
  assert.deepEqual(helpers.textUploadControl({id:'c'},{published:false,publish:null}),{label:'上传网络',disabled:false,force:false});
  const previous={status:'published',carouselId:12,imageIds:[15]};
  assert.equal(helpers.textUploadControl({id:'c'},{published:false,publish:{status:'pending'},lastPublished:previous}).disabled,true);
  assert.deepEqual(helpers.textUploadControl({id:'c'},{published:true,publish:previous}),{label:'再次上传',disabled:false,force:true});
  assert.ok(panel.includes('getTextPublishStatus'));
  assert.ok(panel.includes('window.confirm(\'这个文本合集已上传，确认再次上传？\')'));
});

test('混选漫画与文本会在提交前提示分开选择', () => {
  assert.equal(typeof helpers.collectionUploadKind,'function');
  assert.equal(helpers.collectionUploadKind(['text','text']),'text');
  assert.equal(helpers.collectionUploadKind(['comic']),'comic');
  assert.throws(()=>helpers.collectionUploadKind(['comic','text']),{message:'漫画与文本不能一起上传，请分开选择。'});
});

const browser = [process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(path => path && existsSync(path));
test('Chrome 点击上传、取消和确认再次上传，等待回读期间不会重复提交', {skip:!browser}, async () => {
  const dir = mkdtempSync(join(tmpdir(),'text-publish-'));
  try {
    const result = await build({stdin:{resolveDir:here,loader:'jsx',contents:`
      import React from 'react';
      import {createRoot} from 'react-dom/client';
      import TextMediaPanel from './TextMediaPanel.jsx';
      const root=createRoot(document.querySelector('#root'));
      const sleep=n=>new Promise(done=>setTimeout(done,n));
      const button=label=>Array.from(document.querySelectorAll('button')).find(b=>b.textContent.trim()===label);
      async function until(test){for(let i=0;i<500;i++){if(test())return;await sleep(20)}throw Error(document.body.innerText)}
      let collection=null,status={published:false,publish:null},confirms=0,confirmResult=false;
      const posts=[];
      window.confirm=()=>{confirms++;return confirmResult};
      window.fetch=async(url,opts={})=>{
        const path=String(url).split('?')[0];
        let data;
        if(path.endsWith('text-pages'))data={pages:[],artifacts:[]};
        else if(path.endsWith('text-collections'))data={items:collection?[collection]:[]};
        else if(path.endsWith('text-collections/c'))data=collection;
        else if(path.endsWith('text-collections/c/publish')){
          if(opts.method==='POST'){
            posts.push(JSON.parse(opts.body));
            status={published:false,publish:{status:'pending'}};
          }
          data=status;
        }else throw Error('意外请求 '+path);
        return new Response(JSON.stringify(data),{headers:{'Content-Type':'application/json'}});
      };
      const mount=key=>root.render(<TextMediaPanel key={key} book={{id:'book',title:'测试书'}} batchId="batch" initialTab="collections"/>);
      async function run(){
        mount(1);await until(()=>button('上传网络')&&document.body.innerText.includes('暂无文本合集'));
        if(!button('上传网络').disabled)throw Error('无合集时允许上传');
        collection={id:'c',status:'ready'};mount(2);
        await until(()=>button('上传网络')&&!button('上传网络').disabled);
        button('上传网络').click();await until(()=>button('等待上传确认'));
        if(posts.length!==1||posts[0].force!==false||confirms||!button('等待上传确认').disabled)throw Error('首次提交或待回读状态错误');
        status={published:true,publish:{status:'published',carouselId:11,imageIds:[12]}};
        await until(()=>button('再次上传')&&!button('再次上传').disabled);
        button('再次上传').click();await sleep(40);
        if(confirms!==1||posts.length!==1)throw Error('取消后仍然发起重传');
        confirmResult=true;button('再次上传').click();await until(()=>posts.length===2);
        if(confirms!==2||posts[1].force!==true)throw Error('确认后未按强制重传提交');
        document.body.dataset.result='passed';
      }
      run().catch(e=>document.body.dataset.error=e.message);
    `},bundle:true,write:false,format:'iife',loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'},minify:true});
    writeFileSync(join(dir,'app.js'),result.outputFiles[0].contents);
    const html=join(dir,'index.html');
    writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="app.js"></script>');
    const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=16000',
      '--dump-dom',pathToFileURL(html).href],{timeout:45000,windowsHide:true,encoding:'utf8',stdio:['ignore','pipe','pipe']});
    assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || '浏览器未完成上传回读验证');
  } finally { rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150}); }
});

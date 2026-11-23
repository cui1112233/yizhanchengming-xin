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

test('重试只渲染失败文字页，并按当前版本重建合集', async () => {
  const calls=[];
  const api={
    listTextPages:async()=>({document:{revision:7,activeVersionId:'v',paginationVersion:2},pages:[
      {id:'one',ordinal:1,revision:3,status:'ready'}, {id:'two',ordinal:2,revision:5,status:'failed'},
      {id:'three',ordinal:3,revision:4,status:'ready',locked:true}]}),
    renderTextPage:async(...args)=>calls.push(['render',...args]),
    createTextCollection:async(...args)=>calls.push(['collection',...args])
  };
  assert.equal(typeof helpers.retryTextStage,'function');
  await helpers.retryTextStage(api,'batch','book',{stage:'text_render',pageId:'two'});
  await helpers.retryTextStage(api,'batch','book',{stage:'text_collection'});
  assert.deepEqual(calls,[
    ['render','batch','book','two',{expectedRevision:5}],
    ['collection','batch','book',{expectedRevision:7,versionId:'v',paginationVersion:2,pageIds:['one','two','three']}]
  ]);
  await assert.rejects(()=>helpers.retryTextStage(api,'batch','book',{stage:'text_render',pageId:'three'}));
  assert.equal(calls.length,2);
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

test('批量重试先确认和选合集，再一次提交所有同意的自动书，手动上传单独提交',async()=>{
  const events=[];
  const api={
    getTextPublishStatus:async(batch,id,c)=>{events.push('read:'+id+':'+c);return {published:false,publish:{status:'failed'},lastPublished:id==='b2'?null:{status:'published'}}},
    retryBatchAutomation:async(...args)=>events.push(['retry',...args]),
    publishTextCollection:async(...args)=>events.push(['manual',...args])
  };
  const items=['b','manual','cancel','b2'].map(id=>({book:{id,title:id},stage:'text_publish',
    automatic:id!=='manual',summary:id==='manual'?{textPublishSource:'manual',textPublishRecord:{collectionId:'manual-current'},runs:[{stage:'text_publish',collectionId:'older-auto'}]}:
      {runs:[{stage:'text_publish',id:'run-'+id,collectionId:id==='b2'?'':'c-'+id}]}}));
  const result=await helpers.retryTextBooks(api,'batch',items,{
    chooseCollection:async(book)=>{events.push('select:'+book.id);return 'chosen-'+book.id},
    confirmRepublish:book=>{events.push('confirm:'+book.id);return book.id!=='cancel'}
  });
  assert.equal(result.submitted,3);
  assert.deepEqual(result.errors,[]);
  const firstWrite=events.findIndex(Array.isArray);
  assert.ok(events.slice(0,firstWrite).includes('select:b2'));
  assert.ok(events.slice(0,firstWrite).includes('confirm:cancel'));
  assert.deepEqual(events.filter(Array.isArray),[
    ['retry','batch',['b','b2'],{b:{collectionId:'c-b',force:true},b2:{collectionId:'chosen-b2',force:false}}],
    ['manual','batch','manual','manual-current',{force:true}]
  ]);
});

test('失败的自动重传仍带有旧成功回执时，取消保留旧回执，确认绑定具体合集', async () => {
  const old={kind:'text',bookId:'book',collectionId:'c',status:'published',runId:'old',attempt:1,requestKey:'old-key'};
  const run={stage:'text_publish',id:'new',collectionId:'c',attempt:2,requestId:'new-key',status:'failed'};
  const calls=[];
  const api={getTextPublishStatus:async()=>({published:true,publish:old,lastPublished:old}),
    retryBatchAutomation:async(...args)=>calls.push(args)};
  const item={book:{id:'book'},automatic:true,stage:'text_publish',summary:{runs:[run],textLastPublishedRecord:old}};
  const retry=approved=>helpers.retryTextBooks(api,'batch',[item],{confirmRepublish:()=>approved});
  assert.equal((await retry(false)).submitted,0);
  assert.equal(calls.length,0);
  assert.equal((await retry(true)).submitted,1);
  assert.deepEqual(calls,[['batch',['book'],{book:{collectionId:'c',force:true}}]]);
});

test('已选文本进入自动生产接口，日志收录手动上传回执',()=>{
  assert.ok(list.includes("if (batchProductionMode === 'text') {\n        const started = await startBatchAutomation("));
  assert.ok(list.includes("summary?.textPublishRecord"));
  assert.ok(list.includes('runMode: automationStatus?.runMode, textAutoUpload: batch?.settingsState?.patch?.textAutoUpload'));
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
async function runChrome(contents) {
  const dir = mkdtempSync(join(tmpdir(),'text-publish-'));
  try {
    const result = await build({stdin:{resolveDir:here,loader:'jsx',contents},jsx:'automatic',bundle:true,write:false,format:'iife',loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'},minify:true});
    writeFileSync(join(dir,'app.js'),result.outputFiles[0].contents);
    const html=join(dir,'index.html');
    writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="app.js"></script>');
    const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=16000',
      '--dump-dom',pathToFileURL(html).href],{timeout:45000,windowsHide:true,encoding:'utf8',stdio:['ignore','pipe','pipe']});
    assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || '浏览器未完成验证');
  } finally { rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150}); }
}

test('Chrome 点击上传、取消和确认再次上传，等待回读期间不会重复提交', {skip:!browser}, async () => {
  await runChrome(`
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
            status={published:false,publish:{status:'pending',message:'等待平台返回合集回执；如长时间未更新，请到发布后台核对。'}};
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
        if(!document.body.innerText.includes('等待平台返回合集回执；如长时间未更新，请到发布后台核对。'))throw Error('缺少平台回读提示');
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
    `);
});

test('Chrome 在生产表重试手动上传、打开日志并立即执行所选文本书', {skip:!browser}, async () => {
  await runChrome(`
    import React from 'react';
    import {createRoot} from 'react-dom/client';
    import {BatchFactoryNovelList} from './BatchFactoryNovelList.jsx';
    const sleep=n=>new Promise(done=>setTimeout(done,n));
    const buttons=()=>Array.from(document.querySelectorAll('button'));
    const button=text=>buttons().find(node=>node.textContent.trim()===text);
    const posts=[];
    let step='初始化',approved=false,confirms=0;
    window.confirm=()=>{confirms++;return approved};
    window.onerror=(message,_url,_line,_col,error)=>{document.body.dataset.crash=String(error?.message || message)};
    class Boundary extends React.Component {
      state={error:''};
      static getDerivedStateFromError(error){return {error:error.message}}
      render(){return this.state.error ? <div>{this.state.error}</div> : this.props.children}
    }
    async function until(test){for(let i=0;i<130;i++){if(test())return;await sleep(25)}throw Error(step+' '+(document.body.dataset.crash || document.body.innerText.slice(-750)))}
    const book={id:'b',title:'手动上传的书',bookId:'101',sourceText:'正文。',txtText:'正文。',settingsState:{patch:{}},videos:[]};
    const batch={id:'batch',revision:1,title:'文本批次',settingsState:{patch:{productionMode:'text'}},books:[book]};
    const history={kind:'text',bookId:'b',collectionId:'c',status:'published',carouselId:1,imageIds:[2]};
    const runtime={stageSummaries:{b:{textPublishSource:'manual',textPublishRecord:{kind:'text',bookId:'b',collectionId:'c',status:'failed'},textLastPublishedRecord:history}},automation:{state:'idle',books:[]},production:{jobs:[]},merge:{jobs:[]}};
    window.fetch=async(url,opts={})=>{
      const path=String(url).split('?')[0],method=opts.method||'GET';
      const payload=opts.body?JSON.parse(opts.body):null;
      if(method==='POST')posts.push({path,payload});
      let data={};
      if(path.endsWith('runtime-summary'))data=runtime;
      else if(path.endsWith('text-collections/c/publish'))data={published:false,publish:{status:'failed'},lastPublished:history};
      else if(path.endsWith('/books/b2/text-collections'))data={items:[{id:'do-not-guess'},{id:'choose-b2'}]};
      else if(path.endsWith('/text-collections/choose-b2/publish'))data={published:false,publish:null};
      else if(path.endsWith('/automation/retry'))data={automation:{state:'running'}};
      else if(path.endsWith('automation-presets'))data={presets:[{id:'preset',name:'文本预设',config:{productionMode:'text'}}]};
      else if(path.endsWith('automation/preflight'))data={readOnly:true,books:[{bookId:'b',passed:true}]};
      else if(path.endsWith('automation/start'))data={automation:{state:'running',books:[{bookId:'b',stage:'text_process',status:'running'}]}};
      else if(path.endsWith('config-versions'))data={configVersions:[]};
      return new Response(JSON.stringify(data),{headers:{'Content-Type':'application/json'}});
    };
    async function run(){
      const root=createRoot(document.querySelector('#root'));
      root.render(<Boundary><BatchFactoryNovelList batch={batch} onBatchChanged={async()=>batch}/></Boundary>);
      step='重试取消';await until(()=>button('重试失败步骤'));button('重试失败步骤').click();await until(()=>confirms===1);
      if(posts.some(x=>x.path.endsWith('/publish')))throw Error('取消后仍提交上传');
      step='确认重传';approved=true;await until(()=>!button('重试失败步骤').disabled);button('重试失败步骤').click();await until(()=>posts.some(x=>x.path.endsWith('/text-collections/c/publish')&&x.payload.force));
      step='手动日志';button('任务/日志').click();const logs=()=>Array.from(document.querySelectorAll('.ant-modal')).find(n=>n.textContent.includes('文本任务 / 日志'));
      await until(()=>logs()?.textContent.includes('手动上传 · 手动上传的书'));if(!logs().textContent.includes('上传失败'))throw Error('缺少手动上传失败状态');
      logs().querySelector('.ant-modal-close').click();await sleep(350);
      step='选择书';document.querySelector('.batch-factory-book-row input[type=checkbox]').click();await sleep(60);
      step='批量操作';button('批量操作').dispatchEvent(new MouseEvent('mouseover',{bubbles:true}));await until(()=>Array.from(document.querySelectorAll('[role=menuitem]')).some(n=>n.textContent.includes('立即执行已选')));
      Array.from(document.querySelectorAll('[role=menuitem]')).find(n=>n.textContent.includes('立即执行已选')).click();
      step='已选启动';await until(()=>posts.some(x=>x.path.endsWith('/automation/start')));
      if(posts.some(x=>x.path.includes('/executions')) || posts.find(x=>x.path.endsWith('/automation/start')).payload.bookIds.join(',')!=='b')throw Error('提交到错误模式或书目');
      step='两本书一次重试';
      const book2={...book,id:'b2',bookId:'102',title:'第二本'};
      batch.books=[book,book2];
      const failed=id=>({bookId:id,stage:'text_publish',title:id,status:'failed',error:'上传失败'});
      runtime.automation={batchId:'batch',state:'failed',books:[failed('b'),failed('b2')],counts:{total:2,failed:2}};
      runtime.stageSummaries={
        b:{runs:[{stage:'text_publish',status:'failed',id:'run-b',collectionId:'c'}],textLastPublishedRecord:history},
        b2:{runs:[{stage:'text_publish',status:'failed',id:'run-b2'}]}
      };
      root.render(<Boundary key="batch-retry"><BatchFactoryNovelList batch={batch} onBatchChanged={async()=>batch}/></Boundary>);
      await until(()=>buttons().filter(n=>n.textContent.trim()==='重试失败步骤').length===2);
      const menu=buttons().find(n=>n.textContent.trim().startsWith('自动生产'));
      menu.dispatchEvent(new MouseEvent('mouseover',{bubbles:true}));
      const retryMenu=()=>Array.from(document.querySelectorAll('[role=menuitem]')).find(n=>n.textContent.includes('重试失败'));
      await until(retryMenu); retryMenu().click();
      const select=()=>document.querySelector('select[aria-label="重试文本合集"]');
      await until(select);
      if(posts.some(x=>x.path.endsWith('automation/retry')))throw Error('完成选集之前已提交');
      select().value='choose-b2';select().dispatchEvent(new Event('change',{bubbles:true}));
      await until(()=>button('使用此合集')&&!button('使用此合集').disabled);button('使用此合集').click();
      await until(()=>posts.some(x=>x.path.endsWith('automation/retry')));
      const requests=posts.filter(x=>x.path.endsWith('automation/retry'));
      const sent=requests[0].payload, forced=sent.textPublishOverrides?.b, normal=sent.textPublishOverrides?.b2;
      if(requests.length!==1||sent.bookIds.join(',')!=='b,b2'||forced?.collectionId!=='c'||forced.force!==true||
        normal?.collectionId!=='choose-b2'||normal.force!==false)throw Error('两本书未按各自选集和确认状态一起提交');
      document.body.dataset.result='passed';root.unmount();
    }
    run().catch(e=>document.body.dataset.error=e.message);
  `);
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync, existsSync, mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { execFileSync } from 'node:child_process';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';
import { batchFactoryProjectsFrom } from './batchFactoryProjects.js';

test('项目和书目将文本交给文字媒体库', () => {
  assert.equal(batchFactoryProjectsFrom([{id:'b',productionMode:'text'}])[0].factoryMode, 'text');
  const list = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8');
  assert.match(list, /mediaMode === 'text' \? <TextMediaDialog\b/);
  assert.match(list, /mediaMode === 'comic' \? <ComicMediaPanel/);
  const library = readFileSync(new URL('./ProjectsView.jsx', import.meta.url), 'utf8');
  assert.match(library, /project.factoryMode === 'text'.*<AuthenticatedTextPageImage/);
  assert.match(library, /value:'text',label:'文本'/);
});

test('文本合集预览只列文字页，不把作品封面放进合集', () => {
  const source = readFileSync(new URL('./TextMediaPanel.jsx', import.meta.url), 'utf8');
  assert.match(source, /collectionItems\[0\]\?\.contentUrl/);
  assert.doesNotMatch(source, /currentCollection\?\.coverContentUrl/);
  assert.doesNotMatch(source, /setCollectionImage\(''\)>封面/);
});

const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(path => path && existsSync(path));
test('媒体库标签、翻页、刷新保留选中页，仅在生成中轮询', {skip:!browser}, async () => {
  assert.ok(existsSync(new URL('./TextMediaPanel.jsx', import.meta.url)), '缺少文本媒体库');
  const dir = mkdtempSync(join(tmpdir(), 'text-library-'));
  try {
    const result = await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
      import React from 'react';
      import { createRoot } from 'react-dom/client';
      import TextMediaDialog from './TextMediaDialog.jsx';
      window.addEventListener('error',event=>{document.body.dataset.runtimeError=event.error?.message || event.message});
      const root = createRoot(document.getElementById('root'));
      const sleep = t=>new Promise(r=>setTimeout(r,t));
      async function until(f){for(let i=0;i<300;i++){if(f())return;await sleep(20)}throw Error(document.body.innerText)}
      let reads=0, rendering=true,requests=[],previewBodies=[],saved,aborted=false,restores=0,locked=false,coverLocked=false;
      const mutations=[];
      let confirmRestore=false, confirmCount=0, releaseSave, single=true;
      let bookRevision=7, bookWrites=0, ready=false, collection, publishes=0;
      window.confirm=()=>{confirmCount++;return confirmRestore};
      window.fetch=async(path,opts)=>{
        requests.push([path,opts.method]); if(opts.method==='POST')mutations.push([path,opts.body]);
        let items=[];
        if(path.endsWith('/preview')){
          const body=JSON.parse(opts.body);
          previewBodies.push(body);
          if(body.text==='较慢的草稿。') return new Promise((resolve,reject)=>opts.signal.addEventListener('abort',()=>{aborted=true;reject(new DOMException('取消','AbortError'))},{once:true}));
          if(body.text==='太长的草稿。')return new Response('{"error":"文字超出页面"}',{status:400});
          return new Response(new Blob([Uint8Array.from(atob('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='),c=>c.charCodeAt(0))],{type:'image/png'}),{headers:{'X-Text-Contrast':'low'}});
        }else if(opts.method==='PATCH'){saved=JSON.parse(opts.body);await new Promise(resolve=>{releaseSave=resolve});
          items={affectedPageIds:['p2','downstream'],pages:[{id:'p1',revision:2},{id:'p2',revision:4,ordinal:2,status:'pending'},{id:'downstream',revision:5,ordinal:3,status:'pending'}]};}
        else if(opts.method==='PUT'){
          const data=JSON.parse(opts.body);if(data.expectedRevision!==bookRevision || data.patch.textBackground?.kind!=='color')throw Error('整书背景版本错误');
          bookWrites++;items={revision:++bookRevision,patch:data.patch};
        }
        else if(path.includes('/upload?')){if(path.includes('text-cover'))coverLocked=true;else locked=true;items={id:'new-art'};}
        else if(path.endsWith('/restore')){restores++;if(path.includes('text-cover'))coverLocked=false;else locked=false;items={document:{revision:5},pages:[{id:'p2',revision:5}]};}
        else if(path.endsWith('/split')||path.endsWith('/merge'))items={pages:[{id:'p1'},{id:'p2'}]};
        else if(path.endsWith('text-pages')){
          reads++;
          items={document:{id:'doc',revision:4,activeVersionId:'v',paginationVersion:1,coverLocked},artifacts:[],sentenceBreaks:[6,9],pages:[
            {id:'p1',ordinal:1,startOffset:0,endOffset:9,revision:2,status:rendering?'rendering':'pending',text:'第一段完整。后句。'},
            {id:'p2',ordinal:2,status:'pending',text:saved?.text || '第二段完整。',revision:3,locked}]};
          if(ready){
            items.document.coverArtifactId='cover';
            items.artifacts=[{id:'cover',kind:'cover',sourceKind:'system',renderVersion:1}];
            for(const p of items.pages){
              Object.assign(p,{status:'ready',versionId:'v',paginationVersion:1,renderVersion:1,activeArtifactId:p.id+'art'});
              items.artifacts.push({id:p.activeArtifactId,pageId:p.id,kind:'page',sourceKind:'system',renderVersion:1});
            }
          }
          if(single){items.sentenceBreaks=reads===1?null:[];items.pages=[{...items.pages[0],text:'只有一句。',status:'pending',endOffset:5}]}
        }else if(path.endsWith('text-collections')){
          if(opts.method==='POST'){
            const data=JSON.parse(opts.body);
            if(data.pageIds.join(',')!=='p1,p2'||data.versionId!=='v'||data.expectedRevision!==4)throw Error('合集版本/范围不对');
            collection={id:'c',status:'ready'};items=collection;
          }else items={items:collection?[collection]:[]};
        }else if(path.endsWith('text-collections/c'))items=collection;
        return new Response(JSON.stringify(items),{headers:{'Content-Type':'application/json'}})
      };
      const button=(label)=>Array.from(document.querySelectorAll('button')).find(b=>b.getAttribute('aria-label')===label||b.textContent.trim()===label);
      async function run(){
        root.render(<TextMediaDialog batchId="a" book={{id:'b',title:'测试书',revision:7}} onUpload={()=>{publishes++;collection.status='uploaded'}} onClose={()=>root.unmount()} />);
        await until(()=>document.body.innerText.includes('只有一句。'));
        if(!button('在完整句子后拆页').disabled)throw Error('单句页仍能拆分');
        button('更新文字页').click();await until(()=>reads>=2&&button('更新文字页')?.disabled===false);
        if(!button('在完整句子后拆页').disabled)throw Error('空句末列表仍可拆分');
        single=false;button('更新文字页').click();
        await until(()=>document.body.innerText.includes('第一段完整。'));
        for(const label of ['本页背景','页面内容','封面','文字页','文本合集','页目录','合成当前书','上传网络','页面操作','拆分与合并','整书操作'])if(!document.body.innerText.includes(label))throw Error('缺少 '+label);
        await sleep(350);
        for(const group of document.querySelectorAll('.text-media-operation-grid')){
          const boxes=[...group.children].map(item=>item.getBoundingClientRect());
          if(boxes.some(box=>box.width<70 || box.height<35))throw Error('操作按钮尺寸不一致 '+boxes.map(box=>Math.round(box.width)+'x'+Math.round(box.height)).join(','));
          if(Math.max(...boxes.map(box=>box.top))-Math.min(...boxes.map(box=>box.top))>1)throw Error('分组按钮未对齐');
          if(group.scrollWidth>group.clientWidth+2)throw Error('操作分组横向溢出');
        }
        if(!button('合成当前书').disabled)throw Error('文字页未就绪时未禁用合成');
        button('下一页').click();
        await until(()=>document.body.innerText.includes('第二段完整。'));
        const beforePoll=reads;rendering=false;await until(()=>reads>beforePoll);
        if(!document.body.innerText.includes('第二段完整。'))throw Error('刷新后丢失当前页');
        const n=reads;await sleep(4500);if(reads!==n)throw Error('无生成中页面仍在轮询');
        const edit=button('排版调整');if(!edit)throw Error('缺少当前页排版调整');
        edit.click();
        const area=()=>document.querySelector('textarea[aria-label="页面正文"]');
        await until(area);
        const change=(s)=>{Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype,'value').set.call(area(),s);area().dispatchEvent(new Event('input',{bubbles:true}))};
        change('较慢的草稿。');await until(()=>previewBodies.some(p=>p.text==='较慢的草稿。'));
        change('较新的草稿。');await until(()=>aborted && document.querySelector('img[alt="排版预览"]'));
        if(!document.body.innerText.includes('对比度较低'))throw Error('未显示颜色提示');
        const last=document.querySelector('img[alt="排版预览"]').src;
        change('太长的草稿。'); await until(()=>document.body.innerText.includes('文字超出页面'));
        if(document.querySelector('img[alt="排版预览"]').src!==last)throw Error('预览失败丢失上一张图');
        change('最终手改。'); await until(()=>previewBodies.some(p=>p.text==='最终手改。'));
        const esc=()=>document.querySelector('.ant-modal-wrap').dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',keyCode:27,bubbles:true}));
        const mask=()=>document.querySelector('.ant-modal-wrap').click();
        const close=()=>button('关闭文本媒体库').click();
        for(const request of [esc,mask,close]){
          const n=confirmCount;request();await sleep(35);
          if(confirmCount!==n+1 || area()?.value!=='最终手改。')throw Error('未保存关闭保护失效');
        }
        button('保存本页').click();await until(()=>releaseSave);
        const nDuringSave=confirmCount;confirmRestore=true;
        for(const request of [esc,mask,close]){request();await sleep(35);}
        if(confirmCount!==nDuringSave || !area())throw Error('保存中仍可关闭');
        confirmRestore=false;releaseSave();await until(()=>saved && !area());
        if(saved.text!=='最终手改。'||saved.expectedRevision!==3||saved.expectedDocumentRevision!==4||saved.reflow!==true)throw Error('保存未带草稿、局部分页和两个版本号');
        await until(()=>mutations.some(([path,b])=>path.endsWith('/p2/render')&&JSON.parse(b).expectedRevision===4));
        await until(()=>mutations.some(([path,b])=>path.endsWith('/downstream/render')&&JSON.parse(b).expectedRevision===5));
        for(let i=0;i<2;i++){
          button('排版调整').click();await until(area);
          button('背景应用到整本书').click();await until(()=>!area());
        }
        const prep=mutations.filter(([path])=>path.endsWith('text-prepare'));
        if(prep.length!==4||bookWrites!==2||prep.some(([,b])=>JSON.parse(b).reflow))throw Error('整书背景触发错误分页');
        const upload=label=>{
          const input=document.querySelector('input[type=file][aria-label="'+label+'"]');if(!input)throw Error('缺少'+label);
          const transfer=new DataTransfer();transfer.items.add(new File(['test'],'用户成品.png',{type:'image/png'}));
          input.files=transfer.files;input.dispatchEvent(new Event('change',{bubbles:true}));
        };
        upload('上传成品页');await until(()=>button('恢复系统生成'));
        button('恢复系统生成').click();await sleep(80);if(restores!==0)throw Error('取消后仍然恢复');
        confirmRestore=true;button('恢复系统生成').click();await until(()=>!button('恢复系统生成')&&restores===1);
        if(!mutations.some(([url,body])=>url.endsWith('/p2/render')&&JSON.parse(body).expectedRevision===5))throw Error('未用解锁后的页版本重新出图');
        button('上一页').click();await until(()=>button('在完整句子后拆页')?.disabled===false);
        button('在完整句子后拆页').click();await until(()=>mutations.some(([url,b])=>url.endsWith('/p1/split')&&JSON.parse(b).offset===6&&JSON.parse(b).expectedRevision===4));
        await until(()=>button('与下一页合并')?.disabled===false);button('与下一页合并').click();
        await until(()=>mutations.some(([url,b])=>url.endsWith('/p1/merge')&&JSON.parse(b).expectedRevision===4));
        await until(()=>button('封面')?.disabled===false);
        button('封面').click();await until(()=>document.body.innerText.includes('测试书')&&!document.body.innerText.includes('第二段完整。'));
        upload('上传成品封面');await until(()=>button('恢复系统生成'));
        confirmRestore=false;button('恢复系统生成').click();await sleep(80);if(restores!==1)throw Error('取消恢复封面未生效');
        confirmRestore=true;button('恢复系统生成').click();await until(()=>!button('恢复系统生成')&&restores===2);
        if(!mutations.some(([url,b])=>url.endsWith('text-cover/render')&&JSON.parse(b).expectedRevision===5))throw Error('未重新生成封面');
        button('文本合集').click();await until(()=>document.body.innerText.includes('暂无文本合集'));
        ready=true;button('文字页').click();await until(()=>button('更新文字页'));
        button('更新文字页').click();await until(()=>button('合成当前书')?.disabled===false);
        button('合成当前书').click();await until(()=>button('重新合成')&&button('上传网络')?.disabled===false);
        button('上传网络').click();await until(()=>button('再次上传'));
        confirmRestore=false;button('再次上传').click();await sleep(50);if(publishes!==1)throw Error('取消后重复上传');
        confirmRestore=true;await until(()=>button('再次上传')?.disabled===false);button('再次上传').click();await sleep(100);
        await until(()=>publishes===2&&button('关闭文本媒体库')?.disabled===false);
        esc();await until(()=>!document.querySelector('.text-media-panel'));
        if(requests.some(([url,m])=>m!=='PUT'&&!url.includes('/text-')&&!url.includes('/client-errors')))throw Error('文本库请求了其他模式接口');
        document.body.dataset.result='passed'
      }run().catch(e=>document.body.dataset.error=e.message+'（上传 '+publishes+' 次）');
    `},bundle:true,jsx:'automatic',write:false,format:'iife',loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'},minify:true});
    writeFileSync(join(dir,'app.js'),result.outputFiles[0].contents);
    const html=join(dir,'index.html');
    const styles=readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8')+'\n'+readFileSync(new URL('./text-media-panel.css',import.meta.url),'utf8');
    writeFileSync(join(dir,'styles.css'),styles);
    writeFileSync(html,'<meta charset="utf-8"><link rel="stylesheet" href="styles.css"><style>.ant-modal{animation:none!important;transform:none!important;transition:none!important}</style><div id="root"></div><script>'+result.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
    const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=24000',
      '--dump-dom',pathToFileURL(html).href],{timeout:45000,windowsHide:true,encoding:'utf8',stdio:['ignore','pipe','pipe']});
    assert.ok(dom.includes('data-result="passed"'), dom.match(/data-runtime-error="([^"]*)"/)?.[1] || dom.match(/data-error="([^"]*)"/)?.[1] || '浏览器未完成验证');
  } finally { rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150}); }
});

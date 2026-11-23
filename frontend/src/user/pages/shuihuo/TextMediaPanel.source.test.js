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
  assert.match(list, /mediaMode === 'text' \? <TextMediaPanel\b/);
  assert.match(list, /mediaMode === 'comic' \? <ComicMediaPanel/);
  const library = readFileSync(new URL('./ProjectsView.jsx', import.meta.url), 'utf8');
  assert.match(library, /project.factoryMode === 'text'.*<AuthenticatedTextPageImage/);
  assert.match(library, /value:'text',label:'文本'/);
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
      import TextMediaPanel from './TextMediaPanel.jsx';
      const root = createRoot(document.getElementById('root'));
      const sleep = t=>new Promise(r=>setTimeout(r,t));
      async function until(f){for(let i=0;i<300;i++){if(f())return;await sleep(20)}throw Error(document.body.innerText)}
      let reads=0, rendering=true,requests=[];
      window.fetch=async(path,opts)=>{
        requests.push([path,opts.method]);
        let items=[];
        if(path.endsWith('text-pages')){
          reads++;
          items={document:{id:'doc',revision:4,activeVersionId:'v',paginationVersion:1},artifacts:[],pages:[
            {id:'p1',ordinal:1,status:rendering?'rendering':'pending',text:'第一段完整。'},
            {id:'p2',ordinal:2,status:'pending',text:'第二段完整。'}]};
        }else if(path.endsWith('text-collections'))items={items:[]};
        return new Response(JSON.stringify(items),{headers:{'Content-Type':'application/json'}})
      };
      const button=(label)=>Array.from(document.querySelectorAll('button')).find(b=>b.getAttribute('aria-label')===label||b.textContent.trim()===label);
      async function run(){
        root.render(<TextMediaPanel batchId="a" book={{id:'b',title:'测试书'}} onClose={()=>root.unmount()} />);
        await until(()=>document.body.innerText.includes('第一段完整。'));
        for(const label of ['本页背景','页面内容','封面','文字页','文本合集','展开','合成当前书','上传网络'])if(!document.body.innerText.includes(label))throw Error('缺少 '+label);
        if(!button('合成当前书').disabled)throw Error('缺封面时未禁用合成');
        button('下一页').click();
        await until(()=>document.body.innerText.includes('第二段完整。'));
        rendering=false;await until(()=>reads>=2);
        if(!document.body.innerText.includes('第二段完整。'))throw Error('刷新后丢失当前页');
        const n=reads;await sleep(4500);if(reads!==n)throw Error('无生成中页面仍在轮询');
        button('封面').click();await until(()=>document.body.innerText.includes('测试书')&&!document.body.innerText.includes('第二段完整。'));
        button('文本合集').click();await until(()=>document.body.innerText.includes('暂无文本合集'));
        button('关闭文本媒体库').click();await until(()=>!document.getElementById('root').children.length);
        if(requests.some(([url])=>!url.includes('/text-')))throw Error('文本库请求了其他模式接口');
        document.body.dataset.result='passed'
      }run().catch(e=>document.body.dataset.error=e.message);
    `},bundle:true,write:false,format:'iife',loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'},minify:true});
    writeFileSync(join(dir,'app.js'),result.outputFiles[0].contents);
    const html=join(dir,'index.html');
    writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="app.js"></script>');
    const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=16000',
      '--dump-dom',pathToFileURL(html).href],{timeout:45000,windowsHide:true,encoding:'utf8',stdio:['ignore','pipe','pipe']});
    assert.ok(dom.includes('data-result="passed"'), dom.match(/data-error="([^"]*)"/)?.[1] || '浏览器未完成验证');
  } finally { rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150}); }
});

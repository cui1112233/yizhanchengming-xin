import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';
import { build } from 'esbuild';

const browser = [process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(p=>p&&existsSync(p));

test('认证预览只显示当前页并回收图片，拒绝磁盘地址，失败可重试', {skip:!browser}, async () => {
  const result=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
    import React from 'react';
    import {createRoot} from 'react-dom/client';
    import Preview from './AuthenticatedTextPageImage.jsx';
    const root=createRoot(document.querySelector('#root'));
    const show=url=>root.render(<Preview contentUrl={url} alt="文字页"/>);
    const wait=async f=>{for(let i=0;i<80;i++){if(f())return;await new Promise(r=>setTimeout(r,20));}throw Error('等待超时 '+document.body.innerText)};
    const base='/api/batch-factory/v12/batches/one/books/'+encodeURIComponent('书 2')+'/text-artifacts/';
    const png=Uint8Array.from(atob('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aFYkAAAAASUVORK5CYII='),s=>s.charCodeAt(0));
    const reply=()=>new Response(png,{headers:{'Content-Type':'image/png'}});
    async function run(){
      const revoked=[], created=[], calls=[];
      const create=URL.createObjectURL.bind(URL), revoke=URL.revokeObjectURL.bind(URL);
      URL.createObjectURL=b=>{const u=create(b);created.push(u);return u};
      URL.revokeObjectURL=u=>{revoked.push(u);revoke(u)};
      let resolveSlow, fail=true;
      localStorage.setItem('auth_token','fixture');
      window.fetch=async(path,opts)=>{
        if(path==='/api/client-errors')return new Response('{}');
        calls.push([path,opts.headers.get('Authorization'),opts.signal]);
        if(path.includes('/slow/'))return new Promise(resolve=>resolveSlow=resolve);
        if(path.includes('/again/') && fail){fail=false;return new Response('{}',{status:503});}
        return reply();
      };
      show(base+'slow/content');await wait(()=>resolveSlow);
      show(base+'fast/content');await wait(()=>document.querySelector('img')?.naturalWidth===1);
      if(!calls[0][2].aborted || calls[1][1]!=='Bearer fixture')throw Error('未取消或未认证');
      resolveSlow(reply());
      await new Promise(r=>setTimeout(r,50));
      if(created.length!==1)throw Error('旧请求生成了额外图片');
      show('file:///private/secret.png');await wait(()=>!document.querySelector('img')&&revoked.includes(created[0]));
      if(calls.length!==2 || document.body.innerHTML.includes('secret.png'))throw Error('不应读取或展示磁盘路径');
      show(base+'again/content');await wait(()=>document.querySelector('button'));
      document.querySelector('button').click();
      await wait(()=>document.querySelector('img')?.naturalWidth===1);
      root.unmount();await wait(()=>revoked.length===2);
      document.body.dataset.result='passed';
    }
    run().catch(e=>{document.body.dataset.error=e.message});
  `},bundle:true,write:false,format:'iife',minify:true,define:{'process.env.NODE_ENV':'"production"'}});
  const dir=mkdtempSync(join(tmpdir(),'text-preview-test-'));
  try {
    writeFileSync(join(dir,'app.js'),result.outputFiles[0].contents);
    const html=join(dir,'index.html');
    writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="app.js"></script>');
    const dom=execFileSync(browser,['--headless=new','--disable-gpu','--disable-background-networking','--no-first-run',
      '--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=12000',
      '--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:45000,windowsHide:true,stdio:['ignore','pipe','pipe']});
    assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || '预览测试未完成');
  } finally {
    rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:200});
  }
});

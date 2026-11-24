import test from 'node:test';
import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,readFileSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {pathToFileURL} from 'node:url';
const browser='C:/Program Files/Google/Chrome/Application/chrome.exe';
test('glass preset save remains readable and enabled, while unavailable actions stay disabled',{skip:!existsSync(browser)},()=>{
 const dir=mkdtempSync(join(tmpdir(),'preset-button-'));
 try {
  const css=readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8')+readFileSync(new URL('../../../shared/styles/cinematic-glass.css',import.meta.url),'utf8');
  const html=join(dir,'index.html');
  writeFileSync(html,`<html data-scheme="glass" data-theme="dark"><meta charset="utf-8"><style>
   :root{--surface-base:#111318;--surface-panel:#202833;--surface-2:#303845;--border-default:#58616e;--text-primary:#f0f3f8;--text-tertiary:#9aa4b4;--text-on-accent:#10251f;--theme-primary:#63dfc0;--glass-panel:#29313f;--glass-edge:#566170}
   ${css}
  </style><body class="shuihuo-theme-active"><div class="batch-factory-automation-preset-modal"><button id="save" class="ant-btn ant-btn-primary">保存为新预设</button><button id="rename" class="ant-btn" disabled>重命名所选预设</button></div><script>
   const button=document.getElementById('save');
   const color=getComputedStyle(button).color.match(/\\d+/g).map(Number);
   if(Math.min(...color.slice(0,3))<180)document.body.dataset.error='active primary text is too dark';
   else if(!document.getElementById('rename').disabled)document.body.dataset.error='disabled action was enabled';
   else{button.onclick=()=>document.body.dataset.result='passed';button.click()}
  </script></html>`);
  const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=1000','--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:30000,windowsHide:true,stdio:['ignore','pipe','pipe']});
  assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1]||'button check failed');
 }finally{rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150});}
});

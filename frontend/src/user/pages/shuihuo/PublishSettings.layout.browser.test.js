import assert from 'node:assert/strict';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,readFileSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
import test from 'node:test';
import {build} from 'esbuild';

const browser=['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
test('video comic and text publish forms share aligned fields without horizontal overflow',{skip:!browser},async()=>{
 const dir=mkdtempSync(join(tmpdir(),'publish-layout-'));
 try {
  const result=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
   import React from 'react';import {createRoot} from 'react-dom/client';
   import {BatchFactoryPublishSettingsForm} from './BatchFactoryUnifiedSettingsModal.jsx';
   import {BatchFactoryComicPublishSettingsForm} from './BatchFactoryImagePublishSettingsForm.jsx';
   window.fetch=async()=>Response.json({producerAccounts:['作者一'],producerAccountsWarning:'N8 目录暂时不可用（上游 code=UN_LOGIN：登录已失效），请更新后台登录令牌后重试。',organizations:[{id:'1',name:'组织一'}],audioCategories:[],models:[],groups:[],items:[]});
   const root=createRoot(document.getElementById('root'));
   async function run(){
    for(const width of [920,600]){
     document.getElementById('root').style.width=width+'px';
     for(const mode of ['video','comic','text']){
      root.render(mode==='video'?<BatchFactoryPublishSettingsForm value={{}} onChange={()=>{}} active={false}/>:<BatchFactoryComicPublishSettingsForm key={mode+width} value={{}} onChange={()=>{}} singleBook productionMode={mode}/>);
      await new Promise(r=>setTimeout(r,250));
      const form=document.querySelector('.batch-factory-publish-layout');if(!form)throw Error(mode+' missing shared layout');
      const fields=[...form.querySelectorAll('.batch-factory-engine-field')];
      if(fields.length<2)throw Error(mode+' missing controls');
      for(const field of fields){
       const label=field.querySelector(':scope > span');const control=field.querySelector(':scope > .ant-select,:scope > .ant-segmented');
       if(label&&control&&control.getBoundingClientRect().top<label.getBoundingClientRect().bottom-1)throw Error(mode+' label overlaps control');
       if(field.scrollWidth>field.clientWidth+2)throw Error(mode+' field overflow');
      }
      if(form.scrollWidth>form.clientWidth+2)throw Error(mode+' form overflow');
      const grid=form.querySelector('.batch-factory-engine-card-body');const cols=getComputedStyle(grid).gridTemplateColumns.split(' ').length;
      if(cols!==(innerWidth<=780?1:2))throw Error(mode+' unexpected column count '+cols);
     }
    }
    document.body.dataset.result='passed';
   }
   run().catch(e=>{document.body.dataset.error=e.message});
  `},bundle:true,loader:{'.css':'empty'},jsx:'automatic',write:false,format:'iife',minify:true,define:{'process.env.NODE_ENV':'"production"'}});
  const css=readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8')+'\n'+readFileSync(new URL('./publish-settings.css',import.meta.url),'utf8');
  writeFileSync(join(dir,'styles.css'),css);
  writeFileSync(join(dir,'index.html'),'<meta charset="utf-8"><link rel="stylesheet" href="styles.css"><style>body{--sh-panel:#242a35;--sh-panel-strong:#1b2330;--sh-line:#536071;--sh-text:#eef3f8;--sh-muted:#b9c5d0;--sh-accent:#63dfc0}#root{margin:0 auto}</style><div id="root"></div><script>'+result.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
  for(const width of [1021,720]){
   const dom=execFileSync(browser,['--headless=new','--disable-gpu','--disable-background-networking','--no-first-run','--host-resolver-rules=MAP * ~NOTFOUND','--window-size='+width+',901','--virtual-time-budget=10000','--user-data-dir='+join(dir,'profile-'+width),'--dump-dom',pathToFileURL(join(dir,'index.html')).href],{encoding:'utf8',timeout:60000,maxBuffer:12<<20,windowsHide:true,stdio:['ignore','pipe','pipe']});
   assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || 'browser did not finish');
  }
 } finally {rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:200});}
});

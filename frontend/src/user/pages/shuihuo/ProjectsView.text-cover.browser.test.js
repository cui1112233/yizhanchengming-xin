import test from 'node:test';
import assert from 'node:assert/strict';
import {build} from 'esbuild';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,readFileSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
const browser=['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
test('text work cards display a private default book cover and the same mode badge as comic and video',{skip:!browser},async()=>{
 const dir=mkdtempSync(join(tmpdir(),'text-card-'));
 try{
 const output=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
 import React from 'react';import {createRoot} from 'react-dom/client';import {ProjectsView} from './ProjectsView.jsx';
 const seen=[];window.IntersectionObserver=class{constructor(cb){this.cb=cb}observe(){this.cb([{isIntersecting:true}])}disconnect(){}};
 window.fetch=async(url)=>{seen.push(url);if(url.endsWith('/text-cover/display'))return new Response(new Blob([Uint8Array.from(atob('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='),c=>c.charCodeAt(0))],{type:'image/png'}));return new Response('[]',{headers:{'Content-Type':'application/json'}})};
 const projects=['text','comic','video'].map(mode=>({id:'batch:'+mode,batchId:mode,name:mode+'批次',productionMode:'batch_factory',factoryMode:mode,source:'batch_factory_v11',batch:{books:[{id:'book',title:'封面使用书名'}]}}));
 const sleep=ms=>new Promise(r=>setTimeout(r,ms));async function run(){createRoot(document.querySelector('#root')).render(<ProjectsView projects={projects}/>);for(let i=0;i<150;i++){if(document.querySelector('img[alt="封面使用书名 文本封面"]'))break;await sleep(25)}
 const img=document.querySelector('img[alt="封面使用书名 文本封面"]');if(!img?.src.startsWith('blob:'))throw Error('default book cover was not fetched with the authenticated client');
 const labels=[...document.querySelectorAll('.shuihuo-project-card-kind')].map(el=>el.textContent);if(!['文本','漫画','视频'].every(mode=>labels.includes(mode)))throw Error('inconsistent mode labels: '+labels);
 if(seen.filter(url=>url==='/api/batch-factory/v12/batches/text/books/book/text-cover/display').length!==1)throw Error('wrong cover request or duplicate rendering');document.body.dataset.result='passed';}run().catch(e=>document.body.dataset.error=e.message);
 `},bundle:true,jsx:'automatic',format:'iife',write:false,minify:true,loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'}});
 const css=readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8');
 const html=join(dir,'index.html');writeFileSync(html,'<meta charset="utf-8"><style>'+css+'</style><div id="root"></div><script>'+output.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
 const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--window-size=1021,901','--virtual-time-budget=10000','--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:45000,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:8<<20});
 assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || 'card not completed');
 }finally{rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150});}
});

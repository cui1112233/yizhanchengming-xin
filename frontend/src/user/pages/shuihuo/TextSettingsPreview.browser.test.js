import test from 'node:test';
import assert from 'node:assert/strict';
import {build} from 'esbuild';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
const browser=['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
test('settings preview is explicit, uses revision guards and updates a draft without saving',{skip:!browser},async()=>{
 const dir=mkdtempSync(join(tmpdir(),'text-preview-'));
 try{
 const output=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
 import React,{useState} from 'react';import {createRoot} from 'react-dom/client';
 import Preview from './TextSettingsPreview.jsx';import {defaultTextProductionConfig} from './textProductionConfig.js';
 function App(){const [value,setValue]=useState(defaultTextProductionConfig());return <><button id="change" onClick={()=>setValue(v=>({...v,textFontSize:56}))}>字号</button><Preview batchId="b" bookId="k" value={value}/></>}
 const sleep=ms=>new Promise(r=>setTimeout(r,ms));
 async function until(fn){for(let i=0;i<120;i++){if(fn())return;await sleep(25)}throw Error('timeout')}
 async function run(){createRoot(document.querySelector('#root')).render(<App/>);await until(()=>document.querySelector('.text-settings-preview button'));await sleep(400);
 if(window.reads || window.previews?.length)throw Error('opened settings caused a preview request');
 document.querySelector('.text-settings-preview button').click();await until(()=>document.querySelector('img'));
 if(window.reads!==1 || window.previews[0].expectedRevision!==3 || window.previews[0].expectedDocumentRevision!==8 || window.previews[0].text!=='完整句子。')throw Error('wrong preview input');
 document.querySelector('#change').click();await until(()=>window.previews.some(p=>p.layout.fontSize===56));
 if(window.reads!==1)throw Error('draft refresh reloaded the book');
 document.body.dataset.result='passed';}run().catch(e=>document.body.dataset.error=e.message);
 `},plugins:[{name:'preview-fixture',setup(p){p.onResolve({filter:/shared\/api\/batchFactoryText\.js$/},args=>({path:args.path,namespace:'fixture'}));p.onLoad({filter:/.*/,namespace:'fixture'},()=>({contents:`
 export async function listTextPages(){window.reads=(window.reads||0)+1;return {document:{revision:8},pages:[{id:'p',revision:3,text:'完整句子。'}]}}
 export async function previewTextPage(b,k,p,payload){(window.previews||=[]).push(payload);return {blob:new Blob(['preview'],{type:'image/png'}),warning:''}}
 export const previewTextCover=previewTextPage;
 export async function listTextBackgrounds(){return {items:[]}}
 export async function uploadTextBackground(){throw Error('unexpected upload')}
 `}));}}],bundle:true,jsx:'automatic',format:'iife',write:false,minify:true,loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'}});
 const html=join(dir,'index.html');writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script>'+output.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
 const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=10000','--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:45000,windowsHide:true,stdio:['ignore','pipe','pipe']});
 assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || 'preview not completed');
 }finally{rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150});}
});

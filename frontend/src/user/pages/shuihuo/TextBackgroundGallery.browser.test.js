import test from 'node:test';
import assert from 'node:assert/strict';
import {build} from 'esbuild';
import {execFileSync} from 'node:child_process';
import {existsSync,mkdtempSync,readFileSync,writeFileSync,rmSync} from 'node:fs';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
const browser=['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
test('gallery owns upload and requires confirmation before applying a background',{skip:!browser},async()=>{
 const dir=mkdtempSync(join(tmpdir(),'text-gallery-'));
 try{
 const output=await build({stdin:{resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
 import {ConfigProvider} from 'antd';import React,{useState} from 'react';import {createRoot} from 'react-dom/client';import Gallery from './TextBackgroundGallery.jsx';
 let changes=0,uploads=0,release;
 window.fetch=async()=>new Response(new Blob([Uint8Array.from(atob('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII='),c=>c.charCodeAt(0))],{type:'image/png'}));
 function App(){const [value,setValue]=useState({kind:'color',color:'#212e50'}),[items,setItems]=useState([]);return <Gallery value={value} items={items} onChange={next=>{changes++;window.applied=next;setValue(next)}} onUpload={async file=>{if(file.type!=='image/png')throw Error('wrong file');uploads++;await new Promise(r=>release=r);const item={id:'uploaded',name:'自选纸张',contentUrl:'/api/batch-factory/v12/text-backgrounds/uploaded/content'};setItems([item]);return item}}/>}
 const sleep=ms=>new Promise(r=>setTimeout(r,ms));async function until(fn){for(let i=0;i<150;i++){if(fn())return;await sleep(25)}throw Error('timeout at '+document.body.dataset.step+': '+document.body.innerText)}
 const button=text=>[...document.querySelectorAll('button')].find(b=>b.textContent.replace(/\\s/g,'')===text);
 async function run(){createRoot(document.querySelector('#root')).render(<ConfigProvider theme={{token:{motion:false}}}><App/></ConfigProvider>);await until(()=>button('打开图库'));if(document.querySelector('input[type=file]'))throw Error('upload rendered outside gallery');
 button('打开图库').click();await until(()=>document.querySelector('.ant-modal input[type=file]'));
 const transfer=new DataTransfer();transfer.items.add(new File(['image'],'paper.png',{type:'image/png'}));const input=document.querySelector('.ant-modal input[type=file]');input.files=transfer.files;input.dispatchEvent(new Event('change',{bubbles:true}));
 await until(()=>uploads===1);await sleep(40);if(!button('取消').disabled)throw Error('upload can be interrupted by cancel');release();document.body.dataset.step='upload-complete';await until(()=>button('使用此背景')?.disabled===false);
 if(changes!==0)throw Error('upload applied background without confirmation');
 document.body.dataset.step='cancel-upload';button('取消').click();await until(()=>getComputedStyle(document.querySelector('.ant-modal-wrap')).display==='none');if(changes)throw Error('cancel changed background');
 button('打开图库').click();await until(()=>getComputedStyle(document.querySelector('.ant-modal-wrap')).display!=='none');await until(()=>document.querySelector('.batch-factory-background-gallery-grid button'));
 document.querySelector('.batch-factory-background-gallery-grid button').click();await until(()=>button('使用此背景'));document.body.dataset.step='apply-image';button('使用此背景').click();await until(()=>changes===1);
 if(window.applied.id!=='uploaded'||window.applied.color!=='#212e50')throw Error('chosen image or fallback color lost');
 await until(()=>getComputedStyle(document.querySelector('.ant-modal-wrap')).display==='none');button('打开图库').click();await until(()=>getComputedStyle(document.querySelector('.ant-modal-wrap')).display!=='none');
 document.body.dataset.step='apply-color';button('背景颜色').click();await until(()=>button('使用背景色'));button('使用背景色').click();await until(()=>changes===2);if(window.applied.kind!=='color'||window.applied.color!=='#212e50')throw Error('color fallback failed');
 if(uploads!==1)throw Error('upload duplicated');document.body.dataset.result='passed';}run().catch(e=>document.body.dataset.error=e.message);
 `},bundle:true,jsx:'automatic',format:'iife',write:false,minify:true,loader:{'.css':'empty'},define:{'process.env.NODE_ENV':'"production"'}});
 const css=readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8')+readFileSync(new URL('./publish-settings.css',import.meta.url),'utf8');
 const html=join(dir,'index.html');writeFileSync(html,'<meta charset="utf-8"><style>:root{--sh-panel:#202633;--sh-panel-strong:#293241;--sh-text:#eeeeee;--sh-muted:#bec8d5;--sh-line:#546078;--sh-accent:#67dac1}.ant-modal{animation:none!important;transform:none!important;transition:none!important}'+css+'</style><div id="root"></div><script>'+output.outputFiles[0].text.replaceAll('</script','<\\/script')+'</script>');
 const dom=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--user-data-dir='+join(dir,'profile'),'--virtual-time-budget=14000','--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:45000,windowsHide:true,stdio:['ignore','pipe','pipe']});
 assert.ok(dom.includes('data-result="passed"'),dom.match(/data-error="([^"]*)"/)?.[1] || 'gallery not completed');
 }finally{rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:150});}
});

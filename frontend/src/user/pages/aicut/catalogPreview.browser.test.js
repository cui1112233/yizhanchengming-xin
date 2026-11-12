import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync,mkdirSync,mkdtempSync,writeFileSync } from 'node:fs';
import { fileURLToPath,pathToFileURL } from 'node:url';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

test('actual GPU catalog effects change pixels, honor zero intensity and free resources',()=>{
 const chrome=['C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);assert.ok(chrome);
 const archive=fileURLToPath(new URL('../../../../../_local-archive-20261028/aicut-chain-repair/',import.meta.url));mkdirSync(archive,{recursive:true});const dir=mkdtempSync(path.join(archive,'catalog-gpu-'));
 const html=`<pre id="result">pending</pre><script type="module">
 import {startCatalogPreviewRenderer} from ${JSON.stringify(new URL('./catalogPreview.js',import.meta.url).href)};
 try {const source=document.createElement('canvas');source.width=32;source.height=16;const ctx=source.getContext('2d');ctx.fillStyle='rgb(0,0,0)';ctx.fillRect(0,0,32,16);ctx.fillStyle='red';ctx.fillRect(0,0,16,16);
 const samples=[];
 for(const layers of [[],[{resourceId:'effect-flash',parameters:{intensity:1}}],[{resourceId:'effect-flash',parameters:{intensity:0}}],[{resourceId:'filter-monochrome',parameters:{intensity:1}}],[{resourceId:'effect-flash',parameters:{intensity:1}},{resourceId:'filter-monochrome',parameters:{intensity:1}}]]){
  const canvas=document.createElement('canvas');const gl=canvas.getContext('webgl',{alpha:true,antialias:false,preserveDrawingBuffer:true});if(!gl)throw Error('WebGL unavailable');const textures=[];const create=gl.createTexture.bind(gl);gl.createTexture=()=>{const t=create();textures.push(t);return t;};
  const renderer=startCatalogPreviewRenderer(canvas,source,layers,130);if(!renderer)throw Error('effect shader failed');const pixels=new Uint8Array(32*16*4);gl.readPixels(0,0,32,16,gl.RGBA,gl.UNSIGNED_BYTE,pixels);const red=[...pixels.slice((8*32+8)*4,(8*32+8)*4+4)],black=[...pixels.slice((8*32+24)*4,(8*32+24)*4+4)];const error=gl.getError();renderer.stop();samples.push({red,black,error,freed:textures.every(t=>!gl.isTexture(t))});
 }
 document.getElementById('result').textContent='CATALOG_RESULT:'+JSON.stringify(samples);
 }catch(e){document.getElementById('result').textContent='CATALOG_ERROR:'+e.message;}</script>`;
 const input=path.join(dir,'preview.html');writeFileSync(input,html);const result=spawnSync(chrome,['--headless=new','--no-sandbox','--disable-gpu-sandbox','--use-angle=swiftshader','--enable-unsafe-swiftshader','--allow-file-access-from-files',`--user-data-dir=${path.join(dir,'profile')}`,'--dump-dom','--virtual-time-budget=3000',pathToFileURL(input).href],{encoding:'utf8',timeout:30000,windowsHide:true,maxBuffer:2*1024*1024});
 assert.equal(result.status,0,result.stderr);const match=result.stdout.match(/CATALOG_RESULT:(\[[^<]+\])/);assert.ok(match,result.stdout);const samples=JSON.parse(match[1]);
 assert.deepEqual(samples[0].red,[255,0,0,255]);assert.deepEqual(samples[0].black,[0,0,0,255]);
 assert.ok(samples[1].black[0]>100,'flash must affect real pixels');assert.deepEqual(samples[2].black,samples[0].black,'intensity zero must not change pixels');
 assert.ok(Math.max(...samples[3].red.slice(0,3))-Math.min(...samples[3].red.slice(0,3))<=1,'monochrome must be grayscale');
 assert.ok(samples[4].black[0]>100,'multipass stack must retain flash');assert.ok(samples.every(s=>s.error===0&&s.freed),'GL error or leaked textures');
});

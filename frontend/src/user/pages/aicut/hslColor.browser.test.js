import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

test('real WebGL HSL preview changes red preserves green and alpha then frees texture', () => {
  const chrome = ['C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
  assert.ok(chrome, 'Chrome or Edge required for GPU preview acceptance');
  const archive = fileURLToPath(new URL('../../../../../_local-archive-20261028/', import.meta.url));
  mkdirSync(archive, { recursive: true });
  const dir = mkdtempSync(path.join(archive, 'aicut-hsl-gpu-'));
  const rendererURL = new URL('../aicutVisualProcessing.js', import.meta.url).href;
  const html = `<!doctype html><html><body><pre id="result">pending</pre><script type="module">
import { startVisualProcessingRenderer } from ${JSON.stringify(rendererURL)};
try {
 const source=document.createElement('canvas');source.width=6;source.height=2;
 const ctx=source.getContext('2d');const data=ctx.createImageData(6,2);
 for(let row=0;row<2;row++) data.data.set([255,0,0,255,0,255,0,255,0,0,255,255,255,0,0,77,83,83,83,255,0,0,0,0],row*24);
 ctx.putImageData(data,0,0);const image=new Image();image.src=source.toDataURL();await image.decode();
 const canvas=document.createElement('canvas');document.body.append(canvas);
 const gl=canvas.getContext('webgl',{alpha:true,antialias:false,preserveDrawingBuffer:false});
 if(!gl) throw Error('WebGL unavailable');
 const textures=[];const createTexture=gl.createTexture.bind(gl);
 gl.createTexture=()=>{const texture=createTexture();textures.push(texture);return texture;};
 const renderer=startVisualProcessingRenderer(canvas,image,{color:{hsl:{red:{hue:100},blue:{hue:100}}}});
 if(!renderer) throw Error('WebGL renderer failed');
 const pixels=new Uint8Array(6*2*4);
 gl.readPixels(0,0,6,2,gl.RGBA,gl.UNSIGNED_BYTE,pixels);
 renderer.stop();
 const texturesFreed=textures.length===2&&textures.every(texture=>!gl.isTexture(texture));
 const gradeSource=document.createElement('canvas');gradeSource.width=2;gradeSource.height=2;
 const gradeCtx=gradeSource.getContext('2d');const gradeImageData=gradeCtx.createImageData(2,2);
 for(let offset=0;offset<16;offset+=4)gradeImageData.data.set([128,128,128,77],offset);
 gradeCtx.putImageData(gradeImageData,0,0);const gradeImage=new Image();gradeImage.src=gradeSource.toDataURL();await gradeImage.decode();
 const grades=[];
 for(const color of [{rgbCurves:{red:[{x:0,y:0},{x:.5,y:1},{x:1,y:1}]}},{wheels:{shadows:{hue:0,saturation:100}}}]){
  const current=startVisualProcessingRenderer(canvas,gradeImage,{color});if(!current)throw Error('grade renderer failed');
  const sample=new Uint8Array(16);gl.readPixels(0,0,2,2,gl.RGBA,gl.UNSIGNED_BYTE,sample);grades.push([...sample.slice(0,4)]);current.stop();
 }
 document.getElementById('result').textContent='GPU_RESULT:'+JSON.stringify({pixels:[...pixels],grades,error:gl.getError(),texturesFreed});
} catch(error) { document.getElementById('result').textContent='GPU_ERROR:'+error.message; }
</script></body></html>`;
  const input = path.join(dir, 'preview.html');
  writeFileSync(input, html);
  const result = spawnSync(chrome, ['--headless=new', '--no-sandbox', '--disable-gpu-sandbox', '--use-angle=swiftshader', '--enable-unsafe-swiftshader', '--allow-file-access-from-files', `--user-data-dir=${path.join(dir, 'profile')}`, '--dump-dom', '--virtual-time-budget=3000', pathToFileURL(input).href], { encoding: 'utf8', timeout: 30000, windowsHide: true, maxBuffer: 2 * 1024 * 1024 });
  assert.equal(result.status, 0, result.stderr || result.error?.message);
  const match = result.stdout.match(/GPU_RESULT:(\{[^<]+\})/);
  assert.ok(match, result.stdout);
  const { pixels, grades, error, texturesFreed } = JSON.parse(match[1]);
  assert.equal(error, 0, 'WebGL reported an error');
  assert.equal(texturesFreed, true, 'media and LUT textures must both be deleted');
  for (const [index, expected] of [[0, [255, 128, 128, 77]], [1, [140, 125, 125, 77]]]) {
    assert.ok(expected.every((channel, i) => Math.abs(channel - grades[index][i]) <= 3), `GPU RGB/wheel grade ${index}: ${grades[index]}`);
  }
  for (let offset = 0; offset < pixels.length; offset += 24) {
    assert.ok(pixels[offset] >= 254 && Math.abs(pixels[offset + 1] - 128) <= 2 && pixels[offset + 2] <= 1, 'red must become orange');
    assert.deepEqual(pixels.slice(offset + 4, offset + 8), [0, 255, 0, 255]);
    assert.ok(Math.abs(pixels[offset + 8] - 128) <= 2 && pixels[offset + 9] <= 1 && pixels[offset + 10] >= 254, 'blue slice must become purple');
    assert.ok(Math.abs(pixels[offset + 15] - 77) <= 1, 'partial alpha preserved');
    assert.ok(Math.max(...pixels.slice(offset + 16, offset + 19)) - Math.min(...pixels.slice(offset + 16, offset + 19)) <= 1, 'gray remains neutral');
    assert.equal(pixels[offset + 23], 0);
  }
});

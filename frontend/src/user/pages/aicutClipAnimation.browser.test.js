import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import { fileURLToPath, pathToFileURL } from 'node:url';
import path from 'node:path';
import { spawnSync } from 'node:child_process';

test('real browser renders independent entrance exit and looping output-time samples', () => {
  const chrome = ['C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
  assert.ok(chrome, 'Chrome or Edge required');
  const archive = fileURLToPath(new URL('../../../../_local-archive-20261028/', import.meta.url));
  mkdirSync(archive, { recursive: true });
  const dir = mkdtempSync(path.join(archive, 'aicut-animation-browser-'));
  const html = `<!doctype html><html><body><pre id="result">pending</pre><script type="module">
import { clipAnimationPreviewStyle } from ${JSON.stringify(new URL('./aicutClipAnimation.js', import.meta.url).href)};
const samples=[];
for (const [phase,preset,time] of [['entrance','slide',250],['entrance','zoom',250],['exit','fade',1750],['loop','float',500],['loop','pulse',1000]]) {
 const stage=document.createElement('div');Object.assign(stage.style,{position:'absolute',left:'0px',top:'0px',width:'200px',height:'100px',overflow:'hidden'});
 const layer=document.createElement('div');Object.assign(layer.style,{position:'absolute',inset:'0',background:'red'},clipAnimationPreviewStyle({[phase]:{preset,durationMs:phase==='loop'?2000:500}},2000,time));stage.append(layer);document.body.append(stage);
 const rect=layer.getBoundingClientRect();samples.push({phase,preset,left:rect.left,top:rect.top,width:rect.width,opacity:Number(getComputedStyle(layer).opacity)});stage.remove();
}
document.getElementById('result').textContent='ANIMATION_RESULT:'+JSON.stringify(samples);
</script></body></html>`;
  const input = path.join(dir, 'preview.html');
  writeFileSync(input, html);
  const result = spawnSync(chrome, ['--headless=new', '--no-sandbox', '--allow-file-access-from-files', `--user-data-dir=${path.join(dir, 'profile')}`, '--dump-dom', '--virtual-time-budget=1500', pathToFileURL(input).href], { encoding: 'utf8', timeout: 30000, windowsHide: true, maxBuffer: 2 * 1024 * 1024 });
  assert.equal(result.status, 0, result.stderr || result.error?.message);
  const match = result.stdout.match(/ANIMATION_RESULT:(\[[^<]+\])/);
  assert.ok(match, result.stdout);
  const samples = JSON.parse(match[1]);
  assert.equal(samples[0].left, -100);
  assert.ok(Math.abs(samples[1].width - 180) < .01);
  assert.equal(samples[2].opacity, .5);
  assert.ok(Math.abs(samples[3].top - 1.5) < .01);
  assert.ok(Math.abs(samples[4].width - 210) < .01);
});

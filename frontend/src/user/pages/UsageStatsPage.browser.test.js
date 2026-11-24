import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdtempSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';

const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(p => p && existsSync(p));

test('billing keeps pending usage unknown, member calls private, and server failures retryable', { skip: !browser }, async () => {
  const bundle = await build({ stdin: { resolveDir: here, loader: 'jsx', contents: `
    import React from 'react'; import {createRoot} from 'react-dom/client'; import Page from './UsageStatsPage.jsx';
    let role='member',fail=false,settled=false;const requests=[];
    const summary=()=>({calls:1,billableCalls:1,pricedCalls:settled?1:0,knownUsageCalls:settled?1:0,pendingCalls:settled?0:1,inputTokens:settled?12:0,outputTokens:settled?8:0,totalTokens:settled?20:0,actualCost:settled?.25:null,estimatedCost:settled?.25:null,costCurrency:'CNY',callsByFeature:{image:1},byFeature:{image:settled?20:0}});
    window.fetch=async(url,options={})=>{requests.push(String(url));let data={},status=200;
      if(String(url)==='/api/member/me'){if(fail){data={error:'账单读取失败，请稍后重试'};status=503;}else data={member:{username:'me',role},usage:{historyCoverage:'tracked_requests',day:summary(),month:summary(),recent:[{id:'request1',at:'2026-11-24T03:00:00Z',feature:'image',usageKnown:settled,inputTokens:settled?12:0,outputTokens:settled?8:0,totalTokens:settled?20:0,costKnown:settled,estimatedCost:settled?.25:null,costCurrency:settled?'CNY':'',source:'yzcm_billing_log'}]}};}
      else if(String(url)==='/api/member/team') data={members:[{username:'alice',role:'member',usage:{month:summary()}}]};
      return new Response(JSON.stringify(data),{status,headers:{'Content-Type':'application/json'}});
    };
    const tick=()=>new Promise(r=>setTimeout(r,180)),check=(v,m)=>{if(!v)throw Error(m)},root=createRoot(document.getElementById('root'));
    const button=label=>[...document.querySelectorAll('button')].find(b=>b.textContent.replace(/\\s/g,'').includes(label));
    (async()=>{
      root.render(<Page key="member"/>);await tick();await tick();
      check(!requests.includes('/api/member/team'),'ordinary member requested team');
      check(!document.querySelector('input'),'manual credential required');
      check([...document.querySelectorAll('.ac-metric-value strong')].every(e=>e.textContent==='—'),'pending token metrics masquerade as zero');
      check(document.querySelector('.ac-usage-table-row b').textContent==='—','pending row token count masquerades as zero');
      check(document.querySelectorAll('.ac-bar-column').length===0,'pending receipt fabricated chart bar');
      check(document.body.textContent.includes('待核账'),'pending state missing');
      check(!document.body.textContent.includes('次制作')&&!document.body.textContent.includes('制作调用'),'model calls labelled as works');
      role='manager';root.render(<Page key="manager"/>);await tick();await tick();
      check(document.querySelector('.ac-ranking-value').textContent.includes('—'),'pending member fee masquerades as zero');
      check(!document.querySelector('.ac-ranking-value').textContent.includes('0.0000'),'pending ranking shown as free');
      check(!document.querySelector('.ac-side-summary').textContent.includes('0.0000'),'pending team total shown as free');
      settled=true;button('刷新账单').click();await tick();await tick();
      check(document.querySelectorAll('.ac-metric-value strong')[1].textContent==='20','real known tokens hidden');
      check(document.querySelectorAll('.ac-metric-value strong')[4].textContent==='0.2500','actual cost lost');
      check(document.querySelector('.ac-side-summary').textContent.includes('CNY 0.2500')&&!document.querySelector('.ac-side-summary').textContent.includes('估算'),'actual team charges called estimate');
      fail=true;button('刷新账单').click();await tick();await tick();
      check(document.body.textContent.includes('账单读取失败，请稍后重试'),'safe server failure hidden');
      check(button('重试'),'failed page has no retry');
      fail=false;button('重试').click();await tick();await tick();check(document.querySelectorAll('.ac-metric-value strong')[4].textContent==='0.2500','retry did not recover');
      document.body.dataset.result='passed';
    })().catch(e=>{document.body.dataset.result='failed';document.body.dataset.error=e.message;});
  ` }, bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser', loader: { '.css': 'empty' } });
  const dir = mkdtempSync(path.join(tmpdir(), 'yzcm-candidate-ui-'));
  writeFileSync(path.join(dir, 'test.js'), bundle.outputFiles[0].text);
  const html = path.join(dir, 'test.html');
  writeFileSync(html, '<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output = execFileSync(browser, ['--headless=new', '--disable-gpu', '--no-first-run', '--disable-background-networking', '--host-resolver-rules=MAP * ~NOTFOUND', '--virtual-time-budget=12000', '--user-data-dir=' + path.join(dir, 'profile'), '--dump-dom', pathToFileURL(html).href], { encoding: 'utf8', timeout: 60000, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'], maxBuffer: 4000000 });
  writeFileSync(path.join(dir, 'result.html'), output);
  assert.ok(/data-result="passed"/.test(output), output.match(/data-error="([^"]*)"/)?.[1] || 'browser did not finish: ' + dir);
});

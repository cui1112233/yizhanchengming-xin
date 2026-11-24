import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';
import { build } from 'esbuild';

const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(path => path && existsSync(path));

let renderedFixture;
async function renderFixture() {
  if (renderedFixture) return renderedFixture;
  renderedFixture = (async () => {
    const result = await build({
      stdin: {
        resolveDir: fileURLToPath(new URL('.', import.meta.url)),
        loader: 'jsx',
        contents: `
          import React from 'react';
          import {createRoot} from 'react-dom/client';
          import InlineTextBookPages from './InlineTextBookPages.jsx';
          window.addEventListener('error', event => { document.body.dataset.error = event.error?.message || event.message; });
          window.addEventListener('unhandledrejection', event => { document.body.dataset.error = event.reason?.message || String(event.reason); });
          localStorage.setItem('auth_token', 'fixture');
          window.fetch = async path => {
            if (path === '/api/client-errors') return new Response('{}', {headers:{'Content-Type':'application/json'}});
            return new Response(JSON.stringify({
              pages:[{id:'page-1', ordinal:1, text:'测试正文', activeArtifactId:''}],
              artifacts:[]
            }), {headers:{'Content-Type':'application/json'}});
          };
          const root = createRoot(document.querySelector('#root'));
          try { root.render(<article className="shuihuo-workbench-row batch-factory-book-row"><InlineTextBookPages batchId="batch-1" book={{id:'book-1'}} /></article>); }
          catch (error) { document.body.dataset.error = error.message; }
          const wait = async condition => { for (let i=0;i<80;i++){ if(condition()) return; await new Promise(resolve=>setTimeout(resolve,20)); } throw Error('等待文字页超时'); };
          wait(() => document.querySelectorAll('.batch-factory-text-page-head').length === 2
            && document.querySelectorAll('.batch-factory-cell-pager').length === 2
            && document.querySelector('.batch-factory-text-page-content')?.textContent === '测试正文').then(() => {
            const headers = [...document.querySelectorAll('.batch-factory-text-page-head')];
            const pagers = [...document.querySelectorAll('.batch-factory-cell-pager')];
            document.body.dataset.duplicateTitles = String(headers.some(header => /页面内容|文字页/.test(header.textContent)));
            document.body.dataset.pagerDelta = String(Math.abs(pagers[0].getBoundingClientRect().bottom - pagers[1].getBoundingClientRect().bottom));
            document.body.dataset.pageInfo = headers.map(header => header.textContent).join('|');
          }).catch(error => { if (!document.body.dataset.error) document.body.dataset.error = error.message; });
        `,
      },
      bundle: true,
      write: false,
      format: 'iife',
      jsx: 'automatic',
      minify: true,
      define: {'process.env.NODE_ENV': '"production"'},
    });
    const dir = mkdtempSync(join(tmpdir(), 'inline-text-layout-'));
    try {
      writeFileSync(join(dir, 'styles.css'), readFileSync(fileURLToPath(new URL('../shuihuo-production.css', import.meta.url))));
      const html = join(dir, 'index.html');
      const script = result.outputFiles[0].text.replaceAll('</script', '<\\/script');
      writeFileSync(html, `<meta charset="utf-8"><link rel="stylesheet" href="styles.css"><style>#root .batch-factory-book-row{display:grid;grid-template-columns:330px 300px!important;width:630px}</style><div id="root"></div><script>${script}</script>`);
      return execFileSync(browser, ['--headless=new', '--disable-gpu', '--disable-background-networking', '--no-first-run',
        '--host-resolver-rules=MAP * ~NOTFOUND', '--user-data-dir=' + join(dir, 'profile'), '--virtual-time-budget=8000',
        '--dump-dom', pathToFileURL(html).href], {encoding:'utf8', timeout:45000, windowsHide:true, stdio:['ignore','pipe','pipe']});
    } finally {
      rmSync(dir, {recursive:true, force:true, maxRetries:4, retryDelay:200});
    }
  })();
  return renderedFixture;
}

test('inline text page headers keep page information without repeating column titles', {skip:!browser}, async () => {
  const dom = await renderFixture();
  assert.match(dom, /data-duplicate-titles="false"/);
  assert.match(dom, /data-page-info="[^"]*第 1 页/);
});

test('assigned copy and rendered text page pagers share the same bottom edge', {skip:!browser}, async () => {
  const dom = await renderFixture();
  assert.match(dom, /data-pager-delta="0"/);
});

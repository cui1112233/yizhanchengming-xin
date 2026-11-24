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
            if (String(path).includes('/text-artifacts/artifact-1/content')) return new Response(new Uint8Array([
              137,80,78,71,13,10,26,10,0,0,0,13,73,72,68,82,0,0,0,3,0,0,0,4,8,2,0,0,0,84,178,83,118,
              0,0,0,12,73,68,65,84,8,215,99,248,207,192,0,0,4,1,1,0,233,89,119,218,0,0,0,0,73,69,78,68,174,66,96,130
            ]), {headers:{'Content-Type':'image/png'}});
            return new Response(JSON.stringify({
              pages:[{id:'page-1', ordinal:1, text:'测试正文', activeArtifactId:'artifact-1'}],
              artifacts:[{id:'artifact-1',kind:'page',contentUrl:'/api/batch-factory/v12/batches/batch-1/books/book-1/text-artifacts/artifact-1/content'}]
            }), {headers:{'Content-Type':'application/json'}});
          };
          const root = createRoot(document.querySelector('#root'));
          try { root.render(<div className="batch-factory-workbench-table"><div className="shuihuo-workbench-head" style={{gridTemplateColumns:'330px 300px',minWidth:630}}><div>页面内容</div><div>文字页</div></div><article className="shuihuo-workbench-row batch-factory-book-row" style={{gridTemplateColumns:'330px 300px',minWidth:630}}><InlineTextBookPages batchId="batch-1" book={{id:'book-1'}} /></article></div>); }
          catch (error) { document.body.dataset.error = error.message; }
          const wait = async condition => { for (let i=0;i<80;i++){ if(condition()) return; await new Promise(resolve=>setTimeout(resolve,20)); } throw Error('等待文字页超时'); };
          wait(() => document.querySelectorAll('.batch-factory-text-page-head').length === 2
            && document.querySelectorAll('.batch-factory-cell-pager').length === 2
            && document.querySelector('.batch-factory-text-page-image img')?.complete
            && document.querySelector('.batch-factory-text-page-content')?.textContent === '测试正文').then(() => {
            const headers = [...document.querySelectorAll('.batch-factory-text-page-head')];
            const pagers = [...document.querySelectorAll('.batch-factory-cell-pager')];
            const head = document.querySelector('.shuihuo-workbench-head').getBoundingClientRect();
            const labels = [...document.querySelectorAll('.shuihuo-workbench-head > div')];
            const preview = document.querySelector('.batch-factory-text-page-image').getBoundingClientRect();
            const image = document.querySelector('.batch-factory-text-page-image img').getBoundingClientRect();
            const cell = document.querySelector('.batch-factory-text-rendered-cell').getBoundingClientRect();
            document.body.dataset.previewShellWidthDelta = String(Math.abs(preview.width - cell.width));
            document.body.dataset.duplicateTitles = String(headers.some(header => /页面内容|文字页/.test(header.textContent)));
            document.body.dataset.pagerDelta = String(Math.abs(pagers[0].getBoundingClientRect().bottom - pagers[1].getBoundingClientRect().bottom));
            document.body.dataset.pageInfo = headers.map(header => header.textContent).join('|');
            document.body.dataset.headerHeight = String(head.height);
            document.body.dataset.headerLabelsAligned = String(labels.every(label => Math.abs(label.getBoundingClientRect().height - head.height) <= 1));
            document.body.dataset.previewCenterDelta = String(Math.abs((image.left + image.right) / 2 - (preview.left + preview.right) / 2));
            document.body.dataset.previewWidth = String(image.width);
            document.body.dataset.previewContained = String(image.right <= preview.right && image.bottom <= preview.bottom);
            document.body.dataset.previewRatio = String(image.width / image.height);
          }).catch(error => { if (!document.body.dataset.error) document.body.dataset.error = error.message; });
        `,
      },
      bundle: true,
      plugins:[{name:'fixture-text-page-image',setup(plugin){
        plugin.onResolve({filter:/AuthenticatedTextPageImage\.jsx$/},()=>({path:'fixture-text-page-image',namespace:'fixture'}));
        plugin.onLoad({filter:/.*/,namespace:'fixture'},()=>({loader:'jsx',resolveDir:fileURLToPath(new URL('.',import.meta.url)),contents:`
          export default function FixtureTextPageImage({alt}) {
            return <img alt={alt} src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='300' height='400'%3E%3Crect width='300' height='400' fill='%23111827'/%3E%3C/svg%3E"/>;
          }
        `}));
      }}],
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

test('shared workbench header is compact and its labels fill the row height', {skip:!browser}, async () => {
  const dom = await renderFixture();
  const height = Number(dom.match(/data-header-height="([^"]+)"/)?.[1]);
  assert.ok(height >= 30 && height <= 34, `header height ${height}`);
  assert.match(dom, /data-header-labels-aligned="true"/);
});

test('rendered text page preview is centered, enlarged, contained, and keeps 3:4', {skip:!browser}, async () => {
  const dom = await renderFixture();
  const centerDelta = Number(dom.match(/data-preview-center-delta="([^"]+)"/)?.[1]);
  const width = Number(dom.match(/data-preview-width="([^"]+)"/)?.[1]);
  const ratio = Number(dom.match(/data-preview-ratio="([^"]+)"/)?.[1]);
  assert.ok(centerDelta <= 1, `preview center delta ${centerDelta}`);
  const shellDelta = Number(dom.match(/data-preview-shell-width-delta="([^"]+)"/)?.[1]);
  assert.ok(shellDelta <= 2, `preview shell must fill its column, width delta ${shellDelta}`);
  assert.ok(width > 120, `preview width ${width}`);
  assert.match(dom, /data-preview-contained="true"/);
  assert.ok(Math.abs(ratio - 0.75) < 0.02, `preview ratio ${ratio}`);
});

import test from 'node:test';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { build } from 'esbuild';
import { runNovelBrowserDOM } from './novel-creation/browserEvidence.js';

const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(p => p && existsSync(p));
// Read the real application stylesheet imports, in application order. This
// deliberately has no fixture :root variables that could hide a theme collision.
const appRoot = process.env.QIANTIE_TEST_NOVEL_THEME_ROOT || path.resolve(here, '..');
const main = readFileSync(path.join(appRoot, 'main.jsx'), 'utf8');
const stylePaths = [...main.matchAll(/import\s+'(\.\.\/shared\/styles\/[^']+\.css)'/g)].map(m => path.resolve(appRoot, m[1]));
const styles = stylePaths.map(p => readFileSync(p, 'utf8')).join('\n') + '\n' + readFileSync(path.join(here, 'NovelCreationPage.css'), 'utf8');

for (const theme of ['dark', 'light']) test(`全站真实glass ${theme}小说Modal和Drawer内容及头部完全不透明`, { skip: !browser }, async () => {
  const code = `import React from 'react';import{createRoot}from'react-dom/client';import{Modal,Drawer,ConfigProvider,theme as antdTheme}from'antd';
  document.documentElement.dataset.scheme='glass';document.documentElement.dataset.theme=${JSON.stringify(theme)};
  window.calls=[];const wait=async()=>{for(let n=0;n<100;n++){if(document.querySelector('.ant-modal-content')&&document.querySelector('.ant-drawer-content'))return;await new Promise(r=>setTimeout(r,50))}throw Error('portals missing')};
  createRoot(document.getElementById('root')).render(<ConfigProvider theme={{algorithm:${theme === 'dark' ? 'antdTheme.darkAlgorithm' : 'antdTheme.defaultAlgorithm'}}}><main className="legacy-shell"><div className="legacy-main"><div className="novel-creation-page"><p>底层创作正文必须被弹窗遮住</p></div></div></main><Modal open title="新建小说作品" footer={null}><p>小说弹窗正文</p></Modal><Drawer open title="导入素材"><p>小说资料正文</p></Drawer></ConfigProvider>);
  (async()=>{await wait();if(!CSS.supports('backdrop-filter','blur(1px)'))throw Error('glass support required');const result=[];for(const selector of ['.ant-modal-content','.ant-modal-header','.ant-drawer-content','.ant-drawer-header']){const el=document.querySelector(selector),s=getComputedStyle(el),color=s.backgroundColor;const alpha=color.startsWith('rgba')?Number(color.split(',').at(-1).replace(')','')):color==='transparent'?0:1;result.push({selector,color,alpha,surface:s.getPropertyValue('--surface-base').trim(),blur:s.backdropFilter});if(alpha!==1)throw Error(${JSON.stringify(theme)}+' '+selector+' background alpha='+alpha+' '+color);if(s.backdropFilter!=='none')throw Error('novel dialog still blurred');}document.body.dataset.computed=JSON.stringify(result);document.body.dataset.result='passed'})().catch(e=>{document.body.dataset.error=e.message;document.body.dataset.result='failed'});`;
  const bundle = await build({ stdin: { resolveDir: here, loader: 'jsx', contents: code }, bundle: true, write: false, minify: true, jsx: 'automatic', define: { 'process.env.NODE_ENV': '"production"' }, loader: { '.css': 'empty' } });
  const base = path.resolve(here, '../../../../_local-archive-20261028/novel-task8-theme');
  mkdirSync(base, { recursive: true }); const dir = mkdtempSync(path.join(base, theme + '-'));
  writeFileSync(path.join(dir, 'test.js'), bundle.outputFiles[0].text);
  const html = path.join(dir, 'test.html'); writeFileSync(html, '<meta charset="utf-8"><style>' + styles + '</style><div id="root"></div><script src="test.js"></script>');
  await runNovelBrowserDOM(browser, html, dir);
});

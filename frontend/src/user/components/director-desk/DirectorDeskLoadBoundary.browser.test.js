import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const browser = [
  process.env.CHROME_PATH,
  'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
  'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
  '/usr/bin/google-chrome', '/usr/bin/chromium', '/usr/bin/chromium-browser',
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
].find(candidate => candidate && fs.existsSync(candidate));

test('加载出错后显示提示，点击重新加载调用重试回调', { skip: browser ? false : 'Chrome or Edge is not available' }, async () => {
  const bundle = await build({
    stdin: {
      resolveDir: dirname,
      loader: 'jsx',
      contents: `
        import React from 'react';
        import { createRoot } from 'react-dom/client';
        import { flushSync } from 'react-dom';
        import { DirectorDeskLoadBoundary } from './DirectorDeskLoadBoundary.jsx';
        function Broken() { throw new Error('simulated module load failure'); }
        let retries = 0;
        const root = createRoot(document.getElementById('root'));
        flushSync(() => root.render(<DirectorDeskLoadBoundary onRetry={() => retries++}><Broken /></DirectorDeskLoadBoundary>));
        const alert = document.querySelector('[role="alert"]');
        const button = alert?.querySelector('button');
        const valid = alert?.textContent.includes('导演台资源加载失败') && button?.textContent === '重新加载页面';
        button?.click();
        document.body.dataset.testResult = valid && retries === 1 ? 'passed' : 'failed';
      `
    },
    bundle: true,
    jsx: 'automatic',
    write: false,
    format: 'iife',
    platform: 'browser'
  });
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'director-boundary-test-'));
  try {
    const html = path.join(tempDir, 'test.html');
    fs.writeFileSync(path.join(tempDir, 'test.js'), bundle.outputFiles[0].text);
    fs.writeFileSync(html, '<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
    const output = execFileSync(browser, [
      '--headless', '--disable-gpu', '--no-first-run', '--no-default-browser-check',
      '--disable-background-networking', '--host-resolver-rules=MAP * ~NOTFOUND',
      `--user-data-dir=${path.join(tempDir, 'profile')}`, '--dump-dom', pathToFileURL(html).href
    ], { encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe'] });
    assert.match(output, /data-test-result="passed"/);
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
});

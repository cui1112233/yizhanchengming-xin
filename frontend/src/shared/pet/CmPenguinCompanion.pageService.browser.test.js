import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { build } from 'esbuild';

const here = path.dirname(fileURLToPath(import.meta.url));
const browser = [process.env.CHROME_PATH, 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe', 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe'].find(file => file && fs.existsSync(file));

test('page panel opens from retained request, stays mounted closed, updates in scope and resets on page change', { skip: !browser }, async () => {
  const bundle = await build({
    stdin: { resolveDir: here, loader: 'jsx', contents: `
      import React, { useEffect, useState } from 'react';
      import { createRoot } from 'react-dom/client';
      import { CmPenguinCompanion } from './CmPenguinCompanion.jsx';
      import { registerPetPageService, openPetPageService } from './pageService.js';
      window.fetch = async () => ({ok: true, json: async () => ({pet: 'cm-penguin'})});
      let mounts = 0, unmounts = 0;
      function Panel({revision}) {
        const [count, setCount] = useState(0);
        useEffect(() => { mounts++; return () => { unmounts++; }; }, []);
        return <button id="real-panel" onClick={() => setCount(value => value + 1)}>{revision}:{count}</button>;
      }
      const assert = (value, message) => { if (!value) throw new Error(message); };
      const tick = () => new Promise(resolve => setTimeout(resolve, 40));
      (async () => {
        let unregister = registerPetPageService({id: 'aicut', scope: 'alice/project-1', label: '剪辑服务', render: () => <Panel revision="one" />});
        openPetPageService('aicut');
        createRoot(document.getElementById('root')).render(<CmPenguinCompanion username="alice" accountSessionKey="session" />);
        await tick();
        assert(document.querySelector('#real-panel'), 'real registered panel missing');
        assert(document.querySelector('[role="dialog"]')?.textContent.includes('剪辑服务'), 'service label missing');
        assert(!document.querySelector('[aria-label="分析当前内容"]'), 'generic fake action visible');
        assert(!document.querySelector('.cm-penguin-composer'), 'generic composer visible');
        document.querySelector('#real-panel').click(); await tick();
        document.querySelector('.cm-penguin-header-actions button[title="关闭"]').click(); await tick();
        assert(document.querySelector('#real-panel')?.textContent === 'one:1', 'close unmounted service');
        assert(document.querySelector('#real-panel').closest('[hidden]'), 'closed panel remains visible');
        const updatedCleanup = registerPetPageService({id: 'aicut', scope: 'alice/project-1', label: '剪辑服务', render: () => <Panel revision="two" />});
        unregister(); unregister = updatedCleanup; await tick();
        assert(document.querySelector('#real-panel')?.textContent === 'two:1' && mounts === 1, 'revision reset state');
        openPetPageService('aicut'); await tick();
        assert(!document.querySelector('#real-panel').closest('[hidden]'), 'explicit reopen failed');
        const nextCleanup = registerPetPageService({id: 'aicut', scope: 'alice/project-2', label: '另一个工程', render: () => <Panel revision="three" />});
        unregister(); unregister = nextCleanup; await tick();
        assert(document.querySelector('#real-panel')?.textContent === 'three:0' && mounts === 2 && unmounts === 1, 'scope did not isolate state');
        unregister(); await tick();
        assert(!document.querySelector('#real-panel') && unmounts === 2, 'page exit did not clean panel');
        (document.querySelector('.cm-penguin-chat-trigger') || document.querySelector('[role="button"]')).click(); await tick();
        assert(document.querySelector('.cm-penguin-composer'), 'generic bridge did not return');
        document.body.dataset.testResult = 'passed';
      })().catch(error => { document.body.dataset.testResult = 'failed'; document.body.dataset.error = error.message; });
    ` },
    bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser', loader: { '.css': 'empty' }
  });
  const archive = path.resolve(here, '../../../../_local-archive-20261028/pet-page-service-tests');
  fs.mkdirSync(archive, { recursive: true });
  const directory = fs.mkdtempSync(path.join(archive, 'run-'));
  const html = path.join(directory, 'test.html');
  fs.writeFileSync(path.join(directory, 'test.js'), bundle.outputFiles[0].text);
  fs.writeFileSync(html, '<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output = execFileSync(browser, ['--headless', '--disable-gpu', '--no-first-run', '--no-default-browser-check', '--disable-background-networking', '--host-resolver-rules=MAP * ~NOTFOUND', '--virtual-time-budget=3000', '--allow-file-access-from-files', '--user-data-dir=' + path.join(directory, 'profile'), '--dump-dom', pathToFileURL(html).href], { encoding: 'utf8', timeout: 30000, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] });
  assert.match(output, /data-test-result="passed"/, output.match(/data-error="([^"]*)"/)?.[1] || 'browser did not complete');
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';

test('hidden pet only reappears for a new explicit service open, never ordinary context updates', async () => {
  const browser = ['C:/Program Files/Google/Chrome/Application/chrome.exe', 'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(existsSync);
  assert.ok(browser);
  const here = path.dirname(fileURLToPath(import.meta.url));
  const bundle = await build({
    stdin: { resolveDir: here, loader: 'jsx', contents: `
      import React from 'react'; import {createRoot} from 'react-dom/client';
      import Host from './AICutPetHost.jsx';
      import {registerPetPageService,openPetPageService} from '../../shared/pet/pageService.js';
      const root=createRoot(document.getElementById('root'));
      const tick=()=>new Promise(resolve=>setTimeout(resolve,40));
      const assert=(value,message)=>{if(!value)throw Error(message)};
      const service={id:'aicut-agent',scope:'alice:cut:default',label:'剪辑',render:()=>null};
      (async()=>{
        registerPetPageService(service); root.render(<Host scope={service.scope}/>); await tick();await tick();
        assert(!document.getElementById('pet'),'hidden preference ignored');
        openPetPageService(service.id);await tick(); assert(document.getElementById('pet'),'explicit open failed');
        window.dispatchEvent(new CustomEvent('qiantie:notifications-updated',{detail:{petVisible:false}}));await tick();
        assert(!document.getElementById('pet'),'hide notification ignored');
        registerPetPageService({...service,label:'剪辑·新播放头'});await tick();
        assert(!document.getElementById('pet'),'context update replayed an old open');
        openPetPageService(service.id);await tick(); assert(document.getElementById('pet'),'new explicit open ignored');
        const next={...service,scope:'alice:other:default'};registerPetPageService(next);root.render(<Host scope={next.scope}/>);await tick();await tick();
        assert(!document.getElementById('pet'),'scope change retained visibility override');
        document.body.dataset.result='passed';
      })().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
    ` },
    plugins: [{ name: 'host-services', setup(plugin) {
      plugin.onResolve({ filter: /shared\/api\/(config|auth)\.js$/ }, args => ({ path: args.path.includes('config') ? 'config' : 'auth', namespace: 'fixture' }));
      plugin.onResolve({ filter: /shared\/pet\/CmPenguinCompanion\.jsx$/ }, () => ({ path: 'pet', namespace: 'fixture' }));
      plugin.onLoad({ filter: /.*/, namespace: 'fixture' }, args => ({ loader: 'js', resolveDir: here, contents: args.path === 'config' ? 'export async function getConfig(){return {notifications:{petVisible:false}}}' : args.path === 'auth' ? 'export function getCurrentUsername(){return "alice"}' : 'import React from "react"; export function CmPenguinCompanion(){return React.createElement("div",{id:"pet"})}' }));
    } }],
    bundle: true, jsx: 'automatic', write: false, format: 'iife', platform: 'browser'
  });
  const archive = path.resolve(here, '../../../../_local-archive-20261028/aicut-pet-parity');
  mkdirSync(archive, { recursive: true });
  const dir = mkdtempSync(path.join(archive, 'host-'));
  writeFileSync(path.join(dir, 'test.js'), bundle.outputFiles[0].text);
  const html = path.join(dir, 'test.html'); writeFileSync(html, '<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--virtual-time-budget=4000',`--user-data-dir=${path.join(dir,'profile')}`,'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,windowsHide:true,stdio:['ignore','pipe','pipe']});
  assert.ok(output.includes('data-result="passed"'), output.match(/data-error="([^"]*)"/)?.[1] || 'browser did not finish');
});

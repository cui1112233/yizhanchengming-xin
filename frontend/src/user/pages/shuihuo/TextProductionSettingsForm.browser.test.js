import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';
import { build } from 'esbuild';

const browser = [process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(path => path && existsSync(path));

test('文本统一配置可同步选档，关闭再打开显示保存的名称与生产失败说明', {skip:!browser}, async () => {
  const dir = mkdtempSync(join(tmpdir(),'text-settings-browser-'));
  try {
    const bundle = await build({
      stdin: {resolveDir:fileURLToPath(new URL('.',import.meta.url)),loader:'jsx',contents:`
        import React,{useState} from 'react';
        import {createRoot} from 'react-dom/client';
        import Form from './TextProductionSettingsForm.jsx';
        import {applyTextProductionChanges,normalizeTextProductionConfig} from './textProductionConfig.js';
        function Harness() {
          const [draft, setDraft] = useState({sensitiveMode:'ai'});
          const [open,setOpen] = useState(true);
          return <><button id="toggle" onClick={()=>{
            if(open) window.saved=JSON.stringify(draft);
            else setDraft(JSON.parse(window.saved));
            setOpen(!open);
          }}>切换</button>{open && <Form value={normalizeTextProductionConfig(draft)}
            onChange={changes=>setDraft(current=>applyTextProductionChanges(current,changes))}/>}</>;
        }
        async function waitFor(fn, message) {
          for(let i=0;i<50;i++){const value=fn();if(value)return value;await new Promise(resolve=>setTimeout(resolve,35));}
          throw Error(message);
        }
        const byText=(selector,text)=>[...document.querySelectorAll(selector)].find(el=>el.textContent.trim()===text);
        async function run() {
          createRoot(document.querySelector('#root')).render(<Harness/>);
          await waitFor(()=>document.body.textContent.includes('停止生产并显示失败原因'),'production failure note');
          const publish=()=>waitFor(()=>byText('[role=tab]','发布配置'),'publish tab');
          (await publish()).click();
          (await waitFor(()=>byText('button','同步网站配置档'),'sync button')).click();
          await waitFor(()=>window.synced,'sync request');
          const combo=await waitFor(()=>document.querySelector('[aria-label="网站配置档"]'),'profile field');
          combo.closest('.ant-select').querySelector('.ant-select-selector').dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));
          (await waitFor(()=>byText('.ant-select-item-option-content','8个解压'),'profile option')).click();
          await waitFor(()=>document.querySelector('.ant-select-selection-item[title="8个解压"]'),'selected label');
          document.querySelector('#toggle').click();
          await waitFor(()=>!document.querySelector('[role=tab]'),'closing dialog');
          document.querySelector('#toggle').click();
          (await publish()).click();
          await waitFor(()=>document.querySelector('.ant-select-selection-item[title="8个解压"]'),'label on reopen');
          const saved=JSON.parse(window.saved).publishSettings;
          if(saved.organization!=='7' || saved.websiteProfileId!=='121-8' || window.synced!==1) throw Error('saved mapping');
          document.body.dataset.result='passed';
        }
        run().catch(err=>{ document.body.dataset.error=err.message; });
      `},
      plugins:[{name:'fixture-directory',setup(plugin){
        plugin.onResolve({filter:/shared\/api\/(modelCatalog|batchFactoryV12|novelFetch|batchFactoryText)$/},
          args=>({path:args.path,namespace:'directory'}));
        plugin.onLoad({filter:/.*/,namespace:'directory'},()=>({contents:`
          export const getAvailableModelCatalog=async()=>({models:[]});
          export const modelSelectOptions=()=>[];
          export const getN8ComicPublishStatus=async()=>({});
          export const listPersonalTextBackgrounds=async()=>({items:[]});
          export const uploadPersonalTextBackground=async()=>({});
          export const syncWebSubmitConfigs=async()=>{window.synced=(window.synced||0)+1;
            return {groups:[{id:'121-8',name:'8个解压',organization:'7',advanced:{jieyaNum:8}}]};};
        `}));
      }}],
      bundle:true,jsx:'automatic',write:false,format:'iife',minify:true,define:{'process.env.NODE_ENV':'"production"'}
    });
    writeFileSync(join(dir,'form.js'),bundle.outputFiles[0].contents);
    const html = join(dir,'index.html');
    writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="form.js"></script>');
    const result = execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--virtual-time-budget=10000','--user-data-dir='+join(dir,'profile'),
      '--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,maxBuffer:8<<20,windowsHide:true,stdio:['ignore','pipe','pipe']});
    assert.ok(result.includes('data-result="passed"'), result.match(/data-error="([^"]*)"/)?.[1] || 'browser did not finish');
  } finally {
    rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:200});
  }
});

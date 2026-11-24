import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
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
        window.addEventListener('error',event=>{document.body.dataset.error=event.error?.message||event.message;});
        window.addEventListener('unhandledrejection',event=>{document.body.dataset.error=event.reason?.message||String(event.reason);});
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
          document.body.dataset.step='render';
          createRoot(document.querySelector('#root')).render(<Harness/>);
          await waitFor(()=>document.body.textContent.includes('按关键词标红'),'engine fields');
          if(document.body.textContent.includes('复用已有标红词，本地排版。')) throw Error('redundant highlight note');
          if(document.body.textContent.includes('系统优化在本地执行。')) throw Error('redundant sensitive-mode note');
          (await waitFor(()=>byText('[role=tab]','排版配置'),'layout tab')).click();
          await waitFor(()=>byText('.batch-factory-engine-group-title','正文排版'),'layout group');
          document.body.dataset.step='layout';
          const layoutGrid=document.querySelector('.batch-factory-engine-grid');
          const layoutColumns=getComputedStyle(layoutGrid).gridTemplateColumns;
          if(!layoutColumns.startsWith('repeat(2,') && layoutColumns.split(' ').length!==2) throw Error('layout is not two columns: '+layoutColumns);
          const layoutFields=[...layoutGrid.querySelectorAll('.batch-factory-engine-field')].slice(0,4);
          const layoutRows=layoutFields.map(field=>field.getBoundingClientRect().height);
          if(Math.max(...layoutRows)-Math.min(...layoutRows)>2) throw Error('layout rows are uneven');
          byText('button','纸感手写').click();
          await waitFor(()=>document.querySelector('.ant-select-selection-item[title="龙藏体 · 自然手写"]'),'Long Cang preset');
          byText('button','深色文字').click();
          await waitFor(()=>document.querySelector('[aria-label="背景颜色色值"]')?.value==='#F6EFDF','readable light background');
          byText('button','浅色文字').click();
          await waitFor(()=>document.querySelector('[aria-label="背景颜色色值"]')?.value==='#111318','readable dark background');
          const publish=()=>waitFor(()=>byText('[role=tab]','发布配置'),'publish tab');
          (await publish()).click();
          const publishGroup=await waitFor(()=>byText('.batch-factory-engine-group-title','发布映射'),'publish group');
          document.body.dataset.step='publish';
          const publishFields=[...publishGroup.closest('.batch-factory-engine-card').querySelectorAll('.batch-factory-engine-field')];
          if(publishFields.length!==5) throw Error('text upload controls escaped the shared grid');
          await waitFor(()=>document.body.textContent.includes('文案来源') && document.body.textContent.includes('音频匹配方式') && document.body.textContent.includes('作品匹配方式'),'shared image publish controls');
          (await waitFor(()=>byText('button','同步网站配置档'),'sync button')).click();
          await waitFor(()=>window.synced,'sync request');
          document.body.dataset.step='synced';
          const combo=await waitFor(()=>document.querySelector('[aria-label="网站配置档"]'),'profile field');
          combo.closest('.ant-select').querySelector('.ant-select-selector').dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));
          (await waitFor(()=>byText('.ant-select-item-option-content','8个解压'),'profile option')).click();
          await waitFor(()=>document.querySelector('.ant-select-selection-item[title="8个解压"]'),'selected label');
          document.body.dataset.step='selected';
          document.querySelector('#toggle').click();
          await waitFor(()=>!document.querySelector('[role=tab]'),'closing dialog');
          document.querySelector('#toggle').click();
          (await publish()).click();
          document.body.dataset.step='reopened';
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
      bundle:true,loader:{'.css':'empty'},jsx:'automatic',write:false,format:'iife',minify:true,define:{'process.env.NODE_ENV':'"production"'}
    });
    writeFileSync(join(dir,'styles.css'),readFileSync(new URL('../shuihuo-production.css',import.meta.url),'utf8')+'\n'+readFileSync(new URL('./publish-settings.css',import.meta.url),'utf8'));
    const html = join(dir,'index.html');
    const script=bundle.outputFiles[0].text.replaceAll('</script','<\\/script');
    writeFileSync(html,`<meta charset="utf-8"><link rel="stylesheet" href="styles.css"><div id="root"></div><script>${script}</script>`);
    const result = execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking',
      '--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1021,901','--virtual-time-budget=10000','--user-data-dir='+join(dir,'profile'),
      '--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout:60000,maxBuffer:8<<20,windowsHide:true,stdio:['ignore','pipe','pipe']});
    assert.ok(result.includes('data-result="passed"'), result.match(/data-error="([^"]*)"/)?.[1] || `browser stopped at ${result.match(/data-step="([^"]*)"/)?.[1] || 'unknown'} ${result.match(/<body[^>]*>/)?.[0] || ''}`);
  } finally {
    rmSync(dir,{recursive:true,force:true,maxRetries:4,retryDelay:200});
  }
});

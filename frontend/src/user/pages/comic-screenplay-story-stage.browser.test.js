import test from 'node:test';
import assert from 'node:assert/strict';
import { existsSync,mkdirSync,mkdtempSync,writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath,pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { build } from 'esbuild';
const here=path.dirname(fileURLToPath(import.meta.url));
const browser=[process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(file=>file&&existsSync(file));
const archive=path.resolve(here,'../../../../_local-archive-20261028/luoshui-screenplay-contract-20261114/ui-reconstruction');

async function runCase(name,contents,timeout=60000){
  const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents},bundle:true,jsx:'automatic',write:false,format:'iife',platform:'browser',loader:{'.css':'empty'}});
  mkdirSync(archive,{recursive:true});
  const dir=mkdtempSync(path.join(archive,name+'-'));
  writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);
  const html=path.join(dir,'test.html');
  writeFileSync(html,'<meta charset="utf-8"><div id="root"></div><script src="test.js"></script>');
  const output=execFileSync(browser,['--headless=new','--disable-gpu','--no-first-run','--disable-background-networking','--host-resolver-rules=MAP * ~NOTFOUND','--window-size=1200,900','--virtual-time-budget=12000','--user-data-dir='+path.join(dir,'profile'),'--dump-dom',pathToFileURL(html).href],{encoding:'utf8',timeout,windowsHide:true,stdio:['ignore','pipe','pipe'],maxBuffer:4000000});
  writeFileSync(path.join(dir,'result.html'),output);
  assert.ok(/data-result="passed"/.test(output),output.match(/data-error="([^"]*)"/)?.[1]||'browser did not finish');
}

test('剧本页顶部故事概要：结构化结果渲染标题/题材标签/摘要，空态显示可见的 AI 生成入口', {skip:!browser},async()=>{
 await runCase('story-stage-synopsis',`
import React from 'react';import{createRoot}from'react-dom/client';
import ComicScreenplayStoryStage from'./ComicScreenplayStoryStage.jsx';
const tick=()=>new Promise(resolve=>setTimeout(resolve,120));const assert=(v,m)=>{if(!v)throw Error(m)};
const project={id:'p',title:'并发实测项目',synopsis:''};
const doc1={id:'doc-1',content:'## 第1集\\n雨夜的一盏灯'};
const doc2={id:'doc-2',content:'## 第1集\\n雨夜的一盏灯'};
const synopsisResult={title:'少年爱辱',synopsis:'林川在一座南方小城长大，丧父后与母亲相依为命。高中毕业后进入照相馆工作，意外发现父亲日记里藏着一桩旧案，于是追查二十年前的雨夜车祸，最终真相大白、母子和解。',audience:'女频',genre:'现代言情',tags:['校园','青春','悬疑','甜宠']};
(async()=>{
 let generated=0;
 const root=createRoot(document.getElementById('root'));
 root.render(<ComicScreenplayStoryStage project={project} synopsisResult={synopsisResult} document={doc1} draft={doc1.content} onDraft={()=>{}} onSave={async()=>true} onTools={()=>{}} readOnly={false} working="" sourceReady={true} onGenerateSynopsis={()=>generated++}/>);
 for(let i=0;i<40&&!document.body.textContent.includes('少年爱辱');i++)await tick();
 assert(document.body.textContent.includes('少年爱辱'),'title from synopsis result missing');
 assert(document.querySelector('h1')?.textContent==='少年爱辱','title should prefer synopsis result over project title');
 assert(document.body.textContent.includes('女频')&&document.body.textContent.includes('现代言情'),'audience/genre tags missing');
 for(const t of ['校园','青春','悬疑','甜宠'])assert(document.querySelector('.comic-story-tags')?.textContent.includes(t),'tag '+t+' missing');
 assert(document.querySelector('.comic-story-tags .is-female'),'female audience tag class missing');
 assert(document.body.textContent.includes('林川在一座南方小城长大'),'synopsis text missing');
 assert(document.body.textContent.includes('故事概要'),'summary heading missing');
 // 有结果时主舞台仍能点 AI 生成
 const genBtn=[...document.querySelectorAll('button')].find(b=>b.textContent.includes('AI 生成'));
 assert(genBtn,'AI generate button missing with result');
 genBtn.click();await tick();
 assert(generated===1,'AI generate callback not fired with result');
 // 无结构化结果、无手动概要时：空态卡片 + 可见可用 AI 生成入口
 root.render(<ComicScreenplayStoryStage project={{id:'p2',title:'空态项目',synopsis:''}} synopsisResult={null} document={doc2} draft={doc2.content} onDraft={()=>{}} onSave={async()=>true} onTools={()=>{}} readOnly={false} working="" sourceReady={true} onGenerateSynopsis={()=>generated++}/>);
 for(let i=0;i<30;i++)await tick();
 assert(!document.querySelector('.comic-story-tags'),'tags should not render without result');
 assert(document.querySelector('.comic-story-summary'),'summary section must render in empty state');
 assert(document.body.textContent.includes('尚未生成故事概要'),'empty-state hint missing');
 const emptyGenBtn=[...document.querySelectorAll('button')].find(b=>b.textContent.includes('AI 生成'));
 assert(emptyGenBtn,'AI generate button missing in empty state');
 assert(!emptyGenBtn.disabled,'AI generate button should be enabled when sourceReady');
 emptyGenBtn.click();await tick();
 assert(generated===2,'AI generate callback not fired in empty state');
 document.body.dataset.result='passed';
})().catch(error=>{document.body.dataset.result='failed';document.body.dataset.error=error.message});
`);
});

import test from 'node:test';
import {existsSync,mkdirSync,mkdtempSync,readFileSync,writeFileSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
import {build} from 'esbuild';
import {runNovelBrowserDOM} from './novel-creation/browserEvidence.js';
const here=path.dirname(fileURLToPath(import.meta.url));
const browser=[process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(p=>p&&existsSync(p));
const fixture=String.raw`
import React,{useState} from 'react';import{createRoot}from'react-dom/client';import Dialog from './novel-creation/LongImitationDialog.jsx';import{readLongImitationDraft,writeLongImitationDraft}from'./novel-creation/longImitationDraftStore.js';
localStorage.clear();localStorage.setItem('auth_token','fixture');localStorage.setItem('auth_username','alice');
const tick=()=>new Promise(r=>setTimeout(r,100)),wait=async f=>{for(let n=0;n<100;n++){if(f())return;await tick()}throw Error('timeout '+window.step)},assert=(v,m)=>{if(!v)throw Error(m)},visible=e=>!!e&&e.getClientRects().length>0;
const button=(s,root=document)=>[...root.querySelectorAll('button')].find(e=>visible(e)&&(e.textContent.replace(/\s/g,'')===s.replace(/\s/g,'')||e.getAttribute('aria-label')===s));
const click=e=>{assert(e&&!e.disabled,'missing/disabled button '+window.step);e.click()},type=(label,value)=>{const e=document.querySelector('[aria-label="'+label+'"]');assert(e,'missing input '+label);Object.getOwnPropertyDescriptor(e.tagName==='TEXTAREA'?HTMLTextAreaElement.prototype:HTMLInputElement.prototype,'value').set.call(e,value);e.dispatchEvent(new Event('input',{bubbles:true}));};
const response=(data,status=200)=>new Response(JSON.stringify(data),{status,headers:{'Content-Type':'application/json'}});
window.calls=[];window.navigations=[];window.persisted=[];let saved=null;
const makeResult=body=>({project:{id:'new-book',title:body.title||'蓝图命名前的暂定书名',modelId:body.modelId},session:{id:'new-session',mode:'long_imitation'},run:{id:'new-run',status:'queued'},creation:{input:body}});
window.fetch=async(url,opt={})=>{url=String(url);const method=opt.method||'GET',body=opt.body instanceof FormData?{files:opt.body.getAll('files').map(f=>f.name)}:opt.body?JSON.parse(opt.body):{};window.calls.push({url,method,body});
 if(url.endsWith('/style-profiles'))return response({styles:[{id:'style-own',title:'已保存文风'}]});
 if(url.endsWith('/long-imitation/memos'))return response({memos:[{projectId:'source-project',id:'memo-own',title:'人物底线'}]});
 if(url.endsWith('/long-imitation/preview')){const data={statistics:{avgSentenceLength:body.referenceText.includes('旧')?11:22,sentenceLengthStdDev:1,avgParagraphLength:80,vocabularyDiversity:.5,topPatterns:[],rhetoricalFeatures:[]}};if(window.holdOldPreview&&body.referenceText.includes('旧'))return new Promise(r=>window.releasePreview=()=>r(response(data)));return response(data)}
 if(url.endsWith('/long-imitation/parse'))return response({files:[{filename:body.files[0],chapters:Array.from({length:7},(_,i)=>({id:'c'+(i+1),title:'第'+(i+1)+'章',content:'第'+(i+1)+'章\n独特内容'+(i+1)+'。'.repeat(100),characters:110}))}],failures:[]});
 if(url.includes('/long-imitation/requests/'))return saved?response(saved):response({error:'尚未找到创建结果'},404);
 if(url.endsWith('/long-imitation')&&method==='POST'){if(window.reject)return response({error:'参考配置已失效，请核对'},422);if(window.loseBeforeCommit){window.loseBeforeCommit=false;throw Error('创建回包丢失')}saved=makeResult(body);if(window.holdStart)return new Promise(r=>window.releaseStart=()=>r(response(saved,202)));if(window.loseReply){window.loseReply=false;throw Error('创建回包丢失')}return response(saved,202)}
 return response({});};
let root=createRoot(document.getElementById('root'));
function Host(){const[open,setOpen]=useState(true),[owner,setOwner]=useState('alice'),[entry,setEntry]=useState(1);window.closeDialog=()=>setOpen(false);window.reopen=()=>{setEntry(n=>n+1);setOpen(true)};window.switchOwner=()=>setOwner('bob');window.restoreOwner=()=>setOwner('alice');return <Dialog key={owner} owner={owner} open={open} viewKey={owner+':'+entry} models={[{value:'own-model',label:'我们的文本模型'}]} modelId="own-model" thinkingLevel="deep" onCancel={()=>setOpen(false)} onPersisted={r=>window.persisted.push(r)} onCreated={r=>{window.navigations.push(r);setOpen(false)}}/>}
// Cold IndexedDB on a busy workstation is setup, before any tested operation.
const open=async()=>{root.render(<Host/>);const deadline=Date.now()+45000;while(Date.now()<deadline&&(!button('创建作品')||button('创建作品').disabled))await tick();assert(button('创建作品')&&!button('创建作品').disabled,'initial browser draft read timed out');await tick()};const remount=async()=>{root.unmount();root=createRoot(document.getElementById('root'));root.render(<Host/>);await wait(()=>button('核对创建结果'));await tick()};
const starts=()=>window.calls.filter(c=>c.method==='POST'&&c.url.endsWith('/long-imitation'));
const choose=async(label,text)=>{const el=document.querySelector('[aria-label="'+label+'"]');el.closest('.ant-select').querySelector('.ant-select-selector').dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));await tick();const option=[...document.querySelectorAll('.ant-select-item-option')].find(e=>visible(e)&&e.textContent.includes(text));assert(option,'option absent');option.click();await tick();document.querySelector('[aria-label="暂定标题（可选）"]').click();await tick()};
`;
async function run(code,name){const bundle=await build({stdin:{resolveDir:here,loader:'jsx',contents:fixture+`\n(async()=>{${code};document.body.dataset.result='passed'})().catch(e=>{document.body.dataset.result='failed';document.body.dataset.error=e.message});`},bundle:true,write:false,minify:true,jsx:'automatic',define:{'process.env.NODE_ENV':'"production"'},loader:{'.css':'empty'}});const base=path.resolve(here,'../../../../_local-archive-20261028/long-imitation-browser');mkdirSync(base,{recursive:true});const dir=mkdtempSync(path.join(base,name+'-'));writeFileSync(path.join(dir,'test.js'),bundle.outputFiles[0].text);writeFileSync(path.join(dir,'test.html'),'<meta charset="utf-8"><style>body{margin:0;background:#f5f5f7}:root{--text-primary:#25262c;--text-secondary:#777;--surface-panel:#fff;--surface-elevated:#fff;--border-default:#ddd}'+readFileSync(path.join(here,'novel-creation/LongImitationDialog.css'),'utf8')+'</style><div id="root"></div><script src="test.js"></script>');await runNovelBrowserDOM(browser,path.join(dir,'test.html'),dir)}

test('仿写弹窗空值校验不建书且完整字段提交自家模型',{skip:!browser},async()=>run(String.raw`
 await open();window.step='whole payload';click(button('创建作品'));await tick();assert(document.body.textContent.includes('至少需要 100 字')&&starts().length===0,'empty created book');type('参考原文','字'.repeat(99)+'\n   ');await tick();click(button('创建作品'));await tick();assert(starts().length===0,'whitespace bypassed minimum');type('参考原文','原文'.repeat(60));type('暂定标题（可选）','参考表单新作');type('读者偏好','悬疑');type('原创故事想法（可选）','山村送信人的原创故事');type('仿写方向（可选）','方向二：从失信开始');type('补充信息（可选）','结局保留希望');await tick();click(button('1万'));click(button('300'));const style=[...document.querySelectorAll('.ant-checkbox-wrapper')].find(e=>e.textContent==='已保存文风');style.querySelector('input').click();await choose('关联备忘录（可选）','人物底线');click(button('创建作品'));await wait(()=>window.navigations.length===1);const body=starts()[0].body;assert(body.targetWordCount===10000&&body.targetChapters===300,'preset fields dropped');assert(body.readerPreferences==='悬疑'&&body.storyIdea==='山村送信人的原创故事'&&body.selectedDirection==='方向二：从失信开始'&&body.supplementaryInfo==='结局保留希望','optional fields dropped');assert(body.modelId==='own-model'&&body.thinkingLevel==='deep','own model dropped');assert(body.knowledgeBookIds[0]==='style-own'&&body.memoReferences[0].noteId==='memo-own','references dropped');assert(starts().length===1,'duplicate create');
`,'whole-payload'));

test('仿写上传默认最近五章，自选按原顺序回填，取消选择不改原稿',{skip:!browser},async()=>run(String.raw`
 await open();window.step='chapter selection';const input=document.querySelector('[aria-label="上传参考文件"]');const transfer=new DataTransfer();transfer.items.add(new File(['测试上传'],'参考.txt',{type:'text/plain',lastModified:7}));input.files=transfer.files;input.dispatchEvent(new Event('change',{bubbles:true}));await wait(()=>button('选择章节'));const content=()=>document.querySelector('[aria-label="参考原文"]').value;assert(!content().includes('独特内容1')&&content().includes('独特内容3')&&content().includes('独特内容7'),'not recent five');click(button('选择章节'));await wait(()=>document.querySelector('.long-imitation-chapter-picker .ant-modal'));let picker=document.querySelector('.long-imitation-chapter-picker .ant-modal');click(button('清空',picker));click(button('取消',picker));await wait(()=>!visible(picker));assert(content().includes('独特内容3'),'cancel changed selected chapters');click(button('选择章节'));await wait(()=>visible(document.querySelector('.long-imitation-chapter-picker .ant-modal')));picker=document.querySelector('.long-imitation-chapter-picker .ant-modal');click(button('清空',picker));await tick();const choices=picker.querySelectorAll('.long-imitation-picker-list input[type="checkbox"]');choices[6].click();choices[0].click();await tick();click(button('确认选择',picker));await wait(()=>!visible(picker));assert(content().indexOf('独特内容1')<content().indexOf('独特内容7')&&!content().includes('独特内容3'),'custom selection order incorrect');
`,'chapter-selection'));

test('仿写创建丢回包后刷新只查询原编号，不再调用创建接口',{skip:!browser},async()=>run(String.raw`
 await open();window.step='receipt recovery';type('参考原文','参考'.repeat(60));await tick();window.loseReply=true;click(button('创建作品'));await wait(()=>button('核对创建结果'));await remount();click(button('核对创建结果'));await wait(()=>window.navigations.length===1);assert(starts().length===1,'recovery recreated book');assert(window.calls.some(c=>c.url.includes('/requests/'+starts()[0].body.idempotencyKey)),'receipt key changed');
`,'receipt-recovery'));

test('仿写未找到结果时只明确重提同编号，双击只发一次',{skip:!browser},async()=>run(String.raw`
 await open();window.step='missing receipt';type('参考原文','参考'.repeat(60));await tick();window.loseBeforeCommit=true;click(button('创建作品'));await wait(()=>button('核对创建结果'));click(button('核对创建结果'));await wait(()=>button('重新提交原请求'));const buttonAgain=button('重新提交原请求');buttonAgain.click();buttonAgain.click();await wait(()=>window.navigations.length===1);assert(starts().length===2,'duplicate retry');assert(JSON.stringify(starts()[0].body)===JSON.stringify(starts()[1].body),'retry mutated frozen payload');
`,'missing-receipt'));

test('仿写明确拒绝保留输入并解除锁定，关闭重开读回草稿',{skip:!browser},async()=>run(String.raw`
 await open();window.step='rejection';type('参考原文','保留原文'.repeat(30));type('暂定标题（可选）','不能丢失');await tick();window.reject=true;click(button('创建作品'));await wait(()=>document.body.textContent.includes('参考配置已失效'));await wait(()=>!document.querySelector('[aria-label="参考原文"]').disabled);assert(document.querySelector('[aria-label="参考原文"]').value==='保留原文'.repeat(30),'rejection lost draft');click(button('取消'));await tick();window.reopen();await wait(()=>button('创建作品'));assert(document.querySelector('[aria-label="暂定标题（可选）"]').value==='不能丢失','close lost draft');
`,'definite-rejection'));

test('关闭再打开或切账号时，迟到创建不抢导航且不泄露原文',{skip:!browser},async()=>run(String.raw`
 await open();window.step='late create';type('参考原文','旧账号私有原文'.repeat(20));await tick();window.holdStart=true;click(button('创建作品'));await wait(()=>window.releaseStart);window.switchOwner();await tick();await tick();window.releaseStart();await tick();await tick();assert(window.navigations.length===0&&window.persisted.length===0,'old account navigated');assert(document.querySelector('[aria-label="参考原文"]').value==='','old account draft leaked');
`,'late-owner'));

test('过时文风预览不能覆盖新原文的统计结果',{skip:!browser},async()=>run(String.raw`
 await open();window.step='preview ordering';window.holdOldPreview=true;type('参考原文','旧原文'.repeat(40));await wait(()=>window.releasePreview);type('参考原文','新原文'.repeat(40));await wait(()=>document.querySelector('.long-imitation-statistics')?.textContent.includes('22字'));window.releasePreview();await tick();await tick();assert(document.querySelector('.long-imitation-statistics').textContent.includes('22字'),'old preview replaced new reference');assert(starts().length===0,'preview created book');
`,'preview-ordering'));


test('输入后立即切账号再返回不丢最后一段草稿',{skip:!browser},async()=>run(String.raw`
 await open();window.step='immediate owner switch';type('参考原文','最后一段必须保存'.repeat(20));window.switchOwner();await tick();await wait(()=>button('创建作品')&&!button('创建作品').disabled);assert(document.querySelector('[aria-label="参考原文"]').value==='','bob received alice draft');window.restoreOwner();await tick();await wait(()=>button('创建作品')&&!button('创建作品').disabled);assert(document.querySelector('[aria-label="参考原文"]').value==='最后一段必须保存'.repeat(20),'immediate switch lost final input');assert(starts().length===0,'saving draft created book');
`,'immediate-owner-draft'));


test('长参考草稿存储不占localStorage且按账号隔离',{skip:!browser},async()=>run(String.raw`
 window.step='large durable draft';const text='长参考'.repeat(250000);await writeLongImitationDraft('large-alice',{draft:{referenceText:text},intent:{payload:{idempotencyKey:'frozen-test',referenceText:text}}});const saved=await readLongImitationDraft('large-alice');assert(saved.draft.referenceText===text&&saved.intent.payload.referenceText===text,'large draft truncated');assert(await readLongImitationDraft('large-bob')===null,'large draft crossed owner');assert(!Object.values(localStorage).some(v=>v.includes('长参考')),'large reference used small localStorage quota');await writeLongImitationDraft('large-alice',null);assert(starts().length===0,'storage test generated content');
`,'large-draft-store'));

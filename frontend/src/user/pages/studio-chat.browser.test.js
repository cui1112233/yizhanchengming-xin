import test from 'node:test';
import { existsSync, readFileSync, mkdirSync, mkdtempSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { build } from 'esbuild';
import { runNovelBrowserDOM } from './novel-creation/browserEvidence.js';
const here=path.dirname(fileURLToPath(import.meta.url));
const browser=[process.env.CHROME_PATH,'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(x=>x&&existsSync(x));
const fixture=String.raw`
import React,{useState} from 'react';import {createRoot} from 'react-dom/client';import StudioChat from './novel-creation/StudioChat.jsx';
localStorage.clear();localStorage.setItem('auth_token','fixture');localStorage.setItem('auth_username','alice');
const tick=()=>new Promise(r=>setTimeout(r,80)),assert=(v,m)=>{if(!v)throw Error(m)},wait=async f=>{for(let n=0;n<90;n++){if(f())return;await tick()}throw Error('timeout '+window.step)};
const button=s=>[...document.querySelectorAll('button')].find(e=>e.getClientRects().length&&(e.textContent.replace(/\s/g,'')===s.replace(/\s/g,'')||e.getAttribute('aria-label')===s));
const click=e=>{assert(e&&!e.disabled,'button absent/disabled '+window.step);e.click()},type=(e,v)=>{assert(e,'input absent '+window.step);Object.getOwnPropertyDescriptor(e.tagName==='TEXTAREA'?HTMLTextAreaElement.prototype:HTMLInputElement.prototype,'value').set.call(e,v);e.dispatchEvent(new Event('input',{bubbles:true}))};
const response=(d,s=200)=>new Response(JSON.stringify(d),{status:s,headers:{'Content-Type':'application/json'}});
const session={id:'studio1',mode:'long_original',title:'未建书会话'},proposal={id:'action1',type:'create_book',status:'pending',book:{title:'雨夜归人',genre:'悬疑',world:'旧城',protagonist:'送信人',coreConflict:'一封旧信',targetChapterWords:2000}};
let snap={session,messages:[],turn:null};window.calls=[];window.plannerCalls=0;window.transitions=[];
window.fetch=async(url,opt={})=>{url=String(url);const method=opt.method||'GET',body=opt.body?JSON.parse(opt.body):{};window.calls.push({url,method,body});if(url.endsWith('/studio-sessions')&&method==='POST'){if(window.loseCreation){window.loseCreation=false;throw Error('创建回包丢失')}return response({session},201);}if(url.endsWith('/studio-sessions/studio1')&&method==='GET')return response(snap);
if(url.endsWith('/messages')){if(window.rejectSend)return response({error:'模型当前不可用'},422);window.plannerCalls++;snap={session,messages:[{id:'u1',role:'user',content:body.content},{id:'a1',role:'assistant',content:'可以建立这部悬疑小说，请确认设定。'}],pendingAction:structuredClone(proposal),turn:{id:'turn1',idempotencyKey:body.idempotencyKey,state:'completed',modelId:body.modelId}};if(window.invalidReply)snap={...snap,pendingAction:null,turn:{...snap.turn,state:'invalid_output',errorMessage:'回复格式无效'}};if(window.unknownReply)snap={...snap,pendingAction:null,turn:{...snap.turn,state:'unknown',errorMessage:'结果待确认'}};if(window.holdSend)return new Promise(resolve=>{window.releaseSend=()=>resolve(response(snap))});if(window.loseReply){window.loseReply=false;throw Error('连接中断')}return response(snap)}
if(url.endsWith('/actions/cancel')){snap={...snap,pendingAction:{...proposal,status:'cancelled'}};return response(snap)}
if(url.endsWith('/actions/confirm')){if(window.failConfirm){window.failConfirm=false;return response({error:'确认暂时失败'},500)}return response({session:{...session,projectId:'p-new'},project:{id:'p-new',title:'雨夜归人'},productionSession:{id:'prod1',mode:'long_original'},run:{id:'run1',status:'queued'}})}return response({});};
let root=createRoot(document.getElementById('root'));
function Host(){const[id,setID]=useState(''),[owner,setOwner]=useState('alice');window.switchOwner=()=>{setID('');setOwner('bob')};return <StudioChat key={owner+':'+id} owner={owner} sessionId={id} models={[{value:'m1',label:'授权模型'}]} onSessionSelected={setID} onSessionCreated={()=>{}} onTransition={r=>window.transitions.push(r)} />}
const open=async()=>{root.render(<Host/>);await wait(()=>document.querySelector('[aria-label="创作要求"]'));await tick()};const remount=async()=>{root.unmount();root=createRoot(document.getElementById('root'));await open()};
`;
async function run(code,label){const bundled=await build({stdin:{resolveDir:here,loader:'jsx',contents:fixture+`\n(async()=>{${code};document.body.dataset.result='passed'})().catch(e=>{document.body.dataset.result='failed';document.body.dataset.error=e.message;document.body.dataset.calls=JSON.stringify(window.calls)});`},bundle:true,write:false,minify:true,jsx:'automatic',define:{'process.env.NODE_ENV':'"production"'},loader:{'.css':'empty'}});const base=path.resolve(here,'../../../../_local-archive-20261028/novel-prebook-browser');mkdirSync(base,{recursive:true});const dir=mkdtempSync(path.join(base,label+'-'));writeFileSync(path.join(dir,'test.js'),bundled.outputFiles[0].text);writeFileSync(path.join(dir,'test.html'),'<meta charset="utf-8"><style>body{margin:0}:root{--text-primary:#172033;--text-secondary:#666;--surface-panel:#fff;--surface-elevated:#fff;--border-default:#ddd}'+readFileSync(path.join(here,'novel-creation/StudioChat.css'),'utf8')+'</style><div id="root"></div><script src="test.js"></script>');await runNovelBrowserDOM(browser,path.join(dir,'test.html'),dir)}
test('无书会话真实规划后确认才创建，取消与重开不生成',{skip:!browser},async()=>run(String.raw`
await open();window.step='empty';click(button('发送'));await tick();assert(window.plannerCalls===0,'empty generated');type(document.querySelector('[aria-label="创作要求"]'),'写一部雨夜送信人的悬疑长篇');await tick();click(button('发送'));await wait(()=>button('确认创建作品'));assert(window.plannerCalls===1&&window.transitions.length===0,'proposal created project');assert(document.body.textContent.includes('雨夜归人'),'proposal missing');click(button('取消创建'));await wait(()=>document.body.textContent.includes('已取消'));assert(!window.calls.some(c=>c.url.endsWith('/actions/confirm')),'cancel confirmed');await remount();assert(window.plannerCalls===1,'remount regenerated');
`,'cancel'));
test('规划回包丢失只核对原请求，确认失败保留动作再恢复',{skip:!browser},async()=>run(String.raw`
await open();window.step='lost';window.loseReply=true;type(document.querySelector('[aria-label="创作要求"]'),'写一部悬疑长篇');await tick();click(button('发送'));await wait(()=>button('确认创建作品')||button('核对原请求'));if(button('核对原请求'))click(button('核对原请求'));await wait(()=>button('确认创建作品'));assert(window.plannerCalls===1,'lost reply rebilled');window.failConfirm=true;click(button('确认创建作品'));await wait(()=>document.body.textContent.includes('确认暂时失败'));assert(button('确认创建作品'),'failure lost action');click(button('确认创建作品'));await wait(()=>window.transitions.length===1);assert(window.plannerCalls===1,'confirm planned again');assert(window.calls.filter(c=>c.url.endsWith('/actions/confirm')).every(c=>c.body.actionId==='action1'&&c.body.modelId==='m1'),'confirmation changed action/model');
`,'recovery'));


test('模型坏格式保留原草稿且不自动再请求',{skip:!browser},async()=>run(String.raw`
await open();window.step='invalid reply';window.invalidReply=true;type(document.querySelector('[aria-label="创作要求"]'),'不能清除的用户原稿');await tick();click(button('发送'));await wait(()=>document.body.textContent.includes('回复格式无效'));await wait(()=>!document.querySelector('[aria-label="创作要求"]').disabled);assert(document.querySelector('[aria-label="创作要求"]').value==='不能清除的用户原稿','invalid reply cleared draft');assert(window.plannerCalls===1&&!button('确认创建作品'),'invalid reply created/rebilled');
`,'invalid'));
test('结果未知只GET核对，切账号不接收旧回包',{skip:!browser},async()=>run(String.raw`
await open();window.step='unknown';window.unknownReply=true;type(document.querySelector('[aria-label="创作要求"]'),'旧账号私有创作要求');await tick();click(button('发送'));await wait(()=>document.body.textContent.includes('上游结果待确认'));assert(button('发送').disabled,'unknown enabled resend');click(button('核对原请求'));await tick();assert(window.plannerCalls===1,'unknown called model again');window.switchOwner();await tick();await tick();assert(!document.body.textContent.includes('旧账号私有创作要求'),'account switch exposed old history');assert(document.querySelector('[aria-label="创作要求"]').value==='','account draft leaked');
`,'unknown'));
test('旧账号在途规划回包不抢占新账号',{skip:!browser},async()=>run(String.raw`
await open();window.step='late';window.holdSend=true;type(document.querySelector('[aria-label="创作要求"]'),'旧账号私有创作要求');await tick();click(button('发送'));await wait(()=>window.releaseSend);window.switchOwner();await tick();window.releaseSend();await tick();await tick();assert(!document.body.textContent.includes('旧账号私有创作要求')&&!button('确认创建作品'),'late reply replaced new view');assert(document.querySelector('[aria-label="创作要求"]').value==='','late reply replaced draft');assert(window.transitions.length===0,'late reply navigated');
`,'late'));


test('新会话发送前拒绝后迁移完整草稿并解除锁定',{skip:!browser},async()=>run(String.raw`
await open();window.step='pre-send reject';window.rejectSend=true;type(document.querySelector('[aria-label="创作要求"]'),'创建会话后也不能丢的原稿');await tick();click(button('发送'));await wait(()=>document.body.textContent.includes('模型当前不可用'));await wait(()=>!document.querySelector('[aria-label="创作要求"]').disabled);assert(document.querySelector('[aria-label="创作要求"]').value==='创建会话后也不能丢的原稿','new session binding lost rejected draft');assert(window.plannerCalls===0,'pre-send rejection billed');
`,'rejected-draft'));


test('创建回包丢失核对不发送，明确发送前拒绝仍保留草稿',{skip:!browser},async()=>run(String.raw`
await open();window.step='creation recovery';window.loseCreation=true;type(document.querySelector('[aria-label="创作要求"]'),'创建丢回包也保留的原稿');await tick();click(button('发送'));await wait(()=>button('核对原请求'));assert(window.plannerCalls===0,'creation loss sent model');click(button('核对原请求'));await wait(()=>button('发送原消息'));assert(window.plannerCalls===0,'checking creation generated');window.rejectSend=true;click(button('发送原消息'));await wait(()=>document.body.textContent.includes('模型当前不可用'));await wait(()=>!document.querySelector('[aria-label="创作要求"]').disabled);assert(document.querySelector('[aria-label="创作要求"]').value==='创建丢回包也保留的原稿','recovery rejection lost draft');assert(!button('核对原请求'),'definitive rejection kept intent lock');
`,'creation-recovery'));


test('紧凑输入区的图标设置真实进入发送参数',{skip:!browser},async()=>run(String.raw`
 await open();window.step='compact settings';assert(!document.querySelector('[aria-label="上下文字数"]'),'settings exposed as a large form');click(button('调整上下文设置'));await wait(()=>document.querySelector('[aria-label="上下文字数"]'));type(document.querySelector('[aria-label="上下文字数"]'),'12000');document.querySelector('[aria-label="上下文字数"]').dispatchEvent(new Event('blur',{bubbles:true}));await tick();click(button('调整上下文设置'));click(button('调整思考要求'));await wait(()=>document.querySelector('[aria-label="创作思考要求"]'));const select=document.querySelector('[aria-label="创作思考要求"]').closest('.ant-select');select.querySelector('.ant-select-selector').dispatchEvent(new MouseEvent('mousedown',{bubbles:true}));await tick();const option=[...document.querySelectorAll('.ant-select-item-option')].find(e=>e.textContent==='深度思考');assert(option,'thinking option absent');option.click();await tick();assert(window.plannerCalls===0,'opening settings called model');type(document.querySelector('[aria-label="创作要求"]'),'紧凑输入区的创作要求');await tick();click(button('发送'));await wait(()=>window.plannerCalls===1);const post=window.calls.find(c=>c.url.endsWith('/messages'));assert(post.body.contextBudget===12000,'context setting was decorative');assert(post.body.thinkingLevel==='deep','thinking setting was decorative');
`,'compact-settings'));

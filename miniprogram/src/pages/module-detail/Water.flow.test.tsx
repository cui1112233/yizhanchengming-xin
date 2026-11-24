// @vitest-environment jsdom
import {act} from 'react';
import {createRoot,type Root} from 'react-dom/client';
import {afterEach,beforeEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({show:()=>{},hide:()=>{},api:vi.fn(),modal:vi.fn(),navigate:vi.fn(),projects:[] as any[],tasks:[] as any[]}));
vi.mock('@tarojs/taro',()=>({default:{showModal:h.modal,navigateTo:h.navigate,setClipboardData:vi.fn()},useDidShow:(fn:any)=>{h.show=fn;},useDidHide:(fn:any)=>{h.hide=fn;}}));
vi.mock('@tarojs/components',()=>({View:'div',Text:'span',Button:'button',Input:({onInput,maxlength,...props}:any)=><input {...props} onInput={event=>onInput?.({detail:{value:event.currentTarget.value}})}/>,Textarea:({onInput,maxlength,...props}:any)=><textarea {...props} onInput={event=>onInput?.({detail:{value:event.currentTarget.value}})}/>,Image:'img',Video:'video',Audio:'audio',ScrollView:'div',Picker:({id,range=[],value=0,onChange,children}:any)=><div id={id}>{children}<select id={`${id}-select`} value={value} onChange={event=>onChange({detail:{value:event.target.value}})}>{range.map((label:string,index:number)=><option key={index} value={index}>{label}</option>)}</select></div>}));
vi.mock('../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../services/mediaCache',()=>({loadAuthenticatedMedia:vi.fn()}));
vi.mock('../../theme',()=>({withThemedPage:(p:any)=>p}));
vi.mock('../../components/AppIcon',()=>({default:()=>null}));
vi.mock('../../components/BrandHeader',()=>({default:()=>null}));
vi.mock('../../components/AmbientStarfield',()=>({default:()=>null}));
vi.mock('../batch/BatchCover',()=>({default:()=>null}));
import Workspace from './ShuihuoWorkspace';
let root:Root,container:HTMLDivElement;
const models=[{id:'text-1',name:'文本',kind:'text',enabled:true},{id:'image-1',modelId:'provider-image-1',name:'同名图片模型',kind:'image',enabled:true},{id:'image-2',modelId:'provider-image-2',name:'同名图片模型',kind:'image',enabled:true}];
function doc(id='p1'){return {project:{id,name:id==='p1'?'小城故事':'恢复项目',sourceText:'保存的原文',sourceRevision:1},segments:[{id:'s1',subtitleText:'原分镜正文',speaker:'旁白',confirmed:true}],assets:[{id:'a1',name:'林夏',category:'character',description:'旧描述'}],media:[]};}
beforeEach(()=>{vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT',true);container=document.createElement('div');root=createRoot(container);h.modal.mockResolvedValue({confirm:true});h.api.mockReset();h.projects=[{id:'p1',name:'小城故事'}];h.tasks=[];h.api.mockImplementation(async(path:string,options:any)=>{if(path.endsWith('/batches/summary'))return{batches:[]};if(path.endsWith('/models'))return{models};if(path.endsWith('/projects')&&options?.method==='POST'){const created={id:'p2',name:options.data.name};h.projects.push(created);return created;}if(path.endsWith('/projects'))return{projects:h.projects};if(path.endsWith('/tasks')&&options?.method==='POST'){h.tasks=[{id:'t1',status:'queued',kind:options.data.kind}];return{task:h.tasks[0]};}if(path.endsWith('/tasks'))return{tasks:h.tasks};if(path.endsWith('/segmentation/paragraphs'))return{candidates:[{text:'候选分镜',speaker:'旁白'}]};if(/\/projects\/p[12]$/.test(path))return doc(path.endsWith('p2')?'p2':'p1');return{};});});
afterEach(async()=>{await act(async()=>root.unmount());vi.unstubAllGlobals();});
async function mount(){await act(async()=>root.render(<Workspace/>));await act(async()=>h.show());}
function button(label:string){const element=[...container.querySelectorAll('button')].find(item=>item.textContent===label||item.getAttribute('aria-label')===label);expect(element,label).toBeTruthy();return element!;}
async function click(label:string){await act(async()=>button(label).click());}
async function input(id:string,value:string){const element=container.querySelector(`#${id}`) as HTMLInputElement;expect(element,id).toBeTruthy();await act(async()=>{element.value=value;element.dispatchEvent(new Event('input',{bubbles:true}));});}
async function openProject(){await click('小城故事');}
test('the library and create form are separate views and opening them never creates a project',async()=>{
 await mount();expect(container.querySelector('textarea')).toBeNull();await click('新建漫剧');expect(container.querySelector('#water-create-source')).not.toBeNull();expect(h.api.mock.calls.some(([,o])=>o?.method==='POST')).toBe(false);
});
test('cancelled segment editing leaves persisted source intact and sends no write',async()=>{
 await mount();await openProject();await click('分镜');await click('编辑分镜');await input('water-editor-subtitleText','未保存改稿');await click('取消');expect(container.textContent).toContain('原分镜正文');expect(container.textContent).not.toContain('未保存改稿');expect(h.api.mock.calls.some(([,o])=>o?.method==='PUT')).toBe(false);
});
test('source changes invalidate preview candidates before confirmation',async()=>{
 await mount();await openProject();await click('预览分镜');expect(container.textContent).toContain('候选分镜');await click('原文');await input('water-source','改过的原文');await click('分镜');expect(container.textContent).not.toContain('确认采用候选');expect(h.api.mock.calls.some(([p])=>p.endsWith('/segmentation/confirm'))).toBe(false);
});
test('duplicate model names still select their real ids and double taps create one task',async()=>{
 await mount();await openProject();await click('分镜');await click('生成素材');const select=container.querySelector('#water-generation-model-select') as HTMLSelectElement;expect(select).toBeTruthy();await act(async()=>{select.value='1';select.dispatchEvent(new Event('change',{bubbles:true}));});
 const start=button('开始生成');await act(async()=>{start.click();start.click();});const writes=h.api.mock.calls.filter(([p,o])=>p.endsWith('/tasks')&&o?.method==='POST');expect(writes).toHaveLength(1);expect(writes[0][1].data).toEqual({segmentId:'s1',modelId:'provider-image-2',kind:'image'});
});
test('partial creation keeps the created project available without a second POST',async()=>{
 const normal=h.api.getMockImplementation()!;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/p2/source'))throw Error('保存超时');return normal(p,o);});await mount();await click('新建漫剧');await input('water-create-name','恢复项目');await input('water-create-source','待保存的原文');await click('保存项目');expect(container.textContent).toContain('已创建');await click('打开已创建项目');expect(h.api.mock.calls.filter(([p,o])=>p.endsWith('/projects')&&o?.method==='POST')).toHaveLength(1);
});
test('task retry and cancellation target only the chosen task',async()=>{
 h.tasks=[{id:'failed-1',status:'failed',kind:'image'},{id:'running-1',status:'running',kind:'video'}];await mount();await openProject();await click('任务');await click('重试任务');await click('取消任务');expect(h.api.mock.calls.some(([p,o])=>p.endsWith('/tasks/failed-1/retry')&&o?.method==='POST')).toBe(true);expect(h.api.mock.calls.some(([p,o])=>p.endsWith('/tasks/running-1/cancel')&&o?.method==='PUT')).toBe(true);
});

test('refresh while editing does not silently rebase a source draft over a newer server revision',async()=>{
 await mount();await openProject();await input('water-source','本地保留草稿');
 const normal=h.api.getMockImplementation()!;
 h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/projects/p1')){const next=doc();next.project.sourceRevision=2;next.project.sourceText='其他设备改稿';return next;}return normal(p,o);});
 await click('刷新项目');await click('保存原文');expect(container.textContent).toContain('服务端原文已更新');expect((container.querySelector('#water-source') as HTMLTextAreaElement).value).toBe('本地保留草稿');expect(h.api.mock.calls.some(([p,o])=>p.endsWith('/source')&&o?.method==='PUT')).toBe(false);
});

test('an empty asset name stays editable without a request or an unknown-result lock',async()=>{
 await mount();await openProject();await click('素材');await click('新建资产');await click('保存资产');expect(container.textContent).toContain('请填写资产名称');expect(container.textContent).not.toContain('结果未知');await input('water-editor-name','新人物');await click('保存资产');expect(h.api.mock.calls.filter(([p,o])=>p.endsWith('/p1/assets')&&o?.method==='POST')).toHaveLength(1);
});

test('a recovered project does not prevent creating the next project',async()=>{
 const normal=h.api.getMockImplementation()!;let fail=true;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/p2/source')&&fail)throw Error('保存超时');return normal(p,o);});
 await mount();await click('新建漫剧');await input('water-create-name','恢复项目');await input('water-create-source','待保存的原文');await click('保存项目');await click('打开已创建项目');fail=false;await click('保存原文');await click('返回作品库');await click('新建漫剧');expect(container.textContent).not.toContain('打开已创建项目');await input('water-create-name','下一部');await input('water-create-source','新原文');expect(button('保存项目').disabled).toBe(false);await click('保存项目');expect(h.api.mock.calls.filter(([p,o])=>p.endsWith('/projects')&&o?.method==='POST')).toHaveLength(2);
});

test('the image dialog exposes no single-segment video or selectable audio model',async()=>{
 await mount();await openProject();await click('分镜');await click('生成素材');expect(container.textContent).toContain('生成分镜图片');expect(container.querySelector('#water-generation-kind')).toBeNull();expect(container.textContent).toContain('视频需在作品工作区按整部作品执行');
});

test('a timed-out source save that already persisted is recovered by reading it back',async()=>{
 const normal=h.api.getMockImplementation()!;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/p2/source'))throw Error('响应超时');if(p.endsWith('/projects/p2')){const next=doc('p2');next.project.sourceText='实际已保存原文';return next;}return normal(p,o);});
 await mount();await click('新建漫剧');await input('water-create-name','恢复项目');await input('water-create-source','实际已保存原文');await click('保存项目');await click('打开已创建项目');await click('返回作品库');await click('新建漫剧');expect(container.textContent).not.toContain('打开已创建项目');await input('water-create-name','第二部');await input('water-create-source','第二部原文');expect(button('保存项目').disabled).toBe(false);expect(h.api.mock.calls.filter(([p,o])=>p.endsWith('/projects')&&o?.method==='POST')).toHaveLength(1);
});

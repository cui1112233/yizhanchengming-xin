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
import Workspace from './CmWorkspace';
import {invalidateSessionResources} from '../../services/sessionBoundary';
let root:Root,container:HTMLDivElement;
const document=()=>({version:1,revision:4,viewport:{zoom:0.8},appearance:{edgeColor:'cyan'},nodes:[{id:'n1',type:'image',title:'封面图',version:3,config:{prompt:'原要求',parameters:{ratio:'9:16'}},position:{x:20,y:30},sourceNodeIds:[]}],edges:[{id:'keep',source:'other',target:'n1',relation:'reference',extra:'保留'}]});
beforeEach(()=>{invalidateSessionResources();vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT',true);container=globalThis.document.createElement('div');root=createRoot(container);h.modal.mockResolvedValue({confirm:true});h.api.mockReset();h.api.mockImplementation(async(p:string,o:any)=>{
 if(p==='/api/agent/tasks')return o?.method==='POST'?{task:{id:'p2',title:'新项目'}}:{tasks:[{id:'p1',title:'既有CM项目',projectType:'comic',messageCount:1}]};
 if(p.endsWith('/projects'))throw Error('禁止Studio');if(p==='/api/agent/skills')return{skills:[]};if(p.includes('/api/models?'))return{models:[{id:'channel-1',kind:p.split('kind=')[1],provider:'custom',name:'已配置模型'}]};
 if(p.endsWith('/canvas')){if(o?.method==='PATCH')return{canvas:{...o.data.canvas,revision:5}};return{canvas:document()};}
 if(p.endsWith('/image-generations'))return{generation:{id:'g1',state:'queued'}};
 if(p.endsWith('/generations'))return{tasks:[]};if(p.endsWith('/video-jobs'))return{jobs:[]};if(p.endsWith('/assets'))return{assets:[]};if(p.endsWith('/executions'))return{executions:[]};if(/\/tasks\/p[12]$/.test(p))return{task:{id:p.endsWith('p2')?'p2':'p1',title:'既有CM项目',projectType:'comic',messages:[{role:'assistant',content:'已有CM回复'}]}};return{};
});});
afterEach(async()=>{await act(async()=>root.unmount());vi.unstubAllGlobals();});
async function mount(){await act(async()=>root.render(<Workspace/>));}
function button(label:string){const b=[...container.querySelectorAll('button')].find(v=>v.textContent===label||v.getAttribute('aria-label')===label);expect(b,label).toBeTruthy();return b!;}
async function click(label:string){await act(async()=>button(label).click());}
async function input(id:string,value:string){const e=container.querySelector('#'+id) as HTMLInputElement;await act(async()=>{e.value=value;e.dispatchEvent(new Event('input',{bubbles:true}));});}
async function open(){await click('既有CM项目');}
test('CM loads real online projects with no Studio calls or document editor',async()=>{await mount();expect(container.textContent).toContain('既有CM项目');expect(container.querySelector('textarea')).toBeNull();expect(h.api.mock.calls.every(([p])=>!p.startsWith('/api/studio/'))).toBe(true);await open();expect(container.textContent).toContain('已有CM回复');});
test('cancel a single node edit does not mutate or save canvas',async()=>{await mount();await open();await click('节点');await click('编辑节点');await input('cm-node-title','取消改稿');await click('取消');expect(container.textContent).not.toContain('取消改稿');expect(h.api.mock.calls.some(([,o])=>o?.method==='PATCH')).toBe(false);});
test('node editing preserves canvas metadata and saved revision',async()=>{await mount();await open();await click('节点');await click('编辑节点');await input('cm-node-title','改后封面');await click('保留节点修改');await click('保存画布修改');const write=h.api.mock.calls.find(([p,o])=>p.endsWith('/canvas')&&o?.method==='PATCH');expect(write![1].data.revision).toBe(4);expect(write![1].data.canvas.viewport.zoom).toBe(0.8);expect(write![1].data.canvas.edges[0]).toEqual(document().edges[0]);expect(write![1].data.canvas.nodes[0]).toMatchObject({title:'改后封面',version:4,config:{parameters:{ratio:'9:16'}}});});
test('a canvas conflict retains the local draft',async()=>{const normal=h.api.getMockImplementation()!;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/canvas')&&o?.method==='PATCH')throw Object.assign(Error('版本冲突'),{statusCode:409});return normal(p,o);});await mount();await open();await click('节点');await click('编辑节点');await input('cm-node-title','保留草稿');await click('保留节点修改');await click('保存画布修改');expect(container.textContent).toContain('版本冲突');expect(container.textContent).toContain('保留草稿');expect(button('保存画布修改')).toBeTruthy();});
test('model generation double taps submit one real node request',async()=>{await mount();await open();await click('节点');await click('生成图片');const s=container.querySelector('#cm-generation-model-select') as HTMLSelectElement;await act(async()=>{s.value='0';s.dispatchEvent(new Event('change',{bubbles:true}));});const b=button('确认生成');await act(async()=>{b.click();b.click();});const calls=h.api.mock.calls.filter(([p,o])=>p.endsWith('/image-generations')&&o?.method==='POST');expect(calls).toHaveLength(1);expect(calls[0][1].data).toMatchObject({nodeId:'n1',nodeVersion:3,modelId:'channel-1'});});
test('new project does not automatically invoke CM chat or media',async()=>{await mount();await click('新建画布项目');expect(h.api.mock.calls.filter(([p,o])=>p==='/api/agent/tasks'&&o?.method==='POST')).toHaveLength(1);expect(h.api.mock.calls.some(([p])=>p.endsWith('/chat')||p.endsWith('/image-generations'))).toBe(false);});

test('a read during node editing retains the original canvas revision as the conflict boundary',async()=>{await mount();await open();await click('节点');await click('编辑节点');await input('cm-node-title','旧基线改稿');const normal=h.api.getMockImplementation()!;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/canvas')&&!o?.method){const c=document();c.revision=9;c.nodes[0].version=8;c.nodes[0].title='线上新结果';return{canvas:c};}return normal(p,o);});await act(async()=>h.show());await click('保留节点修改');await click('保存画布修改');const write=h.api.mock.calls.find(([p,o])=>p.endsWith('/canvas')&&o?.method==='PATCH');expect(write![1].data.revision).toBe(4);});
test('the payment confirmation names the actual count and aspect ratio',async()=>{await mount();await open();await click('节点');await click('生成图片');const s=container.querySelector('#cm-generation-model-select') as HTMLSelectElement;await act(async()=>s.dispatchEvent(new Event('change',{bubbles:true})));expect(container.textContent).toContain('1 份');expect(container.textContent).toContain('9:16');await click('确认生成');expect(h.modal.mock.calls.some(([o])=>o.content.includes('1 份')&&o.content.includes('9:16'))).toBe(true);});

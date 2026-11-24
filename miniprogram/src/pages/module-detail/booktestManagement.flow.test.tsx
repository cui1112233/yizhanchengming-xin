// @vitest-environment jsdom
import {act} from 'react';
import {createRoot,type Root} from 'react-dom/client';
import {beforeEach,afterEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({modal:vi.fn(),api:vi.fn(),storage:new Map<string,any>()}));
vi.mock('@tarojs/taro',()=>({default:{showModal:h.modal,navigateTo:vi.fn(),getStorageSync:(k:string)=>h.storage.get(k),setStorageSync:(k:string,v:any)=>h.storage.set(k,v),removeStorageSync:(k:string)=>h.storage.delete(k)},useDidShow:()=>{},useDidHide:()=>{}}));
vi.mock('@tarojs/components',()=>({View:'div',ScrollView:({children}:any)=><div>{children}</div>,Text:'span',Button:'button',Input:({onInput,...p}:any)=><input {...p} onInput={e=>onInput?.({detail:{value:e.currentTarget.value}})}/>,Textarea:({onInput,maxlength,...p}:any)=><textarea {...p} onInput={e=>onInput?.({detail:{value:e.currentTarget.value}})}/>,Picker:({children,range=[],onChange,value=0,disabled}:any)=><div>{children}<select disabled={disabled} value={value} onChange={e=>onChange({detail:{value:e.currentTarget.value}})}>{range.map((v:string,i:number)=><option key={i} value={i}>{v}</option>)}</select></div>}));
vi.mock('../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../components/BrandHeader',()=>({default:()=>null}));
vi.mock('../../components/AppIcon',()=>({default:()=>null}));
import {FetchWorkspaceBody} from './FetchWorkspace';
import Management from './BooktestManagement';
import Scope from './FetchAccountScope';
import {invalidateSessionResources} from '../../services/sessionBoundary';
let root:Root,container:HTMLDivElement;
beforeEach(()=>{h.storage.clear();h.api.mockReset();h.modal.mockReset();h.modal.mockResolvedValue({confirm:true});vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT',true);container=document.createElement('div');root=createRoot(container);h.api.mockImplementation(async(p:string)=>p.endsWith('/batches/current')?{batch:{id:'current',bookIds:['book']}}:p.includes('/realtime/status')?{auto_queue_enabled:false,queue_available:true,processing_count:2,pending_count:4}:{items:[{id:'one',name:'参考文章',collection:'references',content:'正文',extra:'保留'}],configRevision:7});});
afterEach(async()=>{await act(async()=>root.unmount());vi.unstubAllGlobals();});
const ui=(self=true)=><Management request={h.api as any} scopeKey="test" isSelf={self} date=""/>;
async function mount(node=ui()){await act(async()=>root.render(node));}
function button(label:string){const b=[...container.querySelectorAll('button')].find(e=>e.textContent?.trim()===label);expect(b,label).toBeTruthy();return b!;}
async function click(label:string){await act(async()=>button(label).click());}
const writes=()=>h.api.mock.calls.filter(([,o])=>o?.method&&o.method!=='GET');
test('opening queue is read only, counts scoped; failed switch recovers after refresh',async()=>{await mount();await click('自动队列');expect(container.textContent).toContain('处理中 2');expect(writes()).toHaveLength(0);const prior=h.api.getMockImplementation()!;h.api.mockImplementation(async(p,o)=>{if(p.endsWith('/automation/start'))throw Error('timeout');return prior(p,o);});await click('开启自动队列');expect(button('开启自动队列').disabled).toBe(true);await click('刷新核对');expect(button('开启自动队列').disabled).toBe(false);});
test('editing a reference preserves collection and unknown fields with server revision',async()=>{await mount();await click('知识库');await act(async()=>[...container.querySelectorAll('button')].find(b=>b.textContent?.startsWith('参考文章'))!.click());await click('保存条目');expect(writes()[0][1].data).toMatchObject({configRevision:7,item:{id:'one',collection:'references',content:'正文',extra:'保留'}});});
test('foreign account cannot create production batch and preview never writes',async()=>{await mount(ui(false));await click('当前批次');expect(button('预览入厂正文').disabled).toBe(true);expect(writes()).toHaveLength(0);});
test('account changes invalidate old scoped request before and after late replies',async()=>{h.api.mockImplementation(async(p:string)=>p==='/api/member/me'?{member:{username:'self',role:'manager'}}:p==='/api/member/team'?{members:[{username:'other',role:'member',boundTo:'self'}]}:{});let current:any;await mount(<Scope>{scope=>{current=scope;return <div>{scope.control}{scope.key}</div>;}}</Scope>);const old=current.request;await act(async()=>{const picker=container.querySelector('select')!;picker.value='1';picker.dispatchEvent(new Event('change',{bubbles:true}));});await expect(old('/api/batch-rewrite/config',{method:'POST'})).rejects.toThrow('切换');expect(writes()).toHaveLength(0);await current.request('/api/batch-rewrite/config');expect(h.api.mock.calls.slice(-1)[0]?.[1].header['X-Qiantie-Novel-Fetch-Account']).toBe('other');await act(async()=>invalidateSessionResources());expect(container.textContent).toContain('登录账号已改变');});

test('a workspace mounted after account lookup loads config and tasks even after native page show already fired',async()=>{h.api.mockImplementation(async(p:string)=>p.includes('/models')?[{id:'m',name:'模型'}]:p.endsWith('/config')?{platforms:[{id:'one',name:'书城'}],app_config:{text_model_id:'m',fetch:{default_max_txt:4000}}}:p.endsWith('/tasks')?{tasks:[]}:p.includes('/jobs/latest')?{jobs:[]}:{schedules:[]});await mount(<FetchWorkspaceBody title="批量测书" compact request={h.api as any}/>);expect(container.textContent).toContain('1 个平台');expect(h.api.mock.calls.some(([p])=>p==='/api/batch-rewrite/tasks')).toBe(true);expect(writes()).toHaveLength(0);});

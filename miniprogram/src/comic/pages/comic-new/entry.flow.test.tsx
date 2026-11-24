// @vitest-environment jsdom
import {act} from 'react';
import {createRoot,type Root} from 'react-dom/client';
import {beforeEach,afterEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({api:vi.fn(),storage:new Map<string,any>()}));
vi.mock('@tarojs/taro',()=>({default:{getStorageSync:(k:string)=>h.storage.get(k),setStorageSync:(k:string,v:any)=>h.storage.set(k,v),removeStorageSync:(k:string)=>h.storage.delete(k)}}));
vi.mock('@tarojs/components',()=>({Button:'button',View:'div',Text:'span'}));
vi.mock('../../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../../components/BrandHeader',()=>({default:()=>null}));
vi.mock('../../../theme',()=>({ThemedPage:({children}:any)=><div>{children}</div>}));
vi.mock('./ManualComicPrivatePage',()=>({default:()=> <div>新的漫画编辑器</div>}));
vi.mock('./LegacyComicPrivatePage',()=>({default:()=> <div>旧自动任务恢复</div>}));
import Page from './index';
import {invalidateSessionResources} from '../../../services/sessionBoundary';
let root:Root,container:HTMLDivElement;
beforeEach(()=>{invalidateSessionResources();h.storage.clear();h.api.mockReset();h.api.mockResolvedValue({member:{username:'owner'}});vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT',true);container=document.createElement('div');root=createRoot(container);});
afterEach(async()=>{await act(async()=>root.unmount());vi.unstubAllGlobals();});
async function mount(){await act(async()=>root.render(<Page/>));}
test('editor waits for authenticated batch ownership and never displays unverified legacy source',async()=>{h.storage.set('comic:active-run',{batchId:'b',bookId:'old',source:'其他账号正文',startedAt:Date.now(),submitted:true});let finish:any;h.api.mockImplementation(async(p:string)=>p==='/api/member/me'?{member:{username:'owner'}}:new Promise(resolve=>{finish=resolve;}));await mount();expect(container.textContent).toContain('正在核对');expect(container.textContent).not.toContain('编辑器');expect(container.textContent).not.toContain('其他账号正文');await act(async()=>finish({batch:{id:'b',books:[{id:'different'}]}}));expect(container.textContent).toContain('新的漫画编辑器');expect(container.textContent).not.toContain('旧自动任务恢复');});
test('account invalidation immediately unmounts the previous comic editor',async()=>{await mount();expect(container.textContent).toContain('新的漫画编辑器');await act(async()=>invalidateSessionResources());expect(container.textContent).not.toContain('新的漫画编辑器');expect(container.textContent).toContain('账号已切换');});
test('failed identity read is visible and retry can enter the editor without creating tasks',async()=>{h.api.mockRejectedValueOnce(Error('网络暂不可用'));await mount();expect(container.textContent).toContain('网络暂不可用');await act(async()=>container.querySelector('button')!.click());expect(container.textContent).toContain('新的漫画编辑器');expect(h.api.mock.calls.every(([,o])=>!o?.method)).toBe(true);});

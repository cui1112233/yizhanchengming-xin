// @vitest-environment jsdom
import {act,useState} from 'react';
import {createRoot,type Root} from 'react-dom/client';
import {beforeEach,afterEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({api:vi.fn(),modal:vi.fn(),storage:new Map<string,any>()}));
vi.mock('@tarojs/taro',()=>({default:{showModal:h.modal,getStorageSync:(k:string)=>h.storage.get(k),setStorageSync:(k:string,v:any)=>h.storage.set(k,v),removeStorageSync:(k:string)=>h.storage.delete(k)},useDidHide:()=>{},useDidShow:()=>{}}));
vi.mock('@tarojs/components',()=>({View:'div',Text:'span',ScrollView:({children}:any)=><div>{children}</div>,Button:'button',Input:({onInput,maxlength,...p}:any)=><input {...p} onInput={e=>onInput?.({detail:{value:e.currentTarget.value}})}/>,Textarea:({onInput,maxlength,...p}:any)=><textarea {...p} onInput={e=>onInput?.({detail:{value:e.currentTarget.value}})}/>,Picker:({children}:any)=><div>{children}</div>}));
vi.mock('../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../components/AppIcon',()=>({default:()=>null}));
vi.mock('./ScriptEntityMedia',()=>({default:()=>null}));
import Assets from './ScriptAssetsEditor';
import useDraft from './useScriptDraft';
import {invalidateSessionResources} from '../../services/sessionBoundary';
import {authStorageKey} from '../../config/localAuth';
let root:Root,container:HTMLDivElement;
beforeEach(()=>{invalidateSessionResources();h.storage.clear();h.api.mockReset();h.modal.mockReset();h.modal.mockResolvedValue({confirm:true});h.api.mockResolvedValue({member:{username:'owner'}});vi.stubGlobal('IS_REACT_ACT_ENVIRONMENT',true);container=document.createElement('div');root=createRoot(container);});
afterEach(async()=>{await act(async()=>root.unmount());vi.unstubAllGlobals();});
async function mount(node:any){await act(async()=>root.render(node));}
async function click(label:string){await act(async()=>{const button=[...container.querySelectorAll('button')].find(v=>v.textContent===label);expect(button,label).toBeTruthy();button!.click();});}
function Editor(){const [value,set]=useState<any>({characters:[{id:'a',data:{名称:'小明',外形:''}}],scenes:[]});return <Assets value={value} onChange={set} model="" novel="原文" preset="preset" textModel="text-model"/>;}
test('enrichment fills blanks, preserves name and displays original evidence and uncertainty',async()=>{h.api.mockResolvedValue({enrichment:{fields:{名称:'别名',外形:'黑发'},evidence:['原文说他留黑发'],uncertainties:['年龄不明']}});await mount(<Editor/>);await click('根据小说智能补全');expect([...container.querySelectorAll('input')].map(e=>e.value)).toContain('小明');expect([...container.querySelectorAll('textarea')].map(e=>e.value)).toContain('黑发');expect(container.textContent).toContain('年龄不明');expect(h.api.mock.calls[0][1].data).toMatchObject({promptType:'entity_enrich',entityType:'character',textModelId:'text-model'});});
test('cancelled enrichment calls no paid endpoint',async()=>{h.modal.mockResolvedValue({confirm:false});await mount(<Editor/>);await click('根据小说智能补全');expect(h.api).not.toHaveBeenCalled();});
function Draft(){const [state,set]=useState({novelText:''}),draft=useDraft({state,result:'',savedHistoryId:''},'',v=>set(v.state));return <div><span>{state.novelText}</span><span>{draft.status}</span>{draft.recovery?<button onClick={draft.restore}>恢复</button>:null}</div>;}
test('draft is restored only on explicit action and invalidation clears the pending recovery',async()=>{const prefix=encodeURIComponent(authStorageKey('script:owner:new')),key=`yzcm.novel.v1:${prefix}:draft`;h.storage.set(key,{state:{novelText:'已存原文'},result:''});await mount(<Draft/>);expect(container.textContent).not.toContain('已存原文');await click('恢复');expect(container.textContent).toContain('已存原文');await act(async()=>invalidateSessionResources());expect(container.textContent).toContain('旧草稿已隔离');});
test('unknown paid enrichment blocks replay after closing and reopening same entity',async()=>{h.api.mockRejectedValue(Error('timeout'));await mount(<Editor/>);await click('根据小说智能补全');await mount(<div/>);await mount(<Editor/>);await click('根据小说智能补全');expect(h.api).toHaveBeenCalledTimes(1);expect(container.textContent).toContain('阻止重复调用');});

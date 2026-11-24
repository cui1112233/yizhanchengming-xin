import {beforeEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({slots:[] as any[],cursor:0,effects:[] as Array<()=>any>,api:vi.fn(),modal:vi.fn(),revision:0}));
vi.mock('react',()=>({
 useState:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=typeof initial==='function'?initial():initial;return [h.slots[i],(next:any)=>{h.slots[i]=typeof next==='function'?next(h.slots[i]):next;}];},
 useRef:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]={current:initial};return h.slots[i];},
 useEffect:(fn:any,deps:any[])=>{const i=h.cursor++,last=h.slots[i];if(!last||deps.some((v,n)=>v!==last[n])){h.slots[i]=deps;h.effects.push(fn);}},
}));
vi.mock('@tarojs/taro',()=>({default:{showModal:h.modal},useDidShow:()=>{},useDidHide:()=>{}}));
vi.mock('@tarojs/components',()=>Object.fromEntries(['Button','Input','Textarea','Picker','Text','View','ScrollView'].map(n=>[n,n])));
vi.mock('../../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../../services/sessionBoundary',()=>({sessionRevision:()=>h.revision,onSessionInvalidated:()=>()=>{}}));
import Create from './NovelCreate';
import Studio from './NovelStudio';
import Project from './NovelProject';
import {createNovelJournal} from '../../../services/novelRecovery';
import type {NovelMode} from '../../../services/novelCreationInputs';
const storage=new Map<string,any>();
const journal=createNovelJournal({get:key=>storage.get(key),set:(key,value)=>storage.set(key,value),remove:key=>storage.delete(key)},'test-account');
let component:any,props:any;
function render(){h.cursor=0;return component(props);}
function nodes(tree:any):any[]{if(!tree||typeof tree!=='object')return [];if(typeof tree.type==='function')return nodes(tree.type(tree.props));return [tree,...[tree.props?.children].flat(Infinity).flatMap(nodes)];}
function text(tree:any):string{return typeof tree==='string'||typeof tree==='number'?String(tree):tree?.props?[tree.props.children].flat(Infinity).map(text).join(''):'';}
const all=()=>nodes(render());
function button(label:string){const node=all().find(n=>n.type==='Button'&&text(n)===label);expect(node, label).toBeTruthy();return node;}
function input(id:string,value:string){all().find(n=>n.props.id===id).props.onInput({detail:{value}});}
async function flush(){for(let i=0;i<15;i++){await Promise.resolve();for(const effect of h.effects.splice(0))effect();render();}}
async function mount(next:any,extra:any){component=next;props={models:[{id:'m1',name:'文本模型'}],journal,onBack:vi.fn(),onCreated:vi.fn(),onTransition:vi.fn(),...extra};render();await flush();}
const postCalls=()=>h.api.mock.calls.filter(([,options])=>options?.method==='POST');
beforeEach(()=>{storage.clear();h.slots=[];h.cursor=0;h.effects=[];h.revision=0;h.api.mockReset();h.modal.mockResolvedValue({confirm:true});h.api.mockImplementation(async(path:string)=>path.endsWith('/style-profiles')?{styles:[]}:path.endsWith('/memos')?{memos:[]}:{});});
test('opening each special mode reads metadata but never creates a work or calls a model',async()=>{
 for(const mode of ['long_imitation','short_imitation','fanfiction','continuation'] as NovelMode[]){h.slots=[];h.effects=[];await mount(Create,{mode,projects:[]});expect(postCalls()).toHaveLength(0);}
});
test('imitation double tap submits once and an uncertain result only reads the original request',async()=>{
 await mount(Create,{mode:'long_imitation',projects:[]});input('novel-create-referenceText','文'.repeat(100));
 h.api.mockImplementation(async(path:string,options:any)=>{if(options?.method==='POST')throw Error('timeout');if(path.includes('/requests/'))throw Object.assign(Error('not found'),{statusCode:404});return {};});
 const click=all().find(n=>n.props.id==='novel-create-submit').props.onClick;click();click();await flush();
 expect(postCalls()).toHaveLength(1);const original=postCalls()[0][1].data;expect(original).toMatchObject({targetWordCount:5000,targetChapters:100});
 button('核对原提交').props.onClick();await flush();expect(postCalls()).toHaveLength(1);
 button('以原编号重试').props.onClick();await flush();expect(postCalls()).toHaveLength(2);expect(postCalls()[1][1].data).toEqual(original);
});
test('server validation rejection unlocks the form and retains the source',async()=>{
 await mount(Create,{mode:'short_imitation',projects:[]});input('novel-create-referenceText','原'.repeat(100));h.api.mockRejectedValue(Object.assign(Error('输入不正确'),{statusCode:400}));
 button('开始创作').props.onClick();await flush();expect(all().find(n=>n.props.id==='novel-create-referenceText').props).toMatchObject({value:'原'.repeat(100),disabled:false});expect(button('开始创作').props.disabled).toBe(false);
});
test('failed local request persistence prevents all server writes',async()=>{
 const blocked={...journal,save:()=>{throw Error('存储已满');}};await mount(Create,{mode:'continuation',projects:[],journal:blocked});input('novel-create-title','故事');input('novel-create-referenceText','第1章 出发\n正文');
 button('保存正式章节').props.onClick();await flush();expect(postCalls()).toHaveLength(0);
});
test('original discussion uses one studio creation and one message even on double tap',async()=>{
 await mount(Studio,{mode:'long_original',sessionId:''});input('novel-studio-content','一位灯塔守夜人的故事');
 h.api.mockImplementation(async(path:string,options:any)=>path.endsWith('/studio-sessions')?{session:{id:'s1'}}:path.endsWith('/messages')?{session:{id:'s1'},turn:{state:'completed',idempotencyKey:options.data.idempotencyKey},messages:[]}:{session:{id:'s1'},messages:[]});
 const click=button('发送').props.onClick;click();click();await flush();expect(postCalls().map(([path])=>path)).toEqual(['/api/novel-creation/studio-sessions','/api/novel-creation/studio-sessions/s1/messages']);
 expect(all().find(n=>n.props.id==='novel-studio-content').props.value).toBe('');
});
test('chapter version conflict preserves the edited draft and never automatically retries',async()=>{
 const chapter={id:'c1',title:'第1章',content:'原正文',version:4};h.api.mockImplementation(async(path:string,options:any)=>{if(options?.method==='PUT')throw Object.assign(Error('conflict'),{statusCode:409});if(path.endsWith('/chapters'))return {chapters:[chapter]};if(path.endsWith('/sessions'))return {sessions:[{id:'s1',mode:'continuation'}]};return {session:{id:'s1'},messages:[]};});
 await mount(Project,{project:{id:'p1',title:'故事'},initialSession:{id:'s1',mode:'continuation'},onCreate:vi.fn()});button('已保存章节 · 1').props.onClick();button('编辑正文').props.onClick();
 const editor=all().find(n=>n.type==='Textarea'&&n.props.value==='原正文');editor.props.onInput({detail:{value:'保留的新正文'}});button('保存新版本').props.onClick();await flush();
 expect(all().some(n=>n.type==='Textarea'&&n.props.value==='保留的新正文')).toBe(true);expect(h.api.mock.calls.filter(([,o])=>o?.method==='PUT')).toHaveLength(1);
});
test('account switch discards late studio creation and prevents its model message',async()=>{
 await mount(Studio,{mode:'short_original',sessionId:''});input('novel-studio-content','一个短篇故事');let resolve:any;h.api.mockImplementation(()=>new Promise(r=>{resolve=r;}));
 button('发送').props.onClick();await flush();h.revision++;resolve({session:{id:'old-account-session'}});await flush();expect(postCalls()).toHaveLength(1);expect(postCalls()[0][0]).toBe('/api/novel-creation/studio-sessions');
});
test('unknown chapter request checks the snapshot before allowing an explicit retry with the same key',async()=>{
 const data={content:'继续故事',modelId:'m1',idempotencyKey:'original-key',modeInputs:{previousText:'前文'}};
 journal.save('project:p1:s1',{draft:{content:'继续故事',modelId:'m1'},intent:{sessionId:'s1',data}});
 h.api.mockImplementation(async(path:string,options:any)=>{if(options?.method==='POST')return {run:{id:'r1',status:'queued',idempotencyKey:'original-key'}};if(path.endsWith('/chapters'))return {chapters:[{number:1,content:'前文'}]};if(path.endsWith('/sessions'))return {sessions:[{id:'s1',mode:'continuation'}]};return {session:{id:'s1'},messages:[]};});
 await mount(Project,{project:{id:'p1',title:'故事'},initialSession:{id:'s1',mode:'continuation'},onCreate:vi.fn()});
 expect(postCalls()).toHaveLength(0);button('核对原请求').props.onClick();await flush();expect(postCalls()).toHaveLength(0);
 button('以原编号重试').props.onClick();await flush();expect(postCalls()).toHaveLength(1);expect(postCalls()[0][1].data).toEqual(data);
});
test('unknown studio confirmation recovers the bound work by reading without confirming again',async()=>{
 const action={id:'a1',status:'pending',book:{title:'原方案'}};journal.save('studio-action:s1:a1',{actionId:'a1',modelId:'m1',thinkingLevel:'standard'});
 let bound=false;h.api.mockImplementation(async()=>({session:{id:'s1',...(bound?{projectId:'p1',productionSessionId:'prod1',productionRunId:'r1'}:{})},pendingAction:action,messages:[]}));
 await mount(Studio,{mode:'long_original',sessionId:'s1'});bound=true;button('确认方案').props.onClick();await flush();expect(postCalls()).toHaveLength(0);expect(props.onTransition).toHaveBeenCalledWith(expect.objectContaining({project:expect.objectContaining({id:'p1'})}));
});

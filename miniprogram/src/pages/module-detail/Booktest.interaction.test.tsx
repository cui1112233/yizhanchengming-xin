import {beforeEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({slots:[] as any[],cursor:0,effects:[] as Array<()=>any>,api:vi.fn(),show:()=>{}}));
vi.mock('react',()=>({useState:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=typeof initial==='function'?initial():initial;return [h.slots[i],(next:any)=>{h.slots[i]=typeof next==='function'?next(h.slots[i]):next;}];},useRef:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]={current:initial};return h.slots[i];},useEffect:(fn:any,deps:any[])=>{const i=h.cursor++,last=h.slots[i];if(!last||deps.some((v,n)=>v!==last[n])){h.slots[i]=deps;h.effects.push(fn);}}}));
vi.mock('@tarojs/taro',()=>({default:{showToast:vi.fn(),showModal:()=>Promise.resolve({confirm:true})},useDidHide:()=>{}, useDidShow:(fn:any)=>{h.show=fn;}}));
vi.mock('@tarojs/components',()=>Object.fromEntries(['Button','Checkbox','CheckboxGroup','Input','Picker','ScrollView','Text','Textarea','View'].map(n=>[n,n])));
vi.mock('../../theme',()=>({withThemedPage:(component:any)=>component}));
vi.mock('../../components/BrandHeader',()=>({default:'Header'}));
vi.mock('../../components/AppIcon',()=>({default:'Icon'}));
vi.mock('./FetchMaintenancePanel',()=>({default:()=>null}));
vi.mock('../../services/api',()=>({apiRequest:h.api}));
import Workspace from './FetchWorkspace';
let props:any;
function render(){h.cursor=0;return Workspace(props);}
function nodes(tree:any):any[]{return !tree||typeof tree!=='object'?[]:[tree,...[tree.props?.children].flat(Infinity).flatMap(nodes)];}
function text(tree:any):string{return typeof tree==='string'||typeof tree==='number'?String(tree):tree?.props?[tree.props.children].flat(Infinity).map(text).join(''):'';}
const all=()=>nodes(render());
function click(label:string){const node=all().find(n=>n.type==='Button'&&text(n)===label);expect(node,label).toBeTruthy();return node.props.onClick();}
async function flush(){for(let i=0;i<18;i++){await Promise.resolve();for(const effect of h.effects.splice(0))effect();render();}}
async function mount(){render();h.show();await flush();}
const config={platforms:[{id:'p1',name:'书城甲'}],app_config:{text_model_id:'m1',fetch:{default_max_txt:4000}}};
beforeEach(()=>{props={title:'批量测书',compact:true};h.slots=[];h.cursor=0;h.effects=[];h.api.mockReset();h.api.mockImplementation(async(path:string)=>path.includes('/models')?[{id:'m1',name:'文本模型'}]:path.endsWith('/config')?config:path.endsWith('/tasks')?{tasks:[]}:{schedules:[]});});
test('compact testing opens its own title and hides maintenance tools until requested without writes',async()=>{
 await mount();expect(all().find(n=>n.type==='Header').props.title).toBe('批量测书');expect(text(render())).not.toContain('流程与抓取配置');
 click('更多工具 · 定时、规则与提交');expect(text(render())).toContain('流程与抓取配置');click('收起更多工具');expect(text(render())).not.toContain('流程与抓取配置');
 expect(h.api.mock.calls.some(([,o])=>o?.method==='POST')).toBe(false);
});
test('loading and failed reads cannot claim the server has no tasks',async()=>{
 h.api.mockImplementation((path:string)=>path.endsWith('/tasks')?Promise.reject(Error('任务读取失败')):Promise.resolve(path.includes('/models')?[]:path.endsWith('/config')?config:{schedules:[]}));
 render();click('任务');expect(text(render())).toContain('正在读取任务');expect(text(render())).not.toContain('暂无测书任务');await flush();
 expect(text(render())).toContain('任务读取失败');expect(text(render())).not.toContain('暂无测书任务');
});
test('a real task link opens only the matching book detail including independent raw text counts',async()=>{
 props.initialTaskId='bookB';h.api.mockImplementation(async(path:string)=>path.endsWith('/tasks')?{tasks:[{book_id:'bookA',book_name:'甲'},{book_id:'bookB',book_name:'乙'}]}:path.endsWith('/bookB')?{original:'处理正文',original_raw:'完整原文',processed_chars:4000,original_raw_chars:27831}:path.includes('/models')?[]:path.endsWith('/config')?config:{schedules:[]});
 await mount();expect(h.api.mock.calls.filter(([path])=>path.endsWith('/bookB'))).toHaveLength(1);expect(h.api.mock.calls.some(([path])=>path.endsWith('/bookA'))).toBe(false);
 expect(text(render())).toContain('4000 / 27831');expect(all().filter(n=>n.type==='Textarea').map(n=>n.props.value)).toEqual(['处理正文','完整原文']);
 click('关闭');await flush();expect(text(render())).not.toContain('处理后原文');
});
test('missing task ids show an explicit message instead of opening a different book',async()=>{
 props.initialTaskId='missing';h.api.mockImplementation(async(path:string)=>path.endsWith('/tasks')?{tasks:[{book_id:'other',book_name:'其他书'}]}:path.includes('/models')?[]:path.endsWith('/config')?config:{schedules:[]});
 await mount();expect(text(render())).toContain('未找到指定任务');expect(h.api.mock.calls.some(([path])=>path.endsWith('/other'))).toBe(false);
});
test('closing advanced tools preserves the input and current platform selection',async()=>{
 await mount();all().find(n=>n.type==='Textarea').props.onInput({detail:{value:'123456 一本小说'}});
 click('更多工具 · 定时、规则与提交');click('收起更多工具');expect(all().find(n=>n.type==='Textarea').props.value).toBe('123456 一本小说');
 expect(all().find(n=>n.type==='Picker').props.value).toBe(0);
});
test('starting real testing sends the visible AI slot method through the async interface once',async()=>{
 h.api.mockImplementation(async(path:string)=>path.endsWith('/jobs/latest?limit=10')?{jobs:[]}:path.endsWith('/process/start')||path.endsWith('/jobs/go-workshop-1')?{id:'go-workshop-1',status:'done'}:path.includes('/models')?[{id:'m1',name:'模型'}]:path.endsWith('/config')?config:path.endsWith('/tasks')?{tasks:[]}:{schedules:[]});
 await mount();all().find(n=>n.type==='Textarea').props.onInput({detail:{value:'123456'}});
 all().find(n=>n.props.id==='booktest-versions').props.onClick();all().find(n=>n.type==='CheckboxGroup').props.onChange({detail:{value:['ai2']}});
 all().find(n=>n.type==='Picker'&&n.props.range?.includes('批量改文指令库')).props.onChange({detail:{value:3}});click('收起面板');
 const start=all().find(n=>n.type==='Button'&&text(n)==='开始处理').props.onClick;await Promise.all([start(),start()]);await flush();
 const writes=h.api.mock.calls.filter(([path])=>path.endsWith('/process/start'));expect(writes).toHaveLength(1);expect(writes[0][1].data).toMatchObject({selected_versions:['ai2'],rewrite_methods:{ai2:'instruction'},text_model_id:'m1'});
 expect(h.api.mock.calls.some(([path])=>path==='/api/batch-rewrite/process')).toBe(false);
});
test('unknown async start results never automatically replay on refresh or another click',async()=>{
 h.api.mockImplementation(async(path:string)=>{if(path.endsWith('/process/start'))throw Error('timeout');return path.includes('/jobs/latest')?{jobs:[]}:path.includes('/models')?[{id:'m1',name:'模型'}]:path.endsWith('/config')?config:path.endsWith('/tasks')?{tasks:[]}:{schedules:[]};});
 await mount();all().find(n=>n.type==='Textarea').props.onInput({detail:{value:'123456'}});await click('开始处理');await flush();expect(text(render())).toContain('结果未知');await click('开始处理');await flush();
 expect(h.api.mock.calls.filter(([path])=>path.endsWith('/process/start'))).toHaveLength(1);
});
test('mobile scheduling sends flat execution settings and preserves an unchecked automatic submission choice',async()=>{
 await mount();all().find(n=>n.type==='Textarea').props.onInput({detail:{value:'123456'}});click('更多工具 · 定时、规则与提交');
 all().find(n=>n.type==='Button'&&text(n).startsWith('开始定时')).props.onClick();
 all().find(n=>n.type==='Picker'&&n.props.mode==='date').props.onChange({detail:{value:'2099-01-01'}});
 all().find(n=>n.type==='Picker'&&n.props.mode==='time').props.onChange({detail:{value:'12:30'}});
 await click('创建定时任务');await flush();
 const write=h.api.mock.calls.find(([path,options])=>path.endsWith('/schedules')&&options?.method==='POST');expect(write?.[1].data).toMatchObject({inputText:'123456',platform_id:'p1',text_model_id:'m1',selected_versions:['original','ai1'],auto_submit:false});expect(write?.[1].data).not.toHaveProperty('inputSnapshot');
});

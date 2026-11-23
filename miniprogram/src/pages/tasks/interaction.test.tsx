import {beforeEach,expect,test,vi} from 'vitest';
import {readFileSync} from 'node:fs';
import {resolve} from 'node:path';
const h=vi.hoisted(()=>({slots:[] as any[],cursor:0,show:undefined as any,load:vi.fn(),navigate:vi.fn()}));
vi.mock('react',()=>({useState:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=typeof initial==='function'?initial():initial;return [h.slots[i],(next:any)=>{h.slots[i]=typeof next==='function'?next(h.slots[i]):next;}];},useRef:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]={current:initial};return h.slots[i];},useEffect:vi.fn()}));
vi.mock('@tarojs/taro',()=>({default:{navigateTo:h.navigate},useDidShow:(fn:any)=>{h.show=fn;}}));
vi.mock('@tarojs/components',()=>({View:'View',Text:'Text',Button:'Button',Image:'Image'}));
vi.mock('../../services/api',()=>({apiRequest:vi.fn()}));
vi.mock('../../config/runtime',()=>({isApiOriginConfigured:()=>true,API_ORIGIN_CONFIGURATION_ERROR:''}));
vi.mock('../../services/taskSummary',async original=>({...await original<any>(),loadTaskSummary:h.load}));
vi.mock('../../theme',()=>({ThemedPage:'ThemedPage'}));
vi.mock('../../components/BrandHeader',()=>({default:'BrandHeader'}));
vi.mock('../../components/AmbientStarfield',()=>({default:'AmbientStarfield'}));
vi.mock('../../components/AppIcon',()=>({default:'AppIcon'}));
vi.mock('../functions/functionVisuals',()=>({FUNCTION_VISUALS:{script:{asset:'script.webp'},agent:{asset:'agent.webp'},fetch:{asset:'fetch.webp'},batch:{asset:'batch.webp'},shuihuo:{asset:'shuihuo.webp'}}}));
import TasksPage from './index';
const nodes=(tree:any):any[]=>!tree||typeof tree!=='object'?[]:[tree,...[tree.props?.children].flat(Infinity).flatMap(nodes)];
const render=()=>{h.cursor=0;return nodes(TasksPage());};
const ready=async()=>{render();h.show();for(let i=0;i<12;i++)await Promise.resolve();return render();};
const content=(node:any):string=>typeof node==='string'||typeof node==='number'?String(node):!node||typeof node!=='object'?'':[node.props?.children].flat(Infinity).map(content).join('');
beforeEach(()=>{h.slots=[];h.navigate.mockReset();h.load.mockReset().mockResolvedValue({sources:[{source:'fetch',state:'ready'}],entries:[
 {source:'fetch',sourceLabel:'小说获取',id:'run',title:'正在生成',status:'running',progress:{completed:2,total:5}},
 {source:'fetch',sourceLabel:'小说获取',id:'review',title:'等待确认',status:'attention'},
 {source:'fetch',sourceLabel:'小说获取',id:'done',title:'已完成项目',status:'completed'},
 {source:'fetch',sourceLabel:'小说获取',id:'a/b',title:'失败项目',status:'failed'},
]});});
test('task center renders real counters, approved filters, and one shared list surface',async()=>{
 const tree=await ready();
 expect(tree.filter(n=>n.props?.className==='task-studio-hero-art')).toHaveLength(1);
 expect(tree.filter(n=>n.props?.className==='task-studio-mascot-shell')).toHaveLength(0);
 expect(tree.filter(n=>n.props?.className?.includes('task-studio-count-item'))).toHaveLength(3);
 expect(tree.filter(n=>n.props?.className?.includes('task-studio-count-art'))).toHaveLength(3);
 expect(tree.filter(n=>n.props?.className==='task-studio-count-value').map(content)).toEqual(['1','1','1']);
 expect(tree.filter(n=>n.props?.className?.includes('task-studio-count-item')).map(content)).toEqual(['1进行中','1待确认','1已完成']);
 expect(tree.filter(n=>n.props?.className?.split(' ').includes('task-studio-filter')).map(content)).toEqual(['全部','进行中','待确认','已完成']);
 expect(tree.filter(n=>n.props?.className?.includes('task-studio-list'))).toHaveLength(1);
 expect(tree.filter(n=>n.props?.className?.split(' ').includes('task-studio-row'))).toHaveLength(4);
});
test('progress is rendered only when the server returns a valid total',async()=>{
 const tree=await ready();
 expect(tree.filter(n=>n.props?.className==='task-studio-progress-track')).toHaveLength(1);
 expect(content(tree.find(n=>n.props?.className==='task-studio-progress-value'))).toBe('40%');
});
test('flat record opens the existing encoded module destination',async()=>{
 const tree=await ready();expect(tree.filter(n=>n.props?.className?.split(' ').includes('task-studio-row'))).toHaveLength(4);
 tree.find(n=>n.props?.className?.split(' ').includes('task-studio-row')&&content(n).includes('失败项目')).props.onClick();
 expect(h.navigate).toHaveBeenCalledWith({url:'/pages/module-detail/index?key=fetch&taskId=a%2Fb'});
});
test('a record without an ID cannot open an empty or unrelated destination',async()=>{
 h.load.mockResolvedValue({sources:[{source:'fetch',state:'ready'}],entries:[{source:'fetch',sourceLabel:'Fetch',title:'Missing ID',status:'unknown'}]});
 const tree=await ready();const action=tree.find(n=>n.props?.className?.split(' ').includes('task-studio-row'));expect(action.props.className).toContain('is-disabled');action.props.onClick();expect(h.navigate).not.toHaveBeenCalled();
});
test('task content stays above the dynamic glass canvas in WeChat',()=>{
 const style=readFileSync(resolve(__dirname,'tasks-studio.scss'),'utf8');
 expect(style).toMatch(/\.task-studio-header-scene,.task-studio-body,.task-studio-state\s*\{[^}]*position:relative[^}]*z-index:1/);
});
test('task center keeps the compact phone scale from the approved reference',()=>{
 const style=readFileSync(resolve(__dirname,'tasks-studio.scss'),'utf8');
 expect(style).toMatch(/\.task-studio-counts\s*\{[^}]*min-height:72PX/);
 expect(style).toMatch(/\.task-studio-segments\s*\{[^}]*height:42PX/);
 expect(style).toMatch(/\.task-studio-row\s*\{[^}]*min-height:76PX/);
 expect(style).toMatch(/\.task-studio-row-title\s*\{[^}]*font-size:14PX/);
});

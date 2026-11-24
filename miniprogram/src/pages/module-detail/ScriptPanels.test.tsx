import { beforeEach, expect, test, vi } from 'vitest';
const h = vi.hoisted(() => ({ slots: [] as any[], cursor: 0, seen: new Set<number>(), effects: [] as any[], cleanups: [] as any[], api: vi.fn() }));
vi.mock('react', () => ({ useState: (v:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=typeof v==='function'?v():v;return[h.slots[i],(next:any)=>{h.slots[i]=typeof next==='function'?next(h.slots[i]):next;}];}, useRef: (v:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]={current:v};return h.slots[i];}, useEffect:(fn:any)=>{const i=h.cursor++;if(!h.seen.has(i)){h.seen.add(i);h.effects.push(fn);}} }));
vi.mock('@tarojs/components',()=>Object.fromEntries(['Button','Input','Picker','Text','Textarea','View','ScrollView'].map(name=>[name,name])));
vi.mock('../../services/api',()=>({apiRequest:h.api}));
vi.mock('../../components/AppIcon',()=>({default:'Icon'}));
vi.mock('./ScriptEntityMedia',()=>({default:'Media'}));
import ScriptAssetsEditor from './ScriptAssetsEditor';
import ScriptHistoryPanel from './ScriptHistoryPanel';
let page:()=>any;
const render=()=>{h.cursor=0;const tree=page();while(h.effects.length){const off=h.effects.shift()();if(off)h.cleanups.push(off);}return tree;};
const nodes=(t:any):any[]=>t&&typeof t==='object'?[t,...[t.props?.children].flat(Infinity).flatMap(nodes)]:[];
const text=(t:any):string=>typeof t==='string'||typeof t==='number'?String(t):t?.props?[t.props.children].flat(Infinity).map(text).join(''):'';
const flush=async()=>{for(let i=0;i<25;i++)await Promise.resolve();};
const button=(label:string)=>nodes(render()).find(n=>n.type==='Button'&&text(n)===label);
beforeEach(()=>{h.slots=[];h.cursor=0;h.seen.clear();h.effects=[];h.cleanups=[];h.api.mockReset();});
test('manual asset add, rename and removal preserve nested unknown fields and images',()=>{
  let value:any={characters:[{id:'c1',data:{名称:'林夏',身份:'店主',custom:{keep:true}},imageUrls:['/real.png'],mainImageUrl:'/real.png'}],scenes:[]};
  page=()=>ScriptAssetsEditor({value,onChange:next=>{value=next;},model:'',novel:'原文',preset:'p'});
  nodes(render()).find(n=>n.props?.id==='script-entity-name').props.onInput({detail:{value:'新名字'}});
  expect(value.characters[0]).toMatchObject({data:{名称:'新名字',custom:{keep:true}},imageUrls:['/real.png'],mainImageUrl:'/real.png'});
  button('添加人物').props.onClick(); expect(value.characters).toHaveLength(2);
  button('移除此人物').props.onClick(); expect(value.characters).toHaveLength(1); expect(value.characters[0].id).toBe('c1');
});
test('history reads the web history endpoint and selection never creates a task',async()=>{
  const select=vi.fn();h.api.mockResolvedValue({entries:[{id:'h1',novelText:'旧故事',format:'shotlist'}]});page=()=>ScriptHistoryPanel({onSelect:select});
  render();await flush();const open=nodes(render()).find(n=>n.type==='Button'&&text(n).includes('旧故事'));open.props.onClick();
  expect(h.api.mock.calls).toEqual([['/api/history']]);expect(select).toHaveBeenCalledWith('h1');
});
test('history errors and malformed replies remain errors, not empty states',async()=>{
  h.api.mockResolvedValue({});page=()=>ScriptHistoryPanel({onSelect:vi.fn()});render();await flush();
  expect(text(render())).toContain('历史列表格式不正确');expect(text(render())).not.toContain('暂无生成历史');
});
test('late history response cannot update a closed panel',async()=>{
  let resolve!:(v:any)=>void;h.api.mockReturnValue(new Promise(done=>{resolve=done;}));page=()=>ScriptHistoryPanel({onSelect:vi.fn()});render();h.cleanups[0]();resolve({entries:[{id:'h1',novelText:'过期记录'}]});await flush();expect(text(render())).not.toContain('过期记录');
});
